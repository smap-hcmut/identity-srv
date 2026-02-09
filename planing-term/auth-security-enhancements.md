# Auth Service - Enterprise Security Enhancements

## Tổng quan

Đây là 3 bổ sung quan trọng cho Auth Service để đạt chuẩn Enterprise Production-Ready:

1. **Token Blacklist** - Thu hồi quyền truy cập tức thì
2. **Identity Provider Abstraction** - Hỗ trợ đa nhà cung cấp SSO
3. **Key Rotation** - Quản lý và xoay vòng khóa bảo mật tự động

---

## 1. Token Blacklist (Instant Revocation)

### Vấn đề

JWT có TTL 15 phút. Khi Admin block user (nhân viên bị sa thải, mất laptop), token cũ vẫn valid trong 15 phút → Lỗ hổng bảo mật.

### Giải pháp

Thêm Redis blacklist check vào JWT middleware.

### Implementation Timeline

- **Phase 1 (Tuần 1):** Basic blacklist check
- **Phase 2 (Tuần 2):** Admin API để revoke tokens
- **Phase 3 (Tuần 12):** Monitoring & alerting

### Key Features

- Revoke toàn bộ tokens của 1 user
- Revoke 1 token cụ thể (user báo mất laptop)
- Redis TTL tự động cleanup
- Performance impact < 5ms

### Redis Keys

```
blacklist:user:{user_id}     → Block all tokens của user
blacklist:token:{jti}        → Block specific token
```

---

## 2. Identity Provider Abstraction

### Vấn đề

Plan hiện tại hardcode Google Workspace. Khách hàng enterprise thường dùng:

- Microsoft Azure AD
- Okta
- Custom LDAP
- SAML providers

### Giải pháp

Thiết kế theo Interface pattern - dễ thêm provider mới.

### Implementation Timeline

- **Phase 1 (Tuần 1):** Interface design + Google implementation
- **Phase 2 (Tuần 12):** Azure AD implementation
- **Phase 3 (Future):** Okta, LDAP implementations

### Interface Methods

```go
type IdentityProvider interface {
    GetAuthURL(state string) string
    ExchangeCode(code string) (*TokenResponse, error)
    GetUserInfo(accessToken string) (*UserInfo, error)
    GetUserGroups(accessToken, email string) ([]string, error)
    ValidateToken(accessToken string) error
}
```

### Supported Providers

| Provider         | Status     | Priority |
| ---------------- | ---------- | -------- |
| Google Workspace | ✅ Phase 1 | High     |
| Azure AD         | 🔄 Phase 3 | High     |
| Okta             | 📋 Future  | Medium   |
| LDAP             | 📋 Future  | Low      |

---

## 3. Key Rotation Strategy

### Vấn đề

Hiện tại mount cứng file `.pem` vào container. Nếu private key bị lộ:

- Phải redeploy toàn bộ hệ thống
- Downtime trong quá trình thay key
- Không có audit trail

### Giải pháp

Thiết kế key rotation mechanism với multiple active keys.

### Implementation Phases

**Phase 1 (Tuần 1): Flexible Key Loading**

- Support file, env, k8s secrets
- No hardcoded paths
- Easy to change keys manually

**Phase 2 (Tuần 12): Automatic Rotation**

- Generate new key pair every 30 days
- Multiple active keys (old + new)
- Zero-downtime rotation
- Grace period 15 minutes

### Key Rotation Flow

```
Day 0:  Key A (active) → Sign new tokens
Day 30: Key B generated → Key A (rotating), Key B (active)
        - New tokens signed with Key B
        - Old tokens (Key A) still valid for 15 min
Day 30 + 15min: Key A retired → Only Key B active
```

### Database Schema

```sql
CREATE TABLE auth.jwt_keys (
    kid VARCHAR(50) PRIMARY KEY,
    private_key TEXT NOT NULL,
    public_key TEXT NOT NULL,
    status VARCHAR(20), -- active | rotating | retired
    created_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    retired_at TIMESTAMPTZ
);
```

### JWKS Endpoint

Expose multiple public keys:

```json
{
  "keys": [
    { "kid": "2026-01", "kty": "RSA", "n": "...", "e": "AQAB" },
    { "kid": "2026-02", "kty": "RSA", "n": "...", "e": "AQAB" }
  ]
}
```

---

## Comparison Matrix

| Feature                    | Without Enhancement    | With Enhancement        |
| -------------------------- | ---------------------- | ----------------------- |
| **Token Revocation**       | Wait 15 min for expiry | Instant (< 100ms)       |
| **Provider Support**       | Google only            | Google + Azure + Okta   |
| **Key Compromise**         | Redeploy all services  | Rotate key in 1 command |
| **Downtime on Key Change** | 5-10 minutes           | Zero downtime           |
| **Compliance**             | Basic                  | Enterprise-grade        |

---

## Security Benefits

### 1. Instant Response to Security Incidents

- Employee termination → Revoke access immediately
- Lost device → Block specific token
- Suspicious activity → Block user instantly

### 2. Flexibility for Enterprise Customers

- Customer uses Azure AD → No problem
- Customer uses Okta → Easy to add
- Customer uses custom LDAP → Interface ready

### 3. Proactive Security Posture

- Regular key rotation (30 days)
- Audit trail for all key changes
- Compliance with ISO 27001, SOC 2

---

## Implementation Effort

| Enhancement          | Phase 1 (Tuần 1)   | Phase 3 (Tuần 12)  | Total   |
| -------------------- | ------------------ | ------------------ | ------- |
| Token Blacklist      | 2h (Redis check)   | 2h (Admin API)     | 4h      |
| Provider Abstraction | 3h (Interface)     | 4h (Azure impl)    | 7h      |
| Key Rotation         | 2h (Flexible load) | 1d (Auto rotation) | 10h     |
| **Total**            | **7h**             | **1d + 6h**        | **21h** |

---

## Testing Strategy

### Token Blacklist

- Unit test: Redis lookup performance
- Integration test: Revoke user → API returns 401
- Load test: 10k requests/sec with blacklist check

### Provider Abstraction

- Unit test: Mock provider interface
- Integration test: Google OAuth flow
- Integration test: Azure AD OAuth flow

### Key Rotation

- Unit test: Key generation
- Integration test: Sign with Key A, verify with Key B
- E2E test: Rotate key → No downtime

---

## Monitoring & Alerting

### Metrics to Track

- `auth.blacklist.checks` - Blacklist lookup count
- `auth.blacklist.hits` - Revoked token attempts
- `auth.provider.errors` - Provider API failures
- `auth.key.rotation.success` - Successful rotations
- `auth.key.rotation.failures` - Failed rotations

### Alerts

- Blacklist hit rate > 1% → Possible attack
- Provider error rate > 5% → Provider down
- Key rotation failed → Manual intervention needed

---

## References

- JWT Best Practices: https://datatracker.ietf.org/doc/html/rfc8725
- JWKS Specification: https://datatracker.ietf.org/doc/html/rfc7517
- OAuth 2.0 Security: https://datatracker.ietf.org/doc/html/rfc6749
