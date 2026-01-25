# Event Publishing Specification

## ADDED Requirements

### Requirement: Standardized Event Envelope
The system SHALL wrap all event payloads in a standard envelope containing event metadata (event_id, timestamp, payload) to enable event tracing and debugging across services.

#### Scenario: Event envelope structure
- **WHEN** publishing any event to `smap.events`
- **THEN** the message body SHALL contain JSON with structure:
  ```json
  {
    "event_id": "evt_<uuid_v4>",
    "timestamp": "<ISO8601_UTC>",
    "payload": {
      "service": "identity",
      "action": "<specific_action>",
      "data": { /* domain-specific data */ }
    }
  }
  ```

#### Scenario: Event ID uniqueness
- **WHEN** generating an event_id
- **THEN** it SHALL use UUID v4 prefixed with `evt_`
- **AND** be globally unique across all services and time

#### Scenario: Timestamp precision
- **WHEN** setting event timestamp
- **THEN** it SHALL use ISO 8601 format with UTC timezone (e.g., `2025-12-05T10:30:45Z`)
- **AND** include millisecond precision

### Requirement: Content Type Standardization
The system SHALL publish all events with `Content-Type: application/json` header to ensure consistent deserialization across consumers.

#### Scenario: JSON content type header
- **WHEN** publishing an event to RabbitMQ
- **THEN** the message SHALL have `ContentType = "application/json"`
- **AND** the body SHALL be valid JSON (validated before publish)

#### Scenario: Invalid JSON rejection
- **WHEN** attempting to publish malformed JSON
- **THEN** the system SHALL return an error before publishing to RabbitMQ
- **AND** log the validation failure with context

### Requirement: Routing Key Association
The system SHALL include the routing key in the event payload metadata to enable consumers to identify event types without inspecting RabbitMQ message properties.

#### Scenario: Routing key in payload
- **WHEN** publishing an event with routing key `email.verification.requested`
- **THEN** the payload SHALL include:
  ```json
  {
    "payload": {
      "service": "identity",
      "action": "verification.requested",
      "routing_key": "email.verification.requested",
      ...
    }
  }
  ```

### Requirement: Event Publishing Interface Abstraction
The system SHALL provide a unified event publishing interface that abstracts RabbitMQ implementation details and supports multiple event types.

#### Scenario: Generic event publisher interface
- **WHEN** implementing event publishing
- **THEN** the system SHALL provide an interface:
  ```go
  type EventPublisher interface {
      PublishEvent(ctx context.Context, routingKey string, payload interface{}) error
  }
  ```

#### Scenario: Typed event publishing methods
- **WHEN** publishing domain-specific events
- **THEN** the system SHALL provide typed methods wrapping the generic publisher:
  - `PublishEmailVerificationRequested(ctx, emailData) error`
  - `PublishSubscriptionCreated(ctx, subscriptionData) error`
  - `PublishSubscriptionCancelled(ctx, subscriptionData) error`

#### Scenario: Publishing error handling
- **WHEN** RabbitMQ publish fails
- **THEN** the system SHALL:
  - Return a wrapped error with context (event type, routing key)
  - Log the failure with structured fields
  - NOT retry automatically (let caller decide retry strategy)
  - NOT crash the service

### Requirement: Event Schema Validation
The system SHALL validate event payload schemas before publishing to prevent malformed events from entering the message bus.

#### Scenario: Required field validation
- **WHEN** publishing an email verification event
- **THEN** the system SHALL validate that payload contains required fields:
  - `recipient` (non-empty email address)
  - `subject` (non-empty string)
  - `body` (non-empty string)
- **AND** return validation error if any field is missing or invalid

#### Scenario: Email format validation
- **WHEN** validating email verification event
- **THEN** the `recipient` field SHALL match email regex pattern
- **AND** reject invalid email formats (e.g., `invalid@`, `@example.com`)

### Requirement: Event Payload Size Limits
The system SHALL enforce maximum payload size limits to prevent message bus overload and ensure efficient message processing.

#### Scenario: Maximum payload size enforcement
- **WHEN** attempting to publish an event
- **THEN** the serialized JSON payload SHALL NOT exceed 256 KB
- **AND** return error if size limit is exceeded

#### Scenario: Large attachment handling
- **WHEN** email contains large attachments
- **THEN** the system SHALL:
  - Store attachments in MinIO/S3
  - Include only storage references in event payload
  - NOT embed binary data directly in event

### Requirement: Event Publishing Observability
The system SHALL instrument event publishing operations with structured logging and metrics to enable operational monitoring.

#### Scenario: Successful publish logging
- **WHEN** an event is successfully published
- **THEN** the system SHALL log at INFO level with fields:
  - `event_id`
  - `routing_key`
  - `service: identity`
  - `action`
  - `timestamp`

#### Scenario: Failed publish logging
- **WHEN** event publishing fails
- **THEN** the system SHALL log at ERROR level with fields:
  - `error` (error message)
  - `routing_key`
  - `payload_size` (bytes)
  - `retry_attempt` (if applicable)

#### Scenario: Publishing metrics
- **WHEN** events are published
- **THEN** the system SHALL track metrics:
  - `events_published_total` (counter by routing_key)
  - `event_publish_failures_total` (counter by routing_key, error_type)
  - `event_publish_duration_seconds` (histogram)

### Requirement: Idempotency Key Support
The system SHALL support optional idempotency keys in event payloads to enable consumers to deduplicate events.

#### Scenario: Idempotency key inclusion
- **WHEN** publishing an event with idempotency requirement
- **THEN** the payload MAY include an `idempotency_key` field with a unique identifier
- **AND** consumers SHALL use this key to detect duplicate processing

#### Scenario: Subscription event idempotency
- **WHEN** publishing `subscription.created` event
- **THEN** the idempotency_key SHALL be the subscription UUID
- **AND** enable consumers to safely retry processing without side effects
