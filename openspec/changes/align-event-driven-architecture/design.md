# Event-Driven Architecture Alignment - Design Document

## Context

The SMAP platform follows an event-driven choreography architecture as defined in `event-drivent.md`, where services communicate asynchronously through a centralized message broker (RabbitMQ) and coordinate state using Redis. The Identity Service currently implements an ad-hoc messaging pattern that violates these standards:

**Current State (Violations):**
- Uses Fanout exchange `smtp_send_email_exc` instead of Topic exchange `smap.events`
- No routing key-based message distribution (Fanout broadcasts to all queues indiscriminately)
- No Redis integration for distributed state management
- Email-specific message structure instead of generic event envelope
- Cannot participate in SMAP data flow choreography (Project → Collector → Analytics)

**Target State (Compliant):**
- Topic exchange `smap.events` with routing key pattern `<domain>.<action>`
- Redis DB 1 for state management (Hash structure: `smap:proj:{id}`)
- Standardized event envelope with metadata (event_id, timestamp, payload)
- Support for subscription lifecycle events beyond just emails
- Foundation for future project workflow integration

**Constraints:**
- **Zero Downtime Migration:** Must support dual-publishing during transition period (2 weeks)
- **Backward Compatibility:** Existing SMTP consumer must continue receiving messages
- **Performance:** Event publishing latency must remain <50ms (p99)
- **Resilience:** Redis failures must not prevent core authentication flows
- **Go 1.25+ Clean Architecture:** Maintain separation of concerns across layers

## Goals / Non-Goals

### Goals
1. **Compliance:** Align with SMAP event-driven choreography standards (Topic exchange, routing keys, Redis state management)
2. **Extensibility:** Enable subscription lifecycle event publishing (created, cancelled, status_changed)
3. **Observability:** Instrument event publishing with structured logging and metrics
4. **Migration Safety:** Support dual-publishing to prevent service disruption during rollout
5. **State Tracking:** Implement atomic Redis operations for progress tracking (future use by Collector/Analytics)

