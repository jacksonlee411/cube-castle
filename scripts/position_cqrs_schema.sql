-- Position Management CQRS Schema Migration
-- This script implements the Position CQRS architecture with Outbox Pattern
-- Author: Claude Code SuperClaude
-- Date: 2025-08-03

-- ===========================================
-- 1. BACKUP EXISTING DATA
-- ===========================================

CREATE TABLE IF NOT EXISTS positions_backup AS 
SELECT * FROM positions WHERE 1=1;

CREATE TABLE IF NOT EXISTS position_occupancy_history_backup AS 
SELECT * FROM position_occupancy_history WHERE 1=1;

-- ===========================================
-- 2. DROP EXISTING TABLES (CAREFUL!)
-- ===========================================

-- Drop constraints first
DROP TABLE IF EXISTS position_occupancy_history CASCADE;
DROP TABLE IF EXISTS positions CASCADE;

-- ===========================================
-- 3. CREATE CQRS-OPTIMIZED POSITION TABLES
-- ===========================================

-- 职位主表 (Command Side - PostgreSQL)
CREATE TABLE positions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    position_type VARCHAR(50) NOT NULL CHECK (position_type IN ('REGULAR', 'TEMPORARY', 'CONTRACT', 'EXECUTIVE')),
    job_profile_id UUID NOT NULL,
    department_id UUID NOT NULL,
    manager_position_id UUID REFERENCES positions(id),
    status VARCHAR(50) NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'ACTIVE', 'FROZEN', 'PENDING_ELIMINATION')),
    budgeted_fte DECIMAL(3,2) NOT NULL DEFAULT 1.00 CHECK (budgeted_fte > 0 AND budgeted_fte <= 5.00),
    details JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- 业务约束
    CONSTRAINT positions_tenant_job_dept_unique UNIQUE (tenant_id, job_profile_id, department_id, manager_position_id)
);

-- 简化的职位分配表 (替代复杂的PositionOccupancyHistory)
CREATE TABLE position_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    position_id UUID NOT NULL REFERENCES positions(id) ON DELETE CASCADE,
    employee_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    is_current BOOLEAN NOT NULL DEFAULT true,
    fte DECIMAL(3,2) NOT NULL DEFAULT 1.00 CHECK (fte > 0 AND fte <= 5.00),
    assignment_type VARCHAR(50) NOT NULL DEFAULT 'PRIMARY' CHECK (assignment_type IN ('PRIMARY', 'SECONDARY', 'ACTING')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- 业务约束：同一时间每个员工只能有一个primary assignment
    CONSTRAINT position_assignments_unique UNIQUE (employee_id, position_id, start_date),
    -- 确保结束日期晚于开始日期
    CONSTRAINT position_assignments_date_check CHECK (end_date IS NULL OR end_date >= start_date)
);

-- 分配详情表 (复杂业务信息)
CREATE TABLE assignment_details (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    assignment_id UUID NOT NULL REFERENCES position_assignments(id) ON DELETE CASCADE,
    pay_grade_id UUID,
    reporting_manager_id UUID,
    location_id UUID,
    cost_center VARCHAR(50),
    effective_date DATE NOT NULL,
    reason TEXT,
    approval_status VARCHAR(50) DEFAULT 'PENDING' CHECK (approval_status IN ('PENDING', 'APPROVED', 'REJECTED')),
    approved_by UUID,
    approved_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 分配历史表 (审计跟踪)
CREATE TABLE assignment_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    assignment_id UUID NOT NULL REFERENCES position_assignments(id) ON DELETE CASCADE,
    change_type VARCHAR(50) NOT NULL CHECK (change_type IN ('CREATED', 'UPDATED', 'ENDED', 'TRANSFERRED')),
    old_values JSONB,
    new_values JSONB,
    changed_by UUID NOT NULL,
    change_reason TEXT,
    effective_date DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- ===========================================
-- 4. OUTBOX PATTERN IMPLEMENTATION
-- ===========================================

-- Outbox事件表 (确保事务安全)
CREATE TABLE outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    aggregate_id UUID NOT NULL,
    event_data JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'PROCESSED', 'FAILED')),
    attempt_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT
);

-- ===========================================
-- 5. PERFORMANCE INDEXES
-- ===========================================

-- Position表索引
CREATE INDEX idx_positions_tenant_id ON positions(tenant_id);
CREATE INDEX idx_positions_department ON positions(department_id);
CREATE INDEX idx_positions_manager ON positions(manager_position_id);
CREATE INDEX idx_positions_status ON positions(status);
CREATE INDEX idx_positions_type ON positions(position_type);
CREATE INDEX idx_positions_updated ON positions(updated_at);

-- Position Assignments表索引
CREATE INDEX idx_assignments_tenant_id ON position_assignments(tenant_id);
CREATE INDEX idx_assignments_position ON position_assignments(position_id);
CREATE INDEX idx_assignments_employee ON position_assignments(employee_id);
CREATE INDEX idx_assignments_current ON position_assignments(is_current) WHERE is_current = true;
CREATE INDEX idx_assignments_type ON position_assignments(assignment_type);
CREATE INDEX idx_assignments_date_range ON position_assignments(start_date, end_date);

