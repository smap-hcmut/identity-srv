# CORS Middleware Specification

## ADDED Requirements

### Requirement: Environment-Based CORS Configuration
The system SHALL provide different CORS configurations based on the deployment environment (production, staging, development) to balance security and developer accessibility.

#### Scenario: Production environment uses strict origin list
- **GIVEN** the ENV environment variable is set to "production"
- **WHEN** the CORS configuration is initialized
- **THEN** only explicitly allowed production origins SHALL be permitted
- **AND** dynamic origin validation SHALL NOT be enabled
- **AND** the allowed origins list SHALL include only production domains (e.g., "https://smap.tantai.dev", "https://smap-api.tantai.dev")

#### Scenario: Development environment uses dynamic validation
- **GIVEN** the ENV environment variable is set to "dev" or "staging"
- **WHEN** the CORS configuration is initialized
- **THEN** dynamic origin validation SHALL be enabled via AllowOriginFunc
- **AND** production domains SHALL be allowed
- **AND** localhost origins SHALL be allowed
- **AND** private subnet origins SHALL be allowed

#### Scenario: Missing environment variable defaults to production
- **GIVEN** the ENV environment variable is not set or empty
- **WHEN** the CORS configuration is initialized
- **THEN** the system SHALL default to production mode for security
- **AND** only strict origin list SHALL be used

### Requirement: Private Subnet Origin Validation
The system SHALL validate origins against configured private subnet CIDR ranges in non-production environments to enable developer access via VPN.

#### Scenario: Accept origin from allowed private subnet
- **GIVEN** the environment is non-production (dev or staging)
- **AND** private subnets are configured as ["172.16.21.0/24", "172.16.19.0/24", "192.168.1.0/24"]
- **WHEN** a request arrives with origin "http://172.16.21.50:3000"
- **THEN** the origin SHALL be validated as allowed
- **AND** the Access-Control-Allow-Origin header SHALL be set to "http://172.16.21.50:3000"

#### Scenario: Reject origin from non-configured subnet
- **GIVEN** the environment is non-production
- **AND** private subnets are configured as ["172.16.21.0/24", "172.16.19.0/24", "192.168.1.0/24"]
- **WHEN** a request arrives with origin "http://10.0.0.50:3000"
- **THEN** the origin SHALL NOT be allowed
- **AND** the Access-Control-Allow-Origin header SHALL NOT be set

#### Scenario: Private subnet validation disabled in production
- **GIVEN** the environment is production
- **WHEN** a request arrives with origin from a private subnet
- **THEN** the origin SHALL be rejected
- **AND** only explicitly configured production origins SHALL be allowed

### Requirement: Localhost Origin Validation
The system SHALL automatically allow localhost and 127.0.0.1 origins with any port number in non-production environments to support local development.

#### Scenario: Allow localhost with HTTP on any port
- **GIVEN** the environment is non-production
- **WHEN** a request arrives with origin "http://localhost:3000" or "http://localhost:5173" or "http://localhost"
- **THEN** the origin SHALL be allowed
- **AND** the Access-Control-Allow-Origin header SHALL be set appropriately

#### Scenario: Allow 127.0.0.1 with HTTP on any port
- **GIVEN** the environment is non-production
- **WHEN** a request arrives with origin "http://127.0.0.1:3000" or "http://127.0.0.1:8080"
- **THEN** the origin SHALL be allowed
- **AND** the Access-Control-Allow-Origin header SHALL be set appropriately

#### Scenario: Allow localhost with HTTPS
- **GIVEN** the environment is non-production
- **WHEN** a request arrives with origin "https://localhost:3000" or "https://127.0.0.1:3000"
- **THEN** the origin SHALL be allowed
- **AND** the Access-Control-Allow-Origin header SHALL be set appropriately

### Requirement: CORS Credentials Support
The system SHALL maintain support for credentials (HttpOnly cookies) with appropriate origin restrictions as required by browser security policies.

#### Scenario: Credentials supported with specific origins
- **GIVEN** any environment (production or non-production)
- **WHEN** CORS configuration is applied
- **THEN** AllowCredentials SHALL be set to true
- **AND** Access-Control-Allow-Credentials header SHALL be set to "true"
- **AND** wildcard origins ("*") SHALL NOT be used

#### Scenario: Preflight request with credentials
- **GIVEN** any environment
- **WHEN** a preflight OPTIONS request is received
- **AND** the origin is allowed
- **THEN** Access-Control-Allow-Credentials SHALL be set to "true"
- **AND** Access-Control-Allow-Origin SHALL be set to the specific requesting origin
- **AND** the response status SHALL be 204 No Content

### Requirement: CIDR Subnet Parsing and Validation
The system SHALL provide functions to parse CIDR notation and validate IP addresses against subnet ranges for dynamic origin validation.

#### Scenario: Parse valid CIDR and check IP membership
- **GIVEN** a CIDR range "172.16.21.0/24"
- **WHEN** checking if IP "172.16.21.50" is in the range
- **THEN** the validation SHALL return true

#### Scenario: IP outside CIDR range
- **GIVEN** a CIDR range "172.16.21.0/24"
- **WHEN** checking if IP "172.16.22.50" is in the range
- **THEN** the validation SHALL return false

#### Scenario: Handle invalid CIDR format
- **GIVEN** an invalid CIDR string "not-a-cidr"
- **WHEN** attempting to parse the CIDR
- **THEN** the validation SHALL fail gracefully
- **AND** the IP SHALL be rejected

#### Scenario: Extract IP from origin URL
- **GIVEN** an origin URL "http://172.16.21.50:3000"
- **WHEN** extracting the IP address for validation
- **THEN** the extracted IP SHALL be "172.16.21.50"
- **AND** port and scheme SHALL be ignored for CIDR matching

### Requirement: Configurable Private Subnets
The system SHALL define private subnet CIDR ranges in code configuration to support different network topologies without code changes.

#### Scenario: Default private subnets configured
- **GIVEN** the system is initialized
- **WHEN** reading the private subnet configuration
- **THEN** the default subnets SHALL include:
  - "172.16.21.0/24" (K8s cluster subnet)
  - "172.16.19.0/24" (Private network 1)
  - "192.168.1.0/24" (Private network 2)

#### Scenario: Production domains configured
- **GIVEN** the system is initialized
- **WHEN** reading the production origin configuration
- **THEN** the production origins SHALL include:
  - "https://smap.tantai.dev"
  - "https://smap-api.tantai.dev"
