-- Migration: Phase 2 Enhanced Temporal Query Optimizations
-- Version: 2.0.0
-- Date: 2025-07-28
-- Description: Advanced temporal query optimizations and performance enhancements

-- ============================================================================
-- PART 1: ADVANCED INDEXING STRATEGY
-- ============================================================================

-- Drop existing basic indexes (will be replaced with optimized ones)
DROP INDEX IF EXISTS idx_position_history_temporal;
DROP INDEX IF EXISTS idx_position_history_effective_date;
DROP INDEX IF EXISTS idx_position_history_date_range;

-- Create optimized composite index for temporal range queries
-- This index covers the most common query pattern: tenant + employee + date range
CREATE INDEX CONCURRENTLY idx_position_history_temporal_optimized 
ON position_history (tenant_id, employee_id, effective_date DESC, end_date DESC)
INCLUDE (position_title, department, employment_type, is_retroactive);

-- Create specialized index for current position lookups (most frequent query)
CREATE UNIQUE INDEX CONCURRENTLY idx_position_history_current_optimized
ON position_history (tenant_id, employee_id, effective_date DESC) 
WHERE end_date IS NULL
INCLUDE (position_title, department, job_level, location, employment_type, reports_to_employee_id);

-- Create index for as-of-date queries (point-in-time lookups)
CREATE INDEX CONCURRENTLY idx_position_history_as_of_date
ON position_history (tenant_id, employee_id, effective_date DESC, end_date ASC)
WHERE end_date IS NOT NULL OR end_date IS NULL;

-- Create index for timeline queries with department filtering
CREATE INDEX CONCURRENTLY idx_position_history_department_timeline
ON position_history (tenant_id, department, effective_date DESC)
INCLUDE (employee_id, position_title, job_level, is_retroactive);

-- Create index for job level analysis queries
CREATE INDEX CONCURRENTLY idx_position_history_job_level_timeline
ON position_history (tenant_id, job_level, effective_date DESC)
WHERE job_level IS NOT NULL
INCLUDE (employee_id, department, position_title);

-- Create index for retroactive change analysis
CREATE INDEX CONCURRENTLY idx_position_history_retroactive_analysis
ON position_history (tenant_id, is_retroactive, created_at DESC, effective_date DESC)
WHERE is_retroactive = true
INCLUDE (employee_id, change_reason, position_title, department);

-- Create index for manager/reporting relationship queries
CREATE INDEX CONCURRENTLY idx_position_history_reporting_timeline
ON position_history (tenant_id, reports_to_employee_id, effective_date DESC)
WHERE reports_to_employee_id IS NOT NULL AND end_date IS NULL
INCLUDE (employee_id, position_title, department);

-- Create index for overlapping period detection (temporal consistency)
CREATE INDEX CONCURRENTLY idx_position_history_overlap_detection
ON position_history (tenant_id, employee_id, effective_date, COALESCE(end_date, 'infinity'::timestamp));

-- ============================================================================
-- PART 2: ADVANCED QUERY OPTIMIZATION FUNCTIONS
-- ============================================================================

-- Function for optimized as-of-date position lookup
CREATE OR REPLACE FUNCTION get_position_as_of_date(
    p_tenant_id UUID,
    p_employee_id UUID,
    p_as_of_date TIMESTAMP WITH TIME ZONE
) RETURNS TABLE (
    position_history_id UUID,
    position_title VARCHAR(100),
    department VARCHAR(100),
    job_level VARCHAR(50),
    location VARCHAR(100),
    employment_type VARCHAR(20),
    reports_to_employee_id UUID,
    effective_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE,
    is_retroactive BOOLEAN,
    min_salary DECIMAL(15,2),
    max_salary DECIMAL(15,2),
    currency CHAR(3)
) 
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ph.id,
        ph.position_title,
        ph.department,
        ph.job_level,
        ph.location,
        ph.employment_type,
        ph.reports_to_employee_id,
        ph.effective_date,
        ph.end_date,
        ph.is_retroactive,
        ph.min_salary,
        ph.max_salary,
        ph.currency
    FROM position_history ph
    WHERE ph.tenant_id = p_tenant_id
      AND ph.employee_id = p_employee_id
      AND ph.effective_date <= p_as_of_date
      AND (ph.end_date IS NULL OR ph.end_date > p_as_of_date)
    ORDER BY ph.effective_date DESC
    LIMIT 1;
