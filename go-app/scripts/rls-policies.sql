-- PostgreSQL Row Level Security (RLS) 多租户隔离策略
-- 本文件包含为Cube Castle项目实现多租户数据隔离的RLS策略

-- ===========================================
-- 启用RLS并创建基础策略
-- ===========================================

-- 启用所有核心表的RLS
ALTER TABLE corehr.employees ENABLE ROW LEVEL SECURITY;
ALTER TABLE corehr.organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE corehr.positions ENABLE ROW LEVEL SECURITY;
ALTER TABLE corehr.employee_positions ENABLE ROW LEVEL SECURITY;
ALTER TABLE corehr.organization_hierarchies ENABLE ROW LEVEL SECURITY;
ALTER TABLE workflow.workflow_instances ENABLE ROW LEVEL SECURITY;
ALTER TABLE workflow.workflow_activities ENABLE ROW LEVEL SECURITY;
-- Outbox table RLS removed in Phase 4

-- ===========================================
-- 租户上下文管理函数
-- ===========================================

-- 创建或替换设置当前租户上下文的函数
CREATE OR REPLACE FUNCTION set_current_tenant_id(tenant_uuid uuid)
RETURNS void AS $$
BEGIN
    -- 设置会话级别的租户ID
    PERFORM set_config('app.current_tenant_id', tenant_uuid::text, false);
    
    -- 记录租户上下文设置（用于审计）
    INSERT INTO system.tenant_access_log (
        tenant_id, 
        access_time, 
        session_id,
        application_name
    ) VALUES (
        tenant_uuid,
        NOW(),
        pg_backend_pid(),
        current_setting('application_name', true)
    ) ON CONFLICT (session_id) DO UPDATE SET
        tenant_id = EXCLUDED.tenant_id,
        access_time = EXCLUDED.access_time;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 获取当前租户ID的函数
CREATE OR REPLACE FUNCTION get_current_tenant_id()
RETURNS uuid AS $$
BEGIN
    RETURN current_setting('app.current_tenant_id', true)::uuid;
EXCEPTION
    WHEN OTHERS THEN
        RETURN NULL;
END;
$$ LANGUAGE plpgsql STABLE;

-- 清除租户上下文的函数
CREATE OR REPLACE FUNCTION clear_current_tenant_id()
RETURNS void AS $$
BEGIN
    PERFORM set_config('app.current_tenant_id', '', false);
END;
$$ LANGUAGE plpgsql;

-- ===========================================
-- 管理员和系统角色管理
-- ===========================================

-- 创建检查是否为超级管理员的函数
CREATE OR REPLACE FUNCTION is_super_admin()
RETURNS boolean AS $$
BEGIN
    -- 检查当前用户是否具有超级管理员权限
    RETURN EXISTS (
        SELECT 1 
        FROM system.user_roles ur
        JOIN system.roles r ON ur.role_id = r.id
        WHERE ur.user_id = current_setting('app.current_user_id', true)::uuid
        AND r.name = 'super_admin'
        AND ur.is_active = true
    );
EXCEPTION
    WHEN OTHERS THEN
        RETURN false;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- 创建检查是否为租户管理员的函数
CREATE OR REPLACE FUNCTION is_tenant_admin()
RETURNS boolean AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 
        FROM system.user_roles ur
        JOIN system.roles r ON ur.role_id = r.id
        WHERE ur.user_id = current_setting('app.current_user_id', true)::uuid
        AND ur.tenant_id = get_current_tenant_id()
        AND r.name IN ('tenant_admin', 'admin')
        AND ur.is_active = true
    );
EXCEPTION
    WHEN OTHERS THEN
        RETURN false;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- ===========================================
-- CoreHR 表的RLS策略
-- ===========================================

-- 员工表策略
CREATE POLICY employees_tenant_isolation ON corehr.employees
    USING (
        tenant_id = get_current_tenant_id() 
        OR is_super_admin()
    );

