# Chức Năng và Tính Năng - SMAP Identity Service

**Ngày cập nhật**: 14/02/2026  
**Phiên bản**: 2.0.0 (Simplified)

---

## 📋 Tổng Quan Chức Năng

SMAP Identity Service cung cấp các chức năng authentication và authorization cho toàn bộ hệ thống SMAP. Service này hoạt động như một **centralized authentication gateway** cho tất cả các microservices khác.

---

## 🔐 1. AUTHENTICATION (Xác Thực)

### 1.1. OAuth2 Login với Google

**Mô tả**: Người dùng đăng nhập bằng tài khoản Google Workspace/Gmail thông qua OAuth2 flow.

**Flow hoạt động**:
```
1. User click "Login" → GET /authentication/login
2. Service redirect đến Google OAuth consent screen
3. User approve permissions
4. Google redirect về → GET /authentication/callback?code=xxx
5. Service exchange code → lấy user info (email, name, avatar)
6. Service validate domain (allowed_domains)
7. Service check blocklist (blocked_emails)
8. Service map email → role (từ config)
9. Service tạo/update user trong database
10. Service generate JWT token (HS256)
11. Service tạo session trong Redis
12. Service set HttpOnly cookie
13. Service publish audit event (LOGIN)
14. Redirect về frontend với cookie
```

**Endpoints**:
- `GET /authentication/login` - Khởi tạo OAuth flow
- `GET /authentication/callback` - Xử lý OAuth callback

**Security Features**:
- ✅ Domain validation (chỉ cho phép email từ domains được config)
- ✅ Email blocklist (chặn specific emails)
- ✅ HttpOnly cookie (chống XSS)
- ✅ State parameter (chống CSRF)
- ✅ Audit logging (track mọi login attempt)

**Config**:
```yaml
oauth2:
  provider: google
  client_id: YOUR_CLIENT_ID
  client_secret: YOUR_SECRET
  redirect_uri: http://localhost:8080/authentication/callback
  scopes:
    - openid
    - email
    - profile

access_control:
  allowed_domains:
    - gmail.com
    - vinfast.com
  blocked_emails:
    - blocked@example.com
```

**Supported Providers**:
- ✅ Google (hiện tại đang dùng)
- ✅ Azure AD (code có sẵn, chưa config)
- ✅ Okta (code có sẵn, chưa config)

---

### 1.2. JWT Token Generation (HS256)

**Mô tả**: Sau khi login thành công, service generate JWT token với thuật toán HS256 (symmetric key).

**Token Structure**:

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
  "sub": "550e8400-e29b-41d4-a716-446655440000",  // User ID
  "email": "user@vinfast.com",
  "role": "ADMIN",
  "groups": [],  // Empty (không dùng Google Groups nữa)
  "iss": "smap-auth-service",
  "aud": ["smap-api"],
  "exp": 1708012800,  // Expiration time (8 hours)
  "iat": 1707984000,  // Issued at
  "jti": "unique-jwt-id-for-revocation"  // JWT ID (dùng cho blacklist)
}
```

**Signature**: 
```
HMACSHA256(
  base64UrlEncode(header) + "." + base64UrlEncode(payload),
  secret_key
)
```

**Features**:
- ✅ HS256 symmetric signing (đơn giản, nhanh)
- ✅ Secret key từ config (min 32 characters)
- ✅ TTL configurable (default 8 hours)
- ✅ JTI (JWT ID) cho token revocation
- ✅ Role-based claims
- ✅ Audience validation

**Config**:
```yaml
jwt:
  algorithm: HS256
  issuer: smap-auth-service
  audience:
    - smap-api
  secret_key: smap-jwt-secret-key-2024-minimum-32-characters-required
  ttl: 28800  # 8 hours
```

---

### 1.3. Session Management (Redis-backed)

**Mô tả**: Service lưu session information trong Redis để track active sessions và hỗ trợ logout.

**Session Data Structure**:
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "jti": "unique-jwt-id",
  "created_at": "2026-02-14T10:00:00Z",
  "expires_at": "2026-02-14T18:00:00Z"
}
```

