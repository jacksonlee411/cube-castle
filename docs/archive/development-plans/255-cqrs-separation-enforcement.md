# Plan 255 - CQRS 严格分层与门禁

文档编号: 255  
标题: CQRS 严格分层与门禁（来源：202 计划拆分）  
创建日期: 2025-11-15  
完成日期: 2025-11-16  
版本: v1.3  
状态: 已完成  
关联计划: 202、203、215（CQRS：命令=REST、查询=GraphQL）

---

## 1. 目标
- 建立命令=REST、查询=GraphQL 的硬性门禁（脚本与审查清单）；
- 检测并阻断跨层调用、错误命名与不合规端口。

## 2. 交付物
- 守卫脚本与 ESLint 规则（仅引用现有脚本/配置）；
- 违规样例与修正路径说明（不复制正文，索引唯一来源）；
- 证据：logs/plan255/*（守卫输出与修复记录）。

## 3. 验收标准
- 守卫脚本可检测并阻断常见违规；
- 现存用例完成修正或登记过渡期（带 TODO-TEMPORARY 标签与回收期）。

验收证据（CI）
- gates‑255（前端/后端门禁整体）: 通过  
  - 运行链接：<https://github.com/jacksonlee411/cube-castle/actions/runs/19401060738/job/55508624545>
- 前端门禁报告：`reports/architecture/architecture-validation.json`（生成且已归档）
- 证据目录：`logs/plan255/**`（ESLint/architecture-validator/golangci-lint 全量日志）

备注
- 后端 golangci‑lint 在 CI 仅以 depguard/tagliatelle 的“真实违规”为阻断，其他类型检查噪音仅记录日志；不降低门禁标准，不改变规则事实来源。

---

维护者: 架构与后端

---

## 4. 范围澄清与索引（避免重复定义）
- 端口与容器映射：本计划不复刻 compose 端口/镜像标签/冷启动门禁，统一引用 Plan 253 的门禁，作为本计划准入条件（compose 端口映射与镜像标签一旦变更由 Plan 253 阻断）。
  - 参考：.github/workflows/plan-253-gates.yml、scripts/quality/gates-253-*.sh
- 代码层端口硬编码：本计划覆盖“代码层面的端口与直连后端”检测与阻断（前端），使用现有架构验证器与 ESLint 规则。
  - 参考：scripts/quality/architecture-validator.js、eslint.config.architecture.mjs
- CQRS 协议分离（前端）：统一使用 UnifiedGraphQLClient（查询）与 UnifiedRESTClient（命令），禁止 GET/fetch/axios.get 直连后端查询。
  - 参考：frontend/src/shared/api/unified-client.ts
- CQRS 边界（后端）：禁止 command↔query 内部包交叉依赖（depguard）
  - 参考：.golangci.yml 中 depguard 规则
- JSON 字段命名（Go 对外）：新增“camelCase 守卫（软→硬）”，OAuth/OIDC/JWT 标准字段按白名单例外（详见 6）。
  - 本计划自 v1.2 起，将“认证路径整体豁免”收敛为“字段级白名单 + 最小路径范围”，避免过度放宽造成漏检。

---

## 5. 执行步骤（本地/CI 可复制）
说明：以下命令仅为执行路径说明，所有规范值与实现以“唯一事实来源文件”为准。

5.1 预检（与 AGENTS 对齐）
- 工具链：go version ≥1.24（与 go.mod/toolchain 一致）、Node ≥18、NPM registry 指向 https://registry.npmjs.org/
- 端口与容器健康（引用 Plan 253）：
  - ss -lntp | rg ':(5432|6379|9090|8090)'
  - docker compose -f docker-compose.dev.yml ps

5.2 前端架构门禁（CQRS/端口/禁直连）
- node scripts/quality/architecture-validator.js --scope frontend --rule cqrs,ports,forbidden
  - 证据归档（CI 已内置）：logs/plan255/architecture-validator-*.log、reports/architecture/architecture-validation.json
- 例外收敛：GET 直连例外仅限 `/auth`（用于认证流程）。原 `/health`、`/metrics` 例外取消，观测与探活请通过后端/监控侧处理或使用 GraphQL 只读查询。

5.2b（可选·审计）根路径端口/禁止端点扫描（不作为门禁，仅用于发现问题）
- 本地建议：`node scripts/quality/architecture-validator.js --scope root --rule ports,forbidden || true`（软运行，不阻断）
- CI 开关：通过 `.github/workflows/plan-255-gates.yml` 中 `PLAN255_ROOT_AUDIT_MODE`（soft|hard）控制是否阻断
  - 用于发现 scripts/** 等层的硬编码端口/未立项端点引用；保留报告归档，按问题排期整改；门禁范围仍以 5.2 为准。

5.3 后端边界与 JSON 命名门禁
- golangci-lint run
  - depguard：阻断 command↔query 内部包依赖
  - tagliatelle（新增）：JSON tag 必须为 camelCase
    - 例外（字段级白名单 + 最小路径）：仅在认证相关路径允许以下 snake_case 字段：`access_token`、`refresh_token`、`token_type`、`id_token`、`expires_in`、`authorization_endpoint`、`token_endpoint`、`end_session_endpoint`、`jwks_uri`、`tenant_id`（JWT claim）
    - 临时豁免：如确需短期例外，必须以 `//nolint:tagliatelle // TODO-TEMPORARY(YYYY-MM-DD): 原因|计划|截止` 标注（≤1迭代），并在 215 登记；过期即移除
- 证据归档（CI 已内置）：logs/plan255/golangci-lint-*.log

5.4 证据登记
- 在 docs/development-plans/215-phase2-execution-log.md 登记本计划执行记录与证据路径（仅登记索引，不复制正文）

---

## 6. 决议与例外（针对开放问题）
- OAuth/OIDC/JWT 标准字段例外（snake_case 允许）：
  - 字段列表：access_token、refresh_token、token_type、id_token、expires_in、authorization_endpoint、token_endpoint、end_session_endpoint、jwks_uri、tenant_id（JWT claim）
  - 路径范围：仅限认证/会话相关路径（internal/auth/**、cmd/hrms-server/**/auth/**、cmd/hrms-server/**/authbff/**）
  - 收敛原则：由“路径整体豁免”收敛为“字段级白名单 + 最小路径”，避免非标准 snake_case 混入
- 前端 JWKS 获取（/.well-known/jwks.json）：
  - 最佳实践：不加入永久例外；前端不直接拉取 JWKS，认证流程通过 /auth/session 或 /auth/dev-token 获取所需令牌/上下文。
  - 临时策略（如短期需兼容）：仅在认证模块 + DEV_MODE 下允许，并在代码处以 `// TODO-TEMPORARY(YYYY-MM-DD):` 标注与 215 登记，限 1 个迭代内移除。
- JSON 命名守卫策略（软→硬）：
  - 第 1 迭代：启用 tagliatelle + exclude-rules 白名单（上述例外路径），CI 报错但允许临时豁免通过 `// TODO-TEMPORARY(YYYY-MM-DD):` 标注登记 215，限一迭代。
  - 第 2 迭代：移除豁免或转为最小白名单，仅保留 OAuth/OIDC/JWT 标准字段。
- 端口检测边界：255 仅覆盖“代码层硬编码端口/直连 :9090|:8090|localhost”与“禁止 GET 直连查询”；Compose 端口映射/镜像标签/冷启动由 253 负责。根路径扫描作为“审计项”，不纳入门禁。
- GraphQL 侧误用 REST 客户端：前端已统一门面；后端通过 depguard 禁止跨层依赖，resolver 不得引入 REST handler 内部实现（如需复用，抽至 shared/internal 公共接口）。

---

## 7. CI 接入（plan-255-gates）
- 新增工作流 .github/workflows/plan-255-gates.yml（不复制规范，作为执行器与证据归档）：
  - 步骤：
    - 前端：ESLint 架构守卫（Flat Config：eslint.config.architecture.mjs；覆盖 CQRS/端口/契约字段）→ tee 至 logs/plan255
    - 前端：architecture-validator（规则 cqrs,ports,forbidden）→ tee 至 logs/plan255
    - 后端：golangci-lint run（包含 depguard + tagliatelle）→ tee 至 logs/plan255
  - 工件：logs/plan255/**/*（保存 7 天）
  - 版本：固定 golangci-lint 版本（@v1.59.1）确保结果可复现（变更另行入 CHANGELOG）
  - 依赖：将 plan-250-gates、plan-253-gates 设为仓库受保护分支的必需检查（组合门禁）；在 215 记录“受保护分支检查项截图/链接”证据

