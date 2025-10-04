# Pipeline API Documentation

This document provides examples for using the Pipeline API endpoints to manage CI/CD pipelines for applications.

## Authentication

All endpoints require JWT authentication. Include the JWT token in the Authorization header:

```bash
Authorization: Bearer <your-jwt-token>
```

## Endpoints

### Trigger Pipeline

Trigger a new pipeline for an application.

**POST** `/apps/{appId}/pipelines`

**Request Body:**
```json
{
  "branch": "main",
  "commit_sha": "abc123def456"
}
```

**Example:**
```bash
curl -X POST "http://localhost:8080/apps/123e4567-e89b-12d3-a456-426614174000/pipelines" \
  -H "Authorization: Bearer <your-jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "branch": "main",
    "commit_sha": "abc123def456"
  }'
```

**Response:**
```json
{
  "id": "456e7890-e89b-12d3-a456-426614174001",
  "app_id": "123e4567-e89b-12d3-a456-426614174000",
  "repo_id": "789e0123-e89b-12d3-a456-426614174002",
  "commit_sha": "abc123def456",
  "status": "pending",
  "triggered_by": "012e3456-e89b-12d3-a456-426614174003",
  "logs_url": null,
  "started_at": null,
  "finished_at": null,
  "meta": {
    "branch": "main"
  },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### Get Pipelines by Application

Get all pipelines for a specific application.

**GET** `/apps/{appId}/pipelines?limit=20&offset=0`

**Query Parameters:**
- `limit` (optional): Number of pipelines to return (default: 20)
- `offset` (optional): Number of pipelines to skip (default: 0)

**Example:**
```bash
curl -X GET "http://localhost:8080/apps/123e4567-e89b-12d3-a456-426614174000/pipelines?limit=10&offset=0" \
  -H "Authorization: Bearer <your-jwt-token>"
```

**Response:**
```json
[
  {
    "id": "456e7890-e89b-12d3-a456-426614174001",
    "app_id": "123e4567-e89b-12d3-a456-426614174000",
    "repo_id": "789e0123-e89b-12d3-a456-426614174002",
    "commit_sha": "abc123def456",
    "status": "succeeded",
    "triggered_by": "012e3456-e89b-12d3-a456-426614174003",
    "logs_url": "https://logs.example.com/pipeline/456e7890-e89b-12d3-a456-426614174001",
    "started_at": "2024-01-15T10:30:00Z",
    "finished_at": "2024-01-15T10:35:00Z",
    "meta": {
      "branch": "main"
    },
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:35:00Z"
  },
  {
    "id": "567e8901-e89b-12d3-a456-426614174004",
    "app_id": "123e4567-e89b-12d3-a456-426614174000",
    "repo_id": "789e0123-e89b-12d3-a456-426614174002",
    "commit_sha": "def456ghi789",
    "status": "failed",
    "triggered_by": "012e3456-e89b-12d3-a456-426614174003",
    "logs_url": "https://logs.example.com/pipeline/567e8901-e89b-12d3-a456-426614174004",
    "started_at": "2024-01-15T11:00:00Z",
    "finished_at": "2024-01-15T11:02:00Z",
    "meta": {
      "branch": "feature/new-feature"
    },
    "created_at": "2024-01-15T11:00:00Z",
    "updated_at": "2024-01-15T11:02:00Z"
  }
]
```

### Get Pipeline Details

Get detailed information about a specific pipeline including its steps.

**GET** `/pipelines/{pipelineId}`

**Example:**
```bash
curl -X GET "http://localhost:8080/pipelines/456e7890-e89b-12d3-a456-426614174001" \
  -H "Authorization: Bearer <your-jwt-token>"