END;
$$;

-- Function for optimized position timeline with filters
CREATE OR REPLACE FUNCTION get_filtered_position_timeline(
    p_tenant_id UUID,
    p_employee_ids UUID[] DEFAULT NULL,
    p_departments TEXT[] DEFAULT NULL,
    p_job_levels TEXT[] DEFAULT NULL,
    p_start_date TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    p_end_date TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    p_include_retroactive BOOLEAN DEFAULT true,
    p_limit INTEGER DEFAULT 1000,
    p_offset INTEGER DEFAULT 0
) RETURNS TABLE (
    position_history_id UUID,
    employee_id UUID,
    position_title VARCHAR(100),
    department VARCHAR(100),
    job_level VARCHAR(50),
    effective_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE,
    is_retroactive BOOLEAN,
    change_reason TEXT
)
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ph.id,
        ph.employee_id,
        ph.position_title,
        ph.department,
        ph.job_level,
        ph.effective_date,
        ph.end_date,
        ph.is_retroactive,
        ph.change_reason
    FROM position_history ph
    WHERE ph.tenant_id = p_tenant_id
      AND (p_employee_ids IS NULL OR ph.employee_id = ANY(p_employee_ids))
      AND (p_departments IS NULL OR ph.department = ANY(p_departments))
      AND (p_job_levels IS NULL OR ph.job_level = ANY(p_job_levels))
      AND (p_start_date IS NULL OR ph.end_date IS NULL OR ph.end_date >= p_start_date)
      AND (p_end_date IS NULL OR ph.effective_date <= p_end_date)
      AND (p_include_retroactive OR ph.is_retroactive = false)
    ORDER BY ph.effective_date DESC
    LIMIT p_limit
    OFFSET p_offset;
END;
$$;

-- Function for position gap detection
CREATE OR REPLACE FUNCTION detect_position_gaps(
    p_tenant_id UUID,
    p_employee_id UUID
) RETURNS TABLE (
    gap_start TIMESTAMP WITH TIME ZONE,
    gap_end TIMESTAMP WITH TIME ZONE,
    gap_duration INTERVAL,
    previous_position_id UUID,
    next_position_id UUID
)
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
AS $$
BEGIN
    RETURN QUERY
    WITH position_sequence AS (
        SELECT 
            id,
            effective_date,
            end_date,
            LAG(end_date) OVER (ORDER BY effective_date) as prev_end_date,
            LAG(id) OVER (ORDER BY effective_date) as prev_position_id
        FROM position_history
        WHERE tenant_id = p_tenant_id 
          AND employee_id = p_employee_id
        ORDER BY effective_date
    )
    SELECT 
        (prev_end_date + INTERVAL '1 day')::TIMESTAMP WITH TIME ZONE as gap_start,
        (effective_date - INTERVAL '1 day')::TIMESTAMP WITH TIME ZONE as gap_end,
        (effective_date - prev_end_date - INTERVAL '1 day')::INTERVAL as gap_duration,
        prev_position_id,
        id as next_position_id
    FROM position_sequence
    WHERE prev_end_date IS NOT NULL
      AND prev_end_date < effective_date - INTERVAL '1 day';
END;
$$;

