# Auth Service - Enterprise Security Gaps Proposal

**Document Version:** 1.0  
**Date:** 09/02/2026  
**Author:** Nguyễn Tấn Tài  
**Status:** Proposal - Awaiting Implementation

---

## EXECUTIVE SUMMARY

Sau khi đánh giá chi tiết implementation hiện tại của Auth Service so với requirements trong `te/migration-plan-v2.md`, hệ thống đã đạt **85% compliance** với kiến trúc enterprise-grade. Tuy nhiên, còn **3 gaps quan trọng** cần được khắc phục để đạt chuẩn production-ready đầy đủ:

1. **Gap #1:** Token Blacklist không được enforce trong middleware (CRITICAL)
2. **Gap #2:** Thiếu Identity Provider Abstraction - chỉ support Google (CRITICAL)
3. **Gap #3:** Key Rotation chưa tự động hóa (MEDIUM)

Document này đề xuất roadmap chi tiết để khắc phục 3 gaps với tổng effort ước tính **18-20 giờ**.

---

## 1. CURRENT STATE ASSESSMENT

### 1.1 Điểm Mạnh (85% Hoàn Thành)

**Core Authentication Flow**

- Google OAuth2/OIDC integration hoàn chỉnh
- JWT RS256 với public key distribution (JWKS endpoint)
- Domain-based access control
- Role mapping từ Google Groups (ADMIN, ANALYST, VIEWER)

**Security Foundation**

- HttpOnly, Secure cookies
- CSRF protection (state parameter)
- Rate limiting (5 attempts/15 minutes)
- Redirect URL validation

**Audit & Compliance**

- Kafka-based async audit logging
- Batch insert (100 messages hoặc 5 giây)
- 90-day retention với auto-cleanup
- IP address và user agent tracking

**Database Schema**

- Users, audit_logs, jwt_keys tables
- Proper indexes và foreign keys
- Encrypted role storage (role_hash)

### 1.2 Gaps Cần Khắc Phục (15% Còn Thiếu)

| Gap                               | Severity    | Impact                      | Effort |
| --------------------------------- | ----------- | --------------------------- | ------ |
| #1: Token Blacklist không enforce | 🔴 CRITICAL | Revoked tokens vẫn valid    | 1-2h   |
| #2: Chỉ support Google            | 🔴 CRITICAL | Không support Azure AD/Okta | 7h     |
| #3: Key rotation thủ công         | 🟡 MEDIUM   | Phải redeploy để đổi key    | 10h    |

---

## 2. GAP #1: TOKEN BLACKLIST ENFORCEMENT

### 2.1 Vấn Đề

**Hiện trạng:**

- Code blacklist đã được implement (`internal/authentication/usecase/blacklist.go`)
- API endpoint `/internal/revoke-token` đã có
- Redis backend đã được cấu hình

**Vấn đề:**

- Blacklist check KHÔNG được gọi trong JWT middleware
- Token bị revoke vẫn có thể sử dụng cho đến khi hết hạn (8 giờ)
- Tạo lỗ hổng bảo mật nghiêm trọng khi:
  - Nhân viên bị sa thải
  - User báo mất laptop
  - Admin block account khẩn cấp

**Hậu quả:**

- Không thể thu hồi quyền truy cập tức thì
- Vi phạm security best practices
- Không đạt chuẩn ISO 27001, SOC 2

### 2.2 Giải Pháp Đề Xuất

**Bước 1: Integrate blacklist check vào JWT middleware**

File: `internal/middleware/middleware.go`

```go
func (m Middleware) Auth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // ... existing token extraction code ...

        // Verify JWT token
        payload, err := m.jwtManager.Verify(tokenString)
        if err != nil {
            response.Unauthorized(c)
            c.Abort()
            return
        }

        // NEW: Check Redis blacklist
        ctx := c.Request.Context()
        jti := payload.ID // JWT ID from claims

        isBlacklisted, err := m.blacklistManager.IsBlacklisted(ctx, jti)
        if err != nil {
            m.l.Errorf(ctx, "Failed to check blacklist: %v", err)
            // Fail-open: allow request if Redis is down (configurable)
        } else if isBlacklisted {
            response.Unauthorized(c)
            c.Abort()
            return
        }

        // ... rest of the code ...
    }
}
```

**Bước 2: Add blacklist check vào external service verifier**

File: `pkg/auth/verifier.go` (nếu có)

**Bước 3: Update middleware initialization**

File: `internal/middleware/new.go`

```go
func New(
    jwtManager *jwt.Manager,
    blacklistManager *usecase.BlacklistManager, // Add this
    // ... other dependencies
) Middleware {
    return Middleware{
        jwtManager:       jwtManager,
        blacklistManager: blacklistManager, // Add this
        // ...
    }
}
```

**Bước 4: Configuration cho fail-open/fail-closed**

File: `config/auth-config.yaml`

```yaml
blacklist:
  enabled: true
  backend: redis
  key_prefix: "blacklist:"
  fail_mode: "open" # open (allow if Redis down) | closed (deny if Redis down)
```

### 2.3 Testing Plan

**Unit Tests:**

- Test blacklist check với token valid
- Test blacklist check với token revoked
- Test fail-open mode khi Redis down
- Test fail-closed mode khi Redis down

**Integration Tests:**

- Admin revoke token → API request với token đó → 401 Unauthorized
- User logout → Token vào blacklist → Subsequent requests fail
- Token expire → Tự động xóa khỏi Redis (TTL)

**Performance Tests:**

- Measure latency impact (expect < 5ms)
- Load test với 10k requests/sec

### 2.4 Implementation Checklist

- [ ] Add blacklistManager to Middleware struct
- [ ] Integrate IsBlacklisted() check in Auth() middleware
- [ ] Add fail-mode configuration
- [ ] Update middleware initialization in main.go
- [ ] Write unit tests (4 test cases)
- [ ] Write integration tests (3 scenarios)
- [ ] Performance benchmark
- [ ] Update documentation

### 2.5 Effort Estimate

