-- =========================
-- DROP CAMPAIGN TARGETS
-- =========================
DROP TABLE IF EXISTS campaign_targets;

-- =========================
-- REMOVER ÍNDICES ANTES DAS COLUNAS
-- =========================
DROP INDEX IF EXISTS idx_campaigns_smtp_profile_id;
DROP INDEX IF EXISTS idx_campaigns_email_template_id;

-- =========================
-- AGORA SIM remover colunas
-- =========================
ALTER TABLE campaigns DROP COLUMN send_emails;
ALTER TABLE campaigns DROP COLUMN smtp_profile_id;
ALTER TABLE campaigns DROP COLUMN email_template_id;

-- =========================
-- RESTANTE
-- =========================
DROP TABLE IF EXISTS group_targets;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS targets;
DROP TABLE IF EXISTS email_templates;
DROP TABLE IF EXISTS smtp_profiles;