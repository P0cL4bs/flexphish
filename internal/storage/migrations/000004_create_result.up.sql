CREATE TABLE results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    campaign_id INTEGER NOT NULL,
    session_id TEXT NOT NULL UNIQUE,

    email TEXT,
    username TEXT,
    password TEXT,

    ip TEXT,
    user_agent TEXT,
    country TEXT,
    city TEXT,

    device TEXT,
    os TEXT,
    browser TEXT,

    status TEXT NOT NULL DEFAULT 'in_progress'
        CHECK(status IN ('in_progress','completed','abandoned')),

    first_seen DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_seen DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (campaign_id)
        REFERENCES campaigns(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_results_campaign_id ON results(campaign_id);
CREATE INDEX idx_results_email ON results(email);
CREATE INDEX idx_results_status ON results(status);