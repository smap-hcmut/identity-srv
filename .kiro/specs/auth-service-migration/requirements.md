# Requirements Document: Auth Service Migration

## Introduction

This document specifies the requirements for migrating from the Identity Service (SaaS multi-tenant authentication) to a new Auth Service designed for on-premise enterprise SSO. This is a greenfield project with no existing users or production data, allowing for a clean implementation rather than a migration.

The Auth Service will provide enterprise-grade authentication using Google OAuth2/OIDC, JWT-based authorization with asymmetric keys, role-based access control via Google Groups, and comprehensive audit logging for compliance.

**Timeline:** 7 days
**Tech Stack:** Go 1.25, Gin framework, Viper (config), PostgreSQL, Redis, Kafka
**Scope:** Migrate Identity Service to OAuth2/JWT RS256, update all dependent services
**Service Name:** identity-service (keep existing name)

## Glossary

- **Auth_Service**: The migrated authentication and authorization service using OAuth2 and JWT RS256 (service name remains `identity-service`)
- **Identity_Service**: The current authentication service using email/password (to be migrated, not replaced)
- **OAuth2**: Open standard for access delegation, used for SSO with Google Workspace
- **OIDC**: OpenID Connect, authentication layer on top of OAuth2
- **JWT**: JSON Web Token, used for stateless authentication
- **RS256**: RSA Signature with SHA-256, asymmetric algorithm for JWT signing
- **JWKS**: JSON Web Key Set, endpoint exposing public keys for JWT verification
- **SSO**: Single Sign-On, allows users to authenticate once and access multiple services
- **RBAC**: Role-Based Access Control, authorization based on user roles
- **Google_Directory_API**: Google API for accessing organizational data including groups
- **Service_Account**: Google account used by applications to make API calls
- **Audit_Log**: Record of security-relevant events for compliance and investigation
- **Token_Blacklist**: Redis-based list of revoked JWT tokens
- **Session**: Server-side state stored in Redis, linked to JWT token
- **HttpOnly_Cookie**: Cookie that cannot be accessed by JavaScript, used for secure token storage (maintained from Identity Service)
- **Domain_Validation**: Check that user's email domain is in allowed list
- **Role_Mapping**: Configuration mapping Google Groups to application roles (ADMIN, ANALYST, VIEWER)
- **Key_Rotation**: Process of replacing cryptographic keys periodically
- **Project_Service**: Service managing project data, requires authentication
- **Ingest_Service**: Service handling data ingestion, requires authentication
- **Knowledge_Service**: Service managing knowledge base, requires authentication
- **Notification_Service**: Service handling notifications and WebSocket connections, requires authentication

## Requirements

### Requirement 1: Google OAuth2 Authentication

**User Story:** As a user, I want to log in using my Google Workspace account, so that I can access the system without managing separate credentials.

#### Acceptance Criteria

1. WHEN a user accesses the login endpoint, THE Auth_Service SHALL redirect to Google OAuth2 authorization page
2. WHEN Google redirects back with an authorization code, THE Auth_Service SHALL exchange the code for user information
3. WHEN a user's email domain is in the allowed domains list, THE Auth_Service SHALL create or update the user record
4. IF a user's email domain is not in the allowed domains list, THEN THE Auth_Service SHALL reject authentication and return a domain not allowed error
5. IF a user's account is in the blocklist, THEN THE Auth_Service SHALL reject authentication and return an account blocked error
6. WHEN authentication succeeds, THE Auth_Service SHALL create a session in Redis with 8 hour TTL
7. WHEN authentication succeeds, THE Auth_Service SHALL generate a JWT token and set it as an HttpOnly cookie (maintaining current Identity Service cookie approach)
8. THE Auth_Service SHALL set cookie with Secure flag, SameSite=Lax, and appropriate domain configuration
9. WHEN OAuth2 flow fails, THE Auth_Service SHALL return an appropriate error message to the user

### Requirement 2: JWT Token Management

**User Story:** As a system architect, I want JWT tokens signed with asymmetric keys, so that services can verify tokens independently without shared secrets.

#### Acceptance Criteria

