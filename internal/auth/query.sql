-- name: AuthFindUserByEmail :one
SELECT *
FROM users
WHERE email = $1 AND deleted_at IS NULL;

-- name: AuthCreateUser :one
INSERT INTO users (email, name)
VALUES ($1, $2)
RETURNING *;

-- name: AuthSetOTP :exec
UPDATE users
SET otp_code = $2, otp_expires_at = $3, updated_at = NOW()
WHERE id = $1;

-- name: AuthClearOTP :exec
UPDATE users
SET otp_code = NULL, otp_expires_at = NULL, updated_at = NOW()
WHERE id = $1;

-- name: AuthCreateSession :one
INSERT INTO user_sessions (user_id, token, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: AuthDeleteSession :exec
DELETE FROM user_sessions WHERE token = $1;

-- name: AuthDeleteAllUserSessions :exec
DELETE FROM user_sessions WHERE user_id = $1;

-- name: AuthFindUserBySessionToken :one
SELECT u.*
FROM users u
JOIN user_sessions s ON s.user_id = u.id
WHERE s.token = $1
  AND s.expires_at > NOW()
  AND u.deleted_at IS NULL;

-- name: AuthFirstCompanyForUser :one
-- Untuk auto-select company pertama saat user login tanpa selected_company.
-- Order by id biar deterministic.
SELECT cm.company_id
FROM company_members cm
JOIN companies c ON c.id = cm.company_id AND c.deleted_at IS NULL
WHERE cm.user_id = $1 AND cm.deleted_at IS NULL
ORDER BY cm.id
LIMIT 1;

-- name: AuthSetSelectedCompany :exec
UPDATE users
SET selected_company_id = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;
