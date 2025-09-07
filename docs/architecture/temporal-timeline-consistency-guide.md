# 组织架构时态时间轴连贯性实施参考 v1.0

目的
- 给出“写入负荷小、单条当前态查询高频”的简化落地方案，确保在插入/删除中间版本、修改生效日期、停用与启用等操作下，时间轴（effective_date/end_date）保持连续、可审计；同时保证读路径极简、低延迟、易维护。

前提与范围
- 单表实现，不新增物理表。
- 不使用数据库触发器、EXCLUDE 排它约束、热路径 advisory lock；以应用层事务与最小约束达成一致性。

API 优先对齐（按最新 API 文档）
- Create 组织（POST /api/v1/organization-units）
  - 仅用于新 code 的首条记录；不做跨版本回填/重算。
- 版本操作（POST /api/v1/organization-units/{code}/versions）
  - INSERT：prev.end_date = new.effective_date - 1；new.end_date = next.effective_date - 1（若存在）。
  - UPDATE effectiveDate：按“删旧+插新”原子重算相邻边界，保持连续。
  - DELETE one version：桥接相邻；若为尾部，prev.end_date = NULL。
  - 仅非删除版本参与时间线；按 effective_date 升序，禁止重叠与断档。
- 事件端点（POST /api/v1/organization-units/{code}/events）
  - DEACTIVATE：按 recordId 作废该版本（status=DELETED），同一事务执行“全链重算”。
  - 响应返回最新“非删除时间线”（timeline），供前端即时刷新，避免读缓存延迟。
- 更新 `/api/v1/organization-units/{code}`
  - 仅支持 `PUT`；`PATCH` 已弃用，不处理 effectiveDate/endDate 与状态流转。
- DELETE /{code}
  - 整单位软删除（所有版本 DELETED），不做时间线回填。

背景与模型
- 一维业务状态：business_status/status ∈ {ACTIVE, INACTIVE}（INACTIVE 即停用）。停用/启用语义通过 operation_type=SUSPEND/REACTIVATE 表达，而非引入第三状态。
- 有效期表达时态：effective_date（生效，date），end_date（结束，date，可空）。按“日”的闭区间语义管理，上一版本在应用事务中回填为 new.effective_date - 1 天。
- 当前态标记：写侧物化 is_current（boolean），由应用层在事务里维护与翻转；API/GraphQL 可暴露 isCurrent/isFuture 计算字段，读取 is_current 或基于 asOfDate 计算均可。
- 删除标记：is_deleted + deleted_at（独立于业务状态）。

系统不变量与约束（最小必要集）
- 当前唯一（Single Current per Code per Tenant）
  - 同一租户 tenant、同一 code 任一时刻最多一条 is_current=true 记录。
  - 数据库约束：部分唯一索引（推荐 uk_org_current）
    - ON organization_units(tenant_id, code) WHERE is_current = true
- 时点唯一（Temporal Point Uniqueness）
  - UNIQUE (tenant_id, code, effective_date)
- 自动衔接（Auto Back-fill End Date）
  - 新版本插入时，在同一事务内由应用层把上一版本 end_date 设为 new.effective_date - 1，避免重叠/断档。
- 区间不重叠（No Temporal Overlap）
  - 不启用 EXCLUDE 约束。应用层在写入事务中做相邻版本预检（基于前一/后一版本并使用 FOR UPDATE 锁），避免区间重叠。
  - 运行侧提供离线巡检 SQL 以报警与修复（非热路径）。
- 端点唯一（状态变更的唯一命令）
  - 停用：POST /api/v1/organization-units/{code}/suspend → 强制 INACTIVE + operation_type=SUSPEND
  - 启用：POST /api/v1/organization-units/{code}/activate → 强制 ACTIVE + operation_type=REACTIVATE

实现参考（四类关键场景，单表/无触发器）

1) 插入“中间版本”（在历史与当前之间补版本）
- 目标：在 t1 与 t2 之间插入新版本 tX，保证无重叠、无断档。
- 操作步骤（单事务）：
  1. 读取相邻版本：按 `(tenant_id, code, effective_date)` 查前一/后一版本，并 `FOR UPDATE` 锁定相邻行。
  2. 预检：
     - 不存在相同 (tenant_id, code, effective_date=tX)
     - 与后一版本不重叠（若存在 next，则要求 tX <= next.effective_date）
  3. 回填边界：若存在上一版本 prev，设置 `prev.end_date = tX - 1`。
  4. 插入新版本（effective_date=tX）。若 tX 为“即时当前”且 tX=今天（或业务定义的当前），将 `prev.is_current=false`，新版本 `is_current=true`。
  5. 提交事务；由部分唯一索引兜底“单当前”。
  6. 可选：新增前先调用验证端点（docs/api/openapi.yaml:694-717）或后端校验服务。

端点对应：POST /api/v1/organization-units/{code}/versions（INSERT）。

2) 删除“中间版本”（历史数据修复）
- 首选策略：避免硬删除历史版本，采用“更正版本”覆盖（插入修正版本并回填边界），保留审计链。
- 如必须删除（数据修复场景），单事务执行：
  1. 读取相邻版本（prev, next）并 `FOR UPDATE` 锁定。
  2. DELETE tX。
  3. 将 `prev.end_date = next.effective_date - 1`（若 next 存在），桥接区间。
  4. 写入时间线事件/审计。
  5. 若删除的是尾部：`prev.end_date = NULL`，并基于 `prev.effective_date ≤ 今天` 重算 is_current。