### Non-Goals
1. **Full Event Sourcing:** Not implementing event store or event replay capabilities (only publish current events)
2. **Consumer Migration:** Not updating external services (Collector, Analytics) in this change (they don't exist yet for Identity events)
3. **Synchronous Workflows:** Not replacing async events with synchronous API calls
4. **SMTP Consumer Changes:** Not modifying existing SMTP consumer logic (only adding new Topic-based consumer alongside)
5. **Project Workflow:** Not implementing full project creation flow (only infrastructure foundation)

## Decisions

### Decision 1: Redis Database Segregation (DB 0 for Cache, DB 1 for State)

**Choice:** Use Redis DB 1 exclusively for state management, separate from DB 0 (cache).

**Rationale:**
- **Cache Eviction Safety:** DB 0 typically uses LRU eviction policies for caching. If state keys (e.g., `smap:proj:{id}`) were stored in DB 0, Redis might evict them under memory pressure, causing loss of critical progress tracking data.
- **TTL Control:** State tracking requires precise TTL management (7 days for project state). Separating DBs prevents cache operations from interfering with state TTL.
- **Operational Clarity:** Monitoring and debugging are easier when cache and state are isolated.

**Alternatives Considered:**
- **Single DB with Key Prefixes:** Rejected - still vulnerable to eviction policies affecting state keys.
- **Redis Cluster with Separate Nodes:** Rejected - overly complex for current scale, adds operational overhead.

**Implementation:**
```go
// pkg/redis/client.go
func NewStateClient(cfg Config) (*redis.Client, error) {
    return redis.NewClient(&redis.Options{
        Addr: cfg.Host + ":" + cfg.Port,
        Password: cfg.Password,
        DB: 1, // State management DB
    }), nil
}
```

### Decision 2: Dual-Publishing Strategy for Zero-Downtime Migration

**Choice:** Publish to both old Fanout exchange (`smtp_send_email_exc`) and new Topic exchange (`smap.events`) during migration phase, controlled by environment variable `ENABLE_DUAL_PUBLISH`.

**Rationale:**
- **Risk Mitigation:** Prevents message loss if consumers lag in migration.
- **Gradual Rollout:** Allows testing new Topic exchange in production before cutting over.
- **Rollback Safety:** Can disable dual-publishing and revert to old exchange if issues arise.

**Alternatives Considered:**
- **Immediate Cutover:** Rejected - high risk of breaking SMTP consumer if deployment timing is misaligned.
- **Message Replay from Dead Letter Queue:** Rejected - complex recovery process, potential data loss window.
- **Blue-Green Deployment of RabbitMQ:** Rejected - infrastructure complexity, requires maintaining two message brokers.

**Migration Timeline:**
1. **Week 1:** Deploy dual-publishing (publish to both exchanges)
2. **Week 2:** Monitor metrics, verify no consumer errors
3. **Week 3:** Disable dual-publishing (publish only to `smap.events`)
4. **Week 4:** Remove old exchange and legacy code

**Implementation:**
```go
// internal/event/publisher.go
func (p *eventPublisher) PublishEvent(ctx context.Context, routingKey string, payload interface{}) error {
    // Always publish to new Topic exchange
    if err := p.publishToTopic(ctx, routingKey, payload); err != nil {
        return err
    }

    // Conditionally publish to old Fanout exchange (migration phase)
    if p.cfg.EnableDualPublish {
        p.logger.Warn(ctx, "Dual-publishing enabled, using deprecated Fanout exchange")
        return p.publishToLegacyFanout(ctx, payload)
    }

    return nil
}
```

### Decision 3: Redis Hash for Project State (Not Separate Keys)

**Choice:** Use Redis Hash (`HSET smap:proj:{id}`) to store all project fields (status, total, done, errors) instead of separate keys.

**Rationale:**
- **Atomic Multi-Field Updates:** `HSET` with mapping allows updating multiple fields in a single operation (e.g., set `total=1000` and `status=CRAWLING` atomically).
- **Single TTL Management:** One `EXPIRE` command applies to entire Hash, simpler than managing TTL for 4 separate keys.
- **Efficient Querying:** `HGETALL` retrieves all fields in one round-trip, better than 4 separate `GET` calls.
- **Memory Efficiency:** Hashes are optimized for small field counts (<100 fields).

**Alternatives Considered:**
- **Separate Keys (`smap:proj:{id}:status`, `smap:proj:{id}:total`):** Rejected - requires 4 separate `EXPIRE` commands, more network overhead for queries.
- **Single JSON String:** Rejected - no atomic field-level updates (must read-modify-write entire JSON), race condition risks.

**Implementation:**
```go
// pkg/redis/state.go
func InitProjectState(ctx context.Context, projectID string) error {
    key := fmt.Sprintf("smap:proj:%s", projectID)

    pipe := r.Pipeline()
    pipe.HSet(ctx, key, map[string]interface{}{
        "status": "INITIALIZING",
        "total": 0,
        "done": 0,
        "errors": 0,
    })
    pipe.Expire(ctx, key, 7*24*time.Hour) // 7 days
    _, err := pipe.Exec(ctx)
    return err
}
```

### Decision 4: Atomic Finish Line Detection (HINCRBY + Check Pattern)

**Choice:** Use `HINCRBY` to atomically increment `done` counter and immediately check for completion in the same function call (not separate transactions).

**Rationale:**
- **Race Condition Prevention:** 100 workers might increment `done` simultaneously. Using `HINCRBY` (atomic operation) ensures no increments are lost.
- **Completion Detection Accuracy:** By retrieving the new `done` value from `HINCRBY` and comparing it to `total` in the same code path, we avoid the "early completion" problem where `done == total` is detected multiple times.
- **Single Event Publication:** Only the first worker to set `status=DONE` will publish `job.completed` event.

**Alternatives Considered:**
- **Lua Script for Increment+Check:** Rejected - adds complexity, `HINCRBY` + conditional `HSET` is sufficient and easier to debug.
- **Distributed Lock (Redlock):** Rejected - overkill for this use case, adds latency and complexity.
- **Separate Increment and Check Transactions:** Rejected - race condition window between increment and check.

**Implementation:**
```go
// pkg/redis/state.go
func MarkItemDone(ctx context.Context, projectID string, isError bool) (string, error) {
    key := fmt.Sprintf("smap:proj:%s", projectID)

    // 1. Atomic increment (returns new value)
    newDone, err := r.HIncrBy(ctx, key, "done", 1).Result()
    if err != nil {
        return "", err
    }

    if isError {
        r.HIncrBy(ctx, key, "errors", 1)
    }

    // 2. Get total for comparison
    totalStr, err := r.HGet(ctx, key, "total").Result()
    if err != nil {
        return "PROCESSING", nil // Total not set yet
    }
    total, _ := strconv.ParseInt(totalStr, 10, 64)

    // 3. Finish line check (only first worker to hit this will succeed)
    if total > 0 && newDone >= total {
        status, _ := r.HGet(ctx, key, "status").Result()
        if status != "DONE" && status != "CRAWLING" {
            r.HSet(ctx, key, "status", "DONE")
            return "COMPLETED", nil // Signal to publish event
        }
    }

    return "PROCESSING", nil
}
```

### Decision 5: Event Envelope Structure (event_id, timestamp, payload)

**Choice:** Wrap all events in a standard envelope with `event_id` (UUID v4), `timestamp` (ISO 8601 UTC), and `payload` (domain data).

**Rationale:**
- **Event Tracing:** Unique `event_id` enables end-to-end tracing across services (e.g., "which email send event triggered this analytics job?").
- **Temporal Ordering:** Explicit `timestamp` allows consumers to order events correctly even if RabbitMQ delivers out-of-order.
- **Debugging:** Standardized structure simplifies log correlation (search logs by `event_id`).
- **Idempotency:** Consumers can use `event_id` or `payload.idempotency_key` to deduplicate retries.

**Alternatives Considered:**
- **Bare Payload (No Envelope):** Rejected - no standardized way to trace events, harder to debug.
- **RabbitMQ Message Properties Only:** Rejected - properties are lost if messages are persisted to disk or forwarded to other systems.

**Implementation:**
```go
// internal/event/envelope.go
type EventEnvelope struct {
    EventID   string      `json:"event_id"`   // evt_<uuid>
    Timestamp string      `json:"timestamp"`  // ISO 8601 UTC
    Payload   interface{} `json:"payload"`
}

func NewEventEnvelope(payload interface{}) EventEnvelope {
    return EventEnvelope{
        EventID:   fmt.Sprintf("evt_%s", uuid.New().String()),
        Timestamp: time.Now().UTC().Format(time.RFC3339),
        Payload:   payload,
    }
}
```

## Risks / Trade-offs

### Risk 1: Redis Single Point of Failure
- **Risk:** If Redis becomes unavailable, progress tracking stops, potentially blocking job completion detection.
- **Mitigation:**
  - Redis failures do NOT block core authentication flows (user registration/login still work).
  - Implement connection retry with exponential backoff (3 attempts: 1s, 2s, 4s).
  - Log critical alerts for monitoring systems to detect Redis outages.
  - Future: Add Redis Sentinel or Cluster for high availability (not in scope for this change).

### Risk 2: Dual-Publishing Performance Overhead
- **Risk:** Publishing to two exchanges doubles message broker load, potentially increasing latency.
- **Mitigation:**
  - Dual-publishing is temporary (2-week migration window).
  - Monitor `event_publish_duration_seconds` metric to detect latency spikes.
  - If latency exceeds 100ms (p99), immediately disable dual-publishing and investigate.

### Risk 3: Schema Evolution Breaking Consumers
- **Risk:** Future changes to event payload structure could break existing consumers.
- **Mitigation:**
  - Follow additive-only schema changes (add new fields, never remove or rename).
  - Version event payloads if breaking changes are unavoidable (e.g., `payload.schema_version`).
  - Document event schemas in `docs/event-schemas.md`.

### Risk 4: Redis Memory Exhaustion
- **Risk:** If projects are created faster than they expire (7-day TTL), Redis memory could fill up.
- **Mitigation:**
  - Monitor Redis memory usage with alerts at 80% threshold.
  - Set `maxmemory-policy` to `volatile-ttl` for DB 1 (evict keys with nearest TTL).
  - If memory issues arise, reduce TTL to 3 days or archive state to PostgreSQL before expiration.

### Trade-off: Eventual Consistency vs. Immediate Consistency
- **Trade-off:** Event-driven architecture is eventually consistent. A subscription created event might arrive at downstream services 100ms after the database is updated.
- **Acceptance Criteria:** This is acceptable for SMAP's use case (analytics, notifications don't need millisecond accuracy).
- **Alternative:** Synchronous API calls to downstream services would provide immediate consistency but violate choreography principles and create tight coupling.