1. THE Auth_Service SHALL generate an RSA keypair for JWT signing
2. THE Auth_Service SHALL sign JWT tokens using the private key with RS256 algorithm
3. THE Auth_Service SHALL include user_id, email, role, groups, and jti in the JWT payload
4. THE Auth_Service SHALL set JWT expiration to 8 hours for normal sessions
5. THE Auth_Service SHALL expose public keys via JWKS endpoint at `/.well-known/jwks.json`
6. THE Auth_Service SHALL store active JWT keys in the jwt_keys table with status tracking
7. WHEN a JWT token is created, THE Auth_Service SHALL record the jti in Redis for blacklist checking
8. THE Auth_Service SHALL support loading private keys from file, environment variable, or Kubernetes secret

### Requirement 3: User Session Management

**User Story:** As a user, I want my session to persist across browser restarts, so that I don't have to log in frequently.

#### Acceptance Criteria

1. WHEN a user logs in successfully, THE Auth_Service SHALL create a session record in Redis
2. THE Auth_Service SHALL set session TTL to 8 hours for normal login
3. WHERE a user selects "remember me", THE Auth_Service SHALL extend session TTL to 7 days
4. WHEN a user accesses `/auth/me`, THE Auth_Service SHALL return current user information from the JWT token
5. WHEN a user logs out, THE Auth_Service SHALL delete the session from Redis
6. WHEN a user logs out, THE Auth_Service SHALL add the JWT jti to the blacklist with TTL equal to remaining token lifetime
7. WHEN a user logs out, THE Auth_Service SHALL expire the HttpOnly cookie

### Requirement 4: Google Groups Role Mapping

**User Story:** As an administrator, I want user roles determined by Google Groups membership, so that access control is centrally managed in Google Workspace.

#### Acceptance Criteria

1. WHEN a user logs in, THE Auth_Service SHALL fetch the user's Google Groups via Directory API
2. THE Auth_Service SHALL cache fetched groups in Redis with 5 minute TTL
3. THE Auth_Service SHALL map Google Groups to application roles using configuration file
4. THE Auth_Service SHALL assign the highest privilege role if a user belongs to multiple groups
5. THE Auth_Service SHALL default to VIEWER role if no group mapping matches
6. THE Auth_Service SHALL include the assigned role in the JWT token
7. THE Auth_Service SHALL include all Google Groups in the JWT token for fine-grained authorization
8. THE Auth_Service SHALL refresh group membership from cache or Google API on each login

### Requirement 5: Audit Logging

**User Story:** As a compliance officer, I want all authentication events logged, so that I can audit access and investigate security incidents.

#### Acceptance Criteria

1. WHEN a user logs in, THE Auth_Service SHALL publish an audit event to Kafka
2. WHEN a user logs out, THE Auth_Service SHALL publish an audit event to Kafka
3. WHEN authentication fails, THE Auth_Service SHALL publish an audit event to Kafka with failure reason
4. WHEN a token is revoked, THE Auth_Service SHALL publish an audit event to Kafka
5. THE Auth_Service SHALL include user_id, action, resource_type, resource_id, metadata, ip_address, and user_agent in audit events
6. THE Audit_Consumer SHALL consume events from Kafka and batch insert into audit_logs table
7. THE Audit_Consumer SHALL batch insert every 100 messages or every 5 seconds, whichever comes first
8. THE Auth_Service SHALL automatically delete audit logs older than 90 days

### Requirement 6: Token Blacklist

**User Story:** As a security administrator, I want to revoke JWT tokens immediately, so that compromised tokens cannot be used.

#### Acceptance Criteria

1. WHEN an admin revokes a specific token, THE Auth_Service SHALL add the jti to Redis blacklist
2. WHEN an admin revokes all tokens for a user, THE Auth_Service SHALL add all active jtis for that user to Redis blacklist
3. THE Auth_Service SHALL set blacklist entry TTL to the remaining token lifetime
4. WHEN a service validates a JWT, THE JWT_Middleware SHALL check if jti is in the blacklist
5. IF a JWT jti is in the blacklist, THEN THE JWT_Middleware SHALL reject the request with 401 Unauthorized
6. THE Auth_Service SHALL provide an internal endpoint for token revocation accessible only to admins

### Requirement 7: JWT Verification Middleware

**User Story:** As a service developer, I want a reusable middleware package for JWT verification, so that I can easily secure my service endpoints.

#### Acceptance Criteria

