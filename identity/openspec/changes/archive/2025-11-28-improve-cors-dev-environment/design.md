# Design Document: Environment-Based CORS Configuration

## Context

The SMAP Identity Service uses HttpOnly cookies for authentication, which requires `AllowCredentials: true` in CORS configuration. Browser security policies mandate that when credentials are allowed, CORS cannot use wildcard origins (`*`) - only specific origins are permitted.

The current implementation hardcodes a list of allowed origins in `internal/middleware/cors.go`:
```go
AllowedOrigins: []string{
    "http://localhost",
    "http://127.0.0.1",
    "http://localhost:3000",
    "http://127.0.0.1:3000",
    "https://smap.tantai.dev",
    "https://smap-api.tantai.dev",
}
```

**Problem**: Developers working remotely via VPN cannot access the API from private subnets (172.16.21.0/24, 172.16.19.0/24, 192.168.1.0/24) because their device IPs are not in the hardcoded list. Adding all possible IPs (254 IPs × 3 subnets = 762 entries) is unmaintainable.

**Stakeholders**:
- Development team (primary beneficiaries - improved workflow)
- DevOps team (infrastructure configuration)
- Security team (production security concerns)
- End users (no direct impact, but must not introduce vulnerabilities)

**Constraints**:
- Must maintain `AllowCredentials: true` for HttpOnly cookies
- Cannot use wildcard origins with credentials (browser restriction)
- Must not compromise production security
- Should maintain backward compatibility for existing deployments
- Must work in containerized K8s environments

## Goals / Non-Goals

### Goals
1. **Enable VPN development access**: Developers can access API from any IP in configured private subnets
2. **Maintain production security**: Production environment uses strict origin validation only
3. **Zero configuration per developer**: No need to add individual IPs to config
4. **Environment awareness**: Automatically adapt CORS based on deployment environment
5. **Maintainability**: Easy to add/remove subnets without code changes
6. **Backward compatibility**: Existing deployments continue working (default to production mode)

### Non-Goals
1. **Dynamic subnet configuration**: Subnets defined in code, not environment variables (out of scope for MVP)
2. **Per-request origin learning**: No automatic origin discovery or allowlisting
3. **Database-driven configuration**: Keep configuration simple and code-based
4. **Fine-grained RBAC**: All allowed origins have same access (existing RBAC applies post-CORS)
5. **Public cloud VPN**: Focus on private subnet ranges, not public cloud VPN IPs

## Decisions

### Decision 1: Use Environment Variable for Mode Selection

**Choice**: Add `ENV` environment variable with values: `production`, `staging`, `dev`

**Rationale**:
- Standard practice for environment differentiation in containerized apps
- Already using environment variables extensively (12-factor app pattern)
- Easy to configure in Kubernetes ConfigMaps
- Clear and explicit (no guessing which mode is active)
- Fail-safe default: empty/missing ENV → production mode

**Alternatives Considered**:
1. **Detect environment by hostname/domain** - Rejected: fragile, harder to test locally, requires DNS setup
2. **Separate CORS config files** - Rejected: adds complexity, harder to maintain, still needs environment detection
3. **Feature flag system** - Rejected: overkill for single boolean decision, adds dependency

**Implementation**:
```go
type Config struct {
    Environment string `env:"ENV" envDefault:"production"`
}
```

### Decision 2: Use AllowOriginFunc for Dynamic Validation

**Choice**: Set `AllowOriginFunc func(origin string) bool` in non-production environments instead of `AllowedOrigins` list

**Rationale**:
- Gin's CORS middleware already supports `AllowOriginFunc` for dynamic validation
- Allows runtime decision based on origin properties (IP subnet, localhost, etc.)
- More flexible than static list for development scenarios
- Standard Go pattern for validation callbacks
- Zero performance impact (microsecond-scale CIDR checks)

**Alternatives Considered**:
1. **Generate 762 static IPs in config** - Rejected: unmaintainable, config bloat, still misses dynamic IPs
2. **Custom CORS middleware** - Rejected: reinventing the wheel, more code to maintain, higher bug risk
3. **Proxy/gateway with CORS** - Rejected: adds infrastructure complexity, latency, single point of failure

**Implementation**:
```go
config.AllowOriginFunc = func(origin string) bool {
    // Check production domains
    // Check localhost
    // Check private subnets
    return false
}
```

### Decision 3: CIDR-Based Subnet Matching

**Choice**: Use Go's `net` package to parse CIDR ranges and check IP membership

