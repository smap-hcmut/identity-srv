# Service: Identity Service (identity-srv)

> **Template Version**: 1.0  
> **Last Updated**: 15/02/2026  
> **Status**: ✅ Production Ready

---

## 🎯 Business Context

### Chức năng chính

Identity Service là dịch vụ xác thực và phân quyền trung tâm cho nền tảng SMAP, cung cấp giải pháp OAuth2 và JWT authentication toàn diện. Service này giải quyết các vấn đề:

- Xác thực người dùng thông qua Google OAuth2 (hỗ trợ Google Workspace)
- Quản lý phiên đăng nhập và token JWT với khả năng thu hồi tức thì
- Phân quyền dựa trên vai trò (RBAC) với 3 cấp độ: ADMIN, ANALYST, VIEWER
- Ghi nhận audit log đầy đủ cho compliance và security monitoring
- Cung cấp API nội bộ cho các service khác xác thực token

### Luồng xử lý

```
User Browser
    → GET /authentication/login (redirect to Google OAuth)
    → Google Authorization Page
    → GET /authentication/callback?code=xxx (exchange code for token)
    → Create/Update User in PostgreSQL
    → Generate JWT Token (HS256)
    → Store Session in Redis
    → Set HttpOnly Cookie (production) / Return JSON (development)
    → User authenticated with JWT in cookie

Subsequent Requests:
    → Request with Cookie/Bearer Token
    → Middleware validates JWT
    → Check token blacklist in Redis
    → Extract user info from token
    → Allow/Deny access based on role
```

### Giá trị cốt lõi

- **Bảo mật cao**: HttpOnly cookies chống XSS, token blacklist cho instant revocation, role encryption
- **Tích hợp dễ dàng**: Cung cấp internal API cho các service khác validate token
- **Audit trail đầy đủ**: Ghi nhận mọi hành động authentication/authorization vào PostgreSQL và Kafka
- **Hiệu năng tốt**: Redis cache cho session và blacklist, giảm tải database
- **Linh hoạt**: Hỗ trợ cả cookie (browser) và Bearer token (mobile/API)
- **Scalable**: Stateless JWT cho phép horizontal scaling dễ dàng

---

## 🛠 Technical Details

### Protocol & Architecture

- **Protocol**: REST API (HTTP/HTTPS)
- **Pattern**: Clean Architecture (Domain-Driven Design)
- **Design**: Layered Architecture với 4 layers: Delivery → UseCase → Repository → Database

### Tech Stack

| Component | Technology      | Version           | Purpose                          |
| --------- | --------------- | ----------------- | -------------------------------- |
| Language  | Go              | 1.25.4            | Backend service                  |
| Framework | Gin             | 1.11.0            | HTTP routing & middleware        |
| Database  | PostgreSQL      | 15+               | User data, audit logs            |
| Cache     | Redis           | 6+                | Session storage, token blacklist |
| Queue     | Kafka           | Latest            | Async audit event processing     |
| Auth      | OAuth2 (Google) | -                 | User authentication              |
| JWT       | HS256           | golang-jwt/jwt/v5 | Token signing (symmetric)        |
| ORM       | SQLBoiler       | v4.19.7           | Type-safe database queries       |
| Config    | Viper           | 1.19.0            | Configuration management         |
| Logger    | Zap             | 1.27.0            | Structured logging               |
| API Docs  | Swagger         | 1.16.6            | OpenAPI documentation            |

### Database Schema

#### PostgreSQL Tables (Schema: schema_identity)

**1. users** - Lưu trữ thông tin người dùng

```sql
CREATE TABLE schema_identity.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    avatar_url TEXT,
    role_hash VARCHAR(255) NOT NULL,  -- Encrypted role
    is_active BOOLEAN DEFAULT TRUE,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- Indexes: idx_users_email, idx_users_is_active
```

**Lưu ý**:

- `role_hash` được mã hóa bằng SHA256 + base64 để bảo mật
- Email là unique identifier, tự động tạo user khi login lần đầu
- `is_active` cho phép block user mà không xóa dữ liệu

**2. audit_logs** - Ghi nhận mọi hành động quan trọng

```sql
CREATE TABLE schema_identity.audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES schema_identity.users(id),
    action VARCHAR(100) NOT NULL,  -- LOGIN, LOGOUT, TOKEN_REVOKED, etc.
    resource_type VARCHAR(100),
    resource_id VARCHAR(255),
    ip_address INET,
    user_agent TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- Indexes: idx_audit_logs_user_id, idx_audit_logs_created_at, idx_audit_logs_action
```

**3. jwt_keys** - Quản lý RSA key pairs (dự phòng cho tương lai)

```sql
CREATE TABLE schema_identity.jwt_keys (
    kid VARCHAR(50) PRIMARY KEY,
    private_key TEXT NOT NULL,
    public_key TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ,
    retired_at TIMESTAMPTZ
);
-- Indexes: idx_jwt_keys_status, idx_jwt_keys_created_at
```

**Lưu ý**: Hiện tại service dùng HS256 (symmetric key), table này dự phòng cho việc migrate sang RS256 (asymmetric) trong tương lai.

#### Redis Data Structures

**Session Storage**:

```
Key: session:{jti}
Value: JSON {user_id, jti, created_at, expires_at}
TTL: 8 hours (default) hoặc 7 days (remember me)
```

**User Sessions Mapping**:

```
Key: user_sessions:{user_id}
Value: JSON array of JTIs
TTL: 7 days
Purpose: Revoke all user tokens
```

**Token Blacklist**:

```
Key: blacklist:{jti}
Value: "1"
TTL: Remaining token lifetime
Purpose: Instant token revocation
```

---

## 📡 API Endpoints

### Domain 1: Authentication (Public)

#### `GET /authentication/login`

**Purpose**: Khởi tạo OAuth2 flow, redirect user đến Google authorization page

**Authentication**: Public (no auth required)

**Query Parameters**:

```
redirect (optional): URL to redirect after successful login
```

**Response**: HTTP 302 Redirect to Google OAuth

**Business Logic Flow**:

1. Validate redirect URL against whitelist
2. Generate OAuth state for CSRF protection
3. Store state in Redis (5 min TTL)
4. Build Google OAuth authorization URL
5. Redirect user to Google

**Performance**: <10ms (chỉ redirect, không có I/O nặng)

---

#### `GET /authentication/callback`

**Purpose**: Xử lý OAuth2 callback từ Google, tạo JWT token và session

**Authentication**: Public (OAuth code validation)

**Query Parameters**:

```
code (required): Authorization code from Google
state (required): CSRF protection token
```

**Response** (Development - 200):

```json
{
  "error_code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**Response** (Production - 302): Redirect to dashboard with HttpOnly cookie set

**Response** (Error - 403):

```json
{
  "error_code": 1003,
  "message": "Domain not allowed",
  "errors": null
}
```

**Business Logic Flow**:

1. Validate state parameter (CSRF check)
2. Exchange authorization code for Google access token
3. Fetch user info from Google (email, name, avatar)
4. Validate email domain against whitelist
5. Check if user is blocked
6. Upsert user in database (create if new, update if exists)
7. Determine user role from config mapping
8. Generate JWT token (HS256, 8h TTL)
9. Create session in Redis
10. Set HttpOnly cookie (production) or return JSON (development)
11. Publish LOGIN audit event to Kafka
12. Redirect to dashboard or return token

**Performance**:

- Avg latency: 200-500ms (includes Google API call)
- Cache: No (OAuth flow must be fresh)

---

#### `POST /authentication/logout`

**Purpose**: Đăng xuất user, xóa session và expire cookie

**Authentication**: JWT (Cookie or Bearer token)

**Request**: No body required

**Response** (Success - 200):

```json
{
  "error_code": 0,
  "message": "Logged out successfully"
}
```

**Business Logic Flow**:

1. Extract JWT from cookie or Authorization header
2. Verify JWT signature and expiration
3. Delete session from Redis (key: session:{jti})
4. Expire cookie (set MaxAge=-1)
5. Publish LOGOUT audit event
6. Return success

**Performance**: <50ms (Redis delete operation)

---

#### `GET /authentication/me`

**Purpose**: Lấy thông tin user hiện tại từ JWT token

**Authentication**: JWT (Cookie or Bearer token)

**Response** (Success - 200):

```json
{
  "error_code": 0,
  "message": "success",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "full_name": "John Doe"
  }
}
```

**Business Logic Flow**:

1. Extract user_id from JWT claims (via middleware)
2. Query user from database by ID
3. Return user info (không trả về role_hash)

**Performance**: <20ms (single DB query with index)

---

### Domain 2: Audit Logs (Protected)

#### `GET /audit-logs`

**Purpose**: Query audit logs với pagination và filters (ADMIN only)

**Authentication**: JWT + ADMIN role required

**Query Parameters**:

```
user_id (optional): Filter by user ID
action (optional): Filter by action (LOGIN, LOGOUT, TOKEN_REVOKED, etc.)
from (optional): Start date (RFC3339 format)
to (optional): End date (RFC3339 format)
page (optional): Page number (default: 1)
limit (optional): Items per page (default: 50, max: 100)
```

**Response** (Success - 200):

```json
{
  "error_code": 0,
  "message": "success",
  "data": {
    "logs": [
      {
        "id": "uuid",
        "user_id": "uuid",
        "action": "LOGIN",
        "resource_type": "authentication",
        "resource_id": null,
        "ip_address": "192.168.1.1",
        "user_agent": "Mozilla/5.0...",
        "metadata": { "provider": "google" },
        "created_at": "2026-02-15T10:30:00Z",
        "expires_at": "2026-05-16T10:30:00Z"
      }
    ],
    "total_count": 1250,
    "page": 1,
    "limit": 50
  }
}
```

**Business Logic Flow**:

1. Verify user has ADMIN role (middleware)
2. Parse and validate query parameters
3. Build SQL query with filters
4. Execute paginated query with COUNT
5. Return logs with pagination metadata

**Performance**:

- Avg latency: 50-200ms (depends on filters and data volume)
- Indexes: user_id, action, created_at for fast filtering

---

### Domain 3: Internal API (Service-to-Service)

#### `POST /internal/validate`

**Purpose**: Validate JWT token cho các service khác (fallback method)

**Authentication**: X-Service-Key header (internal service authentication)

**Request**:

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response** (Success - 200):

```json
{
  "error_code": 0,
  "message": "success",
  "data": {
    "valid": true,
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "role": "ADMIN",
    "groups": [],
    "expires_at": "2026-02-15T18:30:00Z"
  }
}
```

**Business Logic Flow**:

1. Verify X-Service-Key header
2. Parse and verify JWT signature
3. Check token expiration
4. Check if token is blacklisted in Redis
5. Return validation result with user claims

**Performance**: <30ms (JWT verify + Redis check)

---

#### `POST /internal/revoke-token`

**Purpose**: Thu hồi token cụ thể hoặc tất cả token của user (ADMIN only)

**Authentication**: X-Service-Key + JWT with ADMIN role

**Request**:

```json
{
  "jti": "token-id-to-revoke",
  "user_id": "user-id-to-revoke-all-tokens"
}
```

**Response** (Success - 200):

```json
{
  "error_code": 0,
  "message": "Token(s) revoked successfully"
}
```

**Business Logic Flow**:

1. Verify service key and ADMIN role
2. If jti provided: Add single token to blacklist
3. If user_id provided: Get all user sessions, add all to blacklist
4. Delete sessions from Redis
5. Publish TOKEN_REVOKED audit event
6. Return success

**Performance**: <100ms (multiple Redis operations)

---

#### `GET /internal/users/:id`

**Purpose**: Lấy thông tin user theo ID cho internal services

**Authentication**: X-Service-Key header

**Response** (Success - 200):

```json
{
  "error_code": 0,
  "message": "success",
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe",
    "avatar_url": "https://...",
    "role": "ADMIN",
    "is_active": true,
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-02-15T10:00:00Z"
  }
}
```

**Business Logic Flow**:

1. Verify service key
2. Query user by ID from database
3. Decrypt role_hash to get plain role
4. Return user info

**Performance**: <20ms (single indexed query)

---

### Domain 4: Health Checks

#### `GET /health`, `GET /ready`, `GET /live`

**Purpose**: Kubernetes health checks

**Response**: `{"status": "ok"}`

**Performance**: <5ms (no I/O)

---

## 🔗 Integration & Dependencies

### External Services

**1. Google OAuth2 API** (Upstream)

- **Method**: HTTPS REST API
- **Purpose**: Xác thực người dùng qua Google account
- **Endpoints Used**:
  - `GET https://accounts.google.com/o/oauth2/v2/auth` - Authorization page
  - `POST https://oauth2.googleapis.com/token` - Exchange code for token
  - `GET https://www.googleapis.com/oauth2/v2/userinfo` - Get user profile
