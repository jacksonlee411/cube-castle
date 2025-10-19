-- 046_seed_positions_data.sql
-- 目的：为默认租户注入职位管理所需的真实岗位数据，替代前端演示用 Mock 数据。

BEGIN;

WITH
    tenant AS (
        SELECT '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9'::uuid AS tenant_id,
               '789e0123-e89b-12d3-a456-426614174001'::uuid AS operator_id,
               'System Admin'::text AS operator_name
    ),

    seed_groups AS (
        SELECT t.tenant_id,
               g.code AS family_group_code,
               g.name,
               g.description,
               g.effective_date
        FROM tenant t
        CROSS JOIN (VALUES
            ('PROF', '专业序列', '覆盖专业及职能岗位', '2024-01-01'::date),
            ('OPER', '运营序列', '覆盖现场与支持运营岗位', '2024-01-01'::date)
        ) AS g(code, name, description, effective_date)
    ),

    inserted_groups AS (
        INSERT INTO job_family_groups (
            tenant_id, family_group_code, name, description, status,
            effective_date, is_current, created_at, updated_at
        )
        SELECT sg.tenant_id, sg.family_group_code, sg.name, sg.description,
               'ACTIVE', sg.effective_date, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
        FROM seed_groups sg
        ON CONFLICT (tenant_id, family_group_code, effective_date) DO NOTHING
        RETURNING tenant_id, family_group_code, record_id, name
    ),

    groups AS (
        SELECT tenant_id, family_group_code, record_id, name
        FROM inserted_groups
        UNION ALL
        SELECT sg.tenant_id, sg.family_group_code, jfg.record_id, jfg.name
        FROM seed_groups sg
        JOIN job_family_groups jfg
          ON jfg.tenant_id = sg.tenant_id
         AND jfg.family_group_code = sg.family_group_code
         AND jfg.effective_date = sg.effective_date
    ),

    seed_families AS (
        SELECT g.tenant_id,
               f.family_code,
               g.family_group_code,
               g.record_id AS parent_record_id,
               f.name,
               f.description,
               f.effective_date
        FROM groups g
        JOIN (VALUES
            ('PROF-HR', 'PROF', '人力资源序列', '聚焦 HR 管理与业务伙伴岗位', '2024-01-01'::date),
            ('PROF-IT', 'PROF', '技术研发序列', '覆盖后端与前端研发岗位', '2024-01-01'::date),
            ('PROF-UX', 'PROF', '产品与设计序列', '聚焦产品体验与视觉设计岗位', '2024-01-01'::date),
            ('OPER-SITE', 'OPER', '现场运营序列', '面向现场运维与项目管理岗位', '2024-01-01'::date)
        ) AS f(family_code, family_group_code, name, description, effective_date)
          ON f.family_group_code = g.family_group_code
    ),

    inserted_families AS (
        INSERT INTO job_families (
            tenant_id, family_code, family_group_code, parent_record_id,
            name, description, status, effective_date, is_current,
            created_at, updated_at
        )
        SELECT sf.tenant_id, sf.family_code, sf.family_group_code,
               sf.parent_record_id, sf.name, sf.description,
               'ACTIVE', sf.effective_date, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
        FROM seed_families sf
        ON CONFLICT (tenant_id, family_code, effective_date) DO NOTHING
        RETURNING tenant_id, family_code, record_id, name
    ),

    families AS (
        SELECT tenant_id, family_code, record_id, name
        FROM inserted_families
        UNION ALL
        SELECT sf.tenant_id, sf.family_code, jf.record_id, jf.name
        FROM seed_families sf
        JOIN job_families jf
          ON jf.tenant_id = sf.tenant_id
         AND jf.family_code = sf.family_code
         AND jf.effective_date = sf.effective_date
    ),

    seed_roles AS (
        SELECT f.tenant_id,
               r.role_code,
               f.family_code,
               f.record_id AS parent_record_id,
               r.name,
               r.description,
               r.effective_date
        FROM families f
        JOIN (VALUES
            ('PROF-HR-MGR', 'PROF-HR', '人力资源经理', '负责组织发展与人力规划', '2024-01-01'::date),
            ('PROF-IT-BE', 'PROF-IT', '后端工程师', '负责核心服务开发与维护', '2024-01-01'::date),
            ('PROF-IT-FE', 'PROF-IT', '前端工程师', '负责前端界面与交互实现', '2024-01-01'::date),
            ('PROF-UX-DES', 'PROF-UX', '产品设计师', '负责产品体验与视觉方案', '2024-01-01'::date),
            ('OPER-SITE-SUP', 'OPER-SITE', '运营主管', '负责现场运营统筹与协调', '2024-01-01'::date)
        ) AS r(role_code, family_code, name, description, effective_date)
          ON r.family_code = f.family_code
    ),

    inserted_roles AS (
        INSERT INTO job_roles (
            tenant_id, role_code, family_code, parent_record_id,
            name, description, competency_model, status,
            effective_date, is_current, created_at, updated_at
        )
        SELECT sr.tenant_id, sr.role_code, sr.family_code, sr.parent_record_id,
               sr.name, sr.description, '{}'::jsonb, 'ACTIVE',
               sr.effective_date, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
        FROM seed_roles sr
        ON CONFLICT (tenant_id, role_code, effective_date) DO NOTHING
        RETURNING tenant_id, role_code, record_id, name
    ),

    roles AS (
        SELECT tenant_id, role_code, record_id, name
        FROM inserted_roles
        UNION ALL
        SELECT sr.tenant_id, sr.role_code, jr.record_id, jr.name
        FROM seed_roles sr
        JOIN job_roles jr
          ON jr.tenant_id = sr.tenant_id
         AND jr.role_code = sr.role_code
         AND jr.effective_date = sr.effective_date
    ),

    seed_levels AS (
        SELECT r.tenant_id,
               l.level_code,
               r.role_code,
               r.record_id AS parent_record_id,
               l.level_rank,
               l.name,
               l.effective_date
        FROM roles r
        JOIN (VALUES
            ('P3', 'PROF-HR-MGR', 'MID', 'P3 管理层级', '2024-01-01'::date),
            ('P5', 'PROF-IT-BE', 'SENIOR', 'P5 高级技术层级', '2024-01-01'::date),
            ('P4-FE', 'PROF-IT-FE', 'SENIOR', 'P4 前端研发层级', '2024-01-01'::date),
            ('P4-UX', 'PROF-UX-DES', 'SENIOR', 'P4 设计体验层级', '2024-01-01'::date),
            ('M1', 'OPER-SITE-SUP', 'MANAGER', 'M1 现场管理层级', '2024-01-01'::date)
        ) AS l(level_code, role_code, level_rank, name, effective_date)
          ON l.role_code = r.role_code
    ),

    inserted_levels AS (
        INSERT INTO job_levels (
            tenant_id, level_code, role_code, parent_record_id,
            level_rank, name, description, status,
            effective_date, is_current, created_at, updated_at
        )
        SELECT sl.tenant_id, sl.level_code, sl.role_code, sl.parent_record_id,
               sl.level_rank, sl.name, NULL, 'ACTIVE',
               sl.effective_date, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
        FROM seed_levels sl
        ON CONFLICT (tenant_id, level_code, effective_date) DO NOTHING
        RETURNING tenant_id, level_code, role_code, record_id, name
    ),

    levels AS (
        SELECT tenant_id, level_code, role_code, record_id, name
        FROM inserted_levels
        UNION ALL
        SELECT sl.tenant_id, sl.level_code, sl.role_code, jl.record_id, jl.name
        FROM seed_levels sl
        JOIN job_levels jl
          ON jl.tenant_id = sl.tenant_id
         AND jl.level_code = sl.level_code
         AND jl.effective_date = sl.effective_date
    ),

    position_seed AS (
        SELECT t.tenant_id,
               t.operator_id,
               t.operator_name,
               ps.code,
               ps.title,
               ps.job_profile_code,
               ps.job_profile_name,
               ps.job_family_group_code,
               ps.job_family_code,
               ps.job_role_code,
               ps.job_level_code,
               ps.organization_code,
               ps.organization_name,
               ps.position_type,
               ps.status,
               ps.employment_type,
               ps.headcount_capacity,
               ps.headcount_in_use,
               ps.grade_level,
               ps.reports_to_position_code,
               ps.effective_date,
               ps.end_date,
               ps.is_current,
               ps.operation_reason,
               ps.current_holder_name,
               ps.filled_date,
               ps.current_assignment_type
        FROM tenant t
        CROSS JOIN (VALUES
            ('P9000001', '人力资源经理', 'HR-MGR', '人力资源经理岗位', 'PROF', 'PROF-HR', 'PROF-HR-MGR', 'P3', '1000001', '技术部', 'REGULAR', 'ACTIVE', 'FULL_TIME', 1.0, 1.0, 'L3', NULL, '2024-01-01'::date, NULL::date, true, '组织发展方案落地', '刘敏', '2024-01-15'::date, 'PRIMARY'),
            ('P9000002', '高级后端工程师', 'ENG-BACKEND', '高级后端工程师', 'PROF', 'PROF-IT', 'PROF-IT-BE', 'P5', '1000012', '后端开发组', 'REGULAR', 'ACTIVE', 'FULL_TIME', 2.0, 1.0, NULL, 'P9000001', '2024-02-01'::date, NULL::date, true, '核心服务扩编', '罗明', '2024-02-20'::date, 'PRIMARY'),
            ('P9000003', '前端工程师', 'ENG-FRONTEND', '前端工程师', 'PROF', 'PROF-IT', 'PROF-IT-FE', 'P4-FE', '1000011', '前端开发组', 'REGULAR', 'VACANT', 'FULL_TIME', 2.0, 0.0, NULL, 'P9000001', '2024-03-01'::date, NULL::date, true, 'Web 客户端重构', NULL, NULL, NULL),
            ('P9000004', '产品设计师', 'PD-DESIGN', '产品体验设计师', 'PROF', 'PROF-UX', 'PROF-UX-DES', 'P4-UX', '1000021', 'UI设计组', 'REGULAR', 'ACTIVE', 'FULL_TIME', 1.0, 1.0, NULL, 'P9000005', '2024-03-15'::date, NULL::date, true, 'AI 项目体验升级', '王蕾', '2024-04-05'::date, 'PRIMARY'),
            ('P9000005', '运营主管', 'OPS-SUP', '区域运营主管', 'OPER', 'OPER-SITE', 'OPER-SITE-SUP', 'M1', '1000003', '市场部', 'REGULAR', 'ACTIVE', 'FULL_TIME', 1.0, 1.0, NULL, NULL, '2024-01-10'::date, NULL::date, true, '新运营体系上线', '赵强', '2024-01-25'::date, 'PRIMARY')
        ) AS ps(
            code, title, job_profile_code, job_profile_name,
            job_family_group_code, job_family_code, job_role_code, job_level_code,
            organization_code, organization_name,
            position_type, status, employment_type,
            headcount_capacity, headcount_in_use,
            grade_level, reports_to_position_code,
            effective_date, end_date, is_current,
            operation_reason, current_holder_name, filled_date, current_assignment_type
        )
    ),

    inserted_positions AS (
        INSERT INTO positions (
            tenant_id, code, title,
            job_profile_code, job_profile_name,
            job_family_group_code, job_family_group_name, job_family_group_record_id,
            job_family_code, job_family_name, job_family_record_id,
            job_role_code, job_role_name, job_role_record_id,
            job_level_code, job_level_name, job_level_record_id,
            organization_code, organization_name,
            position_type, status, employment_type,
            headcount_capacity, headcount_in_use,
            grade_level, cost_center_code,
            reports_to_position_code,
            profile,
            effective_date, end_date,
            is_current,
            created_at, updated_at,
            operation_type, operated_by_id, operated_by_name, operation_reason
        )
        SELECT ps.tenant_id,
               ps.code,
               ps.title,
               ps.job_profile_code,
               ps.job_profile_name,
               ps.job_family_group_code,
               g.name AS job_family_group_name,
               g.record_id AS job_family_group_record_id,
               ps.job_family_code,
               f.name AS job_family_name,
               f.record_id AS job_family_record_id,
               ps.job_role_code,
               r.name AS job_role_name,
               r.record_id AS job_role_record_id,
               ps.job_level_code,
               l.name AS job_level_name,
               l.record_id AS job_level_record_id,
               ps.organization_code,
               ps.organization_name,
               ps.position_type,
               ps.status,
               ps.employment_type,
               ps.headcount_capacity,
               ps.headcount_in_use,
               ps.grade_level,
               NULL,
               ps.reports_to_position_code,
               '{}'::jsonb,
               ps.effective_date,
               ps.end_date,
               ps.is_current,
               CURRENT_TIMESTAMP,
               CURRENT_TIMESTAMP,
               'CREATE',
               ps.operator_id,
               ps.operator_name,
               ps.operation_reason
        FROM position_seed ps
        JOIN groups g
          ON g.tenant_id = ps.tenant_id
         AND g.family_group_code = ps.job_family_group_code
        JOIN families f
          ON f.tenant_id = ps.tenant_id
         AND f.family_code = ps.job_family_code
        JOIN roles r
          ON r.tenant_id = ps.tenant_id
         AND r.role_code = ps.job_role_code
        JOIN levels l
          ON l.tenant_id = ps.tenant_id
         AND l.level_code = ps.job_level_code
         AND l.role_code = ps.job_role_code
        ON CONFLICT (tenant_id, code, effective_date) DO NOTHING
        RETURNING tenant_id, code, record_id, status, headcount_in_use
    ),

    selected_positions AS (
        SELECT ip.tenant_id,
               ip.code,
               ip.record_id,
               ps.status,
               ps.headcount_in_use,
               ps.current_holder_name,
               ps.filled_date,
               ps.current_assignment_type
        FROM inserted_positions ip
        JOIN position_seed ps
          ON ps.tenant_id = ip.tenant_id
         AND ps.code = ip.code

        UNION

        SELECT ps.tenant_id,
               ps.code,
               p.record_id,
               ps.status,
               COALESCE(p.headcount_in_use, ps.headcount_in_use),
               ps.current_holder_name,
               ps.filled_date,
               ps.current_assignment_type
        FROM position_seed ps
        JOIN positions p
          ON p.tenant_id = ps.tenant_id
         AND p.code = ps.code
         AND p.effective_date = ps.effective_date
    )

INSERT INTO position_assignments (
    tenant_id, position_code, position_record_id,
    employee_id, employee_name, employee_number,
    assignment_type, assignment_status, fte,
    start_date, end_date, is_current,
    notes, created_at, updated_at
)
SELECT sp.tenant_id,
       sp.code,
       sp.record_id,
       gen_random_uuid(),
       sp.current_holder_name,
       NULL,
       COALESCE(sp.current_assignment_type, 'PRIMARY'),
       'ACTIVE',
       CASE WHEN sp.headcount_in_use > 0 THEN sp.headcount_in_use ELSE 1.0 END,
       COALESCE(sp.filled_date, CURRENT_DATE),
       NULL,
       true,
       '初始职位同步数据',
       CURRENT_TIMESTAMP,
       CURRENT_TIMESTAMP
FROM selected_positions sp
WHERE sp.current_holder_name IS NOT NULL
  AND sp.status <> 'VACANT'
ON CONFLICT (tenant_id, position_code, employee_id, start_date) DO NOTHING;

COMMIT;
