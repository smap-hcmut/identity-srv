# Implementation Tasks: Event-Driven Architecture Alignment

## 1. Infrastructure Setup

### 1.1 Redis Integration
- [ ] 1.1.1 Add Redis service to `docker-compose.yml` (version 7.x, port 6379)
- [ ] 1.1.2 Add Redis configuration to `template.env` (`REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_STATE_DB`)
- [ ] 1.1.3 Create `pkg/redis/config.go` with Redis configuration struct
- [ ] 1.1.4 Create `pkg/redis/client.go` with connection pool and graceful shutdown
- [ ] 1.1.5 Add Redis state client initialization to `cmd/api/main.go`
- [ ] 1.1.6 Add Redis health check endpoint `/health/redis`

### 1.2 Redis State Management Package
- [ ] 1.2.1 Create `pkg/redis/state.go` with state management functions
- [ ] 1.2.2 Implement `InitProjectState(ctx, projectID) error` - Create Hash with initial fields
- [ ] 1.2.3 Implement `SetProjectTotal(ctx, projectID, total) error` - Update total and status
- [ ] 1.2.4 Implement `MarkItemDone(ctx, projectID, isError) (status, error)` - Atomic increment + finish check
- [ ] 1.2.5 Implement `GetProjectProgress(ctx, projectID) (ProgressData, error)` - HGETALL with percentage calc
- [ ] 1.2.6 Write unit tests for atomic operations (concurrent workers scenario)
- [ ] 1.2.7 Write integration tests with real Redis instance

### 1.3 RabbitMQ Topic Exchange Setup
- [ ] 1.3.1 Add `smap.events` Topic exchange declaration in `pkg/rabbitmq/connection.go`
- [ ] 1.3.2 Update `PublishArgs` struct to include `RoutingKey string` field
- [ ] 1.3.3 Keep existing `smtp_send_email_exc` Fanout exchange (for dual-publishing)
- [ ] 1.3.4 Add `ENABLE_DUAL_PUBLISH` environment variable to `template.env` (default: `true`)
- [ ] 1.3.5 Update RabbitMQ health check to verify both exchanges exist

## 2. Event Publishing Infrastructure

### 2.1 Event Envelope and Publisher
- [ ] 2.1.1 Create `internal/event/envelope.go` with `EventEnvelope` struct (event_id, timestamp, payload)
- [ ] 2.1.2 Create `internal/event/publisher.go` with `EventPublisher` interface
- [ ] 2.1.3 Implement `PublishEvent(ctx, routingKey, payload) error` with envelope wrapping
- [ ] 2.1.4 Implement dual-publishing logic (conditional based on `ENABLE_DUAL_PUBLISH`)
- [ ] 2.1.5 Add event validation (required fields, payload size <256KB)
- [ ] 2.1.6 Add structured logging (INFO for success, ERROR for failure)
- [ ] 2.1.7 Add metrics instrumentation (`events_published_total`, `event_publish_failures_total`, `event_publish_duration_seconds`)

### 2.2 Event Routing Keys Constants
- [ ] 2.2.1 Create `internal/event/routing_keys.go` with constants:
  - `RoutingKeyEmailVerificationRequested = "email.verification.requested"`
  - `RoutingKeySubscriptionCreated = "subscription.created"`
  - `RoutingKeySubscriptionCancelled = "subscription.cancelled"`
  - `RoutingKeySubscriptionStatusChanged = "subscription.status_changed"`
  - `RoutingKeyProjectCreated = "project.created"` (future use)

### 2.3 Event Payload Schemas
- [ ] 2.3.1 Create `internal/event/payloads.go` with typed payload structs:
  - `EmailVerificationPayload` (user_id, recipient, otp_code, locale, expires_at)
  - `SubscriptionCreatedPayload` (subscription_id, user_id, plan_code, status, trial_start, trial_end)
  - `SubscriptionCancelledPayload` (subscription_id, user_id, previous_status, cancellation_reason, effective_until)
  - `SubscriptionStatusChangedPayload` (subscription_id, previous_status, new_status, transition_reason)

