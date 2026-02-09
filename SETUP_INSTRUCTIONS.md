# 🚀 SMAP Auth Service - Setup Instructions

## Tổng quan

Bạn đã hoàn thành **Day 1-2: Core Auth Service Setup** với OAuth2 + JWT + Session Management.

## ✅ Đã hoàn thành

- ✅ OAuth2 Google authentication flow
- ✅ JWT token generation với RS256
- ✅ JWKS endpoint cho token verification
- ✅ Session management với Redis
- ✅ HttpOnly cookie handling
- ✅ Database schema migration
- ✅ Configuration management với Viper

## 📋 Setup nhanh (5 phút)

### Bước 1: Chạy setup script

```bash
./scripts/setup-dev.sh
```

Script này sẽ tự động:
- Generate RSA keypair cho JWT
- Generate encryption key
- Tạo config files
- Start Docker containers (PostgreSQL, Redis, Kafka)
- Run database migrations

### Bước 2: Setup Google OAuth

**Chi tiết xem: `docs/GOOGLE_OAUTH_SETUP.md`**

Quick steps:
1. Vào [Google Cloud Console](https://console.cloud.google.com/)
2. Tạo OAuth 2.0 Client ID
3. Add redirect URI: `http://localhost:8080/authentication/callback`
4. Copy Client ID và Secret vào `auth-config.yaml` và `.env`

**Cập nhật `auth-config.yaml`:**
```yaml
oauth2:
  client_id: YOUR_CLIENT_ID.apps.googleusercontent.com
  client_secret: YOUR_CLIENT_SECRET
  redirect_uri: http://localhost:8080/authentication/callback
```

**Cập nhật `.env`:**
```bash
GOOGLE_CLIENT_ID=YOUR_CLIENT_ID.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=YOUR_CLIENT_SECRET
```

### Bước 3: Generate SQLBoiler models

```bash
make models
```

### Bước 4: Start API server

```bash
make run-api
```

Server chạy tại: `http://localhost:8080`

### Bước 5: Test OAuth flow

```bash
# Run automated tests
./scripts/test-oauth.sh

# Or manually test in browser
open http://localhost:8080/authentication/login
```

## 🧪 Testing

### Automated Tests

```bash
./scripts/test-oauth.sh
```

Tests sẽ check:
- ✅ Health endpoint
- ✅ JWKS endpoint
- ✅ OAuth redirect
- ✅ Database connection
- ✅ Redis connection
- ✅ Kafka connection (optional)

### Manual Testing

1. **Login Flow:**
   ```
   http://localhost:8080/authentication/login
   ```
   - Redirect đến Google OAuth
   - Login với Google account (phải thuộc allowed_domains)
   - Redirect về `/dashboard`
   - Cookie `smap_auth_token` được set

2. **Get Current User:**
   ```bash
   curl http://localhost:8080/authentication/me \
     --cookie "smap_auth_token=YOUR_JWT_TOKEN"
   ```

3. **Logout:**
   ```bash
   curl -X POST http://localhost:8080/authentication/logout \
     --cookie "smap_auth_token=YOUR_JWT_TOKEN"
   ```

4. **JWKS Endpoint:**
   ```bash
   curl http://localhost:8080/authentication/.well-known/jwks.json
   ```

## 📁 Files đã tạo

### Configuration
- `auth-config.yaml` - Main configuration
- `.env` - Environment variables
- `docker-compose.dev.yml` - Docker services

### Scripts
- `scripts/setup-dev.sh` - Setup môi trường development
- `scripts/test-oauth.sh` - Test OAuth flow

### Documentation
- `docs/GOOGLE_OAUTH_SETUP.md` - Hướng dẫn setup Google OAuth chi tiết
- `docs/QUICK_START.md` - Quick start guide
- `SETUP_INSTRUCTIONS.md` - File này

### Secrets (gitignored)
- `secrets/jwt-private.pem` - RSA private key
- `secrets/jwt-public.pem` - RSA public key
- `secrets/encrypt.key` - Encryption key

## 🐳 Docker Services

```bash
# Start all services
docker-compose -f docker-compose.dev.yml up -d

# Stop all services
docker-compose -f docker-compose.dev.yml down

# View logs
docker-compose -f docker-compose.dev.yml logs -f

# Restart a service
docker-compose -f docker-compose.dev.yml restart postgres
```

**Services:**
- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`
- Kafka: `localhost:9092`
- Zookeeper: `localhost:2181`

## 🔧 Troubleshooting

### "Failed to load config"

**Giải pháp:**
```bash
# Check auth-config.yaml exists
ls -la auth-config.yaml

# Validate YAML syntax
cat auth-config.yaml
```

### "Failed to connect to PostgreSQL"

**Giải pháp:**
```bash
# Check PostgreSQL status
docker ps | grep postgres

# Check logs
docker logs smap-postgres

# Restart
docker-compose -f docker-compose.dev.yml restart postgres
```

### "Failed to load private key"

**Giải pháp:**
```bash
# Check keys exist
ls -la secrets/

# Regenerate if needed
openssl genrsa -out secrets/jwt-private.pem 2048
openssl rsa -in secrets/jwt-private.pem -pubout -out secrets/jwt-public.pem
```

### "redirect_uri_mismatch"

**Giải pháp:**
1. Check redirect URI trong Google Cloud Console
2. Phải match chính xác với `oauth2.redirect_uri` trong `auth-config.yaml`
3. Không có trailing slash

### "Domain not allowed"

**Giải pháp:**
```yaml
# Thêm domain vào auth-config.yaml
access_control:
  allowed_domains:
    - vinfast.com
    - your-domain.com
```

## 📊 API Endpoints

### Public Endpoints
- `GET /health` - Health check
- `GET /authentication/login` - OAuth login (redirect to Google)
- `GET /authentication/callback` - OAuth callback
- `GET /authentication/.well-known/jwks.json` - JWKS endpoint

### Protected Endpoints (require cookie)
- `GET /authentication/me` - Get current user
- `POST /authentication/logout` - Logout

## 🎯 Next Steps

### Day 3: Google Groups RBAC + Audit Logging
- [ ] Setup Google Directory API
- [ ] Implement Google Groups sync
- [ ] Implement role mapping
- [ ] Setup Kafka audit logging
- [ ] Implement audit consumer

### Day 4: JWT Middleware Package
- [ ] Create pkg/auth package
- [ ] Implement JWT verifier
- [ ] Implement authentication middleware
- [ ] Implement token blacklist

### Day 5: Service Integration
- [ ] Integrate auth into Project Service
- [ ] Integrate auth into Ingest Service
- [ ] Integrate auth into Knowledge Service
- [ ] Integrate auth into Notification Service

## 📚 Documentation

- **Quick Start**: `docs/QUICK_START.md`
- **Google OAuth Setup**: `docs/GOOGLE_OAUTH_SETUP.md`
- **API Documentation**: `http://localhost:8080/swagger/index.html` (after running server)

## 🆘 Support

Nếu gặp vấn đề:
1. Check logs: `docker-compose -f docker-compose.dev.yml logs -f`
2. Check API logs: Terminal output khi chạy `make run-api`
3. Run tests: `./scripts/test-oauth.sh`
4. Check documentation trong `docs/`

---

**Ready to test!** 🎉

Sau khi setup xong, chạy:
```bash
./scripts/test-oauth.sh
```

Và test trong browser:
```
http://localhost:8080/authentication/login
```
