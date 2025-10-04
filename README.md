# OneClick Backend

A Go backend service built with Clean Architecture principles, featuring authentication, user management, organization management, Kubernetes cluster management, repository integration, webhook processing, application deployment with release management, infrastructure service provisioning, self-hosted Git server and CI runner management, custom domain management with SSL certificate automation, real-time pod runtime management with terminal access, and comprehensive monitoring with Prometheus integration.

## Tech Stack

- **Language**: Go 1.20+
- **Framework**: Gin (github.com/gin-gonic/gin)
- **Database**: PostgreSQL
- **Migrations**: golang-migrate
- **Configuration**: Viper + Environment Variables
- **Logging**: Zap (go.uber.org/zap)
- **Authentication**: JWT (github.com/golang-jwt/jwt/v5)
- **Testing**: Testify
- **Architecture**: Clean Architecture
- **Kubernetes**: client-go for cluster management, deployments, and pod operations
- **WebSocket**: gorilla/websocket for real-time terminal connections
- **Encryption**: AES-GCM for kubeconfig and token encryption
- **Webhooks**: HMAC signature verification for Git providers
- **Deployments**: Background worker for Kubernetes deployments
- **Infrastructure**: Helm-based service provisioning with YAML configuration
- **Git Servers**: Self-hosted Gitea instance management
- **CI Runners**: GitHub/GitLab/Custom runner deployment and management
- **Domain Management**: Custom domain configuration with cert-manager integration
- **SSL Certificates**: Automated SSL certificate provisioning with ACME challenges
- **Pod Management**: Real-time pod monitoring, logs streaming, and terminal access
- **Monitoring**: Prometheus integration with metrics aggregation, caching, and rate limiting

## Features

### 🔐 Authentication & User Management

- User registration and login with JWT tokens
- Password hashing with bcrypt
- User profile management
- Secure authentication middleware

### 🏢 Organization Management

- Create and manage organizations
- Role-based access control (Owner, Admin, Member)
- Add/remove organization members
- Update member roles
- Organization-specific resource access

### ☸️ Kubernetes Cluster Management

- Create new clusters with provider/region information
- Import existing clusters via kubeconfig upload
- Real-time cluster health monitoring
- Node information and resource usage
- Encrypted kubeconfig storage
- Cluster status tracking (provisioning/active/error/deleting)
- Kubernetes API integration using client-go

### 📁 Repository Management

- Connect GitHub, GitLab, and Gitea repositories
- Encrypted access token storage
- Repository configuration management
- Organization-based repository access
- Repository listing and details
- Repository deletion with permission checks

### 🔗 Webhook Integration

- Public webhook endpoints for all Git providers
- HMAC signature verification for security
- Support for GitHub, GitLab, and Gitea webhooks
- Automatic pipeline triggering on code pushes
- Webhook payload parsing and processing
- Provider-specific header handling

### 🚀 Applications & Releases

- Application management with repository integration
- Release tracking with version history
- Kubernetes deployment automation
- Background worker for deployment processing
- Rollback capability to previous releases
- Real-time deployment status monitoring
- Kubernetes manifest generation
- Environment and configuration management

### 🏗️ Infrastructure Service Provisioning

- YAML-based infrastructure configuration (`infra-config.yml`)
- Helm chart-based service provisioning
- Secret management with `SECRET::name` markers
- Template substitution for dynamic configuration
- Background service provisioning with status tracking
- Kubernetes secret management
- Service lifecycle management (provision/unprovision)
- Role-based access control for infrastructure operations

### 🐙 Git Server Management

- Self-hosted Gitea instance provisioning
- Domain and storage configuration
- Admin user and credential management
- Repository listing and management
- Background installation with Helm charts
- Status tracking and health monitoring
- Secure credential storage with encryption
- Organization-scoped git server management

### 🏃 CI Runner Management

- GitHub Actions runner deployment
- GitLab CI runner provisioning
- Custom runner configuration
- Label and node selector management
- Resource limits and scaling controls
- Background deployment with job queue
- Runner status monitoring
- Token encryption and secure storage

### 🌐 Domain Management

- Custom domain configuration for applications
- DNS provider integration (Cloudflare, Route53, Manual)
- SSL certificate management with cert-manager
- ACME challenge support (HTTP-01 and DNS-01)
- Encrypted DNS provider credentials storage
- Background domain provisioning with job queue
- Certificate status monitoring and renewal
- Manual DNS challenge instructions for manual providers
- Domain deletion with cleanup jobs
- Organization-scoped domain management

### 🚀 Runtime Management (Pods)

- Real-time pod monitoring and management
- Pod listing with status, restarts, and readiness information
- Detailed pod information including containers, events, and conditions
- Pod logs streaming with container-specific filtering
- Interactive terminal access via WebSocket connections
- kubectl describe-style pod information
- Container-level operations and monitoring
- Pod events and owner reference tracking
- Audit logging for all pod access operations
- Organization-scoped pod access control
- Kubernetes client-go integration for cluster operations
- WebSocket-based terminal with TTY resize support

### 📊 Monitoring & Metrics