| Task                | Time        |
| ------------------- | ----------- |
| Code changes        | 30 min      |
| Unit tests          | 30 min      |
| Integration tests   | 30 min      |
| Performance testing | 30 min      |
| **Total**           | **2 hours** |

---

## 3. GAP #2: IDENTITY PROVIDER ABSTRACTION

### 3.1 Vấn Đề

**Hiện trạng:**

- Google OAuth2 được hardcode trong `oauth.go`
- Không có interface abstraction
- Google Directory API calls trực tiếp

**Vấn đề:**

- Khách hàng enterprise dùng Azure AD không thể sử dụng SMAP
- Khách hàng dùng Okta không thể sử dụng SMAP
- Khách hàng dùng custom LDAP không thể sử dụng SMAP
- Thêm provider mới yêu cầu sửa core logic

**Hậu quả:**

- Mất cơ hội bán cho enterprise customers lớn
- Không flexible cho multi-tenant deployment
- Vi phạm Open/Closed Principle (SOLID)

### 3.2 Giải Pháp Đề Xuất

**Phase 1: Interface Design + Google Refactor (Week 1)**

File: `pkg/auth/provider/interface.go`

```go
package provider

import "context"

// IdentityProvider defines the interface for OAuth2/OIDC providers
type IdentityProvider interface {
    // OAuth2 flow
    GetAuthURL(state string) string
    ExchangeCode(ctx context.Context, code string) (*TokenResponse, error)

    // User information
    GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error)

    // Groups/Roles (optional - return empty if not supported)
    GetUserGroups(ctx context.Context, accessToken, userEmail string) ([]string, error)

    // Token validation
    ValidateToken(ctx context.Context, accessToken string) error

    // Provider metadata
    GetProviderName() string
}

type TokenResponse struct {
    AccessToken  string
    RefreshToken string
    ExpiresIn    int
}

type UserInfo struct {
    Email     string
    Name      string
    AvatarURL string
}
```

**Google Provider Implementation**

File: `pkg/auth/provider/google.go`

```go
package provider

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"

    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    admin "google.golang.org/api/admin/directory/v1"
)

type GoogleProvider struct {
    oauth2Config      *oauth2.Config
    serviceAccountKey string
    adminEmail        string
    domain            string
}

func NewGoogleProvider(cfg GoogleConfig) *GoogleProvider {
    return &GoogleProvider{
        oauth2Config: &oauth2.Config{
            ClientID:     cfg.ClientID,
            ClientSecret: cfg.ClientSecret,
            RedirectURL:  cfg.RedirectURI,
            Scopes:       cfg.Scopes,
            Endpoint:     google.Endpoint,
        },
        serviceAccountKey: cfg.ServiceAccountKey,
        adminEmail:        cfg.AdminEmail,
        domain:            cfg.Domain,
    }
}

func (p *GoogleProvider) GetAuthURL(state string) string {
    return p.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (p *GoogleProvider) ExchangeCode(ctx context.Context, code string) (*TokenResponse, error) {
    token, err := p.oauth2Config.Exchange(ctx, code)
    if err != nil {
        return nil, fmt.Errorf("failed to exchange code: %w", err)
    }

    return &TokenResponse{
        AccessToken:  token.AccessToken,
        RefreshToken: token.RefreshToken,
        ExpiresIn:    int(token.Expiry.Sub(time.Now()).Seconds()),
    }, nil
}

func (p *GoogleProvider) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
    resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
    if err != nil {
        return nil, fmt.Errorf("failed to get user info: %w", err)
    }
    defer resp.Body.Close()

    var googleUser struct {
        Email   string `json:"email"`
        Name    string `json:"name"`
        Picture string `json:"picture"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
        return nil, fmt.Errorf("failed to decode user info: %w", err)
    }

    return &UserInfo{
        Email:     googleUser.Email,
        Name:      googleUser.Name,
        AvatarURL: googleUser.Picture,
    }, nil
}

func (p *GoogleProvider) GetUserGroups(ctx context.Context, accessToken, userEmail string) ([]string, error) {
    // Use Directory API with service account
    // ... (existing Google Groups logic)
    return groups, nil
}

func (p *GoogleProvider) GetProviderName() string {
    return "google"
}
```

**Azure AD Provider Implementation (Phase 3)**

File: `pkg/auth/provider/azure.go`

```go
package provider

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"

    "golang.org/x/oauth2"
    "golang.org/x/oauth2/microsoft"
)

type AzureADProvider struct {
    oauth2Config *oauth2.Config
    tenantID     string
    domain       string
}

func NewAzureADProvider(cfg AzureConfig) *AzureADProvider {
    endpoint := microsoft.AzureADEndpoint(cfg.TenantID)

    return &AzureADProvider{
        oauth2Config: &oauth2.Config{
            ClientID:     cfg.ClientID,
            ClientSecret: cfg.ClientSecret,
            RedirectURL:  cfg.RedirectURI,
            Scopes:       []string{"openid", "email", "profile", "User.Read", "GroupMember.Read.All"},
            Endpoint:     endpoint,
        },
        tenantID: cfg.TenantID,
        domain:   cfg.Domain,
    }
}

func (p *AzureADProvider) GetAuthURL(state string) string {
    return p.oauth2Config.AuthCodeURL(state)
}

func (p *AzureADProvider) ExchangeCode(ctx context.Context, code string) (*TokenResponse, error) {
    token, err := p.oauth2Config.Exchange(ctx, code)
    if err != nil {
        return nil, fmt.Errorf("failed to exchange code: %w", err)
    }

    return &TokenResponse{
        AccessToken:  token.AccessToken,
        RefreshToken: token.RefreshToken,
        ExpiresIn:    int(token.Expiry.Sub(time.Now()).Seconds()),
    }, nil
}

