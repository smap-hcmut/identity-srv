# Auth Service API Reference

## Overview

The Auth Service provides enterprise Single Sign-On (SSO) authentication using Google OAuth2/OIDC, role-based access control through Google Groups mapping, and JWT-based authorization with asymmetric keys (RS256).

**Base URL**: `/authentication`  
**Authentication**: HttpOnly cookies (JWT tokens)  
**API Version**: 1.0

---

## Table of Contents

1. [Public Endpoints](#public-endpoints)
2. [Protected Endpoints](#protected-endpoints)
3. [Internal Endpoints](#internal-endpoints)
4. [OAuth Flow](#oauth-flow)
5. [JWT Structure](#jwt-structure)
6. [Error Codes](#error-codes)

---

## Public Endpoints

### 1. OAuth Login

Initiates Google OAuth2 authentication flow.

**Endpoint**: `GET /authentication/login`

**Query Parameters**:

- `redirect` (optional): URL to redirect after successful login. Must be in allowed redirect URLs list.

**Response**: `302 Redirect` to Google OAuth consent page

**Example**:

```bash
curl -X GET "https://api.example.com/authentication/login?redirect=/dashboard" \
  -L
```

**Security**:

- Rate limited: 5 attempts per 15 minutes per IP
- Redirect URL validated against allowed list
- CSRF protection via state parameter

---

### 2. OAuth Callback

Handles OAuth2 callback from Google.

**Endpoint**: `GET /authentication/callback`

**Query Parameters**:

- `code` (required): Authorization code from Google
- `state` (required): CSRF protection state parameter

**Response**: `302 Redirect` to dashboard or specified redirect URL

**Sets Cookie**:

```
smap_auth_token=<JWT_TOKEN>; HttpOnly; Secure; SameSite=Lax; Max-Age=28800
```

**Error Responses**:

- `400 Bad Request`: Invalid state or missing code
- `403 Forbidden`: Domain not allowed or account blocked
- `500 Internal Server Error`: OAuth exchange failed

**Example Flow**:

```bash
# User is redirected from Google with code and state
# Browser automatically follows redirect
GET /authentication/callback?code=4/0AX4XfWh...&state=state-1234567890
```

---

### 3. JWKS Endpoint

Returns public keys for JWT verification (JSON Web Key Set).

**Endpoint**: `GET /.well-known/jwks.json`

**Response**: `200 OK`

```json
{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "kid": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "alg": "RS256",
      "n": "xGOr-H7A...",
      "e": "AQAB"
    }
  ]
}
```

**Example**:

```bash
curl -X GET "https://api.example.com/.well-known/jwks.json"
```

**Usage**: Other services fetch public keys from this endpoint to verify JWT signatures.

---

## Protected Endpoints

These endpoints require authentication via HttpOnly cookie.

### 4. Get Current User

Returns information about the currently authenticated user.

**Endpoint**: `GET /authentication/me`

**Authentication**: Required (HttpOnly cookie)

**Response**: `200 OK`

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "name": "John Doe",
  "avatar_url": "https://lh3.googleusercontent.com/...",
  "role": "ANALYST",
  "groups": ["analysts@example.com", "data-team@example.com"],
  "is_active": true,
  "last_login_at": "2026-02-09T10:30:00Z",
  "created_at": "2025-01-15T08:00:00Z",
  "updated_at": "2026-02-09T10:30:00Z"
}
```

**Error Responses**:

- `401 Unauthorized`: Missing or invalid JWT token

**Example**:

```bash
curl -X GET "https://api.example.com/authentication/me" \
  --cookie "smap_auth_token=<JWT_TOKEN>"
```

---

### 5. Logout

Invalidates the current session and JWT token.

**Endpoint**: `POST /authentication/logout`

**Authentication**: Required (HttpOnly cookie)

**Response**: `200 OK`

```json
{
  "message": "Logged out successfully"
}
```

**Cookie Action**: Expires the `smap_auth_token` cookie

**Side Effects**:

- Deletes session from Redis
- Adds JWT `jti` to blacklist
- Publishes audit event (LOGOUT)

**Example**:

```bash
curl -X POST "https://api.example.com/authentication/logout" \
  --cookie "smap_auth_token=<JWT_TOKEN>"
```

---

## Internal Endpoints

These endpoints require `X-Service-Key` header for service-to-service authentication.

### 6. Validate Token

Validates a JWT token (fallback for services that cannot verify locally).

**Endpoint**: `POST /internal/validate`

**Authentication**: Required (`X-Service-Key` header)

**Request Body**:

```json
{
  "token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response**: `200 OK`

```json
{
  "valid": true,
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "role": "ANALYST",
  "groups": ["analysts@example.com"],
  "expires_at": "2026-02-09T18:30:00Z"
}
```

**Error Responses**:

- `401 Unauthorized`: Missing or invalid service key
- `400 Bad Request`: Invalid token format
- `403 Forbidden`: Token expired or blacklisted

**Example**:

```bash
curl -X POST "https://api.example.com/internal/validate" \
  -H "X-Service-Key: <ENCRYPTED_SERVICE_KEY>" \
  -H "Content-Type: application/json" \
  -d '{"token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."}'
```

---

### 7. Revoke Token

Revokes a specific token or all tokens for a user (admin only).

**Endpoint**: `POST /internal/revoke-token`

**Authentication**: Required (`X-Service-Key` header + ADMIN role)

**Request Body**:

```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "jti": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}
```

**Parameters**:

- `user_id` (optional): Revoke all tokens for this user
- `jti` (optional): Revoke specific token by JTI
- At least one parameter required

**Response**: `200 OK`

```json
{
  "message": "Token(s) revoked successfully",
  "revoked_count": 3
}
```

**Side Effects**:

- Adds JTI(s) to Redis blacklist
- Deletes session(s) from Redis
- Publishes audit event (TOKEN_REVOKED)

**Example**:

```bash
# Revoke all tokens for a user
curl -X POST "https://api.example.com/internal/revoke-token" \
  -H "X-Service-Key: <ENCRYPTED_SERVICE_KEY>" \
  -H "Content-Type: application/json" \
  -d '{"user_id": "550e8400-e29b-41d4-a716-446655440000"}'

# Revoke specific token
curl -X POST "https://api.example.com/internal/revoke-token" \
  -H "X-Service-Key: <ENCRYPTED_SERVICE_KEY>" \
  -H "Content-Type: application/json" \
  -d '{"jti": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"}'
```

---

### 8. Get User by ID

Retrieves user information by user ID (internal use only).

**Endpoint**: `GET /internal/users/:id`

**Authentication**: Required (`X-Service-Key` header)

**Path Parameters**:

- `id` (required): User UUID

**Response**: `200 OK`

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "name": "John Doe",
  "avatar_url": "https://lh3.googleusercontent.com/...",
  "role": "ANALYST",
  "is_active": true,
  "last_login_at": "2026-02-09T10:30:00Z",
  "created_at": "2025-01-15T08:00:00Z",
  "updated_at": "2026-02-09T10:30:00Z"
}
```

**Error Responses**:

- `401 Unauthorized`: Missing or invalid service key
- `404 Not Found`: User not found

**Example**:

```bash
curl -X GET "https://api.example.com/internal/users/550e8400-e29b-41d4-a716-446655440000" \
  -H "X-Service-Key: <ENCRYPTED_SERVICE_KEY>"
```

---

## OAuth Flow

### Sequence Diagram

```
┌──────┐         ┌─────────┐         ┌──────────────┐         ┌────────┐
│ User │         │ Web UI  │         │ Identity Svc │         │ Google │
└──┬───┘         └────┬────┘         └──────┬───────┘         └───┬────┘
   │                  │                     │                     │
   │ Click "Login"    │                     │                     │
   ├─────────────────>│                     │                     │
   │                  │                     │                     │
   │                  │ GET /auth/login     │                     │
   │                  ├────────────────────>│                     │
   │                  │                     │                     │
   │                  │                     │ Redirect to OAuth   │
   │                  │<────────────────────┤                     │
   │                  │                     │                     │
   │                  │ 302 Redirect        │                     │
   │<─────────────────┤                     │                     │
   │                  │                     │                     │
   │ OAuth Consent Screen                   │                     │
   ├───────────────────────────────────────────────────────────>│
   │                  │                     │                     │
   │                  │                     │ Callback with code  │
   │                  │                     │<────────────────────┤
   │                  │                     │                     │
   │                  │                     │ Exchange code       │
   │                  │                     ├────────────────────>│
   │                  │                     │                     │
   │                  │                     │ User info + token   │
   │                  │                     │<────────────────────┤
   │                  │                     │                     │
   │                  │                     │ Fetch Groups        │
   │                  │                     ├────────────────────>│
   │                  │                     │                     │
   │                  │                     │ Groups list         │
   │                  │                     │<────────────────────┤
   │                  │                     │                     │
   │                  │ Set Cookie + Redirect                     │
   │                  │<────────────────────┤                     │
   │                  │                     │                     │
   │ Dashboard        │                     │                     │
   │<─────────────────┤                     │                     │
   │                  │                     │                     │
```

### Flow Steps

1. **User clicks "Login with Google"** → Frontend redirects to `/authentication/login`
2. **Identity Service generates state** → Stores in cookie for CSRF protection
3. **Redirect to Google OAuth** → User sees consent screen
4. **User approves** → Google redirects to `/authentication/callback?code=...&state=...`
5. **Identity Service validates state** → Prevents CSRF attacks
6. **Exchange code for token** → Calls Google OAuth2 token endpoint
7. **Get user info** → Calls Google UserInfo API
8. **Validate domain** → Checks if email domain is allowed
9. **Check blocklist** → Ensures user is not blocked
10. **Fetch Google Groups** → Calls Directory API (cached 5 minutes)
11. **Map groups to role** → Applies role mapping configuration
12. **Create/update user** → Stores in PostgreSQL
13. **Generate JWT** → Signs with RS256 private key
14. **Create session** → Stores in Redis (8 hours TTL)
15. **Set HttpOnly cookie** → Returns JWT in secure cookie
16. **Publish audit event** → Sends LOGIN event to Kafka
17. **Redirect to dashboard** → User is authenticated

---

## JWT Structure

### Header

```json
{
  "alg": "RS256",
  "typ": "JWT",
  "kid": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}
```

### Payload (Claims)

```json
{
  "iss": "smap-auth-service",
  "aud": ["smap-api"],
  "sub": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "role": "ANALYST",
  "groups": ["analysts@example.com", "data-team@example.com"],
  "jti": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "iat": 1707473400,
  "exp": 1707502200
}
```

### Claims Description

| Claim    | Type   | Description                                                  |
| -------- | ------ | ------------------------------------------------------------ |
| `iss`    | string | Issuer - always "smap-auth-service"                          |
| `aud`    | array  | Audience - list of services that can accept this token       |
| `sub`    | string | Subject - user UUID                                          |
| `email`  | string | User email address                                           |
| `role`   | string | User role (ADMIN, ANALYST, VIEWER)                           |
| `groups` | array  | Google Groups the user belongs to                            |
| `jti`    | string | JWT ID - unique identifier for token tracking and revocation |
| `iat`    | number | Issued At - Unix timestamp                                   |
| `exp`    | number | Expiration - Unix timestamp (8 hours from iat)               |

### Signature

Signed with RSA private key using RS256 algorithm. Public key available at `/.well-known/jwks.json`.

---

## Error Codes

### Authentication Errors

| Code                    | HTTP Status | Description                         | Solution                         |
| ----------------------- | ----------- | ----------------------------------- | -------------------------------- |
| `INVALID_STATE`         | 400         | OAuth state parameter mismatch      | Retry login flow                 |
| `MISSING_CODE`          | 400         | Authorization code missing          | Retry login flow                 |
| `OAUTH_EXCHANGE_FAILED` | 500         | Failed to exchange code for token   | Check Google OAuth configuration |
| `USER_INFO_FAILED`      | 500         | Failed to get user info from Google | Check network connectivity       |
| `DOMAIN_NOT_ALLOWED`    | 403         | Email domain not in allowed list    | Contact admin to add domain      |
| `ACCOUNT_BLOCKED`       | 403         | User account is blocked             | Contact admin to unblock         |
| `JWT_GENERATION_FAILED` | 500         | Failed to generate JWT token        | Check JWT key configuration      |
| `SESSION_CREATE_FAILED` | 500         | Failed to create session in Redis   | Check Redis connectivity         |
| `INVALID_REDIRECT`      | 400         | Redirect URL not in allowed list    | Use allowed redirect URL         |
| `TOO_MANY_REQUESTS`     | 429         | Rate limit exceeded                 | Wait 30 minutes and retry        |

### Authorization Errors

| Code                  | HTTP Status | Description                  | Solution                        |
| --------------------- | ----------- | ---------------------------- | ------------------------------- |
| `UNAUTHORIZED`        | 401         | Missing or invalid JWT token | Login again                     |
| `FORBIDDEN`           | 403         | Insufficient permissions     | Contact admin for role upgrade  |
| `TOKEN_EXPIRED`       | 401         | JWT token has expired        | Login again                     |
| `TOKEN_REVOKED`       | 401         | JWT token has been revoked   | Login again                     |
| `INVALID_SERVICE_KEY` | 401         | X-Service-Key header invalid | Check service key configuration |

### Validation Errors

| Code              | HTTP Status | Description                    | Solution             |
| ----------------- | ----------- | ------------------------------ | -------------------- |
| `INVALID_REQUEST` | 400         | Request body validation failed | Check request format |
| `USER_NOT_FOUND`  | 404         | User ID not found              | Verify user ID       |

---

## Service Key Generation

For internal endpoints, services need an encrypted service key.

### Generate Service Key

```bash
# 1. Generate random key
openssl rand -base64 32

# 2. Encrypt with encrypter key (configured in auth-config.yaml)
# Use the same encryption mechanism as Identity Service

# 3. Add to auth-config.yaml
service_keys:
  project-service: "<ENCRYPTED_KEY>"
  ingest-service: "<ENCRYPTED_KEY>"
  knowledge-service: "<ENCRYPTED_KEY>"
```

### Use Service Key

```bash
curl -X POST "https://api.example.com/internal/validate" \
  -H "X-Service-Key: <ENCRYPTED_SERVICE_KEY>" \
  -H "Content-Type: application/json" \
  -d '{"token": "..."}'
```

---

## Rate Limiting

### Login Endpoints

- **Endpoints**: `/authentication/login`, `/authentication/callback`
- **Limit**: 5 failed attempts per 15 minutes per IP address
- **Block Duration**: 30 minutes
- **Storage**: Redis

### Behavior

1. **Successful login** → Counter reset
2. **Failed login** (domain not allowed, account blocked) → Counter incremented
3. **5th failure** → IP blocked for 30 minutes
4. **During block** → Returns `429 Too Many Requests`

---

## Security Best Practices

### For Frontend Developers

1. **Always use HTTPS** in production
2. **Never access JWT from JavaScript** (HttpOnly cookie)
3. **Set `withCredentials: true`** in axios configuration
4. **Handle 401 errors** → Redirect to login
5. **Handle 403 errors** → Show permission denied message
6. **Validate redirect URLs** before passing to `/auth/login`

### For Backend Developers

1. **Verify JWT signature** using public key from JWKS endpoint
2. **Check JWT expiration** (`exp` claim)
3. **Validate audience** (`aud` claim) matches your service
4. **Check blacklist** before accepting token
5. **Use service keys** for internal API calls
6. **Never log JWT tokens** (contains sensitive data)

---

## Support

For issues or questions:

- **Documentation**: See `documents/identity-service-troubleshooting.md`
- **Deployment**: See `documents/deployment-guide.md`
- **Integration**: See `documents/auth-service-integration.md`
