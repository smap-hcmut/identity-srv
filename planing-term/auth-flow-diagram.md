# Auth Service Flow Diagram

## 1. JWT Verification Flow (Middleware ở các Services)

```
┌─────────────────────────────────────────────────────────────────┐
│              JWT VERIFICATION MIDDLEWARE FLOW                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  User Request                                                   │
│       ↓                                                         │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  Web UI / Mobile App                                    │    │
│  │  Header: Authorization: Bearer eyJhbGc...              │    │
│  └─────────────────────────┬───────────────────────────────┘    │
│                            ↓                                    │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  Project Service (hoặc bất kỳ service nào)             │    │
│  │                                                         │    │
│  │  ┌───────────────────────────────────────────────┐      │    │
│  │  │  JWT Middleware (pkg/auth/middleware.go)     │      │    │
│  │  │                                              │      │    │
│  │  │  1. Extract token from Authorization header │      │    │
│  │  │     tokenString = "eyJhbGc..."              │      │    │
│  │  │                                              │      │    │
│  │  │  2. Parse JWT using PUBLIC KEY (cached)     │      │    │
│  │  │     token, err := jwt.Parse(tokenString,    │      │    │
│  │  │       func(token) { return publicKey })     │      │    │
│  │  │                                              │      │    │
│  │  │  3. Verify signature (RS256)                │      │    │
│  │  │     ✓ Signature valid                       │      │    │
│  │  │                                              │      │    │
│  │  │  4. Verify issuer & audience                │      │    │
│  │  │     ✓ iss = "smap-auth-service"             │      │    │
│  │  │     ✓ aud = ["smap-api"]                    │      │    │
│  │  │                                              │      │    │
│  │  │  5. Extract claims                          │      │    │
│  │  │     user_id = "uuid-123"                    │      │    │
│  │  │     email = "manager@vinfast.com"           │      │    │
│  │  │     role = "ANALYST"                        │      │    │
│  │  │     groups = ["marketing-team@..."]         │      │    │
│  │  │                                              │      │    │
│  │  │  6. Inject into request context             │      │    │
│  │  │     ctx = context.WithValue(ctx, "user_id", │      │    │
│  │  │                              user_id)       │      │    │
│  │  └───────────────────────────────────────────────┘      │    │
│  │                            ↓                            │    │
│  │  ┌───────────────────────────────────────────────┐      │    │
│  │  │  Authorization Middleware (Optional)         │      │    │
│  │  │  RequireRole("ANALYST")                      │      │    │
│  │  │                                              │      │    │
│  │  │  userRole = ctx.Value("role") // "ANALYST"  │      │    │
│  │  │  if !hasPermission(userRole, "ANALYST") {   │      │    │
│  │  │    return 403 Forbidden                     │      │    │
│  │  │  }                                           │      │    │
│  │  └───────────────────────────────────────────────┘      │    │
│  │                            ↓                            │    │
│  │  ┌───────────────────────────────────────────────┐      │    │
│  │  │  Business Logic Handler                      │      │    │
│  │  │  CreateProject(w, r)                         │      │    │
│  │  │                                              │      │    │
│  │  │  userID := r.Context().Value("user_id")     │      │    │
│  │  │  // Execute business logic...               │      │    │
│  │  └───────────────────────────────────────────────┘      │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Key Points:**

- Middleware chạy TRƯỚC business logic
- Verify JWT bằng public key (KHÔNG cần gọi Auth Service)
- Extract user info và inject vào context
- Business logic handler lấy user info từ context

---

## 2. Audit Log Flow (Async via Kafka)

```
┌─────────────────────────────────────────────────────────────────┐
│                    AUDIT LOG ASYNC FLOW                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  Project Service - CreateProject Handler               │    │
│  │                                                         │    │
│  │  func CreateProject(w, r) {                            │    │
│  │    // 1. Extract user from context                     │    │
│  │    userID := r.Context().Value("user_id")              │    │
│  │                                                         │    │
│  │    // 2. Execute business logic                        │    │
│  │    project := service.CreateProject(...)               │    │
│  │    db.Insert(project) // ← Blocking                    │    │
│  │                                                         │    │
│  │    // 3. Publish audit event (NON-BLOCKING)            │    │
│  │    auditPublisher.Log(AuditEvent{                      │    │
│  │      UserID: userID,                                   │    │
│  │      Action: "CREATE_PROJECT",                         │    │
│  │      ResourceID: project.ID,                           │    │
│  │      ...                                               │    │
│  │    }) // ← Fire-and-forget, returns immediately        │    │
│  │                                                         │    │
│  │    // 4. Return response (không đợi audit log)         │    │
│  │    json.Encode(w, project)                             │    │
│  │  }                                                      │    │
│  └─────────────────────────┬───────────────────────────────┘    │
│                            ↓ (async)                            │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  Kafka Topic: audit.events                              │    │
│  │  Message: {                                             │    │
│  │    user_id: "uuid-123",                                 │    │
│  │    action: "CREATE_PROJECT",                            │    │
│  │    resource_type: "project",                            │    │
│  │    resource_id: "proj-456",                             │    │
│  │    metadata: {...},                                     │    │
│  │    ip_address: "10.0.1.5",                              │    │
│  │    user_agent: "Mozilla/5.0..."                         │    │
│  │  }                                                      │    │
│  └─────────────────────────┬───────────────────────────────┘    │
│                            ↓                                    │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  Auth Service - Audit Consumer                         │    │
│  │                                                         │    │
│  │  batch := []AuditEvent{}                               │    │
│  │  ticker := time.NewTicker(5 * time.Second)            │    │
│  │                                                         │    │
│  │  for {                                                  │    │
│  │    select {                                             │    │
│  │    case <-ticker.C:                                     │    │
│  │      // Flush batch every 5 seconds                    │    │
│  │      if len(batch) > 0 {                               │    │
│  │        batchInsertAuditLogs(batch)                     │    │
│  │        batch = batch[:0]                               │    │
│  │      }                                                  │    │
│  │                                                         │    │
│  │    default:                                             │    │
│  │      // Read message from Kafka                        │    │
│  │      msg := reader.ReadMessage()                       │    │
│  │      batch = append(batch, msg)                        │    │
│  │                                                         │    │
│  │      // Flush if batch full (100 messages)             │    │
│  │      if len(batch) >= 100 {                            │    │
│  │        batchInsertAuditLogs(batch)                     │    │
│  │        batch = batch[:0]                               │    │
│  │      }                                                  │    │
│  │    }                                                    │    │
│  │  }                                                      │    │
│  └─────────────────────────┬───────────────────────────────┘    │
│                            ↓                                    │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  PostgreSQL: auth.audit_logs                            │    │
│  │                                                         │    │
│  │  INSERT INTO auth.audit_logs                           │    │
│  │  (user_id, action, resource_type, resource_id, ...)    │    │
│  │  VALUES                                                 │    │
│  │    ('uuid-123', 'CREATE_PROJECT', 'project', ...),     │    │
│  │    ('uuid-456', 'DELETE_SOURCE', 'data_source', ...),  │    │
│  │    ... (batch insert 100 rows)                         │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Key Points:**

