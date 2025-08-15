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
-- Name: organization_versions_backup_before_deletion; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.organization_versions_backup_before_deletion (
    version_id uuid,
    organization_code character varying(10),
    version integer,
    effective_date date,
    end_date date,
    snapshot_data jsonb,
    change_reason character varying(500),
    created_at timestamp with time zone,
    tenant_id uuid
);


ALTER TABLE public.organization_versions_backup_before_deletion OWNER TO "user";

--
-- Data for Name: organization_versions_backup_before_deletion; Type: TABLE DATA; Schema: public; Owner: user
--

COPY public.organization_versions_backup_before_deletion (version_id, organization_code, version, effective_date, end_date, snapshot_data, change_reason, created_at, tenant_id) FROM stdin;
37147860-5961-4f74-9256-b429bb67d643	1000002	2	2025-08-11	\N	{"code": "1000002", "name": "产品部", "path": "/1000000/1000002", "level": 2, "status": "ACTIVE", "version": 2, "tenant_id": "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9", "unit_type": "DEPARTMENT", "created_at": "2025-08-05T11:23:01.426455Z", "is_current": true, "sort_order": 0, "updated_at": "2025-08-11T00:18:19.805932Z", "description": "产品管理部门", "parent_code": "1000000", "change_reason": "部门职能调整 - 产品部将于2025年12月1日开始承担新的产品战略规划职能", "effective_date": "2025-08-11T00:21:32.831Z", "supersedes_version": 1}	部门职能调整 - 产品部将于2025年12月1日开始承担新的产品战略规划职能	2025-08-11 00:21:32.839038+00	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9
eca93bbb-de2a-49f9-a162-2b91c2b5814f	1000056	1	2025-07-12	\N	{"code": "1000056", "name": "重组后的测试部门", "path": "/1000056", "level": 1, "status": "INACTIVE", "tenant_id": "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9", "unit_type": "COST_CENTER", "created_at": "2025-08-09T07:21:10.177689+00:00", "sort_order": 0, "updated_at": "2025-08-11T01:32:53.157447+00:00", "description": "通过事件API更新的部门信息"}	手动同步历史数据	2025-08-11 03:38:18.606802+00	3b99930c-4dc6-4cc9-8e4d-7d960a931cb9
\.


--
-- Name: TABLE organization_versions_backup_before_deletion; Type: ACL; Schema: public; Owner: user
--

GRANT SELECT ON TABLE public.organization_versions_backup_before_deletion TO debezium_user;


--
-- PostgreSQL database dump complete
--

