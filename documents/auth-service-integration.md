# Auth Service Integration Guide

**Complete guide for integrating SMAP Auth Service with other microservices**

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Integration Steps](#integration-steps)
4. [Service-Specific Guides](#service-specific-guides)
5. [Configuration](#configuration)
6. [Testing](#testing)
7. [Troubleshooting](#troubleshooting)

---

## Overview

This guide provides step-by-step instructions for integrating your service with the SMAP Auth Service for authentication and authorization.

### What You'll Get

- **JWT-based authentication**: Stateless token verification
- **HttpOnly cookie support**: Enhanced XSS protection
- **Role-based authorization**: ADMIN, ANALYST, VIEWER roles
- **Group-based permissions**: Fine-grained access control via Google Groups
- **Token blacklist**: Instant revocation capability
- **Audit logging**: Automatic event tracking via Kafka
- **Service-to-service auth**: Internal API access with service keys

### Prerequisites

- Go 1.25+
- Access to Auth Service JWKS endpoint (`/.well-known/jwks.json`)
- Service key from Auth Service admin (for internal APIs)
- Redis connection (for token blacklist checking)
- Kafka connection (optional, for audit logging)

---

## Architecture

### Authentication Flow

```
┌─────────────┐                    ┌────────────────────┐
│   Browser   │  Cookie:           │   Microservice     │
│             │  smap_auth_token   │   (Project, etc)   │
│             │ ─────────────────► │                    │
│             │                    │  1. Extract JWT    │
│             │                    │  2. Verify sig     │
│             │                    │  3. Check claims   │
│             │                    │  4. Check blacklist│
│             │                    │  5. Authorize      │
└─────────────┘                    └────────────────────┘
```

### Key Components

1. **JWT Verifier**: Validates JWT tokens using public key from JWKS
2. **Auth Middleware**: Extracts and verifies tokens from cookies/headers
3. **Role Middleware**: Enforces role-based access control
4. **Blacklist Checker**: Verifies token hasn't been revoked
5. **Audit Publisher**: Publishes events to Kafka for audit trail

---

## Integration Steps

### Step 1: Add Dependencies

Update `go.mod`:

```bash
# Add pkg/auth from Auth Service
go get identity-srv/pkg/auth@latest
go get identity-srv/pkg/redis@latest
go get identity-srv/pkg/kafka@latest  # Optional, for audit logging

# Or if using local path
replace identity-srv/pkg/auth => ../identity-srv/pkg/auth
replace identity-srv/pkg/redis => ../identity-srv/pkg/redis
```

### Step 2: Add Configuration

**File**: `config/config.go`

```go
type Config struct {
    // Server Configuration
    HTTPServer HTTPServerConfig

    // Authentication & Security
    JWT     JWTConfig
    Cookie  CookieConfig
    Redis   RedisConfig

    // Optional: Audit logging
    Kafka   KafkaConfig

    // Internal service authentication
    InternalConfig InternalConfig
}

// JWT Configuration
type JWTConfig struct {
    Algorithm      string
    Issuer         string
    Audience       []string
    PublicKeyPath  string  // Path to public key file
    JWKSEndpoint   string  // Or fetch from JWKS endpoint
}

// Cookie Configuration
type CookieConfig struct {
    Domain         string
    Secure         bool
    SameSite       string
    MaxAge         int
    MaxAgeRemember int
    Name           string
}

// Redis Configuration (for blacklist)
type RedisConfig struct {
    Host     string
    Port     int
    Password string
    DB       int
}

// Kafka Configuration (optional, for audit)
type KafkaConfig struct {
    Brokers []string
    Topic   string
}

// Internal service keys
type InternalConfig struct {
    ServiceKeys map[string]string
}
```

**File**: `config/auth-config.yaml`

```yaml
# JWT Configuration
jwt:
  algorithm: RS256
  issuer: smap-auth-service
  audience:
    - identity-srv
  jwks_endpoint: https://auth-service:8080/.well-known/jwks.json
  # Or use public key file
  public_key_path: ./secrets/jwt-public.pem

# Cookie Configuration
cookie:
  domain: .smap.com
  secure: true
  samesite: Lax
  max_age: 7200 # 2 hours
  max_age_remember: 2592000 # 30 days
  name: smap_auth_token

# Redis Configuration (for blacklist)
redis:
  host: localhost
  port: 6379
  password: ""
  db: 1 # Use DB=1 for blacklist (same as Auth Service)

# Kafka Configuration (optional)
kafka:
  brokers:
    - localhost:9092
  topic: audit.events

# Internal service authentication
internal:
  service_keys:
    auth_service: <SERVICE_KEY>
```

### Step 3: Initialize JWT Verifier

**File**: `cmd/api/main.go`

```go
package main

import (
    "context"
    "time"

    "identity-srv/pkg/auth"
    "identity-srv/pkg/log"
    pkgRedis "identity-srv/pkg/redis"
)

func main() {
    ctx := context.Background()
    logger := log.Init(/* ... */)

    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        logger.Fatalf(ctx, "Failed to load config: %v", err)
    }

    // Initialize Redis for blacklist checking
    redisClient, err := pkgRedis.New(pkgRedis.Config{
        Host:     cfg.Redis.Host,
        Port:     cfg.Redis.Port,
        Password: cfg.Redis.Password,
        DB:       cfg.Redis.DB,
    })
    if err != nil {
        logger.Fatalf(ctx, "Failed to connect to Redis: %v", err)
    }

    // Initialize JWT verifier
    verifier, err := auth.NewVerifier(auth.VerifierConfig{
        JWKSEndpoint:    cfg.JWT.JWKSEndpoint,
        RefreshInterval: 1 * time.Hour,
        Issuer:          cfg.JWT.Issuer,
        Audience:        cfg.JWT.Audience,
        RedisClient:     redisClient.GetClient(),
    })
    if err != nil {
        logger.Fatalf(ctx, "Failed to initialize JWT verifier: %v", err)
    }

    logger.Infof(ctx, "JWT verifier initialized successfully")

    // Create auth middleware
    authMiddleware := auth.NewMiddleware(verifier, cfg.Cookie, logger)

    // Initialize HTTP server with middleware
    httpServer := httpserver.New(logger, httpserver.Config{
        Port:           cfg.HTTPServer.Port,
        AuthMiddleware: authMiddleware,
    })

    httpServer.Run()
}
```

### Step 4: Apply Middleware to Routes

**File**: `internal/httpserver/routes.go`

```go
package httpserver

import (
    "github.com/gin-gonic/gin"
)

func (srv *HTTPServer) mapRoutes() {
    // Public routes (no auth required)
    srv.gin.GET("/health", srv.healthCheck)

    // Protected routes - require authentication
    api := srv.gin.Group("/api")
    api.Use(srv.authMiddleware.Authenticate())
    {
        // All routes in this group require valid JWT

        // Example: Project routes
        projects := api.Group("/projects")
        {
            // VIEWER can read
            projects.GET("",
                srv.authMiddleware.RequireAnyRole("VIEWER", "ANALYST", "ADMIN"),
                srv.listProjects)
            projects.GET("/:id",
                srv.authMiddleware.RequireAnyRole("VIEWER", "ANALYST", "ADMIN"),
                srv.getProject)

            // ANALYST can create/update
            projects.POST("",
                srv.authMiddleware.RequireAnyRole("ANALYST", "ADMIN"),
                srv.createProject)
            projects.PUT("/:id",
                srv.authMiddleware.RequireAnyRole("ANALYST", "ADMIN"),
                srv.updateProject)

            // Only ADMIN can delete
            projects.DELETE("/:id",
                srv.authMiddleware.RequireRole("ADMIN"),
                srv.deleteProject)
        }
    }
}
```

### Step 5: Access User Info in Handlers

**File**: `internal/handler/project_handler.go`

```go
package handler

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "identity-srv/pkg/auth"
)

func (h *Handler) createProject(c *gin.Context) {
    ctx := c.Request.Context()

    // Get user ID from JWT claims
    userID, err := auth.GetUserID(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    // Get user role
    role, err := auth.GetUserRole(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
        return
    }

    // Get user email
    email, err := auth.GetUserEmail(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User email not found"})
        return
    }

    // Get user groups (for fine-grained permissions)
    groups, err := auth.GetUserGroups(c)
    if err != nil {
        groups = []string{} // Default to empty if not found
    }

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

    if err := h.projectService.Create(ctx, project); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
        return
    }

    // Publish audit event (optional)
    h.publishAuditEvent(ctx, AuditEvent{
        UserID:       userID,
        Action:       "CREATE_PROJECT",
        ResourceType: "project",
        ResourceID:   project.ID,
        Metadata: map[string]string{
            "project_name": project.Name,
        },
    })

    c.JSON(http.StatusCreated, project)
}
```

### Step 6: Configure CORS for Cookie Authentication

**File**: `internal/middleware/cors.go`

```go
package middleware

import (
    "github.com/gin-gonic/gin"
)

func CORS(allowedOrigins []string) gin.HandlerFunc {
    return func(c *gin.Context) {
        origin := c.Request.Header.Get("Origin")

        // Check if origin is allowed
        allowed := false
        for _, allowedOrigin := range allowedOrigins {
            if origin == allowedOrigin {
                allowed = true
                break
            }
        }

        if allowed {
            c.Header("Access-Control-Allow-Origin", origin)
            c.Header("Access-Control-Allow-Credentials", "true")
            c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
            c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
        }

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}
```

**IMPORTANT**: For cookie authentication to work:

- `Access-Control-Allow-Credentials` MUST be `true`
- `Access-Control-Allow-Origin` CANNOT be `*` (must be specific origins)
- Frontend MUST set `withCredentials: true` (axios) or `credentials: 'include'` (fetch)

---

## Service-Specific Guides

### Project Service

**Role Requirements**:

- `ANALYST` or `ADMIN`: Create/update projects
- `ADMIN`: Delete projects
- `VIEWER`: Read-only access

**Routes**:

```go
projects.GET("", RequireAnyRole("VIEWER", "ANALYST", "ADMIN"), listProjects)
projects.GET("/:id", RequireAnyRole("VIEWER", "ANALYST", "ADMIN"), getProject)
projects.POST("", RequireAnyRole("ANALYST", "ADMIN"), createProject)
projects.PUT("/:id", RequireAnyRole("ANALYST", "ADMIN"), updateProject)
projects.DELETE("/:id", RequireRole("ADMIN"), deleteProject)
```

### Ingest Service

**Role Requirements**:

- `ANALYST` or `ADMIN`: Ingest data
- `VIEWER`: View ingestion status

**Routes**:

```go
ingest.POST("/upload", RequireAnyRole("ANALYST", "ADMIN"), uploadData)
ingest.POST("/process", RequireAnyRole("ANALYST", "ADMIN"), processData)
ingest.GET("/status/:id", RequireAnyRole("VIEWER", "ANALYST", "ADMIN"), getStatus)
ingest.GET("/history", RequireAnyRole("VIEWER", "ANALYST", "ADMIN"), getHistory)
```

### Knowledge Service

**Role Requirements**:

- `VIEWER`: Read knowledge base
- `ANALYST` or `ADMIN`: Create/update knowledge
- `ADMIN`: Delete knowledge

**Routes**:

```go
knowledge.GET("", RequireAnyRole("VIEWER", "ANALYST", "ADMIN"), listKnowledge)
knowledge.GET("/:id", RequireAnyRole("VIEWER", "ANALYST", "ADMIN"), getKnowledge)
knowledge.GET("/search", RequireAnyRole("VIEWER", "ANALYST", "ADMIN"), searchKnowledge)
knowledge.POST("", RequireAnyRole("ANALYST", "ADMIN"), createKnowledge)
knowledge.PUT("/:id", RequireAnyRole("ANALYST", "ADMIN"), updateKnowledge)
knowledge.DELETE("/:id", RequireRole("ADMIN"), deleteKnowledge)
```

### Notification Service (WebSocket)

**Special Considerations**:

- WebSocket upgrade requires JWT token
- Token can be passed via query parameter or cookie
- Connection must be authenticated before upgrade

**Implementation**:

```go
func (h *Handler) HandleWebSocket(c *gin.Context) {
    // Extract JWT token from query parameter or cookie
    token := c.Query("token")
    if token == "" {
        token, _ = c.Cookie("smap_auth_token")
    }

    if token == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
        return
    }

    // Verify token
    claims, err := h.verifier.VerifyToken(token)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
        return
    }

    // Check blacklist
    if h.verifier.IsBlacklisted(c.Request.Context(), claims.ID) {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Token revoked"})
        return
    }

    // Upgrade connection
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return
    }
    defer conn.Close()

    // Create client with user info
    client := &Client{
        conn:   conn,
        userID: claims.Subject,
        email:  claims.Email,
        role:   claims.Role,
    }

    h.handleClient(client)
}
```

---

## Configuration

### Environment Variables

```env
# JWT Configuration
JWT_ALGORITHM=RS256
JWT_ISSUER=smap-auth-service
JWT_AUDIENCE=identity-srv
JWT_JWKS_ENDPOINT=https://auth-service:8080/.well-known/jwks.json

# Cookie Configuration
COOKIE_DOMAIN=.smap.com
COOKIE_SECURE=true
COOKIE_SAMESITE=Lax
COOKIE_MAX_AGE=7200
COOKIE_NAME=smap_auth_token

# Redis Configuration (for blacklist)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=1

# Kafka Configuration (optional)
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=audit.events

# CORS Configuration
CORS_ALLOWED_ORIGINS=https://app.smap.com,http://localhost:3000
```

### Shared Configuration

**IMPORTANT**: All services MUST use the same:

- JWT issuer: `smap-auth-service`
- JWT audience: `identity-srv`
- Cookie name: `smap_auth_token`
- Cookie domain: `.smap.com`
- Redis DB: `1` (for blacklist)

---

## Testing

### Manual Testing

```bash
# 1. Get JWT token from Auth Service
curl -X POST "https://auth-service.com/authentication/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password"}' \
  --cookie-jar cookies.txt

# 2. Test authenticated endpoint with cookie
curl -X GET "https://your-service.com/api/projects" \
  --cookie cookies.txt

# 3. Test with Authorization header (fallback)
TOKEN=$(cat cookies.txt | grep smap_auth_token | awk '{print $7}')
curl -X GET "https://your-service.com/api/projects" \
  -H "Authorization: Bearer $TOKEN"

# 4. Test authorization (should fail with 403)
curl -X DELETE "https://your-service.com/api/projects/test-id" \
  --cookie cookies.txt
```

### Integration Tests

```go
func TestAuthentication(t *testing.T) {
    // Setup
    router := setupTestRouter()

    // Test without auth - should fail
    req := httptest.NewRequest("GET", "/api/projects", nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    assert.Equal(t, http.StatusUnauthorized, w.Code)

    // Test with valid token - should succeed
    req = httptest.NewRequest("GET", "/api/projects", nil)
    req.AddCookie(&http.Cookie{
        Name:  "smap_auth_token",
        Value: validToken,
    })
    w = httptest.NewRecorder()
    router.ServeHTTP(w, req)
    assert.Equal(t, http.StatusOK, w.Code)
}
```

---

## Troubleshooting

### Issue 1: JWT Verification Fails

**Symptoms**: All requests return 401 Unauthorized

**Solutions**:

- Verify JWKS endpoint is reachable: `curl https://auth-service/.well-known/jwks.json`
- Check issuer and audience configuration match Auth Service
- Verify Redis connection for blacklist checking
- Check logs for specific error messages

### Issue 2: Cookie Not Being Sent

**Symptoms**: Token in cookie but not received by service

**Solutions**:

- Verify cookie domain matches service domain (`.smap.com`)
- Check `Secure` flag - must be `false` for HTTP (dev only)
- Verify CORS `Access-Control-Allow-Credentials: true`
- Check browser DevTools → Application → Cookies

### Issue 3: Role Authorization Not Working

**Symptoms**: Users with correct role get 403 Forbidden

**Solutions**:

- Verify role claim in JWT token
- Check role middleware configuration
- Ensure role names match exactly (case-sensitive: `ADMIN` not `admin`)
- Check if user has required groups

### Issue 4: CORS Errors

**Symptoms**: "Access-Control-Allow-Origin" errors in browser console

**Solutions**:

- Cannot use wildcard `*` with credentials
- Must specify exact origins in CORS config
- Frontend must set `withCredentials: true` or `credentials: 'include'`
- Check `Access-Control-Allow-Credentials: true` in response

### Issue 5: Token Blacklist Not Working

**Symptoms**: Revoked tokens still work

**Solutions**:

- Verify Redis connection
- Check Redis DB number (must be 1)
- Verify blacklist check is called in middleware
- Check Redis key format: `blacklist:{jti}`

---

## Audit Event Publishing

### Setup Kafka Producer

```go
import pkgKafka "identity-srv/pkg/kafka"

// In main.go
kafkaProducer, err := pkgKafka.NewProducer(pkgKafka.Config{
    Brokers: cfg.Kafka.Brokers,
    Topic:   "audit.events",
})
```

### Publish Events

```go
type AuditEvent struct {
    UserID       string            `json:"user_id"`
    Action       string            `json:"action"`
    ResourceType string            `json:"resource_type"`
    ResourceID   string            `json:"resource_id"`
    Metadata     map[string]string `json:"metadata"`
    IPAddress    string            `json:"ip_address"`
    UserAgent    string            `json:"user_agent"`
    Timestamp    time.Time         `json:"timestamp"`
}

func (h *Handler) publishAuditEvent(ctx context.Context, event AuditEvent) {
    if h.kafkaProducer == nil {
        return
    }

    data, _ := json.Marshal(event)
    h.kafkaProducer.Publish([]byte(event.UserID), data)
}
```

### Events to Publish

- `CREATE_PROJECT`, `UPDATE_PROJECT`, `DELETE_PROJECT`
- `CREATE_SOURCE`, `DELETE_SOURCE`, `EXPORT_DATA`
- `CREATE_KNOWLEDGE`, `UPDATE_KNOWLEDGE`, `DELETE_KNOWLEDGE`

---

## Next Steps

After successful integration:

1. **Monitor metrics**: Track auth success/failure rates
2. **Set up alerts**: Alert on high 401/403 rates
3. **Review logs**: Check for auth-related errors
4. **Performance test**: Ensure JWT verification doesn't impact latency
5. **Document**: Update service docs with auth requirements

---

## References

- **API Reference**: `documents/api-reference.md`
- **Deployment Guide**: `documents/deployment-guide.md`
- **Troubleshooting**: `documents/identity-service-troubleshooting.md`
- **Auth Service Gaps**: `documents/auth-service-gaps-proposal.md`

---

**Last Updated**: 09/02/2026  
**Version**: 2.0
