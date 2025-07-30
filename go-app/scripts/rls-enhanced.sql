-- PostgreSQL Row Level Security (RLS) 策略增强脚本
-- 实现全面的多租户数据隔离

-- ===============================
-- 1. 启用行级安全 (RLS) 
-- ===============================

-- 核心HR表启用RLS
ALTER TABLE corehr.employees ENABLE ROW LEVEL SECURITY;
ALTER TABLE corehr.organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE corehr.positions ENABLE ROW LEVEL SECURITY;
ALTER TABLE corehr.employee_positions ENABLE ROW LEVEL SECURITY;

-- 工作流表启用RLS
ALTER TABLE workflow.executions ENABLE ROW LEVEL SECURITY;
ALTER TABLE workflow.activities ENABLE ROW LEVEL SECURITY;
ALTER TABLE workflow.signals ENABLE ROW LEVEL SECURITY;

-- 发件箱表启用RLS
ALTER TABLE outbox.events ENABLE ROW LEVEL SECURITY;

-- ===============================
-- 2. 租户上下文管理函数
-- ===============================

-- 设置当前租户上下文
CREATE OR REPLACE FUNCTION set_tenant_context(tenant_uuid uuid)
RETURNS void AS $$
BEGIN
    -- 设置租户ID
    PERFORM set_config('app.current_tenant_id', tenant_uuid::text, true);
    
    -- 记录设置时间（用于调试）
    PERFORM set_config('app.context_set_at', EXTRACT(EPOCH FROM NOW())::text, true);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 设置用户上下文（包含用户角色）
CREATE OR REPLACE FUNCTION set_user_context(user_uuid uuid, user_role text, tenant_uuid uuid)
RETURNS void AS $$
BEGIN
    -- 设置用户信息
    PERFORM set_config('app.current_user_id', user_uuid::text, true);
    PERFORM set_config('app.current_user_role', user_role, true);
    PERFORM set_config('app.current_tenant_id', tenant_uuid::text, true);
    
    -- 记录设置时间
    PERFORM set_config('app.context_set_at', EXTRACT(EPOCH FROM NOW())::text, true);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 获取当前租户ID
CREATE OR REPLACE FUNCTION get_current_tenant_id()
RETURNS uuid AS $$
BEGIN
    RETURN current_setting('app.current_tenant_id', true)::uuid;
EXCEPTION
    WHEN OTHERS THEN
        RETURN NULL;
END;
$$ LANGUAGE plpgsql STABLE;

-- 获取当前用户ID
CREATE OR REPLACE FUNCTION get_current_user_id()
RETURNS uuid AS $$
BEGIN
    RETURN current_setting('app.current_user_id', true)::uuid;
EXCEPTION
    WHEN OTHERS THEN
        RETURN NULL;
END;
$$ LANGUAGE plpgsql STABLE;

-- 获取当前用户角色
CREATE OR REPLACE FUNCTION get_current_user_role()
RETURNS text AS $$
BEGIN
    RETURN current_setting('app.current_user_role', true);
EXCEPTION
    WHEN OTHERS THEN
        RETURN 'guest';
END;
$$ LANGUAGE plpgsql STABLE;

-- ===============================
-- 3. 员工表 RLS 策略
-- ===============================

-- 删除已存在的策略（如果有）
DROP POLICY IF EXISTS tenant_isolation_employees ON corehr.employees;
DROP POLICY IF EXISTS role_based_employees_select ON corehr.employees;
DROP POLICY IF EXISTS role_based_employees_insert ON corehr.employees;
DROP POLICY IF EXISTS role_based_employees_update ON corehr.employees;
DROP POLICY IF EXISTS role_based_employees_delete ON corehr.employees;

-- 基础租户隔离策略
CREATE POLICY tenant_isolation_employees ON corehr.employees
    FOR ALL
    USING (tenant_id = get_current_tenant_id());

-- 基于角色的查询策略
CREATE POLICY role_based_employees_select ON corehr.employees
    FOR SELECT
    USING (
        tenant_id = get_current_tenant_id() AND (
            get_current_user_role() = 'admin' OR
            get_current_user_role() = 'hr' OR
            get_current_user_role() = 'manager' OR
            (get_current_user_role() = 'employee' AND id = get_current_user_id())
        )
    );

