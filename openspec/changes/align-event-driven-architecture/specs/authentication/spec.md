# Authentication Event Integration Specification

## MODIFIED Requirements

### Requirement: Email Verification Event Publishing
The authentication system SHALL publish email verification requests to the event bus using standardized routing keys and event envelope structure, replacing direct email queue publishing.

#### Scenario: User registration triggers email verification event
- **WHEN** a new user completes registration
- **THEN** the system SHALL publish an event to `smap.events` exchange with:
  - Routing key: `email.verification.requested`
  - Event envelope containing:
    - `event_id`: Unique UUID (e.g., `evt_a1b2c3d4-...`)
    - `timestamp`: ISO 8601 UTC timestamp
    - `payload`:
      - `service`: `identity`
      - `action`: `verification.requested`
      - `user_id`: User UUID
      - `recipient`: User email address
      - `otp_code`: 6-digit verification code
      - `locale`: User's preferred language (`en` or `vi`)
      - `expires_at`: OTP expiration timestamp

#### Scenario: Backward compatibility during migration
- **WHEN** dual-publishing is enabled (migration phase)
- **THEN** the system SHALL publish to both:
  - Legacy: `smtp_send_email_exc` (Fanout) with old message format
  - New: `smap.events` (Topic) with routing key `email.verification.requested` and new event envelope
- **AND** log deprecation warning for legacy exchange usage

#### Scenario: Email verification event validation
- **WHEN** preparing to publish email verification event
- **THEN** the system SHALL validate:
  - `recipient` is a valid email format
  - `otp_code` is exactly 6 digits
  - `locale` is either `en` or `vi`
  - `expires_at` is in the future
- **AND** return error if validation fails (before publishing)

#### Scenario: Email verification event failure handling
- **WHEN** publishing email verification event fails
- **THEN** the system SHALL:
  - Log error with user context (user_id, email)
  - Return HTTP 500 to registration request (transaction not committed)
  - NOT create user account in database
  - Allow user to retry registration

### Requirement: OTP Expiration Tracking
The authentication system SHALL include OTP expiration metadata in email verification events to enable consumers to validate token freshness.

#### Scenario: OTP expiration timestamp generation
- **WHEN** generating an email verification event
- **THEN** the `expires_at` field SHALL be set to `now + 10 minutes` (600 seconds)
- **AND** use ISO 8601 UTC format

#### Scenario: OTP validation with expiration check
- **WHEN** validating an OTP code
- **THEN** the system SHALL reject codes where `now > expires_at`
- **AND** return error: `OTP expired, please request a new one`

### Requirement: Localized Email Content Events
The authentication system SHALL include locale information in email verification events to enable multi-language email template rendering by consumers.

#### Scenario: English locale email event
- **WHEN** user registers with English locale preference
- **THEN** the email verification event SHALL include `locale: "en"`
- **AND** the consumer SHALL render English email template

#### Scenario: Vietnamese locale email event
- **WHEN** user registers with Vietnamese locale preference
- **THEN** the email verification event SHALL include `locale: "vi"`
- **AND** the consumer SHALL render Vietnamese email template

#### Scenario: Default locale fallback
- **WHEN** user registration does not specify locale
- **THEN** the email verification event SHALL include `locale: "en"` (default)

## ADDED Requirements

### Requirement: Authentication Event Logging
The authentication system SHALL log all published authentication-related events with structured fields for audit trails and debugging.

#### Scenario: Email verification event logging
- **WHEN** publishing email verification event
- **THEN** the system SHALL log at INFO level:
  - `event_id`
  - `routing_key: email.verification.requested`
  - `user_id`
  - `recipient` (masked: show only first 2 chars and domain, e.g., `ab***@example.com`)
  - `locale`

#### Scenario: Failed event publish logging
- **WHEN** email verification event publishing fails
- **THEN** the system SHALL log at ERROR level:
  - `error` message
  - `user_id`
  - `recipient` (masked)
  - `retry_count` (if applicable)
