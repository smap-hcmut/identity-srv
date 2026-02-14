# Trạng Thái Hiện Tại của Repository - SMAP Identity Service

**Ngày cập nhật**: 14/02/2026  
**Phiên bản**: 2.0.0 (Simplified)

---

## 📋 Tổng Quan

SMAP Identity Service là một **authentication & authorization service** cho nền tảng SMAP, sử dụng OAuth2 (Google) và JWT để xác thực người dùng. Sau quá trình simplification, service đã được tối ưu hóa để phù hợp với graduation project.

### Kiến Trúc Tổng Thể

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│   Browser   │ Cookie  │ API Service  │  JWT    │   Other     │
│             │────────▶│   (Port      │────────▶│  Services   │
│             │         │    8080)     │         │             │
└─────────────┘         └──────┬───────┘         └─────────────┘
                               │
                    ┌──────────┼──────────┐
                    ▼          ▼          ▼
               ┌─────────┐ ┌──────┐ ┌────────┐
               │Postgres │ │Redis │ │ Kafka  │
               │(Users,  │ │(DB 0)│ │(Audit) │
               │ Audit)  │ │      │ │        │
               └─────────┘ └──────┘ └────────┘

┌──────────────────┐
│ Consumer Service │────▶ Xử lý Audit Log từ Kafka
└──────────────────┘

┌──────────────────┐
│  Test Client     │────▶ HTML test page cho OAuth flow
└──────────────────┘
```

---

## 🏗️ Cấu Trúc Dự Án

### 1. Services (3 services)

#### **API Service** (`cmd/api/`)
- **Mục đích**: Main HTTP server xử lý authentication
- **Port**: 8080
- **Chức năng**:
  - OAuth2 login với Google
  - JWT token generation (HS256)
  - Session management (Redis)
  - Token blacklist
  - Internal API cho service-to-service auth
  - Health check endpoint
  - Swagger documentation

#### **Consumer Service** (`cmd/consumer/`)
- **Mục đích**: Kafka consumer xử lý async tasks
- **Chức năng**:
  - Consume audit events từ Kafka
  - Lưu audit logs vào PostgreSQL
  - Có thể mở rộng cho các async tasks khác

#### **Test Client** (`cmd/test-client/`)
- **Mục đích**: Simple HTML page để test OAuth flow
- **Chức năng**:
  - Test login flow
  - Display user info
  - Test logout

### 2. Internal Modules (`internal/`)

#### **Authentication Module** (`internal/authentication/`)
```
authentication/
├── delivery/
│   └── http/          # HTTP handlers, routes, presenters
├── repository/
│   └── postgre/       # Database operations (nếu cần)
├── usecase/           # Business logic
│   ├── authentication.go  # Core auth logic
│   ├── blacklist.go       # Token blacklist
│   ├── oauth.go           # OAuth flow
│   ├── redirect.go        # Redirect validation
│   ├── roles.go           # Email-to-role mapping
│   ├── session.go         # Session management
│   └── util.go            # Helper functions
├── errors.go          # Custom errors
├── interface.go       # UseCase interface
└── type.go            # Domain types
```

**Chức năng chính**:
- OAuth2 login flow (Google)
- JWT token generation & verification (HS256)
- Session management (Redis-backed)
- Token blacklist (instant revocation)
- Email-to-role mapping (từ config)
- Redirect URL validation

#### **Audit Module** (`internal/audit/`)
```
audit/
├── delivery/
│   ├── http/          # HTTP handlers (nếu có API)
│   └── kafka/
│       ├── consumer/  # Kafka consumer
│       └── producer/  # Kafka producer
├── repository/
│   └── postgre/       # Audit log database operations
├── usecase/           # Business logic
├── error.go
├── interface.go
└── type.go
```

**Chức năng chính**:
- Publish audit events to Kafka
- Consume audit events from Kafka
- Store audit logs in PostgreSQL
- Query audit logs (90-day retention)

#### **User Module** (`internal/user/`)
```
user/
├── repository/
│   └── postgre/       # User CRUD operations
├── usecase/           # User business logic
├── error.go
├── interface.go
└── type.go
```

**Chức năng chính**:
- User CRUD operations
- User profile management
- User lookup by email/ID

#### **HTTP Server** (`internal/httpserver/`)
- HTTP server initialization
- Route registration
- Middleware setup
- Dependency injection

#### **Middleware** (`internal/middleware/`)
- **CORS**: Cross-origin resource sharing
- **Admin**: Admin-only endpoint protection
- **Service Auth**: Internal service authentication
- **Locale**: i18n support
- **Recovery**: Panic recovery

#### **Model** (`internal/model/`)
- Domain models: User, AuditLog, Role, Scope
- Business logic types

#### **SQLBoiler** (`internal/sqlboiler/`)
- Auto-generated database models
- Type-safe database queries

### 3. Packages (`pkg/`)

#### **Authentication & Security**
- **`jwt/`**: HS256 JWT generation & verification
- **`oauth/`**: OAuth2 providers (Google, Azure, Okta)
- **`auth/`**: Auth middleware & helpers
- **`encrypter/`**: Password hashing, encryption
- **`scope/`**: JWT scope management

#### **Database & Cache**
- **`postgre/`**: PostgreSQL utilities
- **`redis/`**: Redis client wrapper

#### **Message Queue**
- **`kafka/`**: Kafka producer & consumer

#### **Utilities**
- **`log/`**: Zap logger wrapper
- **`response/`**: HTTP response helpers
- **`errors/`**: Custom error types
- **`i18n/`**: Internationalization
- **`locale/`**: Locale detection
- **`paginator/`**: Pagination helper
- **`util/`**: Common utilities (OTP, time, validation)

#### **Integrations**
- **`discord/`**: Discord webhook notifications
- **`email/`**: Email sending (nếu cần)
- **`compressor/`**: Data compression

#### **Empty Packages** (có thể xóa)
- **`google/`**: ❌ Đã xóa Google Directory API client
- **`minio/`**: ❌ Không sử dụng MinIO

---

## 🗄️ Database Schema

### **PostgreSQL Tables**

#### 1. **users** table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    avatar_url TEXT,
    role_hash VARCHAR(255) NOT NULL,  -- Encrypted role
    is_active BOOLEAN DEFAULT TRUE,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
```

