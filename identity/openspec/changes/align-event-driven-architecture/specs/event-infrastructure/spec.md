# Event Infrastructure Specification

## ADDED Requirements

### Requirement: Topic Exchange for Event Routing
The system SHALL use a Topic exchange named `smap.events` (durable, non-auto-delete) for all inter-service event communication to enable routing key-based message distribution.

#### Scenario: Topic exchange configuration
- **WHEN** the application initializes RabbitMQ connection
- **THEN** it SHALL declare a Topic exchange with name `smap.events`, durable=true, auto-delete=false, internal=false

#### Scenario: Routing key pattern adherence
- **WHEN** publishing an event to the Topic exchange
- **THEN** the routing key SHALL follow the pattern `<domain>.<action>` (e.g., `email.verification.requested`, `subscription.created`)

#### Scenario: Multiple consumer routing
- **WHEN** multiple services bind queues to `smap.events` with different routing key patterns
- **THEN** each service SHALL receive only events matching its binding pattern (e.g., `email.#` receives all email events)

### Requirement: Standardized Routing Keys
The system SHALL use predefined routing keys for all SMAP event types to ensure consistent choreography across services.

#### Scenario: Email verification event routing
- **WHEN** a user registers and requires email verification
- **THEN** the system SHALL publish an event with routing key `email.verification.requested`

#### Scenario: Subscription lifecycle event routing
- **WHEN** a subscription is created
- **THEN** the system SHALL publish an event with routing key `subscription.created`

#### Scenario: Subscription cancellation event routing
- **WHEN** a subscription is cancelled
- **THEN** the system SHALL publish an event with routing key `subscription.cancelled`

#### Scenario: Future project event routing
- **WHEN** a project is created (future implementation)
- **THEN** the system SHALL publish an event with routing key `project.created`

### Requirement: Backward Compatibility During Migration
The system SHALL support dual-publishing to both legacy Fanout exchange and new Topic exchange during the migration period to prevent service disruption.

#### Scenario: Dual publishing for email events
- **WHEN** an email verification event is published during migration phase
- **THEN** the system SHALL publish to both `smtp_send_email_exc` (Fanout) and `smap.events` (Topic with routing key `email.verification.requested`)

#### Scenario: Migration phase configuration
- **WHEN** the environment variable `ENABLE_DUAL_PUBLISH` is set to `true`
- **THEN** the system SHALL publish events to both exchanges
- **AND** log warnings about deprecated exchange usage

#### Scenario: Migration completion
- **WHEN** the environment variable `ENABLE_DUAL_PUBLISH` is set to `false`
- **THEN** the system SHALL publish events only to `smap.events` Topic exchange
- **AND** NOT publish to `smtp_send_email_exc`

### Requirement: Exchange Health Monitoring
The system SHALL provide health check endpoints for RabbitMQ exchange connectivity to enable operational monitoring.

#### Scenario: RabbitMQ connection health check
- **WHEN** the `/health/rabbitmq` endpoint is called
- **THEN** the system SHALL verify connection to RabbitMQ broker
- **AND** return HTTP 200 with status `healthy` if connected
- **AND** return HTTP 503 with status `unhealthy` if disconnected

#### Scenario: Exchange existence verification
- **WHEN** the health check validates RabbitMQ
- **THEN** it SHALL verify that the `smap.events` exchange exists and is accessible
- **AND** include exchange metadata (type, durability) in the response
