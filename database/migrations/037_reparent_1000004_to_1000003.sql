-- 037_reparent_1000004_to_1000003.sql
-- 目的：将组织 1000004（国际业务部）的当前版本挂载到 1000003（市场部）之下，
--       并同步刷新子树的层级字段（level、hierarchy_depth、path、code_path、name_path）。

DO $$
DECLARE
    parent_exists BOOLEAN;
    target_exists BOOLEAN;
BEGIN
    SELECT EXISTS (
        SELECT 1
        FROM organization_units
        WHERE code = '1000003'
          AND tenant_id = '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9'
          AND is_current = true
          AND status <> 'DELETED'
    ) INTO parent_exists;

    SELECT EXISTS (
        SELECT 1
        FROM organization_units
        WHERE code = '1000004'
          AND tenant_id = '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9'
          AND is_current = true
          AND status <> 'DELETED'
    ) INTO target_exists;

    IF NOT (parent_exists AND target_exists) THEN
        RAISE NOTICE 'Skip 037_reparent_1000004_to_1000003: required organizations absent (parent=%, target=%)', parent_exists, target_exists;
        RETURN;
    END IF;

    WITH RECURSIVE parent_ctx AS (
        SELECT
            tenant_id,
            code,
            name,
            COALESCE(NULLIF(code_path, ''), '/' || code) AS code_path,
            COALESCE(NULLIF(name_path, ''), '/' || name) AS name_path,
            COALESCE(level, 0) AS level,
            COALESCE(hierarchy_depth, COALESCE(level, 0) + 1) AS hierarchy_depth
        FROM organization_units
        WHERE code = '1000003'
          AND tenant_id = '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9'
          AND is_current = true
          AND status <> 'DELETED'
        LIMIT 1
    ),
    subtree AS (
        SELECT
            child.record_id,
            child.tenant_id,
            child.code,
            child.name,
            parent_ctx.code AS new_parent_code,
            CASE
                WHEN parent_ctx.code_path = '' THEN '/' || child.code
                ELSE parent_ctx.code_path || '/' || child.code
            END AS new_code_path,
            CASE
                WHEN parent_ctx.name_path = '' THEN '/' || child.name
                ELSE parent_ctx.name_path || '/' || child.name
            END AS new_name_path,
            parent_ctx.level + 1 AS new_level,
            parent_ctx.hierarchy_depth + 1 AS new_hierarchy_depth
        FROM organization_units child
        JOIN parent_ctx ON child.tenant_id = parent_ctx.tenant_id
        WHERE child.code = '1000004'
          AND child.is_current = true
          AND child.status <> 'DELETED'

        UNION ALL

        SELECT
            c.record_id,
            c.tenant_id,
            c.code,
            c.name,
            c.parent_code AS new_parent_code,
            CASE
                WHEN st.new_code_path = '' THEN '/' || c.code
                ELSE st.new_code_path || '/' || c.code
            END AS new_code_path,
            CASE
                WHEN st.new_name_path = '' THEN '/' || c.name
                ELSE st.new_name_path || '/' || c.name
            END AS new_name_path,
            st.new_level + 1 AS new_level,
            st.new_hierarchy_depth + 1 AS new_hierarchy_depth
        FROM organization_units c
        JOIN subtree st
          ON c.tenant_id = st.tenant_id
         AND c.parent_code = st.code
        WHERE c.is_current = true
          AND c.status <> 'DELETED'
    )
    UPDATE organization_units ou
    SET parent_code = subtree.new_parent_code,
        level = subtree.new_level,
        hierarchy_depth = subtree.new_hierarchy_depth,
        path = subtree.new_code_path,
        code_path = subtree.new_code_path,
        name_path = subtree.new_name_path,
        updated_at = NOW()
    FROM subtree
    WHERE ou.record_id = subtree.record_id;

    RAISE NOTICE 'Reparented organization 1000004 under 1000003';
END
$$;

SELECT 'reparented 1000004 under 1000003 (noop-friendly)' AS info, NOW() AS applied_at;
