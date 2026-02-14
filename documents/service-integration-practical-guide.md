# Hướng Dẫn Thực Tế: Tích Hợp Auth Service Vào Service Khác

> **Mục đích**: Hướng dẫn chi tiết cách implement authentication middleware cho service khác sử dụng HttpOnly Cookies

---

## 📋 Checklist Tích Hợp

- [ ] Copy package `pkg/auth` từ Auth Service
- [ ] Config CORS để nhận cookies từ frontend
- [ ] Config JWT verifier với JWKS endpoint
- [ ] Connect Redis để check blacklist
- [ ] Apply middleware vào routes
- [ ] Test với Postman/Browser

---

## 🚀 Bước 1: Copy Auth Package

Auth Service đã cung cấp sẵn package `pkg/auth` để tái sử dụng.

### Cách 1: Copy Trực Tiếp (Khuyến nghị cho development)

```bash
# Từ service mới của bạn
cp -r ../identity-srv/pkg/auth ./pkg/
cp -r ../identity-srv/pkg/scope ./pkg/
```

### Cách 2: Go Module Replace (Khuyến nghị cho production)

```go
// go.mod của service mới
module your-service

require (
    smap-api/pkg/auth v0.0.0
    smap-api/pkg/scope v0.0.0
)

replace smap-api/pkg/auth => ../identity-srv/pkg/auth
replace smap-api/pkg/scope => ../identity-srv/pkg/scope
```

---

## 🔧 Bước 2: Cấu Hình Service

### File: `config/config.yaml`

```yaml
# Service Configuration
service:
  name: project-service
  port: 8081

# JWT Configuration - PHẢI KHỚP VỚI AUTH SERVICE
jwt:
  algorithm: HS256
  issuer: smap-auth-service # ✅ Phải giống Auth Service
  audience:
    - smap-api # ✅ Phải giống Auth Service
  secret_key: <same-secret-key-as-auth-service> # ✅ Phải giống Auth Service

# Cookie Configuration - PHẢI KHỚP VỚI AUTH SERVICE
cookie:
  domain: localhost # ✅ Development: localhost
  secure: false # ✅ Development: false (HTTP)
  samesite: Lax # ✅ Cho phép cross-site với redirects
  name: smap_auth_token # ✅ Phải giống Auth Service

# Redis Configuration - ĐỂ CHECK BLACKLIST
redis:
  host: localhost
  port: 6379
  password: ""
  db: 0 # ✅ DB 0 cho blacklist (giống Auth Service)

# CORS Configuration
cors:
  allowed_origins:
    - http://localhost:3000 # Frontend dev
    - http://localhost:5173 # Vite dev
    - http://localhost:8080 # Auth service
  allow_credentials: true # ✅ BẮT BUỘC cho cookies
```

### ⚠️ LƯU Ý QUAN TRỌNG

**Các config SAU PHẢI KHỚP với Auth Service**:

- `jwt.issuer`: `smap-auth-service`
- `jwt.audience`: `["smap-api"]`
- `jwt.secret_key`: Same secret key as Auth Service
- `cookie.name`: `smap_auth_token`
- `cookie.domain`: `localhost` (dev) hoặc `.smap.com` (prod)
- `redis.db`: `0` (blacklist database)

---

## 💻 Bước 3: Implement Middleware

### File: `cmd/api/main.go`

