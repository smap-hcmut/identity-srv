# Project Context

## Purpose

The **SMAP Dispatcher Service** (also known as the Collector Service) is the central coordinator for the SMAP data collection system. It serves as an intelligent task router that:

- Receives high-level crawl requests from backend services or schedulers
- Validates and transforms generic requests into platform-specific tasks
- Distributes granular tasks to specialized workers (YouTube, TikTok) via RabbitMQ
- Acts as a decoupling layer between request initiators and scraping workers

**Core Goals:**
- Provide a unified interface for multi-platform data collection
- Ensure type-safe payload mapping and validation
- Enable horizontal scalability through message queue architecture
- Maintain clean separation between business logic and transport mechanisms

## Tech Stack

### Core Technologies
- **Go 1.23.8** - Primary programming language
- **RabbitMQ (amqp091-go)** - Message queue for task distribution
- **MongoDB (mongo-driver)** - Primary database for persistence
- **PostgreSQL** - Secondary database (via sqlboiler)
- **MinIO** - Object storage with async upload capabilities

### Key Libraries
- **Zap (uber/zap)** - Structured logging
- **Env (caarlos0/env)** - Configuration management
- **UUID (google/uuid)** - Unique identifier generation
- **Zstandard (klauspost/compress)** - Data compression
- **Mockery** - Mock generation for testing

### Development Tools
- **Docker & Docker Compose** - Containerization
- **Kubernetes** - Container orchestration
- **Makefile** - Build automation
- **SQLBoiler** - Type-safe ORM code generation

## Project Conventions

### Code Style

**Package Organization:**
- `cmd/` - Application entry point (consumer)
- `internal/` - Private application code (domain logic)
- `pkg/` - Reusable packages (utilities, clients)
- `config/` - Configuration loading and management

**Naming Conventions:**
- Use descriptive, intention-revealing names
- Interfaces: `UseCase`, `Repository`, `Producer`
- Implementations: `implUseCase`, `implRepository`
- Constants: `PascalCase` for exported, `camelCase` for private
- Files: `snake_case.go` (e.g., `dispatch_uc.go`, `uc_interface.go`)

**Error Handling:**
- Define domain-specific errors in `uc_errors.go`
- Wrap errors with context using `fmt.Errorf("%w: %v", ErrType, err)`
- Return errors explicitly, avoid panics in business logic

**Logging:**
- Use structured logging with Zap
- Include context (trace IDs, job IDs) in log messages
- Log levels: Debug (development), Info (production), Error (failures)

### Architecture Patterns

**Clean Architecture:**
The project strictly follows Clean Architecture principles with clear layer separation:

1. **Delivery Layer** (`delivery/`):
   - RabbitMQ consumers and producers
   - HTTP handlers (future expansion)
   - Transport-agnostic; can be swapped without affecting business logic

2. **Use Case Layer** (`usecase/`):
   - Pure business logic (Dispatch, Mapping, Validation)
   - Platform-agnostic
   - No dependencies on external frameworks

3. **Domain Models** (`models/`):
   - Core entities: `CrawlRequest`, `CollectorTask`, `Platform`, `TaskType`
   - DTOs for platform-specific payloads (YouTube, TikTok)

**Design Patterns:**
- **Dependency Injection**: All components receive dependencies via constructors
- **Strategy/Factory Pattern**: `mapPayload()` selects validation and transformation strategies based on platform and task type
- **Producer-Consumer**: RabbitMQ-based async task distribution
- **Fan-out**: Single request dispatched to multiple platform queues

**SOLID Principles:**
- Single Responsibility: Each use case handles one specific concern
- Open/Closed: New platforms can be added without modifying existing code
- Dependency Inversion: Business logic depends on interfaces, not implementations

### Testing Strategy

**Current Test Coverage:**
- Unit tests for critical packages: `compressor`, `minio` (async upload, compression integration)
- Mock generation via Mockery (`.mockery.yaml`)

**Testing Approach:**
- **Unit Tests**: Test business logic in isolation with mocks
- **Integration Tests**: Test RabbitMQ and MinIO interactions
- **Test File Naming**: `*_test.go` alongside source files

**Future Testing Goals:**
- Increase coverage for dispatcher use cases
- Add contract tests for RabbitMQ message formats
- Implement end-to-end tests for full dispatch flow

### Git Workflow

