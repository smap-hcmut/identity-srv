# SMAP Auth Service

> Enterprise-grade authentication and authorization service for the SMAP platform

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

---

## Overview

**SMAP Auth Service** is a production-ready authentication service built with Go, providing secure OAuth2/OIDC integration, JWT-based authentication, and role-based access control for the SMAP ecosystem.

### Key Features

- **OAuth2/OIDC Integration**: Google Workspace (Azure AD, Okta support planned)
- **JWT Authentication**: RS256 asymmetric signing with public key distribution
- **HttpOnly Cookie Auth**: XSS-protected token storage
- **Role-Based Access Control**: ADMIN, ANALYST, VIEWER roles
- **Group-Based Permissions**: Fine-grained access via Google Groups
- **Token Blacklist**: Instant revocation capability
- **Audit Logging**: Kafka-based async event tracking
- **Key Rotation**: Support for zero-downtime key rotation (planned)

---

## Quick Start

### Prerequisites

- Go 1.25+
- PostgreSQL 15+
- Redis 6+
- Docker (optional)

### Installation

```bash
# Clone repository
git clone <repository-url>
cd identity-srv

# Copy configuration template
cp config/auth-config.example.yaml config/auth-config.yaml

# Edit configuration with your credentials
nano config/auth-config.yaml

# Run with Docker Compose (recommended)
docker-compose up -d

# Or run locally
make run-api
```

### First API Call

```bash
# Check health
curl http://localhost:8080/health

# View API documentation
open http://localhost:8080/swagger/index.html

# Login (returns HttpOnly cookie)
curl -X POST http://localhost:8080/authentication/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}' \
  --cookie-jar cookies.txt

# Get current user
curl http://localhost:8080/authentication/me \
  --cookie cookies.txt
```

---

## Architecture

### High-Level Overview

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│   Browser   │ Cookie  │ Auth Service │  JWKS   │   Other     │
│             │────────►│              │────────►│  Services   │
│             │         │  - OAuth2    │         │             │
└─────────────┘         │  - JWT       │         └─────────────┘
                        │  - RBAC      │
                        └──────┬───────┘
                               │
                    ┌──────────┼──────────┐
                    ▼          ▼          ▼
              ┌──────────┐ ┌──────┐ ┌────────┐
              │PostgreSQL│ │Redis │ │ Kafka  │
              └──────────┘ └──────┘ └────────┘
```

### Tech Stack

| Component         | Technology    | Purpose                         |
| ----------------- | ------------- | ------------------------------- |
| **Language**      | Go 1.25       | High performance, strong typing |
| **Framework**     | Gin           | HTTP routing and middleware     |
| **Database**      | PostgreSQL 15 | User data, audit logs, JWT keys |
| **Cache**         | Redis 6       | Token blacklist, session store  |
| **Message Queue** | Kafka         | Async audit event publishing    |
| **Auth**          | OAuth2/OIDC   | Google Workspace integration    |
| **JWT**           | RS256         | Asymmetric token signing        |
| **Container**     | Docker        | Deployment and orchestration    |

---

## Configuration

### Configuration File

All configuration is managed through `config/auth-config.yaml`. Key options:

```yaml
# HTTP Server
http_server:
  host: 0.0.0.0
  port: 8080
  mode: release

# PostgreSQL
postgres:
  host: localhost
  port: 5432
  db_name: smap_auth
  user: postgres
  password: <password>

# Redis
redis:
  host: localhost
  port: 6379
  db: 0  # DB 1 used for blacklist

# OAuth2 (Google Workspace)
oauth2:
  client_id: <your-client-id>
  client_secret: <your-client-secret>
  redirect_uri: http://localhost:8080/authentication/callback

# JWT
jwt:
  algorithm: RS256
  private_key_path: ./keys/jwt-private.pem
  public_key_path: ./keys/jwt-public.pem
  issuer: smap-auth-service
  audience: [smap-api]
  ttl: 28800  # 8 hours

# Cookie
cookie:
  domain: .smap.com
  secure: true
  same_site: Lax
  name: smap_auth_token

