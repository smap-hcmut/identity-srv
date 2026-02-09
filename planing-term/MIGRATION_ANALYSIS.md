# 📊 Phân Tích Migration: Identity Service → Auth Service

**Ngày:** 09/02/2026 | **Phiên bản:** v1.0

---

## 🎯 TÓM TẮT EXECUTIVE

### ✅ KHUYẾN NGHỊ: MIGRATE NGAY (Greenfield Project)

**Context quan trọng:**

- ✅ **Chưa có khách hàng** → Không cần lo migration users
- ✅ **Chưa có production data** → Clean slate
- ✅ **Có AI agents** → Code generation nhanh
- ✅ **Thời gian: 1 tuần** → Đủ cho greenfield

**Lý do chính:**

1. Phù hợp Business Model mới (On-Premise B2B vs SaaS)
2. Đơn giản hóa 70% code không cần thiết
3. Enterprise-ready: SSO, RBAC, Audit Log
4. Kiến trúc tốt hơn: JWT self-validation, stateless
5. **KHÔNG CÓ RỦI RO MIGRATION** vì chưa có users

**Chiến lược:**

- Thời gian: **1 tuần** (không phải 2-3 tuần)
- Phạm vi: Chỉ Auth Service, giữ nguyên services khác
- Approach: **Clean implementation** (không phải migration)

---

## 📊 SO SÁNH TỔNG QUAN

### Identity Service (Hiện tại)

```
🎯 Mục đích: SaaS Multi-tenant Authentication
📦 Features:
   ✓ User Registration (Email + Password)
   ✓ OTP Email Verification
   ✓ JWT Login (HttpOnly Cookie)
   ✓ Password Management
   ✓ Subscription Plans (Free trial 14 days)
   ✓ Role-based Access (USER, ADMIN)

🏗️ Tech Stack:
   - Go 1.25 + Gin
   - PostgreSQL (users, plans, subscriptions)
   - RabbitMQ (Email queue)
   - SMTP (Email sending)
   - bcrypt (Password hashing)
   - JWT (golang-jwt)

📁 Database: 3 tables
   - users (email, password_hash, role_hash, otp)
   - plans (name, code, max_usage)
   - subscriptions (user_id, plan_id, status)
```

### Auth Service (Planning)

```
🎯 Mục đích: On-Premise Enterprise SSO
📦 Features:
   ✓ Google OAuth2/OIDC (SSO)
   ✓ Azure AD / Okta support (pluggable)
   ✓ Domain-based Access Control
   ✓ Google Groups → Role Mapping
   ✓ JWT RS256 (Asymmetric)
   ✓ Public Key Distribution (JWKS)
   ✓ Audit Log (90 days retention)
   ✓ Token Blacklist (Redis)
   ✓ Key Rotation Strategy

   ✗ KHÔNG CÓ: Registration, Password, OTP, Subscriptions

🏗️ Tech Stack:
   - Go 1.25 + Chi Router
   - PostgreSQL (users, audit_logs, jwt_keys)
   - Redis (Session, Blacklist, Groups cache)
   - Kafka (Audit events)
   - Google Directory API
   - JWT RS256 (RSA keypair)

📁 Database: 3 tables
   - users (email, name, role, last_login)
   - audit_logs (user_id, action, resource_type)
   - jwt_keys (kid, public_key, private_key, status)
```

---

## 🔄 THAY ĐỔI CHI TIẾT

### 1. Authentication Flow

| Aspect               | Identity (Cũ)       | Auth Service (Mới)             |
| -------------------- | ------------------- | ------------------------------ |
| **Login Method**     | Email + Password    | Google OAuth2 SSO              |
| **User Creation**    | Manual Registration | Auto-create on first SSO login |
| **Verification**     | OTP via Email       | Google verifies                |
| **Password**         | bcrypt hash         | Không có (Google manages)      |
| **JWT Algorithm**    | HS256 (Symmetric)   | RS256 (Asymmetric)             |
| **Token Storage**    | HttpOnly Cookie     | HttpOnly Cookie (giống)        |
| **Token Validation** | Shared secret       | Public key (self-validation)   |
| **Session**          | Stateless JWT       | Redis-backed + JWT             |