**CI/CD Pipeline (Jenkins):**
1. **Pull Code**: Checkout from SCM
2. **Build API Image**: Build Docker image from `cmd/api/Dockerfile`
3. **Push to Registry**: Push to private registry (`registry.tantai.dev`)
4. **Deploy to Kubernetes**: Update deployment in `smap` namespace
5. **Verify Deployment**: Health check with readiness probes
6. **Cleanup**: Remove old Docker images (retain last 2)
7. **Notify Discord**: Send build status notifications

**Deployment Strategy:**
- **Environment**: Production (`smap` namespace)
- **Registry**: Private Docker registry at `registry.tantai.dev`
- **K8s Deployment**: Rolling updates with health checks
- **Monitoring**: Discord webhook notifications for build status

## Domain Context

### Supported Platforms
- **YouTube** (`PlatformYouTube`): Video search, channel crawling, comment extraction
- **TikTok** (`PlatformTikTok`): Video search, creator profiles, engagement metrics

### Task Types
1. **`research_keyword`**: Search for content based on keywords
   - Returns list of videos/posts matching search criteria
   - Supports sorting, time range filtering, result limits

2. **`crawl_links`**: Scrape specific URLs
   - Extract detailed metadata from provided video/profile URLs
   - Optional: download media, fetch comments, channel info

3. **`research_and_crawl`**: Composite task
   - First searches for content, then immediately scrapes results
   - Combines both research and crawl operations

### Message Flow
```
External Service → RabbitMQ (collector.inbound) → Dispatcher Consumer
                                                        ↓
                                    [Validate & Map Payload]
                                                        ↓
                        ┌───────────────────────────────┴───────────────────────┐
                        ↓                                                       ↓
            RabbitMQ (collector.youtube)                        RabbitMQ (collector.tiktok)
                        ↓                                                       ↓
                YouTube Worker                                          TikTok Worker
```

### Payload Mapping
The dispatcher transforms generic `CrawlRequest.Payload` into strict, platform-specific structs:
- **YouTube**: `YouTubeResearchKeywordPayload`, `YouTubeCrawlLinksPayload`, `YouTubeResearchAndCrawlPayload`
- **TikTok**: `TikTokResearchKeywordPayload`, `TikTokCrawlLinksPayload`, `TikTokResearchAndCrawlPayload`

This ensures type safety and validation at the boundary between generic requests and platform workers.

## Important Constraints

### Technical Constraints
- **Go Version**: Must use Go 1.23.8 or compatible
- **RabbitMQ Dependency**: Service cannot function without RabbitMQ connection
- **Schema Versioning**: All tasks include `schema_version` for backward compatibility
- **Trace IDs**: Required for distributed tracing and debugging

### Business Constraints
- **Retry Logic**: Tasks support configurable retry attempts (`attempt`, `max_attempts`)
- **Time Range Filtering**: All tasks support time-based filtering (e.g., last 7 days)
- **Platform Availability**: Currently limited to YouTube and TikTok

### Operational Constraints
- **Deployment**: Kubernetes-only deployment (no standalone mode)
- **Configuration**: Environment-based config (no config files)
- **Logging**: JSON-formatted logs in production for log aggregation
- **Health Checks**: Must respond to `/health`, `/ready`, `/live` endpoints

## External Dependencies

### Message Queue
- **RabbitMQ**: Core dependency for task distribution
  - **Inbound Exchange**: `collector.inbound` (topic)
  - **Inbound Queue**: `collector.inbound.queue`
  - **Routing Key**: `crawler.#`
  - **Internal Exchanges**: `collector.youtube`, `collector.tiktok` (managed by dispatcher)

### Databases
- **MongoDB**: Primary persistence layer
  - Connection via encrypted URI (`MONGODB_ENCODED_URI`)
  - Optional monitoring enabled via `MONGODB_ENABLE_MONITORING`
- **PostgreSQL**: Secondary database (future use)
  - SQLBoiler for type-safe ORM

### Object Storage
- **MinIO**: Large file storage with claim check pattern
  - Async upload with configurable workers and queue size
  - Zstandard compression for efficient storage
  - Bucket-based organization

### External Services
- **Discord Webhooks**: Build notifications and error reporting
  - Report bugs: `DISCORD_REPORT_BUG_ID` + `DISCORD_REPORT_BUG_TOKEN`
- **JWT Authentication**: Secure API access
- **Internal API Key**: Service-to-service authentication

### Infrastructure
- **Kubernetes Cluster**: `https://172.16.21.111:6443`
- **Docker Registry**: `registry.tantai.dev`
- **Namespace**: `smap`
- **Deployment**: `smap-collector` with rolling updates