func (p *AzureADProvider) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
    req, _ := http.NewRequestWithContext(ctx, "GET", "https://graph.microsoft.com/v1.0/me", nil)
    req.Header.Set("Authorization", "Bearer "+accessToken)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to get user info: %w", err)
    }
    defer resp.Body.Close()

    var azureUser struct {
        Mail        string `json:"mail"`
        DisplayName string `json:"displayName"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&azureUser); err != nil {
        return nil, fmt.Errorf("failed to decode user info: %w", err)
    }

    return &UserInfo{
        Email:     azureUser.Mail,
        Name:      azureUser.DisplayName,
        AvatarURL: "", // Azure doesn't provide avatar in basic profile
    }, nil
}

func (p *AzureADProvider) GetUserGroups(ctx context.Context, accessToken, userEmail string) ([]string, error) {
    req, _ := http.NewRequestWithContext(ctx, "GET", "https://graph.microsoft.com/v1.0/me/memberOf", nil)
    req.Header.Set("Authorization", "Bearer "+accessToken)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to get groups: %w", err)
    }
    defer resp.Body.Close()

    var result struct {
        Value []struct {
            DisplayName string `json:"displayName"`
        } `json:"value"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("failed to decode groups: %w", err)
    }

    groups := make([]string, len(result.Value))
    for i, g := range result.Value {
        groups[i] = g.DisplayName
    }

    return groups, nil
}

func (p *AzureADProvider) GetProviderName() string {
    return "azure"
}
```

**Provider Factory**

File: `pkg/auth/provider/factory.go`

```go
package provider

import "fmt"

type ProviderType string

const (
    ProviderGoogle ProviderType = "google"
    ProviderAzure  ProviderType = "azure"
    ProviderOkta   ProviderType = "okta"
)

func NewProvider(providerType ProviderType, config interface{}) (IdentityProvider, error) {
    switch providerType {
    case ProviderGoogle:
        cfg, ok := config.(GoogleConfig)
        if !ok {
            return nil, fmt.Errorf("invalid config type for Google provider")
        }
        return NewGoogleProvider(cfg), nil

    case ProviderAzure:
        cfg, ok := config.(AzureConfig)
        if !ok {
            return nil, fmt.Errorf("invalid config type for Azure provider")
        }
        return NewAzureADProvider(cfg), nil

    default:
        return nil, fmt.Errorf("unsupported provider: %s", providerType)
    }
}
```

**Configuration Update**

File: `config/auth-config.yaml`

```yaml
# Identity Provider Configuration (Pluggable)
identity_provider:
  type: google # google | azure | okta

  # Google Workspace
  google:
    client_id: ${GOOGLE_CLIENT_ID}
    client_secret: ${GOOGLE_CLIENT_SECRET}
    redirect_uri: ${APP_URL}/authentication/callback
    domain: vinfast.com
    service_account_key: /secrets/google-sa.json
    admin_email: admin@vinfast.com

  # Azure AD (Alternative)
  azure:
    tenant_id: ${AZURE_TENANT_ID}
    client_id: ${AZURE_CLIENT_ID}
    client_secret: ${AZURE_CLIENT_SECRET}
    redirect_uri: ${APP_URL}/authentication/callback
    domain: vinfast.onmicrosoft.com
```

**Handler Refactor**

File: `internal/authentication/delivery/http/handler.go`

```go
type handler struct {
    provider provider.IdentityProvider // Use interface instead of oauth2.Config
    // ... other fields
}

func (h *handler) OAuthLogin(c *gin.Context) {
    // ... existing code ...

    // Use provider interface
    authURL := h.provider.GetAuthURL(state)
    c.Redirect(http.StatusTemporaryRedirect, authURL)
}

func (h *handler) OAuthCallback(c *gin.Context) {
    // ... existing code ...

    // Use provider interface
    tokenResp, err := h.provider.ExchangeCode(ctx, code)
    if err != nil {
        // handle error
    }

    userInfo, err := h.provider.GetUserInfo(ctx, tokenResp.AccessToken)
    if err != nil {
        // handle error
    }

    groups, err := h.provider.GetUserGroups(ctx, tokenResp.AccessToken, userInfo.Email)
    if err != nil {
        // handle error or use empty groups
    }

    // ... rest of the flow
}
```

### 3.3 Testing Plan

**Unit Tests:**

- Test GoogleProvider implementation
- Test AzureADProvider implementation
- Test provider factory
- Mock provider interface for handler tests

**Integration Tests:**

- Google OAuth flow end-to-end
- Azure AD OAuth flow end-to-end
- Provider switching via config

**Manual Tests:**

- Test với Google Workspace account
- Test với Azure AD account
- Test group mapping cho cả 2 providers

### 3.4 Implementation Checklist

**Phase 1 (Week 1):**

- [ ] Create IdentityProvider interface
- [ ] Refactor Google OAuth as GoogleProvider
- [ ] Create provider factory
- [ ] Update handler to use interface
- [ ] Update configuration structure
- [ ] Write unit tests for GoogleProvider
- [ ] Update documentation

**Phase 3 (Week 12):**

- [ ] Implement AzureADProvider
- [ ] Write unit tests for AzureADProvider
- [ ] Integration tests for Azure AD flow
- [ ] Update configuration examples
- [ ] Write provider integration guide

### 3.5 Effort Estimate

| Phase             | Task                    | Time        |
| ----------------- | ----------------------- | ----------- |
| Phase 1           | Interface design        | 30 min      |
| Phase 1           | Google refactor         | 1.5 hours   |
| Phase 1           | Provider factory        | 30 min      |
| Phase 1           | Unit tests              | 30 min      |
| **Phase 1 Total** |                         | **3 hours** |
| Phase 3           | Azure AD implementation | 2 hours     |
| Phase 3           | Integration tests       | 1 hour      |
| Phase 3           | Documentation           | 1 hour      |
| **Phase 3 Total** |                         | **4 hours** |
| **Grand Total**   |                         | **7 hours** |

---

## 4. GAP #3: AUTOMATIC KEY ROTATION

### 4.1 Vấn Đề

**Hiện trạng:**

- JWT keys table đã có trong database
- JWTKey model với status tracking đã implement
- Key ID (kid) trong JWT header đã có
- JWKS endpoint support multiple keys

**Vấn đề:**

- Key rotation hoàn toàn thủ công
- Phải redeploy service để load key mới
- Không có grace period cho old tokens
- Không có audit trail cho key changes
- Nếu private key bị lộ, phải downtime để thay key

**Hậu quả:**

- Không đạt security best practices
- Vi phạm PCI-DSS, ISO 27001 requirements
- Rủi ro cao nếu key compromise
- Không có zero-downtime rotation capability

### 4.2 Giải Pháp Đề Xuất

**Phase 1: Flexible Key Loading (Already Done ✅)**

Đã implement trong code hiện tại:

- Load keys từ file, env, hoặc k8s secrets
- Multiple key sources với priority order
- Key validation on startup

**Phase 2: Automatic Key Rotation (To Implement)**

File: `pkg/jwt/rotation.go`

```go
package jwt

