#!/bin/bash

# OneClick Repositories API Test Script
# This script demonstrates how to use the OneClick repositories API

BASE_URL="http://localhost:8080"

echo "=== OneClick Repositories API Test Script ==="
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

# Create a GitHub repository
echo "3. Creating a GitHub repository..."
CREATE_GITHUB_REPO_RESPONSE=$(curl -s -X POST "$BASE_URL/orgs/$ORG_ID/repos" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "type": "github",
    "url": "https://github.com/user/example-repo.git",
    "default_branch": "main"
  }')

echo "Create GitHub repository response:"
echo "$CREATE_GITHUB_REPO_RESPONSE" | jq .

# Extract repository ID
REPO_ID=$(echo "$CREATE_GITHUB_REPO_RESPONSE" | jq -r '.id')
echo ""
echo "Repository ID: $REPO_ID"
echo ""

if [ "$REPO_ID" = "null" ] || [ -z "$REPO_ID" ]; then
    echo "Failed to create repository. Exiting."
    exit 1
fi

# Create a GitLab repository with token
echo "4. Creating a GitLab repository with token..."
CREATE_GITLAB_REPO_RESPONSE=$(curl -s -X POST "$BASE_URL/orgs/$ORG_ID/repos" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "type": "gitlab",
    "url": "https://gitlab.com/user/example-repo.git",
    "default_branch": "main",
    "token": "glpat-xxxxxxxxxxxxxxxxxxxx"
  }')

echo "Create GitLab repository response:"
echo "$CREATE_GITLAB_REPO_RESPONSE" | jq .

# Create a Gitea repository
echo "5. Creating a Gitea repository..."
CREATE_GITEA_REPO_RESPONSE=$(curl -s -X POST "$BASE_URL/orgs/$ORG_ID/repos" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "type": "gitea",
    "url": "https://gitea.example.com/user/example-repo.git",
    "default_branch": "main"
  }')

echo "Create Gitea repository response:"
echo "$CREATE_GITEA_REPO_RESPONSE" | jq .

# Get repositories in organization
echo "6. Getting repositories in organization..."
curl -s -X GET "$BASE_URL/orgs/$ORG_ID/repos" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
echo ""

# Get repository details
echo "7. Getting repository details..."
curl -s -X GET "$BASE_URL/repos/$REPO_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
echo ""

# Test webhook endpoint
echo "8. Testing webhook endpoint..."
curl -s -X GET "$BASE_URL/hooks/test?provider=github" | jq .
echo ""

# Test webhook with GitHub payload
echo "9. Testing webhook with GitHub payload..."
curl -s -X POST "$BASE_URL/hooks/git?provider=github" \
  -H "Content-Type: application/json" \
  -H "X-Hub-Signature-256: sha256=test-signature" \
  -d '{
    "ref": "refs/heads/main",
    "repository": {
      "full_name": "user/example-repo",
      "clone_url": "https://github.com/user/example-repo.git",
      "ssh_url": "git@github.com:user/example-repo.git"
    },
    "commits": [
      {
        "id": "abc123",
        "message": "Test commit",
        "author": {
          "name": "Test User",
          "email": "test@example.com"
        }
      }
    ],
    "head_commit": {
      "id": "abc123",
      "message": "Test commit",
      "author": {
        "name": "Test User",
        "email": "test@example.com"
      }
    }
  }' | jq .
echo ""

# Test webhook with GitLab payload
echo "10. Testing webhook with GitLab payload..."
curl -s -X POST "$BASE_URL/hooks/git?provider=gitlab" \
  -H "Content-Type: application/json" \
  -H "X-Gitlab-Signature: sha256=test-signature" \
  -d '{
    "ref": "refs/heads/main",
    "repository": {
      "name": "example-repo",
      "url": "https://gitlab.com/user/example-repo.git",
      "homepage": "https://gitlab.com/user/example-repo"
    },
    "commits": [
      {
        "id": "def456",
        "message": "Test commit",
        "author": {
          "name": "Test User",
          "email": "test@example.com"
        }
      }
    ]
  }' | jq .
echo ""

# Test webhook with Gitea payload
echo "11. Testing webhook with Gitea payload..."
curl -s -X POST "$BASE_URL/hooks/git?provider=gitea" \
  -H "Content-Type: application/json" \
  -H "X-Gitea-Signature: sha256=test-signature" \
  -d '{
    "ref": "refs/heads/main",
    "repository": {
      "full_name": "user/example-repo",
      "clone_url": "https://gitea.example.com/user/example-repo.git",
      "ssh_url": "git@gitea.example.com:user/example-repo.git"
    },
    "commits": [
      {
        "id": "ghi789",
        "message": "Test commit",
        "author": {
          "name": "Test User",
          "email": "test@example.com"
        }
      }
    ]
  }' | jq .
echo ""

# Test error cases
echo "12. Testing error cases..."

echo "12a. Trying to create repository with invalid type..."
curl -s -X POST "$BASE_URL/orgs/$ORG_ID/repos" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "type": "invalid",
    "url": "https://github.com/user/repo.git",
    "default_branch": "main"
  }' | jq .
echo ""

echo "12b. Trying to create repository with invalid URL..."
curl -s -X POST "$BASE_URL/orgs/$ORG_ID/repos" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "type": "github",
    "url": "invalid-url",
    "default_branch": "main"
  }' | jq .
echo ""

echo "12c. Trying to access repository without permission..."
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

# Try to access the repository
curl -s -X GET "$BASE_URL/repos/$REPO_ID" \
  -H "Authorization: Bearer $ALICE_TOKEN" | jq .
echo ""

echo "12d. Trying to delete repository without permission..."
curl -s -X DELETE "$BASE_URL/repos/$REPO_ID" \
  -H "Authorization: Bearer $ALICE_TOKEN" | jq .
echo ""

# Clean up - delete repositories
echo "13. Cleaning up - deleting repositories..."

echo "Deleting created repositories..."
curl -s -X DELETE "$BASE_URL/repos/$REPO_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
echo ""

echo "=== Repositories API Test completed ==="
echo ""
echo "Summary of tested endpoints:"
echo "- POST /orgs/{orgId}/repos (create repository)"
echo "- GET /orgs/{orgId}/repos (list organization repositories)"
echo "- GET /repos/{repoId} (get repository details)"
echo "- DELETE /repos/{repoId} (delete repository)"
echo "- POST /hooks/git (webhook endpoint)"
echo "- GET /hooks/test (test webhook endpoint)"
echo ""
echo "Tested providers:"
echo "- GitHub (github)"
echo "- GitLab (gitlab)"
echo "- Gitea (gitea)"
echo ""
echo "Tested error cases:"
echo "- Invalid repository type"
echo "- Invalid repository URL"
echo "- Access without permission"
echo "- Delete without permission"
