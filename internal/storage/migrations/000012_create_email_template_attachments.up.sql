CREATE TABLE email_template_attachments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email_template_id INTEGER NOT NULL,

    filename TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    size INTEGER NOT NULL,
    content BLOB NOT NULL,

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (email_template_id) REFERENCES email_templates(id) ON DELETE CASCADE
);

CREATE INDEX idx_email_template_attachments_template_id ON email_template_attachments(email_template_id);
