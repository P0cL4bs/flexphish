-- =========================
-- SMTP PROFILES
-- =========================
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

CREATE INDEX idx_smtp_profiles_user_id ON smtp_profiles(user_id);

-- =========================
-- EMAIL TEMPLATES
-- =========================
CREATE TABLE email_templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,

    name TEXT NOT NULL,
    subject TEXT NOT NULL,
    body TEXT NOT NULL,

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_email_templates_user_id ON email_templates(user_id);

-- =========================
-- TARGETS
-- =========================
CREATE TABLE targets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
   
    first_name TEXT,
    last_name TEXT,
    email TEXT NOT NULL,
    position TEXT,

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_targets_user_id ON targets(user_id);
CREATE INDEX idx_targets_email ON targets(email);

-- =========================
-- GROUPS
-- =========================
CREATE TABLE groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    is_global BOOLEAN DEFAULT 0,
    name TEXT NOT NULL,

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_groups_user_id ON groups(user_id);

-- =========================
-- GROUP TARGETS
-- =========================
CREATE TABLE group_targets (
    group_id INTEGER NOT NULL,
    target_id INTEGER NOT NULL,

    PRIMARY KEY (group_id, target_id),

    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (target_id) REFERENCES targets(id) ON DELETE CASCADE
);

-- =========================
-- ALTER CAMPAIGNS (SQLite style)
-- =========================
ALTER TABLE campaigns ADD COLUMN send_emails BOOLEAN DEFAULT 0;
ALTER TABLE campaigns ADD COLUMN smtp_profile_id INTEGER;
ALTER TABLE campaigns ADD COLUMN email_template_id INTEGER;

CREATE INDEX idx_campaigns_smtp_profile_id ON campaigns(smtp_profile_id);
CREATE INDEX idx_campaigns_email_template_id ON campaigns(email_template_id);

-- =========================
-- CAMPAIGN TARGETS
-- =========================
CREATE TABLE campaign_targets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    campaign_id INTEGER NOT NULL,
    target_id INTEGER NOT NULL,

    token TEXT NOT NULL UNIQUE,

    status TEXT,

    email_sent_at DATETIME,
    opened_at DATETIME,
    clicked_at DATETIME,
    submitted_at DATETIME,

    ip TEXT,
    user_agent TEXT,

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE CASCADE,
    FOREIGN KEY (target_id) REFERENCES targets(id) ON DELETE CASCADE
);

CREATE INDEX idx_campaign_targets_campaign_id ON campaign_targets(campaign_id);
CREATE INDEX idx_campaign_targets_target_id ON campaign_targets(target_id);
CREATE INDEX idx_campaign_targets_status ON campaign_targets(status);