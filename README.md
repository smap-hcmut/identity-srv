# SMAP API Services

Backend services for the SMAP (Subscription Management and Platform) graduation project, built with Go.

## Overview

This repository contains the microservices architecture for the SMAP platform, implementing modern Go development practices with clean architecture, event-driven design, and containerization.

## Services

### 1. Identity Service

Authentication and subscription management service.

**Location:** [`identity/`](identity/)

**Key Features:**
- User authentication (registration, email verification, login)
- **HttpOnly Cookie-based authentication** (primary method)
- JWT token validation and management
- Subscription and plan management
- Automatic 14-day free trial
- Asynchronous email processing via RabbitMQ

**Tech Stack:**
- Go 1.23+
- PostgreSQL 15
- RabbitMQ 3.x
- Gin Web Framework
- SQLBoiler ORM

**Authentication:**
- **Primary**: HttpOnly Cookies (`smap_auth_token`)
- **Legacy**: Bearer Token (deprecated, for migration)

**Documentation:** See [identity/README.md](identity/README.md)

**API Documentation:** http://localhost:8080/swagger/index.html

---

### 2. WebSocket Service

Real-time notification hub using WebSocket and Redis Pub/Sub.

**Location:** [`websocket/`](websocket/)

**Key Features:**
- **HttpOnly Cookie authentication** (shared with Identity service)
- Persistent WebSocket connections
- Redis Pub/Sub integration for message routing
- Multiple connections per user (multi-tab support)
- Automatic reconnection with retry logic
- Ping/Pong keep-alive mechanism
- Horizontal scaling ready

**Tech Stack:**
- Go 1.25+
- Redis 7.0+
- Gorilla WebSocket
- Gin Web Framework

**Authentication:**
- **Primary**: HttpOnly Cookies (automatic, no token in URL)
- **Legacy**: Query parameter token (deprecated)

**Documentation:** See [websocket/README.md](websocket/README.md)

**Complete Docs:** See [websocket/document/](websocket/document/)

---

### 3. Project Service

Project and campaign management service for SMAP.

**Location:** [`project/`](project/)

**Key Features:**
- Project CRUD with validation and status workflow
- Brand and keyword tracking per project
- Competitor monitoring with keyword mapping
- Date-range enforcement and soft-delete support
- User isolation with JWT authentication
- **HttpOnly Cookie authentication** (shared with Identity service)

**Tech Stack:**
- Go 1.23+
- PostgreSQL 15
- Gin Web Framework
- SQLBoiler ORM

**Authentication:**
- **Primary**: HttpOnly Cookies (`smap_auth_token`)
- **Legacy**: Bearer Token (deprecated, for migration)

**Documentation:** See [project/README.md](project/README.md)

**API Base Path:** `http://localhost:8080/project`

---

### 4. Collector Service

Data collection and task dispatching service for SMAP.

**Location:** [`collector/`](collector/)

**Key Features:**
- Crawl request validation and dispatch
- Platform-specific task routing (YouTube, TikTok)
- RabbitMQ-based task distribution
- Health check and monitoring endpoints
- Clean Architecture with module-first approach

**Tech Stack:**
- Go 1.23+
- RabbitMQ 3.x
- MongoDB (for persistence)
- Gin Web Framework

**Documentation:** See [collector/README.md](collector/README.md)

---

## Architecture

### System Overview

```
┌──────────────────────────────────────────────────────────────────┐
│                         Frontend Clients                         │
│                  (Web, Mobile, Desktop Apps)                     │
└──────────────┬───────────────────────────┬───────────────────────┘
               │                           │
               │ HTTP/REST                 │ WebSocket
               │                           │
               ▼                           ▼
┌──────────────────────────┐    ┌──────────────────────────┐    ┌──────────────────────────┐
│   Identity Service       │    │   Project Service        │    │   WebSocket Service      │
│   (Port 8080)            │    │   (APP_PORT, /project)   │    │   (Port 8081)            │
│                          │    │                          │    │                          │
│ - Authentication         │    │ - Project CRUD           │    │ - Real-time messages     │
│ - User management        │    │ - Brand/keyword tracking │    │ - JWT validation         │
│ - Subscription mgmt      │    │ - Competitor analysis    │    │ - Connection management  │
│ - Plan management        │    │ - JWT validation         │    │                          │
└───────┬──────────────────┘    └──────────┬───────────────┘    └──────────┬───────────────┘
        │                                   │                                │
        │                                   │                                │
        ▼                                   ▼                                ▼
┌──────────────────────┐          ┌──────────────────────┐        ┌──────────────────────┐
│ PostgreSQL (Identity)│          │ PostgreSQL (Projects)│        │       Redis          │
└──────────────────────┘          └──────────────────────┘        │    (Pub/Sub)         │
        │                                                         └──────────────────────┘
        ▼
┌──────────────────────┐
│     RabbitMQ         │
│   (Message Queue)    │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│  Consumer Service    │
│  (Email, Tasks)      │
└──────────────────────┘
```

