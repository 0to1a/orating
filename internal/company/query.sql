-- name: CompanyCreate :one
INSERT INTO companies (name, owner_id)
VALUES ($1, $2)
RETURNING *;

-- name: CompanyAddMember :one
INSERT INTO company_members (company_id, user_id, role)
VALUES ($1, $2, $3)
RETURNING *;

-- name: CompanyGetMembership :one
SELECT *
FROM company_members
WHERE company_id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: CompanyListByUser :many
SELECT
    c.id,
    c.name,
    c.owner_id,
    c.created_at,
    c.updated_at,
    c.deleted_at,
    cm.role
FROM companies c
JOIN company_members cm ON cm.company_id = c.id AND cm.deleted_at IS NULL
WHERE cm.user_id = $1
  AND c.deleted_at IS NULL
ORDER BY c.id;

-- name: CompanyGetByID :one
SELECT *
FROM companies
WHERE id = $1 AND deleted_at IS NULL;

-- name: CompanyListMembers :many
SELECT
    cm.user_id,
    cm.role,
    cm.created_at AS joined_at,
    u.email,
    u.name
FROM company_members cm
JOIN users u ON u.id = cm.user_id AND u.deleted_at IS NULL
WHERE cm.company_id = $1 AND cm.deleted_at IS NULL
ORDER BY cm.id;

-- name: CompanyFindUserByEmail :one
SELECT *
FROM users
WHERE email = $1 AND deleted_at IS NULL;

-- name: CompanyCreateUser :one
INSERT INTO users (email, name)
VALUES ($1, $2)
RETURNING *;

-- name: CompanyRemoveMember :exec
UPDATE company_members
SET deleted_at = NOW()
WHERE company_id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: CompanyUpdateMemberRole :exec
UPDATE company_members
SET role = $3
WHERE company_id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: CompanySetUserSelectedCompany :exec
UPDATE users
SET selected_company_id = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: CompanyExistsOwnedBy :one
-- True kalau user pernah jadi owner suatu company. Dipakai untuk gate
-- "siapa yang boleh create company baru" — non-owner harus di-invite.
SELECT EXISTS(
    SELECT 1 FROM companies
    WHERE owner_id = $1 AND deleted_at IS NULL
) AS owns;
