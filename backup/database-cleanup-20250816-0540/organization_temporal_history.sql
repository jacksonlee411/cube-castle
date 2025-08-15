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

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: organization_temporal_history; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.organization_temporal_history (
    id integer NOT NULL,
    tenant_id uuid NOT NULL,
    code character varying(50) NOT NULL,
    parent_code character varying(50),
    name character varying(255) NOT NULL,
    unit_type character varying(50) NOT NULL,
    status character varying(20) NOT NULL,
    level integer NOT NULL,
    path text NOT NULL,
    sort_order integer DEFAULT 0,
    description text,
    effective_date date NOT NULL,
    end_date date,
    change_reason text,
    is_current boolean DEFAULT false,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.organization_temporal_history OWNER TO "user";

--
-- Name: organization_temporal_history_id_seq; Type: SEQUENCE; Schema: public; Owner: user
--

CREATE SEQUENCE public.organization_temporal_history_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.organization_temporal_history_id_seq OWNER TO "user";

--
-- Name: organization_temporal_history_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: user
--

ALTER SEQUENCE public.organization_temporal_history_id_seq OWNED BY public.organization_temporal_history.id;


--
-- Name: organization_temporal_history id; Type: DEFAULT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.organization_temporal_history ALTER COLUMN id SET DEFAULT nextval('public.organization_temporal_history_id_seq'::regclass);


--
-- Data for Name: organization_temporal_history; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.organization_temporal_history (id, tenant_id, code, parent_code, name, unit_type, status, level, path, sort_order, description, effective_date, end_date, change_reason, is_current, created_at, updated_at) FROM stdin;
6	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	1000056	\N	测试部门	DEPARTMENT	ACTIVE	1	/1000056	0	新成立的测试部门	2023-01-01	2023-06-30	部门成立	f	2023-01-01 00:00:00	2023-01-01 00:00:00
7	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	1000056	\N	扩展测试部门	DEPARTMENT	ACTIVE	1	/1000056	0	部门规模扩展，业务增长	2023-07-01	2023-12-31	业务扩展需要	f	2023-07-01 00:00:00	2023-07-01 00:00:00
8	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	1000056	\N	重组后的测试部门	COST_CENTER	ACTIVE	1	/1000056	0	通过事件API更新的部门信息	2024-01-01	2025-12-31	部门重组，改为成本中心	f	2024-01-01 00:00:00	2025-08-11 03:42:01
9	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9	1000056	\N	创新研发中心	DEPARTMENT	PLANNED	1	/1000056	0	转型为创新研发中心	2026-01-01	\N	战略转型升级	t	2025-08-11 14:00:00	2025-08-11 14:00:00
\.


--
-- Name: organization_temporal_history_id_seq; Type: SEQUENCE SET; Schema: public; Owner: user
--

SELECT pg_catalog.setval('public.organization_temporal_history_id_seq', 9, true);


--
-- Name: organization_temporal_history organization_temporal_history_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.organization_temporal_history
    ADD CONSTRAINT organization_temporal_history_pkey PRIMARY KEY (id);


--
-- Name: idx_org_temporal_code; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_org_temporal_code ON public.organization_temporal_history USING btree (code);


--
-- Name: idx_org_temporal_current; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_org_temporal_current ON public.organization_temporal_history USING btree (code, is_current) WHERE (is_current = true);


--
-- Name: idx_org_temporal_effective_date; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_org_temporal_effective_date ON public.organization_temporal_history USING btree (effective_date);


--
-- Name: TABLE organization_temporal_history; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT ON TABLE public.organization_temporal_history TO debezium_user;


--
-- PostgreSQL database dump complete
--

