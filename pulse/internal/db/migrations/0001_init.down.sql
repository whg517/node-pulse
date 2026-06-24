-- Revert the initial schema baseline (0001_init).
-- Drops all objects created by 0001_init.up.sql. CASCADE handles FK order.

-- Trigger & function
DROP TRIGGER IF EXISTS update_probes_updated_at ON probes;
DROP FUNCTION IF EXISTS update_probes_updated_at_func();

-- Tables (CASCADE removes dependent objects automatically)
DROP TABLE IF EXISTS webhook_logs CASCADE;
DROP TABLE IF EXISTS alert_notes CASCADE;
DROP TABLE IF EXISTS alert_status_history CASCADE;
DROP TABLE IF EXISTS alert_records CASCADE;
DROP TABLE IF EXISTS alert_suppressions CASCADE;
DROP TABLE IF EXISTS alert_events CASCADE;
DROP TABLE IF EXISTS webhooks CASCADE;
DROP TABLE IF EXISTS alerts CASCADE;
DROP TABLE IF EXISTS metrics CASCADE;
DROP TABLE IF EXISTS mtr_results CASCADE;
DROP TABLE IF EXISTS beacon_config_history CASCADE;
DROP TABLE IF EXISTS beacon_configs CASCADE;
DROP TABLE IF EXISTS probes CASCADE;
DROP TABLE IF EXISTS nodes CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS password_reset_tokens CASCADE;
DROP TABLE IF EXISTS rate_limits CASCADE;
DROP TABLE IF EXISTS auth_audit_logs CASCADE;
DROP TABLE IF EXISTS token_blacklist CASCADE;
DROP TABLE IF EXISTS api_keys CASCADE;
DROP TABLE IF EXISTS refresh_tokens CASCADE;
DROP TABLE IF EXISTS service_accounts CASCADE;
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS users CASCADE;
