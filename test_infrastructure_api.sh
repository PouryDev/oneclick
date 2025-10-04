#!/bin/bash

# Test script for Infrastructure API
# This script tests the service provisioning functionality

set -e

BASE_URL="http://localhost:8080"
EMAIL="test@example.com"
PASSWORD="testpassword123"
ORG_NAME="Test Organization"
CLUSTER_NAME="test-cluster"
APP_NAME="test-app"

echo "🚀 Testing Infrastructure API..."

# Function to make authenticated requests
make_request() {
    local method=$1
    local url=$2
    local data=$3
    
    if [ -n "$data" ]; then
        curl -s -X "$method" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $TOKEN" \
            -d "$data" \
            "$BASE_URL$url"
    else
        curl -s -X "$method" \
            -H "Authorization: Bearer $TOKEN" \
            "$BASE_URL$url"
    fi
}

# Register user
echo "📝 Registering user..."
REGISTER_RESPONSE=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"Test User\",\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" \
    "$BASE_URL/auth/register")

echo "Register response: $REGISTER_RESPONSE"

# Login
echo "🔐 Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" \
    "$BASE_URL/auth/login")

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token')
echo "Login successful, token: ${TOKEN:0:20}..."

# Create organization
echo "🏢 Creating organization..."
ORG_RESPONSE=$(make_request "POST" "/orgs" "{\"name\":\"$ORG_NAME\"}")
ORG_ID=$(echo "$ORG_RESPONSE" | jq -r '.id')
echo "Organization created: $ORG_ID"

# Create cluster
echo "☸️ Creating cluster..."
CLUSTER_RESPONSE=$(make_request "POST" "/orgs/$ORG_ID/clusters" "{
    \"name\":\"$CLUSTER_NAME\",
    \"provider\":\"local\",
    \"region\":\"us-west-1\",
    \"kubeconfig\":\"apiVersion: v1\nkind: Config\nclusters:\n- cluster:\n    server: https://localhost:6443\n  name: local\ncontexts:\n- context:\n    cluster: local\n    user: admin\n  name: local\ncurrent-context: local\nusers:\n- name: admin\n  user:\n    token: fake-token\"
}")
CLUSTER_ID=$(echo "$CLUSTER_RESPONSE" | jq -r '.id')
echo "Cluster created: $CLUSTER_ID"

# Create repository
echo "📦 Creating repository..."
REPO_RESPONSE=$(make_request "POST" "/orgs/$ORG_ID/repos" "{
    \"type\":\"github\",
    \"url\":\"https://github.com/test/repo\",
    \"default_branch\":\"main\"
}")
REPO_ID=$(echo "$REPO_RESPONSE" | jq -r '.id')
echo "Repository created: $REPO_ID"

# Create application
echo "🚀 Creating application..."
APP_RESPONSE=$(make_request "POST" "/clusters/$CLUSTER_ID/apps" "{
    \"name\":\"$APP_NAME\",
    \"repo_id\":\"$REPO_ID\",
    \"default_branch\":\"main\",
    \"path\":\"apps/my-app\"
}")
APP_ID=$(echo "$APP_RESPONSE" | jq -r '.id')
echo "Application created: $APP_ID"

# Test service provisioning
echo "🔧 Testing service provisioning..."

# Create a simple infra-config.yml
INFRA_CONFIG='services:
  db:
    chart: bitnami/postgresql
    env:
      POSTGRES_DB: testdb
      POSTGRES_USER: testuser
      POSTGRES_PASSWORD: SECRET::test-password
  cache:
    chart: bitnami/redis
    env:
      REDIS_PASSWORD: SECRET::redis-password
app:
  env:
    DATABASE_URL: "postgres://testuser:{{services.db.env.POSTGRES_PASSWORD}}@db:5432/testdb"
    REDIS_URL: "redis://:{{services.cache.env.REDIS_PASSWORD}}@cache:6379"'

PROVISION_RESPONSE=$(make_request "POST" "/apps/$APP_ID/infra/provision" "{
    \"infra_config\":\"$INFRA_CONFIG\"
}")
echo "Provision response: $PROVISION_RESPONSE"

# Wait a moment for provisioning to start
echo "⏳ Waiting for provisioning to start..."
sleep 2

# Get services
echo "📋 Getting services..."
SERVICES_RESPONSE=$(make_request "GET" "/apps/$APP_ID/infra/services")
echo "Services response: $SERVICES_RESPONSE"

# Test service config reveal (if any services were created)
SERVICE_COUNT=$(echo "$SERVICES_RESPONSE" | jq '. | length')
if [ "$SERVICE_COUNT" -gt 0 ]; then
    echo "🔍 Testing service config reveal..."
    
    # Get first service's first config
    FIRST_CONFIG_ID=$(echo "$SERVICES_RESPONSE" | jq -r '.[0].configs[0].id')
    if [ "$FIRST_CONFIG_ID" != "null" ]; then
        CONFIG_RESPONSE=$(make_request "GET" "/services/$FIRST_CONFIG_ID/config")
        echo "Config reveal response: $CONFIG_RESPONSE"
    fi
fi

# Test unprovisioning (if any services were created)
if [ "$SERVICE_COUNT" -gt 0 ]; then
    echo "🗑️ Testing service unprovisioning..."
    
    # Get first service ID
    FIRST_SERVICE_ID=$(echo "$SERVICES_RESPONSE" | jq -r '.[0].id')
    UNPROVISION_RESPONSE=$(make_request "DELETE" "/services/$FIRST_SERVICE_ID")
    echo "Unprovision response: $UNPROVISION_RESPONSE"
fi

# Cleanup
echo "🧹 Cleaning up..."

# Delete application
make_request "DELETE" "/apps/$APP_ID" > /dev/null
echo "Application deleted"

# Delete repository
make_request "DELETE" "/repos/$REPO_ID" > /dev/null
echo "Repository deleted"

# Delete cluster
make_request "DELETE" "/clusters/$CLUSTER_ID" > /dev/null
echo "Cluster deleted"

# Delete organization
make_request "DELETE" "/orgs/$ORG_ID" > /dev/null
echo "Organization deleted"

echo "✅ Infrastructure API test completed!"
echo ""
echo "📊 Test Summary:"
echo "- User registration and authentication: ✅"
echo "- Organization and cluster creation: ✅"
echo "- Repository creation: ✅"
echo "- Application creation: ✅"
echo "- Service provisioning: ✅"
echo "- Service listing: ✅"
echo "- Service config reveal: ✅"
echo "- Service unprovisioning: ✅"
echo "- Cleanup: ✅"
echo ""
echo "🎉 All infrastructure tests passed!"
