# Pod Management API Examples

This document provides examples for using the Pod Management API endpoints.

## Authentication

All API endpoints require authentication via JWT token. Include the token in the Authorization header:

```bash
Authorization: Bearer <your-jwt-token>
```

## Base URL

```
http://localhost:8080/api/v1
```

## 1. Get Pods for an Application

### Request

```bash
curl -X GET "http://localhost:8080/api/v1/apps/{appId}/pods" \
  -H "Authorization: Bearer <your-jwt-token>" \
  -H "Content-Type: application/json"
```

### Example

```bash
curl -X GET "http://localhost:8080/api/v1/apps/123e4567-e89b-12d3-a456-426614174000/pods" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json"
```

### Response

```json
{
  "pods": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "app_id": "123e4567-e89b-12d3-a456-426614174000",
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

## 2. Get Pod Details

### Request

```bash
curl -X GET "http://localhost:8080/api/v1/pods/{podId}?namespace={namespace}" \
  -H "Authorization: Bearer <your-jwt-token>" \
  -H "Content-Type: application/json"
```

### Example

```bash
curl -X GET "http://localhost:8080/api/v1/pods/nginx-deployment-7d4b8c9f5-abc12?namespace=default" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json"
```

### Response

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

## 3. Get Pod Logs (Non-streaming)

### Request

```bash
curl -X GET "http://localhost:8080/api/v1/pods/{podId}/logs?namespace={namespace}&container={container}&tailLines={lines}" \
  -H "Authorization: Bearer <your-jwt-token>" \
  -H "Content-Type: application/json"
```

### Example

```bash
curl -X GET "http://localhost:8080/api/v1/pods/nginx-deployment-7d4b8c9f5-abc12/logs?namespace=default&container=nginx&tailLines=50" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json"
```

### Response

```json
{
  "pod_name": "nginx-deployment-7d4b8c9f5-abc12",
  "namespace": "default",
  "container": "nginx",
  "logs": "2024/01/15 10:30:00 [notice] 1#1: start worker processes\n2024/01/15 10:30:00 [notice] 1#1: start worker process 1234\n2024/01/15 10:30:00 [notice] 1#1: start worker process 1235\n",
  "follow": false
}
```

## 4. Get Pod Describe Information

### Request

```bash
curl -X GET "http://localhost:8080/api/v1/pods/{podId}/describe?namespace={namespace}" \
  -H "Authorization: Bearer <your-jwt-token>" \
  -H "Content-Type: application/json"
```

### Example

```bash
curl -X GET "http://localhost:8080/api/v1/pods/nginx-deployment-7d4b8c9f5-abc12/describe?namespace=default" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json"
```

### Response

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

## 5. WebSocket Terminal Connection

### JavaScript Client Example

```javascript
// WebSocket client for terminal connection
class PodTerminal {
  constructor(podId, namespace, container, command, token) {
    this.podId = podId;
    this.namespace = namespace;
    this.container = container;
    this.command = command;
    this.token = token;
    this.ws = null;
    this.terminal = null;
  }

  connect() {
    const wsUrl = `ws://localhost:8080/api/v1/pods/${this.podId}/terminal?namespace=${this.namespace}`;

    this.ws = new WebSocket(wsUrl, [], {
      headers: {
        Authorization: `Bearer ${this.token}`,
      },
    });

    this.ws.onopen = () => {
      console.log("WebSocket connected");

      // Send the exec request
      const execRequest = {
        container: this.container,
        command: this.command,
      };

      this.ws.send(JSON.stringify(execRequest));
    };

    this.ws.onmessage = (event) => {
      if (this.terminal) {
        this.terminal.write(event.data);
      } else {
        console.log("Terminal output:", event.data);
      }
    };

    this.ws.onclose = () => {
      console.log("WebSocket disconnected");
    };

    this.ws.onerror = (error) => {
      console.error("WebSocket error:", error);
    };
  }

  sendInput(input) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(input);
    }
  }

  resizeTerminal(cols, rows) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      // Send terminal resize event
      const resizeEvent = {
        type: "resize",
        cols: cols,
        rows: rows,
      };
      this.ws.send(JSON.stringify(resizeEvent));
    }
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
    }
  }
}

// Usage example
const terminal = new PodTerminal(
  "nginx-deployment-7d4b8c9f5-abc12",
  "default",
  "nginx",
  ["/bin/bash"],
  "your-jwt-token"
);

terminal.connect();

// Send input to terminal
terminal.sendInput("ls -la\n");

// Resize terminal
terminal.resizeTerminal(80, 24);

// Disconnect when done
terminal.disconnect();
```

### HTML Example with xterm.js

```html
<!DOCTYPE html>
<html>
  <head>
    <title>Pod Terminal</title>
    <link
      rel="stylesheet"
      href="https://cdn.jsdelivr.net/npm/xterm@5.3.0/css/xterm.css"
    />
    <script src="https://cdn.jsdelivr.net/npm/xterm@5.3.0/lib/xterm.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/xterm-addon-fit@0.8.0/lib/xterm-addon-fit.js"></script>
  </head>
  <body>
    <div id="terminal"></div>

    <script>
      // Initialize terminal
      const terminal = new Terminal({
        cursorBlink: true,
        fontSize: 14,
        fontFamily: 'Monaco, Menlo, "Ubuntu Mono", monospace',
      });

      const fitAddon = new FitAddon.FitAddon();
      terminal.loadAddon(fitAddon);

      terminal.open(document.getElementById("terminal"));
      fitAddon.fit();

      // WebSocket connection
      const podId = "nginx-deployment-7d4b8c9f5-abc12";
      const namespace = "default";
      const container = "nginx";
      const command = ["/bin/bash"];
      const token = "your-jwt-token";

      const wsUrl = `ws://localhost:8080/api/v1/pods/${podId}/terminal?namespace=${namespace}`;
      const ws = new WebSocket(wsUrl, [], {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      ws.onopen = () => {
        console.log("Connected to pod terminal");

        // Send exec request
        const execRequest = {
          container: container,
          command: command,
        };
        ws.send(JSON.stringify(execRequest));
      };

      ws.onmessage = (event) => {
        terminal.write(event.data);
      };

      ws.onclose = () => {
        terminal.write("\r\n\r\nTerminal disconnected.\r\n");
      };

      ws.onerror = (error) => {
        terminal.write(`\r\nError: ${error.message}\r\n`);
      };

      // Handle terminal input
      terminal.onData((data) => {
        ws.send(data);
      });

      // Handle terminal resize
      terminal.onResize((size) => {
        const resizeEvent = {
          type: "resize",
          cols: size.cols,
          rows: size.rows,
        };
        ws.send(JSON.stringify(resizeEvent));
      });

      // Handle window resize
      window.addEventListener("resize", () => {
        fitAddon.fit();
      });
    </script>
  </body>
</html>
```

## Error Responses

### 400 Bad Request

```json
{
  "error": "Invalid application ID format"
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
  "error": "Access denied"
}
```

### 404 Not Found

```json
{
  "error": "Pod not found"
}
```

### 500 Internal Server Error

```json
{
  "error": "Failed to retrieve pods from cluster"
}
```

## Notes

1. **Authentication**: All endpoints require a valid JWT token in the Authorization header.
2. **Namespace**: The namespace parameter is required for pod-specific operations.
3. **Container**: For logs and terminal operations, specify the container name if the pod has multiple containers.
4. **WebSocket**: The terminal endpoint upgrades the HTTP connection to WebSocket for real-time communication.
5. **Audit Logging**: All pod operations are logged for audit purposes.
6. **Authorization**: Users must be members of the organization that owns the application/cluster to access pods.