## 3. Authentication Domain Integration

### 3.1 Refactor Authentication Event Publishing
- [ ] 3.1.1 Update `internal/authentication/delivery/rabbitmq/producer/new.go` to inject `EventPublisher`
- [ ] 3.1.2 Refactor `PublishSendEmail` to use new `EventPublisher.PublishEvent` with routing key `email.verification.requested`
- [ ] 3.1.3 Wrap email data in `EmailVerificationPayload` struct
- [ ] 3.1.4 Add validation for email format, OTP code (6 digits), locale (en/vi)
- [ ] 3.1.5 Calculate and include `expires_at` timestamp (now + 10 minutes)
- [ ] 3.1.6 Update error handling to rollback user registration if event publish fails
- [ ] 3.1.7 Add structured logging with masked email (e.g., `ab***@example.com`)

### 3.2 Authentication UseCase Updates
- [ ] 3.2.1 Update `internal/authentication/usecase/register.go` to handle event publish errors
- [ ] 3.2.2 Ensure database transaction rollback if event publishing fails
- [ ] 3.2.3 Add retry logic for transient RabbitMQ errors (3 attempts with exponential backoff)

## 4. Subscription Domain Event Publishing

### 4.1 Subscription Event Producer
- [ ] 4.1.1 Create `internal/subscription/delivery/rabbitmq/producer/new.go` with `EventPublisher` injection
- [ ] 4.1.2 Implement `PublishSubscriptionCreated(ctx, subscriptionData) error` with routing key `subscription.created`
- [ ] 4.1.3 Implement `PublishSubscriptionCancelled(ctx, subscriptionData) error` with routing key `subscription.cancelled`
- [ ] 4.1.4 Implement `PublishSubscriptionStatusChanged(ctx, statusData) error` with routing key `subscription.status_changed`
- [ ] 4.1.5 Add idempotency_key (subscription_id) to all subscription events
- [ ] 4.1.6 Add event validation (UUID format, status enum, timestamp consistency)

### 4.2 Subscription UseCase Integration
- [ ] 4.2.1 Update `internal/subscription/usecase/create.go` to publish `subscription.created` event after DB commit
- [ ] 4.2.2 Update `internal/subscription/usecase/cancel.go` to publish `subscription.cancelled` event
- [ ] 4.2.3 Add structured logging at WARN level for cancellation events (business-critical)
- [ ] 4.2.4 Ensure events include all required fields (subscription_id, user_id, plan_code, status, timestamps)

## 5. Consumer Update for Topic Exchange

### 5.1 SMTP Consumer Topic Subscription
- [ ] 5.1.1 Create new queue `smtp_send_email_v2` in `internal/smtp/rabbitmq/consumer/consumer.go`
- [ ] 5.1.2 Bind `smtp_send_email_v2` queue to `smap.events` exchange with routing key pattern `email.#`
- [ ] 5.1.3 Update consumer to parse `EventEnvelope` structure
- [ ] 5.1.4 Extract `payload.data` and pass to existing email sending logic
- [ ] 5.1.5 Add logging for `event_id` to enable event tracing
- [ ] 5.1.6 Run both old and new consumers in parallel during migration (dual-subscription)

### 5.2 Consumer Error Handling
- [ ] 5.2.1 Add Dead Letter Queue (DLQ) `email.errors` for failed email sends
- [ ] 5.2.2 Implement retry logic with exponential backoff (3 attempts: 1s, 2s, 4s)
- [ ] 5.2.3 Log failed events to DLQ with error context
- [ ] 5.2.4 Add metrics for consumer processing (`emails_processed_total`, `emails_failed_total`)

## 6. Testing

### 6.1 Unit Tests
- [ ] 6.1.1 Test Redis atomic increment operations (concurrent workers)
- [ ] 6.1.2 Test finish line detection logic (multiple workers completing simultaneously)
- [ ] 6.1.3 Test event envelope wrapping (event_id uniqueness, timestamp format)
- [ ] 6.1.4 Test event validation (required fields, payload size limits)
- [ ] 6.1.5 Test dual-publishing logic (conditional based on env var)

