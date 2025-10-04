# OneClick Backend API Documentation

This directory contains comprehensive API documentation for the OneClick Backend service, including Swagger/OpenAPI specifications and Postman collections.

## 📁 Files Overview

- **`swagger.yaml`** - Complete OpenAPI 3.0.3 specification
- **`postman-collection.json`** - Postman collection with pre-configured requests
- **`event-api.md`** - Detailed documentation for Event Logging & Audit Trail API
- **`monitoring-api.md`** - Detailed documentation for Monitoring & Metrics API
- **`pipeline-api.md`** - Detailed documentation for Pipeline Management API

## 🚀 Quick Start

### Using Swagger UI

1. **Local Development:**

   ```bash
   # Start the OneClick server
   go run cmd/server/main.go

   # Access Swagger UI at:
   # http://localhost:8080/swagger-ui/ (if Swagger UI is integrated)
   # Or use online Swagger Editor: https://editor.swagger.io/
   ```

2. **Import the Swagger file:**
   - Copy the contents of `swagger.yaml`
   - Paste into [Swagger Editor](https://editor.swagger.io/)
   - Or use any OpenAPI-compatible tool

### Using Postman

1. **Import Collection:**

   - Open Postman
   - Click "Import" button
   - Select `postman-collection.json` file
   - The collection will be imported with all endpoints

2. **Set Environment Variables:**

   - Create a new environment in Postman
   - Add the following variables:
     ```
     base_url: http://localhost:8080
     access_token: (will be set automatically after login)
     org_id: (will be set automatically after creating organization)
     user_id: (will be set automatically after login)
     ```

3. **Authentication Flow:**
   - Run "Authentication > Register User" to create an account
   - Run "Authentication > Login User" to get access token
   - The token will be automatically set for all subsequent requests

## 📋 API Endpoints Overview

### 🔐 Authentication

- `POST /auth/register` - Create new user account
- `POST /auth/login` - Authenticate user and get JWT token
- `GET /auth/me` - Get current user information

### 🏢 Organizations

- `POST /orgs` - Create new organization
- `GET /orgs` - Get user's organizations
- `GET /orgs/{orgId}` - Get organization details
- `DELETE /orgs/{orgId}` - Delete organization (Owner only)

### 📋 Event Logging & Audit Trail

- `GET /orgs/{orgId}/events` - Get organization events with filtering
- `GET /events/{eventId}` - Get specific event details
- `GET /orgs/{orgId}/dashboard` - Get dashboard counts
- `POST /orgs/{orgId}/dashboard/refresh` - Refresh dashboard counts
- `GET /orgs/{orgId}/readmodel` - Get read model projects
- `GET /orgs/{orgId}/readmodel/{key}` - Get specific read model project
- `POST /orgs/{orgId}/readmodel` - Create read model project
- `DELETE /orgs/{orgId}/readmodel/{key}` - Delete read model project

### ☸️ Clusters

- `POST /orgs/{orgId}/clusters` - Create new cluster
- `GET /orgs/{orgId}/clusters` - Get organization clusters
- `GET /clusters/{clusterId}` - Get cluster details
- `GET /clusters/{clusterId}/status` - Get cluster health
- `DELETE /clusters/{clusterId}` - Delete cluster (Admin/Owner only)

### 🚀 Applications

- `POST /clusters/{clusterId}/apps` - Create new application
- `GET /clusters/{clusterId}/apps` - Get cluster applications
- `GET /apps/{appId}` - Get application details
- `POST /apps/{appId}/deploy` - Deploy application
- `GET /apps/{appId}/releases` - Get application releases
- `POST /apps/{appId}/releases/{releaseId}/rollback` - Rollback application
- `DELETE /apps/{appId}` - Delete application (Admin/Owner only)

### 🔄 Pipelines

- `POST /apps/{appId}/pipelines` - Trigger pipeline
- `GET /apps/{appId}/pipelines` - Get application pipelines
- `GET /pipelines/{pipelineId}` - Get pipeline details
- `GET /pipelines/{pipelineId}/logs` - Get pipeline logs

### 📊 Monitoring

- `GET /monitoring/health` - Check monitoring service health
- `GET /clusters/{clusterId}/monitoring` - Get cluster metrics
- `GET /clusters/{clusterId}/alerts` - Get cluster alerts
- `GET /apps/{appId}/monitoring` - Get application metrics
- `GET /pods/{podId}/monitoring` - Get pod metrics

### 🔗 Webhooks

- `POST /hooks/git` - Git webhook endpoint
- `GET /hooks/test` - Test webhook endpoint

## 🔑 Authentication

All endpoints (except authentication and webhooks) require JWT authentication:

```http
Authorization: Bearer <jwt-token>
```

### Getting a Token

1. **Register a new user:**

   ```bash
   curl -X POST http://localhost:8080/auth/register \
     -H "Content-Type: application/json" \
     -d '{
       "name": "John Doe",
       "email": "john@example.com",
       "password": "password123"
     }'
   ```

2. **Login to get token:**

   ```bash
   curl -X POST http://localhost:8080/auth/login \
     -H "Content-Type: application/json" \
     -d '{
       "email": "john@example.com",
       "password": "password123"
     }'
   ```

3. **Use the token in subsequent requests:**
   ```bash
   curl -H "Authorization: Bearer <your-jwt-token>" \
     http://localhost:8080/auth/me
   ```

## 📊 Event Logging Examples

### Get Recent Events

```bash
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8080/orgs/{orgId}/events?limit=10"
```

### Filter Events by Action

```bash
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8080/orgs/{orgId}/events?action=app_created&limit=5"
```

### Get Dashboard Summary

```bash
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8080/orgs/{orgId}/dashboard"
```

### Get Failed Pipelines Summary

```bash
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8080/orgs/{orgId}/readmodel/recent_failed_pipelines"
```

## 🔧 Error Handling

The API uses standard HTTP status codes:

- **200** - Success
- **201** - Created
- **204** - No Content
- **400** - Bad Request
- **401** - Unauthorized
- **403** - Forbidden
- **404** - Not Found
- **409** - Conflict
- **500** - Internal Server Error

### Error Response Format

```json
{
  "error": "Error message description"
}
```

## 🚀 Development Workflow

### 1. Start the Server

```bash
# Set environment variables
export DATABASE_URL="postgres://user:password@localhost/oneclick?sslmode=disable"
export JWT_SECRET="your-secret-key"
export ONECLICK_MASTER_KEY="your-32-byte-master-key"

# Run the server
go run cmd/server/main.go
```

### 2. Test with Postman

1. Import the Postman collection
2. Set up environment variables
3. Run the authentication flow
4. Test various endpoints

### 3. Validate with Swagger

1. Open Swagger Editor
2. Import the `swagger.yaml` file
3. Test endpoints directly from the UI

## 📝 API Versioning

The API follows semantic versioning:

- **Current Version:** 1.0.0
- **Version Header:** `Accept: application/vnd.oneclick.v1+json`

## 🔒 Security Considerations

- All sensitive endpoints require JWT authentication
- Organization access is enforced through middleware
- Rate limiting is applied to prevent abuse
- Input validation and sanitization on all endpoints
- CORS headers configured for web applications

## 📚 Additional Resources

- **Main README:** [../README.md](../README.md)
- **Event API Details:** [event-api.md](event-api.md)
- **Monitoring API Details:** [monitoring-api.md](monitoring-api.md)
- **Pipeline API Details:** [pipeline-api.md](pipeline-api.md)

## 🤝 Contributing

When adding new endpoints:

1. Update `swagger.yaml` with new paths and schemas
2. Add corresponding requests to `postman-collection.json`
3. Update this README with endpoint descriptions
4. Add detailed documentation in appropriate `.md` files

## 📞 Support

For API-related questions or issues:

- Create an issue in the repository
- Check the detailed API documentation files
- Review the Swagger specification for complete details