**Redis Keys**:
```
session:{jti}                    → Session data (JSON)
user_sessions:{user_id}          → List of JTIs for user (JSON array)
```

**Features**:
- ✅ **Create Session**: Lưu session khi login
- ✅ **Get Session**: Lấy session info by JTI
- ✅ **Delete Session**: Xóa session khi logout
- ✅ **Get All User Sessions**: Lấy tất cả sessions của 1 user
- ✅ **Delete User Sessions**: Xóa tất cả sessions của 1 user (logout all devices)
- ✅ **Session Exists**: Check session còn tồn tại không
- ✅ **Remember Me**: TTL 7 days thay vì 8 hours
- ✅ **Auto Cleanup**: Redis TTL tự động xóa expired sessions

**TTL**:
- Normal session: 8 hours (28800s)
- Remember me: 7 days (604800s)

**Use Cases**:
1. **Single Logout**: User logout → xóa session hiện tại
2. **Logout All Devices**: Admin revoke all user tokens → xóa tất cả sessions
3. **Session Validation**: Check session còn valid không (ngoài JWT verification)

---

### 1.4. Token Blacklist (Instant Revocation)

**Mô tả**: Service hỗ trợ revoke (thu hồi) JWT tokens ngay lập tức bằng cách thêm JTI vào blacklist.

**Blacklist Structure**:
```
Redis Key: blacklist:{jti}
Value: "1"
TTL: Remaining token lifetime
```

**Features**:
- ✅ **Add Token**: Thêm token vào blacklist
- ✅ **Add All User Tokens**: Thêm tất cả tokens của user vào blacklist
- ✅ **Is Blacklisted**: Check token có bị blacklist không
- ✅ **Remove Token**: Xóa token khỏi blacklist (rarely used)
- ✅ **Auto Expire**: Redis TTL tự động xóa expired blacklist entries

**Use Cases**:
1. **User Logout**: Token bị blacklist ngay lập tức
2. **Admin Revoke**: Admin có thể revoke token của user khác
3. **Security Incident**: Revoke all tokens của compromised user
4. **Account Suspension**: Revoke all tokens khi suspend account

**Flow**:
```
1. User/Admin request revoke token
2. Service extract JTI from JWT
3. Service calculate remaining TTL
4. Service add JTI to Redis blacklist with TTL
5. Service delete session from Redis
6. Token immediately invalid
```

**Verification Flow**:
```
1. Request comes with JWT token
2. Middleware verify JWT signature
3. Middleware check JWT expiration
4. Middleware check JTI in blacklist ← KEY STEP
5. If blacklisted → 401 Unauthorized
6. If not blacklisted → Allow request
```

---

### 1.5. Cookie-based Authentication

**Mô tả**: JWT token được lưu trong HttpOnly cookie để bảo vệ khỏi XSS attacks.

**Cookie Configuration**:
```yaml
cookie:
  domain: localhost
  secure: false  # true in production (HTTPS only)
  samesite: Lax  # CSRF protection
  max_age: 28800  # 8 hours
  max_age_remember: 604800  # 7 days
  name: smap_auth_token
```

**Security Features**:
- ✅ **HttpOnly**: JavaScript không thể access cookie (chống XSS)
- ✅ **Secure**: Chỉ gửi qua HTTPS (production)
- ✅ **SameSite**: Chống CSRF attacks
- ✅ **Domain**: Restrict cookie scope
- ✅ **Max Age**: Auto expire

**Cookie Lifecycle**:
```
Login → Set cookie with JWT
Request → Browser auto send cookie
Logout → Expire cookie (max-age=0)
```

---

## 🔑 2. AUTHORIZATION (Phân Quyền)

### 2.1. Role-Based Access Control (RBAC)

**Mô tả**: Service hỗ trợ 3 roles với permissions khác nhau.

**Roles**:

| Role | Level | Permissions | Use Case |
|------|-------|-------------|----------|
| **ADMIN** | 3 | Full access - Tất cả operations | System administrators, DevOps |
| **ANALYST** | 2 | Create, Read, Analyze - Không delete | Data analysts, Business users |
| **VIEWER** | 1 | Read-only - Chỉ xem | Stakeholders, Managers |