1. THE JWT_Middleware SHALL fetch public keys from Auth_Service JWKS endpoint
2. THE JWT_Middleware SHALL cache public keys in memory with 1 hour TTL
3. THE JWT_Middleware SHALL verify JWT signature using cached public keys
4. THE JWT_Middleware SHALL validate JWT expiration, issuer, and audience claims
5. THE JWT_Middleware SHALL check if JWT jti is in Redis blacklist
6. WHEN JWT verification succeeds, THE JWT_Middleware SHALL extract user_id, email, role, and groups into request context
7. WHEN JWT verification fails, THE JWT_Middleware SHALL return 401 Unauthorized
8. THE JWT_Middleware SHALL provide role-based authorization helpers (RequireRole, RequireAnyRole)

### Requirement 8: Service Integration

**User Story:** As a service owner, I want my service to use the new Auth Service, so that users can access it with SSO.

#### Acceptance Criteria

1. WHEN Project_Service receives a request, THE JWT_Middleware SHALL verify the JWT token
2. WHEN Ingest_Service receives a request, THE JWT_Middleware SHALL verify the JWT token
3. WHEN Knowledge_Service receives a request, THE JWT_Middleware SHALL verify the JWT token
4. WHEN Notification_Service receives a WebSocket connection, THE JWT_Middleware SHALL verify the JWT token
5. THE Project_Service SHALL require ANALYST or ADMIN role for project creation
6. THE Project_Service SHALL require ADMIN role for project deletion
7. THE Ingest_Service SHALL require ANALYST or ADMIN role for data ingestion
8. THE Knowledge_Service SHALL allow VIEWER role for read operations

### Requirement 9: Database Schema

**User Story:** As a database administrator, I want a clean schema for the Auth Service, so that data is properly structured and performant.

#### Acceptance Criteria

1. THE Auth_Service SHALL create a users table with id, email, name, avatar_url, role, is_active, last_login_at, created_at, and updated_at columns
2. THE Auth_Service SHALL create an audit_logs table with id, user_id, action, resource_type, resource_id, metadata, ip_address, user_agent, created_at, and expires_at columns
3. THE Auth_Service SHALL create a jwt_keys table with kid, private_key, public_key, status, created_at, expires_at, and retired_at columns
4. THE Auth_Service SHALL create an index on users.email for fast lookup
5. THE Auth_Service SHALL create an index on audit_logs.user_id for user activity queries
6. THE Auth_Service SHALL create an index on audit_logs.created_at for time-based queries
7. THE Auth_Service SHALL create an index on audit_logs.expires_at for cleanup queries
8. THE Auth_Service SHALL set default expires_at to 90 days from created_at for audit logs

### Requirement 10: API Endpoints

**User Story:** As a frontend developer, I want clear API endpoints for authentication, so that I can implement the login flow.

#### Acceptance Criteria

1. THE Auth_Service SHALL provide GET `/auth/login` endpoint that redirects to Google OAuth
2. THE Auth_Service SHALL provide GET `/auth/callback` endpoint that handles OAuth callback
3. THE Auth_Service SHALL provide POST `/auth/logout` endpoint that invalidates tokens and sessions
4. THE Auth_Service SHALL provide GET `/auth/me` endpoint that returns current user information
5. THE Auth_Service SHALL provide GET `/.well-known/jwks.json` endpoint that returns public keys
6. THE Auth_Service SHALL provide GET `/health` endpoint that returns service health status
7. THE Auth_Service SHALL provide POST `/internal/validate` endpoint for fallback token validation
8. THE Auth_Service SHALL provide POST `/internal/revoke-token` endpoint for admin token revocation
9. THE Auth_Service SHALL provide GET `/audit-logs` endpoint for querying audit logs with pagination
10. THE Auth_Service SHALL authenticate internal endpoints using X-Service-Key header with encrypted service key
11. THE Auth_Service SHALL validate X-Service-Key by decrypting and comparing with configured service keys
12. THE Auth_Service SHALL return 401 Unauthorized if X-Service-Key is missing or invalid for internal endpoints

### Requirement 11: Configuration Management

**User Story:** As a DevOps engineer, I want flexible configuration options, so that I can deploy the Auth Service in different environments.

#### Acceptance Criteria

