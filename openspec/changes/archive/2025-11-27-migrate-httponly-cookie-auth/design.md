# Technical Design: HttpOnly Cookie Authentication

## Context

The SMAP Identity Service currently uses JWT tokens stored in browser LocalStorage and transmitted via `Authorization: Bearer` headers. This is a common pattern for SPAs but has a critical security vulnerability: any JavaScript code (including third-party scripts) can access `localStorage.getItem()` and steal authentication tokens.

This design document outlines the migration to HttpOnly Cookie-based JWT storage while maintaining:
- Stateless JWT architecture (no server-side session database)
- Horizontal scalability for microservices
- Backward compatibility during transition
- Support for multiple client types (web, mobile)

## Goals / Non-Goals

### Goals
- **Eliminate XSS token theft**: HttpOnly cookies cannot be accessed by JavaScript
- **Maintain statelessness**: Continue using JWT (not server-side sessions)
- **OWASP compliance**: Follow modern security best practices
- **Minimize frontend changes**: Browser handles cookies automatically
- **Support CORS properly**: Enable credentials with explicit origin configuration

### Non-Goals
- Not implementing refresh tokens (out of scope for this change)
- Not implementing CSRF token system (SameSite provides adequate protection)
- Not changing JWT structure or claims
- Not implementing session database (staying stateless)
- Not removing Bearer token support immediately (phased deprecation)

## Decisions

### 1. Cookie Attributes Configuration

**Decision**: Use strict security attributes for production environments.

```go
type CookieConfig struct {
    Name     string // Default: "smap_auth_token"
    Domain   string // Default: ".smap.com" (allows subdomain sharing)
    Secure   bool   // Default: true (HTTPS only)
    SameSite string // Default: "Lax" (balance security vs usability)
    MaxAge   int    // Default: 7200 seconds (2 hours)
    MaxAgeRemember int // Default: 2592000 seconds (30 days)
    Path     string // Default: "/identity"
}
```

**Rationale**:
- `HttpOnly: true` - Critical for XSS prevention (always true, not configurable)
- `Secure: true` - Prevent token leakage over HTTP (configurable for local dev)
- `SameSite: Lax` - Balance between security and UX
  - `Strict` would break cross-site navigation (e.g., email links)
  - `Lax` allows top-level navigation while blocking CSRF
  - `None` would require explicit CSRF protection
- `Domain: .smap.com` - Allows cookie sharing across `api.smap.com`, `web.smap.com`
- `Path: /identity` - Scope to API base path to avoid unnecessary cookie transmission

**Alternatives Considered**:
- `SameSite: Strict` - Too restrictive, breaks legitimate use cases
- `SameSite: None` - Requires additional CSRF protection
- No Domain attribute - Would not share across subdomains

### 2. Token Extraction Strategy

**Decision**: Implement cookie-first strategy with optional Authorization header fallback.

```go
func (m Middleware) Auth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Try cookie first
        tokenString, err := c.Cookie(m.cookieName)
        
        // Fallback to Authorization header (for backward compatibility)
        if err != nil || tokenString == "" {
            tokenString = strings.ReplaceAll(c.GetHeader("Authorization"), "Bearer ", "")
        }
        
        if tokenString == "" {
            response.Unauthorized(c)
            c.Abort()
            return
        }
        
        // ... existing verification logic ...
    }
}
```

**Rationale**:
- Cookie-first ensures new clients use secure method
- Fallback maintains backward compatibility during migration
- Clear deprecation path: remove fallback after full migration

**Alternatives Considered**:
- Cookie-only (no fallback) - Too disruptive, forces big-bang migration
- Header-only (no change) - Doesn't address security vulnerability
- Dual-token (cookie + header validation) - Unnecessary complexity

### 3. CORS Configuration Changes

**Decision**: Enable credentials and require explicit origin whitelist.

**Before**:
```go
AllowedOrigins: []string{"*"}
AllowCredentials: false
```