### 6.2 Integration Tests
- [ ] 6.2.1 Test email verification event publishing (end-to-end with RabbitMQ)
- [ ] 6.2.2 Test subscription event publishing (created, cancelled, status_changed)
- [ ] 6.2.3 Test Redis state management (init, increment, progress query)
- [ ] 6.2.4 Test consumer processing of new Topic exchange messages
- [ ] 6.2.5 Test graceful shutdown (Redis and RabbitMQ connection cleanup)

### 6.3 Load Tests
- [ ] 6.3.1 Test Redis under high-throughput (1000 concurrent increments)
- [ ] 6.3.2 Test RabbitMQ event publishing latency (target: <50ms p99)
- [ ] 6.3.3 Test dual-publishing performance overhead
- [ ] 6.3.4 Test consumer throughput (emails processed per second)

## 7. Migration Execution

### 7.1 Phase 1: Staging Deployment (Week 1)
- [ ] 7.1.1 Deploy Redis to staging environment
- [ ] 7.1.2 Deploy updated Identity Service with dual-publishing enabled (`ENABLE_DUAL_PUBLISH=true`)
- [ ] 7.1.3 Verify both old and new consumers receive messages
- [ ] 7.1.4 Monitor metrics for errors and latency spikes

### 7.2 Phase 2: Production Deployment (Week 2)
- [ ] 7.2.1 Deploy Redis to production environment
- [ ] 7.2.2 Deploy updated Identity Service with dual-publishing enabled
- [ ] 7.2.3 Monitor for 1 week (verify no consumer errors, check metrics)
- [ ] 7.2.4 Create alert thresholds (Redis memory >80%, event publish latency >100ms)

### 7.3 Phase 3: Cutover (Week 3)
- [ ] 7.3.1 Disable dual-publishing in production (`ENABLE_DUAL_PUBLISH=false`)
- [ ] 7.3.2 Verify only new Topic exchange receives messages
- [ ] 7.3.3 Monitor for 3 days (ensure no regressions)
- [ ] 7.3.4 Stop old Fanout consumer

### 7.4 Phase 4: Cleanup (Week 4)
- [ ] 7.4.1 Delete `smtp_send_email_exc` Fanout exchange from RabbitMQ
- [ ] 7.4.2 Remove legacy message publishing code from authentication domain
- [ ] 7.4.3 Remove `ENABLE_DUAL_PUBLISH` environment variable and logic
- [ ] 7.4.4 Update Swagger documentation with new event schemas

## 8. Documentation

### 8.1 Technical Documentation
- [ ] 8.1.1 Create `docs/event-routing-keys.md` - Registry of all routing keys and payloads
- [ ] 8.1.2 Create `docs/redis-operations-runbook.md` - Connection, monitoring, troubleshooting
- [ ] 8.1.3 Update `docs/architecture.md` - Add event-driven choreography diagram
- [ ] 8.1.4 Create `docs/event-migration-guide.md` - Guide for other services to consume events

### 8.2 API Documentation
- [ ] 8.2.1 Update Swagger annotations for event publishing (async operations)
- [ ] 8.2.2 Document event schemas in `docs/event-schemas.md` with JSON examples
- [ ] 8.2.3 Add sequence diagrams for event flows (registration → email → verification)

### 8.3 Operational Documentation
- [ ] 8.3.1 Create monitoring dashboard for event metrics (Grafana/Prometheus)
- [ ] 8.3.2 Document alert thresholds and escalation procedures
- [ ] 8.3.3 Create runbook for Redis failover procedures
- [ ] 8.3.4 Document rollback procedures for each migration phase

## 9. Validation and Approval

- [ ] 9.1 Run `openspec validate align-event-driven-architecture --strict` and resolve all issues
- [ ] 9.2 Review proposal with team and get approval for implementation plan
- [ ] 9.3 Schedule migration timeline with stakeholders (Collector, Analytics teams)
- [ ] 9.4 Obtain approval for production deployment schedule