- Prometheus integration for cluster, application, and pod metrics
- Real-time CPU and memory usage monitoring with time series data
- Cluster-level metrics including node health and resource utilization
- Application-level metrics with pod counts and status aggregation
- Pod-level metrics for individual container performance
- Alert management with severity levels and status tracking
- In-memory caching with configurable TTL for performance
- Rate limiting (100 requests per minute per user) to prevent abuse
- PromQL query execution with range and instant queries
- Comprehensive error handling and authorization
- Health check endpoints for monitoring service status
- Organization-scoped monitoring access control
- Background metrics collection and aggregation
- Support for multiple time ranges (5m, 15m, 1h, 6h, 24h)
- Top alerts filtering and limiting for application dashboards

### 🔒 Security Features

- AES-GCM encryption for sensitive data
- JWT-based authentication
- Role-based access control
- Input validation and sanitization
- Secure file upload handling

## Project Structure

```
oneclick/
├── cmd/
│   └── server/           # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/     # HTTP handlers
│   │   └── middleware/   # HTTP middleware
│   ├── app/
│   │   ├── crypto/       # Encryption utilities
│   │   ├── deployment/   # Kubernetes deployment generator
│   │   ├── infra/        # Infrastructure configuration parser
│   │   ├── provisioner/  # Helm-based service provisioning
│   │   ├── services/     # Business logic services
│   │   ├── webhook/     # Webhook signature verification
│   │   └── worker/      # Background workers
│   ├── config/           # Configuration management
│   ├── domain/           # Domain models and interfaces
│   └── repo/             # Data access layer
├── migrations/           # Database migrations
├── Makefile             # Build and development commands
├── Dockerfile           # Container configuration
└── .env.example         # Environment variables template
```

## Prerequisites

- Go 1.20 or higher
- PostgreSQL 12 or higher
- Make (optional, for using Makefile commands)

### Prerequisites

- **Go 1.20+**: Required for building and running the application
- **PostgreSQL**: Database for storing application data
- **Kubernetes Cluster**: For deploying applications and services
- **Helm CLI**: Required for infrastructure service provisioning
- **kubectl**: For Kubernetes cluster management

### Helm Installation

OneClick requires Helm CLI for infrastructure service provisioning. Install Helm:

```bash
# macOS
brew install helm

# Linux/Windows
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

Verify installation:

```bash
helm version
```

## Quick Start

### 1. Clone and Setup

```bash
git clone <repository-url>
cd oneclick
```

### 2. Install Dependencies

```bash
make deps
# or manually:
go mod download
go mod tidy
```

### 3. Database Setup

Create a PostgreSQL database:

```sql
CREATE DATABASE oneclick;
```

### 4. Environment Configuration

Copy the environment template and configure:

```bash
cp .env.example .env
```

Edit `.env` with your configuration:

```env
DATABASE_URL=postgres://username:password@localhost:5432/oneclick?sslmode=disable
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
ONECLICK_MASTER_KEY=your-32-byte-master-key-for-encryption-change-this
PORT=8080
LOG_LEVEL=info
GIN_MODE=debug
```

### 5. Run Database Migrations

Install migrate CLI and run migrations:

```bash
make install-migrate
make migrate-up
```

Or manually:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
migrate -path migrations -database "$DATABASE_URL" up
```

### 6. Start the Server

```bash
make run
# or manually:
go run ./cmd/server
```

The server will start on `http://localhost:8080`

## API Endpoints

### Authentication

#### Register User

```http
POST /auth/register
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "password123"
}
```

**Response (201):**

