#!/bin/bash

# OneClick API Test Script
# This script demonstrates how to use the OneClick authentication API

BASE_URL="http://localhost:8080"

echo "=== OneClick API Test Script ==="
echo "Base URL: $BASE_URL"
echo ""

# Test health endpoint
echo "1. Testing health endpoint..."
curl -s -X GET "$BASE_URL/health" | jq .
echo ""

# Register a new user
echo "2. Registering a new user..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "password123"
  }')

echo "Register response:"
echo "$REGISTER_RESPONSE" | jq .
echo ""

# Login with the user
echo "3. Logging in..."
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

# Get user profile
echo "4. Getting user profile..."
curl -s -X GET "$BASE_URL/auth/me" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
echo ""

# Test invalid login
echo "5. Testing invalid login..."
curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "wrongpassword"
  }' | jq .
echo ""

# Test registration with existing email
echo "6. Testing registration with existing email..."
curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jane Doe",
    "email": "john@example.com",
    "password": "password456"
  }' | jq .
echo ""

echo "=== Test completed ==="