**Role Assignment Flow**:
```
1. User login với email
2. Service check email trong user_roles config
3. If found → assign mapped role
4. If not found → assign default_role (VIEWER)
5. Role được lưu trong JWT claims
6. Role được encrypt và lưu trong database
```

**Config**:
```yaml
access_control:
  user_roles:
    admin@vinfast.com: ADMIN
    analyst@vinfast.com: ANALYST
    viewer@vinfast.com: VIEWER
    tantai@vinfast.com: ADMIN
  default_role: VIEWER
```

**Features**:
- ✅ **Email-to-Role Mapping**: Direct mapping từ config (đơn giản)
- ✅ **Default Role**: Fallback cho unmapped emails
- ✅ **Role in JWT**: Role được embed trong token
- ✅ **Role Encryption**: Role được encrypt trong database
- ✅ **Dynamic Update**: Update config → restart service → new roles apply

**Middleware Protection**:
```go
// Admin-only endpoint
r.POST("/admin/users", mw.Admin(), handler.CreateUser)

// Authenticated endpoint (any role)
r.GET("/me", mw.Auth(), handler.GetMe)

// Public endpoint (no auth)
r.GET("/health", handler.Health)
```

---

### 2.2. Domain Validation

**Mô tả**: Chỉ cho phép users từ specific email domains login.

**Config**:
```yaml
access_control:
  allowed_domains:
    - gmail.com
    - vinfast.com
    - yourdomain.com
```

**Validation Logic**:
```
1. User login với email: user@example.com
2. Extract domain: example.com
3. Check domain in allowed_domains list
4. If not found → Reject with ErrDomainNotAllowed
5. If found → Continue authentication
```

**Use Cases**:
- ✅ Restrict access to company employees only
- ✅ Multi-tenant support (different domains)
- ✅ Prevent unauthorized access

---

### 2.3. Email Blocklist

**Mô tả**: Block specific emails từ việc login (blacklist).

**Config**:
```yaml
access_control:
  blocked_emails:
    - blocked@example.com
    - suspended@vinfast.com
```

**Validation Logic**:
```
1. User login với email
2. Check email in blocked_emails list
3. If found → Reject with ErrAccountBlocked
4. If not found → Continue authentication
```

**Use Cases**:
- ✅ Suspend specific user accounts
- ✅ Block malicious users
- ✅ Temporary access restriction

---

### 2.4. Redirect URL Validation

**Mô tả**: Validate redirect URLs sau OAuth callback để prevent open redirect attacks.

**Config**:
```yaml
access_control:
  allowed_redirect_urls:
    - http://localhost:3000
    - https://smap.vinfast.com
```

**Validation Logic**:
```
1. OAuth callback với redirect_url parameter
2. Check redirect_url in allowed list
3. If not found → Use default redirect
4. If found → Redirect to specified URL
```

**Security**: Chống open redirect vulnerability.

---

## 👤 3. USER MANAGEMENT

### 3.1. User CRUD Operations

**Features**:
- ✅ **Create User**: Tự động tạo user khi first login
- ✅ **Update User**: Update name, avatar, role
- ✅ **Get User**: Get user by ID hoặc email
- ✅ **User Profile**: Get current user info

**User Model**:
```go
type User struct {
    ID           string    // UUID
    Email        string    // Unique
    Name         string
    AvatarURL    string
    RoleHash     string    // Encrypted role
    IsActive     bool
    LastLoginAt  time.Time
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

**Database Table**:
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    avatar_url TEXT,
    role_hash VARCHAR(255) NOT NULL,  -- Encrypted
    is_active BOOLEAN DEFAULT TRUE,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
```

**Endpoints**:
- `GET /authentication/me` - Get current user (protected)
- `GET /internal/users/:id` - Get user by ID (internal)

---

### 3.2. User Profile Information

**Mô tả**: Lấy thông tin user hiện tại từ JWT token.

**Endpoint**: `GET /authentication/me`

