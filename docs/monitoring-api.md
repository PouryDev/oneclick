# Monitoring API Documentation

## Overview

The Monitoring API provides endpoints to query Prometheus metrics for clusters, applications, and pods. It includes caching, rate limiting, and proper authorization.

## API Endpoints

### Cluster Monitoring

#### GET /clusters/{clusterId}/monitoring

Retrieves cluster-level metrics including CPU usage, memory usage, node counts, and health status.

**Query Parameters:**

- `range` (optional): Time range for metrics (`5m`, `15m`, `1h`, `6h`, `24h`). Default: `5m`
- `start` (optional): Start time in RFC3339 format
- `end` (optional): End time in RFC3339 format
- `step` (optional): Query step size. Default: `1m`

**Example Request:**

```bash
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8080/clusters/123e4567-e89b-12d3-a456-426614174000/monitoring?range=5m"
```

**Example Response:**

```json
{
  "cluster_id": "123e4567-e89b-12d3-a456-426614174000",
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

### Application Monitoring

#### GET /apps/{appId}/monitoring

Retrieves application-level metrics including CPU usage, memory usage, pod counts, and top alerts.

**Query Parameters:**

- `range` (optional): Time range for metrics (`5m`, `15m`, `1h`, `6h`, `24h`). Default: `5m`
- `start` (optional): Start time in RFC3339 format
- `end` (optional): End time in RFC3339 format
- `step` (optional): Query step size. Default: `1m`

**Example Request:**

```bash
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8080/apps/123e4567-e89b-12d3-a456-426614174000/monitoring?range=5m"
```

**Example Response:**

```json
{
  "app_id": "123e4567-e89b-12d3-a456-426614174000",
  "cluster_id": "456e7890-e89b-12d3-a456-426614174000",
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
      "id": "789e0123-e89b-12d3-a456-426614174000",
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

### Pod Monitoring

#### GET /pods/{podId}/monitoring

Retrieves pod-level metrics including CPU usage, memory usage, and status.

**Query Parameters:**

- `range` (optional): Time range for metrics (`5m`, `15m`, `1h`, `6h`, `24h`). Default: `5m`
- `start` (optional): Start time in RFC3339 format
- `end` (optional): End time in RFC3339 format
- `step` (optional): Query step size. Default: `1m`

**Example Request:**

```bash
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8080/pods/123e4567-e89b-12d3-a456-426614174000/monitoring?range=5m"
```

**Example Response:**

```json
{
  "pod_id": "123e4567-e89b-12d3-a456-426614174000",
  "pod_name": "my-app-pod-1",
  "namespace": "my-app",
  "cluster_id": "456e7890-e89b-12d3-a456-426614174000",
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

### Alerts

#### GET /clusters/{clusterId}/alerts

Retrieves active alerts for a cluster.

**Query Parameters:**

- `limit` (optional): Maximum number of alerts to return (1-100). Default: `10`

**Example Request:**

```bash
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8080/clusters/123e4567-e89b-12d3-a456-426614174000/alerts?limit=5"
```

**Example Response:**

```json
{
  "alerts": [
    {
      "id": "789e0123-e89b-12d3-a456-426614174000",
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
      "id": "789e0123-e89b-12d3-a456-426614174001",
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

### Health Check

#### GET /monitoring/health

Checks the health of monitoring services.

**Example Request:**

```bash
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8080/monitoring/health"
```

**Example Response:**

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

## PromQL Queries

The monitoring service uses the following PromQL queries:

### Cluster-Level Queries

```promql
# CPU Usage
sum(rate(container_cpu_usage_seconds_total{container!="POD",container!=""}[5m]))

# Memory Usage
sum(container_memory_usage_bytes{container!="POD",container!=""})

# Node Count
count(kube_node_info)

# Healthy Nodes
count(kube_node_status_condition{condition="Ready",status="true"})
```

### Application-Level Queries

```promql
# CPU Usage by Namespace
sum(rate(container_cpu_usage_seconds_total{namespace="my-app",container!="POD",container!=""}[5m])) by (namespace)

# Memory Usage by Namespace
sum(container_memory_usage_bytes{namespace="my-app",container!="POD",container!=""}) by (namespace)

# Pod Count by Namespace
count(kube_pod_info{namespace="my-app"})

# Running Pods by Namespace
count(kube_pod_status_phase{namespace="my-app",phase="Running"})

# Pending Pods by Namespace
count(kube_pod_status_phase{namespace="my-app",phase="Pending"})

# Failed Pods by Namespace
count(kube_pod_status_phase{namespace="my-app",phase="Failed"})
```

### Pod-Level Queries

```promql
# CPU Usage by Pod
rate(container_cpu_usage_seconds_total{pod="my-app-pod-1",namespace="my-app",container!="POD",container!=""}[5m])

# Memory Usage by Pod
container_memory_usage_bytes{pod="my-app-pod-1",namespace="my-app",container!="POD",container!=""}

# Pod Restart Count
kube_pod_container_status_restarts_total{pod="my-app-pod-1",namespace="my-app"}
```

## Error Responses

### Rate Limit Exceeded

```json
{
  "error": "Rate limit exceeded"
}
```

Status: `429 Too Many Requests`

### Access Denied

```json
{
  "error": "Access denied"
}
```

Status: `403 Forbidden`

### Resource Not Found

```json
{
  "error": "Cluster not found"
}
```

Status: `404 Not Found`

### Invalid Parameters

```json
{
  "error": "Invalid time range"
}
```

Status: `400 Bad Request`

## Rate Limiting

- **Limit**: 100 requests per minute per user
- **Window**: 1 minute sliding window
- **Headers**: Rate limit information is not currently exposed in response headers

## Caching

- **Cluster Metrics**: 30 seconds TTL
- **Application Metrics**: 30 seconds TTL
- **Alerts**: 1 minute TTL
- **Cache Type**: In-memory cache with automatic expiration

## Authentication

All endpoints require JWT authentication via the `Authorization: Bearer <token>` header.

## Time Ranges

Supported time ranges:

- `5m` - 5 minutes
- `15m` - 15 minutes
- `1h` - 1 hour
- `6h` - 6 hours
- `24h` - 24 hours

## Implementation Notes

1. **Prometheus Integration**: The service assumes Prometheus is installed in the cluster and accessible via port-forwarding or in-cluster URL.

2. **Authorization**: All requests are authorized against the user's organization membership.

3. **Error Handling**: Comprehensive error handling with specific error messages for different failure scenarios.

4. **Performance**: Caching and rate limiting ensure good performance and prevent abuse.

5. **Extensibility**: The service is designed to be easily extended with additional metrics and query types.
