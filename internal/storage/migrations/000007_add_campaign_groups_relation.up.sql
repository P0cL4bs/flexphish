-- =========================
-- CAMPAIGN GROUPS (pivot)
-- =========================
CREATE TABLE campaign_groups (
    campaign_id INTEGER NOT NULL,
    group_id INTEGER NOT NULL,

    PRIMARY KEY (campaign_id, group_id),

    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
);

-- =========================
-- INDEXES (performance)
-- =========================
CREATE INDEX idx_campaign_groups_campaign_id ON campaign_groups(campaign_id);
CREATE INDEX idx_campaign_groups_group_id ON campaign_groups(group_id);