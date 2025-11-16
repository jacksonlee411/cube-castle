# refactor(health-alerting): migrate JSON tags to camelCase and harden Plan 255 gates

## What / Why
- 将监控/告警导出 JSON 字段统一为 camelCase，消除 Plan 255 的命名违规，准备从“软门禁”切换为“硬门禁”。
- 与 AGENTS.md 的“对外响应 camelCase”与“临时例外限期回收”一致，避免路径白名单扩张。

## Changes
- JSON 字段（snake_case → camelCase）：
  - `resolved_at` → `resolvedAt`
  - `max_retries` → `maxRetries`
  - `enabled_by` → `enabledBy`
  - `status_equals` → `statusEquals`
  - `response_time_gt` → `responseTimeGt`
  - `consecutive_fails` → `consecutiveFails`
- 移除此前 `//nolint:tagliatelle` 与 `// TODO-TEMPORARY` 注记（监控/告警导出路径）。
- 接入与强化 Plan 255 门禁：
  - .github/workflows/plan-255-gates.yml 新增 ESLint 架构守卫（Flat Config：`eslint.config.architecture.mjs`）
  - 自研 architecture-validator 增强：跨行默认 GET 检测；忽略 `unified-client.ts` 底层实现，避免误报
  - 新增根路径端口/禁用端点“审计步骤”（非门禁，不阻断）
- 词表统一：状态字段在 ESLint/arch-validator 统一为 `status / isCurrent / isFuture / isTemporal`。
- 前端策略：GET 例外仅 `/auth`；不设 JWKS 永久前端例外（DEV+auth 模块可临时，已改用 `UnauthenticatedRESTClient` 出站）。

## Compatibility
- 外部消费者（如 Webhook 接收端）若依赖 snake_case，需要在一个迭代内完成字段映射调整。
  - 建议：接收端增加 camelCase 解析，窗口期内支持旧字段，窗口结束后移除旧字段支持。
  - 本仓库不回退对外字段为 snake_case。

## Gates & Evidence
- 前端门禁（软）：
  - ESLint 架构守卫（AST）+ architecture-validator（启发式）均为 0 关键违规（本地证据）
  - `logs/plan255/architecture-validator-20251116_101740.log`
  - `reports/architecture/architecture-validation.json`
- 根路径审计（非门禁）：
  - `logs/plan255/audit-root-20251116_102250.log`（发现 37 个端口硬编码与 1 个禁用端点模式；作为问题清单分批收敛）
- 后端门禁：
  - golangci-lint（CI 固定 v1.59.1）；tagliatelle 在监控/告警导出路径不再报错

## Alignment
- Plan 202：保持“命令=REST、查询=GraphQL”主策略；255 负责行为与命名守卫
- Plan 250/253/254：合流与部署门禁引用；禁止直连 9090/8090，前端走单基址代理
- Plan 258：字段裁决以契约与漂移校验为准；255 不与契约事实来源冲突

## Next Steps
1) 在仓库 Branch protection rules 开启 required checks：`plan-250-gates`、`plan-253-gates`、`plan-255-gates`；在 215 登记“设置截图 + 失败示例链接”。
2) 触发一次 CI，拿到 plan-255 日志工件并在 215 登记索引。
3) 按根路径审计日志将硬编码端口/禁用端点问题分批修复（非门禁）。
4) 一迭代内切换 Plan 255 为“硬门禁”，关闭临时豁免（与 250/253 并列为受保护分支必过）。

## Checklist
- [ ] 分支保护 required checks（250/253/255）已开启并登记证据
- [ ] 外部消费者确认 camelCase 兼容窗口（一个迭代）
- [ ] 首轮 CI 工件登记（logs/plan255/*, reports/architecture/…）
- [ ] 根路径审计问题建立 issue 列表并分批收敛

## Files
- internal/monitoring/health/alerting.go（JSON tag 迁移）
- .github/workflows/plan-255-gates.yml（ESLint 守卫 + 根路径审计）
- eslint.config.architecture.mjs / scripts/quality/architecture-validator.js（规则与检测增强）
- docs/development-plans/255-cqrs-separation-enforcement.md（v1.3 进展）
- docs/development-plans/215-phase2-execution-log.md（证据与待办登记）
- CHANGELOG.md（记录工具固定与字段迁移）