**Indexes**:
- `idx_users_email` (email lookup)
- `idx_users_is_active` (active users)

#### 2. **audit_logs** table
```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    action VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50),
    resource_id UUID,
    metadata JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL  -- 90-day retention
);
```

**Indexes**:
- `idx_audit_logs_user_id`
- `idx_audit_logs_action`
- `idx_audit_logs_created_at`
- `idx_audit_logs_expires_at` (cleanup)
- `idx_audit_logs_resource`

#### 3. **jwt_keys** table ⚠️
```sql
CREATE TABLE jwt_keys (
    kid VARCHAR(50) PRIMARY KEY,
    private_key TEXT NOT NULL,
    public_key TEXT NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ,
    retired_at TIMESTAMPTZ
);
```

**⚠️ QUAN TRỌNG**: Table này **không còn được sử dụng** sau khi chuyển sang HS256. Cần chạy migration để drop:

```bash
make migrate-simplify
# hoặc
psql -d smap_auth -f scripts/drop-jwt-keys-table.sql
```

### **Redis Data Structures**

**Database**: DB 0 (single database cho cả session và blacklist)

#### 1. Session Storage
```
Key: session:{user_id}
Type: String (JSON)
TTL: 8 hours (28800s)
Value: {
  "user_id": "uuid",
  "jti": "jwt-id",
  "created_at": "timestamp",
  "expires_at": "timestamp"
}
```

#### 2. Token Blacklist
```
Key: blacklist:{jti}
Type: String
TTL: Remaining token lifetime
Value: "revoked"
```

---

## ⚙️ Configuration

### Config Structure (`config/auth-config.yaml`)

```yaml
# Environment
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

# Redis (Single DB)
redis:
  host: localhost
  port: 6379
  password: ""
  db: 0  # ✅ Single DB for session + blacklist

# Kafka
kafka:
  brokers:
    - localhost:9092
  topic: audit.events

# OAuth2 (Google)
oauth2:
  provider: google
  client_id: YOUR_GOOGLE_CLIENT_ID
  client_secret: YOUR_GOOGLE_CLIENT_SECRET
  redirect_uri: http://localhost:8080/authentication/callback
  scopes:
    - openid
    - email
    - profile

# JWT (HS256 Symmetric Key)
jwt:
  algorithm: HS256  # ✅ Simplified from RS256
  issuer: smap-auth-service
  audience:
    - smap-api
  secret_key: smap-jwt-secret-key-2024-minimum-32-characters-required
  ttl: 28800  # 8 hours

# Cookie
cookie:
  domain: localhost
  secure: false
  samesite: Lax
  max_age: 28800
  max_age_remember: 604800
  name: smap_auth_token

# Access Control (Email-to-Role Mapping)
access_control:
  allowed_domains:
    - gmail.com
    - vinfast.com
  blocked_emails: []
  user_roles:  # ✅ Direct email-to-role mapping
    admin@vinfast.com: ADMIN
    analyst@vinfast.com: ANALYST
    viewer@vinfast.com: VIEWER
  default_role: VIEWER

# Session
session:
  ttl: 28800
  remember_me_ttl: 604800
  backend: redis

# Token Blacklist
blacklist:
  enabled: true
  backend: redis
  key_prefix: "blacklist:"

# Encrypter
encrypter:
  key: test-encryption-key-32-characters

# Internal Service Auth
internal:
  service_keys:
    project_service: project-service-key
    ingest_service: ingest-service-key
    knowledge_service: knowledge-service-key
    notification_service: notification-service-key

# Discord (Optional)
discord:
  webhook_id: ""
  webhook_token: ""
```

