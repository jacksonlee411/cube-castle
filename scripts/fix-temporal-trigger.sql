-- 修复时态记录创建的触发器问题
-- 解决层级字段NOT NULL约束违反问题
-- 创建日期: 2025-08-11

BEGIN;

-- 1. 创建优化的触发器函数，特殊处理时态记录
CREATE OR REPLACE FUNCTION generate_org_unit_code()
RETURNS TRIGGER AS $$
BEGIN
    -- 自动生成7位编码（如果为空）
    IF NEW.code IS NULL THEN
        NEW.code := LPAD(nextval('org_unit_code_seq')::text, 7, '0');
    END IF;
    
    -- 层级和路径计算逻辑优化
    -- 检查是否为时态记录（已有effective_date或end_date或is_current字段）
    IF NEW.effective_date IS NOT NULL OR NEW.end_date IS NOT NULL OR NEW.is_current IS NOT NULL THEN
        -- 这是时态记录创建，应用程序应该已经提供了完整的层级信息
        -- 如果level为NULL，需要计算或使用默认值
        IF NEW.level IS NULL THEN
            IF NEW.parent_code IS NOT NULL THEN
                -- 尝试从父组织计算层级
                SELECT COALESCE(level, 0) + 1, COALESCE(path, '') || '/' || NEW.code 
                INTO NEW.level, NEW.path
                FROM organization_units 
                WHERE code = NEW.parent_code 
                AND tenant_id = NEW.tenant_id
                ORDER BY CASE WHEN is_current = true THEN 0 ELSE 1 END,
                         effective_date DESC
                LIMIT 1;
                
                -- 如果仍然是NULL，说明父组织不存在，设置默认值
                IF NEW.level IS NULL THEN
                    RAISE WARNING '父组织 % 不存在，设置为根组织', NEW.parent_code;
                    NEW.level := 1;
                    NEW.path := '/' || NEW.code;
                END IF;
            ELSE
                -- 根组织
                NEW.level := 1;
                NEW.path := '/' || NEW.code;
            END IF;
        END IF;
        
        -- 如果path为NULL，根据level和parent_code重新计算
        IF NEW.path IS NULL THEN
            IF NEW.parent_code IS NOT NULL THEN
                SELECT COALESCE(path, '') || '/' || NEW.code 
                INTO NEW.path
                FROM organization_units 
                WHERE code = NEW.parent_code 
                AND tenant_id = NEW.tenant_id
                ORDER BY CASE WHEN is_current = true THEN 0 ELSE 1 END,
                         effective_date DESC
                LIMIT 1;
                
                IF NEW.path IS NULL THEN
                    NEW.path := '/' || NEW.code;
                END IF;
            ELSE
                NEW.path := '/' || NEW.code;
            END IF;
        END IF;
    ELSE
        -- 这是常规记录创建，使用原有逻辑
        IF NEW.parent_code IS NOT NULL THEN
            SELECT level + 1, path || '/' || NEW.code 
            INTO NEW.level, NEW.path
            FROM organization_units 
            WHERE code = NEW.parent_code
            AND tenant_id = NEW.tenant_id
            AND is_current = true;
            
            -- 如果当前记录不存在，查找最新记录
            IF NEW.level IS NULL THEN
                SELECT level + 1, path || '/' || NEW.code 
                INTO NEW.level, NEW.path
                FROM organization_units 
                WHERE code = NEW.parent_code
                AND tenant_id = NEW.tenant_id
                ORDER BY effective_date DESC
                LIMIT 1;
            END IF;
        ELSE
            NEW.level := 1;
            NEW.path := '/' || NEW.code;
        END IF;
    END IF;
    
    -- 确保level和path不为NULL
    IF NEW.level IS NULL THEN
        NEW.level := 1;
    END IF;
    
    IF NEW.path IS NULL THEN
        NEW.path := '/' || NEW.code;
    END IF;
    
    -- 更新时间戳
    NEW.updated_at := NOW();
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 2. 重新创建触发器（如果需要）
DROP TRIGGER IF EXISTS set_org_unit_code ON organization_units;
CREATE TRIGGER set_org_unit_code 
    BEFORE INSERT OR UPDATE ON organization_units 
    FOR EACH ROW EXECUTE FUNCTION generate_org_unit_code();

-- 3. 验证触发器修复
-- 可以通过以下查询测试触发器是否正常工作：
/*
-- 测试时态记录创建
INSERT INTO organization_units (
    code, tenant_id, name, unit_type, status, level, path,
    effective_date, is_current
) VALUES (
    '1000999', '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9', 
    '测试时态组织', 'DEPARTMENT', 'ACTIVE', NULL, NULL,
    '2025-08-11', true
);

-- 查看结果
SELECT code, name, level, path, effective_date, is_current 
FROM organization_units 
WHERE code = '1000999';

-- 清理测试数据
DELETE FROM organization_units WHERE code = '1000999';
*/

COMMIT;

-- 日志记录
SELECT 'generate_org_unit_code 触发器已优化' AS status,
       'level和path字段NULL约束问题已修复' AS fix_description,
       '支持时态记录创建' AS feature;