-- ============================================================
-- users
-- ============================================================
CREATE TABLE IF NOT EXISTS users (
    id                    BIGSERIAL PRIMARY KEY,
    email                 TEXT NOT NULL UNIQUE,
    name                  TEXT NOT NULL,
    otp_code              TEXT,
    otp_expires_at        TIMESTAMP,
    selected_company_id   BIGINT,
    created_at            TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at            TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_active ON users (deleted_at) WHERE deleted_at IS NULL;

-- ============================================================
-- companies
-- ============================================================
CREATE TABLE IF NOT EXISTS companies (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    owner_id    BIGINT NOT NULL REFERENCES users(id),
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_companies_active ON companies (deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_companies_owner  ON companies (owner_id) WHERE deleted_at IS NULL;

-- ============================================================
-- users.selected_company_id FK (after companies exists)
-- ============================================================
ALTER TABLE users
    ADD CONSTRAINT fk_users_selected_company
    FOREIGN KEY (selected_company_id) REFERENCES companies(id);

CREATE INDEX IF NOT EXISTS idx_users_selected_company
    ON users (selected_company_id) WHERE selected_company_id IS NOT NULL;

-- ============================================================
-- company_members
-- ============================================================
CREATE TABLE IF NOT EXISTS company_members (
    id          BIGSERIAL PRIMARY KEY,
    company_id  BIGINT NOT NULL REFERENCES companies(id),
    user_id     BIGINT NOT NULL REFERENCES users(id),
    role        TEXT NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_company_members_active
    ON company_members (company_id, user_id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_company_members_user
    ON company_members (user_id) WHERE deleted_at IS NULL;

-- ============================================================
-- api_keys
-- ============================================================
CREATE TABLE IF NOT EXISTS api_keys (
    id            BIGSERIAL PRIMARY KEY,
    hash          TEXT NOT NULL UNIQUE,
    prefix        TEXT NOT NULL,
    name          TEXT NOT NULL,
    company_id    BIGINT NOT NULL REFERENCES companies(id),
    created_by    BIGINT NOT NULL REFERENCES users(id),
    last_used_at  TIMESTAMP,
    revoked_at    TIMESTAMP,
    created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_api_keys_company
    ON api_keys (company_id) WHERE deleted_at IS NULL AND revoked_at IS NULL;

-- ============================================================
-- feature_flags
-- ============================================================
CREATE TABLE IF NOT EXISTS feature_flags (
    id          BIGSERIAL PRIMARY KEY,
    company_id  BIGINT NOT NULL REFERENCES companies(id),
    flag_key    TEXT NOT NULL,
    enabled     BOOLEAN NOT NULL DEFAULT FALSE,
    payload     JSONB,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_feature_flags_company_key
    ON feature_flags (company_id, flag_key);

-- ============================================================
-- user_sessions
-- ============================================================
CREATE TABLE IF NOT EXISTS user_sessions (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token        TEXT NOT NULL UNIQUE,
    expires_at   TIMESTAMP NOT NULL,
    last_seen_at TIMESTAMP,
    created_at   TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_sessions_user ON user_sessions (user_id);

-- ============================================================
-- Seed data
-- ============================================================
INSERT INTO users (id, email, name)
VALUES (1, 'admin@localhost', 'Admin')
    ON CONFLICT (email) DO NOTHING;

INSERT INTO companies (id, name, owner_id)
VALUES (1, 'Default Company', 1)
    ON CONFLICT (id) DO NOTHING;

INSERT INTO company_members (company_id, user_id, role)
VALUES (1, 1, 'admin')
    ON CONFLICT DO NOTHING;

INSERT INTO feature_flags (company_id, flag_key, enabled)
VALUES (1, 'labPage', TRUE)
    ON CONFLICT (company_id, flag_key) DO NOTHING;

-- Reset sequences after explicit-id seed rows so BIGSERIAL inserts start at 2.
SELECT setval('users_id_seq',         (SELECT MAX(id) FROM users));
SELECT setval('companies_id_seq',     (SELECT MAX(id) FROM companies));
SELECT setval('feature_flags_id_seq', (SELECT MAX(id) FROM feature_flags));