## Migration Plan

### Phase 1: Infrastructure Setup (Week 1)
1. **Add Redis to Infrastructure**
   - Update `docker-compose.yml` to include Redis 7.x service
   - Add environment variables: `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_STATE_DB=1`
   - Deploy Redis to staging environment

2. **Implement Redis Client Package**
   - Create `pkg/redis/client.go` with connection pooling
   - Create `pkg/redis/state.go` with state management functions (`InitProjectState`, `MarkItemDone`, `GetProgress`)
   - Write unit tests for atomic operations

3. **Add Topic Exchange Support**
   - Update `pkg/rabbitmq/types.go` to support routing keys in `PublishArgs`
   - Declare `smap.events` Topic exchange on application startup
   - Keep existing `smtp_send_email_exc` Fanout exchange (dual-publishing phase)

### Phase 2: Event Publishing Refactor (Week 2)
1. **Create Event Publishing Infrastructure**
   - Create `internal/event/publisher.go` with `EventPublisher` interface
   - Implement `PublishEvent(ctx, routingKey, payload)` with dual-publishing logic
   - Add environment variable `ENABLE_DUAL_PUBLISH=true` (default: true during migration)

2. **Migrate Authentication Events**
   - Refactor `internal/authentication/delivery/rabbitmq/producer/producer.go` to use new `EventPublisher`
   - Wrap email verification messages in standard event envelope
   - Use routing key `email.verification.requested`

