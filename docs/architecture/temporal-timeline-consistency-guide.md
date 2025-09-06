# 组织架构时态时间轴连贯性实施参考 v1.0

目的
- 形成一份面向实施的参考指南，确保在插入/删除中间版本、修改生效日期、停用与启用等操作下，组织时间轴（effectiveDate/endDate）保持连续、无重叠、可审计、可回放。

背景与模型
- 一维业务状态：businessStatus/status ∈ {ACTIVE, INACTIVE}（INACTIVE 即停用）
- 有效期表达时态：effectiveDate（生效），endDate（结束，可空）
- 计算字段：isCurrent/isFuture（基于 asOfDate，查询态，不持久化）
- 删除标记：isDeleted + deletedAt（独立于状态）

系统不变量与约束
- 当前唯一（Single Current per Code）
  - 每个 code 任意时间点最多一条当前记录：唯一索引 uk_current_organization（参考 scripts/data-cleanup-and-test-creation.sql:38-41）
- 时点唯一（Temporal Point Uniqueness）
  - (code, effective_date, tenant_id) 在“正常数据”上唯一（scripts/data-cleanup-and-test-creation.sql:41-43）
- 自动衔接（Auto Back-fill End Date）
  - 新版本插入时，触发器自动将上一版本 end_date 设为新版本 effective_date - 1，避免重叠/断档（行为见 scripts/test-five-state-api.sh:141-168）
- 端点唯一（状态变更的唯一命令）
  - 停用：POST /api/v1/organization-units/{code}/suspend → 强制 INACTIVE + operationType=SUSPEND
  - 启用：POST /api/v1/organization-units/{code}/activate → 强制 ACTIVE + operationType=REACTIVATE

实现参考（四类关键场景）

1) 插入“中间版本”（在历史与当前之间补版本）
- 目标：在 t1 与 t2 之间插入新版本 tX，保证无重叠、无断档
- 操作步骤（事务内）：
  1. 预检冲突：
     - 确保不存在相同 (code, effectiveDate)
     - 与前后版本区间不重叠（如 tPrev.effectiveDate ≤ tX < tNext.effectiveDate）
  2. 插入新版本（effectiveDate=tX）；触发器自动回填 tPrev.endDate=tX-1
  3. 重新计算 is_current（如有）或在读侧通过 asOfDate 动态计算
- 推荐：新增前先调用验证端点（docs/api/openapi.yaml:694-717）或后端校验服务

2) 删除“中间版本”（历史数据修复）
- 首选策略：避免硬删除历史版本，采用“更正版本”覆盖：
  - 插入一条更正的中间版本，让触发器重算相邻版本边界，保留审计链
- 如必须删除（数据修复场景），事务内执行：
  1. 读取相邻版本（tPrev, tNext）并加行级锁
  2. 删除 tX
  3. 将 tPrev.endDate 设为 tNext.effectiveDate - 1，桥接区间
  4. 写入时间线事件/审计
- 风险控制：
  - 检查“单当前唯一”与“时点唯一”是否仍然满足
  - 大批量修复时，可按 scripts/data-cleanup-and-test-creation.sql 的顺序临时禁用触发器/索引→导入→重建→一致性检查

3) 修改记录的生效日期（effectiveDate 变更）
- 推荐语义：等价于“删除旧版本 + 插入新版本”（单事务原子化），由触发器自动重算边界
- 步骤：
  1. 预检：目标 effectiveDate 与邻接区间不冲突（不重叠，不相同时间点）
  2. DELETE 旧版本（或标记作废）
  3. INSERT 新版本（使用新 effectiveDate）
  4. 触发器自动设置前一版本 endDate
  5. 写入时间线事件（operationType=UPDATE）

4) 停用与重新启用（即时/计划）
- 停用：/suspend → 强制 businessStatus=INACTIVE，自动写入 SUSPEND 版本
- 启用：/activate → 强制 businessStatus=ACTIVE，自动写入 REACTIVATE 版本
- 计划操作：effectiveDate>today → 生成未来版本（isFuture=true），不影响当前；asOfDate 到达时成为当前
- 幂等：重复启/停同一目标状态 → 200 OK，无新版本（建议在命令端做幂等保护）

预检与一致性校验（建议最小集）
- 入参校验：date 合法性、code 七位数、租户隔离头（X-Tenant-ID）存在
- 冲突校验：
  - (code, effectiveDate) 唯一
  - 与邻接版本无区间重叠
- 当前唯一：若为“即时变更”，确保结果仍仅有一条 isCurrent=true
- 业务规则：启/停用禁止对 isDeleted=true 数据；父子关系约束（如需要）

失败与回滚处理
- 事务边界：所有写操作（DELETE/INSERT/UPDATE）在单事务内完成
- 错误码映射：
  - 409 Conflict：时点冲突/区间重叠/不可操作状态
  - 422 Unprocessable Entity：无效日期/不满足业务前提
  - 403 Forbidden：权限不足（org:activate/org:suspend）

监控与审计
- 时间线事件：
  - 版本化触发记录 timeline 事件（database/migrations/008_temporal_management_schema.sql:267-323, 291-307）
  - operationType ∈ {CREATE, UPDATE, SUSPEND, REACTIVATE, DELETE}
- 告警：
  - 410 Gone（若访问已废弃端点 /reactivate）次数>0 告警
  - 审计事件 DEPRECATED_ENDPOINT_USED 触发告警（已在 ADR-008 给出中间件伪代码）

测试清单（最小 E2E 套件）
- 自动结束日期：创建 t1，再创建 t2（t2>t1），确认 t1.endDate=t2-1（scripts/test-five-state-api.sh:141-168）
- 插入中间版本：在 t1<tX<t2 插入新版本，确认 t1.endDate=tX-1 且无重叠
- 删除中间版本：删除 tX 后桥接 t1.endDate=t2-1
- 修改生效日：变更 effectiveDate，确认边界重算正确
- 停用/启用：立即与计划（future-effective），幂等重复操作
- 权限：org:activate/org:suspend 缺失→403；有权→200

落地建议（实现顺序）
1) 命令端：封装三类存储过程/服务层原子操作
   - insert_intermediate_version(code, effectiveDate, payload)
   - delete_intermediate_version(code, effectiveDate)
   - change_effective_date(code, oldEffectiveDate, newEffectiveDate)
   输出统一信封响应与审计事件。
2) 端点：严格使用 /suspend 与 /activate；禁止通过 PATCH 修改 status
3) 验证端点：在写前调用，统一返回冲突与建议（docs/api/openapi.yaml:694-717）
4) 监控审计：激活 410/DEPRECATED_ENDPOINT_USED 告警，观测迁移/异常调用

参考文件
- docs/api/openapi.yaml:633（activate 端点定义与权限）
- scripts/test-five-state-api.sh:141（自动 end_date 行为校验）
- scripts/data-cleanup-and-test-creation.sql:38（唯一约束/触发器管理）
- database/migrations/008_temporal_management_schema.sql:267（版本化触发/时间线）
- docs/api/schema.graphql:162（status 一维 + isCurrent/isFuture 注释）

