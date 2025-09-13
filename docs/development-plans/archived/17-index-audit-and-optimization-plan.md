# 17 - 组织管理数据库索引审计与优化计划

文档编号: 17  
文档类型: 调查/优化计划  
创建日期: 2025-09-07  
范围: organization_units（组织管理核心表）及相关查询  
优先级: 中高（可在不影响功能的前提下减少维护与写入成本）

---

## 背景与目标

当前 Schema 初始脚本与后续迁移叠加创建了两类索引（“通用/广撒网”与“时态/当前专用”），导致多处重叠。目标是在不牺牲核心查询性能的前提下，精简冗余索引，聚焦核心访问路径，降低写入开销与维护复杂度。

---

## 现状盘点（来源：sql/init/01-schema.sql 与 migrations/025）

主要索引（节选）：
- 通用/广撒网类
  - idx_org_units_tenant_code (tenant_id, code)
  - idx_org_units_parent_code (parent_code) WHERE parent_code IS NOT NULL
  - idx_org_units_status_current (status, is_current) WHERE NOT is_deleted
  - idx_org_units_level_path (level, code_path)
  - idx_org_units_effective_date (effective_date, end_date)
  - idx_org_units_is_current (is_current) WHERE is_current = true
  - idx_org_units_unit_type (unit_type, status)
  - idx_org_units_created_at / idx_org_units_updated_at
  - idx_org_units_operation_type (operation_type, updated_at)
  - GIN/全文: idx_org_units_profile_gin, idx_org_units_name_text, idx_org_units_description_text
  - 监控: idx_org_units_monitoring (tenant_id, created_at, operation_type)

- 时态/当前专用类（migrations/025）
  - uk_org_temporal_point（唯一）: (tenant_id, code, effective_date) WHERE 非删除
  - uk_org_current（唯一）: (tenant_id, code) WHERE is_current=true 且非删除
  - ix_org_temporal_query: (tenant_id, code, effective_date DESC) WHERE 非删除
  - ix_org_adjacent_versions: (tenant_id, code, effective_date, record_id) WHERE 非删除
  - ix_org_current_lookup: (tenant_id, code, is_current) WHERE is_current=true 且非删除
  - ix_org_temporal_boundaries: (code, effective_date, end_date, is_current) WHERE 非删除
  - ix_org_daily_transition: (effective_date, end_date, is_current) WHERE 非删除

结论：同一访问路径存在多组功能重叠索引，且部分单列/通用索引未见代码侧使用。

---

## 访问模式与关键查询（代码参考）

- 相邻/当前版本查询（事务内 FOR UPDATE）：`internal/services/temporal.go`、`internal/repository/temporal_timeline.go`
  - WHERE tenant_id, code AND is_current=true
  - WHERE tenant_id, code AND effective_date [< 或 >] x ORDER BY effective_date [DESC/ASC] LIMIT 1

- 层级查询（当前态子节点/树）：`internal/repository/hierarchy.go`
  - WHERE tenant_id=$2 AND parent_code=$1 AND is_current=true ORDER BY sort_order, code

- 日切/批处理：`scripts/daily-cutover.sql`
  - 基于 effective_date/end_date/is_current 的批量校正（UTC）

---

## 冗余与重叠分析（删除候选）

- 与“当前/时态”专用索引重叠或方向不当：
  - idx_org_units_is_current（仅 is_current）
  - idx_org_units_effective_date（effective_date, end_date）
  - idx_org_units_temporal（code, effective_date, end_date, is_current）
  - idx_org_units_status_current（status, is_current）

- 低价值/未见使用（结合当前代码）：
  - idx_org_units_unit_type、idx_org_units_created_at、idx_org_units_updated_at、idx_org_units_operation_type
  - idx_org_units_level_path（层级主要按 parent_code + tenant_id + is_current 查询）

- 组合与单列重复：
  - idx_org_units_tenant_code 与专用索引存在覆盖关系；如保留后者，可评估移除前者。
  - idx_org_units_monitoring 与 created_at/operation_type 单列索引存在重叠（择一保留）。

- 文本/JSON 索引：
  - idx_org_units_profile_gin、idx_org_units_name_text/description_text：若无搜索功能调用，建议移除；如前端/报表依赖，应保留并记录使用方。