```go
package main

import (
    "context"
    "log"
    "time"

    "your-service/config"
    "your-service/internal/httpserver"

    "smap-api/pkg/auth"
    pkgRedis "smap-api/pkg/redis"

    "github.com/gin-gonic/gin"
)

func main() {
    ctx := context.Background()

    // 1. Load Configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // 2. Initialize Redis (for blacklist checking)
    redisClient, err := pkgRedis.New(pkgRedis.Config{
        Host:     cfg.Redis.Host,
        Port:     cfg.Redis.Port,
        Password: cfg.Redis.Password,
        DB:       cfg.Redis.DB, // Must be 0 for blacklist
    })
    if err != nil {
        log.Fatalf("Failed to connect to Redis: %v", err)
    }
    log.Println("✅ Redis connected")

    // 3. Initialize JWT Verifier
    verifier, err := auth.NewVerifier(auth.VerifierConfig{
        SecretKey: cfg.JWT.SecretKey,
        Issuer:    cfg.JWT.Issuer,
        Audience:  cfg.JWT.Audience,
    })
    if err != nil {
        log.Fatalf("Failed to initialize JWT verifier: %v", err)
    }
    log.Println("✅ JWT Verifier initialized")

    // 4. Create Auth Middleware
    authMiddleware := auth.NewMiddleware(auth.MiddlewareConfig{
        Verifier:       verifier,
        BlacklistRedis: redisClient.GetClient(), // Enable blacklist checking
        CookieName:     cfg.Cookie.Name,
    })
    log.Println("✅ Auth Middleware created")

    // 5. Initialize HTTP Server
    router := gin.Default()

    // Apply CORS middleware
    router.Use(corsMiddleware(cfg.CORS))

    // Map routes with auth middleware
    mapRoutes(router, authMiddleware)

    // 6. Start Server
    log.Printf("🚀 Server starting on port %d", cfg.Service.Port)
    if err := router.Run(fmt.Sprintf(":%d", cfg.Service.Port)); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}
```

---

## 🛣️ Bước 4: Apply Middleware Vào Routes

### File: `internal/httpserver/routes.go`

```go
package httpserver

import (
    "smap-api/pkg/auth"
    "github.com/gin-gonic/gin"
)

func mapRoutes(router *gin.Engine, authMW *auth.Middleware) {
    // ========================================
    // PUBLIC ROUTES (No authentication)
    // ========================================
    router.GET("/health", healthCheck)
    router.GET("/version", getVersion)

    // ========================================
    // PROTECTED ROUTES (Require authentication)
    // ========================================
    api := router.Group("/api")
    api.Use(authMW.Authenticate()) // ✅ Apply auth middleware
    {
        // Projects - Role-based access
        projects := api.Group("/projects")
        {
            // VIEWER, ANALYST, ADMIN can read
            projects.GET("",
                authMW.RequireAnyRole("VIEWER", "ANALYST", "ADMIN"),
                listProjects)

            projects.GET("/:id",
                authMW.RequireAnyRole("VIEWER", "ANALYST", "ADMIN"),
                getProject)

            // ANALYST, ADMIN can create/update
            projects.POST("",
                authMW.RequireAnyRole("ANALYST", "ADMIN"),
                createProject)

            projects.PUT("/:id",
                authMW.RequireAnyRole("ANALYST", "ADMIN"),
                updateProject)

            // Only ADMIN can delete
            projects.DELETE("/:id",
                authMW.RequireRole("ADMIN"),
                deleteProject)
        }

        // Users - Admin only
        users := api.Group("/users")
        users.Use(authMW.RequireRole("ADMIN"))
        {
            users.GET("", listUsers)
            users.GET("/:id", getUser)
            users.PUT("/:id", updateUser)
            users.DELETE("/:id", deleteUser)
        }
    }
}
```

### 📝 Giải Thích Middleware Chain

```go
// 1. authMW.Authenticate()
//    - Extract JWT từ cookie hoặc Authorization header
//    - Verify JWT signature với public key
//    - Check token expiration
//    - Check blacklist (nếu có Redis)
//    - Inject claims vào context
//    - Nếu fail → 401 Unauthorized

// 2. authMW.RequireRole("ADMIN")
//    - Lấy claims từ context
//    - Check user có role "ADMIN" không
//    - Nếu không → 403 Forbidden

// 3. authMW.RequireAnyRole("ANALYST", "ADMIN")
//    - Check user có BẤT KỲ role nào trong list
//    - Nếu không → 403 Forbidden
```

---

## 🎨 Bước 5: Implement CORS Middleware

### File: `internal/middleware/cors.go`

