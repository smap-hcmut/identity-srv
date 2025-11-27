# Middleware Capability - HttpOnly Cookie Auth Changes

## MODIFIED Requirements

### Requirement: JWT Token Extraction
The authentication middleware SHALL extract JWT tokens from cookies as the primary method, with fallback to Authorization header for backward compatibility.

#### Scenario: Extract token from cookie (primary method)
- **GIVEN** an HTTP request with authentication cookie named "smap_auth_token"
- **WHEN** the Auth middleware processes the request
- **THEN** the middleware SHALL attempt to read token from cookie first
- **AND** SHALL use c.Cookie(cookieName) to extract token value
- **AND** if cookie exists and is not empty, use that token
- **AND** proceed to JWT verification

#### Scenario: Fallback to Authorization header
- **GIVEN** an HTTP request without authentication cookie
- **AND** request includes Authorization: Bearer <token> header
- **WHEN** the Auth middleware processes the request
- **THEN** the middleware SHALL fall back to Authorization header
- **AND** SHALL extract token by removing "Bearer " prefix
- **AND** proceed to JWT verification with header token

#### Scenario: Cookie takes precedence over header
- **GIVEN** an HTTP request with both authentication cookie and Authorization header
- **WHEN** the Auth middleware processes the request
- **THEN** the middleware SHALL use token from cookie
- **AND** SHALL NOT check or use Authorization header token

#### Scenario: No token provided
- **GIVEN** an HTTP request without authentication cookie or Authorization header
- **WHEN** the Auth middleware processes the request
- **THEN** the middleware SHALL return HTTP 401 Unauthorized
- **AND** SHALL abort request processing
- **AND** SHALL NOT proceed to route handler

#### Scenario: Invalid token format in cookie
- **GIVEN** an HTTP request with malformed JWT in cookie
- **WHEN** the Auth middleware attempts JWT verification
- **THEN** the middleware SHALL return HTTP 401 Unauthorized
- **AND** SHALL log the verification error
- **AND** SHALL abort request processing

### Requirement: CORS Credential Support
The CORS middleware SHALL support credentials to allow browsers to send authentication cookies with cross-origin requests.

#### Scenario: CORS preflight with credentials
- **GIVEN** a preflight OPTIONS request from allowed origin
- **WHEN** CORS middleware processes the request
- **THEN** the middleware SHALL set Access-Control-Allow-Credentials: true
- **AND** SHALL set Access-Control-Allow-Origin to the specific requesting origin (not wildcard)
- **AND** SHALL include Cookie in Access-Control-Allow-Headers
- **AND** SHALL return HTTP 204 No Content

#### Scenario: Actual request with credentials from allowed origin
- **GIVEN** an HTTP request from origin "https://web.smap.com" (in allowed list)
- **WHEN** CORS middleware processes the request
- **THEN** the middleware SHALL set Access-Control-Allow-Origin: https://web.smap.com
- **AND** SHALL set Access-Control-Allow-Credentials: true
- **AND** SHALL allow request to proceed

#### Scenario: Request from disallowed origin
- **GIVEN** an HTTP request from origin "https://malicious.com" (not in allowed list)
- **WHEN** CORS middleware processes the request
- **THEN** the middleware SHALL NOT set Access-Control-Allow-Origin header
- **AND** SHALL NOT set Access-Control-Allow-Credentials
- **AND** browser will block the request (CORS policy)

## ADDED Requirements

### Requirement: Cookie Configuration in Middleware
The authentication middleware SHALL accept cookie configuration to correctly extract tokens from cookies.

#### Scenario: Initialize middleware with cookie configuration
- **GIVEN** cookie configuration with name "smap_auth_token"
- **WHEN** creating new Auth middleware instance
- **THEN** the middleware constructor SHALL accept cookie configuration
- **AND** SHALL store cookie name in middleware struct
- **AND** SHALL use configured cookie name for token extraction

#### Scenario: Use default cookie name if not configured
- **GIVEN** middleware initialization without explicit cookie configuration
- **WHEN** creating new Auth middleware instance
- **THEN** the middleware SHALL use default cookie name "smap_auth_token"

### Requirement: CORS Origin Validation
The CORS middleware SHALL validate that wildcard origins are not used when credentials are enabled.

#### Scenario: Reject wildcard origin with credentials
- **GIVEN** CORS configuration with AllowedOrigins=["*"] and AllowCredentials=true
- **WHEN** the application attempts to start
- **THEN** the system SHALL log a warning
- **AND** SHALL recommend using specific origins instead of wildcard
- **AND** SHOULD fail validation in strict mode

#### Scenario: Accept specific origins with credentials
- **GIVEN** CORS configuration with AllowedOrigins=["https://web.smap.com", "https://app.smap.com"]
- **AND** AllowCredentials=true
- **WHEN** the application starts
- **THEN** the system SHALL accept the configuration
- **AND** SHALL proceed with initialization

## MODIFIED Requirements

### Requirement: Default CORS Configuration
The default CORS configuration SHALL be updated to support credentials with explicit origin whitelisting.

#### Scenario: Default CORS configuration for development
- **GIVEN** development environment
- **WHEN** DefaultCORSConfig() is called
- **THEN** the configuration SHALL include:
  - AllowedOrigins: ["http://localhost:3000", "http://localhost:3001"]
  - AllowCredentials: true
  - AllowedHeaders: includes "Cookie"
  - AllowedMethods: standard HTTP methods

#### Scenario: Production CORS configuration from environment
- **GIVEN** production environment with CORS_ALLOWED_ORIGINS="https://web.smap.com,https://app.smap.com"
- **WHEN** loading CORS configuration
- **THEN** the system SHALL parse comma-separated origins
- **AND** SHALL set AllowedOrigins to parsed list
- **AND** SHALL enable AllowCredentials
- **AND** SHALL validate no wildcard present

## Technical Notes

### Backward Compatibility Strategy
During the migration period (estimated 4-8 weeks), the middleware will support both authentication methods:
1. **Primary**: Cookie-based (secure, recommended)
2. **Fallback**: Authorization header (deprecated, for compatibility)

After migration completion, the Authorization header fallback will be removed in a future change.

### Security Considerations
- HttpOnly flag prevents JavaScript from accessing token
- Secure flag ensures HTTPS-only transmission
- SameSite=Lax prevents CSRF attacks
- Cookie-first strategy encourages secure method adoption
- Specific origin whitelist prevents unauthorized cross-origin access