**Request**:
```bash
curl http://localhost:8080/authentication/me \
  -H "Cookie: smap_auth_token=<JWT_TOKEN>"
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@vinfast.com",
    "name": "John Doe",
    "avatar_url": "https://lh3.googleusercontent.com/...",
    "role": "ADMIN",
    "is_active": true,
    "last_login_at": "2026-02-14T10:00:00Z",
    "created_at": "2026-01-01T00:00:00Z"
  }
}
```

---

## 📊 4. AUDIT LOGGING

### 4.1. Event Tracking

**Mô tả**: Track tất cả authentication events và user actions để compliance và security monitoring.

**Tracked Events**:
- ✅ `LOGIN` - Successful login
- ✅ `LOGIN_FAILED` - Failed login (domain not allowed, account blocked)
- ✅ `LOGOUT` - User logout
- ✅ `TOKEN_REVOKED` - Token revoked by admin
- ✅ Custom events từ other services

**Audit Event Structure**:
```go
type AuditEvent struct {
    UserID       string            // User ID or email
    Action       string            // LOGIN, LOGOUT, etc.
    ResourceType string            // authentication, project, etc.
    ResourceID   string            // Resource UUID (optional)
    Metadata     map[string]string // Additional context
    IPAddress    string
    UserAgent    string
    Timestamp    time.Time
}
```

**Database Table**:
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

---

### 4.2. Kafka-based Async Processing

**Mô tả**: Audit events được publish to Kafka để async processing, không block main request flow.

**Architecture**:
```
API Service → Kafka Producer → Kafka Topic (audit.events)
                                      ↓
                              Consumer Service → PostgreSQL
```

**Features**:
- ✅ **Non-blocking**: Publish không block HTTP response
- ✅ **Buffering**: In-memory buffer khi Kafka unavailable
- ✅ **Auto Retry**: Consumer tự động retry failed messages
- ✅ **Partitioning**: Partition by user_id
- ✅ **Scalable**: Có thể scale consumer independently

**Kafka Config**:
```yaml
kafka:
  brokers:
    - localhost:9092
  topic: audit.events
```

**Buffer Logic**:
```
1. Try publish to Kafka
2. If Kafka unavailable → Add to in-memory buffer (max 10000 events)
3. When Kafka available → Flush buffer
4. If buffer full → Drop oldest events
```

---

### 4.3. Audit Log Retention (90 days)

**Mô tả**: Audit logs tự động expire sau 90 ngày để comply với data retention policies.

**Retention Strategy**:
- ✅ **expires_at column**: Mỗi log có expires_at = created_at + 90 days
- ✅ **Manual Cleanup**: Script để delete expired logs
- ✅ **Index**: Index trên expires_at để cleanup nhanh

**Cleanup Script**:
```bash
# Manual cleanup
./scripts/cleanup-audit-logs.sh

# Or SQL directly
psql -d smap_auth -f scripts/cleanup-audit-logs.sql
```

**SQL**:
```sql
DELETE FROM audit_logs
WHERE created_at < NOW() - INTERVAL '90 days';
```

**Recommended**: Setup cron job để auto cleanup:
```bash
# Crontab: Run daily at 2 AM
0 2 * * * /path/to/scripts/cleanup-audit-logs.sh
```

---

## 🔌 5. INTERNAL API (Service-to-Service)

### 5.1. Token Validation

**Mô tả**: Other services có thể validate JWT tokens thông qua internal API.

**Endpoint**: `POST /internal/validate`

**Request**:
```bash
curl -X POST http://localhost:8080/internal/validate \
  -H "Content-Type: application/json" \
  -d '{"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."}'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "valid": true,
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@vinfast.com",
    "role": "ADMIN",
    "groups": [],
    "expires_at": "2026-02-14T18:00:00Z"
  }
}
```

**Validation Steps**:
1. Verify JWT signature (HS256)
2. Check JWT expiration
3. Check JTI in blacklist
4. Return validation result

---

### 5.2. Token Revocation (Admin)

**Mô tả**: Admin có thể revoke tokens của users khác.

**Endpoint**: `POST /internal/revoke-token`

**Auth**: Requires ADMIN role

