# Domain Management API Examples

This document provides curl examples for the Domain Management API endpoints.

## Prerequisites

1. Start the server: `go run cmd/server/main.go`
2. Register a user and get a JWT token
3. Create an organization and application
4. Get the application ID for domain management

## Authentication

All requests require a JWT token in the Authorization header:

```bash
Authorization: Bearer <your-jwt-token>
```

## Domain Management Endpoints

### 1. Create a Domain for an Application

**POST** `/apps/{appId}/domains`

#### Example 1: Create domain with Cloudflare DNS provider (DNS-01 challenge)

```bash
curl -X POST "http://localhost:8080/apps/123e4567-e89b-12d3-a456-426614174000/domains" \
  -H "Authorization: Bearer <your-jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "domain": "myapp.example.com",
    "provider": "cloudflare",
    "provider_config": {
      "api_key": "your-cloudflare-api-key",
      "email": "admin@example.com",
      "zone_id": "your-zone-id"
    },
    "challenge_type": "dns-01"
  }'
```

#### Example 2: Create domain with Route53 DNS provider (DNS-01 challenge)

```bash
curl -X POST "http://localhost:8080/apps/123e4567-e89b-12d3-a456-426614174000/domains" \
  -H "Authorization: Bearer <your-jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "domain": "myapp.example.com",
    "provider": "route53",
    "provider_config": {
      "api_key": "your-aws-access-key",
      "secret_key": "your-aws-secret-key",
      "hosted_zone_id": "Z1234567890ABC"
    },
    "challenge_type": "dns-01"
  }'
```

#### Example 3: Create domain with manual DNS configuration (HTTP-01 challenge)

```bash
curl -X POST "http://localhost:8080/apps/123e4567-e89b-12d3-a456-426614174000/domains" \
  -H "Authorization: Bearer <your-jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "domain": "myapp.example.com",
    "provider": "manual",
    "challenge_type": "http-01"
  }'
```

#### Response Example:

```json
{
  "id": "456e7890-e89b-12d3-a456-426614174001",
  "app_id": "123e4567-e89b-12d3-a456-426614174000",
  "domain": "myapp.example.com",
  "provider": "cloudflare",
  "provider_config": {
    "api_key": "***MASKED***",
    "email": "admin@example.com",
    "zone_id": "your-zone-id"
  },
  "cert_status": "pending",
  "cert_secret_name": "",
  "challenge_type": "dns-01",
  "dns_instructions": "",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### 2. Get Domains for an Application

**GET** `/apps/{appId}/domains`

```bash
curl -X GET "http://localhost:8080/apps/123e4567-e89b-12d3-a456-426614174000/domains" \
  -H "Authorization: Bearer <your-jwt-token>"
```

#### Response Example:

```json
[
  {
    "id": "456e7890-e89b-12d3-a456-426614174001",
    "app_id": "123e4567-e89b-12d3-a456-426614174000",
    "domain": "myapp.example.com",
    "provider": "cloudflare",
    "provider_config": {
      "api_key": "***MASKED***",
      "email": "admin@example.com",
      "zone_id": "your-zone-id"
    },
    "cert_status": "active",
    "cert_secret_name": "myapp.example.com-tls",
    "challenge_type": "dns-01",
    "dns_instructions": "",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:35:00Z"
  },
  {
    "id": "789e0123-e89b-12d3-a456-426614174002",
    "app_id": "123e4567-e89b-12d3-a456-426614174000",
    "domain": "api.example.com",
    "provider": "manual",
    "provider_config": {
      "api_key": "***MASKED***",
      "secret_key": "***MASKED***",
      "email": "",
      "zone_id": "",
      "hosted_zone_id": ""
    },
    "cert_status": "pending",
    "cert_secret_name": "",
    "challenge_type": "http-01",
    "dns_instructions": "Please create a TXT record for _acme-challenge.api.example.com with value 'YOUR_DNS_CHALLENGE_TOKEN'.",
    "created_at": "2024-01-15T11:00:00Z",
    "updated_at": "2024-01-15T11:00:00Z"
  }
]
```

### 3. Get Domain Details

**GET** `/domains/{domainId}`

```bash
curl -X GET "http://localhost:8080/domains/456e7890-e89b-12d3-a456-426614174001" \
  -H "Authorization: Bearer <your-jwt-token>"