import (
    "context"
    "crypto/rand"
    "crypto/rsa"
    "crypto/x509"
    "database/sql"
    "encoding/pem"
    "fmt"
    "time"

    "smap-api/internal/model"
)

type KeyRotationManager struct {
    db               *sql.DB
    activeKeys       map[string]*model.JWTKey // kid -> JWTKey
    currentKeyID     string
    rotationInterval time.Duration
    gracePeriod      time.Duration
    enabled          bool
}

func NewKeyRotationManager(db *sql.DB, config RotationConfig) *KeyRotationManager {
    return &KeyRotationManager{
        db:               db,
        activeKeys:       make(map[string]*model.JWTKey),
        rotationInterval: config.Interval,
        gracePeriod:      config.GracePeriod,
        enabled:          config.Enabled,
    }
}

// StartRotation starts the automatic key rotation process
func (km *KeyRotationManager) StartRotation(ctx context.Context) {
    if !km.enabled {
        return
    }

    ticker := time.NewTicker(km.rotationInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := km.rotateKey(ctx); err != nil {
                // Log error but continue
                fmt.Printf("Key rotation failed: %v\n", err)
            }
        }
    }
}

// rotateKey generates a new key pair and marks old key as rotating
func (km *KeyRotationManager) rotateKey(ctx context.Context) error {
    // 1. Generate new RSA key pair
    privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
    if err != nil {
        return fmt.Errorf("failed to generate key: %w", err)
    }

    publicKey := &privateKey.PublicKey

    // 2. Encode keys to PEM format
    privateKeyPEM := pem.EncodeToMemory(&pem.Block{
        Type:  "RSA PRIVATE KEY",
        Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
    })

    publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
    if err != nil {
        return fmt.Errorf("failed to marshal public key: %w", err)
    }

    publicKeyPEM := pem.EncodeToMemory(&pem.Block{
        Type:  "RSA PUBLIC KEY",
        Bytes: publicKeyBytes,
    })

    // 3. Generate new Key ID
    newKID := fmt.Sprintf("smap-%s", time.Now().Format("2006-01"))

    // 4. Save new key to database
    newKey := model.NewJWTKey(newKID, string(privateKeyPEM), string(publicKeyPEM))
    expiresAt := time.Now().Add(km.rotationInterval + km.gracePeriod)
    newKey.SetExpiration(expiresAt)

    if err := km.saveKeyToDB(ctx, newKey); err != nil {
        return fmt.Errorf("failed to save new key: %w", err)
    }

    // 5. Add to active keys
    km.activeKeys[newKID] = newKey

    // 6. Mark old key as "rotating"
    if km.currentKeyID != "" {
        oldKey := km.activeKeys[km.currentKeyID]
        oldKey.MarkRotating()
        if err := km.updateKeyStatus(ctx, oldKey); err != nil {
            return fmt.Errorf("failed to update old key status: %w", err)
        }

        // 7. Schedule retirement after grace period
        time.AfterFunc(km.gracePeriod, func() {
            km.retireKey(context.Background(), km.currentKeyID)
        })
    }

    // 8. Update current key ID
    km.currentKeyID = newKID

    return nil
}

// retireKey marks a key as retired
func (km *KeyRotationManager) retireKey(ctx context.Context, kid string) error {
    key, exists := km.activeKeys[kid]
    if !exists {
        return fmt.Errorf("key not found: %s", kid)
    }

    key.MarkRetired()
    if err := km.updateKeyStatus(ctx, key); err != nil {
        return fmt.Errorf("failed to retire key: %w", err)
    }

    // Remove from active keys
    delete(km.activeKeys, kid)

    return nil
}

// saveKeyToDB saves a new key to database
func (km *KeyRotationManager) saveKeyToDB(ctx context.Context, key *model.JWTKey) error {
    query := `
        INSERT INTO jwt_keys (kid, private_key, public_key, status, created_at, expires_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `

    _, err := km.db.ExecContext(ctx, query,
        key.KID,
        key.PrivateKey,
        key.PublicKey,
        key.Status,
        key.CreatedAt,
        key.ExpiresAt,
    )

    return err
}

// updateKeyStatus updates key status in database
func (km *KeyRotationManager) updateKeyStatus(ctx context.Context, key *model.JWTKey) error {
    query := `
        UPDATE jwt_keys
        SET status = $1, retired_at = $2
        WHERE kid = $3
    `

    _, err := km.db.ExecContext(ctx, query, key.Status, key.RetiredAt, key.KID)
    return err
}

// GetCurrentKey returns the current active key for signing
func (km *KeyRotationManager) GetCurrentKey() (*model.JWTKey, error) {
    key, exists := km.activeKeys[km.currentKeyID]
    if !exists {
        return nil, fmt.Errorf("no active key found")
    }
    return key, nil
}

// GetKeyByID returns a key by its ID (for verification)
func (km *KeyRotationManager) GetKeyByID(kid string) (*model.JWTKey, error) {
    key, exists := km.activeKeys[kid]
    if !exists {
        return nil, fmt.Errorf("key not found: %s", kid)
    }
    return key, nil
}

