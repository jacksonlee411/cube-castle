-- 81号计划：职位管理租户隔离巡检脚本
--
-- ⚠️ 执行时机：Stage 1 数据库迁移完成后（positions、job_family_groups等表创建后）
--
-- 执行方式：
--   psql -h localhost -U user -d cubecastle \
--     -f docs/development-plans/81-tenant-isolation-checks.sql \
--     > reports/architecture/tenant-isolation-check-YYYYMMDD.sql
--
-- 预期结果：全部查询返回空集（0 rows）
-- 如有返回行，说明存在租户隔离缺陷，立即触发回滚并参考 81 号计划第 6 节处理。
--
-- 注意：
--   - Phase 4（契约阶段）执行会报"relation does not exist"，这是正常的
--   - Stage 1（实现阶段）数据库迁移后必须执行此脚本验证

\echo '1) positions vs job_family_groups tenant mismatch'
SELECT p.code AS position_code, p.tenant_id AS position_tenant, jfg.tenant_id AS catalog_tenant
FROM positions p
JOIN job_family_groups jfg ON p.job_family_group_record_id = jfg.record_id
WHERE p.tenant_id <> jfg.tenant_id;

\echo '2) positions vs job_families tenant mismatch'
SELECT p.code AS position_code, p.tenant_id AS position_tenant, jf.tenant_id AS catalog_tenant
FROM positions p
JOIN job_families jf ON p.job_family_record_id = jf.record_id
WHERE p.tenant_id <> jf.tenant_id;

\echo '3) positions vs job_roles tenant mismatch'
SELECT p.code AS position_code, p.tenant_id AS position_tenant, jr.tenant_id AS catalog_tenant
FROM positions p
JOIN job_roles jr ON p.job_role_record_id = jr.record_id
WHERE p.tenant_id <> jr.tenant_id;

\echo '4) positions vs job_levels tenant mismatch'
SELECT p.code AS position_code, p.tenant_id AS position_tenant, jl.tenant_id AS catalog_tenant
FROM positions p
JOIN job_levels jl ON p.job_level_record_id = jl.record_id
WHERE p.tenant_id <> jl.tenant_id;

\echo '5) job_roles current flag duplicates'
SELECT role_code, COUNT(*) FILTER (WHERE is_current) AS current_count
FROM job_roles
GROUP BY role_code
HAVING COUNT(*) FILTER (WHERE is_current) > 1;

\echo '6) job_levels current flag duplicates'
SELECT level_code, COUNT(*) FILTER (WHERE is_current) AS current_count
FROM job_levels
GROUP BY level_code
HAVING COUNT(*) FILTER (WHERE is_current) > 1;

\echo '7) positions referencing missing job catalog versions'
SELECT p.code AS position_code
FROM positions p
LEFT JOIN job_roles jr ON p.job_role_record_id = jr.record_id
WHERE jr.record_id IS NULL;