**Rationale**:
- Standard library support (no external dependencies)
- Proven, battle-tested implementation
- Efficient: O(1) IP-in-subnet check using bit masking
- Handles edge cases (invalid IPs, malformed CIDRs)
- Clear semantics: "172.16.21.0/24" is self-documenting

**Alternatives Considered**:
1. **IP range (start-end)** - Rejected: less flexible, harder to express, not standard notation
2. **Regex matching** - Rejected: error-prone, slower, doesn't understand IP semantics
3. **External CIDR library** - Rejected: unnecessary dependency, stdlib sufficient

**Implementation**:
```go
func isPrivateOrigin(origin string) bool {
    u, _ := url.Parse(origin)
    ip := net.ParseIP(u.Hostname())
    _, subnet, _ := net.ParseCIDR("172.16.21.0/24")
    return subnet.Contains(ip)
}
```

### Decision 4: Hardcoded Subnet List (MVP)

**Choice**: Define private subnets as Go constants in code, not environment variables

**Rationale**:
- MVP simplicity: subnets rarely change
- Version-controlled: subnet changes go through code review
- Type-safe: compile-time validation
- Easy to extend later: can add ENV support if needed
- Current network topology is stable (3 subnets)

**Alternatives Considered**:
1. **ENV variable with comma-separated CIDRs** - Rejected for MVP: string parsing complexity, error-prone, harder to validate
2. **JSON config file** - Rejected: adds file management, deployment complexity
3. **Database configuration** - Rejected: overkill, adds dependency, slower

**Future Extension Path**:
If subnets need frequent changes, add optional ENV override:
```go
subnets := getEnvOrDefault("CORS_PRIVATE_SUBNETS", defaultSubnets)
```

### Decision 5: Fail-Safe Default to Production

**Choice**: If ENV is empty/missing/invalid, default to production mode

**Rationale**:
- Security-first principle: when in doubt, be strict
- Prevents accidental exposure from misconfiguration
- Explicit opt-in to permissive mode
- Aligns with least-privilege principle
- Existing deployments without ENV continue working securely

**Implementation**:
```go
env := os.Getenv("ENV")
if env == "" {
    env = "production"
}
```

## Architecture

### Component Diagram

```
┌─────────────────────────────────────────────────┐
│              HTTP Server (Gin)                  │
│  ┌───────────────────────────────────────────┐  │
│  │         CORS Middleware                   │  │
│  │  ┌─────────────────────────────────────┐  │  │
│  │  │    DefaultCORSConfig(env)           │  │  │
│  │  │  ┌───────────────────────────────┐  │  │  │
│  │  │  │  if env == "production"       │  │  │  │
│  │  │  │    return static origins      │  │  │  │
│  │  │  │  else                         │  │  │  │
│  │  │  │    return AllowOriginFunc     │  │  │  │
│  │  │  └───────────────────────────────┘  │  │  │
│  │  └─────────────────────────────────────┘  │  │
│  │               │                            │  │
│  │               ▼                            │  │
│  │  ┌─────────────────────────────────────┐  │  │
│  │  │      AllowOriginFunc(origin)        │  │  │
│  │  │  ┌───────────────────────────────┐  │  │  │
│  │  │  │ isProductionOrigin(origin)?   │  │  │  │
│  │  │  │        ↓ yes                  │  │  │  │
│  │  │  │ isLocalhostOrigin(origin)?    │  │  │  │
│  │  │  │        ↓ yes                  │  │  │  │
│  │  │  │ isPrivateOrigin(origin)?      │──┼──┼──┼──> net.ParseIP()
│  │  │  │        ↓ yes                  │  │  │  │    net.ParseCIDR()
│  │  │  │ return true/false             │  │  │  │    subnet.Contains()
│  │  │  └───────────────────────────────┘  │  │  │
│  │  └─────────────────────────────────────┘  │  │
│  └───────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
```

### Data Flow

**Production Mode**:
1. Server starts → reads ENV=production
2. DefaultCORSConfig() → returns static origin list
3. Request arrives → CORS middleware checks origin against list
4. Match → allow / No match → block

**Development Mode**:
1. Server starts → reads ENV=dev
2. DefaultCORSConfig() → returns config with AllowOriginFunc
3. Request arrives → CORS middleware calls AllowOriginFunc(origin)
4. AllowOriginFunc checks in order:
   - Is production domain? → allow
   - Is localhost? → allow
   - Is private subnet IP? → parse IP → check CIDR → allow/block
5. Return result to middleware

### Code Structure