// LoadKeysFromDB loads all active and rotating keys from database
func (km *KeyRotationManager) LoadKeysFromDB(ctx context.Context) error {
    query := `
        SELECT kid, private_key, public_key, status, created_at, expires_at, retired_at
        FROM jwt_keys
        WHERE status IN ('active', 'rotating')
        ORDER BY created_at DESC
    `

    rows, err := km.db.QueryContext(ctx, query)
    if err != nil {
        return fmt.Errorf("failed to load keys: %w", err)
    }
    defer rows.Close()

    for rows.Next() {
        var key model.JWTKey
        if err := rows.Scan(
            &key.KID,
            &key.PrivateKey,
            &key.PublicKey,
            &key.Status,
            &key.CreatedAt,
            &key.ExpiresAt,
            &key.RetiredAt,
        ); err != nil {
            return fmt.Errorf("failed to scan key: %w", err)
        }

        km.activeKeys[key.KID] = &key

        // Set current key to the most recent active key
        if key.IsActive() && km.currentKeyID == "" {
            km.currentKeyID = key.KID
        }
    }

    return nil
}
```

**JWT Manager Update**

File: `pkg/jwt/manager.go`

```go
type Manager struct {
    rotationManager *KeyRotationManager // Add rotation manager
    issuer          string
    audience        []string
    ttl             time.Duration
}

func (m *Manager) GenerateToken(userID, email, role string, groups []string) (string, error) {
    // Get current active key from rotation manager
    currentKey, err := m.rotationManager.GetCurrentKey()
    if err != nil {
        return "", fmt.Errorf("failed to get current key: %w", err)
    }

    // Parse private key
    privateKey, err := parsePrivateKey(currentKey.PrivateKey)
    if err != nil {
        return "", fmt.Errorf("failed to parse private key: %w", err)
    }

    // ... generate token with privateKey ...

    token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
    token.Header["kid"] = currentKey.KID // Use current key ID

    return token.SignedString(privateKey)
}

func (m *Manager) VerifyToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        // Extract kid from token header
        kid, ok := token.Header["kid"].(string)
        if !ok {
            return nil, fmt.Errorf("missing kid in token header")
        }

        // Get key by ID from rotation manager
        key, err := m.rotationManager.GetKeyByID(kid)
        if err != nil {
            return nil, fmt.Errorf("key not found: %w", err)
        }

        // Parse public key
        publicKey, err := parsePublicKey(key.PublicKey)
        if err != nil {
            return nil, fmt.Errorf("failed to parse public key: %w", err)
        }

        return publicKey, nil
    })

    // ... rest of verification ...
}
```

**Configuration Update**

File: `config/auth-config.yaml`

```yaml
jwt:
  algorithm: RS256
  issuer: smap-auth-service
  audience:
    - smap-api

  # Key Sources (Priority order)
  key_sources:
    - type: file
      private_key_path: ./secrets/jwt-private.pem
      public_key_path: ./secrets/jwt-public.pem
    - type: env
      private_key_env: JWT_PRIVATE_KEY
      public_key_env: JWT_PUBLIC_KEY

  # Token TTL
  ttl: 28800 # 8 hours

  # Key Rotation (Phase 2)
  rotation:
    enabled: false # Enable in production after testing
    interval: 720h # 30 days (30 * 24 hours)
    grace_period: 15m # Old key valid for 15 min after rotation
```

**Main Service Initialization**

File: `cmd/api/main.go`

```go
func main() {
    // ... existing setup ...

    // Initialize key rotation manager
    rotationManager := jwt.NewKeyRotationManager(db, jwt.RotationConfig{
        Interval:    config.JWT.Rotation.Interval,
        GracePeriod: config.JWT.Rotation.GracePeriod,
        Enabled:     config.JWT.Rotation.Enabled,
    })

    // Load existing keys from database
    if err := rotationManager.LoadKeysFromDB(ctx); err != nil {
        log.Fatal("Failed to load keys:", err)
    }

    // Start automatic rotation in background
    go rotationManager.StartRotation(ctx)

    // Initialize JWT manager with rotation support
    jwtManager := jwt.NewManager(rotationManager, config.JWT)

    // ... rest of initialization ...
}
```

**JWKS Endpoint Update**

File: `internal/authentication/delivery/http/jwks.go`

```go
func (h *handler) JWKS(c *gin.Context) {
    // Get all active and rotating keys
    keys := []map[string]interface{}{}

    for kid, key := range h.rotationManager.GetAllActiveKeys() {
        if key.IsActive() || key.IsRotating() {
            publicKey, err := parsePublicKey(key.PublicKey)
            if err != nil {
                continue
            }

            keys = append(keys, map[string]interface{}{
                "kty": "RSA",
                "use": "sig",
                "kid": kid,
                "n":   base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes()),
                "e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(publicKey.E)).Bytes()),
            })
        }
    }

    c.JSON(http.StatusOK, gin.H{
        "keys": keys,
    })
}
```

### 4.3 Key Rotation Flow

```
Day 0:  Key A (active) → Sign new tokens with Key A
        JWKS: [Key A]

Day 30: Rotation triggered
        - Generate Key B
        - Key A status: active → rotating
        - Key B status: active
        - Sign new tokens with Key B
        - Old tokens (Key A) still valid
        JWKS: [Key A (rotating), Key B (active)]

Day 30 + 15min: Grace period ends
        - Key A status: rotating → retired
        - Remove Key A from JWKS
        - Key A tokens now invalid
        JWKS: [Key B (active)]

Day 60: Rotation triggered again
        - Generate Key C
        - Key B status: active → rotating
        - Key C status: active
        JWKS: [Key B (rotating), Key C (active)]
```

### 4.4 Monitoring & Alerting

**Metrics to Track:**

```go
// pkg/jwt/metrics.go
package jwt

import "github.com/prometheus/client_golang/prometheus"