# Google Workspace
google_workspace:
  service_account_key: /keys/google-sa.json
  admin_email: admin@yourdomain.com
  domain: yourdomain.com

# Access Control
access_control:
  allowed_domains:
    - yourdomain.com
  default_role: VIEWER
```

See `config/auth-config.example.yaml` for complete configuration options.

### Google OAuth Setup

See [docs/GOOGLE_OAUTH_SETUP.md](docs/GOOGLE_OAUTH_SETUP.md) for detailed instructions.

---

## API Documentation

### Endpoints

**Authentication**:

- `GET /authentication/login` - Redirect to OAuth provider
- `GET /authentication/callback` - OAuth callback handler
- `POST /authentication/logout` - Logout and expire cookie
- `GET /authentication/me` - Get current user info

**JWKS** (for other services):

- `GET /.well-known/jwks.json` - Public key distribution

**Internal** (service-to-service):

- `POST /internal/validate` - Token validation
- `POST /internal/revoke-token` - Token revocation (admin)
- `GET /internal/users/:id` - User lookup

### Swagger UI

Interactive API documentation available at:

```
http://localhost:8080/swagger/index.html
```

### Detailed Documentation

- **API Reference**: [documents/api-reference.md](documents/api-reference.md)
- **Integration Guide**: [documents/auth-service-integration.md](documents/auth-service-integration.md)
- **Deployment Guide**: [documents/deployment-guide.md](documents/deployment-guide.md)

---

## Development

### Available Commands

```bash
# Development
make run-api              # Run API server locally
make swagger              # Generate Swagger docs
make lint                 # Run linter
make test                 # Run tests

# Database
make migrate-up           # Run migrations
make migrate-down         # Rollback migrations

# Docker
make docker-build         # Build Docker image
make docker-run           # Build and run in Docker
make docker-clean         # Remove Docker images
```

### Project Structure

```
identity-srv/
├── cmd/
│   └── api/              # API server entry point
├── config/               # Configuration management
├── internal/             # Private application code
│   ├── authentication/   # Auth domain logic
│   ├── audit/           # Audit logging
│   ├── httpserver/      # HTTP server setup
│   ├── middleware/      # HTTP middlewares
│   └── model/           # Domain models
├── pkg/                  # Public packages
│   ├── auth/            # JWT verification
│   ├── jwt/             # JWT generation
│   ├── redis/           # Redis client
│   └── kafka/           # Kafka producer
├── migration/            # Database migrations
├── docs/                 # Setup guides
├── documents/            # Technical documentation
└── te/                   # Requirements & specs
```

---

## Deployment

### Docker

```bash
# Build image
docker build -t smap-auth-service:latest -f cmd/api/Dockerfile .

# Run container
docker run -d \
  --name smap-auth \
  -p 8080:8080 \
  -v $(pwd)/config/auth-config.yaml:/app/config/auth-config.yaml \
  -v $(pwd)/keys:/app/keys \
  smap-auth-service:latest
```

### Docker Compose

```bash
# Start all services (PostgreSQL, Redis, Kafka, Auth Service)
docker-compose up -d

# View logs
docker-compose logs -f auth-service

# Stop services
docker-compose down
```

### Kubernetes

See [documents/deployment-guide.md](documents/deployment-guide.md) for Kubernetes deployment instructions.

---

## Security

### Implemented Security Measures

- **Password Hashing**: bcrypt with salt
- **JWT Signing**: RS256 asymmetric algorithm
- **HttpOnly Cookies**: XSS protection
- **Token Blacklist**: Instant revocation via Redis
- **Domain Validation**: Email domain whitelist
- **Role Encryption**: Encrypted role storage in database
- **CORS Configuration**: Strict origin validation
- **Rate Limiting**: Login attempt throttling
- **Audit Logging**: Complete audit trail via Kafka

### Security Best Practices

1. **Never commit `auth-config.yaml`** - Use `auth-config.example.yaml` as template
2. **Never commit private keys** - Store in `keys/` directory (gitignored)
3. **Rotate JWT keys** regularly (30-day rotation recommended)
4. **Use strong passwords** for database and Redis
5. **Enable TLS/SSL** for all connections in production
6. **Monitor audit logs** for suspicious activities
7. **Keep dependencies updated**: `go get -u ./...`
8. **Review access control** regularly

---

## Testing

### Run Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -v -cover ./...

# Run specific package
go test ./internal/authentication/...
```

