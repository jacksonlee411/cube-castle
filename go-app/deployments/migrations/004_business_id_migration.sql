-- 业务ID系统数据库迁移脚本
-- 文件: 004_business_id_migration.sql
-- 日期: 2025-08-04
-- 描述: 为现有数据添加业务ID支持，实现UUID到业务ID的平滑迁移

-- ==========================================
-- 第一部分: 序列和函数创建
-- ==========================================

-- 创建业务ID生成序列
CREATE SEQUENCE IF NOT EXISTS employee_business_id_seq 
  START WITH 1 
  INCREMENT BY 1 
  MAXVALUE 99999999 
  NO CYCLE;

CREATE SEQUENCE IF NOT EXISTS org_business_id_seq 
  START WITH 0 
  INCREMENT BY 1 
  MAXVALUE 899999 
  NO CYCLE;

CREATE SEQUENCE IF NOT EXISTS position_business_id_seq 
  START WITH 0 
  INCREMENT BY 1 
  MAXVALUE 8999999 
  NO CYCLE;

-- 创建业务ID生成函数
CREATE OR REPLACE FUNCTION generate_business_id(entity_type TEXT) 
RETURNS TEXT AS $$
DECLARE
    new_id TEXT;
BEGIN
    CASE entity_type
        WHEN 'employee' THEN
            SELECT nextval('employee_business_id_seq')::text INTO new_id;
        WHEN 'organization' THEN  
            SELECT (100000 + nextval('org_business_id_seq'))::text INTO new_id;
        WHEN 'position' THEN
            SELECT (1000000 + nextval('position_business_id_seq'))::text INTO new_id;
        ELSE
            RAISE EXCEPTION 'Unknown entity type: %', entity_type;
    END CASE;
    
    RETURN new_id;
END;
$$ LANGUAGE plpgsql;

-- 创建业务ID验证函数
CREATE OR REPLACE FUNCTION validate_business_id(entity_type TEXT, business_id TEXT) 
RETURNS BOOLEAN AS $$
BEGIN
    CASE entity_type
        WHEN 'employee' THEN
            RETURN business_id ~ '^[1-9][0-9]{0,7}$';
        WHEN 'organization' THEN
            RETURN business_id ~ '^[1-9][0-9]{5}$';
        WHEN 'position' THEN
            RETURN business_id ~ '^[1-9][0-9]{6}$';
        ELSE
            RETURN FALSE;
    END CASE;
END;
$$ LANGUAGE plpgsql;

-- ==========================================
-- 第二部分: 员工表迁移
-- ==========================================

-- 检查员工表是否存在business_id字段
DO $$
BEGIN
    -- 添加business_id字段到corehr.employees表
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'corehr' 
        AND table_name = 'employees' 
        AND column_name = 'business_id'
    ) THEN
        ALTER TABLE corehr.employees ADD COLUMN business_id VARCHAR(8);
        RAISE NOTICE '已添加 business_id 字段到 corehr.employees 表';
    END IF;
    
    -- 添加business_id字段到public.employees表 (如果存在)
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'employees') THEN
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_schema = 'public' 
            AND table_name = 'employees' 
            AND column_name = 'business_id'
        ) THEN
            ALTER TABLE public.employees ADD COLUMN business_id VARCHAR(8);
            RAISE NOTICE '已添加 business_id 字段到 public.employees 表';
        END IF;
    END IF;
END $$;

-- 为现有员工生成业务ID (corehr schema)
UPDATE corehr.employees 
SET business_id = nextval('employee_business_id_seq')::text 
WHERE business_id IS NULL;

-- 为现有员工生成业务ID (public schema)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'employees') THEN
        UPDATE public.employees 
        SET business_id = nextval('employee_business_id_seq')::text 
        WHERE business_id IS NULL;
        RAISE NOTICE '已为 public.employees 表生成业务ID';
    END IF;
END $$;

-- 设置员工表约束 (corehr schema)
ALTER TABLE corehr.employees 
  ALTER COLUMN business_id SET NOT NULL,
  ADD CONSTRAINT uk_employees_business_id UNIQUE (business_id),
  ADD CONSTRAINT ck_employees_business_id CHECK (business_id ~ '^[1-9][0-9]{0,7}$');

-- 设置员工表约束 (public schema)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'employees') THEN
        ALTER TABLE public.employees 
          ALTER COLUMN business_id SET NOT NULL,
          ADD CONSTRAINT uk_public_employees_business_id UNIQUE (business_id),
          ADD CONSTRAINT ck_public_employees_business_id CHECK (business_id ~ '^[1-9][0-9]{0,7}$');
        RAISE NOTICE '已为 public.employees 表设置约束';
    END IF;
END $$;

