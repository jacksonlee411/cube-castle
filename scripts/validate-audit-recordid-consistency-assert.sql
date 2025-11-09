-- scripts/validate-audit-recordid-consistency-assert.sql
-- 目的：在校验失败时直接抛错（用于 CI/发布门禁）。
-- 行为：
--  - 空UPDATE>0  → 失败
--  - record_id 与载荷不一致>0 → 失败
--  - OU触发器数>0 → 失败（可通过 psql -v app.assert_triggers_zero=0 暂时跳过）

-- psql 可通过 -v 注入自定义设置：
--   psql ... -v app.assert_triggers_zero=0 -f scripts/validate-audit-recordid-consistency-assert.sql

DO $$
DECLARE
  v_empty_updates bigint := 0;
  v_mismatched    bigint := 0;
  v_triggers      bigint := 0;
  v_assert_trg    text;
BEGIN
  SELECT COUNT(*)
    INTO v_empty_updates
    FROM audit_logs
   WHERE event_type = 'UPDATE'
     AND request_data = response_data
     AND jsonb_array_length(coalesce(changes, '[]'::jsonb)) = 0;

  IF v_empty_updates > 0 THEN
    RAISE EXCEPTION 'AUDIT_EMPTY_UPDATES_GT_ZERO(%)', v_empty_updates;
  END IF;

  SELECT COUNT(*)
    INTO v_mismatched
    FROM audit_logs
   WHERE coalesce((response_data->>'record_id'), (request_data->>'record_id')) IS NOT NULL
     AND record_id IS DISTINCT FROM coalesce((response_data->>'record_id')::uuid, (request_data->>'record_id')::uuid);

  IF v_mismatched > 0 THEN
    RAISE EXCEPTION 'AUDIT_RECORD_ID_MISMATCH_GT_ZERO(%)', v_mismatched;
  END IF;

  -- 仅断言“目标触发器”不存在（组织审计/时态/层级校验相关），避免误伤其他技术性触发器
  SELECT COUNT(*)
    INTO v_triggers
    FROM pg_trigger t
    JOIN pg_class c ON c.oid = t.tgrelid
   WHERE c.relname = 'organization_units'
     AND NOT t.tgisinternal
     AND t.tgname IN (
       'audit_changes_trigger',
       'enforce_temporal_flags_trigger',
       'trg_prevent_update_deleted',
       'update_hierarchy_paths_trigger',
       'validate_parent_available_trigger',
       'validate_parent_available_update_trigger',
       'auto_end_date_trigger',
       'auto_lifecycle_status_trigger',
       'enforce_soft_delete_temporal_flags_trigger'
     );

  -- 允许通过 psql -v app.assert_triggers_zero=0 暂时跳过触发器断言（例如执行022之前）
  BEGIN
    v_assert_trg := current_setting('app.assert_triggers_zero');
  EXCEPTION WHEN others THEN
    v_assert_trg := '1'; -- 默认开启断言
  END;

  IF coalesce(v_assert_trg, '1') = '1' AND v_triggers > 0 THEN
    RAISE EXCEPTION 'OU_TRIGGERS_PRESENT_GT_ZERO(%)', v_triggers;
  END IF;
END $$ LANGUAGE plpgsql;
