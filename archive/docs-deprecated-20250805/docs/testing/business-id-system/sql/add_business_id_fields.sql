-- 业务ID字段添加脚本
-- add_business_id_fields.sql

-- 为employees表添加business_id字段
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'employees' AND column_name = 'business_id') THEN
        ALTER TABLE employees ADD COLUMN business_id VARCHAR(5) UNIQUE;
        CREATE INDEX idx_employees_business_id ON employees(business_id);
        RAISE NOTICE '✅ employees表business_id字段已添加';
    ELSE
        RAISE NOTICE 'ℹ️ employees表business_id字段已存在';
    END IF;
END $$;

-- 为organization_units表添加business_id字段  
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'organization_units' AND column_name = 'business_id') THEN
        ALTER TABLE organization_units ADD COLUMN business_id VARCHAR(6) UNIQUE;
        CREATE INDEX idx_organization_units_business_id ON organization_units(business_id);
        RAISE NOTICE '✅ organization_units表business_id字段已添加';
    ELSE
        RAISE NOTICE 'ℹ️ organization_units表business_id字段已存在';
    END IF;
END $$;

-- 为positions表添加business_id字段
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'positions' AND column_name = 'business_id') THEN
        ALTER TABLE positions ADD COLUMN business_id VARCHAR(7) UNIQUE;
        CREATE INDEX idx_positions_business_id ON positions(business_id);
        RAISE NOTICE '✅ positions表business_id字段已添加';
    ELSE
        RAISE NOTICE 'ℹ️ positions表business_id字段已存在';
    END IF;
END $$;