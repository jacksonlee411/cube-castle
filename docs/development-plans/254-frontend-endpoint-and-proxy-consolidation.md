# Plan 254 - 前端端点与代理整合

文档编号: 254  
标题: 前端端点与代理整合（来源：202 计划拆分）  
创建日期: 2025-11-15  
版本: v1.1  
状态: 已完成  
关联计划: 202、241（前端框架合流）、AGENTS（Docker 强制）

---

## 1. 目标
- 统一前端对后端的访问端点（REST/GraphQL）与本地代理；
- 减少多代理/多端口带来的开发与 E2E 波动，明确 PW_* 与本地 dev server 的约定。
- 明确约束：仅收敛“单基址”，不更改路径前缀（保持 `/api/v1` 与 `/graphql`），与 202 主计划一致。

## 1.1 端口与容器约束（AGENTS 强制）
- 不得修改 docker-compose*.yml 的容器端口映射（如 5432、6379、9090、8090）；
- 如遇端口冲突，必须卸载/停用宿主服务释放端口；禁止通过改映射端口迁就宿主；
- 验证命令示例：
  - `ss -lntp | rg ':(5432|6379|9090|8090)'` 确认无宿主冲突
  - `docker compose -f docker-compose.dev.yml ps` 核对容器端口与健康

## 2. 单一事实来源索引（仅索引，不复制）
- 端口与端点中心：frontend/src/shared/config/ports.ts
  - `SERVICE_PORTS` 与 `CQRS_ENDPOINTS` 定义（端口/端点唯一来源）
- Vite 单基址代理：frontend/vite.config.ts
  - `server.proxy` 将 `/api/v1` → 命令服务，`/graphql` → 查询服务（遵循“命令=REST、查询=GraphQL”）
- Playwright 基址与 server 策略：frontend/playwright.config.ts
  - `PW_SKIP_SERVER` 跳过 webServer、自托管；`PW_BASE_URL` 覆盖基址
- E2E 运行环境配置：frontend/tests/e2e/config/test-environment.ts
- 端口配置验证脚本：frontend/scripts/validate-port-config.ts
- 架构一致性门禁：scripts/quality/architecture-validator.js（CQRS/端口/禁用端点）

以上文件为唯一事实来源；**本计划仅索引实现文件，不在计划文档内复制任何配置值**，所有示例命令仅用于执行路径说明。

---

## 3. 执行步骤（可复制命令）

3.1 预检（与 AGENTS.md 对齐）
- 工具链与镜像
  - `go version` 应与仓库 `toolchain go1.24.9` 一致
  - Node ≥18；NPM registry 锁定为 https://registry.npmjs.org/
- 端口冲突与容器健康
  - `ss -lntp | rg ':(5432|6379|9090|8090)'` 应无宿主服务占用
  - `docker compose -f docker-compose.dev.yml ps` 容器应处于 healthy/运行状态

3.2 启动基础设施与服务（迁移即真源）
- `make docker-up`
- `make run-dev`
- 健康检查（2xx 通过）：
  - `curl -s -o /dev/null -w '%{http_code}\n' http://localhost:9090/health`
  - `curl -s -o /dev/null -w '%{http_code}\n' http://localhost:8090/health`
- 准备鉴权令牌（如需）：`make jwt-dev-mint`（Playwright 会从 `.cache/dev.jwt` 自动读取 `PW_JWT`）

3.3 前端单基址代理验证（本地）
- 启动前端：`cd frontend && npm ci && npm run dev`
- 前端健康：`curl -s http://localhost:3000/health | rg '\"status\":\"ok\"'`
- 核对端点（仅通过前端单基址访问，禁止直连 9090/8090）：
  - REST 命令：`curl -i http://localhost:3000/api/v1/...`
  - GraphQL 查询：`curl -i -X POST http://localhost:3000/graphql`

3.4 端口与架构门禁（可本地/CI 复用）
- 端口配置与硬编码检测：
  - `cd frontend && npm run validate:ports`（失败即需整改）
- 架构一致性（CQRS/端口/禁用端点）：
  - `node scripts/quality/architecture-validator.js --scope frontend --rule cqrs,ports,forbidden`
- 直连检测（源码与产物）：
  - `cd frontend && npm run validate:no-direct-backend`

3.5 E2E 一致性与证据登记
- 运行（默认自启 webServer；如已手启，设置 `PW_SKIP_SERVER=1`）：
  - `cd frontend && npm run test:e2e:254`
- 证据落盘（脚本已内置）：
  - `logs/plan254/playwright-254-run-*.log`
  - `logs/plan254/trace/*.zip`
  - `logs/plan254/report-<timestamp>/`（Playwright HTML 报告）
- 额外校验（防止直连端口）：
  - `rg -n \":(9090|8090)\\b\" logs/plan254/trace logs/plan254/report-* || true`（不应命中）
  - 可选：需要 HAR 时使用 `E2E_SAVE_HAR=1` 再运行（HAR 仍由 Playwright 默认目录产出，执行后将其复制到 `logs/plan254/`）

3.6 登记与同步
- 在 `docs/development-plans/215-phase2-execution-log.md` 登记本计划执行记录与证据路径

注：上述命令仅为执行路径说明；端口/代理/基址等配置请以源文件为准（见 2. 索引）。

---

## 4. 验收标准（可度量/可门禁）
- 单基址代理生效：E2E 期间前端仅向 `/:api/v1` 与 `/graphql` 发起请求，不得出现 `:9090|:8090` 直连（通过 trace/HAR/报告与 `rg` 抽检佐证）
- 架构一致性通过：`architecture-validator` 在 `cqrs,ports,forbidden` 规则下关键违规为 0
- 端口配置通过：`npm run validate:ports` 通过，且未发现问题性硬编码端口
- E2E 通过：`npm run test:e2e:254` 退出码为 0，报告与 trace 按约落盘到 `logs/plan254/*`
- 登记完成：`215-phase2-execution-log.md` 已登记执行证据

---

## 5. 回滚与临时策略
- 如代理合流导致回归，可在短期内通过 `PW_SKIP_SERVER=1` + 手动运行 `npm run dev` 保持现状，同时在相关变更处添加
  `// TODO-TEMPORARY(YYYY-MM-DD): 说明/计划/截止`（≤1 个迭代），并建立清单按期回收
- 严禁为规避宿主端口冲突而修改 docker-compose 端口映射；必须卸载宿主占用服务（见 AGENTS.md）

---

## 6. CI 接入建议（示例片段）
- 预拉取/启动：`make docker-up && make run-dev`
- 前端测试：`cd frontend && PW_SKIP_SERVER=0 npm run test:e2e:254`
- 门禁：`node scripts/quality/architecture-validator.js --scope frontend --rule cqrs,ports,forbidden`；`npm run validate:no-direct-backend`
- 证据归档：打包 `logs/plan254/*` 作为工件

以上仅为执行顺序与变量建议，CI 细节以项目流水线配置为准。

---

## 7. 交付物（更新）
- 端点/代理配置说明与最小可执行路径（本计划文件，索引唯一来源）
- E2E 运行约定（`PW_BASE_URL`、`PW_SKIP_SERVER`）与脚本：`frontend/package.json` 内 `test:e2e:254`
- 证据登记：`logs/plan254/*`（前端端点连通与代理验证输出）

---

维护者: 前端（E2E/后端协作）