-- 创建员工业务ID索引
CREATE INDEX IF NOT EXISTS idx_employees_business_id ON corehr.employees(business_id);
CREATE INDEX IF NOT EXISTS idx_employees_business_id_tenant ON corehr.employees(tenant_id, business_id);

-- 为public schema创建索引 (如果表存在)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'employees') THEN
        CREATE INDEX IF NOT EXISTS idx_public_employees_business_id ON public.employees(business_id);
        RAISE NOTICE '已为 public.employees 表创建索引';
    END IF;
END $$;

-- ==========================================
-- 第三部分: 组织表迁移
-- ==========================================

-- 添加business_id字段到组织表
DO $$
BEGIN
    -- corehr.organizations表
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'corehr' 
        AND table_name = 'organizations' 
        AND column_name = 'business_id'
    ) THEN
        ALTER TABLE corehr.organizations ADD COLUMN business_id VARCHAR(6);
        RAISE NOTICE '已添加 business_id 字段到 corehr.organizations 表';
    END IF;
    
    -- public.organization_units表 (如果存在)
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'organization_units') THEN
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_schema = 'public' 
            AND table_name = 'organization_units' 
            AND column_name = 'business_id'
        ) THEN
            ALTER TABLE public.organization_units ADD COLUMN business_id VARCHAR(6);
            RAISE NOTICE '已添加 business_id 字段到 public.organization_units 表';
        END IF;
    END IF;
END $$;

-- 为现有组织生成业务ID (corehr schema)
UPDATE corehr.organizations 
SET business_id = (100000 + nextval('org_business_id_seq'))::text
WHERE business_id IS NULL;

-- 为现有组织生成业务ID (public schema)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'organization_units') THEN
        UPDATE public.organization_units 
        SET business_id = (100000 + nextval('org_business_id_seq'))::text
        WHERE business_id IS NULL;
        RAISE NOTICE '已为 public.organization_units 表生成业务ID';
    END IF;
END $$;

-- 设置组织表约束 (corehr schema)
ALTER TABLE corehr.organizations 
  ALTER COLUMN business_id SET NOT NULL,
  ADD CONSTRAINT uk_organizations_business_id UNIQUE (business_id),
  ADD CONSTRAINT ck_organizations_business_id CHECK (business_id ~ '^[1-9][0-9]{5}$');

-- 设置组织表约束 (public schema)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'organization_units') THEN
        ALTER TABLE public.organization_units 
          ALTER COLUMN business_id SET NOT NULL,
          ADD CONSTRAINT uk_organization_units_business_id UNIQUE (business_id),
          ADD CONSTRAINT ck_organization_units_business_id CHECK (business_id ~ '^[1-9][0-9]{5}$');
        RAISE NOTICE '已为 public.organization_units 表设置约束';
    END IF;
END $$;

-- 创建组织业务ID索引
CREATE INDEX IF NOT EXISTS idx_organizations_business_id ON corehr.organizations(business_id);
CREATE INDEX IF NOT EXISTS idx_organizations_business_id_tenant ON corehr.organizations(tenant_id, business_id);

-- 为public schema创建索引 (如果表存在)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'organization_units') THEN
        CREATE INDEX IF NOT EXISTS idx_organization_units_business_id ON public.organization_units(business_id);
        RAISE NOTICE '已为 public.organization_units 表创建索引';
    END IF;
END $$;

-- ==========================================
-- 第四部分: 职位表迁移 (如果存在)
-- ==========================================

-- 处理职位表 (如果存在)
DO $$
BEGIN
    -- 检查并处理 corehr.positions 表
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'corehr' AND table_name = 'positions') THEN
        -- 添加business_id字段
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_schema = 'corehr' 
            AND table_name = 'positions' 
            AND column_name = 'business_id'
        ) THEN
            ALTER TABLE corehr.positions ADD COLUMN business_id VARCHAR(7);
            RAISE NOTICE '已添加 business_id 字段到 corehr.positions 表';
        END IF;
        
        -- 生成业务ID
        UPDATE corehr.positions 
        SET business_id = (1000000 + nextval('position_business_id_seq'))::text
        WHERE business_id IS NULL;
        
        -- 设置约束
        ALTER TABLE corehr.positions 
          ALTER COLUMN business_id SET NOT NULL,
          ADD CONSTRAINT uk_positions_business_id UNIQUE (business_id),
          ADD CONSTRAINT ck_positions_business_id CHECK (business_id ~ '^[1-9][0-9]{6}$');
          
        -- 创建索引
        CREATE INDEX IF NOT EXISTS idx_positions_business_id ON corehr.positions(business_id);
        
        RAISE NOTICE '已完成 corehr.positions 表的业务ID迁移';
    END IF;
    
    -- 处理其他可能的职位表
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'positions') THEN
        -- 添加business_id字段
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_schema = 'public' 
            AND table_name = 'positions' 
            AND column_name = 'business_id'
        ) THEN
            ALTER TABLE public.positions ADD COLUMN business_id VARCHAR(7);
            RAISE NOTICE '已添加 business_id 字段到 public.positions 表';
        END IF;
        
        -- 生成业务ID
        UPDATE public.positions 
        SET business_id = (1000000 + nextval('position_business_id_seq'))::text
        WHERE business_id IS NULL;
        
        -- 设置约束
        ALTER TABLE public.positions 
          ALTER COLUMN business_id SET NOT NULL,
          ADD CONSTRAINT uk_public_positions_business_id UNIQUE (business_id),
          ADD CONSTRAINT ck_public_positions_business_id CHECK (business_id ~ '^[1-9][0-9]{6}$');
          
        -- 创建索引
        CREATE INDEX IF NOT EXISTS idx_public_positions_business_id ON public.positions(business_id);
        
        RAISE NOTICE '已完成 public.positions 表的业务ID迁移';
    END IF;
