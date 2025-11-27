# Implementation Tasks

## 1. Configuration Layer

- [x] 1.1 Add cookie configuration struct to `config/config.go`
  - Cookie domain
  - Secure flag (default: true for production)
  - SameSite policy (default: Lax)
  - Max-Age for normal login (default: 7200 seconds / 2 hours)
  - Max-Age for "Remember Me" (default: 2592000 seconds / 30 days)
  - Cookie name (default: "smap_auth_token")
- [x] 1.2 Add environment variables to `template.env`
  - COOKIE_DOMAIN
  - COOKIE_SECURE
  - COOKIE_SAMESITE
  - COOKIE_MAX_AGE
  - COOKIE_MAX_AGE_REMEMBER
  - COOKIE_NAME

## 2. Authentication Domain Updates

- [x] 2.1 Update `internal/authentication/interface.go`
  - Add `Logout` method to UseCase interface
  - Add `GetCurrentUser` method to UseCase interface
- [x] 2.2 Update `internal/authentication/type.go`
  - Add `LogoutInput` and `LogoutOutput` types
  - Add `GetCurrentUserOutput` type
- [x] 2.3 Implement logout usecase in `internal/authentication/usecase/authentication.go`
  - Validate user context exists
  - Return success (actual cookie expiry handled in delivery layer)
- [x] 2.4 Implement get-current-user usecase
  - Extract user ID from scope context
  - Fetch user details from user repository
  - Return user information
- [x] 2.5 Update login handler in `internal/authentication/delivery/http/handler.go`
  - Extract cookie configuration from handler dependencies
  - After successful login, set HttpOnly cookie with JWT
  - Calculate Max-Age based on "Remember Me" flag
  - Keep JSON response but remove token field (backward compatibility transition)
  - Added helper method to set SameSite attribute (Gin doesn't support it directly)
- [x] 2.6 Add logout handler
  - Expire authentication cookie (set Max-Age: -1)
  - Return success response
- [x] 2.7 Add get-me handler
  - Extract user from scope context (set by Auth middleware)
  - Return user details (id, email, full_name, role)
- [x] 2.8 Update `internal/authentication/delivery/http/presenter.go`
  - Remove `Token` field from `loginResp` struct
  - Create `getMeResp` struct for /me endpoint response
- [x] 2.9 Update `internal/authentication/delivery/http/routes.go`
  - Add POST `/authentication/logout` route (requires Auth middleware)
  - Add GET `/authentication/me` route (requires Auth middleware)

## 3. Middleware Updates

- [ ] 3.1 Update `internal/middleware/middleware.go` Auth function
  - First, attempt to read token from cookie
  - Fallback to Authorization header for backward compatibility (optional, remove after full migration)
  - Extract cookie name from middleware configuration
  - Maintain existing JWT verification logic
- [ ] 3.2 Update `internal/middleware/new.go`
  - Accept cookie configuration in constructor
  - Store cookie config in Middleware struct
- [ ] 3.3 Update `internal/middleware/cors.go`
  - Set `AllowCredentials: true` in DefaultCORSConfig
  - Update documentation to require specific origins (not wildcard "*")
  - Add warning comment about security implications

## 4. HTTP Server Integration

- [ ] 4.1 Update `internal/httpserver/new.go`
  - Accept cookie configuration in Config struct
  - Pass cookie config to authentication handler
  - Pass cookie config to middleware
- [ ] 4.2 Update `internal/httpserver/handler.go`
  - Ensure logout and /me routes are registered
  - Verify Auth middleware is applied correctly
- [ ] 4.3 Update `cmd/api/main.go`
  - Load cookie configuration from environment
  - Pass cookie config to HTTP server

## 5. Documentation & Testing

- [ ] 5.1 Update Swagger annotations
  - Document Set-Cookie header in login endpoint
  - Document /logout endpoint
  - Document /me endpoint
  - Add note about credentials requirement in API docs
- [ ] 5.2 Update README.md
  - Document breaking changes
  - Add migration guide for frontend clients
  - Document cookie configuration options
- [ ] 5.3 Add integration tests
  - Test login sets cookie correctly
  - Test authenticated requests read cookie
  - Test logout expires cookie
  - Test /me endpoint returns user info
  - Test fallback to Authorization header (if keeping backward compatibility)
- [ ] 5.4 Update document/api.md
  - Document new authentication flow
  - Update sequence diagrams for cookie-based auth
  - Document CORS requirements

## 6. Migration & Rollout

- [ ] 6.1 Create migration guide document
  - Frontend code examples (axios, fetch)
  - Mobile app considerations
  - Testing procedures
- [ ] 6.2 Remove Authorization header fallback (future phase)
  - Only after confirming all clients migrated
  - Schedule deprecation timeline