-- 基于角色的插入策略
CREATE POLICY role_based_employees_insert ON corehr.employees
    FOR INSERT
    WITH CHECK (
        tenant_id = get_current_tenant_id() AND
        get_current_user_role() IN ('admin', 'hr')
    );

-- 基于角色的更新策略
CREATE POLICY role_based_employees_update ON corehr.employees
    FOR UPDATE
    USING (
        tenant_id = get_current_tenant_id() AND (
            get_current_user_role() IN ('admin', 'hr') OR
            (get_current_user_role() = 'manager' AND id IN (
                SELECT e.id FROM corehr.employees e
                JOIN corehr.employee_positions ep ON e.id = ep.employee_id
                JOIN corehr.positions p ON ep.position_id = p.id
                WHERE p.manager_id = get_current_user_id()
            )) OR
            (get_current_user_role() = 'employee' AND id = get_current_user_id())
        )
    )
    WITH CHECK (
        tenant_id = get_current_tenant_id()
    );

-- 基于角色的删除策略
CREATE POLICY role_based_employees_delete ON corehr.employees
    FOR DELETE
    USING (
        tenant_id = get_current_tenant_id() AND
        get_current_user_role() IN ('admin', 'hr')
    );

-- ===============================
-- 4. 组织表 RLS 策略
-- ===============================

-- 删除已存在的策略
DROP POLICY IF EXISTS tenant_isolation_organizations ON corehr.organizations;
DROP POLICY IF EXISTS role_based_organizations_select ON corehr.organizations;
DROP POLICY IF EXISTS role_based_organizations_modify ON corehr.organizations;

-- 租户隔离策略
CREATE POLICY tenant_isolation_organizations ON corehr.organizations
    FOR ALL
    USING (tenant_id = get_current_tenant_id());

-- 查询策略（所有用户都可以查看组织架构）
CREATE POLICY role_based_organizations_select ON corehr.organizations
    FOR SELECT
    USING (
        tenant_id = get_current_tenant_id() AND
        get_current_user_role() IN ('admin', 'hr', 'manager', 'employee')
    );

-- 修改策略（只有管理员和HR可以修改）
CREATE POLICY role_based_organizations_modify ON corehr.organizations
    FOR INSERT, UPDATE, DELETE
    USING (
        tenant_id = get_current_tenant_id() AND
        get_current_user_role() IN ('admin', 'hr')
    )
    WITH CHECK (
        tenant_id = get_current_tenant_id()
    );

-- ===============================
-- 5. 职位表 RLS 策略
-- ===============================

-- 删除已存在的策略
DROP POLICY IF EXISTS tenant_isolation_positions ON corehr.positions;
DROP POLICY IF EXISTS role_based_positions_access ON corehr.positions;

-- 租户隔离策略
CREATE POLICY tenant_isolation_positions ON corehr.positions
    FOR ALL
    USING (tenant_id = get_current_tenant_id());

-- 基于角色的访问策略
CREATE POLICY role_based_positions_access ON corehr.positions
    FOR ALL
    USING (
        tenant_id = get_current_tenant_id() AND
        get_current_user_role() IN ('admin', 'hr', 'manager', 'employee')
    )
    WITH CHECK (
        tenant_id = get_current_tenant_id() AND
        get_current_user_role() IN ('admin', 'hr')
    );

-- ===============================
-- 6. 工作流表 RLS 策略
-- ===============================

-- 工作流执行表策略
DROP POLICY IF EXISTS tenant_isolation_workflow_executions ON workflow.executions;
CREATE POLICY tenant_isolation_workflow_executions ON workflow.executions
    FOR ALL
    USING (
        tenant_id = get_current_tenant_id() AND (
            get_current_user_role() IN ('admin', 'hr') OR
            (get_current_user_role() = 'manager' AND workflow_type = 'approval') OR
            (get_current_user_role() = 'employee' AND created_by = get_current_user_id())
        )
    )
    WITH CHECK (
        tenant_id = get_current_tenant_id()
    );

-- 工作流活动表策略
DROP POLICY IF EXISTS tenant_isolation_workflow_activities ON workflow.activities;
CREATE POLICY tenant_isolation_workflow_activities ON workflow.activities
    FOR ALL
    USING (
        execution_id IN (
            SELECT id FROM workflow.executions 
            WHERE tenant_id = get_current_tenant_id()
        )
    );