END $$;

-- ==========================================
-- 第五部分: 外键关联更新
-- ==========================================

-- 更新员工表中的组织关联 (如果字段存在)
DO $$
BEGIN
    -- 检查并更新organization_id字段的引用
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'corehr' 
        AND table_name = 'employees' 
        AND column_name = 'organization_id'
    ) THEN
        -- 如果organization_id是UUID类型，需要转换为业务ID
        -- 这里假设已经有了UUID到业务ID的映射关系
        RAISE NOTICE '员工表中的组织关联字段存在，请手动检查是否需要更新引用关系';
    END IF;
    
    -- 检查并更新manager_id字段的引用
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'corehr' 
        AND table_name = 'employees' 
        AND column_name = 'manager_id'
    ) THEN
        RAISE NOTICE '员工表中的经理关联字段存在，请手动检查是否需要更新引用关系';
    END IF;
END $$;

-- ==========================================
-- 第六部分: 权限设置
-- ==========================================

-- 为应用程序角色授予新字段的权限
DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'application_role') THEN
        -- 员工表权限
        GRANT SELECT, UPDATE ON corehr.employees TO application_role;
        IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'employees') THEN
            GRANT SELECT, UPDATE ON public.employees TO application_role;
        END IF;
        
        -- 组织表权限  
        GRANT SELECT, UPDATE ON corehr.organizations TO application_role;
        IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'organization_units') THEN
            GRANT SELECT, UPDATE ON public.organization_units TO application_role;
        END IF;
        
        -- 序列权限
        GRANT ALL ON employee_business_id_seq TO application_role;
        GRANT ALL ON org_business_id_seq TO application_role;
        GRANT ALL ON position_business_id_seq TO application_role;
        
        -- 函数权限
        GRANT EXECUTE ON FUNCTION generate_business_id(TEXT) TO application_role;
        GRANT EXECUTE ON FUNCTION validate_business_id(TEXT, TEXT) TO application_role;
        
        RAISE NOTICE '已为 application_role 授予相关权限';
    ELSE
        RAISE NOTICE 'application_role 不存在，跳过权限设置';
    END IF;
END $$;

-- ==========================================
-- 第七部分: 数据完整性验证
-- ==========================================

-- 创建验证视图
CREATE OR REPLACE VIEW business_id_validation_report AS
SELECT 
    'employees' as table_name,
    'corehr' as schema_name,
    COUNT(*) as total_records,
    COUNT(business_id) as records_with_business_id,
    COUNT(DISTINCT business_id) as unique_business_ids,
    COUNT(*) - COUNT(DISTINCT business_id) as duplicates,
    COUNT(*) - COUNT(business_id) as missing_business_ids
FROM corehr.employees
UNION ALL
SELECT 
    'organizations',
    'corehr',
    COUNT(*),
    COUNT(business_id),
    COUNT(DISTINCT business_id),
    COUNT(*) - COUNT(DISTINCT business_id),
    COUNT(*) - COUNT(business_id)
FROM corehr.organizations
UNION ALL
SELECT 
    'employees' as table_name,
    'public' as schema_name,
    COALESCE(COUNT(*), 0),
    COALESCE(COUNT(business_id), 0),  
    COALESCE(COUNT(DISTINCT business_id), 0),
    COALESCE(COUNT(*) - COUNT(DISTINCT business_id), 0),
    COALESCE(COUNT(*) - COUNT(business_id), 0)