**Request (Revoke specific token)**:
```bash
curl -X POST http://localhost:8080/internal/revoke-token \
  -H "Content-Type: application/json" \
  -H "Cookie: smap_auth_token=<ADMIN_TOKEN>" \
  -d '{"jti": "unique-jwt-id"}'
```

**Request (Revoke all user tokens)**:
```bash
curl -X POST http://localhost:8080/internal/revoke-token \
  -H "Content-Type: application/json" \
  -H "Cookie: smap_auth_token=<ADMIN_TOKEN>" \
  -d '{"user_id": "550e8400-e29b-41d4-a716-446655440000"}'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "message": "Token revoked successfully"
  }
}
```

**Use Cases**:
- ✅ Security incident response
- ✅ Account suspension
- ✅ Force re-authentication
- ✅ Logout all devices

---

### 5.3. Get User by ID

**Mô tả**: Internal services có thể lấy user info by ID.

**Endpoint**: `GET /internal/users/:id`

**Request**:
```bash
curl http://localhost:8080/internal/users/550e8400-e29b-41d4-a716-446655440000
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@vinfast.com",
    "name": "John Doe",
    "role": "ADMIN",
    "is_active": true
  }
}
```

---

### 5.4. Service Authentication

**Mô tả**: Internal endpoints được protect bằng service keys.

**Config**:
```yaml
internal:
  service_keys:
    project_service: project-service-key-123
    ingest_service: ingest-service-key-456
    knowledge_service: knowledge-service-key-789
```

**Request**:
```bash
curl http://localhost:8080/internal/users/123 \
  -H "X-Service-Key: project-service-key-123"
```

**Validation**:
```
1. Extract X-Service-Key header
2. Check key in service_keys config
3. If valid → Allow request
4. If invalid → 401 Unauthorized
```

---

## 🛡️ 6. SECURITY FEATURES

### 6.1. HttpOnly Cookies

**Mô tả**: JWT tokens được lưu trong HttpOnly cookies để chống XSS.

**Benefits**:
- ✅ JavaScript không thể access cookie
- ✅ Tự động gửi với mọi request
- ✅ Chống XSS attacks
- ✅ Secure flag cho HTTPS

---

### 6.2. CORS Protection

**Mô tả**: Middleware CORS để control cross-origin requests.

**Config**:
```go
AllowOrigins: []string{"http://localhost:3000", "https://smap.vinfast.com"}
AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
AllowHeaders: []string{"Origin", "Content-Type", "Authorization"}
AllowCredentials: true  // Allow cookies
```

---

### 6.3. Password Encryption

**Mô tả**: User roles được encrypt trước khi lưu database.

**Algorithm**: AES-256-GCM

**Config**:
```yaml
encrypter:
  key: test-encryption-key-32-characters
```

---

### 6.4. Panic Recovery

**Mô tả**: Middleware recovery để catch panics và prevent service crash.

**Features**:
- ✅ Catch all panics
- ✅ Log stack trace
- ✅ Return 500 error
- ✅ Service continues running

---

## 🌐 7. INTERNATIONALIZATION (i18n)

### 7.1. Multi-language Support

**Mô tả**: Service hỗ trợ multiple languages cho error messages.

**Supported Languages**:
- ✅ English (en)
- ✅ Vietnamese (vi)

**Locale Detection**:
```
1. Check Accept-Language header
2. Extract locale (en, vi)
3. Load translations
4. Return localized messages
```

**Example**:
```
EN: "Domain not allowed"
VI: "Tên miền không được phép"
```

---

## 📈 8. MONITORING & HEALTH CHECK

### 8.1. Health Check Endpoint

**Endpoint**: `GET /health`

**Response**:
```json
{
  "status": "ok",
  "timestamp": "2026-02-14T10:00:00Z"
}
```

**Use Cases**:
- ✅ Kubernetes liveness probe
- ✅ Load balancer health check
- ✅ Monitoring systems

---

### 8.2. Discord Notifications (Optional)

**Mô tả**: Service có thể gửi notifications to Discord webhook.

**Config**:
```yaml
discord:
  webhook_id: "123456789"
  webhook_token: "abcdef..."
```

