# 06 — 跨团队一致性与代码异味整治进展日志（不含敏感凭据项）

最后更新：2025-09-13  
维护团队：架构组（主责）+ 后端组 + 前端组 + QA组  
文档状态：进展记录 + 待办拆解（Development Plans）

—

## 1. 范围声明
- 本次记录聚焦于一致性、健壮性与实现缺陷问题；按用户要求，不纳入“明文凭据与密钥”相关事项。
- 参考与边界：契约唯一来源 `docs/api/`；本文件属于 Development Plans（计划/进展日志），不写入 Reference。

—

## 2. 发现摘要（按优先级）

- P0 功能性缺陷：L2 缓存命中失效（类型丢失）
  - 位置：`internal/cache/unified_cache_manager.go`
  - 现象：`CacheEntry.Data` 使用 `interface{}` 存储。L2(JSON) 反序列化后为通用 `[]interface{}`/`map[string]interface{}`，后续直接断言为 `[]Organization` 或 `*OrganizationStats` 失败，导致 L2 永不命中（只有 L1 有效）。
  - 影响：缓存命中率与性能下降；一致性智能更新策略价值受损。
  - 建议：为缓存数据引入可还原类型的序列化协议（详见第3节）。

- P1 API 一致性：对外 JSON 命名存在 snake_case 风险
  - 位置：`pkg/health/reporter.go` 多处 JSON tag 使用 `snake_case`（如 `response_time`、`last_checked`）。
  - 规范：API 一致性规范要求响应字段一律 camelCase。
  - 风险：若健康仪表板/状态端点外露，将触发规范冲突和前端类型不一致。
  - 建议：统一改为 camelCase（例如 `responseTime`、`lastChecked`）；若仅内部使用也应加注释声明“内部结构，非 API 响应”。

- P1 历史依赖残留的治理（与当前架构相悖）
  - 位置：`scripts/verify-cqrs-data-consistency.py` 等仍引用 Neo4j/CDC 逻辑（已在脚本内写明历史参考）。
  - 规范：PostgreSQL 原生 CQRS，禁止引入 Neo4j/Kafka CDC 同步。
  - 建议：将此类脚本迁移至 `docs/archive/`，并在 `docs/archive/README.md` 标注成因与废弃说明；CI 忽略其执行，仅保留历史学习价值。

- P2 健壮性与可维护性
  - 手工大小写匹配实现：`internal/middleware/graphql_envelope.go` 自实现大小写与子串搜索（`containsIgnoreCase` 等），建议改用标准库 `strings`，降低维护风险。
  - 错误响应去敏：认证/权限中间件在响应中拼接 `err.Error()`（如 GraphQL 权限中间件），建议仅返回企业错误码与通用描述，细节写日志（生产最小暴露）。
  - 前端日志与请求规范：多处 `console.log`（temporal 组件、hooks 与配置）；建议统一日志封装（按环境级别输出），并确保业务请求统一经 `frontend/src/shared/api/unified-client.ts`。
  - 二进制产物与工作区清洁度：大型二进制存在于仓库工作区（已被 `.gitignore` 覆盖）。建议CI增加“工作区清洁度”检查，避免误提交。

—

## 3. 修复建议（方案与权衡）

- L2 缓存类型还原（推荐任一方案）
  - 方案A（类型标注）：为 `CacheEntry` 增加 `kind` 字段（如 `organizations|organization|stats`），并在 `MarshalJSON/UnmarshalJSON` 中按 `kind` 执行强类型还原。
  - 方案B（分结构）：按用途拆分 `ListEntry/StatsEntry/EntityEntry`，避免 `interface{}`；键路由到确切结构，序列化/反序列化零歧义。
  - 方案C（延迟解码）：`CacheEntry.Data` 存为 `json.RawMessage`；读取端按调用场景决定解码目标类型（最少侵入）。
  - 选型建议：B 最清晰（类型安全），C 侵入小（迁移快）。

- 健康仪表板 JSON 命名统一
  - 将 `pkg/health/reporter.go` 的对外 JSON tag 全量改为 camelCase；若存在仅内部用途的结构，补充注释声明“非 API 外露”。

- 历史 Neo4j/CDC 脚本治理
  - 文件迁移至 `docs/archive/`；在 `docs/README.md` 与 `docs/archive/README.md` 增补迁移说明与学习价值保留原因；CI 对该目录不做执行校验。

- 中间件与错误响应
  - `graphql_envelope.go` 使用 `strings.Contains(strings.ToLower(s), strings.ToLower(substr))` 等替换自实现；
  - 错误响应保持统一信封与错误码，`details` 控制在开发态；生产下隐藏内部栈与实现细节。

- 前端治理
  - 建立日志工具（最低限 `info/debug/error` 封装），替换分散 `console.log`；
  - 审核业务层请求路径，确保统一使用 `unified-client`；对遗留直连 `fetch` 标注 `// TODO-TEMPORARY:` 并限期收敛。

—

## 4. 任务拆解与负责人（Agent）

- 后端（backend-agent）
  - [P0] 修复 L2 缓存类型还原（选 B 或 C）
  - [P1] 统一健康仪表板 JSON 命名为 camelCase
  - [P2] 替换 `graphql_envelope.go` 的手工大小写匹配为标准库
  - [P2] 错误响应去敏（认证/权限中间件）

