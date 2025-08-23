-- ===============================================
-- 高级层级管理系统实现 (修正版)
-- 功能：17级深度限制 + 级联路径更新 + 双路径系统
-- 作者：Claude Code Assistant
-- 创建时间：2025-08-23
-- ===============================================

-- 1. 数据库表结构扩展 - 支持双路径系统
-- ===============================================

-- 1.1 添加新的路径字段
ALTER TABLE organization_units 
ADD COLUMN IF NOT EXISTS code_path VARCHAR(2000),        -- 编码路径: /1000000/1000001/1000002
ADD COLUMN IF NOT EXISTS name_path VARCHAR(4000),        -- 名称路径: /高谷集团/爱治理办公室/技术部
ADD COLUMN IF NOT EXISTS hierarchy_depth INTEGER DEFAULT 1; -- 层级深度缓存，便于查询优化

-- 1.2 添加层级深度约束 (最大17级)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'hierarchy_depth_limit') THEN
        ALTER TABLE organization_units 
        ADD CONSTRAINT hierarchy_depth_limit 
        CHECK (hierarchy_depth > 0 AND hierarchy_depth <= 17);
    END IF;
END $$;

-- 1.3 添加字段注释
COMMENT ON COLUMN organization_units.path IS '编码路径别名，与code_path保持同步';
COMMENT ON COLUMN organization_units.code_path IS '编码路径：/1000000/1000001/1000002';
COMMENT ON COLUMN organization_units.name_path IS '名称路径：/高谷集团/爱治理办公室/技术部';
COMMENT ON COLUMN organization_units.hierarchy_depth IS '层级深度：1-17级，与level字段同步';

-- 2. 高性能索引系统 (非并发创建)
-- ===============================================

-- 2.1 层级查询优化索引
DROP INDEX IF EXISTS idx_org_hierarchy_depth;
CREATE INDEX idx_org_hierarchy_depth 
    ON organization_units(tenant_id, hierarchy_depth, status, is_current) 
    WHERE is_current = true;

-- 2.2 路径搜索优化索引 (需要pg_trgm扩展)
CREATE EXTENSION IF NOT EXISTS pg_trgm;

DROP INDEX IF EXISTS idx_org_code_path_gin;
CREATE INDEX idx_org_code_path_gin 
    ON organization_units USING gin(code_path gin_trgm_ops);

DROP INDEX IF EXISTS idx_org_name_path_gin;
CREATE INDEX idx_org_name_path_gin 
    ON organization_units USING gin(name_path gin_trgm_ops);

-- 2.3 父子关系优化索引
DROP INDEX IF EXISTS idx_org_parent_current;
CREATE INDEX idx_org_parent_current 
    ON organization_units(parent_code, tenant_id, is_current) 
    WHERE is_current = true;

-- 3. 层级自动计算核心函数
-- ===============================================

-- 3.1 单个组织层级计算函数
CREATE OR REPLACE FUNCTION calculate_org_hierarchy(
    p_code VARCHAR(10),
    p_tenant_id UUID
) RETURNS TABLE (
    calculated_level INTEGER,
    calculated_code_path VARCHAR(2000),
    calculated_name_path VARCHAR(4000),
    calculated_hierarchy_depth INTEGER
) AS $$
DECLARE
    parent_info RECORD;
    current_name VARCHAR(255);
BEGIN
    -- 获取当前组织名称
    SELECT name INTO current_name 
    FROM organization_units 
    WHERE code = p_code AND tenant_id = p_tenant_id AND is_current = true
    LIMIT 1;
    
    -- 获取父组织信息
    SELECT 
        ou.code,
        ou.level,
        ou.code_path,
        ou.name_path,
        ou.hierarchy_depth
    INTO parent_info
    FROM organization_units ou
    WHERE ou.code = (
        SELECT parent_code 
        FROM organization_units 
        WHERE code = p_code AND tenant_id = p_tenant_id AND is_current = true
        LIMIT 1
    ) 
    AND ou.tenant_id = p_tenant_id 
    AND ou.is_current = true
    LIMIT 1;
    
    -- 计算层级信息
    IF parent_info.code IS NULL THEN
        -- 根组织
        calculated_level := 1;
        calculated_hierarchy_depth := 1;
        calculated_code_path := '/' || p_code;
        calculated_name_path := '/' || COALESCE(current_name, p_code);
    ELSE
        -- 子组织
        calculated_level := parent_info.level + 1;
        calculated_hierarchy_depth := parent_info.hierarchy_depth + 1;
        calculated_code_path := COALESCE(parent_info.code_path, '/' || parent_info.code) || '/' || p_code;
        calculated_name_path := COALESCE(parent_info.name_path, '/' || current_name) || '/' || COALESCE(current_name, p_code);
        
        -- 检查层级深度限制
        IF calculated_hierarchy_depth > 17 THEN
            RAISE EXCEPTION '组织层级超过最大限制17级！当前尝试创建第%级组织。', calculated_hierarchy_depth;
        END IF;
    END IF;
    
    RETURN NEXT;
