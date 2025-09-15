# 06 — 集成团队进展与复核报告（2025-09-15）

最后更新：2025-09-15  
牵头：QA / Backend / Frontend / DevOps  
状态：已完成复核与部分修复；本地E2E验证已跑通环境与认证注入；仍有一项服务端验证待修复

—

## 执行摘要
- 已复核 06 文档列出的问题，完成文档与前端配置修复、E2E 全局认证注入与 CI 集成；本地 E2E 回归用例运行后发现“组织名称含括号的更新”仍被命令端以 400 拒绝，判断为后端名称验证或更新契约限制导致。
- 已将“前端 E2E（Dev Auth）”纳入 CI；当前暂以 continue-on-error 跑通流程，待后端修复后转为强制门禁。
- 下一步请求：由后端放宽名称字符校验并更新 OpenAPI 描述、补充单元测试；完成后我将去除 CI 容错并回归验证。

—

## 已复核问题与结论
1) GraphQL 查询示例与契约不匹配  
- 结论：确有不一致（旧示例使用 `first/hasMore`）。  
- 处置：已更新 docs/api/README.md 与开发者快速参考中的示例为最新分页包装结构（`data + pagination{ total page pageSize hasNext }`）。

2) 认证头一致性与文档化  
- 结论：实现已一致（统一客户端自动注入 Authorization、X-Tenant-ID；后端强制校验；OpenAPI 声明必填）。  
- 处置：在 docs/api/README.md、docs/reference/01-DEVELOPER-QUICK-REFERENCE.md 与 jwt-development-guide.md 补充“必填头部”与 curl 用法。

3) 组织名称验证过严（括号）  
- 结论：前端验证已放宽，但命令端仍返回 400，说明服务端验证或更新契约仍在限制。  
- 处置：新增 E2E 回归用例（含括号名称），用于锁定修复；见“本地 E2E 验证结果”。

4) 前端端口显示与实际不一致  
- 结论：Vite 默认自增端口导致与 Playwright 配置不一致风险。  
- 处置：开启 Vite `server.strictPort=true`，确保与 Playwright `baseURL`（3000）一致。

5) E2E 测试未注入认证  
- 结论：Playwright 缺少全局 JWT/Tenant 注入。  
- 处置：在 `playwright.config.ts` 注入 `extraHTTPHeaders`（从环境变量 `PW_JWT`、`PW_TENANT_ID` 读取）；新增 npm 脚本一键 mint+运行；补充开发/测试指南。

—

## 已落实修复与提交
- 文档与示例：
  - docs/api/README.md：GraphQL 新契约示例（分页包装）与必填头部说明。
  - docs/reference/01-DEVELOPER-QUICK-REFERENCE.md：补充认证头必填、GraphQL 新示例、E2E 认证注入与名称验证说明。
  - docs/development-guides/jwt-development-guide.md：新增 Playwright E2E 认证注入说明。
- 前端配置：
  - frontend/vite.config.ts：`server.strictPort=true`。
  - frontend/playwright.config.ts：全局 `extraHTTPHeaders` 注入 `Authorization`、`X-Tenant-ID`（来自 `PW_JWT`、`PW_TENANT_ID`）。
  - frontend/package.json：新增 `e2e:auth:dev`（mint dev token → 注入 PW_* → 跑测试）。
- 回归用例：
  - frontend/tests/e2e/name-validation-parentheses.spec.ts：验证“组织名称允许括号”；已升级为“GraphQL 读取实体 → 全量 PUT 更新（仅改 name）”。
- CI 集成：
  - .github/workflows/frontend-e2e.yml：新增“Frontend E2E (Dev Auth)”工作流，启动 Postgres/Redis、run-dev 后执行 `npm run e2e:auth:dev`（当前 continue-on-error: true）。

—

## 本地 E2E 验证结果（2025-09-15）
- 环境健康：命令(9090)/查询(8090) healthy；依赖安装与浏览器就绪。
- 执行用例：`tests/e2e/name-validation-parentheses.spec.ts` 2/2 失败（chromium、firefox）。
- 失败详情：PUT /api/v1/organization-units/{code} 返回 400（已采用全量载荷，仅改 name）。
- 结论：命令端名称验证或更新契约（字段/模式）对括号仍有限制，需后端修复。

—

## 下一步请求（需要后端与架构批准并实施）
1) 放宽后端组织名称验证（缺陷修复）  
- 允许常见企业命名字符：中文/英文/数字/空格、连字符（-/_）、圆括号 ()、全角括号 （）、中点 · 等；长度 ≤100/或与现有最大值一致（OpenAPI 当前示例为 ≤255，可按统一标准收敛）。
- 实现建议：优先使用 Unicode 类别（L/N/Zs）+ 小型白名单符号，避免复杂正则边界问题；集中到单一 validator，被 REST/GraphQL 复用。
- 同步更新：
  - OpenAPI name 字段描述（允许字符集合与长度），示例包含括号。
  - 后端单元测试：创建/更新包含括号的名称通过；非法字符用例拒绝。

2) 更新接口契约说明与测试  
- 明确 PUT 为“替换式更新”（必须传完整载荷）；未来若要支持部分更新，另行评审新增 PATCH。
- 扩展合约/集成测试覆盖“名称含括号”的创建与更新。

3) CI 门禁调整（修复完成后）  
- 移除 frontend-e2e.yml 的 `continue-on-error: true`，使“Frontend E2E (Dev Auth)”成为强制门禁。

—

## 风险与影响
- 风险：
  - 放宽字符集后需关注 SQL 注入、XSS 等风控（本系统以参数化与后端编码为全局策略，同时 E2E/单测覆盖回归）。
  - 开启 strictPort 后，如 3000 被占用将导致 dev 启动失败（文档已提示；CI 固定端口无风险）。
- 影响面：
  - 前端表单与后端接口对齐后，含括号名称创建/更新的链路将稳定；E2E 用例可持续守护。

—

## 验收标准（DoD）
- 后端：
  - 名称验证放宽并有单元测试；OpenAPI 描述与示例更新。
  - PUT 更新含括号名称返回 200；非法字符按规则拒绝。
- 前端：
  - 回归用例 `name-validation-parentheses.spec.ts` 稳定通过。
- CI：
  - Frontend E2E (Dev Auth) 无容错运行通过；成为强制门禁。

—

## 附：运行与验证速查
- 本地 mint 并跑 E2E：
```bash
make jwt-dev-mint && eval $(make jwt-dev-export)
cd frontend && npm run e2e:auth:dev
```
- Playwright 全局认证变量：`PW_JWT=$JWT_TOKEN`，`PW_TENANT_ID=3b99930c-4dc6-4cc9-8e4d-7d960a931cb9`
- GraphQL 示例（分页包装 + 必填头部）：见 `docs/api/README.md`

—

## 变更记录
- 2025-09-15：重写本报告为“进展与复核”版本；记录文档修复、前端配置与 E2E 认证注入、CI 集成、新增回归用例与本地执行结果；提出后端修复请求与门禁调整计划。

