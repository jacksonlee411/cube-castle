# Plan 240E – 职位页面回归与 Runbook

编号: 240E  
上游: Plan 240（职位管理页面重构） · 依赖 240A/240B/240C/240D 完成  
状态: 待启动

—

## 目标
- 固化改造经验/回滚路径，清理残留，更新 Plan 06 表格与日志路径，确保持续回归稳定。严格遵循 AGENTS.md 的“资源唯一性与跨层一致性”“Docker 容器化强制”“先契约后实现”等约束。

## 任务清单
T0) 环境与门禁前置校验（以 AGENTS.md 为准；本计划仅补充 240E 专属差异）  
T1) 回归执行（以 Plan 06 标记为 P0 的“职位域”用例集合为准；本地仅作辅助）  
T2) Plan 06 表格更新与日志路径回填；归档录屏与对比截图  
T3) Runbook 精简与指针化（观测/排障/回滚仅指向权威文档，不复制实现细节）  
T4) 质量门禁结果汇总（依赖既有 CI/脚本；本计划仅统一落盘路径）

## 验收标准
- 回归套件在 CI（frontend/playwright.config.ts 已开启 `retries`）下稳定通过；仅使用统一选择器来源；无“偶发绿”。  
- Plan 06 “已执行验证/当前阻塞”两节完成更新并可追溯到本计划产物路径。  
- 文档与质量门禁通过：选择器守卫、架构守卫、临时标签检查、前端 lint/typecheck 全部通过（以既有 CI/脚本为准）。  
- 证据完整落盘（见“证据与落盘”），不引入第二事实来源或平行路径。

## 证据与落盘
- E2E 执行日志：`logs/plan240/E/playwright-{spec}-{browser}.log`  
- Playwright Trace：`logs/plan240/E/trace/{spec}-{browser}-*.zip`  
- 选择器/架构/临时标签/前端 Lint：`logs/plan240/E/*.log`（见 T4；命令以 CI/根级脚本为准）  
- 录屏与截图（可选，作为对比证据）：`reports/plan240/`  
- HAR 说明：由前端配置集中管理，产出统一至 `logs/plan240/B`（或 240BT）；240E 不新增 HAR 路径（历史兼容 + 前端配置集中管理）
- 产物落盘入口：由 CI Workflow 或前端脚本统一负责。推荐统一脚本：`cd frontend && npm run test:e2e:240e`（会将执行日志与 trace 归档至 `../logs/plan240/E`）

---

## T0 – 环境与门禁前置校验（以 AGENTS.md 为准）
- 严格遵循 AGENTS.md“开发前必检”与 Docker 强制；本计划仅补充差异化要求：  
  - 观测开关：真实链路回归需在 CI/本地启用 `VITE_OBS_ENABLED=true`（DEV 默认可用），CI 建议加 `VITE_ENABLE_MUTATION_LOGS=true`（事件通道切换为 mutation）  
  - Mock/真实链路：按用例要求切换 `VITE_POSITIONS_MOCK_MODE`，避免在错误模式下产生“假失败”

---

## 回归套件与执行矩阵（T1）
- 权威来源：以 Plan 06 表格中标记为 P0 的“职位域”用例集合为准（含职位详情/生命周期/观测与相关集成路径）。  
- 过滤规则（客观标准）：优先以 Plan 06 的 `domain=position` 标签/列为准；如 Plan 06 临时缺失该列，兜底集合为 `frontend/tests/e2e/position-*` 与 `frontend/tests/e2e/temporal-management-integration.spec.ts`，但以 Plan 06 为最终事实来源。  
- 本地示例（非权威，仅便于手工复核）：`position-*` 与 `temporal-management-integration` 等规格。具体清单以 Plan 06 为准。

