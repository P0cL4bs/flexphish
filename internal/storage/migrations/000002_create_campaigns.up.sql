CREATE TABLE campaigns (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,

    name TEXT NOT NULL,
    subdomain TEXT NOT NULL UNIQUE,

    status TEXT NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft','scheduled','active','stopped' ,'completed','cancelled')),

    launch_date DATETIME,
    completed_date DATETIME,

    template_id TEXT NOT NULL,

    -- Tracking flags
    track_opens BOOLEAN NOT NULL DEFAULT 1,
    track_clicks BOOLEAN NOT NULL DEFAULT 1,
    track_geo_location BOOLEAN NOT NULL DEFAULT 1,
    track_device_info BOOLEAN NOT NULL DEFAULT 1,
    track_ip BOOLEAN NOT NULL DEFAULT 1,
    track_user_agent BOOLEAN NOT NULL DEFAULT 1,
    track_referrer BOOLEAN NOT NULL DEFAULT 1,
    enable_fingerprinting BOOLEAN NOT NULL DEFAULT 1,

    -- Totals
    total_sent INTEGER NOT NULL DEFAULT 0,
    total_opened INTEGER NOT NULL DEFAULT 0,
    total_clicked INTEGER NOT NULL DEFAULT 0,
    total_submitted INTEGER NOT NULL DEFAULT 0,

    -- Unique metrics
    unique_opened INTEGER NOT NULL DEFAULT 0,
    unique_clicked INTEGER NOT NULL DEFAULT 0,
    unique_submitted INTEGER NOT NULL DEFAULT 0,

    is_archived BOOLEAN NOT NULL DEFAULT 0,
    deleted_at DATETIME,

    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_campaigns_user_id ON campaigns(user_id);
CREATE INDEX idx_campaigns_status ON campaigns(status);
CREATE INDEX idx_campaigns_deleted_at ON campaigns(deleted_at);