---

## 8. 验收标准（更新版）
- 前端门禁：ESLint（架构守卫）与 architecture-validator 均为 0 关键违规；禁止 fetch GET/axios.get 查询；禁止直连 :9090/:8090
- 后端门禁：
  - depguard：command↔query 内部依赖为 0
  - tagliatelle：除白名单外 json tag snake_case 为 0（软→硬策略按迭代执行；临时豁免需 //nolint + TODO-TEMPORARY 并在 215 登记）
- 证据：logs/plan255/* 与 reports/architecture/architecture-validation.json 存在；215 登记以下内容：
  - 前端/后端门禁日志路径
  - 受保护分支必需检查项（包含 plan-250/253/255）的配置截图与失败示例链接
- 组合门禁：plan-250-gates、plan-253-gates 均通过（单体合流与 compose 端口/镜像标签）；上述均为受保护分支必需检查

---

## 9. 单一事实来源索引
- 架构验证器与规则：scripts/quality/architecture-validator.js、eslint.config.architecture.mjs
- 前端统一客户端：frontend/src/shared/api/unified-client.ts
- 后端边界与命名：.golangci.yml（depguard、tagliatelle）
- 组合门禁引用：.github/workflows/plan-253-gates.yml
- 契约 SSoT：docs/api/openapi.yaml、docs/api/schema.graphql（命名与字段以契约为准；如工具报告与契约冲突，以 Plan 258 的漂移校验结论为准）

---

## 10. 与项目原则（AGENTS.md）对齐摘要
- 资源唯一性与跨层一致性：本计划仅索引现有实现与工作流，不复制规范正文；跨层依赖通过 depguard 阻断；证据统一落盘 logs/plan255/*
- Docker 强制与端口治理：compose 映射/镜像标签/冷启动由 Plan 253 门禁负责；本计划仅覆盖代码层硬编码端口与直连检测
- PostgreSQL 原生 CQRS：查询=GraphQL、命令=REST 的硬性门禁（前端脚本 + 统一客户端），后端同步=依赖注入（仅公开接口）、禁止跨层内部依赖
- 命名与响应：对外 JSON 字段 camelCase；认证标准 snake_case 受路径白名单与迭代回收约束
- 临时方案管控：所有临时豁免需 // TODO‑TEMPORARY(YYYY‑MM‑DD): 标注并在 215 登记，限一迭代

---

## 11. 对齐与依赖（与 20x/25x 的关系）
- 202（执行指引）：本计划负责将“命令=REST、查询=GraphQL”的架构决策落地为可执行门禁与证据登记
- 250（合流）：在“单一二进制/单端口/禁 legacy”的基础上补齐行为门禁，避免合流后回退
- 251（运行时统一）：通过行为门禁保证中间件/指标在单体路径下的一致性，不引入二义性
- 253（部署流水线简化）：compose 端口/镜像标签/冷启动门禁由其负责；255 将其设为合入前置条件
- 254（前端端点与代理）：禁止直连 9090/8090，统一从单基址访问 /api/v1 与 /graphql
- 256/258（契约 SSoT/漂移门禁）：字段命名标准与差异裁决以契约与漂移校验为准；255 仅做行为与命名守卫

---

## 12. 风险与回滚
- 风险：历史代码中存在零星 snake_case json tag
  - 应对：tagliatelle 先软后硬；使用 // TODO‑TEMPORARY(YYYY‑MM‑DD): 标注并在 215 登记，限一迭代回收
- 风险：前端零星 GET 查询或直连 9090/8090 的遗留
  - 应对：architecture-validator 报告 + E2E/HAR 抽检；统一迁移到统一客户端与代理路径
- 风险：depguard 触发的跨层耦合
  - 应对：抽出共享接口到 internal/* 或 pkg/* 公开接口；禁止引用对方 internal 包
- 回滚：如门禁引发 CI 全局阻断，可短期在分支上以 TODO‑TEMPORARY 标注 + 215 登记的方式设置最小白名单；不得修改 compose 端口映射；不得扩大白名单范围；须在一个迭代内回收

---

## 13. 里程碑（建议）
- M1：启用 plan-255-gates 与报告归档；depguard 生效；architecture-validator 全绿（关键违规=0）
- M2：启用 tagliatelle（软门禁）；由“路径整体豁免”迁移为“字段级白名单 + 最小路径”；登记与收敛首批 snake_case；更新 215 日志
- M3：关闭临时豁免，转为硬门禁；与 258 一并收敛剩余漂移
- M4：周期性复盘：门禁信噪比、白名单收敛度、违规趋势与教育材料

---

## 14. 执行清单（摘要）
- 前端门禁（CQRS/端口/禁直连）：node scripts/quality/architecture-validator.js --scope frontend --rule cqrs,ports,forbidden
- 后端门禁（跨层/命名）：golangci-lint run（depguard + tagliatelle）
- 证据登记：logs/plan255/* 与 reports/architecture/architecture-validation.json；在 215 登记索引与结论
- 组合门禁：确保 plan-253-gates 也为受保护分支必需检查

---

## 15. 变更记录
- v1.3（2025-11-16）：新增 ESLint 架构守卫接入；明确 JWKS 不设永久前端例外（提供临时 DEV_MODE 策略）；统一状态词表为 status/isCurrent/isFuture/isTemporal；组合门禁补充 plan-250；完善验收与证据登记
- v1.2（2025-11-16）：收敛 snake_case 例外到“字段级白名单 + 最小路径”；收紧前端 GET 直连例外为仅 `/auth`；CI 固定 golangci-lint 版本并增加受保护分支证据要求；补充根路径审计步骤（非门禁）
- v1.1（2025-11-16）：补齐范围澄清、执行步骤、例外与 CI 接入、验收标准、对齐/依赖、风险回滚、里程碑与执行清单；对齐 AGENTS 与 253/258
- v1.0（2025-11-15）：初版（目标/交付物/验收）

---

## 16. 执行进展（滚动）
更新时间：2025-11-16

- 已完成（软门禁可启动）
  - 工作流接入：.github/workflows/plan-255-gates.yml 已新增 ESLint 架构守卫（Flat Config：eslint.config.architecture.mjs），与 architecture-validator 组合执行；golangci-lint 固定 v1.59.1
  - 静态守卫增强：architecture-validator 支持跨行“默认 GET（fetch + options 无 method）”检测；忽略 unified-client 底层实现，避免误报
  - 策略收敛：前端 GET 例外仅限 `/auth`；前端不设 JWKS 永久例外，已改用 UnauthenticatedRESTClient 统一出站；状态字段词表统一为 `status/isCurrent/isFuture/isTemporal`
  - 原则与记录：AGENTS 增加“决策建议原则”；本计划 v1.3 与 CHANGELOG 已登记

- 待办/前置（硬门禁切换前必须）
  - 受保护分支：在仓库 Branch protection rules 中将 plan-250-gates、plan-253-gates、plan-255-gates 配置为必需检查；在 215 登记“设置截图 + 失败示例链接”
  - 后端命名收敛：监控/告警导出 JSON 字段 snake_case → camelCase（已完成：internal/monitoring/health/alerting.go；移除临时 nolint；对外字段为 `resolvedAt/maxRetries/enabledBy/statusEquals/responseTimeGt/consecutiveFails`）
  - 首轮证据：CI 首次运行产物（logs/plan255/*、reports/architecture/architecture-validation.json）落盘后，在 215 登记索引

- 风险与缓解
  - 字段变更影响下游消费方：提供一迭代窗口，必要时边缘层最小适配/双写；严格禁止扩散路径白名单

- 下一步（建议顺序）
  1) 完成 Branch protection 设置并在 215 登记
  2) 提交后端 camelCase 迁移 MR，移除对应 `//nolint` 与 TODO
  3) 关闭临时豁免并切换 255 为硬门禁（与 250/253 并列为受保护分支必需检查）
