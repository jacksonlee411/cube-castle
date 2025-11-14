# Plan 244 – Temporal Timeline & Status 抽象执行记录（Go 契约同步）

date: 2025-11-11
window: Day 7
status: completed

## delta
- TemporalTimelineManager 现在输出 `unitType/level/codePath/namePath/sortOrder` 等 `TemporalEntityTimelineVersion` 字段，`organization_update`/`organization_events` 响应结构同步，REST timeline 与前端 `TemporalEntityTimelineAdapter` 完全对齐。
- OpenAPI `organization-units` 事件响应文档补充 `codePath/namePath/sortOrder` 字段说明，Implementation Inventory 与命名清单登记 Go 层事实来源。
- 运行 `node scripts/generate-implementation-inventory.js` 触发最新快照；Plan 242 日志与 docs/reference 更新反向引用。

## verification
- `go test ./cmd/hrms-server/...`
- `make test`
- `npm run lint`
- `cd frontend && npm run test`
- `node scripts/quality/architecture-validator.js`
- `node scripts/generate-implementation-inventory.js`

## notes
- 根目录 `package.json` 未定义 `test` 脚本，执行 `npm run test` 报错（见 shell 日志）；前端测试改在 `frontend/` 目录运行 Vitest 以满足 Plan 244 的验证要求。
- Playwright `npm run test:e2e -- --project=chromium --project=firefox` 依赖完整浏览器与后端环境，当前容器未安装 Playwright；请在本地工作站或 CI 阶段执行三轮连跑以完成 Plan 244 的 E2E 验证。