CREATE POLICY employees_select_policy ON corehr.employees
    FOR SELECT
    USING (
        tenant_id = get_current_tenant_id()
        OR is_super_admin()
    );

CREATE POLICY employees_insert_policy ON corehr.employees
    FOR INSERT
    WITH CHECK (
        tenant_id = get_current_tenant_id()
        OR is_super_admin()
    );

CREATE POLICY employees_update_policy ON corehr.employees
    FOR UPDATE
    USING (
        tenant_id = get_current_tenant_id()
        OR is_super_admin()
    )
    WITH CHECK (
        tenant_id = get_current_tenant_id()
        OR is_super_admin()
    );

CREATE POLICY employees_delete_policy ON corehr.employees
    FOR DELETE
    USING (
        (tenant_id = get_current_tenant_id() AND is_tenant_admin())
        OR is_super_admin()
    );

-- 组织架构表策略
CREATE POLICY organizations_tenant_isolation ON corehr.organizations
    USING (
        tenant_id = get_current_tenant_id()
        OR is_super_admin()
    );

CREATE POLICY organizations_select_policy ON corehr.organizations
    FOR SELECT
    USING (
        tenant_id = get_current_tenant_id()
        OR is_super_admin()
    );

CREATE POLICY organizations_insert_policy ON corehr.organizations
    FOR INSERT
    WITH CHECK (
        tenant_id = get_current_tenant_id()
        OR is_super_admin()
    );

CREATE POLICY organizations_update_policy ON corehr.organizations
    FOR UPDATE
    USING (
        tenant_id = get_current_tenant_id()
        OR is_super_admin()
    )
    WITH CHECK (
        tenant_id = get_current_tenant_id()
        OR is_super_admin()
    );

CREATE POLICY organizations_delete_policy ON corehr.organizations
    FOR DELETE
    USING (
        (tenant_id = get_current_tenant_id() AND is_tenant_admin())
        OR is_super_admin()
    );

-- 职位表策略
CREATE POLICY positions_tenant_isolation ON corehr.positions
    USING (
        tenant_id = get_current_tenant_id()
        OR is_super_admin()
    );

-- 员工职位关联表策略
CREATE POLICY employee_positions_tenant_isolation ON corehr.employee_positions
    USING (
        EXISTS (
            SELECT 1 FROM corehr.employees e 
            WHERE e.id = employee_positions.employee_id 
            AND e.tenant_id = get_current_tenant_id()
        )
        OR is_super_admin()
    );

-- 组织层级表策略
CREATE POLICY organization_hierarchies_tenant_isolation ON corehr.organization_hierarchies
    USING (
        EXISTS (
            SELECT 1 FROM corehr.organizations o 
            WHERE o.id = organization_hierarchies.parent_id 
            AND o.tenant_id = get_current_tenant_id()
        )
        OR is_super_admin()
    );

-- ===========================================
-- 工作流表的RLS策略
-- ===========================================

-- 工作流实例表策略
CREATE POLICY workflow_instances_tenant_isolation ON workflow.workflow_instances
    USING (
        tenant_id = get_current_tenant_id()
        OR is_super_admin()
    );

CREATE POLICY workflow_instances_insert_policy ON workflow.workflow_instances
    FOR INSERT
    WITH CHECK (
        tenant_id = get_current_tenant_id()
        OR is_super_admin()
    );

-- 工作流活动表策略
CREATE POLICY workflow_activities_tenant_isolation ON workflow.workflow_activities
    USING (
        EXISTS (
            SELECT 1 FROM workflow.workflow_instances wi 
            WHERE wi.id = workflow_activities.workflow_instance_id 
            AND wi.tenant_id = get_current_tenant_id()
        )
        OR is_super_admin()
    );

-- Outbox RLS policies removed in Phase 4

-- ===========================================
-- 性能优化索引
-- ===========================================