**🔑 Thay đổi lớn nhất:**

- **Cũ:** Self-managed users (email/password)
- **Mới:** Delegate to Google Workspace (SSO)

### 2. Authorization & RBAC

| Aspect               | Identity (Cũ)        | Auth Service (Mới)            |
| -------------------- | -------------------- | ----------------------------- |
| **Roles**            | USER, ADMIN          | ADMIN, ANALYST, VIEWER        |
| **Role Source**      | Database field       | Google Groups mapping         |
| **Permission Check** | Middleware reads JWT | Middleware reads JWT + Groups |
| **Role Update**      | Manual DB update     | Auto-sync from Google Groups  |

**🔑 Thay đổi lớn nhất:**

- **Cũ:** Static roles trong DB
- **Mới:** Dynamic roles từ Google Groups (sync mỗi 5 phút)

### 3. Database Schema Changes

**BỎ HOÀN TOÀN:**

```sql
-- ❌ Không cần nữa
DROP TABLE subscriptions;
DROP TABLE plans;

-- ❌ Bỏ các cột
ALTER TABLE users DROP COLUMN password_hash;
ALTER TABLE users DROP COLUMN otp;
ALTER TABLE users DROP COLUMN otp_expired_at;
ALTER TABLE users DROP COLUMN is_active;
```

**THÊM MỚI:**

```sql
-- ✅ Audit log table
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    action VARCHAR(50),
    resource_type VARCHAR(50),
    resource_id UUID,
    metadata JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ DEFAULT (NOW() + INTERVAL '90 days')
);

-- ✅ JWT keys table (for key rotation)
CREATE TABLE jwt_keys (
    kid VARCHAR(50) PRIMARY KEY,
    private_key TEXT NOT NULL,
    public_key TEXT NOT NULL,
    status VARCHAR(20), -- active | rotating | retired
    created_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ
);
```

### 4. API Endpoints Changes

**BỎ HOÀN TOÀN:**

```
❌ POST /auth/register
❌ POST /auth/send-otp
❌ POST /auth/verify-otp
❌ POST /auth/change-password
❌ GET  /plans
❌ POST /plans
❌ GET  /subscriptions
❌ POST /subscriptions
```

**THÊM MỚI:**

```
✅ GET  /auth/login              → Redirect to Google OAuth
✅ GET  /auth/callback           → OAuth callback handler
✅ GET  /.well-known/jwks.json   → Public keys (JWKS)
✅ POST /internal/revoke-token   → Blacklist token (Admin)
✅ GET  /audit-logs              → Query audit logs
```

**GIỮ NGUYÊN:**

```
✓ POST /auth/logout
✓ GET  /auth/me
✓ GET  /health
```

---

## 📈 ĐÁNH GIÁ TÁC ĐỘNG

### Tích cực ✅

1. **Đơn giản hóa Code**
   - Bỏ 70% code: Registration, OTP, Password, Subscriptions
   - Giảm từ ~3000 LOC xuống ~1200 LOC
   - Ít bug hơn, dễ maintain hơn

2. **Bảo mật tốt hơn**
   - Không tự quản lý password → Giảm attack surface
   - JWT RS256 → Self-validation, không cần shared secret
   - Token blacklist → Revoke tức thì
   - Audit log → Compliance (ISO 27001, SOC 2)

3. **Enterprise-ready**
   - SSO integration (Google, Azure AD, Okta)
   - Domain-based access control
   - Google Groups → Role mapping
   - Audit trail cho compliance

4. **Scalability**
   - JWT self-validation → Không cần gọi Auth Service
   - Stateless → Dễ scale horizontal
   - Redis cache → Giảm load Google API

### Tiêu cực ❌

1. **Phụ thuộc Google Workspace**
   - Khách hàng PHẢI có Google Workspace
   - ~~Nếu Google down → Không login được~~ **KHÔNG LO** vì:
     - Cache Google Groups trong Redis (TTL 5 phút)
     - JWT vẫn valid trong 15 phút
     - Services tự verify JWT bằng public key (không cần Auth Service)
   - **Giải pháp:** Support Azure AD, Okta (pluggable)