END;
$$ LANGUAGE plpgsql;

-- 3.2 批量层级重计算函数 (用于级联更新)
CREATE OR REPLACE FUNCTION recalculate_hierarchy_cascade(
    p_parent_code VARCHAR(10),
    p_tenant_id UUID
) RETURNS INTEGER AS $$
DECLARE
    affected_count INTEGER := 0;
    child_record RECORD;
    hierarchy_info RECORD;
BEGIN
    -- 递归更新所有子组织
    FOR child_record IN 
        WITH RECURSIVE org_children AS (
            -- 直接子组织
            SELECT code, parent_code, name, level
            FROM organization_units 
            WHERE parent_code = p_parent_code 
                AND tenant_id = p_tenant_id 
                AND is_current = true
            
            UNION ALL
            
            -- 间接子组织 (递归)
            SELECT o.code, o.parent_code, o.name, o.level
            FROM organization_units o
            INNER JOIN org_children oc ON o.parent_code = oc.code
            WHERE o.tenant_id = p_tenant_id AND o.is_current = true
        )
        SELECT * FROM org_children
    LOOP
        -- 计算当前子组织的层级信息
        SELECT * INTO hierarchy_info 
        FROM calculate_org_hierarchy(child_record.code, p_tenant_id);
        
        -- 更新子组织的层级信息
        UPDATE organization_units 
        SET 
            level = hierarchy_info.calculated_level,
            hierarchy_depth = hierarchy_info.calculated_hierarchy_depth,
            code_path = hierarchy_info.calculated_code_path,
            name_path = hierarchy_info.calculated_name_path,
            path = hierarchy_info.calculated_code_path, -- 保持向后兼容
            updated_at = CURRENT_TIMESTAMP
        WHERE code = child_record.code 
            AND tenant_id = p_tenant_id 
            AND is_current = true;
        
        affected_count := affected_count + 1;
        
        -- 记录级联更新日志
        RAISE NOTICE '级联更新组织 %: level=%, depth=%, code_path=%', 
            child_record.code, 
            hierarchy_info.calculated_level,
            hierarchy_info.calculated_hierarchy_depth,
            hierarchy_info.calculated_code_path;
    END LOOP;
    
    RETURN affected_count;
END;
$$ LANGUAGE plpgsql;

-- 4. 智能层级管理触发器
-- ===============================================

-- 4.1 层级自动计算和级联更新触发器
CREATE OR REPLACE FUNCTION smart_hierarchy_trigger() RETURNS TRIGGER AS $$
DECLARE
    hierarchy_info RECORD;
BEGIN
    -- INSERT操作：计算新组织的层级信息
    IF TG_OP = 'INSERT' THEN
        SELECT * INTO hierarchy_info 
        FROM calculate_org_hierarchy(NEW.code, NEW.tenant_id);
        
        NEW.level := hierarchy_info.calculated_level;
        NEW.hierarchy_depth := hierarchy_info.calculated_hierarchy_depth;
        NEW.code_path := hierarchy_info.calculated_code_path;
        NEW.name_path := hierarchy_info.calculated_name_path;
        NEW.path := hierarchy_info.calculated_code_path;
        
        RETURN NEW;
    END IF;
    
    -- UPDATE操作：检查是否需要级联更新
    IF TG_OP = 'UPDATE' THEN
        -- 检查关键字段变化
        IF OLD.parent_code IS DISTINCT FROM NEW.parent_code 
           OR OLD.name IS DISTINCT FROM NEW.name THEN
            
            -- 重新计算当前组织层级
            SELECT * INTO hierarchy_info 
            FROM calculate_org_hierarchy(NEW.code, NEW.tenant_id);
            
            NEW.level := hierarchy_info.calculated_level;
            NEW.hierarchy_depth := hierarchy_info.calculated_hierarchy_depth;
            NEW.code_path := hierarchy_info.calculated_code_path;
            NEW.name_path := hierarchy_info.calculated_name_path;
            NEW.path := hierarchy_info.calculated_code_path;
        END IF;
        
        RETURN NEW;
    END IF;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- 4.2 绑定触发器到表
