package resource_test

import (
	"errors"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"github.com/concourse/atc"
	"github.com/concourse/atc/dbng/dbngfakes"
	. "github.com/concourse/atc/resource"
	"github.com/concourse/atc/worker"
	"github.com/concourse/atc/worker/workerfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ResourceInstance", func() {
	var (
		logger                   lager.Logger
		resourceInstance         ResourceInstance
		fakeWorkerClient         *workerfakes.FakeClient
		fakeResourceCacheFactory *dbngfakes.FakeResourceCacheFactory
		disaster                 error
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")
		fakeWorkerClient = new(workerfakes.FakeClient)
		fakeResourceCacheFactory = new(dbngfakes.FakeResourceCacheFactory)
		disaster = errors.New("disaster")

		resourceInstance = NewBuildResourceInstance(
			"some-resource-type",
			atc.Version{"some": "version"},
			atc.Source{"some": "source"},
			atc.Params{"some": "params"},
			42,
			43,
			atc.ResourceTypes{},
			fakeResourceCacheFactory,
		)
	})

	Describe("FindInitializedOn", func() {
		var (
			foundVolume worker.Volume
			found       bool
			findErr     error
		)

		JustBeforeEach(func() {
			foundVolume, found, findErr = resourceInstance.FindInitializedOn(logger, fakeWorkerClient)
		})

		Context("when failing to find or create cache in database", func() {
			BeforeEach(func() {
				fakeResourceCacheFactory.FindOrCreateResourceCacheForBuildReturns(nil, disaster)
			})

			It("returns the error", func() {
				Expect(findErr).To(Equal(disaster))
			})
		})

		Context("when initialized volume for resource cache exists on worker", func() {
			var fakeVolume *workerfakes.FakeVolume

			BeforeEach(func() {
				fakeVolume = new(workerfakes.FakeVolume)
				fakeWorkerClient.FindInitializedVolumeForResourceCacheReturns(fakeVolume, true, nil)
			})

			It("returns found volume", func() {
				Expect(findErr).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(foundVolume).To(Equal(fakeVolume))
			})
		})

		Context("when initialized volume for resource cache does not exist on worker", func() {
			BeforeEach(func() {
				fakeWorkerClient.FindInitializedVolumeForResourceCacheReturns(nil, false, nil)
			})

			It("does not return any volume", func() {
				Expect(findErr).NotTo(HaveOccurred())
				Expect(found).To(BeFalse())
				Expect(foundVolume).To(BeNil())
			})
		})

		Context("when worker errors in finding the cache", func() {
			BeforeEach(func() {
				fakeWorkerClient.FindInitializedVolumeForResourceCacheReturns(nil, false, disaster)
			})

			It("returns the error", func() {
				Expect(findErr).To(Equal(disaster))
				Expect(found).To(BeFalse())
				Expect(foundVolume).To(BeNil())
			})
		})
	})

	Context("FindOrCreateOn", func() {
		var createdVolume worker.Volume
		var createErr error

		JustBeforeEach(func() {
			createdVolume, createErr = resourceInstance.FindOrCreateOn(logger, fakeWorkerClient)
		})

		Context("when creating the volume succeeds", func() {
			var volume *workerfakes.FakeVolume

			BeforeEach(func() {
				volume = new(workerfakes.FakeVolume)
				fakeWorkerClient.FindOrCreateVolumeForResourceCacheReturns(volume, nil)
			})

			It("succeeds", func() {
				Expect(createErr).ToNot(HaveOccurred())
			})

			It("returns the volume", func() {
				Expect(createdVolume).To(Equal(volume))
			})

			It("created with the right properties", func() {
				_, spec, _ := fakeWorkerClient.FindOrCreateVolumeForResourceCacheArgsForCall(0)
				Expect(spec).To(Equal(worker.VolumeSpec{
					Strategy: worker.ResourceCacheStrategy{
						ResourceHash:    `some-resource-type{"some":"source"}`,
						ResourceVersion: atc.Version{"some": "version"},
					},
					Properties: worker.VolumeProperties{
						"resource-type":    "some-resource-type",
						"resource-version": `{"some":"version"}`,
						"resource-source":  "968e27f71617a029e58a09fb53895f1e1875b51bdaa11293ddc2cb335960875cb42c19ae8bc696caec88d55221f33c2bcc3278a7d15e8d13f23782d1a05564f1",
						"resource-params":  "fe7d9dbc2ac75030c3e8c88e54a33676c38d8d9d2876700bc01d4961caf898e7cbe8e738232e86afcf6a5f64a9527c458a130277b08d72fb339962968d0d0967",
					},
					Privileged: true,
					TTL:        0,
				}))
			})
		})

		Context("when creating the volume fails", func() {
			BeforeEach(func() {
				fakeWorkerClient.FindOrCreateVolumeForResourceCacheReturns(nil, disaster)
			})

			It("returns the error", func() {
				Expect(createErr).To(Equal(disaster))
			})
		})
	})

	Context("ResourceCacheIdentifier", func() {
		It("returns a volume identifier corrsponding to the resource that the identifier is tracking", func() {
			expectedIdentifier := worker.ResourceCacheIdentifier{
				ResourceVersion: atc.Version{"some": "version"},
				ResourceHash:    `some-resource-type{"some":"source"}`,
			}

			Expect(resourceInstance.ResourceCacheIdentifier()).To(Equal(expectedIdentifier))
		})
	})
})

var _ = Describe("GenerateResourceHash", func() {
	It("returns a hash of the source and resource type", func() {
		Expect(GenerateResourceHash(atc.Source{"some": "source"}, "git")).To(Equal(`git{"some":"source"}`))
	})
})
