-- hierarchy-consistency-check.sql
-- 目的：检测 organization_units 表层级字段（path/code_path/name_path/level）的一致性
-- 使用方式：由 scripts/maintenance/run-hierarchy-consistency-check.sh 或 CI 守卫脚本调用

WITH current_units AS (
    SELECT
        tenant_id,
        code,
        parent_code,
        code_path,
        name_path,
        level,
        status
    FROM organization_units
    WHERE is_current = true
      AND status <> 'DELETED'
),
annotated AS (
    SELECT
        cu.*,
        parent.code_path   AS parent_code_path,
        parent.name_path   AS parent_name_path,
        parent.level       AS parent_level,
        NULLIF(TRIM(BOTH '/' FROM cu.code_path), '')  AS trimmed_code_path,
        NULLIF(TRIM(BOTH '/' FROM cu.name_path), '')  AS trimmed_name_path
    FROM current_units cu
    LEFT JOIN current_units parent
        ON parent.tenant_id = cu.tenant_id
       AND parent.code = cu.parent_code
),
metrics AS (
    SELECT
        a.*,
        CASE
            WHEN trimmed_code_path IS NULL THEN 0
            ELSE array_length(string_to_array(trimmed_code_path, '/'), 1)
        END AS code_depth,
        CASE
            WHEN trimmed_name_path IS NULL THEN 0
            ELSE array_length(string_to_array(trimmed_name_path, '/'), 1)
        END AS name_depth
    FROM annotated a
),
anomalies AS (
    SELECT tenant_id,
           code,
           parent_code,
           level,
           status,
           'missing_code_path' AS anomaly_type,
           'code_path 为空或缺失' AS anomaly_detail,
           code_path,
           name_path
    FROM metrics
    WHERE code_depth = 0

    UNION ALL

    SELECT tenant_id,
           code,
           parent_code,
           level,
           status,
           'missing_name_path' AS anomaly_type,
           'name_path 为空或缺失' AS anomaly_detail,
           code_path,
           name_path
    FROM metrics
    WHERE name_depth = 0

    UNION ALL

    SELECT tenant_id,
           code,
           parent_code,
           level,
           status,
           'depth_level_mismatch' AS anomaly_type,
           FORMAT('code_path 深度 %s 与 level %s 不一致', code_depth, level) AS anomaly_detail,
           code_path,
           name_path
    FROM metrics
    WHERE code_depth > 0
      AND code_depth <> level

    UNION ALL

    SELECT tenant_id,
           code,
           parent_code,
           level,
           status,
           'code_tail_mismatch' AS anomaly_type,
           'code_path 最末段与组织代码不一致' AS anomaly_detail,
           code_path,
           name_path
    FROM metrics
    WHERE code_depth > 0
      AND split_part(trimmed_code_path, '/', code_depth) <> code

    UNION ALL

    SELECT tenant_id,
           code,
           parent_code,
           level,
           status,
           'parent_missing' AS anomaly_type,
           '父组织不存在或非当前记录' AS anomaly_detail,
           code_path,
           name_path
    FROM metrics
    WHERE parent_code IS NOT NULL
      AND parent_code <> ''
      AND parent_level IS NULL

    UNION ALL

    SELECT tenant_id,
           code,
           parent_code,
           level,
           status,
           'parent_path_mismatch' AS anomaly_type,
           'code_path 未以父组织路径为前缀' AS anomaly_detail,
           code_path,
           name_path
    FROM metrics
    WHERE parent_code IS NOT NULL
      AND parent_code <> ''
      AND parent_level IS NOT NULL
      AND parent_code_path IS NOT NULL
      AND NOT (TRIM(BOTH '/' FROM code_path) LIKE TRIM(BOTH '/' FROM parent_code_path) || '/%')

    UNION ALL

    SELECT tenant_id,
           code,
           parent_code,
           level,
           status,
           'root_level_mismatch' AS anomaly_type,
           '根组织 level 应为 1' AS anomaly_detail,
           code_path,
           name_path
    FROM metrics
    WHERE parent_code IS NULL
      AND level <> 1
)
SELECT tenant_id,
       code,
       parent_code,
       level,
       status,
       anomaly_type,
       anomaly_detail,
       code_path,
       name_path
FROM anomalies
ORDER BY tenant_id, code, anomaly_type;