- **Error Handling**: Retry 3 lần với exponential backoff, fallback error message cho user
- **SLA**: Google SLA 99.9%, timeout 10s

**2. PostgreSQL Database** (Infrastructure)

- **Method**: Direct connection via lib/pq driver
- **Purpose**: Persistent storage cho users và audit logs
- **Connection Pool**: Max 25 connections
- **Error Handling**: Connection retry, graceful degradation
- **SLA**: <50ms query time, 99.99% availability

**3. Redis Cache** (Infrastructure)

- **Method**: Direct connection via go-redis/v9
- **Purpose**: Session storage, token blacklist, OAuth state
- **Data Structures**:
  - String: session data, blacklist flags
  - TTL: Auto-expiration cho cleanup
- **Error Handling**: Fail-fast (Redis down = service unavailable)
- **SLA**: <5ms operation time, 99.99% availability

**4. Kafka Message Queue** (Infrastructure)

- **Method**: Sarama producer/consumer
- **Purpose**: Async audit event processing
- **Topics**:
  - `audit.events` - Audit log events
- **Error Handling**: Retry with backoff, dead letter queue
- **SLA**: At-least-once delivery, <100ms publish time

### Infrastructure Dependencies

**Message Queue** (Kafka)

```
Topic: audit.events
Consumer Group: audit-consumer-group
Message Format: {
  "user_id": "uuid",
  "action": "LOGIN",
  "resource_type": "authentication",
  "resource_id": null,
  "ip_address": "192.168.1.1",
  "user_agent": "Mozilla/5.0...",
  "metadata": {"provider": "google"},
  "timestamp": "2026-02-15T10:30:00Z"
}
Handler: Consumer service writes to PostgreSQL audit_logs table
```

**Cache** (Redis)

```
Key Patterns:
- session:{jti} → SessionData JSON (TTL: 8h or 7d)
- user_sessions:{user_id} → Array of JTIs (TTL: 7d)
- blacklist:{jti} → "1" (TTL: remaining token lifetime)
- oauth_state:{state} → redirect_url (TTL: 5m)

Invalidation Strategy:
- Session: Manual delete on logout
- Blacklist: Auto-expire when token expires
- OAuth state: Auto-expire after 5 minutes
```

**Database** (PostgreSQL)

```
Connection Pool: Max 25 connections
Schema: schema_identity
Migrations: Manual SQL scripts in migration/ folder
Backup: Daily automated backups
Indexes: Optimized for email lookup, audit log queries
```

---

## 🎨 Key Features & Highlights

### 1. OAuth2 + JWT Hybrid Authentication

**Description**: Kết hợp OAuth2 (user login) với JWT (stateless auth) để có được ưu điểm của cả hai

**Implementation**:

- OAuth2 flow với Google cho user-friendly login experience
- JWT HS256 cho stateless authentication (không cần query DB mỗi request)
- Redis session cho instant revocation capability
- HttpOnly cookie cho browser security (chống XSS)
- Bearer token support cho mobile/API clients

**Benefits**:

- User không cần nhớ password, dùng Google account
- Stateless JWT cho phép horizontal scaling dễ dàng
- Instant token revocation qua blacklist (best of both worlds)
- Secure cookie storage chống XSS attacks

### 2. Role-Based Access Control (RBAC)

**Description**: Hệ thống phân quyền 3 cấp với role encryption

**Implementation**:

- 3 roles: ADMIN (full access), ANALYST (create/analyze), VIEWER (read-only)
- Role được mã hóa bằng SHA256 + base64 trong database
- Email-to-role mapping trong config cho flexible assignment
- Default role (VIEWER) cho users không có mapping
- Role được embed trong JWT claims cho fast authorization