- 架构（architecture-agent）
  - [P1] 迁移 Neo4j/CDC 历史脚本至 `docs/archive/` 并补充归档说明
  - [P1] 在 Reference 中补一条注记：组织单元路径参数统一 `{code}`（文档表格展示其他域 `{id}` 示例不代表本域）

- 前端（frontend-agent）
  - [P2] 导入统一日志工具并替换散落 `console.log`
  - [P2] 自查直连 `fetch`，统一走 `unified-client`（保留必要例外并注记）

- QA（qa-agent）
  - [P1] 添加契约回归：健康端点字段命名 camelCase 校验
  - [P2] 缓存命中路径的契约/集成测试（L1/L2 命中与回填行为）

—

## 5. 里程碑与截止（一个迭代内）

- 2025-09-16（D1）：
  - 缓存修复方案评审与选型（B/C）
  - 历史脚本迁移至 `docs/archive/`（文档说明就位）

- 2025-09-18（D3）：
  - 实现并合入 L2 缓存修复；补最小集成测试
  - 健康仪表板 JSON 命名统一；新增 QA 校验

- 2025-09-20（D5，迭代收口）：
  - 中间件去除手工大小写匹配；错误响应去敏
  - 前端日志与请求治理完成 80%+ 收敛，遗留处标注 `// TODO-TEMPORARY:` 并在下一迭代清理

—

## 6. 风险与回退
- 若 L2 修复上线后出现兼容性问题，可快速切换为“仅 L1 有效 + L3 直读”的降级策略，并保留事件驱动的失效机制；
- 健康仪表板字段改名可能影响监控对接，需提前同步仪表板使用方并提供字段映射说明。

—

## 7. 影响文件（摘录）
- `internal/cache/unified_cache_manager.go`（缓存存取/反序列化）
- `internal/cache/cache_events.go`（模型命名与事件映射，仅作为参照）
- `pkg/health/reporter.go`（JSON 输出字段）
- `internal/middleware/graphql_envelope.go`（字符串处理与错误包装）
- `scripts/verify-cqrs-data-consistency.py`（迁移归档）
- `frontend/src/**`（日志与请求统一，按自查范围增量改造）

—

## 8. 验收标准（Definition of Done）
- 缓存：L2 命中路径可被复现（列表/单体/统计）；回填 L1 生效；命中率提升可观察；
- API 命名：对外端点无 snake_case 字段；前端类型与后端输出一致；
- 历史脚本：已归档且 CI 不再尝试执行；
- 中间件：不再包含自实现大小写匹配；错误响应对外去敏，日志保留细节；
- 前端：新增日志封装与 80%+ 替换覆盖；请求统一通过 `unified-client`（例外有注记与计划）。

—

（注：本日志为 Development Plans 用于跨团队协作与阶段推进；完成项将归档至 `docs/archive/development-plans/`，规范性参考内容请见 `docs/reference/`。）

—

## 9. 相关规范与参考
- 项目原则与单一事实来源索引：`../../CLAUDE.md`
- 代理/实现强制规范：`../../AGENTS.md`
- API 契约（唯一事实来源）：`../api/openapi.yaml`、`../api/schema.graphql`
- 文档治理与目录边界：`../DOCUMENT-MANAGEMENT-GUIDELINES.md`、`../README.md`
const QUERY = gql`
  query OrganizationVersions($code: String!) {
    organizationVersions(code: $code) {
      recordId code name unitType status level
      effectiveDate endDate isCurrent createdAt updatedAt parentCode description
    }
  }
`;

async function loadVersions(isRetry = false) {
  try {
    setIsLoading(true);
    setLoadingError(null);
    if (!isRetry) setRetryCount(0);

    const data = await unifiedGraphQLClient.request<{ organizationVersions: Org[] }>(QUERY, { code: organizationCode });
    const list = data?.organizationVersions ?? [];
    const mapped: TimelineVersion[] = list.map(o => ({
      recordId: o.recordId,
      code: o.code,
      name: o.name,
      unitType: o.unitType,
      status: o.status,
      level: o.level,
      effectiveDate: o.effectiveDate,
      endDate: o.endDate ?? null,
      isCurrent: o.isCurrent,
      createdAt: o.createdAt,
      updatedAt: o.updatedAt,
      parentCode: o.parentCode ?? undefined,
      description: o.description ?? undefined,
      lifecycleStatus: o.isCurrent ? 'CURRENT' : 'HISTORICAL',
      businessStatus: o.status === 'ACTIVE' ? 'ACTIVE' : 'INACTIVE',
      dataStatus: 'NORMAL',
      path: '',
      sortOrder: 1,
      changeReason: '',
    }));
    const sorted = mapped.sort((a,b) => new Date(a.effectiveDate).getTime() - new Date(b.effectiveDate).getTime());
    setVersions(sorted);
    setSelectedVersion(sorted.find(v => v.isCurrent) ?? sorted.at(-1) ?? null);
  } catch (e) {
    setLoadingError(e instanceof Error ? e.message : String(e));
    // 回退：旧的单体快照逻辑（可选）
    await loadSingleSnapshotFallback();
  } finally {
    setIsLoading(false);
  }
}
```