```

#### Response Example:

```json
{
  "id": "456e7890-e89b-12d3-a456-426614174001",
  "app_id": "123e4567-e89b-12d3-a456-426614174000",
  "domain": "myapp.example.com",
  "provider": "cloudflare",
  "provider_config": {
    "api_key": "***MASKED***",
    "email": "admin@example.com",
    "zone_id": "your-zone-id"
  },
  "cert_status": "active",
  "cert_secret_name": "myapp.example.com-tls",
  "challenge_type": "dns-01",
  "dns_instructions": "",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:35:00Z"
}
```

### 4. Request Certificate for a Domain

**POST** `/domains/{domainId}/certificates`

```bash
curl -X POST "http://localhost:8080/domains/456e7890-e89b-12d3-a456-426614174001/certificates" \
  -H "Authorization: Bearer <your-jwt-token>"
```

#### Response Example:

```json
{
  "message": "Certificate request initiated. Status will be updated shortly.",
  "domain": {
    "id": "456e7890-e89b-12d3-a456-426614174001",
    "app_id": "123e4567-e89b-12d3-a456-426614174000",
    "domain": "myapp.example.com",
    "provider": "cloudflare",
    "provider_config": {
      "api_key": "***MASKED***",
      "email": "admin@example.com",
      "zone_id": "your-zone-id"
    },
    "cert_status": "pending",
    "cert_secret_name": "myapp.example.com-tls",
    "challenge_type": "dns-01",
    "dns_instructions": "",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T11:00:00Z"
  }
}
```

### 5. Get Certificate Status

**GET** `/domains/{domainId}/certificates`

```bash
curl -X GET "http://localhost:8080/domains/456e7890-e89b-12d3-a456-426614174001/certificates" \
  -H "Authorization: Bearer <your-jwt-token>"
```

#### Response Example:

```json
{
  "domain_id": "456e7890-e89b-12d3-a456-426614174001",
  "domain": "myapp.example.com",
  "cert_status": "active",
  "cert_secret_name": "myapp.example.com-tls",
  "issued_at": "2024-01-15T10:35:00Z",
  "expires_at": "2024-04-15T10:35:00Z",
  "issuer": "Let's Encrypt",
  "serial_number": "1234567890abcdef"
}
```

### 6. Delete a Domain

**DELETE** `/domains/{domainId}`

```bash
curl -X DELETE "http://localhost:8080/domains/456e7890-e89b-12d3-a456-426614174001" \
  -H "Authorization: Bearer <your-jwt-token>"
```

#### Response:

- Status: `204 No Content` (successful deletion)

## Error Responses

### 400 Bad Request

```json
{
  "error": "Invalid request body"
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
  "error": "Domain not found"
}
```

### 409 Conflict

```json
{
  "error": "Domain already exists for this application"
}
```

### 500 Internal Server Error

```json
{
  "error": "Failed to create domain"
}
```

## Notes

1. **Provider Credentials**: API keys and secret keys are encrypted in the database and masked in responses
2. **Certificate Status**: The status can be `pending`, `active`, or `failed`
3. **Challenge Types**:
   - `http-01`: HTTP challenge (requires manual DNS configuration)
   - `dns-01`: DNS challenge (automated with provider credentials)
4. **Manual Provider**: When using manual provider, DNS instructions are provided in the response
5. **Background Jobs**: Domain provisioning and certificate requests are processed asynchronously
6. **Cert-manager Integration**: The system creates cert-manager Certificate resources for automatic SSL certificate management

## Testing the API

You can use these curl commands to test the domain management functionality:

1. Create a domain with Cloudflare provider
2. Check the domain status
3. Request a certificate
4. Monitor certificate status
5. Delete the domain when done

Make sure to replace the placeholder values (JWT token, application ID, domain ID, etc.) with actual values from your system.
