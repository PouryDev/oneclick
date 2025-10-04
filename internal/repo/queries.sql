-- name: InsertUser :one
INSERT INTO
    users (name, email, password_hash)
VALUES ($1, $2, $3) RETURNING id,
    name,
    email,
    created_at,
    updated_at;

-- name: GetUserByEmail :one
SELECT
    id,
    name,
    email,
    password_hash,
    created_at,
    updated_at
FROM users
WHERE
    email = $1;

-- name: GetUserByID :one
SELECT
    id,
    name,
    email,
    password_hash,
    created_at,
    updated_at
FROM users
WHERE
    id = $1;

-- name: UpdateUserPassword :exec
UPDATE users
SET
    password_hash = $2,
    updated_at = NOW()
WHERE
    id = $1;

-- Organization queries
-- name: CreateOrganization :one
INSERT INTO
    organizations (name)
VALUES ($1) RETURNING id,
    name,
    created_at,
    updated_at;

-- name: GetOrganizationByID :one
SELECT id, name, created_at, updated_at
FROM organizations
WHERE
    id = $1;

-- name: UpdateOrganization :one
UPDATE organizations
SET
    name = $2,
    updated_at = NOW()
WHERE
    id = $1 RETURNING id,
    name,
    created_at,
    updated_at;

-- name: DeleteOrganization :exec
DELETE FROM organizations WHERE id = $1;

-- User-Organization queries
-- name: AddUserToOrganization :one
INSERT INTO
    user_organizations (user_id, org_id, role)
VALUES ($1, $2, $3) RETURNING user_id,
    org_id,
    role,
    created_at,
    updated_at;

-- name: GetUserOrganizations :many
SELECT o.id, o.name, o.created_at, o.updated_at, uo.role
FROM
    organizations o
    JOIN user_organizations uo ON o.id = uo.org_id
WHERE
    uo.user_id = $1
ORDER BY o.created_at DESC;

-- name: GetOrganizationMembers :many
SELECT u.id, u.name, u.email, uo.role, uo.created_at, uo.updated_at
FROM
    users u
    JOIN user_organizations uo ON u.id = uo.user_id
WHERE
    uo.org_id = $1
ORDER BY uo.created_at ASC;

-- name: GetUserRoleInOrganization :one
SELECT role
FROM user_organizations
WHERE
    user_id = $1
    AND org_id = $2;

-- name: UpdateUserRoleInOrganization :one
UPDATE user_organizations
SET role = $3,
updated_at = NOW()
WHERE
    user_id = $1
    AND org_id = $2 RETURNING user_id,
    org_id,
    role,
    created_at,
    updated_at;

-- name: RemoveUserFromOrganization :exec
DELETE FROM user_organizations WHERE user_id = $1 AND org_id = $2;

-- name: GetUserByEmailForOrg :one
SELECT
    id,
    name,
    email,
    created_at,
    updated_at
FROM users
WHERE
    email = $1;