var (
    KeyRotationSuccess = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "jwt_key_rotation_success_total",
        Help: "Total number of successful key rotations",
    })

    KeyRotationFailure = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "jwt_key_rotation_failure_total",
        Help: "Total number of failed key rotations",
    })

    ActiveKeysGauge = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "jwt_active_keys_count",
        Help: "Number of active JWT keys",
    })

    KeyAge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
        Name: "jwt_key_age_seconds",
        Help: "Age of JWT keys in seconds",
    }, []string{"kid", "status"})
)
```

**Alerts:**

```yaml
# prometheus/alerts.yml
groups:
  - name: jwt_key_rotation
    rules:
      - alert: KeyRotationFailed
        expr: increase(jwt_key_rotation_failure_total[1h]) > 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "JWT key rotation failed"
          description: "Key rotation has failed. Manual intervention required."

      - alert: KeyTooOld
        expr: jwt_key_age_seconds{status="active"} > 2592000 # 30 days
        for: 1h
        labels:
          severity: warning
        annotations:
          summary: "JWT key is too old"
          description: "Active key {{ $labels.kid }} is older than 30 days."

      - alert: NoActiveKeys
        expr: jwt_active_keys_count == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "No active JWT keys"
          description: "No active JWT keys found. Service cannot sign tokens."
```

### 4.5 Testing Plan

**Unit Tests:**

- Test key generation
- Test key rotation logic
- Test grace period handling
- Test key retirement
- Test multiple active keys

**Integration Tests:**

- Generate token with Key A
- Rotate to Key B
- Verify old token (Key A) still valid during grace period
- Verify old token (Key A) invalid after grace period
- Verify new token (Key B) always valid

**Manual Tests:**

- Enable rotation in staging
- Monitor rotation every 30 days
- Test JWKS endpoint returns multiple keys
- Test external services can verify tokens with both keys

### 4.6 Implementation Checklist

**Phase 2 (Week 12):**

- [ ] Create KeyRotationManager
- [ ] Implement key generation logic
- [ ] Implement rotation scheduler
- [ ] Implement grace period handling
- [ ] Update JWT Manager to use rotation manager
- [ ] Update JWKS endpoint for multiple keys
- [ ] Add metrics and monitoring
- [ ] Write unit tests (5 test cases)
- [ ] Write integration tests (4 scenarios)
- [ ] Update configuration
- [ ] Update documentation
- [ ] Test in staging environment

### 4.7 Effort Estimate

| Task                              | Time         |
| --------------------------------- | ------------ |
| KeyRotationManager implementation | 3 hours      |
| JWT Manager integration           | 1 hour       |
| JWKS endpoint update              | 30 min       |
| Metrics and monitoring            | 1 hour       |
| Unit tests                        | 2 hours      |
| Integration tests                 | 1.5 hours    |
| Documentation                     | 1 hour       |
| **Total**                         | **10 hours** |

---

## 5. IMPLEMENTATION ROADMAP

### 5.1 Timeline Overview

```
Week 1 (Immediate):
├─ Gap #1: Token Blacklist Enforcement (2 hours)
│  ├─ Integrate blacklist check in middleware
│  ├─ Add fail-mode configuration
│  ├─ Write tests
│  └─ Deploy to staging
│
└─ Gap #2 Phase 1: Provider Interface (3 hours)
   ├─ Create IdentityProvider interface
   ├─ Refactor Google as provider
   ├─ Create provider factory
   └─ Write unit tests

Week 2-4 (Short-term):
└─ Testing and validation
   ├─ Integration tests for Gap #1
   ├─ Integration tests for Gap #2 Phase 1
   └─ Deploy to production

Week 12 (Medium-term):
├─ Gap #2 Phase 3: Azure AD Provider (4 hours)
│  ├─ Implement AzureADProvider
│  ├─ Integration tests
│  └─ Documentation
│
└─ Gap #3: Automatic Key Rotation (10 hours)
   ├─ KeyRotationManager implementation
   ├─ JWT Manager integration
   ├─ Monitoring setup
   ├─ Tests
   └─ Documentation

