# 12. 时态命令契约回正计划

**文档类型**: 契约修复 / 去重专项  
**创建日期**: 2025-09-24  
**优先级**: P0（唯一性失守 + 生产阻断风险）  
**负责团队**: 命令服务团队（Owner） / 前端组织域团队（Co-owner）  
**关联文档**: `CLAUDE.md`、`AGENTS.md`、`docs/reference/02-IMPLEMENTATION-INVENTORY.md`、`docs/api/openapi.yaml`

---

## 1. 背景与触发

- 组织时态功能的合同事实来源始终是 `/api/v1/organization-units/{code}/versions` 等既有 REST 端点，且当前命令服务依旧通过 `cmd/organization-command-service/internal/handlers/organization.go:167` 提供该能力。
- 历史上曾计划引入新的 `/organization-units/temporal` 路由，并准备了 `temporal_handlers.go.disabled` 中的处理器草稿，但重构在契约更新前即被搁置，未进入主干、未获得 OpenAPI 支持。
- 前端在未更新契约的情况下，自行调用 `/organization-units/temporal` / `/organization-units/{code}/temporal`，导致运行时 404 与重复实现，直接违反“先契约后实现”“资源唯一性与跨层一致性为最高优先级”的原则。
- 该重复路径风险已触发 IIG 与质量门禁告警，必须回收半途而废的实现草稿，恢复单一事实来源。

## 2. 问题画像（证据）

| 证据位置 | 问题说明 |
| --- | --- |
| `frontend/src/features/organizations/components/OrganizationForm/index.tsx` & `frontend/src/shared/hooks/useOrganizationMutations.ts:24-72` | 表单与 Hook 新增了 `/temporal` 调用，绕过契约端点。 |
| `docs/api/openapi.yaml` | `rg "organization-units/temporal" docs/api/openapi.yaml` 无匹配结果，说明契约未更新。 |
| `cmd/organization-command-service/internal/handlers/organization.go:167` | 仍以 `/versions`、`/events` 为唯一可执行端点。 |
| `cmd/organization-command-service/internal/handlers/temporal_handlers.go.disabled` | 未编译的草稿处理器，证明重构停滞。 |
| `reports/implementation-inventory.json` | IIG 未登记 `/temporal` 系列，显示事实来源仍指向 `/versions`。 |

## 3. 根因与约束

1. **半途而废的重构计划**：后端草稿未合并却残留在仓库，给调用方造成“新端点可用”的错觉。
2. **契约流程缺失**：前端跳过 `docs/api/openapi.yaml` 变更与评审，违反 `CLAUDE.md` 的契约先行原则。
3. **跨团队同步失败**：IIG 报表和实现清单未发现重复调用，导致前端误判可以试用 `/temporal`。
4. **唯一性原则被破坏**：同一能力出现两个事实来源（契约内 `/versions` vs. 未发布 `/temporal`），直接违背最高优先级指令。

## 4. 目标与验收标准

| 目标 | 验收标准 |
| --- | --- |
| 恢复唯一事实来源 | OpenAPI、命令服务路由、前端调用全部回归 `/api/v1/organization-units/{code}/versions` 等既有端点；IIG 报表仅登记该能力。 |
| 端到端一致 | 前端 payload 字段与契约字段完全一致（`operationReason`、`effectiveDate` 等），通过 Jest/Vitest + Go 集成测试。 |
| 去除遗留草稿 | `temporal_handlers.go.disabled` 删除或存档至 `archive/`，确保仓库不再出现未发布处理器。 |
| 质量门禁恢复 | `node scripts/generate-implementation-inventory.js`、`node scripts/quality/architecture-validator.js` 与 Playwright 时态场景全部绿灯。 |

## 5. 指导原则（强制）

- ⚠️ **最高优先级**：资源唯一性与跨层一致性优先于一切交付，如果任何环节仍依赖 `/temporal`，必须立即阻断合并。
- 契约 → 实现 → 调用的顺序不得倒置；不得在契约之外创建新端点或代理层。
- 所有修复提交需在 PR 描述中说明“如何验证唯一事实来源与契约一致性”。

## 6. 方案评估与决策

| 方案 | 概述 | 结论 |
| --- | --- | --- |
| A. 完成 `/temporal` 重构 | 重新启用草稿处理器，补充契约并迁移调用 | ❌ 与目标相悖，会再次制造双事实来源，增加维护面。 |
| B. 回退至契约端点并清理遗留 | 前端撤回 `/temporal` 调用，后端删除草稿，维持 `/versions` 为唯一入口 | ✅ 直接恢复唯一性与一致性，符合 CLAUDE.md 与 AGENTS.md 要求。 |
| C. 新增 BFF 代理层 | 由 BFF 翻译 `/temporal` 请求至 `/versions` | ❌ 引入额外事实来源与复杂度，风险更高。 |

