-- 021_audit_and_temporal_sane_updates.sql
-- 目的：
-- 1) 统一 log_audit_changes()：UPDATE 无实际变更则跳过写审计；写入 business_context 上下文；与统一契约字段对齐
-- 2) 调整 auto_manage_end_dates()：仅在值发生变化时执行 UPDATE，避免空 UPDATE

BEGIN;

-- 计算字段变更（沿用013的思想，裁剪为必要最小集）
CREATE OR REPLACE FUNCTION calculate_field_changes_min(old_record JSONB, new_record JSONB)
RETURNS JSONB AS $$
DECLARE
  key TEXT;
  old_value JSONB;
  new_value JSONB;
  change_array JSONB := '[]'::JSONB;
BEGIN
  IF old_record IS NULL OR old_record = 'null'::JSONB THEN
    RETURN '[]'::JSONB;
  END IF;
  FOR key IN SELECT jsonb_object_keys(new_record)
  LOOP
    IF key IN ('record_id','created_at','updated_at','tenant_id','code','path','code_path','name_path','hierarchy_depth') THEN
      CONTINUE;
    END IF;
    old_value := old_record -> key;
    new_value := new_record -> key;
    IF old_value IS DISTINCT FROM new_value THEN
      change_array := change_array || jsonb_build_array(jsonb_build_object('field', key, 'oldValue', old_value, 'newValue', new_value));
    END IF;
  END LOOP;
  RETURN change_array;
END;
$$ LANGUAGE plpgsql;

-- 统一并增强审计触发器函数
DROP FUNCTION IF EXISTS log_audit_changes() CASCADE;
CREATE OR REPLACE FUNCTION log_audit_changes()
RETURNS TRIGGER AS $$
DECLARE
  op_type VARCHAR(20);
  old_json JSONB;
  new_json JSONB;
  changes JSONB := '[]'::JSONB;
  actor_id_val VARCHAR(100) := '550e8400-e29b-41d4-a716-446655440000';
  actor_type_val VARCHAR(50) := 'SYSTEM';
  action_name_val VARCHAR(100);
  req_id_val VARCHAR(100) := COALESCE(current_setting('app.request_id', true), gen_random_uuid()::text);
  ctx JSONB := jsonb_build_object('source', COALESCE(current_setting('app.context', true), 'unknown'));
  rec_id UUID := COALESCE(NEW.record_id, OLD.record_id);
BEGIN
  IF TG_OP = 'INSERT' THEN
    op_type := 'CREATE';
  ELSIF TG_OP = 'UPDATE' THEN
    op_type := 'UPDATE';
  ELSIF TG_OP = 'DELETE' THEN
    op_type := 'DELETE';
  END IF;

  old_json := CASE WHEN TG_OP = 'INSERT' THEN NULL ELSE to_jsonb(OLD) END;
  new_json := CASE WHEN TG_OP = 'DELETE' THEN NULL ELSE to_jsonb(NEW) END;

  IF TG_OP = 'UPDATE' THEN
    changes := calculate_field_changes_min(old_json, new_json);
    -- 跳过空UPDATE：无字段变化或整体行等价
    IF changes = '[]'::JSONB OR jsonb_array_length(changes) = 0 OR old_json = new_json THEN
      RETURN COALESCE(NEW, OLD);
    END IF;
  END IF;

  action_name_val := op_type || '_ORGANIZATION';

  INSERT INTO audit_logs (
    tenant_id,
    event_type,
    resource_type,
    actor_id,
    actor_type,
    action_name,
    request_id,
    resource_id,
    operation_reason,
    before_data,
    after_data,
    changes,
    business_context,
    record_id
  ) VALUES (
    COALESCE(NEW.tenant_id, OLD.tenant_id),
    op_type,
    'ORGANIZATION',
    actor_id_val,
    actor_type_val,
    action_name_val,
    req_id_val,
    rec_id,
    COALESCE(NEW.change_reason, OLD.change_reason, 'System operation'),
    old_json,
    new_json,
    changes,
    ctx,
    rec_id
  );

  RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- 重建触发器
