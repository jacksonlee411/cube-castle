-- 初始化 PostgreSQL 扩展，供 Goose 迁移与测试使用
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Goose 元数据表在 goose up 期间自动创建，如需自定义可在迁移中维护
