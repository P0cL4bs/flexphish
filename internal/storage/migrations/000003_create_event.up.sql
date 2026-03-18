CREATE TABLE events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    campaign_id INTEGER NOT NULL,
    result_id INTEGER,

    type TEXT NOT NULL,
    step_id TEXT,
    path TEXT,

    ip TEXT,
    user_agent TEXT,
    referrer TEXT,

    metadata TEXT, -- JSON armazenado como TEXT

    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (campaign_id)
        REFERENCES campaigns(id)
        ON DELETE CASCADE,

    FOREIGN KEY (result_id)
        REFERENCES results(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_events_campaign_id ON events(campaign_id);
CREATE INDEX idx_events_result_id ON events(result_id);
CREATE INDEX idx_events_type ON events(type);
CREATE INDEX idx_events_created_at ON events(created_at);