### Service Communication

**Identity Service:**
- Publishes messages to RabbitMQ (email verification, notifications)
- Stores data in PostgreSQL
- Issues JWT tokens for authentication

**Project Service:**
- Provides REST APIs for project lifecycle management
- Stores project data, brand keywords, and competitor metadata in PostgreSQL
- Validates JWT tokens from HttpOnly cookies (shared with Identity Service)

**WebSocket Service:**
- Validates JWT tokens from HttpOnly cookies (shared with Identity Service)
- Subscribes to Redis Pub/Sub channels (`user_noti:*`)
- Delivers real-time messages to connected clients

**Collector Service:**
- Receives crawl requests and dispatches to platform-specific workers
- Routes tasks via RabbitMQ to YouTube and TikTok workers
- Validates and maps payloads for different platforms

**Integration:**
- **Authentication**: All services share HttpOnly cookie (`smap_auth_token`) from Identity Service
- **Real-time**: Other services publish to Redis: `PUBLISH user_noti:user123 {...}`
- **WebSocket**: Routes messages to connected users automatically
- **Tasks**: Collector dispatches crawl tasks via RabbitMQ to workers

---

## Authentication

All SMAP services use **HttpOnly Cookie-based authentication** for enhanced security.

### How It Works

1. **Login via Identity Service**: User authenticates and receives `smap_auth_token` cookie
2. **Automatic Cookie Transmission**: Browser automatically sends cookie with all requests
3. **Cross-Service Sharing**: Cookie is shared across all SMAP services (`.smap.com` domain)
4. **Secure Attributes**: HttpOnly, Secure, SameSite=Lax for XSS and CSRF protection

### Cookie Configuration

- **Cookie Name**: `smap_auth_token`
- **Domain**: `.smap.com` (shared across services)
- **Session Duration**: 2 hours (normal) / 30 days (remember me)
- **Attributes**: HttpOnly, Secure, SameSite=Lax

### Frontend Integration

```javascript
// Axios example
const api = axios.create({
  baseURL: 'https://smap-api.tantai.dev',
  withCredentials: true  // REQUIRED for cookie authentication
});

// Login
await api.post('/identity/authentication/login', {
  email: 'user@example.com',
  password: 'password123'
});

// Cookie is automatically sent with all subsequent requests
```

**Legacy Support**: Bearer token authentication is still supported during migration but will be removed in future versions.

---

## Getting Started

### Prerequisites

- **Go**: 1.23+ (Identity, Project, Collector), 1.25+ (WebSocket)
- **PostgreSQL**: 15+ (Identity & Project services)
- **Redis**: 7.0+ (WebSocket service)
- **RabbitMQ**: 3.x (Identity & Collector services)
- **MongoDB**: (Collector service, optional)
- **Docker**: 20.10+ (optional)
- **Make**: For build automation

### Quick Start (Docker Compose)

The fastest way to run all services:

```bash
# Clone repository
git clone <repository-url>
cd smap-api

# Start all services
docker-compose up -d

# Check services
docker-compose ps

# View logs
docker-compose logs -f
```

Services will be available at:
- Identity API: http://localhost:8080
- Identity Swagger: http://localhost:8080/swagger/index.html
- WebSocket: ws://localhost:8081/ws
- WebSocket Health: http://localhost:8081/health
- RabbitMQ Management: http://localhost:15672 (guest/guest)

### Individual Service Setup

#### Identity Service

> **Note**: Identity Service sets the authentication cookie that all other services use.

```bash
cd identity

# Copy environment template
cp template.env .env

# Edit configuration
nano .env

# Start PostgreSQL and RabbitMQ
docker run -d --name postgres -p 5432:5432 \
  -e POSTGRES_DB=smap_identity \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=password \
  postgres:15-alpine

docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 \
  rabbitmq:3-management-alpine

# Run migrations
make migrate-up

# Generate Swagger docs
make swagger

# Run API server
make run-api

# In another terminal, run consumer
make run-consumer
```

