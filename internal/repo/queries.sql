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

-- Cluster queries
-- name: CreateCluster :one
INSERT INTO
    clusters (
        org_id,
        name,
        provider,
        region,
        kubeconfig_encrypted,
        status
    )
VALUES ($1, $2, $3, $4, $5, $6) RETURNING id,
    org_id,
    name,
    provider,
    region,
    node_count,
    status,
    kube_version,
    last_health_check,
    created_at,
    updated_at;

-- name: GetClusterByID :one
SELECT
    id,
    org_id,
    name,
    provider,
    region,
    kubeconfig_encrypted,
    node_count,
    status,
    kube_version,
    last_health_check,
    created_at,
    updated_at
FROM clusters
WHERE
    id = $1;

-- name: GetClustersByOrgID :many
SELECT
    id,
    org_id,
    name,
    provider,
    region,
    node_count,
    status,
    kube_version,
    last_health_check,
    created_at,
    updated_at
FROM clusters
WHERE
    org_id = $1
ORDER BY created_at DESC;

-- name: UpdateClusterStatus :one
UPDATE clusters
SET
    status = $2,
    updated_at = NOW()
WHERE
    id = $1 RETURNING id,
    org_id,
    name,
    provider,
    region,
    node_count,
    status,
    kube_version,
    last_health_check,
    created_at,
    updated_at;

-- name: UpdateClusterKubeconfig :one
UPDATE clusters
SET
    kubeconfig_encrypted = $2,
    status = $3,
    updated_at = NOW()
WHERE
    id = $1 RETURNING id,
    org_id,
    name,
    provider,
    region,
    node_count,
    status,
    kube_version,
    last_health_check,
    created_at,
    updated_at;

-- name: UpdateClusterHealth :one
UPDATE clusters
SET
    kube_version = $2,
    last_health_check = NOW(),
    updated_at = NOW()
WHERE
    id = $1 RETURNING id,
    org_id,
    name,
    provider,
    region,
    node_count,
    status,
    kube_version,
    last_health_check,
    created_at,
    updated_at;

-- name: UpdateClusterNodeCount :one
UPDATE clusters
SET
    node_count = $2,
    updated_at = NOW()
WHERE
    id = $1 RETURNING id,
    org_id,
    name,
    provider,
    region,
    node_count,
    status,
    kube_version,
    last_health_check,
    created_at,
    updated_at;

-- name: DeleteCluster :exec
DELETE FROM clusters WHERE id = $1;

-- Repository queries
-- name: CreateRepository :one
INSERT INTO
    repositories (
        org_id,
        type,
        url,
        default_branch,
        config
    )
VALUES ($1, $2, $3, $4, $5) RETURNING id,
    org_id,
    type,
    url,
    default_branch,
    config,
    created_at,
    updated_at;

-- name: GetRepositoryByID :one
SELECT
    id,
    org_id,
    type,
    url,
    default_branch,
    config,
    created_at,
    updated_at
FROM repositories
WHERE
    id = $1;

-- name: GetRepositoriesByOrgID :many
SELECT
    id,
    org_id,
    type,
    url,
    default_branch,
    config,
    created_at,
    updated_at
FROM repositories
WHERE
    org_id = $1
ORDER BY created_at DESC;

-- name: GetRepositoryByURL :one
SELECT
    id,
    org_id,
    type,
    url,
    default_branch,
    config,
    created_at,
    updated_at
FROM repositories
WHERE
    org_id = $1
    AND url = $2;

-- name: UpdateRepositoryConfig :one
UPDATE repositories
SET
    config = $2,
    updated_at = NOW()
WHERE
    id = $1 RETURNING id,
    org_id,
    type,
    url,
    default_branch,
    config,
    created_at,
    updated_at;

-- name: DeleteRepository :exec
DELETE FROM repositories WHERE id = $1;

-- Application queries
-- name: CreateApplication :one
INSERT INTO
    applications (
        org_id,
        cluster_id,
        name,
        repo_id,
        path,
        default_branch
    )
VALUES ($1, $2, $3, $4, $5, $6) RETURNING id,
    org_id,
    cluster_id,
    name,
    repo_id,
    path,
    default_branch,
    created_at,
    updated_at;

-- name: GetApplicationByID :one
SELECT
    id,
    org_id,
    cluster_id,
    name,
    repo_id,
    path,
    default_branch,
    created_at,
    updated_at