DROP TRIGGER IF EXISTS smart_hierarchy_management ON organization_units;
CREATE TRIGGER smart_hierarchy_management
    BEFORE INSERT OR UPDATE ON organization_units
    FOR EACH ROW 
    EXECUTE FUNCTION smart_hierarchy_trigger();

-- 5. 层级验证和约束函数
-- ===============================================

-- 5.1 循环引用检测函数
CREATE OR REPLACE FUNCTION check_circular_reference(
    p_code VARCHAR(10),
    p_parent_code VARCHAR(10),
    p_tenant_id UUID
) RETURNS BOOLEAN AS $$
DECLARE
    current_code VARCHAR(10) := p_parent_code;
    depth_counter INTEGER := 0;
BEGIN
    -- 如果parent_code为空，不存在循环引用
    IF p_parent_code IS NULL THEN
        RETURN FALSE;
    END IF;
    
    -- 向上追溯检查循环引用
    WHILE current_code IS NOT NULL LOOP
        depth_counter := depth_counter + 1;
        
        -- 防止无限循环
        IF depth_counter > 20 THEN
            RAISE EXCEPTION '检测到潜在的循环引用或层级过深！组织编码: %', p_code;
        END IF;
        
        -- 如果找到了目标组织编码，说明存在循环引用
        IF current_code = p_code THEN
            RETURN TRUE;
        END IF;
        
        -- 继续向上查找
        SELECT parent_code INTO current_code
        FROM organization_units
        WHERE code = current_code AND tenant_id = p_tenant_id AND is_current = true
        LIMIT 1;
    END LOOP;
    
    RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

-- 5.2 层级变更验证触发器
CREATE OR REPLACE FUNCTION validate_hierarchy_changes() RETURNS TRIGGER AS $$
BEGIN
    -- 检查循环引用
    IF NEW.parent_code IS NOT NULL THEN
        IF check_circular_reference(NEW.code, NEW.parent_code, NEW.tenant_id) THEN
            RAISE EXCEPTION '不能设置父组织，会导致循环引用！组织 % 尝试设置父组织 %', 
                NEW.code, NEW.parent_code;
        END IF;
    END IF;
    
    -- 检查父组织是否存在
    IF NEW.parent_code IS NOT NULL THEN
        IF NOT EXISTS (
            SELECT 1 FROM organization_units 
            WHERE code = NEW.parent_code 
                AND tenant_id = NEW.tenant_id 
                AND is_current = true
        ) THEN
            RAISE EXCEPTION '父组织不存在！父组织编码: %', NEW.parent_code;
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 绑定验证触发器
DROP TRIGGER IF EXISTS validate_hierarchy ON organization_units;
CREATE TRIGGER validate_hierarchy
    BEFORE INSERT OR UPDATE ON organization_units
    FOR EACH ROW 
    EXECUTE FUNCTION validate_hierarchy_changes();

-- 6. 数据修复和初始化
-- ===============================================

-- 6.1 修复现有数据的层级信息
WITH RECURSIVE org_hierarchy AS (
    -- 根组织 (parent_code IS NULL)
    SELECT 
        code, 
        parent_code, 
        tenant_id,
        name,
        1 as calculated_level,
        1 as calculated_depth,
        ('/' || code) as calculated_code_path,
        ('/' || name) as calculated_name_path
    FROM organization_units 
    WHERE parent_code IS NULL AND is_current = true
    
    UNION ALL
    
    -- 子组织 (递归)
    SELECT 
        o.code,
        o.parent_code,
        o.tenant_id,
        o.name,
        oh.calculated_level + 1,
        oh.calculated_depth + 1,
        (oh.calculated_code_path || '/' || o.code),
        (oh.calculated_name_path || '/' || o.name)
    FROM organization_units o
    INNER JOIN org_hierarchy oh ON o.parent_code = oh.code
    WHERE o.is_current = true AND oh.calculated_depth < 17
)
UPDATE organization_units 
SET 
    level = org_hierarchy.calculated_level,
    hierarchy_depth = org_hierarchy.calculated_depth,
    code_path = org_hierarchy.calculated_code_path,
    name_path = org_hierarchy.calculated_name_path,
    path = org_hierarchy.calculated_code_path,
    updated_at = CURRENT_TIMESTAMP
FROM org_hierarchy 
WHERE organization_units.code = org_hierarchy.code 
    AND organization_units.tenant_id = org_hierarchy.tenant_id
    AND organization_units.is_current = true;

