# Implementation Plan: Auth Service Migration

## Overview

This implementation plan covers the migration from Identity Service to Auth Service for on-premise enterprise deployments. The plan is reorganized to focus on **Identity Service implementation** first, then provide **comprehensive migration documentation** for other services and frontend.

**Timeline**: 5 days (Identity Service only)
**Tech Stack**: Go 1.25, Gin framework, Viper (config), PostgreSQL, Redis, Kafka
**Approach**: Greenfield implementation (no data migration needed)
**Service Name**: identity-service (keep existing name)
**Code Structure**: Follow existing pattern (internal/domain/delivery/usecase/repository)

**Note**: Service integration (Day 4) and Frontend integration (Day 5) are **documentation-only** - no code implementation required. Other teams will use the guides to migrate their services.

## Tasks

- [x] 1. Day 1-2: Core Auth Service Setup (OAuth2 + JWT)
  - [x] 1.0 Remove legacy authentication code
    - Delete internal/authentication/usecase/authentication.go (old email/password logic)
    - Delete internal/authentication/delivery/http/handler.go methods: Register, SendOTP, VerifyOTP, ChangePassword
    - Delete internal/plan package entirely (no longer needed)
    - Delete internal/subscription package entirely (no longer needed)
    - Delete internal/smtp package entirely (no longer needed for OTP)
    - Remove RabbitMQ initialization from cmd/api/main.go (no longer needed for email queue)
    - Remove RabbitMQ consumer from cmd/consumer/main.go (no longer needed)
    - Clean up unused imports and dependencies
    - _Requirements: Migration from SaaS to Enterprise SSO_
    - _Pattern: Clean removal of unused code_

  - [x] 1.1 Setup project structure and dependencies
    - Update go.mod with new dependencies: golang-jwt/jwt, google oauth2, redis, kafka clients
    - Add Viper for configuration management (replace env parsing)
    - Keep existing Gin framework and middleware structure
    - Update config/config.go to use Viper instead of env tags
    - Setup logging with structured logger (keep existing)
    - _Requirements: 11.1, 11.2_
    - _Pattern: Follow existing config/config.go structure_

  - [x] 1.2 Implement database schema and migrations
    - Delete old migration files: migration/01_add_user_indexes.sql.sql, migration/02_add_role_hash_to_users.sql
    - Create NEW migration file: migration/01_auth_service_schema.sql with complete schema:
      * CREATE TABLE users (id, email, name, avatar_url, role, is_active, last_login_at, created_at, updated_at)
      * CREATE TABLE audit_logs (id, user_id, action, resource_type, resource_id, metadata, ip_address, user_agent, created_at, expires_at)
      * CREATE TABLE jwt_keys (kid, private_key, public_key, status, created_at, expires_at, retired_at)
      * CREATE indexes on users.email, audit_logs.user_id, audit_logs.created_at, audit_logs.expires_at
    - Delete internal/sqlboiler/* (all generated files)
    - Regenerate sqlboiler models from new schema: `sqlboiler psql`
    - Update internal/model/user.go to match new schema (remove password_hash, otp fields; add avatar_url, last_login_at)
    - Delete internal/model/plan.go (no longer needed)
    - Delete internal/model/subscription.go (no longer needed)
    - Create internal/model/audit_log.go for new audit_logs table
    - Create internal/model/jwt_key.go for new jwt_keys table
    - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5, 9.6, 9.7_
    - _Pattern: Follow existing 2-layer pattern (sqlboiler → domain model)_

  - [x] 1.3 Implement Google OAuth2 flow
    - Create internal/authentication/delivery/http/oauth.go for OAuth handlers
    - Implement GET /auth/login endpoint (redirect to Google OAuth)
    - Implement GET /auth/callback endpoint (handle OAuth callback)
    - Add domain validation logic in usecase layer
    - Add blocklist checking in usecase layer
    - Update internal/authentication/interface.go with new methods
    - Remove old endpoints from routes.go: POST /auth/register, POST /auth/send-otp, POST /auth/verify-otp, POST /auth/change-password
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_
    - _Pattern: Follow existing handler.go structure with Gin context_

  - [x] 1.4 Implement JWT token generation (RS256)
    - Create pkg/jwt package for JWT management (follow existing pkg structure)
    - Generate RSA keypair (2048-bit) on startup or load from config
    - Implement JWT signing with private key using RS256 algorithm
    - Add all required claims (iss, aud, sub, email, role, groups, jti, iat, exp)
    - Store keypair in jwt_keys table with status tracking
    - Support loading keys from file/env/k8s secret via Viper config
    - Delete pkg/scope if it only contains HS256 logic (or update to use RS256)
    - Remove old JWT HS256 signing code from authentication usecase
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.6, 2.8_
    - _Pattern: Create new pkg/jwt similar to existing pkg/encrypter_

  - [x] 1.5 Implement JWKS endpoint
    - Create GET /.well-known/jwks.json endpoint in internal/authentication/delivery/http/
    - Expose public keys in JWKS format (standard JSON Web Key Set)
    - Support multiple active keys (for future rotation)
    - Add route to internal/authentication/delivery/http/routes.go
    - Remove old /plans and /subscriptions routes from routes.go
    - _Requirements: 2.5_
    - _Pattern: Follow existing handler pattern with Gin_

  - [x] 1.6 Implement session management with Redis
    - Create pkg/redis package for Redis client (follow existing pkg structure)
    - Create session manager in internal/authentication/usecase/session.go
    - Implement session creation with TTL (8h normal, 7d remember me)
    - Implement session retrieval and deletion (logout)
    - Store session data: user_id, jti, created_at, expires_at
    - Initialize Redis in cmd/api/main.go
    - _Requirements: 1.6, 3.1, 3.2, 3.3_
    - _Pattern: Similar to existing RabbitMQ initialization_

  - [x] 1.7 Update HttpOnly cookie handling
    - Update Login handler to set JWT as HttpOnly cookie (keep existing pattern)
    - Ensure cookie attributes: Secure, SameSite=Lax, Domain from config
    - Update Logout handler to expire cookie (keep existing pattern)
    - Keep existing addSameSiteAttribute helper method
    - Update config/config.go CookieConfig if needed
    - _Requirements: 1.7, 1.8, 3.7_
    - _Pattern: Keep existing cookie handling in handler.go_

  - [x] 1.8 Update authentication endpoints
    - Update GET /auth/me in internal/authentication/delivery/http/handler.go
    - Return current user info from JWT claims (keep existing pattern)
    - Update POST /auth/logout to invalidate session and token
    - Keep existing GET /health endpoint
    - Update internal/authentication/delivery/http/routes.go
    - _Requirements: 3.4, 10.3, 10.4, 10.6_
    - _Pattern: Follow existing handler methods in handler.go_

  - [x] 1.9 Checkpoint - Core auth flow working
    - Manually test OAuth login flow end-to-end
    - Verify JWT tokens are generated correctly
    - Verify sessions are created in Redis
    - Verify cookies are set with correct attributes
    - Ask user if questions arise

- [x] 2. Day 3: Google Groups RBAC + Audit Logging
  - [x] 2.1 Setup Google Directory API integration
    - Create pkg/google package for Directory API client
    - Implement service account credentials loading via Viper
    - Add admin email configuration for domain-wide delegation
    - Test connection to Directory API
    - Initialize in cmd/api/main.go
    - _Requirements: 4.1_
    - _Pattern: Similar to existing pkg structure_

  - [x] 2.2 Implement Google Groups synchronization
    - Create internal/authentication/usecase/groups.go
    - Fetch user groups via Directory API
    - Cache groups in Redis with 5-minute TTL
    - Implement cache-first lookup (check cache before API call)
    - Handle API errors gracefully (use cached data if available)
    - _Requirements: 4.1, 4.2, 4.8_
    - _Pattern: Follow existing usecase pattern_

  - [x] 2.3 Implement role mapping configuration
    - Load role_mapping from auth-config.yaml via Viper
    - Create internal/authentication/usecase/roles.go
    - Implement group-to-role mapping logic
    - Assign highest privilege role for multiple group memberships
    - Default to VIEWER role if no matches
    - _Requirements: 4.3, 4.4, 4.5_
    - _Pattern: Follow existing usecase pattern_

  - [x] 2.4 Integrate groups sync into login flow
    - Update OAuth callback handler to fetch groups
    - Map groups to role using role mapping logic
    - Include role and groups in JWT claims
    - Update user record with assigned role
    - Update internal/authentication/delivery/http/handler.go
    - _Requirements: 4.6, 4.7_
    - _Pattern: Extend existing OAuth handler_

  - [x] 2.5 Setup Kafka for audit logging
    - Create pkg/kafka package for Kafka client
    - Configure Kafka connection via Viper
    - Create audit.events topic (or document manual creation)
    - Implement Kafka producer client
    - Add connection health check
    - Initialize in cmd/api/main.go
    - _Requirements: 5.1_
    - _Pattern: Similar to existing RabbitMQ setup_

  - [x] 2.6 Implement audit event publisher
    - Create internal/audit package with event types
    - Create AuditEvent struct with all required fields
    - Implement async publish to Kafka (non-blocking)
    - Add in-memory buffer for Kafka unavailability
    - Log audit events for: LOGIN, LOGOUT, LOGIN_FAILED, TOKEN_REVOKED
    - Include user_id, action, ip_address, user_agent, metadata
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_
    - _Pattern: Similar to existing RabbitMQ producer pattern_

  - [x] 2.7 Implement audit consumer service
    - Update cmd/consumer/main.go for audit consumer
    - Consume events from audit.events topic
    - Implement batch insert logic (100 messages or 5 seconds)
    - Set expires_at to created_at + 90 days
    - Handle consumer errors and retries
    - Create internal/audit/repository for database operations
    - _Requirements: 5.6, 5.7, 5.8_
    - _Pattern: Follow existing consumer pattern in cmd/consumer/_

  - [x] 2.8 Implement audit log cleanup job
    - Create internal/audit/usecase/cleanup.go
    - Create scheduled job (cron) to delete expired logs
    - Run daily at 2 AM
    - Delete records where expires_at < NOW()
    - Log cleanup statistics
    - Add to cmd/api/main.go or separate worker
    - _Requirements: 5.8_
    - _Pattern: Create new scheduled job_

  - [x] 2.9 Setup Redis for token blacklist
    - Extend existing Redis connection for blacklist
    - Create separate Redis DB for blacklist (db=1)
    - Add connection health check
    - Update pkg/redis package
    - _Requirements: 6.1_
    - _Pattern: Extend existing Redis setup_

  - [x] 2.10 Checkpoint - RBAC and audit logging working
    - Manually test Google Groups fetching and caching
    - Verify roles are mapped correctly
    - Verify audit events are published to Kafka
    - Verify audit consumer batch inserts to database
    - Ask user if questions arise

- [-] 3. Day 4: JWT Middleware Package + Token Blacklist
  - [x] 3.1 Create pkg/auth package structure
    - Create pkg/auth directory
    - Define public interfaces and types
    - Create README with usage examples
    - _Requirements: 7.1_
    - _Pattern: Follow existing pkg structure (similar to pkg/encrypter)_

  - [x] 3.2 Implement JWT verifier component
    - Create pkg/auth/verifier.go
    - Fetch public keys from JWKS endpoint on startup
    - Cache public keys in memory with 1-hour TTL
    - Implement background refresh of public keys
    - Verify JWT signature using cached public key
    - Validate exp, iss, aud claims
    - Extract claims into struct
    - _Requirements: 7.1, 7.2, 7.3, 7.4_
    - _Pattern: Create new pkg component_

  - [x] 3.3 Implement authentication middleware
    - Create pkg/auth/middleware.go
    - Extract JWT from Authorization header or cookie
    - Verify JWT using verifier component
    - Check if jti is in Redis blacklist
    - Inject user claims into Gin context
    - Return 401 if verification fails
    - _Requirements: 7.5, 7.6, 7.7_
    - _Pattern: Similar to existing middleware pattern_

  - [x] 3.4 Implement authorization helpers
    - Create pkg/auth/helpers.go
    - RequireRole(role string) middleware
    - RequireAnyRole(roles ...string) middleware
    - HasPermission(ctx, permission) helper
    - GetUserID(ctx) helper
    - GetUserRole(ctx) helper
    - GetUserGroups(ctx) helper
    - _Requirements: 7.8_
    - _Pattern: Gin middleware pattern_

  - [x] 3.5 Implement token blacklist manager
    - Create internal/authentication/usecase/blacklist.go
    - Add token to blacklist by jti
    - Add all user tokens to blacklist by user_id
    - Check if token is blacklisted
    - Set TTL to remaining token lifetime
    - Use Redis for storage
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_
    - _Pattern: Follow existing usecase pattern_

  - [x] 3.6 Implement internal admin endpoints
    - Create internal/authentication/delivery/http/internal.go
    - Create internal/middleware/service_auth.go for X-Service-Key validation
    - Implement service key decryption using pkg/encrypter
    - Load service_keys from auth-config.yaml via Viper
    - POST /internal/validate - fallback token validation (requires X-Service-Key)
    - POST /internal/revoke-token - revoke specific token or all user tokens (requires X-Service-Key + ADMIN role)
    - GET /internal/users/:id - get user by ID (requires X-Service-Key)
    - Apply service_auth middleware to all /internal/* routes
    - Return 401 if X-Service-Key is missing or invalid
    - Update routes.go
    - _Requirements: 6.6, 10.7, 10.8, 10.10, 10.11, 10.12_
    - _Pattern: Follow existing handler pattern + new middleware_

  - [x] 3.7 Implement audit log query endpoint
    - Create internal/audit/delivery/http package
    - GET /audit-logs with query params (user_id, action, from, to, page, limit)
    - Implement pagination logic in repository
    - Require ADMIN role
    - Return total count and paginated results
    - Add routes to main router
    - _Requirements: 10.9_
    - _Pattern: Follow existing delivery/http pattern_

  - [x] 3.8 Write comprehensive middleware documentation
    - Document installation and setup in pkg/auth/README.md
    - Provide usage examples for each helper
    - Document configuration options
    - Add troubleshooting guide
    - _Requirements: 14.4_

  - [x] 3.9 Checkpoint - Middleware package ready
    - Manually test JWT verification
    - Verify blacklist checking works
    - Verify authorization helpers work
    - Verify documentation is complete
    - Ask user if questions arise

- [x] 4. Day 4: Security Enhancements (Identity Service)
  - [x] 4.1 Implement redirect URL validation
    - Create internal/authentication/usecase/redirect.go
    - Validate redirect URLs against allowed list from config
    - Prevent open redirect attacks
    - Return error for invalid redirect URLs
    - Add to OAuth callback handler
    - _Requirements: 15.4_
    - _Pattern: Follow existing usecase pattern_

  - [x] 4.2 Implement login rate limiting
    - Create internal/authentication/usecase/ratelimit.go
    - Track failed login attempts by IP address
    - Block login attempts after 5 failures in 15 minutes
    - Return 429 Too Many Requests when rate limited
    - Implement exponential backoff
    - Use Redis for storage
    - Apply to /auth/login and /auth/callback endpoints
    - _Requirements: 15.5_
    - _Pattern: Create new usecase component_

  - [x] 4.3 Verify secure JTI generation
    - Review pkg/jwt/jwt.go JTI generation
    - Confirm using crypto/rand via uuid.New()
    - Verify UUID v4 implementation
    - Add comment documenting security requirement
    - _Requirements: 15.7_
    - _Pattern: Code review + documentation_

  - [x] 4.4 Implement configuration validation
    - Update config/config.go validate() function
    - Add validation for all required fields
    - Test missing config fails startup with clear error
    - Test invalid config fails startup with clear error
    - Add validation for format (URLs, email, etc.)
    - _Requirements: 11.3, 11.10_
    - _Pattern: Extend existing validate() function_

  - [x] 4.5 Checkpoint - Security features complete
    - Test redirect URL validation with various URLs
    - Test rate limiting with multiple failed attempts
    - Verify JTI randomness (visual inspection)
    - Test config validation with missing/invalid values
    - Ask user if questions arise

- [ ] 5. Day 5: Comprehensive Documentation
  
  **5.1 Identity Service Documentation**
  
  - [x] 5.1.1 Write Identity Service API documentation
    - Create documents/identity-service-api.md
    - Document all public endpoints with examples:
      * GET /auth/login - OAuth redirect
      * GET /auth/callback - OAuth callback
      * GET /auth/me - Get current user
      * POST /auth/logout - Logout
      * GET /.well-known/jwks.json - Public keys
    - Document OAuth flow with sequence diagram
    - Document JWT structure and claims (iss, aud, sub, email, role, groups, jti, iat, exp)
    - Document error codes and responses
    - Include curl examples for each endpoint
    - _Requirements: 14.1, 14.2, 14.3_

  - [x] 5.1.2 Write Internal API documentation
    - Add to documents/identity-service-api.md
    - Document internal endpoints:
      * POST /internal/validate - Token validation
      * POST /internal/revoke-token - Token revocation
      * GET /internal/users/:id - Get user by ID
    - Document X-Service-Key authentication
    - Document service key generation and encryption
    - Include code examples for calling internal APIs
    - _Requirements: 10.7, 10.8, 10.10, 10.11, 10.12_

  - [x] 5.1.3 Write deployment guide
    - Create documents/identity-service-deployment.md
    - Document Google OAuth setup steps (with screenshots)
    - Document service account creation for Directory API
    - Document environment variables reference (complete list)
    - Document service API keys generation and configuration
    - Document Kubernetes manifests (with examples)
    - Document Docker Compose setup for local development
    - Document database migration steps
    - Document Redis setup (2 DBs: sessions + blacklist)
    - Document Kafka setup (audit.events topic)
    - _Requirements: 14.5, 14.6_

  - [x] 5.1.4 Write troubleshooting guide
    - Create documents/identity-service-troubleshooting.md
    - Document common issues and solutions:
      * OAuth callback errors
      * JWT verification failures
      * Google Directory API connection issues
      * Redis connection issues
      * Kafka connection issues
      * Session management issues
      * Token blacklist issues
    - Document debugging techniques
    - Document log analysis (what to look for)
    - Document health check interpretation
    - Include diagnostic commands (curl, redis-cli, etc.)
    - _Requirements: 14.8_

  **5.2 Service Integration Documentation**
  
  - [ ] 5.2.1 Write service integration guide (Project Service)
    - Create documents/service-integration-guide.md
    - Section 1: Project Service Integration
    - Document step-by-step integration:
      * Add pkg/auth dependency to go.mod
      * Initialize JWT verifier with JWKS endpoint
      * Create middleware instance
      * Apply Authenticate() to all routes
      * Add role-based authorization (RequireRole, RequireAnyRole)
      * Update error handling for 401/403
      * Configure SERVICE_KEY environment variable
      * Add X-Service-Key header for internal API calls
    - Provide complete code examples (copy-paste ready)
    - Document testing steps
    - Document rollback procedure
    - _Requirements: 8.1, 8.5, 8.6, 10.10_

  - [ ] 5.2.2 Write service integration guide (Ingest Service)
    - Add to documents/service-integration-guide.md
    - Section 2: Ingest Service Integration
    - Document step-by-step integration (similar to Project Service)
    - Specific role requirements for data ingestion endpoints
    - Code examples for Ingest Service
    - _Requirements: 8.2, 8.7_

  - [ ] 5.2.3 Write service integration guide (Knowledge Service)
    - Add to documents/service-integration-guide.md
    - Section 3: Knowledge Service Integration
    - Document step-by-step integration
    - Specific role requirements (VIEWER for GET, ANALYST for POST)
    - Code examples for Knowledge Service
    - _Requirements: 8.3, 8.8_

  - [ ] 5.2.4 Write service integration guide (Notification Service)
    - Add to documents/service-integration-guide.md
    - Section 4: Notification Service Integration (WebSocket)
    - Document WebSocket authentication flow
    - Code examples for JWT extraction from query param/cookie
    - Code examples for WebSocket upgrade with auth
    - Document error handling for WebSocket auth failures
    - _Requirements: 8.4_

  - [ ] 5.2.5 Write audit event publishing guide
    - Add to documents/service-integration-guide.md
    - Section 5: Audit Event Publishing
    - Document how to add audit publisher to each service
    - List of audit events to publish:
      * CREATE_PROJECT, DELETE_PROJECT
      * CREATE_SOURCE, DELETE_SOURCE
      * EXPORT_DATA
    - Code examples for publishing audit events
    - Document metadata structure
    - _Requirements: 5.1_

  **5.3 Frontend Integration Documentation**
  
  - [ ] 5.3.1 Write frontend migration guide
    - Create documents/frontend-oauth-migration.md
    - Document changes from email/password to OAuth
    - Section 1: Login Page Migration
      * Remove email/password form
      * Add "Login with Google" button
      * Redirect to /auth/login
      * Handle OAuth callback redirect
      * Remove localStorage token management
    - Section 2: Axios Configuration
      * Ensure withCredentials: true
      * Verify CORS configuration
      * Test cookie handling
    - Section 3: Authentication State Management
      * Call GET /auth/me on app load
      * Store user info in React context/state
      * Redirect to login on 401
      * Display user info in header
    - Section 4: Error Handling
      * Add axios interceptor for 401 (redirect to login)
      * Add axios interceptor for 403 (show permission denied)
    - Section 5: Logout Implementation
      * Call POST /auth/logout
      * Clear user context/state
      * Redirect to login page
    - Section 6: Role-Based UI Rendering
      * Hide/show UI elements based on role
      * Disable buttons for insufficient permissions
      * Show role badge in user profile
    - Provide complete code examples (React/Vue)
    - Document testing steps
    - _Requirements: 12.1-12.8, 14.7_

  - [ ] 5.3.2 Write frontend testing guide
    - Add to documents/frontend-oauth-migration.md
    - Section 7: Testing Checklist
    - Document manual testing steps:
      * Test login flow end-to-end
      * Verify cookies are set correctly
      * Verify API requests include cookies
      * Test logout flow
      * Test 401 redirect to login
      * Test 403 permission denied message
      * Test role-based UI rendering
    - Document automated testing approach
    - _Requirements: 12.1-12.8_

  - [ ] 5.4 Checkpoint - Documentation complete
    - Review all documentation for completeness
    - Verify all code examples are correct
    - Verify all links work
    - Ask user if questions arise

- [ ] 8. Comprehensive Testing Suite (After Code Review & Approval)
  - [ ] 8.1 Unit Tests - Core Authentication
    - Test OAuth callback handler with valid/invalid tokens
    - Test JWT generation with all claim types
    - Test session creation/retrieval/deletion
    - Test cookie setting with various attributes
    - Test /auth/me endpoint with valid/invalid sessions
    - Test logout endpoint
    - Target: >85% coverage for internal/authentication

  - [ ] 8.2 Unit Tests - Google Groups & RBAC
    - Test Directory API client connection
    - Test groups fetching and caching logic
    - Test cache expiration and refresh
    - Test role mapping with various group combinations
    - Test default role assignment
    - Test highest privilege role selection
    - Target: >85% coverage for groups/roles logic

  - [ ] 8.3 Unit Tests - Audit Logging
    - Test audit event creation with all fields
    - Test Kafka producer publish (with mock)
    - Test in-memory buffer behavior
    - Test audit consumer batch processing
    - Test audit log cleanup job
    - Test audit log query endpoint with pagination
    - Target: >85% coverage for internal/audit

  - [ ] 8.4 Unit Tests - JWT Middleware Package
    - Test JWT extraction from header and cookie
    - Test JWT signature verification
    - Test JWT claims validation (exp, iss, aud)
    - Test public key caching and refresh
    - Test blacklist checking
    - Test context injection on success
    - Test 401 response on failure
    - Test RequireRole and RequireAnyRole helpers
    - Test GetUserID, GetUserRole, GetUserGroups helpers
    - Target: >90% coverage for pkg/auth

  - [ ] 8.5 Unit Tests - Token Blacklist
    - Test adding token to blacklist by jti
    - Test adding all user tokens to blacklist
    - Test blacklist TTL calculation
    - Test blacklist enforcement
    - Test Redis connection errors
    - Target: >85% coverage for blacklist logic

  - [ ] 8.6 Unit Tests - Security Features
    - Test redirect URL validation with allowed/blocked URLs
    - Test rate limiting with various attempt patterns
    - Test JTI generation randomness
    - Test private key encryption/decryption
    - Test configuration validation
    - Target: >85% coverage for security features

  - [ ] 8.7 Property-Based Tests - OAuth & JWT
    - **Property 1: OAuth Domain Validation** - Valid domains always succeed, invalid always fail
    - **Property 2: OAuth Blocklist Enforcement** - Blocked emails always rejected
    - **Property 3: Session TTL Correctness** - Sessions expire at correct time
    - **Property 4: Remember Me Duration** - Remember me sessions last 7 days
    - **Property 5: JWT Expiration** - Tokens expire at correct time
    - **Property 6: JWT Signature Verification** - Round trip sign/verify always succeeds
    - **Property 7: JWT Claims Completeness** - All required claims present
    - **Property 8: Cookie Attributes** - HttpOnly, Secure, SameSite always set
    - **Property 9: Cookie Expiration** - Cookie expires with session
    - **Property 10: Logout Invalidation** - Logout always invalidates session
    - **Property 11: Current User Retrieval** - /auth/me returns correct user info
    - _Validates: Requirements 1.1-1.8, 2.1-2.8, 3.1-3.7_
    - _Minimum 100 iterations per property_

  - [ ] 8.8 Property-Based Tests - Google Groups & RBAC
    - **Property 12: Groups Fetching** - Groups always fetched for valid users
    - **Property 13: Groups Caching** - Cached groups returned within TTL
    - **Property 14: Groups Cache Refresh** - Cache refreshed after expiration
    - **Property 15: Role Mapping Logic** - Groups always map to correct role
    - **Property 16: Default Role Assignment** - No group match defaults to VIEWER
    - **Property 17: Highest Privilege Selection** - Multiple groups select highest role
    - **Property 18: Role in JWT Claims** - Role always included in JWT
    - _Validates: Requirements 4.1-4.8_
    - _Minimum 100 iterations per property_

  - [ ] 8.9 Property-Based Tests - Audit Logging
    - **Property 19: Audit Event Publishing** - Events always published to Kafka
    - **Property 20: Audit Event Fields** - All required fields present
    - **Property 21: Audit Consumer Batch** - Batches processed correctly
    - **Property 22: Audit Log Expiration** - Logs expire after 90 days
    - **Property 23: Audit Log Cleanup** - Cleanup deletes expired logs only
    - **Property 24: Audit Log Pagination** - Pagination returns correct results
    - _Validates: Requirements 5.1-5.8_
    - _Minimum 100 iterations per property_

  - [ ] 8.10 Property-Based Tests - Token Blacklist
    - **Property 25: Token Revocation** - Revoked tokens always rejected
    - **Property 26: User Token Revocation** - All user tokens revoked
    - **Property 27: Blacklist TTL** - Blacklist entries expire correctly
    - **Property 28: Blacklist Enforcement** - Blacklisted tokens always fail
    - _Validates: Requirements 6.1-6.6_
    - _Minimum 100 iterations per property_

  - [ ] 8.11 Property-Based Tests - JWT Middleware
    - **Property 29: Public Key Caching** - Keys cached for 1 hour
    - **Property 30: JWT Claims Validation** - Invalid claims always rejected
    - **Property 31: Context Injection** - User claims always in context
    - **Property 32: Unauthorized Response** - Invalid JWT returns 401
    - **Property 33: Internal Token Validation** - /internal/validate works correctly
    - _Validates: Requirements 7.1-7.8, 10.7_
    - _Minimum 100 iterations per property_

  - [ ] 8.12 Property-Based Tests - Service Integration
    - **Property 34: Universal JWT Verification** - All services verify JWT correctly
    - **Property 35: Role-Based Authorization** - Role checks work in all services
    - _Validates: Requirements 8.1-8.8_
    - _Minimum 100 iterations per property_

  - [ ] 8.13 Property-Based Tests - Security Features
    - **Property 36: Redirect URL Validation** - Only allowed URLs accepted
    - **Property 37: Login Rate Limiting** - Rate limit enforced correctly
    - **Property 38: JTI Randomness** - JTI values are unique and random
    - **Property 39: Configuration Validation** - Invalid config always fails
    - _Validates: Requirements 15.4, 15.5, 15.7, 11.3, 11.10_
    - _Minimum 100 iterations per property_

  - [ ] 8.14 Integration Tests - End-to-End Flows
    - Test complete OAuth login flow (browser simulation)
    - Test JWT verification across all 4 services
    - Test role-based access in Project Service
    - Test role-based access in Ingest Service
    - Test role-based access in Knowledge Service
    - Test WebSocket auth in Notification Service
    - Test audit events from all services
    - Test token revocation across services
    - Test logout and session cleanup
    - _Validates: Requirements 8.1-8.8, 12.1-12.8_

  - [ ] 8.15 Integration Tests - Error Scenarios
    - Test expired JWT handling
    - Test revoked JWT handling
    - Test invalid signature handling
    - Test missing claims handling
    - Test blocked user handling
    - Test invalid domain handling
    - Test rate limit exceeded
    - Test Kafka unavailable (audit buffer)
    - Test Redis unavailable (graceful degradation)
    - _Validates: Requirements 10.1-10.10, 15.1-15.10_

  - [ ] 8.16 Performance Tests
    - Benchmark JWT generation (target < 10ms)
    - Benchmark JWT verification (target < 5ms)
    - Benchmark blacklist check (target < 2ms)
    - Benchmark session operations (target < 3ms)
    - Benchmark audit event publish (target < 1ms)
    - Load test with 1000 concurrent users
    - Load test with 10000 requests/second
    - _Validates: Requirements 13.1-13.5_

  - [ ] 8.17 Security Tests
    - Test JWT forgery attempts (should fail)
    - Test token replay attacks (should fail)
    - Test open redirect attacks (should fail)
    - Test brute force login (should be rate limited)
    - Test SQL injection in audit logs (should be sanitized)
    - Test XSS in user data (should be escaped)
    - Run gosec security scanner
    - Run dependency vulnerability scan
    - _Validates: Requirements 15.1-15.10_

  - [ ] 8.18 Test Coverage Analysis
    - Generate coverage report for all packages
    - Ensure >80% overall coverage
    - Ensure >85% coverage for critical paths (auth, jwt, blacklist)
    - Identify and test uncovered edge cases
    - Document coverage gaps and justification

  - [ ] 8.19 Test Documentation
    - Document test strategy and approach
    - Document how to run all tests
    - Document how to run specific test suites
    - Document test data setup and teardown
    - Document CI/CD integration
    - Create documents/testing-guide.md

## Notes

- All implementation tasks (Day 1-7) focus on writing code and documentation
- Testing is consolidated in Section 8 after code review and approval
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation and allow for course correction
- Property tests validate universal correctness properties with randomized inputs (minimum 100 iterations each)
- Unit tests validate specific examples and edge cases
- Integration tests validate end-to-end flows across services
- The 7-day timeline assumes full-time work with AI assistance for code generation
- Greenfield project means no data migration or backward compatibility concerns
- HttpOnly cookie approach is maintained from Identity Service (no breaking changes for frontend)
- Service name remains "identity-service" (not renamed to "auth-service")
- Framework: Gin (existing), NOT Chi router
- Config: Viper (new), NOT env tags
- Code structure: Follow existing internal/domain/delivery/usecase/repository pattern

