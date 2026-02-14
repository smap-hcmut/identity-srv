-- Auth Service Schema Migration
-- Description: Complete schema for OAuth2/JWT authentication with Google Workspace integration
-- Date: 2026-02-09
-- Schema: schema_identity

-- NOTE: Schema schema_identity should already exist and be owned by identity_prod
-- If not, ask DBA to run: CREATE SCHEMA schema_identity AUTHORIZATION identity_prod;

-- Set search path to schema_identity
SET search_path TO schema_identity;

-- Enable UUID extension (may require superuser)
-- If this fails, ask DBA to run: CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- USERS TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS schema_identity.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    avatar_url TEXT,
    role_hash VARCHAR(255) NOT NULL, -- Encrypted role (ADMIN, ANALYST, VIEWER)
    is_active BOOLEAN DEFAULT TRUE,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index for fast email lookup
CREATE INDEX IF NOT EXISTS idx_users_email ON schema_identity.users(email);

-- Index for active users
CREATE INDEX IF NOT EXISTS idx_users_is_active ON schema_identity.users(is_active);

-- ============================================================================
-- JWT KEYS TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS schema_identity.jwt_keys (
    kid VARCHAR(50) PRIMARY KEY, -- Key ID
    private_key TEXT NOT NULL, -- RSA private key (encrypted at rest)
    public_key TEXT NOT NULL, -- RSA public key
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, rotating, retired
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ,
    retired_at TIMESTAMPTZ
);

-- Index for active keys lookup
CREATE INDEX IF NOT EXISTS idx_jwt_keys_status ON schema_identity.jwt_keys(status);

-- Index for key rotation queries
CREATE INDEX IF NOT EXISTS idx_jwt_keys_created_at ON schema_identity.jwt_keys(created_at);

-- ============================================================================
-- AUDIT LOGS TABLE (if needed)
-- ============================================================================
CREATE TABLE IF NOT EXISTS schema_identity.audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES schema_identity.users(id),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100),
    resource_id VARCHAR(255),
    ip_address INET,
    user_agent TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index for user audit logs
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON schema_identity.audit_logs(user_id);

-- Index for timestamp queries
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON schema_identity.audit_logs(created_at);

-- Index for action queries
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON schema_identity.audit_logs(action);

-- ============================================================================
-- COMMENTS
-- ============================================================================
COMMENT ON SCHEMA schema_identity IS 'Identity and authentication service schema';

COMMENT ON TABLE schema_identity.users IS 'User accounts created via OAuth2 SSO';
COMMENT ON COLUMN schema_identity.users.email IS 'User email from OAuth provider (unique identifier)';
COMMENT ON COLUMN schema_identity.users.role_hash IS 'Encrypted user role: ADMIN (full access), ANALYST (create/analyze), VIEWER (read-only)';
COMMENT ON COLUMN schema_identity.users.is_active IS 'Account status - false for blocked users';
COMMENT ON COLUMN schema_identity.users.last_login_at IS 'Last successful login timestamp';

COMMENT ON TABLE schema_identity.jwt_keys IS 'RSA key pairs for JWT signing (supports rotation)';
COMMENT ON COLUMN schema_identity.jwt_keys.kid IS 'Key ID (used in JWT header)';
COMMENT ON COLUMN schema_identity.jwt_keys.status IS 'Key status: active (signing new tokens), rotating (grace period), retired (no longer used)';
COMMENT ON COLUMN schema_identity.jwt_keys.private_key IS 'RSA private key (encrypted with AES-256-GCM)';
COMMENT ON COLUMN schema_identity.jwt_keys.public_key IS 'RSA public key (exposed via JWKS endpoint)';

COMMENT ON TABLE schema_identity.audit_logs IS 'Audit trail for all authentication and authorization events';