### Integration Testing

```bash
# Start test environment
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
go test -tags=integration ./...

# Cleanup
docker-compose -f docker-compose.test.yml down
```

---

## Documentation

### Available Documentation

| Document                                                         | Description                          |
| ---------------------------------------------------------------- | ------------------------------------ |
| [API Reference](documents/api-reference.md)                      | Complete API endpoint documentation  |
| [Integration Guide](documents/auth-service-integration.md)       | Guide for integrating other services |
| [Deployment Guide](documents/deployment-guide.md)                | Production deployment instructions   |
| [Troubleshooting](documents/identity-service-troubleshooting.md) | Common issues and solutions          |
| [Google OAuth Setup](docs/GOOGLE_OAUTH_SETUP.md)                 | OAuth2 configuration guide           |
| [Quick Start](docs/QUICK_START.md)                               | 5-minute setup guide                 |
| [Gaps Proposal](documents/auth-service-gaps-proposal.md)         | Future enhancements roadmap          |

### Requirements & Specs

| Document                                                  | Description                  |
| --------------------------------------------------------- | ---------------------------- |
| [Auth Flow Diagram](te/auth-flow-diagram.md)              | Authentication flow diagrams |
| [Security Enhancements](te/auth-security-enhancements.md) | Enterprise security features |
| [Migration Plan](te/migration-plan-v2.md)                 | Complete migration plan v2.9 |

---

## Roadmap

### Current Status (v1.0)

- ✅ OAuth2/OIDC with Google Workspace
- ✅ JWT RS256 authentication
- ✅ HttpOnly cookie support
- ✅ Role-based access control
- ✅ Token blacklist
- ✅ Audit logging via Kafka
- ✅ JWKS endpoint for public key distribution

### Planned Features (v2.0)

See [documents/auth-service-gaps-proposal.md](documents/auth-service-gaps-proposal.md) for details:

- ⏳ **Token Blacklist Enforcement** (2 hours) - CRITICAL
- ⏳ **Identity Provider Abstraction** (7 hours) - CRITICAL
  - Azure AD support
  - Okta support
  - LDAP support
- ⏳ **Automatic Key Rotation** (10 hours) - MEDIUM
  - 30-day rotation cycle
  - Zero-downtime rotation
  - Grace period handling

---

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit your changes: `git commit -m 'Add amazing feature'`
4. Push to the branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

### Development Guidelines

- Follow **Clean Architecture** principles
- Write **unit tests** for business logic
- Add **Swagger annotations** for new endpoints
- Update **documentation** for significant changes
- Use **conventional commits**: `feat:`, `fix:`, `docs:`, etc.

---

## Troubleshooting

### Common Issues

**Cannot connect to PostgreSQL**:

```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Test connection
psql -h localhost -U postgres -d smap_auth
```

**JWT verification fails**:

```bash
# Check JWKS endpoint
curl http://localhost:8080/.well-known/jwks.json

# Verify public key
cat secrets/jwt-public.pem
```

**Cookie not being set**:

- Check `COOKIE_SECURE=false` for HTTP (dev only)
- Verify `COOKIE_DOMAIN` matches your domain
- Check CORS `Access-Control-Allow-Credentials: true`

See [documents/identity-service-troubleshooting.md](documents/identity-service-troubleshooting.md) for more solutions.

---

## Support

- **Documentation**: See `documents/` folder
- **Issues**: Open an issue on GitHub
- **Email**: support@smap.com

---

## License

This project is part of the SMAP graduation project.

---

## Acknowledgments

Built with ❤️ using:

- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [golang-jwt](https://github.com/golang-jwt/jwt)
- [SQLBoiler](https://github.com/volatiletech/sqlboiler)
- [Swagger](https://github.com/swaggo/swag)

---

**Last Updated**: 09/02/2026  
**Version**: 1.0.0
