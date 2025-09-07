-- 时态时间轴连贯性索引和约束
-- 基于 docs/architecture/temporal-timeline-consistency-guide.md v1.0

-- =============================================
-- 核心约束索引（按文档第123-126行）
-- =============================================

-- 1. 时点唯一约束 (tenant_id, code, effective_date)
CREATE UNIQUE INDEX IF NOT EXISTS uk_org_temporal_point 
ON organization_units(tenant_id, code, effective_date)
WHERE status != 'DELETED' AND deleted_at IS NULL;

-- 2. 当前唯一约束 (tenant_id, code) WHERE is_current = true  
CREATE UNIQUE INDEX IF NOT EXISTS uk_org_current 
ON organization_units(tenant_id, code) 
WHERE is_current = true AND status != 'DELETED' AND deleted_at IS NULL;

-- 3. 时态查询性能索引 (tenant_id, code, effective_date DESC)
CREATE INDEX IF NOT EXISTS ix_org_temporal_query 
ON organization_units(tenant_id, code, effective_date DESC)
WHERE status != 'DELETED' AND deleted_at IS NULL;

-- 4. 相邻版本锁定查询优化索引 (用于 FOR UPDATE)
CREATE INDEX IF NOT EXISTS ix_org_adjacent_versions 
ON organization_units(tenant_id, code, effective_date, record_id)
WHERE status != 'DELETED' AND deleted_at IS NULL;

-- =============================================  
-- 性能优化索引
-- =============================================

-- 5. 当前态高频查询索引
CREATE INDEX IF NOT EXISTS ix_org_current_lookup
ON organization_units(tenant_id, code, is_current)
WHERE is_current = true AND status != 'DELETED' AND deleted_at IS NULL;

-- 6. 时态边界查询索引 (effective_date, end_date)
CREATE INDEX IF NOT EXISTS ix_org_temporal_boundaries
ON organization_units(code, effective_date, end_date, is_current)
WHERE status != 'DELETED' AND deleted_at IS NULL;

-- 7. 日切任务专用索引 (昨天结束/今天生效的记录) - 移除CURRENT_DATE条件
CREATE INDEX IF NOT EXISTS ix_org_daily_transition
ON organization_units(effective_date, end_date, is_current)
WHERE status != 'DELETED' AND deleted_at IS NULL;

-- =============================================
-- 数据完整性检查函数
-- =============================================

-- 检查时态连续性的函数
CREATE OR REPLACE FUNCTION check_temporal_continuity(
    p_tenant_id UUID,
    p_code VARCHAR(7)
) RETURNS TABLE(
    issue_type TEXT,
    effective_date DATE,
    end_date DATE,
    message TEXT
) AS $$
BEGIN
    -- 检查区间重叠
    RETURN QUERY
    WITH ordered_versions AS (
        SELECT 
            effective_date,
            end_date,
            ROW_NUMBER() OVER (ORDER BY effective_date) as rn
        FROM organization_units 
        WHERE tenant_id = p_tenant_id 
          AND code = p_code 
          AND status != 'DELETED' AND deleted_at IS NULL
        ORDER BY effective_date
    ),
    version_overlaps AS (
        SELECT 
            curr.effective_date,
            curr.end_date,
            'OVERLAP' as issue_type,
            'Version overlaps with next version' as message
        FROM ordered_versions curr
        JOIN ordered_versions next ON next.rn = curr.rn + 1
        WHERE curr.end_date IS NOT NULL 
          AND curr.end_date >= next.effective_date
    )
    SELECT 
        o.issue_type::TEXT,
        o.effective_date,
        o.end_date,
        o.message::TEXT
    FROM version_overlaps o;
    
    -- 检查断档
    RETURN QUERY
    WITH ordered_versions AS (
        SELECT 
            effective_date,
            end_date,
            ROW_NUMBER() OVER (ORDER BY effective_date) as rn
        FROM organization_units 
        WHERE tenant_id = p_tenant_id 
          AND code = p_code 
          AND status != 'DELETED' AND deleted_at IS NULL
        ORDER BY effective_date
    ),
    gaps AS (
        SELECT 
            curr.effective_date,
            curr.end_date,
            'GAP' as issue_type,
            'Gap between versions' as message
        FROM ordered_versions curr
        JOIN ordered_versions next ON next.rn = curr.rn + 1
        WHERE curr.end_date IS NOT NULL 
          AND curr.end_date + INTERVAL '1 day' < next.effective_date
    )
    SELECT 
        g.issue_type::TEXT,
        g.effective_date,
        g.end_date,
        g.message::TEXT
    FROM gaps g;
END;
$$ LANGUAGE plpgsql;

-- =============================================
-- 约束验证
-- =============================================

-- 验证当前索引和约束是否正确创建
DO $$
DECLARE
    index_count INTEGER;
    unique_count INTEGER;
BEGIN
    -- 检查索引创建
    SELECT COUNT(*) INTO index_count
    FROM pg_indexes 
    WHERE tablename = 'organization_units' 
      AND indexname IN (
          'uk_org_temporal_point',
          'uk_org_current',
          'ix_org_temporal_query',
          'ix_org_adjacent_versions',
          'ix_org_current_lookup',
          'ix_org_temporal_boundaries',
          'ix_org_daily_transition'
      );
    
    -- 检查唯一约束
    SELECT COUNT(*) INTO unique_count
    FROM pg_indexes
    WHERE tablename = 'organization_units'
      AND indexdef LIKE '%UNIQUE%'
      AND indexname IN ('uk_org_temporal_point', 'uk_org_current');
      
    RAISE NOTICE '时态连续性索引创建完成: 总计 % 个索引, % 个唯一约束', index_count, unique_count;
    
    IF index_count < 7 THEN
        RAISE WARNING '部分索引创建失败，请检查日志';
    END IF;
    
    IF unique_count < 2 THEN
        RAISE WARNING '部分唯一约束创建失败，请检查约束冲突';
    END IF;
END $$;

-- =============================================
-- 完成报告
-- =============================================
SELECT 
    '时态时间轴连贯性索引部署完成' as status,
    'Temporal timeline consistency indexes deployed' as details,
    NOW() as completed_at;