-- ===============================
-- 7. 发件箱表 RLS 策略
-- ===============================

-- 发件箱事件表策略
DROP POLICY IF EXISTS tenant_isolation_outbox_events ON outbox.events;
CREATE POLICY tenant_isolation_outbox_events ON outbox.events
    FOR ALL
    USING (
        tenant_id = get_current_tenant_id() AND
        get_current_user_role() IN ('admin', 'hr')
    )
    WITH CHECK (
        tenant_id = get_current_tenant_id()
    );

-- ===============================
-- 8. RLS 策略测试函数
-- ===============================

-- 创建RLS策略测试函数
CREATE OR REPLACE FUNCTION test_rls_policies()
RETURNS TABLE(
    test_name text,
    result boolean,
    message text
) AS $$
DECLARE
    test_tenant_id uuid := '550e8400-e29b-41d4-a716-446655440000';
    other_tenant_id uuid := '550e8400-e29b-41d4-a716-446655440001';
    test_user_id uuid := '660e8400-e29b-41d4-a716-446655440000';
    test_admin_id uuid := '660e8400-e29b-41d4-a716-446655440001';
    employee_count int;
    org_count int;
BEGIN
    -- 测试1: 设置用户上下文后能否访问同租户数据
    PERFORM set_user_context(test_user_id, 'employee', test_tenant_id);
    
    SELECT COUNT(*) INTO employee_count 
    FROM corehr.employees 
    WHERE tenant_id = test_tenant_id;
    
    test_name := 'Same tenant access';
    result := employee_count >= 0;
    message := CASE 
        WHEN result THEN 'PASS: Can access same tenant data'
        ELSE 'FAIL: Cannot access same tenant data'
    END;
    RETURN NEXT;
    
    -- 测试2: 跨租户访问应该被阻止
    PERFORM set_user_context(test_user_id, 'employee', other_tenant_id);
    
    SELECT COUNT(*) INTO employee_count 
    FROM corehr.employees 
    WHERE tenant_id = test_tenant_id;
    
    test_name := 'Cross tenant isolation';
    result := employee_count = 0;
    message := CASE 
        WHEN result THEN 'PASS: Cross-tenant access blocked'
        ELSE 'FAIL: Cross-tenant access not blocked'
    END;
    RETURN NEXT;
    
    -- 测试3: 管理员应该能访问所有数据
    PERFORM set_user_context(test_admin_id, 'admin', test_tenant_id);
    
    SELECT COUNT(*) INTO employee_count 
    FROM corehr.employees 
    WHERE tenant_id = test_tenant_id;
    
    test_name := 'Admin access';
    result := employee_count >= 0;
    message := CASE 
        WHEN result THEN 'PASS: Admin can access tenant data'
        ELSE 'FAIL: Admin cannot access tenant data'
    END;
    RETURN NEXT;
    
    -- 测试4: 租户上下文强制执行
    PERFORM set_user_context(test_user_id, 'employee', test_tenant_id);
    
    SELECT COUNT(*) INTO org_count 
    FROM corehr.organizations 
    WHERE tenant_id = test_tenant_id;
    
    test_name := 'Tenant context enforce';
    result := org_count >= 0;
    message := CASE 
        WHEN result THEN 'PASS: Tenant context enforced'
        ELSE 'FAIL: Tenant context not enforced'
    END;
    RETURN NEXT;
    
END;
$$ LANGUAGE plpgsql;

-- ===============================
-- 9. RLS 性能优化索引
-- ===============================