-- 7. 性能监控和统计视图
-- ===============================================

-- 7.1 层级分布统计视图
CREATE OR REPLACE VIEW v_hierarchy_statistics AS
SELECT 
    tenant_id,
    hierarchy_depth,
    COUNT(*) as org_count,
    COUNT(*) * 100.0 / SUM(COUNT(*)) OVER (PARTITION BY tenant_id) as percentage,
    array_agg(code ORDER BY code) as sample_codes
FROM organization_units 
WHERE is_current = true
GROUP BY tenant_id, hierarchy_depth
ORDER BY tenant_id, hierarchy_depth;

-- 7.2 路径完整性检查视图
CREATE OR REPLACE VIEW v_path_integrity_check AS
SELECT 
    code,
    name,
    parent_code,
    level,
    hierarchy_depth,
    code_path,
    name_path,
    -- 检查路径一致性
    CASE 
        WHEN level != hierarchy_depth THEN 'level_depth_mismatch'
        WHEN code_path IS NULL OR name_path IS NULL THEN 'missing_paths'
        WHEN array_length(string_to_array(code_path, '/'), 1) - 1 != level THEN 'code_path_level_mismatch'
        ELSE 'ok'
    END as integrity_status
FROM organization_units 
WHERE is_current = true;

-- 8. 实用工具函数
-- ===============================================

-- 8.1 获取组织完整路径信息
CREATE OR REPLACE FUNCTION get_org_full_path(p_code VARCHAR(10), p_tenant_id UUID)
RETURNS TABLE (
    code VARCHAR(10),
    name VARCHAR(255),
    level INTEGER,
    hierarchy_depth INTEGER,
    code_path VARCHAR(2000),
    name_path VARCHAR(4000),
    parent_chain TEXT[]
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ou.code,
        ou.name,
        ou.level,
        ou.hierarchy_depth,
        ou.code_path,
        ou.name_path,
        string_to_array(trim(both '/' from ou.code_path), '/') as parent_chain
    FROM organization_units ou
    WHERE ou.code = p_code 
        AND ou.tenant_id = p_tenant_id 
        AND ou.is_current = true;
END;
$$ LANGUAGE plpgsql;

-- 8.2 获取组织子树
CREATE OR REPLACE FUNCTION get_org_subtree(p_code VARCHAR(10), p_tenant_id UUID)
RETURNS TABLE (
    code VARCHAR(10),
    name VARCHAR(255),
    level INTEGER,
    hierarchy_depth INTEGER,
    code_path VARCHAR(2000),
    name_path VARCHAR(4000)
) AS $$
BEGIN
    RETURN QUERY
    WITH RECURSIVE subtree AS (
        -- 起始节点
        SELECT 
            ou.code, ou.name, ou.level, ou.hierarchy_depth,
            ou.code_path, ou.name_path
        FROM organization_units ou
        WHERE ou.code = p_code AND ou.tenant_id = p_tenant_id AND ou.is_current = true
        
        UNION ALL
        
        -- 子节点
        SELECT 
            o.code, o.name, o.level, o.hierarchy_depth,
            o.code_path, o.name_path
        FROM organization_units o
        INNER JOIN subtree st ON o.parent_code = st.code
        WHERE o.tenant_id = p_tenant_id AND o.is_current = true
    )
    SELECT * FROM subtree ORDER BY hierarchy_depth, code;
END;
$$ LANGUAGE plpgsql;

-- 9. 验证和测试
-- ===============================================

-- 验证层级计算结果
DO $$
DECLARE
    total_orgs INTEGER;
    max_depth INTEGER;
    integrity_issues INTEGER;
BEGIN
    SELECT COUNT(*), MAX(hierarchy_depth) 
    INTO total_orgs, max_depth
    FROM organization_units WHERE is_current = true;
    
    SELECT COUNT(*) 
    INTO integrity_issues
    FROM v_path_integrity_check 
    WHERE integrity_status != 'ok';
    
    RAISE NOTICE '=== 高级层级管理系统初始化完成 ===';
    RAISE NOTICE '总组织数: %', total_orgs;
    RAISE NOTICE '最大层级深度: %', max_depth;
    RAISE NOTICE '完整性问题: %', integrity_issues;
    
    IF integrity_issues > 0 THEN
        RAISE WARNING '发现 % 个层级完整性问题，请检查 v_path_integrity_check 视图', integrity_issues;
    ELSE
        RAISE NOTICE '✅ 所有组织层级信息完整性验证通过';
    END IF;
END $$;