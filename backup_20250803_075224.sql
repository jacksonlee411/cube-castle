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
-- Name: update_updated_at_column(); Type: FUNCTION; Schema: public; Owner: user
--

CREATE FUNCTION public.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
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
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    employee_id uuid NOT NULL,
    position_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    start_date date NOT NULL,
    end_date date,
    is_primary boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.employee_positions OWNER TO "user";

--
-- Name: employees; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.employees (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    employee_type character varying(50) NOT NULL,
    first_name character varying(100) NOT NULL,
    last_name character varying(100) NOT NULL,
    email character varying(255) NOT NULL,
    position_id uuid,
    hire_date date NOT NULL,
    termination_date date,
    employment_status character varying(50) DEFAULT 'PENDING_START'::character varying NOT NULL,
    personal_info jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT employees_employee_type_check CHECK (((employee_type)::text = ANY ((ARRAY['FULL_TIME'::character varying, 'PART_TIME'::character varying, 'CONTRACTOR'::character varying, 'INTERN'::character varying])::text[]))),
    CONSTRAINT employees_employment_status_check CHECK (((employment_status)::text = ANY ((ARRAY['PENDING_START'::character varying, 'ACTIVE'::character varying, 'TERMINATED'::character varying, 'ON_LEAVE'::character varying])::text[])))
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
-- Name: organization_units; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.organization_units (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    unit_type character varying(50) NOT NULL,
    name character varying(100) NOT NULL,
    description text,
    parent_unit_id uuid,
    profile jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    status character varying(20) DEFAULT 'ACTIVE'::character varying NOT NULL,
    level integer DEFAULT 0 NOT NULL,
    employee_count integer DEFAULT 0 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    CONSTRAINT organization_units_status_check CHECK (((status)::text = ANY ((ARRAY['ACTIVE'::character varying, 'INACTIVE'::character varying, 'PLANNED'::character varying])::text[]))),
    CONSTRAINT organization_units_unit_type_check CHECK (((unit_type)::text = ANY ((ARRAY['DEPARTMENT'::character varying, 'COST_CENTER'::character varying, 'COMPANY'::character varying, 'PROJECT_TEAM'::character varying])::text[])))
);


ALTER TABLE public.organization_units OWNER TO "user";

--
-- Name: organization_units_backup; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.organization_units_backup (
    id uuid,
    tenant_id uuid,
    unit_type character varying,
    name character varying,
    description character varying,
    status character varying,
    profile jsonb,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    parent_unit_id uuid
);


ALTER TABLE public.organization_units_backup OWNER TO "user";

--
-- Name: outbox_events; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.outbox_events (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    event_type character varying NOT NULL,
    payload bytea NOT NULL,
    destination character varying NOT NULL,
    retry_count bigint DEFAULT 0 NOT NULL,
    next_retry_at timestamp with time zone,
    processed_at timestamp with time zone,
    error_message character varying,
    created_at timestamp with time zone NOT NULL
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
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    title character varying(100) NOT NULL,
    department character varying(100) NOT NULL,
    level character varying(50) NOT NULL,
    description text,
    requirements text,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.positions OWNER TO "user";

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
-- Data for Name: employees; Type: TABLE DATA; Schema: corehr; Owner: user
--

COPY corehr.employees (id, tenant_id, employee_number, first_name, last_name, email, status, created_at, phone_number, "position", department, hire_date, manager_id, updated_at) FROM stdin;
6bc3fa3a-a761-4df3-957c-11bccfd47fdc	62c5f693-95b0-4d0b-bf1f-5f3d86e296fb	FINAL-TEST-1753938539	最终	测试	final-test-1753938539@example.com	active	2025-07-31 05:08:59.627038+00	13800138000	\N	\N	2025-07-31	\N	2025-07-31 05:08:59.627038+00
6e5009c2-d8c2-4ad4-8f9b-8909c6462418	00000000-0000-0000-0000-000000000000	EMP001	张	伟强	zhang.weiqiang@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	CTO	技术部	2020-01-01	\N	2025-08-01 00:58:54.771142+00
eee08f7b-833c-437e-84f4-4e8ee0a25223	00000000-0000-0000-0000-000000000000	EMP002	李	芳芳	li.fangfang@techcorp.com	active	2025-08-01 00:58:54.771142+00	\N	CPO	产品部	2020-01-01	\N	2025-08-01 00:58:54.771142+00
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
-- Data for Name: business_process_events; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.business_process_events (id, tenant_id, event_type, entity_type, entity_id, effective_date, event_data, initiated_by, correlation_id, status, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: employee_positions; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.employee_positions (id, employee_id, position_id, tenant_id, start_date, end_date, is_primary, created_at) FROM stdin;
\.


--
-- Data for Name: employees; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.employees (id, tenant_id, employee_type, first_name, last_name, email, position_id, hire_date, termination_date, employment_status, personal_info, created_at, updated_at) FROM stdin;
ad09e010-75ef-437a-9160-c07062041009	9e5116f7-b3ce-40ea-a2ff-157d555771ba	FULL_TIME	Phoenix	TestEmployee	phoenix.test@cubecastle.com	\N	2025-08-02	\N	ACTIVE	\N	2025-08-02 03:52:43.956326+00	2025-08-02 03:52:43.956326+00
71150bab-801c-4fb3-b7bd-35989845b324	d23d2d4f-8dae-4c2f-8ba7-cba81dee7364	FULL_TIME	CQRS	TestEmployee	cqrs.test@cubecastle.com	\N	2025-08-02	\N	ACTIVE	\N	2025-08-02 03:53:11.161095+00	2025-08-02 03:53:11.161095+00
a60394bd-3020-4b82-8b52-4af598560968	1b3d9937-8d6a-418d-a07a-93897e01cf2a	FULL_TIME	Phoenix	TestUser	phoenix.test@cubecastle.com	\N	2025-08-02	\N	ACTIVE	\N	2025-08-02 06:01:00.767856+00	2025-08-02 06:01:00.767856+00
88172843-6487-4bd2-99e2-28048b898b72	fd99b8c3-59bc-4a47-848f-adb82d93638d	FULL_TIME	Phoenix	CDCTest	phoenix.cdc@cubecastle.com	\N	2025-08-02	\N	ACTIVE	\N	2025-08-02 06:05:54.4592+00	2025-08-02 06:05:54.4592+00
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

COPY public.organization_units (id, tenant_id, unit_type, name, description, parent_unit_id, profile, created_at, updated_at, status, level, employee_count, is_active) FROM stdin;
5cfdb01d-9dcc-49f4-b9bd-4f43453520c5	550e8400-e29b-41d4-a716-446655440000	COMPANY	高谷集团	CQRS Phase 3 Real Database Test	\N	\N	2025-08-02 14:24:56.454698+00	2025-08-02 23:15:06.989562+00	ACTIVE	0	0	t
2f86d7e2-742f-4a84-9ab0-5eb0f9d79ae6	550e8400-e29b-41d4-a716-446655440000	DEPARTMENT	技术研发部	负责公司核心技术研发、产品架构设计、技术创新和技术团队管理	5cfdb01d-9dcc-49f4-b9bd-4f43453520c5	{}	2025-08-02 22:59:14.951084+00	2025-08-02 23:15:06.989562+00	ACTIVE	1	0	t
5629c5e0-db37-4e0e-84bd-bf87e8523b38	550e8400-e29b-41d4-a716-446655440000	DEPARTMENT	人力资源部	负责人才招聘、员工培训、绩效管理、薪酬福利和企业文化建设	5cfdb01d-9dcc-49f4-b9bd-4f43453520c5	{}	2025-08-02 23:26:31.492055+00	2025-08-02 23:26:31.492055+00	ACTIVE	1	0	t
b1f8ae08-b1d4-4e15-9e07-dc235ae27e15	550e8400-e29b-41d4-a716-446655440000	DEPARTMENT	财务管理部	负责财务规划、成本控制、资金管理、财务分析和合规审计	5cfdb01d-9dcc-49f4-b9bd-4f43453520c5	{}	2025-08-02 23:26:31.492055+00	2025-08-02 23:26:31.492055+00	ACTIVE	1	0	t
e7e38f64-8da3-42a9-b478-4fe6043c35ce	550e8400-e29b-41d4-a716-446655440000	DEPARTMENT	产品管理部	负责公司产品规划、产品设计、用户体验、项目管理和产品运营	5cfdb01d-9dcc-49f4-b9bd-4f43453520c5	{}	2025-08-02 23:24:25.257656+00	2025-08-02 23:27:03.790748+00	ACTIVE	1	0	t
75db946b-3138-4dd7-9145-33025409c185	550e8400-e29b-41d4-a716-446655440000	DEPARTMENT	市场营销部	负责市场推广、品牌建设、客户关系管理、销售支持和市场分析	5cfdb01d-9dcc-49f4-b9bd-4f43453520c5	{}	2025-08-02 23:25:52.073502+00	2025-08-02 23:27:03.790748+00	ACTIVE	1	0	t
cc7cdb48-9e04-4c58-811a-1185daa43127	550e8400-e29b-41d4-a716-446655440000	PROJECT_TEAM	前端开发组	负责Web前端、移动端UI开发和用户体验优化	2f86d7e2-742f-4a84-9ab0-5eb0f9d79ae6	{}	2025-08-02 23:28:05.256212+00	2025-08-02 23:28:05.256212+00	ACTIVE	2	0	t
3903647b-0a2b-4ec1-b31b-a5d9f3600ae8	550e8400-e29b-41d4-a716-446655440000	PROJECT_TEAM	后端开发组	负责服务端开发、API设计和数据库架构	2f86d7e2-742f-4a84-9ab0-5eb0f9d79ae6	{}	2025-08-02 23:28:05.256212+00	2025-08-02 23:28:05.256212+00	ACTIVE	2	0	t
26f504e7-f7a2-41e9-9469-1e7487504800	550e8400-e29b-41d4-a716-446655440000	PROJECT_TEAM	测试质量组	负责软件测试、质量保证和自动化测试	2f86d7e2-742f-4a84-9ab0-5eb0f9d79ae6	{}	2025-08-02 23:28:05.256212+00	2025-08-02 23:28:05.256212+00	ACTIVE	2	0	t
2d1abfad-501d-4e69-b8e5-2a393264eae7	550e8400-e29b-41d4-a716-446655440000	PROJECT_TEAM	运维架构组	负责系统运维、CI/CD和基础设施架构	2f86d7e2-742f-4a84-9ab0-5eb0f9d79ae6	{}	2025-08-02 23:28:05.256212+00	2025-08-02 23:28:05.256212+00	ACTIVE	2	0	t
e370c5e9-e193-4cff-8944-5f76235e5d82	550e8400-e29b-41d4-a716-446655440000	PROJECT_TEAM	产品规划组	负责产品策略制定、需求分析和产品路线图	e7e38f64-8da3-42a9-b478-4fe6043c35ce	{}	2025-08-02 23:28:05.256212+00	2025-08-02 23:28:05.256212+00	ACTIVE	2	0	t
6b1a55cd-4b90-4d9e-94a4-bee7d2c77bc4	550e8400-e29b-41d4-a716-446655440000	PROJECT_TEAM	用户体验组	负责用户研究、交互设计和界面设计	e7e38f64-8da3-42a9-b478-4fe6043c35ce	{}	2025-08-02 23:28:05.256212+00	2025-08-02 23:28:05.256212+00	ACTIVE	2	0	t
6064c299-e1fb-4478-bd5d-037f9511d3be	550e8400-e29b-41d4-a716-446655440000	PROJECT_TEAM	项目管理组	负责项目协调、进度管理和资源调配	e7e38f64-8da3-42a9-b478-4fe6043c35ce	{}	2025-08-02 23:28:05.256212+00	2025-08-02 23:28:05.256212+00	ACTIVE	2	0	t
a394c695-13ee-40ce-b052-a7569b6165c1	550e8400-e29b-41d4-a716-446655440000	PROJECT_TEAM	数据分析组	负责用户行为分析、产品数据分析和业务洞察	e7e38f64-8da3-42a9-b478-4fe6043c35ce	{}	2025-08-02 23:28:05.256212+00	2025-08-02 23:28:05.256212+00	ACTIVE	2	0	t
ab6dbacd-ac5c-4eef-90d2-590523d089f5	550e8400-e29b-41d4-a716-446655440000	PROJECT_TEAM	品牌推广组	负责品牌建设、公关活动和媒体合作	75db946b-3138-4dd7-9145-33025409c185	{}	2025-08-02 23:28:05.256212+00	2025-08-02 23:28:05.256212+00	ACTIVE	2	0	t
56ad0e7a-75c8-435f-8c0e-1b5c54ce882f	550e8400-e29b-41d4-a716-446655440000	PROJECT_TEAM	数字营销组	负责线上推广、SEM/SEO和社交媒体营销	75db946b-3138-4dd7-9145-33025409c185	{}	2025-08-02 23:28:05.256212+00	2025-08-02 23:28:05.256212+00	ACTIVE	2	0	t
69c93c6a-915e-4789-bbe3-78e19f06adca	550e8400-e29b-41d4-a716-446655440000	PROJECT_TEAM	客户关系组	负责客户维护、客户服务和客户满意度	75db946b-3138-4dd7-9145-33025409c185	{}	2025-08-02 23:28:05.256212+00	2025-08-02 23:28:05.256212+00	ACTIVE	2	0	t
a5a90dee-fad4-45a5-a1d7-fbfb4939071a	550e8400-e29b-41d4-a716-446655440000	PROJECT_TEAM	销售支持组	负责销售工具、销售培训和销售数据分析	75db946b-3138-4dd7-9145-33025409c185	{}	2025-08-02 23:28:05.256212+00	2025-08-02 23:28:05.256212+00	ACTIVE	2	0	t
c330c171-96b4-4a3b-aa52-af70fff5d906	550e8400-e29b-41d4-a716-446655440000	PROJECT_TEAM	招聘培训组	负责人才招聘、入职培训和员工发展	5629c5e0-db37-4e0e-84bd-bf87e8523b38	{}	2025-08-02 23:28:05.256212+00	2025-08-02 23:28:05.256212+00	ACTIVE	2	0	t
e8dc712a-5dde-4d46-8b7d-e402d2f916d1	550e8400-e29b-41d4-a716-446655440000	PROJECT_TEAM	绩效薪酬组	负责绩效考核、薪酬管理和激励机制	5629c5e0-db37-4e0e-84bd-bf87e8523b38	{}	2025-08-02 23:28:05.256212+00	2025-08-02 23:28:05.256212+00	ACTIVE	2	0	t
0495eeb4-d559-4eec-9a96-c495e3c5e4dd	550e8400-e29b-41d4-a716-446655440000	PROJECT_TEAM	员工关系组	负责员工关怀、劳动关系和企业文化	5629c5e0-db37-4e0e-84bd-bf87e8523b38	{}	2025-08-02 23:28:05.256212+00	2025-08-02 23:28:05.256212+00	ACTIVE	2	0	t
c359b9a0-4428-43c2-ac37-2e8810e6cad9	550e8400-e29b-41d4-a716-446655440000	PROJECT_TEAM	人事行政组	负责人事档案、考勤管理和行政支持	5629c5e0-db37-4e0e-84bd-bf87e8523b38	{}	2025-08-02 23:28:05.256212+00	2025-08-02 23:28:05.256212+00	ACTIVE	2	0	t
581eda0b-f212-4ea0-b5f4-a1597fc55cb9	550e8400-e29b-41d4-a716-446655440000	PROJECT_TEAM	财务核算组	负责日常记账、财务报表和税务申报	b1f8ae08-b1d4-4e15-9e07-dc235ae27e15	{}	2025-08-02 23:28:05.256212+00	2025-08-02 23:28:05.256212+00	ACTIVE	2	0	t
a95762ad-394a-4216-857f-e572ab73a7f9	550e8400-e29b-41d4-a716-446655440000	PROJECT_TEAM	成本控制组	负责成本分析、预算管理和费用控制	b1f8ae08-b1d4-4e15-9e07-dc235ae27e15	{}	2025-08-02 23:28:05.256212+00	2025-08-02 23:28:05.256212+00	ACTIVE	2	0	t
83a619fc-0a00-44b3-992f-4355972ef2df	550e8400-e29b-41d4-a716-446655440000	PROJECT_TEAM	资金管理组	负责现金流管理、投资决策和资金调配	b1f8ae08-b1d4-4e15-9e07-dc235ae27e15	{}	2025-08-02 23:28:05.256212+00	2025-08-02 23:28:05.256212+00	ACTIVE	2	0	t
618bed72-dd46-4bd0-9a43-14e7c625b59c	550e8400-e29b-41d4-a716-446655440000	PROJECT_TEAM	审计风控组	负责内部审计、风险控制和合规监督	b1f8ae08-b1d4-4e15-9e07-dc235ae27e15	{}	2025-08-02 23:28:05.256212+00	2025-08-02 23:28:05.256212+00	ACTIVE	2	0	t
\.


--
-- Data for Name: organization_units_backup; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.organization_units_backup (id, tenant_id, unit_type, name, description, status, profile, created_at, updated_at, parent_unit_id) FROM stdin;
ec3afce7-4466-420d-bfa8-b569880b984a	550e8400-e29b-41d4-a716-446655440000	DEPARTMENT	工程技术部	负责产品技术开发	ACTIVE	{"cost_center": "CC-ENG-001", "capabilities": ["software_development", "system_architecture", "devops"], "functional_area": "engineering", "has_budget_responsibility": true}	2025-07-29 12:53:50.550483+00	2025-07-29 12:53:50.550484+00	\N
5fa32f0c-0a6f-4242-8372-632565d26731	550e8400-e29b-41d4-a716-446655440000	DEPARTMENT	UAT测试部门	UAT测试专用部门	ACTIVE	\N	2025-07-30 02:17:29.072512+00	2025-07-30 02:17:29.072515+00	\N
3659e057-e60f-4559-9517-e782c61ada3a	550e8400-e29b-41d4-a716-446655440000	DEPARTMENT	UAT测试部门	UAT测试专用部门	ACTIVE	\N	2025-07-30 03:07:12.705018+00	2025-07-30 03:07:12.705018+00	\N
ac2b16df-7713-4ec1-b21b-ff501f6b7264	550e8400-e29b-41d4-a716-446655440000	DEPARTMENT	UAT测试部门	UAT测试专用部门	ACTIVE	\N	2025-07-30 03:36:56.219389+00	2025-07-30 03:36:56.21939+00	\N
07b0b20f-629c-4a4a-9fd6-7fdd8dfdc87c	550e8400-e29b-41d4-a716-446655440000	DEPARTMENT	UAT测试部门	UAT测试专用部门	ACTIVE	\N	2025-07-30 03:37:29.279577+00	2025-07-30 03:37:29.279578+00	\N
fdfd725f-8641-409f-86a0-6353a5013b67	550e8400-e29b-41d4-a716-446655440000	DEPARTMENT	AI治理办公室	负责AI技术治理和规范制定	ACTIVE	\N	2025-07-30 10:59:47.929929+00	2025-07-30 10:59:47.92993+00	\N
e0af7ce9-fb79-4f14-850d-60a1886daf85	123e4567-e89b-12d3-a456-426614174000	COMPANY	总公司	集团总公司	ACTIVE	\N	2025-07-30 23:54:28.441275+00	2025-07-30 23:54:28.441276+00	\N
88043b12-9253-41eb-b8aa-77601685d473	123e4567-e89b-12d3-a456-426614174000	DEPARTMENT	人力资源部	人力资源管理部门	ACTIVE	\N	2025-07-30 23:54:45.550237+00	2025-07-30 23:54:45.550237+00	e0af7ce9-fb79-4f14-850d-60a1886daf85
e7b68e6a-268e-4fc6-999f-9e7bee5e0b3c	123e4567-e89b-12d3-a456-426614174000	PROJECT_TEAM	后端开发组	后端系统开发小组	ACTIVE	\N	2025-07-30 23:55:37.430467+00	2025-07-30 23:55:37.430468+00	806e9d1f-bd95-436a-8fdc-b8e603bff8db
806e9d1f-bd95-436a-8fdc-b8e603bff8db	123e4567-e89b-12d3-a456-426614174000	DEPARTMENT	技术研发部	软件技术研发与创新部门	ACTIVE	\N	2025-07-30 23:54:36.366686+00	2025-07-30 23:56:06.34444+00	e0af7ce9-fb79-4f14-850d-60a1886daf85
76df127f-6ad6-41ad-8a4b-73950d80df67	00000000-0000-0000-0000-000000000000	COMPANY	CubeCastle Technology	领先的企业级软件解决方案提供商	ACTIVE	{"size": "medium", "founded": "2018", "industry": "software"}	2025-08-01 00:02:41.551049+00	2025-08-01 00:02:41.551049+00	\N
0a09b4ee-3c96-4601-b8fe-2ac0140e7be5	00000000-0000-0000-0000-000000000000	DEPARTMENT	前端开发部	负责用户界面和用户体验开发	ACTIVE	{"headcount": 8, "tech_stack": ["React", "Vue", "TypeScript"]}	2025-08-01 00:02:41.559311+00	2025-08-01 00:02:41.559311+00	76df127f-6ad6-41ad-8a4b-73950d80df67
7c222568-eb3e-4032-86d1-91f5fd9f0f69	00000000-0000-0000-0000-000000000000	DEPARTMENT	后端开发部	负责服务端开发和API设计	ACTIVE	{"headcount": 10, "tech_stack": ["Go", "Python", "Node.js"]}	2025-08-01 00:02:41.559311+00	2025-08-01 00:02:41.559311+00	76df127f-6ad6-41ad-8a4b-73950d80df67
f715fcd7-9c73-489c-a2af-5406ceaaa38a	00000000-0000-0000-0000-000000000000	DEPARTMENT	移动开发部	负责移动应用开发	ACTIVE	{"headcount": 6, "tech_stack": ["React Native", "Flutter", "iOS", "Android"]}	2025-08-01 00:02:41.559311+00	2025-08-01 00:02:41.559311+00	76df127f-6ad6-41ad-8a4b-73950d80df67
02fabf96-183a-46b5-9bfa-6974312b603d	00000000-0000-0000-0000-000000000000	DEPARTMENT	数据工程部	负责大数据处理和数据平台建设	ACTIVE	{"headcount": 5, "tech_stack": ["Kafka", "Spark", "Elasticsearch"]}	2025-08-01 00:02:41.559311+00	2025-08-01 00:02:41.559311+00	76df127f-6ad6-41ad-8a4b-73950d80df67
09bebe19-2981-45ef-9226-9f0d6ae88549	00000000-0000-0000-0000-000000000000	DEPARTMENT	DevOps部	负责基础设施和CI/CD	ACTIVE	{"headcount": 4, "tech_stack": ["Docker", "Kubernetes", "AWS", "Jenkins"]}	2025-08-01 00:02:41.559311+00	2025-08-01 00:02:41.559311+00	76df127f-6ad6-41ad-8a4b-73950d80df67
6db67d50-9019-4d77-8584-c6931d623f39	00000000-0000-0000-0000-000000000000	DEPARTMENT	测试部	负责软件质量保证和自动化测试	ACTIVE	{"headcount": 6, "tech_stack": ["Selenium", "Jest", "Cypress"]}	2025-08-01 00:02:41.559311+00	2025-08-01 00:02:41.559311+00	76df127f-6ad6-41ad-8a4b-73950d80df67
3edda3f0-7c54-47a3-97e9-0a22ed6c250e	00000000-0000-0000-0000-000000000000	DEPARTMENT	架构部	负责技术架构设计和技术选型	ACTIVE	{"focus": ["system_design", "performance", "scalability"], "headcount": 3}	2025-08-01 00:02:41.559311+00	2025-08-01 00:02:41.559311+00	76df127f-6ad6-41ad-8a4b-73950d80df67
251375cb-f234-4158-b15a-140854dc9b12	00000000-0000-0000-0000-000000000000	DEPARTMENT	产品管理部	负责产品规划和需求分析	ACTIVE	{"focus": ["product_strategy", "user_research"], "headcount": 4}	2025-08-01 00:02:41.559311+00	2025-08-01 00:02:41.559311+00	76df127f-6ad6-41ad-8a4b-73950d80df67
a5f43444-35b2-4b30-b649-f4f2a809b0c7	00000000-0000-0000-0000-000000000000	DEPARTMENT	UX设计部	负责用户体验设计	ACTIVE	{"tools": ["Figma", "Sketch", "Adobe XD"], "headcount": 3}	2025-08-01 00:02:41.559311+00	2025-08-01 00:02:41.559311+00	76df127f-6ad6-41ad-8a4b-73950d80df67
dc83c05b-e5ff-40c7-a651-35a1efba93c5	00000000-0000-0000-0000-000000000000	DEPARTMENT	销售部	负责客户开发和销售	ACTIVE	{"focus": ["enterprise_sales", "channel_partnership"], "headcount": 5}	2025-08-01 00:02:41.559311+00	2025-08-01 00:02:41.559311+00	76df127f-6ad6-41ad-8a4b-73950d80df67
3c46526a-8e6f-4a99-b057-387f80bda484	00000000-0000-0000-0000-000000000000	DEPARTMENT	市场部	负责品牌推广和市场营销	ACTIVE	{"focus": ["digital_marketing", "content_marketing"], "headcount": 3}	2025-08-01 00:02:41.559311+00	2025-08-01 00:02:41.559311+00	76df127f-6ad6-41ad-8a4b-73950d80df67
3bcf62e2-f6c7-4240-b848-153da43fcd79	00000000-0000-0000-0000-000000000000	DEPARTMENT	客户成功部	负责客户支持和成功	ACTIVE	{"focus": ["customer_support", "success_management"], "headcount": 4}	2025-08-01 00:02:41.559311+00	2025-08-01 00:02:41.559311+00	76df127f-6ad6-41ad-8a4b-73950d80df67
c09da93a-6ee8-4deb-8fa2-abf4e66012e6	00000000-0000-0000-0000-000000000000	DEPARTMENT	人力资源部	负责人才招聘和员工发展	ACTIVE	{"focus": ["recruitment", "training", "performance"], "headcount": 2}	2025-08-01 00:02:41.559311+00	2025-08-01 00:02:41.559311+00	76df127f-6ad6-41ad-8a4b-73950d80df67
6e80a9e9-beac-42a0-bcdd-45371610721b	00000000-0000-0000-0000-000000000000	DEPARTMENT	财务部	负责财务管理和成本控制	ACTIVE	{"focus": ["financial_planning", "budgeting"], "headcount": 2}	2025-08-01 00:02:41.559311+00	2025-08-01 00:02:41.559311+00	76df127f-6ad6-41ad-8a4b-73950d80df67
0540cd68-01e9-41e9-9675-a7ea62736d05	00000000-0000-0000-0000-000000000000	DEPARTMENT	法务部	负责法律事务和合规	ACTIVE	{"focus": ["contract_management", "compliance"], "headcount": 1}	2025-08-01 00:02:41.559311+00	2025-08-01 00:02:41.559311+00	76df127f-6ad6-41ad-8a4b-73950d80df67
9aeccaea-a3b4-471c-b4ac-bd399bbb02c2	00000000-0000-0000-0000-000000000000	DEPARTMENT	行政部	负责日常行政事务	ACTIVE	{"focus": ["office_management", "procurement"], "headcount": 2}	2025-08-01 00:02:41.559311+00	2025-08-01 00:02:41.559311+00	76df127f-6ad6-41ad-8a4b-73950d80df67
980191ba-0af2-487d-894d-b98be815b80c	00000000-0000-0000-0000-000000000000	DEPARTMENT	创新实验室	负责新技术研究和原型开发	ACTIVE	{"focus": ["AI", "blockchain", "IoT"], "headcount": 3}	2025-08-01 00:02:41.559311+00	2025-08-01 00:02:41.559311+00	76df127f-6ad6-41ad-8a4b-73950d80df67
698fddc0-6eb5-4726-a087-01af65c9e8c8	00000000-0000-0000-0000-000000000000	DEPARTMENT	安全部	负责信息安全和数据保护	ACTIVE	{"focus": ["cybersecurity", "data_protection"], "headcount": 2}	2025-08-01 00:02:41.559311+00	2025-08-01 00:02:41.559311+00	76df127f-6ad6-41ad-8a4b-73950d80df67
75c434d5-2875-4b19-968f-db1f2d477d00	00000000-0000-0000-0000-000000000000	DEPARTMENT	技术写作部	负责技术文档和API文档	ACTIVE	{"focus": ["technical_writing", "documentation"], "headcount": 2}	2025-08-01 00:02:41.559311+00	2025-08-01 00:02:41.559311+00	76df127f-6ad6-41ad-8a4b-73950d80df67
479fa572-b0ff-4a12-b62e-80b32548e0f8	00000000-0000-0000-0000-000000000000	DEPARTMENT	业务分析部	负责业务需求分析和流程优化	ACTIVE	{"focus": ["business_analysis", "process_optimization"], "headcount": 2}	2025-08-01 00:02:41.559311+00	2025-08-01 00:02:41.559311+00	76df127f-6ad6-41ad-8a4b-73950d80df67
a19274b4-bc08-4d4d-884b-3907c608f41c	00000000-0000-0000-0000-000000000000	COMPANY	CubeCastle Technology	领先的企业级软件解决方案提供商	ACTIVE	{"size": "medium", "founded": "2018", "industry": "software"}	2025-08-01 00:08:13.281262+00	2025-08-01 00:08:13.281262+00	\N
ce432ed1-5bc4-43dc-ac55-6410b1c78a6c	550e8400-e29b-41d4-a716-446655440000	COMPANY	高谷集团	集团总部	ACTIVE	\N	2025-08-01 13:12:11.947147+00	2025-08-01 13:12:11.947147+00	\N
07bfe14a-35f6-4b9a-b851-e27f2baae7bc	550e8400-e29b-41d4-a716-446655440000	COMPANY	高谷集团	高山出平湖，幽兰远空谷	ACTIVE	\N	2025-08-01 13:20:03.18548+00	2025-08-01 13:20:03.185481+00	\N
4f2fe27d-f40d-4935-a510-916e06893983	550e8400-e29b-41d4-a716-446655440000	DEPARTMENT	测试部门	测试创建组织	ACTIVE	\N	2025-08-01 16:01:00.696581+00	2025-08-01 16:01:00.696582+00	\N
64765477-9075-4771-9d22-ea9c813de18a	550e8400-e29b-41d4-a716-446655440000	DEPARTMENT	数据科学部	负责数据分析、机器学习和人工智能相关的技术研发工作	ACTIVE	{"managerName": "张伟", "maxCapacity": 20}	2025-08-01 22:17:43.045942+00	2025-08-01 22:17:43.045944+00	\N
\.


--
-- Data for Name: outbox_events; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.outbox_events (id, tenant_id, event_type, payload, destination, retry_count, next_retry_at, processed_at, error_message, created_at) FROM stdin;
\.


--
-- Data for Name: person; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.person (id, tenant_id, name, email, employee_id, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: position_attribute_histories; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.position_attribute_histories (id, tenant_id, position_type, job_profile_id, department_id, manager_position_id, status, budgeted_fte, details, effective_date, end_date, change_reason, changed_by, change_type, source_event_id, created_at, position_id) FROM stdin;
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

COPY public.positions (id, tenant_id, title, department, level, description, requirements, is_active, created_at, updated_at) FROM stdin;
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
-- Name: business_process_events business_process_events_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.business_process_events
    ADD CONSTRAINT business_process_events_pkey PRIMARY KEY (id);


--
-- Name: employee_positions employee_positions_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.employee_positions
    ADD CONSTRAINT employee_positions_pkey PRIMARY KEY (id);


--
-- Name: employee_positions employee_positions_unique; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.employee_positions
    ADD CONSTRAINT employee_positions_unique UNIQUE (employee_id, position_id, start_date);


--
-- Name: employees employees_email_tenant_unique; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.employees
    ADD CONSTRAINT employees_email_tenant_unique UNIQUE (email, tenant_id);


--
-- Name: employees employees_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.employees
    ADD CONSTRAINT employees_pkey PRIMARY KEY (id);


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
-- Name: organization_units organization_units_name_tenant_unique; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.organization_units
    ADD CONSTRAINT organization_units_name_tenant_unique UNIQUE (name, tenant_id);


--
-- Name: organization_units organization_units_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.organization_units
    ADD CONSTRAINT organization_units_pkey PRIMARY KEY (id);


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
-- Name: position_attribute_histories position_attribute_histories_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.position_attribute_histories
    ADD CONSTRAINT position_attribute_histories_pkey PRIMARY KEY (id);


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
    ADD CONSTRAINT positions_pkey PRIMARY KEY (id);


--
-- Name: positions positions_title_dept_tenant_unique; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.positions
    ADD CONSTRAINT positions_title_dept_tenant_unique UNIQUE (title, department, tenant_id);


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
-- Name: idx_emp_positions_employee; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_emp_positions_employee ON public.employee_positions USING btree (employee_id);


--
-- Name: idx_emp_positions_position; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_emp_positions_position ON public.employee_positions USING btree (position_id);


--
-- Name: idx_emp_positions_primary; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_emp_positions_primary ON public.employee_positions USING btree (is_primary) WHERE (is_primary = true);


--
-- Name: idx_emp_positions_tenant; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_emp_positions_tenant ON public.employee_positions USING btree (tenant_id);


--
-- Name: idx_employees_email; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_employees_email ON public.employees USING btree (email);


--
-- Name: idx_employees_hire_date; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_employees_hire_date ON public.employees USING btree (hire_date);


--
-- Name: idx_employees_position; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_employees_position ON public.employees USING btree (position_id);


--
-- Name: idx_employees_status; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_employees_status ON public.employees USING btree (employment_status);


--
-- Name: idx_employees_tenant_id; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_employees_tenant_id ON public.employees USING btree (tenant_id);


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
-- Name: idx_org_units_parent; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_org_units_parent ON public.organization_units USING btree (parent_unit_id);


--
-- Name: idx_org_units_tenant_id; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_org_units_tenant_id ON public.organization_units USING btree (tenant_id);


--
-- Name: idx_org_units_type; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_org_units_type ON public.organization_units USING btree (unit_type);


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
-- Name: idx_positions_active; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_positions_active ON public.positions USING btree (is_active);


--
-- Name: idx_positions_dept; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_positions_dept ON public.positions USING btree (department);


--
-- Name: idx_positions_tenant_id; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_positions_tenant_id ON public.positions USING btree (tenant_id);


--
-- Name: outboxevent_created_at; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX outboxevent_created_at ON public.outbox_events USING btree (created_at);


--
-- Name: outboxevent_destination; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX outboxevent_destination ON public.outbox_events USING btree (destination);


--
-- Name: outboxevent_next_retry_at; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX outboxevent_next_retry_at ON public.outbox_events USING btree (next_retry_at);


--
-- Name: outboxevent_processed_at; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX outboxevent_processed_at ON public.outbox_events USING btree (processed_at);


--
-- Name: outboxevent_tenant_id_event_type; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX outboxevent_tenant_id_event_type ON public.outbox_events USING btree (tenant_id, event_type);


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
-- Name: position_history trigger_auto_close_previous_positions; Type: TRIGGER; Schema: public; Owner: user
--

CREATE TRIGGER trigger_auto_close_previous_positions AFTER INSERT ON public.position_history FOR EACH ROW EXECUTE FUNCTION public.auto_close_previous_positions();


--
-- Name: position_history trigger_validate_position_history_temporal_consistency; Type: TRIGGER; Schema: public; Owner: user
--

CREATE TRIGGER trigger_validate_position_history_temporal_consistency BEFORE INSERT OR UPDATE ON public.position_history FOR EACH ROW EXECUTE FUNCTION public.validate_position_history_temporal_consistency();


--
-- Name: employees update_employees_updated_at; Type: TRIGGER; Schema: public; Owner: user
--

CREATE TRIGGER update_employees_updated_at BEFORE UPDATE ON public.employees FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: organization_units update_organization_units_updated_at; Type: TRIGGER; Schema: public; Owner: user
--

CREATE TRIGGER update_organization_units_updated_at BEFORE UPDATE ON public.organization_units FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: positions update_positions_updated_at; Type: TRIGGER; Schema: public; Owner: user
--

CREATE TRIGGER update_positions_updated_at BEFORE UPDATE ON public.positions FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


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
-- Name: employee_positions employee_positions_employee_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.employee_positions
    ADD CONSTRAINT employee_positions_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES public.employees(id) ON DELETE CASCADE;


--
-- Name: employee_positions employee_positions_position_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.employee_positions
    ADD CONSTRAINT employee_positions_position_id_fkey FOREIGN KEY (position_id) REFERENCES public.positions(id) ON DELETE CASCADE;


--
-- Name: metacontract_editor_sessions metacontract_editor_sessions_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.metacontract_editor_sessions
    ADD CONSTRAINT metacontract_editor_sessions_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.metacontract_editor_projects(id) ON DELETE CASCADE;


--
-- Name: organization_units organization_units_parent_unit_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.organization_units
    ADD CONSTRAINT organization_units_parent_unit_id_fkey FOREIGN KEY (parent_unit_id) REFERENCES public.organization_units(id);


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
-- Name: organization_publication; Type: PUBLICATION; Schema: -; Owner: user
--

CREATE PUBLICATION organization_publication WITH (publish = 'insert, update, delete, truncate');


ALTER PUBLICATION organization_publication OWNER TO "user";

--
-- Name: organization_publication employee_positions; Type: PUBLICATION TABLE; Schema: public; Owner: user
--

ALTER PUBLICATION organization_publication ADD TABLE ONLY public.employee_positions;


--
-- Name: organization_publication employees; Type: PUBLICATION TABLE; Schema: public; Owner: user
--

ALTER PUBLICATION organization_publication ADD TABLE ONLY public.employees;


--
-- Name: organization_publication organization_units; Type: PUBLICATION TABLE; Schema: public; Owner: user
--

ALTER PUBLICATION organization_publication ADD TABLE ONLY public.organization_units;


--
-- Name: organization_publication positions; Type: PUBLICATION TABLE; Schema: public; Owner: user
--

ALTER PUBLICATION organization_publication ADD TABLE ONLY public.positions;


--
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: pg_database_owner
--

GRANT USAGE ON SCHEMA public TO debezium_user;


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
-- Name: TABLE organization_units_backup; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT ON TABLE public.organization_units_backup TO debezium_user;


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
-- Name: TABLE position_attribute_histories; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT ON TABLE public.position_attribute_histories TO debezium_user;


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

