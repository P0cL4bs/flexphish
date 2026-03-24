DROP INDEX IF EXISTS idx_campaigns_email_dispatch_status;

ALTER TABLE campaigns DROP COLUMN email_dispatch_status;
ALTER TABLE campaigns DROP COLUMN email_dispatch_queued_at;
ALTER TABLE campaigns DROP COLUMN email_dispatch_started_at;
ALTER TABLE campaigns DROP COLUMN email_dispatch_completed_at;
ALTER TABLE campaigns DROP COLUMN email_dispatch_last_attempt_at;
ALTER TABLE campaigns DROP COLUMN email_dispatch_last_error;
ALTER TABLE campaigns DROP COLUMN email_dispatch_total_targets;
ALTER TABLE campaigns DROP COLUMN email_dispatch_sent;
ALTER TABLE campaigns DROP COLUMN email_dispatch_failed;
ALTER TABLE campaigns DROP COLUMN email_dispatch_pending;
