BEGIN;
  DROP INDEX IF EXISTS order_builds_with_reruns_for_job_idx;
  CREATE INDEX ready_and_scheduled_ordering_with_rerun_builds_idx ON builds (job_id, COALESCE(rerun_of, id) DESC, id DESC) WHERE inputs_ready AND scheduled;
COMMIT;
