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
CREATE TABLE public.audit_logs (
CREATE TABLE public.job_families (
CREATE TABLE public.job_family_groups (
CREATE TABLE public.job_levels (
CREATE TABLE public.job_roles (
CREATE TABLE public.organization_units (
CREATE VIEW public.organization_current AS
CREATE VIEW public.organization_stats_view AS
CREATE VIEW public.organization_temporal_current AS
CREATE TABLE public.organization_units_backup_temporal (
CREATE TABLE public.organization_units_unittype_backup (
CREATE TABLE public.position_assignments (
CREATE TABLE public.positions (
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
CREATE TRIGGER audit_changes_trigger AFTER INSERT OR DELETE OR UPDATE ON public.organization_units FOR EACH ROW EXECUTE FUNCTION public.log_audit_changes();
CREATE TRIGGER enforce_temporal_flags_trigger BEFORE INSERT OR UPDATE ON public.organization_units FOR EACH ROW EXECUTE FUNCTION public.enforce_temporal_flags();
CREATE TRIGGER trg_prevent_update_deleted BEFORE UPDATE ON public.organization_units FOR EACH ROW WHEN (((old.status)::text = 'DELETED'::text)) EXECUTE FUNCTION public.prevent_update_deleted();
CREATE TRIGGER update_hierarchy_paths_trigger BEFORE INSERT OR UPDATE ON public.organization_units FOR EACH ROW EXECUTE FUNCTION public.update_hierarchy_paths();
CREATE TRIGGER validate_parent_available_trigger BEFORE INSERT ON public.organization_units FOR EACH ROW EXECUTE FUNCTION public.validate_parent_available();
CREATE TRIGGER validate_parent_available_update_trigger BEFORE UPDATE ON public.organization_units FOR EACH ROW WHEN (((new.parent_code IS NOT NULL) AND ((new.parent_code)::text IS DISTINCT FROM (old.parent_code)::text))) EXECUTE FUNCTION public.validate_parent_available();