-- 为RLS策略创建优化索引
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_employees_tenant_id_active 
ON corehr.employees (tenant_id) WHERE status = 'active';

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_organizations_tenant_id_active 
ON corehr.organizations (tenant_id) WHERE status = 'active';

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_positions_tenant_id 
ON corehr.positions (tenant_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_workflow_instances_tenant_id 
ON workflow.workflow_instances (tenant_id);

-- Outbox index removed in Phase 4

-- 复合索引用于常见查询模式
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_employees_tenant_number 
ON corehr.employees (tenant_id, employee_number);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_organizations_tenant_code 
ON corehr.organizations (tenant_id, code);

-- ===========================================
-- 审计和监控
-- ===========================================

-- 创建租户访问日志表
CREATE TABLE IF NOT EXISTS system.tenant_access_log (
    id SERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    access_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    session_id INTEGER NOT NULL,
    application_name TEXT,
    UNIQUE(session_id)
);

-- 创建RLS策略违规日志表
CREATE TABLE IF NOT EXISTS system.rls_violation_log (
    id SERIAL PRIMARY KEY,
    violation_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    session_id INTEGER,
    attempted_tenant_id UUID,
    current_tenant_id UUID,
    table_name TEXT,
    operation TEXT,
    user_id UUID,
    details JSONB
);

-- 创建监控RLS违规的函数
CREATE OR REPLACE FUNCTION log_rls_violation(
    p_attempted_tenant_id UUID,
    p_table_name TEXT,
    p_operation TEXT,
    p_details JSONB DEFAULT NULL
)
RETURNS void AS $$
BEGIN
    INSERT INTO system.rls_violation_log (
        attempted_tenant_id,
        current_tenant_id,
        table_name,
        operation,
        user_id,
        details,
        session_id
    ) VALUES (
        p_attempted_tenant_id,
        get_current_tenant_id(),
        p_table_name,
        p_operation,
        current_setting('app.current_user_id', true)::uuid,
        p_details,
        pg_backend_pid()
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- ===========================================
-- RLS策略测试和验证函数
-- ===========================================

-- 创建测试RLS策略的函数
CREATE OR REPLACE FUNCTION test_rls_policies()
RETURNS TABLE (
    test_name TEXT,
    result BOOLEAN,
    message TEXT
) AS $$
DECLARE
    test_tenant_1 UUID := gen_random_uuid();
    test_tenant_2 UUID := gen_random_uuid();
    test_emp_id UUID;
    test_org_id UUID;
    record_count INTEGER;
BEGIN
    -- 测试准备：创建测试数据
    PERFORM set_current_tenant_id(test_tenant_1);
    
    -- 插入测试员工
    INSERT INTO corehr.employees (id, tenant_id, employee_number, first_name, last_name, email, status)
    VALUES (gen_random_uuid(), test_tenant_1, 'TEST001', 'Test', 'User', 'test@example.com', 'active')
    RETURNING id INTO test_emp_id;
    
    -- 插入测试组织
    INSERT INTO corehr.organizations (id, tenant_id, name, code, level, status)
    VALUES (gen_random_uuid(), test_tenant_1, 'Test Org', 'TESTORG', 1, 'active')
    RETURNING id INTO test_org_id;
    
    -- 测试1：同租户访问应该成功
    SELECT COUNT(*) INTO record_count FROM corehr.employees WHERE tenant_id = test_tenant_1;
    RETURN QUERY SELECT 
        'Same tenant access'::TEXT,
        record_count > 0,
        CASE WHEN record_count > 0 THEN 'PASS: Can access same tenant data' 
             ELSE 'FAIL: Cannot access same tenant data' END;
    
    -- 测试2：切换租户后应该无法访问其他租户数据
    PERFORM set_current_tenant_id(test_tenant_2);
    SELECT COUNT(*) INTO record_count FROM corehr.employees WHERE tenant_id = test_tenant_1;
    RETURN QUERY SELECT 
        'Cross tenant isolation'::TEXT,
        record_count = 0,
        CASE WHEN record_count = 0 THEN 'PASS: Cross-tenant access blocked' 
             ELSE 'FAIL: Cross-tenant access allowed' END;
    
    -- 测试3：插入数据应该自动使用当前租户ID
    INSERT INTO corehr.employees (id, tenant_id, employee_number, first_name, last_name, email, status)
    VALUES (gen_random_uuid(), test_tenant_2, 'TEST002', 'Test2', 'User2', 'test2@example.com', 'active');
    
    SELECT COUNT(*) INTO record_count FROM corehr.employees WHERE tenant_id = test_tenant_2;
    RETURN QUERY SELECT 
        'Tenant context enforcement'::TEXT,
        record_count > 0,
        CASE WHEN record_count > 0 THEN 'PASS: Tenant context enforced on insert' 
             ELSE 'FAIL: Tenant context not enforced' END;
    
    -- 清理测试数据
    PERFORM clear_current_tenant_id();
    DELETE FROM corehr.employees WHERE id = test_emp_id OR tenant_id = test_tenant_2;
    DELETE FROM corehr.organizations WHERE id = test_org_id;
    
END;
$$ LANGUAGE plpgsql;

-- ===========================================
-- 租户数据统计函数
-- ===========================================

-- 创建获取租户数据统计的函数
CREATE OR REPLACE FUNCTION get_tenant_statistics(p_tenant_id UUID)
RETURNS JSONB AS $$
DECLARE
    stats JSONB;
BEGIN
    -- 确保只有超级管理员或租户管理员可以查看统计
    IF NOT (is_super_admin() OR (get_current_tenant_id() = p_tenant_id AND is_tenant_admin())) THEN
        RAISE EXCEPTION 'Insufficient permissions to view tenant statistics';
    END IF;
    
    SELECT jsonb_build_object(
        'tenant_id', p_tenant_id,
        'employees', (
            SELECT jsonb_build_object(
                'total', COUNT(*),
                'active', COUNT(*) FILTER (WHERE status = 'active'),
                'inactive', COUNT(*) FILTER (WHERE status = 'inactive')
            )
            FROM corehr.employees 
            WHERE tenant_id = p_tenant_id
        ),
        'organizations', (
            SELECT jsonb_build_object(
                'total', COUNT(*),
                'active', COUNT(*) FILTER (WHERE status = 'active')
            )
            FROM corehr.organizations 
            WHERE tenant_id = p_tenant_id
        ),
        'workflows', (
            SELECT jsonb_build_object(
                'total', COUNT(*),
                'running', COUNT(*) FILTER (WHERE status = 'running'),
                'completed', COUNT(*) FILTER (WHERE status = 'completed')
            )
            FROM workflow.workflow_instances 
            WHERE tenant_id = p_tenant_id
        ),
        'last_updated', NOW()
    ) INTO stats;
    
    RETURN stats;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- ===========================================
-- 数据迁移和租户管理
-- ===========================================

-- 创建租户数据迁移函数
CREATE OR REPLACE FUNCTION migrate_data_to_tenant(
    p_source_tenant_id UUID,
    p_target_tenant_id UUID,
    p_data_types TEXT[] DEFAULT ARRAY['employees', 'organizations']
)
RETURNS JSONB AS $$
DECLARE
    result JSONB := jsonb_build_object();
    migrated_count INTEGER;
BEGIN
    -- 只有超级管理员可以执行数据迁移
    IF NOT is_super_admin() THEN
        RAISE EXCEPTION 'Only super administrators can perform data migration';
    END IF;
    
    -- 迁移员工数据
    IF 'employees' = ANY(p_data_types) THEN
        UPDATE corehr.employees 
        SET tenant_id = p_target_tenant_id 
        WHERE tenant_id = p_source_tenant_id;
        
        GET DIAGNOSTICS migrated_count = ROW_COUNT;
        result := jsonb_set(result, '{employees_migrated}', to_jsonb(migrated_count));
    END IF;
    
    -- 迁移组织数据
    IF 'organizations' = ANY(p_data_types) THEN
        UPDATE corehr.organizations 
        SET tenant_id = p_target_tenant_id 
        WHERE tenant_id = p_source_tenant_id;
        
        GET DIAGNOSTICS migrated_count = ROW_COUNT;
        result := jsonb_set(result, '{organizations_migrated}', to_jsonb(migrated_count));
    END IF;
    
    -- 记录迁移操作
    INSERT INTO system.tenant_migration_log (
        source_tenant_id,
        target_tenant_id,
        data_types,
        migration_result,
        migrated_by,
        migration_time
    ) VALUES (
        p_source_tenant_id,
        p_target_tenant_id,
        p_data_types,
        result,
        current_setting('app.current_user_id', true)::uuid,
        NOW()
    );
    
    RETURN result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 创建租户迁移日志表
CREATE TABLE IF NOT EXISTS system.tenant_migration_log (
    id SERIAL PRIMARY KEY,
    source_tenant_id UUID NOT NULL,
    target_tenant_id UUID NOT NULL,
    data_types TEXT[],
    migration_result JSONB,
    migrated_by UUID,
    migration_time TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ===========================================
-- 权限和角色管理
-- ===========================================

-- 为应用程序角色授予必要权限
GRANT USAGE ON SCHEMA corehr TO cube_castle_app;
GRANT USAGE ON SCHEMA workflow TO cube_castle_app;
-- Outbox schema permissions removed in Phase 4
GRANT USAGE ON SCHEMA system TO cube_castle_app;

-- 授予表权限
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA corehr TO cube_castle_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA workflow TO cube_castle_app;
-- Outbox table permissions removed in Phase 4
GRANT SELECT, INSERT ON ALL TABLES IN SCHEMA system TO cube_castle_app;

-- 授予序列权限
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA corehr TO cube_castle_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA workflow TO cube_castle_app;
-- Outbox sequence permissions removed in Phase 4
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA system TO cube_castle_app;

-- 授予函数执行权限
GRANT EXECUTE ON FUNCTION set_current_tenant_id(uuid) TO cube_castle_app;
GRANT EXECUTE ON FUNCTION get_current_tenant_id() TO cube_castle_app;
GRANT EXECUTE ON FUNCTION clear_current_tenant_id() TO cube_castle_app;
GRANT EXECUTE ON FUNCTION is_super_admin() TO cube_castle_app;
GRANT EXECUTE ON FUNCTION is_tenant_admin() TO cube_castle_app;
GRANT EXECUTE ON FUNCTION get_tenant_statistics(uuid) TO cube_castle_app;

-- ===========================================
-- 备注和文档
-- ===========================================

COMMENT ON FUNCTION set_current_tenant_id(uuid) IS 
'设置当前会话的租户上下文，用于RLS策略。必须在每个数据库连接开始时调用。';

COMMENT ON FUNCTION get_current_tenant_id() IS 
'获取当前会话的租户ID，用于RLS策略评估。';

COMMENT ON FUNCTION is_super_admin() IS 
'检查当前用户是否为超级管理员，超级管理员可以访问所有租户数据。';

COMMENT ON FUNCTION is_tenant_admin() IS 
'检查当前用户是否为当前租户的管理员。';

COMMENT ON FUNCTION test_rls_policies() IS 
'测试RLS策略是否正确工作，返回测试结果。';

COMMENT ON FUNCTION get_tenant_statistics(uuid) IS 
'获取指定租户的数据统计信息，需要管理员权限。';

COMMENT ON TABLE system.tenant_access_log IS 
'记录租户访问日志，用于审计和监控。';

COMMENT ON TABLE system.rls_violation_log IS 
'记录RLS策略违规尝试，用于安全监控。';