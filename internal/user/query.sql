-- name: UserGetByID :one
SELECT *
FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: UserUpdateName :exec
UPDATE users
SET name = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: UserSetSelectedCompany :exec
UPDATE users
SET selected_company_id = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;