端点对应：
- 推荐：POST /api/v1/organization-units/{code}/events（DEACTIVATE，返回最新 timeline）。
- 管理：POST /api/v1/organization-units/{code}/versions（DELETE one version）。
- 风险控制：
  - 核对“单当前唯一”“时点唯一”“区间不重叠”。
  - 批量修复：建议串行处理同一 (tenant_id, code) 的变更，或外层应用级互斥；不将 advisory lock 引入热路径。

3) 修改记录的生效日期（effective_date 变更）
- 推荐语义：等价于“删除旧版本 + 插入新版本”（单事务原子化）。
- 步骤：
  1. 预检：目标 effective_date 与邻接区间不冲突（不重叠，不相同时间点）。
  2. DELETE 旧版本（或标记作废）。
  3. INSERT 新版本（使用新 effective_date），并回填相邻边界、维护 is_current。
  4. 写入时间线事件（operation_type=UPDATE）。

端点对应：
- 推荐：POST /api/v1/organization-units/{code}/versions（UPDATE effectiveDate）。
- 历史修正（非日期字段）：PUT /api/v1/organization-units/{code}/history/{record_id}。

4) 停用与重新启用（即时/计划）
- 停用：/suspend → 强制 business_status=INACTIVE，写入 SUSPEND 版本。
- 启用：/activate → 强制 business_status=ACTIVE，写入 REACTIVATE 版本。
- 计划操作：effective_date>today → 生成未来版本（isFuture=true），不影响当前；asOfDate 到达时成为当前。
- 幂等：重复启/停同一目标状态 → 200 OK，无新版本（命令端做幂等保护）。
  - 建议幂等键： (tenant_id, code, operation_type, effective_date)

预检与一致性校验（建议最小集）
- 入参校验：date 合法性、code 规范、租户隔离头（X-Tenant-ID）存在。
- 冲突校验：
  - UNIQUE (tenant_id, code, effective_date) 兜底重复时点。
  - 相邻版本无区间重叠（应用层预检）。
- 当前唯一：若为“即时变更”，确保事务后仅一条 `is_current=true`（由部分唯一索引兜底）。
- 读侧路径：
  - 当前态高频查询：`WHERE tenant_id=$1 AND code=$2 AND is_current=true`。
  - asOfDate 低频查询：`effective_date <= :asOf AND (end_date IS NULL OR end_date > :asOf)`。

全链重算（统一算法）
- 触发：版本删除（DELETE one version / DEACTIVATE）后必须执行；也可用于批量修复。
- 输入：同一 (tenant_id, code) 的“非删除版本”，按 effective_date 升序。
- 步骤：
  1. 对序列中每条 i：`end_date[i] = next.effective_date - 1`（若存在），最后一条 `end_date=NULL`。
  2. 清空该 code 所有 `is_current` 标记（不区分删除与否，避免残留冲突）。
  3. 若存在 `effective_date ≤ 今天` 的版本，选择其中生效日最大的一条设为 `is_current=true`；否则无当前态。
- 输出：无断档、无重叠、尾部开放、单当前。

错误模型与返回
- 冲突：`TEMPORAL_POINT_CONFLICT`（时点唯一）、`TEMPORAL_OVERLAP_CONFLICT`（区间重叠）。
- 事件响应：`data.timeline[]` 返回最新“非删除时间线”（recordId, code, effectiveDate, endDate, isCurrent, status …）。

运行与维护
- 约束与索引（PostgreSQL 示例）：
  - `CREATE UNIQUE INDEX uk_org_ver ON organization_units(tenant_id, code, effective_date);`
  - `CREATE UNIQUE INDEX uk_org_current ON organization_units(tenant_id, code) WHERE is_current;`
  - `CREATE INDEX ix_org_tce ON organization_units(tenant_id, code, effective_date DESC);`
- 日切任务（推荐开启）：
  - 运行时机：每日 00:05。
  - 作用：将“昨天结束/今天生效”的极少量行翻转 is_current（原当前→false，新当前→true）。
  - 特性：仅影响小集合，可重试，失败不影响历史可回放。
- 巡检任务（离线）：
  - 周期性检查同一 (tenant_id, code) 是否存在区间重叠或断档；发现异常报警并生成修复建议。
- 并发控制：
  - 写少场景下，应用层对相邻版本 `SELECT ... FOR UPDATE` 即可；不引入 advisory lock 到热路径。

性能建议
- 读路径优先：
  - 当前态点查命中部分唯一索引，计划稳定、延迟低。
  - 如有热点，可加应用缓存（键：tenant_id+code），写时主动失效，TTL 兜底。
- asOf 查询：
  - 保持简单过滤；仅在离线/报表类任务中使用，不在在线热路径引入 GiST/EXCLUDE。

结论与演进
- 在“写入少、当前态读多”的前提下，单表 + 应用事务回填 + 部分唯一索引 + 日切的小而稳组合即可满足正确性与性能要求。
- 不引入触发器、EXCLUDE、advisory lock 至热路径，降低复杂度与维护成本。
- 若未来出现事故频发或规模化性能瓶颈，可在不改变读路径的前提下，按需评估引入更强约束（例如只在导入/修复窗口临时开启更严格校验）。