DROP TRIGGER IF EXISTS audit_changes_trigger ON organization_units;
CREATE TRIGGER audit_changes_trigger
  AFTER INSERT OR UPDATE OR DELETE ON organization_units
  FOR EACH ROW
  EXECUTE FUNCTION log_audit_changes();

-- 调整 auto_manage_end_dates()：仅值变更才更新
CREATE OR REPLACE FUNCTION auto_manage_end_dates()
RETURNS TRIGGER AS $$
DECLARE
  new_prev_end DATE := (NEW.effective_date - INTERVAL '1 day')::date;
  new_curr_end DATE;
BEGIN
  IF TG_OP = 'INSERT' THEN
    -- 邻接：仅当 end_date 将发生变化时更新
    UPDATE organization_units 
       SET end_date = new_prev_end,
           updated_at = NOW()
     WHERE code = NEW.code 
       AND tenant_id = NEW.tenant_id
       AND data_status = 'NORMAL'
       AND effective_date < NEW.effective_date
       AND end_date IS NULL
       AND record_id != NEW.record_id
       AND (end_date IS DISTINCT FROM new_prev_end);

    -- 当前记录：计算新值并仅在变化时更新
    SELECT MIN(effective_date - INTERVAL '1 day')::date 
      INTO new_curr_end
      FROM organization_units future 
     WHERE future.code = NEW.code 
       AND future.tenant_id = NEW.tenant_id
       AND future.data_status = 'NORMAL'
       AND future.effective_date > NEW.effective_date
       AND future.record_id != NEW.record_id;

    UPDATE organization_units 
       SET end_date = new_curr_end
     WHERE record_id = NEW.record_id
       AND (end_date IS DISTINCT FROM new_curr_end);

    RETURN NEW;
  END IF;

  IF TG_OP = 'UPDATE' THEN
    IF OLD.effective_date != NEW.effective_date AND NEW.data_status = 'NORMAL' THEN
      UPDATE organization_units 
         SET end_date = (
               SELECT MIN(effective_date - INTERVAL '1 day')::date 
                 FROM organization_units next_records 
                WHERE next_records.code = NEW.code 
                  AND next_records.tenant_id = NEW.tenant_id
                  AND next_records.data_status = 'NORMAL'
                  AND next_records.effective_date > organization_units.effective_date
                  AND next_records.record_id != NEW.record_id
             ),
             updated_at = NOW()
       WHERE code = NEW.code 
         AND tenant_id = NEW.tenant_id
         AND data_status = 'NORMAL'
         AND effective_date < NEW.effective_date
         AND record_id != NEW.record_id
         AND (end_date IS DISTINCT FROM (
               SELECT MIN(effective_date - INTERVAL '1 day')::date 
                 FROM organization_units next_records 
                WHERE next_records.code = NEW.code 
                  AND next_records.tenant_id = NEW.tenant_id
                  AND next_records.data_status = 'NORMAL'
                  AND next_records.effective_date > organization_units.effective_date
                  AND next_records.record_id != NEW.record_id));

      UPDATE organization_units 
         SET end_date = (
               SELECT MIN(effective_date - INTERVAL '1 day')::date 
                 FROM organization_units future 
                WHERE future.code = NEW.code 
                  AND future.tenant_id = NEW.tenant_id
                  AND future.data_status = 'NORMAL'
                  AND future.effective_date > NEW.effective_date
                  AND future.record_id != NEW.record_id
             )
       WHERE record_id = NEW.record_id
         AND (end_date IS DISTINCT FROM (
               SELECT MIN(effective_date - INTERVAL '1 day')::date 
                 FROM organization_units future 
                WHERE future.code = NEW.code 
                  AND future.tenant_id = NEW.tenant_id
                  AND future.data_status = 'NORMAL'
                  AND future.effective_date > NEW.effective_date
                  AND future.record_id != NEW.record_id));
    END IF;
    RETURN NEW;
  END IF;

  RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS auto_end_date_trigger ON organization_units;
CREATE TRIGGER auto_end_date_trigger
  AFTER INSERT OR UPDATE ON organization_units
  FOR EACH ROW
  EXECUTE FUNCTION auto_manage_end_dates();

COMMIT;