**Benefits**:

- Không thể đọc trực tiếp role từ database (security)
- Flexible role assignment qua config (không cần code change)
- Fast authorization check (role trong JWT, không query DB)
- Audit trail đầy đủ cho role changes

### 3. Comprehensive Audit Logging

**Description**: Ghi nhận mọi hành động quan trọng cho compliance và security

**Implementation**:

- Dual-write: Sync write to PostgreSQL + async publish to Kafka
- Consumer service process Kafka events (decoupled architecture)
- Rich metadata: IP address, user agent, resource info, custom metadata
- Auto-expiration: Audit logs tự động xóa sau 90 ngày
- Indexed queries: Fast filtering by user, action, date range

**Benefits**:

- Compliance với security standards (SOC2, ISO27001)
- Forensics và incident investigation
- User activity monitoring
- Performance: Async Kafka không block main flow

### 4. Token Blacklist & Session Management

**Description**: Instant token revocation với Redis blacklist

**Implementation**:

- Session storage: Redis với TTL tự động cleanup
- Blacklist: Redis set với TTL = remaining token lifetime
- Revoke single token: Add JTI to blacklist
- Revoke all user tokens: Batch add all JTIs to blacklist
- Middleware check: Verify token not in blacklist mỗi request

**Benefits**:

- Instant revocation (không cần đợi token expire)
- Memory efficient (TTL auto-cleanup)
- Fast check (<5ms Redis lookup)
- Support "logout all devices" feature

### 5. Performance Optimizations

**Caching Strategy**:

- Redis session cache giảm DB load
- JWT stateless auth (không query DB mỗi request)
- Connection pooling cho PostgreSQL (max 25)
- Index optimization cho audit log queries

**Database Optimization**:

- Composite indexes: (user_id, created_at), (action, created_at)
- Partial indexes: is_active users only
- Query optimization: Pagination với LIMIT/OFFSET

**Concurrency**:

- Goroutines cho async audit publishing
- Non-blocking Kafka producer
- Connection pool reuse

### 6. Reliability Features

**Retry Logic**:

- Google OAuth API: 3 retries với exponential backoff
- Kafka producer: Retry với backoff, dead letter queue
- Database: Connection retry on transient errors

**Circuit Breaker**:

- Redis connection failure = service unavailable (fail-fast)
- Kafka failure = log error but continue (non-critical)

**Graceful Degradation**:

- Discord webhook failure: Log warning, continue
- Audit Kafka failure: Still write to PostgreSQL

**Health Checks**:

- `/health`: Overall health
- `/ready`: Ready to serve traffic (DB + Redis connected)
- `/live`: Process alive (for Kubernetes liveness probe)

---

## 🚧 Status & Roadmap

### ✅ Done (Implemented & Tested)

- [x] OAuth2 login flow với Google
- [x] JWT token generation và verification (HS256)
- [x] Session management với Redis
- [x] Token blacklist cho instant revocation
- [x] Role-based access control (ADMIN, ANALYST, VIEWER)
- [x] Email-to-role mapping từ config
- [x] Audit logging với PostgreSQL + Kafka
- [x] Internal API cho service-to-service auth
- [x] HttpOnly cookie cho browser security
- [x] Bearer token support cho mobile/API
- [x] CORS middleware với origin validation
- [x] Swagger API documentation
- [x] Docker deployment (API + Consumer)
- [x] Kubernetes manifests
- [x] Health check endpoints
- [x] Graceful shutdown
- [x] Structured logging với Zap
- [x] Configuration management với Viper

### 🔄 In Progress

- [ ] Load testing và performance benchmarking - Status: 60% complete
- [ ] Monitoring dashboard với Grafana - Status: Planning
- [ ] Rate limiting middleware - Status: Design phase

### 📋 Todo (Planned)

- [ ] Multi-provider OAuth2 (Azure AD, Okta) - Priority: Medium
- [ ] JWT key rotation (migrate to RS256) - Priority: Low
- [ ] 2FA/MFA support - Priority: High
- [ ] API rate limiting per user - Priority: High
- [ ] Metrics export (Prometheus) - Priority: High
- [ ] Distributed tracing (Jaeger) - Priority: Medium
- [ ] User management API (CRUD) - Priority: Medium
- [ ] Bulk user import - Priority: Low
- [ ] Password reset flow (fallback) - Priority: Low

### 🐛 Known Bugs

- [ ] Bug #001: OAuth state cleanup job chưa implement - Severity: Low
  - Workaround: Redis TTL tự động cleanup sau 5 phút
- [ ] Bug #002: Audit log pagination performance với large dataset - Severity: Medium
  - Workaround: Limit max page size = 100

---

## ⚠️ Known Issues & Limitations

### 1. Performance - Large Audit Log Queries

**Issue**: Query audit logs với date range rộng (>1 tháng) và không có user_id filter có thể chậm

- **Current**: Query full table scan nếu không có user_id
- **Problem**: Với >1M audit logs, query có thể mất >2s
- **Impact**: ADMIN users experience slow response khi query large date range
- **Workaround**: Luôn filter theo user_id hoặc giới hạn date range <7 ngày
- **TODO**: Implement partition by month cho audit_logs table

**Code location**: `internal/audit/repository/postgre/audit_query.go`