2. **Mất tính linh hoạt**
   - Không thể tạo user thủ công
   - Không thể test local dễ dàng
   - **Giải pháp:** Dev mode với mock OAuth

3. ~~**Migration effort**~~ **KHÔNG CÓN** vì greenfield
   - ~~Phải migrate existing users~~ → Không có users
   - ~~Phải update frontend (OAuth flow)~~ → Có document chi tiết
   - ~~Phải update các services khác (JWT middleware)~~ → AI agents code

4. **Không phù hợp cho SaaS public**
   - Nếu muốn bán cho SME không có Google Workspace
   - **Giải pháp:** Giữ Identity Service cho SaaS, dùng Auth cho Enterprise

---

## ⚠️ RỦI RO & MITIGATION (Updated - Greenfield Context)

| Rủi ro                             | Mức độ        | Mitigation                      | Status       |
| ---------------------------------- | ------------- | ------------------------------- | ------------ |
| ~~**Google API downtime**~~        | ~~🟡 Medium~~ | Cache + JWT self-validation     | ✅ Không lo  |
| ~~**Existing users migration**~~   | ~~🔴 High~~   | N/A - Chưa có users             | ✅ Không có  |
| **Frontend breaking changes**      | 🟢 Low        | Document chi tiết OAuth flow    | ✅ Có doc    |
| **Services không verify JWT đúng** | 🟡 Medium     | AI agents code + Testing        | ✅ Có agents |
| **Key rotation phức tạp**          | 🟢 Low        | Phase 1: Manual, Phase 2: Auto  | ✅ OK        |
| **Audit log performance**          | 🟢 Low        | Async Kafka queue, batch insert | ✅ OK        |
| **Learning curve OAuth2**          | 🟢 Low        | Planning docs có flow diagrams  | ✅ Có docs   |

**Kết luận:** Hầu hết rủi ro đã được loại bỏ do greenfield project!

---

## 🗓️ KẾ HOẠCH IMPLEMENTATION (Updated - 1 Tuần)

### 🚀 Greenfield Implementation Plan

**Context:**

- ✅ Không có users cũ → Không cần migration
- ✅ Không có production → Clean implementation
- ✅ Có AI agents → Code generation nhanh
- ✅ Có planning docs → Spec rõ ràng

### Day 1-2: Core Auth Service (2 ngày)

**Mục tiêu:** Implement OAuth2 + JWT core

- [ ] Setup project structure (Go + Chi router)
- [ ] Implement Google OAuth2 flow
  - [ ] `/auth/login` - Redirect to Google
  - [ ] `/auth/callback` - Handle OAuth callback
  - [ ] Domain validation (allowed_domains)
- [ ] Implement JWT RS256
  - [ ] Generate RSA keypair
  - [ ] Sign tokens with private key
  - [ ] JWKS endpoint `/.well-known/jwks.json`
- [ ] Database schema
  - [ ] `users` table (simplified)
  - [ ] `audit_logs` table
  - [ ] `jwt_keys` table
- [ ] Basic endpoints
  - [ ] `GET /auth/me`
  - [ ] `POST /auth/logout`
  - [ ] `GET /health`

**Deliverables:**

- Auth Service running locally
- OAuth flow working
- JWT tokens issued

**AI Agent Tasks:**

- Generate boilerplate code
- Generate database migrations
- Generate test cases

### Day 3: Google Groups + Audit Log (1 ngày)

**Mục tiêu:** RBAC + Compliance features

- [ ] Google Directory API integration
  - [ ] Service account setup
  - [ ] Fetch user groups
  - [ ] Cache groups in Redis (TTL 5 min)
  - [ ] Map groups → roles (config file)
- [ ] Audit log system
  - [ ] Kafka publisher (shared package)
  - [ ] Kafka consumer (batch insert)
  - [ ] Audit log endpoints
- [ ] Redis setup
  - [ ] Session storage
  - [ ] Groups cache
  - [ ] Token blacklist (future)

**Deliverables:**

- RBAC working (ADMIN, ANALYST, VIEWER)
- Audit log recording actions
- Redis caching working

**AI Agent Tasks:**

