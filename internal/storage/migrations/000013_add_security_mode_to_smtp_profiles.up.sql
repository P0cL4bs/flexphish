ALTER TABLE smtp_profiles ADD COLUMN security_mode TEXT NOT NULL DEFAULT 'starttls';

UPDATE smtp_profiles
SET security_mode = CASE
    WHEN port = 465 THEN 'implicit_tls'
    ELSE 'starttls'
END
WHERE security_mode IS NULL OR TRIM(security_mode) = '';
