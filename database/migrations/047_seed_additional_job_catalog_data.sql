-- 047_seed_additional_job_catalog_data.sql
-- 为默认租户批量补充数据智能域 Job Catalog 数据，覆盖 10+ 条职类/职种/职务/职级记录

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
            ('DATA', '数据智能序列', '覆盖数据科学与算法工程岗位', '2024-04-01'::date),
            ('AI', '人工智能应用序列', '聚焦 AI 推理与模型服务', '2024-04-01'::date),
            ('STAT', '统计建模序列', '负责预测分析与统计实验', '2024-04-01'::date),
            ('BI', '商业智能序列', '建设指标体系与可视化分析', '2024-04-01'::date),
            ('OPS', '数据运营序列', '保障数据资产治理与运维', '2024-04-01'::date),
            ('RISK', '风控建模序列', '构建信用与反欺诈模型', '2024-04-01'::date),
            ('GOV', '数据治理序列', '覆盖主数据与标准管理', '2024-04-01'::date),
            ('IOT', '物联网数据序列', '处理设备与传感器数据', '2024-04-01'::date),
            ('AUTO', '自动化智能序列', '建设自动化与智能调度平台', '2024-04-01'::date),
            ('CX', '客户洞察序列', '面向客户体验与旅程分析', '2024-04-01'::date)
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
        RETURNING tenant_id, family_group_code, record_id
    ),

    groups AS (
        SELECT tenant_id, family_group_code, record_id
        FROM inserted_groups
        UNION ALL
        SELECT sg.tenant_id, sg.family_group_code, jfg.record_id
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
            ('DATA-CORE', 'DATA', '数据平台职种', '建设与维护数据基础设施', '2024-04-01'::date),
            ('AI-SERVICE', 'AI', 'AI 服务职种', '提供模型推理与 API 服务', '2024-04-01'::date),
            ('STAT-RES', 'STAT', '统计研究职种', '开展实验设计与统计推断', '2024-04-01'::date),
            ('BI-DEV', 'BI', '商业智能开发职种', '构建报表与分析应用', '2024-04-01'::date),
            ('OPS-QUALITY', 'OPS', '数据质量职种', '监控数据链路与校验规则', '2024-04-01'::date),
            ('RISK-CREDIT', 'RISK', '信用风控职种', '搭建信用评分与策略模型', '2024-04-01'::date),
            ('GOV-MDM', 'GOV', '主数据管理职种', '维护主数据模型与字典', '2024-04-01'::date),
            ('IOT-STREAM', 'IOT', '流式数据职种', '处理大规模设备流数据', '2024-04-01'::date),
            ('AUTO-ORCH', 'AUTO', '智能调度职种', '实现跨系统自动化编排', '2024-04-01'::date),
            ('CX-INSIGHT', 'CX', '客户洞察职种', '洞察客户行为与体验指标', '2024-04-01'::date)
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
        RETURNING tenant_id, family_code, record_id
    ),

    families AS (
        SELECT tenant_id, family_code, record_id
        FROM inserted_families
        UNION ALL
        SELECT sf.tenant_id, sf.family_code, jf.record_id
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
            ('DATA-CORE-ENG', 'DATA-CORE', '数据平台工程师', '负责数据仓库与计算引擎', '2024-04-01'::date),
            ('AI-SERVICE-MGR', 'AI-SERVICE', 'AI 服务经理', '协调模型服务上线与 SLA', '2024-04-01'::date),
            ('STAT-RES-SCI', 'STAT-RES', '统计科学家', '提供统计建模方法支持', '2024-04-01'::date),
            ('BI-DEV-ARCH', 'BI-DEV', 'BI 架构师', '设计商业智能平台架构', '2024-04-01'::date),
            ('OPS-QUALITY-LEAD', 'OPS-QUALITY', '数据质量负责人', '制定数据质量策略与指标', '2024-04-01'::date),
            ('RISK-CREDIT-MODEL', 'RISK-CREDIT', '信用建模专家', '搭建信用评分和策略模型', '2024-04-01'::date),
            ('GOV-MDM-OWNER', 'GOV-MDM', '主数据负责人', '负责主数据模型与治理流程', '2024-04-01'::date),
            ('IOT-STREAM-ENGINEER', 'IOT-STREAM', '流式数据工程师', '构建实时流处理管道', '2024-04-01'::date),
            ('AUTO-ORCH-ENGINEER', 'AUTO-ORCH', '自动化编排工程师', '实现跨系统自动化流程', '2024-04-01'::date),
            ('CX-INSIGHT-LEAD', 'CX-INSIGHT', '客户洞察负责人', '统筹客户体验数据洞察', '2024-04-01'::date)
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
        RETURNING tenant_id, role_code, record_id
    ),

    roles AS (
        SELECT tenant_id, role_code, record_id
        FROM inserted_roles
        UNION ALL
        SELECT sr.tenant_id, sr.role_code, jr.record_id
        FROM seed_roles sr
        JOIN job_roles jr
          ON jr.tenant_id = sr.tenant_id
         AND jr.role_code = sr.role_code
         AND jr.effective_date = sr.effective_date
    )

INSERT INTO job_levels (
    tenant_id, level_code, role_code, parent_record_id,
    level_rank, name, description, status,
    effective_date, is_current, created_at, updated_at
)
SELECT r.tenant_id,
       l.level_code,
       l.role_code,
       r.record_id,
       l.level_rank,
       l.name,
       l.description,
       'ACTIVE',
       l.effective_date,
       true,
       CURRENT_TIMESTAMP,
       CURRENT_TIMESTAMP
FROM roles r
JOIN (VALUES
    ('IC5-DATA', 'DATA-CORE-ENG', 'SENIOR', 'IC5 高级数据平台层级', '负责核心组件性能与扩展', '2024-04-01'::date),
    ('M2-AI', 'AI-SERVICE-MGR', 'MANAGER', 'M2 AI 服务管理层级', '统筹服务发布与 SLA 管控', '2024-04-01'::date),
    ('IC6-STAT', 'STAT-RES-SCI', 'EXPERT', 'IC6 统计科学层级', '提供高阶统计方法与指导', '2024-04-01'::date),
    ('IC5-BI', 'BI-DEV-ARCH', 'SENIOR', 'IC5 商业智能层级', '规划全局 BI 架构与标准', '2024-04-01'::date),
    ('M1-OPS', 'OPS-QUALITY-LEAD', 'MANAGER', 'M1 数据质量管理层级', '推进质量制度与巡检流程', '2024-04-01'::date),
    ('IC5-RISK', 'RISK-CREDIT-MODEL', 'SENIOR', 'IC5 风控建模层级', '构建核心信用策略模型', '2024-04-01'::date),
    ('M1-GOV', 'GOV-MDM-OWNER', 'MANAGER', 'M1 主数据治理层级', '负责企业级数据标准落地', '2024-04-01'::date),
    ('IC5-IOT', 'IOT-STREAM-ENGINEER', 'SENIOR', 'IC5 流式数据层级', '构建实时处理与监控平台', '2024-04-01'::date),
    ('IC5-AUTO', 'AUTO-ORCH-ENGINEER', 'SENIOR', 'IC5 自动化编排层级', '设计自动化流程与规则引擎', '2024-04-01'::date),
    ('M2-CX', 'CX-INSIGHT-LEAD', 'MANAGER', 'M2 客户洞察管理层级', '统筹体验指标与洞察交付', '2024-04-01'::date)
) AS l(level_code, role_code, level_rank, name, description, effective_date)
  ON l.role_code = r.role_code
ON CONFLICT (tenant_id, level_code, effective_date) DO NOTHING;

COMMIT;
