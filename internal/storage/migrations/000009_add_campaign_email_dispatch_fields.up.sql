ALTER TABLE campaigns ADD COLUMN email_dispatch_status TEXT DEFAULT 'idle';
ALTER TABLE campaigns ADD COLUMN email_dispatch_queued_at DATETIME;
ALTER TABLE campaigns ADD COLUMN email_dispatch_started_at DATETIME;
ALTER TABLE campaigns ADD COLUMN email_dispatch_completed_at DATETIME;
ALTER TABLE campaigns ADD COLUMN email_dispatch_last_attempt_at DATETIME;
ALTER TABLE campaigns ADD COLUMN email_dispatch_last_error TEXT;

ALTER TABLE campaigns ADD COLUMN email_dispatch_total_targets INTEGER DEFAULT 0;
ALTER TABLE campaigns ADD COLUMN email_dispatch_sent INTEGER DEFAULT 0;
ALTER TABLE campaigns ADD COLUMN email_dispatch_failed INTEGER DEFAULT 0;
ALTER TABLE campaigns ADD COLUMN email_dispatch_pending INTEGER DEFAULT 0;

CREATE INDEX idx_campaigns_email_dispatch_status ON campaigns(email_dispatch_status);
