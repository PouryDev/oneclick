#!/bin/bash

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | xargs)
else
    echo "Error: .env file not found. Please create one with DATABASE_URL, JWT_SECRET, ONECLICK_MASTER_KEY, PORT."
    exit 1
fi

# Ensure the server is running (optional, for local testing)
# You might want to start your server in the background before running this script
# Example: go run cmd/server/main.go &
# SERVER_PID=$!
# trap "kill $SERVER_PID" EXIT # Kill server on script exit

PORT=${PORT:-8080}
BASE_URL="http://localhost:$PORT"

echo "Starting API tests for Git Server and Runner features..."
echo "Base URL: $BASE_URL"

# --- Helper Functions ---
function assert_status() {
    local expected=$1
    local actual=$2
    local message=$3
    if [ "$expected" -ne "$actual" ]; then
        echo "FAIL: $message (Expected: $expected, Got: $actual)"
        exit 1
    else
        echo "PASS: $message (Status: $actual)"
    fi
}

function extract_json_value() {
    local json_string=$1
    local key=$2
    echo "$json_string" | jq -r ".$key"
}

# --- 1. Register a new user ---
echo -e "\n--- Registering User ---"
REGISTER_PAYLOAD='{"name":"gituser","email":"git@example.com","password":"password123"}'
REGISTER_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" -d "$REGISTER_PAYLOAD" "$BASE_URL/auth/register")
REGISTER_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST -H "Content-Type: application/json" -d "$REGISTER_PAYLOAD" "$BASE_URL/auth/register")
assert_status 201 "$REGISTER_STATUS" "User registration"

# --- 2. Login the user ---
echo -e "\n--- Logging in User ---"
LOGIN_PAYLOAD='{"email":"git@example.com","password":"password123"}'
LOGIN_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" -d "$LOGIN_PAYLOAD" "$BASE_URL/auth/login")
LOGIN_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST -H "Content-Type: application/json" -d "$LOGIN_PAYLOAD" "$BASE_URL/auth/login")
assert_status 200 "$LOGIN_STATUS" "User login"

JWT_TOKEN=$(extract_json_value "$LOGIN_RESPONSE" "token")
USER_ID=$(extract_json_value "$LOGIN_RESPONSE" "user.id")
echo "Logged in as User ID: $USER_ID"

if [ -z "$JWT_TOKEN" ]; then
    echo "FAIL: JWT_TOKEN not received."
    exit 1
fi

# --- 3. Create an Organization ---
echo -e "\n--- Creating Organization ---"
ORG_PAYLOAD='{"name":"GitOrg"}'
ORG_RESPONSE=$(curl -s -X POST -H "Authorization: Bearer $JWT_TOKEN" -H "Content-Type: application/json" -d "$ORG_PAYLOAD" "$BASE_URL/orgs")
ORG_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST -H "Authorization: Bearer $JWT_TOKEN" -H "Content-Type: application/json" -d "$ORG_PAYLOAD" "$BASE_URL/orgs")
assert_status 201 "$ORG_STATUS" "Organization creation"

ORG_ID=$(extract_json_value "$ORG_RESPONSE" "id")
echo "Created Organization ID: $ORG_ID"

# --- 4. Create a Git Server ---
echo -e "\n--- Creating Git Server ---"
GIT_SERVER_PAYLOAD='{"type":"gitea","domain":"gitea.example.com","storage":"10Gi"}'
GIT_SERVER_RESPONSE=$(curl -s -X POST -H "Authorization: Bearer $JWT_TOKEN" -H "Content-Type: application/json" -d "$GIT_SERVER_PAYLOAD" "$BASE_URL/orgs/$ORG_ID/gitservers")
GIT_SERVER_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST -H "Authorization: Bearer $JWT_TOKEN" -H "Content-Type: application/json" -d "$GIT_SERVER_PAYLOAD" "$BASE_URL/orgs/$ORG_ID/gitservers")
assert_status 201 "$GIT_SERVER_STATUS" "Git server creation"

