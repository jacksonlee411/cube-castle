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
CREATE VIEW public.organization_current AS
CREATE VIEW public.organization_stats_view AS
CREATE VIEW public.organization_temporal_current AS
CREATE TRIGGER audit_changes_trigger AFTER INSERT OR DELETE OR UPDATE ON public.organization_units FOR EACH ROW EXECUTE FUNCTION public.log_audit_changes();
CREATE TRIGGER enforce_temporal_flags_trigger BEFORE INSERT OR UPDATE ON public.organization_units FOR EACH ROW EXECUTE FUNCTION public.enforce_temporal_flags();
CREATE TRIGGER trg_prevent_update_deleted BEFORE UPDATE ON public.organization_units FOR EACH ROW WHEN (((old.status)::text = 'DELETED'::text)) EXECUTE FUNCTION public.prevent_update_deleted();
CREATE TRIGGER update_hierarchy_paths_trigger BEFORE INSERT OR UPDATE ON public.organization_units FOR EACH ROW EXECUTE FUNCTION public.update_hierarchy_paths();
CREATE TRIGGER validate_parent_available_trigger BEFORE INSERT ON public.organization_units FOR EACH ROW EXECUTE FUNCTION public.validate_parent_available();
CREATE TRIGGER validate_parent_available_update_trigger BEFORE UPDATE ON public.organization_units FOR EACH ROW WHEN (((new.parent_code IS NOT NULL) AND ((new.parent_code)::text IS DISTINCT FROM (old.parent_code)::text))) EXECUTE FUNCTION public.validate_parent_available();
