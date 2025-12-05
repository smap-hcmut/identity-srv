# State Management Specification

## ADDED Requirements

### Requirement: Redis Database Segregation
The system SHALL use Redis DB 1 exclusively for distributed state management, separate from cache operations in DB 0, to prevent cache eviction policies from deleting state tracking data.

#### Scenario: Redis state database connection
- **WHEN** the application initializes Redis for state management
- **THEN** it SHALL connect to Redis database 1 (not DB 0)
- **AND** use a separate Redis client instance for state operations

#### Scenario: Cache and state isolation
- **WHEN** Redis cache eviction occurs in DB 0 (LRU policy)
- **THEN** state management keys in DB 1 SHALL remain unaffected
- **AND** no state tracking data SHALL be lost

#### Scenario: Connection pool separation
- **WHEN** creating Redis connections
- **THEN** the system SHALL maintain separate connection pools for cache (DB 0) and state (DB 1)

### Requirement: Project State Tracking with Redis Hash
The system SHALL use Redis Hash data structure with key pattern `smap:proj:{project_id}` to store project state, enabling atomic operations and efficient multi-field updates.

#### Scenario: Project state initialization
- **WHEN** a new project is created
- **THEN** the system SHALL create a Redis Hash at `smap:proj:{project_id}` with fields:
  - `status` = `INITIALIZING` (String)
  - `total` = `0` (Integer)
  - `done` = `0` (Integer)
  - `errors` = `0` (Integer)
- **AND** set TTL to 604800 seconds (7 days)

#### Scenario: State field atomicity
- **WHEN** updating multiple fields in project state
- **THEN** the system SHALL use Redis Pipeline or HSET with mapping for atomic multi-field updates

#### Scenario: TTL refresh on activity
- **WHEN** project state is updated (progress increment, status change)
- **THEN** the system SHALL refresh the TTL to 604800 seconds
- **AND** prevent premature expiration of active projects

### Requirement: Atomic Progress Tracking
The system SHALL use Redis HINCRBY for atomic increment operations on `done` and `errors` counters to prevent race conditions in distributed processing environments.

#### Scenario: Atomic done counter increment
- **WHEN** an item processing completes successfully
- **THEN** the system SHALL execute `HINCRBY smap:proj:{id} done 1`
- **AND** retrieve the new `done` value atomically
- **AND** ensure no race condition occurs with concurrent increments

#### Scenario: Atomic error counter increment
- **WHEN** an item processing fails
- **THEN** the system SHALL execute both:
  - `HINCRBY smap:proj:{id} done 1` (count as processed)
  - `HINCRBY smap:proj:{id} errors 1` (track failure)
- **AND** ensure both increments are atomic

#### Scenario: Concurrent worker safety
- **WHEN** 100 workers concurrently increment `done` counter
- **THEN** the final value SHALL equal exactly 100
- **AND** no increments SHALL be lost due to race conditions

### Requirement: Finish Line Detection
The system SHALL implement atomic finish line detection to identify when all project items are processed and trigger completion events exactly once.

#### Scenario: Completion detection with atomic check
- **WHEN** incrementing the `done` counter
- **THEN** the system SHALL:
  1. Execute `HINCRBY smap:proj:{id} done 1` and capture the new value
  2. Execute `HGET smap:proj:{id} total` to get total items
  3. Compare `done >= total` (when `total > 0`)
  4. If true and `status != DONE`, execute `HSET smap:proj:{id} status DONE`
  5. Return `COMPLETED` signal to trigger `job.completed` event publication

#### Scenario: Prevent duplicate completion events
- **WHEN** multiple workers detect completion simultaneously
- **THEN** only the first worker to set `status = DONE` SHALL publish `job.completed` event
- **AND** subsequent workers SHALL detect `status == DONE` and skip event publication

#### Scenario: Handle early completion race condition
- **WHEN** analytics processes items faster than collector reports total count
- **THEN** the system SHALL NOT trigger completion until:
  - `done >= total` AND
  - `status != CRAWLING` (collector must finish discovery phase)

### Requirement: Project Progress Query
The system SHALL provide efficient project progress queries using Redis HGETALL to retrieve all state fields in a single operation.

#### Scenario: Progress percentage calculation
- **WHEN** querying project progress
- **THEN** the system SHALL execute `HGETALL smap:proj:{id}`
- **AND** calculate progress as `(done / total) * 100` (avoiding division by zero)
- **AND** return a structured response with `status`, `total`, `done`, `errors`, `percent`

#### Scenario: Missing project handling
- **WHEN** querying a non-existent project (expired or never created)
- **THEN** the system SHALL return `status: NOT_FOUND`, `percent: 0`
- **AND** NOT create a new Hash entry

#### Scenario: Real-time dashboard updates
- **WHEN** frontend polls project progress every 2 seconds
- **THEN** the Redis query SHALL complete in <10ms (p99 latency)
- **AND** support at least 1000 concurrent progress queries

### Requirement: Redis Connection Resilience
The system SHALL implement connection retry logic and graceful degradation for Redis failures to maintain service availability.

#### Scenario: Redis connection failure handling
- **WHEN** Redis connection fails during operation
- **THEN** the system SHALL:
  - Retry connection up to 3 times with exponential backoff (1s, 2s, 4s)
  - Log error with context (operation, project_id)
  - Return error to caller (do not silently fail)
  - Continue processing other requests (do not crash service)

#### Scenario: State tracking degradation
- **WHEN** Redis is unavailable for extended period
- **THEN** the system SHALL:
  - Log critical alert for monitoring
  - Allow project creation to proceed (store in PostgreSQL)
  - Disable progress tracking features temporarily
  - Re-initialize state when Redis reconnects

#### Scenario: Graceful shutdown with Redis
- **WHEN** the application receives shutdown signal (SIGTERM)
- **THEN** the system SHALL:
  - Stop accepting new Redis operations
  - Complete in-flight Redis operations (with 5-second timeout)
  - Close Redis connection pool gracefully
  - Exit without data loss