**After**:
```go
AllowedOrigins: []string{
    "https://web.smap.com",
    "https://app.smap.com",
    "http://localhost:3000", // dev only
}
AllowCredentials: true
```

**Rationale**:
- Browsers block `Access-Control-Allow-Origin: *` with credentials enabled
- Explicit whitelist improves security posture
- Configuration via environment variables allows per-environment customization

**Migration Path**:
1. Add CORS_ALLOWED_ORIGINS environment variable
2. Parse comma-separated list of origins
3. Validate no wildcard when credentials enabled
4. Log warning if misconfigured

### 4. Login Response Structure

**Decision**: Remove token from JSON response body, rely solely on Set-Cookie header.

**Before**:
```json
{
  "code": 200,
  "data": {
    "user": { "id": "...", "email": "...", "role": "..." },
    "token": "eyJhbGc..."
  }
}
```

**After**:
```json
{
  "code": 200,
  "data": {
    "user": { "id": "...", "email": "...", "role": "..." }
  }
}
```
Plus HTTP header: `Set-Cookie: smap_auth_token=eyJhbGc...; HttpOnly; Secure; SameSite=Lax; ...`

**Rationale**:
- Removes token from client-accessible response
- Forces proper cookie usage
- Simplifies frontend code (no manual storage)

**Alternatives Considered**:
- Keep token in response temporarily - Defeats security purpose
- Return both cookie and token - Confusing, encourages wrong usage

### 5. Logout Mechanism

**Decision**: Add `/authentication/logout` endpoint that expires the cookie.

```go
func (h handler) Logout(c *gin.Context) {
    c.SetCookie(
        h.cookieName,
        "",                  // Empty value
        -1,                  // MaxAge: -1 (expire immediately)
        h.cookiePath,
        h.cookieDomain,
        h.cookieSecure,
        true,                // HttpOnly
        http.SameSiteLaxMode,
    )
    response.OK(c, nil)
}
```

**Rationale**:
- Server-controlled logout ensures cookie properly expired
- Cannot rely on client-side clearing (HttpOnly prevents JavaScript access)
- Stateless: no server-side session to invalidate

**Note**: For immediate invalidation across all instances, consider adding JWT blacklist (future enhancement, out of scope).

### 6. Current User Endpoint

**Decision**: Add `/authentication/me` endpoint to retrieve user information.

**Rationale**:
- Frontend can no longer decode JWT client-side (HttpOnly cookie)
- Need server endpoint to check authentication status
- Provides user profile without parsing token

```go
GET /identity/authentication/me
Authorization: Cookie (automatic)

Response:
{
  "code": 200,
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "full_name": "John Doe",
    "role": "USER"
  }
}
```

**Alternatives Considered**:
- Embed user info in every API response - Wasteful bandwidth
- Use separate /users/me endpoint - Authentication concern, belongs in auth domain

### 7. Mobile App Support

**Decision**: Support both cookie and Bearer token authentication (detect via User-Agent or endpoint).

**Option A - User-Agent Detection** (Recommended for gradual migration):
```go
func (m Middleware) Auth() gin.HandlerFunc {
    return func(c *gin.Context) {
        userAgent := c.GetHeader("User-Agent")
        
        // Mobile apps use Bearer token
        if strings.Contains(userAgent, "SMAPMobile") {
            tokenString = c.GetHeader("Authorization")
        } else {
            // Web clients use cookie
            tokenString, _ = c.Cookie(m.cookieName)
        }
        // ... validation ...
    }
}
```

**Option B - Separate Endpoints** (Clean separation):
- `/authentication/login` - Sets cookie (web clients)
- `/authentication/mobile/login` - Returns token in JSON (mobile clients)

**Decision Rationale**:
- Option A chosen for simplicity and unified codebase
- Mobile HTTP clients can support cookies but may have limitations
- Allows per-client migration timeline
- Can remove mobile fallback when apps updated

**Alternatives Considered**:
- Force mobile apps to use cookies - May break existing apps
- Maintain two separate auth systems - Code duplication, drift risk

## Risks / Trade-offs