- Business logic KHÔNG bị block bởi audit log
- Kafka buffer messages nếu Auth Service down
- Batch insert giảm DB load (100 messages hoặc 5 giây)
- Audit Consumer có thể scale độc lập

---

## 3. Public Key Distribution Flow

```
┌─────────────────────────────────────────────────────────────────┐
│              PUBLIC KEY DISTRIBUTION FLOW                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  Auth Service Startup                                   │    │
│  │                                                         │    │
│  │  1. Load or generate RSA key pair                      │    │
│  │     privateKey := loadPrivateKey("/secrets/jwt-private.pem") │
│  │     publicKey := loadPublicKey("/secrets/jwt-public.pem")    │
│  │                                                         │    │
│  │  2. Expose JWKS endpoint                               │    │
│  │     GET /.well-known/jwks.json                         │    │
│  │     Response: {                                        │    │
│  │       "keys": [{                                       │    │
│  │         "kty": "RSA",                                  │    │
│  │         "use": "sig",                                  │    │
│  │         "kid": "smap-2026",                            │    │
│  │         "n": "base64-encoded-modulus",                 │    │
│  │         "e": "AQAB"                                    │    │
│  │       }]                                               │    │
│  │     }                                                  │    │
│  └─────────────────────────────────────────────────────────┘    │
│                            ↑                                    │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  Project Service Startup                                │    │
│  │                                                         │    │
│  │  1. Fetch public key from Auth Service                 │    │
│  │     resp := http.Get("http://auth-service:8080/.well-known/jwks.json") │
│  │                                                         │    │
│  │  2. Parse and cache public key                         │    │
│  │     publicKey := parseJWKS(resp.Body)                  │    │
│  │     cache.Set("jwt_public_key", publicKey)             │    │
│  │                                                         │    │
│  │  3. Refresh public key every 1 hour (background)       │    │
│  │     go func() {                                        │    │
│  │       ticker := time.NewTicker(1 * time.Hour)         │    │
│  │       for range ticker.C {                            │    │
│  │         refreshPublicKey()                            │    │
│  │       }                                               │    │
│  │     }()                                               │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  Ingest Service Startup (same flow)                     │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  Knowledge Service Startup (same flow)                  │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Key Points:**

- Auth Service publish public key qua JWKS endpoint (standard)
- Các services khác fetch public key on startup
- Cache public key và refresh mỗi giờ
- Nếu Auth Service down, services vẫn verify được JWT (dùng cached key)

---

## 4. Complete Authentication Flow (End-to-End)

```
┌─────────────────────────────────────────────────────────────────┐
│           COMPLETE AUTHENTICATION FLOW (E2E)                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  User                Web UI         Auth Service      Google    │
│   │                   │                  │              │       │
│   │ 1. Click Login    │                  │              │       │
│   ├──────────────────>│                  │              │       │
│   │                   │                  │              │       │
│   │                   │ 2. GET /auth/login              │       │
│   │                   ├─────────────────>│              │       │
│   │                   │                  │              │       │
│   │                   │ 3. Redirect to Google OAuth     │       │
│   │                   │<─────────────────┤              │       │
│   │                   │                  │              │       │
│   │ 4. Redirect       │                  │              │       │
│   │<──────────────────┤                  │              │       │
│   │                   │                  │              │       │
│   │ 5. Google Login Page                 │              │       │
│   ├──────────────────────────────────────┼─────────────>│       │
│   │                   │                  │              │       │
│   │ 6. Enter credentials & consent       │              │       │
│   ├──────────────────────────────────────┼─────────────>│       │
│   │                   │                  │              │       │
│   │ 7. Redirect to callback with code    │              │       │
│   │<──────────────────────────────────────┼──────────────┤       │
│   │                   │                  │              │       │
│   │ 8. GET /auth/callback?code=xxx       │              │       │
│   ├──────────────────────────────────────>│              │       │
│   │                   │                  │              │       │
│   │                   │ 9. Exchange code for token      │       │
│   │                   │                  ├─────────────>│       │
│   │                   │                  │              │       │
│   │                   │ 10. Return user info            │       │
│   │                   │                  │<─────────────┤       │
│   │                   │                  │              │       │
│   │                   │ 11. Verify domain (vinfast.com) │       │
│   │                   │                  │ ✓ OK         │       │
│   │                   │                  │              │       │
│   │                   │ 12. Fetch Google Groups         │       │
│   │                   │                  ├─────────────>│       │
│   │                   │                  │<─────────────┤       │
│   │                   │                  │              │       │
│   │                   │ 13. Map role (ANALYST)          │       │
│   │                   │                  │              │       │
│   │                   │ 14. Create/update user in DB    │       │
│   │                   │                  │              │       │
│   │                   │ 15. Generate JWT (RS256)        │       │
│   │                   │                  │ privateKey   │       │
│   │                   │                  │              │       │
│   │ 16. Set cookie & redirect to dashboard              │       │
│   │<──────────────────────────────────────┤              │       │
│   │   Set-Cookie: token=eyJhbGc...       │              │       │
│   │                   │                  │              │       │
│   │ 17. GET /dashboard                   │              │       │
│   ├──────────────────>│                  │              │       │
│   │   Cookie: token=eyJhbGc...           │              │       │
│   │                   │                  │              │       │
│   │                   │ 18. GET /api/projects           │       │
│   │                   │    Authorization: Bearer eyJ... │       │
│   │                   ├─────────────────────────────────────>   │
│   │                   │                  │         Project      │
│   │                   │                  │         Service      │
│   │                   │                  │              │       │
│   │                   │ 19. JWT Middleware verify       │       │
│   │                   │    (using cached public key)    │       │
│   │                   │    ✓ Valid                      │       │
│   │                   │                  │              │       │
│   │                   │ 20. Return projects             │       │
│   │                   │<────────────────────────────────────┤   │
│   │                   │                  │              │       │
│   │ 21. Render dashboard                 │              │       │
│   │<──────────────────┤                  │              │       │
│   │                   │                  │              │       │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Key Points:**

