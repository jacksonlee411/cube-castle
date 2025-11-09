--
-- PostgreSQL database dump
--

-- Dumped from database version 16.9
-- Dumped by pg_dump version 16.9

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


--
-- Name: calculate_field_changes(jsonb, jsonb); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.calculate_field_changes(old_record jsonb, new_record jsonb) RETURNS TABLE(changes jsonb, modified_fields jsonb)
    LANGUAGE plpgsql
    AS $$
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
$$;


--
-- Name: calculate_org_hierarchy(character varying, uuid); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.calculate_org_hierarchy(p_code character varying, p_tenant_id uuid) RETURNS TABLE(calculated_level integer, calculated_code_path character varying, calculated_name_path character varying, calculated_hierarchy_depth integer)
    LANGUAGE plpgsql
    AS $$
DECLARE
    parent_info RECORD;
    current_name VARCHAR(255);
BEGIN
    SELECT name INTO current_name
      FROM organization_units
     WHERE code = p_code AND tenant_id = p_tenant_id AND is_current = true
     LIMIT 1;

    SELECT
        ou.code,
        ou.level,
        ou.code_path,
        ou.name_path,
        ou.hierarchy_depth
      INTO parent_info
      FROM organization_units ou
     WHERE ou.code = (
            SELECT parent_code
              FROM organization_units
             WHERE code = p_code AND tenant_id = p_tenant_id AND is_current = true
             LIMIT 1
          )
       AND ou.tenant_id = p_tenant_id
       AND ou.is_current = true
       AND ou.status <> 'DELETED'
     LIMIT 1;

    IF parent_info.code IS NULL THEN
        calculated_level := 1;
        calculated_hierarchy_depth := 1;
        calculated_code_path := '/' || p_code;
        calculated_name_path := '/' || COALESCE(current_name, p_code);
    ELSE
        calculated_level := parent_info.level + 1;
        calculated_hierarchy_depth := parent_info.hierarchy_depth + 1;
        calculated_code_path := COALESCE(parent_info.code_path, '/' || parent_info.code) || '/' || p_code;
        calculated_name_path := COALESCE(parent_info.name_path, '/' || current_name) || '/' || COALESCE(current_name, p_code);
        IF calculated_hierarchy_depth > 17 THEN
            RAISE EXCEPTION '组织层级超过最大限制17级！当前尝试创建第%级组织。', calculated_hierarchy_depth;
        END IF;
    END IF;
    RETURN NEXT;
END;
$$;


--
-- Name: check_temporal_continuity(uuid, character varying); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.check_temporal_continuity(p_tenant_id uuid, p_code character varying) RETURNS TABLE(issue_type text, effective_date date, end_date date, message text)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    WITH ordered_versions AS (
        SELECT
            effective_date,
            end_date,
            ROW_NUMBER() OVER (ORDER BY effective_date) AS rn
        FROM organization_units
        WHERE tenant_id = p_tenant_id
          AND code = p_code
          AND status <> 'DELETED'
        ORDER BY effective_date
    ), version_overlaps AS (
        SELECT
            curr.effective_date,
            curr.end_date,
            'OVERLAP'::TEXT AS issue_type,
            'Version overlaps with next version'::TEXT AS message
        FROM ordered_versions curr
        JOIN ordered_versions nxt ON nxt.rn = curr.rn + 1
        WHERE curr.end_date IS NOT NULL
          AND curr.end_date >= nxt.effective_date
    ), gaps AS (
        SELECT
            curr.effective_date,
            curr.end_date,
            'GAP'::TEXT AS issue_type,
            'Gap between versions'::TEXT AS message
        FROM ordered_versions curr
        JOIN ordered_versions nxt ON nxt.rn = curr.rn + 1
        WHERE curr.end_date IS NOT NULL
          AND curr.end_date + INTERVAL '1 day' < nxt.effective_date
    )
    SELECT * FROM version_overlaps
    UNION ALL
    SELECT * FROM gaps;
END;
$$;