```go
// ❌ Current implementation
func (r *implRepository) List(ctx context.Context, opts QueryOptions) ([]model.AuditLog, int, error) {
    query := "SELECT * FROM audit_logs WHERE 1=1"
    // No partition, full table scan possible
}

// ✅ Proposed solution
// Partition audit_logs by month
CREATE TABLE audit_logs_2026_02 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
// Add partition pruning to query
```

### 2. Security - HS256 vs RS256

**Issue**: Hiện tại dùng HS256 (symmetric key) cho JWT signing

- **Current**: Secret key được share giữa issuer và verifier
- **Problem**: Nếu secret key bị leak, attacker có thể forge tokens
- **Impact**: Security risk nếu secret key không được bảo vệ tốt
- **Workaround**: Rotate secret key định kỳ, strict access control
- **TODO**: Migrate sang RS256 (asymmetric) với public/private key pair

**Code location**: `pkg/jwt/jwt.go`

```go
// ❌ Current implementation
token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
tokenString, err := token.SignedString(m.secretKey)

// ✅ Proposed solution
token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
tokenString, err := token.SignedString(m.privateKey)
// Expose public key via /jwks endpoint for verifiers
```

### 3. Scalability - Redis Single Point of Failure

**Issue**: Redis là single instance, không có replication

- **Current**: Single Redis instance cho session và blacklist
- **Problem**: Nếu Redis down, toàn bộ service unavailable
- **Impact**: Service downtime khi Redis maintenance hoặc failure
- **Workaround**: Redis persistence (AOF), frequent backups
- **TODO**: Setup Redis Sentinel hoặc Redis Cluster cho HA

**Code location**: `pkg/redis/redis.go`

```go
// ❌ Current implementation
client := redis.NewClient(&redis.Options{
    Addr: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
})

// ✅ Proposed solution
client := redis.NewFailoverClient(&redis.FailoverOptions{
    MasterName:    "mymaster",
    SentinelAddrs: []string{"sentinel1:26379", "sentinel2:26379"},
})
```

### 4. Configuration - Hardcoded Service Keys

**Issue**: Internal service keys được lưu trong config file (plaintext)

- **Current**: Service keys trong auth-config.yaml
- **Problem**: Keys có thể bị expose qua git history hoặc logs
- **Impact**: Security risk nếu config file bị leak
- **Workaround**: Use environment variables, .gitignore config file
- **TODO**: Integrate với secret management (Vault, AWS Secrets Manager)

**Code location**: `config/config.go`

```yaml
# ❌ Current implementation
internal:
  service_keys:
    project_service: project-service-key-plaintext

# ✅ Proposed solution
# Use environment variables
INTERNAL_SERVICE_KEY_PROJECT_SERVICE=encrypted-or-from-vault
```

### 5. Monitoring - Lack of Metrics

**Issue**: Không có metrics export cho monitoring

- **Current**: Chỉ có structured logs
- **Problem**: Khó monitor performance, error rate, latency
- **Impact**: Không có visibility vào service health real-time
- **Workaround**: Parse logs để extract metrics (không ideal)
- **TODO**: Implement Prometheus metrics export

**Code location**: New file needed `pkg/metrics/prometheus.go`

```go
// ✅ Proposed implementation
var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{Name: "http_requests_total"},
        []string{"method", "endpoint", "status"},
    )
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{Name: "http_request_duration_seconds"},
        []string{"method", "endpoint"},
    )
)
```

### 6. Testing - Missing Integration Tests

**Issue**: Chưa có integration tests cho OAuth flow

- **Current**: Chỉ có unit tests cho individual functions
- **Problem**: Không test end-to-end OAuth flow
- **Impact**: Regression bugs có thể xảy ra khi refactor
- **Workaround**: Manual testing với test client
- **TODO**: Implement integration tests với mock OAuth provider

---

## 🔮 Future Enhancements

### Short-term (1-2 months)

- [ ] Prometheus metrics export - Cần cho production monitoring
- [ ] Rate limiting middleware - Prevent abuse và DDoS
- [ ] 2FA/MFA support - Enhance security cho sensitive accounts
- [ ] Integration tests - Improve code quality và confidence
- [ ] Load testing results - Validate performance targets

### Mid-term (3-6 months)

- [ ] Redis Sentinel/Cluster - High availability
- [ ] Multi-provider OAuth2 - Support Azure AD, Okta
- [ ] User management API - CRUD operations cho ADMIN
- [ ] Distributed tracing - Better debugging và performance analysis
- [ ] Secret management integration - Vault hoặc AWS Secrets Manager

### Long-term (6+ months)

- [ ] Migrate to RS256 JWT - Asymmetric key signing
- [ ] JWT key rotation - Automated key lifecycle management
- [ ] Audit log partitioning - Better performance với large dataset
- [ ] Multi-region deployment - Global availability
- [ ] Advanced analytics - User behavior analysis, anomaly detection

---

## 🔧 Configuration

**File**: `config/auth-config.yaml`

