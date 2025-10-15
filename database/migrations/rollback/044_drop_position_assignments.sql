-- Rollback script for 044_create_position_assignments.sql

BEGIN;

DROP TABLE IF EXISTS position_assignments CASCADE;

COMMIT;