**决策**：执行方案 B。

## 7. 行动计划

| 阶段 | 任务 | Owner | 截止 |
| --- | --- | --- | --- |
| Phase 1 | 在 IIG 与架构周会上通报 `/temporal` 重复问题，冻结相关 PR；更新本计划至 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 变更日志 | IIG Guardian | 2025-09-25 |
| Phase 1 | 执行实现清单与架构校验基线：运行 `node scripts/generate-implementation-inventory.js`、`node scripts/quality/architecture-validator.js`，并将输出保存到 `reports/architecture/temporal-contract-baseline-YYYYMMDD.log`，记录摘要与附件路径 | Architecture QA | 2025-09-25 |
| Phase 1 | 删除前端 `/temporal` 调用：统一使用契约 Hook，修正请求 payload（`operationReason`、`effectiveDate` 为 `YYYY-MM-DD`），补充单测 | Frontend Org Team | 2025-09-27 |
| Phase 1 | 清理后端遗留：移除 `temporal_handlers.go.disabled` 并将背景说明归档到 `docs/archive/deprecated-api-design/temporal-handlers.md`（保留成因与回退记录），确认路由表仅包含契约端点 | Command Service Team | 2025-09-27 |
| Phase 1 | 将 `/organization-units/temporal` 列入 CI deny list：在 `scripts/quality/architecture-validator.js` 增加阻断规则，并于 `.github/workflows/agents-compliance.yml` 中强制执行，验收标准为流水线检测到该路径即失败 | Architecture QA | 2025-09-27 |
| Phase 2 | 运行 `make test`、`make test-integration`、`npm --prefix frontend run test`、Playwright 时态场景（详见下文最新剧本需求），更新实现清单与 QA 记录 | 联合 QA | 2025-09-28 |
| Phase 2 | 在 `docs/reference/02-IMPLEMENTATION-INVENTORY.md`、`06-integrated-teams-progress-log.md` 登记结果，关闭本计划 | IIG Guardian | 2025-09-29 |

## 8. 验证步骤

1. `node scripts/generate-implementation-inventory.js`，确认输出无 `/temporal`。  
2. `node scripts/quality/architecture-validator.js`，验证 CQRS 路由与文档一致。  
3. `curl -X POST http://localhost:9090/api/v1/organization-units/{code}/versions` 搭配新 payload，验证 2xx 与审计记录。  
4. Playwright `organization-create.spec.ts` / `temporal-management-integration.spec.ts` 通过。  
5. 记录在 `reports/iig-guardian/` 中，作为唯一事实来源恢复的证据。

### 8.1 最新实施说明（2025-09-27 更新）

- `frontend/tests/e2e/temporal-management-integration.spec.ts` 已重写：
  - UI 场景统一访问 `/organizations/{code}/temporal`，依赖命令服务返回 `/versions` 数据。
  - 后端断言覆盖 `/api/v1/organization-units/{code}/versions` 正常与错误分支，并显式验证 `/organization-units/{code}/temporal` 返回 404。
- 第一次复测命令：`npm --prefix frontend run test:e2e -- --grep temporal`
  - 失败原因：`GET http://localhost:9090/health` 返回非 200，说明命令服务未启动；后续 GraphQL 校验因此未执行。
  - 附件：Playwright 自动生成的 `test-results/temporal-management-integration-*.png/webm`（保留于仓库，用作失败证据）。
- 第二次复测：同一命令在 2025-09-27 重新执行，命令服务仍未启动，`restHealth.ok()` 断言失败，导致 12 项中的 8 项在测试前置阶段终止；Playwright 报告已追加至 09-27 验证日志。

### 8.2 待执行测试前置条件

1. 启动基础服务：依次执行 `make docker-up`、`make run-dev`（或至少启动 `cmd/organization-command-service`），确保 `curl http://localhost:9090/health` 和 `curl http://localhost:8090/health` 均返回 200。
2. 准备认证：运行 `make jwt-dev-setup`（首次）与 `make jwt-dev-mint USER_ID=dev TENANT_ID=default ROLES=ADMIN,USER`，将 `.cache/dev.jwt` 内容分别注入：
   - `export PW_JWT=$(cat .cache/dev.jwt)`
   - `export PW_TENANT_ID=3b99930c-4dc6-4cc9-8e4d-7d960a931cb9`
3. 若命令/查询服务提供自定义端点，请同步设置：
   - `export E2E_COMMAND_API_URL=http://localhost:9090`
   - `export E2E_GRAPHQL_API_URL=http://localhost:8090/graphql`
   - `export E2E_BASE_URL=http://localhost:3000`

