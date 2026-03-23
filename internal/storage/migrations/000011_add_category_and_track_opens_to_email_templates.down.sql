ALTER TABLE email_templates RENAME TO email_templates_new;

CREATE TABLE email_templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    is_global BOOLEAN DEFAULT 0,

    name TEXT NOT NULL,
    subject TEXT NOT NULL,
    body TEXT NOT NULL,

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO email_templates (
    id,
    user_id,
    is_global,
    name,
    subject,
    body,
    created_at,
    updated_at
)
SELECT
    id,
    user_id,
    is_global,
    name,
    subject,
    body,
    created_at,
    updated_at
FROM email_templates_new;

DROP TABLE email_templates_new;

CREATE INDEX idx_email_templates_user_id ON email_templates(user_id);
