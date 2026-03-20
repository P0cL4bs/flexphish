-- =========================
-- SMTP PROFILES: allow global scope
-- =========================
ALTER TABLE smtp_profiles RENAME TO smtp_profiles_old;

CREATE TABLE smtp_profiles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    is_global BOOLEAN DEFAULT 0,
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
    is_global,
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
    user_id,
    0,
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
FROM smtp_profiles_old;

DROP TABLE smtp_profiles_old;

CREATE INDEX idx_smtp_profiles_user_id ON smtp_profiles(user_id);

-- =========================
-- EMAIL TEMPLATES: allow global scope
-- =========================
ALTER TABLE email_templates RENAME TO email_templates_old;

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
    0,
    name,
    subject,
    body,
    created_at,
    updated_at
FROM email_templates_old;

DROP TABLE email_templates_old;

CREATE INDEX idx_email_templates_user_id ON email_templates(user_id);