```yaml
# Environment (development = token in JSON, production = cookie only)
environment:
  name: development

# HTTP Server
http_server:
  host: ""
  port: 8080
  mode: debug

# Logger
logger:
  level: debug
  mode: debug
  encoding: console
  color_enabled: true

# PostgreSQL
postgres:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: smap_auth
  sslmode: disable
  schema: schema_identity

# Redis
redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

# Kafka
kafka:
  brokers:
    - localhost:9092
  topic: audit.events

# OAuth2
oauth2:
  provider: google
  client_id: YOUR_GOOGLE_CLIENT_ID
  client_secret: YOUR_GOOGLE_CLIENT_SECRET
  redirect_uri: http://localhost:8080/authentication/callback
  scopes:
    - openid
    - email
    - profile

# JWT
jwt:
  algorithm: HS256
  issuer: smap-auth-service
  audience:
    - identity-srv
  secret_key: your-secret-key-min-32-characters
  ttl: 28800 # 8 hours

# Cookie
cookie:
  domain: localhost
  secure: false # true in production
  samesite: Lax
  max_age: 28800 # 8 hours
  max_age_remember: 604800 # 7 days
  name: smap_auth_token

# Access Control
access_control:
  allowed_domains:
    - gmail.com
    - yourdomain.com
  blocked_emails: []
  allowed_redirect_urls:
    - /dashboard
    - /
    - http://localhost:3000
  user_roles:
    admin@yourdomain.com: ADMIN
    analyst@yourdomain.com: ANALYST
  default_role: VIEWER

# Session
session:
  ttl: 28800 # 8 hours
  remember_me_ttl: 604800 # 7 days
  backend: redis

# Blacklist
blacklist:
  enabled: true
  backend: redis
  key_prefix: "blacklist:"

# Encrypter
encrypter:
  key: your-encryption-key-32-chars

# Internal Service Keys
internal:
  service_keys:
    project_service: project-service-key
    ingest_service: ingest-service-key

# Discord Webhook (Optional)
discord:
  webhook_id: ""
  webhook_token: ""
```

**Environment Variables** (override config file):

```bash
# Required
OAUTH2_CLIENT_ID=xxx.apps.googleusercontent.com
OAUTH2_CLIENT_SECRET=xxx
JWT_SECRET_KEY=your-secret-key-min-32-characters
POSTGRES_PASSWORD=xxx
REDIS_PASSWORD=xxx

# Optional
HTTP_SERVER_PORT=8080
ENVIRONMENT_NAME=production
LOGGER_LEVEL=info
POSTGRES_HOST=localhost
REDIS_HOST=localhost
KAFKA_BROKERS=localhost:9092
```

---

## 📊 Performance Metrics

### Estimated Performance

**Note**: Đây là estimates dựa trên code analysis và local testing, chưa có production load tests

**API Endpoints**:

- `GET /authentication/login`: ~10ms (redirect only)
- `GET /authentication/callback`: ~200-500ms (includes Google API call)
- `POST /authentication/logout`: ~50ms (Redis delete)
- `GET /authentication/me`: ~20ms (single DB query)
- `GET /audit-logs`: ~50-200ms (depends on filters)
- `POST /internal/validate`: ~30ms (JWT verify + Redis check)
- `POST /internal/revoke-token`: ~100ms (multiple Redis ops)
- `GET /internal/users/:id`: ~20ms (single DB query)

**Resource Usage** (estimated for 1000 req/s):

- CPU: ~30-40% (2 cores)
- Memory: ~200-300MB
- Database Connections: ~10-15 active
- Redis Connections: ~5-10 active

**Throughput Estimates**:

- Login flow: ~50 req/s (limited by Google API)
- Token validation: ~2000 req/s (JWT + Redis)
- Audit log queries: ~100 req/s (DB limited)

**TODO**: Run load tests với k6 hoặc Gatling để có accurate numbers

---

## 🔐 Security

### Authentication

- **Method**: OAuth2 (Google) + JWT (HS256)
- **Token Storage**: HttpOnly cookie (browser), Bearer token (mobile/API)
- **Token Expiry**: 8 hours (default), 7 days (remember me)
- **Refresh Strategy**: Re-login via OAuth (no refresh token)

### Authorization

- **Model**: RBAC (Role-Based Access Control)
- **Permissions**:
  - ADMIN: Full access (all endpoints)
  - ANALYST: Create/analyze resources
  - VIEWER: Read-only access
- **Scope Validation**: Middleware checks role from JWT claims

### Data Protection

- **Encryption at Rest**:
  - Role hash: SHA256 + base64
  - Database: PostgreSQL encryption (if enabled)
- **Encryption in Transit**: TLS 1.2+ (production)
- **PII Handling**:
  - Email, name, avatar stored in database
  - No sensitive data in logs
  - Audit logs include IP and user agent
- **Secrets Management**:
  - Config file (development)
  - Environment variables (production)
  - TODO: Vault integration

### Security Best Practices

- [x] Input validation and sanitization (Gin binding)
- [x] SQL injection prevention (SQLBoiler parameterized queries)
- [x] XSS protection (HttpOnly cookies)
- [x] CSRF protection (OAuth state parameter)
- [x] Rate limiting (TODO: implement)
- [x] API authentication (JWT + service keys)
- [x] Audit logging (comprehensive)
- [ ] Dependency scanning (TODO: implement)
- [x] Security headers (CORS, Content-Type)

### Security Headers

```go
// CORS middleware
Access-Control-Allow-Origin: https://app.example.com
Access-Control-Allow-Credentials: true
Access-Control-Allow-Methods: GET, POST, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization

// Cookie security
Set-Cookie: smap_auth_token=xxx; HttpOnly; Secure; SameSite=Lax
```

---

## 🧪 Testing

### Test Coverage

- **Unit Tests**: ~40% coverage (core business logic)
- **Integration Tests**: 0% (TODO: implement)
- **E2E Tests**: Manual testing với test client
- **Load Tests**: Not yet performed

### Running Tests

```bash
# Unit tests
go test ./...

# Unit tests with coverage
go test -cover ./...

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Specific package
go test ./internal/authentication/usecase/...

# Verbose output
go test -v ./...
```

### Test Strategy

