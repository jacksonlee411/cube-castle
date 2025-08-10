--
-- PostgreSQL database dump
--

-- Dumped from database version 16.9
-- Dumped by pg_dump version 16.9 (Ubuntu 16.9-0ubuntu0.24.04.1)

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
-- Name: corehr; Type: SCHEMA; Schema: -; Owner: user
--

CREATE SCHEMA corehr;


ALTER SCHEMA corehr OWNER TO "user";

--
-- Name: identity; Type: SCHEMA; Schema: -; Owner: user
--

CREATE SCHEMA identity;


ALTER SCHEMA identity OWNER TO "user";

--
-- Name: intelligence; Type: SCHEMA; Schema: -; Owner: user
--

CREATE SCHEMA intelligence;


ALTER SCHEMA intelligence OWNER TO "user";

--
-- Name: outbox; Type: SCHEMA; Schema: -; Owner: user
--

CREATE SCHEMA outbox;


ALTER SCHEMA outbox OWNER TO "user";

--
-- Name: tenancy; Type: SCHEMA; Schema: -; Owner: user
--

CREATE SCHEMA tenancy;


ALTER SCHEMA tenancy OWNER TO "user";

--
-- Name: workflow; Type: SCHEMA; Schema: -; Owner: user
--

CREATE SCHEMA workflow;


ALTER SCHEMA workflow OWNER TO "user";

--
-- Name: pg_trgm; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;


--
-- Name: EXTENSION pg_trgm; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION pg_trgm IS 'text similarity measurement and index searching based on trigrams';


--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


--
-- Name: audit_position_history_changes(); Type: FUNCTION; Schema: public; Owner: user
--

CREATE FUNCTION public.audit_position_history_changes() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- Log position changes to audit table (if exists)
    IF TG_OP = 'INSERT' THEN
        INSERT INTO audit_log (
            table_name, operation, record_id, tenant_id, 
            changed_by, changed_at, new_values
        ) VALUES (
            'position_history', 'INSERT', NEW.id, NEW.tenant_id,
            NEW.created_by, NOW(), row_to_json(NEW)
        );
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO audit_log (
            table_name, operation, record_id, tenant_id,
            changed_by, changed_at, old_values, new_values
        ) VALUES (
            'position_history', 'UPDATE', NEW.id, NEW.tenant_id,
            NEW.created_by, NOW(), row_to_json(OLD), row_to_json(NEW)
        );
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$;


ALTER FUNCTION public.audit_position_history_changes() OWNER TO "user";

--
-- Name: auto_close_previous_positions(); Type: FUNCTION; Schema: public; Owner: user
--

CREATE FUNCTION public.auto_close_previous_positions() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- If this is a new current position (end_date is NULL), close previous open positions
    IF NEW.end_date IS NULL THEN
        UPDATE position_history 
        SET end_date = NEW.effective_date - INTERVAL '1 day'
        WHERE tenant_id = NEW.tenant_id 
          AND employee_id = NEW.employee_id
          AND id != NEW.id
          AND end_date IS NULL
          AND effective_date < NEW.effective_date;
    END IF;
    
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.auto_close_previous_positions() OWNER TO "user";

--
-- Name: auto_generate_position_code(); Type: FUNCTION; Schema: public; Owner: user
--

CREATE FUNCTION public.auto_generate_position_code() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NEW.code IS NULL THEN
        NEW.code := generate_position_code(NEW.tenant_id);
    END IF;
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.auto_generate_position_code() OWNER TO "user";

--
-- Name: generate_employee_code(); Type: FUNCTION; Schema: public; Owner: user
--

CREATE FUNCTION public.generate_employee_code() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NEW.code IS NULL THEN
        NEW.code := LPAD(nextval('employee_code_seq')::TEXT, 8, '0');
    END IF;
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.generate_employee_code() OWNER TO "user";

--
-- Name: generate_org_unit_code(); Type: FUNCTION; Schema: public; Owner: user
--

CREATE FUNCTION public.generate_org_unit_code() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- 自动生成7位编码（如果为空）
    IF NEW.code IS NULL THEN
        NEW.code := LPAD(nextval('org_unit_code_seq')::text, 7, '0');
    END IF;
    
    -- 自动计算层级和路径
    IF NEW.parent_code IS NOT NULL THEN
        SELECT level + 1, path || '/' || NEW.code 
        INTO NEW.level, NEW.path
        FROM organization_units 
        WHERE code = NEW.parent_code;
    ELSE
        NEW.level := 1;
        NEW.path := '/' || NEW.code;
    END IF;
    
    -- 更新时间戳
    NEW.updated_at := NOW();
    
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.generate_org_unit_code() OWNER TO "user";

--
-- Name: generate_position_code(uuid); Type: FUNCTION; Schema: public; Owner: user
--

CREATE FUNCTION public.generate_position_code(p_tenant_id uuid) RETURNS character varying
    LANGUAGE plpgsql
    AS $$
DECLARE
    current_code INTEGER;
    new_code VARCHAR(7);
BEGIN
    INSERT INTO position_code_sequence (tenant_id, last_code)
    VALUES (p_tenant_id, 1000000)
    ON CONFLICT (tenant_id) DO NOTHING;
    
    UPDATE position_code_sequence 
    SET last_code = last_code + 1,
        updated_at = CURRENT_TIMESTAMP
    WHERE tenant_id = p_tenant_id
    RETURNING last_code INTO current_code;
    
    IF current_code > 9999999 THEN
        RAISE EXCEPTION 'Position code overflow for tenant %', p_tenant_id;
    END IF;
    
    new_code := LPAD(current_code::TEXT, 7, '0');
    RETURN new_code;
END;
$$;


ALTER FUNCTION public.generate_position_code(p_tenant_id uuid) OWNER TO "user";

--
-- Name: get_current_tenant_id(); Type: FUNCTION; Schema: public; Owner: user
--

CREATE FUNCTION public.get_current_tenant_id() RETURNS uuid
    LANGUAGE plpgsql STABLE
    AS $$
BEGIN
    RETURN current_setting('app.current_tenant_id', true)::uuid;
EXCEPTION
    WHEN OTHERS THEN
        RETURN NULL;
END;
$$;


ALTER FUNCTION public.get_current_tenant_id() OWNER TO "user";

--
-- Name: get_sync_status(); Type: FUNCTION; Schema: public; Owner: user
--

CREATE FUNCTION public.get_sync_status() RETURNS TABLE(total_pending integer, total_success integer, total_failed integer, last_sync_time timestamp with time zone)
    LANGUAGE plpgsql
    AS $$
		BEGIN
			RETURN QUERY 
			SELECT 
				(SELECT COUNT(*)::INTEGER FROM sync_monitoring WHERE sync_status = 'PENDING') as total_pending,
				(SELECT COUNT(*)::INTEGER FROM sync_monitoring WHERE sync_status = 'SUCCESS') as total_success,
				(SELECT COUNT(*)::INTEGER FROM sync_monitoring WHERE sync_status = 'FAILED') as total_failed,
				(SELECT MAX(synced_at) FROM sync_monitoring WHERE sync_status = 'SUCCESS') as last_sync_time;
		END;
		$$;


ALTER FUNCTION public.get_sync_status() OWNER TO "user";

--
-- Name: notify_organization_change(); Type: FUNCTION; Schema: public; Owner: user
--

CREATE FUNCTION public.notify_organization_change() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF TG_OP = 'DELETE' then
        PERFORM pg_notify('organization_change', json_build_object('operation', TG_OP, 'code', OLD.code)::text);
        RETURN OLD;
    ELSE
        PERFORM pg_notify('organization_change', json_build_object('operation', TG_OP, 'code', NEW.code)::text);
        RETURN NEW;
    END IF;
END;
$$;


ALTER FUNCTION public.notify_organization_change() OWNER TO "user";

--
-- Name: repair_organization_sync(); Type: FUNCTION; Schema: public; Owner: user
--

CREATE FUNCTION public.repair_organization_sync() RETURNS TABLE(repaired_count integer, failed_count integer, details text)
    LANGUAGE plpgsql
    AS $$
		DECLARE
			pending_count INTEGER;
			failed_sync_count INTEGER;
			repair_details TEXT := '';
		BEGIN
			-- 获取待同步数量
			SELECT COUNT(*) INTO pending_count 
			FROM sync_monitoring 
			WHERE sync_status = 'PENDING' 
			AND created_at > NOW() - INTERVAL '24 hours';
			
			-- 获取失败同步数量
			SELECT COUNT(*) INTO failed_sync_count 
			FROM sync_monitoring 
			WHERE sync_status = 'FAILED' 
			AND retry_count < 3;
			
			-- 标记超时的待同步记录为失败
			UPDATE sync_monitoring 
			SET sync_status = 'FAILED', 
				error_message = 'Sync timeout after 1 hour',
				updated_at = NOW()
			WHERE sync_status = 'PENDING' 
			AND created_at < NOW() - INTERVAL '1 hour';
			
			-- 重置失败次数不超过3次的记录为待同步
			UPDATE sync_monitoring 
			SET sync_status = 'PENDING', 
				retry_count = retry_count + 1,
				updated_at = NOW()
			WHERE sync_status = 'FAILED' 
			AND retry_count < 3
			AND created_at > NOW() - INTERVAL '24 hours';
			
			repair_details := format(
				'待同步: %s, 重试失败: %s, 修复时间: %s',
				pending_count,
				failed_sync_count,
				NOW()
			);
			
			RETURN QUERY SELECT pending_count, failed_sync_count, repair_details;
		END;
		$$;


ALTER FUNCTION public.repair_organization_sync() OWNER TO "user";

--
-- Name: set_tenant_context(uuid); Type: FUNCTION; Schema: public; Owner: user
--

CREATE FUNCTION public.set_tenant_context(tenant_uuid uuid) RETURNS void
    LANGUAGE plpgsql SECURITY DEFINER
    AS $$
BEGIN
    PERFORM set_config('app.current_tenant_id', tenant_uuid::text, true);
END;
$$;


ALTER FUNCTION public.set_tenant_context(tenant_uuid uuid) OWNER TO "user";

--
-- Name: update_employee_updated_at(); Type: FUNCTION; Schema: public; Owner: user
--

CREATE FUNCTION public.update_employee_updated_at() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_employee_updated_at() OWNER TO "user";

--
-- Name: update_position_updated_at(); Type: FUNCTION; Schema: public; Owner: user
--

CREATE FUNCTION public.update_position_updated_at() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_position_updated_at() OWNER TO "user";

--
-- Name: update_sync_monitoring_updated_at(); Type: FUNCTION; Schema: public; Owner: user
--

CREATE FUNCTION public.update_sync_monitoring_updated_at() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
		BEGIN
			NEW.updated_at = NOW();
			RETURN NEW;
		END;
		$$;


ALTER FUNCTION public.update_sync_monitoring_updated_at() OWNER TO "user";

--
-- Name: update_updated_at_column(); Type: FUNCTION; Schema: public; Owner: user
--

CREATE FUNCTION public.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_updated_at_column() OWNER TO "user";

--
-- Name: validate_position_history_temporal_consistency(); Type: FUNCTION; Schema: public; Owner: user
--

CREATE FUNCTION public.validate_position_history_temporal_consistency() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- Validate no overlapping periods for the same employee
    IF EXISTS (
        SELECT 1 FROM position_history 
        WHERE tenant_id = NEW.tenant_id 
          AND employee_id = NEW.employee_id
          AND id != COALESCE(NEW.id, '00000000-0000-0000-0000-000000000000'::UUID)
          AND effective_date <= COALESCE(NEW.end_date, 'infinity'::timestamp)
          AND COALESCE(end_date, 'infinity'::timestamp) > NEW.effective_date
    ) THEN
        RAISE EXCEPTION 'Temporal conflict: overlapping position periods for employee %', NEW.employee_id;
    END IF;
    
    -- Validate effective date is not in the far future (more than 2 years)
    IF NEW.effective_date > NOW() + INTERVAL '2 years' THEN
        RAISE EXCEPTION 'Effective date cannot be more than 2 years in the future';
    END IF;
    
    -- Validate retroactive flag is set correctly
    IF NEW.effective_date < NOW() - INTERVAL '1 day' AND NOT NEW.is_retroactive THEN
        NEW.is_retroactive = TRUE;
    END IF;
    
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.validate_position_history_temporal_consistency() OWNER TO "user";

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: employees; Type: TABLE; Schema: corehr; Owner: user
--