---

## 建议的“最小必要集合”（第一版）

- 唯一约束（必保留）
  - uk_org_temporal_point（tenant_id, code, effective_date）WHERE 非删除
  - uk_org_current（tenant_id, code）WHERE is_current=true 且非删除

- 时态查询
  - ix_org_temporal_query（tenant_id, code, effective_date DESC）WHERE 非删除
  - ix_org_adjacent_versions（tenant_id, code, effective_date, record_id）WHERE 非删除
  - ix_org_temporal_boundaries（code, effective_date, end_date, is_current）WHERE 非删除（与 idx_org_units_temporal 二选一，建议保留本索引）

- 当前查找
  - ix_org_current_lookup（tenant_id, code, is_current）WHERE is_current=true 且非删除

- 子节点（建议新增）
  - (tenant_id, parent_code, sort_order, code) WHERE is_current=true AND status<>'DELETED' AND deleted_at IS NULL

- 日切
  - ix_org_daily_transition（effective_date, end_date, is_current）WHERE 非删除

- 监控（二选一）
  - 保留 idx_org_units_monitoring（组合索引），删除 created_at/operation_type 单列索引；或反之。

---

## 建议删除清单（按优先级）

1) 明显冗余（优先删除）：
   - idx_org_units_is_current
   - idx_org_units_effective_date
   - idx_org_units_temporal（保留 ix_org_temporal_boundaries 即可）
   - idx_org_units_status_current

2) 低价值/未见使用：
   - idx_org_units_unit_type
   - idx_org_units_created_at、idx_org_units_updated_at
   - idx_org_units_operation_type
   - idx_org_units_level_path

3) 组合/单列重叠（择优保留）：
   - idx_org_units_tenant_code（如专用索引覆盖则移除）
   - idx_org_units_monitoring 与 created_at/operation_type 单列索引（二选一）

4) 文本/JSON（如无需求则移除）：
   - idx_org_units_profile_gin、idx_org_units_name_text、idx_org_units_description_text

---

## 验证与落地流程（灰度）

1) 线上热度采样（7–14天）
```sql
SELECT indexrelname AS index, idx_scan, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes
WHERE relname='organization_units'
ORDER BY idx_scan DESC;
```

2) 压测关键路径
- 插入/删除/改生效日/停用/激活（验证相邻版本查询与事务锁）
- 层级查询（直接子节点、有序遍历）
- 日切脚本（UTC）

3) 灰度删除顺序
- 先删除“明显冗余单列/通用索引”→ 观察慢查询与执行计划（EXPLAIN ANALYZE）→ 再处理组合/覆盖冲突

4) 新增子节点部分索引
```sql
CREATE INDEX CONCURRENTLY IF NOT EXISTS ix_org_children_lookup
ON organization_units(tenant_id, parent_code, sort_order, code)
WHERE is_current = true AND status <> 'DELETED' AND deleted_at IS NULL;
```

5) 观测与回滚
- 观察 pg_stat_statements、慢查询日志、应用延迟指标；如出现回归，立即回滚 DROP 操作（保留回滚脚本）。

---

## 风险与缓解

- 写入回归风险：逐步删除 + CONCURRENTLY 创建；保留回滚脚本。
- 查询回归风险：以真实执行计划为准，必要时调整字段顺序/WHERE 条件匹配。
- 需求变更：文本/JSON 索引删除前确认无下游依赖。

---

## 后续行动（可由我来提交迁移草案）

- 生成两阶段迁移：
  1) DROP 冗余索引（优先集）+ CREATE 子节点部分索引
  2) 根据热度采样与压测，进一步精简（组合/覆盖冲突组）

- 提供回滚脚本与观测手册（含常用 SQL 与 EXPLAIN 分析模版）。

---

附：常用诊断 SQL
```sql
-- 列出 organization_units 的索引
SELECT indexname, indexdef FROM pg_indexes WHERE tablename='organization_units';

-- 索引热度
SELECT indexrelname AS index, idx_scan, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes
WHERE relname='organization_units'
ORDER BY idx_scan DESC;

-- 示例执行计划
EXPLAIN (ANALYZE, BUFFERS)
SELECT status FROM organization_units
WHERE tenant_id=$1 AND code=$2 AND is_current=true;
```

