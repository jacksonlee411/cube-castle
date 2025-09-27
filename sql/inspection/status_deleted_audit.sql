-- status_deleted_audit.sql
-- 用于生成软删除状态审计 JSON 报告，请将输出重定向至
-- `reports/temporal/status-only-audit.json`

\set ON_ERROR_STOP on
\pset pager off
\pset footer off
\pset format unaligned
\pset fieldsep ''
\pset tuples_only on

WITH summary AS (
    SELECT jsonb_build_object(
        'totalRecords', COUNT(*)::bigint,
        'deletedWithoutTimestamp', SUM(CASE WHEN status = 'DELETED' AND deleted_at IS NULL THEN 1 ELSE 0 END)::bigint,
        'timestampWithoutDeleted', SUM(CASE WHEN status <> 'DELETED' AND deleted_at IS NOT NULL THEN 1 ELSE 0 END)::bigint,
        'consistentDeleted', SUM(CASE WHEN status = 'DELETED' AND deleted_at IS NOT NULL THEN 1 ELSE 0 END)::bigint
    ) AS data
    FROM organization_units
), deleted_without_timestamp AS (
    SELECT COALESCE(
        jsonb_agg(
            jsonb_build_object(
                'tenantId', tenant_id::text,
                'code', code,
                'effectiveDate', to_char(effective_date, 'YYYY-MM-DD'),
                'status', status,
                'deletedAt', NULL
            )
            ORDER BY tenant_id, code, effective_date
        ),
        '[]'::jsonb
    ) AS data
    FROM organization_units
    WHERE status = 'DELETED' AND deleted_at IS NULL
), timestamp_without_deleted AS (
    SELECT COALESCE(
        jsonb_agg(
            jsonb_build_object(
                'tenantId', tenant_id::text,
                'code', code,
                'effectiveDate', to_char(effective_date, 'YYYY-MM-DD'),
                'status', status,
                'deletedAt', to_char(timezone('UTC', deleted_at), 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')
            )
            ORDER BY tenant_id, code, effective_date
        ),
        '[]'::jsonb
    ) AS data
    FROM organization_units
    WHERE status <> 'DELETED' AND deleted_at IS NOT NULL
), conflicting_deleted_at_base AS (
    SELECT
        tenant_id,
        code,
        effective_date,
        COUNT(*)::bigint as version_count,
        MIN(deleted_at) as min_deleted_at,
        MAX(deleted_at) as max_deleted_at
    FROM organization_units
    GROUP BY tenant_id, code, effective_date
    HAVING COUNT(DISTINCT deleted_at) > 1
), conflicting_deleted_at AS (
    SELECT COALESCE(
        jsonb_agg(
            jsonb_build_object(
                'tenantId', tenant_id::text,
                'code', code,
                'effectiveDate', to_char(effective_date, 'YYYY-MM-DD'),
                'versionCount', version_count,
                'minDeletedAt', to_char(timezone('UTC', min_deleted_at), 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'),
                'maxDeletedAt', to_char(timezone('UTC', max_deleted_at), 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')
            )
            ORDER BY tenant_id, code, effective_date
        ),
        '[]'::jsonb
    ) AS data
    FROM conflicting_deleted_at_base
)
SELECT jsonb_pretty(
    jsonb_build_object(
        'generatedAt', to_char(timezone('UTC', now()), 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'),
        'summary', COALESCE((SELECT data FROM summary), '{}'::jsonb),
        'deletedWithoutTimestamp', (SELECT data FROM deleted_without_timestamp),
        'timestampWithoutDeleted', (SELECT data FROM timestamp_without_deleted),
        'conflictingDeletedAtValues', (SELECT data FROM conflicting_deleted_at)
    )
);