See [identity/README.md](identity/README.md) for detailed setup.

#### WebSocket Service

```bash
cd websocket

# Copy environment template
cp template.env .env

# Edit configuration
nano .env

# Start Redis
redis-server --requirepass 21042004

# Or using Docker
docker run -d --name redis -p 6379:6379 \
  redis:7-alpine redis-server --requirepass 21042004

# Run service
make run

# Test with example client
go run tests/client_example.go YOUR_JWT_TOKEN
```

See [websocket/README.md](websocket/README.md) for detailed setup.

#### Project Service

```bash
cd project

# Copy environment template
cp template.env .env

# Edit configuration
nano .env

# Important: Set JWT_SECRET to match Identity service
# Important: Configure cookie settings to match Identity service

# Start PostgreSQL (or reuse existing instance)
docker run -d --name postgres-project -p 5442:5432 \
  -e POSTGRES_DB=smap_project \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=password \
  postgres:15-alpine

# Run migrations
make migrate-up

# Generate SQLBoiler models
make sqlboiler

# Run API server
make run-api
```

See [project/README.md](project/README.md) for detailed setup.

#### Collector Service

```bash
cd collector

# Copy environment template
cp env.template .env

# Edit configuration
nano .env

# Start RabbitMQ (or reuse existing instance)
docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 \
  rabbitmq:3-management-alpine

# Start MongoDB (optional, for persistence)
docker run -d --name mongodb -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=admin \
  -e MONGO_INITDB_ROOT_PASSWORD=password \
  mongo:7

# Run consumer (dispatcher worker)
go run cmd/consumer/main.go

# Run API server (optional, for health checks)
go run cmd/api/main.go
```

See [collector/README.md](collector/README.md) for detailed setup.

---

## Development

### Project Structure

```
smap-api/
├── identity/                 # Identity and subscription service
│   ├── cmd/                 # Application entry points
│   │   ├── api/            # API server
│   │   └── consumer/       # Consumer service
│   ├── internal/           # Business logic
│   ├── pkg/                # Shared packages
│   ├── document/           # Documentation
│   └── README.md           # Service documentation
│
├── websocket/               # Real-time WebSocket service
│   ├── cmd/server/         # Application entry point
│   ├── internal/           # Core implementation
│   ├── pkg/                # Shared packages
│   ├── document/           # Comprehensive docs
│   │   ├── README.md      # Documentation index
│   │   ├── OVERVIEW.md    # Service overview
│   │   ├── ARCHITECTURE.md # Technical architecture
│   │   ├── API_REFERENCE.md # API documentation
│   │   ├── DEPLOYMENT.md   # Deployment guide
│   │   ├── integration.md  # Integration guide
│   │   └── TESTING_GUIDE.md # Testing procedures
│   ├── tests/              # Test utilities
│   └── README.md           # Service documentation
│
├── project/                 # Project management service
│   ├── cmd/api/            # Project API server
│   ├── internal/           # Project domain and business logic
│   ├── pkg/                # Shared packages
│   ├── document/           # Service documentation
│   └── README.md           # Service documentation
│
├── collector/               # Data collection and dispatch service
│   ├── cmd/                # API and consumer entry points
│   ├── internal/          # Dispatcher domain and business logic
│   ├── pkg/               # Shared packages
│   └── README.md          # Service documentation
│
├── docker-compose.yml       # Full stack setup
└── README.md                # This file
```

### Technology Stack

**Languages & Frameworks:**
- Go 1.23+ / 1.25+
- Gin Web Framework
- Gorilla WebSocket

**Databases:**
- PostgreSQL 15 (Identity & Project Services)
- Redis 7.0+ (WebSocket Service)

**Message Queue:**
- RabbitMQ 3.x (Identity Service)

**Infrastructure:**
- Docker & Docker Compose
- Multi-stage builds with distroless images
- BuildKit optimization

**Libraries & Tools:**
- SQLBoiler (ORM for Identity)
- go-redis (Redis client)
- golang-jwt (JWT authentication)
- Zap (structured logging)
- Swaggo (API documentation)
- Validator (request validation)

### Development Workflow

**Identity Service:**
```bash
cd identity
make run-api              # Run API server
make run-consumer         # Run consumer service
make swagger              # Generate Swagger docs
make test                 # Run tests
make docker-build         # Build Docker image
```