- Generate Kafka publisher/consumer code
- Generate Redis client wrapper
- Generate audit log queries

### Day 4: JWT Middleware Package (1 ngày)

**Mục tiêu:** Shared library cho các services khác

- [ ] Create `pkg/auth` package
  - [ ] JWT verification middleware
  - [ ] Public key fetching (JWKS)
  - [ ] Public key caching
  - [ ] Role-based authorization helpers
- [ ] Documentation
  - [ ] Usage examples
  - [ ] Integration guide
  - [ ] API reference
- [ ] Testing
  - [ ] Unit tests
  - [ ] Integration tests
  - [ ] Mock OAuth for testing

**Deliverables:**

- `pkg/auth` package ready
- Documentation complete
- Test coverage > 80%

**AI Agent Tasks:**

- Generate middleware code
- Generate documentation
- Generate test cases

### Day 5: Services Integration (1 ngày)

**Mục tiêu:** Integrate Auth với các services khác

- [ ] Update Project Service
  - [ ] Add `pkg/auth` middleware
  - [ ] Update routes (require auth)
  - [ ] Add role checks (ANALYST, ADMIN)
- [ ] Update Ingest Service
  - [ ] Add `pkg/auth` middleware
  - [ ] Update routes
- [ ] Update Knowledge Service
  - [ ] Add `pkg/auth` middleware
  - [ ] Update routes
- [ ] Update Notification Service
  - [ ] Add `pkg/auth` middleware
  - [ ] Update WebSocket auth

**Deliverables:**

- All services integrated
- Auth working end-to-end

**AI Agent Tasks:**

- Generate integration code for each service
- Update route definitions
- Generate integration tests

### Day 6-7: Frontend + Documentation (2 ngày)

**Day 6: Frontend OAuth**

- [ ] OAuth login flow
  - [ ] Login button → `/auth/login`
  - [ ] Handle callback
  - [ ] Axios config (`withCredentials: true`)
- [ ] Update API calls
  - [ ] Remove localStorage code
  - [ ] Test authenticated requests
- [ ] Error handling (401, 403)

**Day 7: Documentation**

- [ ] **Auth Service API Docs**
  - [ ] Endpoints reference
  - [ ] OAuth flow diagram
  - [ ] JWT structure
- [ ] **Integration Guide**
  - [ ] JWT middleware usage
  - [ ] Role-based authorization
- [ ] **Deployment Guide**
  - [ ] Google OAuth setup
  - [ ] Environment variables
  - [ ] K8s manifests
- [ ] **Frontend Migration Guide**
  - [ ] OAuth implementation
  - [ ] Axios configuration

**Deliverables:**

- Frontend OAuth working
- Complete documentation
- Deployment ready

**AI Agent Tasks:**

- Generate frontend OAuth code
- Generate documentation
- Generate deployment manifests

---

## 💰 EFFORT ESTIMATION (Updated - 1 Tuần)

### Development Effort (Greenfield)

| Day         | Task                   | Hours   | AI Agent Help                     |
| ----------- | ---------------------- | ------- | --------------------------------- |
| **Day 1-2** | Auth Service Core      | 16h     | ✅ Boilerplate, migrations, tests |
| **Day 3**   | Google Groups + Audit  | 8h      | ✅ Kafka code, Redis wrapper      |
| **Day 4**   | JWT Middleware Package | 8h      | ✅ Middleware, docs, tests        |
| **Day 5**   | Services Integration   | 8h      | ✅ Integration code               |
| **Day 6**   | Frontend OAuth         | 8h      | ✅ OAuth flow code                |
| **Day 7**   | Documentation          | 8h      | ✅ Docs generation                |
| **TOTAL**   | **7 days**             | **56h** | **~70% AI-generated**             |

**Lý do nhanh hơn:**

- ✅ Không có migration users
- ✅ Không có backward compatibility
- ✅ AI agents generate 70% code
- ✅ Planning docs rõ ràng
- ✅ Clean implementation

---

## 📚 DOCUMENTS CẦN TẠO

### 1. Auth Service API Documentation

**File:** `docs/auth-service-api.md`

**Nội dung:**

````markdown
# Auth Service API Documentation

## Authentication Flow