CREATE TABLE corehr.employees (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    employee_number character varying(50) NOT NULL,
    first_name character varying(100) NOT NULL,
    last_name character varying(100) NOT NULL,
    email character varying(255) NOT NULL,
    status character varying(20) DEFAULT 'active'::character varying,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    phone_number character varying(20),
    "position" character varying(100),
    department character varying(100),
    hire_date date DEFAULT CURRENT_DATE NOT NULL,
    manager_id uuid,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE corehr.employees OWNER TO "user";

--
-- Name: organizations; Type: TABLE; Schema: corehr; Owner: user
--

CREATE TABLE corehr.organizations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    name character varying(255) NOT NULL,
    code character varying(50) NOT NULL,
    parent_id uuid,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    level integer DEFAULT 1,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE corehr.organizations OWNER TO "user";

--
-- Name: positions; Type: TABLE; Schema: corehr; Owner: user
--

CREATE TABLE corehr.positions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tenant_id uuid NOT NULL,
    title character varying(255) NOT NULL,
    code character varying(50) NOT NULL,
    department_id uuid,
    level integer DEFAULT 1,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE corehr.positions OWNER TO "user";

--
-- Name: permissions; Type: TABLE; Schema: identity; Owner: user
--

CREATE TABLE identity.permissions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tenant_id uuid NOT NULL,
    resource character varying(100) NOT NULL,
    action character varying(100) NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE identity.permissions OWNER TO "user";

--
-- Name: role_permissions; Type: TABLE; Schema: identity; Owner: user
--

CREATE TABLE identity.role_permissions (
    role_id uuid NOT NULL,
    permission_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE identity.role_permissions OWNER TO "user";

--
-- Name: roles; Type: TABLE; Schema: identity; Owner: user
--

CREATE TABLE identity.roles (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tenant_id uuid NOT NULL,
    name character varying(100) NOT NULL,
    description text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE identity.roles OWNER TO "user";

--
-- Name: user_roles; Type: TABLE; Schema: identity; Owner: user
--

CREATE TABLE identity.user_roles (
    user_id uuid NOT NULL,
    role_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE identity.user_roles OWNER TO "user";

--
-- Name: users; Type: TABLE; Schema: identity; Owner: user
--

CREATE TABLE identity.users (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tenant_id uuid NOT NULL,
    employee_id uuid,
    username character varying(100) NOT NULL,
    email character varying(255) NOT NULL,
    password_hash character varying(255) NOT NULL,
    status character varying(20) DEFAULT 'active'::character varying,
    last_login_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE identity.users OWNER TO "user";

--
-- Name: conversations; Type: TABLE; Schema: intelligence; Owner: user
--

CREATE TABLE intelligence.conversations (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tenant_id uuid NOT NULL,
    user_id uuid,
    session_id character varying(255) NOT NULL,
    status character varying(20) DEFAULT 'active'::character varying,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE intelligence.conversations OWNER TO "user";

--
-- Name: messages; Type: TABLE; Schema: intelligence; Owner: user
--

CREATE TABLE intelligence.messages (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    conversation_id uuid,
    user_text text,
    ai_response text,
    intent character varying(100),
    entities jsonb,
    confidence numeric(3,2),
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE intelligence.messages OWNER TO "user";

--
-- Name: events; Type: TABLE; Schema: outbox; Owner: user
--

CREATE TABLE outbox.events (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    aggregate_id uuid NOT NULL,
    aggregate_type character varying(100) NOT NULL,
    event_type character varying(100) NOT NULL,
    event_version integer DEFAULT 1,
    payload jsonb NOT NULL,
    metadata jsonb,
    processed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE outbox.events OWNER TO "user";

--
-- Name: assignment_details; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.assignment_details (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    assignment_id uuid NOT NULL,
    pay_grade_id uuid,
    reporting_manager_id uuid,
    location_id uuid,
    cost_center character varying(50),
    effective_date date NOT NULL,
    reason text,
    approval_status character varying(50) DEFAULT 'PENDING'::character varying,
    approved_by uuid,
    approved_at timestamp with time zone,
    metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT assignment_details_approval_status_check CHECK (((approval_status)::text = ANY ((ARRAY['PENDING'::character varying, 'APPROVED'::character varying, 'REJECTED'::character varying])::text[])))
);


ALTER TABLE public.assignment_details OWNER TO "user";

--
-- Name: TABLE assignment_details; Type: COMMENT; Schema: public; Owner: user
--

COMMENT ON TABLE public.assignment_details IS '分配详情表 - 复杂业务信息';


--
-- Name: assignment_history; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.assignment_history (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    assignment_id uuid NOT NULL,
    change_type character varying(50) NOT NULL,
    old_values jsonb,
    new_values jsonb,
    changed_by uuid NOT NULL,
    change_reason text,
    effective_date date NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT assignment_history_change_type_check CHECK (((change_type)::text = ANY ((ARRAY['CREATED'::character varying, 'UPDATED'::character varying, 'ENDED'::character varying, 'TRANSFERRED'::character varying])::text[])))
);


ALTER TABLE public.assignment_history OWNER TO "user";

--
-- Name: TABLE assignment_history; Type: COMMENT; Schema: public; Owner: user
--

COMMENT ON TABLE public.assignment_history IS '分配历史表 - 审计跟踪';


--
-- Name: business_process_events; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.business_process_events (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    event_type character varying NOT NULL,
    entity_type character varying NOT NULL,
    entity_id uuid NOT NULL,
    effective_date timestamp with time zone NOT NULL,
    event_data jsonb NOT NULL,
    initiated_by uuid NOT NULL,
    correlation_id character varying,
    status character varying DEFAULT 'PENDING'::character varying NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


ALTER TABLE public.business_process_events OWNER TO "user";

--
-- Name: employee_business_id_seq; Type: SEQUENCE; Schema: public; Owner: user
--

CREATE SEQUENCE public.employee_business_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    MAXVALUE 99999
    CACHE 1;


ALTER SEQUENCE public.employee_business_id_seq OWNER TO "user";

--
-- Name: employee_code_seq; Type: SEQUENCE; Schema: public; Owner: user
--

CREATE SEQUENCE public.employee_code_seq
    START WITH 10000000
    INCREMENT BY 1
    NO MINVALUE
    MAXVALUE 99999999
    CACHE 1;


ALTER SEQUENCE public.employee_code_seq OWNER TO "user";

--
-- Name: position_history; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.position_history (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    employee_id uuid NOT NULL,
    position_title character varying(100) NOT NULL,
    department character varying(100) NOT NULL,
    job_level character varying(50),
    location character varying(100),
    employment_type character varying(20) NOT NULL,
    reports_to_employee_id uuid,
    effective_date timestamp with time zone NOT NULL,
    end_date timestamp with time zone,
    change_reason text,
    is_retroactive boolean DEFAULT false NOT NULL,
    created_by uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    min_salary numeric(15,2),
    max_salary numeric(15,2),
    currency character(3) DEFAULT 'CNY'::bpchar,
    CONSTRAINT position_history_employment_type_check CHECK (((employment_type)::text = ANY ((ARRAY['FULL_TIME'::character varying, 'PART_TIME'::character varying, 'CONTRACT'::character varying, 'INTERN'::character varying])::text[]))),
    CONSTRAINT valid_date_range CHECK (((end_date IS NULL) OR (end_date > effective_date))),
    CONSTRAINT valid_salary_range CHECK (((max_salary IS NULL) OR (min_salary IS NULL) OR (max_salary >= min_salary)))
);


ALTER TABLE public.position_history OWNER TO "user";

--
-- Name: TABLE position_history; Type: COMMENT; Schema: public; Owner: user
--

COMMENT ON TABLE public.position_history IS 'Employee position history with temporal tracking. Supports point-in-time queries and complete audit trail.';


--
-- Name: employee_department_summary; Type: VIEW; Schema: public; Owner: user
--

CREATE VIEW public.employee_department_summary AS
 SELECT department AS department_name,
    count(DISTINCT employee_id) AS employee_count,
    round(avg(((min_salary + max_salary) / 2.0))) AS avg_salary,
    (min(effective_date))::date AS earliest_hire_date,
    (max(effective_date))::date AS latest_hire_date,
    count(
        CASE
            WHEN ((job_level)::text = 'DIRECTOR'::text) THEN 1
            ELSE NULL::integer
        END) AS directors,
    count(
        CASE
            WHEN ((job_level)::text = 'SENIOR'::text) THEN 1
            ELSE NULL::integer
        END) AS senior_staff,
    count(
        CASE
            WHEN ((job_level)::text = 'REGULAR'::text) THEN 1
            ELSE NULL::integer
        END) AS regular_staff,
    count(
        CASE
            WHEN ((job_level)::text = 'JUNIOR'::text) THEN 1
            ELSE NULL::integer
        END) AS junior_staff,
    count(
        CASE
            WHEN ((job_level)::text = 'INTERN'::text) THEN 1
            ELSE NULL::integer
        END) AS interns
   FROM public.position_history ph
  WHERE ((tenant_id = '00000000-0000-0000-0000-000000000000'::uuid) AND (end_date IS NULL))
  GROUP BY department
  ORDER BY (count(DISTINCT employee_id)) DESC, (round(avg(((min_salary + max_salary) / 2.0)))) DESC;


ALTER VIEW public.employee_department_summary OWNER TO "user";

--
-- Name: employee_positions; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.employee_positions (
    id integer NOT NULL,
    employee_code character varying(8) NOT NULL,
    position_code character varying(7) NOT NULL,
    assignment_type character varying(20) DEFAULT 'PRIMARY'::character varying NOT NULL,
    status character varying(20) DEFAULT 'ACTIVE'::character varying NOT NULL,
    start_date date NOT NULL,
    end_date date,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT employee_positions_assignment_type_check CHECK (((assignment_type)::text = ANY ((ARRAY['PRIMARY'::character varying, 'SECONDARY'::character varying, 'TEMPORARY'::character varying, 'ACTING'::character varying])::text[]))),
    CONSTRAINT employee_positions_check CHECK (((end_date IS NULL) OR (end_date >= start_date))),
    CONSTRAINT employee_positions_status_check CHECK (((status)::text = ANY ((ARRAY['ACTIVE'::character varying, 'INACTIVE'::character varying, 'PENDING'::character varying, 'ENDED'::character varying])::text[])))
);


ALTER TABLE public.employee_positions OWNER TO "user";

--
-- Name: employee_positions_backup_20250805; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.employee_positions_backup_20250805 (
    id uuid,
    employee_id uuid,
    position_id uuid,
    tenant_id uuid,
    start_date date,
    end_date date,
    is_primary boolean,
    created_at timestamp with time zone
);


ALTER TABLE public.employee_positions_backup_20250805 OWNER TO "user";

--
-- Name: employee_positions_id_seq; Type: SEQUENCE; Schema: public; Owner: user
--

CREATE SEQUENCE public.employee_positions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.employee_positions_id_seq OWNER TO "user";

--
-- Name: employee_positions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: user
--

ALTER SEQUENCE public.employee_positions_id_seq OWNED BY public.employee_positions.id;


--
-- Name: employees; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.employees (
    code character varying(8) NOT NULL,
    organization_code character varying(7) NOT NULL,
    primary_position_code character varying(7),
    employee_type character varying(20) NOT NULL,
    employment_status character varying(20) DEFAULT 'ACTIVE'::character varying NOT NULL,
    first_name character varying(100) NOT NULL,
    last_name character varying(100) NOT NULL,
    email character varying(255) NOT NULL,
    personal_email character varying(255),
    phone_number character varying(20),
    hire_date date NOT NULL,
    termination_date date,
    personal_info jsonb,
    employee_details jsonb,
    tenant_id uuid NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT employees_code_check CHECK ((((code)::text ~ '^[0-9]{8}$'::text) AND (((code)::integer >= 10000000) AND ((code)::integer <= 99999999)))),
    CONSTRAINT employees_employee_type_check CHECK (((employee_type)::text = ANY ((ARRAY['FULL_TIME'::character varying, 'PART_TIME'::character varying, 'CONTRACTOR'::character varying, 'INTERN'::character varying])::text[]))),
    CONSTRAINT employees_employment_status_check CHECK (((employment_status)::text = ANY ((ARRAY['ACTIVE'::character varying, 'TERMINATED'::character varying, 'ON_LEAVE'::character varying, 'PENDING_START'::character varying])::text[])))
);


ALTER TABLE public.employees OWNER TO "user";

--
-- Name: employees_backup; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.employees_backup (
    id character varying,
    name character varying,
    email character varying,
    "position" character varying,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    uuid_id uuid
);


ALTER TABLE public.employees_backup OWNER TO "user";

--
-- Name: metacontract_editor_projects; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.metacontract_editor_projects (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    content text DEFAULT ''::text NOT NULL,
    version character varying(50) DEFAULT '1.0.0'::character varying NOT NULL,
    status character varying(20) DEFAULT 'draft'::character varying NOT NULL,
    tenant_id uuid NOT NULL,
    created_by uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    last_compiled timestamp with time zone,
    compile_error text,
    CONSTRAINT metacontract_editor_projects_status_check CHECK (((status)::text = ANY ((ARRAY['draft'::character varying, 'compiling'::character varying, 'valid'::character varying, 'error'::character varying, 'published'::character varying])::text[])))
);


ALTER TABLE public.metacontract_editor_projects OWNER TO "user";

--
-- Name: metacontract_editor_sessions; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.metacontract_editor_sessions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    user_id uuid NOT NULL,
    started_at timestamp with time zone DEFAULT now() NOT NULL,
    last_seen timestamp with time zone DEFAULT now() NOT NULL,
    active boolean DEFAULT true NOT NULL
);


ALTER TABLE public.metacontract_editor_sessions OWNER TO "user";

--
-- Name: metacontract_editor_settings; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.metacontract_editor_settings (
    user_id uuid NOT NULL,
    theme character varying(50) DEFAULT 'light'::character varying NOT NULL,
    font_size integer DEFAULT 14 NOT NULL,
    auto_save boolean DEFAULT true NOT NULL,
    auto_compile boolean DEFAULT false NOT NULL,
    key_bindings character varying(50) DEFAULT 'default'::character varying NOT NULL,
    settings jsonb DEFAULT '{}'::jsonb,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.metacontract_editor_settings OWNER TO "user";

--
-- Name: metacontract_editor_templates; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.metacontract_editor_templates (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    category character varying(50) DEFAULT 'general'::character varying NOT NULL,
    content text NOT NULL,
    tags text[] DEFAULT '{}'::text[],
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.metacontract_editor_templates OWNER TO "user";

--
-- Name: org_business_id_seq; Type: SEQUENCE; Schema: public; Owner: user
--

CREATE SEQUENCE public.org_business_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    MAXVALUE 899999
    CACHE 1;


ALTER SEQUENCE public.org_business_id_seq OWNER TO "user";

--
-- Name: org_unit_code_seq; Type: SEQUENCE; Schema: public; Owner: user
--

CREATE SEQUENCE public.org_unit_code_seq
    START WITH 1000000
    INCREMENT BY 1
    NO MINVALUE
    MAXVALUE 9999999
    CACHE 1;


ALTER SEQUENCE public.org_unit_code_seq OWNER TO "user";

--
-- Name: organization_units; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.organization_units (
    code character varying(10) NOT NULL,
    parent_code character varying(10),
    tenant_id uuid NOT NULL,
    name character varying(255) NOT NULL,
    unit_type character varying(50) NOT NULL,
    status character varying(20) DEFAULT 'ACTIVE'::character varying,
    level integer DEFAULT 1 NOT NULL,
    path character varying(1000),
    sort_order integer DEFAULT 0,
    description text,
    profile jsonb DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT organization_units_status_check CHECK (((status)::text = ANY ((ARRAY['ACTIVE'::character varying, 'INACTIVE'::character varying, 'PLANNED'::character varying])::text[]))),
    CONSTRAINT organization_units_unit_type_check CHECK (((unit_type)::text = ANY ((ARRAY['DEPARTMENT'::character varying, 'COST_CENTER'::character varying, 'COMPANY'::character varying, 'PROJECT_TEAM'::character varying])::text[])))
);


ALTER TABLE public.organization_units OWNER TO "user";

--
-- Name: outbox_events; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.outbox_events (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    event_type character varying(100) NOT NULL,
    aggregate_id uuid NOT NULL,
    event_data jsonb NOT NULL,
    status character varying(50) DEFAULT 'PENDING'::character varying NOT NULL,
    attempt_count integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    processed_at timestamp with time zone,
    error_message text,
    CONSTRAINT outbox_events_status_check CHECK (((status)::text = ANY ((ARRAY['PENDING'::character varying, 'PROCESSED'::character varying, 'FAILED'::character varying])::text[])))
);


ALTER TABLE public.outbox_events OWNER TO "user";

--
-- Name: person; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.person (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid DEFAULT '00000000-0000-0000-0000-000000000000'::uuid NOT NULL,
    name character varying(255) NOT NULL,
    email character varying(255) NOT NULL,
    employee_id character varying(100) NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.person OWNER TO "user";

--
-- Name: position_assignments; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.position_assignments (
    id integer NOT NULL,
    position_code character varying(7) NOT NULL,
    employee_code character varying(8),
    assignment_date date NOT NULL,
    end_date date,
    assignment_reason character varying(100),
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.position_assignments OWNER TO "user";

--
-- Name: position_assignments_backup_20250805; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.position_assignments_backup_20250805 (
    id uuid,
    tenant_id uuid,
    position_id uuid,
    employee_id uuid,
    start_date date,
    end_date date,
    is_current boolean,
    fte numeric(3,2),
    assignment_type character varying(50),
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.position_assignments_backup_20250805 OWNER TO "user";

--
-- Name: position_assignments_id_seq; Type: SEQUENCE; Schema: public; Owner: user
--

CREATE SEQUENCE public.position_assignments_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.position_assignments_id_seq OWNER TO "user";

--
-- Name: position_assignments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: user
--

ALTER SEQUENCE public.position_assignments_id_seq OWNED BY public.position_assignments.id;


--
-- Name: position_attribute_histories; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.position_attribute_histories (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    position_type character varying NOT NULL,
    job_profile_id uuid NOT NULL,
    department_id uuid NOT NULL,
    manager_position_id uuid,
    status character varying NOT NULL,
    budgeted_fte double precision NOT NULL,
    details jsonb,
    effective_date timestamp with time zone NOT NULL,
    end_date timestamp with time zone,
    change_reason character varying,
    changed_by uuid NOT NULL,
    change_type character varying,
    source_event_id uuid,
    created_at timestamp with time zone NOT NULL,
    position_id uuid NOT NULL
);


ALTER TABLE public.position_attribute_histories OWNER TO "user";

--
-- Name: position_business_id_seq; Type: SEQUENCE; Schema: public; Owner: user
--

CREATE SEQUENCE public.position_business_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    MAXVALUE 8999999
    CACHE 1;


ALTER SEQUENCE public.position_business_id_seq OWNER TO "user";

--
-- Name: position_code_sequence; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.position_code_sequence (
    tenant_id uuid NOT NULL,
    last_code integer DEFAULT 1000000 NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.position_code_sequence OWNER TO "user";

--
-- Name: position_histories; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.position_histories (
    id character varying NOT NULL,
    employee_id character varying NOT NULL,
    organization_id character varying NOT NULL,
    position_title character varying NOT NULL,
    department character varying NOT NULL,
    effective_date timestamp with time zone NOT NULL,
    end_date timestamp with time zone,
    is_active boolean DEFAULT true NOT NULL,
    is_retroactive boolean DEFAULT false NOT NULL,
    salary_data jsonb,
    change_reason character varying,
    approval_status character varying DEFAULT 'approved'::character varying NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


ALTER TABLE public.position_histories OWNER TO "user";

--
-- Name: position_occupancy_histories; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.position_occupancy_histories (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    employee_id uuid NOT NULL,
    start_date timestamp with time zone NOT NULL,
    end_date timestamp with time zone,
    is_active boolean DEFAULT true NOT NULL,
    assignment_type character varying DEFAULT 'REGULAR'::character varying NOT NULL,
    assignment_reason character varying,
    fte_percentage double precision DEFAULT 1 NOT NULL,
    work_arrangement character varying,
    approved_by uuid,
    approval_date timestamp with time zone,
    approval_reference character varying,
    compensation_data jsonb,
    performance_review_cycle character varying,
    source_event_id uuid,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    position_id uuid NOT NULL
);


ALTER TABLE public.position_occupancy_histories OWNER TO "user";

--
-- Name: positions; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.positions (
    code character varying(7) NOT NULL,
    organization_code character varying(7) NOT NULL,
    manager_position_code character varying(7),
    position_type character varying(50) NOT NULL,
    job_profile_id uuid NOT NULL,
    status character varying(20) DEFAULT 'OPEN'::character varying NOT NULL,
    budgeted_fte numeric(3,2) DEFAULT 1.00 NOT NULL,
    details jsonb,
    tenant_id uuid NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT positions_budgeted_fte_check CHECK (((budgeted_fte > (0)::numeric) AND (budgeted_fte <= 5.00))),
    CONSTRAINT positions_code_check CHECK ((((code)::text ~ '^[0-9]{7}$'::text) AND (((code)::integer >= 1000000) AND ((code)::integer <= 9999999)))),
    CONSTRAINT positions_position_type_check CHECK (((position_type)::text = ANY ((ARRAY['FULL_TIME'::character varying, 'PART_TIME'::character varying, 'CONTINGENT_WORKER'::character varying, 'INTERN'::character varying])::text[]))),
    CONSTRAINT positions_status_check CHECK (((status)::text = ANY ((ARRAY['OPEN'::character varying, 'FILLED'::character varying, 'FROZEN'::character varying, 'PENDING_ELIMINATION'::character varying])::text[])))
);


ALTER TABLE public.positions OWNER TO "user";

--
-- Name: positions_backup; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.positions_backup (
    id uuid,
    tenant_id uuid,
    title character varying(100),
    department character varying(100),
    level character varying(50),
    description text,
    requirements text,
    is_active boolean,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.positions_backup OWNER TO "user";

--
-- Name: positions_backup_20250805; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.positions_backup_20250805 (
    id uuid,
    tenant_id uuid,
    position_type character varying(50),
    job_profile_id uuid,
    department_id uuid,
    manager_position_id uuid,
    status character varying(50),
    budgeted_fte numeric(3,2),
    details jsonb,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    business_id character varying(7)
);


ALTER TABLE public.positions_backup_20250805 OWNER TO "user";

--
-- Name: sync_monitoring; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.sync_monitoring (
    id integer NOT NULL,
    operation_type character varying(20) NOT NULL,
    entity_id uuid NOT NULL,
    entity_data jsonb NOT NULL,
    sync_status character varying(20) DEFAULT 'PENDING'::character varying,
    error_message text,
    retry_count integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    synced_at timestamp with time zone
);


ALTER TABLE public.sync_monitoring OWNER TO "user";

--
-- Name: sync_monitoring_id_seq; Type: SEQUENCE; Schema: public; Owner: user
--

CREATE SEQUENCE public.sync_monitoring_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.sync_monitoring_id_seq OWNER TO "user";

--
-- Name: sync_monitoring_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: user
--

ALTER SEQUENCE public.sync_monitoring_id_seq OWNED BY public.sync_monitoring.id;


--
-- Name: workflow_instances; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.workflow_instances (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    workflow_type character varying NOT NULL,
    current_state character varying NOT NULL,
    state_history jsonb NOT NULL,
    context jsonb NOT NULL,
    initiated_by uuid NOT NULL,
    correlation_id character varying NOT NULL,
    started_at timestamp with time zone NOT NULL,
    completed_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


ALTER TABLE public.workflow_instances OWNER TO "user";

--
-- Name: workflow_steps; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.workflow_steps (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    step_name character varying NOT NULL,
    step_type character varying NOT NULL,
    status character varying DEFAULT 'PENDING'::character varying NOT NULL,
    assigned_to uuid,
    input_data jsonb,
    output_data jsonb,
    due_date timestamp with time zone,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    workflow_instance_id uuid NOT NULL
);


ALTER TABLE public.workflow_steps OWNER TO "user";

--
-- Name: tenant_configs; Type: TABLE; Schema: tenancy; Owner: user
--

CREATE TABLE tenancy.tenant_configs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tenant_id uuid,
    config_key character varying(100) NOT NULL,
    config_value text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE tenancy.tenant_configs OWNER TO "user";

--
-- Name: tenants; Type: TABLE; Schema: tenancy; Owner: user
--

CREATE TABLE tenancy.tenants (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name character varying(255) NOT NULL,
    domain character varying(255),
    status character varying(20) DEFAULT 'active'::character varying,
    subscription_plan character varying(50) DEFAULT 'basic'::character varying,
    max_users integer DEFAULT 10,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE tenancy.tenants OWNER TO "user";

--
-- Name: employee_positions id; Type: DEFAULT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.employee_positions ALTER COLUMN id SET DEFAULT nextval('public.employee_positions_id_seq'::regclass);


--
-- Name: position_assignments id; Type: DEFAULT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.position_assignments ALTER COLUMN id SET DEFAULT nextval('public.position_assignments_id_seq'::regclass);


--
-- Name: sync_monitoring id; Type: DEFAULT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.sync_monitoring ALTER COLUMN id SET DEFAULT nextval('public.sync_monitoring_id_seq'::regclass);


--
-- Data for Name: employees; Type: TABLE DATA; Schema: corehr; Owner: user
--

COPY corehr.employees (id, tenant_id, employee_number, first_name, last_name, email, status, created_at, phone_number, "position", department, hire_date, manager_id, updated_at) FROM stdin;
6bc3fa3a-a761-4df3-957c-11bccfd47fdc	62c5f693-95b0-4d0b-bf1f-5f3d86e296fb	FINAL-TEST-1753938539	最终	测试	final-test-1753938539@example.com	active	2025-07-31 05:08:59.627038+00	13800138000	\N	\N	2025-07-31	\N	2025-07-31 05:08:59.627038+00
6e5009c2-d8c2-4ad4-8f9b-8909c6462418	00000000-0000-0000-0000-000000000000	EMP001	张	伟强	zhang.weiqiang@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	CTO	技术部	2020-01-01	\N	2025-08-01 00:58:54.771142+00
a7a297d4-6714-4232-9999-6fadaacd8157	00000000-0000-0000-0000-000000000000	EMP003	王	建国	wang.jianguo@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	VP Engineering	技术部	2020-06-01	\N	2025-08-01 00:58:54.771142+00
ff4f9ae9-2445-464c-b653-8f4eeed61e6a	00000000-0000-0000-0000-000000000000	EMP004	刘	美丽	liu.meili@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	VP Sales	销售部	2020-06-01	\N	2025-08-01 00:58:54.771142+00
865d0ea2-177a-4e67-b067-4f309fb74fad	00000000-0000-0000-0000-000000000000	EMP005	陈	志华	chen.zhihua@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	CFO	财务部	2020-01-01	\N	2025-08-01 00:58:54.771142+00
55b4c36a-7062-41f5-9871-c740409b681f	00000000-0000-0000-0000-000000000000	EMP006	赵	晓明	zhao.xiaoming@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	前端开发总监	前端开发部	2020-06-01	\N	2025-08-01 00:58:54.771142+00
c98f8a05-e17c-46a2-a957-0794589dab20	00000000-0000-0000-0000-000000000000	EMP007	吴	小丽	wu.xiaoli@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	高级前端工程师	前端开发部	2021-03-01	\N	2025-08-01 00:58:54.771142+00
35c536df-53c2-4148-8d7e-d6372cb5251c	00000000-0000-0000-0000-000000000000	EMP008	周	大伟	zhou.dawei@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	高级前端工程师	前端开发部	2021-08-01	\N	2025-08-01 00:58:54.771142+00
14738b73-302a-4906-881b-b1a53ef5698d	00000000-0000-0000-0000-000000000000	EMP009	郑	晓红	zheng.xiaohong@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	前端工程师	前端开发部	2022-01-01	\N	2025-08-01 00:58:54.771142+00
67bb77b9-d3c9-4900-96f1-a29da9bac0b9	00000000-0000-0000-0000-000000000000	EMP010	孙	志强	sun.zhiqiang@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	前端工程师	前端开发部	2022-06-01	\N	2025-08-01 00:58:54.771142+00
6f0685f6-1e9b-44d7-a60a-85428124b0c6	00000000-0000-0000-0000-000000000000	EMP011	朱	小芳	zhu.xiaofang@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	前端工程师	前端开发部	2023-01-01	\N	2025-08-01 00:58:54.771142+00
1b776ce5-a431-43a2-b5b7-8a3424117f85	00000000-0000-0000-0000-000000000000	EMP012	胡	建华	hu.jianhua@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	初级前端工程师	前端开发部	2023-09-01	\N	2025-08-01 00:58:54.771142+00
32b70e5f-f24b-488c-b44b-a4a2eefdaf0d	00000000-0000-0000-0000-000000000000	EMP013	高	小雨	gao.xiaoyu@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	前端实习生	前端开发部	2024-09-01	\N	2025-08-01 00:58:54.771142+00
386121ed-e685-4df0-a892-5bdb4411b3ff	00000000-0000-0000-0000-000000000000	EMP014	许	文博	xu.wenbo@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	后端开发总监	后端开发部	2020-06-01	\N	2025-08-01 00:58:54.771142+00
1427b312-e5ef-441b-a9a3-1dabb49faa0b	00000000-0000-0000-0000-000000000000	EMP015	何	志远	he.zhiyuan@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	架构师	后端开发部	2021-01-01	\N	2025-08-01 00:58:54.771142+00
e25024c3-916c-456c-91e5-d910274a5729	00000000-0000-0000-0000-000000000000	EMP016	韩	小强	han.xiaoqiang@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	高级后端工程师	后端开发部	2021-03-01	\N	2025-08-01 00:58:54.771142+00
1151b4f5-9538-4ff1-a1f0-9c9909791ce5	00000000-0000-0000-0000-000000000000	EMP017	冯	大明	feng.daming@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	高级后端工程师	后端开发部	2021-08-01	\N	2025-08-01 00:58:54.771142+00
65c82c35-4418-4f8a-af26-39ab76bee2fa	00000000-0000-0000-0000-000000000000	EMP018	邓	晓丽	deng.xiaoli@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	后端工程师	后端开发部	2022-01-01	\N	2025-08-01 00:58:54.771142+00
bd6c8433-d984-443c-98f6-ce6fdbb14085	00000000-0000-0000-0000-000000000000	EMP019	曹	志华	cao.zhihua@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	后端工程师	后端开发部	2022-06-01	\N	2025-08-01 00:58:54.771142+00
948a02dd-3cd5-4c46-840d-9338266cc9d0	00000000-0000-0000-0000-000000000000	EMP020	彭	小芳	peng.xiaofang@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	后端工程师	后端开发部	2023-01-01	\N	2025-08-01 00:58:54.771142+00
3fdf43ea-5a76-485b-b8ff-93a4f47abe85	00000000-0000-0000-0000-000000000000	EMP021	吕	建国	lv.jianguo@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	后端工程师	后端开发部	2023-06-01	\N	2025-08-01 00:58:54.771142+00
107a84cd-4b6c-41d7-ac70-1bc20089b742	00000000-0000-0000-0000-000000000000	EMP022	苏	小雨	su.xiaoyu@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	初级后端工程师	后端开发部	2023-09-01	\N	2025-08-01 00:58:54.771142+00
a52aee3f-25ea-46ff-aa52-e67abdef9048	00000000-0000-0000-0000-000000000000	EMP023	丁	志强	ding.zhiqiang@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	后端实习生	后端开发部	2024-09-01	\N	2025-08-01 00:58:54.771142+00
eab17a44-4b9a-4f18-9712-82504ff2d2f2	00000000-0000-0000-0000-000000000000	EMP024	任	小明	ren.xiaoming@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	移动开发总监	移动开发部	2020-06-01	\N	2025-08-01 00:58:54.771142+00
834dbbe4-b6be-4796-8861-bc67edf842e2	00000000-0000-0000-0000-000000000000	EMP025	姜	大伟	jiang.dawei@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	高级iOS工程师	移动开发部	2021-08-01	\N	2025-08-01 00:58:54.771142+00
c01fa714-1714-45c3-a928-e2c7c8c63fe6	00000000-0000-0000-0000-000000000000	EMP026	谢	晓红	xie.xiaohong@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	高级Android工程师	移动开发部	2022-01-01	\N	2025-08-01 00:58:54.771142+00
a8879c82-b6d1-4f0f-a6d3-7253ec5b5b9d	00000000-0000-0000-0000-000000000000	EMP027	沈	志华	shen.zhihua@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	React Native工程师	移动开发部	2022-06-01	\N	2025-08-01 00:58:54.771142+00
59514a7c-67c4-4040-9eec-0758d1c9aa2a	00000000-0000-0000-0000-000000000000	EMP028	韦	小芳	wei.xiaofang@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	Flutter工程师	移动开发部	2023-01-01	\N	2025-08-01 00:58:54.771142+00
596516df-0077-4b90-852d-c7195c39bd86	00000000-0000-0000-0000-000000000000	EMP029	段	建华	duan.jianhua@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	移动开发实习生	移动开发部	2024-09-01	\N	2025-08-01 00:58:54.771142+00
c92a9138-2c00-481a-b5a4-008a76bb1767	00000000-0000-0000-0000-000000000000	EMP030	毛	建华	mao.jianhua@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	数据工程总监	数据工程部	2020-06-01	\N	2025-08-01 00:58:54.771142+00
3a3dda8d-b62d-4edc-93ad-58019d42b5b4	00000000-0000-0000-0000-000000000000	EMP031	薛	小雨	xue.xiaoyu@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	数据架构师	数据工程部	2021-01-01	\N	2025-08-01 00:58:54.771142+00
a614f411-b7a9-4cdd-9e01-d007b05c57c4	00000000-0000-0000-0000-000000000000	EMP032	白	志强	bai.zhiqiang@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	大数据工程师	数据工程部	2021-08-01	\N	2025-08-01 00:58:54.771142+00
2fd5e286-2bbb-4236-8823-95e238f17da8	00000000-0000-0000-0000-000000000000	EMP033	崔	小明	cui.xiaoming@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	数据分析师	数据工程部	2022-01-01	\N	2025-08-01 00:58:54.771142+00
c874edee-b9e1-4057-92ed-8241720ff052	00000000-0000-0000-0000-000000000000	EMP034	田	大伟	tian.dawei@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	机器学习工程师	数据工程部	2022-06-01	\N	2025-08-01 00:58:54.771142+00
2ecb0bf5-a018-4b9e-84ad-b66ee2dd875b	00000000-0000-0000-0000-000000000000	EMP035	侯	伟光	hou.weiguang@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	DevOps总监	DevOps部	2020-06-01	\N	2025-08-01 00:58:54.771142+00
44831b41-5ae1-46ad-a260-a67ccc81a0eb	00000000-0000-0000-0000-000000000000	EMP036	邹	晓红	zou.xiaohong@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	高级DevOps工程师	DevOps部	2021-03-01	\N	2025-08-01 00:58:54.771142+00
adff6e5a-e1d5-41a7-9ad0-f40bce511c75	00000000-0000-0000-0000-000000000000	EMP037	石	志华	shi.zhihua@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	DevOps工程师	DevOps部	2021-08-01	\N	2025-08-01 00:58:54.771142+00
af7eb0d9-15b8-428c-b603-10422898a30a	00000000-0000-0000-0000-000000000000	EMP038	龙	小芳	long.xiaofang@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	云平台工程师	DevOps部	2022-01-01	\N	2025-08-01 00:58:54.771142+00
a1d7b779-62af-4515-8fc3-1ffb48d500cc	00000000-0000-0000-0000-000000000000	EMP039	谭	建平	tan.jianping@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	测试总监	测试部	2020-06-01	\N	2025-08-01 00:58:54.771142+00
eee08f7b-833c-437e-84f4-4e8ee0a25223	00000000-0000-0000-0000-000000000000	EMP002	李	芳芳	li.fangfang.test@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	CPO	产品部	2020-01-01	\N	2025-08-05 01:47:59.425713+00
e1cc151d-969d-43a9-a42c-75b304a5fd20	00000000-0000-0000-0000-000000000000	EMP040	黎	小雨	li.xiaoyu@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	高级测试工程师	测试部	2021-08-01	\N	2025-08-01 00:58:54.771142+00
ca53e84d-ac12-421b-a708-f22c5f2ebe3c	00000000-0000-0000-0000-000000000000	EMP041	严	志强	yan.zhiqiang@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	自动化测试工程师	测试部	2022-01-01	\N	2025-08-01 00:58:54.771142+00
fbd3c9fa-7dd6-4d5f-bcad-e58ad342f7a7	00000000-0000-0000-0000-000000000000	EMP042	文	小明	wen.xiaoming@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	性能测试工程师	测试部	2022-06-01	\N	2025-08-01 00:58:54.771142+00
868107f4-7741-4da0-813f-65a039480b5a	00000000-0000-0000-0000-000000000000	EMP043	尹	大伟	yin.dawei@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	测试工程师	测试部	2023-01-01	\N	2025-08-01 00:58:54.771142+00
59802120-59b6-42f9-9aa5-0a70b96defef	00000000-0000-0000-0000-000000000000	EMP044	卢	晓红	lu.xiaohong@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	测试实习生	测试部	2024-09-01	\N	2025-08-01 00:58:54.771142+00
ff729829-2ef4-46bf-ac1b-56d3ed120087	00000000-0000-0000-0000-000000000000	EMP045	常	晓东	chang.xiaodong@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	产品总监	产品部	2020-06-01	\N	2025-08-01 00:58:54.771142+00
ffc0f730-7d77-42a8-b3e6-6678c3eb3d24	00000000-0000-0000-0000-000000000000	EMP046	马	志华	ma.zhihua@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	高级产品经理	产品部	2021-03-01	\N	2025-08-01 00:58:54.771142+00
43f7367e-7af8-41e5-925a-fae960a82190	00000000-0000-0000-0000-000000000000	EMP047	方	小芳	fang.xiaofang@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	产品经理	产品部	2021-08-01	\N	2025-08-01 00:58:54.771142+00
d8c6ddca-2cce-4b01-b73a-49812b4fd5b7	00000000-0000-0000-0000-000000000000	EMP048	夏	志华	xia.zhihua@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	人力资源经理	人力资源部	2021-08-01	\N	2025-08-01 00:58:54.771142+00
2a7e60fc-da35-4a8f-ab4a-76cc745e3d0e	00000000-0000-0000-0000-000000000000	EMP049	华	小明	hua.xiaoming@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	UX设计师	UX设计部	2022-06-01	\N	2025-08-01 00:58:54.771142+00
80269685-f20a-42c7-b172-9222eaf59fd2	00000000-0000-0000-0000-000000000000	EMP050	金	晓丽	jin.xiaoli@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	销售经理	销售部	2020-06-01	\N	2025-08-01 00:58:54.771142+00
066190de-cfb2-41ad-be62-7e50ec7bac33	550e8400-e29b-41d4-a716-446655440000	EMP000150	吴	敏	test_employee_149@company.com	active	2025-08-05 03:58:41.252838+00	\N	\N	财务部-8	2025-03-08	\N	2025-08-05 04:13:27.746228+00
\.


--
-- Data for Name: organizations; Type: TABLE DATA; Schema: corehr; Owner: user
--

COPY corehr.organizations (id, tenant_id, name, code, parent_id, created_at, level, updated_at) FROM stdin;
\.


--
-- Data for Name: positions; Type: TABLE DATA; Schema: corehr; Owner: user
--

COPY corehr.positions (id, tenant_id, title, code, department_id, level, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: permissions; Type: TABLE DATA; Schema: identity; Owner: user
--

COPY identity.permissions (id, tenant_id, resource, action, created_at) FROM stdin;
\.


--
-- Data for Name: role_permissions; Type: TABLE DATA; Schema: identity; Owner: user
--

COPY identity.role_permissions (role_id, permission_id, created_at) FROM stdin;
\.


--
-- Data for Name: roles; Type: TABLE DATA; Schema: identity; Owner: user
--

COPY identity.roles (id, tenant_id, name, description, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: user_roles; Type: TABLE DATA; Schema: identity; Owner: user
--

COPY identity.user_roles (user_id, role_id, created_at) FROM stdin;
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: identity; Owner: user
--

COPY identity.users (id, tenant_id, employee_id, username, email, password_hash, status, last_login_at, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: conversations; Type: TABLE DATA; Schema: intelligence; Owner: user
--

COPY intelligence.conversations (id, tenant_id, user_id, session_id, status, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: messages; Type: TABLE DATA; Schema: intelligence; Owner: user
--

COPY intelligence.messages (id, conversation_id, user_text, ai_response, intent, entities, confidence, created_at) FROM stdin;
\.


--
-- Data for Name: events; Type: TABLE DATA; Schema: outbox; Owner: user
--

COPY outbox.events (id, aggregate_id, aggregate_type, event_type, event_version, payload, metadata, processed_at, created_at) FROM stdin;
a8b4b00a-97f7-4b05-a862-093df8a8f098	6bc3fa3a-a761-4df3-957c-11bccfd47fdc	Employee	employee.created	1	{"created_at": "2025-07-31T13:08:59+08:00", "employee_id": "6bc3fa3a-a761-4df3-957c-11bccfd47fdc", "employee_data": {"email": "final-test-1753938539@example.com", "position": null, "hire_date": "2025-07-31T13:08:59+08:00", "last_name": "测试", "department": null, "first_name": "最终", "employee_number": "FINAL-TEST-1753938539"}}	\N	\N	2025-07-31 05:08:59.630229+00
\.


--
-- Data for Name: assignment_details; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.assignment_details (id, assignment_id, pay_grade_id, reporting_manager_id, location_id, cost_center, effective_date, reason, approval_status, approved_by, approved_at, metadata, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: assignment_history; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.assignment_history (id, assignment_id, change_type, old_values, new_values, changed_by, change_reason, effective_date, created_at) FROM stdin;
\.


--
-- Data for Name: business_process_events; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.business_process_events (id, tenant_id, event_type, entity_type, entity_id, effective_date, event_data, initiated_by, correlation_id, status, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: employee_positions; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.employee_positions (id, employee_code, position_code, assignment_type, status, start_date, end_date, created_at, updated_at) FROM stdin;
1	10000000	1000001	PRIMARY	ACTIVE	2024-01-15	\N	2025-08-05 13:20:29.27323	2025-08-05 13:20:29.27323
2	10000001	1000002	PRIMARY	ACTIVE	2024-02-01	\N	2025-08-05 13:20:29.27323	2025-08-05 13:20:29.27323
3	10000002	1000003	PRIMARY	ACTIVE	2024-03-10	\N	2025-08-05 13:20:29.27323	2025-08-05 13:20:29.27323
4	10000003	1000004	PRIMARY	ACTIVE	2024-04-05	\N	2025-08-05 13:20:29.27323	2025-08-05 13:20:29.27323
5	10000004	1000005	PRIMARY	ACTIVE	2024-05-20	\N	2025-08-05 13:20:29.27323	2025-08-05 13:20:29.27323
\.


--
-- Data for Name: employee_positions_backup_20250805; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.employee_positions_backup_20250805 (id, employee_id, position_id, tenant_id, start_date, end_date, is_primary, created_at) FROM stdin;
\.


--
-- Data for Name: employees; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.employees (code, organization_code, primary_position_code, employee_type, employment_status, first_name, last_name, email, personal_email, phone_number, hire_date, termination_date, personal_info, employee_details, tenant_id, created_at, updated_at) FROM stdin;
10000000	1000000	1000001	FULL_TIME	ACTIVE	张	伟	zhang.wei@company.com	\N	\N	2024-01-15	\N	{"age": 30, "gender": "M", "address": "北京市朝阳区"}	{"level": "P6", "title": "高级软件工程师", "salary": 28000}	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	2025-08-05 13:20:29.27323	2025-08-05 13:20:29.27323
10000001	1000000	1000002	FULL_TIME	ACTIVE	李	娜	li.na@company.com	\N	\N	2024-02-01	\N	{"age": 28, "gender": "F", "address": "上海市浦东新区"}	{"level": "P7", "title": "软件架构师", "salary": 35000}	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	2025-08-05 13:20:29.27323	2025-08-05 13:20:29.27323
10000002	1000001	1000003	FULL_TIME	ACTIVE	王	强	wang.qiang@company.com	\N	\N	2024-03-10	\N	{"age": 32, "gender": "M", "address": "深圳市南山区"}	{"level": "P6", "title": "产品经理", "salary": 30000}	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	2025-08-05 13:20:29.27323	2025-08-05 13:20:29.27323
10000003	1000001	1000004	PART_TIME	ACTIVE	刘	敏	liu.min@company.com	\N	\N	2024-04-05	\N	{"age": 26, "gender": "F", "address": "广州市天河区"}	{"level": "P4", "title": "UI设计师", "hourly_rate": 200}	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	2025-08-05 13:20:29.27323	2025-08-05 13:20:29.27323
10000004	1000002	1000005	INTERN	ACTIVE	陈	阳	chen.yang@company.com	\N	\N	2024-05-20	\N	{"age": 22, "gender": "M", "address": "杭州市西湖区"}	{"title": "前端开发实习生", "stipend": 3000, "university": "浙江大学"}	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	2025-08-05 13:20:29.27323	2025-08-05 13:20:29.27323
\.


--
-- Data for Name: employees_backup; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.employees_backup (id, name, email, "position", created_at, updated_at, uuid_id) FROM stdin;
emp_001	张伟强	zhang.weiqiang@cubecastle.com	CTO & 联合创始人	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	fc92fb08-9da0-44a2-95ad-99264f866e30
emp_002	李芳芳	li.fangfang@cubecastle.com	CPO & 产品副总裁	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	a02c2f1c-0856-45df-9891-429e7b9c4be9
emp_003	王建国	wang.jianguo@cubecastle.com	VP Engineering	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	11de579b-3ddc-4c35-9816-aee99240d2f4
emp_004	刘美丽	liu.meili@cubecastle.com	VP Sales & Marketing	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	e99fa884-1c96-42a0-b912-5a93bfd6a449
emp_005	陈志华	chen.zhihua@cubecastle.com	CFO & 运营副总裁	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	85ea22ee-da9a-4f7a-8324-71d8799dc74c
emp_006	赵晓明	zhao.xiaoming@cubecastle.com	前端开发总监	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	93be3d76-e486-4e14-b92a-cfbf25cd537a
emp_007	孙丽娟	sun.lijuan@cubecastle.com	高级前端工程师	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	805b75b9-d8a9-4086-873b-29976d289f4c
emp_008	周强	zhou.qiang@cubecastle.com	高级前端工程师	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	59d809b7-b869-4c39-b175-5221af9b4936
emp_009	吴敏	wu.min@cubecastle.com	前端工程师	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	5c24a73f-229a-4e13-837e-c5c8800220d8
emp_010	郑海洋	zheng.haiyang@cubecastle.com	前端工程师	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	197d1c35-8114-42f1-96dd-17f56ec195fc
emp_011	冯雪梅	feng.xuemei@cubecastle.com	前端工程师	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	95505194-8306-4999-a84a-ce92ef0d065f
emp_012	蒋大伟	jiang.dawei@cubecastle.com	初级前端工程师	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	f24b3375-272b-487d-99e5-30bb7004faa5
emp_013	韩小红	han.xiaohong@cubecastle.com	前端实习生	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	33bf4431-7a8a-40ee-8d25-d6b18a49f3ba
emp_014	许文博	xu.wenbo@cubecastle.com	后端开发总监	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	56f28f28-7189-47b1-a24a-bd723506aaf5
emp_015	何晓峰	he.xiaofeng@cubecastle.com	首席后端架构师	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	fadc2600-635c-4751-9857-7cfe28ed0a99
emp_016	沈佳琪	shen.jiaqi@cubecastle.com	高级后端工程师	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	2235ff60-31af-4329-bf8f-fe7bbe607092
emp_017	卢志强	lu.zhiqiang@cubecastle.com	高级后端工程师	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	822a0ced-f020-40e3-a0b5-3d20cd0bf511
emp_018	施雨婷	shi.yuting@cubecastle.com	后端工程师	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	458a82e5-3eaf-4872-bf30-508900c7c7c1
emp_019	姚伟华	yao.weihua@cubecastle.com	后端工程师	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	73ac8c7e-80ce-4a40-8e90-46a7c461338f
emp_020	傅小丽	fu.xiaoli@cubecastle.com	后端工程师	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	8f170f3b-5d02-4dea-b879-112cd8f969aa
emp_021	邓建军	deng.jianjun@cubecastle.com	后端工程师	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	e9d4ac31-b31a-4a7d-9bbf-d5be46d83d4d
emp_022	曹明明	cao.mingming@cubecastle.com	初级后端工程师	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	b53cdd68-fabe-407b-8891-a59c56a315bf
emp_023	彭小强	peng.xiaoqiang@cubecastle.com	后端实习生	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	f27e67c2-c4b2-4408-ac07-18bfb6cab43a
emp_024	范志刚	fan.zhigang@cubecastle.com	移动开发总监	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	a9312f67-bd34-4733-8ff7-f54852f75e63
emp_025	苏美玲	su.meiling@cubecastle.com	高级移动开发工程师	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	ce9b9457-c329-42dc-bee7-b1016de8249d
emp_026	程晓燕	cheng.xiaoyan@cubecastle.com	移动开发工程师	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	d8c9d9cb-cea0-4627-adab-9c401f4b3ba2
emp_027	丁伟东	ding.weidong@cubecastle.com	移动开发工程师	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	bed3b2d3-79e6-493d-a7ba-963bac63e9f5
emp_028	白雪莹	bai.xueying@cubecastle.com	移动开发工程师	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	7668ea60-a111-4d78-8c6b-1480721ec133
emp_029	石磊	shi.lei@cubecastle.com	移动开发实习生	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	f00ba6ec-fd34-4010-97ea-b4e6280e4d06
emp_030	毛建华	mao.jianhua@cubecastle.com	数据工程总监	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	ce9543cf-b7f5-44d0-a6d8-5b917d33a05b
emp_031	文小芳	wen.xiaofang@cubecastle.com	高级数据工程师	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	a3ce8178-0fac-46b6-be39-15f6fdb94a76
emp_032	方志敏	fang.zhimin@cubecastle.com	数据工程师	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	ab333ccf-3252-47d0-a8c8-b14aa9a2a9cd
emp_033	宋雨桐	song.yutong@cubecastle.com	数据工程师	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	63ed8289-3a47-4889-947d-1d69cd2d2e19
emp_034	戴小明	dai.xiaoming@cubecastle.com	数据分析师	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	56906a42-2278-4e2a-8e1a-b8db8175082a
emp_035	侯伟光	hou.weiguang@cubecastle.com	DevOps总监	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	2e4b679c-91da-460c-8f94-585d37e71835
emp_036	薛晓琳	xue.xiaolin@cubecastle.com	高级DevOps工程师	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	ebf9f5d4-cca3-457c-9600-728b36a2b403
emp_037	顾志华	gu.zhihua@cubecastle.com	DevOps工程师	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	c0ec233e-219e-4a78-9006-53cd2c60558e
emp_038	廖小梅	liao.xiaomei@cubecastle.com	DevOps工程师	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	99c7ac75-8e1e-423c-ac35-7953c921c8f0
emp_039	谭建平	tan.jianping@cubecastle.com	测试总监	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	9ae2bca3-4537-48c4-a4a2-d67082c195f3
emp_040	洪美华	hong.meihua@cubecastle.com	高级测试工程师	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	86dc71ca-a6c6-4637-91fe-64ffb8e07c02
emp_041	黎志强	li.zhiqiang@cubecastle.com	测试工程师	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	b66f3fad-21cb-4455-92fb-d73f7dd223dc
emp_042	康小红	kang.xiaohong@cubecastle.com	测试工程师	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	215cf283-76a1-4a94-9eaf-a6792d1d11ab
emp_043	贺文静	he.wenjing@cubecastle.com	自动化测试工程师	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	afc929fb-46ed-4c93-adbd-378986fe09c7
emp_044	龙小飞	long.xiaofei@cubecastle.com	测试实习生	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	fd94fecc-be44-4e06-b1a3-9bfc7dc2b04c
emp_045	常晓东	chang.xiaodong@cubecastle.com	产品总监	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	e1a9f77e-9a17-4572-97be-f46e6c4f9a4a
emp_046	包雪芳	bao.xuefang@cubecastle.com	高级产品经理	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	f1bcdedf-c121-4e10-bf7d-f2ca6cdaad63
emp_047	华小明	hua.xiaoming@cubecastle.com	UX设计师	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	203ba039-e1a8-4123-a6a2-44f80c008d69
emp_048	金晓丽	jin.xiaoli@cubecastle.com	销售经理	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	1ce14bda-5f5f-45e3-9fcf-6f84154dbceb
emp_049	夏志华	xia.zhihua@cubecastle.com	人力资源经理	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	e394ecd1-7ca3-46be-a74e-f8d0528e5750
emp_050	武小强	wu.xiaoqiang@cubecastle.com	财务经理	2025-08-01 00:08:13.297323+00	2025-08-01 00:08:13.297323+00	cc68797e-6f09-4622-b948-c20a98d180bf
\.


--
-- Data for Name: metacontract_editor_projects; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.metacontract_editor_projects (id, name, description, content, version, status, tenant_id, created_by, created_at, updated_at, last_compiled, compile_error) FROM stdin;
\.


--
-- Data for Name: metacontract_editor_sessions; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.metacontract_editor_sessions (id, project_id, user_id, started_at, last_seen, active) FROM stdin;
\.


--
-- Data for Name: metacontract_editor_settings; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.metacontract_editor_settings (user_id, theme, font_size, auto_save, auto_compile, key_bindings, settings, updated_at) FROM stdin;
\.


--
-- Data for Name: metacontract_editor_templates; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.metacontract_editor_templates (id, name, description, category, content, tags, created_at, updated_at) FROM stdin;
cd35b143-a863-4d4b-97e7-b4b1cb903808	Employee Management Template	Basic template for employee management meta-contract	hr	# Employee Management Meta-Contract\n\nversion: "1.0.0"\nname: "employee_management"\ndescription: "Employee management system"\n\nentities:\n  Employee:\n    fields:\n      - name: id\n        type: UUID\n        required: true\n        primary_key: true\n      - name: first_name\n        type: String\n        required: true\n      - name: last_name\n        type: String\n        required: true\n      - name: email\n        type: String\n        required: true\n        unique: true\n      - name: hire_date\n        type: Date\n        required: true\n\nworkflows:\n  employee_onboarding:\n    description: "Employee onboarding process"\n    steps:\n      - name: create_employee\n        action: create\n        entity: Employee\n      - name: send_welcome_email\n        action: notify\n        template: welcome_email	{hr,employee,management,basic}	2025-07-30 23:43:33.491778+00	2025-07-30 23:43:33.491778+00
\.


--
-- Data for Name: organization_units; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.organization_units (code, parent_code, tenant_id, name, unit_type, status, level, path, sort_order, description, profile, created_at, updated_at) FROM stdin;
1000000	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	高谷集团	COMPANY	ACTIVE	1	/1000000	0	集团总公司	{"type": "headquarters"}	2025-08-05 11:23:01.426455+00	2025-08-05 11:23:01.426455+00
1000013	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	前端修复测试组织	DEPARTMENT	INACTIVE	1	/1000013	0	测试前端修复是否生效	{}	2025-08-07 12:59:43.272937+00	2025-08-08 08:29:48.077206+00
1000015	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	修复后测试部门	DEPARTMENT	INACTIVE	1	/1000015	0		{}	2025-08-07 13:13:16.726148+00	2025-08-08 08:29:50.749104+00
1000038	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	部门十一	DEPARTMENT	INACTIVE	1	/1000038	0		{}	2025-08-08 07:08:40.505431+00	2025-08-08 08:29:53.348862+00
1000039	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	部门十二	DEPARTMENT	INACTIVE	1	/1000039	0		{}	2025-08-08 08:04:49.327119+00	2025-08-08 08:29:55.720836+00
CUSTOM-001	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	测试部门自定义编码	DEPARTMENT	INACTIVE	1	/CUSTOM-001	0	测试用户自定义编码功能	{}	2025-08-08 05:54:01.321143+00	2025-08-08 08:29:59.435129+00
1000037	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	部门十	DEPARTMENT	INACTIVE	1	/1000037	0	测试新部门	{}	2025-08-08 07:04:27.723253+00	2025-08-08 08:30:02.006375+00
1000002	1000000	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	产品部	DEPARTMENT	ACTIVE	2	/1000000/1000002	0	产品管理部门	{"type": "product"}	2025-08-05 11:23:01.426455+00	2025-08-06 06:13:47.072807+00
1000003	1000000	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	销售部	DEPARTMENT	ACTIVE	2	/1000000/1000003	0	销售业务部门	{"type": "sales"}	2025-08-05 11:23:01.426455+00	2025-08-06 06:13:47.072807+00
1000032	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	测试部门自动编码	DEPARTMENT	INACTIVE	1	/1000032	0		{}	2025-08-08 05:53:51.209306+00	2025-08-08 08:30:04.197299+00
1000031	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	部门6	DEPARTMENT	INACTIVE	1	/1000031	0		{}	2025-08-08 05:40:22.517792+00	2025-08-08 08:30:06.055407+00
1000008	1000000	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	性能测试部门	DEPARTMENT	INACTIVE	2	/1000000/1000008	0	性能测试创建的部门	{}	2025-08-06 08:20:40.514638+00	2025-08-06 08:20:58.686919+00
1000030	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	部门五	COMPANY	INACTIVE	1	/1000030	0		{}	2025-08-08 05:39:18.546437+00	2025-08-08 08:30:10.056275+00
1000029	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	最终验收测试部门	DEPARTMENT	INACTIVE	1	/1000029	0	验证表单重复提交修复的最终测试	{}	2025-08-08 05:34:37.145632+00	2025-08-08 08:30:12.169852+00
1000012	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	测试组织创建	COMPANY	INACTIVE	1	/1000012	0	API测试	{}	2025-08-07 12:54:53.171785+00	2025-08-08 08:30:16.049528+00
1000014	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	最终测试验证组织	DEPARTMENT	INACTIVE	1	/1000014	0	确认端口修复成功	{}	2025-08-07 13:07:58.098469+00	2025-08-08 08:30:18.151684+00
1000022	1000000	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	页面测试部门	DEPARTMENT	INACTIVE	2	/1000000/1000022	0		{}	2025-08-08 04:28:16.951531+00	2025-08-08 08:30:20.610268+00
1000011	1000000	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	实时监控测试部门	DEPARTMENT	INACTIVE	2	/1000000/1000011	0	实时监控同步的测试部门	{}	2025-08-06 09:14:34.213943+00	2025-08-08 02:37:46.497481+00
1000018	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	Canvas Kit Modal测试部门	DEPARTMENT	ACTIVE	1	/1000018	0	验证Canvas Kit v13 Modal API修复后的端到端功能	{}	2025-08-08 03:46:21.395223+00	2025-08-08 03:46:21.39713+00
1000026	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	重复提交测试部门	DEPARTMENT	INACTIVE	1	/1000026	0		{}	2025-08-08 05:04:48.899194+00	2025-08-08 08:30:43.63678+00
1000016	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	CQRS测试部门	DEPARTMENT	INACTIVE	1	/1000016	0	用于验证CQRS命令服务功能	{}	2025-08-07 22:53:56.116101+00	2025-08-08 04:06:33.426054+00
1000027	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	修复验证测试部门-已更新	DEPARTMENT	INACTIVE	1	/1000027	0	更新测试成功	{}	2025-08-08 05:18:23.200839+00	2025-08-08 08:30:47.876905+00
1000019	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	测试部门	DEPARTMENT	ACTIVE	1	/1000019	0	\N	{}	2025-08-08 03:54:57.175359+00	2025-08-08 03:54:57.175716+00
1000020	1000000	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	API修复测试部门	DEPARTMENT	ACTIVE	2	/1000000/1000020	0		{}	2025-08-08 04:00:32.996385+00	2025-08-08 04:00:32.996721+00
1000028	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	部门四	COMPANY	INACTIVE	1	/1000028	0		{}	2025-08-08 05:34:03.038786+00	2025-08-08 08:30:48.424042+00
1000021	1000000	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	E2E测试部门_更新版	DEPARTMENT	INACTIVE	2	/1000000/1000021	0	\N	{}	2025-08-08 04:21:36.617388+00	2025-08-08 04:22:24.960447+00
1000017	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	托尔斯泰2	DEPARTMENT	INACTIVE	1	/1000017	0		{}	2025-08-07 23:54:58.159621+00	2025-08-08 04:55:32.135861+00
1000024	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	验收测试部门_已编辑	DEPARTMENT	INACTIVE	1	/1000024	0		{}	2025-08-08 05:03:04.121248+00	2025-08-08 23:35:31.076786+00
1000025	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	错误测试部门	DEPARTMENT	INACTIVE	1	/1000025	0		{}	2025-08-08 05:03:45.715432+00	2025-08-08 23:36:05.754032+00
1000023	1000000	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	修复成功测试部门	DEPARTMENT	INACTIVE	2	/1000000/1000023	0		{}	2025-08-08 04:53:42.734942+00	2025-08-09 00:38:10.615696+00
1000035	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	部门九	DEPARTMENT	INACTIVE	1	/1000035	0		{}	2025-08-08 06:06:12.680371+00	2025-08-08 06:10:57.246611+00
1000034	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	测试自动编码部门	DEPARTMENT	INACTIVE	1	/1000034	0		{}	2025-08-08 06:05:15.725628+00	2025-08-08 06:11:17.502145+00
1000033	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	页面测试部门2	DEPARTMENT	INACTIVE	1	/1000033	0	测试页面编码自动生成功能	{}	2025-08-08 06:04:58.171327+00	2025-08-08 06:11:53.307689+00
1000036	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	状态更新测试部门	DEPARTMENT	INACTIVE	1	/1000036	0		{}	2025-08-08 06:15:42.18495+00	2025-08-08 06:16:32.022373+00
1000010	1000000	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	数据同步测试部门_FINAL	DEPARTMENT	ACTIVE	2	/1000000/1000010	0	用于验证数据同步的测试部门	{}	2025-08-06 09:11:59.778608+00	2025-08-09 05:41:18.00958+00
1000009	1000000	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	GraphQL端到端测试部门	DEPARTMENT	ACTIVE	2	/1000000/1000009	0	用于验证GraphQL集成的测试部门	{}	2025-08-06 08:53:30.92514+00	2025-08-09 07:45:01.087376+00
1000005	1000000	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	运维部	DEPARTMENT	ACTIVE	2	/1000000/1000005	0	测试修复后的CDC实时同步	{}	2025-08-06 06:43:14.71988+00	2025-08-09 10:42:21.152661+00
1000001	1000000	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	AI治理办公室	DEPARTMENT	ACTIVE	2	/1000000/1000001	0	技术研发部门	{"type": "rd"}	2025-08-05 11:23:01.426455+00	2025-08-09 12:07:15.838099+00
1000004	1000000	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	人事部	DEPARTMENT	ACTIVE	2	/1000000/1000004	0	人力资源部门	{"type": "hr"}	2025-08-05 11:23:01.426455+00	2025-08-10 03:15:49.936921+00
1000007	1000000	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	CoreHR测试部门	DEPARTMENT	INACTIVE	2	/1000000/1000007	0	修改后的部门描述 - MCP验证测试	{}	2025-08-06 07:09:08.005666+00	2025-08-10 02:49:58.073305+00
1000006	1000000	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	同步测试部门禁用-公司-3级-描述4-排序7	COMPANY	INACTIVE	2	/1000000/1000006	7	描述4	{}	2025-08-06 07:00:33.177944+00	2025-08-10 03:55:14.63456+00
1000040	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	测试部门_1754668553755	DEPARTMENT	ACTIVE	1	/1000040	0	这是一个测试部门	{}	2025-08-08 15:55:53.759589+00	2025-08-08 15:55:53.760634+00
1000041	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	测试部门_1754669118696	DEPARTMENT	ACTIVE	1	/1000041	0	这是一个测试部门	{}	2025-08-08 16:05:18.711949+00	2025-08-08 16:05:18.712276+00
1000042	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	测试部门_1754669124970	DEPARTMENT	ACTIVE	1	/1000042	0	这是一个测试部门	{}	2025-08-08 16:05:24.97627+00	2025-08-08 16:05:24.976671+00
1000043	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	测试部门_1754669129397	DEPARTMENT	ACTIVE	1	/1000043	0	这是一个测试部门	{}	2025-08-08 16:05:29.412342+00	2025-08-08 16:05:29.412733+00
1000044	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	测试部门_1754669139984	DEPARTMENT	ACTIVE	1	/1000044	0	这是一个测试部门	{}	2025-08-08 16:05:40.001238+00	2025-08-08 16:05:40.001654+00
1000045	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	全面测试部门	DEPARTMENT	ACTIVE	1	/1000045	0		{}	2025-08-08 23:33:02.396731+00	2025-08-08 23:33:02.398649+00
1000046	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	测试部门_1754703523513	DEPARTMENT	ACTIVE	1	/1000046	0	这是一个测试部门	{}	2025-08-09 01:38:43.274837+00	2025-08-09 01:38:43.276731+00
1000047	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	测试部门_1754703527894	DEPARTMENT	ACTIVE	1	/1000047	0	这是一个测试部门	{}	2025-08-09 01:38:47.922057+00	2025-08-09 01:38:47.922372+00
1000048	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	部门二十	DEPARTMENT	ACTIVE	1	/1000048	0		{}	2025-08-09 02:04:47.443405+00	2025-08-09 02:04:47.443898+00
1000049	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	稳定性测试部门0	DEPARTMENT	ACTIVE	1	/1000049	0	\N	{}	2025-08-09 03:01:03.157302+00	2025-08-09 03:01:03.158256+00
1000050	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	稳定性测试部门1	DEPARTMENT	ACTIVE	1	/1000050	0	\N	{}	2025-08-09 03:01:03.324159+00	2025-08-09 03:01:03.324592+00
1000051	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	稳定性测试部门2	DEPARTMENT	ACTIVE	1	/1000051	0	\N	{}	2025-08-09 03:01:03.451726+00	2025-08-09 03:01:03.452036+00
1000052	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	稳定性测试部门3	DEPARTMENT	ACTIVE	1	/1000052	0	\N	{}	2025-08-09 03:01:03.576029+00	2025-08-09 03:01:03.576217+00
1000054	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	Phase2测试部门	DEPARTMENT	ACTIVE	1	/1000054	0	验证简化测试	{}	2025-08-09 05:05:54.217338+00	2025-08-09 05:05:54.218356+00
1000055	1000054	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	DDD简化更新测试	DEPARTMENT	INACTIVE	2	/1000054/1000055	0	测试层次计算	{}	2025-08-09 05:11:37.2748+00	2025-08-09 05:12:13.111427+00
C722694	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	CDC端到端测试	DEPARTMENT	ACTIVE	1	/C722694	1	验证CDC功能	{}	2025-08-09 06:58:14.346048+00	2025-08-09 06:58:14.346048+00
S722950	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	数据同步测试	DEPARTMENT	ACTIVE	1	/S722950	1	验证Neo4j同步功能	{}	2025-08-09 07:02:30.110333+00	2025-08-09 07:02:30.110333+00
U723048	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	解析修复测试	DEPARTMENT	ACTIVE	1	/U723048	1	验证unwrapped CDC事件	{}	2025-08-09 07:04:08.52621+00	2025-08-09 07:04:08.52621+00
F723128	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	CDC修复验证	DEPARTMENT	ACTIVE	1	/F723128	1	验证schema消息解析	{}	2025-08-09 07:05:28.554088+00	2025-08-09 07:05:28.554088+00
1000056	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	测试更新缓存_同步修复	DEPARTMENT	ACTIVE	1	/1000056	0	测试同步服务修复后的缓存失效	{}	2025-08-09 07:21:10.177689+00	2025-08-09 08:13:00.514444+00
1000053	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	稳定性测试部门4	DEPARTMENT	INACTIVE	1	/1000053	0	\N	{}	2025-08-09 03:01:03.703919+00	2025-08-09 08:16:23.560124+00
1000058	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	同步测试部门_验证	DEPARTMENT	ACTIVE	1	/1000058	0		{}	2025-08-09 08:47:46.220656+00	2025-08-09 08:47:46.221894+00
1000059	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	重新同步测试_部门	DEPARTMENT	ACTIVE	1	/1000059	0		{}	2025-08-09 08:49:46.600642+00	2025-08-09 08:49:46.602372+00
1000057	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	部门二十一	DEPARTMENT	ACTIVE	1	/1000057	0		{}	2025-08-09 08:40:36.610696+00	2025-08-09 08:52:50.589984+00
1000060	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	前端修复验证部门	DEPARTMENT	ACTIVE	1	/1000060	0	验证前端修复是否成功的测试部门	{}	2025-08-09 09:59:25.060107+00	2025-08-09 09:59:25.063766+00
1000061	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	部门二十二	DEPARTMENT	INACTIVE	1	/1000061	0		{}	2025-08-09 10:07:00.322212+00	2025-08-09 10:07:21.47672+00
1000062	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	E2E测试部门_缓存失效测试	DEPARTMENT	ACTIVE	1	/1000062	0		{}	2025-08-09 10:50:35.015368+00	2025-08-09 10:53:17.206794+00
1000063	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	性能测试组织_1754736868	DEPARTMENT	ACTIVE	1	/1000063	0		{}	2025-08-09 10:54:28.700026+00	2025-08-09 10:54:28.70107+00
1000064	1000000	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	性能测试组织(已更新)	DEPARTMENT	ACTIVE	2	/1000000/1000064	0		{}	2025-08-09 12:09:43.433069+00	2025-08-09 12:09:50.151959+00
1000065	1000000	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	端到端测试部门_已更新	DEPARTMENT	INACTIVE	2	/1000000/1000065	0		{}	2025-08-09 12:24:34.885664+00	2025-08-09 12:30:03.646997+00
1000066	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	健康检查测试部门	DEPARTMENT	ACTIVE	1	/1000066	0		{}	2025-08-09 12:30:31.109551+00	2025-08-09 12:30:31.111004+00
1000067	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	删除测试部门_1754742927	DEPARTMENT	INACTIVE	1	/1000067	0	用于测试删除功能	{}	2025-08-09 12:35:27.966243+00	2025-08-09 12:35:27.983511+00
1000068	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	CDC测试部门_1754742983	DEPARTMENT	ACTIVE	1	/1000068	0	CDC同步测试	{}	2025-08-09 12:36:28.515331+00	2025-08-09 12:36:28.516233+00
1000069	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	缓存失效测试_1754743228	DEPARTMENT	ACTIVE	1	/1000069	0	测试缓存失效机制	{}	2025-08-09 12:40:35.286829+00	2025-08-09 12:40:35.288004+00
1000070	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	缓存失效测试2_1754743246	DEPARTMENT	ACTIVE	1	/1000070	0	再次测试缓存失效	{}	2025-08-09 12:40:46.337728+00	2025-08-09 12:40:46.338665+00
1000071	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	性能测试部门_1754743988_1	DEPARTMENT	ACTIVE	1	/1000071	0	性能压力测试	{}	2025-08-09 12:53:08.788862+00	2025-08-09 12:53:08.789868+00
1000072	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	性能测试部门_1754743988_2	DEPARTMENT	ACTIVE	1	/1000072	0	性能压力测试	{}	2025-08-09 12:53:08.810222+00	2025-08-09 12:53:08.811267+00
1000073	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	性能测试部门_1754743988_3	DEPARTMENT	ACTIVE	1	/1000073	0	性能压力测试	{}	2025-08-09 12:53:08.825371+00	2025-08-09 12:53:08.826261+00
1000074	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	性能测试部门_1754743988_4	DEPARTMENT	ACTIVE	1	/1000074	0	性能压力测试	{}	2025-08-09 12:53:08.837663+00	2025-08-09 12:53:08.838478+00
1000075	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	性能测试部门_1754743988_5	DEPARTMENT	ACTIVE	1	/1000075	0	性能压力测试	{}	2025-08-09 12:53:08.850776+00	2025-08-09 12:53:08.851811+00
1000076	\N	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	测试编辑部门	COST_CENTER	ACTIVE	1	/1000076	0	完整编辑模式测试 - 所有字段可编辑	{}	2025-08-10 03:02:51.529857+00	2025-08-10 03:15:03.677593+00
\.


--
-- Data for Name: outbox_events; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.outbox_events (id, tenant_id, event_type, aggregate_id, event_data, status, attempt_count, created_at, processed_at, error_message) FROM stdin;
07fc1cc1-d951-4847-80ac-d044bba0cb4f	22b7afd0-4e27-40b5-85d7-3649cf7d2214	PositionCreatedEvent	7453981f-5740-4428-ae0d-ede318216462	{"status": "ACTIVE", "details": {"level": "Senior", "title": "Test Software Engineer", "description": "Test position for CQRS integration"}, "tenant_id": "22b7afd0-4e27-40b5-85d7-3649cf7d2214", "position_id": "7453981f-5740-4428-ae0d-ede318216462", "budgeted_fte": 1, "department_id": "4d97ed03-9bcd-44d4-880e-6ca5cc9a1a9c", "position_type": "REGULAR"}	PENDING	0	2025-08-03 22:59:03.89751+00	\N	\N
159fc695-427d-4108-b041-cb23cc7de3cc	22b7afd0-4e27-40b5-85d7-3649cf7d2214	PositionCreatedEvent	a10567d3-67bb-4998-ab70-472fdc54ad24	{"status": "ACTIVE", "details": {"title": "Test Position for Outbox"}, "tenant_id": "22b7afd0-4e27-40b5-85d7-3649cf7d2214", "position_id": "a10567d3-67bb-4998-ab70-472fdc54ad24", "budgeted_fte": 1, "department_id": "448c90b1-7836-4e6a-b6e3-3f4aabe46ef0", "position_type": "REGULAR"}	PENDING	0	2025-08-03 22:59:03.905795+00	\N	\N
\.


--
-- Data for Name: person; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.person (id, tenant_id, name, email, employee_id, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: position_assignments; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.position_assignments (id, position_code, employee_code, assignment_date, end_date, assignment_reason, created_at) FROM stdin;
\.


--
-- Data for Name: position_assignments_backup_20250805; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.position_assignments_backup_20250805 (id, tenant_id, position_id, employee_id, start_date, end_date, is_current, fte, assignment_type, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: position_attribute_histories; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.position_attribute_histories (id, tenant_id, position_type, job_profile_id, department_id, manager_position_id, status, budgeted_fte, details, effective_date, end_date, change_reason, changed_by, change_type, source_event_id, created_at, position_id) FROM stdin;
\.


--
-- Data for Name: position_code_sequence; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.position_code_sequence (tenant_id, last_code, updated_at) FROM stdin;
3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	1000008	2025-08-07 03:33:15.469337
\.


--
-- Data for Name: position_histories; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.position_histories (id, employee_id, organization_id, position_title, department, effective_date, end_date, is_active, is_retroactive, salary_data, change_reason, approval_status, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: position_history; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.position_history (id, tenant_id, employee_id, position_title, department, job_level, location, employment_type, reports_to_employee_id, effective_date, end_date, change_reason, is_retroactive, created_by, created_at, min_salary, max_salary, currency) FROM stdin;
98a23475-c0be-4ca8-929c-cda24d89796c	00000000-0000-0000-0000-000000000000	fc92fb08-9da0-44a2-95ad-99264f866e30	CTO & 联合创始人	架构部	EXECUTIVE	上海总部	FULL_TIME	\N	2020-01-01 00:00:00+00	\N	公司创立	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	800000.00	1200000.00	CNY
b8282b40-0cd3-46a9-a4ac-a0c567a8305f	00000000-0000-0000-0000-000000000000	203ba039-e1a8-4123-a6a2-44f80c008d69	UX设计师	UX设计部	REGULAR	上海总部	FULL_TIME	\N	2022-06-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	160000.00	300000.00	CNY
d98c82aa-d48d-46b9-97cb-af0c56be085b	00000000-0000-0000-0000-000000000000	1ce14bda-5f5f-45e3-9fcf-6f84154dbceb	销售经理	销售部	MANAGER	上海总部	FULL_TIME	\N	2021-08-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	180000.00	350000.00	CNY
3b465dfd-d583-4093-9044-7efb1dfea70e	00000000-0000-0000-0000-000000000000	e394ecd1-7ca3-46be-a74e-f8d0528e5750	人力资源经理	人力资源部	MANAGER	上海总部	FULL_TIME	\N	2021-08-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	180000.00	350000.00	CNY
de80e9d9-d17b-4b7b-a9ad-837ad86a1be5	00000000-0000-0000-0000-000000000000	cc68797e-6f09-4622-b948-c20a98d180bf	财务经理	财务部	MANAGER	上海总部	FULL_TIME	\N	2021-08-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	180000.00	350000.00	CNY
9975aae0-f6e4-45e0-8e1c-55c580add56e	00000000-0000-0000-0000-000000000000	a02c2f1c-0856-45df-9891-429e7b9c4be9	CPO & 产品副总裁	产品管理部	EXECUTIVE	上海总部	FULL_TIME	fc92fb08-9da0-44a2-95ad-99264f866e30	2020-01-01 00:00:00+00	\N	公司创立	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	600000.00	900000.00	CNY
94f80dae-14c6-4348-a320-c3a4f7f58dc6	00000000-0000-0000-0000-000000000000	11de579b-3ddc-4c35-9816-aee99240d2f4	VP Engineering	架构部	EXECUTIVE	上海总部	FULL_TIME	fc92fb08-9da0-44a2-95ad-99264f866e30	2020-06-01 00:00:00+00	\N	高级管理层加入	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	500000.00	800000.00	CNY
d3163b4b-7912-4828-a2fe-33d89a471a75	00000000-0000-0000-0000-000000000000	e99fa884-1c96-42a0-b912-5a93bfd6a449	VP Sales & Marketing	销售部	EXECUTIVE	上海总部	FULL_TIME	fc92fb08-9da0-44a2-95ad-99264f866e30	2020-06-01 00:00:00+00	\N	高级管理层加入	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	500000.00	800000.00	CNY
8fbd1fe4-fc47-49de-adda-b3eed31768fc	00000000-0000-0000-0000-000000000000	85ea22ee-da9a-4f7a-8324-71d8799dc74c	CFO & 运营副总裁	财务部	EXECUTIVE	上海总部	FULL_TIME	fc92fb08-9da0-44a2-95ad-99264f866e30	2020-01-01 00:00:00+00	\N	公司创立	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	600000.00	900000.00	CNY
71d0ec26-3a16-46ad-9877-2cd254dd5291	00000000-0000-0000-0000-000000000000	93be3d76-e486-4e14-b92a-cfbf25cd537a	前端开发总监	前端开发部	DIRECTOR	上海总部	FULL_TIME	11de579b-3ddc-4c35-9816-aee99240d2f4	2020-06-01 00:00:00+00	\N	部门负责人任命	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	350000.00	600000.00	CNY
de5400d1-76c7-4ba1-8271-c079589919b0	00000000-0000-0000-0000-000000000000	805b75b9-d8a9-4086-873b-29976d289f4c	高级前端工程师	前端开发部	SENIOR	上海总部	FULL_TIME	93be3d76-e486-4e14-b92a-cfbf25cd537a	2021-03-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	250000.00	420000.00	CNY
68798a7f-b585-4a7e-a4d7-35e787ea28d4	00000000-0000-0000-0000-000000000000	59d809b7-b869-4c39-b175-5221af9b4936	高级前端工程师	前端开发部	SENIOR	上海总部	FULL_TIME	93be3d76-e486-4e14-b92a-cfbf25cd537a	2021-03-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	250000.00	420000.00	CNY
c9f72822-a7e7-4c86-bb33-7a18aecd26ec	00000000-0000-0000-0000-000000000000	5c24a73f-229a-4e13-837e-c5c8800220d8	前端工程师	前端开发部	REGULAR	上海总部	FULL_TIME	805b75b9-d8a9-4086-873b-29976d289f4c	2022-01-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	180000.00	320000.00	CNY
6598477f-163e-4e01-9b61-07107dfd7395	00000000-0000-0000-0000-000000000000	197d1c35-8114-42f1-96dd-17f56ec195fc	前端工程师	前端开发部	REGULAR	上海总部	FULL_TIME	805b75b9-d8a9-4086-873b-29976d289f4c	2022-01-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	180000.00	320000.00	CNY
e89f8078-75a7-402e-a004-df8870681825	00000000-0000-0000-0000-000000000000	95505194-8306-4999-a84a-ce92ef0d065f	前端工程师	前端开发部	REGULAR	上海总部	FULL_TIME	805b75b9-d8a9-4086-873b-29976d289f4c	2022-01-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	180000.00	320000.00	CNY
e41b155e-362c-4fb7-bf51-cc250a2ec4d3	00000000-0000-0000-0000-000000000000	f24b3375-272b-487d-99e5-30bb7004faa5	初级前端工程师	前端开发部	JUNIOR	上海总部	FULL_TIME	805b75b9-d8a9-4086-873b-29976d289f4c	2023-07-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	120000.00	200000.00	CNY
f33fdaad-b27e-4f4f-a93f-901a0b48c741	00000000-0000-0000-0000-000000000000	33bf4431-7a8a-40ee-8d25-d6b18a49f3ba	前端实习生	前端开发部	INTERN	上海总部	INTERN	805b75b9-d8a9-4086-873b-29976d289f4c	2024-09-01 00:00:00+00	\N	实习项目	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	8000.00	15000.00	CNY
7980254e-e32d-4e66-bfca-4482082a60c4	00000000-0000-0000-0000-000000000000	56f28f28-7189-47b1-a24a-bd723506aaf5	后端开发总监	后端开发部	DIRECTOR	上海总部	FULL_TIME	11de579b-3ddc-4c35-9816-aee99240d2f4	2020-06-01 00:00:00+00	\N	部门负责人任命	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	350000.00	600000.00	CNY
118dadfe-f3f5-40dc-b39c-afd784dfb547	00000000-0000-0000-0000-000000000000	fadc2600-635c-4751-9857-7cfe28ed0a99	首席后端架构师	后端开发部	PRINCIPAL	上海总部	FULL_TIME	56f28f28-7189-47b1-a24a-bd723506aaf5	2022-06-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	500000.00	700000.00	CNY
d27c72bc-3ed7-43c3-939d-42673f574da5	00000000-0000-0000-0000-000000000000	2235ff60-31af-4329-bf8f-fe7bbe607092	高级后端工程师	后端开发部	SENIOR	上海总部	FULL_TIME	56f28f28-7189-47b1-a24a-bd723506aaf5	2021-03-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	250000.00	420000.00	CNY
5bffdbaf-5ad4-429a-9754-106fa341ee7e	00000000-0000-0000-0000-000000000000	822a0ced-f020-40e3-a0b5-3d20cd0bf511	高级后端工程师	后端开发部	SENIOR	上海总部	FULL_TIME	56f28f28-7189-47b1-a24a-bd723506aaf5	2021-03-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	250000.00	420000.00	CNY
3d865934-0964-4f0d-87e7-f2a67c35b997	00000000-0000-0000-0000-000000000000	458a82e5-3eaf-4872-bf30-508900c7c7c1	后端工程师	后端开发部	REGULAR	上海总部	FULL_TIME	2235ff60-31af-4329-bf8f-fe7bbe607092	2022-01-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	180000.00	320000.00	CNY
5862137a-3a89-4240-97e0-3300391ed909	00000000-0000-0000-0000-000000000000	73ac8c7e-80ce-4a40-8e90-46a7c461338f	后端工程师	后端开发部	REGULAR	上海总部	FULL_TIME	2235ff60-31af-4329-bf8f-fe7bbe607092	2022-01-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	180000.00	320000.00	CNY
d1f90b92-fdbe-4c3a-863c-4ce81f355a2b	00000000-0000-0000-0000-000000000000	8f170f3b-5d02-4dea-b879-112cd8f969aa	后端工程师	后端开发部	REGULAR	上海总部	FULL_TIME	2235ff60-31af-4329-bf8f-fe7bbe607092	2022-01-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	180000.00	320000.00	CNY
ec6c5648-1ee7-4849-9365-df02ec64457a	00000000-0000-0000-0000-000000000000	e9d4ac31-b31a-4a7d-9bbf-d5be46d83d4d	后端工程师	后端开发部	REGULAR	上海总部	FULL_TIME	2235ff60-31af-4329-bf8f-fe7bbe607092	2022-01-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	180000.00	320000.00	CNY
2628e1ab-cb8e-4982-92ed-1bd4162b1921	00000000-0000-0000-0000-000000000000	b53cdd68-fabe-407b-8891-a59c56a315bf	初级后端工程师	后端开发部	JUNIOR	上海总部	FULL_TIME	2235ff60-31af-4329-bf8f-fe7bbe607092	2023-07-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	120000.00	200000.00	CNY
7060400e-7c6b-4d8e-afa1-31d260ba0d0c	00000000-0000-0000-0000-000000000000	f27e67c2-c4b2-4408-ac07-18bfb6cab43a	后端实习生	后端开发部	INTERN	上海总部	INTERN	2235ff60-31af-4329-bf8f-fe7bbe607092	2024-09-01 00:00:00+00	\N	实习项目	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	8000.00	15000.00	CNY
a626e9b1-e515-41a0-a733-f1fca4e82f3f	00000000-0000-0000-0000-000000000000	a9312f67-bd34-4733-8ff7-f54852f75e63	移动开发总监	移动开发部	DIRECTOR	上海总部	FULL_TIME	11de579b-3ddc-4c35-9816-aee99240d2f4	2020-06-01 00:00:00+00	\N	部门负责人任命	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	350000.00	600000.00	CNY
4072992b-881c-44e4-b234-07ed8d6b27de	00000000-0000-0000-0000-000000000000	ce9b9457-c329-42dc-bee7-b1016de8249d	高级移动开发工程师	移动开发部	SENIOR	上海总部	FULL_TIME	a9312f67-bd34-4733-8ff7-f54852f75e63	2021-03-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	250000.00	420000.00	CNY
a0a8032f-a6e3-4735-9e92-9f356d9e23b9	00000000-0000-0000-0000-000000000000	d8c9d9cb-cea0-4627-adab-9c401f4b3ba2	移动开发工程师	移动开发部	REGULAR	上海总部	FULL_TIME	ce9b9457-c329-42dc-bee7-b1016de8249d	2022-01-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	180000.00	320000.00	CNY
cfdfa539-cb93-470e-9e2f-bd1efc943faf	00000000-0000-0000-0000-000000000000	bed3b2d3-79e6-493d-a7ba-963bac63e9f5	移动开发工程师	移动开发部	REGULAR	上海总部	FULL_TIME	ce9b9457-c329-42dc-bee7-b1016de8249d	2022-01-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	180000.00	320000.00	CNY
53257845-3af1-4e4e-b0ea-e25e0fb62788	00000000-0000-0000-0000-000000000000	7668ea60-a111-4d78-8c6b-1480721ec133	移动开发工程师	移动开发部	REGULAR	上海总部	FULL_TIME	ce9b9457-c329-42dc-bee7-b1016de8249d	2022-01-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	180000.00	320000.00	CNY
7e45223e-3342-4d64-baac-e805718f12c0	00000000-0000-0000-0000-000000000000	f00ba6ec-fd34-4010-97ea-b4e6280e4d06	移动开发实习生	移动开发部	INTERN	上海总部	INTERN	ce9b9457-c329-42dc-bee7-b1016de8249d	2024-09-01 00:00:00+00	\N	实习项目	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	8000.00	15000.00	CNY
64073e04-9844-446c-b7ec-65fa33a5885f	00000000-0000-0000-0000-000000000000	ce9543cf-b7f5-44d0-a6d8-5b917d33a05b	数据工程总监	数据工程部	DIRECTOR	上海总部	FULL_TIME	11de579b-3ddc-4c35-9816-aee99240d2f4	2020-06-01 00:00:00+00	\N	部门负责人任命	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	350000.00	600000.00	CNY
4bb73aea-83f8-4042-b66a-043f2c20cf3d	00000000-0000-0000-0000-000000000000	a3ce8178-0fac-46b6-be39-15f6fdb94a76	高级数据工程师	数据工程部	SENIOR	上海总部	FULL_TIME	ce9543cf-b7f5-44d0-a6d8-5b917d33a05b	2021-03-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	250000.00	420000.00	CNY
4f21e6e7-8104-4c28-b3d9-1b663dfc5dcd	00000000-0000-0000-0000-000000000000	ab333ccf-3252-47d0-a8c8-b14aa9a2a9cd	数据工程师	数据工程部	REGULAR	上海总部	FULL_TIME	a3ce8178-0fac-46b6-be39-15f6fdb94a76	2022-01-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	180000.00	320000.00	CNY
49c58168-cb72-42d8-91fb-c863cfa702cc	00000000-0000-0000-0000-000000000000	63ed8289-3a47-4889-947d-1d69cd2d2e19	数据工程师	数据工程部	REGULAR	上海总部	FULL_TIME	a3ce8178-0fac-46b6-be39-15f6fdb94a76	2022-01-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	180000.00	320000.00	CNY
9db1af40-c0f3-4460-9e6d-08ac0864fd8f	00000000-0000-0000-0000-000000000000	56906a42-2278-4e2a-8e1a-b8db8175082a	数据分析师	数据工程部	REGULAR	上海总部	FULL_TIME	a3ce8178-0fac-46b6-be39-15f6fdb94a76	2022-06-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	160000.00	300000.00	CNY
f5534b67-1bf2-4bcb-a14a-dc85d3e3bd16	00000000-0000-0000-0000-000000000000	2e4b679c-91da-460c-8f94-585d37e71835	DevOps总监	DevOps部	DIRECTOR	上海总部	FULL_TIME	11de579b-3ddc-4c35-9816-aee99240d2f4	2020-06-01 00:00:00+00	\N	部门负责人任命	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	350000.00	600000.00	CNY
02d05dcc-de8d-4c15-a4ce-e442657006ce	00000000-0000-0000-0000-000000000000	ebf9f5d4-cca3-457c-9600-728b36a2b403	高级DevOps工程师	DevOps部	SENIOR	上海总部	FULL_TIME	2e4b679c-91da-460c-8f94-585d37e71835	2021-03-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	250000.00	420000.00	CNY
af2805dd-57cc-441c-8fab-599cc426c93c	00000000-0000-0000-0000-000000000000	c0ec233e-219e-4a78-9006-53cd2c60558e	DevOps工程师	DevOps部	REGULAR	上海总部	FULL_TIME	ebf9f5d4-cca3-457c-9600-728b36a2b403	2022-01-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	180000.00	320000.00	CNY
a4cc4e5f-7116-4941-a343-7559ee07357a	00000000-0000-0000-0000-000000000000	99c7ac75-8e1e-423c-ac35-7953c921c8f0	DevOps工程师	DevOps部	REGULAR	上海总部	FULL_TIME	ebf9f5d4-cca3-457c-9600-728b36a2b403	2022-01-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	180000.00	320000.00	CNY
6b9814a6-5723-4f29-ae29-09851ff61827	00000000-0000-0000-0000-000000000000	9ae2bca3-4537-48c4-a4a2-d67082c195f3	测试总监	测试部	DIRECTOR	上海总部	FULL_TIME	11de579b-3ddc-4c35-9816-aee99240d2f4	2020-06-01 00:00:00+00	\N	部门负责人任命	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	350000.00	600000.00	CNY
91339f87-4327-4180-ae67-886688378401	00000000-0000-0000-0000-000000000000	86dc71ca-a6c6-4637-91fe-64ffb8e07c02	高级测试工程师	测试部	SENIOR	上海总部	FULL_TIME	9ae2bca3-4537-48c4-a4a2-d67082c195f3	2021-03-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	250000.00	420000.00	CNY
beec4881-ca11-4d37-8cfd-cdbf005182e8	00000000-0000-0000-0000-000000000000	b66f3fad-21cb-4455-92fb-d73f7dd223dc	测试工程师	测试部	REGULAR	上海总部	FULL_TIME	86dc71ca-a6c6-4637-91fe-64ffb8e07c02	2022-01-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	180000.00	320000.00	CNY
7d8b5908-80a4-4afc-8e30-046d3afae1f8	00000000-0000-0000-0000-000000000000	215cf283-76a1-4a94-9eaf-a6792d1d11ab	测试工程师	测试部	REGULAR	上海总部	FULL_TIME	86dc71ca-a6c6-4637-91fe-64ffb8e07c02	2022-01-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	180000.00	320000.00	CNY
72db9a53-05a1-4d54-8678-d466333c23e6	00000000-0000-0000-0000-000000000000	afc929fb-46ed-4c93-adbd-378986fe09c7	自动化测试工程师	测试部	REGULAR	上海总部	FULL_TIME	86dc71ca-a6c6-4637-91fe-64ffb8e07c02	2022-01-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	180000.00	320000.00	CNY
186367b0-7687-4b49-b45d-4b4f9e31db86	00000000-0000-0000-0000-000000000000	fd94fecc-be44-4e06-b1a3-9bfc7dc2b04c	测试实习生	测试部	INTERN	上海总部	INTERN	86dc71ca-a6c6-4637-91fe-64ffb8e07c02	2024-09-01 00:00:00+00	\N	实习项目	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	8000.00	15000.00	CNY
ec07b8a8-8e7f-4300-97a5-9172e9a90a3b	00000000-0000-0000-0000-000000000000	e1a9f77e-9a17-4572-97be-f46e6c4f9a4a	产品总监	产品管理部	DIRECTOR	上海总部	FULL_TIME	a02c2f1c-0856-45df-9891-429e7b9c4be9	2020-06-01 00:00:00+00	\N	部门负责人任命	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	350000.00	600000.00	CNY
a206d7e6-927d-4fb4-828d-ed46478ff95e	00000000-0000-0000-0000-000000000000	f1bcdedf-c121-4e10-bf7d-f2ca6cdaad63	高级产品经理	产品管理部	SENIOR	上海总部	FULL_TIME	e1a9f77e-9a17-4572-97be-f46e6c4f9a4a	2021-03-01 00:00:00+00	\N	团队扩张	t	11111111-1111-1111-1111-111111111111	2025-08-01 00:54:07.445413+00	250000.00	420000.00	CNY
\.


--
-- Data for Name: position_occupancy_histories; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.position_occupancy_histories (id, tenant_id, employee_id, start_date, end_date, is_active, assignment_type, assignment_reason, fte_percentage, work_arrangement, approved_by, approval_date, approval_reference, compensation_data, performance_review_cycle, source_event_id, created_at, updated_at, position_id) FROM stdin;
\.


--
-- Data for Name: positions; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.positions (code, organization_code, manager_position_code, position_type, job_profile_id, status, budgeted_fte, details, tenant_id, created_at, updated_at) FROM stdin;
1000001	1000000	\N	FULL_TIME	550e8400-e29b-41d4-a716-446655440000	OPEN	1.00	{"title": "高级软件工程师", "salary_range": {"max": 35000, "min": 20000, "currency": "CNY"}}	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	2025-08-05 12:51:52.988398	2025-08-05 12:51:52.988398
1000002	1000000	\N	FULL_TIME	550e8400-e29b-41d4-a716-446655440000	OPEN	1.00	{"title": "软件架构师", "salary_range": {"max": 50000, "min": 30000, "currency": "CNY"}}	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	2025-08-05 12:51:52.988398	2025-08-05 12:51:52.988398
1000004	1000001	\N	PART_TIME	550e8400-e29b-41d4-a716-446655440000	OPEN	0.50	{"title": "UI设计师", "hourly_rate": 200, "max_hours_per_week": 20}	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	2025-08-05 12:51:52.988398	2025-08-05 12:51:52.988398
1000003	1000001	1000001	FULL_TIME	550e8400-e29b-41d4-a716-446655440000	FILLED	1.00	{"title": "产品经理", "salary_range": {"max": 40000, "min": 25000, "currency": "CNY"}}	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	2025-08-05 12:51:52.988398	2025-08-05 12:52:15.792605
1000005	1000002	1000002	INTERN	550e8400-e29b-41d4-a716-446655440000	FILLED	0.80	{"title": "前端开发实习生", "stipend": 3000, "internship_duration": "3m"}	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	2025-08-05 12:51:52.988398	2025-08-05 12:52:15.792605
1000006	1000001	1000001	FULL_TIME	550e8400-e29b-41d4-a716-446655440000	OPEN	1.00	{"title": "高级产品经理", "salary_range": {"max": 40000, "min": 25000, "currency": "CNY"}}	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	2025-08-05 12:59:07.227494	2025-08-05 12:59:07.227494
1000007	1000010	\N	FULL_TIME	550e8400-e29b-41d4-a716-446655440000	OPEN	1.00	{"title": "端到端测试工程师", "salary_range": {"max": 28000, "min": 18000, "currency": "CNY"}}	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	2025-08-07 03:26:31.060852	2025-08-07 03:26:31.060852
\.


--
-- Data for Name: positions_backup; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.positions_backup (id, tenant_id, title, department, level, description, requirements, is_active, created_at, updated_at) FROM stdin;
9d09861c-e247-4ef7-9ea4-2a4496e0c8e4	9c3f27f9-15d0-45b6-a8ec-931cb07dbd0c	前端开发工程师	技术研发部	JUNIOR	负责Web前端开发	React, TypeScript	t	2025-08-03 08:51:35.728373+00	2025-08-03 08:51:35.728373+00
15c4afe5-5242-473a-bf82-74682c314951	9c3f27f9-15d0-45b6-a8ec-931cb07dbd0c	后端开发工程师	技术研发部	JUNIOR	负责后端API开发	Go, PostgreSQL	t	2025-08-03 08:51:35.728373+00	2025-08-03 08:51:35.728373+00
544f6f0d-1631-4a5d-a30a-e2ae8bb4a0a6	9c3f27f9-15d0-45b6-a8ec-931cb07dbd0c	人事专员	人力资源部	JUNIOR	负责招聘和员工关系	人力资源管理	t	2025-08-03 08:51:35.728373+00	2025-08-03 08:51:35.728373+00
f7d1c950-eb48-4e04-91a4-cc243caa5732	9c3f27f9-15d0-45b6-a8ec-931cb07dbd0c	市场营销专员	市场营销部	JUNIOR	负责市场推广	市场营销、数字营销	t	2025-08-03 08:51:35.728373+00	2025-08-03 08:51:35.728373+00
\.


--
-- Data for Name: positions_backup_20250805; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.positions_backup_20250805 (id, tenant_id, position_type, job_profile_id, department_id, manager_position_id, status, budgeted_fte, details, created_at, updated_at, business_id) FROM stdin;
06acc8c3-7478-48ad-a1f0-4132978a6d53	550e8400-e29b-41d4-a716-446655440000	REGULAR	15e4cfd0-9344-4e71-af23-38a80183cfd8	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000000
f83be542-2673-4775-be77-aa04bef72da4	550e8400-e29b-41d4-a716-446655440000	REGULAR	41857c6c-6d82-46a7-97e4-f66bea1d2278	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000001
0d8541ba-4cbb-4cbf-bc80-72e44c432254	550e8400-e29b-41d4-a716-446655440000	REGULAR	161b44a7-70ec-48ae-baed-d6fd84272ad2	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000002
17acc75b-83ca-4b94-bc24-b1c7d6e566fd	550e8400-e29b-41d4-a716-446655440000	REGULAR	8c3d8e89-3918-41ea-aa73-8e2dd7173ba7	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000003
0414eb7a-96fe-455f-99b3-161ce8bdfee3	550e8400-e29b-41d4-a716-446655440000	REGULAR	58fe6d4b-005d-4ec4-8035-e1bd75a09082	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000004
b57f64c5-0d38-4129-8a08-05084432bf32	550e8400-e29b-41d4-a716-446655440000	REGULAR	861f3984-a574-423b-8f7b-6400abdd090a	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000005
360e61e3-3dca-4b99-9e62-bcbca83ade11	550e8400-e29b-41d4-a716-446655440000	REGULAR	cde2d89f-9355-4a7c-9a15-c923993ab1ef	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000006
18dadfaa-8a2a-4568-9cba-9b48a6e1439f	550e8400-e29b-41d4-a716-446655440000	REGULAR	4150277d-8c4b-458f-9131-585dc0a37157	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000007
5eb49f8b-c63f-4254-9b25-f5f49603e862	550e8400-e29b-41d4-a716-446655440000	REGULAR	f5ff49f1-9742-4478-96f1-8367c1365150	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000008
fea6f82e-abab-483e-af50-448493379100	550e8400-e29b-41d4-a716-446655440000	REGULAR	f4803e29-e5c1-4898-b6e3-ca7c0690276c	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000009
b3958664-4659-4728-9b4b-0e09527d41a3	550e8400-e29b-41d4-a716-446655440000	REGULAR	154d279a-5fb6-4b93-98db-207ecc2b7f39	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000010
1878c3d8-c4ca-4975-bd7f-567da25a9dda	550e8400-e29b-41d4-a716-446655440000	REGULAR	f9ab44bc-d2b2-4640-b779-0667fcbb1592	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000011
b59f6e78-7817-47c8-97ff-41f4cfe071e7	550e8400-e29b-41d4-a716-446655440000	REGULAR	a707de46-4197-4bca-bbc0-be567902ff90	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000012
888b9464-989e-4c06-9011-fcd55212e00c	550e8400-e29b-41d4-a716-446655440000	REGULAR	9ee8c9f9-ed90-43da-9596-f93d5fbd15ef	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000013
af308ee9-9a96-443b-ac49-99fd61896c39	550e8400-e29b-41d4-a716-446655440000	REGULAR	56e9091f-b15f-480f-8843-acb7bed4655e	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000014
fe71d14c-7a4b-41d9-a906-68f7ebfd86f3	550e8400-e29b-41d4-a716-446655440000	REGULAR	db50d18b-e16e-4ef9-9687-4c825a4eabb2	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000015
5b8ef7b4-6812-4fce-b7b0-0c8ce5e0e40a	550e8400-e29b-41d4-a716-446655440000	REGULAR	2a8af1ef-70bf-43bc-9d96-0be47a543453	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000016
f3b26e7f-cd2b-4c50-826f-3d8b48cc3f3f	550e8400-e29b-41d4-a716-446655440000	REGULAR	955b1828-3592-4b21-a101-eef7cb99ead7	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000017
a582c840-fc2a-4bcc-aaf6-8b96bcb1faae	550e8400-e29b-41d4-a716-446655440000	REGULAR	29c630a7-8d1b-40bc-8362-2e8a0d9097d8	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000018
4bd2fc5a-2efb-416f-8dd7-0f6a05dfa1ac	550e8400-e29b-41d4-a716-446655440000	REGULAR	6b207ab6-b240-4bbb-a4c3-309f288d1f19	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000019
fc891f54-2f55-4046-82c1-637eaa619fa8	550e8400-e29b-41d4-a716-446655440000	REGULAR	7ee2041e-e2b9-459d-b285-6a231f69bd2f	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000020
c897f28c-185e-4d56-8591-16f18aaf5bed	550e8400-e29b-41d4-a716-446655440000	REGULAR	7b318bce-aa9d-4807-8c39-1d48095e03f7	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000021
95dcbcbc-ef6a-4e52-82b9-4df044ede087	550e8400-e29b-41d4-a716-446655440000	REGULAR	99489b7f-7e51-48d6-93dc-1f7c6904c490	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000022
3102937c-b969-48a4-8ae7-f0555bbc4b0e	550e8400-e29b-41d4-a716-446655440000	REGULAR	33f4a068-3932-487d-ba05-a91f5bd2f558	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000023
0af4e50a-7508-4905-9cc6-e43fb546a404	550e8400-e29b-41d4-a716-446655440000	REGULAR	d51e282e-727a-49d8-8bea-344625725548	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000024
ac512751-cbb1-4a85-afe1-43421d75b044	550e8400-e29b-41d4-a716-446655440000	REGULAR	73ce146c-bfef-44aa-a860-7fcb96eb0efc	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000025
546ab95e-e509-48b4-97f6-95df214514ba	550e8400-e29b-41d4-a716-446655440000	REGULAR	2a2f210f-36f8-4e0e-a0f3-95aa18e20c2b	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000026
94e3fee4-41d3-4db5-aacc-ddc35e1686c9	550e8400-e29b-41d4-a716-446655440000	REGULAR	ba2b275c-c6b5-49ec-a485-83fd1b602f18	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000027
8aa1f81e-50f0-4c50-a088-c1a497ea9bee	550e8400-e29b-41d4-a716-446655440000	REGULAR	94ee84f2-1e9c-4e48-885d-a7a3276caa68	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000028
863889c8-aae3-4e52-b3f3-5038658f3ae4	550e8400-e29b-41d4-a716-446655440000	REGULAR	2b93218a-e14c-4375-bcad-aff0224e84ed	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000029
ee0f6a77-47a3-4138-88f0-f8ee2f77082b	550e8400-e29b-41d4-a716-446655440000	REGULAR	db15e284-c619-4a1f-8836-e28cf723c4a4	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000030
a0e7974f-95c5-4ae2-af42-f2e4215b3830	550e8400-e29b-41d4-a716-446655440000	REGULAR	94e4236d-c9ba-47e0-baf6-64095fd8c5bf	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000031
f2d71f30-eaa6-49df-8583-428ac7263c6d	550e8400-e29b-41d4-a716-446655440000	REGULAR	d0b35faa-f1d4-4f34-8ff0-c2a5a64b6888	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000032
017233e6-7620-422d-8cde-50d841630bad	550e8400-e29b-41d4-a716-446655440000	REGULAR	1c6713f9-f603-4b97-93a3-aed03c1e7a03	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000033
fbcc54ae-332b-42fb-842c-25bf7b5e34ed	550e8400-e29b-41d4-a716-446655440000	REGULAR	9b7ef8f5-0c79-4eca-a38c-09525021e7b3	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000034
4f86e191-feb7-412f-b84a-ae5976108fa0	550e8400-e29b-41d4-a716-446655440000	REGULAR	c698fac1-486d-400e-a013-1c0c1e922ea6	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000035
8d6d11ce-7705-4a2c-b5e8-efa2ebeaaee9	550e8400-e29b-41d4-a716-446655440000	REGULAR	5162e063-c45b-4e56-bb6d-d34282eabaa0	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000036
d13fd19e-664f-4a80-8933-756c66b3edd2	550e8400-e29b-41d4-a716-446655440000	REGULAR	83646f3f-8349-4c1f-960f-4aabb275d1fa	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000037
a993ef36-2017-46b4-b402-a5c18eda3484	550e8400-e29b-41d4-a716-446655440000	REGULAR	25bd55b6-2d72-4f2e-aeb5-7c98ab2b3aad	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000038
44cc1f19-b05a-48c0-a288-e04358a741fa	550e8400-e29b-41d4-a716-446655440000	REGULAR	cc5a628f-aa92-461c-8221-f19f0d19c459	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000039
311f6f8a-fa60-4a10-ae7d-6b950c132f2b	550e8400-e29b-41d4-a716-446655440000	REGULAR	9d57fbb2-289a-441f-98e8-7e6f6d48d86b	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000040
94ced1fc-ce4d-47b7-a6b5-488218881e63	550e8400-e29b-41d4-a716-446655440000	REGULAR	1197ae37-614d-4c53-9b6f-abc36a160169	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000041
61adeb62-05f5-481c-8bdd-d0a2d2127c2c	550e8400-e29b-41d4-a716-446655440000	REGULAR	7631cb69-8dfd-4c35-8a08-956aed3fd434	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000042
1ae89b4f-5101-4894-84b2-da849c37a8b8	550e8400-e29b-41d4-a716-446655440000	REGULAR	428935be-b6a3-48cc-99e3-e9aa8691baaf	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000043
7847c21b-61a4-4675-a5fc-484fbfa85642	550e8400-e29b-41d4-a716-446655440000	REGULAR	731765ce-6af4-4248-bdd1-21bdeb3d4cb7	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000044
1d705097-5939-4adc-9771-568cdce0094e	550e8400-e29b-41d4-a716-446655440000	REGULAR	e1cd7527-6e09-439d-acf4-b7183326e00b	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000045
46797a8b-34e4-4cae-9acc-12615b0384df	550e8400-e29b-41d4-a716-446655440000	REGULAR	10a2dc31-ba59-4bb2-ab52-e204a111608a	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000046
efb46b7d-c633-40f6-92ed-cdf001ecbe7d	550e8400-e29b-41d4-a716-446655440000	REGULAR	87e12259-9697-4c02-8567-074407922d07	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000047
4d711c81-2813-4977-aba1-67da21335602	550e8400-e29b-41d4-a716-446655440000	REGULAR	cd052b81-9d05-429f-a15e-b16e713e0908	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000048
7038220e-2daf-4966-97f0-de6d2300d621	550e8400-e29b-41d4-a716-446655440000	REGULAR	fbdb6d97-63b3-4e02-9caf-ee0ce9f875dd	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000049
a8e97605-c8d3-4f4a-8fe1-925c9803c781	550e8400-e29b-41d4-a716-446655440000	REGULAR	96076bc6-72ca-4fd3-a69d-723fd84b4550	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000050
af5264ac-3718-479c-8412-65be091b6ee4	550e8400-e29b-41d4-a716-446655440000	REGULAR	d7949a55-f1fb-4a45-a9b5-e60138f6f209	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000051
5394c57a-3ffd-42af-b876-5bbff0e6d9c3	550e8400-e29b-41d4-a716-446655440000	REGULAR	2d88151f-fda5-4b6e-979b-218b65954a01	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000052
002ac680-9205-418a-8669-81391824274a	550e8400-e29b-41d4-a716-446655440000	REGULAR	910cba8a-90e3-43df-b01d-c18b61d6787f	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000053
f3d58a3c-bbce-4f06-8871-c6579be24c94	550e8400-e29b-41d4-a716-446655440000	REGULAR	1fd9081e-d0dc-4559-8f55-fe43a0ad2afd	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000054
e03b9bbd-fa6d-4328-abcb-a8660d68866a	550e8400-e29b-41d4-a716-446655440000	REGULAR	4b17afff-b8c3-4cc1-b17f-9bcc22c70078	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000055
7d476e1d-c896-4d00-868b-6e8271f45429	550e8400-e29b-41d4-a716-446655440000	REGULAR	4645f704-fd93-4e20-8df1-b7c2902fcd4b	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000056
6836abdc-d312-4691-b174-804c54685a03	550e8400-e29b-41d4-a716-446655440000	REGULAR	a8258c8a-6a28-4ca3-8df0-6bbf5d3b5cd4	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000057
8e0df9a7-53c5-41ea-a385-cd08d1f63949	550e8400-e29b-41d4-a716-446655440000	REGULAR	645d80fe-a25f-4424-a083-de14192b2ad0	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000058
1c8143fc-bddc-403e-b881-63dea7b973dd	550e8400-e29b-41d4-a716-446655440000	REGULAR	ff7531de-3c9e-4a62-baa2-e982a3996056	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000059
e3cc4556-d1db-4daf-8ba2-f151636dcf90	550e8400-e29b-41d4-a716-446655440000	REGULAR	84e58947-8bec-4435-8222-08a334b67295	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000060
817f74ce-72dc-460e-bc4a-b498abc8679b	550e8400-e29b-41d4-a716-446655440000	REGULAR	9c37a02c-6042-4471-8a33-ddce7e4f8e71	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000061
eec0dec0-f7b4-4361-a191-3ca3da6541a9	550e8400-e29b-41d4-a716-446655440000	REGULAR	3efe63d6-09d1-4cc2-b5a0-3d66ac329f21	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000062
b74c26af-888f-4b77-98ad-87136d0f7e5e	550e8400-e29b-41d4-a716-446655440000	REGULAR	5d7fcf73-527c-43b8-b39c-6e496ed8c623	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000063
2ca757ec-be43-42b4-8e1b-05cb4cc28ee8	550e8400-e29b-41d4-a716-446655440000	REGULAR	604380d6-6dd3-4af3-bf99-bf40990b4378	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000064
0f0aadb5-2e72-4c1a-96f4-4bddba149022	550e8400-e29b-41d4-a716-446655440000	REGULAR	5d7af970-d1de-49c5-b1d1-a651eace6e8a	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000065
1a219cfb-a43f-4c1e-93bb-38c3a5778681	550e8400-e29b-41d4-a716-446655440000	REGULAR	9bb1543d-49ea-4040-9286-6892bcd772db	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000066
730652c9-e7b9-40b9-97a3-12a60c4b55e6	550e8400-e29b-41d4-a716-446655440000	REGULAR	262e4305-ff67-430a-8641-e8745b51dcbd	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000067
9b736286-3677-4568-877a-48969143b9bf	550e8400-e29b-41d4-a716-446655440000	REGULAR	3137a262-9e18-41c5-ac5a-db955d5c3937	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000068
fd26ecfc-4029-4064-aa65-8d8186c09537	550e8400-e29b-41d4-a716-446655440000	REGULAR	f6b838fb-3fe2-425f-9099-aee64d69826f	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000069
b85d4edc-24aa-4c4c-b283-d115382c1fd3	550e8400-e29b-41d4-a716-446655440000	REGULAR	b2670708-d3e5-4a2c-a56e-d1d3bef8056c	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000070
d253fce0-0562-4755-9cbd-f57388f1bfe4	550e8400-e29b-41d4-a716-446655440000	REGULAR	b9b498d6-177e-4e16-b365-219127ed9193	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000071
69f99701-3b91-48b4-a223-d431ffd6d21a	550e8400-e29b-41d4-a716-446655440000	REGULAR	c23f4e38-d663-4423-8795-ce069460c581	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000072
6a1388e3-9af9-4b14-b917-e4647907001d	550e8400-e29b-41d4-a716-446655440000	REGULAR	d12ea2a1-e2c3-4e57-9c4e-26df7a92edc8	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000073
c76ecf69-82a2-4792-a7c3-80814b1e858e	550e8400-e29b-41d4-a716-446655440000	REGULAR	82b59778-dbe4-48f1-b808-16f66efe9cbb	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000074
73e12081-5e45-46a1-8670-a76a4375625e	550e8400-e29b-41d4-a716-446655440000	REGULAR	e34617a4-4784-4348-bbb2-e6785440f5c9	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000075
48cfd479-ef26-4d1a-89bd-c9b6fe168915	550e8400-e29b-41d4-a716-446655440000	REGULAR	515451b4-0968-4362-8c89-2d2db8f55358	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000076
60e3bfb6-73b9-43dd-8b4b-0838d2be4e2e	550e8400-e29b-41d4-a716-446655440000	REGULAR	e670ec79-3006-406b-8965-d6a04fdf3d1d	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000077
aae938ba-4572-4657-a9d6-cfd6bdff6410	550e8400-e29b-41d4-a716-446655440000	REGULAR	31b41811-74cf-43ee-9e3e-393e2bf6c43b	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000078
f6b4e54f-e23a-495c-b1b1-d14a5f9bbbfb	550e8400-e29b-41d4-a716-446655440000	REGULAR	1e555d23-84e7-4155-b6ea-eeebc0f950fa	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000079
2cb1bd34-db11-489e-8c64-c3fbf7ab8c5c	550e8400-e29b-41d4-a716-446655440000	REGULAR	984ecd62-b14b-4167-822e-66b6a107925f	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000080
4b3f75ed-1b5b-47ab-ac74-05d6bf4ae3b7	550e8400-e29b-41d4-a716-446655440000	REGULAR	f5bb18c9-73f9-459e-ae9e-1efd5e92fbce	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000081
633f969f-6745-4f52-a7fa-717e07d78b1d	550e8400-e29b-41d4-a716-446655440000	REGULAR	a570edc0-7a75-4d9c-9cfb-44d0711032da	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000082
f6e9105e-7af7-49f1-a8bd-f00130d725e2	550e8400-e29b-41d4-a716-446655440000	REGULAR	78fbac25-c969-4412-a082-91845d863525	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000083
08a06ddc-07d1-4a99-a9fa-7e8d3ee241b8	550e8400-e29b-41d4-a716-446655440000	REGULAR	1453cfad-140f-4dc4-951a-f522961e8cef	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000084
0a73f114-59d7-4c48-b0e1-01d8f3626efe	550e8400-e29b-41d4-a716-446655440000	REGULAR	81dc8b47-c41a-4674-8977-8d11a8a34cc9	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000085
3bd74846-9d76-4364-8b1b-72b3c3ba389b	550e8400-e29b-41d4-a716-446655440000	REGULAR	0ec108b4-dc5a-4ce4-81a0-c88352c284a1	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000086
96a39b35-9231-448f-9669-fe278f5ea1d4	550e8400-e29b-41d4-a716-446655440000	REGULAR	2094305e-c47b-4a1f-9eda-bdfbc2a3e46c	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000087
cf1301df-fea0-4e22-9dc3-20f098ffa3a2	550e8400-e29b-41d4-a716-446655440000	REGULAR	82ad6fc5-d24b-4fd2-83d6-6b416be7cb5b	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000088
1768ebb7-a7ef-4c03-8b12-2a8c1a78a8e5	550e8400-e29b-41d4-a716-446655440000	REGULAR	f014bb21-85ab-43bb-a81a-886021221913	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000089
d8f67fdc-b87c-44d2-b33e-5d027cbbc5d1	550e8400-e29b-41d4-a716-446655440000	REGULAR	b7df07cc-e25d-4c3c-9cc9-d55526d71cb9	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000090
1e325b16-6783-4a30-b72e-d93bfa0e48c8	550e8400-e29b-41d4-a716-446655440000	REGULAR	d043edda-b68c-4d62-8f25-85fec6a89287	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000091
34fb88fb-a620-4220-b567-a3f66853d69a	550e8400-e29b-41d4-a716-446655440000	REGULAR	18011b4a-caa1-462e-9169-9458c6671010	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000092
01177c44-9f0f-4d32-ba8f-f182efd9b2c5	550e8400-e29b-41d4-a716-446655440000	REGULAR	573d1af6-b740-49f2-aa2f-5ed1369d97fc	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000093
ac0e8ebd-b949-4c09-a21e-738c827acf7d	550e8400-e29b-41d4-a716-446655440000	REGULAR	3724ba11-50e0-4cb0-8741-70e427f9b1f5	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000094
e1653742-648e-4c70-9a59-b175f1a2c62d	550e8400-e29b-41d4-a716-446655440000	REGULAR	a38825e1-6341-4aa5-8d5c-2b0cbf45c335	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000095
d9216b0e-7a06-4d4a-8607-3bb43bf3c720	550e8400-e29b-41d4-a716-446655440000	REGULAR	129ddc29-ef25-44f9-8c0b-c2b20b67ee1a	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000096
c8b43de4-6bc1-4a5e-8380-4472312a956c	550e8400-e29b-41d4-a716-446655440000	REGULAR	c209f36a-328d-4603-8fdf-0d9308df3c04	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000097
fd321169-a8a0-4b71-b64e-03a3b3e1a24d	550e8400-e29b-41d4-a716-446655440000	REGULAR	675fc56b-de07-4a7d-8c9a-2055f04064ce	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000098
479d232b-04e4-4547-92c6-ea7fffbe40e9	550e8400-e29b-41d4-a716-446655440000	REGULAR	3a7a0aa6-9b9d-4b6a-b301-95b36b596d27	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.870653+00	2025-08-04 03:40:27.870653+00	1000099
b1199703-d7ca-49c5-9139-bf309798f448	550e8400-e29b-41d4-a716-446655440000	REGULAR	8733ff9f-3d82-4264-a76f-b75effc7f8d3	11111111-1111-1111-1111-111111111111	\N	ACTIVE	1.00	\N	2025-08-04 03:40:27.905365+00	2025-08-04 03:40:27.905365+00	9999999
\.


--
-- Data for Name: sync_monitoring; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.sync_monitoring (id, operation_type, entity_id, entity_data, sync_status, error_message, retry_count, created_at, updated_at, synced_at) FROM stdin;
1	CREATE	9690290b-4c8b-4494-b31f-f4a39c2e45c5	{"new_data": {"id": "9690290b-4c8b-4494-b31f-f4a39c2e45c5", "name": "测试同步组织_1754180219", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "ddb5a408-e9cc-486b-b5ce-a1b970003982", "unit_type": "DEPARTMENT", "created_at": "2025-08-03T00:16:59.091374+00:00", "updated_at": "2025-08-03T00:16:59.091374+00:00", "description": "测试同步机制", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-03T00:16:59.091374+00:00", "table_name": "organization_units"}	SUCCESS	\N	0	2025-08-03 00:16:59.091374+00	2025-08-03 00:17:01.218971+00	2025-08-03 00:17:01.218971+00
2	DELETE	9690290b-4c8b-4494-b31f-f4a39c2e45c5	{"old_data": {"id": "9690290b-4c8b-4494-b31f-f4a39c2e45c5", "name": "测试同步组织_1754180219", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "ddb5a408-e9cc-486b-b5ce-a1b970003982", "unit_type": "DEPARTMENT", "created_at": "2025-08-03T00:16:59.091374+00:00", "updated_at": "2025-08-03T00:16:59.091374+00:00", "description": "测试同步机制", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-03T00:17:01.222674+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-03 00:17:01.222674+00	2025-08-03 00:17:01.222674+00	\N
3	CREATE	621f6880-76f8-45d8-94f3-ac3811b2143f	{"new_data": {"id": "621f6880-76f8-45d8-94f3-ac3811b2143f", "name": "技术部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100000", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
4	CREATE	924be0e5-9da9-4174-b17d-263f85f5b1fe	{"new_data": {"id": "924be0e5-9da9-4174-b17d-263f85f5b1fe", "name": "产品部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100001", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
5	CREATE	885c2a42-c205-4023-8710-d5ae5655aae2	{"new_data": {"id": "885c2a42-c205-4023-8710-d5ae5655aae2", "name": "销售部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100002", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
6	CREATE	408eafeb-06cd-4e7a-9ecc-e78715626595	{"new_data": {"id": "408eafeb-06cd-4e7a-9ecc-e78715626595", "name": "人事部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100003", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
7	CREATE	77dd1a38-0852-4abe-8335-90dfb5e77983	{"new_data": {"id": "77dd1a38-0852-4abe-8335-90dfb5e77983", "name": "财务部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100004", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
8	CREATE	b8618d55-42db-4bc5-84e7-f4f1b2ad69b9	{"new_data": {"id": "b8618d55-42db-4bc5-84e7-f4f1b2ad69b9", "name": "技术部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100005", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
9	CREATE	8afed392-5967-4202-a153-dc8f628eb0ae	{"new_data": {"id": "8afed392-5967-4202-a153-dc8f628eb0ae", "name": "产品部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100006", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
10	CREATE	4808064d-d252-450f-b5d1-b07a8acc4d8b	{"new_data": {"id": "4808064d-d252-450f-b5d1-b07a8acc4d8b", "name": "销售部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100007", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
11	CREATE	67bb192b-8996-45fe-a13b-83a972cc68a5	{"new_data": {"id": "67bb192b-8996-45fe-a13b-83a972cc68a5", "name": "人事部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100008", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
12	CREATE	88ad70cb-691b-4d2b-8ef1-22d13994b4bc	{"new_data": {"id": "88ad70cb-691b-4d2b-8ef1-22d13994b4bc", "name": "财务部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100009", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
13	CREATE	f9573014-bedd-45c5-a742-8dd2bf7d22ff	{"new_data": {"id": "f9573014-bedd-45c5-a742-8dd2bf7d22ff", "name": "技术部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100010", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
14	CREATE	bf4da899-660f-4c0b-bdd0-c7cdda24eb02	{"new_data": {"id": "bf4da899-660f-4c0b-bdd0-c7cdda24eb02", "name": "产品部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100011", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
15	CREATE	c7b74a72-7b24-4176-9c11-031e59d14683	{"new_data": {"id": "c7b74a72-7b24-4176-9c11-031e59d14683", "name": "销售部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100012", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
16	CREATE	b531ca52-34ab-4fe8-b678-ae8a4c692941	{"new_data": {"id": "b531ca52-34ab-4fe8-b678-ae8a4c692941", "name": "人事部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100013", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
17	CREATE	03286230-0546-4775-9cdf-afa02d417f7d	{"new_data": {"id": "03286230-0546-4775-9cdf-afa02d417f7d", "name": "财务部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100014", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
18	CREATE	e1da0448-be36-42e2-af24-29129f809189	{"new_data": {"id": "e1da0448-be36-42e2-af24-29129f809189", "name": "技术部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100015", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
19	CREATE	60136276-6457-418b-9311-f90224c9a7ce	{"new_data": {"id": "60136276-6457-418b-9311-f90224c9a7ce", "name": "产品部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100016", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
20	CREATE	7263846f-ef0d-4050-b210-e1a3afd8e3ec	{"new_data": {"id": "7263846f-ef0d-4050-b210-e1a3afd8e3ec", "name": "销售部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100017", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
21	CREATE	66e6d75c-1238-4d7c-84c0-cfdbc0f89edd	{"new_data": {"id": "66e6d75c-1238-4d7c-84c0-cfdbc0f89edd", "name": "人事部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100018", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
22	CREATE	e9e04579-339b-4e70-81dc-4396dcbabbf9	{"new_data": {"id": "e9e04579-339b-4e70-81dc-4396dcbabbf9", "name": "财务部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100019", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
23	CREATE	7e41b3cd-8931-4c33-92e1-b356c83fa79e	{"new_data": {"id": "7e41b3cd-8931-4c33-92e1-b356c83fa79e", "name": "技术部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100020", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
24	CREATE	d10995e9-74f0-41be-8a26-78a23143a2ad	{"new_data": {"id": "d10995e9-74f0-41be-8a26-78a23143a2ad", "name": "产品部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100021", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
25	CREATE	54359b84-44c3-4615-a55a-dfa0e8677f65	{"new_data": {"id": "54359b84-44c3-4615-a55a-dfa0e8677f65", "name": "销售部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100022", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
26	CREATE	f1dc9253-39c3-4bcc-8001-37504a0fec22	{"new_data": {"id": "f1dc9253-39c3-4bcc-8001-37504a0fec22", "name": "人事部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100023", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
27	CREATE	a03b584a-922d-4b3f-a812-539f5c483e2d	{"new_data": {"id": "a03b584a-922d-4b3f-a812-539f5c483e2d", "name": "财务部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100024", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
28	CREATE	c32d523c-e237-4cd1-aa40-3e27e6695d92	{"new_data": {"id": "c32d523c-e237-4cd1-aa40-3e27e6695d92", "name": "技术部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100025", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
29	CREATE	bc299d05-d044-4a5f-bc3b-b87ca34e50d5	{"new_data": {"id": "bc299d05-d044-4a5f-bc3b-b87ca34e50d5", "name": "产品部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100026", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
30	CREATE	2bf97da8-3438-46d7-b12b-bd2aef1ddd71	{"new_data": {"id": "2bf97da8-3438-46d7-b12b-bd2aef1ddd71", "name": "销售部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100027", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
31	CREATE	0d91e416-67cd-4197-9b04-d935fea423bf	{"new_data": {"id": "0d91e416-67cd-4197-9b04-d935fea423bf", "name": "人事部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100028", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
32	CREATE	a79140e0-d004-4168-83a5-fc0230e76817	{"new_data": {"id": "a79140e0-d004-4168-83a5-fc0230e76817", "name": "财务部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100029", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
33	CREATE	b6d4fcf1-743c-4a8a-be21-063c183be6e0	{"new_data": {"id": "b6d4fcf1-743c-4a8a-be21-063c183be6e0", "name": "技术部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100030", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
34	CREATE	acd73f45-15ff-4a6e-af63-1e5fb442a22e	{"new_data": {"id": "acd73f45-15ff-4a6e-af63-1e5fb442a22e", "name": "产品部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100031", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
35	CREATE	6cb078c5-3823-4f7c-9191-41cdc984146c	{"new_data": {"id": "6cb078c5-3823-4f7c-9191-41cdc984146c", "name": "销售部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100032", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
36	CREATE	5adb1ca6-b19b-4682-8522-6ca3ee70713e	{"new_data": {"id": "5adb1ca6-b19b-4682-8522-6ca3ee70713e", "name": "人事部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100033", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
37	CREATE	674e398e-4c54-487d-8037-e355ff987296	{"new_data": {"id": "674e398e-4c54-487d-8037-e355ff987296", "name": "财务部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100034", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
38	CREATE	44192384-510f-4a02-9d22-a48d72d5e624	{"new_data": {"id": "44192384-510f-4a02-9d22-a48d72d5e624", "name": "技术部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100035", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
39	CREATE	9d23ee22-763e-472f-b42b-81f74a6f96b7	{"new_data": {"id": "9d23ee22-763e-472f-b42b-81f74a6f96b7", "name": "产品部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100036", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
40	CREATE	0816000a-5127-4858-8d87-ded8a78a45ba	{"new_data": {"id": "0816000a-5127-4858-8d87-ded8a78a45ba", "name": "销售部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100037", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
41	CREATE	75733cec-d0d7-4290-ab77-01434186db88	{"new_data": {"id": "75733cec-d0d7-4290-ab77-01434186db88", "name": "人事部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100038", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
42	CREATE	d56bce51-8e89-408c-8c3b-01ff91e659a6	{"new_data": {"id": "d56bce51-8e89-408c-8c3b-01ff91e659a6", "name": "财务部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100039", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
43	CREATE	c7eb7416-34d5-4633-bced-2cc627a880d6	{"new_data": {"id": "c7eb7416-34d5-4633-bced-2cc627a880d6", "name": "技术部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100040", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
44	CREATE	cf78fb4e-d774-4f93-acb4-15d814627040	{"new_data": {"id": "cf78fb4e-d774-4f93-acb4-15d814627040", "name": "产品部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100041", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
45	CREATE	d15a72c3-e2ab-47d6-b4c2-aed32902bca2	{"new_data": {"id": "d15a72c3-e2ab-47d6-b4c2-aed32902bca2", "name": "销售部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100042", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
46	CREATE	4cfbe8d6-67f7-430e-89ce-9fb58f9c975a	{"new_data": {"id": "4cfbe8d6-67f7-430e-89ce-9fb58f9c975a", "name": "人事部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100043", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
47	CREATE	141459e2-43ca-4b14-9cc4-3c5cc5afd959	{"new_data": {"id": "141459e2-43ca-4b14-9cc4-3c5cc5afd959", "name": "财务部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100044", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
48	CREATE	80ffda67-2958-4c57-9fe8-8203b0264ba7	{"new_data": {"id": "80ffda67-2958-4c57-9fe8-8203b0264ba7", "name": "技术部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100045", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
49	CREATE	f6fa2159-c52f-4b73-ac02-b80ad279d153	{"new_data": {"id": "f6fa2159-c52f-4b73-ac02-b80ad279d153", "name": "产品部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100046", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
50	CREATE	37b9b528-b899-4a57-85a8-a20a8ceec16a	{"new_data": {"id": "37b9b528-b899-4a57-85a8-a20a8ceec16a", "name": "销售部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100047", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
51	CREATE	ba88c966-a082-4085-a3ba-e57f0652fd70	{"new_data": {"id": "ba88c966-a082-4085-a3ba-e57f0652fd70", "name": "人事部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100048", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
52	CREATE	1dbed342-31d3-4258-be0a-3951cce1cec8	{"new_data": {"id": "1dbed342-31d3-4258-be0a-3951cce1cec8", "name": "财务部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100049", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:15:14.357115+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:15:14.357115+00	2025-08-04 03:15:14.357115+00	\N
53	CREATE	ae30d3c3-f4f3-42f5-9b1b-64e5a7482c4b	{"new_data": {"id": "ae30d3c3-f4f3-42f5-9b1b-64e5a7482c4b", "name": "边界组织999998", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:16:23.648038+00:00", "updated_at": "2025-08-04T03:16:23.648038+00:00", "business_id": "999998", "description": null, "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:16:23.648038+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:16:23.648038+00	2025-08-04 03:16:23.648038+00	\N
54	CREATE	28b98968-f122-41b9-b67c-e2ae126db2ec	{"new_data": {"id": "28b98968-f122-41b9-b67c-e2ae126db2ec", "name": "边界组织999999", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:16:23.648038+00:00", "updated_at": "2025-08-04T03:16:23.648038+00:00", "business_id": "999999", "description": null, "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:16:23.648038+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:16:23.648038+00	2025-08-04 03:16:23.648038+00	\N
55	DELETE	0816000a-5127-4858-8d87-ded8a78a45ba	{"old_data": {"id": "0816000a-5127-4858-8d87-ded8a78a45ba", "name": "销售部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100037", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
56	DELETE	75733cec-d0d7-4290-ab77-01434186db88	{"old_data": {"id": "75733cec-d0d7-4290-ab77-01434186db88", "name": "人事部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100038", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
57	DELETE	d56bce51-8e89-408c-8c3b-01ff91e659a6	{"old_data": {"id": "d56bce51-8e89-408c-8c3b-01ff91e659a6", "name": "财务部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100039", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
58	DELETE	5cfdb01d-9dcc-49f4-b9bd-4f43453520c5	{"old_data": {"id": "5cfdb01d-9dcc-49f4-b9bd-4f43453520c5", "name": "高谷集团", "level": 0, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "COMPANY", "created_at": "2025-08-02T14:24:56.454698+00:00", "updated_at": "2025-08-02T23:15:06.989562+00:00", "business_id": null, "description": "CQRS Phase 3 Real Database Test", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
59	DELETE	c7eb7416-34d5-4633-bced-2cc627a880d6	{"old_data": {"id": "c7eb7416-34d5-4633-bced-2cc627a880d6", "name": "技术部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100040", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
60	DELETE	2f86d7e2-742f-4a84-9ab0-5eb0f9d79ae6	{"old_data": {"id": "2f86d7e2-742f-4a84-9ab0-5eb0f9d79ae6", "name": "技术研发部", "level": 1, "status": "ACTIVE", "profile": {}, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-02T22:59:14.951084+00:00", "updated_at": "2025-08-02T23:15:06.989562+00:00", "business_id": null, "description": "负责公司核心技术研发、产品架构设计、技术创新和技术团队管理", "employee_count": 0, "parent_unit_id": "5cfdb01d-9dcc-49f4-b9bd-4f43453520c5"}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
61	DELETE	cf78fb4e-d774-4f93-acb4-15d814627040	{"old_data": {"id": "cf78fb4e-d774-4f93-acb4-15d814627040", "name": "产品部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100041", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
62	DELETE	d15a72c3-e2ab-47d6-b4c2-aed32902bca2	{"old_data": {"id": "d15a72c3-e2ab-47d6-b4c2-aed32902bca2", "name": "销售部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100042", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
63	DELETE	4cfbe8d6-67f7-430e-89ce-9fb58f9c975a	{"old_data": {"id": "4cfbe8d6-67f7-430e-89ce-9fb58f9c975a", "name": "人事部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100043", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
64	DELETE	5629c5e0-db37-4e0e-84bd-bf87e8523b38	{"old_data": {"id": "5629c5e0-db37-4e0e-84bd-bf87e8523b38", "name": "人力资源部", "level": 1, "status": "ACTIVE", "profile": {}, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-02T23:26:31.492055+00:00", "updated_at": "2025-08-02T23:26:31.492055+00:00", "business_id": null, "description": "负责人才招聘、员工培训、绩效管理、薪酬福利和企业文化建设", "employee_count": 0, "parent_unit_id": "5cfdb01d-9dcc-49f4-b9bd-4f43453520c5"}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
65	DELETE	b1f8ae08-b1d4-4e15-9e07-dc235ae27e15	{"old_data": {"id": "b1f8ae08-b1d4-4e15-9e07-dc235ae27e15", "name": "财务管理部", "level": 1, "status": "ACTIVE", "profile": {}, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-02T23:26:31.492055+00:00", "updated_at": "2025-08-02T23:26:31.492055+00:00", "business_id": null, "description": "负责财务规划、成本控制、资金管理、财务分析和合规审计", "employee_count": 0, "parent_unit_id": "5cfdb01d-9dcc-49f4-b9bd-4f43453520c5"}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
66	DELETE	e7e38f64-8da3-42a9-b478-4fe6043c35ce	{"old_data": {"id": "e7e38f64-8da3-42a9-b478-4fe6043c35ce", "name": "产品管理部", "level": 1, "status": "ACTIVE", "profile": {}, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-02T23:24:25.257656+00:00", "updated_at": "2025-08-02T23:27:03.790748+00:00", "business_id": null, "description": "负责公司产品规划、产品设计、用户体验、项目管理和产品运营", "employee_count": 0, "parent_unit_id": "5cfdb01d-9dcc-49f4-b9bd-4f43453520c5"}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
67	DELETE	75db946b-3138-4dd7-9145-33025409c185	{"old_data": {"id": "75db946b-3138-4dd7-9145-33025409c185", "name": "市场营销部", "level": 1, "status": "ACTIVE", "profile": {}, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-02T23:25:52.073502+00:00", "updated_at": "2025-08-02T23:27:03.790748+00:00", "business_id": null, "description": "负责市场推广、品牌建设、客户关系管理、销售支持和市场分析", "employee_count": 0, "parent_unit_id": "5cfdb01d-9dcc-49f4-b9bd-4f43453520c5"}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
68	DELETE	cc7cdb48-9e04-4c58-811a-1185daa43127	{"old_data": {"id": "cc7cdb48-9e04-4c58-811a-1185daa43127", "name": "前端开发组", "level": 2, "status": "ACTIVE", "profile": {}, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "PROJECT_TEAM", "created_at": "2025-08-02T23:28:05.256212+00:00", "updated_at": "2025-08-02T23:28:05.256212+00:00", "business_id": null, "description": "负责Web前端、移动端UI开发和用户体验优化", "employee_count": 0, "parent_unit_id": "2f86d7e2-742f-4a84-9ab0-5eb0f9d79ae6"}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
69	DELETE	3903647b-0a2b-4ec1-b31b-a5d9f3600ae8	{"old_data": {"id": "3903647b-0a2b-4ec1-b31b-a5d9f3600ae8", "name": "后端开发组", "level": 2, "status": "ACTIVE", "profile": {}, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "PROJECT_TEAM", "created_at": "2025-08-02T23:28:05.256212+00:00", "updated_at": "2025-08-02T23:28:05.256212+00:00", "business_id": null, "description": "负责服务端开发、API设计和数据库架构", "employee_count": 0, "parent_unit_id": "2f86d7e2-742f-4a84-9ab0-5eb0f9d79ae6"}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
70	DELETE	26f504e7-f7a2-41e9-9469-1e7487504800	{"old_data": {"id": "26f504e7-f7a2-41e9-9469-1e7487504800", "name": "测试质量组", "level": 2, "status": "ACTIVE", "profile": {}, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "PROJECT_TEAM", "created_at": "2025-08-02T23:28:05.256212+00:00", "updated_at": "2025-08-02T23:28:05.256212+00:00", "business_id": null, "description": "负责软件测试、质量保证和自动化测试", "employee_count": 0, "parent_unit_id": "2f86d7e2-742f-4a84-9ab0-5eb0f9d79ae6"}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
71	DELETE	2d1abfad-501d-4e69-b8e5-2a393264eae7	{"old_data": {"id": "2d1abfad-501d-4e69-b8e5-2a393264eae7", "name": "运维架构组", "level": 2, "status": "ACTIVE", "profile": {}, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "PROJECT_TEAM", "created_at": "2025-08-02T23:28:05.256212+00:00", "updated_at": "2025-08-02T23:28:05.256212+00:00", "business_id": null, "description": "负责系统运维、CI/CD和基础设施架构", "employee_count": 0, "parent_unit_id": "2f86d7e2-742f-4a84-9ab0-5eb0f9d79ae6"}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
72	DELETE	e370c5e9-e193-4cff-8944-5f76235e5d82	{"old_data": {"id": "e370c5e9-e193-4cff-8944-5f76235e5d82", "name": "产品规划组", "level": 2, "status": "ACTIVE", "profile": {}, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "PROJECT_TEAM", "created_at": "2025-08-02T23:28:05.256212+00:00", "updated_at": "2025-08-02T23:28:05.256212+00:00", "business_id": null, "description": "负责产品策略制定、需求分析和产品路线图", "employee_count": 0, "parent_unit_id": "e7e38f64-8da3-42a9-b478-4fe6043c35ce"}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
73	DELETE	6b1a55cd-4b90-4d9e-94a4-bee7d2c77bc4	{"old_data": {"id": "6b1a55cd-4b90-4d9e-94a4-bee7d2c77bc4", "name": "用户体验组", "level": 2, "status": "ACTIVE", "profile": {}, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "PROJECT_TEAM", "created_at": "2025-08-02T23:28:05.256212+00:00", "updated_at": "2025-08-02T23:28:05.256212+00:00", "business_id": null, "description": "负责用户研究、交互设计和界面设计", "employee_count": 0, "parent_unit_id": "e7e38f64-8da3-42a9-b478-4fe6043c35ce"}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
74	DELETE	6064c299-e1fb-4478-bd5d-037f9511d3be	{"old_data": {"id": "6064c299-e1fb-4478-bd5d-037f9511d3be", "name": "项目管理组", "level": 2, "status": "ACTIVE", "profile": {}, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "PROJECT_TEAM", "created_at": "2025-08-02T23:28:05.256212+00:00", "updated_at": "2025-08-02T23:28:05.256212+00:00", "business_id": null, "description": "负责项目协调、进度管理和资源调配", "employee_count": 0, "parent_unit_id": "e7e38f64-8da3-42a9-b478-4fe6043c35ce"}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
75	DELETE	a394c695-13ee-40ce-b052-a7569b6165c1	{"old_data": {"id": "a394c695-13ee-40ce-b052-a7569b6165c1", "name": "数据分析组", "level": 2, "status": "ACTIVE", "profile": {}, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "PROJECT_TEAM", "created_at": "2025-08-02T23:28:05.256212+00:00", "updated_at": "2025-08-02T23:28:05.256212+00:00", "business_id": null, "description": "负责用户行为分析、产品数据分析和业务洞察", "employee_count": 0, "parent_unit_id": "e7e38f64-8da3-42a9-b478-4fe6043c35ce"}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
76	DELETE	ab6dbacd-ac5c-4eef-90d2-590523d089f5	{"old_data": {"id": "ab6dbacd-ac5c-4eef-90d2-590523d089f5", "name": "品牌推广组", "level": 2, "status": "ACTIVE", "profile": {}, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "PROJECT_TEAM", "created_at": "2025-08-02T23:28:05.256212+00:00", "updated_at": "2025-08-02T23:28:05.256212+00:00", "business_id": null, "description": "负责品牌建设、公关活动和媒体合作", "employee_count": 0, "parent_unit_id": "75db946b-3138-4dd7-9145-33025409c185"}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
77	DELETE	56ad0e7a-75c8-435f-8c0e-1b5c54ce882f	{"old_data": {"id": "56ad0e7a-75c8-435f-8c0e-1b5c54ce882f", "name": "数字营销组", "level": 2, "status": "ACTIVE", "profile": {}, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "PROJECT_TEAM", "created_at": "2025-08-02T23:28:05.256212+00:00", "updated_at": "2025-08-02T23:28:05.256212+00:00", "business_id": null, "description": "负责线上推广、SEM/SEO和社交媒体营销", "employee_count": 0, "parent_unit_id": "75db946b-3138-4dd7-9145-33025409c185"}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
78	DELETE	141459e2-43ca-4b14-9cc4-3c5cc5afd959	{"old_data": {"id": "141459e2-43ca-4b14-9cc4-3c5cc5afd959", "name": "财务部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100044", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
79	DELETE	80ffda67-2958-4c57-9fe8-8203b0264ba7	{"old_data": {"id": "80ffda67-2958-4c57-9fe8-8203b0264ba7", "name": "技术部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100045", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
80	DELETE	f6fa2159-c52f-4b73-ac02-b80ad279d153	{"old_data": {"id": "f6fa2159-c52f-4b73-ac02-b80ad279d153", "name": "产品部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100046", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
81	DELETE	37b9b528-b899-4a57-85a8-a20a8ceec16a	{"old_data": {"id": "37b9b528-b899-4a57-85a8-a20a8ceec16a", "name": "销售部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100047", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
82	DELETE	ba88c966-a082-4085-a3ba-e57f0652fd70	{"old_data": {"id": "ba88c966-a082-4085-a3ba-e57f0652fd70", "name": "人事部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100048", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
83	DELETE	1dbed342-31d3-4258-be0a-3951cce1cec8	{"old_data": {"id": "1dbed342-31d3-4258-be0a-3951cce1cec8", "name": "财务部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100049", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
84	DELETE	69c93c6a-915e-4789-bbe3-78e19f06adca	{"old_data": {"id": "69c93c6a-915e-4789-bbe3-78e19f06adca", "name": "客户关系组", "level": 2, "status": "ACTIVE", "profile": {}, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "PROJECT_TEAM", "created_at": "2025-08-02T23:28:05.256212+00:00", "updated_at": "2025-08-02T23:28:05.256212+00:00", "business_id": null, "description": "负责客户维护、客户服务和客户满意度", "employee_count": 0, "parent_unit_id": "75db946b-3138-4dd7-9145-33025409c185"}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
85	DELETE	a5a90dee-fad4-45a5-a1d7-fbfb4939071a	{"old_data": {"id": "a5a90dee-fad4-45a5-a1d7-fbfb4939071a", "name": "销售支持组", "level": 2, "status": "ACTIVE", "profile": {}, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "PROJECT_TEAM", "created_at": "2025-08-02T23:28:05.256212+00:00", "updated_at": "2025-08-02T23:28:05.256212+00:00", "business_id": null, "description": "负责销售工具、销售培训和销售数据分析", "employee_count": 0, "parent_unit_id": "75db946b-3138-4dd7-9145-33025409c185"}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
86	DELETE	c330c171-96b4-4a3b-aa52-af70fff5d906	{"old_data": {"id": "c330c171-96b4-4a3b-aa52-af70fff5d906", "name": "招聘培训组", "level": 2, "status": "ACTIVE", "profile": {}, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "PROJECT_TEAM", "created_at": "2025-08-02T23:28:05.256212+00:00", "updated_at": "2025-08-02T23:28:05.256212+00:00", "business_id": null, "description": "负责人才招聘、入职培训和员工发展", "employee_count": 0, "parent_unit_id": "5629c5e0-db37-4e0e-84bd-bf87e8523b38"}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
87	DELETE	e8dc712a-5dde-4d46-8b7d-e402d2f916d1	{"old_data": {"id": "e8dc712a-5dde-4d46-8b7d-e402d2f916d1", "name": "绩效薪酬组", "level": 2, "status": "ACTIVE", "profile": {}, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "PROJECT_TEAM", "created_at": "2025-08-02T23:28:05.256212+00:00", "updated_at": "2025-08-02T23:28:05.256212+00:00", "business_id": null, "description": "负责绩效考核、薪酬管理和激励机制", "employee_count": 0, "parent_unit_id": "5629c5e0-db37-4e0e-84bd-bf87e8523b38"}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
88	DELETE	0495eeb4-d559-4eec-9a96-c495e3c5e4dd	{"old_data": {"id": "0495eeb4-d559-4eec-9a96-c495e3c5e4dd", "name": "员工关系组", "level": 2, "status": "ACTIVE", "profile": {}, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "PROJECT_TEAM", "created_at": "2025-08-02T23:28:05.256212+00:00", "updated_at": "2025-08-02T23:28:05.256212+00:00", "business_id": null, "description": "负责员工关怀、劳动关系和企业文化", "employee_count": 0, "parent_unit_id": "5629c5e0-db37-4e0e-84bd-bf87e8523b38"}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
89	DELETE	c359b9a0-4428-43c2-ac37-2e8810e6cad9	{"old_data": {"id": "c359b9a0-4428-43c2-ac37-2e8810e6cad9", "name": "人事行政组", "level": 2, "status": "ACTIVE", "profile": {}, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "PROJECT_TEAM", "created_at": "2025-08-02T23:28:05.256212+00:00", "updated_at": "2025-08-02T23:28:05.256212+00:00", "business_id": null, "description": "负责人事档案、考勤管理和行政支持", "employee_count": 0, "parent_unit_id": "5629c5e0-db37-4e0e-84bd-bf87e8523b38"}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
90	DELETE	581eda0b-f212-4ea0-b5f4-a1597fc55cb9	{"old_data": {"id": "581eda0b-f212-4ea0-b5f4-a1597fc55cb9", "name": "财务核算组", "level": 2, "status": "ACTIVE", "profile": {}, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "PROJECT_TEAM", "created_at": "2025-08-02T23:28:05.256212+00:00", "updated_at": "2025-08-02T23:28:05.256212+00:00", "business_id": null, "description": "负责日常记账、财务报表和税务申报", "employee_count": 0, "parent_unit_id": "b1f8ae08-b1d4-4e15-9e07-dc235ae27e15"}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
91	DELETE	a95762ad-394a-4216-857f-e572ab73a7f9	{"old_data": {"id": "a95762ad-394a-4216-857f-e572ab73a7f9", "name": "成本控制组", "level": 2, "status": "ACTIVE", "profile": {}, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "PROJECT_TEAM", "created_at": "2025-08-02T23:28:05.256212+00:00", "updated_at": "2025-08-02T23:28:05.256212+00:00", "business_id": null, "description": "负责成本分析、预算管理和费用控制", "employee_count": 0, "parent_unit_id": "b1f8ae08-b1d4-4e15-9e07-dc235ae27e15"}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
92	DELETE	83a619fc-0a00-44b3-992f-4355972ef2df	{"old_data": {"id": "83a619fc-0a00-44b3-992f-4355972ef2df", "name": "资金管理组", "level": 2, "status": "ACTIVE", "profile": {}, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "PROJECT_TEAM", "created_at": "2025-08-02T23:28:05.256212+00:00", "updated_at": "2025-08-02T23:28:05.256212+00:00", "business_id": null, "description": "负责现金流管理、投资决策和资金调配", "employee_count": 0, "parent_unit_id": "b1f8ae08-b1d4-4e15-9e07-dc235ae27e15"}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
93	DELETE	618bed72-dd46-4bd0-9a43-14e7c625b59c	{"old_data": {"id": "618bed72-dd46-4bd0-9a43-14e7c625b59c", "name": "审计风控组", "level": 2, "status": "ACTIVE", "profile": {}, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "PROJECT_TEAM", "created_at": "2025-08-02T23:28:05.256212+00:00", "updated_at": "2025-08-02T23:28:05.256212+00:00", "business_id": null, "description": "负责内部审计、风险控制和合规监督", "employee_count": 0, "parent_unit_id": "b1f8ae08-b1d4-4e15-9e07-dc235ae27e15"}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
94	DELETE	621f6880-76f8-45d8-94f3-ac3811b2143f	{"old_data": {"id": "621f6880-76f8-45d8-94f3-ac3811b2143f", "name": "技术部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100000", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
95	DELETE	924be0e5-9da9-4174-b17d-263f85f5b1fe	{"old_data": {"id": "924be0e5-9da9-4174-b17d-263f85f5b1fe", "name": "产品部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100001", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
96	DELETE	885c2a42-c205-4023-8710-d5ae5655aae2	{"old_data": {"id": "885c2a42-c205-4023-8710-d5ae5655aae2", "name": "销售部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100002", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
97	DELETE	408eafeb-06cd-4e7a-9ecc-e78715626595	{"old_data": {"id": "408eafeb-06cd-4e7a-9ecc-e78715626595", "name": "人事部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100003", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
98	DELETE	77dd1a38-0852-4abe-8335-90dfb5e77983	{"old_data": {"id": "77dd1a38-0852-4abe-8335-90dfb5e77983", "name": "财务部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100004", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
99	DELETE	b8618d55-42db-4bc5-84e7-f4f1b2ad69b9	{"old_data": {"id": "b8618d55-42db-4bc5-84e7-f4f1b2ad69b9", "name": "技术部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100005", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
100	DELETE	8afed392-5967-4202-a153-dc8f628eb0ae	{"old_data": {"id": "8afed392-5967-4202-a153-dc8f628eb0ae", "name": "产品部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100006", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
101	DELETE	4808064d-d252-450f-b5d1-b07a8acc4d8b	{"old_data": {"id": "4808064d-d252-450f-b5d1-b07a8acc4d8b", "name": "销售部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100007", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
102	DELETE	67bb192b-8996-45fe-a13b-83a972cc68a5	{"old_data": {"id": "67bb192b-8996-45fe-a13b-83a972cc68a5", "name": "人事部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100008", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
103	DELETE	88ad70cb-691b-4d2b-8ef1-22d13994b4bc	{"old_data": {"id": "88ad70cb-691b-4d2b-8ef1-22d13994b4bc", "name": "财务部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100009", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
104	DELETE	f9573014-bedd-45c5-a742-8dd2bf7d22ff	{"old_data": {"id": "f9573014-bedd-45c5-a742-8dd2bf7d22ff", "name": "技术部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100010", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
105	DELETE	bf4da899-660f-4c0b-bdd0-c7cdda24eb02	{"old_data": {"id": "bf4da899-660f-4c0b-bdd0-c7cdda24eb02", "name": "产品部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100011", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
106	DELETE	c7b74a72-7b24-4176-9c11-031e59d14683	{"old_data": {"id": "c7b74a72-7b24-4176-9c11-031e59d14683", "name": "销售部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100012", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
107	DELETE	b531ca52-34ab-4fe8-b678-ae8a4c692941	{"old_data": {"id": "b531ca52-34ab-4fe8-b678-ae8a4c692941", "name": "人事部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100013", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
108	DELETE	03286230-0546-4775-9cdf-afa02d417f7d	{"old_data": {"id": "03286230-0546-4775-9cdf-afa02d417f7d", "name": "财务部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100014", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
109	DELETE	e1da0448-be36-42e2-af24-29129f809189	{"old_data": {"id": "e1da0448-be36-42e2-af24-29129f809189", "name": "技术部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100015", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
110	DELETE	60136276-6457-418b-9311-f90224c9a7ce	{"old_data": {"id": "60136276-6457-418b-9311-f90224c9a7ce", "name": "产品部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100016", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
111	DELETE	7263846f-ef0d-4050-b210-e1a3afd8e3ec	{"old_data": {"id": "7263846f-ef0d-4050-b210-e1a3afd8e3ec", "name": "销售部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100017", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
112	DELETE	66e6d75c-1238-4d7c-84c0-cfdbc0f89edd	{"old_data": {"id": "66e6d75c-1238-4d7c-84c0-cfdbc0f89edd", "name": "人事部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100018", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
113	DELETE	e9e04579-339b-4e70-81dc-4396dcbabbf9	{"old_data": {"id": "e9e04579-339b-4e70-81dc-4396dcbabbf9", "name": "财务部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100019", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
114	DELETE	7e41b3cd-8931-4c33-92e1-b356c83fa79e	{"old_data": {"id": "7e41b3cd-8931-4c33-92e1-b356c83fa79e", "name": "技术部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100020", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
115	DELETE	d10995e9-74f0-41be-8a26-78a23143a2ad	{"old_data": {"id": "d10995e9-74f0-41be-8a26-78a23143a2ad", "name": "产品部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100021", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
116	DELETE	54359b84-44c3-4615-a55a-dfa0e8677f65	{"old_data": {"id": "54359b84-44c3-4615-a55a-dfa0e8677f65", "name": "销售部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100022", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
117	DELETE	f1dc9253-39c3-4bcc-8001-37504a0fec22	{"old_data": {"id": "f1dc9253-39c3-4bcc-8001-37504a0fec22", "name": "人事部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100023", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
118	DELETE	a03b584a-922d-4b3f-a812-539f5c483e2d	{"old_data": {"id": "a03b584a-922d-4b3f-a812-539f5c483e2d", "name": "财务部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100024", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
119	DELETE	c32d523c-e237-4cd1-aa40-3e27e6695d92	{"old_data": {"id": "c32d523c-e237-4cd1-aa40-3e27e6695d92", "name": "技术部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100025", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
120	DELETE	bc299d05-d044-4a5f-bc3b-b87ca34e50d5	{"old_data": {"id": "bc299d05-d044-4a5f-bc3b-b87ca34e50d5", "name": "产品部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100026", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
121	DELETE	2bf97da8-3438-46d7-b12b-bd2aef1ddd71	{"old_data": {"id": "2bf97da8-3438-46d7-b12b-bd2aef1ddd71", "name": "销售部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100027", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
122	DELETE	0d91e416-67cd-4197-9b04-d935fea423bf	{"old_data": {"id": "0d91e416-67cd-4197-9b04-d935fea423bf", "name": "人事部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100028", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
123	DELETE	a79140e0-d004-4168-83a5-fc0230e76817	{"old_data": {"id": "a79140e0-d004-4168-83a5-fc0230e76817", "name": "财务部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100029", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
124	DELETE	b6d4fcf1-743c-4a8a-be21-063c183be6e0	{"old_data": {"id": "b6d4fcf1-743c-4a8a-be21-063c183be6e0", "name": "技术部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100030", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
125	DELETE	acd73f45-15ff-4a6e-af63-1e5fb442a22e	{"old_data": {"id": "acd73f45-15ff-4a6e-af63-1e5fb442a22e", "name": "产品部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100031", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
126	DELETE	6cb078c5-3823-4f7c-9191-41cdc984146c	{"old_data": {"id": "6cb078c5-3823-4f7c-9191-41cdc984146c", "name": "销售部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100032", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
127	DELETE	5adb1ca6-b19b-4682-8522-6ca3ee70713e	{"old_data": {"id": "5adb1ca6-b19b-4682-8522-6ca3ee70713e", "name": "人事部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100033", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
128	DELETE	674e398e-4c54-487d-8037-e355ff987296	{"old_data": {"id": "674e398e-4c54-487d-8037-e355ff987296", "name": "财务部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100034", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
129	DELETE	44192384-510f-4a02-9d22-a48d72d5e624	{"old_data": {"id": "44192384-510f-4a02-9d22-a48d72d5e624", "name": "技术部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100035", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
130	DELETE	9d23ee22-763e-472f-b42b-81f74a6f96b7	{"old_data": {"id": "9d23ee22-763e-472f-b42b-81f74a6f96b7", "name": "产品部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:15:14.357115+00:00", "updated_at": "2025-08-04T03:15:14.357115+00:00", "business_id": "100036", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
131	DELETE	ae30d3c3-f4f3-42f5-9b1b-64e5a7482c4b	{"old_data": {"id": "ae30d3c3-f4f3-42f5-9b1b-64e5a7482c4b", "name": "边界组织999998", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:16:23.648038+00:00", "updated_at": "2025-08-04T03:16:23.648038+00:00", "business_id": "999998", "description": null, "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
132	DELETE	28b98968-f122-41b9-b67c-e2ae126db2ec	{"old_data": {"id": "28b98968-f122-41b9-b67c-e2ae126db2ec", "name": "边界组织999999", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:16:23.648038+00:00", "updated_at": "2025-08-04T03:16:23.648038+00:00", "business_id": "999999", "description": null, "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:27:15.020764+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:27:15.020764+00	2025-08-04 03:27:15.020764+00	\N
133	CREATE	9b03a1fe-5ea6-4b60-830b-2be95aa96b1a	{"new_data": {"id": "9b03a1fe-5ea6-4b60-830b-2be95aa96b1a", "name": "技术部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100000", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
134	CREATE	9217706f-3a49-4d8c-bdc6-7c168ee8a0c1	{"new_data": {"id": "9217706f-3a49-4d8c-bdc6-7c168ee8a0c1", "name": "产品部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100001", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
135	CREATE	0cc7139f-3cbb-4251-8558-e0928a2f14a6	{"new_data": {"id": "0cc7139f-3cbb-4251-8558-e0928a2f14a6", "name": "销售部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100002", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
136	CREATE	7fcf1329-d2de-44ac-9b36-2cf0980f69f9	{"new_data": {"id": "7fcf1329-d2de-44ac-9b36-2cf0980f69f9", "name": "人事部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100003", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
137	CREATE	c70c0fa6-32c1-4c51-9d2e-ea2b5f208e33	{"new_data": {"id": "c70c0fa6-32c1-4c51-9d2e-ea2b5f208e33", "name": "财务部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100004", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
138	CREATE	e659aa75-7d1c-49f4-a849-778a351c32aa	{"new_data": {"id": "e659aa75-7d1c-49f4-a849-778a351c32aa", "name": "技术部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100005", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
139	CREATE	29e79e25-2c8e-4f2e-ab40-7769aad5299f	{"new_data": {"id": "29e79e25-2c8e-4f2e-ab40-7769aad5299f", "name": "产品部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100006", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
140	CREATE	771b676e-b37a-4370-bdb4-238f135e82c1	{"new_data": {"id": "771b676e-b37a-4370-bdb4-238f135e82c1", "name": "销售部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100007", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
141	CREATE	005badb6-54d6-4bb6-8e19-d9ce253d933c	{"new_data": {"id": "005badb6-54d6-4bb6-8e19-d9ce253d933c", "name": "人事部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100008", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
142	CREATE	04027ca2-272b-4dcb-83d9-60ab6885f241	{"new_data": {"id": "04027ca2-272b-4dcb-83d9-60ab6885f241", "name": "财务部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100009", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
143	CREATE	edccce0f-5b90-43c8-a9c1-f2f56d50b951	{"new_data": {"id": "edccce0f-5b90-43c8-a9c1-f2f56d50b951", "name": "技术部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100010", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
144	CREATE	004300fc-0cf5-4ff3-861e-0026ae859c21	{"new_data": {"id": "004300fc-0cf5-4ff3-861e-0026ae859c21", "name": "产品部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100011", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
145	CREATE	c8a192a0-c873-46fc-93ab-35330c243573	{"new_data": {"id": "c8a192a0-c873-46fc-93ab-35330c243573", "name": "销售部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100012", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
146	CREATE	a5405662-39ef-4f28-a805-5d2461ae6c67	{"new_data": {"id": "a5405662-39ef-4f28-a805-5d2461ae6c67", "name": "人事部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100013", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
147	CREATE	ef501aa0-508a-4daf-9b51-da629f12ab9a	{"new_data": {"id": "ef501aa0-508a-4daf-9b51-da629f12ab9a", "name": "财务部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100014", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
148	CREATE	112a04c4-e33d-400a-9bba-f7fb5b28c0e4	{"new_data": {"id": "112a04c4-e33d-400a-9bba-f7fb5b28c0e4", "name": "技术部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100015", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
149	CREATE	e4cfd40d-8224-4739-8ead-0046730aee0d	{"new_data": {"id": "e4cfd40d-8224-4739-8ead-0046730aee0d", "name": "产品部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100016", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
150	CREATE	9e9e137f-8be5-41be-9405-ad5d23f556bd	{"new_data": {"id": "9e9e137f-8be5-41be-9405-ad5d23f556bd", "name": "销售部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100017", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
151	CREATE	c11afa55-8921-46a1-b1c7-16b995b2436b	{"new_data": {"id": "c11afa55-8921-46a1-b1c7-16b995b2436b", "name": "人事部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100018", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
152	CREATE	c1b99353-ea94-4a1c-82f3-e0b7a2065dc7	{"new_data": {"id": "c1b99353-ea94-4a1c-82f3-e0b7a2065dc7", "name": "财务部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100019", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
153	CREATE	113f8ed7-28a0-4800-9b03-7d9694d9147c	{"new_data": {"id": "113f8ed7-28a0-4800-9b03-7d9694d9147c", "name": "技术部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100020", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
154	CREATE	a39f68f1-c173-4e69-9b15-50997d12a797	{"new_data": {"id": "a39f68f1-c173-4e69-9b15-50997d12a797", "name": "产品部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100021", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
155	CREATE	761644aa-5912-42e5-8920-aafba932a1ae	{"new_data": {"id": "761644aa-5912-42e5-8920-aafba932a1ae", "name": "销售部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100022", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
156	CREATE	2343dbb2-e971-473f-a5b2-f631e37f7b8b	{"new_data": {"id": "2343dbb2-e971-473f-a5b2-f631e37f7b8b", "name": "人事部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100023", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
157	CREATE	f9b2476a-a863-475f-a16e-4186e8e96432	{"new_data": {"id": "f9b2476a-a863-475f-a16e-4186e8e96432", "name": "财务部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100024", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
158	CREATE	fae833c6-bf98-44d9-b24c-0d605c1419fe	{"new_data": {"id": "fae833c6-bf98-44d9-b24c-0d605c1419fe", "name": "技术部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100025", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
159	CREATE	45db32ba-2e75-4895-a786-74e2e1a46e1f	{"new_data": {"id": "45db32ba-2e75-4895-a786-74e2e1a46e1f", "name": "产品部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100026", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
160	CREATE	3b3b2ff9-4713-441f-b292-ec2c41f59c44	{"new_data": {"id": "3b3b2ff9-4713-441f-b292-ec2c41f59c44", "name": "销售部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100027", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
161	CREATE	9ea190c9-2ee1-4d27-b4fd-01a0f7bf9d25	{"new_data": {"id": "9ea190c9-2ee1-4d27-b4fd-01a0f7bf9d25", "name": "人事部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100028", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
162	CREATE	3f16c7d3-b636-4516-a312-94f0234b9558	{"new_data": {"id": "3f16c7d3-b636-4516-a312-94f0234b9558", "name": "财务部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100029", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
163	CREATE	f9b83c1a-784f-4fce-bde1-83bbe1543a20	{"new_data": {"id": "f9b83c1a-784f-4fce-bde1-83bbe1543a20", "name": "技术部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100030", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
164	CREATE	cccb4a22-6a7f-45de-ab16-d25a006ee3e9	{"new_data": {"id": "cccb4a22-6a7f-45de-ab16-d25a006ee3e9", "name": "产品部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100031", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
165	CREATE	00e8f70d-0abc-46c2-957b-e4bd6cf15703	{"new_data": {"id": "00e8f70d-0abc-46c2-957b-e4bd6cf15703", "name": "销售部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100032", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
166	CREATE	292b18f7-1216-483b-9b76-c33d4716748f	{"new_data": {"id": "292b18f7-1216-483b-9b76-c33d4716748f", "name": "人事部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100033", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
167	CREATE	01763dce-d0e5-488f-8f8e-8d4a43e4a5ea	{"new_data": {"id": "01763dce-d0e5-488f-8f8e-8d4a43e4a5ea", "name": "财务部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100034", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
168	CREATE	ce083058-c85f-48c2-b74b-2f334e128316	{"new_data": {"id": "ce083058-c85f-48c2-b74b-2f334e128316", "name": "技术部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100035", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
169	CREATE	6f7b8b46-987d-4066-b708-ebc020cadbd4	{"new_data": {"id": "6f7b8b46-987d-4066-b708-ebc020cadbd4", "name": "产品部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100036", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
170	CREATE	92f7e3a6-78d6-487e-b0b9-5c76c9b613f4	{"new_data": {"id": "92f7e3a6-78d6-487e-b0b9-5c76c9b613f4", "name": "销售部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100037", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
171	CREATE	5c2a26e8-fbe7-4051-9ca8-a01ed0990189	{"new_data": {"id": "5c2a26e8-fbe7-4051-9ca8-a01ed0990189", "name": "人事部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100038", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
172	CREATE	19d05916-3f84-4229-8eae-f0b57b63635b	{"new_data": {"id": "19d05916-3f84-4229-8eae-f0b57b63635b", "name": "财务部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100039", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
173	CREATE	7c0a94b8-3957-4521-a755-c094ec56d153	{"new_data": {"id": "7c0a94b8-3957-4521-a755-c094ec56d153", "name": "技术部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100040", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
174	CREATE	ea65aed8-d569-4a24-b149-4cae9014dba9	{"new_data": {"id": "ea65aed8-d569-4a24-b149-4cae9014dba9", "name": "产品部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100041", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
175	CREATE	99eaf937-fb6c-4744-b679-8eaa02eccf51	{"new_data": {"id": "99eaf937-fb6c-4744-b679-8eaa02eccf51", "name": "销售部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100042", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
176	CREATE	3de462f8-2c99-4807-921a-b31b26dc1470	{"new_data": {"id": "3de462f8-2c99-4807-921a-b31b26dc1470", "name": "人事部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100043", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
177	CREATE	38db7e77-c519-49f0-bc4e-a41257c65cba	{"new_data": {"id": "38db7e77-c519-49f0-bc4e-a41257c65cba", "name": "财务部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100044", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
178	CREATE	0461a857-b08d-4341-b2de-c79fc14d529c	{"new_data": {"id": "0461a857-b08d-4341-b2de-c79fc14d529c", "name": "技术部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100045", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
179	CREATE	de4f4ae5-d53d-4ff9-b810-94b6f13c09fc	{"new_data": {"id": "de4f4ae5-d53d-4ff9-b810-94b6f13c09fc", "name": "产品部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100046", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
180	CREATE	d431f9e0-8fd9-4a90-836c-b75e4f735ca7	{"new_data": {"id": "d431f9e0-8fd9-4a90-836c-b75e4f735ca7", "name": "销售部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100047", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
181	CREATE	ee8b2e30-ce71-435c-975d-ca3be78094f2	{"new_data": {"id": "ee8b2e30-ce71-435c-975d-ca3be78094f2", "name": "人事部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100048", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
182	CREATE	bba4cdbc-c8d8-4987-b2ed-bf447d02e415	{"new_data": {"id": "bba4cdbc-c8d8-4987-b2ed-bf447d02e415", "name": "财务部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100049", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.160792+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.160792+00	2025-08-04 03:29:52.160792+00	\N
183	CREATE	d2fa55ee-00bd-41cc-b4d1-57f7564661a9	{"new_data": {"id": "d2fa55ee-00bd-41cc-b4d1-57f7564661a9", "name": "边界组织999999", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.20178+00:00", "updated_at": "2025-08-04T03:29:52.20178+00:00", "business_id": "999999", "description": null, "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:29:52.20178+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:29:52.20178+00	2025-08-04 03:29:52.20178+00	\N
184	DELETE	9b03a1fe-5ea6-4b60-830b-2be95aa96b1a	{"old_data": {"id": "9b03a1fe-5ea6-4b60-830b-2be95aa96b1a", "name": "技术部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100000", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
185	DELETE	9217706f-3a49-4d8c-bdc6-7c168ee8a0c1	{"old_data": {"id": "9217706f-3a49-4d8c-bdc6-7c168ee8a0c1", "name": "产品部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100001", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
186	DELETE	0cc7139f-3cbb-4251-8558-e0928a2f14a6	{"old_data": {"id": "0cc7139f-3cbb-4251-8558-e0928a2f14a6", "name": "销售部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100002", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
187	DELETE	7fcf1329-d2de-44ac-9b36-2cf0980f69f9	{"old_data": {"id": "7fcf1329-d2de-44ac-9b36-2cf0980f69f9", "name": "人事部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100003", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
188	DELETE	c70c0fa6-32c1-4c51-9d2e-ea2b5f208e33	{"old_data": {"id": "c70c0fa6-32c1-4c51-9d2e-ea2b5f208e33", "name": "财务部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100004", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
189	DELETE	e659aa75-7d1c-49f4-a849-778a351c32aa	{"old_data": {"id": "e659aa75-7d1c-49f4-a849-778a351c32aa", "name": "技术部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100005", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
190	DELETE	29e79e25-2c8e-4f2e-ab40-7769aad5299f	{"old_data": {"id": "29e79e25-2c8e-4f2e-ab40-7769aad5299f", "name": "产品部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100006", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
191	DELETE	771b676e-b37a-4370-bdb4-238f135e82c1	{"old_data": {"id": "771b676e-b37a-4370-bdb4-238f135e82c1", "name": "销售部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100007", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
192	DELETE	005badb6-54d6-4bb6-8e19-d9ce253d933c	{"old_data": {"id": "005badb6-54d6-4bb6-8e19-d9ce253d933c", "name": "人事部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100008", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
193	DELETE	04027ca2-272b-4dcb-83d9-60ab6885f241	{"old_data": {"id": "04027ca2-272b-4dcb-83d9-60ab6885f241", "name": "财务部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100009", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
194	DELETE	edccce0f-5b90-43c8-a9c1-f2f56d50b951	{"old_data": {"id": "edccce0f-5b90-43c8-a9c1-f2f56d50b951", "name": "技术部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100010", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
195	DELETE	004300fc-0cf5-4ff3-861e-0026ae859c21	{"old_data": {"id": "004300fc-0cf5-4ff3-861e-0026ae859c21", "name": "产品部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100011", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
196	DELETE	c8a192a0-c873-46fc-93ab-35330c243573	{"old_data": {"id": "c8a192a0-c873-46fc-93ab-35330c243573", "name": "销售部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100012", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
197	DELETE	a5405662-39ef-4f28-a805-5d2461ae6c67	{"old_data": {"id": "a5405662-39ef-4f28-a805-5d2461ae6c67", "name": "人事部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100013", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
198	DELETE	ef501aa0-508a-4daf-9b51-da629f12ab9a	{"old_data": {"id": "ef501aa0-508a-4daf-9b51-da629f12ab9a", "name": "财务部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100014", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
199	DELETE	112a04c4-e33d-400a-9bba-f7fb5b28c0e4	{"old_data": {"id": "112a04c4-e33d-400a-9bba-f7fb5b28c0e4", "name": "技术部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100015", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
200	DELETE	e4cfd40d-8224-4739-8ead-0046730aee0d	{"old_data": {"id": "e4cfd40d-8224-4739-8ead-0046730aee0d", "name": "产品部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100016", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
201	DELETE	9e9e137f-8be5-41be-9405-ad5d23f556bd	{"old_data": {"id": "9e9e137f-8be5-41be-9405-ad5d23f556bd", "name": "销售部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100017", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
202	DELETE	c11afa55-8921-46a1-b1c7-16b995b2436b	{"old_data": {"id": "c11afa55-8921-46a1-b1c7-16b995b2436b", "name": "人事部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100018", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
203	DELETE	c1b99353-ea94-4a1c-82f3-e0b7a2065dc7	{"old_data": {"id": "c1b99353-ea94-4a1c-82f3-e0b7a2065dc7", "name": "财务部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100019", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
204	DELETE	113f8ed7-28a0-4800-9b03-7d9694d9147c	{"old_data": {"id": "113f8ed7-28a0-4800-9b03-7d9694d9147c", "name": "技术部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100020", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
205	DELETE	a39f68f1-c173-4e69-9b15-50997d12a797	{"old_data": {"id": "a39f68f1-c173-4e69-9b15-50997d12a797", "name": "产品部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100021", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
206	DELETE	761644aa-5912-42e5-8920-aafba932a1ae	{"old_data": {"id": "761644aa-5912-42e5-8920-aafba932a1ae", "name": "销售部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100022", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
207	DELETE	2343dbb2-e971-473f-a5b2-f631e37f7b8b	{"old_data": {"id": "2343dbb2-e971-473f-a5b2-f631e37f7b8b", "name": "人事部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100023", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
208	DELETE	f9b2476a-a863-475f-a16e-4186e8e96432	{"old_data": {"id": "f9b2476a-a863-475f-a16e-4186e8e96432", "name": "财务部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100024", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
209	DELETE	fae833c6-bf98-44d9-b24c-0d605c1419fe	{"old_data": {"id": "fae833c6-bf98-44d9-b24c-0d605c1419fe", "name": "技术部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100025", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
210	DELETE	45db32ba-2e75-4895-a786-74e2e1a46e1f	{"old_data": {"id": "45db32ba-2e75-4895-a786-74e2e1a46e1f", "name": "产品部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100026", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
211	DELETE	3b3b2ff9-4713-441f-b292-ec2c41f59c44	{"old_data": {"id": "3b3b2ff9-4713-441f-b292-ec2c41f59c44", "name": "销售部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100027", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
212	DELETE	9ea190c9-2ee1-4d27-b4fd-01a0f7bf9d25	{"old_data": {"id": "9ea190c9-2ee1-4d27-b4fd-01a0f7bf9d25", "name": "人事部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100028", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
213	DELETE	3f16c7d3-b636-4516-a312-94f0234b9558	{"old_data": {"id": "3f16c7d3-b636-4516-a312-94f0234b9558", "name": "财务部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100029", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
214	DELETE	f9b83c1a-784f-4fce-bde1-83bbe1543a20	{"old_data": {"id": "f9b83c1a-784f-4fce-bde1-83bbe1543a20", "name": "技术部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100030", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
215	DELETE	cccb4a22-6a7f-45de-ab16-d25a006ee3e9	{"old_data": {"id": "cccb4a22-6a7f-45de-ab16-d25a006ee3e9", "name": "产品部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100031", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
216	DELETE	00e8f70d-0abc-46c2-957b-e4bd6cf15703	{"old_data": {"id": "00e8f70d-0abc-46c2-957b-e4bd6cf15703", "name": "销售部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100032", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
217	DELETE	292b18f7-1216-483b-9b76-c33d4716748f	{"old_data": {"id": "292b18f7-1216-483b-9b76-c33d4716748f", "name": "人事部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100033", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
218	DELETE	01763dce-d0e5-488f-8f8e-8d4a43e4a5ea	{"old_data": {"id": "01763dce-d0e5-488f-8f8e-8d4a43e4a5ea", "name": "财务部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100034", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
219	DELETE	ce083058-c85f-48c2-b74b-2f334e128316	{"old_data": {"id": "ce083058-c85f-48c2-b74b-2f334e128316", "name": "技术部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100035", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
220	DELETE	6f7b8b46-987d-4066-b708-ebc020cadbd4	{"old_data": {"id": "6f7b8b46-987d-4066-b708-ebc020cadbd4", "name": "产品部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100036", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
221	DELETE	92f7e3a6-78d6-487e-b0b9-5c76c9b613f4	{"old_data": {"id": "92f7e3a6-78d6-487e-b0b9-5c76c9b613f4", "name": "销售部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100037", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
222	DELETE	5c2a26e8-fbe7-4051-9ca8-a01ed0990189	{"old_data": {"id": "5c2a26e8-fbe7-4051-9ca8-a01ed0990189", "name": "人事部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100038", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
223	DELETE	19d05916-3f84-4229-8eae-f0b57b63635b	{"old_data": {"id": "19d05916-3f84-4229-8eae-f0b57b63635b", "name": "财务部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100039", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
224	DELETE	7c0a94b8-3957-4521-a755-c094ec56d153	{"old_data": {"id": "7c0a94b8-3957-4521-a755-c094ec56d153", "name": "技术部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100040", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
225	DELETE	ea65aed8-d569-4a24-b149-4cae9014dba9	{"old_data": {"id": "ea65aed8-d569-4a24-b149-4cae9014dba9", "name": "产品部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100041", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
226	DELETE	99eaf937-fb6c-4744-b679-8eaa02eccf51	{"old_data": {"id": "99eaf937-fb6c-4744-b679-8eaa02eccf51", "name": "销售部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100042", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
227	DELETE	3de462f8-2c99-4807-921a-b31b26dc1470	{"old_data": {"id": "3de462f8-2c99-4807-921a-b31b26dc1470", "name": "人事部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100043", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
228	DELETE	38db7e77-c519-49f0-bc4e-a41257c65cba	{"old_data": {"id": "38db7e77-c519-49f0-bc4e-a41257c65cba", "name": "财务部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100044", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
229	DELETE	0461a857-b08d-4341-b2de-c79fc14d529c	{"old_data": {"id": "0461a857-b08d-4341-b2de-c79fc14d529c", "name": "技术部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100045", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
230	DELETE	de4f4ae5-d53d-4ff9-b810-94b6f13c09fc	{"old_data": {"id": "de4f4ae5-d53d-4ff9-b810-94b6f13c09fc", "name": "产品部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100046", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
231	DELETE	d431f9e0-8fd9-4a90-836c-b75e4f735ca7	{"old_data": {"id": "d431f9e0-8fd9-4a90-836c-b75e4f735ca7", "name": "销售部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100047", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
232	DELETE	ee8b2e30-ce71-435c-975d-ca3be78094f2	{"old_data": {"id": "ee8b2e30-ce71-435c-975d-ca3be78094f2", "name": "人事部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100048", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
233	DELETE	bba4cdbc-c8d8-4987-b2ed-bf447d02e415	{"old_data": {"id": "bba4cdbc-c8d8-4987-b2ed-bf447d02e415", "name": "财务部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.160792+00:00", "updated_at": "2025-08-04T03:29:52.160792+00:00", "business_id": "100049", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
234	DELETE	d2fa55ee-00bd-41cc-b4d1-57f7564661a9	{"old_data": {"id": "d2fa55ee-00bd-41cc-b4d1-57f7564661a9", "name": "边界组织999999", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:29:52.20178+00:00", "updated_at": "2025-08-04T03:29:52.20178+00:00", "business_id": "999999", "description": null, "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:32:19.30751+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.30751+00	2025-08-04 03:32:19.30751+00	\N
235	CREATE	b134bce5-017a-47ac-9c67-c75d33096414	{"new_data": {"id": "b134bce5-017a-47ac-9c67-c75d33096414", "name": "技术部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100000", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
236	CREATE	25337b0b-1272-4834-aa60-fe60b300c0ba	{"new_data": {"id": "25337b0b-1272-4834-aa60-fe60b300c0ba", "name": "产品部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100001", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
237	CREATE	65aa3b33-4d14-4b82-ad44-1d1d3de7b61d	{"new_data": {"id": "65aa3b33-4d14-4b82-ad44-1d1d3de7b61d", "name": "销售部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100002", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
238	CREATE	3886ddb9-92ed-460f-b444-41e8dd4f1566	{"new_data": {"id": "3886ddb9-92ed-460f-b444-41e8dd4f1566", "name": "人事部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100003", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
239	CREATE	21887e79-483f-4258-852e-c9c7492ca501	{"new_data": {"id": "21887e79-483f-4258-852e-c9c7492ca501", "name": "财务部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100004", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
240	CREATE	e4c0c0fd-951d-4532-8e20-001b8b3c5481	{"new_data": {"id": "e4c0c0fd-951d-4532-8e20-001b8b3c5481", "name": "技术部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100005", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
241	CREATE	06e250b6-e96a-451b-b200-af932bdbdf03	{"new_data": {"id": "06e250b6-e96a-451b-b200-af932bdbdf03", "name": "产品部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100006", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
242	CREATE	af1ca8b1-503a-48a4-a954-67d3feaca82b	{"new_data": {"id": "af1ca8b1-503a-48a4-a954-67d3feaca82b", "name": "销售部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100007", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
243	CREATE	8fabe4ac-7387-4d51-91ec-e6331c1fa186	{"new_data": {"id": "8fabe4ac-7387-4d51-91ec-e6331c1fa186", "name": "人事部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100008", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
244	CREATE	092b53ef-48ca-422f-b3ac-d172576e3cd3	{"new_data": {"id": "092b53ef-48ca-422f-b3ac-d172576e3cd3", "name": "财务部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100009", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
245	CREATE	2da4fbe5-c21a-4ada-a9a4-58a9252869a3	{"new_data": {"id": "2da4fbe5-c21a-4ada-a9a4-58a9252869a3", "name": "技术部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100010", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
246	CREATE	1db2e107-dc03-4731-a979-d0df16e672fd	{"new_data": {"id": "1db2e107-dc03-4731-a979-d0df16e672fd", "name": "产品部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100011", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
247	CREATE	ad31a477-65d4-4081-9c60-1a2f9767cfb5	{"new_data": {"id": "ad31a477-65d4-4081-9c60-1a2f9767cfb5", "name": "销售部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100012", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
248	CREATE	79645b7f-13a8-48f3-816b-058c66964a29	{"new_data": {"id": "79645b7f-13a8-48f3-816b-058c66964a29", "name": "人事部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100013", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
249	CREATE	f28e5afd-d1a6-46dc-8fec-cd9113d2ae44	{"new_data": {"id": "f28e5afd-d1a6-46dc-8fec-cd9113d2ae44", "name": "财务部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100014", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
250	CREATE	194a6094-87a5-43c3-8250-fad01d03209e	{"new_data": {"id": "194a6094-87a5-43c3-8250-fad01d03209e", "name": "技术部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100015", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
251	CREATE	d675dfd5-5b52-4bb3-878a-9d004a1b0cc2	{"new_data": {"id": "d675dfd5-5b52-4bb3-878a-9d004a1b0cc2", "name": "产品部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100016", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
252	CREATE	4193adda-39ea-45f6-bc6f-b105ab0d7a2e	{"new_data": {"id": "4193adda-39ea-45f6-bc6f-b105ab0d7a2e", "name": "销售部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100017", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
253	CREATE	a3f4a66f-dabc-4990-be4c-644b8f717843	{"new_data": {"id": "a3f4a66f-dabc-4990-be4c-644b8f717843", "name": "人事部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100018", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
254	CREATE	d5950d0c-3807-452b-8c1e-03ff11a8153e	{"new_data": {"id": "d5950d0c-3807-452b-8c1e-03ff11a8153e", "name": "财务部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100019", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
255	CREATE	38786a13-5fc4-4bf3-bd30-c68647cfb7ac	{"new_data": {"id": "38786a13-5fc4-4bf3-bd30-c68647cfb7ac", "name": "技术部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100020", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
256	CREATE	c240ecd2-c378-4051-810d-b1b5c0266c5a	{"new_data": {"id": "c240ecd2-c378-4051-810d-b1b5c0266c5a", "name": "产品部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100021", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
257	CREATE	5f261937-a696-436d-aabf-1e1adbd6eeb4	{"new_data": {"id": "5f261937-a696-436d-aabf-1e1adbd6eeb4", "name": "销售部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100022", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
258	CREATE	9ccbaafe-069f-4b84-951b-95b8d4da6eee	{"new_data": {"id": "9ccbaafe-069f-4b84-951b-95b8d4da6eee", "name": "人事部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100023", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
259	CREATE	e718e2ad-c0ae-449f-ae02-817c03b5a054	{"new_data": {"id": "e718e2ad-c0ae-449f-ae02-817c03b5a054", "name": "财务部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100024", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
260	CREATE	fff5a011-9e70-4f9c-8a4a-d711b4cd2c4e	{"new_data": {"id": "fff5a011-9e70-4f9c-8a4a-d711b4cd2c4e", "name": "技术部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100025", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
261	CREATE	b8c5f580-9918-4b5e-86b9-6155d3fe0b98	{"new_data": {"id": "b8c5f580-9918-4b5e-86b9-6155d3fe0b98", "name": "产品部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100026", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
262	CREATE	7356653b-6690-4e78-957c-0f8a0177a97b	{"new_data": {"id": "7356653b-6690-4e78-957c-0f8a0177a97b", "name": "销售部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100027", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
263	CREATE	2fbaf774-71d1-4db0-b359-aa70134717be	{"new_data": {"id": "2fbaf774-71d1-4db0-b359-aa70134717be", "name": "人事部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100028", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
264	CREATE	75253133-4eac-4d28-b4df-83d54eed83c6	{"new_data": {"id": "75253133-4eac-4d28-b4df-83d54eed83c6", "name": "财务部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100029", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
265	CREATE	7af09c72-75cf-4b86-9db5-1046a2e6b7f1	{"new_data": {"id": "7af09c72-75cf-4b86-9db5-1046a2e6b7f1", "name": "技术部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100030", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
266	CREATE	26a249b8-e94f-4e39-8596-ae1a62f92265	{"new_data": {"id": "26a249b8-e94f-4e39-8596-ae1a62f92265", "name": "产品部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100031", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
267	CREATE	6027c432-5f5a-4bfe-8b62-249efee6593f	{"new_data": {"id": "6027c432-5f5a-4bfe-8b62-249efee6593f", "name": "销售部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100032", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
268	CREATE	f6cf2a87-855a-44dd-8747-0e1d0605f5ee	{"new_data": {"id": "f6cf2a87-855a-44dd-8747-0e1d0605f5ee", "name": "人事部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100033", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
269	CREATE	d1541653-4ee7-47fe-abd7-e3211d66bd4d	{"new_data": {"id": "d1541653-4ee7-47fe-abd7-e3211d66bd4d", "name": "财务部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100034", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
270	CREATE	47ed1946-6002-42ad-8978-a493049bd721	{"new_data": {"id": "47ed1946-6002-42ad-8978-a493049bd721", "name": "技术部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100035", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
271	CREATE	8fca144d-8f98-46df-a9ff-28a4183787ee	{"new_data": {"id": "8fca144d-8f98-46df-a9ff-28a4183787ee", "name": "产品部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100036", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
272	CREATE	67c9fd26-c1d3-4820-a1b4-519f8096df52	{"new_data": {"id": "67c9fd26-c1d3-4820-a1b4-519f8096df52", "name": "销售部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100037", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
273	CREATE	bad5defe-c477-4953-bb4b-d7088fbb7bc8	{"new_data": {"id": "bad5defe-c477-4953-bb4b-d7088fbb7bc8", "name": "人事部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100038", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
274	CREATE	1106094f-df3a-45c9-b031-7d7894c87954	{"new_data": {"id": "1106094f-df3a-45c9-b031-7d7894c87954", "name": "财务部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100039", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
275	CREATE	3ad5d80b-4b6e-4a3c-9ad8-68f9fc8ddea6	{"new_data": {"id": "3ad5d80b-4b6e-4a3c-9ad8-68f9fc8ddea6", "name": "技术部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100040", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
276	CREATE	a73796ee-2aea-48b5-bb6f-a1bb0338ea7b	{"new_data": {"id": "a73796ee-2aea-48b5-bb6f-a1bb0338ea7b", "name": "产品部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100041", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
277	CREATE	cb163d29-6b04-4a17-9206-706fdc12264e	{"new_data": {"id": "cb163d29-6b04-4a17-9206-706fdc12264e", "name": "销售部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100042", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
278	CREATE	c6d12b8b-bea1-4e58-8a4f-360f86295b8e	{"new_data": {"id": "c6d12b8b-bea1-4e58-8a4f-360f86295b8e", "name": "人事部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100043", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
279	CREATE	de51e905-6e10-4a4d-87d6-276f2797a0ba	{"new_data": {"id": "de51e905-6e10-4a4d-87d6-276f2797a0ba", "name": "财务部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100044", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
280	CREATE	bc4caa20-0415-48a9-b8fd-17aa06d34860	{"new_data": {"id": "bc4caa20-0415-48a9-b8fd-17aa06d34860", "name": "技术部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100045", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
281	CREATE	5a25a8f1-86ea-462a-8742-c7e171de1284	{"new_data": {"id": "5a25a8f1-86ea-462a-8742-c7e171de1284", "name": "产品部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100046", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
282	CREATE	45fef2c9-14f1-440a-8f80-4a47605b0247	{"new_data": {"id": "45fef2c9-14f1-440a-8f80-4a47605b0247", "name": "销售部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100047", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
283	CREATE	eae8eefa-b507-4b6e-ae9d-eb1c6e518e96	{"new_data": {"id": "eae8eefa-b507-4b6e-ae9d-eb1c6e518e96", "name": "人事部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100048", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
284	CREATE	702bb76b-c934-4f87-b2ee-7a0754d62c86	{"new_data": {"id": "702bb76b-c934-4f87-b2ee-7a0754d62c86", "name": "财务部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100049", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.474494+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.474494+00	2025-08-04 03:32:19.474494+00	\N
285	CREATE	c433e3f0-6b8e-418f-a940-5d706fe3edab	{"new_data": {"id": "c433e3f0-6b8e-418f-a940-5d706fe3edab", "name": "边界组织999999", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.508287+00:00", "updated_at": "2025-08-04T03:32:19.508287+00:00", "business_id": "999999", "description": null, "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.508287+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.508287+00	2025-08-04 03:32:19.508287+00	\N
286	CREATE	11111111-1111-1111-1111-111111111111	{"new_data": {"id": "11111111-1111-1111-1111-111111111111", "name": "默认部门", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.539905+00:00", "updated_at": "2025-08-04T03:32:19.539905+00:00", "business_id": "100050", "description": null, "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:32:19.539905+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:32:19.539905+00	2025-08-04 03:32:19.539905+00	\N
287	DELETE	b134bce5-017a-47ac-9c67-c75d33096414	{"old_data": {"id": "b134bce5-017a-47ac-9c67-c75d33096414", "name": "技术部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100000", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
288	DELETE	25337b0b-1272-4834-aa60-fe60b300c0ba	{"old_data": {"id": "25337b0b-1272-4834-aa60-fe60b300c0ba", "name": "产品部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100001", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
289	DELETE	65aa3b33-4d14-4b82-ad44-1d1d3de7b61d	{"old_data": {"id": "65aa3b33-4d14-4b82-ad44-1d1d3de7b61d", "name": "销售部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100002", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
290	DELETE	3886ddb9-92ed-460f-b444-41e8dd4f1566	{"old_data": {"id": "3886ddb9-92ed-460f-b444-41e8dd4f1566", "name": "人事部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100003", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
291	DELETE	21887e79-483f-4258-852e-c9c7492ca501	{"old_data": {"id": "21887e79-483f-4258-852e-c9c7492ca501", "name": "财务部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100004", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
292	DELETE	e4c0c0fd-951d-4532-8e20-001b8b3c5481	{"old_data": {"id": "e4c0c0fd-951d-4532-8e20-001b8b3c5481", "name": "技术部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100005", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
293	DELETE	06e250b6-e96a-451b-b200-af932bdbdf03	{"old_data": {"id": "06e250b6-e96a-451b-b200-af932bdbdf03", "name": "产品部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100006", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
294	DELETE	af1ca8b1-503a-48a4-a954-67d3feaca82b	{"old_data": {"id": "af1ca8b1-503a-48a4-a954-67d3feaca82b", "name": "销售部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100007", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
295	DELETE	8fabe4ac-7387-4d51-91ec-e6331c1fa186	{"old_data": {"id": "8fabe4ac-7387-4d51-91ec-e6331c1fa186", "name": "人事部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100008", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
296	DELETE	092b53ef-48ca-422f-b3ac-d172576e3cd3	{"old_data": {"id": "092b53ef-48ca-422f-b3ac-d172576e3cd3", "name": "财务部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100009", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
297	DELETE	2da4fbe5-c21a-4ada-a9a4-58a9252869a3	{"old_data": {"id": "2da4fbe5-c21a-4ada-a9a4-58a9252869a3", "name": "技术部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100010", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
298	DELETE	1db2e107-dc03-4731-a979-d0df16e672fd	{"old_data": {"id": "1db2e107-dc03-4731-a979-d0df16e672fd", "name": "产品部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100011", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
299	DELETE	ad31a477-65d4-4081-9c60-1a2f9767cfb5	{"old_data": {"id": "ad31a477-65d4-4081-9c60-1a2f9767cfb5", "name": "销售部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100012", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
300	DELETE	79645b7f-13a8-48f3-816b-058c66964a29	{"old_data": {"id": "79645b7f-13a8-48f3-816b-058c66964a29", "name": "人事部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100013", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
301	DELETE	f28e5afd-d1a6-46dc-8fec-cd9113d2ae44	{"old_data": {"id": "f28e5afd-d1a6-46dc-8fec-cd9113d2ae44", "name": "财务部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100014", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
302	DELETE	194a6094-87a5-43c3-8250-fad01d03209e	{"old_data": {"id": "194a6094-87a5-43c3-8250-fad01d03209e", "name": "技术部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100015", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
303	DELETE	d675dfd5-5b52-4bb3-878a-9d004a1b0cc2	{"old_data": {"id": "d675dfd5-5b52-4bb3-878a-9d004a1b0cc2", "name": "产品部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100016", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
304	DELETE	4193adda-39ea-45f6-bc6f-b105ab0d7a2e	{"old_data": {"id": "4193adda-39ea-45f6-bc6f-b105ab0d7a2e", "name": "销售部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100017", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
305	DELETE	a3f4a66f-dabc-4990-be4c-644b8f717843	{"old_data": {"id": "a3f4a66f-dabc-4990-be4c-644b8f717843", "name": "人事部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100018", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
306	DELETE	d5950d0c-3807-452b-8c1e-03ff11a8153e	{"old_data": {"id": "d5950d0c-3807-452b-8c1e-03ff11a8153e", "name": "财务部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100019", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
307	DELETE	38786a13-5fc4-4bf3-bd30-c68647cfb7ac	{"old_data": {"id": "38786a13-5fc4-4bf3-bd30-c68647cfb7ac", "name": "技术部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100020", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
308	DELETE	c240ecd2-c378-4051-810d-b1b5c0266c5a	{"old_data": {"id": "c240ecd2-c378-4051-810d-b1b5c0266c5a", "name": "产品部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100021", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
309	DELETE	5f261937-a696-436d-aabf-1e1adbd6eeb4	{"old_data": {"id": "5f261937-a696-436d-aabf-1e1adbd6eeb4", "name": "销售部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100022", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
310	DELETE	9ccbaafe-069f-4b84-951b-95b8d4da6eee	{"old_data": {"id": "9ccbaafe-069f-4b84-951b-95b8d4da6eee", "name": "人事部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100023", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
311	DELETE	e718e2ad-c0ae-449f-ae02-817c03b5a054	{"old_data": {"id": "e718e2ad-c0ae-449f-ae02-817c03b5a054", "name": "财务部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100024", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
312	DELETE	fff5a011-9e70-4f9c-8a4a-d711b4cd2c4e	{"old_data": {"id": "fff5a011-9e70-4f9c-8a4a-d711b4cd2c4e", "name": "技术部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100025", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
313	DELETE	b8c5f580-9918-4b5e-86b9-6155d3fe0b98	{"old_data": {"id": "b8c5f580-9918-4b5e-86b9-6155d3fe0b98", "name": "产品部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100026", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
314	DELETE	7356653b-6690-4e78-957c-0f8a0177a97b	{"old_data": {"id": "7356653b-6690-4e78-957c-0f8a0177a97b", "name": "销售部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100027", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
315	DELETE	2fbaf774-71d1-4db0-b359-aa70134717be	{"old_data": {"id": "2fbaf774-71d1-4db0-b359-aa70134717be", "name": "人事部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100028", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
316	DELETE	75253133-4eac-4d28-b4df-83d54eed83c6	{"old_data": {"id": "75253133-4eac-4d28-b4df-83d54eed83c6", "name": "财务部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100029", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
317	DELETE	7af09c72-75cf-4b86-9db5-1046a2e6b7f1	{"old_data": {"id": "7af09c72-75cf-4b86-9db5-1046a2e6b7f1", "name": "技术部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100030", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
318	DELETE	26a249b8-e94f-4e39-8596-ae1a62f92265	{"old_data": {"id": "26a249b8-e94f-4e39-8596-ae1a62f92265", "name": "产品部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100031", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
319	DELETE	6027c432-5f5a-4bfe-8b62-249efee6593f	{"old_data": {"id": "6027c432-5f5a-4bfe-8b62-249efee6593f", "name": "销售部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100032", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
320	DELETE	f6cf2a87-855a-44dd-8747-0e1d0605f5ee	{"old_data": {"id": "f6cf2a87-855a-44dd-8747-0e1d0605f5ee", "name": "人事部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100033", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
321	DELETE	d1541653-4ee7-47fe-abd7-e3211d66bd4d	{"old_data": {"id": "d1541653-4ee7-47fe-abd7-e3211d66bd4d", "name": "财务部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100034", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
322	DELETE	47ed1946-6002-42ad-8978-a493049bd721	{"old_data": {"id": "47ed1946-6002-42ad-8978-a493049bd721", "name": "技术部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100035", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
323	DELETE	8fca144d-8f98-46df-a9ff-28a4183787ee	{"old_data": {"id": "8fca144d-8f98-46df-a9ff-28a4183787ee", "name": "产品部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100036", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
324	DELETE	67c9fd26-c1d3-4820-a1b4-519f8096df52	{"old_data": {"id": "67c9fd26-c1d3-4820-a1b4-519f8096df52", "name": "销售部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100037", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
325	DELETE	bad5defe-c477-4953-bb4b-d7088fbb7bc8	{"old_data": {"id": "bad5defe-c477-4953-bb4b-d7088fbb7bc8", "name": "人事部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100038", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
326	DELETE	1106094f-df3a-45c9-b031-7d7894c87954	{"old_data": {"id": "1106094f-df3a-45c9-b031-7d7894c87954", "name": "财务部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100039", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
327	DELETE	3ad5d80b-4b6e-4a3c-9ad8-68f9fc8ddea6	{"old_data": {"id": "3ad5d80b-4b6e-4a3c-9ad8-68f9fc8ddea6", "name": "技术部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100040", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
328	DELETE	a73796ee-2aea-48b5-bb6f-a1bb0338ea7b	{"old_data": {"id": "a73796ee-2aea-48b5-bb6f-a1bb0338ea7b", "name": "产品部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100041", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
329	DELETE	cb163d29-6b04-4a17-9206-706fdc12264e	{"old_data": {"id": "cb163d29-6b04-4a17-9206-706fdc12264e", "name": "销售部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100042", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
330	DELETE	c6d12b8b-bea1-4e58-8a4f-360f86295b8e	{"old_data": {"id": "c6d12b8b-bea1-4e58-8a4f-360f86295b8e", "name": "人事部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100043", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
331	DELETE	de51e905-6e10-4a4d-87d6-276f2797a0ba	{"old_data": {"id": "de51e905-6e10-4a4d-87d6-276f2797a0ba", "name": "财务部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100044", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
332	DELETE	bc4caa20-0415-48a9-b8fd-17aa06d34860	{"old_data": {"id": "bc4caa20-0415-48a9-b8fd-17aa06d34860", "name": "技术部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100045", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
333	DELETE	5a25a8f1-86ea-462a-8742-c7e171de1284	{"old_data": {"id": "5a25a8f1-86ea-462a-8742-c7e171de1284", "name": "产品部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100046", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
334	DELETE	45fef2c9-14f1-440a-8f80-4a47605b0247	{"old_data": {"id": "45fef2c9-14f1-440a-8f80-4a47605b0247", "name": "销售部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100047", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
335	DELETE	eae8eefa-b507-4b6e-ae9d-eb1c6e518e96	{"old_data": {"id": "eae8eefa-b507-4b6e-ae9d-eb1c6e518e96", "name": "人事部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100048", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
336	DELETE	702bb76b-c934-4f87-b2ee-7a0754d62c86	{"old_data": {"id": "702bb76b-c934-4f87-b2ee-7a0754d62c86", "name": "财务部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.474494+00:00", "updated_at": "2025-08-04T03:32:19.474494+00:00", "business_id": "100049", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
337	DELETE	c433e3f0-6b8e-418f-a940-5d706fe3edab	{"old_data": {"id": "c433e3f0-6b8e-418f-a940-5d706fe3edab", "name": "边界组织999999", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.508287+00:00", "updated_at": "2025-08-04T03:32:19.508287+00:00", "business_id": "999999", "description": null, "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
338	DELETE	11111111-1111-1111-1111-111111111111	{"old_data": {"id": "11111111-1111-1111-1111-111111111111", "name": "默认部门", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:32:19.539905+00:00", "updated_at": "2025-08-04T03:32:19.539905+00:00", "business_id": "100050", "description": null, "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:04.8775+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:04.8775+00	2025-08-04 03:40:04.8775+00	\N
339	CREATE	51dbe2d2-180f-4f54-bfc6-156ab37655e7	{"new_data": {"id": "51dbe2d2-180f-4f54-bfc6-156ab37655e7", "name": "技术部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100000", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
340	CREATE	910fda21-ce80-4dbc-9caa-25633800cdee	{"new_data": {"id": "910fda21-ce80-4dbc-9caa-25633800cdee", "name": "产品部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100001", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
341	CREATE	8d914d84-c6e9-4dce-9e81-a77dd418e106	{"new_data": {"id": "8d914d84-c6e9-4dce-9e81-a77dd418e106", "name": "销售部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100002", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
342	CREATE	ae2686a8-454c-4300-9c7a-166f0f2ee82e	{"new_data": {"id": "ae2686a8-454c-4300-9c7a-166f0f2ee82e", "name": "人事部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100003", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
343	CREATE	f1a8d7ab-ee72-4ebf-8f6a-ef922d99c25d	{"new_data": {"id": "f1a8d7ab-ee72-4ebf-8f6a-ef922d99c25d", "name": "财务部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100004", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
344	CREATE	61adece5-fccf-4682-bf0c-935be73967bb	{"new_data": {"id": "61adece5-fccf-4682-bf0c-935be73967bb", "name": "技术部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100005", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
345	CREATE	7ffa1b60-c707-4ea3-b809-3c8edebb9119	{"new_data": {"id": "7ffa1b60-c707-4ea3-b809-3c8edebb9119", "name": "产品部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100006", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
346	CREATE	4f302844-4191-49d6-a2f1-168ae3ec3f73	{"new_data": {"id": "4f302844-4191-49d6-a2f1-168ae3ec3f73", "name": "销售部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100007", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
347	CREATE	c36859d4-7a8f-419f-9218-83964b8fdd8a	{"new_data": {"id": "c36859d4-7a8f-419f-9218-83964b8fdd8a", "name": "人事部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100008", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
348	CREATE	e71d6ba6-e0fb-4cef-8d67-e78401444939	{"new_data": {"id": "e71d6ba6-e0fb-4cef-8d67-e78401444939", "name": "财务部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100009", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
349	CREATE	78bd7ce4-a0e3-4d7e-ab8c-651ef2baf24e	{"new_data": {"id": "78bd7ce4-a0e3-4d7e-ab8c-651ef2baf24e", "name": "技术部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100010", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
350	CREATE	843152a0-29d8-4e4c-8aa8-fed5acc4024b	{"new_data": {"id": "843152a0-29d8-4e4c-8aa8-fed5acc4024b", "name": "产品部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100011", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
351	CREATE	9d792488-d06d-4258-9582-47f564d5a47d	{"new_data": {"id": "9d792488-d06d-4258-9582-47f564d5a47d", "name": "销售部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100012", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
352	CREATE	b889e603-8240-4bb0-aaef-3c11fd8cb83b	{"new_data": {"id": "b889e603-8240-4bb0-aaef-3c11fd8cb83b", "name": "人事部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100013", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
353	CREATE	74725e57-acd3-4864-b02c-9b0633bcff56	{"new_data": {"id": "74725e57-acd3-4864-b02c-9b0633bcff56", "name": "财务部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100014", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
354	CREATE	5e7ad705-7ded-4f58-8a5a-7b2432271233	{"new_data": {"id": "5e7ad705-7ded-4f58-8a5a-7b2432271233", "name": "技术部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100015", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
355	CREATE	c0569f87-4417-4751-bf32-90ba9df7747c	{"new_data": {"id": "c0569f87-4417-4751-bf32-90ba9df7747c", "name": "产品部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100016", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
356	CREATE	67113395-ac7f-4d24-8230-d0751007ad49	{"new_data": {"id": "67113395-ac7f-4d24-8230-d0751007ad49", "name": "销售部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100017", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
357	CREATE	3f206997-433c-45f4-87d3-42649167c589	{"new_data": {"id": "3f206997-433c-45f4-87d3-42649167c589", "name": "人事部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100018", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
358	CREATE	7dd26f05-7914-4195-902f-d94f6968b23f	{"new_data": {"id": "7dd26f05-7914-4195-902f-d94f6968b23f", "name": "财务部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100019", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
359	CREATE	dcd76c46-a4bf-482d-8930-9c91faa85777	{"new_data": {"id": "dcd76c46-a4bf-482d-8930-9c91faa85777", "name": "技术部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100020", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
360	CREATE	dca7751f-35c0-4a84-9588-e689cd11da58	{"new_data": {"id": "dca7751f-35c0-4a84-9588-e689cd11da58", "name": "产品部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100021", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
361	CREATE	85f21d88-4b77-4c05-9dde-cc2883a9aedd	{"new_data": {"id": "85f21d88-4b77-4c05-9dde-cc2883a9aedd", "name": "销售部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100022", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
362	CREATE	00fc7190-167b-483c-af16-e760dda6dce8	{"new_data": {"id": "00fc7190-167b-483c-af16-e760dda6dce8", "name": "人事部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100023", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
363	CREATE	e95cc0b3-6de2-4c62-a14d-bd5eb37ebbb0	{"new_data": {"id": "e95cc0b3-6de2-4c62-a14d-bd5eb37ebbb0", "name": "财务部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100024", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
364	CREATE	a051afe9-3e5b-45ec-94c9-01d5cc42818b	{"new_data": {"id": "a051afe9-3e5b-45ec-94c9-01d5cc42818b", "name": "技术部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100025", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
365	CREATE	55c7c2e0-fd27-476e-acc6-e8ad706dd09f	{"new_data": {"id": "55c7c2e0-fd27-476e-acc6-e8ad706dd09f", "name": "产品部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100026", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
366	CREATE	c79e3a67-d146-49f2-bd29-f40d5822abb9	{"new_data": {"id": "c79e3a67-d146-49f2-bd29-f40d5822abb9", "name": "销售部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100027", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
367	CREATE	c4a5c0ba-4682-4793-9967-5108a56c0eef	{"new_data": {"id": "c4a5c0ba-4682-4793-9967-5108a56c0eef", "name": "人事部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100028", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
368	CREATE	61eb3bb3-5f84-4164-b155-7afde4bb2d90	{"new_data": {"id": "61eb3bb3-5f84-4164-b155-7afde4bb2d90", "name": "财务部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100029", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
369	CREATE	589d1b1b-5182-4360-8d8f-1269f8d49897	{"new_data": {"id": "589d1b1b-5182-4360-8d8f-1269f8d49897", "name": "技术部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100030", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
370	CREATE	c75cc074-8fb6-40cc-a6de-54b034c1dd58	{"new_data": {"id": "c75cc074-8fb6-40cc-a6de-54b034c1dd58", "name": "产品部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100031", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
371	CREATE	e3f62ff3-ea4f-467e-82ba-f1bbc54b2f91	{"new_data": {"id": "e3f62ff3-ea4f-467e-82ba-f1bbc54b2f91", "name": "销售部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100032", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
372	CREATE	f8eebe45-f7e8-4765-bd61-d67ddb189a22	{"new_data": {"id": "f8eebe45-f7e8-4765-bd61-d67ddb189a22", "name": "人事部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100033", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
373	CREATE	afb3ff24-5167-44b1-ae71-e4709650f7e4	{"new_data": {"id": "afb3ff24-5167-44b1-ae71-e4709650f7e4", "name": "财务部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100034", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
374	CREATE	565925e9-b2f3-4a20-b455-2d59a29dbdf2	{"new_data": {"id": "565925e9-b2f3-4a20-b455-2d59a29dbdf2", "name": "技术部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100035", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
375	CREATE	9bc340fb-5d73-4a69-8f86-0279bac088ae	{"new_data": {"id": "9bc340fb-5d73-4a69-8f86-0279bac088ae", "name": "产品部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100036", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
376	CREATE	0e0e9d1d-724d-4b8b-95b2-0b0cf2e0fb4f	{"new_data": {"id": "0e0e9d1d-724d-4b8b-95b2-0b0cf2e0fb4f", "name": "销售部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100037", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
377	CREATE	777a08ce-37d8-4497-9163-b20cad22e92c	{"new_data": {"id": "777a08ce-37d8-4497-9163-b20cad22e92c", "name": "人事部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100038", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
378	CREATE	91213f54-8bcd-4e36-9732-1bcfbb97102e	{"new_data": {"id": "91213f54-8bcd-4e36-9732-1bcfbb97102e", "name": "财务部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100039", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
379	CREATE	1e02f6ac-284f-4e4e-aed8-fe493d13e89a	{"new_data": {"id": "1e02f6ac-284f-4e4e-aed8-fe493d13e89a", "name": "技术部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100040", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
380	CREATE	175f6c8c-6057-4769-9e17-e45f3a856b89	{"new_data": {"id": "175f6c8c-6057-4769-9e17-e45f3a856b89", "name": "产品部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100041", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
381	CREATE	53cbad6a-3af9-4d90-a865-d319a6f8e1a8	{"new_data": {"id": "53cbad6a-3af9-4d90-a865-d319a6f8e1a8", "name": "销售部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100042", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
382	CREATE	4cd00d0f-41e9-43cb-a912-88b2616b6857	{"new_data": {"id": "4cd00d0f-41e9-43cb-a912-88b2616b6857", "name": "人事部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100043", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
383	CREATE	50f4d973-95d9-46fb-bec2-84e91bc517c1	{"new_data": {"id": "50f4d973-95d9-46fb-bec2-84e91bc517c1", "name": "财务部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100044", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
384	CREATE	7fc4524d-cac1-4513-a508-e4a3111d7c98	{"new_data": {"id": "7fc4524d-cac1-4513-a508-e4a3111d7c98", "name": "技术部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100045", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
385	CREATE	0d98a255-e913-489e-8a90-de0ecf568772	{"new_data": {"id": "0d98a255-e913-489e-8a90-de0ecf568772", "name": "产品部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100046", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
386	CREATE	50b46fb3-828b-4503-bbdc-92e5478f4d52	{"new_data": {"id": "50b46fb3-828b-4503-bbdc-92e5478f4d52", "name": "销售部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100047", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
387	CREATE	8f6c2996-990a-4400-884b-fa6b5c374378	{"new_data": {"id": "8f6c2996-990a-4400-884b-fa6b5c374378", "name": "人事部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100048", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
388	CREATE	b511e9c9-56ac-46c7-a5ae-0b406ac2d088	{"new_data": {"id": "b511e9c9-56ac-46c7-a5ae-0b406ac2d088", "name": "财务部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100049", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.05874+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.05874+00	2025-08-04 03:40:05.05874+00	\N
389	CREATE	be986bfd-1ea3-4f4a-b2ce-e8a5c402b94c	{"new_data": {"id": "be986bfd-1ea3-4f4a-b2ce-e8a5c402b94c", "name": "边界组织999999", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.095245+00:00", "updated_at": "2025-08-04T03:40:05.095245+00:00", "business_id": "999999", "description": null, "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.095245+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.095245+00	2025-08-04 03:40:05.095245+00	\N
390	CREATE	11111111-1111-1111-1111-111111111111	{"new_data": {"id": "11111111-1111-1111-1111-111111111111", "name": "默认部门", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.127268+00:00", "updated_at": "2025-08-04T03:40:05.127268+00:00", "business_id": "100050", "description": null, "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:05.127268+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:05.127268+00	2025-08-04 03:40:05.127268+00	\N
391	DELETE	51dbe2d2-180f-4f54-bfc6-156ab37655e7	{"old_data": {"id": "51dbe2d2-180f-4f54-bfc6-156ab37655e7", "name": "技术部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100000", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
392	DELETE	910fda21-ce80-4dbc-9caa-25633800cdee	{"old_data": {"id": "910fda21-ce80-4dbc-9caa-25633800cdee", "name": "产品部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100001", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
393	DELETE	8d914d84-c6e9-4dce-9e81-a77dd418e106	{"old_data": {"id": "8d914d84-c6e9-4dce-9e81-a77dd418e106", "name": "销售部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100002", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
394	DELETE	ae2686a8-454c-4300-9c7a-166f0f2ee82e	{"old_data": {"id": "ae2686a8-454c-4300-9c7a-166f0f2ee82e", "name": "人事部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100003", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
395	DELETE	f1a8d7ab-ee72-4ebf-8f6a-ef922d99c25d	{"old_data": {"id": "f1a8d7ab-ee72-4ebf-8f6a-ef922d99c25d", "name": "财务部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100004", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
396	DELETE	61adece5-fccf-4682-bf0c-935be73967bb	{"old_data": {"id": "61adece5-fccf-4682-bf0c-935be73967bb", "name": "技术部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100005", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
397	DELETE	7ffa1b60-c707-4ea3-b809-3c8edebb9119	{"old_data": {"id": "7ffa1b60-c707-4ea3-b809-3c8edebb9119", "name": "产品部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100006", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
398	DELETE	4f302844-4191-49d6-a2f1-168ae3ec3f73	{"old_data": {"id": "4f302844-4191-49d6-a2f1-168ae3ec3f73", "name": "销售部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100007", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
399	DELETE	c36859d4-7a8f-419f-9218-83964b8fdd8a	{"old_data": {"id": "c36859d4-7a8f-419f-9218-83964b8fdd8a", "name": "人事部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100008", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
400	DELETE	e71d6ba6-e0fb-4cef-8d67-e78401444939	{"old_data": {"id": "e71d6ba6-e0fb-4cef-8d67-e78401444939", "name": "财务部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100009", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
401	DELETE	78bd7ce4-a0e3-4d7e-ab8c-651ef2baf24e	{"old_data": {"id": "78bd7ce4-a0e3-4d7e-ab8c-651ef2baf24e", "name": "技术部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100010", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
402	DELETE	843152a0-29d8-4e4c-8aa8-fed5acc4024b	{"old_data": {"id": "843152a0-29d8-4e4c-8aa8-fed5acc4024b", "name": "产品部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100011", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
403	DELETE	9d792488-d06d-4258-9582-47f564d5a47d	{"old_data": {"id": "9d792488-d06d-4258-9582-47f564d5a47d", "name": "销售部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100012", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
404	DELETE	b889e603-8240-4bb0-aaef-3c11fd8cb83b	{"old_data": {"id": "b889e603-8240-4bb0-aaef-3c11fd8cb83b", "name": "人事部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100013", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
405	DELETE	74725e57-acd3-4864-b02c-9b0633bcff56	{"old_data": {"id": "74725e57-acd3-4864-b02c-9b0633bcff56", "name": "财务部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100014", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
406	DELETE	5e7ad705-7ded-4f58-8a5a-7b2432271233	{"old_data": {"id": "5e7ad705-7ded-4f58-8a5a-7b2432271233", "name": "技术部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100015", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
407	DELETE	c0569f87-4417-4751-bf32-90ba9df7747c	{"old_data": {"id": "c0569f87-4417-4751-bf32-90ba9df7747c", "name": "产品部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100016", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
408	DELETE	67113395-ac7f-4d24-8230-d0751007ad49	{"old_data": {"id": "67113395-ac7f-4d24-8230-d0751007ad49", "name": "销售部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100017", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
409	DELETE	3f206997-433c-45f4-87d3-42649167c589	{"old_data": {"id": "3f206997-433c-45f4-87d3-42649167c589", "name": "人事部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100018", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
410	DELETE	7dd26f05-7914-4195-902f-d94f6968b23f	{"old_data": {"id": "7dd26f05-7914-4195-902f-d94f6968b23f", "name": "财务部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100019", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
411	DELETE	dcd76c46-a4bf-482d-8930-9c91faa85777	{"old_data": {"id": "dcd76c46-a4bf-482d-8930-9c91faa85777", "name": "技术部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100020", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
412	DELETE	dca7751f-35c0-4a84-9588-e689cd11da58	{"old_data": {"id": "dca7751f-35c0-4a84-9588-e689cd11da58", "name": "产品部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100021", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
413	DELETE	85f21d88-4b77-4c05-9dde-cc2883a9aedd	{"old_data": {"id": "85f21d88-4b77-4c05-9dde-cc2883a9aedd", "name": "销售部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100022", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
414	DELETE	00fc7190-167b-483c-af16-e760dda6dce8	{"old_data": {"id": "00fc7190-167b-483c-af16-e760dda6dce8", "name": "人事部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100023", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
415	DELETE	e95cc0b3-6de2-4c62-a14d-bd5eb37ebbb0	{"old_data": {"id": "e95cc0b3-6de2-4c62-a14d-bd5eb37ebbb0", "name": "财务部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100024", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
416	DELETE	a051afe9-3e5b-45ec-94c9-01d5cc42818b	{"old_data": {"id": "a051afe9-3e5b-45ec-94c9-01d5cc42818b", "name": "技术部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100025", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
417	DELETE	55c7c2e0-fd27-476e-acc6-e8ad706dd09f	{"old_data": {"id": "55c7c2e0-fd27-476e-acc6-e8ad706dd09f", "name": "产品部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100026", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
418	DELETE	c79e3a67-d146-49f2-bd29-f40d5822abb9	{"old_data": {"id": "c79e3a67-d146-49f2-bd29-f40d5822abb9", "name": "销售部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100027", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
419	DELETE	c4a5c0ba-4682-4793-9967-5108a56c0eef	{"old_data": {"id": "c4a5c0ba-4682-4793-9967-5108a56c0eef", "name": "人事部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100028", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
420	DELETE	61eb3bb3-5f84-4164-b155-7afde4bb2d90	{"old_data": {"id": "61eb3bb3-5f84-4164-b155-7afde4bb2d90", "name": "财务部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100029", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
421	DELETE	589d1b1b-5182-4360-8d8f-1269f8d49897	{"old_data": {"id": "589d1b1b-5182-4360-8d8f-1269f8d49897", "name": "技术部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100030", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
422	DELETE	c75cc074-8fb6-40cc-a6de-54b034c1dd58	{"old_data": {"id": "c75cc074-8fb6-40cc-a6de-54b034c1dd58", "name": "产品部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100031", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
423	DELETE	e3f62ff3-ea4f-467e-82ba-f1bbc54b2f91	{"old_data": {"id": "e3f62ff3-ea4f-467e-82ba-f1bbc54b2f91", "name": "销售部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100032", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
424	DELETE	f8eebe45-f7e8-4765-bd61-d67ddb189a22	{"old_data": {"id": "f8eebe45-f7e8-4765-bd61-d67ddb189a22", "name": "人事部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100033", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
425	DELETE	afb3ff24-5167-44b1-ae71-e4709650f7e4	{"old_data": {"id": "afb3ff24-5167-44b1-ae71-e4709650f7e4", "name": "财务部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100034", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
426	DELETE	565925e9-b2f3-4a20-b455-2d59a29dbdf2	{"old_data": {"id": "565925e9-b2f3-4a20-b455-2d59a29dbdf2", "name": "技术部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100035", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
427	DELETE	9bc340fb-5d73-4a69-8f86-0279bac088ae	{"old_data": {"id": "9bc340fb-5d73-4a69-8f86-0279bac088ae", "name": "产品部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100036", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
428	DELETE	0e0e9d1d-724d-4b8b-95b2-0b0cf2e0fb4f	{"old_data": {"id": "0e0e9d1d-724d-4b8b-95b2-0b0cf2e0fb4f", "name": "销售部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100037", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
429	DELETE	777a08ce-37d8-4497-9163-b20cad22e92c	{"old_data": {"id": "777a08ce-37d8-4497-9163-b20cad22e92c", "name": "人事部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100038", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
430	DELETE	91213f54-8bcd-4e36-9732-1bcfbb97102e	{"old_data": {"id": "91213f54-8bcd-4e36-9732-1bcfbb97102e", "name": "财务部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100039", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
431	DELETE	1e02f6ac-284f-4e4e-aed8-fe493d13e89a	{"old_data": {"id": "1e02f6ac-284f-4e4e-aed8-fe493d13e89a", "name": "技术部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100040", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
432	DELETE	175f6c8c-6057-4769-9e17-e45f3a856b89	{"old_data": {"id": "175f6c8c-6057-4769-9e17-e45f3a856b89", "name": "产品部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100041", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
433	DELETE	53cbad6a-3af9-4d90-a865-d319a6f8e1a8	{"old_data": {"id": "53cbad6a-3af9-4d90-a865-d319a6f8e1a8", "name": "销售部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100042", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
434	DELETE	4cd00d0f-41e9-43cb-a912-88b2616b6857	{"old_data": {"id": "4cd00d0f-41e9-43cb-a912-88b2616b6857", "name": "人事部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100043", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
435	DELETE	50f4d973-95d9-46fb-bec2-84e91bc517c1	{"old_data": {"id": "50f4d973-95d9-46fb-bec2-84e91bc517c1", "name": "财务部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100044", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
436	DELETE	7fc4524d-cac1-4513-a508-e4a3111d7c98	{"old_data": {"id": "7fc4524d-cac1-4513-a508-e4a3111d7c98", "name": "技术部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100045", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
437	DELETE	0d98a255-e913-489e-8a90-de0ecf568772	{"old_data": {"id": "0d98a255-e913-489e-8a90-de0ecf568772", "name": "产品部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100046", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
438	DELETE	50b46fb3-828b-4503-bbdc-92e5478f4d52	{"old_data": {"id": "50b46fb3-828b-4503-bbdc-92e5478f4d52", "name": "销售部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100047", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
439	DELETE	8f6c2996-990a-4400-884b-fa6b5c374378	{"old_data": {"id": "8f6c2996-990a-4400-884b-fa6b5c374378", "name": "人事部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100048", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
440	DELETE	b511e9c9-56ac-46c7-a5ae-0b406ac2d088	{"old_data": {"id": "b511e9c9-56ac-46c7-a5ae-0b406ac2d088", "name": "财务部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.05874+00:00", "updated_at": "2025-08-04T03:40:05.05874+00:00", "business_id": "100049", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
441	DELETE	be986bfd-1ea3-4f4a-b2ce-e8a5c402b94c	{"old_data": {"id": "be986bfd-1ea3-4f4a-b2ce-e8a5c402b94c", "name": "边界组织999999", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.095245+00:00", "updated_at": "2025-08-04T03:40:05.095245+00:00", "business_id": "999999", "description": null, "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
442	DELETE	11111111-1111-1111-1111-111111111111	{"old_data": {"id": "11111111-1111-1111-1111-111111111111", "name": "默认部门", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:05.127268+00:00", "updated_at": "2025-08-04T03:40:05.127268+00:00", "business_id": "100050", "description": null, "employee_count": 0, "parent_unit_id": null}, "operation": "DELETE", "timestamp": "2025-08-04T03:40:27.58836+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.58836+00	2025-08-04 03:40:27.58836+00	\N
443	CREATE	f47ae9de-810c-481d-b266-1e485252548a	{"new_data": {"id": "f47ae9de-810c-481d-b266-1e485252548a", "name": "技术部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100000", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
444	CREATE	0658bdd7-3e4e-4b60-86cc-a115743d81b0	{"new_data": {"id": "0658bdd7-3e4e-4b60-86cc-a115743d81b0", "name": "产品部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100001", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
445	CREATE	8e93fa5f-8fb1-489e-ae22-f12589ac38a9	{"new_data": {"id": "8e93fa5f-8fb1-489e-ae22-f12589ac38a9", "name": "销售部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100002", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
446	CREATE	20c12a1c-2854-45bd-a45f-876f9e77276f	{"new_data": {"id": "20c12a1c-2854-45bd-a45f-876f9e77276f", "name": "人事部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100003", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
447	CREATE	40fd2093-6d02-47d8-a583-a5277480d928	{"new_data": {"id": "40fd2093-6d02-47d8-a583-a5277480d928", "name": "财务部", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100004", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
448	CREATE	370fa14f-1108-4a5e-a8b7-af7914437ddb	{"new_data": {"id": "370fa14f-1108-4a5e-a8b7-af7914437ddb", "name": "技术部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100005", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
449	CREATE	d20a693e-8b6b-41bc-aa91-31f20bee8a9a	{"new_data": {"id": "d20a693e-8b6b-41bc-aa91-31f20bee8a9a", "name": "产品部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100006", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
450	CREATE	8830c7d8-4fda-4079-9175-2fe8ae8b76bf	{"new_data": {"id": "8830c7d8-4fda-4079-9175-2fe8ae8b76bf", "name": "销售部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100007", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
451	CREATE	76e7c802-0cd1-4e18-aa66-9ec43dbcd945	{"new_data": {"id": "76e7c802-0cd1-4e18-aa66-9ec43dbcd945", "name": "人事部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100008", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
452	CREATE	c3d2855c-512c-4c40-9093-0ea620ade03d	{"new_data": {"id": "c3d2855c-512c-4c40-9093-0ea620ade03d", "name": "财务部-2", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100009", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
453	CREATE	4b4bea88-81a9-48c8-895d-932f46acf30d	{"new_data": {"id": "4b4bea88-81a9-48c8-895d-932f46acf30d", "name": "技术部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100010", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
454	CREATE	ec238cdb-e097-4bbd-b8ef-62057d6b6bfb	{"new_data": {"id": "ec238cdb-e097-4bbd-b8ef-62057d6b6bfb", "name": "产品部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100011", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
455	CREATE	f220bd54-c155-425b-b3ee-eafe42cb87d0	{"new_data": {"id": "f220bd54-c155-425b-b3ee-eafe42cb87d0", "name": "销售部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100012", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
456	CREATE	0c2d9ee1-44dd-4bd9-a01f-52f73d26307f	{"new_data": {"id": "0c2d9ee1-44dd-4bd9-a01f-52f73d26307f", "name": "人事部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100013", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
457	CREATE	36883130-de7a-4de0-ae23-f1aa4f7269f0	{"new_data": {"id": "36883130-de7a-4de0-ae23-f1aa4f7269f0", "name": "财务部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100014", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
458	CREATE	fc845f6d-9606-464f-9967-211ca720dcd3	{"new_data": {"id": "fc845f6d-9606-464f-9967-211ca720dcd3", "name": "技术部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100015", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
459	CREATE	6718b207-99e2-4bd8-bc73-b9b52e2ac509	{"new_data": {"id": "6718b207-99e2-4bd8-bc73-b9b52e2ac509", "name": "产品部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100016", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
460	CREATE	c64495dd-daf3-4f5d-83ed-a3be7df1918a	{"new_data": {"id": "c64495dd-daf3-4f5d-83ed-a3be7df1918a", "name": "销售部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100017", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
461	CREATE	7ae44c5f-4fc9-4065-aa67-5ba00f84d4cc	{"new_data": {"id": "7ae44c5f-4fc9-4065-aa67-5ba00f84d4cc", "name": "人事部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100018", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
462	CREATE	45b791a3-f457-4bea-99e6-be2bd52a320f	{"new_data": {"id": "45b791a3-f457-4bea-99e6-be2bd52a320f", "name": "财务部-4", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100019", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
463	CREATE	6e8b018f-e2e6-4bf0-bece-38c0b3ec1cf6	{"new_data": {"id": "6e8b018f-e2e6-4bf0-bece-38c0b3ec1cf6", "name": "技术部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100020", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
464	CREATE	a6d61e23-6df1-4c1e-8a9a-22ab19a033e3	{"new_data": {"id": "a6d61e23-6df1-4c1e-8a9a-22ab19a033e3", "name": "产品部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100021", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
465	CREATE	cc1df976-7ee7-4842-8077-97800c5ac99b	{"new_data": {"id": "cc1df976-7ee7-4842-8077-97800c5ac99b", "name": "销售部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100022", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
466	CREATE	7ac3f5b2-234e-4975-8cb4-56e01cd84bcf	{"new_data": {"id": "7ac3f5b2-234e-4975-8cb4-56e01cd84bcf", "name": "人事部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100023", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
467	CREATE	4ba0341f-512a-4e3a-a05d-0202cdd14957	{"new_data": {"id": "4ba0341f-512a-4e3a-a05d-0202cdd14957", "name": "财务部-5", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100024", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
468	CREATE	6f0b0e67-f889-40ae-a89e-22a9ed7bc857	{"new_data": {"id": "6f0b0e67-f889-40ae-a89e-22a9ed7bc857", "name": "技术部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100025", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
469	CREATE	e4b177de-86d6-4e5b-88ec-674cfaff941b	{"new_data": {"id": "e4b177de-86d6-4e5b-88ec-674cfaff941b", "name": "产品部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100026", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
470	CREATE	fba2f085-3da1-4624-8e6f-4ff6936b99f9	{"new_data": {"id": "fba2f085-3da1-4624-8e6f-4ff6936b99f9", "name": "销售部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100027", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
471	CREATE	86e635c1-073d-4aec-becd-bbda26800119	{"new_data": {"id": "86e635c1-073d-4aec-becd-bbda26800119", "name": "人事部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100028", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
472	CREATE	606f017e-cd7f-49bd-80c7-114c40f05c28	{"new_data": {"id": "606f017e-cd7f-49bd-80c7-114c40f05c28", "name": "财务部-6", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100029", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
473	CREATE	66bbaaff-e244-4286-a8c3-da25bc24b750	{"new_data": {"id": "66bbaaff-e244-4286-a8c3-da25bc24b750", "name": "技术部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100030", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
474	CREATE	f813293c-9796-4ea6-971d-d272cb74e4c9	{"new_data": {"id": "f813293c-9796-4ea6-971d-d272cb74e4c9", "name": "产品部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100031", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
475	CREATE	6137a442-3338-4863-bcc8-2c8e8603aac9	{"new_data": {"id": "6137a442-3338-4863-bcc8-2c8e8603aac9", "name": "销售部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100032", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
476	CREATE	54df8696-571d-42f7-bbc2-c2451339d0a9	{"new_data": {"id": "54df8696-571d-42f7-bbc2-c2451339d0a9", "name": "人事部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100033", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
477	CREATE	1d6f06d6-c7cc-4909-bf4c-520802f60058	{"new_data": {"id": "1d6f06d6-c7cc-4909-bf4c-520802f60058", "name": "财务部-7", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100034", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
478	CREATE	b4d4032e-24a2-43b4-817d-9569c8c6bfff	{"new_data": {"id": "b4d4032e-24a2-43b4-817d-9569c8c6bfff", "name": "技术部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100035", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
479	CREATE	35f25791-673f-4e0a-82a6-15086748fdef	{"new_data": {"id": "35f25791-673f-4e0a-82a6-15086748fdef", "name": "产品部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100036", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
480	CREATE	d3ba3dc5-1bd4-40d9-8b2c-4be745411166	{"new_data": {"id": "d3ba3dc5-1bd4-40d9-8b2c-4be745411166", "name": "销售部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100037", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
481	CREATE	fec61690-e747-4de5-a54b-8b2a429b7f59	{"new_data": {"id": "fec61690-e747-4de5-a54b-8b2a429b7f59", "name": "人事部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100038", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
482	CREATE	ec51d9d2-554d-4e10-b3f1-aae4b48b3a10	{"new_data": {"id": "ec51d9d2-554d-4e10-b3f1-aae4b48b3a10", "name": "财务部-8", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100039", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
483	CREATE	c6bb9f0f-33a6-48b6-9ae3-c35011639008	{"new_data": {"id": "c6bb9f0f-33a6-48b6-9ae3-c35011639008", "name": "技术部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100040", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
484	CREATE	ec7b4783-61d3-4210-a9bf-169dd60d8447	{"new_data": {"id": "ec7b4783-61d3-4210-a9bf-169dd60d8447", "name": "产品部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100041", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
485	CREATE	df202dd6-c2de-41b2-be9c-03de8669a68a	{"new_data": {"id": "df202dd6-c2de-41b2-be9c-03de8669a68a", "name": "销售部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100042", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
486	CREATE	26d44134-ecad-4b16-bc49-c486ff153271	{"new_data": {"id": "26d44134-ecad-4b16-bc49-c486ff153271", "name": "人事部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100043", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
487	CREATE	30113637-9539-480b-804e-fff9cdaa4e15	{"new_data": {"id": "30113637-9539-480b-804e-fff9cdaa4e15", "name": "财务部-9", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100044", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
488	CREATE	ca13be57-075e-46b2-8caf-1ae1fb4cf5e3	{"new_data": {"id": "ca13be57-075e-46b2-8caf-1ae1fb4cf5e3", "name": "技术部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100045", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
489	CREATE	f9a483d6-0b54-4474-9a26-d3a283793601	{"new_data": {"id": "f9a483d6-0b54-4474-9a26-d3a283793601", "name": "产品部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100046", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
490	CREATE	617d749d-eef6-4e07-a444-ede80f769a3f	{"new_data": {"id": "617d749d-eef6-4e07-a444-ede80f769a3f", "name": "销售部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100047", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
491	CREATE	47971292-845f-4b23-bd0c-cc5096221bea	{"new_data": {"id": "47971292-845f-4b23-bd0c-cc5096221bea", "name": "人事部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100048", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
492	CREATE	cd89afbc-05dd-4a08-8e86-be77a39f0a5c	{"new_data": {"id": "cd89afbc-05dd-4a08-8e86-be77a39f0a5c", "name": "财务部-10", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100049", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.769331+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.769331+00	2025-08-04 03:40:27.769331+00	\N
493	CREATE	ab3cdf98-36b9-4292-981a-78bce50fc6b9	{"new_data": {"id": "ab3cdf98-36b9-4292-981a-78bce50fc6b9", "name": "边界组织999999", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.805204+00:00", "updated_at": "2025-08-04T03:40:27.805204+00:00", "business_id": "999999", "description": null, "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.805204+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.805204+00	2025-08-04 03:40:27.805204+00	\N
494	CREATE	11111111-1111-1111-1111-111111111111	{"new_data": {"id": "11111111-1111-1111-1111-111111111111", "name": "默认部门", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.836529+00:00", "updated_at": "2025-08-04T03:40:27.836529+00:00", "business_id": "100050", "description": null, "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-04T03:40:27.836529+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 03:40:27.836529+00	2025-08-04 03:40:27.836529+00	\N
495	UPDATE	ec238cdb-e097-4bbd-b8ef-62057d6b6bfb	{"new_data": {"id": "ec238cdb-e097-4bbd-b8ef-62057d6b6bfb", "name": "高谷集团", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T06:32:28.339383+00:00", "business_id": "100011", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "old_data": {"id": "ec238cdb-e097-4bbd-b8ef-62057d6b6bfb", "name": "产品部-3", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.769331+00:00", "updated_at": "2025-08-04T03:40:27.769331+00:00", "business_id": "100011", "description": "测试部门描述", "employee_count": 0, "parent_unit_id": null}, "operation": "UPDATE", "timestamp": "2025-08-04T06:32:28.339383+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 06:32:28.339383+00	2025-08-04 06:32:28.339383+00	\N
496	UPDATE	ab3cdf98-36b9-4292-981a-78bce50fc6b9	{"new_data": {"id": "ab3cdf98-36b9-4292-981a-78bce50fc6b9", "name": "边界组织999999", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.805204+00:00", "updated_at": "2025-08-04T06:32:36.160609+00:00", "business_id": "999999", "description": null, "employee_count": 0, "parent_unit_id": "ec238cdb-e097-4bbd-b8ef-62057d6b6bfb"}, "old_data": {"id": "ab3cdf98-36b9-4292-981a-78bce50fc6b9", "name": "边界组织999999", "level": 1, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "550e8400-e29b-41d4-a716-446655440000", "unit_type": "DEPARTMENT", "created_at": "2025-08-04T03:40:27.805204+00:00", "updated_at": "2025-08-04T03:40:27.805204+00:00", "business_id": "999999", "description": null, "employee_count": 0, "parent_unit_id": null}, "operation": "UPDATE", "timestamp": "2025-08-04T06:32:36.160609+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-04 06:32:36.160609+00	2025-08-04 06:32:36.160609+00	\N
497	CREATE	0cbe1aad-32e6-4e98-86fc-d22ebe9dac33	{"new_data": {"id": "0cbe1aad-32e6-4e98-86fc-d22ebe9dac33", "name": "高谷集团", "level": 0, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000001", "unit_type": "COMPANY", "created_at": "2025-08-05T04:33:38.085885+00:00", "updated_at": "2025-08-05T04:33:38.085887+00:00", "business_id": "100053", "description": null, "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-05T04:33:38.092738+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-05 04:33:38.092738+00	2025-08-05 04:33:38.092738+00	\N
498	UPDATE	0cbe1aad-32e6-4e98-86fc-d22ebe9dac33	{"new_data": {"id": "0cbe1aad-32e6-4e98-86fc-d22ebe9dac33", "name": "AI治理办公室", "level": 0, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000001", "unit_type": "COMPANY", "created_at": "2025-08-05T04:33:38.085885+00:00", "updated_at": "2025-08-05T04:38:08.043356+00:00", "business_id": "100053", "description": null, "employee_count": 0, "parent_unit_id": null}, "old_data": {"id": "0cbe1aad-32e6-4e98-86fc-d22ebe9dac33", "name": "高谷集团", "level": 0, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000001", "unit_type": "COMPANY", "created_at": "2025-08-05T04:33:38.085885+00:00", "updated_at": "2025-08-05T04:33:38.085887+00:00", "business_id": "100053", "description": null, "employee_count": 0, "parent_unit_id": null}, "operation": "UPDATE", "timestamp": "2025-08-05T04:38:08.043356+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-05 04:38:08.043356+00	2025-08-05 04:38:08.043356+00	\N
499	CREATE	ebf13067-635b-4138-800e-eae8ba1b43ad	{"new_data": {"id": "ebf13067-635b-4138-800e-eae8ba1b43ad", "name": "测试部门", "level": 0, "status": "ACTIVE", "profile": null, "is_active": true, "tenant_id": "00000000-0000-0000-0000-000000000001", "unit_type": "DEPARTMENT", "created_at": "2025-08-05T04:41:57.929855+00:00", "updated_at": "2025-08-05T04:41:57.929856+00:00", "business_id": "100054", "description": null, "employee_count": 0, "parent_unit_id": null}, "operation": "INSERT", "timestamp": "2025-08-05T04:41:57.930954+00:00", "table_name": "organization_units"}	PENDING	\N	0	2025-08-05 04:41:57.930954+00	2025-08-05 04:41:57.930954+00	\N
\.


--
-- Data for Name: workflow_instances; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.workflow_instances (id, tenant_id, workflow_type, current_state, state_history, context, initiated_by, correlation_id, started_at, completed_at, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: workflow_steps; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.workflow_steps (id, tenant_id, step_name, step_type, status, assigned_to, input_data, output_data, due_date, started_at, completed_at, created_at, updated_at, workflow_instance_id) FROM stdin;
\.


--
-- Data for Name: tenant_configs; Type: TABLE DATA; Schema: tenancy; Owner: user
--

COPY tenancy.tenant_configs (id, tenant_id, config_key, config_value, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: tenants; Type: TABLE DATA; Schema: tenancy; Owner: user
--

COPY tenancy.tenants (id, name, domain, status, subscription_plan, max_users, created_at, updated_at) FROM stdin;
99999999-9999-9999-9999-999999999999	测试租户	test.example.com	active	basic	10	2025-07-31 05:00:38.831272+00	2025-07-31 05:00:38.831272+00
\.


--
-- Name: employee_business_id_seq; Type: SEQUENCE SET; Schema: public; Owner: user
--

SELECT pg_catalog.setval('public.employee_business_id_seq', 501, true);


--
-- Name: employee_code_seq; Type: SEQUENCE SET; Schema: public; Owner: user
--

SELECT pg_catalog.setval('public.employee_code_seq', 10000004, true);


--
-- Name: employee_positions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: user
--

SELECT pg_catalog.setval('public.employee_positions_id_seq', 5, true);


--
-- Name: org_business_id_seq; Type: SEQUENCE SET; Schema: public; Owner: user
--

SELECT pg_catalog.setval('public.org_business_id_seq', 54, true);


--
-- Name: org_unit_code_seq; Type: SEQUENCE SET; Schema: public; Owner: user
--

SELECT pg_catalog.setval('public.org_unit_code_seq', 1000004, true);


--
-- Name: position_assignments_id_seq; Type: SEQUENCE SET; Schema: public; Owner: user
--

SELECT pg_catalog.setval('public.position_assignments_id_seq', 1, false);


--
-- Name: position_business_id_seq; Type: SEQUENCE SET; Schema: public; Owner: user
--

SELECT pg_catalog.setval('public.position_business_id_seq', 101, true);


--
-- Name: sync_monitoring_id_seq; Type: SEQUENCE SET; Schema: public; Owner: user
--

SELECT pg_catalog.setval('public.sync_monitoring_id_seq', 499, true);


--
-- Name: employees employees_email_key; Type: CONSTRAINT; Schema: corehr; Owner: user
--

ALTER TABLE ONLY corehr.employees
    ADD CONSTRAINT employees_email_key UNIQUE (email);


--
-- Name: employees employees_employee_number_key; Type: CONSTRAINT; Schema: corehr; Owner: user
--

ALTER TABLE ONLY corehr.employees
    ADD CONSTRAINT employees_employee_number_key UNIQUE (employee_number);


--
-- Name: employees employees_pkey; Type: CONSTRAINT; Schema: corehr; Owner: user
--

ALTER TABLE ONLY corehr.employees
    ADD CONSTRAINT employees_pkey PRIMARY KEY (id);


--
-- Name: organizations organizations_code_key; Type: CONSTRAINT; Schema: corehr; Owner: user
--

ALTER TABLE ONLY corehr.organizations
    ADD CONSTRAINT organizations_code_key UNIQUE (code);


--
-- Name: organizations organizations_pkey; Type: CONSTRAINT; Schema: corehr; Owner: user
--

ALTER TABLE ONLY corehr.organizations
    ADD CONSTRAINT organizations_pkey PRIMARY KEY (id);


--
-- Name: positions positions_code_key; Type: CONSTRAINT; Schema: corehr; Owner: user
--

ALTER TABLE ONLY corehr.positions
    ADD CONSTRAINT positions_code_key UNIQUE (code);


--
-- Name: positions positions_pkey; Type: CONSTRAINT; Schema: corehr; Owner: user
--

ALTER TABLE ONLY corehr.positions
    ADD CONSTRAINT positions_pkey PRIMARY KEY (id);


--
-- Name: permissions permissions_pkey; Type: CONSTRAINT; Schema: identity; Owner: user
--

ALTER TABLE ONLY identity.permissions
    ADD CONSTRAINT permissions_pkey PRIMARY KEY (id);


--
-- Name: permissions permissions_tenant_id_resource_action_key; Type: CONSTRAINT; Schema: identity; Owner: user
--

ALTER TABLE ONLY identity.permissions
    ADD CONSTRAINT permissions_tenant_id_resource_action_key UNIQUE (tenant_id, resource, action);


--
-- Name: role_permissions role_permissions_pkey; Type: CONSTRAINT; Schema: identity; Owner: user
--

ALTER TABLE ONLY identity.role_permissions
    ADD CONSTRAINT role_permissions_pkey PRIMARY KEY (role_id, permission_id);


--
-- Name: roles roles_pkey; Type: CONSTRAINT; Schema: identity; Owner: user
--

ALTER TABLE ONLY identity.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);


--
-- Name: roles roles_tenant_id_name_key; Type: CONSTRAINT; Schema: identity; Owner: user
--

ALTER TABLE ONLY identity.roles
    ADD CONSTRAINT roles_tenant_id_name_key UNIQUE (tenant_id, name);


--
-- Name: user_roles user_roles_pkey; Type: CONSTRAINT; Schema: identity; Owner: user
--

ALTER TABLE ONLY identity.user_roles
    ADD CONSTRAINT user_roles_pkey PRIMARY KEY (user_id, role_id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: identity; Owner: user
--

ALTER TABLE ONLY identity.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: identity; Owner: user
--

ALTER TABLE ONLY identity.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_username_key; Type: CONSTRAINT; Schema: identity; Owner: user
--

ALTER TABLE ONLY identity.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- Name: conversations conversations_pkey; Type: CONSTRAINT; Schema: intelligence; Owner: user
--

ALTER TABLE ONLY intelligence.conversations
    ADD CONSTRAINT conversations_pkey PRIMARY KEY (id);


--
-- Name: messages messages_pkey; Type: CONSTRAINT; Schema: intelligence; Owner: user
--

ALTER TABLE ONLY intelligence.messages
    ADD CONSTRAINT messages_pkey PRIMARY KEY (id);


--
-- Name: events events_pkey; Type: CONSTRAINT; Schema: outbox; Owner: user
--

ALTER TABLE ONLY outbox.events
    ADD CONSTRAINT events_pkey PRIMARY KEY (id);


--
-- Name: assignment_details assignment_details_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.assignment_details
    ADD CONSTRAINT assignment_details_pkey PRIMARY KEY (id);


--
-- Name: assignment_history assignment_history_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.assignment_history
    ADD CONSTRAINT assignment_history_pkey PRIMARY KEY (id);


--
-- Name: business_process_events business_process_events_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.business_process_events
    ADD CONSTRAINT business_process_events_pkey PRIMARY KEY (id);


--
-- Name: employee_positions employee_positions_employee_code_position_code_assignment_t_key; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.employee_positions
    ADD CONSTRAINT employee_positions_employee_code_position_code_assignment_t_key UNIQUE (employee_code, position_code, assignment_type, start_date);


--
-- Name: employee_positions employee_positions_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.employee_positions
    ADD CONSTRAINT employee_positions_pkey PRIMARY KEY (id);


--
-- Name: employees employees_email_tenant_id_key; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.employees
    ADD CONSTRAINT employees_email_tenant_id_key UNIQUE (email, tenant_id);


--
-- Name: employees employees_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.employees
    ADD CONSTRAINT employees_pkey PRIMARY KEY (code);


--
-- Name: metacontract_editor_projects metacontract_editor_projects_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.metacontract_editor_projects
    ADD CONSTRAINT metacontract_editor_projects_pkey PRIMARY KEY (id);


--
-- Name: metacontract_editor_sessions metacontract_editor_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.metacontract_editor_sessions
    ADD CONSTRAINT metacontract_editor_sessions_pkey PRIMARY KEY (id);


--
-- Name: metacontract_editor_settings metacontract_editor_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.metacontract_editor_settings
    ADD CONSTRAINT metacontract_editor_settings_pkey PRIMARY KEY (user_id);


--
-- Name: metacontract_editor_templates metacontract_editor_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.metacontract_editor_templates
    ADD CONSTRAINT metacontract_editor_templates_pkey PRIMARY KEY (id);


--
-- Name: organization_units organization_units_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.organization_units
    ADD CONSTRAINT organization_units_pkey PRIMARY KEY (code);


--
-- Name: outbox_events outbox_events_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.outbox_events
    ADD CONSTRAINT outbox_events_pkey PRIMARY KEY (id);


--
-- Name: person person_email_key; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.person
    ADD CONSTRAINT person_email_key UNIQUE (email);


--
-- Name: person person_employee_id_key; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.person
    ADD CONSTRAINT person_employee_id_key UNIQUE (employee_id);


--
-- Name: person person_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.person
    ADD CONSTRAINT person_pkey PRIMARY KEY (id);


--
-- Name: position_assignments position_assignments_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.position_assignments
    ADD CONSTRAINT position_assignments_pkey PRIMARY KEY (id);


--
-- Name: position_attribute_histories position_attribute_histories_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.position_attribute_histories
    ADD CONSTRAINT position_attribute_histories_pkey PRIMARY KEY (id);


--
-- Name: position_code_sequence position_code_sequence_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.position_code_sequence
    ADD CONSTRAINT position_code_sequence_pkey PRIMARY KEY (tenant_id);


--
-- Name: position_histories position_histories_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.position_histories
    ADD CONSTRAINT position_histories_pkey PRIMARY KEY (id);


--
-- Name: position_history position_history_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.position_history
    ADD CONSTRAINT position_history_pkey PRIMARY KEY (id);


--
-- Name: position_occupancy_histories position_occupancy_histories_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.position_occupancy_histories
    ADD CONSTRAINT position_occupancy_histories_pkey PRIMARY KEY (id);


--
-- Name: positions positions_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.positions
    ADD CONSTRAINT positions_pkey PRIMARY KEY (code);


--
-- Name: sync_monitoring sync_monitoring_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.sync_monitoring
    ADD CONSTRAINT sync_monitoring_pkey PRIMARY KEY (id);


--
-- Name: organization_units uk_tenant_code; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.organization_units
    ADD CONSTRAINT uk_tenant_code UNIQUE (tenant_id, code);


--
-- Name: organization_units uk_tenant_name; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.organization_units
    ADD CONSTRAINT uk_tenant_name UNIQUE (tenant_id, name);


--
-- Name: workflow_instances workflow_instances_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.workflow_instances
    ADD CONSTRAINT workflow_instances_pkey PRIMARY KEY (id);


--
-- Name: workflow_steps workflow_steps_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.workflow_steps
    ADD CONSTRAINT workflow_steps_pkey PRIMARY KEY (id);


--
-- Name: tenant_configs tenant_configs_pkey; Type: CONSTRAINT; Schema: tenancy; Owner: user
--

ALTER TABLE ONLY tenancy.tenant_configs
    ADD CONSTRAINT tenant_configs_pkey PRIMARY KEY (id);


--
-- Name: tenant_configs tenant_configs_tenant_id_config_key_key; Type: CONSTRAINT; Schema: tenancy; Owner: user
--

ALTER TABLE ONLY tenancy.tenant_configs
    ADD CONSTRAINT tenant_configs_tenant_id_config_key_key UNIQUE (tenant_id, config_key);


--
-- Name: tenants tenants_domain_key; Type: CONSTRAINT; Schema: tenancy; Owner: user
--

ALTER TABLE ONLY tenancy.tenants
    ADD CONSTRAINT tenants_domain_key UNIQUE (domain);


--
-- Name: tenants tenants_pkey; Type: CONSTRAINT; Schema: tenancy; Owner: user
--

ALTER TABLE ONLY tenancy.tenants
    ADD CONSTRAINT tenants_pkey PRIMARY KEY (id);


--
-- Name: idx_employees_manager_id; Type: INDEX; Schema: corehr; Owner: user
--

CREATE INDEX idx_employees_manager_id ON corehr.employees USING btree (manager_id);


--
-- Name: idx_employees_tenant_id; Type: INDEX; Schema: corehr; Owner: user
--

CREATE INDEX idx_employees_tenant_id ON corehr.employees USING btree (tenant_id);


--
-- Name: idx_organizations_parent_id; Type: INDEX; Schema: corehr; Owner: user
--

CREATE INDEX idx_organizations_parent_id ON corehr.organizations USING btree (parent_id);


--
-- Name: idx_organizations_tenant_id; Type: INDEX; Schema: corehr; Owner: user
--

CREATE INDEX idx_organizations_tenant_id ON corehr.organizations USING btree (tenant_id);


--
-- Name: idx_permissions_tenant_id; Type: INDEX; Schema: identity; Owner: user
--

CREATE INDEX idx_permissions_tenant_id ON identity.permissions USING btree (tenant_id);


--
-- Name: idx_roles_tenant_id; Type: INDEX; Schema: identity; Owner: user
--

CREATE INDEX idx_roles_tenant_id ON identity.roles USING btree (tenant_id);


--
-- Name: idx_users_employee_id; Type: INDEX; Schema: identity; Owner: user
--

CREATE INDEX idx_users_employee_id ON identity.users USING btree (employee_id);


--
-- Name: idx_users_tenant_id; Type: INDEX; Schema: identity; Owner: user
--

CREATE INDEX idx_users_tenant_id ON identity.users USING btree (tenant_id);


--
-- Name: idx_conversations_session_id; Type: INDEX; Schema: intelligence; Owner: user
--

CREATE INDEX idx_conversations_session_id ON intelligence.conversations USING btree (session_id);


--
-- Name: idx_outbox_events_aggregate; Type: INDEX; Schema: outbox; Owner: user
--

CREATE INDEX idx_outbox_events_aggregate ON outbox.events USING btree (aggregate_id, aggregate_type);


--
-- Name: idx_outbox_events_processed; Type: INDEX; Schema: outbox; Owner: user
--

CREATE INDEX idx_outbox_events_processed ON outbox.events USING btree (processed_at) WHERE (processed_at IS NULL);


--
-- Name: businessprocessevent_correlation_id; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX businessprocessevent_correlation_id ON public.business_process_events USING btree (correlation_id);


--
-- Name: businessprocessevent_status; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX businessprocessevent_status ON public.business_process_events USING btree (status);


--
-- Name: businessprocessevent_tenant_id_effective_date; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX businessprocessevent_tenant_id_effective_date ON public.business_process_events USING btree (tenant_id, effective_date);


--
-- Name: businessprocessevent_tenant_id_entity_type_entity_id; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX businessprocessevent_tenant_id_entity_type_entity_id ON public.business_process_events USING btree (tenant_id, entity_type, entity_id);


--
-- Name: businessprocessevent_tenant_id_event_type; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX businessprocessevent_tenant_id_event_type ON public.business_process_events USING btree (tenant_id, event_type);


--
-- Name: idx_assignment_details_approver; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_assignment_details_approver ON public.assignment_details USING btree (approved_by);


--
-- Name: idx_assignment_details_assignment; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_assignment_details_assignment ON public.assignment_details USING btree (assignment_id);


--
-- Name: idx_assignment_details_effective; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_assignment_details_effective ON public.assignment_details USING btree (effective_date);


--
-- Name: idx_assignment_details_status; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_assignment_details_status ON public.assignment_details USING btree (approval_status);


--
-- Name: idx_assignment_history_assignment; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_assignment_history_assignment ON public.assignment_history USING btree (assignment_id);


--
-- Name: idx_assignment_history_changed_by; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_assignment_history_changed_by ON public.assignment_history USING btree (changed_by);


--
-- Name: idx_assignment_history_date; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_assignment_history_date ON public.assignment_history USING btree (effective_date);


--
-- Name: idx_assignment_history_type; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_assignment_history_type ON public.assignment_history USING btree (change_type);


--
-- Name: idx_emp_pos_active; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_emp_pos_active ON public.employee_positions USING btree (employee_code, status) WHERE ((status)::text = 'ACTIVE'::text);


--
-- Name: idx_emp_pos_assignment; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_emp_pos_assignment ON public.employee_positions USING btree (assignment_type);


--
-- Name: idx_emp_pos_dates; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_emp_pos_dates ON public.employee_positions USING btree (start_date, end_date);


--
-- Name: idx_emp_pos_employee; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_emp_pos_employee ON public.employee_positions USING btree (employee_code);


--
-- Name: idx_emp_pos_position; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_emp_pos_position ON public.employee_positions USING btree (position_code);


--
-- Name: idx_emp_pos_primary; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_emp_pos_primary ON public.employee_positions USING btree (employee_code, assignment_type) WHERE ((assignment_type)::text = 'PRIMARY'::text);


--
-- Name: idx_emp_pos_status; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_emp_pos_status ON public.employee_positions USING btree (status);


--
-- Name: idx_employees_active; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_employees_active ON public.employees USING btree (employment_status) WHERE ((employment_status)::text = 'ACTIVE'::text);


--
-- Name: idx_employees_email; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_employees_email ON public.employees USING btree (email);


--
-- Name: idx_employees_hire_date; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_employees_hire_date ON public.employees USING btree (hire_date);


--
-- Name: idx_employees_name; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_employees_name ON public.employees USING btree (first_name, last_name);


--
-- Name: idx_employees_org_status; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_employees_org_status ON public.employees USING btree (organization_code, employment_status);


--
-- Name: idx_employees_organization; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_employees_organization ON public.employees USING btree (organization_code);


--
-- Name: idx_employees_primary_position; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_employees_primary_position ON public.employees USING btree (primary_position_code);


--
-- Name: idx_employees_status; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_employees_status ON public.employees USING btree (employment_status);


--
-- Name: idx_employees_tenant; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_employees_tenant ON public.employees USING btree (tenant_id);


--
-- Name: idx_employees_type; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_employees_type ON public.employees USING btree (employee_type);


--
-- Name: idx_employees_type_status; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_employees_type_status ON public.employees USING btree (employee_type, employment_status);


--
-- Name: idx_metacontract_projects_created_by; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_metacontract_projects_created_by ON public.metacontract_editor_projects USING btree (created_by);


--
-- Name: idx_metacontract_projects_status; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_metacontract_projects_status ON public.metacontract_editor_projects USING btree (status);


--
-- Name: idx_metacontract_projects_tenant_id; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_metacontract_projects_tenant_id ON public.metacontract_editor_projects USING btree (tenant_id);


--
-- Name: idx_metacontract_projects_updated_at; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_metacontract_projects_updated_at ON public.metacontract_editor_projects USING btree (updated_at);


--
-- Name: idx_metacontract_sessions_active; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_metacontract_sessions_active ON public.metacontract_editor_sessions USING btree (active);


--
-- Name: idx_metacontract_sessions_project_id; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_metacontract_sessions_project_id ON public.metacontract_editor_sessions USING btree (project_id);


--
-- Name: idx_metacontract_sessions_user_id; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_metacontract_sessions_user_id ON public.metacontract_editor_sessions USING btree (user_id);


--
-- Name: idx_metacontract_templates_category; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_metacontract_templates_category ON public.metacontract_editor_templates USING btree (category);


--
-- Name: idx_metacontract_templates_tags; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_metacontract_templates_tags ON public.metacontract_editor_templates USING gin (tags);


--
-- Name: idx_org_units_name_gin; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_org_units_name_gin ON public.organization_units USING gin (name public.gin_trgm_ops);


--
-- Name: idx_org_units_parent_code; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_org_units_parent_code ON public.organization_units USING btree (parent_code);


--
-- Name: idx_org_units_path_gin; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_org_units_path_gin ON public.organization_units USING gin (path public.gin_trgm_ops);


--
-- Name: idx_org_units_tenant_status; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_org_units_tenant_status ON public.organization_units USING btree (tenant_id, status);


--
-- Name: idx_org_units_type_level; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_org_units_type_level ON public.organization_units USING btree (unit_type, level);


--
-- Name: idx_outbox_events_aggregate; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_outbox_events_aggregate ON public.outbox_events USING btree (aggregate_id);


--
-- Name: idx_outbox_events_created; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_outbox_events_created ON public.outbox_events USING btree (created_at);


--
-- Name: idx_outbox_events_status; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_outbox_events_status ON public.outbox_events USING btree (status);


--
-- Name: idx_outbox_events_tenant; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_outbox_events_tenant ON public.outbox_events USING btree (tenant_id);


--
-- Name: idx_outbox_events_tenant_status; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_outbox_events_tenant_status ON public.outbox_events USING btree (tenant_id, status);


--
-- Name: idx_outbox_events_type; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_outbox_events_type ON public.outbox_events USING btree (event_type);


--
-- Name: idx_person_email; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_person_email ON public.person USING btree (email);


--
-- Name: idx_person_employee_id; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_person_employee_id ON public.person USING btree (employee_id);


--
-- Name: idx_person_tenant_id; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_person_tenant_id ON public.person USING btree (tenant_id);


--
-- Name: idx_position_assignments_employee; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_position_assignments_employee ON public.position_assignments USING btree (employee_code);


--
-- Name: idx_position_assignments_position; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_position_assignments_position ON public.position_assignments USING btree (position_code);


--
-- Name: idx_position_history_current; Type: INDEX; Schema: public; Owner: user
--

CREATE UNIQUE INDEX idx_position_history_current ON public.position_history USING btree (tenant_id, employee_id) WHERE (end_date IS NULL);


--
-- Name: INDEX idx_position_history_current; Type: COMMENT; Schema: public; Owner: user
--

COMMENT ON INDEX public.idx_position_history_current IS 'Optimized index for current position lookups';


--
-- Name: idx_position_history_date_range; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_position_history_date_range ON public.position_history USING btree (tenant_id, employee_id, effective_date DESC, end_date DESC);


--
-- Name: INDEX idx_position_history_date_range; Type: COMMENT; Schema: public; Owner: user
--

COMMENT ON INDEX public.idx_position_history_date_range IS 'Composite index for efficient timeline queries';


--
-- Name: idx_position_history_effective_date; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_position_history_effective_date ON public.position_history USING btree (tenant_id, effective_date);


--
-- Name: idx_position_history_reports_to; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_position_history_reports_to ON public.position_history USING btree (tenant_id, reports_to_employee_id, effective_date) WHERE (end_date IS NULL);


--
-- Name: idx_position_history_retroactive; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_position_history_retroactive ON public.position_history USING btree (tenant_id, is_retroactive, created_at);


--
-- Name: idx_position_history_temporal; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_position_history_temporal ON public.position_history USING btree (tenant_id, employee_id, effective_date, end_date);


--
-- Name: INDEX idx_position_history_temporal; Type: COMMENT; Schema: public; Owner: user
--

COMMENT ON INDEX public.idx_position_history_temporal IS 'Primary temporal index for efficient as-of-date and range queries';


--
-- Name: idx_positions_job_profile; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_positions_job_profile ON public.positions USING btree (job_profile_id);


--
-- Name: idx_positions_manager; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_positions_manager ON public.positions USING btree (manager_position_code);


--
-- Name: idx_positions_org_status; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_positions_org_status ON public.positions USING btree (organization_code, status);


--
-- Name: idx_positions_organization; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_positions_organization ON public.positions USING btree (organization_code);


--
-- Name: idx_positions_status; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_positions_status ON public.positions USING btree (status);


--
-- Name: idx_positions_tenant; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_positions_tenant ON public.positions USING btree (tenant_id);


--
-- Name: idx_positions_tenant_org; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_positions_tenant_org ON public.positions USING btree (tenant_id, organization_code);


--
-- Name: idx_positions_type; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_positions_type ON public.positions USING btree (position_type);


--
-- Name: idx_positions_type_status; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_positions_type_status ON public.positions USING btree (position_type, status);


--
-- Name: idx_positions_updated; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_positions_updated ON public.positions USING btree (updated_at);


--
-- Name: idx_sync_monitoring_created_at; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_sync_monitoring_created_at ON public.sync_monitoring USING btree (created_at);


--
-- Name: idx_sync_monitoring_entity_id; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_sync_monitoring_entity_id ON public.sync_monitoring USING btree (entity_id);


--
-- Name: idx_sync_monitoring_status; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_sync_monitoring_status ON public.sync_monitoring USING btree (sync_status);


--
-- Name: positionattributehistory_changed_by_created_at; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX positionattributehistory_changed_by_created_at ON public.position_attribute_histories USING btree (changed_by, created_at);


--
-- Name: positionattributehistory_department_id_effective_date; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX positionattributehistory_department_id_effective_date ON public.position_attribute_histories USING btree (department_id, effective_date);


--
-- Name: positionattributehistory_effective_date_end_date; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX positionattributehistory_effective_date_end_date ON public.position_attribute_histories USING btree (effective_date, end_date);


--
-- Name: positionattributehistory_position_id_effective_date; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX positionattributehistory_position_id_effective_date ON public.position_attribute_histories USING btree (position_id, effective_date);


--
-- Name: positionattributehistory_position_id_status_effective_date; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX positionattributehistory_position_id_status_effective_date ON public.position_attribute_histories USING btree (position_id, status, effective_date);


--
-- Name: positionattributehistory_source_event_id; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX positionattributehistory_source_event_id ON public.position_attribute_histories USING btree (source_event_id);


--
-- Name: positionattributehistory_tenant_id; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX positionattributehistory_tenant_id ON public.position_attribute_histories USING btree (tenant_id);


--
-- Name: positionoccupancyhistory_approved_by_approval_date; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX positionoccupancyhistory_approved_by_approval_date ON public.position_occupancy_histories USING btree (approved_by, approval_date);


--
-- Name: positionoccupancyhistory_assignment_type_start_date; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX positionoccupancyhistory_assignment_type_start_date ON public.position_occupancy_histories USING btree (assignment_type, start_date);


--
-- Name: positionoccupancyhistory_employee_id_start_date; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX positionoccupancyhistory_employee_id_start_date ON public.position_occupancy_histories USING btree (employee_id, start_date);


--
-- Name: positionoccupancyhistory_employee_id_start_date_end_date; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX positionoccupancyhistory_employee_id_start_date_end_date ON public.position_occupancy_histories USING btree (employee_id, start_date, end_date);


--
-- Name: positionoccupancyhistory_performance_review_cycle_start_date; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX positionoccupancyhistory_performance_review_cycle_start_date ON public.position_occupancy_histories USING btree (performance_review_cycle, start_date);


--
-- Name: positionoccupancyhistory_position_id_fte_percentage_start_date; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX positionoccupancyhistory_position_id_fte_percentage_start_date ON public.position_occupancy_histories USING btree (position_id, fte_percentage, start_date);


--
-- Name: positionoccupancyhistory_position_id_start_date; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX positionoccupancyhistory_position_id_start_date ON public.position_occupancy_histories USING btree (position_id, start_date);


--
-- Name: positionoccupancyhistory_source_event_id; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX positionoccupancyhistory_source_event_id ON public.position_occupancy_histories USING btree (source_event_id);


--
-- Name: positionoccupancyhistory_start_date_end_date; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX positionoccupancyhistory_start_date_end_date ON public.position_occupancy_histories USING btree (start_date, end_date);


--
-- Name: positionoccupancyhistory_tenant_id; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX positionoccupancyhistory_tenant_id ON public.position_occupancy_histories USING btree (tenant_id);


--
-- Name: workflowinstance_correlation_id; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX workflowinstance_correlation_id ON public.workflow_instances USING btree (correlation_id);


--
-- Name: workflowinstance_initiated_by; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX workflowinstance_initiated_by ON public.workflow_instances USING btree (initiated_by);


--
-- Name: workflowinstance_started_at; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX workflowinstance_started_at ON public.workflow_instances USING btree (started_at);


--
-- Name: workflowinstance_tenant_id_current_state; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX workflowinstance_tenant_id_current_state ON public.workflow_instances USING btree (tenant_id, current_state);


--
-- Name: workflowinstance_tenant_id_workflow_type; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX workflowinstance_tenant_id_workflow_type ON public.workflow_instances USING btree (tenant_id, workflow_type);


--
-- Name: workflowstep_assigned_to_status; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX workflowstep_assigned_to_status ON public.workflow_steps USING btree (assigned_to, status);


--
-- Name: workflowstep_due_date; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX workflowstep_due_date ON public.workflow_steps USING btree (due_date);


--
-- Name: workflowstep_step_type; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX workflowstep_step_type ON public.workflow_steps USING btree (step_type);


--
-- Name: workflowstep_tenant_id_status; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX workflowstep_tenant_id_status ON public.workflow_steps USING btree (tenant_id, status);


--
-- Name: workflowstep_tenant_id_workflow_instance_id; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX workflowstep_tenant_id_workflow_instance_id ON public.workflow_steps USING btree (tenant_id, workflow_instance_id);


--
-- Name: employees update_employees_updated_at; Type: TRIGGER; Schema: corehr; Owner: user
--

CREATE TRIGGER update_employees_updated_at BEFORE UPDATE ON corehr.employees FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: organizations update_organizations_updated_at; Type: TRIGGER; Schema: corehr; Owner: user
--

CREATE TRIGGER update_organizations_updated_at BEFORE UPDATE ON corehr.organizations FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: positions update_positions_updated_at; Type: TRIGGER; Schema: corehr; Owner: user
--

CREATE TRIGGER update_positions_updated_at BEFORE UPDATE ON corehr.positions FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: roles update_roles_updated_at; Type: TRIGGER; Schema: identity; Owner: user
--

CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON identity.roles FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: users update_users_updated_at; Type: TRIGGER; Schema: identity; Owner: user
--

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON identity.users FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: conversations update_conversations_updated_at; Type: TRIGGER; Schema: intelligence; Owner: user
--

CREATE TRIGGER update_conversations_updated_at BEFORE UPDATE ON intelligence.conversations FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: employees employee_code_trigger; Type: TRIGGER; Schema: public; Owner: user
--

CREATE TRIGGER employee_code_trigger BEFORE INSERT ON public.employees FOR EACH ROW EXECUTE FUNCTION public.generate_employee_code();


--
-- Name: employee_positions employee_positions_updated_at_trigger; Type: TRIGGER; Schema: public; Owner: user
--

CREATE TRIGGER employee_positions_updated_at_trigger BEFORE UPDATE ON public.employee_positions FOR EACH ROW EXECUTE FUNCTION public.update_employee_updated_at();


--
-- Name: employees employee_updated_at_trigger; Type: TRIGGER; Schema: public; Owner: user
--

CREATE TRIGGER employee_updated_at_trigger BEFORE UPDATE ON public.employees FOR EACH ROW EXECUTE FUNCTION public.update_employee_updated_at();


--
-- Name: organization_units organization_units_change_trigger; Type: TRIGGER; Schema: public; Owner: user
--

CREATE TRIGGER organization_units_change_trigger AFTER INSERT OR DELETE OR UPDATE ON public.organization_units FOR EACH ROW EXECUTE FUNCTION public.notify_organization_change();


--
-- Name: organization_units set_org_unit_code; Type: TRIGGER; Schema: public; Owner: user
--

CREATE TRIGGER set_org_unit_code BEFORE INSERT OR UPDATE ON public.organization_units FOR EACH ROW EXECUTE FUNCTION public.generate_org_unit_code();


--
-- Name: position_history trigger_auto_close_previous_positions; Type: TRIGGER; Schema: public; Owner: user
--

CREATE TRIGGER trigger_auto_close_previous_positions AFTER INSERT ON public.position_history FOR EACH ROW EXECUTE FUNCTION public.auto_close_previous_positions();


--
-- Name: positions trigger_auto_position_code; Type: TRIGGER; Schema: public; Owner: user
--

CREATE TRIGGER trigger_auto_position_code BEFORE INSERT ON public.positions FOR EACH ROW EXECUTE FUNCTION public.auto_generate_position_code();


--
-- Name: positions trigger_update_position_timestamp; Type: TRIGGER; Schema: public; Owner: user
--

CREATE TRIGGER trigger_update_position_timestamp BEFORE UPDATE ON public.positions FOR EACH ROW EXECUTE FUNCTION public.update_position_updated_at();


--
-- Name: position_history trigger_validate_position_history_temporal_consistency; Type: TRIGGER; Schema: public; Owner: user
--

CREATE TRIGGER trigger_validate_position_history_temporal_consistency BEFORE INSERT OR UPDATE ON public.position_history FOR EACH ROW EXECUTE FUNCTION public.validate_position_history_temporal_consistency();


--
-- Name: assignment_details update_assignment_details_updated_at; Type: TRIGGER; Schema: public; Owner: user
--

CREATE TRIGGER update_assignment_details_updated_at BEFORE UPDATE ON public.assignment_details FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: organization_units update_organization_units_updated_at; Type: TRIGGER; Schema: public; Owner: user
--

CREATE TRIGGER update_organization_units_updated_at BEFORE UPDATE ON public.organization_units FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: sync_monitoring update_sync_monitoring_updated_at_trigger; Type: TRIGGER; Schema: public; Owner: user
--

CREATE TRIGGER update_sync_monitoring_updated_at_trigger BEFORE UPDATE ON public.sync_monitoring FOR EACH ROW EXECUTE FUNCTION public.update_sync_monitoring_updated_at();


--
-- Name: tenant_configs update_tenant_configs_updated_at; Type: TRIGGER; Schema: tenancy; Owner: user
--

CREATE TRIGGER update_tenant_configs_updated_at BEFORE UPDATE ON tenancy.tenant_configs FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: tenants update_tenants_updated_at; Type: TRIGGER; Schema: tenancy; Owner: user
--

CREATE TRIGGER update_tenants_updated_at BEFORE UPDATE ON tenancy.tenants FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: employees employees_manager_id_fkey; Type: FK CONSTRAINT; Schema: corehr; Owner: user
--

ALTER TABLE ONLY corehr.employees
    ADD CONSTRAINT employees_manager_id_fkey FOREIGN KEY (manager_id) REFERENCES corehr.employees(id);


--
-- Name: positions positions_department_id_fkey; Type: FK CONSTRAINT; Schema: corehr; Owner: user
--

ALTER TABLE ONLY corehr.positions
    ADD CONSTRAINT positions_department_id_fkey FOREIGN KEY (department_id) REFERENCES corehr.organizations(id);


--
-- Name: role_permissions role_permissions_permission_id_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: user
--

ALTER TABLE ONLY identity.role_permissions
    ADD CONSTRAINT role_permissions_permission_id_fkey FOREIGN KEY (permission_id) REFERENCES identity.permissions(id) ON DELETE CASCADE;


--
-- Name: role_permissions role_permissions_role_id_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: user
--

ALTER TABLE ONLY identity.role_permissions
    ADD CONSTRAINT role_permissions_role_id_fkey FOREIGN KEY (role_id) REFERENCES identity.roles(id) ON DELETE CASCADE;


--
-- Name: user_roles user_roles_role_id_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: user
--

ALTER TABLE ONLY identity.user_roles
    ADD CONSTRAINT user_roles_role_id_fkey FOREIGN KEY (role_id) REFERENCES identity.roles(id) ON DELETE CASCADE;


--
-- Name: user_roles user_roles_user_id_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: user
--

ALTER TABLE ONLY identity.user_roles
    ADD CONSTRAINT user_roles_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id) ON DELETE CASCADE;


--
-- Name: users users_employee_id_fkey; Type: FK CONSTRAINT; Schema: identity; Owner: user
--

ALTER TABLE ONLY identity.users
    ADD CONSTRAINT users_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES corehr.employees(id);


--
-- Name: conversations conversations_user_id_fkey; Type: FK CONSTRAINT; Schema: intelligence; Owner: user
--

ALTER TABLE ONLY intelligence.conversations
    ADD CONSTRAINT conversations_user_id_fkey FOREIGN KEY (user_id) REFERENCES identity.users(id);


--
-- Name: messages messages_conversation_id_fkey; Type: FK CONSTRAINT; Schema: intelligence; Owner: user
--

ALTER TABLE ONLY intelligence.messages
    ADD CONSTRAINT messages_conversation_id_fkey FOREIGN KEY (conversation_id) REFERENCES intelligence.conversations(id) ON DELETE CASCADE;


--
-- Name: employee_positions employee_positions_employee_code_fkey; Type: FK CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.employee_positions
    ADD CONSTRAINT employee_positions_employee_code_fkey FOREIGN KEY (employee_code) REFERENCES public.employees(code) ON DELETE CASCADE;


--
-- Name: employee_positions employee_positions_position_code_fkey; Type: FK CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.employee_positions
    ADD CONSTRAINT employee_positions_position_code_fkey FOREIGN KEY (position_code) REFERENCES public.positions(code) ON DELETE CASCADE;


--
-- Name: employees employees_organization_code_fkey; Type: FK CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.employees
    ADD CONSTRAINT employees_organization_code_fkey FOREIGN KEY (organization_code) REFERENCES public.organization_units(code) ON DELETE RESTRICT;


--
-- Name: employees employees_primary_position_code_fkey; Type: FK CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.employees
    ADD CONSTRAINT employees_primary_position_code_fkey FOREIGN KEY (primary_position_code) REFERENCES public.positions(code) ON DELETE SET NULL;


--
-- Name: organization_units fk_parent_code; Type: FK CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.organization_units
    ADD CONSTRAINT fk_parent_code FOREIGN KEY (parent_code) REFERENCES public.organization_units(code) ON DELETE CASCADE;


--
-- Name: position_assignments fk_position_assignments_position; Type: FK CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.position_assignments
    ADD CONSTRAINT fk_position_assignments_position FOREIGN KEY (position_code) REFERENCES public.positions(code) ON DELETE CASCADE;


--
-- Name: positions fk_positions_manager; Type: FK CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.positions
    ADD CONSTRAINT fk_positions_manager FOREIGN KEY (manager_position_code) REFERENCES public.positions(code);


--
-- Name: positions fk_positions_organization; Type: FK CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.positions
    ADD CONSTRAINT fk_positions_organization FOREIGN KEY (organization_code) REFERENCES public.organization_units(code);


--
-- Name: metacontract_editor_sessions metacontract_editor_sessions_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.metacontract_editor_sessions
    ADD CONSTRAINT metacontract_editor_sessions_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.metacontract_editor_projects(id) ON DELETE CASCADE;


--
-- Name: workflow_steps workflow_steps_workflow_instances_steps; Type: FK CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.workflow_steps
    ADD CONSTRAINT workflow_steps_workflow_instances_steps FOREIGN KEY (workflow_instance_id) REFERENCES public.workflow_instances(id);


--
-- Name: tenant_configs tenant_configs_tenant_id_fkey; Type: FK CONSTRAINT; Schema: tenancy; Owner: user
--

ALTER TABLE ONLY tenancy.tenant_configs
    ADD CONSTRAINT tenant_configs_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenancy.tenants(id) ON DELETE CASCADE;


--
-- Name: employees; Type: ROW SECURITY; Schema: corehr; Owner: user
--

ALTER TABLE corehr.employees ENABLE ROW LEVEL SECURITY;

--
-- Name: organizations; Type: ROW SECURITY; Schema: corehr; Owner: user
--

ALTER TABLE corehr.organizations ENABLE ROW LEVEL SECURITY;

--
-- Name: employees tenant_isolation_employees; Type: POLICY; Schema: corehr; Owner: user
--

CREATE POLICY tenant_isolation_employees ON corehr.employees USING ((tenant_id = public.get_current_tenant_id()));


--
-- Name: organizations tenant_isolation_organizations; Type: POLICY; Schema: corehr; Owner: user
--

CREATE POLICY tenant_isolation_organizations ON corehr.organizations USING ((tenant_id = public.get_current_tenant_id()));


--
-- Name: position_history; Type: ROW SECURITY; Schema: public; Owner: user
--

ALTER TABLE public.position_history ENABLE ROW LEVEL SECURITY;

--
-- Name: position_history position_history_salary_access; Type: POLICY; Schema: public; Owner: user
--

CREATE POLICY position_history_salary_access ON public.position_history FOR SELECT TO application_role USING (
CASE
    WHEN (current_setting('app.user_permissions'::text, true) ~~ '%hr.compensation.read%'::text) THEN true
    ELSE ((min_salary IS NULL) AND (max_salary IS NULL))
END);


--
-- Name: position_history position_history_tenant_isolation; Type: POLICY; Schema: public; Owner: user
--

CREATE POLICY position_history_tenant_isolation ON public.position_history TO application_role USING ((tenant_id = (current_setting('app.current_tenant_id'::text))::uuid));


--
-- Name: debezium_org_publication; Type: PUBLICATION; Schema: -; Owner: user
--

CREATE PUBLICATION debezium_org_publication FOR ALL TABLES WITH (publish = 'insert, update, delete, truncate');


ALTER PUBLICATION debezium_org_publication OWNER TO "user";

--
-- Name: organization_publication; Type: PUBLICATION; Schema: -; Owner: user
--

CREATE PUBLICATION organization_publication WITH (publish = 'insert, update, delete, truncate');


ALTER PUBLICATION organization_publication OWNER TO "user";

--
-- Name: organization_publication_v2; Type: PUBLICATION; Schema: -; Owner: user
--

CREATE PUBLICATION organization_publication_v2 FOR ALL TABLES WITH (publish = 'insert, update, delete, truncate');


ALTER PUBLICATION organization_publication_v2 OWNER TO "user";

--
-- Name: organization_publication_v3; Type: PUBLICATION; Schema: -; Owner: user
--

CREATE PUBLICATION organization_publication_v3 FOR ALL TABLES WITH (publish = 'insert, update, delete, truncate');


ALTER PUBLICATION organization_publication_v3 OWNER TO "user";

--
-- Name: organization_publication_v4; Type: PUBLICATION; Schema: -; Owner: user
--

CREATE PUBLICATION organization_publication_v4 FOR ALL TABLES WITH (publish = 'insert, update, delete, truncate');


ALTER PUBLICATION organization_publication_v4 OWNER TO "user";

--
-- Name: organization_publication organization_units; Type: PUBLICATION TABLE; Schema: public; Owner: user
--

ALTER PUBLICATION organization_publication ADD TABLE ONLY public.organization_units;


--
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: pg_database_owner
--

GRANT USAGE ON SCHEMA public TO debezium_user;


--
-- Name: TABLE assignment_details; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT ON TABLE public.assignment_details TO debezium_user;


--
-- Name: TABLE assignment_history; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT ON TABLE public.assignment_history TO debezium_user;


--
-- Name: TABLE business_process_events; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT ON TABLE public.business_process_events TO debezium_user;


--
-- Name: TABLE position_history; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT,INSERT,UPDATE ON TABLE public.position_history TO application_role;
GRANT SELECT ON TABLE public.position_history TO debezium_user;


--
-- Name: TABLE employee_department_summary; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT ON TABLE public.employee_department_summary TO debezium_user;


--
-- Name: TABLE employee_positions; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT ON TABLE public.employee_positions TO debezium_user;


--
-- Name: TABLE employee_positions_backup_20250805; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT ON TABLE public.employee_positions_backup_20250805 TO debezium_user;


--
-- Name: TABLE employees; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT ON TABLE public.employees TO debezium_user;


--
-- Name: TABLE employees_backup; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT ON TABLE public.employees_backup TO debezium_user;


--
-- Name: TABLE metacontract_editor_projects; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT ON TABLE public.metacontract_editor_projects TO debezium_user;


--
-- Name: TABLE metacontract_editor_sessions; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT ON TABLE public.metacontract_editor_sessions TO debezium_user;


--
-- Name: TABLE metacontract_editor_settings; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT ON TABLE public.metacontract_editor_settings TO debezium_user;


--
-- Name: TABLE metacontract_editor_templates; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT ON TABLE public.metacontract_editor_templates TO debezium_user;


--
-- Name: TABLE organization_units; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT ON TABLE public.organization_units TO debezium_user;


--
-- Name: TABLE outbox_events; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT ON TABLE public.outbox_events TO debezium_user;


--
-- Name: TABLE person; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT,INSERT,DELETE,UPDATE ON TABLE public.person TO application_role;
GRANT SELECT ON TABLE public.person TO debezium_user;


--
-- Name: TABLE position_assignments; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT ON TABLE public.position_assignments TO debezium_user;


--
-- Name: TABLE position_assignments_backup_20250805; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT ON TABLE public.position_assignments_backup_20250805 TO debezium_user;


--
-- Name: TABLE position_attribute_histories; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT ON TABLE public.position_attribute_histories TO debezium_user;


--
-- Name: TABLE position_code_sequence; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT ON TABLE public.position_code_sequence TO debezium_user;


--
-- Name: TABLE position_histories; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT ON TABLE public.position_histories TO debezium_user;


--
-- Name: TABLE position_occupancy_histories; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT ON TABLE public.position_occupancy_histories TO debezium_user;


--
-- Name: TABLE positions; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT ON TABLE public.positions TO debezium_user;


--
-- Name: TABLE positions_backup; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT ON TABLE public.positions_backup TO debezium_user;


--
-- Name: TABLE positions_backup_20250805; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT ON TABLE public.positions_backup_20250805 TO debezium_user;


--
-- Name: TABLE sync_monitoring; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT ON TABLE public.sync_monitoring TO debezium_user;


--
-- Name: TABLE workflow_instances; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT ON TABLE public.workflow_instances TO debezium_user;


--
-- Name: TABLE workflow_steps; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT ON TABLE public.workflow_steps TO debezium_user;


--
-- PostgreSQL database dump complete
--