**WebSocket Service:**
```bash
cd websocket
make run                  # Run service
make build                # Build binary
make docker-build         # Build Docker image
make test                 # Run tests
```

**Project Service:**
```bash
cd project
make run-api              # Run API server
make migrate-up           # Run database migrations
make sqlboiler            # Generate SQLBoiler models
make test                 # Run tests
make docker-build         # Build Docker image
```

**Collector Service:**
```bash
cd collector
go run cmd/consumer/main.go  # Run dispatcher worker
go run cmd/api/main.go       # Run API server (optional)
make test                    # Run tests
```

### Code Standards

Both services follow:
- **Clean Architecture** principles
- **Domain-Driven Design** patterns
- **SOLID** principles
- **Dependency Injection** via constructors
- **Interface-based** design
- **Unit testing** for business logic
- **Structured logging** with Zap
- **Environment-based** configuration

---

## API Documentation

### Identity Service API

**Base URL:** `http://localhost:8080/identity`

**Key Endpoints:**
- `POST /authentication/register` - Register new user
- `POST /authentication/verify-otp` - Verify email with OTP
- `POST /authentication/login` - Login and get JWT token
- `GET /plans` - List subscription plans
- `GET /subscriptions/me` - Get current user's subscription
- `POST /subscriptions` - Create subscription

**Swagger UI:** http://localhost:8080/swagger/index.html