- **Unit Tests**:
  - UseCase layer: Business logic validation
  - Repository layer: Database operations (with mock)
  - Middleware: Auth, CORS, error handling
- **Integration Tests** (TODO):
  - OAuth flow end-to-end
  - Token validation with Redis
  - Audit log publishing to Kafka
- **Mocking Strategy**:
  - Database: SQLBoiler mock
  - Redis: go-redis mock
  - OAuth: Mock HTTP client
- **Test Data**:
  - Fixtures in `testdata/` folder
  - Factory functions for model creation

---

## 🚀 Deployment

### Docker

```bash
# Build API image
make docker-build
# Or: docker build -t identity-srv:latest -f cmd/api/Dockerfile .

# Build Consumer image
make consumer-build
# Or: docker build -t smap-consumer:latest -f cmd/consumer/Dockerfile .

# Run API container
docker run -d -p 8080:8080 \
  -v $(pwd)/config:/app/config \
  -e POSTGRES_PASSWORD=xxx \
  -e REDIS_PASSWORD=xxx \
  identity-srv:latest

# Run Consumer container
docker run -d \
  -v $(pwd)/config:/app/config \
  -e POSTGRES_PASSWORD=xxx \
  -e KAFKA_BROKERS=kafka:9092 \
  smap-consumer:latest
```

### Kubernetes

**API Deployment**:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: identity-srv
  namespace: smap
spec:
  replicas: 3
  selector:
    matchLabels:
      app: identity-srv
  template:
    metadata:
      labels:
        app: identity-srv
    spec:
      containers:
        - name: identity-srv
          image: identity-srv:latest
          ports:
            - containerPort: 8080
          env:
            - name: ENVIRONMENT_NAME
              value: "production"
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: identity-secrets
                  key: postgres-password
            - name: REDIS_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: identity-secrets
                  key: redis-password
            - name: JWT_SECRET_KEY
              valueFrom:
                secretKeyRef:
                  name: identity-secrets
                  key: jwt-secret
          livenessProbe:
            httpGet:
              path: /live
              port: 8080
            initialDelaySeconds: 30
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /ready
              port: 8080
            initialDelaySeconds: 10
            periodSeconds: 5
          resources:
            requests:
              memory: "256Mi"
              cpu: "250m"
            limits:
              memory: "512Mi"
              cpu: "500m"
```

**Consumer Deployment**:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: smap-consumer
  namespace: smap
spec:
  replicas: 2
  selector:
    matchLabels:
      app: smap-consumer
  template:
    metadata:
      labels:
        app: smap-consumer
    spec:
      containers:
        - name: smap-consumer
          image: smap-consumer:latest
          env:
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: identity-secrets
                  key: postgres-password
            - name: KAFKA_BROKERS
              value: "kafka-service:9092"
          resources:
            requests:
              memory: "128Mi"
              cpu: "100m"
            limits:
              memory: "256Mi"
              cpu: "200m"
```

**Service**:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: identity-srv
  namespace: smap
spec:
  selector:
    app: identity-srv
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
  type: ClusterIP
```

**Ingress**:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: identity-srv
  namespace: smap
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
    - hosts:
        - identity-srv.example.com
      secretName: identity-srv-tls
  rules:
    - host: identity-srv.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: identity-srv
                port:
                  number: 80
```

### CI/CD Pipeline

**GitHub Actions** (`.github/workflows/deploy.yml`):

```yaml
name: Deploy Identity Service
on:
  push:
    branches: [main]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.25
      - name: Run tests
        run: go test ./...

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Build Docker images
        run: |
          docker build -t ${{ secrets.REGISTRY }}/identity-srv:${{ github.sha }} -f cmd/api/Dockerfile .
          docker build -t ${{ secrets.REGISTRY }}/smap-consumer:${{ github.sha }} -f cmd/consumer/Dockerfile .
      - name: Push to registry
        run: |
          echo ${{ secrets.REGISTRY_PASSWORD }} | docker login -u ${{ secrets.REGISTRY_USER }} --password-stdin
          docker push ${{ secrets.REGISTRY }}/identity-srv:${{ github.sha }}
          docker push ${{ secrets.REGISTRY }}/smap-consumer:${{ github.sha }}

  deploy:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to Kubernetes
        run: |
          kubectl set image deployment/identity-srv identity-srv=${{ secrets.REGISTRY }}/identity-srv:${{ github.sha }}
          kubectl set image deployment/smap-consumer smap-consumer=${{ secrets.REGISTRY }}/smap-consumer:${{ github.sha }}
          kubectl rollout status deployment/identity-srv
          kubectl rollout status deployment/smap-consumer
```

---

## 📚 Documentation

### API Documentation

- **Swagger/OpenAPI**: `http://localhost:8080/swagger/index.html` (when running)
- **Swagger JSON**: `docs/swagger.json`
- **Swagger YAML**: `docs/swagger.yaml`
- **Postman Collection**: TODO (create and link)

### Architecture Docs

- **System Design**: `documents/auth-service-integration.md`
- **Database Schema**: `migration/01_auth_service_schema.sql`
- **Deployment Guide**: `documents/deployment-guide.md`
- **Local Testing Guide**: `documents/local-testing-guide.md`
- **Service Integration Guide**: `documents/service-integration-practical-guide.md`

### Developer Guides

- **Setup Guide**: `README.md` (Quick Start section)
- **Contributing Guide**: TODO (create CONTRIBUTING.md)
- **Code Conventions**: `documents/convention/` folder
  - `convention.md` - General conventions
  - `convention_delivery.md` - HTTP handler conventions
  - `convention_repository.md` - Repository layer conventions
  - `convention_usecase.md` - Business logic conventions

