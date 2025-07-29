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
-- Name: corehr; Type: SCHEMA; Schema: -; Owner: user
--

CREATE SCHEMA corehr;


ALTER SCHEMA corehr OWNER TO "user";

--
-- Name: outbox; Type: SCHEMA; Schema: -; Owner: user
--

CREATE SCHEMA outbox;


ALTER SCHEMA outbox OWNER TO "user";

--
-- Name: workflow; Type: SCHEMA; Schema: -; Owner: user
--

CREATE SCHEMA workflow;


ALTER SCHEMA workflow OWNER TO "user";

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
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
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
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE corehr.organizations OWNER TO "user";

--
-- Name: employees; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.employees (
    id character varying NOT NULL,
    name character varying NOT NULL,
    email character varying NOT NULL,
    "position" character varying NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


ALTER TABLE public.employees OWNER TO "user";

--
-- Name: organization_units; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.organization_units (
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    unit_type character varying NOT NULL,
    name character varying NOT NULL,
    description character varying,
    status character varying DEFAULT 'ACTIVE'::character varying NOT NULL,
    profile jsonb,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    parent_unit_id uuid
);


ALTER TABLE public.organization_units OWNER TO "user";

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
    id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    position_type character varying NOT NULL,
    job_profile_id uuid NOT NULL,
    status character varying DEFAULT 'OPEN'::character varying NOT NULL,
    budgeted_fte double precision DEFAULT 1 NOT NULL,
    details jsonb,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    department_id uuid NOT NULL,
    manager_position_id uuid
);


ALTER TABLE public.positions OWNER TO "user";

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
-- Name: employees employees_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.employees
    ADD CONSTRAINT employees_pkey PRIMARY KEY (id);


--
-- Name: organization_units organization_units_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.organization_units
    ADD CONSTRAINT organization_units_pkey PRIMARY KEY (id);


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
-- Name: organizationunit_parent_unit_id; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX organizationunit_parent_unit_id ON public.organization_units USING btree (parent_unit_id);


--
-- Name: organizationunit_tenant_id_name; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX organizationunit_tenant_id_name ON public.organization_units USING btree (tenant_id, name);


--
-- Name: organizationunit_tenant_id_status; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX organizationunit_tenant_id_status ON public.organization_units USING btree (tenant_id, status);


--
-- Name: organizationunit_tenant_id_unit_type; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX organizationunit_tenant_id_unit_type ON public.organization_units USING btree (tenant_id, unit_type);


--
-- Name: organizationunit_tenant_id_unit_type_status; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX organizationunit_tenant_id_unit_type_status ON public.organization_units USING btree (tenant_id, unit_type, status);


--
-- Name: position_department_id; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX position_department_id ON public.positions USING btree (department_id);


--
-- Name: position_job_profile_id; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX position_job_profile_id ON public.positions USING btree (job_profile_id);


--
-- Name: position_manager_position_id; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX position_manager_position_id ON public.positions USING btree (manager_position_id);


--
-- Name: position_tenant_id_budgeted_fte; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX position_tenant_id_budgeted_fte ON public.positions USING btree (tenant_id, budgeted_fte);


--
-- Name: position_tenant_id_department_id_status; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX position_tenant_id_department_id_status ON public.positions USING btree (tenant_id, department_id, status);


--
-- Name: position_tenant_id_position_type; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX position_tenant_id_position_type ON public.positions USING btree (tenant_id, position_type);


--
-- Name: position_tenant_id_status; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX position_tenant_id_status ON public.positions USING btree (tenant_id, status);


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
-- Name: position_history trigger_auto_close_previous_positions; Type: TRIGGER; Schema: public; Owner: user
--

CREATE TRIGGER trigger_auto_close_previous_positions AFTER INSERT ON public.position_history FOR EACH ROW EXECUTE FUNCTION public.auto_close_previous_positions();


--
-- Name: position_history trigger_validate_position_history_temporal_consistency; Type: TRIGGER; Schema: public; Owner: user
--

CREATE TRIGGER trigger_validate_position_history_temporal_consistency BEFORE INSERT OR UPDATE ON public.position_history FOR EACH ROW EXECUTE FUNCTION public.validate_position_history_temporal_consistency();


--
-- Name: organization_units organization_units_organization_units_children; Type: FK CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.organization_units
    ADD CONSTRAINT organization_units_organization_units_children FOREIGN KEY (parent_unit_id) REFERENCES public.organization_units(id) ON DELETE SET NULL;


--
-- Name: position_attribute_histories position_attribute_histories_positions_attribute_history; Type: FK CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.position_attribute_histories
    ADD CONSTRAINT position_attribute_histories_positions_attribute_history FOREIGN KEY (position_id) REFERENCES public.positions(id);


--
-- Name: position_occupancy_histories position_occupancy_histories_positions_occupancy_history; Type: FK CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.position_occupancy_histories
    ADD CONSTRAINT position_occupancy_histories_positions_occupancy_history FOREIGN KEY (position_id) REFERENCES public.positions(id);


--
-- Name: positions positions_organization_units_positions; Type: FK CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.positions
    ADD CONSTRAINT positions_organization_units_positions FOREIGN KEY (department_id) REFERENCES public.organization_units(id);


--
-- Name: positions positions_positions_direct_reports; Type: FK CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.positions
    ADD CONSTRAINT positions_positions_direct_reports FOREIGN KEY (manager_position_id) REFERENCES public.positions(id) ON DELETE SET NULL;


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
-- Name: TABLE person; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT,INSERT,DELETE,UPDATE ON TABLE public.person TO application_role;


--
-- Name: TABLE position_history; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT,INSERT,UPDATE ON TABLE public.position_history TO application_role;


--
-- PostgreSQL database dump complete
--

