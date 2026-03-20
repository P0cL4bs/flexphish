DROP INDEX IF EXISTS idx_results_campaign_target_id;
ALTER TABLE results DROP COLUMN campaign_target_id;

DROP INDEX IF EXISTS idx_campaign_targets_result_id;
ALTER TABLE campaign_targets DROP COLUMN result_id;
