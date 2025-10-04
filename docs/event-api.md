# Event Logging & Audit Trail API Documentation

This document describes the Event Logging and Audit Trail API endpoints for the OneClick backend service.

## Overview

The Event Logging system provides comprehensive audit trails for all system operations, including:

- Application lifecycle events (created, updated, deleted)
- Cluster management events (imported, updated, deleted)
- Pipeline execution events (started, completed, failed)
- Release management events
- User activity tracking

## Event Types

### Event Actions

- `app_created` - Application was created
- `app_updated` - Application was updated
- `app_deleted` - Application was deleted
- `cluster_imported` - Cluster was imported
- `cluster_updated` - Cluster was updated
- `cluster_deleted` - Cluster was deleted
- `pipeline_started` - Pipeline execution started
- `pipeline_completed` - Pipeline execution completed
- `pipeline_failed` - Pipeline execution failed
- `release_created` - Release was created
- `release_updated` - Release was updated
- `release_deleted` - Release was deleted
- `user_joined` - User joined organization
- `user_left` - User left organization

### Resource Types

- `app` - Application resource
- `cluster` - Cluster resource
- `pipeline` - Pipeline resource
- `release` - Release resource
- `user` - User resource
- `organization` - Organization resource

## API Endpoints

### Get Organization Events

Retrieve events for a specific organization with optional filtering.

```http
GET /orgs/{orgId}/events?limit=50&offset=0&action=app_created&resource_type=app
Authorization: Bearer <jwt-token>
```

**Query Parameters:**

- `limit` (optional): Number of events to return (default: 50, max: 100)
- `offset` (optional): Number of events to skip (default: 0)
- `action` (optional): Filter by event action
- `resource_type` (optional): Filter by resource type

**Response (200):**

```json
{
  "events": [
    {
      "id": "uuid",
      "org_id": "uuid",
      "user_id": "uuid",
      "action": "app_created",
      "resource_type": "app",
      "resource_id": "uuid",
      "details": {
        "name": "my-app",
        "cluster_id": "uuid",
        "repository_id": "uuid"
      },
      "created_at": "2024-01-01T12:00:00Z"
    }
  ],
  "limit": 50,
  "offset": 0,
  "count": 1
}
```

### Get Event by ID

Retrieve a specific event by its ID.

```http
GET /events/{eventId}
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
{
  "id": "uuid",
  "org_id": "uuid",
  "user_id": "uuid",
  "action": "pipeline_started",
  "resource_type": "pipeline",
  "resource_id": "uuid",
  "details": {
    "app_name": "my-app",
    "branch": "main",
    "commit_sha": "abc123"
  },
  "created_at": "2024-01-01T12:00:00Z"
}
```

### Get Dashboard Counts

Retrieve aggregated counts for the organization dashboard.

```http
GET /orgs/{orgId}/dashboard
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
{
  "org_id": "uuid",
  "apps_count": 15,
  "clusters_count": 3,
  "running_pipelines": 2,
  "updated_at": "2024-01-01T12:00:00Z"
}
```

### Refresh Dashboard Counts

Manually refresh the dashboard counts for an organization.

```http
POST /orgs/{orgId}/dashboard/refresh
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
{
  "org_id": "uuid",
  "apps_count": 15,
  "clusters_count": 3,
  "running_pipelines": 2,
  "updated_at": "2024-01-01T12:00:00Z"
}
```

### Get Read Model Projects

Retrieve all read model projects for an organization.

```http
GET /orgs/{orgId}/readmodel
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
{
  "projects": [
    {
      "id": "uuid",
      "org_id": "uuid",
      "key": "recent_failed_pipelines",
      "value": {
        "pipelines": [
          {
            "id": "uuid",
            "app_id": "uuid",
            "app_name": "my-app",
            "commit_sha": "abc123",
            "finished_at": "2024-01-01T12:00:00Z"
          }
        ],
        "count": 1,
        "updated_at": "2024-01-01T12:00:00Z"
      },
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  ],
  "count": 1
}
```

### Get Read Model Project by Key

Retrieve a specific read model project by its key.

```http
GET /orgs/{orgId}/readmodel/{key}
Authorization: Bearer <jwt-token>
```

**Response (200):**

```json
{
  "id": "uuid",
  "org_id": "uuid",
  "key": "top_apps_by_deployments",
  "value": {
    "apps": [
      {
        "id": "uuid",
        "name": "my-app",
        "deployment_count": 25
      }
    ],
    "updated_at": "2024-01-01T12:00:00Z"
  },
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

### Create Read Model Project

Create a new read model project.

```http
POST /orgs/{orgId}/readmodel
Authorization: Bearer <jwt-token>
Content-Type: application/json

{
  "key": "custom_analytics",
  "value": {
    "metric": "custom_value",
    "data": "analytics_data"
  }
}
```

**Response (201):**

```json
{
  "id": "uuid",
  "org_id": "uuid",
  "key": "custom_analytics",
  "value": {
    "metric": "custom_value",
    "data": "analytics_data"
  },
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

### Delete Read Model Project

Delete a read model project by its key.

```http
DELETE /orgs/{orgId}/readmodel/{key}
Authorization: Bearer <jwt-token>
```

**Response (204):** No Content

## Error Responses

### 400 Bad Request

```json
{
  "error": "Invalid organization ID"
}
```

### 401 Unauthorized

```json
{
  "error": "User not authenticated"
}
```

### 403 Forbidden

```json
{
  "error": "User does not have access to organization"
}
```

### 404 Not Found

```json
{
  "error": "Event not found"
}
```

### 500 Internal Server Error

```json
{
  "error": "Internal server error"
}
```

## Background Processing

The system includes a background projector worker that runs every minute to:

1. **Update Dashboard Counts**: Aggregates counts for applications, clusters, and running pipelines
2. **Generate Read Model Projections**: Creates denormalized data for performance optimization
3. **Maintain Data Consistency**: Ensures read models stay synchronized with source data

### Read Model Projections

The system automatically generates the following read model projections:

- `recent_failed_pipelines`: List of recent failed pipelines with details
- `top_apps_by_deployments`: Applications ranked by deployment count
- `cluster_health_summary`: Cluster health and application distribution

## Security Considerations

- All endpoints require authentication via JWT tokens
- Organization access is enforced through middleware
- Event data includes user identification for audit trails
- Sensitive information in event details is properly handled
- Rate limiting is applied to prevent abuse

## Usage Examples

### Get Recent Application Events

```bash
curl -H "Authorization: Bearer <token>" \
  "https://api.oneclick.com/orgs/{orgId}/events?action=app_created&limit=10"
```

### Get Dashboard Summary

```bash
curl -H "Authorization: Bearer <token>" \
  "https://api.oneclick.com/orgs/{orgId}/dashboard"
```

### Get Failed Pipelines Summary

```bash
curl -H "Authorization: Bearer <token>" \
  "https://api.oneclick.com/orgs/{orgId}/readmodel/recent_failed_pipelines"
```