---

## 🔐 Authentication Flow

### 1. OAuth2 Login Flow

```
1. User clicks "Login" → Browser redirects to /authentication/login
2. API redirects to Google OAuth consent screen
3. User approves → Google redirects to /authentication/callback?code=xxx
4. API exchanges code for user info (email, name, avatar)
5. API validates email domain (allowed_domains)
6. API maps email to role (user_roles config)
7. API creates/updates user in database
8. API generates JWT token (HS256)
9. API stores session in Redis
10. API sets HttpOnly cookie with JWT
11. API redirects to frontend
```

### 2. JWT Token Structure

**Header**:
```json
{
  "alg": "HS256",
  "typ": "JWT"
}
```

**Payload**:
```json
{
  "sub": "user-uuid",
  "email": "user@example.com",
  "role": "ADMIN",
  "groups": [],
  "iss": "smap-auth-service",
  "aud": ["smap-api"],
  "exp": 1708012800,
  "iat": 1707984000,
  "jti": "unique-jwt-id"
}
```

**Signature**: HMACSHA256(base64UrlEncode(header) + "." + base64UrlEncode(payload), secret_key)

### 3. Token Verification Flow

```
1. Client sends request with cookie
2. Middleware extracts JWT from cookie
3. Middleware verifies JWT signature (HS256)
4. Middleware checks token expiration
5. Middleware checks blacklist (Redis)
6. Middleware extracts user info from claims
7. Request proceeds with user context
```

### 4. Logout Flow

```
1. Client sends POST /authentication/logout
2. API extracts JTI from JWT
3. API adds JTI to blacklist (Redis)
4. API deletes session from Redis
5. API clears cookie
6. Token is immediately revoked
```

---

## 🔑 Role-Based Access Control (RBAC)

### Roles

| Role | Permissions | Use Case |
|------|------------|----------|
| **ADMIN** | Full access | System administrators |
| **ANALYST** | Create, read, analyze | Data analysts |
| **VIEWER** | Read-only | Stakeholders, viewers |

### Email-to-Role Mapping

**Config** (`config/auth-config.yaml`):
```yaml
access_control:
  user_roles:
    admin@vinfast.com: ADMIN
    analyst@vinfast.com: ANALYST
    viewer@vinfast.com: VIEWER
  default_role: VIEWER  # For unmapped emails
```

**Logic** (`internal/authentication/usecase/roles.go`):
```go
func (rm *RoleMapper) MapEmailToRole(email string) string {
    if role, ok := rm.userRoles[email]; ok {
        return role
    }
    return rm.defaultRole
}
```

---

## 📡 API Endpoints

### Public Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/authentication/login` | Redirect to Google OAuth |
| GET | `/authentication/callback` | OAuth callback handler |
| GET | `/health` | Health check |
| GET | `/swagger/*` | API documentation |

### Protected Endpoints (Cookie Required)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/authentication/logout` | Logout user |
| GET | `/authentication/me` | Get current user info |

### Internal Endpoints (Service-to-Service)

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | `/internal/validate` | Validate JWT token | Service key |
| POST | `/internal/revoke-token` | Revoke token | Admin only |
| GET | `/internal/users/:id` | Get user by ID | Service key |

---

## 🛠️ Development Commands

### Makefile Targets

```bash
# Development
make models              # Generate SQLBoiler models
make swagger             # Generate Swagger docs
make run-api             # Run API server
make run-consumer        # Run consumer service

# Docker - API
make docker-build        # Build API image (local platform)
make docker-build-amd64  # Build for AMD64 servers
make docker-run          # Build and run container
make docker-push         # Push to registry

# Docker - Consumer
make consumer-build      # Build consumer image
make consumer-run        # Build and run consumer
make consumer-push       # Push to registry

# Database
make migrate-simplify    # Drop jwt_keys table (after HS256 migration)
```

### Manual Scripts

```bash
# Cleanup old audit logs (90+ days)
./scripts/cleanup-audit-logs.sh

# Drop jwt_keys table
psql -d smap_auth -f scripts/drop-jwt-keys-table.sql
```

---

## 📦 Dependencies

### Core Dependencies