### 1. Login (OAuth2)

GET /auth/login
→ Redirects to Google OAuth

### 2. Callback

GET /auth/callback?code=xxx
→ Exchanges code for token
→ Sets HttpOnly cookie
→ Redirects to dashboard

### 3. Get Current User

GET /auth/me
Authorization: Cookie (automatic)
→ Returns user info

### 4. Logout

POST /auth/logout
→ Expires cookie

## JWT Structure

```json
{
  "iss": "smap-auth-service",
  "aud": ["smap-api"],
  "sub": "user-uuid",
  "email": "user@vinfast.com",
  "role": "ANALYST",
  "groups": ["marketing-team@vinfast.com"],
  "jti": "token-uuid",
  "exp": 1234567890
}
```
````

## JWKS Endpoint

GET /.well-known/jwks.json
→ Returns public keys for JWT verification

## Error Codes

| Code               | HTTP | Description                   |
| ------------------ | ---- | ----------------------------- |
| DOMAIN_NOT_ALLOWED | 403  | Email domain not in whitelist |
| ACCOUNT_BLOCKED    | 403  | User blocked by admin         |
| INVALID_TOKEN      | 401  | JWT invalid or expired        |

````

### 2. JWT Middleware Integration Guide

**File:** `docs/jwt-middleware-guide.md`

**Nội dung:**
```markdown
# JWT Middleware Integration Guide

## Installation

```go
import "smap-api/pkg/auth"
````

## Usage

### Basic Authentication

```go
func main() {
    // Initialize middleware
    authMW, _ := auth.NewJWTMiddleware(
        "http://auth-service:8080",
        redisClient,
    )

    // Apply to routes
    r := chi.NewRouter()
    r.Use(authMW.Authenticate)

    r.Get("/projects", listProjects)
}
```

### Role-Based Authorization

```go
// Require ANALYST role
r.With(auth.RequireRole("ANALYST")).Post("/projects", createProject)

// Require ADMIN role
r.With(auth.RequireRole("ADMIN")).Delete("/projects/{id}", deleteProject)
```

### Extract User Info

```go
func handler(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("user_id").(string)
    email := r.Context().Value("email").(string)
    role := r.Context().Value("role").(string)
}
```

## Testing

### Mock OAuth for Tests

```go
func TestWithMockAuth(t *testing.T) {
    authMW := auth.NewMockMiddleware(auth.MockUser{
        ID: "test-user",
        Email: "test@vinfast.com",
        Role: "ANALYST",
    })
}
```

````

### 3. Frontend OAuth Migration Guide

**File:** `docs/frontend-oauth-guide.md`

**Nội dung:**
```markdown
# Frontend OAuth Migration Guide

## Before (Identity Service)

```javascript
// Login with email/password
const response = await api.post('/auth/login', {
  email: 'user@example.com',
  password: 'password123'
});

// Token in response body
const token = response.data.token;
localStorage.setItem('token', token);

// Manual Authorization header
axios.defaults.headers.common['Authorization'] = `Bearer ${token}`;
````

## After (Auth Service)

```javascript
// Login with Google OAuth
window.location.href = "https://api.smap.com/auth/login";

// Callback handled automatically
// Cookie set by server (HttpOnly)

// Configure axios
const api = axios.create({
  baseURL: "https://api.smap.com",
  withCredentials: true, // ← REQUIRED for cookies
});

// No manual token management needed!
// Cookie sent automatically with every request
```

## Error Handling

```javascript
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Redirect to login
      window.location.href = "/auth/login";
    }
    if (error.response?.status === 403) {
      // Show permission error
      toast.error("You don't have permission");
    }
    return Promise.reject(error);
  },
);
```

## Testing

```javascript
// Mock OAuth in tests
jest.mock("axios");
axios.get.mockResolvedValue({
  data: { user: { id: "123", email: "test@vinfast.com" } },
});
```

````

### 4. Deployment Guide

**File:** `docs/deployment-guide.md`

