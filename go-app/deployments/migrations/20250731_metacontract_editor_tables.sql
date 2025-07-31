-- Create metacontract editor tables
-- +goose Up

-- Projects table
CREATE TABLE IF NOT EXISTS metacontract_editor_projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    content TEXT NOT NULL DEFAULT '',
    version VARCHAR(50) NOT NULL DEFAULT '1.0.0',
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'compiling', 'valid', 'error', 'published')),
    tenant_id UUID NOT NULL,
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_compiled TIMESTAMP WITH TIME ZONE,
    compile_error TEXT
);

-- Sessions table
CREATE TABLE IF NOT EXISTS metacontract_editor_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES metacontract_editor_projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_seen TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    active BOOLEAN NOT NULL DEFAULT true
);

-- Templates table
CREATE TABLE IF NOT EXISTS metacontract_editor_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL DEFAULT 'general',
    content TEXT NOT NULL,
    tags TEXT[] DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- User settings table
CREATE TABLE IF NOT EXISTS metacontract_editor_settings (
    user_id UUID PRIMARY KEY,
    theme VARCHAR(50) NOT NULL DEFAULT 'light',
    font_size INTEGER NOT NULL DEFAULT 14,
    auto_save BOOLEAN NOT NULL DEFAULT true,
    auto_compile BOOLEAN NOT NULL DEFAULT false,
    key_bindings VARCHAR(50) NOT NULL DEFAULT 'default',
    settings JSONB DEFAULT '{}',
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_metacontract_projects_tenant_id ON metacontract_editor_projects(tenant_id);
CREATE INDEX IF NOT EXISTS idx_metacontract_projects_created_by ON metacontract_editor_projects(created_by);
CREATE INDEX IF NOT EXISTS idx_metacontract_projects_status ON metacontract_editor_projects(status);
CREATE INDEX IF NOT EXISTS idx_metacontract_projects_updated_at ON metacontract_editor_projects(updated_at);

CREATE INDEX IF NOT EXISTS idx_metacontract_sessions_project_id ON metacontract_editor_sessions(project_id);
CREATE INDEX IF NOT EXISTS idx_metacontract_sessions_user_id ON metacontract_editor_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_metacontract_sessions_active ON metacontract_editor_sessions(active);

CREATE INDEX IF NOT EXISTS idx_metacontract_templates_category ON metacontract_editor_templates(category);
CREATE INDEX IF NOT EXISTS idx_metacontract_templates_tags ON metacontract_editor_templates USING GIN(tags);

-- Insert some sample templates
INSERT INTO metacontract_editor_templates (name, description, category, content, tags) VALUES 
(
    'Employee Management Template',
    'Basic template for employee management meta-contract',
    'hr',
    E'# Employee Management Meta-Contract\n\nversion: "1.0.0"\nname: "employee_management"\ndescription: "Employee management system"\n\nentities:\n  Employee:\n    fields:\n      - name: id\n        type: UUID\n        required: true\n        primary_key: true\n      - name: first_name\n        type: String\n        required: true\n      - name: last_name\n        type: String\n        required: true\n      - name: email\n        type: String\n        required: true\n        unique: true\n      - name: hire_date\n        type: Date\n        required: true\n\nworkflows:\n  employee_onboarding:\n    description: "Employee onboarding process"\n    steps:\n      - name: create_employee\n        action: create\n        entity: Employee\n      - name: send_welcome_email\n        action: notify\n        template: welcome_email',
    ARRAY['hr', 'employee', 'management', 'basic']
),
(
    'Organization Structure Template',
    'Template for defining organizational hierarchy',
    'organization',
    E'# Organization Structure Meta-Contract\n\nversion: "1.0.0"\nname: "organization_structure"\ndescription: "Organization hierarchy management"\n\nentities:\n  OrganizationUnit:\n    fields:\n      - name: id\n        type: UUID\n        required: true\n        primary_key: true\n      - name: name\n        type: String\n        required: true\n      - name: code\n        type: String\n        required: true\n        unique: true\n      - name: parent_id\n        type: UUID\n        required: false\n        references: OrganizationUnit.id\n      - name: level\n        type: Integer\n        required: true\n\nworkflows:\n  restructure_organization:\n    description: "Organization restructuring process"\n    steps:\n      - name: validate_hierarchy\n        action: validate\n        entity: OrganizationUnit\n      - name: update_reporting_structure\n        action: update\n        entity: OrganizationUnit',
    ARRAY['organization', 'hierarchy', 'structure', 'management']
),
(
    'Position Management Template',
    'Template for position and job role management',
    'hr',
    E'# Position Management Meta-Contract\n\nversion: "1.0.0"\nname: "position_management"\ndescription: "Position and job role management"\n\nentities:\n  Position:\n    fields:\n      - name: id\n        type: UUID\n        required: true\n        primary_key: true\n      - name: title\n        type: String\n        required: true\n      - name: department\n        type: String\n        required: true\n      - name: level\n        type: String\n        required: true\n      - name: status\n        type: String\n        required: true\n        enum: [\"active\", \"inactive\", \"draft\"]\n\nworkflows:\n  position_creation:\n    description: "Position creation and approval process"\n    steps:\n      - name: create_position\n        action: create\n        entity: Position\n      - name: require_approval\n        action: approve\n        role: hr_manager',
    ARRAY['position', 'job', 'role', 'hr', 'management']
) ON CONFLICT DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS metacontract_editor_settings;
DROP TABLE IF EXISTS metacontract_editor_templates;
DROP TABLE IF EXISTS metacontract_editor_sessions;
DROP TABLE IF EXISTS metacontract_editor_projects;