```
internal/middleware/cors.go
├── CORSConfig struct (existing, add AllowOriginFunc field)
├── DefaultCORSConfig(env string) (refactored)
├── CORS(config) gin.HandlerFunc (updated to check AllowOriginFunc)
├── isOriginAllowed(origin, list) (existing, unchanged)
├── isPrivateOrigin(origin) bool (new)
├── isLocalhostOrigin(origin) bool (new)
└── constants: privateSubnets, productionOrigins (new)

config/config.go
└── Config.Environment string (new)

internal/httpserver/handler.go
└── Initialize CORS with config.Environment (updated)
```

## Security Analysis

### Threat Model

| Threat | Risk | Mitigation |
|--------|------|------------|
| **Permissive CORS in production** | HIGH | Default to production mode, require explicit ENV=dev |
| **CSRF attacks from dev environment** | MEDIUM | Dev environments not internet-accessible, VPN-only |
| **IP spoofing in private subnets** | LOW | Private subnets behind VPN, not routable from internet |
| **ENV misconfiguration** | MEDIUM | Startup logging, fail-safe default, validation tests |
| **CIDR parsing vulnerabilities** | LOW | Use stdlib (battle-tested), handle errors gracefully |
| **Localhost as production origin** | LOW | Localhost rejected in production mode |
| **Developer machine compromise** | LOW | Existing risk, CORS doesn't add attack surface (auth still required) |

### Security Guarantees

**Production Environment**:
- ✅ Only explicit origins allowed
- ✅ No dynamic validation
- ✅ No localhost allowed
- ✅ No private subnets allowed
- ✅ Same security posture as current implementation

**Non-Production Environment**:
- ✅ Still requires valid authentication (HttpOnly cookie with JWT)
- ✅ CORS is defense-in-depth, not primary security layer
- ✅ Private subnets not internet-routable (VPN-only)
- ✅ Development environments not exposed to public internet

**Defense-in-Depth**:
CORS is layer 1 (browser enforcement). Additional layers:
1. **Authentication**: JWT validation (layer 2)
2. **Authorization**: Role-based access control (layer 3)
3. **Network**: VPN required for private subnets (layer 4)
4. **Audit**: Request logging (layer 5)

## Risks / Trade-offs

### Risk 1: Misconfiguration Enables Permissive CORS in Production

**Likelihood**: Low
**Impact**: High (production security breach)

**Mitigation**:
- Default ENV to "production" if not set
- Add startup log: "CORS mode: production (strict origins only)"
- Add validation tests that fail if production mode allows localhost
- Document ENV in README with security warnings
- Code review checklist includes ENV verification

**Detection**:
- Monitor CORS rejections in production (should be rare)
- Security audit of deployed configs
- Automated tests in CI/CD

### Risk 2: Developer Confusion About Environment Modes

**Likelihood**: Medium
**Impact**: Low (dev workflow disruption, easy to fix)

**Mitigation**:
- Clear documentation in README
- Startup logs indicate mode
- Error messages hint at ENV configuration when CORS blocks in dev
- Example .env files for each environment

### Risk 3: Performance Overhead from CIDR Checking

**Likelihood**: Low
**Impact**: Low (negligible latency)

**Analysis**:
- CIDR check is O(1) bit operation (~microseconds)
- Only runs in non-production environments
- Only runs on requests from private subnets (subset of traffic)
- AllowOriginFunc short-circuits on production domain match

**Measurement**:
- Benchmark CIDR check: expect <1μs
- Load test dev environment: expect <1ms overhead

### Risk 4: Private Subnet IPs Exposed in Logs

**Likelihood**: Medium
**Impact**: Low (internal IPs, not publicly routable)

**Mitigation**:
- Do not log full origin URLs in CORS middleware
- If logging needed, use log level DEBUG only
- Private IPs already in VPN DNS/DHCP logs

## Migration Plan

### Phase 1: Development (Day 1)
1. Implement changes in feature branch
2. Add unit tests (100% coverage for new functions)
3. Add integration tests (CORS scenarios)
4. Test locally with ENV=production → verify strict mode
5. Test locally with ENV=dev → verify permissive mode

### Phase 2: Staging Deployment (Day 2-3)
1. Deploy to staging cluster
2. Set ENV=staging in ConfigMap
3. Verify dev team can access from VPN subnets
4. Verify HttpOnly cookies still work
5. Verify CORS headers in browser DevTools
6. Monitor logs for 24 hours

### Phase 3: Production Deployment (Day 4)
1. Set ENV=production in production ConfigMap
2. Deploy to production
3. Verify no regression (existing origins still work)
4. Verify private subnets blocked in production
5. Monitor logs for 48 hours
6. Check error rates (should be unchanged)

### Phase 4: Monitoring (Day 5-7)
1. Track CORS rejection metrics
2. Collect developer feedback
3. Monitor security alerts
4. Document lessons learned

