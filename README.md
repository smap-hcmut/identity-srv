# SMAP Identity Service

Authentication and authorization service for SMAP platform using OAuth2 and JWT.

---

## Architecture

```
┌─────────┐         ┌─────────┐         ┌──────────┐
│ Browser │ Cookie  │   API   │  JWT    │  Other   │
│         │────────▶│ Service │────────▶│ Services │
└─────────┘         └────┬────┘         └──────────┘
                         │
              ┌──────────┼──────────┐
              ▼          ▼          ▼
         ┌────────┐ ┌───────┐ ┌───────┐
         │Postgres│ │ Redis │ │ Kafka │
         └────────┘ └───────┘ └───────┘

┌──────────┐
│ Consumer │────▶ Audit Log Processing
└──────────┘
```

**3 Services:**
- **API Service**: OAuth2 login, JWT generation, authentication
- **Consumer Service**: Kafka consumer for audit log processing
- **Test Client**: Simple HTML page for testing OAuth flow

---

## Tech Stack

| Component | Technology | Purpose |
|-----------|-----------|---------|
| Language | Go 1.25 | Backend |
| Framework | Gin | HTTP routing |
| Database | PostgreSQL | User data, audit logs |
| Cache | Redis | Session, token blacklist |
| Queue | Kafka | Async audit events |
| Auth | OAuth2 (Google) | User authentication |
| JWT | HS256 | Token signing |

---

## Features

- **OAuth2 Login**: Google Workspace integration
- **JWT Authentication**: HS256 symmetric signing
- **HttpOnly Cookies**: XSS-protected token storage
- **Role-Based Access**: ADMIN, ANALYST, VIEWER roles
- **Email-to-Role Mapping**: Direct role assignment from config
- **Token Blacklist**: Instant token revocation
- **Audit Logging**: Kafka-based event tracking
- **Session Management**: Redis-backed sessions

---

## Quick Start

### Prerequisites

- Go 1.25+
- PostgreSQL 15+
- Redis 6+
- Kafka (optional, for audit logging)

### 1. Clone & Configure

```bash
git clone <repository-url>
cd identity-srv

# Copy config template
cp config/auth-config.example.yaml config/auth-config.yaml

# Edit with your Google OAuth credentials
nano config/auth-config.yaml
```

### 2. Setup Database

```bash
# Create database
createdb smap_auth

# Run migration
psql -d smap_auth -f migration/01_auth_service_schema.sql
```

### 3. Configure Google OAuth

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create OAuth 2.0 Client ID
3. Add redirect URI: `http://localhost:8080/authentication/callback`
4. Update `config/auth-config.yaml`:

```yaml
oauth2:
  client_id: YOUR_CLIENT_ID.apps.googleusercontent.com
  client_secret: YOUR_CLIENT_SECRET
  redirect_uri: http://localhost:8080/authentication/callback
```

### 4. Run Services

```bash
# Start dependencies (PostgreSQL, Redis, Kafka)
docker-compose up -d postgres redis kafka

# Run API service
make run-api

# Run Consumer service (optional, for audit logging)
make run-consumer
```

### 5. Test

```bash
# Health check
curl http://localhost:8080/health

# Login (opens browser)
open http://localhost:8080/authentication/login

# Get current user
curl http://localhost:8080/authentication/me \
  --cookie "smap_auth_token=<YOUR_TOKEN>"
```

---

## Configuration

Key settings in `config/auth-config.yaml`:

```yaml
# JWT (HS256 symmetric key)
jwt:
  algorithm: HS256
  secret_key: your-secret-key-min-32-characters
  ttl: 28800  # 8 hours

# Access Control (email-to-role mapping)
access_control:
  allowed_domains:
    - gmail.com
    - yourdomain.com
  user_roles:
    admin@yourdomain.com: ADMIN
    analyst@yourdomain.com: ANALYST
  default_role: VIEWER

# Redis (single DB for session + blacklist)
redis:
  host: localhost
  port: 6379
  db: 0
```

---

## API Endpoints

### Public
- `GET /authentication/login` - Redirect to Google OAuth
- `GET /authentication/callback` - OAuth callback handler

### Protected (requires cookie)
- `POST /authentication/logout` - Logout
- `GET /authentication/me` - Get current user

### Internal (service-to-service)
- `POST /internal/validate` - Validate JWT token
- `POST /internal/revoke-token` - Revoke token (admin only)
- `GET /internal/users/:id` - Get user by ID

### System
- `GET /health` - Health check
- `GET /swagger/*` - API documentation

---

## Project Structure

```
identity-srv/
├── cmd/
│   ├── api/              # API server
│   ├── consumer/         # Kafka consumer
│   └── test-client/      # Test HTML page
├── config/               # Configuration
├── internal/
│   ├── authentication/   # Auth logic
│   ├── audit/           # Audit logging
│   ├── user/            # User management
│   ├── httpserver/      # HTTP server
│   └── middleware/      # Middlewares
├── pkg/
│   ├── jwt/             # JWT generation
│   ├── oauth/           # OAuth providers
│   ├── redis/           # Redis client
│   └── kafka/           # Kafka producer
├── migration/           # Database migrations
├── scripts/             # Cleanup scripts
└── docs/                # Documentation
```

---

## Development

```bash
# Run API locally
make run-api

# Run Consumer locally
make run-consumer

# Generate Swagger docs
make swagger

# Run tests
go test ./...

# Build Docker images
make docker-build
make consumer-build
```

---

## Deployment

### Docker

```bash
# Build images
docker build -t smap-api:latest -f cmd/api/Dockerfile .
docker build -t smap-consumer:latest -f cmd/consumer/Dockerfile .

# Run containers
docker run -d -p 8080:8080 \
  -v $(pwd)/config:/app/config \
  smap-api:latest
```

### Kubernetes

```bash
# Apply manifests
kubectl apply -f manifests/configmap.yaml
kubectl apply -f manifests/secret.yaml
kubectl apply -f cmd/api/deployment.yaml
kubectl apply -f cmd/consumer/deployment.yaml
```

---

## Security

- **JWT Signing**: HS256 with 32+ character secret key
- **HttpOnly Cookies**: XSS protection
- **Token Blacklist**: Instant revocation via Redis
- **Domain Validation**: Email domain whitelist
- **CORS**: Strict origin validation
- **Audit Logging**: Complete audit trail
---

## Documentation

- [Quick Start Guide](docs/QUICK_START.md)
- [Google OAuth Setup](docs/GOOGLE_OAUTH_SETUP.md)
- [API Reference](documents/api-reference.md)
- [Integration Guide](documents/auth-service-integration.md)
- [Deployment Guide](documents/deployment-guide.md)
- [Troubleshooting](documents/identity-service-troubleshooting.md)

---

## License

Part of SMAP graduation project.

---

**Version**: 2.0.0 (Simplified)  
**Last Updated**: 14/02/2026  