执行要求与环境
- CI：以 frontend/playwright.config.ts 的配置为准（`retries` 打开），Chromium + Firefox 全部稳定通过；产物由 CI 收集并落盘  
- 本地（辅助）：建议仅做一次通过 + 保留 Trace/Report 以便排障；需要时开启 `PW_SKIP_SERVER=1` 与 `E2E_SAVE_HAR=1`

---

## Runbook（T3）
本节提供执行“指针”，避免复制实现细节；事件定义与采集策略以《时态实体体验指南》为准。

1) 环境自检
- 满足 AGENTS.md 的“开发前必检”；确保 9090/8090 health=200，`.cache/dev.jwt` 可读  
- 确认 `VITE_POSITIONS_MOCK_MODE` 与目标用例匹配（Mock 只读 vs 真实链路）

2) 观测与采集
- 真实链路用例：开启 `VITE_OBS_ENABLED=true`（CI 建议叠加 `VITE_ENABLE_MUTATION_LOGS=true`）  
- 事件词汇表、输出通道与采集落盘：参见 `docs/reference/temporal-entity-experience-guide.md` §7（唯一事实来源）。本计划不复述事件语义/字段，避免漂移。  
- 选择器来源：`frontend/src/shared/testids/temporalEntity.ts`（唯一事实来源；禁止硬编码）

3) 常见故障→定位→解法（示例）
- “任职历史”页签缺失或不可点击  
  - 定位：检查 data-testid 是否引用 `temporalEntitySelectors.position.tabId('assignmentHistory')`（只读用例在 Mock 下仍应渲染静态骨架）  
  - 观测：确认是否有 hydrate/tab 相关 `[OBS] position.*` 事件  
  - 解法：修正选择器/路由接入，或在 Mock 模式下补齐骨架占位
- 职位详情白屏/`position-detail-card` 缺失  
  - 定位：GraphQL 请求/错误（参见“观测指南”的 error 事件）  
  - 解法：先在 Mock 模式验证 DOM 完整性，再回到真实链路检查鉴权头与租户
- 组织仪表板加载失败（集成用例）  
  - 定位：入口路由与权限；检查 `PW_TENANT_ID` 是否与后端租户一致  
  - 解法：先通过 `simple-connection-test.spec.ts` 复核可达性，再执行集成用例

4) 回滚路径（仅文档化与开关，不新增运行时代码落盘）
- 前端降级：临时切换 `VITE_POSITIONS_MOCK_MODE=true`（只读保护）；必要时隐藏/禁用版本导出入口  
- MR 级回滚：回退至 240D 前的稳定提交，复跑本计划的“回归套件与执行矩阵”  
- 回滚记录：将操作与结果落盘至 `logs/plan240/E/rollback-{ts}.log`

---

## 文档与质量门禁（T4）
- 必须通过既有门禁（CI/脚本为权威入口）：选择器守卫、计划守卫、架构守卫、临时标签检查、前端 lint/typecheck  
- 本计划仅统一结果落盘：`logs/plan240/E/*.log`（例如 `selector-guard.log`、`architecture-validator.log`、`temporary-tags.log`、`frontend-lint.log`）

---

## Plan 06 对齐项（T2）
- 在 `docs/development-plans/06-integrated-teams-progress-log.md` 中更新：
  - “2. 已执行验证”：新增本计划回归用例的通过/失败记录，附 `logs/plan240/E/*.log` 与 trace 路径  
  - “3. 当前阻塞”：如仍存在 P0 失败，用“故障→定位→解法”格式登记，并引用本计划 Runbook 条目

---

## 附：执行注意事项
- CQRS 一致性：命令=REST、查询=GraphQL，禁止在测试中引入第二数据源或绕过契约（AGENTS.md:7,19）  
- Docker 强制：如端口冲突，请卸载宿主 PostgreSQL/Redis 服务，禁止修改容器端口映射（AGENTS.md:5,67）  
- 证据唯一：不复制同一日志到多处；HAR 路径以前端配置为准（`logs/plan240/B`/`logs/plan240/BT`），E2E 执行日志统一落 `logs/plan240/E`
