-- 013_enhanced_audit_changes_tracking.sql
-- 增强审计触发器，生成详细的字段变动信息
-- 计算changes和modified_fields字段，支持前端显示具体变动内容

-- 创建字段变动计算函数
CREATE OR REPLACE FUNCTION calculate_field_changes(
  old_record JSONB,
  new_record JSONB
) RETURNS TABLE(
  changes JSONB,
  modified_fields JSONB
) AS $$
DECLARE
  change_array JSONB := '[]'::JSONB;
  fields_array JSONB := '[]'::JSONB;
  key TEXT;
  old_value JSONB;
  new_value JSONB;
  change_item JSONB;
  field_name_mapping JSONB;
BEGIN
  -- 字段名映射（数据库字段名 -> 前端显示名）
  field_name_mapping := '{
    "name": "名称",
    "description": "描述", 
    "unit_type": "单位类型",
    "parent_code": "上级单位",
    "status": "状态",
    "effective_date": "生效日期",
    "end_date": "结束日期",
    "change_reason": "变更原因",
    "level": "层级",
    "sort_order": "排序",
    "profile": "配置信息"
  }'::JSONB;
  
  -- 如果old_record为空（INSERT操作），返回空结果
  IF old_record IS NULL OR old_record = 'null'::JSONB THEN
    RETURN QUERY SELECT '[]'::JSONB, '[]'::JSONB;
    RETURN;
  END IF;
  
  -- 遍历所有字段，比较变化
  FOR key IN SELECT jsonb_object_keys(new_record)
  LOOP
    -- 跳过系统字段和时间戳字段
    IF key IN ('record_id', 'created_at', 'updated_at', 'tenant_id', 'code', 'path', 'code_path', 'name_path', 'hierarchy_depth') THEN
      CONTINUE;
    END IF;
    
    old_value := old_record -> key;
    new_value := new_record -> key;
    
    -- 比较值是否发生变化
    IF old_value IS DISTINCT FROM new_value THEN
      -- 构建变更项
      change_item := jsonb_build_object(
        'field', key,
        'fieldLabel', COALESCE(field_name_mapping ->> key, key),
        'oldValue', CASE 
          WHEN old_value = 'null'::JSONB OR old_value IS NULL THEN 'null'::JSONB
          WHEN jsonb_typeof(old_value) = 'string' THEN to_jsonb(old_value #>> '{}')
          ELSE old_value
        END,
        'newValue', CASE 
          WHEN new_value = 'null'::JSONB OR new_value IS NULL THEN 'null'::JSONB
          WHEN jsonb_typeof(new_value) = 'string' THEN to_jsonb(new_value #>> '{}')
          ELSE new_value  
        END
      );
      
      -- 添加到变更数组
      change_array := change_array || jsonb_build_array(change_item);
      
      -- 添加字段名到修改字段数组
      fields_array := fields_array || jsonb_build_array(COALESCE(field_name_mapping ->> key, key));
    END IF;
  END LOOP;
  
  RETURN QUERY SELECT change_array, fields_array;
END;
$$ LANGUAGE plpgsql;

-- 更新审计触发器，生成详细变动信息
DROP FUNCTION IF EXISTS log_audit_changes() CASCADE;

CREATE OR REPLACE FUNCTION log_audit_changes()
RETURNS TRIGGER AS $$
DECLARE
  operation_type_val VARCHAR(20);
  changes_summary_val TEXT;
  operated_by_id_val UUID;
  operated_by_name_val VARCHAR(255);
  operation_reason_val TEXT;
  old_json JSONB;
  new_json JSONB;
  calculated_changes JSONB;
  calculated_fields JSONB;
BEGIN
  -- 确定操作类型
  IF TG_OP = 'INSERT' THEN
    operation_type_val := 'CREATE';
    changes_summary_val := 'Created organization unit: ' || NEW.name;
  ELSIF TG_OP = 'UPDATE' THEN
    operation_type_val := 'UPDATE';
    changes_summary_val := 'Updated organization unit: ' || NEW.name;
  ELSIF TG_OP = 'DELETE' THEN
    operation_type_val := 'DELETE';
    changes_summary_val := 'Deleted organization unit: ' || OLD.name;
  END IF;
  
  -- 设置操作者信息
  operated_by_id_val := '550e8400-e29b-41d4-a716-446655440000'::UUID;
  operated_by_name_val := 'System';
  operation_reason_val := COALESCE(NEW.change_reason, OLD.change_reason, 'System operation');
  
  -- 转换记录为JSONB
  old_json := CASE WHEN TG_OP = 'INSERT' THEN NULL ELSE to_jsonb(OLD) END;
  new_json := CASE WHEN TG_OP = 'DELETE' THEN NULL ELSE to_jsonb(NEW) END;
  
  -- 计算字段变更（仅对UPDATE操作）
  IF TG_OP = 'UPDATE' THEN
    SELECT changes, modified_fields 
    INTO calculated_changes, calculated_fields
    FROM calculate_field_changes(old_json, new_json);
  ELSE
    calculated_changes := '[]'::JSONB;
    calculated_fields := '[]'::JSONB;
  END IF;
  
  -- 插入增强的审计记录
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
    modified_fields,
    record_id
  ) VALUES (
    COALESCE(NEW.tenant_id, OLD.tenant_id),
    operation_type_val,
    'ORGANIZATION',
    operated_by_id_val::VARCHAR,
    'SYSTEM',
    operation_type_val || '_ORGANIZATION',
    gen_random_uuid()::VARCHAR,
    COALESCE(NEW.record_id, OLD.record_id),
    operation_reason_val,
    old_json,
    new_json,
    calculated_changes,
    calculated_fields,
    COALESCE(NEW.record_id, OLD.record_id)
  );
  
  RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- 重新创建触发器
CREATE TRIGGER audit_changes_trigger
    AFTER INSERT OR UPDATE OR DELETE ON organization_units
    FOR EACH ROW
    EXECUTE FUNCTION log_audit_changes();