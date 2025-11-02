




























































































































































































































































































































































  
  
  
    
    
      
      
                        'dataType', data_type
                        'field', key,
                        'newValue', new_value,
                        'oldValue', old_value,
                        ELSE new_value
                        WHEN new_value IS NULL OR new_value = 'null'::JSONB THEN old_value
                    )
                    CASE
                    END
                    jsonb_build_object(
                  FROM jsonb_array_elements_text(modified_fields) AS t(elem)
                );
                );
                SELECT jsonb_agg(DISTINCT elem)
                USING ERRCODE = 'foreign_key_violation';
                change_items := change_items || jsonb_build_array(
                data_type := infer_audit_change_datatype(
                modified_fields := modified_fields || jsonb_build_array(key);
               AND is_current = true
               AND status <> 'DELETED'
               AND tenant_id = NEW.tenant_id
              FROM organization_units
             LIMIT 1
             WHERE code = NEW.parent_code
             WHERE code = p_code AND tenant_id = p_tenant_id AND is_current = true
            'GAP'::TEXT AS issue_type,
            'Gap between versions'::TEXT AS message
            'OVERLAP'::TEXT AS issue_type,
            'Version overlaps with next version'::TEXT AS message
            'actor_name', actor_name,
            'change_reason', change_reason,
            'operation_reason', change_reason,
            'trigger', 'log_audit_changes'
            );
            ELSE NULL::integer
            ELSE NULL::integer
            END IF;
            IF old_value IS DISTINCT FROM new_value THEN
            NEW.code_path := '/' || NEW.code;
            NEW.level := 1;
            NEW.name_path := '/' || NEW.name;
            NEW.parent_code := NULL;
            RAISE EXCEPTION 'PARENT_NOT_AVAILABLE: parent % is not current or has been deleted', NEW.parent_code
            RAISE EXCEPTION '不能设置父组织，会导致循环引用！组织 % 尝试设置父组织 %', NEW.code, NEW.parent_code;
            RAISE EXCEPTION '父组织不可用（不存在/已删除/非当前）！父组织编码: %', NEW.parent_code;
            RAISE EXCEPTION '组织层级超过最大限制17级！当前尝试创建第%级组织。', calculated_hierarchy_depth;
            RETURN NEW;
            ROW_NUMBER() OVER (ORDER BY effective_date) AS rn
            SELECT 1 FROM organization_units
            SELECT parent_code
            USING ERRCODE = '55000';
            WHEN (is_current = false) THEN 1
            WHEN (is_current = true) THEN 1
            after_snapshot := after_snapshot - key;
            before_snapshot := before_snapshot - key;
            curr.effective_date,
            curr.effective_date,
            curr.end_date,
            curr.end_date,
            effective_date,
            end_date,
            modified_fields := (
            new_value := COALESCE(after_snapshot -> key, 'null'::JSONB);
            old_value := COALESCE(before_snapshot -> key, 'null'::JSONB);
            parent.code_path || '/' || NEW.code,
            parent.level + 1
            parent.name_path || '/' || NEW.name,
           AND p.deleted_at IS NULL
           AND p.is_current = TRUE
           AND p.status <> 'DELETED'
           AND p.tenant_id = NEW.tenant_id
           AND parent.deleted_at IS NULL
           AND parent.is_current = TRUE
           AND parent.status <> 'DELETED'
           AND parent.tenant_id = NEW.tenant_id
          )
          AND code = p_code
          AND curr.end_date + INTERVAL '1 day' < nxt.effective_date
          AND curr.end_date >= nxt.effective_date
          AND status <> 'DELETED'
          ELSE new_value  
          ELSE old_value
          FROM organization_units p
          FROM organization_units parent
          INTO NEW.code_path, NEW.name_path, NEW.level
          WHEN jsonb_typeof(new_value) = 'string' THEN to_jsonb(new_value #>> '{}')
          WHEN jsonb_typeof(old_value) = 'string' THEN to_jsonb(old_value #>> '{}')
          WHEN new_value = 'null'::JSONB OR new_value IS NULL THEN 'null'::JSONB
          WHEN old_value = 'null'::JSONB OR old_value IS NULL THEN 'null'::JSONB
         LIMIT 1;
         LIMIT 1;
         WHERE p.code = NEW.parent_code
         WHERE parent.code = NEW.parent_code
        'ORGANIZATION',
        'created_at','updated_at','tenant_id','record_id','path','code_path','name_path',
        'field', key,
        'fieldLabel', COALESCE(field_name_mapping ->> key, key),
        'hierarchy_depth','metadata','changed_by','approved_by','request_id','is_current',
        'is_temporal','is_future'
        'newValue', CASE 
        'oldValue', CASE 
        ) THEN
        ))
        CASE
        CASE
        CASE WHEN actor_id_text = 'system' THEN 'System' ELSE actor_id_text END
        COALESCE(after_snapshot, '{}'::JSONB),
        COALESCE(before_snapshot, '{}'::JSONB),
        COALESCE(changed_by_val::TEXT, 'system')
        COALESCE(modified_fields, '[]'::JSONB),
        END
        END IF;
        END IF;
        END IF;
        END IF;
        END IF;
        END IF;
        END IF;
        END LOOP;
        END LOOP;
        END LOOP;
        END) AS current_count,
        END) AS historical_count
        END,
        FOR key IN SELECT jsonb_object_keys(after_snapshot)
        FOREACH key IN ARRAY excluded_keys LOOP
        FOREACH key IN ARRAY excluded_keys LOOP
        FROM ordered_versions curr
        FROM ordered_versions curr
        FROM organization_units
        IF NOT EXISTS (
        IF NOT FOUND THEN
        IF NOT FOUND THEN
        IF calculated_hierarchy_depth > 17 THEN
        IF check_circular_reference(NEW.code, NEW.parent_code, NEW.tenant_id) THEN
        IF jsonb_array_length(change_items) = 0 THEN
        IF jsonb_array_length(modified_fields) > 0 THEN
        JOIN ordered_versions nxt ON nxt.rn = curr.rn + 1
        JOIN ordered_versions nxt ON nxt.rn = curr.rn + 1
        LOOP
        NEW.code_path := '/' || NEW.code;
        NEW.is_current := FALSE;
        NEW.is_current := FALSE;
        NEW.is_current := FALSE;
        NEW.is_current := FALSE;
        NEW.is_current := FALSE;
        NEW.is_current := FALSE;
        NEW.is_current := TRUE;
        NEW.level := 1;
        NEW.name_path := '/' || NEW.name;
        ORDER BY effective_date
        PERFORM 1
        RAISE EXCEPTION 'READ_ONLY_DELETED: cannot modify deleted record %', OLD.record_id
        RETURN 'array';
        RETURN 'boolean';
        RETURN 'number';
        RETURN 'object';
        RETURN 'string';
        RETURN 'unknown';
        RETURN 'unknown';
        RETURN NEW;
        RETURN NEW;
        RETURN NEW;
        RETURN NEW;
        SELECT
        SELECT
        SELECT
        SELECT 
        TRUE,
        WHERE curr.end_date IS NOT NULL
        WHERE curr.end_date IS NOT NULL
        WHERE tenant_id = p_tenant_id
        action_name,
        actor_id,
        actor_id_text,
        actor_type,
        actor_type,
        after_snapshot := to_jsonb(NEW);
        before_snapshot := to_jsonb(OLD);
        business_context
        calculated_code_path := '/' || p_code;
        calculated_code_path := COALESCE(parent_info.code_path, '/' || parent_info.code) || '/' || p_code;
        calculated_hierarchy_depth := 1;
        calculated_hierarchy_depth := parent_info.hierarchy_depth + 1;
        calculated_level := 1;
        calculated_level := parent_info.level + 1;
        calculated_name_path := '/' || COALESCE(current_name, p_code);
        calculated_name_path := COALESCE(parent_info.name_path, '/' || current_name) || '/' || COALESCE(current_name, p_code);
        change_items,
        change_reason
        change_reason := COALESCE(NEW.change_reason, OLD.change_reason);
        change_reason := COALESCE(NEW.change_reason, OLD.change_reason);
        change_reason := NEW.change_reason;
        change_reason := OLD.change_reason;
        change_reason,
        changed_by_val := COALESCE(NEW.changed_by, OLD.changed_by);
        changed_by_val := COALESCE(NEW.changed_by, OLD.changed_by);
        changed_by_val := NEW.changed_by;
        changed_by_val := OLD.changed_by;
        changes,
        current_setting('app.request_id', true),
        current_setting('cube.actor_id', true),
        current_setting('cube.actor_name', true),
        current_setting('cube.change_reason', true),
        current_setting('cube.request_id', true),
        event_type,
        gen_random_uuid()::TEXT
        jsonb_strip_nulls(jsonb_build_object(
        modified_fields,
        op_type := 'CREATE';
        op_type := 'DELETE';
        op_type := 'UPDATE';
        op_type := TG_OP;
        op_type || '_ORGANIZATION',
        op_type,
        operation_reason,
        ou.change_reason
        ou.code,
        ou.code,
        ou.code_path,
        ou.effective_date,
        ou.end_date,
        ou.hierarchy_depth
        ou.is_current,
        ou.level,
        ou.name,
        ou.name_path,
        ou.parent_code,
        ou.status,
        ou.unit_type,
        record_id,
        request_data,
        request_id,
        request_token,
        resource_id,
        resource_type,
        response_data,
        success,
        target_record := COALESCE(NEW.record_id, OLD.record_id);
        target_record := COALESCE(NEW.record_id, OLD.record_id);
        target_record := NEW.record_id;
        target_record := OLD.record_id;
        target_record,
        target_record::TEXT,
        target_tenant := COALESCE(NEW.tenant_id, OLD.tenant_id);
        target_tenant := COALESCE(NEW.tenant_id, OLD.tenant_id);
        target_tenant := NEW.tenant_id;
        target_tenant := OLD.tenant_id;
        target_tenant,
        tenant_id,
       AND ou.is_current = true
       AND ou.status <> 'DELETED'
       AND ou.tenant_id = p_tenant_id
      );
      -- 构建变更项
      -- 添加到变更数组
      -- 添加字段名到修改字段数组
      AND (ou.end_date IS NULL OR ou.end_date > p_as_of_date)
      AND COALESCE(ou.effective_date, CURRENT_DATE) <= p_as_of_date
      AND ou.code = p_code
      CONTINUE;
      FROM organization_units
      FROM organization_units ou
      INTO parent_info
      change_array := change_array || jsonb_build_array(change_item);
      change_item := jsonb_build_object(
      fields_array := fields_array || jsonb_build_array(COALESCE(field_name_mapping ->> key, key));
     LIMIT 1;
     LIMIT 1;
     WHERE code = p_code AND tenant_id = p_tenant_id AND is_current = true
     WHERE ou.code = (
    "change_reason": "变更原因",
    "description": "描述", 
    "effective_date": "生效日期",
    "end_date": "结束日期",
    "level": "层级",
    "name": "名称",
    "parent_code": "上级单位",
    "profile": "配置信息"
    "sort_order": "排序",
    "status": "状态",
    "timestamp" timestamp with time zone DEFAULT now() NOT NULL,
    "unit_type": "单位类型",
    )
    ) VALUES (
    ), gaps AS (
    ), version_overlaps AS (
    );
    );
    );
    );
    );
    -- 其它情况保留调用方指定值（交由时间轴重算流程统一处理）
    -- 删除状态或已标记删除的记录始终不可作为当前版本
    -- 未来生效或已过期的记录不应标记为当前版本
    -- 比较值是否发生变化
    -- 跳过系统字段和时间戳字段
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);
    ADD CONSTRAINT fk_job_families_group FOREIGN KEY (parent_record_id, tenant_id) REFERENCES public.job_family_groups(record_id, tenant_id);
    ADD CONSTRAINT fk_job_levels_role FOREIGN KEY (parent_record_id, tenant_id) REFERENCES public.job_roles(record_id, tenant_id);
    ADD CONSTRAINT fk_job_roles_family FOREIGN KEY (parent_record_id, tenant_id) REFERENCES public.job_families(record_id, tenant_id);
    ADD CONSTRAINT fk_position_assignments_position FOREIGN KEY (tenant_id, position_code, position_record_id) REFERENCES public.positions(tenant_id, code, record_id);
    ADD CONSTRAINT fk_positions_family FOREIGN KEY (job_family_record_id, tenant_id) REFERENCES public.job_families(record_id, tenant_id);
    ADD CONSTRAINT fk_positions_family_group FOREIGN KEY (job_family_group_record_id, tenant_id) REFERENCES public.job_family_groups(record_id, tenant_id);
    ADD CONSTRAINT fk_positions_level FOREIGN KEY (job_level_record_id, tenant_id) REFERENCES public.job_levels(record_id, tenant_id);
    ADD CONSTRAINT fk_positions_role FOREIGN KEY (job_role_record_id, tenant_id) REFERENCES public.job_roles(record_id, tenant_id);
    ADD CONSTRAINT job_families_pkey PRIMARY KEY (record_id);
    ADD CONSTRAINT job_families_record_id_tenant_id_key UNIQUE (record_id, tenant_id);
    ADD CONSTRAINT job_families_tenant_id_family_code_effective_date_key UNIQUE (tenant_id, family_code, effective_date);
    ADD CONSTRAINT job_family_groups_pkey PRIMARY KEY (record_id);
    ADD CONSTRAINT job_family_groups_record_id_tenant_id_key UNIQUE (record_id, tenant_id);
    ADD CONSTRAINT job_family_groups_tenant_id_family_group_code_effective_dat_key UNIQUE (tenant_id, family_group_code, effective_date);
    ADD CONSTRAINT job_levels_pkey PRIMARY KEY (record_id);
    ADD CONSTRAINT job_levels_record_id_tenant_id_key UNIQUE (record_id, tenant_id);
    ADD CONSTRAINT job_levels_tenant_id_level_code_effective_date_key UNIQUE (tenant_id, level_code, effective_date);
    ADD CONSTRAINT job_roles_pkey PRIMARY KEY (record_id);
    ADD CONSTRAINT job_roles_record_id_tenant_id_key UNIQUE (record_id, tenant_id);
    ADD CONSTRAINT job_roles_tenant_id_role_code_effective_date_key UNIQUE (tenant_id, role_code, effective_date);
    ADD CONSTRAINT pk_org_record_id PRIMARY KEY (record_id);
    ADD CONSTRAINT position_assignments_pkey PRIMARY KEY (assignment_id);
    ADD CONSTRAINT positions_pkey PRIMARY KEY (record_id);
    ADD CONSTRAINT positions_tenant_id_code_effective_date_key UNIQUE (tenant_id, code, effective_date);
    ADD CONSTRAINT positions_tenant_id_code_record_id_key UNIQUE (tenant_id, code, record_id);
    AS $$
    AS $$
    AS $$
    AS $$
    AS $$
    AS $$
    AS $$
    AS $$
    AS $$
    AS $$
    AS $$
    AS $$
    CONSTRAINT audit_logs_event_type_check_v2 CHECK (((event_type)::text = ANY ((ARRAY['CREATE'::character varying, 'UPDATE'::character varying, 'DELETE'::character varying, 'SUSPEND'::character varying, 'REACTIVATE'::character varying, 'QUERY'::character varying, 'VALIDATION'::character varying, 'AUTHENTICATION'::character varying, 'ERROR'::character varying])::text[])))
    CONSTRAINT chk_deleted_not_current CHECK (
    CONSTRAINT chk_org_units_not_deleted_current CHECK (
    CONSTRAINT chk_position_assignments_auto_revert CHECK (((auto_revert = false) OR (((assignment_type)::text = 'ACTING'::text) AND (acting_until IS NOT NULL)))),
    CONSTRAINT chk_position_assignments_dates CHECK ((((end_date IS NULL) OR (end_date > effective_date)) AND ((acting_until IS NULL) OR (acting_until > effective_date)))),
    CONSTRAINT chk_position_assignments_fte CHECK (((fte >= (0)::numeric) AND (fte <= (1)::numeric))),
    CONSTRAINT chk_position_assignments_status CHECK (((assignment_status)::text = ANY ((ARRAY['PENDING'::character varying, 'ACTIVE'::character varying, 'ENDED'::character varying])::text[]))),
    CONSTRAINT chk_position_assignments_type CHECK (((assignment_type)::text = ANY ((ARRAY['PRIMARY'::character varying, 'SECONDARY'::character varying, 'ACTING'::character varying])::text[])))
    CONSTRAINT valid_unit_type CHECK (((unit_type)::text = ANY ((ARRAY['DEPARTMENT'::character varying, 'ORGANIZATION_UNIT'::character varying, 'PROJECT_TEAM'::character varying])::text[])))
    ELSE
    ELSE
    ELSE
    ELSE
    ELSE
    ELSE true
    ELSE true
    ELSIF NEW.end_date IS NOT NULL AND NEW.end_date <= CURRENT_DATE THEN
    ELSIF TG_OP = 'DELETE' THEN
    ELSIF TG_OP = 'UPDATE' THEN
    ELSIF value_type = 'array' THEN
    ELSIF value_type = 'boolean' THEN
    ELSIF value_type = 'number' THEN
    ELSIF value_type = 'object' THEN
    END IF;
    END IF;
    END IF;
    END IF;
    END IF;
    END IF;
    END IF;
    END IF;
    END IF;
    END IF;
    END IF;
    END IF;
    END IF;
    END IF;
    END IF;
    END IF;
    END IF;
    END IF;
    END IF;
    END IF;
    END IF;
    FROM organization_units ou
    IF (OLD.status = 'DELETED') THEN
    IF NEW.effective_date > CURRENT_DATE THEN
    IF NEW.effective_date > utc_date THEN
    IF NEW.end_date IS NOT NULL AND NEW.end_date <= utc_date THEN
    IF NEW.parent_code IS NOT NULL THEN
    IF NEW.parent_code IS NOT NULL THEN
    IF NEW.parent_code IS NOT NULL THEN
    IF NEW.parent_code IS NULL THEN
    IF NEW.status = 'DELETED' OR NEW.deleted_at IS NOT NULL THEN
    IF NEW.status = 'DELETED' THEN
    IF TG_OP <> 'DELETE' THEN
    IF TG_OP <> 'INSERT' THEN
    IF TG_OP = 'INSERT' THEN
    IF TG_OP = 'UPDATE' THEN
    IF after_snapshot IS NOT NULL THEN
    IF before_snapshot IS NOT NULL THEN
    IF key IN ('record_id', 'created_at', 'updated_at', 'tenant_id', 'code', 'path', 'code_path', 'name_path', 'hierarchy_depth') THEN
    IF old_value IS DISTINCT FROM new_value THEN
    IF parent_info.code IS NULL THEN
    IF value IS NULL OR value = 'null'::JSONB THEN
    IF value_type = 'string' THEN
    INSERT INTO audit_logs (
    LANGUAGE plpgsql
    LANGUAGE plpgsql
    LANGUAGE plpgsql
    LANGUAGE plpgsql
    LANGUAGE plpgsql
    LANGUAGE plpgsql
    LANGUAGE plpgsql
    LANGUAGE plpgsql
    LANGUAGE plpgsql
    LANGUAGE plpgsql
    LANGUAGE plpgsql IMMUTABLE
    LANGUAGE sql STABLE
    LIMIT 1;
    NEW.hierarchy_depth := NEW.level;
    ORDER BY ou.effective_date DESC
    RETURN COALESCE(NEW, OLD);
    RETURN NEW;
    RETURN NEW;
    RETURN NEW;
    RETURN NEW;
    RETURN NEW;
    RETURN NEW;
    RETURN NEXT;
    RETURN QUERY
    RETURN QUERY SELECT '[]'::JSONB, '[]'::JSONB;
    RETURN;
    SELECT
    SELECT 
    SELECT * FROM gaps;
    SELECT * FROM version_overlaps
    SELECT name INTO current_name
    UNION ALL
    WHEN (((status)::text = 'DELETED'::text) OR (deleted_at IS NOT NULL)) THEN (is_current = false)
    WHEN ((status)::text = 'DELETED'::text) THEN (is_current = false)
    WHERE ou.tenant_id = p_tenant_id
    WITH ordered_versions AS (
    ];
    acting_until date,
    action_name character varying(100),
    actor_id character varying(100),
    actor_id_text := COALESCE(
    actor_id_text TEXT;
    actor_name := COALESCE(
    actor_name TEXT;
    actor_type := CASE WHEN actor_id_text = 'system' THEN 'SYSTEM' ELSE 'USER' END;
    actor_type TEXT;
    actor_type character varying(50),
    after_snapshot JSONB := NULL;
    approved_by
    approved_by uuid,
    approved_by uuid,
    approved_by uuid,
    assignment_id uuid DEFAULT gen_random_uuid() NOT NULL,
    assignment_status character varying(20) DEFAULT 'ACTIVE'::character varying NOT NULL,
    assignment_type character varying(20) NOT NULL,
    auto_revert boolean DEFAULT false NOT NULL,
    before_snapshot JSONB := NULL;
    business_context jsonb DEFAULT '{}'::jsonb NOT NULL,
    change_items JSONB := '[]'::JSONB;
    change_reason := COALESCE(
    change_reason := NULLIF(change_reason, '');
    change_reason TEXT;
    change_reason text,
    change_reason text,
    change_reason text,
    change_reason,
    change_reason,
    changed_by uuid,
    changed_by uuid,
    changed_by uuid,
    changed_by,
    changed_by_val UUID;
    changes jsonb DEFAULT '[]'::jsonb NOT NULL,
    code character varying(12) NOT NULL,
    code character varying(12),
    code character varying(12),
    code character varying(8) NOT NULL,
    code,
    code,
    code_path text DEFAULT ''::text NOT NULL,
    code_path text,
    code_path text,
    code_path,
    code_path,
    competency_model jsonb DEFAULT '{}'::jsonb,
    cost_center_code character varying(50),
    count(
    count(
    count(*) AS count,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_at timestamp with time zone,
    created_at timestamp with time zone,
    created_at,
    created_at,
    current_name VARCHAR(255);
    data_type TEXT;
    deleted_at timestamp with time zone,
    deleted_at timestamp with time zone,
    deleted_at timestamp with time zone,
    deleted_at timestamp with time zone,
    deleted_at,
    deleted_by uuid,
    deleted_by uuid,
    deleted_by uuid,
    deleted_by,
    deletion_reason text,
    deletion_reason text,
    deletion_reason text,
    deletion_reason,
    description text,
    description text,
    description text,
    description text,
    description text,
    description text,
    description text,
    description,
    description,
    effective_date date DEFAULT CURRENT_DATE NOT NULL,
    effective_date date NOT NULL,
    effective_date date NOT NULL,
    effective_date date NOT NULL,
    effective_date date NOT NULL,
    effective_date date NOT NULL,
    effective_date date NOT NULL,
    effective_date date,
    effective_date date,
    effective_date,
    effective_date,
    effective_from timestamp with time zone,
    effective_from timestamp with time zone,
    effective_from timestamp with time zone,
    effective_from,
    effective_to timestamp with time zone,
    effective_to timestamp with time zone,
    effective_to timestamp with time zone,
    effective_to,
    employee_id uuid NOT NULL,
    employee_name character varying(255) NOT NULL,
    employee_number character varying(64),
    employment_type character varying(50) NOT NULL,
    end_date date,
    end_date date,
    end_date date,
    end_date date,
    end_date date,
    end_date date,
    end_date date,
    end_date date,
    end_date date,
    end_date,
    end_date,
    error_code character varying(100),
    error_message text,
    event_type character varying(50) NOT NULL,
    excluded_keys TEXT[] := ARRAY[
    family_code character varying(20) NOT NULL,
    family_code character varying(20) NOT NULL,
    family_group_code character varying(20) NOT NULL,
    family_group_code character varying(20) NOT NULL,
    fte numeric(5,2) DEFAULT 1.0 NOT NULL,
    grade_level character varying(20),
    headcount_capacity numeric(5,2) DEFAULT 1.0 NOT NULL,
    headcount_in_use numeric(5,2) DEFAULT 0.0 NOT NULL,
    hierarchy_depth integer DEFAULT 0 NOT NULL,
    hierarchy_depth integer,
    hierarchy_depth integer,
    hierarchy_depth,
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    is_current boolean DEFAULT false NOT NULL,
    is_current boolean DEFAULT false NOT NULL,
    is_current boolean DEFAULT false NOT NULL,
    is_current boolean DEFAULT false NOT NULL,
    is_current boolean DEFAULT false NOT NULL,
    is_current boolean DEFAULT false NOT NULL,
    is_current boolean DEFAULT false NOT NULL,
    is_current boolean,
    is_current boolean,
    is_current,
    is_current,
    is_temporal boolean
    is_temporal boolean
    job_family_code character varying(20) NOT NULL,
    job_family_group_code character varying(20) NOT NULL,
    job_family_group_name character varying(255) NOT NULL,
    job_family_group_record_id uuid NOT NULL,
    job_family_name character varying(255) NOT NULL,
    job_family_record_id uuid NOT NULL,
    job_level_code character varying(20) NOT NULL,
    job_level_name character varying(255) NOT NULL,
    job_level_record_id uuid NOT NULL,
    job_profile_code character varying(64),
    job_profile_name character varying(255),
    job_role_code character varying(20) NOT NULL,
    job_role_name character varying(255) NOT NULL,
    job_role_record_id uuid NOT NULL,
    key TEXT;
    level integer DEFAULT 1 NOT NULL,
    level integer,
    level integer,
    level,
    level,
    level_code character varying(20) NOT NULL,
    level_rank character varying(20) NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    metadata jsonb,
    metadata jsonb,
    metadata,
    modified_fields JSONB := '[]'::JSONB;
    modified_fields jsonb DEFAULT '[]'::jsonb NOT NULL,
    name character varying(255) NOT NULL,
    name character varying(255) NOT NULL,
    name character varying(255) NOT NULL,
    name character varying(255) NOT NULL,
    name character varying(255) NOT NULL,
    name character varying(255),
    name character varying(255),
    name,
    name,
    name_path text DEFAULT ''::text NOT NULL,
    name_path text,
    name_path text,
    name_path,
    name_path,
    new_value := new_record -> key;
    new_value JSONB;
    notes text,
    old_value := old_record -> key;
    old_value JSONB;
    op_type TEXT;
    operated_by_id uuid NOT NULL,
    operated_by_id uuid,
    operated_by_id uuid,
    operated_by_id uuid,
    operated_by_id,
    operated_by_name character varying(255) NOT NULL,
    operated_by_name text,
    operated_by_name text,
    operated_by_name text,
    operated_by_name,
    operation_reason text
    operation_reason text,
    operation_type character varying(20) DEFAULT 'CREATE'::character varying NOT NULL,
    operation_type character varying(20) DEFAULT 'CREATE'::character varying,
    operation_type character varying(20),
    operation_type character varying(20),
    organization_code character varying(7) NOT NULL,
    organization_name character varying(255),
    parent_code character varying(12),
    parent_code character varying(12),
    parent_code character varying(12),
    parent_code,
    parent_code,
    parent_info RECORD;
    parent_record_id uuid NOT NULL,
    parent_record_id uuid NOT NULL,
    parent_record_id uuid NOT NULL,
    position_code character varying(8) NOT NULL,
    position_record_id uuid NOT NULL,
    position_type character varying(50) NOT NULL,
    profile jsonb DEFAULT '{}'::jsonb NOT NULL,
    profile jsonb DEFAULT '{}'::jsonb NOT NULL,
    profile jsonb,
    profile jsonb,
    profile,
    profile,
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    record_id uuid,
    record_id uuid,
    record_id uuid,
    reminder_sent_at timestamp with time zone,
    reports_to_position_code character varying(8),
    request_data jsonb DEFAULT '{}'::jsonb NOT NULL,
    request_id character varying(100),
    request_token := COALESCE(
    request_token TEXT;
    resource_id character varying(100),
    resource_type character varying(50) NOT NULL,
    response_data jsonb DEFAULT '{}'::jsonb NOT NULL,
    role_code character varying(20) NOT NULL,
    role_code character varying(20) NOT NULL,
    salary_band jsonb,
    sort_order integer DEFAULT 0 NOT NULL,
    sort_order integer,
    sort_order integer,
    sort_order,
    sort_order,
    status character varying(20) DEFAULT 'ACTIVE'::character varying NOT NULL,
    status character varying(20) DEFAULT 'ACTIVE'::character varying NOT NULL,
    status character varying(20) DEFAULT 'ACTIVE'::character varying NOT NULL,
    status character varying(20) DEFAULT 'ACTIVE'::character varying NOT NULL,
    status character varying(20) DEFAULT 'ACTIVE'::character varying NOT NULL,
    status character varying(20) DEFAULT 'PLANNED'::character varying NOT NULL,
    status character varying(20),
    status character varying(20),
    status,
    status,
    success boolean DEFAULT true NOT NULL,
    suspended_at timestamp with time zone,
    suspended_at timestamp with time zone,
    suspended_at timestamp with time zone,
    suspended_at,
    suspended_by uuid,
    suspended_by uuid,
    suspended_by uuid,
    suspended_by,
    suspension_reason text,
    suspension_reason text,
    suspension_reason text,
    suspension_reason,
    target_record UUID;
    target_tenant UUID;
    tenant_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    tenant_id uuid,
    tenant_id uuid,
    tenant_id,
    title character varying(120) NOT NULL,
    unit_type character varying(64) NOT NULL,
    unit_type character varying(64),
    unit_type character varying(64),
    unit_type,
    unit_type,
    unit_type,
    updated_at
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone,
    updated_at timestamp with time zone,
    updated_at,
    utc_date DATE := (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::date;
    value_type := jsonb_typeof(value);
    value_type TEXT;
   FROM public.organization_units
   FROM public.organization_units
   FROM public.organization_units ou
  -- 如果old_record为空（INSERT操作），返回空结果
  -- 字段名映射（数据库字段名 -> 前端显示名）
  -- 遍历所有字段，比较变化
  END IF;
  END LOOP;
  FOR key IN SELECT jsonb_object_keys(new_record)
  GROUP BY tenant_id, unit_type;
  IF old_record IS NULL OR old_record = 'null'::JSONB THEN
  LOOP
  RETURN QUERY SELECT change_array, fields_array;
  WHERE ((is_current = true) AND ((end_date IS NULL) OR (end_date > CURRENT_DATE)));
  WHERE (deleted_at IS NULL)
  WHERE (is_current = true);
  change_array JSONB := '[]'::JSONB;
  change_item JSONB;
  field_name_mapping := '{
  field_name_mapping JSONB;
  fields_array JSONB := '[]'::JSONB;
  key TEXT;
  new_value JSONB;
  old_value JSONB;
  }'::JSONB;
 SELECT record_id,
 SELECT tenant_id,
 SELECT tenant_id,
$$;
$$;
$$;
$$;
$$;
$$;
$$;
$$;
$$;
$$;
$$;
$$;
);
);
);
);
);
);
);
);
);
);
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
--
-- Dumped by pg_dump version 16.9
-- Dumped from database version 16.9
-- Name: COLUMN audit_logs.record_id; Type: COMMENT; Schema: public; Owner: -
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: -
-- Name: audit_logs audit_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
-- Name: audit_logs; Type: TABLE; Schema: public; Owner: -
-- Name: calculate_field_changes(jsonb, jsonb); Type: FUNCTION; Schema: public; Owner: -
-- Name: calculate_org_hierarchy(character varying, uuid); Type: FUNCTION; Schema: public; Owner: -
-- Name: check_temporal_continuity(uuid, character varying); Type: FUNCTION; Schema: public; Owner: -
-- Name: enforce_soft_delete_temporal_flags(); Type: FUNCTION; Schema: public; Owner: -
-- Name: enforce_temporal_flags(); Type: FUNCTION; Schema: public; Owner: -
-- Name: get_organization_temporal(uuid, character varying, date); Type: FUNCTION; Schema: public; Owner: -
-- Name: idx_audit_logs_record_id_time; Type: INDEX; Schema: public; Owner: -
-- Name: idx_audit_logs_resource; Type: INDEX; Schema: public; Owner: -
-- Name: idx_audit_logs_resource_timestamp; Type: INDEX; Schema: public; Owner: -
-- Name: idx_audit_logs_timestamp; Type: INDEX; Schema: public; Owner: -
-- Name: idx_org_unit_type_optimized; Type: INDEX; Schema: public; Owner: -
-- Name: idx_org_units_code_current_active; Type: INDEX; Schema: public; Owner: -
-- Name: idx_org_units_parent; Type: INDEX; Schema: public; Owner: -
-- Name: idx_org_units_tenant; Type: INDEX; Schema: public; Owner: -
-- Name: idx_organization_current_only; Type: INDEX; Schema: public; Owner: -
-- Name: idx_organization_date_range; Type: INDEX; Schema: public; Owner: -
-- Name: idx_organization_temporal_main; Type: INDEX; Schema: public; Owner: -
-- Name: idx_organization_units_effective_from; Type: INDEX; Schema: public; Owner: -
-- Name: idx_organization_units_effective_to; Type: INDEX; Schema: public; Owner: -
-- Name: idx_position_assignments_auto_revert_due; Type: INDEX; Schema: public; Owner: -
-- Name: idx_position_assignments_employee; Type: INDEX; Schema: public; Owner: -
-- Name: idx_position_assignments_position; Type: INDEX; Schema: public; Owner: -
-- Name: idx_position_assignments_status; Type: INDEX; Schema: public; Owner: -
-- Name: idx_positions_current; Type: INDEX; Schema: public; Owner: -
-- Name: idx_positions_effective_date; Type: INDEX; Schema: public; Owner: -
-- Name: idx_positions_job_family; Type: INDEX; Schema: public; Owner: -
-- Name: idx_positions_job_family_group; Type: INDEX; Schema: public; Owner: -
-- Name: idx_positions_job_role; Type: INDEX; Schema: public; Owner: -
-- Name: idx_positions_org_code; Type: INDEX; Schema: public; Owner: -
-- Name: idx_positions_status; Type: INDEX; Schema: public; Owner: -
-- Name: infer_audit_change_datatype(jsonb); Type: FUNCTION; Schema: public; Owner: -
-- Name: ix_org_adjacent_versions; Type: INDEX; Schema: public; Owner: -
-- Name: ix_org_current_lookup; Type: INDEX; Schema: public; Owner: -
-- Name: ix_org_daily_transition; Type: INDEX; Schema: public; Owner: -
-- Name: ix_org_temporal_boundaries; Type: INDEX; Schema: public; Owner: -
-- Name: ix_org_temporal_query; Type: INDEX; Schema: public; Owner: -
-- Name: job_families fk_job_families_group; Type: FK CONSTRAINT; Schema: public; Owner: -
-- Name: job_families job_families_pkey; Type: CONSTRAINT; Schema: public; Owner: -
-- Name: job_families job_families_record_id_tenant_id_key; Type: CONSTRAINT; Schema: public; Owner: -
-- Name: job_families job_families_tenant_id_family_code_effective_date_key; Type: CONSTRAINT; Schema: public; Owner: -
-- Name: job_families; Type: TABLE; Schema: public; Owner: -
-- Name: job_family_groups job_family_groups_pkey; Type: CONSTRAINT; Schema: public; Owner: -
-- Name: job_family_groups job_family_groups_record_id_tenant_id_key; Type: CONSTRAINT; Schema: public; Owner: -
-- Name: job_family_groups job_family_groups_tenant_id_family_group_code_effective_dat_key; Type: CONSTRAINT; Schema: public; Owner: -
-- Name: job_family_groups; Type: TABLE; Schema: public; Owner: -
-- Name: job_levels fk_job_levels_role; Type: FK CONSTRAINT; Schema: public; Owner: -
-- Name: job_levels job_levels_pkey; Type: CONSTRAINT; Schema: public; Owner: -
-- Name: job_levels job_levels_record_id_tenant_id_key; Type: CONSTRAINT; Schema: public; Owner: -
-- Name: job_levels job_levels_tenant_id_level_code_effective_date_key; Type: CONSTRAINT; Schema: public; Owner: -
-- Name: job_levels; Type: TABLE; Schema: public; Owner: -
-- Name: job_roles fk_job_roles_family; Type: FK CONSTRAINT; Schema: public; Owner: -
-- Name: job_roles job_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
-- Name: job_roles job_roles_record_id_tenant_id_key; Type: CONSTRAINT; Schema: public; Owner: -
-- Name: job_roles job_roles_tenant_id_role_code_effective_date_key; Type: CONSTRAINT; Schema: public; Owner: -
-- Name: job_roles; Type: TABLE; Schema: public; Owner: -
-- Name: log_audit_changes(); Type: FUNCTION; Schema: public; Owner: -
-- Name: organization_current; Type: VIEW; Schema: public; Owner: -
-- Name: organization_stats_view; Type: VIEW; Schema: public; Owner: -
-- Name: organization_temporal_current; Type: VIEW; Schema: public; Owner: -
-- Name: organization_units audit_changes_trigger; Type: TRIGGER; Schema: public; Owner: -
-- Name: organization_units enforce_temporal_flags_trigger; Type: TRIGGER; Schema: public; Owner: -
-- Name: organization_units pk_org_record_id; Type: CONSTRAINT; Schema: public; Owner: -
-- Name: organization_units trg_prevent_update_deleted; Type: TRIGGER; Schema: public; Owner: -
-- Name: organization_units update_hierarchy_paths_trigger; Type: TRIGGER; Schema: public; Owner: -
-- Name: organization_units validate_parent_available_trigger; Type: TRIGGER; Schema: public; Owner: -
-- Name: organization_units validate_parent_available_update_trigger; Type: TRIGGER; Schema: public; Owner: -
-- Name: organization_units; Type: TABLE; Schema: public; Owner: -
-- Name: organization_units_backup_temporal; Type: TABLE; Schema: public; Owner: -
-- Name: organization_units_unittype_backup; Type: TABLE; Schema: public; Owner: -
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
-- Name: position_assignments fk_position_assignments_position; Type: FK CONSTRAINT; Schema: public; Owner: -
-- Name: position_assignments position_assignments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
-- Name: position_assignments; Type: TABLE; Schema: public; Owner: -
-- Name: positions fk_positions_family; Type: FK CONSTRAINT; Schema: public; Owner: -
-- Name: positions fk_positions_family_group; Type: FK CONSTRAINT; Schema: public; Owner: -
-- Name: positions fk_positions_level; Type: FK CONSTRAINT; Schema: public; Owner: -
-- Name: positions fk_positions_role; Type: FK CONSTRAINT; Schema: public; Owner: -
-- Name: positions positions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
-- Name: positions positions_tenant_id_code_effective_date_key; Type: CONSTRAINT; Schema: public; Owner: -
-- Name: positions positions_tenant_id_code_record_id_key; Type: CONSTRAINT; Schema: public; Owner: -
-- Name: positions; Type: TABLE; Schema: public; Owner: -
-- Name: prevent_update_deleted(); Type: FUNCTION; Schema: public; Owner: -
-- Name: uidx_org_record_id; Type: INDEX; Schema: public; Owner: -
-- Name: uk_job_families_current; Type: INDEX; Schema: public; Owner: -
-- Name: uk_job_families_record; Type: INDEX; Schema: public; Owner: -
-- Name: uk_job_family_groups_current; Type: INDEX; Schema: public; Owner: -
-- Name: uk_job_family_groups_record; Type: INDEX; Schema: public; Owner: -
-- Name: uk_job_levels_current; Type: INDEX; Schema: public; Owner: -
-- Name: uk_job_levels_record; Type: INDEX; Schema: public; Owner: -
-- Name: uk_job_roles_current; Type: INDEX; Schema: public; Owner: -
-- Name: uk_job_roles_record; Type: INDEX; Schema: public; Owner: -
-- Name: uk_org_current; Type: INDEX; Schema: public; Owner: -
-- Name: uk_org_current_active_only; Type: INDEX; Schema: public; Owner: -
-- Name: uk_org_temporal_point; Type: INDEX; Schema: public; Owner: -
-- Name: uk_org_ver_active_only; Type: INDEX; Schema: public; Owner: -
-- Name: uk_position_assignments_active; Type: INDEX; Schema: public; Owner: -
-- Name: uk_position_assignments_effective; Type: INDEX; Schema: public; Owner: -
-- Name: uk_positions_current_active; Type: INDEX; Schema: public; Owner: -
-- Name: update_hierarchy_paths(); Type: FUNCTION; Schema: public; Owner: -
-- Name: validate_hierarchy_changes(); Type: FUNCTION; Schema: public; Owner: -
-- Name: validate_parent_available(); Type: FUNCTION; Schema: public; Owner: -
-- PostgreSQL database dump
-- PostgreSQL database dump complete
ALTER TABLE ONLY public.audit_logs
ALTER TABLE ONLY public.job_families
ALTER TABLE ONLY public.job_families
ALTER TABLE ONLY public.job_families
ALTER TABLE ONLY public.job_families
ALTER TABLE ONLY public.job_family_groups
ALTER TABLE ONLY public.job_family_groups
ALTER TABLE ONLY public.job_family_groups
ALTER TABLE ONLY public.job_levels
ALTER TABLE ONLY public.job_levels
ALTER TABLE ONLY public.job_levels
ALTER TABLE ONLY public.job_levels
ALTER TABLE ONLY public.job_roles
ALTER TABLE ONLY public.job_roles
ALTER TABLE ONLY public.job_roles
ALTER TABLE ONLY public.job_roles
ALTER TABLE ONLY public.organization_units
ALTER TABLE ONLY public.position_assignments
ALTER TABLE ONLY public.position_assignments
ALTER TABLE ONLY public.positions
ALTER TABLE ONLY public.positions
ALTER TABLE ONLY public.positions
ALTER TABLE ONLY public.positions
ALTER TABLE ONLY public.positions
ALTER TABLE ONLY public.positions
ALTER TABLE ONLY public.positions
BEGIN
BEGIN
BEGIN
BEGIN
BEGIN
BEGIN
BEGIN
BEGIN
BEGIN
BEGIN
BEGIN
CASE
CASE
COMMENT ON COLUMN public.audit_logs.record_id IS '组织单元时态版本的唯一标识，用于精确审计查询';
COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';
CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;
CREATE FUNCTION public.calculate_field_changes(old_record jsonb, new_record jsonb) RETURNS TABLE(changes jsonb, modified_fields jsonb)
CREATE FUNCTION public.calculate_org_hierarchy(p_code character varying, p_tenant_id uuid) RETURNS TABLE(calculated_level integer, calculated_code_path character varying, calculated_name_path character varying, calculated_hierarchy_depth integer)
CREATE FUNCTION public.check_temporal_continuity(p_tenant_id uuid, p_code character varying) RETURNS TABLE(issue_type text, effective_date date, end_date date, message text)
CREATE FUNCTION public.enforce_soft_delete_temporal_flags() RETURNS trigger
CREATE FUNCTION public.enforce_temporal_flags() RETURNS trigger
CREATE FUNCTION public.get_organization_temporal(p_tenant_id uuid, p_code character varying, p_as_of_date date DEFAULT CURRENT_DATE) RETURNS TABLE(code character varying, name character varying, unit_type character varying, status character varying, parent_code character varying, effective_date date, end_date date, is_current boolean, change_reason text)
CREATE FUNCTION public.infer_audit_change_datatype(value jsonb) RETURNS text
CREATE FUNCTION public.log_audit_changes() RETURNS trigger
CREATE FUNCTION public.prevent_update_deleted() RETURNS trigger
CREATE FUNCTION public.update_hierarchy_paths() RETURNS trigger
CREATE FUNCTION public.validate_hierarchy_changes() RETURNS trigger
CREATE FUNCTION public.validate_parent_available() RETURNS trigger
CREATE INDEX idx_audit_logs_record_id_time ON public.audit_logs USING btree (record_id, "timestamp" DESC);
CREATE INDEX idx_audit_logs_resource ON public.audit_logs USING btree (resource_type, resource_id);
CREATE INDEX idx_audit_logs_resource_timestamp ON public.audit_logs USING btree (resource_type, resource_id, "timestamp" DESC);
CREATE INDEX idx_audit_logs_timestamp ON public.audit_logs USING btree ("timestamp");
CREATE INDEX idx_org_unit_type_optimized ON public.organization_units USING btree (tenant_id, unit_type, is_current) WHERE (is_current = true);
CREATE INDEX idx_org_units_code_current_active ON public.organization_units USING btree (code) WHERE ((is_current = true) AND ((status)::text <> 'DELETED'::text));
CREATE INDEX idx_org_units_parent ON public.organization_units USING btree (parent_code);
CREATE INDEX idx_org_units_tenant ON public.organization_units USING btree (tenant_id);
CREATE INDEX idx_organization_current_only ON public.organization_units USING btree (tenant_id, code) WHERE (is_current = true);
CREATE INDEX idx_organization_date_range ON public.organization_units USING btree (tenant_id, effective_date, end_date);
CREATE INDEX idx_organization_temporal_main ON public.organization_units USING btree (tenant_id, code, effective_date DESC NULLS LAST, is_current);
CREATE INDEX idx_organization_units_effective_from ON public.organization_units USING btree (effective_from);
CREATE INDEX idx_organization_units_effective_to ON public.organization_units USING btree (effective_to);
CREATE INDEX idx_position_assignments_auto_revert_due ON public.position_assignments USING btree (tenant_id, auto_revert, acting_until) WHERE ((assignment_type)::text = 'ACTING'::text);
CREATE INDEX idx_position_assignments_employee ON public.position_assignments USING btree (tenant_id, employee_id, effective_date DESC);
CREATE INDEX idx_position_assignments_position ON public.position_assignments USING btree (tenant_id, position_code, effective_date DESC);
CREATE INDEX idx_position_assignments_status ON public.position_assignments USING btree (tenant_id, assignment_status, is_current);
CREATE INDEX idx_positions_current ON public.positions USING btree (tenant_id) WHERE (is_current = true);
CREATE INDEX idx_positions_effective_date ON public.positions USING btree (tenant_id, effective_date);
CREATE INDEX idx_positions_job_family ON public.positions USING btree (tenant_id, job_family_code, is_current);
CREATE INDEX idx_positions_job_family_group ON public.positions USING btree (tenant_id, job_family_group_code, is_current);
CREATE INDEX idx_positions_job_role ON public.positions USING btree (tenant_id, job_role_code, is_current);
CREATE INDEX idx_positions_org_code ON public.positions USING btree (tenant_id, organization_code, is_current);
CREATE INDEX idx_positions_status ON public.positions USING btree (tenant_id, status, is_current);
CREATE INDEX ix_org_adjacent_versions ON public.organization_units USING btree (tenant_id, code, effective_date, record_id) WHERE ((status)::text <> 'DELETED'::text);
CREATE INDEX ix_org_current_lookup ON public.organization_units USING btree (tenant_id, code, is_current) WHERE ((is_current = true) AND ((status)::text <> 'DELETED'::text));
CREATE INDEX ix_org_daily_transition ON public.organization_units USING btree (effective_date, end_date, is_current) WHERE ((status)::text <> 'DELETED'::text);
CREATE INDEX ix_org_temporal_boundaries ON public.organization_units USING btree (code, effective_date, end_date, is_current) WHERE ((status)::text <> 'DELETED'::text);
CREATE INDEX ix_org_temporal_query ON public.organization_units USING btree (tenant_id, code, effective_date DESC) WHERE ((status)::text <> 'DELETED'::text);
CREATE TABLE public.audit_logs (
CREATE TABLE public.job_families (
CREATE TABLE public.job_family_groups (
CREATE TABLE public.job_levels (
CREATE TABLE public.job_roles (
CREATE TABLE public.organization_units (
CREATE TABLE public.organization_units_backup_temporal (
CREATE TABLE public.organization_units_unittype_backup (
CREATE TABLE public.position_assignments (
CREATE TABLE public.positions (
CREATE TRIGGER audit_changes_trigger AFTER INSERT OR DELETE OR UPDATE ON public.organization_units FOR EACH ROW EXECUTE FUNCTION public.log_audit_changes();
CREATE TRIGGER enforce_temporal_flags_trigger BEFORE INSERT OR UPDATE ON public.organization_units FOR EACH ROW EXECUTE FUNCTION public.enforce_temporal_flags();
CREATE TRIGGER trg_prevent_update_deleted BEFORE UPDATE ON public.organization_units FOR EACH ROW WHEN (((old.status)::text = 'DELETED'::text)) EXECUTE FUNCTION public.prevent_update_deleted();
CREATE TRIGGER update_hierarchy_paths_trigger BEFORE INSERT OR UPDATE ON public.organization_units FOR EACH ROW EXECUTE FUNCTION public.update_hierarchy_paths();
CREATE TRIGGER validate_parent_available_trigger BEFORE INSERT ON public.organization_units FOR EACH ROW EXECUTE FUNCTION public.validate_parent_available();
CREATE TRIGGER validate_parent_available_update_trigger BEFORE UPDATE ON public.organization_units FOR EACH ROW WHEN (((new.parent_code IS NOT NULL) AND ((new.parent_code)::text IS DISTINCT FROM (old.parent_code)::text))) EXECUTE FUNCTION public.validate_parent_available();
CREATE UNIQUE INDEX uidx_org_record_id ON public.organization_units USING btree (record_id);
CREATE UNIQUE INDEX uk_job_families_current ON public.job_families USING btree (tenant_id, family_code) WHERE (is_current = true);
CREATE UNIQUE INDEX uk_job_families_record ON public.job_families USING btree (record_id, tenant_id, family_code);
CREATE UNIQUE INDEX uk_job_family_groups_current ON public.job_family_groups USING btree (tenant_id, family_group_code) WHERE (is_current = true);
CREATE UNIQUE INDEX uk_job_family_groups_record ON public.job_family_groups USING btree (record_id, tenant_id, family_group_code);
CREATE UNIQUE INDEX uk_job_levels_current ON public.job_levels USING btree (tenant_id, level_code) WHERE (is_current = true);
CREATE UNIQUE INDEX uk_job_levels_record ON public.job_levels USING btree (record_id, tenant_id, level_code);
CREATE UNIQUE INDEX uk_job_roles_current ON public.job_roles USING btree (tenant_id, role_code) WHERE (is_current = true);
CREATE UNIQUE INDEX uk_job_roles_record ON public.job_roles USING btree (record_id, tenant_id, role_code);
CREATE UNIQUE INDEX uk_org_current ON public.organization_units USING btree (tenant_id, code) WHERE ((is_current = true) AND ((status)::text <> 'DELETED'::text));
CREATE UNIQUE INDEX uk_org_current_active_only ON public.organization_units USING btree (tenant_id, code) WHERE ((is_current = true) AND ((status)::text <> 'DELETED'::text));
CREATE UNIQUE INDEX uk_org_temporal_point ON public.organization_units USING btree (tenant_id, code, effective_date) WHERE ((status)::text <> 'DELETED'::text);
CREATE UNIQUE INDEX uk_org_ver_active_only ON public.organization_units USING btree (tenant_id, code, effective_date) WHERE ((status)::text <> 'DELETED'::text);
CREATE UNIQUE INDEX uk_position_assignments_active ON public.position_assignments USING btree (tenant_id, position_code, employee_id) WHERE ((is_current = true) AND ((assignment_status)::text = 'ACTIVE'::text));
CREATE UNIQUE INDEX uk_position_assignments_effective ON public.position_assignments USING btree (tenant_id, position_code, employee_id, effective_date);
CREATE UNIQUE INDEX uk_positions_current_active ON public.positions USING btree (tenant_id, code) WHERE ((is_current = true) AND ((status)::text <> 'DELETED'::text));
CREATE VIEW public.organization_current AS
CREATE VIEW public.organization_stats_view AS
CREATE VIEW public.organization_temporal_current AS
DECLARE
DECLARE
DECLARE
DECLARE
DECLARE
END),
END),
END;
END;
END;
END;
END;
END;
END;
END;
END;
END;
END;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET client_encoding = 'UTF8';
SET client_min_messages = warning;
SET default_table_access_method = heap;
SET default_tablespace = '';
SET idle_in_transaction_session_timeout = 0;
SET lock_timeout = 0;
SET row_security = off;
SET standard_conforming_strings = on;
SET statement_timeout = 0;
SET xmloption = content;
