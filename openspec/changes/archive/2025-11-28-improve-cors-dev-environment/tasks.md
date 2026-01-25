# Implementation Tasks

## 1. Configuration Layer

- [x] 1.1 Add `Environment` field to `config/config.go` Config struct with `env:"ENV" envDefault:"production"`
- [x] 1.2 Update `config/config.go` documentation to explain environment modes (production, staging, dev)
- [x] 1.3 Update `template.env` file with ENV configuration example and explanation

## 2. CORS Middleware Core Implementation

- [x] 2.1 Add `AllowOriginFunc` field to `CORSConfig` struct in `internal/middleware/cors.go` with type `func(origin string) bool`
- [x] 2.2 Create `isPrivateOrigin(origin string) bool` function to validate IPs against private subnet CIDRs using Go's `net` package
- [x] 2.3 Create `isLocalhostOrigin(origin string) bool` function to detect localhost and 127.0.0.1 with any protocol/port
- [x] 2.4 Define `privateSubnets []string` variable with default CIDR ranges: "172.16.21.0/24", "172.16.19.0/24", "192.168.1.0/24"
- [x] 2.5 Define `productionOrigins []string` variable with production domains: "https://smap.tantai.dev", "https://smap-api.tantai.dev"
- [x] 2.6 Refactor `DefaultCORSConfig()` to accept environment parameter and return environment-specific configuration
- [x] 2.7 Implement production mode logic: return strict origin list without `AllowOriginFunc`
- [x] 2.8 Implement non-production mode logic: set `AllowOriginFunc` to validate production origins, localhost, and private subnets
- [x] 2.9 Update `CORS()` middleware function to handle `AllowOriginFunc` when set (check func before AllowedOrigins list)

## 3. HTTP Server Integration

- [x] 3.1 Update `internal/httpserver/handler.go` to pass environment from config to `DefaultCORSConfig()`
- [x] 3.2 Ensure CORS middleware is initialized with environment-aware configuration during server startup
- [x] 3.3 Add startup log message indicating which CORS mode is active (production/non-production)

## 4. Testing

- [x] 4.1 Create unit tests for `isPrivateOrigin()` function covering:
  - Valid IPs in each configured subnet
  - IPs outside configured subnets
  - Invalid IP formats
  - Invalid origin URLs
  - Edge cases (boundary IPs in CIDR range)
- [x] 4.2 Create unit tests for `isLocalhostOrigin()` function covering:
  - http://localhost with various ports
  - https://localhost with various ports
  - http://127.0.0.1 with various ports
  - https://127.0.0.1 with various ports
  - localhost without port
  - Non-localhost origins
- [x] 4.3 Create unit tests for `DefaultCORSConfig()` covering:
  - Production mode returns strict origins
  - Dev/staging mode returns AllowOriginFunc
  - Empty ENV defaults to production
  - AllowOriginFunc validates production origins
  - AllowOriginFunc validates localhost
  - AllowOriginFunc validates private subnets
  - AllowOriginFunc rejects unauthorized origins

## 5. Documentation

- [x] 5.1 Update README.md "Configuration" section to document ENV variable and its values
- [x] 5.2 Update README.md "CORS Configuration" section with environment-based behavior explanation
- [x] 5.3 Update README.md "Development Setup" section with VPN/private subnet access instructions
- [x] 5.4 Create or update deployment documentation with Kubernetes ConfigMap examples for different environments
- [x] 5.5 Document private subnet CIDR configuration and how to modify for different network topologies
- [x] 5.6 Add security notes explaining why production uses strict origins and non-production uses dynamic validation

## 6. Infrastructure Configuration

- [x] 6.1 Create Kubernetes ConfigMap example for development environment with ENV=dev
- [x] 6.2 Create Kubernetes ConfigMap example for staging environment with ENV=staging
- [x] 6.3 Create Kubernetes ConfigMap example for production environment with ENV=production
- [ ] 6.4 Update Kubernetes Deployment manifest to mount ENV from ConfigMap
- [ ] 6.5 Verify environment variable propagation in containerized environments

## 7. Cleanup

- [x] 7.1 Remove hardcoded origin list comments from `internal/middleware/cors.go` that are now outdated
- [x] 7.2 Update inline documentation to reflect environment-based configuration

## 8. Update README

- [x] 8.1 Review and update the README to ensure all changes regarding CORS, ENV, and development/deployment instructions are fully documented.

## 9. Proposal for Microservice-wide CORS Standardization

- [x] 9.1 Write a short proposal (markdown or doc) describing the dynamic CORS solution for dev, staging, and production environments.
- [x] 9.2 Suggest steps/roadmap to standardize CORS configuration across other microservices by following the guidelines and patterns implemented in this service.
- [x] 9.3 Compile a guide to facilitate easy adoption and replication of this approach across the entire microservices system.

## Dependencies and Parallelization

**Sequential Dependencies:**
- Tasks 1.x must complete before 2.x (config needed for middleware)
- Tasks 2.x must complete before 3.x (middleware must be implemented before integration)
- Tasks 2.x and 3.x must complete before 4.x (implementation needed before testing)
- Tasks 4.x should complete before 7.x (tests should pass before deployment)

**Parallelizable Work:**
- Tasks 4.1, 4.2, 4.3, 4.4 can be done in parallel (independent test suites)
- Tasks 5.x (documentation) can be done in parallel with 4.x (testing)
- Tasks 6.x (infrastructure) can be done in parallel with 5.x (documentation)
- Tasks 9.x và 10.x có thể thực hiện song song sau khi đã hoàn thiện từ task 1 đến 8.

**Critical Path:**
1. Configuration (1.x) → 2. Implementation (2.x) → 3. Integration (3.x) → 4. Testing (4.x) → 7. Deployment (7.x)

**Estimated Effort:**
- Configuration: 1 hour
- Implementation: 4 hours
- Integration: 1 hour
- Testing: 6 hours
- Documentation: 3 hours
- Infrastructure: 2 hours
- Validation: 3 days (includes monitoring periods)
- Update README: 1 hour
- Proposal: 1 hour
- Total development time: ~17 hours + 3 days monitoring
