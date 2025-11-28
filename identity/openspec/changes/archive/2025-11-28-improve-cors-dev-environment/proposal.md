# Change: Improve CORS Configuration for Development Environments

## Why

The current CORS configuration uses a hardcoded list of allowed origins, which prevents developers from accessing the API when working remotely via VPN from private subnets (172.16.21.0/24, 172.16.19.0/24, 192.168.1.0/24). This disrupts development workflow and forces developers to work exclusively on localhost, reducing team productivity and making integration testing from different devices impossible.

The system requires `AllowCredentials: true` for HttpOnly cookie authentication, which means we cannot use wildcard origins. We need an environment-aware solution that maintains production security while enabling flexible development access.

## What Changes

- **Add environment-based CORS configuration** that differentiates between production, staging, and development environments
- **Implement dynamic origin validation** using `AllowOriginFunc` for non-production environments
- **Add private subnet detection** to automatically allow IPs within configured CIDR ranges (172.16.21.0/24, 172.16.19.0/24, 192.168.1.0/24)
- **Add localhost detection** to support both HTTP and HTTPS localhost origins on any port
- **Add ENV configuration** variable to control environment mode (production, staging, dev)
- **Maintain strict origin list** for production environment with no dynamic validation

**BREAKING**: This change requires adding an `ENV` environment variable to distinguish between environments. Systems without this variable will default to production mode (strict origins only) for security.

## Impact

### Affected Specs
- `cors-middleware` - New capability spec (currently not specified, will be created)

### Affected Code
- `internal/middleware/cors.go` - Complete refactor of `DefaultCORSConfig()` function and addition of new validation functions
- `config/config.go` - Add environment configuration field
- `internal/httpserver/handler.go` - Pass environment config to CORS middleware
- Kubernetes ConfigMaps - Add ENV variable configuration for different environments
- Documentation - Update deployment and development setup guides

### Benefits
- 100% of developers can access API from any device within VPN subnets
- No need to manually add each developer's IP to config
- Maintains production security with strict origin validation
- Supports horizontal scaling across multiple developer subnets
- Reduces environment setup time by ~50%

### Risks
- **Medium**: Misconfiguration could accidentally enable permissive CORS in production
  - **Mitigation**: Default to production mode, add startup validation
- **Low**: Performance overhead from CIDR subnet checking
  - **Mitigation**: CIDR checking is microsecond-scale, negligible impact
- **Low**: Private subnet IPs could be exposed
  - **Mitigation**: Only enabled in non-production environments

## Related Changes
- This change builds on the archived change `2025-11-27-migrate-httponly-cookie-auth` which established HttpOnly cookie authentication with `AllowCredentials: true`

## Success Metrics
- **Before**: 0% developers can access API via VPN from private subnets
- **After**: 100% developers can access API from all configured private subnets
- **Security**: 0 CORS-related security incidents in production
- **Performance**: <1ms overhead for CORS validation