```

**Response:**
```json
{
  "id": "456e7890-e89b-12d3-a456-426614174001",
  "app_id": "123e4567-e89b-12d3-a456-426614174000",
  "repo_id": "789e0123-e89b-12d3-a456-426614174002",
  "commit_sha": "abc123def456",
  "status": "succeeded",
  "triggered_by": "012e3456-e89b-12d3-a456-426614174003",
  "logs_url": "https://logs.example.com/pipeline/456e7890-e89b-12d3-a456-426614174001",
  "started_at": "2024-01-15T10:30:00Z",
  "finished_at": "2024-01-15T10:35:00Z",
  "meta": {
    "branch": "main"
  },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:35:00Z",
  "steps": [
    {
      "id": "678e9012-e89b-12d3-a456-426614174005",
      "pipeline_id": "456e7890-e89b-12d3-a456-426614174001",
      "name": "checkout",
      "status": "succeeded",
      "started_at": "2024-01-15T10:30:00Z",
      "finished_at": "2024-01-15T10:30:30Z",
      "logs": "DRY RUN: Executing step 'checkout'\nDRY RUN: Step 'checkout' completed successfully\n",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:30Z"
    },
    {
      "id": "789e0123-e89b-12d3-a456-426614174006",
      "pipeline_id": "456e7890-e89b-12d3-a456-426614174001",
      "name": "build",
      "status": "succeeded",
      "started_at": "2024-01-15T10:30:30Z",
      "finished_at": "2024-01-15T10:32:00Z",
      "logs": "DRY RUN: Executing step 'build'\nDRY RUN: Step 'build' completed successfully\n",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:32:00Z"
    },
    {
      "id": "890e1234-e89b-12d3-a456-426614174007",
      "pipeline_id": "456e7890-e89b-12d3-a456-426614174001",
      "name": "test",
      "status": "succeeded",
      "started_at": "2024-01-15T10:32:00Z",
      "finished_at": "2024-01-15T10:33:30Z",
      "logs": "DRY RUN: Executing step 'test'\nDRY RUN: Step 'test' completed successfully\n",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:33:30Z"
    },
    {
      "id": "901e2345-e89b-12d3-a456-426614174008",
      "pipeline_id": "456e7890-e89b-12d3-a456-426614174001",
      "name": "deploy",
      "status": "succeeded",
      "started_at": "2024-01-15T10:33:30Z",
      "finished_at": "2024-01-15T10:35:00Z",
      "logs": "DRY RUN: Executing step 'deploy'\nDRY RUN: Step 'deploy' completed successfully\n",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:35:00Z"
    }
  ]
}
```

### Get Pipeline Logs

Get aggregated logs for a specific pipeline.

**GET** `/pipelines/{pipelineId}/logs`

**Example:**
```bash
curl -X GET "http://localhost:8080/pipelines/456e7890-e89b-12d3-a456-426614174001/logs" \
  -H "Authorization: Bearer <your-jwt-token>"
```

**Response:**
```json
{
  "pipeline_id": "456e7890-e89b-12d3-a456-426614174001",
  "logs": "=== Pipeline Logs ===\n\n--- Step: checkout ---\nDRY RUN: Executing step 'checkout'\nDRY RUN: Step 'checkout' completed successfully\n\n--- Step: build ---\nDRY RUN: Executing step 'build'\nDRY RUN: Step 'build' completed successfully\n\n--- Step: test ---\nDRY RUN: Executing step 'test'\nDRY RUN: Step 'test' completed successfully\n\n--- Step: deploy ---\nDRY RUN: Executing step 'deploy'\nDRY RUN: Step 'deploy' completed successfully\n\n=== Pipeline Completed Successfully ===\n",
  "status": "succeeded",
  "started_at": "2024-01-15T10:30:00Z",
  "finished_at": "2024-01-15T10:35:00Z"
}
```

## Pipeline Status Values

- `pending`: Pipeline is queued and waiting to start
- `running`: Pipeline is currently executing
- `succeeded`: Pipeline completed successfully
- `failed`: Pipeline failed during execution
- `cancelled`: Pipeline was cancelled

## Pipeline Step Status Values

- `pending`: Step is waiting to execute
- `running`: Step is currently executing
- `succeeded`: Step completed successfully
- `failed`: Step failed during execution
- `cancelled`: Step was cancelled

## Error Responses

### 400 Bad Request
```json
{
  "error": "Invalid request",
  "message": "Branch and commit_sha are required"
}
```

### 401 Unauthorized
```json
{
  "error": "Unauthorized",
  "message": "Invalid or missing JWT token"
}
```

### 403 Forbidden
```json
{
  "error": "Forbidden",
  "message": "You don't have permission to access this resource"
}
```

### 404 Not Found
```json
{
  "error": "Not Found",
  "message": "Pipeline not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "Internal Server Error",
  "message": "An unexpected error occurred"
}
```

## MVP Implementation Notes

The current implementation includes:

1. **Dry-run Mode**: All pipeline executions run in dry-run mode for security reasons
2. **Simulated Steps**: Default steps (checkout, build, test, deploy) are simulated
3. **Mock Logs**: Generated logs simulate successful execution
4. **Job Queue Integration**: Pipelines are processed asynchronously via the job queue
5. **Database Storage**: Pipeline and step data is stored in PostgreSQL

## Security Considerations

⚠️ **Important**: The current MVP implementation includes dry-run mode to prevent remote code execution risks. For production use:

1. **Implement proper runner controllers** (actions-runner-controller, GitLab Runner Controller)
2. **Add security isolation** and sandboxing
3. **Implement resource limits** and validation
4. **Add input sanitization** and validation
5. **Use self-hosted runners** with proper security controls

## Future Enhancements

- Integration with CI/CD platforms (GitHub Actions, GitLab CI)
- Artifact storage and management
- Pipeline caching and optimization
- Integration with container registries
- Security scanning and compliance checks
- Custom pipeline step definitions
- Pipeline templates and reusable workflows
