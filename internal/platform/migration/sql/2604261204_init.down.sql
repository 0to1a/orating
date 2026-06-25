-- Drop in reverse order of dependencies.
DROP TABLE IF EXISTS user_sessions;

DROP INDEX IF EXISTS uq_feature_flags_company_key;
DROP TABLE IF EXISTS feature_flags;

DROP INDEX IF EXISTS idx_api_keys_company;
DROP TABLE IF EXISTS api_keys;

DROP INDEX IF EXISTS idx_company_members_user;
DROP INDEX IF EXISTS uq_company_members_active;
DROP TABLE IF EXISTS company_members;

DROP INDEX IF EXISTS idx_users_selected_company;
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_selected_company;

DROP INDEX IF EXISTS idx_companies_owner;
DROP INDEX IF EXISTS idx_companies_active;
DROP TABLE IF EXISTS companies;

DROP INDEX IF EXISTS idx_users_active;
DROP TABLE IF EXISTS users;
