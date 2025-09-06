-- fix-audit-recordid-misplaced.sql
-- 目的：修复审计表 audit_logs 中 record_id 误填为 actor_id（如 550e8400-... 系统用户）的数据
-- 条件：当存在列 actor_id / resource_id / record_id，且 record_id = actor_id 且 resource_id 不为空时，将 record_id 更正为 resource_id

BEGIN;

-- 报表：统计疑似误填的行数（record_id = actor_id）
-- 注意：若不存在这些列，执行将失败；请根据实际表结构调整。
-- SELECT count(*) FROM audit_logs WHERE record_id = actor_id AND resource_id IS NOT NULL;

UPDATE audit_logs
SET record_id = resource_id::uuid
WHERE record_id = actor_id::uuid
  AND resource_id IS NOT NULL
  AND resource_id ~* '^[0-9a-f-]{36}$'
  AND record_id <> resource_id::uuid;

COMMIT;

-- 可选：再运行 scripts/validate-audit-recordid-consistency.sql 检查是否还有异常