--
-- Name: enforce_soft_delete_temporal_flags(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.enforce_soft_delete_temporal_flags() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NEW.status = 'DELETED' THEN
        NEW.is_current := FALSE;
        RETURN NEW;
    END IF;

    IF NEW.effective_date > CURRENT_DATE THEN
        NEW.is_current := FALSE;
    ELSIF NEW.end_date IS NOT NULL AND NEW.end_date <= CURRENT_DATE THEN
        NEW.is_current := FALSE;
    ELSE
        NEW.is_current := TRUE;
    END IF;
    RETURN NEW;
END;
$$;


--
-- Name: get_organization_temporal(uuid, character varying, date); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.get_organization_temporal(p_tenant_id uuid, p_code character varying, p_as_of_date date DEFAULT CURRENT_DATE) RETURNS TABLE(code character varying, name character varying, unit_type character varying, status character varying, parent_code character varying, effective_date date, end_date date, is_current boolean, change_reason text)
    LANGUAGE sql STABLE
    AS $$
    SELECT 
        ou.code,
        ou.name,
        ou.unit_type,
        ou.status,
        ou.parent_code,
        ou.effective_date,
        ou.end_date,
        ou.is_current,
        ou.change_reason
    FROM organization_units ou
    WHERE ou.tenant_id = p_tenant_id
      AND ou.code = p_code
      AND COALESCE(ou.effective_date, CURRENT_DATE) <= p_as_of_date
      AND (ou.end_date IS NULL OR ou.end_date > p_as_of_date)
    ORDER BY ou.effective_date DESC
    LIMIT 1;
$$;


--
-- Name: infer_audit_change_datatype(jsonb); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.infer_audit_change_datatype(value jsonb) RETURNS text
    LANGUAGE plpgsql IMMUTABLE
    AS $$
DECLARE
    value_type TEXT;
BEGIN
    IF value IS NULL OR value = 'null'::JSONB THEN
        RETURN 'unknown';
    END IF;

    value_type := jsonb_typeof(value);

    IF value_type = 'string' THEN
        RETURN 'string';
    ELSIF value_type = 'number' THEN
        RETURN 'number';
    ELSIF value_type = 'boolean' THEN
        RETURN 'boolean';
    ELSIF value_type = 'array' THEN
        RETURN 'array';
    ELSIF value_type = 'object' THEN
        RETURN 'object';
    ELSE
        RETURN 'unknown';
    END IF;
END;
$$;


--
-- Name: validate_hierarchy_changes(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.validate_hierarchy_changes() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NEW.parent_code IS NOT NULL THEN
        IF check_circular_reference(NEW.code, NEW.parent_code, NEW.tenant_id) THEN
            RAISE EXCEPTION '不能设置父组织，会导致循环引用！组织 % 尝试设置父组织 %', NEW.code, NEW.parent_code;
        END IF;
    END IF;

    IF NEW.parent_code IS NOT NULL THEN
        IF NOT EXISTS (
            SELECT 1 FROM organization_units
             WHERE code = NEW.parent_code
               AND tenant_id = NEW.tenant_id
               AND is_current = true
               AND status <> 'DELETED'
        ) THEN
            RAISE EXCEPTION '父组织不可用（不存在/已删除/非当前）！父组织编码: %', NEW.parent_code;
        END IF;
    END IF;
    RETURN NEW;
END;
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: audit_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.audit_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    event_type character varying(50) NOT NULL,
    resource_type character varying(50) NOT NULL,
    resource_id character varying(100),
    actor_id character varying(100),
    actor_type character varying(50),
    action_name character varying(100),
    request_id character varying(100),
    operation_reason text,
    "timestamp" timestamp with time zone DEFAULT now() NOT NULL,
    success boolean DEFAULT true NOT NULL,
    error_code character varying(100),
    error_message text,
    request_data jsonb DEFAULT '{}'::jsonb NOT NULL,
    response_data jsonb DEFAULT '{}'::jsonb NOT NULL,
    modified_fields jsonb DEFAULT '[]'::jsonb NOT NULL,
    changes jsonb DEFAULT '[]'::jsonb NOT NULL,
    record_id uuid,
    business_context jsonb DEFAULT '{}'::jsonb NOT NULL,
    CONSTRAINT audit_logs_event_type_check_v2 CHECK (((event_type)::text = ANY ((ARRAY['CREATE'::character varying, 'UPDATE'::character varying, 'DELETE'::character varying, 'SUSPEND'::character varying, 'REACTIVATE'::character varying, 'QUERY'::character varying, 'VALIDATION'::character varying, 'AUTHENTICATION'::character varying, 'ERROR'::character varying])::text[])))
);


--
-- Name: COLUMN audit_logs.record_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.audit_logs.record_id IS '组织单元时态版本的唯一标识，用于精确审计查询';


--
-- Name: goose_db_version; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.goose_db_version (
    id integer NOT NULL,
    version_id bigint NOT NULL,
    is_applied boolean NOT NULL,
    tstamp timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: goose_db_version_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.goose_db_version ALTER COLUMN id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.goose_db_version_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: job_families; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.job_families (
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    family_code character varying(20) NOT NULL,
    family_group_code character varying(20) NOT NULL,
    parent_record_id uuid NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    status character varying(20) DEFAULT 'ACTIVE'::character varying NOT NULL,
    effective_date date NOT NULL,
    end_date date,
    is_current boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- Name: job_family_groups; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.job_family_groups (
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    family_group_code character varying(20) NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    status character varying(20) DEFAULT 'ACTIVE'::character varying NOT NULL,
    effective_date date NOT NULL,
    end_date date,
    is_current boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- Name: job_levels; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.job_levels (
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    level_code character varying(20) NOT NULL,
    role_code character varying(20) NOT NULL,
    parent_record_id uuid NOT NULL,
    level_rank character varying(20) NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    salary_band jsonb,
    status character varying(20) DEFAULT 'ACTIVE'::character varying NOT NULL,
    effective_date date NOT NULL,
    end_date date,
    is_current boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- Name: job_roles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.job_roles (
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    role_code character varying(20) NOT NULL,
    family_code character varying(20) NOT NULL,
    parent_record_id uuid NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    competency_model jsonb DEFAULT '{}'::jsonb,
    status character varying(20) DEFAULT 'ACTIVE'::character varying NOT NULL,
    effective_date date NOT NULL,
    end_date date,
    is_current boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- Name: organization_units; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.organization_units (
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    code character varying(12) NOT NULL,
    parent_code character varying(12),
    name character varying(255) NOT NULL,
    unit_type character varying(64) NOT NULL,
    status character varying(20) DEFAULT 'ACTIVE'::character varying NOT NULL,
    level integer DEFAULT 1 NOT NULL,
    hierarchy_depth integer DEFAULT 0 NOT NULL,
    code_path text DEFAULT ''::text NOT NULL,
    name_path text DEFAULT ''::text NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL,
    description text,
    profile jsonb DEFAULT '{}'::jsonb NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    effective_date date DEFAULT CURRENT_DATE NOT NULL,
    end_date date,
    change_reason text,
    is_current boolean DEFAULT false NOT NULL,
    deleted_at timestamp with time zone,
    deleted_by uuid,
    deletion_reason text,
    suspended_at timestamp with time zone,
    suspended_by uuid,
    suspension_reason text,
    operated_by_id uuid,
    operated_by_name text,
    operation_type character varying(20) DEFAULT 'CREATE'::character varying,
    effective_from timestamp with time zone,
    effective_to timestamp with time zone,
    changed_by uuid,
    approved_by uuid,
    CONSTRAINT chk_deleted_not_current CHECK (
CASE
    WHEN ((status)::text = 'DELETED'::text) THEN (is_current = false)
    ELSE true
END),
    CONSTRAINT chk_org_units_not_deleted_current CHECK (
CASE
    WHEN (((status)::text = 'DELETED'::text) OR (deleted_at IS NOT NULL)) THEN (is_current = false)
    ELSE true
END),
    CONSTRAINT valid_unit_type CHECK (((unit_type)::text = ANY ((ARRAY['DEPARTMENT'::character varying, 'ORGANIZATION_UNIT'::character varying, 'PROJECT_TEAM'::character varying])::text[])))
);


--
-- Name: organization_current; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.organization_current AS
 SELECT tenant_id,
    code,
    parent_code,
    name,
    unit_type,
    status,
    level,
    code_path,
    name_path,
    sort_order,
    description,
    profile,
    effective_date,
    end_date,
    is_current,
    change_reason,
    created_at,
    updated_at
   FROM public.organization_units ou
  WHERE ((is_current = true) AND ((end_date IS NULL) OR (end_date > CURRENT_DATE)));


--
-- Name: organization_stats_view; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.organization_stats_view AS
 SELECT tenant_id,
    unit_type,
    count(*) AS count,
    count(
        CASE
            WHEN (is_current = true) THEN 1
            ELSE NULL::integer
        END) AS current_count,
    count(
        CASE
            WHEN (is_current = false) THEN 1
            ELSE NULL::integer
        END) AS historical_count
   FROM public.organization_units
  WHERE (deleted_at IS NULL)
  GROUP BY tenant_id, unit_type;


--
-- Name: organization_temporal_current; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.organization_temporal_current AS
 SELECT record_id,
    tenant_id,
    code,
    parent_code,
    name,
    unit_type,
    status,
    level,
    hierarchy_depth,
    code_path,
    name_path,
    sort_order,
    description,
    profile,
    created_at,
    updated_at,
    effective_date,
    end_date,
    change_reason,
    is_current,
    deleted_at,
    deleted_by,
    deletion_reason,
    suspended_at,
    suspended_by,
    suspension_reason,
    operated_by_id,
    operated_by_name,
    metadata,
    effective_from,
    effective_to,
    changed_by,
    approved_by
   FROM public.organization_units
  WHERE (is_current = true);


--
-- Name: organization_units_backup_temporal; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.organization_units_backup_temporal (
    record_id uuid,
    tenant_id uuid,
    code character varying(12),
    parent_code character varying(12),
    name character varying(255),
    unit_type character varying(64),
    status character varying(20),
    level integer,
    hierarchy_depth integer,
    code_path text,
    name_path text,
    sort_order integer,
    description text,
    profile jsonb,
    metadata jsonb,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    effective_date date,
    end_date date,
    change_reason text,
    is_current boolean,
    deleted_at timestamp with time zone,
    deleted_by uuid,
    deletion_reason text,
    suspended_at timestamp with time zone,
    suspended_by uuid,
    suspension_reason text,
    operated_by_id uuid,
    operated_by_name text,
    operation_type character varying(20),
    effective_from timestamp with time zone,
    effective_to timestamp with time zone,
    changed_by uuid,
    approved_by uuid,
    is_temporal boolean
);


--
-- Name: organization_units_unittype_backup; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.organization_units_unittype_backup (
    record_id uuid,
    tenant_id uuid,
    code character varying(12),
    parent_code character varying(12),
    name character varying(255),
    unit_type character varying(64),
    status character varying(20),
    level integer,
    hierarchy_depth integer,
    code_path text,
    name_path text,
    sort_order integer,
    description text,
    profile jsonb,
    metadata jsonb,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    effective_date date,
    end_date date,
    change_reason text,
    is_current boolean,
    deleted_at timestamp with time zone,
    deleted_by uuid,
    deletion_reason text,
    suspended_at timestamp with time zone,
    suspended_by uuid,
    suspension_reason text,
    operated_by_id uuid,
    operated_by_name text,
    operation_type character varying(20),
    effective_from timestamp with time zone,
    effective_to timestamp with time zone,
    changed_by uuid,
    approved_by uuid,
    is_temporal boolean
);


--
-- Name: position_assignments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.position_assignments (
    assignment_id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    position_code character varying(8) NOT NULL,
    position_record_id uuid NOT NULL,
    employee_id uuid NOT NULL,
    employee_name character varying(255) NOT NULL,
    employee_number character varying(64),
    assignment_type character varying(20) NOT NULL,
    assignment_status character varying(20) DEFAULT 'ACTIVE'::character varying NOT NULL,
    fte numeric(5,2) DEFAULT 1.0 NOT NULL,
    effective_date date NOT NULL,
    end_date date,
    is_current boolean DEFAULT false NOT NULL,
    notes text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    acting_until date,
    auto_revert boolean DEFAULT false NOT NULL,
    reminder_sent_at timestamp with time zone,
    CONSTRAINT chk_position_assignments_auto_revert CHECK (((auto_revert = false) OR (((assignment_type)::text = 'ACTING'::text) AND (acting_until IS NOT NULL)))),
    CONSTRAINT chk_position_assignments_dates CHECK ((((end_date IS NULL) OR (end_date > effective_date)) AND ((acting_until IS NULL) OR (acting_until > effective_date)))),
    CONSTRAINT chk_position_assignments_fte CHECK (((fte >= (0)::numeric) AND (fte <= (1)::numeric))),
    CONSTRAINT chk_position_assignments_status CHECK (((assignment_status)::text = ANY ((ARRAY['PENDING'::character varying, 'ACTIVE'::character varying, 'ENDED'::character varying])::text[]))),
    CONSTRAINT chk_position_assignments_type CHECK (((assignment_type)::text = ANY ((ARRAY['PRIMARY'::character varying, 'SECONDARY'::character varying, 'ACTING'::character varying])::text[])))
);


--
-- Name: positions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.positions (
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    code character varying(8) NOT NULL,
    title character varying(120) NOT NULL,
    job_profile_code character varying(64),
    job_profile_name character varying(255),
    job_family_group_code character varying(20) NOT NULL,
    job_family_group_name character varying(255) NOT NULL,
    job_family_group_record_id uuid NOT NULL,
    job_family_code character varying(20) NOT NULL,
    job_family_name character varying(255) NOT NULL,
    job_family_record_id uuid NOT NULL,
    job_role_code character varying(20) NOT NULL,
    job_role_name character varying(255) NOT NULL,
    job_role_record_id uuid NOT NULL,
    job_level_code character varying(20) NOT NULL,
    job_level_name character varying(255) NOT NULL,
    job_level_record_id uuid NOT NULL,
    organization_code character varying(7) NOT NULL,
    organization_name character varying(255),
    position_type character varying(50) NOT NULL,
    status character varying(20) DEFAULT 'PLANNED'::character varying NOT NULL,
    employment_type character varying(50) NOT NULL,
    headcount_capacity numeric(5,2) DEFAULT 1.0 NOT NULL,
    headcount_in_use numeric(5,2) DEFAULT 0.0 NOT NULL,
    grade_level character varying(20),
    cost_center_code character varying(50),
    reports_to_position_code character varying(8),
    profile jsonb DEFAULT '{}'::jsonb NOT NULL,
    effective_date date NOT NULL,
    end_date date,
    is_current boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at timestamp with time zone,
    operation_type character varying(20) DEFAULT 'CREATE'::character varying NOT NULL,
    operated_by_id uuid NOT NULL,
    operated_by_name character varying(255) NOT NULL,
    operation_reason text
);


--
-- Name: audit_logs audit_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);


--
-- Name: goose_db_version goose_db_version_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.goose_db_version
    ADD CONSTRAINT goose_db_version_pkey PRIMARY KEY (id);


--
-- Name: job_families job_families_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.job_families
    ADD CONSTRAINT job_families_pkey PRIMARY KEY (record_id);


--
-- Name: job_families job_families_record_id_tenant_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.job_families
    ADD CONSTRAINT job_families_record_id_tenant_id_key UNIQUE (record_id, tenant_id);


--
-- Name: job_families job_families_tenant_id_family_code_effective_date_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.job_families
    ADD CONSTRAINT job_families_tenant_id_family_code_effective_date_key UNIQUE (tenant_id, family_code, effective_date);


--
-- Name: job_family_groups job_family_groups_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.job_family_groups
    ADD CONSTRAINT job_family_groups_pkey PRIMARY KEY (record_id);


--
-- Name: job_family_groups job_family_groups_record_id_tenant_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.job_family_groups
    ADD CONSTRAINT job_family_groups_record_id_tenant_id_key UNIQUE (record_id, tenant_id);


--
-- Name: job_family_groups job_family_groups_tenant_id_family_group_code_effective_dat_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.job_family_groups
    ADD CONSTRAINT job_family_groups_tenant_id_family_group_code_effective_dat_key UNIQUE (tenant_id, family_group_code, effective_date);


--
-- Name: job_levels job_levels_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.job_levels
    ADD CONSTRAINT job_levels_pkey PRIMARY KEY (record_id);


--
-- Name: job_levels job_levels_record_id_tenant_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.job_levels
    ADD CONSTRAINT job_levels_record_id_tenant_id_key UNIQUE (record_id, tenant_id);


--
-- Name: job_levels job_levels_tenant_id_level_code_effective_date_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.job_levels
    ADD CONSTRAINT job_levels_tenant_id_level_code_effective_date_key UNIQUE (tenant_id, level_code, effective_date);


--
-- Name: job_roles job_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.job_roles
    ADD CONSTRAINT job_roles_pkey PRIMARY KEY (record_id);


--
-- Name: job_roles job_roles_record_id_tenant_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.job_roles
    ADD CONSTRAINT job_roles_record_id_tenant_id_key UNIQUE (record_id, tenant_id);


--
-- Name: job_roles job_roles_tenant_id_role_code_effective_date_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.job_roles
    ADD CONSTRAINT job_roles_tenant_id_role_code_effective_date_key UNIQUE (tenant_id, role_code, effective_date);


--
-- Name: organization_units pk_org_record_id; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organization_units
    ADD CONSTRAINT pk_org_record_id PRIMARY KEY (record_id);


--
-- Name: position_assignments position_assignments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.position_assignments
    ADD CONSTRAINT position_assignments_pkey PRIMARY KEY (assignment_id);


--
-- Name: positions positions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.positions
    ADD CONSTRAINT positions_pkey PRIMARY KEY (record_id);


--
-- Name: positions positions_tenant_id_code_effective_date_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.positions
    ADD CONSTRAINT positions_tenant_id_code_effective_date_key UNIQUE (tenant_id, code, effective_date);


--
-- Name: positions positions_tenant_id_code_record_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.positions
    ADD CONSTRAINT positions_tenant_id_code_record_id_key UNIQUE (tenant_id, code, record_id);


--
-- Name: idx_audit_logs_record_id_time; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_record_id_time ON public.audit_logs USING btree (record_id, "timestamp" DESC);


--
-- Name: idx_audit_logs_resource; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_resource ON public.audit_logs USING btree (resource_type, resource_id);


--
-- Name: idx_audit_logs_resource_timestamp; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_resource_timestamp ON public.audit_logs USING btree (resource_type, resource_id, "timestamp" DESC);


--
-- Name: idx_audit_logs_timestamp; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_timestamp ON public.audit_logs USING btree ("timestamp");


--
-- Name: idx_org_unit_type_optimized; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_org_unit_type_optimized ON public.organization_units USING btree (tenant_id, unit_type, is_current) WHERE (is_current = true);


--
-- Name: idx_org_units_code_current_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_org_units_code_current_active ON public.organization_units USING btree (code) WHERE ((is_current = true) AND ((status)::text <> 'DELETED'::text));


--
-- Name: idx_org_units_parent; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_org_units_parent ON public.organization_units USING btree (parent_code);


--
-- Name: idx_org_units_tenant; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_org_units_tenant ON public.organization_units USING btree (tenant_id);


--
-- Name: idx_organization_current_only; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_organization_current_only ON public.organization_units USING btree (tenant_id, code) WHERE (is_current = true);


--
-- Name: idx_organization_date_range; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_organization_date_range ON public.organization_units USING btree (tenant_id, effective_date, end_date);


--
-- Name: idx_organization_temporal_main; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_organization_temporal_main ON public.organization_units USING btree (tenant_id, code, effective_date DESC NULLS LAST, is_current);


--
-- Name: idx_organization_units_effective_from; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_organization_units_effective_from ON public.organization_units USING btree (effective_from);


--
-- Name: idx_organization_units_effective_to; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_organization_units_effective_to ON public.organization_units USING btree (effective_to);


--
-- Name: idx_position_assignments_auto_revert_due; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_position_assignments_auto_revert_due ON public.position_assignments USING btree (tenant_id, auto_revert, acting_until) WHERE ((assignment_type)::text = 'ACTING'::text);


--
-- Name: idx_position_assignments_employee; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_position_assignments_employee ON public.position_assignments USING btree (tenant_id, employee_id, effective_date DESC);


--
-- Name: idx_position_assignments_position; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_position_assignments_position ON public.position_assignments USING btree (tenant_id, position_code, effective_date DESC);


--
-- Name: idx_position_assignments_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_position_assignments_status ON public.position_assignments USING btree (tenant_id, assignment_status, is_current);


--
-- Name: idx_positions_current; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_positions_current ON public.positions USING btree (tenant_id) WHERE (is_current = true);


--
-- Name: idx_positions_effective_date; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_positions_effective_date ON public.positions USING btree (tenant_id, effective_date);


--
-- Name: idx_positions_job_family; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_positions_job_family ON public.positions USING btree (tenant_id, job_family_code, is_current);


--
-- Name: idx_positions_job_family_group; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_positions_job_family_group ON public.positions USING btree (tenant_id, job_family_group_code, is_current);


--
-- Name: idx_positions_job_role; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_positions_job_role ON public.positions USING btree (tenant_id, job_role_code, is_current);


--
-- Name: idx_positions_org_code; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_positions_org_code ON public.positions USING btree (tenant_id, organization_code, is_current);


--
-- Name: idx_positions_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_positions_status ON public.positions USING btree (tenant_id, status, is_current);


--
-- Name: ix_org_adjacent_versions; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_org_adjacent_versions ON public.organization_units USING btree (tenant_id, code, effective_date, record_id) WHERE ((status)::text <> 'DELETED'::text);


--
-- Name: ix_org_current_lookup; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_org_current_lookup ON public.organization_units USING btree (tenant_id, code, is_current) WHERE ((is_current = true) AND ((status)::text <> 'DELETED'::text));


--
-- Name: ix_org_daily_transition; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_org_daily_transition ON public.organization_units USING btree (effective_date, end_date, is_current) WHERE ((status)::text <> 'DELETED'::text);


--
-- Name: ix_org_temporal_boundaries; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_org_temporal_boundaries ON public.organization_units USING btree (code, effective_date, end_date, is_current) WHERE ((status)::text <> 'DELETED'::text);


--
-- Name: ix_org_temporal_query; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_org_temporal_query ON public.organization_units USING btree (tenant_id, code, effective_date DESC) WHERE ((status)::text <> 'DELETED'::text);


--
-- Name: uidx_org_record_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uidx_org_record_id ON public.organization_units USING btree (record_id);


--
-- Name: uk_job_families_current; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_job_families_current ON public.job_families USING btree (tenant_id, family_code) WHERE (is_current = true);


--
-- Name: uk_job_families_record; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_job_families_record ON public.job_families USING btree (record_id, tenant_id, family_code);


--
-- Name: uk_job_family_groups_current; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_job_family_groups_current ON public.job_family_groups USING btree (tenant_id, family_group_code) WHERE (is_current = true);


--
-- Name: uk_job_family_groups_record; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_job_family_groups_record ON public.job_family_groups USING btree (record_id, tenant_id, family_group_code);


--
-- Name: uk_job_levels_current; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_job_levels_current ON public.job_levels USING btree (tenant_id, level_code) WHERE (is_current = true);


--
-- Name: uk_job_levels_record; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_job_levels_record ON public.job_levels USING btree (record_id, tenant_id, level_code);


--
-- Name: uk_job_roles_current; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_job_roles_current ON public.job_roles USING btree (tenant_id, role_code) WHERE (is_current = true);


--
-- Name: uk_job_roles_record; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_job_roles_record ON public.job_roles USING btree (record_id, tenant_id, role_code);


--
-- Name: uk_org_current; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_org_current ON public.organization_units USING btree (tenant_id, code) WHERE ((is_current = true) AND ((status)::text <> 'DELETED'::text));


--
-- Name: uk_org_current_active_only; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_org_current_active_only ON public.organization_units USING btree (tenant_id, code) WHERE ((is_current = true) AND ((status)::text <> 'DELETED'::text));


--
-- Name: uk_org_temporal_point; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_org_temporal_point ON public.organization_units USING btree (tenant_id, code, effective_date) WHERE ((status)::text <> 'DELETED'::text);


--
-- Name: uk_org_ver_active_only; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_org_ver_active_only ON public.organization_units USING btree (tenant_id, code, effective_date) WHERE ((status)::text <> 'DELETED'::text);


--
-- Name: uk_position_assignments_active; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_position_assignments_active ON public.position_assignments USING btree (tenant_id, position_code, employee_id) WHERE ((is_current = true) AND ((assignment_status)::text = 'ACTIVE'::text));


--
-- Name: uk_position_assignments_effective; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_position_assignments_effective ON public.position_assignments USING btree (tenant_id, position_code, employee_id, effective_date);


--
-- Name: uk_positions_current_active; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_positions_current_active ON public.positions USING btree (tenant_id, code) WHERE ((is_current = true) AND ((status)::text <> 'DELETED'::text));


--
-- Name: job_families fk_job_families_group; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.job_families
    ADD CONSTRAINT fk_job_families_group FOREIGN KEY (parent_record_id, tenant_id) REFERENCES public.job_family_groups(record_id, tenant_id);


--
-- Name: job_levels fk_job_levels_role; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.job_levels
    ADD CONSTRAINT fk_job_levels_role FOREIGN KEY (parent_record_id, tenant_id) REFERENCES public.job_roles(record_id, tenant_id);


--
-- Name: job_roles fk_job_roles_family; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.job_roles
    ADD CONSTRAINT fk_job_roles_family FOREIGN KEY (parent_record_id, tenant_id) REFERENCES public.job_families(record_id, tenant_id);


--
-- Name: position_assignments fk_position_assignments_position; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.position_assignments
    ADD CONSTRAINT fk_position_assignments_position FOREIGN KEY (tenant_id, position_code, position_record_id) REFERENCES public.positions(tenant_id, code, record_id);


--
-- Name: positions fk_positions_family; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.positions
    ADD CONSTRAINT fk_positions_family FOREIGN KEY (job_family_record_id, tenant_id) REFERENCES public.job_families(record_id, tenant_id);


--
-- Name: positions fk_positions_family_group; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.positions
    ADD CONSTRAINT fk_positions_family_group FOREIGN KEY (job_family_group_record_id, tenant_id) REFERENCES public.job_family_groups(record_id, tenant_id);


--
-- Name: positions fk_positions_level; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.positions
    ADD CONSTRAINT fk_positions_level FOREIGN KEY (job_level_record_id, tenant_id) REFERENCES public.job_levels(record_id, tenant_id);


--
-- Name: positions fk_positions_role; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.positions
    ADD CONSTRAINT fk_positions_role FOREIGN KEY (job_role_record_id, tenant_id) REFERENCES public.job_roles(record_id, tenant_id);


--
-- PostgreSQL database dump complete
--