1. THE Auth_Service SHALL load configuration from environment variables
2. THE Auth_Service SHALL support configuration file (auth-config.yaml) for complex settings
3. THE Auth_Service SHALL validate required configuration on startup
4. THE Auth_Service SHALL support configuring allowed email domains
5. THE Auth_Service SHALL support configuring Google Groups to role mappings
6. THE Auth_Service SHALL support configuring JWT key sources (file, env, k8s secret)
7. THE Auth_Service SHALL support configuring session TTL and refresh token TTL
8. THE Auth_Service SHALL support configuring Redis connection parameters
9. THE Auth_Service SHALL support configuring Kafka connection parameters
10. THE Auth_Service SHALL fail fast with clear error messages if configuration is invalid
11. THE Auth_Service SHALL support configuring service API keys for internal endpoint authentication
12. THE Auth_Service SHALL encrypt service API keys using the same encryption mechanism as other secrets

### Requirement 12: Frontend OAuth Integration

**User Story:** As a frontend developer, I want to integrate OAuth login, so that users can authenticate with Google.

#### Acceptance Criteria

1. WHEN a user clicks login, THE Frontend SHALL redirect to `/auth/login`
2. WHEN OAuth callback completes, THE Frontend SHALL be redirected to the dashboard
3. THE Frontend SHALL maintain existing axios configuration with `withCredentials: true` for HttpOnly cookie authentication
4. WHEN an API request returns 401, THE Frontend SHALL redirect to login page
5. WHEN an API request returns 403, THE Frontend SHALL display permission denied message
6. THE Frontend SHALL continue using HttpOnly cookie approach (no changes to cookie handling from Identity Service)
7. THE Frontend SHALL call `/auth/me` to get current user information on app load
8. WHEN a user logs out, THE Frontend SHALL call `/auth/logout` and redirect to login page
9. THE Frontend SHALL maintain existing CORS configuration for cookie-based authentication

### Requirement 13: Testing and Quality

**User Story:** As a quality engineer, I want comprehensive tests, so that the Auth Service is reliable and secure.

#### Acceptance Criteria

1. THE Auth_Service SHALL have unit tests for OAuth flow with test coverage above 80%
2. THE Auth_Service SHALL have unit tests for JWT signing and verification
3. THE Auth_Service SHALL have unit tests for Google Groups synchronization
4. THE Auth_Service SHALL have integration tests for complete authentication flow
5. THE Auth_Service SHALL provide mock OAuth provider for testing
6. THE JWT_Middleware SHALL have unit tests for token verification
7. THE JWT_Middleware SHALL have unit tests for role-based authorization
8. THE Audit_Consumer SHALL have unit tests for batch processing

### Requirement 14: Documentation

**User Story:** As a developer, I want comprehensive documentation, so that I can understand and integrate with the Auth Service.

#### Acceptance Criteria

1. THE Auth_Service SHALL provide API documentation with endpoint descriptions and examples
2. THE Auth_Service SHALL provide OAuth flow diagram showing the complete authentication sequence
3. THE Auth_Service SHALL provide JWT structure documentation with claim descriptions
4. THE Auth_Service SHALL provide integration guide for JWT middleware usage
5. THE Auth_Service SHALL provide deployment guide with Google OAuth setup instructions
6. THE Auth_Service SHALL provide environment variable reference
7. THE Auth_Service SHALL provide frontend migration guide with code examples
8. THE Auth_Service SHALL provide troubleshooting guide for common issues

### Requirement 15: Security and Compliance

**User Story:** As a security officer, I want the Auth Service to follow security best practices, so that the system is protected against common attacks.

#### Acceptance Criteria

1. THE Auth_Service SHALL use HTTPS-only cookies in production
2. THE Auth_Service SHALL set HttpOnly flag on authentication cookies
3. THE Auth_Service SHALL set SameSite=Lax on authentication cookies for CSRF protection
4. THE Auth_Service SHALL validate redirect URLs to prevent open redirect attacks
5. THE Auth_Service SHALL rate limit login attempts to prevent brute force attacks
6. THE Auth_Service SHALL log all authentication failures with IP address
7. THE Auth_Service SHALL use secure random number generation for JWT jti
8. THE Auth_Service SHALL validate JWT audience claim to prevent token reuse across services
9. THE Auth_Service SHALL implement key rotation strategy for JWT signing keys
10. THE Auth_Service SHALL store private keys encrypted at rest in production