### Risk 1: CORS Misconfiguration

**Risk**: If CORS configured incorrectly, cookies won't be sent, breaking authentication.

**Mitigation**:
- Validate CORS config on startup (fail-fast if invalid)
- Log warnings for common mistakes (wildcard with credentials)
- Provide clear error messages with troubleshooting steps
- Document CORS requirements prominently in README

**Indicators**:
- 401 Unauthorized despite valid cookie
- Browser console: "CORS policy blocked"
- Missing `Access-Control-Allow-Credentials: true` in response

### Risk 2: Cross-Domain Cookie Limitations

**Risk**: If frontend and API on different domains (not subdomains), cookies won't work.

**Example Problem**:
- Frontend: `https://myapp.com`
- API: `https://api-smap.com`
- Cannot share cookies across different domains

**Mitigation**:
- Document domain architecture requirements
- Recommend subdomain architecture: `app.smap.com` + `api.smap.com`
- For different domains, keep Bearer token option available
- Consider API gateway/proxy pattern to unify domains

### Risk 3: Mobile Cookie Jar Complexity

**Risk**: Mobile HTTP clients may not handle cookies automatically or persistently.

**Mitigation**:
- Provide clear mobile integration guide
- Test on iOS (URLSession) and Android (OkHttp)
- Offer Bearer token fallback for mobile
- Document cookie jar configuration for popular mobile HTTP libraries

### Risk 4: Testing Complexity

**Risk**: Manual API testing (Postman, curl) becomes more complex with cookies.

**Mitigation**:
- Provide Postman collection with cookie handling scripts
- Document curl commands with cookie jar: `curl -c cookies.txt -b cookies.txt`
- Create test utilities to extract cookies
- Update integration test fixtures

### Risk 5: Logout Doesn't Invalidate Existing Tokens

**Risk**: Stateless JWT means logout only expires cookie; token still valid if extracted.

**Mitigation**:
- Document that logout is client-side only (cookie removal)
- For critical apps, implement token blacklist (future enhancement)
- Keep JWT expiration short (2 hours default)
- Use "Remember Me" only with user understanding

**Future Enhancement**: JWT blacklist using Redis with expiration matching token lifetime.

## Migration Plan

### Phase 1: Deploy with Dual Support (Week 1)
- Backend supports both cookie and Authorization header
- Frontend continues using Authorization header
- No breaking changes yet
- Monitor logs for cookie usage

### Phase 2: Frontend Migration (Week 2-3)
- Update web frontend to use cookies
- Set `withCredentials: true` in API client
- Remove localStorage token management
- Update to call `/logout` endpoint
- Use `/me` endpoint for user info

### Phase 3: Mobile App Update (Week 4-6)
- Update mobile apps to handle cookies OR continue using Bearer token
- Test cookie persistence across app restarts
- Update login flow documentation

### Phase 4: Deprecate Authorization Header (Week 8+)
- Remove Authorization header fallback from middleware
- Fully enforce cookie-based authentication
- Monitor for any remaining clients using old method

### Rollback Plan
If critical issues discovered:
1. Re-enable Authorization header primary path
2. Make cookies optional
3. Investigate and fix issues
4. Retry migration when resolved

## Open Questions

1. **Should we implement refresh tokens alongside this change?**
   - **Answer**: No, out of scope. Address in separate proposal if needed.

2. **What about API clients (machine-to-machine)?**
   - **Answer**: Use separate authentication method (API keys) for M2M, not cookies.

3. **Should we implement CSRF tokens since we're using cookies?**
   - **Answer**: `SameSite=Lax` provides adequate CSRF protection for our use case. Explicit CSRF tokens only needed if using `SameSite=None`.

4. **How to handle concurrent sessions (multiple devices)?**
   - **Answer**: Each device gets its own cookie/token. JWT payload already stateless. No server-side session management needed.

5. **What about token revocation for account deletion/password reset?**
   - **Answer**: Current limitation of stateless JWT. Future enhancement: implement token blacklist using Redis. For now, rely on short expiration times.

