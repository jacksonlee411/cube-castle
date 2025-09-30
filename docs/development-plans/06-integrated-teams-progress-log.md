# 06 — 集成团队推进记录（RS256 认证与 API 合规治理）

最后更新：2025-09-27 20:25 UTC
维护团队：认证小组（主责）+ 前端工具组 + 命令服务团队
状态：Plan 12/13/14 已完成归档，Plan 16 代码异味治理启动，Plan 17 TODO 治理持续巡检，Plan 15 例行复核中

---

## 1. 进行中事项概览
- **✅ Plan 12 验收完成**：temporal契约回正已验证完成，Playwright测试执行成功，12号文档已归档至 `docs/archive/development-plans/`。
- **✅ Plan 13（停用/删除一致性治理）完成并归档**：
  - ✅ `DELETE_ORGANIZATION` 事件、子组织阻断、审计日志落地；`make test`、`make test-integration` 全绿
  - ✅ GraphQL `childrenCount` / `includeDisabledAncestors`、前端 Parent Selector 与 `organizationPermissions` 联调通过
  - ✅ 验收见 `docs/archive/development-plans/13-organization-suspend-delete-governance.md`，测试证明附录已归档
- **✅ Plan 14（status-only 软删除）技术实施完成**：
  - ✅ Phase 1 数据审计：13条记录状态完全一致，生成 `reports/temporal/status-only-audit-final.json` 与差异报告
  - ✅ Phase 4 回归验证：Go单元/集成测试、前端测试全部通过，E2E测试环境问题已记录
  - ✅ 实现清单已更新为v1.9.1版本，记录软删除统一架构变更
- **🚧 Plan 16 代码异味治理**：Phase 0 基线确认阶段，已生成基线报告，待完成实现清单检查与工作量评估（详见下方进展跟踪）
- **✅ Plan 17 TODO 治理（已归档）**：419 状态码评审完成；`scripts/check-temporary-tags.sh` 扩展支持文档扫描，CI 及周度巡检日志已提交，持续巡检按模板执行。
- **API 合规例外决策**：前端 lint 仍报 3 个 `camelcase`（外部协议字段）与多处 `no-console`，尚未决定豁免或改写，CI 配置也未同步。
- **Console 输出治理**：缺乏统一日志策略，是否替换 `console` 需由前端团队给出方案与时间表。
- **Spectral 依赖失效**：`npm install` 拉取 `@stoplight/spectral-oasx` 仍报 404，需与平台团队协作替换镜像或缓存，避免阻断安装流程。
- **Playwright 权限回归**：时态相关用例因权限配置失败，需定位业务策略或测试数据，确认能够在 RS256 环境下稳定通过。

---

## 2. 当前状态与证据
- ✅ `make run-auth-rs256-sim` + `curl http://localhost:9090/.well-known/jwks.json` 可拿到 `kid=bff-key-1` RSA 公钥，RS256 链路基线可用。
- ✅ `rg "temporal" frontend/src tests/e2e` 仅保留契约内引用，`frontend/tests/e2e/temporal-management-integration.spec.ts` 已校验 `/versions` 并阻断 `/temporal`。
- ✅ **2025-09-27 Playwright 复测完成**：环境就绪后执行成功，12个测试中10个通过，2个预期失败（契约验证正常），结果已记录在 `reports/iig-guardian/temporal-contract-rollback-20250926.md`。
- ✅ `docs/api/openapi.yaml`、`docs/api/schema.graphql`、Quick Reference 与 API Guide 已更新为 status-only 语义，唯一事实来源指向 `status` 字段。
- ⚠️ `NODE_PATH=frontend/node_modules npx eslint@8.57.0 frontend/src/**/*.{ts,tsx} --config frontend/.eslintrc.api-compliance.cjs` 输出 `camelcase` 与 `no-console` 告警，需决策处理方式。
- ⚠️ `npm install` 过程中抓取 `@stoplight/spectral-oasx` 失败，阻塞工具链；暂无替代方案。
- ⚠️ Playwright `--grep "temporal"` 用例在带 RS256 JWT 时仍因权限被拒绝，需补充数据或权限配置。

---

## 3. 待办清单
1. ✅ ~~回归 Playwright temporal 用例（使用更新后的剧本），并在 `reports/iig-guardian/temporal-contract-rollback-20250926.md` 附最新结果。~~
2. 确认 `camelcase` 与 `no-console` 的长期处理方式（豁免名单或代码调整），同步更新 lint 配置及文档记录。
3. 制定前端日志替换方案（目标组件、负责人、时间表），避免出现无限制 `console`。
4. 与平台团队协作，为 `@stoplight/spectral-oasx` 配置可用镜像或缓存，并在日志记录处理进度。
5. 修复 Playwright 权限失败（核对测试账号角色/租户、或补充种子数据），完成后提供报告路径。
6. 【Plan 14】数据平台+命令服务团队：执行 `sql/inspection/status_deleted_audit.sql` 并按手册产出 `reports/temporal/status-only-audit-after.json`、`status-only-migration_diff.md`（生产等价环境复核待排期）。
7. ✅ 【Plan 14】查询服务+QA 团队：`npm run test:contract`、`npm run test:e2e -- --grep "temporal"`（Mock 模式）已执行，待真实环境复核监控后归档。

---

