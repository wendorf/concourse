BEGIN;
  DROP INDEX IF EXISTS ready_and_scheduled_ordering_with_rerun_builds_idx;
  CREATE INDEX order_builds_with_reruns_for_job_idx on builds (job_id, COALESCE(rerun_of, id) DESC, id DESC);
COMMIT;
