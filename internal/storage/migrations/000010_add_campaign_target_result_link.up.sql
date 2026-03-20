ALTER TABLE campaign_targets ADD COLUMN result_id INTEGER;
CREATE INDEX idx_campaign_targets_result_id ON campaign_targets(result_id);

ALTER TABLE results ADD COLUMN campaign_target_id INTEGER;
CREATE INDEX idx_results_campaign_target_id ON results(campaign_target_id);