See [identity/README.md](identity/README.md#api-documentation) for complete API documentation.

### Project Service API

**Base URL:** `http://localhost:8080/project`

**Key Endpoints:**
- `GET /projects` - List authenticated user's projects
- `GET /projects/page` - Paginated projects with filters
- `GET /projects/:id` - Retrieve project details
- `POST /projects` - Create new project
- `PUT /projects/:id` - Update existing project
- `DELETE /projects/:id` - Soft-delete project

**Authentication:** HttpOnly Cookie (automatic) or Bearer Token (legacy)

See [project/README.md](project/README.md#api-endpoints) for payload examples and additional documentation.

### Collector Service API

**Base URL:** `http://localhost:8080` (if API server enabled)

**Key Endpoints:**
- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /live` - Liveness check

**Task Dispatch:** Via RabbitMQ (not HTTP endpoints)

See [collector/README.md](collector/README.md) for detailed documentation.

### WebSocket Service API

**Connection:** `ws://localhost:8081/ws` (cookie authentication, automatic)

**Legacy Connection:** `ws://localhost:8081/ws?token=JWT_TOKEN` (deprecated)

**Health Check:** `GET http://localhost:8081/health`

**Metrics:** `GET http://localhost:8081/metrics`

**Authentication:** HttpOnly Cookie (automatic, recommended) or Query Parameter (legacy)

**Message Format:**
```json
{
  "type": "notification",
  "payload": {...},
  "timestamp": "2025-01-21T10:30:00Z"
}
```

See [websocket/document/API_REFERENCE.md](websocket/document/API_REFERENCE.md) for complete API documentation.

---

## Deployment

### Docker Deployment

**Build all images:**
```bash
# Identity Service
cd identity
make docker-build-amd64

# WebSocket Service
cd websocket
make docker-build-amd64

# Project Service
cd project
make docker-build-amd64

# Collector Service
cd collector
# Build commands (if available)
```

**Run with Docker Compose:**
```bash
docker-compose up -d
```

**Production deployment:**
- Use Docker Compose or Kubernetes
- Enable TLS/SSL for all services
- Use managed databases (PostgreSQL, Redis)
- Configure load balancers
- Set up monitoring and logging
- Use secrets management (not .env files)

See service-specific documentation for detailed deployment guides:
- [identity/README.md#deployment](identity/README.md#deployment)
- [websocket/document/DEPLOYMENT.md](websocket/document/DEPLOYMENT.md)

### Environment Configuration

**Identity Service:**
- See [identity/template.env](identity/template.env)
- Key vars: `API_PORT`, `POSTGRES_*`, `RABBITMQ_URL`, `SMTP_*`, `JWT_SECRET_KEY`

**WebSocket Service:**
- See [websocket/template.env](websocket/template.env)
- Key vars: `WS_PORT`, `REDIS_*`, `JWT_SECRET_KEY`

**Project Service:**
- See [project/template.env](project/template.env)
- Key vars: `APP_PORT`, `POSTGRES_*`, `JWT_SECRET`, `COOKIE_*`, `LOGGER_*`

**Collector Service:**
- See [collector/env.template](collector/env.template)
- Key vars: `PORT`, `AMQP_URL`, `MONGO_*`, `LOGGER_*`

**Important:** 
- Use the same `JWT_SECRET_KEY` for Identity, Project, and WebSocket services
- Configure cookie settings (`COOKIE_NAME`, `COOKIE_DOMAIN`, etc.) consistently across all services
- Cookie domain should be `.smap.com` (with leading dot) for cross-subdomain sharing

---

## Integration

### Publishing Notifications to WebSocket Service

From any backend service (including Identity Service):

```go
import (
    "context"
    "encoding/json"
    "github.com/redis/go-redis/v9"
)

// Publish notification to user
func publishNotification(userID string, title, message string) error {
    client := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "your-redis-password",
    })
    defer client.Close()
    
    msg := map[string]interface{}{
        "type": "notification",
        "payload": map[string]interface{}{
            "title":   title,
            "message": message,
        },
    }
    
    data, _ := json.Marshal(msg)
    channel := "user_noti:" + userID
    return client.Publish(context.Background(), channel, data).Err()
}
```

See [websocket/document/integration.md](websocket/document/integration.md) for complete integration guide with examples in multiple languages.

---

## Testing

### Identity Service

```bash
cd identity

# Unit tests
go test ./...

# Test with Swagger UI
open http://localhost:8080/swagger/index.html

# Test registration flow
curl -X POST http://localhost:8080/identity/authentication/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Password123"}'
```

### WebSocket Service

```bash
cd websocket

# Generate JWT token
go run tests/generate_token.go

# Test with example client
go run tests/client_example.go YOUR_JWT_TOKEN

# Publish test message
redis-cli -a 21042004
PUBLISH user_noti:user123 '{"type":"notification","payload":{"title":"Test"}}'
```

See [websocket/document/TESTING_GUIDE.md](websocket/document/TESTING_GUIDE.md) for comprehensive testing procedures.

### Project Service

```bash
cd project

# Run unit tests
go test ./...

# Create a sample project
curl -X POST http://localhost:8080/project/projects \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Launch Plan","status":"draft","from_date":"2025-01-01T00:00:00Z","to_date":"2025-03-31T23:59:59Z"}'
```

Refer to [project/README.md](project/README.md#getting-started) for complete testing scenarios.

---

## Documentation

### Identity Service

- **[README.md](identity/README.md)** - Complete service documentation
- **[document/](identity/document/)** - Detailed technical documentation
  - API sequence diagrams
  - Implementation summaries
  - Consumer service guide
  - Docker optimization guide

### WebSocket Service

- **[README.md](websocket/README.md)** - Quick start guide
- **[document/](websocket/document/)** - Comprehensive documentation
  - **[README.md](websocket/document/README.md)** - Documentation index
  - **[OVERVIEW.md](websocket/document/OVERVIEW.md)** - Service overview
  - **[ARCHITECTURE.md](websocket/document/ARCHITECTURE.md)** - Technical architecture
  - **[API_REFERENCE.md](websocket/document/API_REFERENCE.md)** - Complete API docs
  - **[DEPLOYMENT.md](websocket/document/DEPLOYMENT.md)** - Deployment guide
  - **[integration.md](websocket/document/integration.md)** - Integration guide
  - **[TESTING_GUIDE.md](websocket/document/TESTING_GUIDE.md)** - Testing procedures

### Project Service

- **[README.md](project/README.md)** - Service overview and API usage
- **[document/](project/document/)** - Architecture and implementation notes
- **[openspec/project.md](project/openspec/project.md)** - Project context for AI assistants

### Collector Service

- **[README.md](collector/README.md)** - Service overview and architecture

---

## Security

### Implemented Security Measures

**Identity Service:**
- Password hashing with bcrypt
- HttpOnly Cookie-based authentication (primary method)
- JWT token-based authentication
- OTP email verification
- SQL injection prevention (parameterized queries)
- Input validation
- CORS configuration
- Distroless container (minimal attack surface)
- Non-root container user

**WebSocket Service:**
- JWT token validation
- TLS support for Redis
- Connection limits
- Distroless container
- Non-root container user
- No secrets in code

**Project Service:**
- HttpOnly Cookie authentication (shared with Identity)
- JWT validation for all endpoints
- Soft delete for audit trails
- Input validation and request sanitization
- Structured logging with correlation IDs
- Configurable Discord alerting

**WebSocket Service:**
- HttpOnly Cookie authentication (shared with Identity)
- JWT token validation
- Connection limits and rate limiting
- TLS support for Redis
- Distroless container
- Non-root container user
- No secrets in code

**Collector Service:**
- RabbitMQ authentication
- Input validation for crawl requests
- Task routing and validation
- Health check endpoints

### Security Best Practices

1. Never commit `.env` files to version control
2. Rotate JWT secrets regularly in production
3. Use strong passwords for databases
4. Enable TLS/SSL for all services in production
5. Use app passwords for SMTP (not account passwords)
6. Keep dependencies updated: `go get -u ./...`
7. Monitor logs for suspicious activities
8. Implement rate limiting in production
9. Use secrets management tools (Vault, AWS Secrets Manager)
10. Regular security audits

---

## Monitoring

### Health Checks

**Identity Service:**
```bash
curl http://localhost:8080/health
```

**Project Service:**
```bash
curl http://localhost:8080/project/health
```

**Collector Service:**
```bash
curl http://localhost:8080/health  # If API server enabled
```

**WebSocket Service:**
```bash
curl http://localhost:8081/health
```

### Metrics

**WebSocket Service:**
```bash
curl http://localhost:8081/metrics
```

Returns:
- Active connections count
- Unique users connected
- Messages sent/received/failed
- Service uptime

### Logging

Both services use structured logging with Zap:
- JSON format in production
- Console format in development
- Configurable log levels (debug, info, warn, error)

---

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit your changes: `git commit -m 'Add amazing feature'`
4. Push to the branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

### Development Guidelines

- Follow Clean Architecture principles
- Write unit tests for business logic
- Add Swagger annotations for new endpoints
- Update documentation for significant changes
- Use conventional commits: `feat:`, `fix:`, `docs:`, `refactor:`, etc.
- Ensure linter passes before committing
- Keep services independent and loosely coupled

---

## Troubleshooting

### Common Issues

**Cannot connect to PostgreSQL:**
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Test connection
psql -h localhost -U postgres -d smap_identity
```

**Project database migrations failing:**
```bash
# Verify SQLBoiler configuration
cat template.sqlboiler.toml

# Regenerate models
make sqlboiler
```

**Cannot connect to Redis:**
```bash
# Check if Redis is running
docker ps | grep redis

# Test connection
redis-cli -a YOUR_PASSWORD ping
```

**Port already in use:**
```bash
# Find process using port
lsof -i :8080  # or :8081

# Kill process
kill -9 <PID>
```

**Email not sent:**
- For Gmail, enable 2FA and use App Password
- Check RabbitMQ logs: `docker logs rabbitmq`
- Check consumer logs: `docker logs <consumer-container>`

**WebSocket connection failed:**
- Ensure JWT token is valid (not expired)
- Check token includes `sub` claim with user ID
- Verify Redis is running
- Check service logs for errors

---

## License

This project is part of the SMAP graduation project.

---

## About SMAP

**SMAP (Subscription Management and Platform)** is a graduation project demonstrating modern software architecture and development practices in Go.

**Project Goals:**
- Build production-grade microservices
- Implement Clean Architecture
- Apply event-driven patterns
- Practice DevOps best practices
- Real-time communication systems
- Scalable and maintainable code

**Technologies Demonstrated:**
- Microservices architecture
- RESTful API design
- WebSocket real-time communication
- Event-driven architecture (RabbitMQ, Redis Pub/Sub)
- Database design and ORM usage
- JWT authentication
- Docker containerization
- Async task processing
- API documentation with Swagger
- Comprehensive testing

---

## Quick Links

**Identity Service:**
- Swagger UI: http://localhost:8080/swagger/index.html
- Health Check: http://localhost:8080/health
- Documentation: [identity/README.md](identity/README.md)

**Project Service:**
- Health Check: http://localhost:8080/project/health
- Documentation: [project/README.md](project/README.md)

**Collector Service:**
- Health Check: http://localhost:8080/health (if API server enabled)
- Documentation: [collector/README.md](collector/README.md)

**WebSocket Service:**
- Health Check: http://localhost:8081/health
- Metrics: http://localhost:8081/metrics
- Documentation: [websocket/document/](websocket/document/)

**Infrastructure:**
- RabbitMQ Management: http://localhost:15672 (guest/guest)
- PostgreSQL: localhost:5432
- Redis: localhost:6379

---

**Built for SMAP Graduation Project**

*Last updated: 2025-11-21*

