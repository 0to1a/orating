-- name: APIKeyCreate :one
INSERT INTO api_keys (hash, prefix, name, company_id, created_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: APIKeyListByCompany :many
SELECT *
FROM api_keys
WHERE company_id = $1 AND deleted_at IS NULL AND revoked_at IS NULL
ORDER BY id DESC;

-- name: APIKeyGetByID :one
SELECT *
FROM api_keys
WHERE id = $1 AND deleted_at IS NULL;

-- name: APIKeyRevoke :exec
UPDATE api_keys
SET revoked_at = NOW()
WHERE id = $1 AND deleted_at IS NULL AND revoked_at IS NULL;

-- name: APIKeyListAllActive :many
SELECT *
FROM api_keys
WHERE deleted_at IS NULL AND revoked_at IS NULL;

-- name: APIKeyTouchLastUsed :exec
UPDATE api_keys
SET last_used_at = NOW()
WHERE id = $1;

-- name: APIKeyResolve :one
-- Resolve api key + caller user dalam satu query.
-- Match user state SAAT INI (selected_company_id, name, email) — bukan
-- snapshot saat key dibuat. Akses semantically identical to user session
-- token; bedanya cuma persistence.
SELECT
    ak.id           AS api_key_id,
    u.id            AS user_id,
    u.email         AS user_email,
    u.name          AS user_name,
    u.selected_company_id
FROM api_keys ak
JOIN users u ON u.id = ak.created_by AND u.deleted_at IS NULL
WHERE ak.hash = $1
  AND ak.deleted_at IS NULL
  AND ak.revoked_at IS NULL;
