-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE VIEW public.organization_units_v AS
SELECT
    so.id AS record_id,
    so.tenant_code::uuid AS tenant_id,
    so.code,
    NULLIF(sov.payload->>'parentCode', '') AS parent_code,
    COALESCE(NULLIF(sov.payload->>'name', ''), so.display_name) AS name,
    COALESCE(NULLIF(sov.payload->>'unitType', ''), so.labels->>'unitType', 'ORGANIZATION_UNIT') AS unit_type,
    COALESCE(NULLIF(sov.payload->>'status', ''), so.status) AS status,
    COALESCE((NULLIF(sov.payload->>'level', ''))::integer, 1) AS level,
    COALESCE((NULLIF(sov.payload->>'level', ''))::integer, 1) AS hierarchy_depth,
    COALESCE(NULLIF(sov.payload->>'codePath', ''), '/' || so.code) AS code_path,
    COALESCE(NULLIF(sov.payload->>'namePath', ''), '/' || so.display_name) AS name_path,
    COALESCE((NULLIF(sov.payload->>'sortOrder', ''))::integer, 0) AS sort_order,
    NULLIF(sov.payload->>'description', '') AS description,
    COALESCE(sov.payload->'profile', '{}'::jsonb) AS profile,
    COALESCE(sov.payload->'metadata', '{}'::jsonb) AS metadata,
    so.created_at,
    so.updated_at,
    sov.effective_date,
    sov.end_date,
    NULLIF(sov.audit->>'changeReason', '') AS change_reason,
    sov.is_current,
    NULL::timestamptz AS deleted_at,
    NULL::uuid AS deleted_by,
    NULL::text AS deletion_reason,
    NULL::timestamptz AS suspended_at,
    NULL::uuid AS suspended_by,
    NULL::text AS suspension_reason,
    NULL::uuid AS operated_by_id,
    NULL::text AS operated_by_name,
    'UPSERT'::text AS operation_type,
    lower(sov.validity_range) AS effective_from,
    NULLIF(upper(sov.validity_range), 'infinity'::timestamptz) AS effective_to,
    NULL::uuid AS changed_by,
    NULL::uuid AS approved_by
FROM public.standard_objects so
JOIN public.standard_object_versions sov ON sov.object_id = so.id
WHERE so.object_type = 'ORGANIZATION_UNIT';

CREATE OR REPLACE VIEW public.positions_v AS
SELECT
    so.id AS record_id,
    so.tenant_code::uuid AS tenant_id,
    so.code,
    COALESCE(NULLIF(sov.payload->>'title', ''), so.display_name) AS title,
    NULLIF(sov.payload->>'jobProfileCode', '') AS job_profile_code,
    NULLIF(sov.payload->>'jobProfileName', '') AS job_profile_name,
    NULLIF(sov.payload->>'jobFamilyGroupCode', '') AS job_family_group_code,
    NULL::text AS job_family_group_name,
    NULL::uuid AS job_family_group_record_id,
    NULLIF(sov.payload->>'jobFamilyCode', '') AS job_family_code,
    NULL::text AS job_family_name,
    NULL::uuid AS job_family_record_id,
    NULLIF(sov.payload->>'jobRoleCode', '') AS job_role_code,
    NULL::text AS job_role_name,
    NULL::uuid AS job_role_record_id,
    NULLIF(sov.payload->>'jobLevelCode', '') AS job_level_code,
    NULL::text AS job_level_name,
    NULL::uuid AS job_level_record_id,
    NULLIF(sov.payload->>'organizationCode', '') AS organization_code,
    NULLIF(sov.payload->>'organizationName', '') AS organization_name,
    NULLIF(sov.payload->>'positionType', '') AS position_type,
    COALESCE(NULLIF(sov.payload->>'status', ''), so.status) AS status,
    NULLIF(sov.payload->>'employmentType', '') AS employment_type,
    COALESCE((NULLIF(sov.payload->>'headcountCapacity', ''))::numeric, 0.0) AS headcount_capacity,
    COALESCE((NULLIF(sov.payload->>'headcountInUse', ''))::numeric, 0.0) AS headcount_in_use,
    NULLIF(sov.payload->>'gradeLevel', '') AS grade_level,
    NULLIF(sov.payload->>'costCenterCode', '') AS cost_center_code,
    NULLIF(sov.payload->>'reportsToPositionCode', '') AS reports_to_position_code,
    COALESCE(sov.payload->'profile', '{}'::jsonb) AS profile,
    so.created_at,
    so.updated_at,
    sov.effective_date,
    sov.end_date,
    sov.is_current,
    NULLIF(sov.audit->>'operationType', '') AS operation_type,
    NULLIF(sov.audit->'operatedBy'->>'id', '')::uuid AS operated_by_id,
    NULLIF(sov.audit->'operatedBy'->>'name', '') AS operated_by_name,
    NULLIF(sov.audit->>'operationReason', '') AS operation_reason,
    NULL::timestamptz AS deleted_at,
    lower(sov.validity_range) AS effective_from,
    NULLIF(upper(sov.validity_range), 'infinity'::timestamptz) AS effective_to
FROM public.standard_objects so
JOIN public.standard_object_versions sov ON sov.object_id = so.id
WHERE so.object_type = 'POSITION_ROLE';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS public.positions_v;
DROP VIEW IF EXISTS public.organization_units_v;
-- +goose StatementEnd
