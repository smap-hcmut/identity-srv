# SMAP Identity Service

> Authentication and subscription management service for the SMAP project

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![RabbitMQ](https://img.shields.io/badge/RabbitMQ-3.x-FF6600?style=flat&logo=rabbitmq)](https://www.rabbitmq.com/)
[![Docker](https://img.shields.io/badge/Docker-Optimized-2496ED?style=flat&logo=docker)](https://www.docker.com/)

---

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Tech Stack](#tech-stack)
- [Getting Started](#getting-started)
- [API Documentation](#api-documentation)
- [Project Structure](#project-structure)
- [Development](#development)
- [Deployment](#deployment)
- [Documentation](#documentation)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

**SMAP Identity Service** is a comprehensive authentication and subscription management system built with Go. It provides secure user authentication, subscription-based access control, and asynchronous task processing for the SMAP platform.

### Key Capabilities

- **User Authentication**: Registration, email verification via OTP, and JWT-based login
- **Subscription Management**: Plan-based subscription system with trial periods
- **Automatic Free Trial**: 14-day free trial subscription upon email verification
- **Async Task Processing**: Email sending via RabbitMQ and SMTP
- **RESTful API**: Clean and well-documented HTTP APIs
- **Scalable Architecture**: Microservice-ready with separate API and consumer services

---

## Features

### Authentication
- User registration with email and password
- OTP-based email verification
- Secure password hashing (bcrypt)
- JWT token-based authentication
- Automatic account activation

### Subscription System
- Multiple subscription plans (Free, Premium, etc.)
- Automatic free trial creation (14 days)
- Subscription status management (trialing, active, cancelled, expired)
- User-to-plan mapping with usage limits
- Subscription cancellation support

### Asynchronous Email
- Email verification with OTP
- Localized email templates (EN, VI)
- RabbitMQ-based message queue
- SMTP integration (Gmail, SendGrid, custom)
- Separate consumer service for email processing

### Additional Features
- Pagination and filtering support
- Soft delete for data retention
- Comprehensive error handling
- Request validation
- Swagger API documentation
- Health check endpoints
- CORS support
- Graceful shutdown

---

## Architecture

### System Overview

```
┌─────────────────┐         ┌──────────────┐         ┌─────────────────┐
│   API Service   │ ──────► │   RabbitMQ   │ ──────► │ Consumer Service│
│   (Port 8080)   │ Publish │   (Queue)    │ Consume │  (Email, etc.)  │
└────────┬────────┘         └──────────────┘         └─────────────────┘
         │
         │ Read/Write
         ▼
┌─────────────────┐
│   PostgreSQL    │
│   (Database)    │
└─────────────────┘
```

### Clean Architecture

The service follows **Clean Architecture** principles with clear separation of concerns:

```
├── cmd/                    # Application entry points
│   ├── api/               # API server
│   └── consumer/          # Consumer service
├── internal/              # Business logic
│   ├── authentication/    # Auth domain
│   ├── plan/             # Plan domain
│   ├── subscription/     # Subscription domain
│   ├── user/             # User domain
│   ├── smtp/             # Email domain
│   └── httpserver/       # HTTP server setup
├── pkg/                   # Shared packages
└── config/               # Configuration
```

**Layers:**
1. **Delivery Layer**: HTTP handlers, RabbitMQ consumers
2. **UseCase Layer**: Business logic and orchestration
3. **Repository Layer**: Data access and persistence
4. **Domain Layer**: Entities and interfaces

---

## Tech Stack

### Core
- **Language**: Go 1.23+
- **Framework**: Gin Web Framework
- **Database**: PostgreSQL 15
- **Message Queue**: RabbitMQ 3.x
- **Authentication**: JWT (golang-jwt)
- **Email**: SMTP (go-mail)

### Infrastructure
- **Containerization**: Docker with multi-stage builds
- **Runtime**: Distroless (secure, minimal)
- **Build System**: BuildKit with cache optimization
- **Orchestration**: Docker Compose (development)

### Libraries
- **ORM**: SQLBoiler (type-safe, code generation)
- **Validation**: go-playground/validator
- **Logging**: Uber Zap
- **Configuration**: caarlos0/env
- **Password**: bcrypt
- **UUID**: gofrs/uuid
- **API Docs**: Swaggo/swag

---

## Getting Started

### Prerequisites

- **Go**: 1.23 or higher
- **PostgreSQL**: 15 or higher
- **RabbitMQ**: 3.x
- **Docker**: 20.10+ (optional, for containerized setup)
- **Make**: For using Makefile commands

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd smap-api/identity
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Setup PostgreSQL**
   ```bash
   # Using Docker
   docker run -d \
     --name postgres \
     -e POSTGRES_USER=postgres \
     -e POSTGRES_PASSWORD=password \
     -e POSTGRES_DB=smap_identity \
     -p 5432:5432 \
     postgres:15-alpine
   
   # Run migrations
   make migrate-up
   ```

4. **Setup RabbitMQ**
   ```bash
   docker run -d \
     --name rabbitmq \
     -p 5672:5672 \
     -p 15672:15672 \
     rabbitmq:3-management-alpine
   ```

5. **Configure environment**
   ```bash
   cp template.env .env
   # Edit .env with your configuration
   ```

6. **Generate Swagger docs**
   ```bash
   make swagger
   ```

### Running the Services

#### Option 1: Local Development

```bash
# Terminal 1: Run API server
make run-api

# Terminal 2: Run consumer service
make run-consumer
```

#### Option 2: Docker

```bash
# Build and run API server
make docker-run

# Build and run consumer service
make consumer-run
```

#### Option 3: Docker Compose (Full Stack)

```bash
docker-compose up -d
```

### Verify Installation

```bash
# Check API health
curl http://localhost:8080/health

# Check Swagger documentation
open http://localhost:8080/swagger/index.html

# Check RabbitMQ management
open http://localhost:15672  # guest/guest
```

---

## API Documentation

### Base URL
```
http://localhost:8080/api/v1
```

### Authentication Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/authentication/register` | Register new user | ❌ |
| POST | `/authentication/send-otp` | Send OTP to email | ❌ |
| POST | `/authentication/verify-otp` | Verify OTP and activate account | ❌ |
| POST | `/authentication/login` | Login and get JWT token | ❌ |

### Plan Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/plans` | List all plans | ❌ |
| GET | `/plans/page` | Get plans with pagination | ❌ |
| GET | `/plans/:id` | Get plan details | ❌ |
| POST | `/plans` | Create new plan | ✅ |
| PUT | `/plans/:id` | Update plan | ✅ |
| DELETE | `/plans/:id` | Delete plan | ✅ |

### Subscription Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/subscriptions/me` | Get my active subscription | ✅ |
| GET | `/subscriptions` | List subscriptions | ✅ |
| GET | `/subscriptions/page` | Get subscriptions with pagination | ✅ |
| GET | `/subscriptions/:id` | Get subscription details | ✅ |
| POST | `/subscriptions` | Create subscription | ✅ |
| PUT | `/subscriptions/:id` | Update subscription | ✅ |
| DELETE | `/subscriptions/:id` | Delete subscription | ✅ |
| POST | `/subscriptions/:id/cancel` | Cancel subscription | ✅ |

### Example Request

**Register User:**
```bash
curl -X POST http://localhost:8080/api/v1/authentication/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123"
  }'
```

**Login:**
```bash
curl -X POST http://localhost:8080/api/v1/authentication/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123"
  }'
```

**Get My Subscription:**
```bash
curl http://localhost:8080/api/v1/subscriptions/me \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

For complete API documentation, visit:
- **Swagger UI**: `http://localhost:8080/swagger/index.html`
- **Sequence Diagrams**: See `document/api_01_sequence_diagrams.md`

---

## Project Structure

```
identity/
├── cmd/                          # Application entry points
│   ├── api/                     # API server
│   │   ├── main.go             # API entry point
│   │   └── Dockerfile          # API Docker image
│   └── consumer/                # Consumer service
│       ├── main.go             # Consumer entry point
│       └── Dockerfile          # Consumer Docker image
│
├── internal/                     # Private application code
│   ├── authentication/          # Authentication domain
│   │   ├── delivery/
│   │   │   ├── http/           # HTTP handlers
│   │   │   └── rabbitmq/       # Message producers
│   │   ├── usecase/            # Business logic
│   │   ├── interface.go        # Domain interfaces
│   │   ├── type.go             # DTOs
│   │   └── error.go            # Domain errors
│   │
│   ├── plan/                    # Plan domain
│   │   ├── delivery/http/      # HTTP handlers
│   │   ├── repository/         # Data access
│   │   ├── usecase/            # Business logic
│   │   └── ...
│   │
│   ├── subscription/            # Subscription domain
│   │   └── ...
│   │
│   ├── user/                    # User domain
│   │   └── ...
│   │
│   ├── smtp/                    # Email domain
│   │   ├── rabbitmq/consumer/  # Email consumer
│   │   └── usecase/            # Email sending logic
│   │
│   ├── consumer/                # Consumer orchestration
│   │   ├── consumer.go         # Service coordinator
│   │   └── new.go              # Dependency injection
│   │
│   ├── httpserver/              # HTTP server setup
│   │   ├── httpserver.go       # Server initialization
│   │   └── handler.go          # Route mapping
│   │
│   ├── middleware/              # HTTP middlewares
│   │   ├── cors.go
│   │   ├── errors.go
│   │   └── locale.go
│   │
│   ├── model/                   # Domain models
│   │   ├── user.go
│   │   ├── plan.go
│   │   └── subscription.go
│   │
│   └── sqlboiler/              # Generated DB models
│
├── pkg/                         # Public packages
│   ├── log/                    # Logging
│   ├── email/                  # Email utilities
│   ├── encrypter/              # Password hashing
│   ├── rabbitmq/               # RabbitMQ client
│   ├── scope/                  # JWT management
│   ├── response/               # HTTP responses
│   └── ...
│
├── config/                      # Configuration
│   ├── config.go               # Config loader
│   └── postgre/                # PostgreSQL setup
│
├── document/                    # Documentation
│   ├── api_01_sequence_diagrams.md
│   ├── api_02_implementation_summary.md
│   ├── consumer_01_readme.md
│   ├── docker_01_optimization_guide.md
│   └── ...
│
├── migration/                   # Database migrations
│   └── 01_add_user_indexes.sql
│
├── docs/                        # Swagger docs (generated)
│   ├── swagger.json
│   └── swagger.yaml
│
├── .env                         # Environment variables (not in git)
├── template.env                 # Environment template
├── go.mod                       # Go dependencies
├── Makefile                     # Build automation
├── docker-compose.yml           # Full stack setup
└── README.md                    # This file
```

---

## Development

### Available Make Commands

```bash
# Development
make run-api              # Run API server locally
make run-consumer         # Run consumer service locally
make swagger              # Generate Swagger docs
make lint                 # Run linter
make test                 # Run tests

# Database
make migrate-up           # Run database migrations
make migrate-down         # Rollback migrations
make sqlboiler            # Generate SQLBoiler models

# Docker - API
make docker-build         # Build API Docker image (local platform)
make docker-build-amd64   # Build API for AMD64 servers
make docker-run           # Build and run API in Docker
make docker-clean         # Remove API Docker images

# Docker - Consumer
make consumer-build       # Build consumer Docker image
make consumer-run         # Build and run consumer in Docker
make consumer-clean       # Remove consumer Docker images

# Utilities
make help                 # Show all available commands
```

### Environment Variables

Key environment variables (see `template.env` for complete list):

```bash
# Server
API_PORT=8080
JWT_SECRET_KEY=your-secret-key

# PostgreSQL
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=smap_identity
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password

# RabbitMQ
RABBITMQ_URL=amqp://guest:guest@localhost:5672/

# SMTP (Gmail example)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=noreply@smap.com

# Logger
LOGGER_LEVEL=info
LOGGER_MODE=production
LOGGER_ENCODING=json
```

### Code Generation

```bash
# Generate Swagger documentation
swag init -g cmd/api/main.go

# Generate SQLBoiler models (requires sqlboiler.toml)
sqlboiler psql
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -v -cover ./...

# Run specific package tests
go test ./internal/authentication/usecase/...
```

### Adding a New Module

1. Create module structure:
   ```
   internal/
   └── newmodule/
       ├── delivery/http/
       ├── repository/
       ├── usecase/
       ├── interface.go
       ├── type.go
       └── error.go
   ```

2. Implement layers (repository → usecase → delivery)

3. Wire up in `internal/httpserver/handler.go`

4. Add routes and Swagger annotations

5. Update documentation

---

## Deployment

### Docker Deployment

**Build optimized images:**
```bash
# API server (for AMD64 servers)
make docker-build-amd64

# Consumer service
make consumer-build-amd64
```

**Push to registry:**
```bash
export REGISTRY=docker.io/yourname
make docker-push
make consumer-push
```

**Run in production:**
```bash
docker run -d \
  --name smap-identity-api \
  -p 8080:8080 \
  --env-file .env \
  --restart unless-stopped \
  yourname/smap-identity:latest

docker run -d \
  --name smap-identity-consumer \
  --env-file .env \
  --restart unless-stopped \
  yourname/smap-consumer:latest
```

### Docker Compose

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: smap_identity
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  rabbitmq:
    image: rabbitmq:3-management-alpine
    ports:
      - "5672:5672"
      - "15672:15672"

  api:
    build:
      context: .
      dockerfile: cmd/api/Dockerfile
    ports:
      - "8080:8080"
    environment:
      POSTGRES_HOST: postgres
      RABBITMQ_URL: amqp://guest:guest@rabbitmq:5672/
    depends_on:
      - postgres
      - rabbitmq

  consumer:
    build:
      context: .
      dockerfile: cmd/consumer/Dockerfile
    environment:
      POSTGRES_HOST: postgres
      RABBITMQ_URL: amqp://guest:guest@rabbitmq:5672/
    depends_on:
      - postgres
      - rabbitmq

volumes:
  postgres_data:
```

### Kubernetes (Example)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: smap-identity-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: smap-identity-api
  template:
    spec:
      containers:
      - name: api
        image: yourname/smap-identity:latest
        ports:
        - containerPort: 8080
        envFrom:
        - secretRef:
            name: smap-identity-secrets
        resources:
          requests:
            memory: "256Mi"
            cpu: "200m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

---

## Documentation

Comprehensive documentation is available in the `document/` folder:

### API Documentation
- **[Sequence Diagrams](document/api_01_sequence_diagrams.md)**: Detailed flow diagrams for all API endpoints
- **[Implementation Summary](document/api_02_implementation_summary.md)**: Technical implementation details

### Consumer Service
- **[Consumer README](document/consumer_01_readme.md)**: Consumer service overview
- **[Implementation Summary](document/consumer_02_implementation_summary.md)**: Technical details
- **[Flow Diagrams](document/consumer_03_flow_diagrams.md)**: Architecture and message flows
- **[Setup Guide](document/consumer_04_setup_guide.md)**: Step-by-step setup instructions

### Docker
- **[Optimization Guide](document/docker_01_optimization_guide.md)**: Docker optimization strategies
- **[Build Guide](document/docker_02_build_guide.md)**: Building and deploying Docker images

---

## Security

### Implemented Security Measures

- Password Hashing: bcrypt with salt
- JWT Authentication: Token-based access control
- SQL Injection Prevention: Parameterized queries via SQLBoiler
- CORS Configuration: Configurable allowed origins
- Distroless Container: No shell, minimal attack surface
- Non-root User: Container runs as UID 65532
- Environment Secrets: No hardcoded credentials
- Input Validation: Request validation middleware
- Soft Delete: Data retention for audit

### Security Best Practices

1. **Never commit `.env` file** - Use `.env.example` instead
2. **Rotate JWT secrets** regularly in production
3. **Use app passwords** for SMTP (not account passwords)
4. **Enable 2FA** on email accounts used for SMTP
5. **Use TLS/SSL** for database and RabbitMQ in production
6. **Monitor logs** for suspicious activities
7. **Keep dependencies updated**: `go get -u ./...`

---

## Contributing

Contributions are welcome! Please follow these guidelines:

1. **Fork the repository**
2. **Create a feature branch**
   ```bash
   git checkout -b feature/amazing-feature
   ```
3. **Commit your changes**
   ```bash
   git commit -m 'Add amazing feature'
   ```
4. **Push to the branch**
   ```bash
   git push origin feature/amazing-feature
   ```
5. **Open a Pull Request**

### Development Guidelines

- Follow **Clean Architecture** principles
- Write **unit tests** for business logic
- Add **Swagger annotations** for new endpoints
- Update **documentation** for significant changes
- Use **conventional commits**: `feat:`, `fix:`, `docs:`, etc.
- Ensure **linter passes**: `make lint`

---

## License

This project is part of the SMAP graduation project.

---

## Support

### Common Issues

**Issue: Cannot connect to PostgreSQL**
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Check connection
psql -h localhost -U postgres -d smap_identity
```

**Issue: Email not sent**
```bash
# For Gmail, enable 2FA and use App Password
# https://myaccount.google.com/apppasswords

# Check RabbitMQ logs
docker logs rabbitmq

# Check consumer logs
docker logs smap-consumer-dev
```

**Issue: "Port already in use"**
```bash
# Find process using port 8080
lsof -i :8080

# Kill process
kill -9 <PID>
```

### Get Help

- Email: support@smap.com
- Documentation: See `document/` folder
- Issues: Open an issue on GitHub

---

## About SMAP

SMAP is a graduation project focused on building a scalable, production-ready subscription management platform with modern architecture and best practices.

**Project Goals:**
- Demonstrate **Clean Architecture** in practice
- Implement **microservices** patterns
- Apply **DevOps** best practices (Docker, CI/CD)
- Build **production-grade** APIs
- Practice **async processing** patterns

---

## Quick Links

- **Swagger UI**: http://localhost:8080/swagger/index.html
- **Health Check**: http://localhost:8080/health
- **RabbitMQ Management**: http://localhost:15672 (guest/guest)
- **Sequence Diagrams**: [document/api_01_sequence_diagrams.md](document/api_01_sequence_diagrams.md)

---

**Built with love for SMAP Graduation Project**

*Last updated: November 2025*