-- Assignment Details表索引
CREATE INDEX idx_assignment_details_assignment ON assignment_details(assignment_id);
CREATE INDEX idx_assignment_details_effective ON assignment_details(effective_date);
CREATE INDEX idx_assignment_details_status ON assignment_details(approval_status);
CREATE INDEX idx_assignment_details_approver ON assignment_details(approved_by);

-- Assignment History表索引
CREATE INDEX idx_assignment_history_assignment ON assignment_history(assignment_id);
CREATE INDEX idx_assignment_history_type ON assignment_history(change_type);
CREATE INDEX idx_assignment_history_date ON assignment_history(effective_date);
CREATE INDEX idx_assignment_history_changed_by ON assignment_history(changed_by);

-- Outbox Events表索引
CREATE INDEX idx_outbox_events_tenant ON outbox_events(tenant_id);
CREATE INDEX idx_outbox_events_status ON outbox_events(status);
CREATE INDEX idx_outbox_events_type ON outbox_events(event_type);
CREATE INDEX idx_outbox_events_aggregate ON outbox_events(aggregate_id);
CREATE INDEX idx_outbox_events_created ON outbox_events(created_at);

-- ===========================================
-- 6. BUSINESS CONSTRAINTS & TRIGGERS
-- ===========================================

-- 确保每个员工只能有一个当前的主要职位分配
CREATE UNIQUE INDEX idx_assignments_employee_primary_current 
ON position_assignments(employee_id) 
WHERE is_current = true AND assignment_type = 'PRIMARY';

-- 自动更新updated_at字段的触发器
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_positions_updated_at 
    BEFORE UPDATE ON positions 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_assignments_updated_at 
    BEFORE UPDATE ON position_assignments 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_assignment_details_updated_at 
    BEFORE UPDATE ON assignment_details 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ===========================================
-- 7. DATA MIGRATION (if needed)
-- ===========================================

-- 从备份表迁移数据的示例 (根据实际需要调整)
/*
INSERT INTO positions (id, tenant_id, position_type, job_profile_id, department_id, status, budgeted_fte, created_at, updated_at)
SELECT 
    id, 
    tenant_id, 
    'REGULAR' as position_type,
    job_profile_id,
    department_id,
    CASE 
        WHEN is_active = true THEN 'ACTIVE'
        ELSE 'FROZEN'
    END as status,
    1.00 as budgeted_fte,
    created_at,
    updated_at
FROM positions_backup
WHERE tenant_id IS NOT NULL;
*/

-- ===========================================
-- 8. VALIDATION QUERIES
-- ===========================================

-- 验证表创建
SELECT 
    schemaname,
    tablename,
    tableowner
FROM pg_tables 
WHERE tablename IN ('positions', 'position_assignments', 'assignment_details', 'assignment_history', 'outbox_events')
ORDER BY tablename;

-- 验证索引创建
SELECT 
    indexname,
    tablename,
    indexdef
FROM pg_indexes 
WHERE tablename IN ('positions', 'position_assignments', 'assignment_details', 'assignment_history', 'outbox_events')
ORDER BY tablename, indexname;

-- 验证约束
SELECT
    conname,
    conrelid::regclass AS table_name,
    pg_get_constraintdef(c.oid) AS constraint_definition
FROM pg_constraint c
JOIN pg_namespace n ON n.oid = c.connamespace
WHERE conrelid::regclass::text IN ('positions', 'position_assignments', 'assignment_details', 'assignment_history', 'outbox_events')
ORDER BY table_name, conname;

-- ===========================================
-- MIGRATION COMPLETE
-- ===========================================

-- 添加注释说明
COMMENT ON TABLE positions IS 'CQRS职位管理主表 - Command Side (PostgreSQL)';
COMMENT ON TABLE position_assignments IS '简化的职位分配表 - 替代复杂的PositionOccupancyHistory';
COMMENT ON TABLE assignment_details IS '分配详情表 - 复杂业务信息';
COMMENT ON TABLE assignment_history IS '分配历史表 - 审计跟踪';
COMMENT ON TABLE outbox_events IS 'Outbox模式事件表 - 确保事务安全';

COMMENT ON COLUMN positions.status IS '职位状态: DRAFT(草稿), ACTIVE(活跃), FROZEN(冻结), PENDING_ELIMINATION(待删除)';
COMMENT ON COLUMN position_assignments.assignment_type IS '分配类型: PRIMARY(主要), SECONDARY(次要), ACTING(代理)';
COMMENT ON COLUMN outbox_events.status IS '事件状态: PENDING(待处理), PROCESSED(已处理), FAILED(失败)';

-- 输出完成信息
SELECT 'Position CQRS Schema Migration Completed Successfully!' AS status;