-- Function for position overlap detection
CREATE OR REPLACE FUNCTION detect_position_overlaps(
    p_tenant_id UUID,
    p_employee_id UUID
) RETURNS TABLE (
    overlap_start TIMESTAMP WITH TIME ZONE,
    overlap_end TIMESTAMP WITH TIME ZONE,
    overlap_duration INTERVAL,
    position1_id UUID,
    position2_id UUID
)
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
AS $$
BEGIN
    RETURN QUERY
    SELECT 
        GREATEST(p1.effective_date, p2.effective_date) as overlap_start,
        LEAST(
            COALESCE(p1.end_date, 'infinity'::timestamp), 
            COALESCE(p2.end_date, 'infinity'::timestamp)
        ) as overlap_end,
        LEAST(
            COALESCE(p1.end_date, 'infinity'::timestamp), 
            COALESCE(p2.end_date, 'infinity'::timestamp)
        ) - GREATEST(p1.effective_date, p2.effective_date) as overlap_duration,
        p1.id as position1_id,
        p2.id as position2_id
    FROM position_history p1
    INNER JOIN position_history p2 ON (
        p1.tenant_id = p2.tenant_id 
        AND p1.employee_id = p2.employee_id 
        AND p1.id < p2.id
    )
    WHERE p1.tenant_id = p_tenant_id
      AND p1.employee_id = p_employee_id
      AND p1.effective_date < COALESCE(p2.end_date, 'infinity'::timestamp)
      AND COALESCE(p1.end_date, 'infinity'::timestamp) > p2.effective_date;
END;
$$;

-- ============================================================================
-- PART 3: PERFORMANCE MONITORING AND ANALYTICS
-- ============================================================================

-- Create table for query performance metrics
CREATE TABLE IF NOT EXISTS temporal_query_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    query_type VARCHAR(50) NOT NULL,
    execution_time_ms INTEGER NOT NULL,
    records_scanned INTEGER NOT NULL,
    records_returned INTEGER NOT NULL,
    query_parameters JSONB,
    executed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    user_id UUID
);

-- Create index for metrics analysis
CREATE INDEX idx_temporal_query_metrics_analysis
ON temporal_query_metrics (tenant_id, query_type, executed_at DESC);

-- Create index for performance monitoring
CREATE INDEX idx_temporal_query_metrics_performance
ON temporal_query_metrics (executed_at DESC, execution_time_ms DESC);

-- Function to log query metrics
CREATE OR REPLACE FUNCTION log_temporal_query_metrics(
    p_tenant_id UUID,
    p_query_type VARCHAR(50),
    p_execution_time_ms INTEGER,
    p_records_scanned INTEGER,
    p_records_returned INTEGER,
    p_query_parameters JSONB DEFAULT NULL,
    p_user_id UUID DEFAULT NULL
) RETURNS UUID
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_metric_id UUID;
BEGIN
    INSERT INTO temporal_query_metrics (
        tenant_id, query_type, execution_time_ms, 
        records_scanned, records_returned, query_parameters, user_id
    ) VALUES (
        p_tenant_id, p_query_type, p_execution_time_ms,
        p_records_scanned, p_records_returned, p_query_parameters, p_user_id
    )
    RETURNING id INTO v_metric_id;
    
    RETURN v_metric_id;
END;
$$;

-- ============================================================================
-- PART 4: MATERIALIZED VIEWS FOR COMMON QUERIES
-- ============================================================================

-- Materialized view for current positions (most frequently accessed)
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_current_positions AS
SELECT 
    ph.tenant_id,
    ph.employee_id,
    ph.id as position_history_id,
    ph.position_title,
    ph.department,
    ph.job_level,
    ph.location,
    ph.employment_type,
    ph.reports_to_employee_id,
    ph.effective_date,
    ph.min_salary,
    ph.max_salary,
    ph.currency,
    p.legal_name as employee_name,
    p.email as employee_email,
    manager.legal_name as manager_name
FROM position_history ph
JOIN person p ON ph.tenant_id = p.tenant_id AND ph.employee_id = p.id
LEFT JOIN person manager ON ph.tenant_id = manager.tenant_id 
    AND ph.reports_to_employee_id = manager.id
WHERE ph.end_date IS NULL;

