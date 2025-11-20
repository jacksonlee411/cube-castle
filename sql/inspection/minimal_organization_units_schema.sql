-- Minimal schema for temporal consistency checks.
-- Only create essential table/columns required by sql/inspection/check_temporal_consistency.sql.
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS organization_units (
  record_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  code TEXT NOT NULL,
  parent_code TEXT,
  name TEXT NOT NULL,
  unit_type TEXT NOT NULL,
  status TEXT NOT NULL,
  level INT,
  path TEXT,
  sort_order INT,
  description TEXT,
  effective_date DATE NOT NULL,
  end_date DATE,
  is_current BOOLEAN NOT NULL DEFAULT false,
  is_temporal BOOLEAN NOT NULL DEFAULT false,
  change_reason TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMP
);