满足以上条件后（务必确认服务已启动且健康检查返回 200），重新执行：

```bash
PW_SKIP_SERVER=1 \
PW_JWT=$PW_JWT \
PW_TENANT_ID=$PW_TENANT_ID \
E2E_COMMAND_API_URL=${E2E_COMMAND_API_URL:-http://localhost:9090} \
E2E_GRAPHQL_API_URL=${E2E_GRAPHQL_API_URL:-http://localhost:8090/graphql} \
E2E_BASE_URL=${E2E_BASE_URL:-http://localhost:3000} \
npm --prefix frontend run test:e2e -- --grep "temporal"
```

请将成功/失败结果追加到 `reports/iig-guardian/temporal-contract-rollback-20250926.md`，并在上述命令失败时保留 `test-results/` 目录作为唯一事实来源。

## 9. 风险与应对

| 风险 | 影响 | 缓解措施 |
| --- | --- | --- |
| 前端移除 `/temporal` 后仍有隐藏调用 | 用户仍会触发 404 | 通过 `rg "temporal" frontend/` 全量扫描；CI deny list 由 `.github/workflows/agents-compliance.yml` + `scripts/quality/architecture-validator.js` 阻断，负责团队需在 PR 中附带更新记录。 |
| 删除草稿影响后续架构讨论 | 历史资料丢失 | 将原草稿说明归档到 `docs/archive/deprecated-api-design/temporal-handlers.md`，保留成因、替代方案与回退指引。 |
| Payload 格式未一次调整完成 | 契约测试失败 | 使用共享转换工具 + Vitest 覆盖，强制类型校验。 |

## 10. 交付物

- 前端修复 PR（撤销 `/temporal`、统一契约字段、单元测试）。
- 后端清理 PR（删除草稿处理器、确认路由表、更新审计文档）。
- 归档文档：`docs/archive/deprecated-api-design/temporal-handlers.md` 记录草稿成因与退场理由。
- 基线与复核日志：`reports/architecture/temporal-contract-baseline-20250926.log`、`reports/architecture/temporal-contract-verification-20250926.log`。
- 更新后的 IIG 报表与实现清单条目，明确“时态命令契约回正”状态。  
- QA 验证记录：`reports/iig-guardian/temporal-contract-rollback-20250926.md`，后续复测结果在同一文件追加。

## 11. 评审结论（更新）
- **评审状态**：核心整改已完成，剩余 Playwright 场景需在前端规格调整后再行关闭。
- **关键证据**：
  - `tests/go/temporal_support.go` 已删除，仓库搜索不再出现 `/organization-units/temporal` 路由实现。
  - `scripts/quality/architecture-validator.js` 新增 Forbidden Endpoint 规则；`.github/workflows/agents-compliance.yml` 已强制执行该校验。
  - 基线/复核日志 `reports/architecture/temporal-contract-{baseline,verification}-20250926.log` 显示违规计数为 0。
  - QA 报告 `reports/iig-guardian/temporal-contract-rollback-20250926.md` 追加了 2025-09-26 本地重跑结果：命令/查询/前端服务与样例数据均就绪，REST/GraphQL 探针成功，但 24 个 Playwright 用例因访问不存在的 `/temporal-demo` 页面超时（详见 `test-results/temporal-management-integration-*.png/webm`）。

## 12. 待决事项
- **Playwright 复测**：待前端团队确认演示页面路由（或补齐 `/temporal-demo` 页面）并更新测试脚本后，再次执行 `npm --prefix frontend run test:e2e -- --grep "temporal"`；通过结果需追加至 `reports/iig-guardian/temporal-contract-rollback-20250926.md`。
- **Playwright 复测（更新）**：服务与认证就绪后，按 8.2 步骤重新执行 `npm --prefix frontend run test:e2e -- --grep "temporal"`；运行日志与报告需补录到 `reports/iig-guardian/temporal-contract-rollback-20250926.md`。
- **生产端口验证（可选）**：落地环境可补充 `curl -X POST http://localhost:9090/api/v1/organization-units/{code}/versions` 实测截图，增强审计说明。
- **计划归档**：待 Playwright 复测完成并附证据后，将本计划移至 `docs/archive/development-plans/` 并同步进度日志。

## 13. 建议与后续步骤
- **统一回收策略**：持续以 `rg "temporal"` 覆盖 `tests/`、`scripts/`、`reports/` 等目录，防止二次引入。
- **证据维护**：基线、复核与 QA 产物已统一存放，后续复测结果请在同一文件更新，保持单一事实来源。
- **CI 监控**：首次触发 Forbidden Endpoint 违规时，应在 PR 说明中引用本计划并确认回滚路径；Playwright 场景修复后可在 CI 加入烟囱路由检查。
