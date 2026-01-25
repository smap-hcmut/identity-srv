# Change: Align Event-Driven Architecture with SMAP Choreography Standards

## Why

The Identity Service currently uses an ad-hoc event messaging approach with a Fanout exchange (`smtp_send_email_exc`) that violates the SMAP event-driven architecture standards defined in `event-drivent.md`. This creates several critical issues:

1. **Incorrect Exchange Type:** Using Fanout instead of Topic exchange prevents proper routing key-based message distribution
2. **Non-Standard Naming:** Exchange and routing keys don't follow the `smap.events` / `service.action` pattern, making cross-service choreography impossible
3. **Missing State Management:** No Redis integration for distributed state tracking, breaking the ability to monitor project progress across services
4. **Broken Choreography:** Cannot participate in the standard SMAP data flow: Project → Collector → Analytics → Dashboard

The event-driven documentation mandates a Topic exchange (`smap.events`) with standardized routing keys (`project.created`, `data.collected`, `analysis.finished`, `job.completed`) and Redis-based state management (DB 1, Hash structure: `smap:proj:{id}`). Without these, the Identity Service cannot properly integrate with the Collector and Analytics services.

## What Changes

### 1. **BREAKING** RabbitMQ Infrastructure Migration
- **Replace** Fanout exchange `smtp_send_email_exc` with Topic exchange `smap.events`
- **Implement** standardized routing key pattern for all events:
  - `project.created` - New project initialization (future use)
  - `email.verification.requested` - OTP email sending
  - `subscription.created` - Subscription lifecycle events
  - `subscription.cancelled` - Subscription termination events
- **Update** message structure to include event metadata (event_id, timestamp, payload)
- **Maintain** backward compatibility during migration with dual-publishing

### 2. Redis State Management Integration
- **Add** Redis client configuration (DB 1 for state management, separate from cache DB 0)
- **Implement** Redis Hash structure for project tracking: `smap:proj:{id}`
  - Fields: `status`, `total`, `done`, `errors`
  - TTL: 7 days (604800 seconds)
- **Create** atomic increment operations for progress tracking
- **Add** finish line check logic for job completion detection

### 3. Event-Driven Message Structure Standardization
- **Define** standard event envelope:
  ```json
  {
    "event_id": "evt_<uuid>",
    "timestamp": "ISO8601",
    "payload": {
      "project_id": "proj_<id>",
      "service": "identity",
      "action": "specific_action",
      "data": {...}
    }
  }
  ```
- **Migrate** existing `PublishSendEmailMsg` to new event structure
- **Add** routing key parameter to all publish operations

### 4. Infrastructure Components
- **Create** new `pkg/redis` package for Redis operations
- **Add** environment variables: `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_STATE_DB`
- **Implement** connection pooling and graceful shutdown for Redis
- **Add** health checks for Redis connectivity

### 5. Code Organization Updates
- **Refactor** `internal/authentication/delivery/rabbitmq/` to support multiple event types
- **Create** `internal/event/` domain for cross-cutting event concerns
- **Extract** common event publishing logic to shared package
- **Update** consumer to handle topic-based routing

## Impact

### Affected Specs
- **NEW:** `event-infrastructure` - RabbitMQ Topic exchange and routing key standards
- **NEW:** `state-management` - Redis-based distributed state tracking
- **NEW:** `event-publishing` - Standardized event envelope and publishing patterns
- **MODIFIED:** `authentication` - Integration with event-driven messaging
- **MODIFIED:** `subscription` - Subscription lifecycle event publishing

### Affected Code
**Core Infrastructure:**
- `pkg/rabbitmq/` - Add Topic exchange support
- `pkg/redis/` - New Redis client package
- `cmd/api/main.go` - Redis initialization
- `cmd/consumer/main.go` - Topic-based message routing

**Domain Integration:**
- `internal/authentication/delivery/rabbitmq/` - Event standardization
- `internal/authentication/usecase/` - Event publishing integration
- `internal/subscription/delivery/rabbitmq/` - New subscription events
- `internal/subscription/usecase/` - Subscription lifecycle events

**Configuration:**
- `template.env` - Add Redis configuration
- `config/config.go` - Redis config struct
- `docker-compose.yml` - Add Redis service

### Migration Strategy
1. **Phase 1 (Non-Breaking):** Add Redis and new Topic exchange alongside existing Fanout exchange
2. **Phase 2 (Dual Publishing):** Publish to both old and new exchanges for 2 weeks
3. **Phase 3 (Consumer Migration):** Update consumers to subscribe to Topic exchange
4. **Phase 4 (Cleanup):** Remove old Fanout exchange and legacy code

### Breaking Changes
- **RabbitMQ Exchange:** Services consuming `smtp_send_email_exc` must migrate to `smap.events` with routing key subscriptions
- **Message Format:** Event payload structure changed to include event envelope
- **Queue Bindings:** Queues must bind with routing key patterns (e.g., `email.#`, `subscription.#`)

### Non-Breaking Additions
- Redis state management (opt-in for new features)
- New event types for subscription lifecycle
- Health check endpoints for Redis

### Testing Requirements
- Unit tests for Redis operations (atomic increments, finish checks)
- Integration tests for dual-publishing migration
- End-to-end tests for event choreography flow
- Load tests for Redis under high-throughput scenarios
- Chaos testing for partial message delivery

### Documentation Updates
- Architecture decision record (ADR) for event-driven migration
- Redis operation runbook (connection, monitoring, troubleshooting)
- Event routing key registry
- Migration guide for dependent services
