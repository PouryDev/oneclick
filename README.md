# OneClick Backend

A Go backend service built with Clean Architecture principles, featuring authentication, user management, organization management, Kubernetes cluster management, repository integration, webhook processing, and application deployment with release management.

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
- **Kubernetes**: client-go for cluster management and deployments
- **Encryption**: AES-GCM for kubeconfig and token encryption
- **Webhooks**: HMAC signature verification for Git providers
- **Deployments**: Background worker for Kubernetes deployments

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

For detailed webhook documentation, see `webhook-payloads.md`.

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

These scripts will:

1. Register a test user
2. Create an organization
3. Test all organization/cluster/repository/application operations
4. Test webhook functionality
5. Test deployment and rollback operations
6. Clean up test data

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

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run tests and ensure they pass
6. Submit a pull request

## License

[Add your license here]
