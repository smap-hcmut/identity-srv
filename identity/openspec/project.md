# Project Context

## Purpose
SMAP Identity Service is a comprehensive authentication and subscription management system built with Go. It provides secure user authentication, subscription-based access control, and asynchronous task processing for the SMAP platform.

**Key Goals:**
- Provide secure user authentication with email verification
- Manage subscription-based access control with plan-based system
- Support automatic free trial creation (14 days) upon email verification
- Enable asynchronous email processing via RabbitMQ and SMTP
- Build production-ready RESTful APIs following Clean Architecture principles
- Support horizontal scaling with separate API and consumer services

## Tech Stack

### Core Technologies
- **Go**: 1.25+ (primary programming language)
- **Gin**: HTTP web framework for RESTful APIs
- **PostgreSQL**: 15+ (primary relational database)
- **SQLBoiler**: Type-safe ORM with code generation
- **RabbitMQ**: 3.x (message broker for async task processing)

### Security & Authentication
- **golang-jwt**: JWT token generation and validation
- **bcrypt**: Password hashing with salt
- **SHA256**: Role encryption

### Infrastructure & DevOps
- **Docker**: Containerization with multi-stage builds
- **BuildKit**: Docker build optimization
- **Distroless**: Minimal runtime image (debian12)
- **MinIO**: S3-compatible object storage (optional)

### Development Tools
- **Swagger/Swag**: API documentation generation
- **Uber Zap**: Structured logging
- **go-mail**: SMTP email sending
- **Make**: Build automation

## Project Conventions

### Code Style
- Follow standard Go formatting (`gofmt`, `goimports`)
- Use `golangci-lint` for code quality checks
- Package naming: lowercase, single word when possible
- File naming: `snake_case.go` for multi-word files
- Function naming: `PascalCase` for exported, `camelCase` for private
- Error handling: Always return errors, use `errors.New()` or `fmt.Errorf()`
- Context propagation: Pass `context.Context` as first parameter
- Structured logging: Use Zap logger with context-aware methods (`Infof`, `Errorf`, `Warnf`)

### Architecture Patterns
**Clean Architecture with 4-Layer Separation:**

1. **Delivery Layer** (`internal/*/delivery/`)
   - HTTP handlers (Gin framework)
   - Request/Response DTOs (presenters)
   - Input validation
   - RabbitMQ producers
   - Middleware (Auth, CORS, Error handling, Locale)

2. **UseCase Layer** (`internal/*/usecase/`)
   - Business logic and orchestration
   - Validation rules
   - Error handling
   - Transaction management
   - Calls to repository layer

3. **Repository Layer** (`internal/*/repository/`)
   - Data access abstraction
   - Query builders
   - Database operations (CRUD)
   - Pagination logic
   - Filtering & sorting

4. **Domain Layer** (`internal/model/`, `internal/*/interface.go`)
   - Domain entities and interfaces
   - Domain errors
   - Type definitions

**Key Patterns:**
- Dependency Injection: Constructor functions (`New()`) for all components
- Interface Segregation: Small, focused interfaces per domain
- Error Wrapping: Domain-specific errors with clear messages
- Soft Delete: `deleted_at` timestamp for data retention
- Async Processing: RabbitMQ for email sending and future background tasks

### Testing Strategy
- Use Go's standard `testing` package
- Table-driven tests for multiple scenarios
- Test file naming: `*_test.go` in same package
- Test function naming: `TestFunctionName`
- Coverage target: Focus on business logic (UseCase layer)
- Mock external dependencies when needed
- Integration tests for critical flows (authentication, subscription)