### Rollback Plan

**If regression detected in production**:
1. Immediate: Set ENV=production (if not set) → forces strict mode
2. Or: Redeploy previous version (standard k8s rollback)
3. Rollback takes <5 minutes

**Rollback is safe because**:
- Change is additive (new ENV variable)
- Default behavior is secure (production mode)
- No database migrations
- No breaking API changes

## Performance Considerations

### CIDR Check Performance

**Benchmark target**: <1μs per check

**Implementation**: Go's `net.IPNet.Contains()` uses bit masking:
```go
// Pseudocode of net.Contains()
func (n *IPNet) Contains(ip IP) bool {
    return ip & mask == network
}
```

**Optimization**:
- Pre-parse CIDRs at startup (not per-request)
- Short-circuit on production domain match
- Localhost check is string prefix (faster than IP parse)
- Only 3 CIDR checks maximum per request

### Expected Overhead

| Scenario | Overhead | Frequency |
|----------|----------|-----------|
| Production domain | 0μs (static list) | 99% of production traffic |
| Localhost in dev | <0.1μs (string prefix) | 80% of dev traffic |
| Private subnet in dev | <1μs (CIDR check) | 20% of dev traffic |
| Rejected origin | <1μs (full validation) | <1% of traffic |

**Conclusion**: Negligible performance impact (<0.01% latency increase)

## Testing Strategy

### Unit Tests

**Test file**: `internal/middleware/cors_test.go`

**Coverage targets**:
- `isPrivateOrigin()` → 100% (8 test cases)
- `isLocalhostOrigin()` → 100% (6 test cases)
- `DefaultCORSConfig()` → 100% (4 test cases)
- `AllowOriginFunc` → 100% (6 test cases)

**Key scenarios**:
- Valid IPs in each subnet (3 tests)
- IPs outside subnets (3 tests)
- Boundary IPs (first/last in range) (6 tests)
- Invalid formats (malformed IPs, invalid CIDRs) (4 tests)
- Localhost variants (HTTP/HTTPS, with/without port) (6 tests)
- Production mode rejects private subnets (1 test)
- Dev mode allows private subnets (1 test)

### Integration Tests

**Test file**: `internal/middleware/cors_integration_test.go`

**Test approach**: Spin up test HTTP server with Gin + CORS middleware

**Scenarios**:
1. Preflight OPTIONS from allowed origin → 204 with CORS headers
2. Preflight OPTIONS from disallowed origin → 403 or no CORS headers
3. GET request from private subnet (dev mode) → allowed with credentials
4. GET request from private subnet (production mode) → blocked
5. POST request from localhost (dev mode) → allowed
6. POST request from production domain (both modes) → allowed

### Manual Testing

**Checklist**:
- [ ] Local development with ENV=dev → private IPs work
- [ ] Local development with ENV=production → private IPs blocked
- [ ] Staging deployment → VPN access works
- [ ] Production deployment → no regression
- [ ] Browser DevTools → CORS headers correct
- [ ] HttpOnly cookie → still set and sent correctly

## Open Questions

### Q1: Should subnets be configurable via ENV in MVP?

**Decision**: No, hardcode for MVP. Add later if needed.
**Rationale**: YAGNI principle, subnets stable, premature optimization.
**Revisit**: If subnets change >2 times in 6 months.

### Q2: Should we log CORS rejections for security monitoring?

**Decision**: Yes, log at INFO level in production, DEBUG in dev.
**Rationale**: Security visibility, detect misconfiguration/attacks.
**Format**: "CORS rejected origin: <origin> (mode: production)"

### Q3: Should staging use same subnets as dev?

**Decision**: Yes, staging and dev use same private subnets.
**Rationale**: Same VPN, same network topology, simpler config.
**Exception**: If staging needs different subnets, use ENV=staging with different subnet list (future).

### Q4: Should localhost be allowed in staging?

**Decision**: Yes, staging behaves like dev (permissive).
**Rationale**: Staging is for testing, not production-equivalent security.
**Alternative**: Add ENV=staging-strict if needed (future).

## References

- [MDN: CORS with Credentials](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS#requests_with_credentials)
- [OWASP: HttpOnly Cookie Security](https://owasp.org/www-community/HttpOnly)
- [Go net Package](https://pkg.go.dev/net)
- [RFC 4632: CIDR Notation](https://datatracker.ietf.org/doc/html/rfc4632)
- [Gin CORS Middleware](https://github.com/gin-gonic/gin)
- OpenSpec archived change: `2025-11-27-migrate-httponly-cookie-auth`
