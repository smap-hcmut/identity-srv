-- Drop unused audit_logs table.
-- The audit domain (internal/audit, Kafka topic audit.events) was a
-- never-completed skeleton: no producer, no UI integration, OAuth login
-- did not record entries synchronously. Removed in the v3 minimization pass.
-- Date: 2026-06-06

SET search_path TO identity;

DROP TABLE IF EXISTS identity.audit_logs CASCADE;