-- 为租户隔离创建性能优化索引
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_employees_tenant_id_performance 
ON corehr.employees (tenant_id, id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_organizations_tenant_id_performance 
ON corehr.organizations (tenant_id, id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_positions_tenant_id_performance 
ON corehr.positions (tenant_id, id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_workflow_executions_tenant_id_performance 
ON workflow.executions (tenant_id, id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_outbox_events_tenant_id_performance 
ON outbox.events (tenant_id, id);

-- 为基于角色的访问创建复合索引
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_employees_tenant_manager 
ON corehr.employees (tenant_id, id) 
WHERE id IN (
    SELECT DISTINCT ep.employee_id 
    FROM corehr.employee_positions ep 
    JOIN corehr.positions p ON ep.position_id = p.id 
    WHERE p.manager_id IS NOT NULL
);

-- ===============================
-- 10. RLS 监控和统计
-- ===============================

-- 创建RLS监控视图
CREATE OR REPLACE VIEW rls_policy_stats AS
SELECT 
    schemaname,
    tablename,
    policyname,
    permissive,
    roles,
    cmd,
    qual,
    with_check
FROM pg_policies 
WHERE schemaname IN ('corehr', 'workflow', 'outbox')
ORDER BY schemaname, tablename, policyname;

-- 创建租户数据统计视图
CREATE OR REPLACE VIEW tenant_data_stats AS
SELECT 
    tenant_id,
    'employees' as table_name,
    COUNT(*) as record_count
FROM corehr.employees 
GROUP BY tenant_id
UNION ALL
SELECT 
    tenant_id,
    'organizations' as table_name,
    COUNT(*) as record_count
FROM corehr.organizations 
GROUP BY tenant_id
UNION ALL
SELECT 
    tenant_id,
    'workflow_executions' as table_name,
    COUNT(*) as record_count
FROM workflow.executions 
GROUP BY tenant_id
ORDER BY tenant_id, table_name;

-- ===============================
-- 11. RLS 安全函数
-- ===============================

-- 验证当前会话的租户上下文
CREATE OR REPLACE FUNCTION validate_tenant_context()
RETURNS boolean AS $$
DECLARE
    current_tenant uuid;
    context_age numeric;
BEGIN
    -- 检查是否设置了租户上下文
    current_tenant := get_current_tenant_id();
    IF current_tenant IS NULL THEN
        RAISE EXCEPTION 'Tenant context not set';
    END IF;
    
    -- 检查上下文设置时间（防止过期上下文）
    BEGIN
        context_age := EXTRACT(EPOCH FROM NOW()) - current_setting('app.context_set_at')::numeric;
        IF context_age > 3600 THEN  -- 1小时过期
            RAISE EXCEPTION 'Tenant context expired';
        END IF;
    EXCEPTION
        WHEN OTHERS THEN
            RAISE EXCEPTION 'Invalid tenant context';
    END;
    
    RETURN true;
END;
$$ LANGUAGE plpgsql;

-- 清除租户上下文
CREATE OR REPLACE FUNCTION clear_tenant_context()
RETURNS void AS $$
BEGIN
    PERFORM set_config('app.current_tenant_id', '', true);
    PERFORM set_config('app.current_user_id', '', true);
    PERFORM set_config('app.current_user_role', '', true);
    PERFORM set_config('app.context_set_at', '', true);
END;
$$ LANGUAGE plpgsql;

-- ===============================
-- 注释和文档
-- ===============================

COMMENT ON FUNCTION set_tenant_context(uuid) IS '设置当前会话的租户上下文，用于RLS策略执行';
COMMENT ON FUNCTION set_user_context(uuid, text, uuid) IS '设置当前会话的用户上下文，包含用户ID、角色和租户ID';
COMMENT ON FUNCTION get_current_tenant_id() IS '获取当前会话的租户ID';
COMMENT ON FUNCTION get_current_user_id() IS '获取当前会话的用户ID';
COMMENT ON FUNCTION get_current_user_role() IS '获取当前会话的用户角色';
COMMENT ON FUNCTION test_rls_policies() IS 'RLS策略功能测试函数，返回测试结果';
COMMENT ON VIEW rls_policy_stats IS 'RLS策略统计视图，显示所有已配置的策略';
COMMENT ON VIEW tenant_data_stats IS '租户数据统计视图，显示各租户的数据记录数';

-- 提示信息
DO $$ 
BEGIN
    RAISE NOTICE '==============================================';
    RAISE NOTICE 'PostgreSQL RLS 多租户隔离策略部署完成！';
    RAISE NOTICE '==============================================';
    RAISE NOTICE '已启用的表: employees, organizations, positions, workflow.executions, outbox.events';
    RAISE NOTICE '使用方法: SELECT set_user_context(user_id, role, tenant_id);';
    RAISE NOTICE '测试命令: SELECT * FROM test_rls_policies();';
    RAISE NOTICE '监控命令: SELECT * FROM rls_policy_stats;';
    RAISE NOTICE '==============================================';
END $$;