-- +goose Up
WITH params AS (
    SELECT
        '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9'::uuid AS tenant_id,
        DATE '2024-01-01' AS effective_date
),
upsert_group AS (
    INSERT INTO public.job_family_groups (
        tenant_id,
        family_group_code,
        name,
        description,
        status,
        effective_date,
        end_date,
        is_current
    )
    SELECT
        tenant_id,
        'OPER',
        '运营管理职类组',
        'Position CRUD / Playwright 参考数据：运营类岗位',
        'ACTIVE',
        effective_date,
        NULL,
        true
    FROM params
    ON CONFLICT (tenant_id, family_group_code, effective_date)
    DO UPDATE
        SET name        = EXCLUDED.name,
            description = EXCLUDED.description,
            status      = 'ACTIVE',
            end_date    = NULL,
            is_current  = true,
            updated_at  = NOW()
    RETURNING record_id
),
upsert_family AS (
    INSERT INTO public.job_families (
        tenant_id,
        family_code,
        family_group_code,
        parent_record_id,
        name,
        description,
        status,
        effective_date,
        end_date,
        is_current
    )
    SELECT
        p.tenant_id,
        'OPER-OPS',
        'OPER',
        g.record_id,
        '运营运营（OPS）',
        'Operations family used by position CRUD E2E',
        'ACTIVE',
        p.effective_date,
        NULL,
        true
    FROM params p
    CROSS JOIN upsert_group g
    ON CONFLICT (tenant_id, family_code, effective_date)
    DO UPDATE
        SET parent_record_id  = EXCLUDED.parent_record_id,
            family_group_code = EXCLUDED.family_group_code,
            name              = EXCLUDED.name,
            description       = EXCLUDED.description,
            status            = 'ACTIVE',
            end_date          = NULL,
            is_current        = true,
            updated_at        = NOW()
    RETURNING record_id
),
upsert_role AS (
    INSERT INTO public.job_roles (
        tenant_id,
        role_code,
        family_code,
        parent_record_id,
        name,
        description,
        competency_model,
        status,
        effective_date,
        end_date,
        is_current
    )
    SELECT
        p.tenant_id,
        'OPER-OPS-MGR',
        'OPER-OPS',
        f.record_id,
        '运营经理',
        '负责跨区域运营排班、流程优化',
        '{}'::jsonb,
        'ACTIVE',
        p.effective_date,
        NULL,
        true
    FROM params p
    CROSS JOIN upsert_family f
    ON CONFLICT (tenant_id, role_code, effective_date)
    DO UPDATE
        SET family_code      = EXCLUDED.family_code,
            parent_record_id = EXCLUDED.parent_record_id,
            name             = EXCLUDED.name,
            description      = EXCLUDED.description,
            competency_model = EXCLUDED.competency_model,
            status           = 'ACTIVE',
            end_date         = NULL,
            is_current       = true,
            updated_at       = NOW()
    RETURNING record_id
),
level_configs AS (
    SELECT *
    FROM (VALUES
        ('S1', 'S1', '运营经理 S1', '初级运营经理'),
        ('S2', 'S2', '运营经理 S2', '中级运营经理'),
        ('S3', 'S3', '运营经理 S3', '高级运营经理')
    ) AS t(level_code, level_rank, name, description)
)
INSERT INTO public.job_levels (
    tenant_id,
    level_code,
    role_code,
    parent_record_id,
    level_rank,
    name,
    description,
    status,
    effective_date,
    end_date,
    is_current
)
SELECT
    p.tenant_id,
    lc.level_code,
    'OPER-OPS-MGR',
    r.record_id,
    lc.level_rank,
    lc.name,
    lc.description,
    'ACTIVE',
    p.effective_date,
    NULL,
    true
FROM params p
CROSS JOIN upsert_role r
CROSS JOIN level_configs lc
ON CONFLICT (tenant_id, level_code, effective_date)
DO UPDATE
    SET role_code        = EXCLUDED.role_code,
        parent_record_id = EXCLUDED.parent_record_id,
        level_rank       = EXCLUDED.level_rank,
        name             = EXCLUDED.name,
        description      = EXCLUDED.description,
        status           = 'ACTIVE',
        end_date         = NULL,
        is_current       = true,
        updated_at       = NOW();

-- +goose Down
DELETE FROM public.job_levels
WHERE tenant_id = '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9'
  AND role_code = 'OPER-OPS-MGR'
  AND level_code IN ('S1', 'S2', 'S3');

DELETE FROM public.job_roles
WHERE tenant_id = '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9'
  AND role_code = 'OPER-OPS-MGR';

DELETE FROM public.job_families
WHERE tenant_id = '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9'
  AND family_code = 'OPER-OPS';

DELETE FROM public.job_family_groups
WHERE tenant_id = '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9'
  AND family_group_code = 'OPER';
