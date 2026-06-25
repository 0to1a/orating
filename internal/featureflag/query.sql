-- name: FFIsAdmin :one
-- Admin role check untuk admin-only endpoints. TRUE kalau user adalah
-- admin di company. Ditulis ulang di sini (bukan share dengan
-- company.requireAdmin) karena domain TIDAK import domain lain.
SELECT EXISTS(
    SELECT 1 FROM company_members
    WHERE company_id = $1 AND user_id = $2
      AND role = 'admin'
      AND deleted_at IS NULL
) AS is_admin;

-- name: FFListAll :many
-- All flags semua company — dipakai cron refresh untuk rebuild snapshot cache.
SELECT id, company_id, flag_key, enabled, updated_at
FROM feature_flags;

-- name: FFAdminListAll :many
-- Admin view — semua flag dari semua company, JOIN companies untuk
-- display name. Order: company_name, flag_key supaya stable.
SELECT
    ff.id,
    ff.company_id,
    c.name AS company_name,
    ff.flag_key,
    ff.enabled,
    ff.updated_at
FROM feature_flags ff
JOIN companies c ON c.id = ff.company_id AND c.deleted_at IS NULL
ORDER BY c.name, ff.flag_key;

-- name: FFListCompaniesAdmin :many
-- Daftar company untuk multi-select target di create dialog.
SELECT id, name
FROM companies
WHERE deleted_at IS NULL
ORDER BY id;

-- name: FFListDistinctKeys :many
-- Daftar flag_key unik (across all companies) untuk autocomplete.
SELECT DISTINCT flag_key
FROM feature_flags
ORDER BY flag_key;

-- name: FFUpsert :one
-- Toggle / create flag. Pakai ON CONFLICT (company_id, flag_key) DO UPDATE
-- supaya idempotent.
INSERT INTO feature_flags (company_id, flag_key, enabled)
VALUES ($1, $2, $3)
ON CONFLICT (company_id, flag_key)
DO UPDATE SET enabled = EXCLUDED.enabled, updated_at = NOW()
RETURNING id, company_id, flag_key, enabled, updated_at;
