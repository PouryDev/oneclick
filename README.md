# OneClick Backend

A Go backend service built with Clean Architecture principles, featuring authentication and user management.

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
│   │   └── services/     # Business logic services
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

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run tests and ensure they pass
6. Submit a pull request

## License

[Add your license here]