**Example Test Structure:**
```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name    string
        input   InputType
        want    OutputType
        wantErr bool
    }{
        {
            name: "success case",
            input: InputType{...},
            want: OutputType{...},
            wantErr: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

### Git Workflow
- **Branching Strategy**: Feature branches (`feature/amazing-feature`)
- **Commit Conventions**: Conventional commits format
  - `feat:` - New features
  - `fix:` - Bug fixes
  - `docs:` - Documentation changes
  - `refactor:` - Code refactoring
  - `test:` - Test additions/changes
  - `chore:` - Build/tooling changes
- **Pull Requests**: Required for all changes
- **Code Review**: All PRs must be reviewed before merge
- **Main Branch**: `master-identity-api` (current active branch)

## Domain Context

### Core Domains

**Authentication Domain** (`internal/authentication/`)
- User registration with email and password
- OTP-based email verification (6-digit code)
- JWT-based login and token generation
- Password hashing and verification
- Account activation flow

**User Domain** (`internal/user/`)
- User profile management
- User listing with pagination (admin)
- User detail queries
- Soft delete support

**Plan Domain** (`internal/plan/`)
- Subscription plan CRUD operations
- Plan definitions with usage limits
- Plan code uniqueness

**Subscription Domain** (`internal/subscription/`)
- Subscription lifecycle management
- Automatic free trial creation (14 days)
- Subscription status tracking: `trialing`, `active`, `cancelled`, `expired`, `past_due`
- Subscription cancellation
- User-to-plan relationships

**Email Domain** (`internal/smtp/`)
- Async email sending via RabbitMQ
- Email verification with OTP
- Localized email templates (EN, VI)
- SMTP integration (Gmail, SendGrid, custom)

### Key Business Rules
- Users must verify email before account activation
- Free trial automatically created upon email verification (14 days)
- Subscriptions track trial end dates and subscription end dates
- Soft delete preserves data for audit purposes
- JWT tokens contain user ID, role, and expiration
- Role-based access control: `USER` and `ADMIN` roles

## Important Constraints

### Technical Constraints
- **Go Version**: Must use Go 1.25+ (specified in `go.mod`)
- **Database**: PostgreSQL 15+ required
- **Message Queue**: RabbitMQ 3.x required for async processing
- **Container Security**: Use distroless images, non-root user (UID 65532)
- **API Base Path**: All APIs under `/identity` prefix
- **Port**: API server runs on port 8080 by default

### Security Constraints
- **Password Security**: Must use bcrypt hashing (never store plaintext)
- **JWT Secrets**: Must be rotated regularly in production
- **SMTP Credentials**: Use app passwords (not account passwords)
- **Environment Variables**: Never commit `.env` file, use `template.env`
- **SQL Injection**: Use parameterized queries via SQLBoiler (no raw SQL)
- **CORS**: Configurable allowed origins
- **Input Validation**: All requests must be validated

### Business Constraints
- **Email Verification**: Required before account activation
- **Free Trial**: Fixed 14-day duration
- **Soft Delete**: Data retention for audit purposes
- **Subscription Status**: Must track lifecycle states accurately
- **Async Processing**: Email sending must be asynchronous (via RabbitMQ)

### Performance Constraints
- **Horizontal Scaling**: API and consumer services must support multiple instances
- **Database Connections**: Use connection pooling
- **Message Queue**: Support multiple consumers for load balancing
- **Graceful Shutdown**: Services must handle shutdown signals properly

## External Dependencies

### Required Services

**PostgreSQL Database**
- **Purpose**: Primary data storage
- **Port**: 5432
- **Tables**: `users`, `plans`, `subscriptions`
- **Connection**: `postgresql://user:password@host:5432/database?sslmode=disable`
- **Features**: Foreign keys, indexes, soft delete (deleted_at), UUID primary keys

**RabbitMQ Message Broker**
- **Purpose**: Async task queue, message routing
- **Ports**: 5672 (AMQP), 15672 (Management UI)
- **Exchange**: `smtp_send_email_exc` (fanout, durable)
- **Queue**: `smtp_send_email` (durable)
- **Connection**: `amqp://guest:guest@localhost:5672/`
- **Features**: Fanout routing, message persistence, manual acknowledgment

**SMTP Email Server**
- **Purpose**: Email delivery
- **Port**: 587 (TLS) or 465 (SSL)
- **Supported Providers**: Gmail, SendGrid, Mailgun, custom SMTP
- **Gmail Setup**: Requires 2FA and App Password
- **Configuration**: Host, port, username, password, from address

### Optional Services

**MinIO Object Storage**
- **Purpose**: File storage, avatar uploads
- **Port**: 9000 (API), 9001 (Console)
- **S3-Compatible**: Yes
- **Features**: Bucket management, object upload/download, presigned URLs
- **Connection**: Endpoint, access key, secret key

**Discord Webhook** (Monitoring)
- **Purpose**: Error notifications and monitoring
- **Configuration**: Webhook ID and token
- **Usage**: Error reporting and system alerts

### Service Dependencies Map
```
API Server:
├─ PostgreSQL (Required) - Data persistence
├─ RabbitMQ (Required) - Async task publishing
└─ MinIO (Optional) - File storage

Consumer Service:
├─ RabbitMQ (Required) - Task consumption
├─ SMTP Server (Required) - Email delivery
└─ PostgreSQL (Optional) - Future features
```