-- Create indexes on materialized view
CREATE UNIQUE INDEX idx_mv_current_positions_primary
ON mv_current_positions (tenant_id, employee_id);

CREATE INDEX idx_mv_current_positions_department
ON mv_current_positions (tenant_id, department);

CREATE INDEX idx_mv_current_positions_manager
ON mv_current_positions (tenant_id, reports_to_employee_id)
WHERE reports_to_employee_id IS NOT NULL;

-- Materialized view for position change analytics
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_position_analytics AS
WITH position_changes AS (
    SELECT 
        ph.tenant_id,
        ph.employee_id,
        ph.department,
        ph.job_level,
        ph.effective_date,
        ph.is_retroactive,
        ph.created_at,
        LAG(ph.position_title) OVER (
            PARTITION BY ph.tenant_id, ph.employee_id 
            ORDER BY ph.effective_date
        ) as previous_position_title,
        LAG(ph.department) OVER (
            PARTITION BY ph.tenant_id, ph.employee_id 
            ORDER BY ph.effective_date
        ) as previous_department,
        CASE 
            WHEN LAG(ph.position_title) OVER (
                PARTITION BY ph.tenant_id, ph.employee_id 
                ORDER BY ph.effective_date
            ) IS NULL THEN 'INITIAL_HIRE'
            WHEN LAG(ph.department) OVER (
                PARTITION BY ph.tenant_id, ph.employee_id 
                ORDER BY ph.effective_date
            ) != ph.department THEN 'TRANSFER'
            ELSE 'PROMOTION'
        END as change_type
    FROM position_history ph
)
SELECT 
    tenant_id,
    DATE_TRUNC('month', effective_date) as month,
    department,
    job_level,
    change_type,
    COUNT(*) as change_count,
    COUNT(*) FILTER (WHERE is_retroactive) as retroactive_count,
    AVG(EXTRACT(days FROM created_at - effective_date)) as avg_retroactive_days
FROM position_changes
WHERE effective_date >= CURRENT_DATE - INTERVAL '2 years'
GROUP BY tenant_id, DATE_TRUNC('month', effective_date), department, job_level, change_type;

-- Create indexes on analytics view
CREATE INDEX idx_mv_position_analytics_tenant_month
ON mv_position_analytics (tenant_id, month DESC);

CREATE INDEX idx_mv_position_analytics_department
ON mv_position_analytics (tenant_id, department, month DESC);

-- ============================================================================
-- PART 5: AUTOMATED MAINTENANCE AND OPTIMIZATION
-- ============================================================================

-- Function to refresh materialized views
CREATE OR REPLACE FUNCTION refresh_temporal_materialized_views()
RETURNS VOID
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_current_positions;
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_position_analytics;
    
    -- Log the refresh
    INSERT INTO temporal_query_metrics (
        tenant_id, query_type, execution_time_ms, 
        records_scanned, records_returned
    ) VALUES (
        '00000000-0000-0000-0000-000000000000'::UUID, 
        'MATERIALIZED_VIEW_REFRESH', 
        0, 0, 0
    );
END;
$$;

-- Function for automated index maintenance
CREATE OR REPLACE FUNCTION maintain_temporal_indexes()
RETURNS VOID
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
BEGIN
    -- Reindex temporal indexes if fragmentation is high
    -- This would be called by a maintenance job
    
    -- Update table statistics
    ANALYZE position_history;
    ANALYZE temporal_query_metrics;
    
    -- Log maintenance activity
    INSERT INTO temporal_query_metrics (
        tenant_id, query_type, execution_time_ms, 
        records_scanned, records_returned
    ) VALUES (
        '00000000-0000-0000-0000-000000000000'::UUID, 
        'INDEX_MAINTENANCE', 
        0, 0, 0
    );
END;
$$;

-- ============================================================================
-- PART 6: QUERY OPTIMIZATION HINTS AND CONFIGURATION
-- ============================================================================