GIT_SERVER_ID=$(extract_json_value "$GIT_SERVER_RESPONSE" "id")
echo "Created Git Server ID: $GIT_SERVER_ID"

# --- 5. Get Git Servers for Organization ---
echo -e "\n--- Getting Git Servers for Organization ---"
GET_GIT_SERVERS_RESPONSE=$(curl -s -X GET -H "Authorization: Bearer $JWT_TOKEN" "$BASE_URL/orgs/$ORG_ID/gitservers")
GET_GIT_SERVERS_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X GET -H "Authorization: Bearer $JWT_TOKEN" "$BASE_URL/orgs/$ORG_ID/gitservers")
assert_status 200 "$GET_GIT_SERVERS_STATUS" "Get git servers for organization"

echo "Git Servers: $GET_GIT_SERVERS_RESPONSE"

# --- 6. Get Specific Git Server ---
echo -e "\n--- Getting Specific Git Server ---"
GET_GIT_SERVER_RESPONSE=$(curl -s -X GET -H "Authorization: Bearer $JWT_TOKEN" "$BASE_URL/gitservers/$GIT_SERVER_ID")
GET_GIT_SERVER_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X GET -H "Authorization: Bearer $JWT_TOKEN" "$BASE_URL/gitservers/$GIT_SERVER_ID")
assert_status 200 "$GET_GIT_SERVER_STATUS" "Get specific git server"

echo "Git Server Details: $GET_GIT_SERVER_RESPONSE"

# --- 7. Create a Runner ---
echo -e "\n--- Creating Runner ---"
RUNNER_PAYLOAD='{"name":"github-runner","type":"github","labels":["ubuntu","docker"],"nodeSelector":{"kubernetes.io/os":"linux"},"resources":{"cpu":"500m","memory":"1Gi"},"token":"ghp_example_token","url":"https://github.com"}'
RUNNER_RESPONSE=$(curl -s -X POST -H "Authorization: Bearer $JWT_TOKEN" -H "Content-Type: application/json" -d "$RUNNER_PAYLOAD" "$BASE_URL/orgs/$ORG_ID/runners")
RUNNER_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST -H "Authorization: Bearer $JWT_TOKEN" -H "Content-Type: application/json" -d "$RUNNER_PAYLOAD" "$BASE_URL/orgs/$ORG_ID/runners")
assert_status 201 "$RUNNER_STATUS" "Runner creation"

RUNNER_ID=$(extract_json_value "$RUNNER_RESPONSE" "id")
echo "Created Runner ID: $RUNNER_ID"

# --- 8. Get Runners for Organization ---
echo -e "\n--- Getting Runners for Organization ---"
GET_RUNNERS_RESPONSE=$(curl -s -X GET -H "Authorization: Bearer $JWT_TOKEN" "$BASE_URL/orgs/$ORG_ID/runners")
GET_RUNNERS_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X GET -H "Authorization: Bearer $JWT_TOKEN" "$BASE_URL/orgs/$ORG_ID/runners")
assert_status 200 "$GET_RUNNERS_STATUS" "Get runners for organization"

echo "Runners: $GET_RUNNERS_RESPONSE"

# --- 9. Get Specific Runner ---
echo -e "\n--- Getting Specific Runner ---"
GET_RUNNER_RESPONSE=$(curl -s -X GET -H "Authorization: Bearer $JWT_TOKEN" "$BASE_URL/runners/$RUNNER_ID")
GET_RUNNER_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X GET -H "Authorization: Bearer $JWT_TOKEN" "$BASE_URL/runners/$RUNNER_ID")
assert_status 200 "$GET_RUNNER_STATUS" "Get specific runner"

echo "Runner Details: $GET_RUNNER_RESPONSE"