**Nội dung:**
```markdown
# Auth Service Deployment Guide

## Prerequisites

1. **Google Workspace Account**
   - Admin access
   - OAuth2 credentials

2. **Infrastructure**
   - Kubernetes cluster
   - PostgreSQL database
   - Redis cluster
   - Kafka cluster

## Step 1: Google OAuth Setup

1. Go to Google Cloud Console
2. Create OAuth2 credentials
3. Set redirect URI: `https://api.smap.com/auth/callback`
4. Download credentials JSON
5. Create Kubernetes secret:

```bash
kubectl create secret generic google-oauth \
  --from-file=credentials.json
````

## Step 2: Generate JWT Keys

```bash
# Generate RSA keypair
openssl genrsa -out jwt-private.pem 2048
openssl rsa -in jwt-private.pem -pubout -out jwt-public.pem

# Create Kubernetes secret
kubectl create secret generic jwt-keys \
  --from-file=private.pem=jwt-private.pem \
  --from-file=public.pem=jwt-public.pem
```

## Step 3: Deploy Auth Service

```bash
# Apply manifests
kubectl apply -f manifests/auth-service/

# Check status
kubectl get pods -l app=auth-service
kubectl logs -f deployment/auth-service
```

## Step 4: Configure Other Services

Update each service's deployment:

```yaml
env:
  - name: AUTH_SERVICE_URL
    value: "http://auth-service:8080"
  - name: REDIS_URL
    value: "redis://redis:6379"
```

## Step 5: Verify

```bash
# Test OAuth flow
curl https://api.smap.com/auth/login

# Test JWKS endpoint
curl https://api.smap.com/.well-known/jwks.json

# Test authenticated endpoint
curl -b cookies.txt https://api.smap.com/auth/me
```

```

---

## 🎯 KHUYẾN NGHỊ CUỐI CÙNG (Updated)

### ✅ MIGRATE NGAY - Greenfield Advantage

**Lý do quyết định:**

1. **Không có rủi ro migration**
   - Chưa có users → Không cần migrate
   - Chưa có production → Clean slate
   - Chưa có technical debt

2. **Thời gian hợp lý: 1 tuần**
   - Day 1-2: Core Auth (OAuth + JWT)
   - Day 3: RBAC + Audit
   - Day 4: JWT Middleware
   - Day 5: Services Integration
   - Day 6: Frontend
   - Day 7: Documentation

3. **AI Agents giúp 70% công việc**
   - Boilerplate code generation
   - Test case generation
   - Documentation generation
   - Integration code

4. **Planning documents rất chi tiết**
   - Flow diagrams có sẵn
   - Database schema có sẵn
   - API specs có sẵn
   - Security best practices

5. **Enterprise-ready từ đầu**
   - SSO (Google, Azure AD, Okta)
   - RBAC (ADMIN, ANALYST, VIEWER)
   - Audit Log (compliance)
   - JWT self-validation (scalable)

### ❌ KHÔNG giữ Identity Service

**Lý do:**
- Không có users cũ cần support
- Không có backward compatibility concerns
- Code cũ sẽ thành technical debt
- Maintain 2 auth systems tốn effort

### 📋 Action Items

**Ngay lập tức:**
1. [ ] Setup Google Workspace test account
2. [ ] Create OAuth2 credentials
3. [ ] Setup development environment (Redis, Kafka)
4. [ ] Clone planning documents vào project

**Tuần này:**
1. [ ] Implement Auth Service (Day 1-3)
2. [ ] Create JWT Middleware package (Day 4)
3. [ ] Integrate services (Day 5)
4. [ ] Update frontend (Day 6)
5. [ ] Write documentation (Day 7)

**Tuần sau:**
1. [ ] Deploy to staging
2. [ ] E2E testing
3. [ ] Performance testing
4. [ ] Security audit

---

## 📞 SUPPORT

Nếu cần hỗ trợ implementation:
- **Planning Docs:** `planing-term/migration-plan-v2.md`
- **Flow Diagrams:** `planing-term/auth-flow-diagram.md`
- **Security Guide:** `planing-term/auth-security-enhancements.md`
- **This Analysis:** `MIGRATION_ANALYSIS.md`

---

**✨ Kết luận: MIGRATE NGAY - Đây là thời điểm tốt nhất!**

*Last updated: 09/02/2026 - Greenfield Context*
```
