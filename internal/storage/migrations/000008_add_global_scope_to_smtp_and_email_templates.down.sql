-- =========================
-- SMTP PROFILES: rollback global scope
-- =========================
ALTER TABLE smtp_profiles RENAME TO smtp_profiles_new;

CREATE TABLE smtp_profiles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,

    host TEXT NOT NULL,
    port INTEGER NOT NULL,
    username TEXT NOT NULL,
    password TEXT NOT NULL,

    from_name TEXT,
    from_email TEXT,

    is_active BOOLEAN DEFAULT 1,

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO smtp_profiles (
    id,
    user_id,
    name,
    host,
    port,
    username,
    password,
    from_name,
    from_email,
    is_active,
    created_at,
    updated_at
)
SELECT
    id,
    COALESCE(user_id, 0),
    name,
    host,
    port,
    username,
    password,
    from_name,
    from_email,
    is_active,
    created_at,
    updated_at
FROM smtp_profiles_new;

DROP TABLE smtp_profiles_new;

CREATE INDEX idx_smtp_profiles_user_id ON smtp_profiles(user_id);

-- =========================
-- EMAIL TEMPLATES: rollback global scope
-- =========================
ALTER TABLE email_templates RENAME TO email_templates_new;

CREATE TABLE email_templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,

    name TEXT NOT NULL,
    subject TEXT NOT NULL,
    body TEXT NOT NULL,

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO email_templates (
    id,
    user_id,
    name,
    subject,
    body,
    created_at,
    updated_at
)
SELECT
    id,
    COALESCE(user_id, 0),
    name,
    subject,
    body,
    created_at,
    updated_at
FROM email_templates_new;

DROP TABLE email_templates_new;

CREATE INDEX idx_email_templates_user_id ON email_templates(user_id);