---

## 📞 Contact & Support

### Team

- **Service Owner**: SMAP Platform Team
- **Tech Lead**: [Name]
- **On-call**: #smap-oncall Slack channel

### Resources

- **Repository**: [GitHub/GitLab URL]
- **Issue Tracker**: GitHub Issues
- **Monitoring**: TODO (Grafana dashboard)
- **Logs**: TODO (Kibana/CloudWatch)
- **Slack Channel**: #smap-identity-service

### SLA & Support

- **Availability Target**: 99.9% (planned)
- **Response Time**:
  - P0 (Service down): 15 minutes
  - P1 (Critical bug): 1 hour
  - P2 (Major bug): 4 hours
  - P3 (Minor bug): 1 business day
- **Support Hours**: 24/7 (on-call rotation)

---

## 📝 Changelog

### [2.0.0] - 2026-02-15

**Added**

- OAuth2 authentication với Google
- JWT token generation (HS256)
- Session management với Redis
- Token blacklist cho instant revocation
- Role-based access control (ADMIN, ANALYST, VIEWER)
- Audit logging với PostgreSQL + Kafka
- Internal API cho service-to-service auth
- HttpOnly cookie support
- Bearer token support
- CORS middleware
- Swagger documentation
- Docker deployment
- Kubernetes manifests
- Health check endpoints

**Changed**

- Migrate từ RabbitMQ sang Kafka
- Improve error handling và logging
- Optimize database queries với indexes

**Fixed**

- OAuth state validation bug
- Session cleanup race condition
- CORS preflight handling

**Security**

- Add role encryption
- Implement token blacklist
- Add CSRF protection với OAuth state

---

## 🎓 Learning Resources

### Internal Docs

- [Auth Service Integration](documents/auth-service-integration.md)
- [Deployment Guide](documents/deployment-guide.md)
- [Local Testing Guide](documents/local-testing-guide.md)
- [Service Integration Practical Guide](documents/service-integration-practical-guide.md)

### External Resources

- [OAuth 2.0 RFC 6749](https://tools.ietf.org/html/rfc6749)
- [JWT RFC 7519](https://tools.ietf.org/html/rfc7519)
- [Go Gin Framework](https://gin-gonic.com/docs/)
- [SQLBoiler Documentation](https://github.com/volatiletech/sqlboiler)
- [Redis Best Practices](https://redis.io/docs/manual/patterns/)
- [Kafka Documentation](https://kafka.apache.org/documentation/)

---

## 📋 Appendix

### Glossary

- **OAuth2**: Open Authorization protocol cho delegated access
- **JWT**: JSON Web Token, compact token format cho claims transfer
- **HS256**: HMAC with SHA-256, symmetric signing algorithm
- **RS256**: RSA Signature with SHA-256, asymmetric signing algorithm
- **JTI**: JWT ID, unique identifier cho mỗi token
- **RBAC**: Role-Based Access Control
- **CSRF**: Cross-Site Request Forgery
- **XSS**: Cross-Site Scripting
- **TTL**: Time To Live
- **SLA**: Service Level Agreement

### Related Services

- **Project Service**: Sử dụng Identity Service để validate user tokens
- **Ingest Service**: Sử dụng Internal API để get user info
- **Knowledge Service**: Sử dụng JWT validation cho authorization
- **Notification Service**: Sử dụng audit logs để trigger notifications

### Migration Guides

- **v1 to v2**: Migrate từ RabbitMQ sang Kafka
  - Update config: Replace `rabbitmq` section với `kafka`
  - Update consumer: Use Sarama instead of amqp091-go
  - No API changes, backward compatible

---

**Document Version**: 1.0  
**Last Updated**: 15/02/2026  
**Maintained By**: SMAP Platform Team  
**Review Cycle**: Monthly

---

## 📌 Quick Reference Card

```
Service: Identity Service (identity-srv)
Port: 8080
Health: GET /health
Ready: GET /ready
Live: GET /live
Swagger: http://localhost:8080/swagger/index.html

Quick Start:
1. Clone repo: git clone <repo-url>
2. Copy config: cp config/auth-config.example.yaml config/auth-config.yaml
3. Edit config: Add Google OAuth credentials
4. Setup DB: psql -f migration/01_auth_service_schema.sql
5. Run API: make run-api
6. Run Consumer: make run-consumer
7. Test: open http://localhost:8080/authentication/login

Common Commands:
- Start API: make run-api
- Start Consumer: make run-consumer
- Generate Swagger: make swagger
- Build Docker: make docker-build
- Build Consumer: make consumer-build
- Run tests: go test ./...

Emergency Contacts:
- On-call: #smap-oncall
- Escalation: Platform Team Lead

Key Endpoints:
- Login: GET /authentication/login
- Callback: GET /authentication/callback
- Logout: POST /authentication/logout
- Me: GET /authentication/me
- Validate: POST /internal/validate
- Audit Logs: GET /audit-logs (ADMIN only)

Dependencies:
- PostgreSQL: localhost:5432 (schema_identity)
- Redis: localhost:6379 (DB 0)
- Kafka: localhost:9092 (topic: audit.events)
- Google OAuth: accounts.google.com

Configuration:
- File: config/auth-config.yaml
- Env vars: Override config values
- Required: OAUTH2_CLIENT_ID, OAUTH2_CLIENT_SECRET, JWT_SECRET_KEY
```