```go
package middleware

import (
    "strings"
    "github.com/gin-gonic/gin"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
    AllowedOrigins   []string
    AllowCredentials bool
}

// CORS returns CORS middleware
func CORS(cfg CORSConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        origin := c.GetHeader("Origin")

        // Check if origin is allowed
        allowed := false
        for _, allowedOrigin := range cfg.AllowedOrigins {
            if origin == allowedOrigin {
                allowed = true
                break
            }
        }

        if allowed {
            // ✅ QUAN TRỌNG: Set origin cụ thể, KHÔNG dùng "*"
            c.Header("Access-Control-Allow-Origin", origin)

            // ✅ QUAN TRỌNG: Phải có để browser gửi cookies
            c.Header("Access-Control-Allow-Credentials", "true")

            c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
            c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
            c.Header("Access-Control-Max-Age", "86400") // 24 hours
        }

        // Handle preflight requests
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}
```

### ⚠️ CORS Requirements Cho HttpOnly Cookies

**BẮT BUỘC**:

1. `Access-Control-Allow-Origin` KHÔNG thể là `*` → Phải là origin cụ thể
2. `Access-Control-Allow-Credentials` PHẢI là `true`
3. Frontend PHẢI gửi `credentials: 'include'` (fetch) hoặc `withCredentials: true` (axios)

---

## 🔍 Bước 6: Sử Dụng User Info Trong Handler

### File: `internal/handler/project_handler.go`

```go
package handler

import (
    "net/http"
    "smap-api/pkg/auth"
    "github.com/gin-gonic/gin"
)

func createProject(c *gin.Context) {
    ctx := c.Request.Context()

    // ✅ Lấy user ID từ JWT claims
    userID, ok := auth.GetUserIDFromContext(ctx)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    // ✅ Lấy user role
    role, ok := auth.GetUserRoleFromContext(ctx)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
        return
    }

    // ✅ Lấy user email
    claims, ok := auth.GetClaimsFromContext(ctx)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Claims not found"})
        return
    }
    email := claims.Email

    // ✅ Lấy user groups (for fine-grained permissions)
    groups := claims.Groups

    // Your business logic
    var req CreateProjectRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    project := &Project{
        Name:      req.Name,
        CreatedBy: userID,
    }

    // Save to database
    if err := projectRepo.Create(ctx, project); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "data": project,
        "meta": gin.H{
            "created_by": email,
            "role":       role,
        },
    })
}
```

### 📚 Available Helper Functions

```go
// From pkg/auth/helpers.go

// Get user information
userID := auth.GetUserID(ctx)
role := auth.GetUserRole(ctx)
email := auth.GetUserEmail(ctx)
groups := auth.GetUserGroups(ctx)

// Check authentication
if !auth.IsAuthenticated(ctx) {
    // Not authenticated
}

// Check roles
if auth.IsAdmin(ctx) {
    // User is ADMIN
}

if auth.IsAnalyst(ctx) {
    // User is ANALYST
}

// Check permissions
if auth.HasPermission(ctx, "projects:delete") {
    // User has permission
}

// Check resource ownership
if auth.CanAccessResource(ctx, resourceOwnerID) {
    // User can access this resource
}
```

---

## 🧪 Bước 7: Testing

### Test 1: Health Check (Public Route)

```bash
curl http://localhost:8081/health
# Expected: 200 OK
```

### Test 2: Protected Route Without Auth

```bash
curl http://localhost:8081/api/projects
# Expected: 401 Unauthorized
```

### Test 3: Login và Get Cookie

```bash
# 1. Login qua Auth Service
curl -c cookies.txt -L \
  http://localhost:8080/authentication/login

# 2. Sau khi login qua browser, extract cookie
# Mở DevTools → Application → Cookies → Copy value
```

### Test 4: Request Với Cookie

```bash
# Tạo file cookies.txt thủ công
echo "localhost	FALSE	/	FALSE	0	smap_auth_token	YOUR_JWT_TOKEN" > cookies.txt

# Test request
curl -b cookies.txt \
  http://localhost:8081/api/projects

# Expected: 200 OK với danh sách projects
```

