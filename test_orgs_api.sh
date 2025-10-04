#!/bin/bash

# OneClick Organizations API Test Script
# This script demonstrates how to use the OneClick organizations API

BASE_URL="http://localhost:8080"

echo "=== OneClick Organizations API Test Script ==="
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
    echo "You can create a user by running:"
    echo "curl -X POST $BASE_URL/auth/register -H 'Content-Type: application/json' -d '{\"name\":\"John Doe\",\"email\":\"john@example.com\",\"password\":\"password123\"}'"
    exit 1
fi

# Create an organization
echo "2. Creating an organization..."
CREATE_ORG_RESPONSE=$(curl -s -X POST "$BASE_URL/orgs" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "name": "My Test Organization"
  }')

echo "Create organization response:"
echo "$CREATE_ORG_RESPONSE" | jq .

# Extract organization ID
ORG_ID=$(echo "$CREATE_ORG_RESPONSE" | jq -r '.id')
echo ""
echo "Organization ID: $ORG_ID"
echo ""

if [ "$ORG_ID" = "null" ] || [ -z "$ORG_ID" ]; then
    echo "Failed to create organization. Exiting."
    exit 1
fi

# Get user's organizations
echo "3. Getting user's organizations..."
curl -s -X GET "$BASE_URL/orgs" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
echo ""

# Get organization details
echo "4. Getting organization details..."
curl -s -X GET "$BASE_URL/orgs/$ORG_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
echo ""

# Create another user to add as member
echo "5. Creating another user to add as member..."
curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jane Doe",
    "email": "jane@example.com",
    "password": "password123"
  }' | jq .
echo ""

# Add member to organization
echo "6. Adding member to organization..."
ADD_MEMBER_RESPONSE=$(curl -s -X POST "$BASE_URL/orgs/$ORG_ID/members" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "email": "jane@example.com",
    "role": "member"
  }')

echo "Add member response:"
echo "$ADD_MEMBER_RESPONSE" | jq .
echo ""

# Get organization details again to see the new member
echo "7. Getting organization details with members..."
curl -s -X GET "$BASE_URL/orgs/$ORG_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
echo ""

# Update member role
echo "8. Updating member role to admin..."
MEMBER_ID=$(echo "$ADD_MEMBER_RESPONSE" | jq -r '.id')
if [ "$MEMBER_ID" != "null" ] && [ -n "$MEMBER_ID" ]; then
    curl -s -X PATCH "$BASE_URL/orgs/$ORG_ID/members/$MEMBER_ID" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $ACCESS_TOKEN" \
      -d '{
        "role": "admin"
      }' | jq .
    echo ""
fi

# Get organization details to see updated role
echo "9. Getting organization details with updated member role..."
curl -s -X GET "$BASE_URL/orgs/$ORG_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
echo ""

# Test error cases
echo "10. Testing error cases..."

echo "10a. Trying to add non-existent user..."
curl -s -X POST "$BASE_URL/orgs/$ORG_ID/members" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "email": "nonexistent@example.com",
    "role": "member"
  }' | jq .
echo ""

echo "10b. Trying to access organization without permission..."
# Create a third user who is not a member
curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Bob Smith",
    "email": "bob@example.com",
    "password": "password123"
  }' > /dev/null

# Login as the third user
BOB_LOGIN=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "bob@example.com",
    "password": "password123"
  }')

BOB_TOKEN=$(echo "$BOB_LOGIN" | jq -r '.access_token')

# Try to access the organization
curl -s -X GET "$BASE_URL/orgs/$ORG_ID" \
  -H "Authorization: Bearer $BOB_TOKEN" | jq .
echo ""

echo "10c. Trying to add member without permission..."
curl -s -X POST "$BASE_URL/orgs/$ORG_ID/members" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $BOB_TOKEN" \
  -d '{
    "email": "alice@example.com",
    "role": "member"
  }' | jq .
echo ""

# Clean up - remove member
echo "11. Removing member from organization..."
if [ "$MEMBER_ID" != "null" ] && [ -n "$MEMBER_ID" ]; then
    curl -s -X DELETE "$BASE_URL/orgs/$ORG_ID/members/$MEMBER_ID" \
      -H "Authorization: Bearer $ACCESS_TOKEN"
    echo ""
fi

# Get organization details to confirm member removal
echo "12. Getting organization details after member removal..."
curl -s -X GET "$BASE_URL/orgs/$ORG_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
echo ""

# Delete organization
echo "13. Deleting organization..."
curl -s -X DELETE "$BASE_URL/orgs/$ORG_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
echo ""

# Verify organization is deleted
echo "14. Verifying organization is deleted..."
curl -s -X GET "$BASE_URL/orgs/$ORG_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
echo ""

echo "=== Organizations API Test completed ==="
echo ""
echo "Summary of tested endpoints:"
echo "- POST /orgs (create organization)"
echo "- GET /orgs (list user's organizations)"
echo "- GET /orgs/{orgId} (get organization details)"
echo "- POST /orgs/{orgId}/members (add member)"
echo "- PATCH /orgs/{orgId}/members/{userId} (update member role)"
echo "- DELETE /orgs/{orgId}/members/{userId} (remove member)"
echo "- DELETE /orgs/{orgId} (delete organization)"
echo ""
echo "Tested error cases:"
echo "- Adding non-existent user"
echo "- Accessing organization without permission"
echo "- Adding members without permission"
