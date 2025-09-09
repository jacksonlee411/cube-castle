-- P1级系统修复脚本
-- 基于CLAUDE.md原则：诚实分析+健壮修复
-- 修复audit_logs表缺失字段问题

-- 1. 检查当前audit_logs表结构
\d audit_logs;

-- 2. 添加缺失的business_entity_type字段（如果不存在）
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'audit_logs' 
        AND column_name = 'business_entity_type'
    ) THEN
        ALTER TABLE audit_logs 
        ADD COLUMN business_entity_type VARCHAR(50) DEFAULT 'ORGANIZATION';
        
        -- 为现有记录设置默认值
        UPDATE audit_logs 
        SET business_entity_type = 'ORGANIZATION' 
        WHERE business_entity_type IS NULL;
        
        RAISE NOTICE '✅ 添加business_entity_type字段成功';
    ELSE
        RAISE NOTICE '✅ business_entity_type字段已存在';
    END IF;
END $$;

-- 3. 验证修复结果
SELECT COUNT(*) as total_records, 
       COUNT(business_entity_type) as with_entity_type
FROM audit_logs;