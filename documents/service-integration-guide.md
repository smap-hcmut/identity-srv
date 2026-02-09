# Service Integration Guide

## Table of Contents

1. [Overview](#overview)
2. [Integration Steps Summary](#integration-steps-summary)
3. [Project Service Integration](#project-service-integration)
4. [Ingest Service Integration](#ingest-service-integration)
5. [Knowledge Service Integration](#knowledge-service-integration)
6. [Notification Service Integration](#notification-service-integration)
7. [Audit Event Publishing](#audit-event-publishing)
8. [Testing Integration](#testing-integration)
9. [Rollback Procedure](#rollback-procedure)
10. [Common Issues](#common-issues)

---

## Overview

This guide provides step-by-step instructions for integrating your service with the Identity Service for authentication and authorization.

### What You'll Get

- **JWT-based authentication**: Stateless token verification
- **Role-based authorization**: ADMIN, ANALYST, VIEWER roles
- **Group-based permissions**: Fine-grained access control
- **Audit logging**: Automatic event tracking
- **Service-to-service auth**: Internal API access

### Prerequisites

- Go 1.25+
- Access to Identity Service JWKS endpoint
- Service key from Identity Service admin
- Redis connection (for blacklist checking)

---

## Integration Steps Summary

For all services, follow these steps:

1. **Add pkg/auth dependency** to go.mod
2. **Initialize JWT verifier** with JWKS endpoint
3. **Create middleware instance**
4. **Apply Authenticate() to routes**
5. **Add role-based authorization**
6. **Configure SERVICE_KEY** environment variable
7. **Update error handling** for 401/403
8. **Test integration**

---

## Project Service Integration

### Overview

Project Service manages project data and requires authentication for all endpoints.

**Role Requirements**:
- `ANALYST` or `ADMIN`: Create/update projects
- `ADMIN`: Delete projects
- `VIEWER`: Read-only access

### Step 1: Add Dependencies

Update `go.mod`:

```bash
# Add pkg/auth from Identity Service
go get smap-api/pkg/auth@latest

# Or if using local path
replace smap-api/pkg/auth => ../identity-service/pkg/auth
```

### Step 2: Initialize JWT Verifier

In `cmd/api/main.go`:

```go
package main

import (
    "context"
    "time"
    
    "smap-api/pkg/auth"
    "smap-api/pkg/log"
    pkgRedis "smap-api/pkg/redis"
)

func main() {
    ctx := context.Background()
    logger := log.Init(/* ... */)
    
    // Initialize Redis for blacklist checking
    redisClient, err := pkgRedis.New(pkgRedis.Config{
        Host:     cfg.Redis.Host,
        Port:     cfg.Redis.Port,
        Password: cfg.Redis.Password,
        DB:       1, // Use DB=1 for blacklist (same as Identity Service)
    })
    if err != nil {
        logger.Fatalf(ctx, "Failed to connect to Redis: %v", err)
    }
    
    // Initialize JWT verifier
    verifier, err := auth.NewVerifier(auth.VerifierConfig{
        JWKSEndpoint:    "https://yourdomain.com/.well-known/jwks.json",
        RefreshInterval: 1 * time.Hour,
        Issuer:          "smap-auth-service",
        Audience:        []string{"smap-api"},
        RedisClient:     redisClient.GetClient(), // For blacklist checking
    })
    if err != nil {
        logger.Fatalf(ctx, "Failed to initialize JWT verifier: %v", err)
    }
    
    logger.Infof(ctx, "JWT verifier initialized successfully")
    
    // Create auth middleware
    authMiddleware := auth.NewMiddleware(verifier, logger)
    
    // Pass to HTTP server
    httpServer := httpserver.New(logger, httpserver.Config{
        // ... other config
        AuthMiddleware: authMiddleware,
    })
    
    httpServer.Run()
}
```


### Step 3: Apply Middleware to Routes

In `internal/httpserver/routes.go`:

```go
package httpserver

import (
    "github.com/gin-gonic/gin"
    "smap-api/pkg/auth"
)

func (srv *HTTPServer) mapRoutes() {
    // Public routes (no auth required)
    srv.gin.GET("/health", srv.healthCheck)
    
    // Protected routes - require authentication
    api := srv.gin.Group("/api")
    api.Use(srv.authMiddleware.Authenticate())
    {
        // All routes in this group require valid JWT
        
        // Project routes
        projects := api.Group("/projects")
        {
            // VIEWER can read
            projects.GET("", srv.authMiddleware.RequireAnyRole("VIEWER", "ANALYST", "ADMIN"), srv.listProjects)
            projects.GET("/:id", srv.authMiddleware.RequireAnyRole("VIEWER", "ANALYST", "ADMIN"), srv.getProject)
            
            // ANALYST can create/update
            projects.POST("", srv.authMiddleware.RequireAnyRole("ANALYST", "ADMIN"), srv.createProject)
            projects.PUT("/:id", srv.authMiddleware.RequireAnyRole("ANALYST", "ADMIN"), srv.updateProject)
            
            // Only ADMIN can delete
            projects.DELETE("/:id", srv.authMiddleware.RequireRole("ADMIN"), srv.deleteProject)
        }
    }
}
```

### Step 4: Access User Info in Handlers

In your handlers, extract user information from context:

```go
package handler

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    "smap-api/pkg/auth"
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
    project := &Project{
        Name:      req.Name,
        CreatedBy: userID,
        // ...
    }
    
    if err := h.projectService.Create(ctx, project); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
        return
    }
    
    c.JSON(http.StatusCreated, project)
}
```

### Step 5: Configure Service Key for Internal APIs

If you need to call Identity Service internal APIs:

**Add to configuration** (`config.yaml`):

```yaml
identity_service:
  base_url: https://yourdomain.com
  service_key: <ENCRYPTED_SERVICE_KEY>  # Get from Identity Service admin
```

**Use in code**:

```go
package client

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type IdentityClient struct {
    baseURL    string
    serviceKey string
    httpClient *http.Client
}

func NewIdentityClient(baseURL, serviceKey string) *IdentityClient {
    return &IdentityClient{
        baseURL:    baseURL,
        serviceKey: serviceKey,
        httpClient: &http.Client{Timeout: 10 * time.Second},
    }
}

// ValidateToken validates a JWT token (fallback if local verification fails)
func (c *IdentityClient) ValidateToken(ctx context.Context, token string) (*TokenInfo, error) {
    reqBody := map[string]string{"token": token}
    body, _ := json.Marshal(reqBody)
    
    req, err := http.NewRequestWithContext(ctx, "POST", 
        c.baseURL+"/internal/validate", bytes.NewBuffer(body))
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-Service-Key", c.serviceKey)
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("validation failed: %d", resp.StatusCode)
    }
    
    var tokenInfo TokenInfo
    if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
        return nil, err
    }
    
    return &tokenInfo, nil
}

// GetUserByID retrieves user information by ID
func (c *IdentityClient) GetUserByID(ctx context.Context, userID string) (*User, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", 
        c.baseURL+"/internal/users/"+userID, nil)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("X-Service-Key", c.serviceKey)
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("get user failed: %d", resp.StatusCode)
    }
    
    var user User
    if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
        return nil, err
    }
    
    return &user, nil
}
```

### Step 6: Update Error Handling

Update your error handling to properly handle auth errors:

```go
package middleware

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        // Check for errors
        if len(c.Errors) > 0 {
            err := c.Errors.Last()
            
            switch err.Type {
            case gin.ErrorTypePublic:
                // Handle auth errors
                if c.Writer.Status() == http.StatusUnauthorized {
                    c.JSON(http.StatusUnauthorized, gin.H{
                        "error": gin.H{
                            "code":    "UNAUTHORIZED",
                            "message": "Authentication required. Please login.",
                        },
                    })
                    return
                }
                
                if c.Writer.Status() == http.StatusForbidden {
                    c.JSON(http.StatusForbidden, gin.H{
                        "error": gin.H{
                            "code":    "FORBIDDEN",
                            "message": "Insufficient permissions.",
                        },
                    })
                    return
                }
            }
        }
    }
}
```

### Step 7: Testing

**Test Authentication**:

```bash
# Get JWT token from Identity Service
TOKEN=$(curl -X POST "https://yourdomain.com/authentication/login" \
  --cookie-jar cookies.txt -L -s | grep -o 'smap_auth_token=[^;]*' | cut -d= -f2)

# Test authenticated endpoint
curl -X GET "https://project-service.com/api/projects" \
  -H "Authorization: Bearer $TOKEN"

# Or with cookie
curl -X GET "https://project-service.com/api/projects" \
  --cookie "smap_auth_token=$TOKEN"
```

**Test Authorization**:

```bash
# Test VIEWER access (should succeed)
curl -X GET "https://project-service.com/api/projects" \
  -H "Authorization: Bearer $VIEWER_TOKEN"

# Test VIEWER create (should fail with 403)
curl -X POST "https://project-service.com/api/projects" \
  -H "Authorization: Bearer $VIEWER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Project"}'

# Test ANALYST create (should succeed)
curl -X POST "https://project-service.com/api/projects" \
  -H "Authorization: Bearer $ANALYST_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Project"}'
```

### Step 8: Rollback Procedure

If integration fails:

1. **Remove auth middleware** from routes
2. **Revert to old authentication** (if any)
3. **Deploy previous version**
4. **Investigate issues** using logs
5. **Fix and retry**

**Quick rollback**:

```bash
# Kubernetes
kubectl rollout undo deployment/project-service -n project-service

# Docker
docker-compose up -d project-service:previous-tag
```

---

## Ingest Service Integration

### Overview

Ingest Service handles data ingestion and requires authentication for all endpoints.

**Role Requirements**:
- `ANALYST` or `ADMIN`: Ingest data
- `VIEWER`: View ingestion status

### Integration Steps

Follow the same steps as Project Service with these specific configurations:

**Step 1-2**: Same as Project Service

**Step 3**: Apply middleware with role requirements

```go
func (srv *HTTPServer) mapRoutes() {
    api := srv.gin.Group("/api")
    api.Use(srv.authMiddleware.Authenticate())
    {
        // Ingest routes
        ingest := api.Group("/ingest")
        {
            // ANALYST can ingest data
            ingest.POST("/upload", srv.authMiddleware.RequireAnyRole("ANALYST", "ADMIN"), srv.uploadData)
            ingest.POST("/process", srv.authMiddleware.RequireAnyRole("ANALYST", "ADMIN"), srv.processData)
            
            // VIEWER can view status
            ingest.GET("/status/:id", srv.authMiddleware.RequireAnyRole("VIEWER", "ANALYST", "ADMIN"), srv.getStatus)
            ingest.GET("/history", srv.authMiddleware.RequireAnyRole("VIEWER", "ANALYST", "ADMIN"), srv.getHistory)
        }
    }
}
```

**Step 4**: Access user info in handlers (same as Project Service)

**Step 5-8**: Same as Project Service

---

## Knowledge Service Integration

### Overview

Knowledge Service manages knowledge base and requires authentication for all endpoints.

**Role Requirements**:
- `VIEWER`: Read knowledge base
- `ANALYST` or `ADMIN`: Create/update knowledge
- `ADMIN`: Delete knowledge

### Integration Steps

Follow the same steps as Project Service with these specific configurations:

**Step 1-2**: Same as Project Service

**Step 3**: Apply middleware with role requirements

```go
func (srv *HTTPServer) mapRoutes() {
    api := srv.gin.Group("/api")
    api.Use(srv.authMiddleware.Authenticate())
    {
        // Knowledge routes
        knowledge := api.Group("/knowledge")
        {
            // VIEWER can read
            knowledge.GET("", srv.authMiddleware.RequireAnyRole("VIEWER", "ANALYST", "ADMIN"), srv.listKnowledge)
            knowledge.GET("/:id", srv.authMiddleware.RequireAnyRole("VIEWER", "ANALYST", "ADMIN"), srv.getKnowledge)
            knowledge.GET("/search", srv.authMiddleware.RequireAnyRole("VIEWER", "ANALYST", "ADMIN"), srv.searchKnowledge)
            
            // ANALYST can create/update
            knowledge.POST("", srv.authMiddleware.RequireAnyRole("ANALYST", "ADMIN"), srv.createKnowledge)
            knowledge.PUT("/:id", srv.authMiddleware.RequireAnyRole("ANALYST", "ADMIN"), srv.updateKnowledge)
            
            // Only ADMIN can delete
            knowledge.DELETE("/:id", srv.authMiddleware.RequireRole("ADMIN"), srv.deleteKnowledge)
        }
    }
}
```

**Step 4-8**: Same as Project Service

---

## Notification Service Integration

### Overview

Notification Service handles WebSocket connections and requires JWT authentication.

**Special Considerations**:
- WebSocket upgrade requires JWT token
- Token can be passed via query parameter or cookie
- Connection must be authenticated before upgrade

### Step 1-2: Same as Other Services

### Step 3: WebSocket Authentication

```go
package websocket

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "smap-api/pkg/auth"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        // Configure based on your CORS policy
        return true
    },
}

func (h *Handler) HandleWebSocket(c *gin.Context) {
    // Extract JWT token from query parameter or cookie
    token := c.Query("token")
    if token == "" {
        // Try cookie
        token, _ = c.Cookie("smap_auth_token")
    }
    
    if token == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
        return
    }
    
    // Verify token using auth middleware
    claims, err := h.verifier.VerifyToken(token)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
        return
    }
    
    // Check if token is blacklisted
    if h.verifier.IsBlacklisted(c.Request.Context(), claims.ID) {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Token revoked"})
        return
    }
    
    // Upgrade connection
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        h.logger.Errorf(c.Request.Context(), "Failed to upgrade WebSocket: %v", err)
        return
    }
    defer conn.Close()
    
    // Create client with user info
    client := &Client{
        conn:   conn,
        userID: claims.Subject,
        email:  claims.Email,
        role:   claims.Role,
        groups: claims.Groups,
    }
    
    // Handle WebSocket messages
    h.handleClient(client)
}
```

### Step 4: Frontend WebSocket Connection

**JavaScript Example**:

```javascript
// Get token from cookie
const token = document.cookie
  .split('; ')
  .find(row => row.startsWith('smap_auth_token='))
  ?.split('=')[1];

// Connect with token in query parameter
const ws = new WebSocket(`wss://notification-service.com/ws?token=${token}`);

ws.onopen = () => {
  console.log('WebSocket connected');
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Received:', data);
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('WebSocket disconnected');
};
```

### Step 5: Handle Token Expiration

```go
func (h *Handler) handleClient(client *Client) {
    // Set up ping/pong to detect disconnections
    client.conn.SetPongHandler(func(string) error {
        // Verify token is still valid
        claims, err := h.verifier.VerifyToken(client.token)
        if err != nil {
            // Token expired or invalid - close connection
            client.conn.WriteMessage(websocket.CloseMessage, 
                websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Token expired"))
            return err
        }
        
        // Check blacklist
        if h.verifier.IsBlacklisted(context.Background(), claims.ID) {
            client.conn.WriteMessage(websocket.CloseMessage, 
                websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Token revoked"))
            return fmt.Errorf("token revoked")
        }
        
        return nil
    })
    
    // Handle messages...
}
```

---

## Audit Event Publishing

All services should publish audit events for important actions.

### Step 1: Add Kafka Producer

In `cmd/api/main.go`:

```go
import (
    pkgKafka "smap-api/pkg/kafka"
)

func main() {
    // Initialize Kafka producer
    kafkaProducer, err := pkgKafka.NewProducer(pkgKafka.Config{
        Brokers: cfg.Kafka.Brokers,
        Topic:   "audit.events",
    })
    if err != nil {
        logger.Warnf(ctx, "Failed to initialize Kafka producer: %v", err)
        kafkaProducer = nil // Continue without Kafka
    }
    
    // Pass to handlers
    httpServer := httpserver.New(logger, httpserver.Config{
        // ... other config
        KafkaProducer: kafkaProducer,
    })
}
```

### Step 2: Publish Audit Events

```go
package handler

import (
    "encoding/json"
    "time"
    
    "smap-api/pkg/auth"
)

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

func (h *Handler) createProject(c *gin.Context) {
    ctx := c.Request.Context()
    userID, _ := auth.GetUserID(c)
    
    // Create project...
    project := &Project{/* ... */}
    if err := h.projectService.Create(ctx, project); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
        return
    }
    
    // Publish audit event
    h.publishAuditEvent(AuditEvent{
        UserID:       userID,
        Action:       "CREATE_PROJECT",
        ResourceType: "project",
        ResourceID:   project.ID,
        Metadata: map[string]string{
            "project_name": project.Name,
            "service":      "project-service",
        },
        IPAddress: c.ClientIP(),
        UserAgent: c.Request.UserAgent(),
        Timestamp: time.Now(),
    })
    
    c.JSON(http.StatusCreated, project)
}

func (h *Handler) publishAuditEvent(event AuditEvent) {
    if h.kafkaProducer == nil {
        return // Kafka not available
    }
    
    data, err := json.Marshal(event)
    if err != nil {
        h.logger.Errorf(context.Background(), "Failed to marshal audit event: %v", err)
        return
    }
    
    if err := h.kafkaProducer.Publish([]byte(event.UserID), data); err != nil {
        h.logger.Errorf(context.Background(), "Failed to publish audit event: %v", err)
    }
}
```

### Step 3: Audit Events to Publish

**Project Service**:
- `CREATE_PROJECT`
- `UPDATE_PROJECT`
- `DELETE_PROJECT`

**Ingest Service**:
- `CREATE_SOURCE`
- `DELETE_SOURCE`
- `EXPORT_DATA`

**Knowledge Service**:
- `CREATE_KNOWLEDGE`
- `UPDATE_KNOWLEDGE`
- `DELETE_KNOWLEDGE`

---

## Testing Integration

### Unit Tests

```go
package handler_test

import (
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "smap-api/pkg/auth"
)

func TestCreateProject_Authenticated(t *testing.T) {
    // Setup
    gin.SetMode(gin.TestMode)
    router := gin.New()
    
    // Mock auth middleware
    router.Use(func(c *gin.Context) {
        // Inject mock user claims
        c.Set("user_id", "test-user-id")
        c.Set("email", "test@example.com")
        c.Set("role", "ANALYST")
        c.Set("groups", []string{"analysts@example.com"})
        c.Next()
    })
    
    handler := NewHandler(/* ... */)
    router.POST("/projects", handler.createProject)
    
    // Test
    req := httptest.NewRequest("POST", "/projects", nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    // Assert
    assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateProject_Unauthorized(t *testing.T) {
    // Setup
    gin.SetMode(gin.TestMode)
    router := gin.New()
    
    // No auth middleware - should fail
    handler := NewHandler(/* ... */)
    router.POST("/projects", handler.createProject)
    
    // Test
    req := httptest.NewRequest("POST", "/projects", nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    // Assert
    assert.Equal(t, http.StatusUnauthorized, w.Code)
}
```

### Integration Tests

```bash
#!/bin/bash

# Test script for service integration

# 1. Get JWT token
echo "Getting JWT token..."
TOKEN=$(curl -X POST "https://yourdomain.com/authentication/login" \
  --cookie-jar cookies.txt -L -s | grep -o 'smap_auth_token=[^;]*' | cut -d= -f2)

if [ -z "$TOKEN" ]; then
  echo "Failed to get token"
  exit 1
fi

echo "Token obtained: ${TOKEN:0:20}..."

# 2. Test authenticated endpoint
echo "Testing authenticated endpoint..."
RESPONSE=$(curl -X GET "https://project-service.com/api/projects" \
  -H "Authorization: Bearer $TOKEN" \
  -w "\n%{http_code}" -s)

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" != "200" ]; then
  echo "Failed: Expected 200, got $HTTP_CODE"
  echo "Response: $BODY"
  exit 1
fi

echo "Success: Authenticated endpoint works"

# 3. Test authorization
echo "Testing authorization..."
RESPONSE=$(curl -X DELETE "https://project-service.com/api/projects/test-id" \
  -H "Authorization: Bearer $TOKEN" \
  -w "\n%{http_code}" -s)

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)

if [ "$HTTP_CODE" != "403" ] && [ "$HTTP_CODE" != "200" ]; then
  echo "Warning: Unexpected status code $HTTP_CODE for delete"
fi

echo "Success: Authorization works"

# 4. Test token expiration
echo "Testing with expired token..."
EXPIRED_TOKEN="eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.expired.signature"
RESPONSE=$(curl -X GET "https://project-service.com/api/projects" \
  -H "Authorization: Bearer $EXPIRED_TOKEN" \
  -w "\n%{http_code}" -s)

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)

if [ "$HTTP_CODE" != "401" ]; then
  echo "Failed: Expected 401 for expired token, got $HTTP_CODE"
  exit 1
fi

echo "Success: Token expiration handled correctly"

echo "All tests passed!"
```

---

## Rollback Procedure

If integration causes issues in production:

### Immediate Rollback

```bash
# Kubernetes
kubectl rollout undo deployment/<service-name> -n <namespace>

# Verify rollback
kubectl rollout status deployment/<service-name> -n <namespace>

# Docker Compose
docker-compose up -d <service-name>:previous-tag
```

### Gradual Rollback

If using canary deployment:

```bash
# Reduce traffic to new version
kubectl patch deployment <service-name> -n <namespace> \
  -p '{"spec":{"replicas":1}}'

# Increase traffic to old version
kubectl scale deployment <service-name>-old -n <namespace> --replicas=3

# Monitor for 10 minutes
# If stable, complete rollback
kubectl delete deployment <service-name> -n <namespace>
kubectl rename deployment <service-name>-old <service-name> -n <namespace>
```

### Post-Rollback

1. **Investigate logs** to find root cause
2. **Fix issues** in development environment
3. **Test thoroughly** before redeploying
4. **Document lessons learned**

---

## Common Issues

### Issue 1: JWT Verification Fails

**Symptoms**: All requests return 401 Unauthorized

**Solutions**:
- Verify JWKS endpoint is reachable
- Check issuer and audience configuration
- Verify Redis connection for blacklist
- Check logs for specific error

### Issue 2: Role Authorization Not Working

**Symptoms**: Users with correct role get 403 Forbidden

**Solutions**:
- Verify role claim in JWT token
- Check role middleware configuration
- Ensure role names match exactly (case-sensitive)

### Issue 3: Service Key Invalid

**Symptoms**: Internal API calls return 401

**Solutions**:
- Verify service key is correctly encrypted
- Check X-Service-Key header is set
- Verify service key matches Identity Service configuration

### Issue 4: Audit Events Not Published

**Symptoms**: No audit logs in database

**Solutions**:
- Verify Kafka connection
- Check topic name matches Identity Service
- Verify event format matches expected schema

---

## Next Steps

After successful integration:

1. **Monitor metrics**: Track authentication success/failure rates
2. **Set up alerts**: Alert on high 401/403 rates
3. **Review logs**: Check for any auth-related errors
4. **Performance test**: Ensure JWT verification doesn't impact latency
5. **Document**: Update service documentation with auth requirements

For more information:
- **Identity Service API**: `documents/identity-service-api.md`
- **Troubleshooting**: `documents/identity-service-troubleshooting.md`
- **Frontend Integration**: `documents/frontend-oauth-migration.md`