# --- 10. Get Jobs for Organization ---
echo -e "\n--- Getting Jobs for Organization ---"
GET_JOBS_RESPONSE=$(curl -s -X GET -H "Authorization: Bearer $JWT_TOKEN" "$BASE_URL/orgs/$ORG_ID/jobs")
GET_JOBS_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X GET -H "Authorization: Bearer $JWT_TOKEN" "$BASE_URL/orgs/$ORG_ID/jobs")
assert_status 200 "$GET_JOBS_STATUS" "Get jobs for organization"

echo "Jobs: $GET_JOBS_RESPONSE"

# --- 11. Test Error Cases ---
echo -e "\n--- Testing Error Cases ---"

# Test creating git server with invalid domain
echo "Testing invalid git server creation..."
INVALID_GIT_SERVER_PAYLOAD='{"type":"gitea","domain":"","storage":"10Gi"}'
INVALID_GIT_SERVER_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST -H "Authorization: Bearer $JWT_TOKEN" -H "Content-Type: application/json" -d "$INVALID_GIT_SERVER_PAYLOAD" "$BASE_URL/orgs/$ORG_ID/gitservers")
assert_status 400 "$INVALID_GIT_SERVER_STATUS" "Invalid git server creation (empty domain)"

# Test creating runner with invalid name
echo "Testing invalid runner creation..."
INVALID_RUNNER_PAYLOAD='{"name":"","type":"github","labels":["ubuntu"]}'
INVALID_RUNNER_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST -H "Authorization: Bearer $JWT_TOKEN" -H "Content-Type: application/json" -d "$INVALID_RUNNER_PAYLOAD" "$BASE_URL/orgs/$ORG_ID/runners")
assert_status 400 "$INVALID_RUNNER_STATUS" "Invalid runner creation (empty name)"

# Test accessing non-existent git server
echo "Testing non-existent git server access..."
NONEXISTENT_GIT_SERVER_ID="00000000-0000-0000-0000-000000000000"
NONEXISTENT_GIT_SERVER_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X GET -H "Authorization: Bearer $JWT_TOKEN" "$BASE_URL/gitservers/$NONEXISTENT_GIT_SERVER_ID")
assert_status 404 "$NONEXISTENT_GIT_SERVER_STATUS" "Non-existent git server access"

# Test accessing non-existent runner
echo "Testing non-existent runner access..."
NONEXISTENT_RUNNER_ID="00000000-0000-0000-0000-000000000000"
NONEXISTENT_RUNNER_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X GET -H "Authorization: Bearer $JWT_TOKEN" "$BASE_URL/runners/$NONEXISTENT_RUNNER_ID")
assert_status 404 "$NONEXISTENT_RUNNER_STATUS" "Non-existent runner access"

# --- 12. Cleanup ---
echo -e "\n--- Cleaning Up ---"

# Delete runner
echo "Deleting runner..."
DELETE_RUNNER_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE -H "Authorization: Bearer $JWT_TOKEN" "$BASE_URL/runners/$RUNNER_ID")
assert_status 204 "$DELETE_RUNNER_STATUS" "Runner deletion"

# Delete git server
echo "Deleting git server..."
DELETE_GIT_SERVER_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE -H "Authorization: Bearer $JWT_TOKEN" "$BASE_URL/gitservers/$GIT_SERVER_ID")
assert_status 204 "$DELETE_GIT_SERVER_STATUS" "Git server deletion"

# Delete organization
echo "Deleting organization..."
DELETE_ORG_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE -H "Authorization: Bearer $JWT_TOKEN" "$BASE_URL/orgs/$ORG_ID")
assert_status 204 "$DELETE_ORG_STATUS" "Organization deletion"

echo -e "\n✅ All Git Server and Runner API tests completed successfully!"
echo "Summary:"
echo "- Git server creation, retrieval, and deletion"
echo "- Runner creation, retrieval, and deletion"
echo "- Job queue management"
echo "- Error handling and validation"
echo "- Proper authentication and authorization"