FROM applications
WHERE
    id = $1;

-- name: GetApplicationsByClusterID :many
SELECT
    id,
    org_id,
    cluster_id,
    name,
    repo_id,
    path,
    default_branch,
    created_at,
    updated_at
FROM applications
WHERE
    cluster_id = $1
ORDER BY created_at DESC;

-- name: GetApplicationByNameInCluster :one
SELECT
    id,
    org_id,
    cluster_id,
    name,
    repo_id,
    path,
    default_branch,
    created_at,
    updated_at
FROM applications
WHERE
    cluster_id = $1
    AND name = $2;

-- name: DeleteApplication :exec
DELETE FROM applications WHERE id = $1;

-- Release queries
-- name: CreateRelease :one
INSERT INTO
    releases (
        app_id,
        image,
        tag,
        created_by,
        status,
        meta
    )
VALUES ($1, $2, $3, $4, $5, $6) RETURNING id,
    app_id,
    image,
    tag,
    created_by,
    status,
    started_at,
    finished_at,
    meta,
    created_at,
    updated_at;

-- name: GetReleaseByID :one
SELECT
    id,
    app_id,
    image,
    tag,
    created_by,
    status,
    started_at,
    finished_at,
    meta,
    created_at,
    updated_at
FROM releases
WHERE
    id = $1;

-- name: GetReleasesByAppID :many
SELECT
    id,
    app_id,
    image,
    tag,
    created_by,
    status,
    started_at,
    finished_at,
    meta,
    created_at,
    updated_at
FROM releases
WHERE
    app_id = $1
ORDER BY created_at DESC;

-- name: GetLatestReleaseByAppID :one
SELECT
    id,
    app_id,
    image,
    tag,
    created_by,
    status,
    started_at,
    finished_at,
    meta,
    created_at,
    updated_at
FROM releases
WHERE
    app_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: UpdateReleaseStatus :one
UPDATE releases
SET
    status = $2,
    started_at = $3,
    finished_at = $4,
    updated_at = NOW()
WHERE
    id = $1 RETURNING id,
    app_id,
    image,
    tag,
    created_by,
    status,
    started_at,
    finished_at,
    meta,
    created_at,
    updated_at;

-- name: UpdateReleaseMeta :one
UPDATE releases
SET
    meta = $2,
    updated_at = NOW()
WHERE
    id = $1 RETURNING id,
    app_id,
    image,
    tag,
    created_by,
    status,
    started_at,
    finished_at,
    meta,
    created_at,
    updated_at;

-- name: DeleteRelease :exec
DELETE FROM releases WHERE id = $1;

-- Service queries
-- name: CreateService :one
INSERT INTO
    services (
        app_id,
        name,
        chart,
        status,
        namespace
    )
VALUES ($1, $2, $3, $4, $5) RETURNING id,
    app_id,
    name,
    chart,
    status,
    namespace,
    created_at,
    updated_at;

-- name: GetServiceByID :one
SELECT
    id,
    app_id,
    name,
    chart,
    status,
    namespace,
    created_at,
    updated_at
FROM services
WHERE
    id = $1;

-- name: GetServicesByAppID :many
SELECT
    id,
    app_id,
    name,
    chart,
    status,
    namespace,
    created_at,
    updated_at
FROM services
WHERE
    app_id = $1
ORDER BY created_at DESC;

-- name: GetServiceByNameInApp :one
SELECT
    id,
    app_id,
    name,
    chart,
    status,
    namespace,
    created_at,
    updated_at
FROM services
WHERE
    app_id = $1
    AND name = $2;

-- name: UpdateServiceStatus :one
UPDATE services
SET
    status = $2,
    updated_at = NOW()
WHERE
    id = $1 RETURNING id,
    app_id,
    name,
    chart,
    status,
    namespace,
    created_at,
    updated_at;

-- name: DeleteService :exec
DELETE FROM services WHERE id = $1;

-- Service Config queries
-- name: CreateServiceConfig :one
INSERT INTO
    service_configs (
        service_id,
        key,
        value,
        is_secret
    )
VALUES ($1, $2, $3, $4) RETURNING id,
    service_id,
    key,
    value,
    is_secret,
    created_at,
    updated_at;

-- name: GetServiceConfigByID :one
SELECT
    id,
    service_id,
    key,
    value,
    is_secret,
    created_at,
    updated_at
