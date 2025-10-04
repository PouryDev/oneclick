#!/bin/bash

# OneClick Applications API Test Script
# This script demonstrates how to use the OneClick applications API

BASE_URL="http://localhost:8080"

echo "=== OneClick Applications API Test Script ==="
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

# Get clusters in the organization
echo "3. Getting clusters in organization..."
CLUSTERS_RESPONSE=$(curl -s -X GET "$BASE_URL/orgs/$ORG_ID/clusters" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "Clusters response:"
echo "$CLUSTERS_RESPONSE" | jq .

# Extract cluster ID
CLUSTER_ID=$(echo "$CLUSTERS_RESPONSE" | jq -r '.[0].id')
echo ""
echo "Cluster ID: $CLUSTER_ID"
echo ""

if [ "$CLUSTER_ID" = "null" ] || [ -z "$CLUSTER_ID" ]; then
    echo "No clusters found. Please create a cluster first."
    exit 1
fi

# Get repositories in the organization
echo "4. Getting repositories in organization..."
REPOS_RESPONSE=$(curl -s -X GET "$BASE_URL/orgs/$ORG_ID/repos" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "Repositories response:"
echo "$REPOS_RESPONSE" | jq .

# Extract repository ID
REPO_ID=$(echo "$REPOS_RESPONSE" | jq -r '.[0].id')
echo ""
echo "Repository ID: $REPO_ID"
echo ""

if [ "$REPO_ID" = "null" ] || [ -z "$REPO_ID" ]; then
    echo "No repositories found. Please create a repository first."
    exit 1
fi

# Create an application
echo "5. Creating an application..."
CREATE_APP_RESPONSE=$(curl -s -X POST "$BASE_URL/clusters/$CLUSTER_ID/apps" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"name\": \"test-app\",
    \"repo_id\": \"$REPO_ID\",
    \"default_branch\": \"main\",
    \"path\": \"apps/test-app\"
  }")

echo "Create application response:"
echo "$CREATE_APP_RESPONSE" | jq .

# Extract application ID
APP_ID=$(echo "$CREATE_APP_RESPONSE" | jq -r '.id')
echo ""
echo "Application ID: $APP_ID"
echo ""

if [ "$APP_ID" = "null" ] || [ -z "$APP_ID" ]; then
    echo "Failed to create application. Exiting."
    exit 1
fi

# Create another application
echo "6. Creating another application..."
CREATE_APP2_RESPONSE=$(curl -s -X POST "$BASE_URL/clusters/$CLUSTER_ID/apps" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"name\": \"api-service\",
    \"repo_id\": \"$REPO_ID\",
    \"default_branch\": \"main\"
  }")

echo "Create second application response:"
echo "$CREATE_APP2_RESPONSE" | jq .

# Get applications in cluster
echo "7. Getting applications in cluster..."
curl -s -X GET "$BASE_URL/clusters/$CLUSTER_ID/apps" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
echo ""

# Get application details
echo "8. Getting application details..."
curl -s -X GET "$BASE_URL/apps/$APP_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
echo ""

# Deploy application
echo "9. Deploying application..."
DEPLOY_RESPONSE=$(curl -s -X POST "$BASE_URL/apps/$APP_ID/deploy" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "image": "myapp:latest",
    "tag": "latest"
  }')

echo "Deploy application response:"
echo "$DEPLOY_RESPONSE" | jq .

# Extract release ID
RELEASE_ID=$(echo "$DEPLOY_RESPONSE" | jq -r '.release_id')
echo ""
echo "Release ID: $RELEASE_ID"
echo ""

# Wait a moment for deployment to process
echo "10. Waiting for deployment to process..."
sleep 3

# Get releases for application
echo "11. Getting releases for application..."
curl -s -X GET "$BASE_URL/apps/$APP_ID/releases" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
echo ""

# Deploy another version
echo "12. Deploying another version..."
DEPLOY_RESPONSE2=$(curl -s -X POST "$BASE_URL/apps/$APP_ID/deploy" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "image": "myapp:v2.0.0",
    "tag": "v2.0.0"
  }')

echo "Deploy second version response:"
echo "$DEPLOY_RESPONSE2" | jq .

# Extract second release ID
RELEASE_ID2=$(echo "$DEPLOY_RESPONSE2" | jq -r '.release_id')
echo ""
echo "Second Release ID: $RELEASE_ID2"
echo ""

# Wait for second deployment
echo "13. Waiting for second deployment to process..."
sleep 3

# Get releases again to see both
echo "14. Getting all releases for application..."
curl -s -X GET "$BASE_URL/apps/$APP_ID/releases" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
echo ""

# Rollback to previous release
echo "15. Rolling back to previous release..."
ROLLBACK_RESPONSE=$(curl -s -X POST "$BASE_URL/apps/$APP_ID/releases/$RELEASE_ID/rollback" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "Rollback response:"
echo "$ROLLBACK_RESPONSE" | jq .
echo ""

# Wait for rollback
echo "16. Waiting for rollback to process..."
sleep 3

# Get releases after rollback
echo "17. Getting releases after rollback..."
curl -s -X GET "$BASE_URL/apps/$APP_ID/releases" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
echo ""

# Test error cases
echo "18. Testing error cases..."

echo "18a. Trying to create application with invalid cluster ID..."
curl -s -X POST "$BASE_URL/clusters/invalid-cluster-id/apps" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"name\": \"test-app\",
    \"repo_id\": \"$REPO_ID\",
    \"default_branch\": \"main\"
  }" | jq .
echo ""

echo "18b. Trying to create application with duplicate name..."
curl -s -X POST "$BASE_URL/clusters/$CLUSTER_ID/apps" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"name\": \"test-app\",
    \"repo_id\": \"$REPO_ID\",
    \"default_branch\": \"main\"
  }" | jq .
echo ""

echo "18c. Trying to deploy without image..."
curl -s -X POST "$BASE_URL/apps/$APP_ID/deploy" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "tag": "latest"
  }" | jq .
