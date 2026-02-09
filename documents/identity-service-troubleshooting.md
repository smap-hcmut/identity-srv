# Identity Service Troubleshooting Guide

## Table of Contents

1. [OAuth Issues](#oauth-issues)
2. [JWT Issues](#jwt-issues)
3. [Google Directory API Issues](#google-directory-api-issues)
4. [Redis Issues](#redis-issues)
5. [Kafka Issues](#kafka-issues)
6. [Session Issues](#session-issues)
7. [Token Blacklist Issues](#token-blacklist-issues)
8. [Rate Limiting Issues](#rate-limiting-issues)
9. [Configuration Issues](#configuration-issues)
10. [Performance Issues](#performance-issues)
11. [Diagnostic Commands](#diagnostic-commands)
12. [Log Analysis](#log-analysis)

---

## OAuth Issues

### Issue 1: OAuth Callback Returns "Invalid State"

**Symptoms**:
- Error message: `INVALID_STATE`
- HTTP Status: `400 Bad Request`

**Causes**:
- State cookie expired (> 5 minutes)
- State parameter tampered with
- Cookie not sent by browser

**Solutions**:
1. Check browser cookie settings (must allow cookies)
2. Verify cookie domain configuration:
   ```yaml
   cookie:
     domain: .yourdomain.com  # Note the leading dot
   ```
3. Check if HTTPS is enabled (required for Secure cookies)
4. Clear browser cookies and retry

**Debug**:
```bash
# Check cookie in browser DevTools
# Application → Cookies → oauth_state

# Expected: state value matching URL parameter
```

---

### Issue 2: OAuth Exchange Fails

**Symptoms**:
- Error message: `OAUTH_EXCHANGE_FAILED`
- HTTP Status: `500 Internal Server Error`

**Causes**:
- Invalid Client ID or Client Secret
- Redirect URI mismatch
- Network connectivity issues

**Solutions**:
1. Verify OAuth credentials in config:
   ```yaml
   oauth2:
     client_id: 123456789-abc123def456.apps.googleusercontent.com
     client_secret: GOCSPX-abc123def456ghi789jkl012
   ```
2. Check redirect URI matches Google Console:
   ```bash
   # Config
   redirect_uri: https://yourdomain.com/authentication/callback
   
   # Google Console → Credentials → OAuth 2.0 Client IDs
   # Authorized redirect URIs must match exactly
   ```
3. Test network connectivity:
   ```bash
   curl -v https://oauth2.googleapis.com/token
   ```

**Logs to Check**:
```
Failed to exchange code for token: oauth2: cannot fetch token: ...
```

---

### Issue 3: Domain Not Allowed

**Symptoms**:
- Error message: `DOMAIN_NOT_ALLOWED`
- HTTP Status: `403 Forbidden`

**Causes**:
- User email domain not in allowed list
- Configuration not loaded correctly

**Solutions**:
1. Add domain to allowed list:
   ```yaml
   access_control:
     allowed_domains:
       - yourdomain.com
       - partner.com
   ```
2. Verify configuration loaded:
   ```bash
   # Check logs for:
   "Loaded configuration with X allowed domains"
   ```
3. Restart service after config change

**Debug**:
```bash
# Check user email domain
# Logs will show: "Domain not allowed: user@example.com"
```

---

### Issue 4: Account Blocked

**Symptoms**:
- Error message: `ACCOUNT_BLOCKED`
- HTTP Status: `403 Forbidden`

**Causes**:
- User email in blocked list
- Account manually blocked by admin

**Solutions**:
1. Remove from blocked list:
   ```yaml
   access_control:
     blocked_emails:
       # - blocked@yourdomain.com  # Comment out or remove
   ```
2. Restart service after config change

---

## JWT Issues

### Issue 1: JWT Signature Verification Fails

**Symptoms**:
- Error: `Invalid signature`
- HTTP Status: `401 Unauthorized`

**Causes**:
- Public key mismatch
- JWT signed with different private key
- Key rotation not synchronized

**Solutions**:
1. Verify public key matches private key:
   ```bash
   # Extract public key from private key
   openssl rsa -in jwt-private.pem -pubout
   
   # Compare with jwt-public.pem
   cat jwt-public.pem
   ```
2. Check JWKS endpoint returns correct key:
   ```bash
   curl https://yourdomain.com/.well-known/jwks.json
   ```
3. Verify key ID (kid) in JWT header matches JWKS:
   ```bash
   # Decode JWT header
   echo "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6ImExYjJjM2Q0In0" | base64 -d
   ```

**Debug**:
```bash
# Verify JWT manually
jwt decode <TOKEN>

# Check kid in header
# Check signature with public key
```

---

### Issue 2: JWT Expired

**Symptoms**:
- Error: `Token expired`
- HTTP Status: `401 Unauthorized`

**Causes**:
- Token TTL exceeded (default 8 hours)
- System clock skew

**Solutions**:
1. User needs to login again
2. Check system clock synchronization:
   ```bash
   # Check current time
   date
   
   # Sync with NTP
   ntpdate -s time.nist.gov
   ```
3. Adjust TTL if needed:
   ```yaml
   jwt:
     ttl: 28800  # 8 hours in seconds
   ```

**Debug**:
```bash
# Decode JWT and check exp claim
jwt decode <TOKEN> | jq '.exp'

# Compare with current time
date +%s
```

---

### Issue 3: JWT Missing Claims

**Symptoms**:
- Error: `Missing required claim`
- HTTP Status: `401 Unauthorized`

**Causes**:
- JWT generated with old code
- Claims not populated correctly

**Solutions**:
1. Verify JWT contains all required claims:
   ```bash
   jwt decode <TOKEN> | jq '.'
   
   # Required claims:
   # - iss (issuer)
   # - aud (audience)
   # - sub (user ID)
   # - email
   # - role
   # - groups
   # - jti
   # - iat
   # - exp
   ```
2. User needs to login again to get new JWT

---

## Google Directory API Issues

### Issue 1: Failed to Fetch User Groups

**Symptoms**:
- Warning: `Failed to fetch user groups (using default role)`
- User assigned default role instead of mapped role

**Causes**:
- Service account not configured
- Domain-wide delegation not enabled
- Insufficient permissions

**Solutions**:
1. Verify service account configuration:
   ```yaml
   google_workspace:
     service_account_key: /app/keys/service-account.json
     admin_email: admin@yourdomain.com
     domain: yourdomain.com
   ```
2. Check domain-wide delegation in Google Admin Console:
   - Go to Security → API Controls → Domain-wide delegation
   - Verify Client ID is authorized
   - Verify scope: `https://www.googleapis.com/auth/admin.directory.group.readonly`
3. Verify admin email has admin privileges
4. Test Directory API access:
   ```bash
   # Check logs for:
   "Google Directory API connected successfully for domain: yourdomain.com"
   ```

**Debug**:
```bash
# Check service account file
cat /app/keys/service-account.json | jq '.'

# Verify client_email and private_key exist
```

---

### Issue 2: Groups Cache Not Working

**Symptoms**:
- Directory API called on every login (slow)
- High API quota usage

**Causes**:
- Redis connection issues
- Cache TTL too short

**Solutions**:
1. Verify Redis connection:
   ```bash
   redis-cli -h <host> -p <port> -a <password> PING
   ```
2. Check cache TTL (default 5 minutes):
   ```bash
   # Check Redis keys
   redis-cli -h <host> -p <port> -a <password> KEYS "groups:*"
   
   # Check TTL
   redis-cli -h <host> -p <port> -a <password> TTL "groups:user@example.com"
   ```
3. Monitor cache hit rate in logs:
   ```
   "Groups cache hit for user@example.com"
   "Groups cache miss for user@example.com, fetching from API"
   ```

---

## Redis Issues

### Issue 1: Connection Refused

**Symptoms**:
- Error: `Failed to connect to Redis: connection refused`
- Service fails to start

**Causes**:
- Redis not running
- Wrong host/port configuration
- Firewall blocking connection

**Solutions**:
1. Verify Redis is running:
   ```bash
   # Docker
   docker ps | grep redis
   
   # Kubernetes
   kubectl get pods -n redis
   
   # Systemd
   systemctl status redis
   ```
2. Test connectivity:
   ```bash
   redis-cli -h <host> -p <port> -a <password> PING
   ```
3. Check firewall rules:
   ```bash
   # Allow Redis port
   sudo ufw allow 6379/tcp
   ```
4. Verify configuration:
   ```yaml
   redis:
     host: redis.example.com
     port: 6379
     password: <REDIS_PASSWORD>
   ```

---

### Issue 2: Authentication Failed

**Symptoms**:
- Error: `NOAUTH Authentication required`

**Causes**:
- Wrong Redis password
- Redis requirepass not set

**Solutions**:
1. Verify Redis password:
   ```bash
   redis-cli -h <host> -p <port> -a <password> PING
   ```
2. Check Redis configuration:
   ```bash
   redis-cli -h <host> -p <port> CONFIG GET requirepass
   ```
3. Update config with correct password:
   ```yaml
   redis:
     password: <CORRECT_PASSWORD>
   ```

---

### Issue 3: Out of Memory

**Symptoms**:
- Error: `OOM command not allowed when used memory > 'maxmemory'`
- Sessions/cache not working

**Causes**:
- Redis maxmemory limit reached
- Too many keys stored
- Memory leak

**Solutions**:
1. Check Redis memory usage:
   ```bash
   redis-cli -h <host> -p <port> -a <password> INFO memory
   ```
2. Increase maxmemory:
   ```bash
   redis-cli -h <host> -p <port> -a <password> CONFIG SET maxmemory 2gb
   ```
3. Set eviction policy:
   ```bash
   redis-cli -h <host> -p <port> -a <password> CONFIG SET maxmemory-policy allkeys-lru
   ```
4. Clean up old keys:
   ```bash
   # Check key count
   redis-cli -h <host> -p <port> -a <password> DBSIZE
   
   # Find keys without TTL
   redis-cli -h <host> -p <port> -a <password> --scan --pattern "*" | while read key; do
     ttl=$(redis-cli -h <host> -p <port> -a <password> TTL "$key")
     if [ "$ttl" = "-1" ]; then
       echo "$key has no TTL"
     fi
   done
   ```

---

## Kafka Issues

### Issue 1: Failed to Connect to Kafka

**Symptoms**:
- Warning: `Failed to initialize Kafka producer (audit logging will be buffered)`
- Audit events not persisted

**Causes**:
- Kafka brokers not reachable
- Wrong broker addresses
- Network issues

**Solutions**:
1. Verify Kafka brokers:
   ```bash
   # Test connectivity
   telnet kafka-1.example.com 9092
   ```
2. Check broker list:
   ```yaml
   kafka:
     brokers:
       - kafka-1.example.com:9092
       - kafka-2.example.com:9092
   ```
3. Verify topic exists:
   ```bash
   kafka-topics.sh --list --bootstrap-server kafka-1.example.com:9092
   ```

**Note**: Service continues to work without Kafka. Audit events are buffered in memory and will be sent when Kafka becomes available.

---

### Issue 2: Audit Events Not Consumed

**Symptoms**:
- Audit events published but not in database
- Consumer lag increasing

**Causes**:
- Consumer not running
- Consumer group offset issues
- Database connection issues

**Solutions**:
1. Check consumer status:
   ```bash
   # Kubernetes
   kubectl get pods -l app=identity-consumer -n identity-service
   
   # Docker
   docker ps | grep consumer
   ```
2. Check consumer lag:
   ```bash
   kafka-consumer-groups.sh --bootstrap-server kafka-1.example.com:9092 \
     --group identity-service-audit-consumer \
     --describe
   ```
3. Reset consumer offset if needed:
   ```bash
   kafka-consumer-groups.sh --bootstrap-server kafka-1.example.com:9092 \
     --group identity-service-audit-consumer \
     --reset-offsets --to-earliest --topic audit.events \
     --execute
   ```
4. Check consumer logs:
   ```bash
   kubectl logs -f deployment/identity-consumer -n identity-service
   ```

---

## Session Issues

### Issue 1: Session Not Found

**Symptoms**:
- User logged in but `/auth/me` returns 401
- Session expires immediately

**Causes**:
- Redis connection issues
- Session TTL too short
- Session key mismatch

**Solutions**:
1. Check Redis connection (see Redis Issues)
2. Verify session TTL:
   ```yaml
   session:
     ttl: 28800  # 8 hours
   ```
3. Check session in Redis:
   ```bash
   # List all sessions
   redis-cli -h <host> -p <port> -a <password> KEYS "user_sessions:*"
   
   # Check specific session
   redis-cli -h <host> -p <port> -a <password> GET "user_sessions:<user_id>"
   ```
4. Check JWT jti matches session:
   ```bash
   # Decode JWT
   jwt decode <TOKEN> | jq '.jti'
   
   # Check in Redis
   redis-cli -h <host> -p <port> -a <password> GET "user_sessions:<user_id>"
   ```

---

### Issue 2: Multiple Sessions Not Working

**Symptoms**:
- User can only login from one device
- Second login invalidates first session

**Causes**:
- Old code using single session per user
- Session manager not updated

**Solutions**:
1. Verify using latest code (supports multiple sessions)
2. Check session storage format:
   ```bash
   # Should be array of JTIs
   redis-cli -h <host> -p <port> -a <password> GET "user_sessions:<user_id>"
   # Expected: ["jti1", "jti2", "jti3"]
   ```
3. Clear old sessions and re-login:
   ```bash
   redis-cli -h <host> -p <port> -a <password> DEL "user_sessions:<user_id>"
   ```

---

## Token Blacklist Issues

### Issue 1: Revoked Token Still Works

**Symptoms**:
- Token revoked but still accepted
- User can access after logout

**Causes**:
- Blacklist Redis (DB=1) connection issues
- JTI not added to blacklist
- Services not checking blacklist

**Solutions**:
1. Verify blacklist Redis connection:
   ```bash
   redis-cli -h <host> -p <port> -a <password> -n 1 PING
   ```
2. Check if JTI in blacklist:
   ```bash
   # Get JTI from JWT
   jwt decode <TOKEN> | jq '.jti'
   
   # Check blacklist (DB=1)
   redis-cli -h <host> -p <port> -a <password> -n 1 EXISTS "blacklist:<jti>"
   ```
3. Verify services are checking blacklist:
   ```go
   // Services should use pkg/auth middleware
   // which automatically checks blacklist
   ```

---

### Issue 2: Blacklist Growing Too Large

**Symptoms**:
- Redis memory usage increasing
- Blacklist contains expired tokens

**Causes**:
- TTL not set correctly
- Old tokens not expiring

**Solutions**:
1. Check blacklist TTL:
   ```bash
   # List blacklist keys
   redis-cli -h <host> -p <port> -a <password> -n 1 KEYS "blacklist:*"
   
   # Check TTL (should match remaining token lifetime)
   redis-cli -h <host> -p <port> -a <password> -n 1 TTL "blacklist:<jti>"
   ```
2. Clean up expired entries:
   ```bash
   # Redis automatically removes expired keys
   # Force cleanup
   redis-cli -h <host> -p <port> -a <password> -n 1 --scan --pattern "blacklist:*" | while read key; do
     ttl=$(redis-cli -h <host> -p <port> -a <password> -n 1 TTL "$key")
     if [ "$ttl" = "-2" ]; then
       echo "Removing expired key: $key"
       redis-cli -h <host> -p <port> -a <password> -n 1 DEL "$key"
     fi
   done
   ```

---

## Rate Limiting Issues

### Issue 1: Legitimate Users Blocked

**Symptoms**:
- Error: `TOO_MANY_REQUESTS`
- Users cannot login

**Causes**:
- Rate limit too strict
- Shared IP address (NAT, proxy)
- Failed login attempts

**Solutions**:
1. Check rate limit configuration:
   ```go
   // In cmd/api/main.go
   rateLimiter := authUsecase.NewRateLimiter(
       redisClient.GetClient(),
       5,              // Max attempts
       15*time.Minute, // Window
       30*time.Minute, // Block duration
   )
   ```
2. Manually unblock IP:
   ```bash
   # Remove rate limit counter
   redis-cli -h <host> -p <port> -a <password> DEL "ratelimit:login:<IP>"
   
   # Remove block
   redis-cli -h <host> -p <port> -a <password> DEL "ratelimit:block:<IP>"
   ```
3. Whitelist trusted IPs (requires code change)

---

### Issue 2: Rate Limit Not Working

**Symptoms**:
- Brute force attacks succeeding
- No rate limiting applied

**Causes**:
- Rate limiter not initialized
- Middleware not applied
- Redis connection issues

**Solutions**:
1. Verify rate limiter initialized:
   ```bash
   # Check logs for:
   "Rate limiter initialized (max 5 attempts per 15 minutes)"
   ```
2. Verify middleware applied to routes:
   ```go
   // In routes.go
   r.GET("/login", rateLimitMW, h.OAuthLogin)
   r.GET("/callback", rateLimitMW, h.OAuthCallback)
   ```
3. Check Redis connection (see Redis Issues)

---

## Configuration Issues

### Issue 1: Configuration Not Loaded

**Symptoms**:
- Service fails to start
- Error: `Failed to load config`

**Causes**:
- Config file not found
- Invalid YAML syntax
- Missing required fields

**Solutions**:
1. Verify config file path:
   ```bash
   # Check environment variable
   echo $CONFIG_FILE
   
   # Check file exists
   ls -la /app/auth-config.yaml
   ```
2. Validate YAML syntax:
   ```bash
   # Use yamllint
   yamllint auth-config.yaml
   
   # Or Python
   python -c "import yaml; yaml.safe_load(open('auth-config.yaml'))"
   ```
3. Check for required fields:
   ```bash
   # Service will log specific missing fields:
   "oauth2.client_id is required"
   "jwt.private_key_path is required"
   ```

---

### Issue 2: Configuration Validation Fails

**Symptoms**:
- Service fails to start
- Error: `Configuration validation failed`

**Causes**:
- Invalid values
- Wrong format
- Missing required fields

**Solutions**:
1. Check validation errors in logs:
   ```
   "oauth2.redirect_uri must be a valid HTTP/HTTPS URL"
   "encrypter.key must be at least 32 characters"
   "access_control.default_role must be one of: ADMIN, ANALYST, VIEWER"
   ```
2. Fix configuration based on error message
3. Restart service

---

## Performance Issues

### Issue 1: Slow Login Response

**Symptoms**:
- Login takes > 5 seconds
- Users complain about slow authentication

**Causes**:
- Google API latency
- Database slow queries
- Redis latency
- Network issues

**Solutions**:
1. Check Google API response time:
   ```bash
   # Monitor logs for:
   "OAuth exchange took 2.5s"
   "Directory API call took 1.2s"
   ```
2. Enable groups caching (should be enabled by default):
   ```yaml
   # Groups cached for 5 minutes
   # Check cache hit rate in logs
   ```
3. Check database query performance:
   ```sql
   -- Enable query logging
   ALTER DATABASE identity_service SET log_statement = 'all';
   
   -- Check slow queries
   SELECT * FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 10;
   ```
4. Check Redis latency:
   ```bash
   redis-cli -h <host> -p <port> -a <password> --latency
   ```

---

### Issue 2: High Memory Usage

**Symptoms**:
- Service using > 1GB memory
- OOM kills in Kubernetes

**Causes**:
- Memory leak
- Too many cached items
- Large audit event buffer

**Solutions**:
1. Check memory usage:
   ```bash
   # Docker
   docker stats identity-service
   
   # Kubernetes
   kubectl top pod -l app=identity-service -n identity-service
   ```
2. Profile memory usage:
   ```bash
   # Enable pprof endpoint
   curl http://localhost:8080/debug/pprof/heap > heap.prof
   
   # Analyze with go tool
   go tool pprof heap.prof
   ```
3. Adjust resource limits:
   ```yaml
   # Kubernetes
   resources:
     limits:
       memory: "1Gi"
     requests:
       memory: "512Mi"
   ```

---

## Diagnostic Commands

### Check Service Health

```bash
# Health check
curl https://yourdomain.com/health

# Readiness check
curl https://yourdomain.com/ready

# Liveness check
curl https://yourdomain.com/live
```

### Check JWT

```bash
# Decode JWT (header + payload)
jwt decode <TOKEN>

# Verify JWT signature
jwt verify <TOKEN> --key jwt-public.pem

# Check expiration
jwt decode <TOKEN> | jq '.exp' | xargs -I {} date -d @{}
```

### Check Redis

```bash
# Connection
redis-cli -h <host> -p <port> -a <password> PING

# Memory usage
redis-cli -h <host> -p <port> -a <password> INFO memory

# Key count
redis-cli -h <host> -p <port> -a <password> DBSIZE

# List sessions
redis-cli -h <host> -p <port> -a <password> KEYS "user_sessions:*"

# List blacklist (DB=1)
redis-cli -h <host> -p <port> -a <password> -n 1 KEYS "blacklist:*"

# Monitor commands
redis-cli -h <host> -p <port> -a <password> MONITOR
```

### Check Kafka

```bash
# List topics
kafka-topics.sh --list --bootstrap-server kafka-1.example.com:9092

# Describe topic
kafka-topics.sh --describe --bootstrap-server kafka-1.example.com:9092 --topic audit.events

# Check consumer group
kafka-consumer-groups.sh --bootstrap-server kafka-1.example.com:9092 \
  --group identity-service-audit-consumer --describe

# Consume messages
kafka-console-consumer.sh --bootstrap-server kafka-1.example.com:9092 \
  --topic audit.events --from-beginning --max-messages 10
```

### Check Database

```bash
# Connection
psql -h <host> -U identity_user -d identity_service -c "SELECT 1"

# Table sizes
psql -h <host> -U identity_user -d identity_service -c "
  SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
  FROM pg_tables
  WHERE schemaname = 'public'
  ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
"

# Row counts
psql -h <host> -U identity_user -d identity_service -c "
  SELECT 'users' AS table, COUNT(*) FROM users
  UNION ALL
  SELECT 'audit_logs', COUNT(*) FROM audit_logs
  UNION ALL
  SELECT 'jwt_keys', COUNT(*) FROM jwt_keys;
"

# Recent audit logs
psql -h <host> -U identity_user -d identity_service -c "
  SELECT * FROM audit_logs ORDER BY created_at DESC LIMIT 10;
"
```

---

## Log Analysis

### Important Log Messages

**Successful Login**:
```
INFO  OAuth callback successful for user@example.com
INFO  User role mapped to ANALYST based on groups
INFO  Session created for user <user_id>
INFO  Audit event published: LOGIN
```

**Failed Login**:
```
WARN  Domain not allowed: user@example.com
ERROR Failed to exchange code for token: ...
WARN  Account blocked: user@example.com
```

**Rate Limiting**:
```
WARN  Rate limit exceeded for IP 192.168.1.100
INFO  Failed login attempt recorded for IP 192.168.1.100
```

**Google API**:
```
INFO  Google Directory API connected successfully
WARN  Failed to fetch user groups (using default role)
INFO  Groups cache hit for user@example.com
INFO  Groups cache miss, fetching from API
```

**Redis**:
```
INFO  Redis connected successfully to redis.example.com:6379 (DB 0)
INFO  Redis blacklist connected successfully (DB 1)
ERROR Failed to connect to Redis: connection refused
```

**Kafka**:
```
INFO  Kafka producer initialized successfully
WARN  Failed to initialize Kafka producer (audit logging will be buffered)
INFO  Audit event buffered (Kafka unavailable)
INFO  Flushed 10 buffered audit events to Kafka
```

### Log Levels

- **DEBUG**: Detailed information for debugging
- **INFO**: General informational messages
- **WARN**: Warning messages (service continues)
- **ERROR**: Error messages (operation failed)

### Enable Debug Logging

```yaml
logger:
  level: debug
  mode: development
  encoding: console
  color_enabled: true
```

---

## Getting Help

If you cannot resolve the issue:

1. **Collect diagnostic information**:
   - Service logs (last 1000 lines)
   - Configuration file (redact secrets)
   - Health check output
   - Redis/Kafka/Database status

2. **Check documentation**:
   - API Documentation: `documents/identity-service-api.md`
   - Deployment Guide: `documents/identity-service-deployment.md`

3. **Contact support**:
   - Include diagnostic information
   - Describe steps to reproduce
   - Include error messages and logs
