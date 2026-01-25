-- Migration: Add role_hash column to users table
-- Description: Add encrypted role field to support role-based access control (RBAC)
-- Date: 2025-11-18

-- Add role_hash column
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS role_hash TEXT;

-- Add index for role_hash (for efficient role-based queries)
CREATE INDEX IF NOT EXISTS idx_users_role_hash ON users(role_hash);

-- Add comment to explain the column
COMMENT ON COLUMN users.role_hash IS 'Encrypted role value (USER or ADMIN) using SHA256 hash';

-- Optional: Set default role for existing users without role
-- Uncomment if you want to set default USER role for all existing users
-- UPDATE users 
-- SET role_hash = 'encrypted_user_role_hash' 
-- WHERE role_hash IS NULL;

