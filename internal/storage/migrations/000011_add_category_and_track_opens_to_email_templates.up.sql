ALTER TABLE email_templates ADD COLUMN category TEXT NOT NULL DEFAULT '';
ALTER TABLE email_templates ADD COLUMN track_opens BOOLEAN NOT NULL DEFAULT 1;

CREATE INDEX IF NOT EXISTS idx_email_templates_category ON email_templates(category);