```json
{
  "id": "uuid",
  "name": "John Doe",
  "email": "john@example.com",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

#### Login

```http
POST /auth/login
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "password123"
}
```

**Response (200):**

```json
{
  "access_token": "jwt-token",
  "user": {
    "id": "uuid",
    "name": "John Doe",
    "email": "john@example.com",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

#### Get Current User

```http
GET /auth/me
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
{
  "id": "uuid",
  "name": "John Doe",
  "email": "john@example.com",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

#### Health Check

```http
GET /health
```

**Response (200):**

```json
{
  "status": "ok"
}
```

### Organizations

#### Create Organization

```http
POST /orgs
Authorization: Bearer <jwt-token>
Content-Type: application/json

{
  "name": "My Organization"
}
```

**Response (201):**

```json
{
  "id": "uuid",
  "name": "My Organization",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

#### Get User's Organizations

```http
GET /orgs
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
[
  {
    "id": "uuid",
    "name": "My Organization",
    "role": "owner",
    "clusters_count": 0,
    "apps_count": 0,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

#### Get Organization Details

```http
GET /orgs/{orgId}
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
{
  "id": "uuid",
  "name": "My Organization",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "members": [
    {
      "id": "uuid",
      "name": "John Doe",
      "email": "john@example.com",
      "role": "owner",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "clusters_count": 0,
  "apps_count": 0
}
```

#### Add Member to Organization

```http
POST /orgs/{orgId}/members
Authorization: Bearer <jwt-token>
Content-Type: application/json

{
  "email": "member@example.com",
  "role": "member"
}
```

**Response (201):**

```json
{
  "id": "uuid",
  "name": "Member Name",
  "email": "member@example.com",
  "role": "member",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

#### Update Member Role

```http
PATCH /orgs/{orgId}/members/{userId}
Authorization: Bearer <jwt-token>
Content-Type: application/json

{
  "role": "admin"
}
```

**Response (200):**

```json
{
  "id": "uuid",
  "name": "Member Name",
  "email": "member@example.com",
  "role": "admin",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

#### Remove Member from Organization

```http
DELETE /orgs/{orgId}/members/{userId}
Authorization: Bearer <jwt-token>
```

**Response (204):** No content

#### Delete Organization

```http
DELETE /orgs/{orgId}
Authorization: Bearer <jwt-token>
```

**Response (204):** No content

### Clusters

#### Create Cluster

```http
POST /orgs/{orgId}/clusters
Authorization: Bearer <jwt-token>
Content-Type: application/json

{
  "name": "My Cluster",
  "provider": "aws",
  "region": "us-west-2",
  "kubeconfig": "base64-encoded-kubeconfig" // optional
}
```

**Response (201):**

```json
{
  "id": "uuid",
  "name": "My Cluster",
  "provider": "aws",
  "region": "us-west-2",
  "node_count": 0,
  "status": "provisioning",
  "kube_version": null,
  "last_health_check": null,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

#### Import Cluster

```http
POST /orgs/{orgId}/clusters/import
Authorization: Bearer <jwt-token>
Content-Type: multipart/form-data

name: "Imported Cluster"
kubeconfig: <file>
```

**Response (201):**

```json
{
  "id": "uuid",
  "name": "Imported Cluster",
  "provider": "imported",
  "region": "unknown",
  "node_count": 0,
  "status": "active",
  "kube_version": null,
  "last_health_check": null,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

#### Get Organization Clusters

```http
GET /orgs/{orgId}/clusters
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
[
  {
    "id": "uuid",
    "name": "My Cluster",
    "provider": "aws",
    "region": "us-west-2",
    "node_count": 3,
    "status": "active",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

#### Get Cluster Details

```http
GET /clusters/{clusterId}
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
{
  "id": "uuid",
  "name": "My Cluster",
  "provider": "aws",
  "region": "us-west-2",
  "node_count": 3,
  "status": "active",
  "kube_version": "v1.25.0",
  "last_health_check": "2024-01-01T00:00:00Z",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "has_kubeconfig": true
}
```

#### Get Cluster Health

```http
GET /clusters/{clusterId}/status
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
{
  "status": "active",
  "kube_version": "v1.25.0",
  "nodes": [
    {
      "name": "node-1",
      "status": "Ready",
      "cpu": "2",
      "memory": "4Gi"
    },
    {
      "name": "node-2",
      "status": "Ready",
      "cpu": "2",
      "memory": "4Gi"
    }
  ],
  "last_check": "2024-01-01T00:00:00Z"
}
```

#### Delete Cluster

```http
DELETE /clusters/{clusterId}
Authorization: Bearer <jwt-token>
```

**Response (204):** No content

### Repositories

#### Create Repository

```http
POST /orgs/{orgId}/repos
Authorization: Bearer <jwt-token>
Content-Type: application/json

{
  "type": "github",
  "url": "https://github.com/user/repo.git",
  "default_branch": "main",
  "token": "ghp_xxxxxxxxxxxxxxxxxxxx" // optional
}
```

**Response (201):**

```json
{
  "id": "uuid",
  "type": "github",
  "url": "https://github.com/user/repo.git",
  "default_branch": "main",
  "config": "{}",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

#### Get Organization Repositories

```http
GET /orgs/{orgId}/repos
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
[
  {
    "id": "uuid",
    "type": "github",
    "url": "https://github.com/user/repo.git",
    "default_branch": "main",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

#### Get Repository Details

```http
GET /repos/{repoId}
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
{
  "id": "uuid",
  "type": "github",
  "url": "https://github.com/user/repo.git",
  "default_branch": "main",
  "config": "{}",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

#### Delete Repository

```http
DELETE /repos/{repoId}
Authorization: Bearer <jwt-token>
```

**Response (204):** No content

### Webhooks

#### Git Webhook Endpoint

```http
POST /hooks/git?provider=github&secret=your-webhook-secret
Content-Type: application/json
X-Hub-Signature-256: sha256=your-signature

{
  "ref": "refs/heads/main",
  "repository": {
    "full_name": "user/repo",
    "clone_url": "https://github.com/user/repo.git"
  },
  "commits": [
    {
      "id": "abc123",
      "message": "Add new feature",
      "author": {
        "name": "John Doe",
        "email": "john@example.com"
      }
    }
  ]
}
```

**Response (202):**

```json
{
  "message": "Webhook processed successfully"
}
```

#### Test Webhook Endpoint

```http
GET /hooks/test?provider=github
```

**Response (200):**

```json
{
  "message": "Webhook endpoint is working",
  "provider": "github",
  "status": "ready"
}
```

### Applications

#### Create Application

```http
POST /clusters/{clusterId}/apps
Authorization: Bearer <jwt-token>
Content-Type: application/json

{
  "name": "my-app",
  "repo_id": "uuid",
  "default_branch": "main",
  "path": "apps/my-app"
}
```

**Response (201):**

```json
{
  "id": "uuid",
  "org_id": "uuid",
  "cluster_id": "uuid",
  "name": "my-app",
  "repo_id": "uuid",
  "path": "apps/my-app",
  "default_branch": "main",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

#### Get Cluster Applications

```http
GET /clusters/{clusterId}/apps
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
[
  {
    "id": "uuid",
    "name": "my-app",
    "repo_id": "uuid",
    "path": "apps/my-app",
    "default_branch": "main",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

#### Get Application Details

```http
GET /apps/{appId}
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
{
  "id": "uuid",
  "org_id": "uuid",
  "cluster_id": "uuid",
  "name": "my-app",
  "repo_id": "uuid",
  "path": "apps/my-app",
  "default_branch": "main",
  "current_release": {
    "id": "uuid",
    "image": "myapp:latest",
    "tag": "latest",
    "status": "succeeded",
    "created_at": "2024-01-01T00:00:00Z"
  },
  "release_count": 3,
  "status": "succeeded",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

#### Deploy Application

```http
POST /apps/{appId}/deploy
Authorization: Bearer <jwt-token>
Content-Type: application/json

{
  "image": "myapp:v2.0.0",
  "tag": "v2.0.0"
}
```

**Response (200):**

```json
{
  "release_id": "uuid",
  "status": "pending",
  "message": "Deployment initiated"
}
```

#### Get Application Releases

```http
GET /apps/{appId}/releases
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
[
  {
    "id": "uuid",
    "image": "myapp:v2.0.0",
    "tag": "v2.0.0",
    "created_by": "uuid",
    "status": "succeeded",
    "started_at": "2024-01-01T00:00:00Z",
    "finished_at": "2024-01-01T00:05:00Z",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:05:00Z"
  }
]
```

#### Rollback Application

```http
POST /apps/{appId}/releases/{releaseId}/rollback
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
{
  "release_id": "uuid",
  "status": "pending",
  "message": "Rollback initiated"
}
```

#### Delete Application

```http
DELETE /apps/{appId}
Authorization: Bearer <jwt-token>
```

**Response (204):** No content

### Infrastructure Service Provisioning

#### Provision Services

```http
POST /apps/{appId}/infra/provision
Authorization: Bearer <jwt-token>
Content-Type: application/json

{
  "infra_config": "services:\n  db:\n    chart: bitnami/postgresql\n    env:\n      POSTGRES_DB: webshop\n      POSTGRES_PASSWORD: SECRET::db-password\n  cache:\n    chart: bitnami/redis\n    env:\n      REDIS_PASSWORD: SECRET::redis-password\napp:\n  env:\n    DATABASE_URL: \"postgres://shop:{{services.db.env.POSTGRES_PASSWORD}}@db:5432/webshop\""
}
```

**Response (202):**

```json
{
  "services": [
    {
      "id": "uuid",
      "name": "db",
      "chart": "bitnami/postgresql",
      "status": "pending",
      "namespace": "my-app",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "message": "Provisioning initiated for 1 services"
}
```

#### Get Application Services

```http
GET /apps/{appId}/infra/services
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
[
  {
    "id": "uuid",
    "app_id": "uuid",
    "name": "db",
    "chart": "bitnami/postgresql",
    "status": "running",
    "namespace": "my-app",
    "configs": [
      {
        "id": "uuid",
        "key": "POSTGRES_DB",
        "value": "webshop",
        "is_secret": false
      },
      {
        "id": "uuid",
        "key": "POSTGRES_PASSWORD",
        "value": "***MASKED***",
        "is_secret": true
      }
    ],
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

#### Get Service Configuration

```http
GET /services/{configId}/config
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
{
  "config": {
    "id": "uuid",
    "key": "POSTGRES_PASSWORD",
    "value": "***MASKED***",
    "is_secret": true
  }
}
```

#### Unprovision Service

```http
DELETE /services/{serviceId}
Authorization: Bearer <jwt-token>
```

**Response (202):**

```json
{
  "message": "Service unprovisioning initiated"
}
```

### Git Server Management

#### Create Git Server

```http
POST /orgs/{orgId}/gitservers
Authorization: Bearer <jwt-token>
Content-Type: application/json

{
  "type": "gitea",
  "domain": "gitea.example.com",
  "storage": "10Gi"
}
```

**Response (201):**

```json
{
  "id": "uuid",
  "org_id": "uuid",
  "type": "gitea",
  "domain": "gitea.example.com",
  "storage": "10Gi",
  "status": "pending",
  "config": {
    "admin_user": "",
    "admin_password": "***MASKED***",
    "admin_email": "",
    "repositories": [],
    "settings": {}
  },
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

#### Get Git Servers

```http
GET /orgs/{orgId}/gitservers
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
[
  {
    "id": "uuid",
    "org_id": "uuid",
    "type": "gitea",
    "domain": "gitea.example.com",
    "storage": "10Gi",
    "status": "running",
    "config": {
      "admin_user": "admin",
      "admin_password": "***MASKED***",
      "admin_email": "admin@gitea.example.com",
      "repositories": ["my-repo", "another-repo"],
      "settings": {
        "domain": "gitea.example.com",
        "storage": "10Gi",
        "namespace": "gitea-abc123",
        "release": "gitea-abc123"
      }
    },
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

#### Get Git Server Details

```http
GET /gitservers/{gitServerId}
Authorization: Bearer <jwt-token>
```

**Response (200):** Same as above

#### Delete Git Server

```http
DELETE /gitservers/{gitServerId}
Authorization: Bearer <jwt-token>
```

**Response (204):** No content

### CI Runner Management

#### Create Runner

```http
POST /orgs/{orgId}/runners
Authorization: Bearer <jwt-token>
Content-Type: application/json

{
  "name": "github-runner",
  "type": "github",
  "labels": ["ubuntu", "docker"],
  "nodeSelector": {
    "kubernetes.io/os": "linux"
  },
  "resources": {
    "cpu": "500m",
    "memory": "1Gi"
  },
  "token": "ghp_example_token",
  "url": "https://github.com"
}
```

**Response (201):**

```json
{
  "id": "uuid",
  "org_id": "uuid",
  "name": "github-runner",
  "type": "github",
  "config": {
    "labels": ["ubuntu", "docker"],
    "nodeSelector": {
      "kubernetes.io/os": "linux"
    },
    "resources": {
      "cpu": "500m",
      "memory": "1Gi"
    },
    "token": "***MASKED***",
    "url": "https://github.com",
    "settings": {}
  },
  "status": "pending",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

#### Get Runners

```http
GET /orgs/{orgId}/runners
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
[
  {
    "id": "uuid",
    "org_id": "uuid",
    "name": "github-runner",
    "type": "github",
    "config": {
      "labels": ["ubuntu", "docker"],
      "nodeSelector": {
        "kubernetes.io/os": "linux"
      },
      "resources": {
        "cpu": "500m",
        "memory": "1Gi"
      },
      "token": "***MASKED***",
      "url": "https://github.com",
      "settings": {
        "namespace": "runner-abc123",
        "release": "runner-abc123",
        "deployed_at": "2024-01-01T00:00:00Z"
      }
    },
    "status": "running",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

#### Get Runner Details

```http
GET /runners/{runnerId}
Authorization: Bearer <jwt-token>
```

**Response (200):** Same as above

#### Delete Runner

```http
DELETE /runners/{runnerId}
Authorization: Bearer <jwt-token>
```

**Response (204):** No content

### Job Queue Management

#### Get Jobs for Organization

```http
GET /orgs/{orgId}/jobs
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
[
  {
    "id": "uuid",
    "org_id": "uuid",
    "type": "git_server_install",
    "status": "completed",
    "payload": {
      "git_server_id": "uuid",
      "config": {
        "type": "gitea",
        "domain": "gitea.example.com",
        "storage": "10Gi"
      }
    },
    "error_message": "",
    "created_at": "2024-01-01T00:00:00Z",
    "started_at": "2024-01-01T00:01:00Z",
    "completed_at": "2024-01-01T00:05:00Z"
  }
]
```

## Development

### Available Make Commands

```bash
make help                 # Show all available commands
make build               # Build the application
make run                 # Run the application
make test                # Run tests
make test-coverage       # Run tests with coverage
make clean               # Clean build artifacts
make deps                # Install dependencies
make migrate-up          # Run database migrations up
make migrate-down        # Run database migrations down
make migrate-create      # Create a new migration
make migrate-status      # Check migration status
make docker-build        # Build Docker image
make docker-run          # Run Docker container
make dev-setup           # Setup development environment
make fmt                 # Format code
make lint                # Lint code
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test ./internal/api/handlers/...
```

### Database Migrations

```bash
# Create a new migration
make migrate-create

# Apply migrations
make migrate-up

# Rollback migrations
make migrate-down

# Check migration status
make migrate-status
```

## Docker

### Build and Run with Docker

```bash
# Build Docker image
make docker-build

# Run Docker container
make docker-run
```

### Manual Docker Commands

```bash
# Build image
docker build -t oneclick .

# Run container
docker run --env-file .env -p 8080:8080 oneclick
```

## Testing the API

### Using curl

#### Register a new user:

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "password123"
  }'
```

#### Login:

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

#### Get user profile (replace TOKEN with actual JWT):

```bash
curl -X GET http://localhost:8080/auth/me \
  -H "Authorization: Bearer TOKEN"
```

### Webhook Payload Samples

Sample webhook payloads are available in the `samples/` directory:

- `samples/github-push-payload.json` - GitHub push event payload
- `samples/gitlab-push-payload.json` - GitLab push event payload
- `samples/gitea-push-payload.json` - Gitea push event payload
- `samples/infra-config.yml` - Example infrastructure configuration

For detailed webhook documentation, see `webhook-payloads.md`.

### Infrastructure Configuration Format

OneClick uses YAML-based infrastructure configuration files (`infra-config.yml`) to define services and their dependencies:

```yaml
services:
  # PostgreSQL Database
  db:
    chart: bitnami/postgresql
    env:
      POSTGRES_DB: webshop
      POSTGRES_USER: shop
      POSTGRES_PASSWORD: SECRET::webshop-postgres-password

  # Redis Cache
  cache:
    chart: bitnami/redis
    env:
      REDIS_PASSWORD: SECRET::redis-password

app:
  env:
    # Template substitution using service configurations
    DATABASE_URL: "postgres://shop:{{services.db.env.POSTGRES_PASSWORD}}@db:5432/webshop"
    REDIS_URL: "redis://:{{services.cache.env.REDIS_PASSWORD}}@cache:6379"
```

**Key Features:**

- **Service Definitions**: Define services with Helm charts and environment variables
- **Secret Management**: Use `SECRET::name` markers for sensitive data
- **Template Substitution**: Reference service configurations in app environment variables
- **Helm Integration**: Automatic Helm chart installation and management
- **Background Processing**: Async service provisioning with status tracking

### Deployment Architecture

OneClick uses a background worker architecture for Kubernetes deployments:

1. **Deployment Request**: User triggers deployment via API
2. **Release Creation**: System creates a release record with status "pending"
3. **Background Processing**: Worker picks up deployment job
4. **Kubernetes Deployment**: Worker deploys to cluster using encrypted kubeconfig
5. **Status Updates**: Worker updates release status (running → succeeded/failed)
6. **Manifest Generation**: Automatic Kubernetes YAML generation for deployments

The deployment worker supports:

- Namespace creation and management
- Deployment, Service, ConfigMap, and Ingress creation
- Health check monitoring
- Rollback capabilities
- Error handling and retry logic

### Using the Test Scripts

#### Test Organizations API:

```bash
./test_orgs_api.sh
```

#### Test Clusters API:

```bash
./test_clusters_api.sh
```

#### Test Repositories API:

```bash
./test_repositories_api.sh
```

#### Test Applications API:

```bash
./test_applications_api.sh
```

#### Test Infrastructure API:

```bash
./test_infrastructure_api.sh
```

#### Test Git Servers and Runners API:

```bash
./test_git_runners_api.sh
```

These scripts will:

1. Register a test user
2. Create an organization
3. Test all organization/cluster/repository/application operations
4. Test webhook functionality
5. Test deployment and rollback operations
6. Test infrastructure service provisioning
7. Test git server and runner management
8. Test domain management and SSL certificate automation
9. Clean up test data

### Domain Management

#### Create Domain for Application

```http
POST /apps/{appId}/domains
Authorization: Bearer <jwt-token>
Content-Type: application/json

{
  "domain": "myapp.example.com",
  "provider": "cloudflare",
  "provider_config": {
    "api_key": "your-cloudflare-api-key",
    "email": "admin@example.com",
    "zone_id": "your-zone-id"
  },
  "challenge_type": "dns-01"
}
```

**Response (201):**

```json
{
  "id": "uuid",
  "app_id": "uuid",
  "domain": "myapp.example.com",
  "provider": "cloudflare",
  "provider_config": {
    "api_key": "***MASKED***",
    "email": "admin@example.com",
    "zone_id": "your-zone-id"
  },
  "cert_status": "pending",
  "cert_secret_name": "",
  "challenge_type": "dns-01",
  "dns_instructions": "",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

#### Get Domains for Application

```http
GET /apps/{appId}/domains
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
[
  {
    "id": "uuid",
    "app_id": "uuid",
    "domain": "myapp.example.com",
    "provider": "cloudflare",
    "provider_config": {
      "api_key": "***MASKED***",
      "email": "admin@example.com",
      "zone_id": "your-zone-id"
    },
    "cert_status": "active",
    "cert_secret_name": "myapp.example.com-tls",
    "challenge_type": "dns-01",
    "dns_instructions": "",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:05:00Z"
  }
]
```

#### Get Domain Details

```http
GET /domains/{domainId}
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
{
  "id": "uuid",
  "app_id": "uuid",
  "domain": "myapp.example.com",
  "provider": "cloudflare",
  "provider_config": {
    "api_key": "***MASKED***",
    "email": "admin@example.com",
    "zone_id": "your-zone-id"
  },
  "cert_status": "active",
  "cert_secret_name": "myapp.example.com-tls",
  "challenge_type": "dns-01",
  "dns_instructions": "",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:05:00Z"
}
```

#### Request Certificate

```http
POST /domains/{domainId}/certificates
Authorization: Bearer <jwt-token>
```

**Response (202):**

```json
{
  "message": "Certificate request initiated. Status will be updated shortly.",
  "domain": {
    "id": "uuid",
    "app_id": "uuid",
    "domain": "myapp.example.com",
    "provider": "cloudflare",
    "provider_config": {
      "api_key": "***MASKED***",
      "email": "admin@example.com",
      "zone_id": "your-zone-id"
    },
    "cert_status": "pending",
    "cert_secret_name": "myapp.example.com-tls",
    "challenge_type": "dns-01",
    "dns_instructions": "",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:10:00Z"
  }
}
```

#### Get Certificate Status

```http
GET /domains/{domainId}/certificates
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
{
  "domain_id": "uuid",
  "domain": "myapp.example.com",
  "cert_status": "active",
  "cert_secret_name": "myapp.example.com-tls",
  "issued_at": "2024-01-01T00:05:00Z",
  "expires_at": "2024-04-01T00:05:00Z",
  "issuer": "Let's Encrypt",
  "serial_number": "1234567890abcdef"
}
```

#### Delete Domain

```http
DELETE /domains/{domainId}
Authorization: Bearer <jwt-token>
```

**Response (204):** No content

### Pod Management (Runtime)

#### Get Pods for Application

```http
GET /apps/{appId}/pods
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
{
  "pods": [
    {
      "id": "uuid",
      "app_id": "uuid",
      "name": "nginx-deployment-7d4b8c9f5-abc12",
      "namespace": "default",
      "status": "Running",
      "restarts": 0,
      "ready": "1/1",
      "age": "2h",
      "node_name": "worker-node-1",
      "labels": {
        "app": "nginx",
        "version": "1.0"
      },
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 1
}
```

#### Get Pod Details

```http
GET /pods/{podId}?namespace={namespace}
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
{
  "name": "nginx-deployment-7d4b8c9f5-abc12",
  "namespace": "default",
  "status": "Running",
  "restarts": 0,
  "ready": "1/1",
  "age": "2h",
  "node_name": "worker-node-1",
  "labels": {
    "app": "nginx",
    "version": "1.0"
  },
  "containers": [
    {
      "name": "nginx",
      "image": "nginx:1.21",
      "ready": true,
      "restart_count": 0,
      "state": "Running",
      "started_at": "2024-01-15T10:30:00Z"
    }
  ],
  "events": [
    {
      "type": "Normal",
      "reason": "Created",
      "message": "Created container nginx",
      "count": 1,
      "first_seen": "2024-01-15T10:30:00Z",
      "last_seen": "2024-01-15T10:30:00Z"
    }
  ],
  "owner_refs": [
    {
      "kind": "ReplicaSet",
      "name": "nginx-deployment-7d4b8c9f5",
      "api_version": "apps/v1",
      "controller": true
    }
  ],
  "created_at": "2024-01-15T10:30:00Z",
  "ip": "10.244.1.5",
  "host_ip": "192.168.1.100",
  "phase": "Running",
  "conditions": [
    {
      "type": "Ready",
      "status": "True",
      "last_transition_time": "2024-01-15T10:30:00Z",
      "reason": "PodReady",
      "message": "Pod is ready"
    }
  ]
}
```

#### Get Pod Logs

```http
GET /pods/{podId}/logs?namespace={namespace}&container={container}&tailLines={lines}
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
{
  "pod_name": "nginx-deployment-7d4b8c9f5-abc12",
  "namespace": "default",
  "container": "nginx",
  "logs": "2024/01/15 10:30:00 [notice] 1#1: start worker processes\n2024/01/15 10:30:00 [notice] 1#1: start worker process 1234\n",
  "follow": false
}
```

#### Get Pod Describe Information

```http
GET /pods/{podId}/describe?namespace={namespace}
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
{
  "pod_detail": {
    "name": "nginx-deployment-7d4b8c9f5-abc12",
    "namespace": "default",
    "status": "Running",
    "restarts": 0,
    "ready": "1/1",
    "age": "2h",
    "node_name": "worker-node-1",
    "labels": {
      "app": "nginx",
      "version": "1.0"
    },
    "containers": [
      {
        "name": "nginx",
        "image": "nginx:1.21",
        "ready": true,
        "restart_count": 0,
        "state": "Running",
        "started_at": "2024-01-15T10:30:00Z"
      }
    ],
    "events": [
      {
        "type": "Normal",
        "reason": "Created",
        "message": "Created container nginx",
        "count": 1,
        "first_seen": "2024-01-15T10:30:00Z",
        "last_seen": "2024-01-15T10:30:00Z"
      }
    ],
    "owner_refs": [
      {
        "kind": "ReplicaSet",
        "name": "nginx-deployment-7d4b8c9f5",
        "api_version": "apps/v1",
        "controller": true
      }
    ],
    "created_at": "2024-01-15T10:30:00Z",
    "ip": "10.244.1.5",
    "host_ip": "192.168.1.100",
    "phase": "Running",
    "conditions": [
      {
        "type": "Ready",
        "status": "True",
        "last_transition_time": "2024-01-15T10:30:00Z",
        "reason": "PodReady",
        "message": "Pod is ready"
      }
    ]
  },
  "raw_yaml": null
}
```

#### Execute Command in Pod Terminal (WebSocket)

```http
POST /pods/{podId}/terminal?namespace={namespace}
Authorization: Bearer <jwt-token>
Content-Type: application/json

{
  "container": "nginx",
  "command": ["/bin/bash"]
}
```

**Response (101 Switching Protocols):**

Upgrades to WebSocket connection for real-time terminal interaction.

**WebSocket Message Format:**

- **Input**: Raw terminal input data
- **Output**: Raw terminal output data
- **Resize**: `{"type": "resize", "cols": 80, "rows": 24}`

### Monitoring & Metrics

#### Get Cluster Metrics

```http
GET /clusters/{clusterId}/monitoring?range=5m
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
{
  "cluster_id": "uuid",
  "time_range": "5m",
  "cpu_usage": {
    "metric_name": "cpu_usage",
    "labels": {
      "__name__": "cpu_usage"
    },
    "data_points": [
      {
        "timestamp": "2024-01-01T12:00:00Z",
        "value": 0.5
      },
      {
        "timestamp": "2024-01-01T12:01:00Z",
        "value": 0.6
      }
    ]
  },
  "memory_usage": {
    "metric_name": "memory_usage",
    "labels": {
      "__name__": "memory_usage"
    },
    "data_points": [
      {
        "timestamp": "2024-01-01T12:00:00Z",
        "value": 1073741824
      },
      {
        "timestamp": "2024-01-01T12:01:00Z",
        "value": 1073741824
      }
    ]
  },
  "node_count": 3,
  "healthy_nodes": 3,
  "unhealthy_nodes": 0,
  "timestamp": "2024-01-01T12:05:00Z"
}
```

#### Get Application Metrics

```http
GET /apps/{appId}/monitoring?range=5m
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
{
  "app_id": "uuid",
  "cluster_id": "uuid",
  "time_range": "5m",
  "cpu_usage": {
    "metric_name": "cpu_usage",
    "labels": {
      "namespace": "my-app"
    },
    "data_points": [
      {
        "timestamp": "2024-01-01T12:00:00Z",
        "value": 0.2
      },
      {
        "timestamp": "2024-01-01T12:01:00Z",
        "value": 0.3
      }
    ]
  },
  "memory_usage": {
    "metric_name": "memory_usage",
    "labels": {
      "namespace": "my-app"
    },
    "data_points": [
      {
        "timestamp": "2024-01-01T12:00:00Z",
        "value": 536870912
      },
      {
        "timestamp": "2024-01-01T12:01:00Z",
        "value": 536870912
      }
    ]
  },
  "pod_count": 2,
  "running_pods": 2,
  "pending_pods": 0,
  "failed_pods": 0,
  "top_alerts": [
    {
      "id": "uuid",
      "name": "HighCPUUsage",
      "description": "CPU usage is above 80%",
      "severity": "warning",
      "status": "firing",
      "labels": {
        "namespace": "my-app",
        "instance": "pod-1"
      },
      "starts_at": "2024-01-01T12:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T12:05:00Z"
}
```

#### Get Pod Metrics

```http
GET /pods/{podId}/monitoring?range=5m
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
{
  "pod_id": "uuid",
  "pod_name": "my-app-pod-1",
  "namespace": "my-app",
  "cluster_id": "uuid",
  "cpu_usage": {
    "metric_name": "cpu_usage",
    "labels": {
      "pod": "my-app-pod-1",
      "namespace": "my-app"
    },
    "data_points": [
      {
        "timestamp": "2024-01-01T12:00:00Z",
        "value": 0.1
      },
      {
        "timestamp": "2024-01-01T12:01:00Z",
        "value": 0.15
      }
    ]
  },
  "memory_usage": {
    "metric_name": "memory_usage",
    "labels": {
      "pod": "my-app-pod-1",
      "namespace": "my-app"
    },
    "data_points": [
      {
        "timestamp": "2024-01-01T12:00:00Z",
        "value": 268435456
      },
      {
        "timestamp": "2024-01-01T12:01:00Z",
        "value": 268435456
      }
    ]
  },
  "status": "Running",
  "restarts": 0,
  "timestamp": "2024-01-01T12:05:00Z"
}
```

#### Get Cluster Alerts

```http
GET /clusters/{clusterId}/alerts?limit=10
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
{
  "alerts": [
    {
      "id": "uuid",
      "name": "HighCPUUsage",
      "description": "CPU usage is above 80%",
      "severity": "warning",
      "status": "firing",
      "labels": {
        "instance": "node1",
        "job": "kubernetes-nodes"
      },
      "starts_at": "2024-01-01T12:00:00Z"
    },
    {
      "id": "uuid",
      "name": "HighMemoryUsage",
      "description": "Memory usage is above 90%",
      "severity": "critical",
      "status": "firing",
      "labels": {
        "instance": "node2",
        "job": "kubernetes-nodes"
      },
      "starts_at": "2024-01-01T12:00:00Z"
    }
  ],
  "count": 2
}
```

#### Get Monitoring Health

```http
GET /monitoring/health
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "services": {
    "prometheus": "healthy",
    "cache": "healthy",
    "rate_limit": "healthy"
  }
}
```

## Configuration

The application uses environment variables for configuration. See `.env.example` for all available options:

- `DATABASE_URL`: PostgreSQL connection string
- `JWT_SECRET`: Secret key for JWT token signing
- `ONECLICK_MASTER_KEY`: Master key for encryption (32 bytes)
- `PORT`: Server port (default: 8080)
- `LOG_LEVEL`: Logging level (debug, info, warn, error)
- `GIN_MODE`: Gin mode (debug, release, test)

## Security Notes

- Change default JWT secret and master key in production
- Use HTTPS in production
- Implement rate limiting for production use
- Consider implementing refresh tokens for better security
- Validate and sanitize all inputs
- Kubeconfigs are encrypted using AES-GCM before storage
- Ensure `ONECLICK_MASTER_KEY` is exactly 32 bytes for proper encryption
- Cluster health checks validate kubeconfig connectivity before storing
- Repository access tokens are encrypted using AES-GCM before storage
- Webhook signatures are verified using HMAC-SHA256 for security
- Webhook endpoints are public but require signature verification
- Application deployments use encrypted kubeconfigs for Kubernetes access
- Background workers process deployments securely with proper error handling
- Release metadata is stored securely with deployment history tracking
- Infrastructure service provisioning requires Admin/Owner permissions
- Service configurations with secrets are properly masked in API responses
- Helm chart installations use secure Kubernetes client connections

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run tests and ensure they pass
6. Submit a pull request

## License

[Add your license here]