### Test 5: Test Với Browser/Postman

Xem chi tiết tại: `documents/local-testing-guide.md`

---

## ⚠️ Common Issues & Solutions

### Issue 1: 401 Unauthorized - "Invalid signature"

**Nguyên nhân**: JWT được sign bằng key khác với secret key.

**Giải pháp**:

```bash
# Verify secret key matches Auth Service
grep secret_key config/auth-config.yaml

# Check JWT issuer and audience match
grep -A 3 "jwt:" config/auth-config.yaml
```

### Issue 3: Cookie Không Được Gửi

**Nguyên nhân**: CORS không được config đúng.

**Giải pháp**:

```yaml
# config.yaml
cors:
  allowed_origins:
    - http://localhost:3000 # ✅ Phải là origin cụ thể
  allow_credentials: true # ✅ Phải là true
```

```javascript
// Frontend
fetch("http://localhost:8081/api/projects", {
  credentials: "include", // ✅ Phải có
});
```

### Issue 4: 403 Forbidden - Role Check Failed

**Nguyên nhân**: User không có role yêu cầu.

**Debug**:

```go
// Thêm log để debug
claims, _ := auth.GetClaimsFromContext(ctx)
log.Printf("User role: %s, Required: ADMIN", claims.Role)
```

**Giải pháp**:

- Verify user có đúng role trong Auth Service
- Check role name match exactly (case-sensitive)
- Decode JWT tại jwt.io để xem claims

### Issue 5: Blacklist Check Failed

**Nguyên nhân**: Redis connection issue hoặc wrong DB.

**Giải pháp**:

```bash
# Test Redis connection
docker exec -it redis redis-cli

# Check blacklist DB (should be 0)
SELECT 0
KEYS blacklist:*

# If empty, blacklist is working (no revoked tokens)
```

---

## 📊 Testing Checklist

### ✅ Pre-Integration Testing

- [ ] Auth Service đang chạy
- [ ] Redis đang chạy: `docker ps | grep redis`
- [ ] Config file đã được setup đúng
- [ ] JWT secret key matches Auth Service

### ✅ Integration Testing

- [ ] Public routes work without auth
- [ ] Protected routes return 401 without auth
- [ ] Login qua Auth Service thành công
- [ ] Cookie được set sau login
- [ ] Protected routes work với cookie
- [ ] Role-based routes enforce permissions
- [ ] Logout revokes token (blacklist works)

### ✅ CORS Testing

- [ ] Preflight OPTIONS requests return 204
- [ ] `Access-Control-Allow-Origin` header present
- [ ] `Access-Control-Allow-Credentials: true` header present
- [ ] Browser không show CORS errors
- [ ] Cookies được gửi từ frontend

---

## 🎯 Best Practices

### 1. Environment-Specific Configuration

```yaml
# Development
cookie:
  domain: localhost
  secure: false
  samesite: Lax

# Production
cookie:
  domain: .smap.com
  secure: true
  samesite: Strict
```

### 2. Error Handling

```go
func (h *Handler) createProject(c *gin.Context) {
    claims, ok := auth.GetClaimsFromContext(c.Request.Context())
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{
            "error": gin.H{
                "code":    "UNAUTHORIZED",
                "message": "Authentication required",
            },
        })
        return
    }

    // Business logic...
}
```

### 3. Logging

```go
// Log authentication events
log.Printf("User %s (%s) accessed %s %s",
    claims.Email,
    claims.Role,
    c.Request.Method,
    c.Request.URL.Path)
```

### 4. Monitoring

```go
// Track auth failures
if err := authMiddleware.Authenticate()(c); err != nil {
    metrics.IncrementAuthFailures()
}
```

---

## 📚 Tài Liệu Liên Quan

- **Auth Package README**: `pkg/auth/README.md`
- **Integration Guide**: `documents/auth-service-integration.md`
- **Local Testing Guide**: `documents/local-testing-guide.md`
- **API Reference**: `documents/api-reference.md`

---

**Cập nhật lần cuối**: 14/02/2026
