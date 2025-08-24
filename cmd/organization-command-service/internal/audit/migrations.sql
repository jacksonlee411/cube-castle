-- 结构化审计日志表创建脚本
-- 支持详细的操作审计、字段变更追踪和业务上下文记录

CREATE TABLE IF NOT EXISTS audit_logs (
    -- 基础标识字段
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    
    -- 事件分类
    event_type VARCHAR(20) NOT NULL CHECK (event_type IN ('CREATE', 'UPDATE', 'DELETE', 'SUSPEND', 'ACTIVATE', 'QUERY', 'VALIDATION', 'AUTHENTICATION', 'ERROR')),
    resource_type VARCHAR(50) NOT NULL CHECK (resource_type IN ('ORGANIZATION', 'HIERARCHY', 'USER', 'SYSTEM')),
    resource_id VARCHAR(100) NOT NULL,
    
    -- 操作者信息
    actor_id VARCHAR(100) NOT NULL,
    actor_type VARCHAR(20) NOT NULL CHECK (actor_type IN ('USER', 'SYSTEM', 'SERVICE')),
    action_name VARCHAR(100) NOT NULL,
    
    -- 请求追踪
    request_id VARCHAR(100) NOT NULL,
    ip_address INET,
    user_agent TEXT,
    
    -- 时间戳
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- 操作结果
    success BOOLEAN NOT NULL DEFAULT true,
    error_code VARCHAR(50),
    error_message TEXT,
    
    -- 结构化数据字段 (JSONB格式)
    request_data JSONB DEFAULT '{}',
    response_data JSONB DEFAULT '{}',
    changes JSONB DEFAULT '[]', -- 字段变更数组
    business_context JSONB DEFAULT '{}', -- 业务上下文
    
    -- 审计元数据
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 创建索引优化查询性能
CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_resource ON audit_logs(tenant_id, resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_event_type ON audit_logs(event_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor ON audit_logs(actor_id, actor_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_request_id ON audit_logs(request_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_success ON audit_logs(success);

-- 创建复合索引优化常见查询场景
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_timestamp ON audit_logs(resource_type, resource_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_timestamp ON audit_logs(tenant_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_timestamp ON audit_logs(actor_id, timestamp DESC);

-- 创建GIN索引优化JSONB字段查询
CREATE INDEX IF NOT EXISTS idx_audit_logs_request_data_gin ON audit_logs USING GIN(request_data);
CREATE INDEX IF NOT EXISTS idx_audit_logs_response_data_gin ON audit_logs USING GIN(response_data);
CREATE INDEX IF NOT EXISTS idx_audit_logs_changes_gin ON audit_logs USING GIN(changes);
CREATE INDEX IF NOT EXISTS idx_audit_logs_business_context_gin ON audit_logs USING GIN(business_context);

-- 表注释说明
COMMENT ON TABLE audit_logs IS '结构化审计日志表：记录所有系统操作的详细审计信息，支持字段级变更追踪和业务上下文';

COMMENT ON COLUMN audit_logs.id IS '审计记录唯一标识UUID';
COMMENT ON COLUMN audit_logs.tenant_id IS '租户ID，支持多租户审计隔离';
COMMENT ON COLUMN audit_logs.event_type IS '事件类型：CREATE, UPDATE, DELETE, SUSPEND, ACTIVATE, QUERY, VALIDATION, AUTHENTICATION, ERROR';
COMMENT ON COLUMN audit_logs.resource_type IS '资源类型：ORGANIZATION, HIERARCHY, USER, SYSTEM';
COMMENT ON COLUMN audit_logs.resource_id IS '资源标识符，通常是业务实体的主键';
COMMENT ON COLUMN audit_logs.actor_id IS '操作执行者ID（用户ID、系统服务名等）';
COMMENT ON COLUMN audit_logs.actor_type IS '操作者类型：USER, SYSTEM, SERVICE';
COMMENT ON COLUMN audit_logs.action_name IS '具体操作名称（如CreateOrganization, UpdateOrganization）';
COMMENT ON COLUMN audit_logs.request_id IS '请求追踪ID，用于关联同一个请求的多个审计记录';
COMMENT ON COLUMN audit_logs.ip_address IS '客户端IP地址';
COMMENT ON COLUMN audit_logs.user_agent IS '用户代理字符串';
COMMENT ON COLUMN audit_logs.timestamp IS '操作发生的时间戳';
COMMENT ON COLUMN audit_logs.success IS '操作是否成功';
COMMENT ON COLUMN audit_logs.error_code IS '错误代码（操作失败时）';
COMMENT ON COLUMN audit_logs.error_message IS '错误消息（操作失败时）';
COMMENT ON COLUMN audit_logs.request_data IS '请求数据的JSON表示';
COMMENT ON COLUMN audit_logs.response_data IS '响应数据的JSON表示';
COMMENT ON COLUMN audit_logs.changes IS '字段变更记录的JSON数组，包含oldValue和newValue';
COMMENT ON COLUMN audit_logs.business_context IS '业务上下文信息的JSON对象，包含额外的业务相关数据';

-- 数据保留策略 (可选：根据业务需求配置)
-- 创建分区表按月分区（生产环境建议）
-- CREATE TABLE audit_logs_2025_08 PARTITION OF audit_logs FOR VALUES FROM ('2025-08-01') TO ('2025-09-01');

-- 创建审计日志清理函数（生产环境可配置定期清理）
CREATE OR REPLACE FUNCTION cleanup_old_audit_logs(retention_days INTEGER DEFAULT 365) 
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM audit_logs 
    WHERE timestamp < NOW() - INTERVAL '1 day' * retention_days;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION cleanup_old_audit_logs IS '清理指定天数之前的审计日志记录，默认保留365天';

-- 创建审计统计视图
CREATE OR REPLACE VIEW audit_logs_stats AS
SELECT 
    tenant_id,
    event_type,
    resource_type,
    actor_type,
    DATE_TRUNC('day', timestamp) as audit_date,
    COUNT(*) as event_count,
    COUNT(*) FILTER (WHERE success = true) as success_count,
    COUNT(*) FILTER (WHERE success = false) as error_count,
    COUNT(DISTINCT actor_id) as unique_actors,
    COUNT(DISTINCT resource_id) as unique_resources
FROM audit_logs 
GROUP BY tenant_id, event_type, resource_type, actor_type, DATE_TRUNC('day', timestamp);

COMMENT ON VIEW audit_logs_stats IS '审计日志统计视图：按租户、事件类型、资源类型统计每日审计数据';

-- 示例查询语句 
/*
-- 查询特定组织的审计历史
SELECT * FROM audit_logs 
WHERE resource_type = 'ORGANIZATION' 
  AND resource_id = 'ORG001' 
  AND tenant_id = 'your-tenant-uuid'
ORDER BY timestamp DESC 
LIMIT 50;

-- 查询特定用户的操作记录
SELECT * FROM audit_logs 
WHERE actor_id = 'user123' 
  AND timestamp >= NOW() - INTERVAL '7 days'
ORDER BY timestamp DESC;

-- 查询失败的操作记录
SELECT * FROM audit_logs 
WHERE success = false 
  AND timestamp >= NOW() - INTERVAL '1 day'
ORDER BY timestamp DESC;

-- 查询字段变更记录
SELECT 
    resource_id,
    action_name,
    timestamp,
    changes
FROM audit_logs 
WHERE event_type = 'UPDATE' 
  AND jsonb_array_length(changes) > 0
ORDER BY timestamp DESC;

-- 统计每日操作量
SELECT 
    DATE_TRUNC('day', timestamp) as date,
    event_type,
    COUNT(*) as count
FROM audit_logs 
WHERE timestamp >= NOW() - INTERVAL '30 days'
GROUP BY DATE_TRUNC('day', timestamp), event_type
ORDER BY date DESC, count DESC;
*/