FROM service_configs
WHERE
    id = $1;

-- name: GetServiceConfigsByServiceID :many
SELECT
    id,
    service_id,
    key,
    value,
    is_secret,
    created_at,
    updated_at
FROM service_configs
WHERE
    service_id = $1
ORDER BY key;

-- name: GetServiceConfigByKey :one
SELECT
    id,
    service_id,
    key,
    value,
    is_secret,
    created_at,
    updated_at
FROM service_configs
WHERE
    service_id = $1
    AND key = $2;

-- name: UpdateServiceConfigValue :one
UPDATE service_configs
SET
    value = $2,
    updated_at = NOW()
WHERE
    id = $1 RETURNING id,
    service_id,
    key,
    value,
    is_secret,
    created_at,
    updated_at;

-- Git Server queries
-- name: CreateGitServer :one
INSERT INTO
    git_servers (
        org_id,
        type,
        domain,
        storage,
        status,
        config
    )
VALUES ($1, $2, $3, $4, $5, $6) RETURNING id,
    org_id,
    type,
    domain,
    storage,
    status,
    config,
    created_at,
    updated_at;

-- name: GetGitServerByID :one
SELECT
    id,
    org_id,
    type,
    domain,
    storage,
    status,
    config,
    created_at,
    updated_at
FROM git_servers
WHERE
    id = $1;

-- name: GetGitServersByOrgID :many
SELECT
    id,
    org_id,
    type,
    domain,
    storage,
    status,
    config,
    created_at,
    updated_at
FROM git_servers
WHERE
    org_id = $1
ORDER BY created_at DESC;

-- name: GetGitServerByDomainInOrg :one
SELECT
    id,
    org_id,
    type,
    domain,
    storage,
    status,
    config,
    created_at,
    updated_at
FROM git_servers
WHERE
    org_id = $1
    AND domain = $2;

-- name: UpdateGitServerStatus :one
UPDATE git_servers
SET
    status = $2,
    updated_at = NOW()
WHERE
    id = $1 RETURNING id,
    org_id,
    type,
    domain,
    storage,
    status,
    config,
    created_at,
    updated_at;

-- name: UpdateGitServerConfig :one
UPDATE git_servers
SET
    config = $2,
    updated_at = NOW()
WHERE
    id = $1 RETURNING id,
    org_id,
    type,
    domain,
    storage,
    status,
    config,
    created_at,
    updated_at;

-- name: DeleteGitServer :exec
DELETE FROM git_servers WHERE id = $1;

-- Runner queries
-- name: CreateRunner :one
INSERT INTO
    runners (
        org_id,
        name,
        type,
        config,
        status
    )
VALUES ($1, $2, $3, $4, $5) RETURNING id,
    org_id,
    name,
    type,
    config,
    status,
    created_at,
    updated_at;

-- name: GetRunnerByID :one
SELECT
    id,
    org_id,
    name,
    type,
    config,
    status,
    created_at,
    updated_at
FROM runners
WHERE
    id = $1;

-- name: GetRunnersByOrgID :many
SELECT
    id,
    org_id,
    name,
    type,
    config,
    status,
    created_at,
    updated_at
FROM runners
WHERE
    org_id = $1
ORDER BY created_at DESC;

-- name: GetRunnerByNameInOrg :one
SELECT
    id,
    org_id,
    name,
    type,
    config,
    status,
    created_at,
    updated_at
FROM runners
WHERE
    org_id = $1
    AND name = $2;

-- name: UpdateRunnerStatus :one
UPDATE runners
SET
    status = $2,
    updated_at = NOW()
WHERE
    id = $1 RETURNING id,
    org_id,
    name,
    type,
    config,
    status,
    created_at,
    updated_at;

-- name: UpdateRunnerConfig :one
UPDATE runners
SET
    config = $2,
    updated_at = NOW()
WHERE
    id = $1 RETURNING id,
    org_id,
    name,
    type,
    config,
    status,
    created_at,
    updated_at;

-- name: DeleteRunner :exec
DELETE FROM runners WHERE id = $1;

-- Job Queue queries
-- name: CreateJob :one
INSERT INTO
    job_queue (org_id, type, status, payload)
VALUES ($1, $2, $3, $4) RETURNING id,
    org_id,
    type,
    status,
    payload,
    error_message,
    created_at,
    started_at,
    completed_at;

