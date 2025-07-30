-- 003_metacontract_editor.sql
-- Meta-contract editor tables

-- Projects table
CREATE TABLE IF NOT EXISTS metacontract_editor_projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    content TEXT NOT NULL,
    version VARCHAR(50) NOT NULL DEFAULT '0.1.0',
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    tenant_id UUID NOT NULL,
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_compiled TIMESTAMP WITH TIME ZONE,
    compile_error TEXT,
    
    CONSTRAINT check_status CHECK (status IN ('draft', 'compiling', 'valid', 'error', 'published'))
);

-- Add RLS policy for tenant isolation
ALTER TABLE metacontract_editor_projects ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_projects ON metacontract_editor_projects 
FOR ALL USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

-- Indexes
CREATE INDEX idx_projects_tenant_id ON metacontract_editor_projects(tenant_id);
CREATE INDEX idx_projects_created_by ON metacontract_editor_projects(created_by);
CREATE INDEX idx_projects_updated_at ON metacontract_editor_projects(updated_at DESC);
CREATE INDEX idx_projects_status ON metacontract_editor_projects(status);

-- Sessions table for tracking active editing sessions
CREATE TABLE IF NOT EXISTS metacontract_editor_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES metacontract_editor_projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_seen TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    active BOOLEAN NOT NULL DEFAULT true
);

-- Indexes for sessions
CREATE INDEX idx_sessions_project_id ON metacontract_editor_sessions(project_id);
CREATE INDEX idx_sessions_user_id ON metacontract_editor_sessions(user_id);
CREATE INDEX idx_sessions_active ON metacontract_editor_sessions(active);

-- Templates table for project templates
CREATE TABLE IF NOT EXISTS metacontract_editor_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100) NOT NULL,
    content TEXT NOT NULL,
    tags TEXT[] DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for templates
CREATE INDEX idx_templates_category ON metacontract_editor_templates(category);
CREATE INDEX idx_templates_tags ON metacontract_editor_templates USING GIN(tags);

-- User settings table for editor preferences
CREATE TABLE IF NOT EXISTS metacontract_editor_settings (
    user_id UUID PRIMARY KEY,
    theme VARCHAR(50) NOT NULL DEFAULT 'vs-dark',
    font_size INTEGER NOT NULL DEFAULT 14,
    auto_save BOOLEAN NOT NULL DEFAULT true,
    auto_compile BOOLEAN NOT NULL DEFAULT true,
    key_bindings VARCHAR(50) NOT NULL DEFAULT 'default',
    settings JSONB DEFAULT '{}',
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Insert some default templates
INSERT INTO metacontract_editor_templates (name, description, category, content) VALUES
('Basic Employee', 'A basic employee entity template', 'CoreHR', 
'specification_version: "1.0"
api_id: "550e8400-e29b-41d4-a716-446655440000"
namespace: "corehr"
resource_name: "employee"
version: "1.0.0"

data_structure:
  primary_key: "id"
  data_classification: "pii"
  fields:
    - name: "id"
      type: "uuid"
      required: true
      unique: true
      data_classification: "public"
    - name: "first_name"
      type: "string"
      required: true
      data_classification: "pii"
    - name: "last_name"
      type: "string"
      required: true
      data_classification: "pii"
    - name: "email"
      type: "string"
      required: true
      unique: true
      data_classification: "pii"
    - name: "employee_id"
      type: "string"
      required: true
      unique: true
      data_classification: "internal"

security_model:
  tenant_isolation: true
  access_control: "rbac"
  data_classification: "pii"
  compliance_tags: ["gdpr", "ccpa"]

temporal_behavior:
  temporality_paradigm: "event_sourced"
  state_transition_model: "status_based"
  history_retention: "indefinite"
  event_driven: true

api_behavior:
  rest_enabled: true
  graphql_enabled: true
  events_enabled: true

relationships: []'),

('Organization Unit', 'A basic organization unit template', 'CoreHR',
'specification_version: "1.0"
api_id: "550e8400-e29b-41d4-a716-446655440001"
namespace: "corehr"
resource_name: "organization_unit"
version: "1.0.0"

data_structure:
  primary_key: "id"
  data_classification: "internal"
  fields:
    - name: "id"
      type: "uuid"
      required: true
      unique: true
      data_classification: "public"
    - name: "name"
      type: "string"
      required: true
      data_classification: "internal"
    - name: "description"
      type: "text"
      required: false
      data_classification: "internal"
    - name: "parent_id"
      type: "uuid"
      required: false
      data_classification: "internal"
    - name: "level"
      type: "integer"
      required: true
      data_classification: "internal"

security_model:
  tenant_isolation: true
  access_control: "rbac"
  data_classification: "internal"
  compliance_tags: []

temporal_behavior:
  temporality_paradigm: "snapshot"
  state_transition_model: "lifecycle_based"
  history_retention: "7_years"
  event_driven: true

api_behavior:
  rest_enabled: true
  graphql_enabled: true
  events_enabled: true

relationships:
  - name: "parent"
    type: "one-to-one"
    target_entity: "organization_unit"
    cardinality: "0..1"
    is_optional: true
  - name: "children"
    type: "one-to-many"
    target_entity: "organization_unit"
    cardinality: "*"
    is_optional: true');

-- Add updated_at trigger function if it doesn't exist
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Add triggers for updated_at
CREATE TRIGGER update_projects_updated_at BEFORE UPDATE ON metacontract_editor_projects 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_templates_updated_at BEFORE UPDATE ON metacontract_editor_templates 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_settings_updated_at BEFORE UPDATE ON metacontract_editor_settings 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();