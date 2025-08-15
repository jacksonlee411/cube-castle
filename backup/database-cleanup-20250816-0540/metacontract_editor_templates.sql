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
-- Data for Name: metacontract_editor_templates; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.metacontract_editor_templates (id, name, description, category, content, tags, created_at, updated_at) FROM stdin;
cd35b143-a863-4d4b-97e7-b4b1cb903808	Employee Management Template	Basic template for employee management meta-contract	hr	# Employee Management Meta-Contract\n\nversion: "1.0.0"\nname: "employee_management"\ndescription: "Employee management system"\n\nentities:\n  Employee:\n    fields:\n      - name: id\n        type: UUID\n        required: true\n        primary_key: true\n      - name: first_name\n        type: String\n        required: true\n      - name: last_name\n        type: String\n        required: true\n      - name: email\n        type: String\n        required: true\n        unique: true\n      - name: hire_date\n        type: Date\n        required: true\n\nworkflows:\n  employee_onboarding:\n    description: "Employee onboarding process"\n    steps:\n      - name: create_employee\n        action: create\n        entity: Employee\n      - name: send_welcome_email\n        action: notify\n        template: welcome_email	{hr,employee,management,basic}	2025-07-30 23:43:33.491778+00	2025-07-30 23:43:33.491778+00
\.


--
-- Name: metacontract_editor_templates metacontract_editor_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.metacontract_editor_templates
    ADD CONSTRAINT metacontract_editor_templates_pkey PRIMARY KEY (id);


--
-- Name: idx_metacontract_templates_category; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_metacontract_templates_category ON public.metacontract_editor_templates USING btree (category);


--
-- Name: idx_metacontract_templates_tags; Type: INDEX; Schema: public; Owner: user
--

CREATE INDEX idx_metacontract_templates_tags ON public.metacontract_editor_templates USING gin (tags);


--
-- Name: TABLE metacontract_editor_templates; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT ON TABLE public.metacontract_editor_templates TO debezium_user;


--
-- PostgreSQL database dump complete
--

