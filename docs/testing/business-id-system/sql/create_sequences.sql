-- 业务ID序列创建脚本
-- create_sequences.sql

-- 创建员工业务ID序列
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_sequences WHERE sequencename = 'employee_business_id_seq') THEN
        CREATE SEQUENCE employee_business_id_seq START 1 INCREMENT 1 MINVALUE 1 MAXVALUE 99999;
        RAISE NOTICE '✅ 员工业务ID序列已创建';
    ELSE
        RAISE NOTICE 'ℹ️ 员工业务ID序列已存在';
    END IF;
END $$;

-- 创建组织业务ID序列
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_sequences WHERE sequencename = 'org_business_id_seq') THEN
        CREATE SEQUENCE org_business_id_seq START 1 INCREMENT 1 MINVALUE 1 MAXVALUE 899999;
        RAISE NOTICE '✅ 组织业务ID序列已创建';
    ELSE
        RAISE NOTICE 'ℹ️ 组织业务ID序列已存在';
    END IF;
END $$;

-- 创建职位业务ID序列  
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_sequences WHERE sequencename = 'position_business_id_seq') THEN
        CREATE SEQUENCE position_business_id_seq START 1 INCREMENT 1 MINVALUE 1 MAXVALUE 8999999;
        RAISE NOTICE '✅ 职位业务ID序列已创建';
    ELSE
        RAISE NOTICE 'ℹ️ 职位业务ID序列已存在';
    END IF;
END $$;

-- 验证序列创建结果
SELECT 
    schemaname,
    sequencename, 
    start_value,
    min_value,
    max_value,
    increment_by
FROM pg_sequences 
WHERE sequencename LIKE '%business_id%'
ORDER BY sequencename;