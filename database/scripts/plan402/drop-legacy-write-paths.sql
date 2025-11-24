-- Plan 402D · 切换窗口最小回收脚本
-- 将旧组织/职位表切换为只读，防止上线期间出现遗留写入。
-- 该脚本可重复执行；若在 402E 彻底清理旧表前需要临时恢复写权限，请更新本脚本并重新执行。

-- 全局防护函数：拒绝所有 DML
CREATE OR REPLACE FUNCTION plan402_legacy_readonly_guard()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  RAISE EXCEPTION 'Plan 402D: legacy table % is read-only. Update SOM tables instead.', TG_TABLE_NAME
    USING ERRCODE = 'read_only_sql_transaction';
END;
$$;

DO $$
DECLARE
  target_tables CONSTANT text[] := ARRAY[
    'organization_units',
    'positions'
  ];
  target_role CONSTANT text := 'user';
  target text;
BEGIN
  FOREACH target IN ARRAY target_tables LOOP
    BEGIN
      EXECUTE format('REVOKE INSERT, UPDATE, DELETE ON %I FROM %I', target, target_role);
      EXECUTE format('COMMENT ON TABLE %I IS %L', target,
        'Plan 402D · 切换窗口内已锁定写权限，禁止直接更新（参见 docs/development-plans/402D-cutover-and-recovery.md）');
      EXECUTE format($fmt$
        DO $do$
        BEGIN
          IF NOT EXISTS (
            SELECT 1 FROM pg_trigger
            WHERE tgname = 'plan402_%1$s_readonly_guard'
          ) THEN
            EXECUTE 'CREATE TRIGGER plan402_%1$s_readonly_guard
                     BEFORE INSERT OR UPDATE OR DELETE ON %1$s
                     FOR EACH ROW
                     EXECUTE FUNCTION plan402_legacy_readonly_guard()';
          END IF;
        END;
        $do$;
      $fmt$, target);
    EXCEPTION
      WHEN undefined_table THEN
        RAISE NOTICE 'Legacy table % does not exist, skipping write lock', target;
    END;
  END LOOP;
END;
$$;

-- 额外的防御：阻止触发器重新启用写入，将其设置为 DISABLE TRIGGER ALL。
-- 显示当前权限状态，供日志记录
\echo '🔒 Legacy tables write privileges after lock:'
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