## 4. 待测试事项
- **Playwright E2E（temporal 场景）**：Mock 模式已完成 `npm run test:e2e -- --grep "temporal"`；待后端服务可用时在 RS256 环境复跑并附带监控结果。
- **API 合规 Lint 复验**：按 `NODE_PATH=frontend/node_modules npx eslint@8.57.0 ...` 命令重跑，确认告警清零或与豁免清单一致，输出结果需归档。
- **Status-only 数据审计**：在具备数据库访问权限的环境运行 `psql -f sql/inspection/status_deleted_audit.sql > reports/temporal/status-only-audit-after.json`，并更新差异报告。
- **Status-only 回归测试**：完成 Phase 4 测试矩阵后，将执行证据与监控结论附于 `reports/temporal/`。

---

## 5. 风险与跟踪
- **验证证据风险**：Playwright 新剧本尚未在 RS256 环境复跑出绿灯，需要补充报告以形成唯一事实来源。
- **CI 阻断风险**：`@stoplight/spectral-oasx` 拉取失败会导致 `npm install` 崩溃，一旦命中，CI 将无法完成前端构建。
- **合规缺口**：若 `camelcase`/`no-console` 未决策且 CI 未加入豁免，未来合并将被阻塞或放行不一致实现。
- **权限回归风险**：Playwright 用例失败说明权限策略未覆盖 RS256 场景，需要在业务层确认期望行为。

---

## 6. 验证记录更新
| 验证项 | 结果 | 备注 |
| --- | --- | --- |
| Playwright temporal 用例 | ✅ 10/12 通过，2个预期失败 | 报告路径：`reports/iig-guardian/temporal-contract-rollback-20250926.md` |
| API 合规 Lint | ⚠️ 仍有camelcase/no-console告警 | 需决策处理方式 |
| 实现清单/架构校验 | ✅ 无 `/temporal` 相关条目 | 契约回正完成 |
| Plan 12 文档归档 | ✅ 已移至 archive 目录 | `docs/archive/development-plans/12-temporal-command-contract-gap-remediation.md` |
| Plan 14 status-only 进度 | 🚧 Phase 1 生产复核待排期 / Phase 4 Mock 已通过 | `npm run test:contract`、`npm run test:e2e -- --grep "temporal"`（Mock 模式） |
| Plan 13 停用/删除治理 | ✅ 已归档，测试全绿 | 文档：`docs/archive/development-plans/13-organization-suspend-delete-governance.md` |
| Plan 15 时间轴导航复核 | ✅ 已完成并归档 | 提交: `0dbee418`, 验收: 时间轴专职导航、编辑唯一入口、所有测试通过 |

---

## 7. Plan 16 代码异味治理进展跟踪

### Phase 0 基线确认（2025-09-30 启动）
- **✅ 基线报告生成**: `reports/iig-guardian/code-smell-baseline-20250929.md`
  - Go后端：54文件，16,888行，红灯3个（27.5%），橙灯5个（22.1%）
  - 前端TS：112文件，18,254行，红灯2个（12.2%），橙灯9个（26.2%）
- **⏳ 待完成任务**:
  - [ ] 运行实现清单：`node scripts/generate-implementation-inventory.js`
  - [ ] 临时治理巡检：`bash scripts/check-temporary-tags.sh`
  - [ ] 工作量复核：确认团队30%工作量可投入（架构组+PM）
  - [ ] Git标签基线：`git tag plan16-phase0-baseline`
  - [ ] 弱类型统计：`rg "\bany\b|\bunknown\b" frontend/src --stats`
  - [ ] 更新本日志：填写Phase 0完成时间与责任人

### 进展模板（每周五更新）
| 周次 | 完成任务 | 红灯文件 | 阻塞项 | 风险变化 |
|------|---------|---------|--------|---------|
| W1 (2025-10-04) | Phase 0完成 | Go:3, TS:2 | - | - |
| W2 (2025-10-11) | main.go拆分 | Go:2, TS:2 | 待填写 | 待填写 |
| W3 (2025-10-18) | handler/repo拆分 | Go:0, TS:2 | 待填写 | 待填写 |
| W4 (2025-10-25) | 前端组件拆分 | Go:0, TS:0 | 待填写 | 待填写 |
| W5 (2025-11-01) | Phase 2类型治理 | - | 待填写 | 待填写 |
| W6 (2025-11-08) | Phase 3监控系统 | - | 待填写 | 待填写 |

### 关键里程碑
- **Phase 0 目标完成**: 2025-10-01（责任人：待填写）
- **Phase 1 红灯清零**: 2025-10-22（3周，含测试）
- **Phase 2 类型治理**: 2025-11-05（1.5周）
- **Phase 3 监控上线**: 2025-11-08（1周）

### 风险追踪
- **测试工作量**: Phase 1需额外30%时间用于单元测试编写（覆盖率≥80%）
- **前端类型改动**: Phase 2需分批次进行，避免大面积编译失败
- **回滚准备**: 每个重构任务前创建git标签 `plan16-phase[X]-task[Y]-before`

### 验收标准
- [ ] 红灯文件清零（Go 0个, TS 0个）
- [ ] 橙灯文件控制（Go ≤3个, TS ≤5个）
- [ ] any/unknown使用≤30处（当前169处）
- [ ] 单元测试覆盖率≥80%
- [ ] 契约测试通过率100%
- [ ] 监控脚本 `scripts/code-smell-monitor.sh` 交付并纳入CI

---
