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
         │(schema │ │ (DB0) │ │       │
         │identity│ │       │ │       │
         └────────┘ └───────┘ └───┬───┘
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

**Note**: API service writes audit logs directly to PostgreSQL. Kafka is only used by Consumer service for async processing.

---

## Tech Stack

| Component | Technology      | Purpose                  |
| --------- | --------------- | ------------------------ |
| Language  | Go 1.25         | Backend                  |
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
# Create database (if not exists)
createdb smap

# Run migration (creates schema_identity and tables)
psql -h 172.16.19.10 -U master -d smap -f migration/01_auth_service_schema.sql

# Or using Docker
docker run --rm \
  -v "$(pwd)/migration:/migration" \
  -e PGPASSWORD="your_password" \
  postgres:15-alpine \
  psql -h 172.16.19.10 -p 5432 -U master -d smap \
  -f /migration/01_auth_service_schema.sql
```

**Note**: All tables are created in `schema_identity` schema.

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

# Get current user
curl http://localhost:8080/authentication/me \
  --cookie "smap_auth_token=<YOUR_TOKEN>"
```

---

## Configuration

Key settings in `config/auth-config.yaml`:

```yaml
# Environment (affects authentication mode)
environment:
  name: development # "development" = token in JSON, "production" = cookie only

# PostgreSQL
postgres:
  host: localhost
  port: 5432
  dbname: smap
  schema: schema_identity # All tables in this schema

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
  allowed_redirect_urls: # Prevent open redirect attacks
    - /dashboard
    - /
    - http://localhost:3000
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

**Authentication Modes:**

- **Development**: Token returned in JSON response, supports Authorization header
- **Production**: Token in HttpOnly cookie, automatic redirect

See [AUTHENTICATION_MODES.md](docs/AUTHENTICATION_MODES.md) for details.

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
docker build -t identity-srv:latest -f cmd/api/Dockerfile .
docker build -t smap-consumer:latest -f cmd/consumer/Dockerfile .

# Run containers
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

- [Quick Start Guide](docs/QUICK_START.md)
- [Authentication Modes](docs/AUTHENTICATION_MODES.md) - Development vs Production
- [Authentication Quick Reference](docs/AUTHENTICATION_QUICK_REFERENCE.md)
- [Google OAuth Setup](docs/GOOGLE_OAUTH_SETUP.md)
- [Testing Examples](docs/examples/README.md)
- [Features & Capabilities](documents/FEATURES_AND_CAPABILITIES.md)
- [Current Repository State](documents/CURRENT_REPOSITORY_STATE.md)
- [Integration Guide](documents/auth-service-integration.md)
- [Deployment Guide](documents/deployment-guide.md)

---

## License

Part of SMAP graduation project.

---

**Version**: 2.0.0 (Simplified)  
**Last Updated**: 14/02/2026
