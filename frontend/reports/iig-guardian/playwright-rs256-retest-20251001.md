# Playwright RS256 E2E 复测报告（2025-10-01 第二轮）

**执行时间**: 2025-10-01 23:35 UTC+8
**执行环境**: 本地开发环境（make run-dev + frontend dev）
**认证**: RS256 JWT（通过 `/auth/dev-token` 生成）

---

## 执行摘要

⚠️ **测试失败原因确认：Playwright E2E 测试用例尚未更新至新 GraphQL 契约**

虽然 GraphQL 测试脚本（`tests/e2e-test.sh` 等）已对齐使用 `filter.codes + pagination` 查询 `data`，但 **Playwright E2E 测试套件中的查询仍使用旧的直接字段访问方式**。

---

## 根因分析

### 1. GraphQL Schema 不一致（P0 - 已定位）

**失败测试**: `architecture-e2e.spec.ts:35` - Phase 1: GraphQL统一查询接口验证

**旧查询**（Playwright E2E 当前使用）:
```graphql
{
  organizations {
    code
    name
    unitType
    status
  }
}
```

**新契约**（`docs/api/schema.graphql` + 测试脚本已对齐）:
```graphql
{
  organizations(
    filter: { codes: ["CODE001"] }
    pagination: { limit: 10, offset: 0 }
  ) {
    data {
      code
      name
      unitType
      status
    }
    pagination {
      total
      limit
      offset
      hasMore
    }
  }
}
```

**错误信息**:
```
Cannot query field \"code\" on type \"OrganizationConnection\".
Cannot query field \"name\" on type \"OrganizationConnection\".
Cannot query field \"unitType\" on type \"OrganizationConnection\".
Cannot query field \"status\" on type \"OrganizationConnection\".
```

**修复方案**: 更新 `frontend/tests/e2e/architecture-e2e.spec.ts:43-50` 使用新的查询结构

---

### 2. 测试页面交互元素缺失（P1）

**失败测试**: `basic-functionality-test.spec.ts:60` - 测试页面功能验证

**错误**: `hasButtons = 0`（期望 > 0）

**根因**: 测试页面路由 `/test` 或组件加载异常

**修复方案**: 排查测试页面路由配置，确认组件正确加载

---

### 3. 业务流程页面加载超时（P1）

**失败测试**: `business-flow-e2e.spec.ts` - CRUD 和分页测试

**错误**: `beforeEach` hook 超时（120秒），找不到 `'组织架构管理'` 文本

**根因**: `/organizations` 页面性能问题或权限拦截导致组件未渲染

**修复方案**: 
- 前端团队排查页面加载性能
- 确认 RS256 JWT 权限配置
- 考虑延长超时配置或优化页面加载

---

## 测试执行状态

| 测试用例 | 状态 | 备注 |
|---------|------|------|
| Phase 1: 服务合并验证 | ✅ PASS | 双核心服务架构正常 |
| Phase 1: GraphQL统一查询接口验证 | ❌ FAIL | 查询结构不匹配（需更新测试） |
| 应用基础加载测试 | ✅ PASS | 305ms |
| 系统响应性测试 | ✅ PASS | 63ms |
| 测试页面功能验证 | ❌ FAIL | 交互元素缺失 |
| 错误处理基础验证 | ✅ PASS | 404路由正常 |
| CRUD业务流程测试 | ❌ FAIL | 页面加载超时 |
| 分页和筛选功能测试 | ❌ FAIL | 页面加载超时 |

**测试进度**: 11/154 用例执行后停止（与上次一致）

---

## 关键发现

### ✅ 契约对齐进展（2025-10-08）

根据 06 号文档第 73 行，**GraphQL 测试脚本已全部对齐**：
- `tests/e2e-test.sh`
- `scripts/tests/test-api-consistency.sh`
- `scripts/tests/test-redis-cache-performance.sh`
- `scripts/e2e-test.sh`

以上脚本已改用 `PaginationInput` + `filter.codes` 查询 `data` 字段。

### ❌ Playwright E2E 测试未对齐

**需要更新的文件**:
- `frontend/tests/e2e/architecture-e2e.spec.ts:43-50`
- 其他可能直接查询 `organizations` 的 E2E 测试文件

---

## 修复优先级

| 优先级 | 任务 | 责任人 | 预计耗时 |
|-------|------|--------|---------|
| **P0** | 更新 Playwright E2E GraphQL 查询至新契约 | QA + 后端团队 | 1-2 天 |
| **P1** | 排查 `/organizations` 页面加载性能 | 前端团队 | 2-3 天 |
| **P1** | 修复测试页面交互元素加载 | 前端团队 | 1 天 |

---

## 后续行动

1. **立即**: QA 团队更新 `architecture-e2e.spec.ts` 使用新 GraphQL 查询结构
2. **本周**: 前端团队排查页面加载性能问题
3. **下周**: 重新执行完整 154 项 E2E 回归测试
4. **验证**: 测试通过后更新 06 号文档阻塞项状态

---

## 结论

**RS256 认证链路可用**，测试失败源于 **Playwright E2E 测试用例未同步更新至新 GraphQL 契约**。

虽然后端 GraphQL schema 正确且测试脚本已对齐，但 Playwright E2E 套件仍使用旧的查询方式，导致契约不一致错误。

**关键差异**: Playwright E2E 需要将直接字段查询改为 `filter + pagination + data` 结构。

---

**报告生成**: 2025-10-01 23:38 UTC+8
**下次复测**: P0 修复完成后（预计 2025-10-03）