FROM (
    SELECT business_id FROM public.employees 
    WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'employees')
    UNION ALL SELECT NULL WHERE NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'employees')
) AS emp
UNION ALL
SELECT 
    'organization_units' as table_name,
    'public' as schema_name,
    COALESCE(COUNT(*), 0),
    COALESCE(COUNT(business_id), 0),
    COALESCE(COUNT(DISTINCT business_id), 0), 
    COALESCE(COUNT(*) - COUNT(DISTINCT business_id), 0),
    COALESCE(COUNT(*) - COUNT(business_id), 0)
FROM (
    SELECT business_id FROM public.organization_units
    WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'organization_units')
    UNION ALL SELECT NULL WHERE NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'organization_units')
) AS org;

-- 创建格式验证视图
CREATE OR REPLACE VIEW business_id_format_validation AS
SELECT 
    'employees' as table_name,
    'corehr' as schema_name,
    'Invalid business_id format' as issue,
    COUNT(*) as count
FROM corehr.employees 
WHERE business_id IS NOT NULL AND business_id !~ '^[1-9][0-9]{0,7}$'
UNION ALL
SELECT 
    'organizations',
    'corehr',
    'Invalid business_id format',
    COUNT(*)
FROM corehr.organizations
WHERE business_id IS NOT NULL AND business_id !~ '^[1-9][0-9]{5}$';

-- 显示验证结果
SELECT 'Business ID Validation Report' as report_type;
SELECT * FROM business_id_validation_report ORDER BY schema_name, table_name;

SELECT 'Format Validation Report' as report_type;
SELECT * FROM business_id_format_validation WHERE count > 0;

-- ==========================================
-- 第八部分: 触发器设置 (可选)
-- ==========================================

-- 创建自动生成业务ID的触发器函数
CREATE OR REPLACE FUNCTION auto_generate_business_id() 
RETURNS TRIGGER AS $$
BEGIN
    -- 员工表
    IF TG_TABLE_NAME = 'employees' AND (NEW.business_id IS NULL OR NEW.business_id = '') THEN
        NEW.business_id := generate_business_id('employee');
    END IF;
    
    -- 组织表  
    IF TG_TABLE_NAME IN ('organizations', 'organization_units') AND (NEW.business_id IS NULL OR NEW.business_id = '') THEN
        NEW.business_id := generate_business_id('organization');
    END IF;
    
    -- 职位表
    IF TG_TABLE_NAME = 'positions' AND (NEW.business_id IS NULL OR NEW.business_id = '') THEN
        NEW.business_id := generate_business_id('position');
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 为员工表创建触发器
CREATE TRIGGER trigger_employees_business_id
    BEFORE INSERT ON corehr.employees
    FOR EACH ROW
    EXECUTE FUNCTION auto_generate_business_id();

-- 为组织表创建触发器  
CREATE TRIGGER trigger_organizations_business_id
    BEFORE INSERT ON corehr.organizations
    FOR EACH ROW
    EXECUTE FUNCTION auto_generate_business_id();

-- 为public schema创建触发器 (如果表存在)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'employees') THEN
        CREATE TRIGGER trigger_public_employees_business_id
            BEFORE INSERT ON public.employees
            FOR EACH ROW
            EXECUTE FUNCTION auto_generate_business_id();
        RAISE NOTICE '已为 public.employees 创建业务ID生成触发器';
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'organization_units') THEN
        CREATE TRIGGER trigger_organization_units_business_id
            BEFORE INSERT ON public.organization_units
            FOR EACH ROW
            EXECUTE FUNCTION auto_generate_business_id();
        RAISE NOTICE '已为 public.organization_units 创建业务ID生成触发器';
    END IF;
END $$;

-- ==========================================
-- 迁移完成通知
-- ==========================================

DO $$
BEGIN
    RAISE NOTICE '==============================================';
    RAISE NOTICE '业务ID系统数据库迁移已完成!';
    RAISE NOTICE '==============================================';
    RAISE NOTICE '完成的工作:';
    RAISE NOTICE '1. 创建了业务ID生成序列和函数';
    RAISE NOTICE '2. 为现有员工和组织数据生成了业务ID';
    RAISE NOTICE '3. 设置了数据完整性约束和索引';
    RAISE NOTICE '4. 配置了自动生成业务ID的触发器';
    RAISE NOTICE '5. 创建了数据验证视图';
    RAISE NOTICE '';
    RAISE NOTICE '下一步操作:';
    RAISE NOTICE '1. 运行 SELECT * FROM business_id_validation_report; 检查迁移结果';
    RAISE NOTICE '2. 运行 SELECT * FROM business_id_format_validation; 检查格式验证';
    RAISE NOTICE '3. 更新应用程序代码以使用业务ID';
    RAISE NOTICE '4. 进行全面的功能测试';
    RAISE NOTICE '==============================================';
END $$;