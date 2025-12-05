# Subscription Event Integration Specification

## ADDED Requirements

### Requirement: Subscription Creation Event Publishing
The subscription system SHALL publish subscription creation events to enable downstream services to react to new subscription activations.

#### Scenario: Free trial subscription created event
- **WHEN** a user verifies their email and a free trial subscription is automatically created
- **THEN** the system SHALL publish an event to `smap.events` exchange with:
  - Routing key: `subscription.created`
  - Event envelope containing:
    - `event_id`: Unique UUID
    - `timestamp`: ISO 8601 UTC timestamp
    - `payload`:
      - `service`: `identity`
      - `action`: `created`
      - `subscription_id`: Subscription UUID
      - `user_id`: User UUID
      - `plan_code`: Plan identifier (e.g., `FREE_TRIAL`)
      - `status`: `trialing`
      - `trial_start`: Trial start timestamp
      - `trial_end`: Trial end timestamp (14 days from start)
      - `created_at`: Subscription creation timestamp

#### Scenario: Paid subscription created event
- **WHEN** a user upgrades to a paid subscription plan
- **THEN** the system SHALL publish an event to `smap.events` exchange with:
  - Routing key: `subscription.created`
  - `status`: `active`
  - `plan_code`: Paid plan identifier (e.g., `PRO_MONTHLY`)
  - `subscription_start`: Subscription start timestamp
  - `subscription_end`: Subscription end timestamp (based on billing cycle)

#### Scenario: Subscription event idempotency
- **WHEN** publishing `subscription.created` event
- **THEN** the payload SHALL include `idempotency_key` equal to `subscription_id`
- **AND** consumers can use this to deduplicate retries

### Requirement: Subscription Cancellation Event Publishing
The subscription system SHALL publish subscription cancellation events to notify dependent services of access termination.

#### Scenario: User-initiated subscription cancellation
- **WHEN** a user cancels their active subscription
- **THEN** the system SHALL publish an event to `smap.events` exchange with:
  - Routing key: `subscription.cancelled`
  - Event envelope containing:
    - `event_id`: Unique UUID
    - `timestamp`: ISO 8601 UTC timestamp
    - `payload`:
      - `service`: `identity`
      - `action`: `cancelled`
      - `subscription_id`: Subscription UUID
      - `user_id`: User UUID
      - `plan_code`: Previous plan code
      - `previous_status`: Status before cancellation (e.g., `active`, `trialing`)
      - `new_status`: `cancelled`
      - `cancelled_at`: Cancellation timestamp
      - `cancellation_reason`: Optional user-provided reason
      - `effective_until`: Date when access actually ends (e.g., end of billing period)

#### Scenario: Trial expiration cancellation event
- **WHEN** a free trial expires without conversion to paid plan
- **THEN** the system SHALL publish `subscription.cancelled` event with:
  - `previous_status`: `trialing`
  - `cancellation_reason`: `trial_expired`
  - `effective_until`: Trial end date

#### Scenario: Admin-initiated cancellation event
- **WHEN** an admin cancels a user's subscription
- **THEN** the system SHALL publish `subscription.cancelled` event with:
  - `cancellation_reason`: `admin_action`
  - `admin_user_id`: ID of admin who performed action

### Requirement: Subscription Status Transition Events
The subscription system SHALL publish events for all subscription status transitions to enable comprehensive lifecycle tracking.

#### Scenario: Trial-to-active transition event
- **WHEN** a user upgrades from trial to paid plan
- **THEN** the system SHALL publish an event with routing key `subscription.status_changed` containing:
  - `previous_status`: `trialing`
  - `new_status`: `active`
  - `transition_reason`: `upgraded`

#### Scenario: Active-to-past_due transition event
- **WHEN** a subscription payment fails
- **THEN** the system SHALL publish an event with routing key `subscription.status_changed` containing:
  - `previous_status`: `active`
  - `new_status`: `past_due`
  - `transition_reason`: `payment_failed`

#### Scenario: Active-to-expired transition event
- **WHEN** a subscription reaches its end date without renewal
- **THEN** the system SHALL publish an event with routing key `subscription.status_changed` containing:
  - `previous_status`: `active`
  - `new_status`: `expired`
  - `transition_reason`: `subscription_ended`

### Requirement: Subscription Event Validation
The subscription system SHALL validate all subscription event payloads before publishing to ensure data integrity.

#### Scenario: Required field validation for subscription events
- **WHEN** preparing to publish a subscription event
- **THEN** the system SHALL validate:
  - `subscription_id` is a valid UUID
  - `user_id` is a valid UUID
  - `plan_code` is non-empty and matches existing plan
  - `status` is one of: `trialing`, `active`, `cancelled`, `expired`, `past_due`
- **AND** return error if any validation fails

#### Scenario: Timestamp consistency validation
- **WHEN** publishing `subscription.created` event for trial
- **THEN** the system SHALL validate:
  - `trial_end > trial_start`
  - `trial_end - trial_start == 14 days`
- **AND** return error if constraints violated

### Requirement: Subscription Event Logging
The subscription system SHALL log all published subscription events with structured fields for audit and compliance.

#### Scenario: Subscription creation event logging
- **WHEN** publishing `subscription.created` event
- **THEN** the system SHALL log at INFO level:
  - `event_id`
  - `routing_key: subscription.created`
  - `subscription_id`
  - `user_id`
  - `plan_code`
  - `status`

#### Scenario: Subscription cancellation event logging
- **WHEN** publishing `subscription.cancelled` event
- **THEN** the system SHALL log at WARN level (business-critical action):
  - `event_id`
  - `routing_key: subscription.cancelled`
  - `subscription_id`
  - `user_id`
  - `cancellation_reason`
  - `effective_until`

#### Scenario: Failed subscription event publish logging
- **WHEN** subscription event publishing fails
- **THEN** the system SHALL log at ERROR level:
  - `error` message
  - `event_type` (created/cancelled/status_changed)
  - `subscription_id`
  - `user_id`
  - Stack trace for debugging
