-- name: UpsertStandardObject :one
INSERT INTO standard_objects (
    object_type,
    code,
    tenant_code,
    display_name,
    status,
    labels,
    schema_version,
    data_classification,
    retention_policy,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
ON CONFLICT (tenant_code, object_type, code)
DO UPDATE SET
    display_name      = EXCLUDED.display_name,
    status            = EXCLUDED.status,
    labels            = EXCLUDED.labels,
    schema_version    = EXCLUDED.schema_version,
    data_classification = EXCLUDED.data_classification,
    retention_policy  = EXCLUDED.retention_policy,
    updated_at        = now()
RETURNING *;

-- name: InsertStandardObjectVersion :one
INSERT INTO standard_object_versions (
    object_id,
    version_code,
    effective_date,
    end_date,
    validity_range,
    transaction_range,
    is_current,
    payload,
    audit,
    checksum
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: ListStandardObjectVersionsForObject :many
SELECT *
FROM standard_object_versions
WHERE object_id = $1
ORDER BY effective_date;

-- name: GetStandardObjectKernel :one
SELECT *
FROM standard_objects
WHERE tenant_code = $1
  AND object_type = $2
  AND code = $3;

-- name: GetVersionAsOf :one
SELECT sov.*
FROM standard_object_versions sov
JOIN standard_objects so ON so.id = sov.object_id
WHERE so.tenant_code = $1
  AND so.object_type = $2
  AND so.code = $3
  AND sov.validity_range @> $4::timestamptz
  AND sov.transaction_range @> $5::timestamptz
ORDER BY sov.effective_date DESC
LIMIT 1;

-- name: UpsertStandardObjectLink :one
INSERT INTO standard_object_links (
    link_type,
    source_object_id,
    target_object_id,
    tenant_code,
    validity_range,
    transaction_range,
    attributes,
    created_by,
    updated_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $8
)
ON CONFLICT (link_type, source_object_id, target_object_id)
DO UPDATE SET
    validity_range    = EXCLUDED.validity_range,
    transaction_range = EXCLUDED.transaction_range,
    attributes        = EXCLUDED.attributes,
    updated_by        = EXCLUDED.updated_by,
    updated_at        = now()
RETURNING *;

-- name: ListStandardObjectLinks :many
SELECT *
FROM standard_object_links
WHERE tenant_code = $1
  AND link_type = $2;

-- name: ListLinksForSource :many
SELECT l.id,
       l.link_type,
       l.source_object_id,
       l.target_object_id,
       l.tenant_code,
       l.validity_range,
       l.transaction_range,
       l.attributes,
       l.created_by,
       l.created_at,
       l.updated_by,
       l.updated_at,
       src.code AS source_code,
       tgt.code AS target_code
FROM standard_object_links l
JOIN standard_objects src ON src.id = l.source_object_id
JOIN standard_objects tgt ON tgt.id = l.target_object_id
WHERE l.source_object_id = $1;

-- name: InsertStandardObjectTranslation :one
INSERT INTO standard_object_translations (
    object_id,
    tenant_code,
    locale,
    display_name,
    description,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6
)
ON CONFLICT (object_id, locale)
DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description  = EXCLUDED.description,
    metadata     = EXCLUDED.metadata,
    updated_at   = now()
RETURNING *;

-- name: InsertStandardObjectAttachment :one
INSERT INTO standard_object_attachments (
    object_id,
    tenant_code,
    attachment_type,
    uri,
    labels,
    metadata,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: UpsertStandardObjectMetadata :one
INSERT INTO standard_object_metadata (
    object_id,
    tenant_code,
    meta_key,
    meta_value
) VALUES (
    $1, $2, $3, $4
)
ON CONFLICT (object_id, meta_key)
DO UPDATE SET
    meta_value = EXCLUDED.meta_value,
    updated_at = now()
RETURNING *;

-- name: InsertStandardObjectMetric :one
INSERT INTO standard_object_metrics (
    object_id,
    metric_type,
    metric_value,
    labels,
    tenant_code,
    recorded_at
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: ListObjectsForTenant :many
SELECT *
FROM standard_objects
WHERE tenant_code = $1
  AND object_type = $2;