- User chỉ login 1 lần qua Google OAuth
- JWT token được lưu trong cookie (HttpOnly, Secure)
- Mọi API request đều gửi JWT trong Authorization header
- Services tự verify JWT (không cần gọi Auth Service)
- Audit log được ghi async qua Kafka

---

## 5. Error Handling Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                    ERROR HANDLING FLOW                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  CASE 1: Domain Not Allowed                                    │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  User: hacker@gmail.com                                 │    │
│  │    ↓                                                    │    │
│  │  Auth Service: Check domain                            │    │
│  │    allowed_domains = [vinfast.com, agency-partner.com] │    │
│  │    gmail.com NOT IN allowed_domains                    │    │
│  │    ↓                                                    │    │
│  │  Return: 403 Forbidden                                 │    │
│  │  Redirect: /auth/error/domain-blocked                  │    │
│  │    ↓                                                    │    │
│  │  Show user-friendly page:                              │    │
│  │  "Your email domain is not allowed.                    │    │
│  │   Please contact admin@vinfast.com"                    │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
│  CASE 2: Account Blocked                                       │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  User: ex-employee@vinfast.com                          │    │
│  │    ↓                                                    │    │
│  │  Auth Service: Check blocklist                         │    │
│  │    blocked = [ex-employee@vinfast.com]                 │    │
│  │    MATCH FOUND                                         │    │
│  │    ↓                                                    │    │
│  │  Return: 403 Forbidden                                 │    │
│  │  Redirect: /auth/error/account-blocked                 │    │
│  │    ↓                                                    │    │
│  │  Show user-friendly page:                              │    │
│  │  "Your account has been blocked.                       │    │
│  │   Contact admin for assistance."                       │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
│  CASE 3: Invalid JWT Token                                     │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  Request: GET /api/projects                             │    │
│  │  Header: Authorization: Bearer invalid_token           │    │
│  │    ↓                                                    │    │
│  │  Project Service Middleware: Verify JWT                │    │
│  │    jwt.Parse(token, publicKey)                         │    │
│  │    ERROR: signature invalid                            │    │
│  │    ↓                                                    │    │
│  │  Return: 401 Unauthorized                              │    │
│  │  Body: {"error": "Invalid token"}                      │    │
│  │    ↓                                                    │    │
│  │  Web UI: Detect 401 → Redirect to /auth/login          │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
│  CASE 4: Insufficient Permissions                              │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  User: viewer@vinfast.com (role=VIEWER)                │    │
│  │  Request: POST /api/projects (create project)          │    │
│  │    ↓                                                    │    │
│  │  Project Service Middleware:                           │    │
│  │    1. JWT valid ✓                                      │    │
│  │    2. Extract role = "VIEWER"                          │    │
│  │    3. Check permission: RequireRole("ANALYST")         │    │
│  │       VIEWER < ANALYST → FAIL                          │    │
│  │    ↓                                                    │    │
│  │  Return: 403 Forbidden                                 │    │
│  │  Body: {"error": "Insufficient permissions"}           │    │
│  │    ↓                                                    │    │
│  │  Web UI: Show error toast                              │    │
│  │  "You don't have permission to create projects"        │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Key Points:**

- Mọi error đều có user-friendly message
- Không expose technical details (security)
- Log detailed errors server-side cho debugging
- Web UI handle 401/403 gracefully
