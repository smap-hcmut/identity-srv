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
         ┌─────────┐ ┌───────┐ ┌───────┐
         │Postgres │ │ Redis │ │ Kafka │
         │(schema  │ │ (DB0) │ │       │
         │identity)│ │       │ │       │
         └─────────┘ └───────┘ └───┬───┘
                                  │
                         ┌────────▼────────┐
                         │    Consumer     │
                         │  (Audit Logs)   │
                         └─────────────────┘
```

**3 Services:**

- **API Service**: OAuth2 login, JWT generation, authentication (NO Kafka producer)
- **Consumer Service**: Kafka consumer for audit log processing
- **Test Client**: Simple HTML page for testing OAuth flow

**Note**: The API service writes audit logs directly to PostgreSQL. Kafka is used only by the Consumer service for async processing.

---

## Tech Stack

| Component | Technology      | Purpose                  |
| --------- | --------------- | ------------------------ |
| Language  | Go 1.25+        | Backend                  |
| Framework | Gin             | HTTP routing             |
| Database  | PostgreSQL      | User data, audit logs    |
| Cache     | Redis           | Session, token blacklist |
| Queue     | Kafka           | Async audit events       |
| Auth      | OAuth2 (Google) | User authentication      |
| JWT       | HS256           | Token signing            |

---

## Features

- **OAuth2 Login**: Google Workspace integration
- **JWT Authentication**: HS256 symmetric signing
- **HttpOnly Cookies**: XSS-protected token storage
- **Role-Based Access**: ADMIN, ANALYST, VIEWER roles
- **Email-to-Role Mapping**: Direct role assignment from config
- **Token Blacklist**: Instant token revocation
- **Audit Logging**: Audit written to PostgreSQL; Consumer processes events from Kafka
- **Session Management**: Redis-backed sessions

---

## Quick Start

### Prerequisites

- Go 1.25+
- PostgreSQL 15+
- Redis 6+
- Kafka (optional; required only for Consumer service audit processing)

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
# Create database (name must match postgres.dbname in config)
createdb smap_auth

# Run migration (creates schema_identity and tables)
psql -h localhost -U postgres -d smap_auth -f migration/01_auth_service_schema.sql

# Or using Docker
docker run --rm \
  -v "$(pwd)/migration:/migration" \
  -e PGPASSWORD="your_password" \
  postgres:15-alpine \
  psql -h host.docker.internal -p 5432 -U postgres -d smap_auth \
  -f /migration/01_auth_service_schema.sql
```

**Note**: All tables live in schema `schema_identity`. Set `postgres.schema` to `schema_identity` and `postgres.dbname` to the database you created (e.g. `smap_auth`) in `auth-config.yaml`.

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
# Start dependencies
# Note: You need PostgreSQL, Redis, and Kafka running
# Either install locally or use Docker containers

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

# Get current user (use cookie from browser or Bearer token in dev)
curl http://localhost:8080/authentication/me --cookie "smap_auth_token=<YOUR_TOKEN>"
# In development, or: curl -H "Authorization: Bearer <YOUR_TOKEN>" http://localhost:8080/authentication/me

# OAuth test page (dev only): http://localhost:8080/test
```

---

## Configuration

Key settings in `config/auth-config.yaml`:

```yaml
# Environment: development = token in JSON + Authorization header; production = cookie only, redirect
environment:
  name: development

# PostgreSQL (dbname and schema must match the migrated database)
postgres:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: smap_auth
  sslmode: disable
  schema: schema_identity

# JWT (HS256 symmetric key)
jwt:
  algorithm: HS256
  secret_key: your-secret-key-min-32-characters
  ttl: 28800 # 8 hours

# Access Control (email-to-role mapping)
access_control:
  allowed_domains:
    - gmail.com
    - yourdomain.com
  allowed_redirect_urls: # Prevent open redirect
    - /dashboard
    - /
    - http://localhost:3000
  user_roles:
    admin@yourdomain.com: ADMIN
    analyst@yourdomain.com: ANALYST
  default_role: VIEWER

# Redis (shared for session and blacklist)
redis:
  host: localhost
  port: 6379
  password: ""
  db: 0
```

**Authentication modes:**

- **development**: Token in JSON response; supports `Authorization: Bearer <token>` header
- **production**: HttpOnly cookie only; redirect after login

---

## API Endpoints

### Public

- `GET /authentication/login` — Redirect to Google OAuth
- `GET /authentication/callback` — OAuth callback handler

### Protected (cookie or Bearer token required)

- `POST /authentication/logout` — Logout
- `GET /authentication/me` — Current user info
- `GET /audit-logs` — List audit logs (ADMIN only; pagination and date filters)

### Internal (service-to-service; X-Service-Key header when enabled)

- `POST /authentication/internal/validate` — Validate JWT
- `POST /authentication/internal/revoke-token` — Revoke token (ADMIN only)
- `GET /authentication/internal/users/:id` — Get user by ID

### System

- `GET /health` — Health check
- `GET /ready` — Readiness check
- `GET /live` — Liveness check
- `GET /swagger/*any` — Swagger docs (e.g. `/swagger/index.html`)
- `GET /test` — OAuth test page (only when `environment.name` is not `production`)

---

## Project Structure

```
identity-srv/
├── cmd/
│   ├── api/              # API server (OAuth2, JWT, auth)
│   ├── consumer/         # Kafka consumer (audit events → PostgreSQL)
│   └── test-client/      # OAuth test HTML (also served at /test in dev)
├── config/               # Configuration (auth-config.yaml, config.go)
├── internal/
│   ├── authentication/   # OAuth login, session, blacklist, roles
│   ├── audit/            # Audit (HTTP handler + Kafka producer/consumer)
│   ├── user/             # User repository & usecase
│   ├── consumer/         # Kafka consumer bootstrap
│   ├── httpserver/       # Router, middleware, health
│   ├── middleware/      # CORS, auth, admin, locale, recovery, service auth
│   ├── model/            # Domain models (User, Role, AuditLog, ...)
│   └── sqlboiler/        # Generated DB models
├── pkg/
│   ├── jwt/              # JWT issue/verify
│   ├── oauth/            # OAuth providers (Google, Okta, Azure)
│   ├── redis/            # Redis client
│   ├── kafka/            # Kafka consumer
│   ├── auth/             # JWT verification, middleware
│   └── ...               # log, i18n, response, encrypter, scope, etc.
├── migration/            # SQL schema (schema_identity)
├── scripts/              # Docker build (build-api.sh, build-consumer.sh)
├── docs/                 # Swagger (docs.go, swagger.json, swagger.yaml)
└── documents/            # Integration, deployment, testing docs
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
# Build (recommended: use Makefile)
make docker-build          # API image (local)
make consumer-build        # Consumer image (local)

# Hoặc build trực tiếp
docker build -t identity-srv:latest -f cmd/api/Dockerfile .
docker build -t smap-consumer:latest -f cmd/consumer/Dockerfile .

# Chạy container API
docker run -d -p 8080:8080 \
  -v $(pwd)/config:/app/config \
  identity-srv:latest
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

- [Auth Service Integration](documents/auth-service-integration.md)
- [Deployment Guide](documents/deployment-guide.md)
- [Local Testing Guide](documents/local-testing-guide.md)
- [Service Integration Practical Guide](documents/service-integration-practical-guide.md)
- Swagger UI: `/swagger/index.html` when the API is running

---

## License

Part of SMAP graduation project.

---

**Version**: 2.0.0  
**Last Updated**: 15/02/2026
