# Plan 250A - 模块化单体运行态统一（合流到 :9090）

文档编号: 250A  
标题: 模块化单体运行态统一（将查询服务并入单体进程，统一对外端口 9090）  
创建日期: 2025-11-16  
版本: v1.0  
状态: 待执行（收尾提案）  
关联计划: 202（总体蓝图）、250（合流与边界治理）、251（运行时统一）、253（流水线门禁）、255（CQRS 分层与门禁）

---

## 1. 背景与目标
- 背景：Plan 250 标注“已完成”，其验收目标包含“单体进程（单端口 9090）”与“不得监听 8090”。当前开发环境仍在 docker-compose.dev.yml 中拉起 `graphql-service:8090` 与 `rest-service:9090` 两个容器，运行态未完全合流。
- 目标：在不改变“命令=REST、查询=GraphQL（PostgreSQL 单数据源）”原则的前提下，将查询服务并入单体进程（9090），统一对外暴露 `/api/v1/*` 与 `/graphql` 等入口，移除独立的 8090 容器，实现运行态的一致性与可观测性统一。

提示：本提案仅“索引与收口”，不重复规范正文；一切规则与契约以唯一事实来源为准（见文末索引）。

---

## 2. 范围与不做的事
- 范围：
  - 合流运行态：将 GraphQL 查询路由挂载至 9090 进程，compose 中移除 `graphql-service`；Makefile `run-dev` 统一为单服务拉起。
  - 门禁对齐：引用 Plan 250/253/255 的工作流与规则作为准入与回归保障（不新增第二事实来源）。
  - 文档与证据登记：对 215 执行日志与本计划进行索引登记（仅登记路径与链接）。
- 不做的事：
  - 不更改 CQRS 原则（命令=REST、查询=GraphQL）；不引入第二数据源；不在文档中复制规范正文或实现细节。

---

## 3. 交付物
- 合流运行态变更：
  - docker-compose.dev.yml 移除 `graphql-service:8090`，保留单一 `rest-service:9090`。
  - 在 `cmd/hrms-server/command` 进程内挂载 `/graphql` 与相关健康/指标入口（参考现有查询模块实现与中间件封装）。
  - Makefile `run-dev` 仅启动单体服务；健康检查维持 `/health`、`/metrics`、`/.well-known/jwks.json` 可用。
- 门禁对齐与验证：
  - `plan-250-gates`：唯一二进制/禁 legacy 双服务/禁 8090 监听通过。
  - `plan-253-gates`：compose 端口映射与镜像标签固定通过。
  - `plan-255-gates`：CQRS 分层与端口直连禁用通过（已硬门禁）。
- 文档与证据：
  - 在 `docs/development-plans/215-phase2-execution-log.md` 登记“合流提交、工作流 run 链接、产物路径与回滚方式”；本计划仅索引。

---

## 4. 验收标准
- 运行态：仅有单一 9090 进程对外暴露；`/api/v1/*`（命令）与 `/graphql`（查询）在同进程提供，`/health`、`/metrics` 正常。
- Compose：开发 compose 不再含 `graphql-service:8090`；`plan-253-gates` 对端口映射与镜像标签的门禁通过。
- 门禁：`plan-250-gates`、`plan-253-gates`、`plan-255-gates` 均通过且已设为 Required checks（在 215 登记截图/链接）。
- E2E：前端 E2E 基于单基址（PW_BASE_URL）运行通过（已完成端口直连收敛，索引 logs/plan255/*）。
- 回滚：仅允许本地短期排障（DEV），CI 禁止双服务；任何临时开关须 `// TODO-TEMPORARY(YYYY-MM-DD)` 注记并在 215 登记。

---

## 5. 执行步骤（索引）
说明：以下为执行路径索引，值与规则以唯一事实来源文件为准。

1) 合流实现（进程内挂载查询层）  
   - 在 `cmd/hrms-server/command` 中挂载 `/graphql` 路由与相关中间件（认证/租户/权限），复用现有查询模块公共组件（internal/*）。  
   - 健康与指标：保持 `/health`、`/metrics` 对等可用；JWKS 仍由单体进程提供（`/.well-known/jwks.json`）。

2) Compose 与 Makefile 更新  
   - docker-compose.dev.yml：移除 `graphql-service` 服务、端口 8090 映射与依赖，保留 Postgres/Redis/单体服务；镜像标签固定。  
   - Makefile：`run-dev` 拉起最小依赖（Postgres/Redis）与单体服务；健康检查统一指向 9090；原“两个服务”提示移除。

3) 门禁与 CI  
   - `plan-250-gates`：唯一二进制/禁 8090 监听/禁 legacy 双服务开关 通过。  
   - `plan-253-gates`：compose 端口与镜像标签门禁通过（复用）。  
   - `plan-255-gates`：前端架构守卫与根审计均通过（已硬门禁）。  
   - 在受保护分支启用三项 Required checks，并在 215 登记链接与截图（索引）。

4) 证据登记（215）  
   - 提交哈希、工作流 run 链接、artifact 名称；必要时登记回滚路径与临时例外清单（如存在）。

---

## 6. 风险与回滚
- 风险：合流过程中路由注册/中间件顺序导致的认证/租户/权限回归；健康探针与指标端点兼容性。  
- 回滚：仅允许本地 DEV 场景短期启用 legacy 开关；CI 永久禁止；超期临时方案会被门禁阻断。  
- 监控：利用现有可观测性基线与 E2E 烟测快速发现回归；必要时在 215 登记缓解措施与修复进度。

---

## 7. 单一事实来源索引（权威）
- 架构与流程：AGENTS.md（Docker 强制、SSoT、CQRS 原则、临时方案管控）  
- API 契约：`docs/api/openapi.yaml`（REST+权限 scopes）、`docs/api/schema.graphql`（GraphQL）  
- 工作流与门禁：`.github/workflows/plan-250-gates.yml`、`plan-253-gates.yml`、`plan-255-gates.yml`  
- 开发与执行日志：`docs/development-plans/215-phase2-execution-log.md`  
- 当前运行态来源：`docker-compose.dev.yml`、`Makefile`

---

## 8. 时间与负责人（建议）
- 时间：一个迭代内完成（含验证与回滚预案）。
- 负责人：后端/DevOps 联合；前端配合验证（E2E 单基址已完成）。

---

## 9. 附：对齐说明（与已完成项）
- Plan 250：本计划为“运行态收尾”，使“合流已完成”的状态与 compose/Makefile 的实际运行保持一致。  
- Plan 251：不变更运行时健康与指标契约，仅统一承载位置（9090）。  
- Plan 254：前端端点与代理已收敛为“单基址”；本计划与其策略一致。  
- Plan 255：CQRS 分层与门禁不变；合流后依然是命令=REST、查询=GraphQL（接口路径不变）。