```go
// Web Framework
github.com/gin-gonic/gin v1.11.0

// JWT
github.com/golang-jwt/jwt/v5 v5.3.1

// OAuth2
golang.org/x/oauth2 v0.34.0

// Database
github.com/lib/pq v1.10.9                    // PostgreSQL driver
github.com/aarondl/sqlboiler/v4 v4.19.5      // ORM

// Cache
github.com/redis/go-redis/v9 v9.17.3

// Message Queue
github.com/IBM/sarama v1.46.3                // Kafka client

// Logging
go.uber.org/zap v1.27.0

// Config
github.com/spf13/viper v1.19.0

// Utilities
github.com/google/uuid v1.6.0
golang.org/x/crypto v0.47.0
```

### Removed Dependencies (After Simplification)

- ❌ `google.golang.org/api` (Google Directory API)
- ❌ Rate limiting libraries
- ❌ RS256 crypto libraries (nếu không dùng cho mục đích khác)

---

## 🚀 Deployment

### Docker Images

**API Service**:
```bash
docker build -t smap-api:latest -f cmd/api/Dockerfile .
docker run -d -p 8080:8080 \
  -v $(pwd)/config:/app/config \
  smap-api:latest
```

**Consumer Service**:
```bash
docker build -t smap-consumer:latest -f cmd/consumer/Dockerfile .
docker run -d \
  -v $(pwd)/config:/app/config \
  smap-consumer:latest
```

### Kubernetes

```bash
kubectl apply -f cmd/api/deployment.yaml
kubectl apply -f cmd/consumer/deployment.yaml
```

---

## 📊 Metrics & Monitoring

### Health Check

```bash
curl http://localhost:8080/health
```

**Response**:
```json
{
  "status": "ok",
  "timestamp": "2026-02-14T10:00:00Z"
}
```

### Audit Logging

**Actions tracked**:
- `LOGIN` - Successful login
- `LOGIN_FAILED` - Failed login attempt
- `LOGOUT` - User logout
- `TOKEN_REVOKED` - Token revoked by admin
- Custom actions from other services

**Retention**: 90 days (auto-cleanup via manual script)

---

## 🔧 Troubleshooting

### Common Issues

**1. JWT verification fails**
```bash
# Check secret key length (min 32 chars)
# Check algorithm is HS256
```

**2. Cookie not being set**
```bash
# Set cookie.secure: false for HTTP (dev only)
# Verify cookie.domain matches your domain
```

**3. Cannot connect to PostgreSQL**
```bash
docker ps | grep postgres
psql -h localhost -U postgres -d smap_auth
```

**4. Redis connection error**
```bash
redis-cli ping
# Check redis.db is 0
```

---

## 📝 TODO / Known Issues

### Cần Làm

1. ✅ **Drop jwt_keys table**: Chạy `make migrate-simplify`
2. ⚠️ **Xóa empty packages**: `pkg/google/`, `pkg/minio/`
3. ⚠️ **Update SQLBoiler models**: Regenerate sau khi drop jwt_keys table
4. ⚠️ **Cleanup unused imports**: Run `go mod tidy`

### Có Thể Cải Thiện

- [ ] Add unit tests cho JWT manager
- [ ] Add integration tests cho OAuth flow
- [ ] Setup cron job cho audit log cleanup
- [ ] Add Prometheus metrics
- [ ] Add distributed tracing (OpenTelemetry)
- [ ] Add rate limiting cho internal endpoints (nếu cần)

---

## 📚 Documentation Files

| File | Description |
|------|-------------|
| `README.md` | Main documentation |
| `docs/QUICK_START.md` | Quick start guide |
| `docs/GOOGLE_OAUTH_SETUP.md` | Google OAuth setup |
| `documents/api-reference.md` | API reference |
| `documents/auth-service-integration.md` | Integration guide |
| `documents/deployment-guide.md` | Deployment guide |
| `documents/identity-service-troubleshooting.md` | Troubleshooting |
| `documents/SIMPLIFICATION_PLAN.md` | Simplification plan |
| `documents/CURRENT_REPOSITORY_STATE.md` | This file |

---

## 🎯 Summary

### Điểm Mạnh

✅ **Đơn giản hóa thành công**: Giảm từ 4 services xuống 3  
✅ **JWT đơn giản**: HS256 thay vì RS256 (dễ test, dễ deploy)  
✅ **Role mapping đơn giản**: Email-to-role từ config (không cần Google Groups)  
✅ **Single Redis DB**: Merge session + blacklist vào DB 0  
✅ **No rate limiting**: Phù hợp với graduation project  
✅ **Manual cleanup scripts**: Thay thế Scheduler service  

### Điểm Cần Lưu Ý

⚠️ **jwt_keys table**: Cần drop sau khi migrate  
⚠️ **Empty packages**: Có thể xóa `pkg/google/`, `pkg/minio/`  
⚠️ **Audit cleanup**: Manual script (không tự động)  
⚠️ **Secret key**: Cần secure trong production (env variable)  

---

**Version**: 2.0.0 (Simplified)  
**Last Updated**: 14/02/2026  
**Author**: Kiro AI Assistant
