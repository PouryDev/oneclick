#!/bin/bash

# OneClick Clusters API Test Script
# This script demonstrates how to use the OneClick clusters API

BASE_URL="http://localhost:8080"

echo "=== OneClick Clusters API Test Script ==="
echo "Base URL: $BASE_URL"
echo ""

# First, we need to authenticate and get a token
echo "1. Authenticating user..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }')

echo "Login response:"
echo "$LOGIN_RESPONSE" | jq .

# Extract the access token
ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token')
echo ""
echo "Access token: $ACCESS_TOKEN"
echo ""

if [ "$ACCESS_TOKEN" = "null" ] || [ -z "$ACCESS_TOKEN" ]; then
    echo "Failed to get access token. Please make sure the user exists and the server is running."
    exit 1
fi

# Get user's organizations to find an org ID
echo "2. Getting user's organizations..."
ORGS_RESPONSE=$(curl -s -X GET "$BASE_URL/orgs" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "Organizations response:"
echo "$ORGS_RESPONSE" | jq .

# Extract organization ID
ORG_ID=$(echo "$ORGS_RESPONSE" | jq -r '.[0].id')
echo ""
echo "Organization ID: $ORG_ID"
echo ""

if [ "$ORG_ID" = "null" ] || [ -z "$ORG_ID" ]; then
    echo "No organizations found. Please create an organization first."
    exit 1
fi

# Create a cluster
echo "3. Creating a cluster..."
CREATE_CLUSTER_RESPONSE=$(curl -s -X POST "$BASE_URL/orgs/$ORG_ID/clusters" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "name": "My Test Cluster",
    "provider": "aws",
    "region": "us-west-2"
  }')

echo "Create cluster response:"
echo "$CREATE_CLUSTER_RESPONSE" | jq .

# Extract cluster ID
CLUSTER_ID=$(echo "$CREATE_CLUSTER_RESPONSE" | jq -r '.id')
echo ""
echo "Cluster ID: $CLUSTER_ID"
echo ""

if [ "$CLUSTER_ID" = "null" ] || [ -z "$CLUSTER_ID" ]; then
    echo "Failed to create cluster. Exiting."
    exit 1
fi

# Get clusters in organization
echo "4. Getting clusters in organization..."
curl -s -X GET "$BASE_URL/orgs/$ORG_ID/clusters" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
echo ""

# Get cluster details
echo "5. Getting cluster details..."
curl -s -X GET "$BASE_URL/clusters/$CLUSTER_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
echo ""

# Create a sample kubeconfig file for import testing
echo "6. Creating sample kubeconfig for import testing..."
cat > /tmp/sample-kubeconfig.yaml << 'EOF'
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://kubernetes.default.svc
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
EOF

# Import cluster with kubeconfig
echo "7. Importing cluster with kubeconfig..."
IMPORT_RESPONSE=$(curl -s -X POST "$BASE_URL/orgs/$ORG_ID/clusters/import" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -F "name=Imported Cluster" \
  -F "kubeconfig=@/tmp/sample-kubeconfig.yaml")

echo "Import cluster response:"
echo "$IMPORT_RESPONSE" | jq .

# Extract imported cluster ID
IMPORTED_CLUSTER_ID=$(echo "$IMPORT_RESPONSE" | jq -r '.id')
echo ""
echo "Imported Cluster ID: $IMPORTED_CLUSTER_ID"
echo ""

# Get cluster health (this will fail with our fake kubeconfig, but shows the API)
echo "8. Getting cluster health status..."
curl -s -X GET "$BASE_URL/clusters/$IMPORTED_CLUSTER_ID/status" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
echo ""

# Test error cases
echo "9. Testing error cases..."

echo "9a. Trying to create cluster with invalid data..."
curl -s -X POST "$BASE_URL/orgs/$ORG_ID/clusters" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "name": "",
    "provider": "aws",
    "region": "us-west-2"
  }' | jq .
echo ""

echo "9b. Trying to access cluster without permission..."
# Create another user who is not a member
curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice Smith",
    "email": "alice@example.com",
    "password": "password123"
  }' > /dev/null

# Login as the other user
ALICE_LOGIN=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "password": "password123"
  }')

ALICE_TOKEN=$(echo "$ALICE_LOGIN" | jq -r '.access_token')

# Try to access the cluster
curl -s -X GET "$BASE_URL/clusters/$CLUSTER_ID" \
  -H "Authorization: Bearer $ALICE_TOKEN" | jq .
echo ""

echo "9c. Trying to delete cluster without permission..."
curl -s -X DELETE "$BASE_URL/clusters/$CLUSTER_ID" \
  -H "Authorization: Bearer $ALICE_TOKEN" | jq .
echo ""

# Clean up - delete clusters
echo "10. Cleaning up - deleting clusters..."

if [ "$IMPORTED_CLUSTER_ID" != "null" ] && [ -n "$IMPORTED_CLUSTER_ID" ]; then
    echo "Deleting imported cluster..."
    curl -s -X DELETE "$BASE_URL/clusters/$IMPORTED_CLUSTER_ID" \
      -H "Authorization: Bearer $ACCESS_TOKEN"
    echo ""
fi

echo "Deleting created cluster..."
curl -s -X DELETE "$BASE_URL/clusters/$CLUSTER_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
echo ""

# Clean up temp file
rm -f /tmp/sample-kubeconfig.yaml

echo "=== Clusters API Test completed ==="
echo ""
echo "Summary of tested endpoints:"
echo "- POST /orgs/{orgId}/clusters (create cluster)"
echo "- POST /orgs/{orgId}/clusters/import (import cluster)"
echo "- GET /orgs/{orgId}/clusters (list organization clusters)"
echo "- GET /clusters/{clusterId} (get cluster details)"
echo "- GET /clusters/{clusterId}/status (get cluster health)"
echo "- DELETE /clusters/{clusterId} (delete cluster)"
echo ""
echo "Tested error cases:"
echo "- Invalid cluster data"
echo "- Access without permission"
echo "- Delete without permission"