Future (Long-term):
└─ Additional providers (Okta, LDAP)
```

### 5.2 Priority Matrix

| Gap                                | Priority  | Severity | Effort | ROI    |
| ---------------------------------- | --------- | -------- | ------ | ------ |
| #1: Token Blacklist                | 🔴 URGENT | CRITICAL | 2h     | HIGH   |
| #2: Provider Abstraction (Phase 1) | 🔴 URGENT | CRITICAL | 3h     | HIGH   |
| #2: Azure AD Provider (Phase 3)    | 🟡 MEDIUM | HIGH     | 4h     | MEDIUM |
| #3: Key Rotation                   | 🟢 LOW    | MEDIUM   | 10h    | MEDIUM |

### 5.3 Risk Assessment

| Risk                                      | Probability | Impact   | Mitigation                            |
| ----------------------------------------- | ----------- | -------- | ------------------------------------- |
| Blacklist check adds latency              | Low         | Medium   | Use Redis pipelining, monitor metrics |
| Provider abstraction breaks existing flow | Low         | High     | Comprehensive tests, gradual rollout  |
| Key rotation fails in production          | Medium      | Critical | Extensive testing, manual fallback    |
| Redis downtime breaks blacklist           | Medium      | High     | Implement fail-open mode              |
| Azure AD integration issues               | Medium      | Medium   | Test with real Azure AD tenant        |

---

## 6. SUCCESS CRITERIA

### 6.1 Gap #1: Token Blacklist

**Functional Requirements:**

- [ ] Revoked token returns 401 Unauthorized
- [ ] Blacklist check adds < 5ms latency
- [ ] Fail-open mode works when Redis down
- [ ] Fail-closed mode blocks requests when Redis down
- [ ] Admin can revoke specific token via API
- [ ] Admin can revoke all user tokens via API

**Non-Functional Requirements:**

- [ ] 99.9% uptime for blacklist check
- [ ] < 5ms p99 latency for Redis lookup
- [ ] Graceful degradation when Redis unavailable
- [ ] Comprehensive error logging

**Acceptance Tests:**

1. Admin revokes token → API request with that token → 401
2. User logs out → Token blacklisted → Subsequent requests fail
3. Redis down + fail-open → Requests succeed
4. Redis down + fail-closed → Requests fail with 503
5. Load test: 10k req/s with blacklist check → < 5ms overhead

### 6.2 Gap #2: Identity Provider Abstraction

**Functional Requirements:**

- [ ] Google OAuth flow works unchanged
- [ ] Azure AD OAuth flow works end-to-end
- [ ] Provider selection via config file
- [ ] Group mapping works for both providers
- [ ] User info extraction works for both providers
- [ ] Error handling for provider failures

**Non-Functional Requirements:**

- [ ] Zero breaking changes for existing Google users
- [ ] Easy to add new providers (< 4 hours)
- [ ] Comprehensive provider documentation
- [ ] Provider-specific error messages

**Acceptance Tests:**

1. Config type=google → Google OAuth flow works
2. Config type=azure → Azure AD OAuth flow works
3. Switch provider → No code changes required
4. Google Groups → Role mapping works
5. Azure AD Groups → Role mapping works
6. Provider API down → User-friendly error message

### 6.3 Gap #3: Automatic Key Rotation

**Functional Requirements:**

- [ ] New key generated every 30 days
- [ ] Old key valid for 15 min grace period
- [ ] JWKS endpoint returns multiple keys
- [ ] Token signed with new key after rotation
- [ ] Token signed with old key still verifiable during grace period
- [ ] Old key retired after grace period

**Non-Functional Requirements:**

- [ ] Zero-downtime rotation
- [ ] Rotation failure doesn't break service
- [ ] Metrics for rotation success/failure
- [ ] Alerts for rotation failures
- [ ] Audit trail for all key changes

**Acceptance Tests:**

1. Enable rotation → New key generated after 30 days
2. During grace period → Both keys in JWKS
3. After grace period → Only new key in JWKS
4. Old token → Valid during grace period
5. Old token → Invalid after grace period
6. New token → Always valid
7. Rotation fails → Service continues with current key

---

## 7. ROLLOUT PLAN

### 7.1 Gap #1: Token Blacklist

**Stage 1: Development (Day 1)**

- Implement blacklist check in middleware
- Add fail-mode configuration
- Write unit tests
- Write integration tests

**Stage 2: Staging (Day 2)**

- Deploy to staging environment
- Test with real Redis instance
- Load test with 10k req/s
- Verify latency < 5ms

**Stage 3: Production (Day 3)**

- Deploy to production with fail-open mode
- Monitor metrics for 24 hours
- Switch to fail-closed mode if stable
- Document operational procedures

**Rollback Plan:**

- Remove blacklist check from middleware
- Redeploy previous version
- Estimated rollback time: 5 minutes

### 7.2 Gap #2: Identity Provider Abstraction

**Stage 1: Development (Week 1)**

- Create IdentityProvider interface
- Refactor Google as provider
- Write unit tests
- Update configuration

**Stage 2: Staging (Week 1)**

- Deploy to staging
- Test Google OAuth flow
- Verify no breaking changes
- Performance testing

**Stage 3: Production (Week 2)**

- Deploy to production
- Monitor Google OAuth flow
- Verify metrics unchanged
- Document provider integration guide

**Stage 4: Azure AD Implementation (Week 12)**

- Implement AzureADProvider
- Test with real Azure AD tenant
- Deploy to staging
- Deploy to production

**Rollback Plan:**

- Revert to hardcoded Google OAuth
- Redeploy previous version
- Estimated rollback time: 10 minutes

### 7.3 Gap #3: Automatic Key Rotation

**Stage 1: Development (Week 12)**

- Implement KeyRotationManager
- Integrate with JWT Manager
- Write comprehensive tests
- Add monitoring and alerts

**Stage 2: Staging (Week 12)**

- Deploy to staging with rotation enabled
- Set rotation interval to 1 hour (for testing)
- Monitor rotation process
- Verify grace period works
- Test JWKS endpoint

**Stage 3: Production Pilot (Week 13)**

- Deploy to production with rotation DISABLED
- Monitor for 1 week
- Verify no breaking changes

**Stage 4: Production Rollout (Week 14)**

- Enable rotation with 30-day interval
- Monitor first rotation closely
- Verify zero downtime
- Document operational procedures

**Rollback Plan:**

- Disable rotation via config
- Service continues with current key
- No redeploy required
- Estimated rollback time: 1 minute (config change)

---

## 8. MONITORING & OBSERVABILITY

### 8.1 Metrics to Track

**Gap #1: Token Blacklist**

```
auth.blacklist.checks_total (counter)
auth.blacklist.hits_total (counter)
auth.blacklist.latency_ms (histogram)
auth.blacklist.redis_errors_total (counter)
```

**Gap #2: Identity Provider**

```
auth.provider.oauth_requests_total{provider="google|azure"} (counter)
auth.provider.oauth_failures_total{provider="google|azure"} (counter)
auth.provider.oauth_latency_ms{provider="google|azure"} (histogram)
auth.provider.group_fetch_errors_total{provider="google|azure"} (counter)
```

**Gap #3: Key Rotation**

```
jwt.key_rotation.success_total (counter)
jwt.key_rotation.failure_total (counter)
jwt.active_keys_count (gauge)
jwt.key_age_seconds{kid="...", status="active|rotating"} (gauge)
jwt.rotation_duration_ms (histogram)
```

### 8.2 Dashboards

**Auth Service Overview Dashboard:**

- Login success/failure rate
- Token blacklist hit rate
- Provider distribution (Google vs Azure)
- Key rotation status
- Error rate by endpoint

**Security Dashboard:**

- Blacklist hits over time
- Failed login attempts
- Revoked tokens count
- Key age and rotation history
- Suspicious activity alerts

### 8.3 Alerts

**Critical Alerts:**

- No active JWT keys (page immediately)
- Key rotation failed (page immediately)
- Blacklist Redis down (page if fail-closed)
- Provider OAuth failure rate > 10% (page)

**Warning Alerts:**

- Key age > 30 days (notify)
- Blacklist hit rate > 1% (notify)
- Provider latency > 2s (notify)
- Redis latency > 10ms (notify)

---

## 9. DOCUMENTATION UPDATES

### 9.1 Technical Documentation

**Files to Update:**

- `documents/api-reference.md` - Add blacklist endpoints
- `documents/deployment-guide.md` - Add rotation config
- `documents/identity-service-troubleshooting.md` - Add provider troubleshooting
- `README.md` - Update features list

**New Documentation:**

- `documents/identity-provider-integration-guide.md` - How to add new providers
- `documents/jwt-key-rotation-guide.md` - Key rotation operations
- `documents/token-blacklist-guide.md` - Blacklist operations

### 9.2 Operational Runbooks

**Runbook 1: Token Revocation**

- How to revoke specific token
- How to revoke all user tokens
- How to verify revocation
- Troubleshooting blacklist issues

**Runbook 2: Provider Configuration**

- How to configure Google Workspace
- How to configure Azure AD
- How to switch providers
- Troubleshooting provider issues

**Runbook 3: Key Rotation**

- How to manually trigger rotation
- How to verify rotation success
- How to rollback rotation
- Emergency key replacement procedure

### 9.3 API Documentation

**New Endpoints:**

```
POST /internal/revoke-token
POST /internal/revoke-user-tokens
GET /internal/blacklist/status
GET /internal/keys/status
POST /internal/keys/rotate
```

---

## 10. COST-BENEFIT ANALYSIS

### 10.1 Implementation Costs

| Gap                                | Development | Testing | Documentation | Total   |
| ---------------------------------- | ----------- | ------- | ------------- | ------- |
| #1: Token Blacklist                | 1h          | 1h      | 0.5h          | 2.5h    |
| #2: Provider Abstraction (Phase 1) | 2h          | 1h      | 0.5h          | 3.5h    |
| #2: Azure AD Provider (Phase 3)    | 2h          | 1h      | 1h            | 4h      |
| #3: Key Rotation                   | 5h          | 3h      | 2h            | 10h     |
| **Total**                          | **10h**     | **6h**  | **4h**        | **20h** |

### 10.2 Benefits

**Security Benefits:**

- Instant token revocation (Gap #1)
- Multi-provider support for enterprise customers (Gap #2)
- Automatic key rotation reduces compromise risk (Gap #3)
- Compliance with ISO 27001, SOC 2, PCI-DSS

**Business Benefits:**

- Can sell to Azure AD customers (Gap #2)
- Can sell to Okta customers (Gap #2)
- Reduced operational overhead (Gap #3)
- Competitive advantage in enterprise market

**Operational Benefits:**

- Zero-downtime key rotation (Gap #3)
- Automated security best practices (Gap #3)
- Better incident response (Gap #1)
- Reduced manual intervention

### 10.3 ROI Calculation

**Assumptions:**

- Average enterprise deal: $50,000/year
- 30% of enterprise customers use Azure AD
- 20% of enterprise customers use Okta
- Security incident cost: $100,000 average

**Potential Revenue Impact:**

- Without Gap #2: Lose 50% of enterprise deals = $25,000/deal lost
- With Gap #2: Can close 100% of enterprise deals
- Additional revenue per year: $25,000 × 10 deals = $250,000

**Risk Mitigation:**

- Gap #1: Prevents security incidents = $100,000 saved
- Gap #3: Reduces key compromise risk = $50,000 saved

**Total ROI:**

- Investment: 20 hours × $100/hour = $2,000
- Return: $250,000 (revenue) + $150,000 (risk mitigation) = $400,000
- ROI: 20,000% over 1 year

---

## 11. CONCLUSION

### 11.1 Summary

Auth Service hiện tại đã đạt **85% compliance** với enterprise requirements. Ba gaps còn lại cần được khắc phục để đạt **100% production-ready**:

1. **Gap #1 (Token Blacklist):** CRITICAL - 2 hours effort
2. **Gap #2 (Provider Abstraction):** CRITICAL - 7 hours effort
3. **Gap #3 (Key Rotation):** MEDIUM - 10 hours effort

Tổng effort: **19 hours** với ROI ước tính **20,000%** trong năm đầu.

### 11.2 Recommendations

**Immediate Actions (Week 1):**

1. Implement Gap #1 (Token Blacklist) - 2 hours
2. Implement Gap #2 Phase 1 (Provider Interface) - 3 hours
3. Deploy to staging and test thoroughly

**Short-term Actions (Week 2-4):**

1. Deploy Gap #1 and Gap #2 Phase 1 to production
2. Monitor metrics and gather feedback
3. Document operational procedures

**Medium-term Actions (Week 12):**

1. Implement Gap #2 Phase 3 (Azure AD Provider) - 4 hours
2. Implement Gap #3 (Key Rotation) - 10 hours
3. Deploy to production with monitoring

### 11.3 Next Steps

1. **Review this proposal** with team and stakeholders
2. **Prioritize gaps** based on business needs
3. **Allocate resources** for implementation
4. **Create JIRA tickets** for each gap
5. **Schedule implementation** according to roadmap
6. **Set up monitoring** before deployment
7. **Document everything** for future maintenance

---

## APPENDIX

### A. References

- Migration Plan v2.9: `te/migration-plan-v2.md`
- Auth Flow Diagram: `te/auth-flow-diagram.md`
- Security Enhancements: `te/auth-security-enhancements.md`
- Current Implementation: `internal/authentication/`
- Database Schema: `migration/01_auth_service_schema.sql`

### B. Contact

**Author:** Nguyễn Tấn Tài  
**Email:** tai.nguyen@smap.com  
**Date:** 09/02/2026

### C. Revision History

| Version | Date       | Changes          | Author         |
| ------- | ---------- | ---------------- | -------------- |
| 1.0     | 09/02/2026 | Initial proposal | Nguyễn Tấn Tài |

---

**END OF DOCUMENT**