-- Set optimal configuration for temporal queries
ALTER TABLE position_history SET (
    fillfactor = 85,  -- Leave room for updates
    parallel_workers = 4,
    autovacuum_vacuum_scale_factor = 0.05,
    autovacuum_analyze_scale_factor = 0.02
);

-- Set optimal configuration for metrics table
ALTER TABLE temporal_query_metrics SET (
    fillfactor = 100,  -- Insert-only table
    parallel_workers = 2,
    autovacuum_vacuum_scale_factor = 0.1,
    autovacuum_analyze_scale_factor = 0.05
);

-- ============================================================================
-- PART 7: PERMISSIONS AND SECURITY
-- ============================================================================

-- Grant permissions for new functions
GRANT EXECUTE ON FUNCTION get_position_as_of_date(UUID, UUID, TIMESTAMP WITH TIME ZONE) TO application_role;
GRANT EXECUTE ON FUNCTION get_filtered_position_timeline(UUID, UUID[], TEXT[], TEXT[], TIMESTAMP WITH TIME ZONE, TIMESTAMP WITH TIME ZONE, BOOLEAN, INTEGER, INTEGER) TO application_role;
GRANT EXECUTE ON FUNCTION detect_position_gaps(UUID, UUID) TO application_role;
GRANT EXECUTE ON FUNCTION detect_position_overlaps(UUID, UUID) TO application_role;
GRANT EXECUTE ON FUNCTION log_temporal_query_metrics(UUID, VARCHAR, INTEGER, INTEGER, INTEGER, JSONB, UUID) TO application_role;

-- Grant permissions for materialized views
GRANT SELECT ON mv_current_positions TO application_role;
GRANT SELECT ON mv_position_analytics TO application_role;

-- Grant permissions for metrics table
GRANT SELECT, INSERT ON temporal_query_metrics TO application_role;

-- Create RLS policies for metrics table
ALTER TABLE temporal_query_metrics ENABLE ROW LEVEL SECURITY;

CREATE POLICY temporal_query_metrics_tenant_isolation ON temporal_query_metrics
    FOR ALL TO application_role
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID OR tenant_id = '00000000-0000-0000-0000-000000000000'::UUID);

-- ============================================================================
-- PART 8: MONITORING AND ALERTING SETUP
-- ============================================================================

-- Create view for query performance monitoring
CREATE OR REPLACE VIEW v_temporal_query_performance AS
SELECT 
    tenant_id,
    query_type,
    COUNT(*) as query_count,
    AVG(execution_time_ms) as avg_execution_time_ms,
    MIN(execution_time_ms) as min_execution_time_ms,
    MAX(execution_time_ms) as max_execution_time_ms,
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY execution_time_ms) as p95_execution_time_ms,
    AVG(records_returned::float / NULLIF(records_scanned, 0)) as avg_efficiency_ratio,
    DATE_TRUNC('hour', executed_at) as hour
FROM temporal_query_metrics
WHERE executed_at >= CURRENT_TIMESTAMP - INTERVAL '24 hours'
GROUP BY tenant_id, query_type, DATE_TRUNC('hour', executed_at)
ORDER BY hour DESC, avg_execution_time_ms DESC;

-- Grant permissions for performance view
GRANT SELECT ON v_temporal_query_performance TO application_role;

-- Add helpful comments
COMMENT ON FUNCTION get_position_as_of_date IS 'Optimized function for point-in-time position lookups with sub-10ms performance target';
COMMENT ON FUNCTION get_filtered_position_timeline IS 'Advanced timeline query function with comprehensive filtering and pagination';
COMMENT ON FUNCTION detect_position_gaps IS 'Identifies gaps in employment history for data quality analysis';
COMMENT ON FUNCTION detect_position_overlaps IS 'Detects overlapping position periods for temporal consistency validation';
COMMENT ON MATERIALIZED VIEW mv_current_positions IS 'Pre-computed current positions for sub-millisecond lookups';
COMMENT ON MATERIALIZED VIEW mv_position_analytics IS 'Pre-aggregated position change analytics for dashboard queries';

COMMIT;