-- Auth Service Schema Migration
-- Description: Complete schema for OAuth2/JWT authentication with Google Workspace integration
-- Date: 2026-02-09

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- USERS TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS users (
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
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Index for active users
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);

-- ============================================================================
-- JWT KEYS TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS jwt_keys (
    kid VARCHAR(50) PRIMARY KEY, -- Key ID
    private_key TEXT NOT NULL, -- RSA private key (encrypted at rest)
    public_key TEXT NOT NULL, -- RSA public key
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, rotating, retired
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ,
    retired_at TIMESTAMPTZ
);

-- Index for active keys lookup
CREATE INDEX IF NOT EXISTS idx_jwt_keys_status ON jwt_keys(status);

-- Index for key rotation queries
CREATE INDEX IF NOT EXISTS idx_jwt_keys_created_at ON jwt_keys(created_at);

-- ============================================================================
-- COMMENTS
-- ============================================================================
COMMENT ON TABLE users IS 'User accounts created via OAuth2 SSO';
COMMENT ON COLUMN users.email IS 'User email from OAuth provider (unique identifier)';
COMMENT ON COLUMN users.role_hash IS 'Encrypted user role: ADMIN (full access), ANALYST (create/analyze), VIEWER (read-only)';
COMMENT ON COLUMN users.is_active IS 'Account status - false for blocked users';
COMMENT ON COLUMN users.last_login_at IS 'Last successful login timestamp';

COMMENT ON TABLE jwt_keys IS 'RSA key pairs for JWT signing (supports rotation)';
COMMENT ON COLUMN jwt_keys.kid IS 'Key ID (used in JWT header)';
COMMENT ON COLUMN jwt_keys.status IS 'Key status: active (signing new tokens), rotating (grace period), retired (no longer used)';
COMMENT ON COLUMN jwt_keys.private_key IS 'RSA private key (encrypted with AES-256-GCM)';
COMMENT ON COLUMN jwt_keys.public_key IS 'RSA public key (exposed via JWKS endpoint)';

-- ============================================================================
-- INITIAL DATA (Optional)
-- ============================================================================
-- Insert default admin user (will be updated on first OAuth login)
-- INSERT INTO users (email, name, role, is_active) 
-- VALUES ('admin@vinfast.com', 'System Admin', 'ADMIN', true)
-- ON CONFLICT (email) DO NOTHING;