**Use Cases**:
- ✅ Error notifications
- ✅ Security alerts
- ✅ System events

---

## 📚 9. API DOCUMENTATION

### 9.1. Swagger/OpenAPI

**Endpoint**: `GET /swagger/index.html`

**Features**:
- ✅ Interactive API documentation
- ✅ Try-it-out functionality
- ✅ Request/response examples
- ✅ Authentication testing

**Generate Docs**:
```bash
make swagger
```

---

## 🔄 10. LOGOUT FUNCTIONALITY

### 10.1. Single Device Logout

**Endpoint**: `POST /authentication/logout`

**Flow**:
```
1. Extract JTI from JWT
2. Add JTI to blacklist
3. Delete session from Redis
4. Expire cookie
5. Publish LOGOUT audit event
```

**Request**:
```bash
curl -X POST http://localhost:8080/authentication/logout \
  -H "Cookie: smap_auth_token=<TOKEN>"
```

**Response**:
```json
{
  "status": "success",
  "data": null
}
```

---

### 10.2. Logout All Devices

**Mô tả**: Admin có thể logout user khỏi tất cả devices.

**Flow**:
```
1. Get all JTIs for user from Redis
2. Add all JTIs to blacklist
3. Delete all sessions
4. All tokens immediately invalid
```

**Use Cases**:
- ✅ Security incident
- ✅ Password change
- ✅ Account suspension

---

## 📊 SUMMARY - Tổng Hợp Chức Năng

### ✅ Authentication Features (7)
1. OAuth2 Login (Google, Azure, Okta)
2. JWT Token Generation (HS256)
3. Session Management (Redis)
4. Token Blacklist (Instant Revocation)
5. Cookie-based Auth (HttpOnly)
6. Token Validation
7. Logout (Single & All Devices)

### ✅ Authorization Features (4)
1. Role-Based Access Control (ADMIN, ANALYST, VIEWER)
2. Domain Validation
3. Email Blocklist
4. Redirect URL Validation

### ✅ User Management Features (3)
1. Auto User Creation (First Login)
2. User Profile Management
3. Get User Info

### ✅ Audit & Compliance Features (3)
1. Event Tracking (LOGIN, LOGOUT, etc.)
2. Kafka-based Async Processing
3. 90-day Retention Policy

### ✅ Internal API Features (4)
1. Token Validation (Service-to-Service)
2. Token Revocation (Admin)
3. Get User by ID
4. Service Authentication

### ✅ Security Features (6)
1. HttpOnly Cookies (XSS Protection)
2. CORS Protection
3. Password/Role Encryption
4. Panic Recovery
5. Domain Validation
6. Blacklist System

### ✅ Infrastructure Features (4)
1. Health Check Endpoint
2. Swagger Documentation
3. i18n Support (EN, VI)
4. Discord Notifications

---

## 🎯 Use Cases Thực Tế

### Use Case 1: User Login Flow
```
1. User mở browser → click "Login"
2. Redirect to Google OAuth
3. User approve permissions
4. System validate domain (vinfast.com)
5. System map email → role (ADMIN)
6. System generate JWT token
7. System set HttpOnly cookie
8. User được redirect về dashboard
9. Audit event: LOGIN được log
```

### Use Case 2: API Request Authentication
```
1. Frontend gửi request với cookie
2. Middleware extract JWT from cookie
3. Middleware verify JWT signature
4. Middleware check blacklist
5. Middleware extract user info
6. Request proceeds với user context
```

### Use Case 3: Admin Revoke User Access
```
1. Admin detect security incident
2. Admin call POST /internal/revoke-token
3. System add all user tokens to blacklist
4. System delete all user sessions
5. User immediately logged out from all devices
6. Audit event: TOKEN_REVOKED được log
```

### Use Case 4: Service-to-Service Auth
```
1. Project Service cần validate user token
2. Project Service call POST /internal/validate
3. Identity Service verify token
4. Identity Service return user info
5. Project Service proceed với request
```

---

**Version**: 2.0.0 (Simplified)  
**Last Updated**: 14/02/2026  
**Total Features**: 31 features across 10 categories
