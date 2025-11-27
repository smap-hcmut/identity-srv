# Authentication Capability - HttpOnly Cookie Auth Changes

## MODIFIED Requirements

### Requirement: User Login Authentication
The system SHALL authenticate users via email and password, and return authentication credentials via secure HttpOnly cookies instead of JSON response body.

#### Scenario: Successful login with standard session
- **GIVEN** a registered and verified user with email "user@example.com" and valid password
- **WHEN** the user submits login credentials with `remember: false`
- **THEN** the system validates credentials
- **AND** generates a JWT access token
- **AND** sets a Set-Cookie header with the token
- **AND** cookie SHALL have HttpOnly=true, Secure=true, SameSite=Lax attributes
- **AND** cookie SHALL have Max-Age=7200 (2 hours)
- **AND** cookie SHALL have Path=/identity
- **AND** cookie SHALL have Domain configured from environment
- **AND** returns HTTP 200 with user details in JSON body
- **AND** JSON response SHALL NOT include the token field

#### Scenario: Successful login with Remember Me
- **GIVEN** a registered and verified user with valid credentials
- **WHEN** the user submits login credentials with `remember: true`
- **THEN** the system authenticates the user
- **AND** sets a Set-Cookie header with the token
- **AND** cookie SHALL have Max-Age=2592000 (30 days)
- **AND** all other cookie attributes same as standard session

#### Scenario: Failed login with invalid credentials
- **GIVEN** a user attempts login with incorrect password
- **WHEN** the login request is processed
- **THEN** the system returns HTTP 400 with error code errWrongPassword(110005)
- **AND** SHALL NOT set any authentication cookie

#### Scenario: Login for inactive account
- **GIVEN** a registered but unverified user (is_active=false)
- **WHEN** the user attempts to login
- **THEN** the system returns HTTP 400 with error code errUserNotVerified
- **AND** SHALL NOT set any authentication cookie

## ADDED Requirements

### Requirement: User Logout
The system SHALL provide an endpoint to log out authenticated users by expiring their authentication cookie.

#### Scenario: Successful logout
- **GIVEN** an authenticated user with valid authentication cookie
- **WHEN** the user calls POST /authentication/logout
- **THEN** the system responds with HTTP 200 success
- **AND** sets a Set-Cookie header with the same cookie name
- **AND** cookie SHALL have Max-Age=-1 (immediate expiration)
- **AND** cookie SHALL have empty value
- **AND** all other cookie attributes match original cookie

#### Scenario: Logout without authentication
- **GIVEN** an unauthenticated request (no valid cookie)
- **WHEN** the user attempts to logout
- **THEN** the system returns HTTP 401 Unauthorized

### Requirement: Get Current User Information
The system SHALL provide an endpoint to retrieve the currently authenticated user's information from the JWT token context.

#### Scenario: Retrieve current user details
- **GIVEN** an authenticated user with valid cookie containing user ID
- **WHEN** the user calls GET /authentication/me
- **THEN** the system extracts user ID from JWT payload
- **AND** fetches user details from database
- **AND** returns HTTP 200 with user object containing:
  - id (UUID)
  - email (string)
  - full_name (string, nullable)
  - role (string: "USER" or "ADMIN")

#### Scenario: Get current user without authentication
- **GIVEN** an unauthenticated request (no valid cookie)
- **WHEN** the user calls GET /authentication/me
- **THEN** the system returns HTTP 401 Unauthorized

#### Scenario: Get current user with expired token
- **GIVEN** an expired JWT token in cookie
- **WHEN** the user calls GET /authentication/me
- **THEN** the system returns HTTP 401 Unauthorized
- **AND** SHALL NOT attempt to fetch user from database

### Requirement: Cookie Configuration
The system SHALL provide configurable cookie security attributes via environment variables.

#### Scenario: Load cookie configuration from environment
- **GIVEN** environment variables for cookie configuration:
  - COOKIE_NAME (default: "smap_auth_token")
  - COOKIE_DOMAIN (default: ".smap.com")
  - COOKIE_SECURE (default: "true")
  - COOKIE_SAMESITE (default: "Lax")
  - COOKIE_MAX_AGE (default: "7200")
  - COOKIE_MAX_AGE_REMEMBER (default: "2592000")
  - COOKIE_PATH (default: "/identity")
- **WHEN** the application starts
- **THEN** the system loads cookie configuration
- **AND** validates configuration values
- **AND** uses defaults for missing values
- **AND** logs loaded configuration (excluding sensitive values)

#### Scenario: Invalid cookie configuration
- **GIVEN** invalid environment variable values (e.g., COOKIE_MAX_AGE="invalid")
- **WHEN** the application attempts to start
- **THEN** the system SHALL fail to start with clear error message
- **AND** SHALL log which configuration value is invalid

### Requirement: Backward Compatibility During Migration
The system SHALL support both cookie-based and Authorization header-based authentication during the migration period.

#### Scenario: Authentication with Bearer token header (backward compatibility)
- **GIVEN** a client using the old authentication method
- **WHEN** the client sends request with Authorization: Bearer <token> header
- **AND** no authentication cookie present
- **THEN** the system extracts token from Authorization header
- **AND** validates the JWT token normally
- **AND** grants access if token is valid

#### Scenario: Cookie takes precedence over header
- **GIVEN** a request with both authentication cookie and Authorization header
- **WHEN** the authentication middleware processes the request
- **THEN** the system SHALL use the token from cookie (cookie-first strategy)
- **AND** SHALL ignore the Authorization header

## REMOVED Requirements

### Requirement: Token in Login Response Body
**Reason**: Security improvement - tokens should not be exposed in JSON responses accessible to JavaScript.

**Migration**: Frontend clients must update to read authentication from cookies automatically sent by browsers instead of manually storing and sending tokens from JSON responses.

The following response structure is REMOVED:
```json
{
  "user": { ... },
  "token": "eyJhbGc..."
}
```

New structure:
```json
{
  "user": { ... }
}
```
Token now transmitted exclusively via Set-Cookie header.

