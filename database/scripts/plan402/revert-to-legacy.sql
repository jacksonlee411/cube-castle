-- Plan 402D · 回滚到 legacy 写路径
-- 该脚本撤销 drop-legacy-write-paths.sql 注入的写锁，允许命令服务重新写入
-- organization_units / positions 表，便于紧急切回旧实现。
--
-- 用法：
--   psql "$DATABASE_URL" -f database/scripts/plan402/revert-to-legacy.sql \
--     | tee logs/plan402/rollback/$(date +%Y%m%d-%H%M%S)-revert-to-legacy.log

INSERT INTO plan402_runtime_flags (id, enforce_legacy_lock)
VALUES (1, true)
ON CONFLICT (id) DO NOTHING;

UPDATE plan402_runtime_flags
SET enforce_legacy_lock = false,
    updated_at = now()
WHERE id = 1;

\echo '🟡 Plan 402D · legacy 写锁已关闭（enforce_legacy_lock=false）'
SELECT enforce_legacy_lock AS legacy_lock_enabled,
       updated_at
FROM plan402_runtime_flags
WHERE id = 1;

-- 移除 READONLY 触发器，恢复写入能力
DO $$
DECLARE
  target_tables CONSTANT text[] := ARRAY['organization_units', 'positions'];
  target text;
BEGIN
  FOREACH target IN ARRAY target_tables LOOP
    BEGIN
      EXECUTE format('DROP TRIGGER IF EXISTS plan402_%I_readonly_guard ON %I', target, target);
    EXCEPTION
      WHEN undefined_table THEN
        RAISE NOTICE 'Legacy table % does not exist, skipping trigger removal', target;
    END;
  END LOOP;
END;
$$;

-- 恢复 user 角色的写权限，便于旧命令层继续运行
DO $$
DECLARE
  target_tables CONSTANT text[] := ARRAY['organization_units', 'positions'];
  target_role CONSTANT text := 'user';
  target text;
BEGIN
  FOREACH target IN ARRAY target_tables LOOP
    BEGIN
      EXECUTE format('GRANT INSERT, UPDATE, DELETE ON %I TO %I', target, target_role);
      EXECUTE format('COMMENT ON TABLE %I IS %L', target,
        'Plan 402D rollback: 恢复 legacy 写入（参见 docs/development-plans/402D-cutover-and-recovery.md §D3）');
    EXCEPTION
      WHEN undefined_table THEN
        RAISE NOTICE 'Legacy table % does not exist, skipping privilege restore', target;
    END;
  END LOOP;
END;
$$;

\echo '✳️ Legacy table privileges after rollback:'
SELECT tab.relname                                                    AS table_name,
       rol.rolname                                                    AS role_name,
       has_table_privilege(rol.rolname, tab.oid, 'insert')            AS can_insert,
       has_table_privilege(rol.rolname, tab.oid, 'update')            AS can_update,
       has_table_privilege(rol.rolname, tab.oid, 'delete')            AS can_delete
FROM pg_class tab
JOIN pg_namespace ns ON ns.oid = tab.relnamespace
JOIN (SELECT 'user'::name AS rolname) rol ON TRUE
WHERE tab.relname IN ('organization_units', 'positions')
  AND ns.nspname = 'public';
