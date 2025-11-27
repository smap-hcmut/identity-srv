# Change: Migrate to HttpOnly Cookie Authentication

## Why

The current authentication system stores JWT tokens in LocalStorage and transmits them via the `Authorization: Bearer` header. This approach exposes a critical XSS (Cross-Site Scripting) vulnerability: any malicious JavaScript code (from compromised third-party libraries, analytics scripts, or injected code) can access `localStorage.getItem()` and exfiltrate user tokens.

**Security Impact:**
- XSS attacks can steal authentication tokens
- Compromised user sessions lead to unauthorized access
- Non-compliance with OWASP security best practices
- Risk of large-scale credential theft if third-party dependencies are compromised

This migration to HttpOnly Cookies eliminates JavaScript access to authentication tokens while maintaining the stateless JWT architecture for horizontal scalability.

## What Changes

1. **Login Response Structure** - Remove token from JSON response body
2. **Cookie-Based Token Delivery** - Set JWT as HttpOnly cookie in `Set-Cookie` header on successful login
3. **Authentication Middleware** - Read token from cookie instead of Authorization header
4. **CORS Configuration** - Enable `Access-Control-Allow-Credentials: true` and configure allowed origins
5. **Logout Endpoint** - Add new endpoint to expire authentication cookie
6. **Get Current User Endpoint** - Add `/authentication/me` endpoint to retrieve user info from token context

**Breaking Changes:**
- **BREAKING**: Login response no longer includes `token` field in JSON body
- **BREAKING**: Frontend must use `credentials: 'include'` (fetch) or `withCredentials: true` (axios)
- **BREAKING**: CORS configuration now requires specific origins (no wildcard `*` with credentials)

**Cookie Security Attributes:**
- `HttpOnly: true` - Prevents JavaScript access
- `Secure: true` - HTTPS-only transmission
- `SameSite: Lax` - CSRF protection (allow same-site navigation)
- `Path: /identity` - Scope to API base path
- `Max-Age: 7200` (2 hours default) or `2592000` (30 days with "Remember Me")
- `Domain: .smap.com` (configurable for subdomain sharing)

## Impact

### Affected Capabilities
- **Authentication** (`internal/authentication/`) - Login/logout flow changes
- **Middleware** (`internal/middleware/`) - Token extraction logic changes
- **CORS Configuration** - Credentials and origin handling

### Affected Code
- `internal/authentication/delivery/http/handler.go` - Login handler response
- `internal/authentication/delivery/http/presenter.go` - Remove token from loginResp
- `internal/authentication/usecase/authentication.go` - Add logout and get-me usecases
- `internal/middleware/middleware.go` - Auth middleware token extraction
- `internal/middleware/cors.go` - Enable credentials support
- `internal/httpserver/handler.go` - Add logout and /me routes
- `config/config.go` - Add cookie configuration

### Frontend Impact
All frontend clients must update to:
1. Set `withCredentials: true` in HTTP client configuration
2. Remove manual token storage (localStorage/sessionStorage)
3. Remove manual Authorization header injection
4. Call `/authentication/logout` endpoint instead of clearing localStorage
5. Use `/authentication/me` endpoint to check authentication status

### Mobile App Compatibility
- Mobile HTTP clients typically support cookie jars (URLSession, OkHttp, etc.)
- Alternatively, provide parallel Bearer token support for mobile (detect via User-Agent)
- Consider maintaining dual authentication modes during transition period

### Testing Impact
- Postman/Insomnia testing requires cookie management or browser interceptor
- Integration tests need to handle cookie extraction and storage
- E2E tests must enable credentials in HTTP client configuration