3. **Add Subscription Events**
   - Create `internal/subscription/delivery/rabbitmq/producer.go`
   - Implement `PublishSubscriptionCreated` and `PublishSubscriptionCancelled`
   - Integrate with `subscription.usecase` layer

### Phase 3: Consumer Update (Week 3)
1. **Deploy New Topic-Based Consumer**
   - Update `internal/smtp/rabbitmq/consumer/consumer.go` to subscribe to `smap.events` with routing key `email.#`
   - Parse new event envelope structure
   - Extract `payload.data` for email sending

2. **Run Both Consumers in Parallel**
   - Old consumer on `smtp_send_email` queue (Fanout binding)
   - New consumer on `smtp_send_email_v2` queue (Topic binding with `email.#`)
   - Both should process the same messages during dual-publishing

### Phase 4: Cutover (Week 4)
1. **Disable Dual-Publishing**
   - Set `ENABLE_DUAL_PUBLISH=false` in production
   - Monitor for errors (consumers should only receive from Topic exchange now)

2. **Remove Legacy Infrastructure**
   - Stop old Fanout consumer
   - Delete `smtp_send_email_exc` exchange
   - Remove legacy message publishing code
   - Remove `ENABLE_DUAL_PUBLISH` environment variable

3. **Documentation Update**
   - Document event routing keys in `docs/event-routing-keys.md`
   - Update architecture diagrams to show Topic exchange
   - Write Redis operations runbook for on-call engineers

### Rollback Strategy
- **If issues detected in Phase 3:** Set `ENABLE_DUAL_PUBLISH=true` and revert consumer to old queue.
- **If Redis fails in production:** Authentication still works (Redis only used for progress tracking, which is future feature).
- **If event publishing breaks:** Consumers won't receive messages, but user data is still persisted in PostgreSQL (can replay events manually if needed).

## Open Questions

1. **Q:** Should we implement event replay capabilities for failed event publishes?
   - **A:** Not in this change. If publishing fails, the transaction is rolled back (user registration fails). Users can retry. Future enhancement could add Dead Letter Queue for event replay.

2. **Q:** What happens if `total` is updated after `done` reaches the old `total` value (collector finds more items)?
   - **A:** Per `event-drivent.md`, the collector must set `status=PROCESSING_WAIT` when done discovering. Finish line check requires `done >= total AND status != CRAWLING`. This prevents premature completion.

3. **Q:** Should subscription events include PII (user email, name)?
   - **A:** No. Subscription events should only include `user_id` (UUID). Downstream services can fetch user details via API if needed (GDPR compliance).

4. **Q:** How to handle events larger than 256 KB (e.g., email with large attachment)?
   - **A:** Attachments should be stored in MinIO, and events should only include storage references (S3 object key). The payload size limit enforces this pattern.

5. **Q:** Should we version the event schema in the payload?
   - **A:** Not immediately. Start with implicit v1 schema. Add `schema_version` field only when breaking changes are required (follow semantic versioning).
