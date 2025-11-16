# Plan 255 – 根路径端口/禁用端点审计整改清单

文档编号: 255-AUDIT-ROOT
创建日期: 2025-11-16
版本: v1.0
状态: 进行中（迭代回收项）

—— 本清单仅登记“整改任务与证据索引”，不复制门禁/规则正文。规则与运行流程的唯一事实来源：
- 守卫与脚本：scripts/quality/architecture-validator.js、eslint.config.architecture.mjs
- CI 工作流：.github/workflows/plan-255-gates.yml（PLAN255_ROOT_AUDIT_MODE 控制软/硬门禁）
- 远程登记与执行日志：docs/development-plans/215-phase2-execution-log.md

## 摘要（本轮审计）
- 报告文件（最新）：logs/plan255/architecture-root-20251116_163330.json
- 统计（最新）：总 1（端口违规 0、禁用端点 1；自检常量，忽略）
- 历史快照：logs/plan255/architecture-root-20251116_162605.json（总 36：端口 22、禁用端点 14）
- 范围：root 审计（非门禁），用于发现 frontend/src 之外的硬编码端口与禁用端点
- 处理策略：
  - 测试/E2E 文件：替换直连/硬编码端口为统一注入（PW_BASE_URL、SERVICE_PORTS 等）；保留少量“端口占位值”在专属测试常量
  - 产物与第三方：已更新审计脚本忽略 third_party、playwright-report 目录（避免噪音）
  - 工具本体（architecture-validator.js）的模式常量：属自检匹配，保留（不计为整改项）

## 分组整改（按目录）

1) frontend/tests/e2e/*
- 发现文件：
  - frontend/tests/e2e/activate-suspend-workflow.spec.ts（ports+forbidden）
  - frontend/tests/e2e/architecture-e2e.spec.ts（多处 ports + forbidden）
  - frontend/tests/e2e/business-flow-e2e.spec.ts（ports+forbidden）
  - frontend/tests/e2e/name-validation-parentheses.spec.ts（ports+forbidden）
  - frontend/tests/e2e/operational-monitoring-e2e.spec.ts（ports+forbidden）
  - frontend/tests/e2e/position-crud-full-lifecycle.spec.ts（ports+forbidden）
  - frontend/tests/e2e/position-observability.spec.ts（ports+forbidden）
  - frontend/tests/e2e/temporal-editform-defaults-smoke.spec.ts（ports+forbidden）
  - frontend/tests/e2e/temporal-graphql-comprehensive.spec.ts（ports）
  - frontend/tests/e2e/utils/authToken.ts（ports+forbidden）
- 整改建议（默认方案）：
  - 使用环境变量 PW_BASE_URL 注入后端基址（单基址代理），禁止写死 :9090/:8090；示例：`const base = process.env.PW_BASE_URL || '/'; fetch(\`\${base}api/v1/...'\`)`
  - GraphQL 查询仅走 `/graphql` 单入口；命令类 REST 仅走 `/api/v1` 单入口
  - 若确需端口占位（本地模拟），集中到测试常量（例如 tests/e2e/config.ts），并统一以 SERVICE_PORTS 映射导出（避免分散硬编码）
- 证据：见上文报告 JSON 中逐文件条目（filePath+line+message）

2) cmd/oauth-service/main.js
- 发现：行 13 同时出现 3000/3001 端口硬编码（ports）
- 建议：读环境变量或统一配置（SERVICE_PORTS.FRONTEND_DEV / SERVICE_PORTS.FRONTEND_PREVIEW）

3) scripts/quality/architecture-validator.js
- 发现：禁止端点模式匹配（forbidden，行 52），属于工具自检常量
- 建议：无需整改（已将 third_party 与 playwright-report 排除，避免工具/产物误报；本条可忽略）

4) 已忽略（不作为整改项）
- third_party/**：外部示例与依赖产物（已在扫描器中排除）
- frontend/playwright-report/**：报告与产物（已在扫描器中排除）

## 执行与回收
- 子任务拆分：按文件建立 issue（含路径、行号、建议替换方式），批量提交 PR（可按模块或用例分批）
- 门禁切换：完成本清单后，将 `PLAN255_ROOT_AUDIT_MODE=hard`（在 plan-255-gates.yml 中），root 审计转为阻断
- 截止与标签：`// TODO-TEMPORARY(2025-11-30)`；过期未回收项需在 215 文件中记录延期原因与新期限

## 证据与索引
- 审计 JSON：logs/plan255/architecture-root-20251116_162605.json
- 审计日志：logs/plan255/audit-root-20251116_162605.log
- 215 登记：docs/development-plans/215-phase2-execution-log.md（“Plan 255 · CI 远程收尾（索引）”）