-- name: GetJobByID :one
SELECT
    id,
    org_id,
    type,
    status,
    payload,
    error_message,
    created_at,
    started_at,
    completed_at
FROM job_queue
WHERE
    id = $1;

-- name: GetJobsByOrgID :many
SELECT
    id,
    org_id,
    type,
    status,
    payload,
    error_message,
    created_at,
    started_at,
    completed_at
FROM job_queue
WHERE
    org_id = $1
ORDER BY created_at DESC;

-- name: GetPendingJobs :many
SELECT
    id,
    org_id,
    type,
    status,
    payload,
    error_message,
    created_at,
    started_at,
    completed_at
FROM job_queue
WHERE
    status = 'pending'
ORDER BY created_at ASC;

-- name: UpdateJobStatus :one
UPDATE job_queue
SET
    status = $2,
    updated_at = NOW()
WHERE
    id = $1 RETURNING id,
    org_id,
    type,
    status,
    payload,
    error_message,
    created_at,
    started_at,
    completed_at;

-- name: StartJob :one
UPDATE job_queue
SET
    status = 'processing',
    started_at = NOW(),
    updated_at = NOW()
WHERE
    id = $1
    AND status = 'pending' RETURNING id,
    org_id,
    type,
    status,
    payload,
    error_message,
    created_at,
    started_at,
    completed_at;

-- name: CompleteJob :one
UPDATE job_queue
SET
    status = 'completed',
    completed_at = NOW(),
    updated_at = NOW()
WHERE
    id = $1 RETURNING id,
    org_id,
    type,
    status,
    payload,
    error_message,
    created_at,
    started_at,
    completed_at;

-- name: FailJob :one
UPDATE job_queue
SET
    status = 'failed',
    error_message = $2,
    completed_at = NOW(),
    updated_at = NOW()
WHERE
    id = $1 RETURNING id,
    org_id,
    type,
    status,
    payload,
    error_message,
    created_at,
    started_at,
    completed_at;

-- Domain queries
-- name: CreateDomain :one
INSERT INTO
    domains (
        app_id,
        domain,
        provider,
        provider_config,
        cert_status,
        cert_secret_name,
        challenge_type
    )
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id,
    app_id,
    domain,
    provider,
    provider_config,
    cert_status,
    cert_secret_name,
    challenge_type,
    created_at,
    updated_at;

-- name: GetDomainByID :one
SELECT
    id,
    app_id,
    domain,
    provider,
    provider_config,
    cert_status,
    cert_secret_name,
    challenge_type,
    created_at,
    updated_at
FROM domains
WHERE
    id = $1;

-- name: GetDomainsByAppID :many
SELECT
    id,
    app_id,
    domain,
    provider,
    provider_config,
    cert_status,
    cert_secret_name,
    challenge_type,
    created_at,
    updated_at
FROM domains
WHERE
    app_id = $1
ORDER BY created_at DESC;

-- name: GetDomainByDomainInApp :one
SELECT
    id,
    app_id,
    domain,
    provider,
    provider_config,
    cert_status,
    cert_secret_name,
    challenge_type,
    created_at,
    updated_at
FROM domains
WHERE
    app_id = $1
    AND domain = $2;

-- name: UpdateDomainCertStatus :one
UPDATE domains
SET
    cert_status = $2,
    updated_at = NOW()
WHERE
    id = $1 RETURNING id,
    app_id,
    domain,
    provider,
    provider_config,
    cert_status,
    cert_secret_name,
    challenge_type,
    created_at,
    updated_at;

-- name: UpdateDomainCertSecret :one
UPDATE domains
SET
    cert_secret_name = $2,
    updated_at = NOW()
WHERE
    id = $1 RETURNING id,
    app_id,
    domain,
    provider,
    provider_config,
    cert_status,
    cert_secret_name,
    challenge_type,
    created_at,
    updated_at;

-- name: UpdateDomainProviderConfig :one
UPDATE domains
SET
    provider_config = $2,
    updated_at = NOW()
WHERE
    id = $1 RETURNING id,
    app_id,
    domain,
    provider,
    provider_config,
    cert_status,
    cert_secret_name,
    challenge_type,
    created_at,
    updated_at;

-- name: DeleteDomain :exec
DELETE FROM domains WHERE id = $1;