echo ""

echo "18d. Trying to deploy without tag..."
curl -s -X POST "$BASE_URL/apps/$APP_ID/deploy" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "image": "myapp:latest"
  }" | jq .
echo ""

echo "18e. Trying to access application without permission..."
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

# Try to access the application
curl -s -X GET "$BASE_URL/apps/$APP_ID" \
  -H "Authorization: Bearer $ALICE_TOKEN" | jq .
echo ""

echo "18f. Trying to delete application without permission..."
curl -s -X DELETE "$BASE_URL/apps/$APP_ID" \
  -H "Authorization: Bearer $ALICE_TOKEN" | jq .
echo ""

# Clean up - delete applications
echo "19. Cleaning up - deleting applications..."

echo "Deleting created applications..."
curl -s -X DELETE "$BASE_URL/apps/$APP_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
echo ""

# Get second app ID if it was created
APP_ID2=$(echo "$CREATE_APP2_RESPONSE" | jq -r '.id')
if [ "$APP_ID2" != "null" ] && [ -n "$APP_ID2" ]; then
    curl -s -X DELETE "$BASE_URL/apps/$APP_ID2" \
      -H "Authorization: Bearer $ACCESS_TOKEN"
    echo ""
fi

echo "=== Applications API Test completed ==="
echo ""
echo "Summary of tested endpoints:"
echo "- POST /clusters/{clusterId}/apps (create application)"
echo "- GET /clusters/{clusterId}/apps (list cluster applications)"
echo "- GET /apps/{appId} (get application details)"
echo "- POST /apps/{appId}/deploy (deploy application)"
echo "- GET /apps/{appId}/releases (list application releases)"
echo "- POST /apps/{appId}/releases/{releaseId}/rollback (rollback application)"
echo "- DELETE /apps/{appId} (delete application)"
echo ""
echo "Tested features:"
echo "- Application creation with repository integration"
echo "- Application deployment with image and tag"
echo "- Release management and history"
echo "- Application rollback functionality"
echo "- Permission-based access control"
echo ""
echo "Tested error cases:"
echo "- Invalid cluster ID"
echo "- Duplicate application names"
echo "- Missing deployment parameters"
echo "- Access without permission"
echo "- Delete without permission"
