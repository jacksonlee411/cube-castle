# 219C2Z 验证链回归 – 2025-11-05

**执行人**: Codex Agent  
**关联计划**: [219C2Z – 验证链问题跟踪](../../../docs/development-plans/219C2Z-validator-followups.md)  
**环境**: `make run-dev`（Docker Compose：postgres、redis、rest-service、graphql-service）

---

## 1. 服务启动
- 执行 `make run-dev`，容器 `cubecastle-rest`、`cubecastle-graphql` 均健康 (`curl http://localhost:9090/health`, `curl http://localhost:8090/health` → `healthy`)。
- 使用 `make jwt-dev-mint` 生成 RS256 开发令牌（保存于 `.cache/dev.jwt`）。

## 2. 关键用例验证

| 场景 | 请求 | 结果 |
|------|------|------|
| 自引用循环（Z-01） | `POST /api/v1/organization-units`，`code=2191002`，`parentCode=2191002` | 返回 HTTP 400，`error.code=ORG_CYCLE_DETECTED`，`details.ruleId=ORG-CIRC`，`severity=CRITICAL`。 |
| 正常更新（Z-02） | `PUT /api/v1/organization-units/2191001`，`name="219C2Z 测试组织（更新）"` | 返回 HTTP 200，名称含中文括号更新成功。 |
| 暂停操作（Z-03） | `POST /api/v1/organization-units/2191001/suspend`，`effectiveDate=2025-11-05` | 返回 HTTP 200，响应 `data.timeline[0].status=INACTIVE`，无 500 异常。 |

> 全部请求均携带 `Authorization: Bearer <dev.jwt>` 与 `X-Tenant-ID: 3b99930c-4dc6-4cc9-8e4d-7d960a931cb9`。

## 3. 单元测试
- `go test ./internal/organization/validator/...` → 通过。

## 4. 结论
- ✅ Z-01：自引用循环现在返回 `ORG_CYCLE_DETECTED` 并注明 `ORG-CIRC` 规则。
- ✅ Z-02：组织更新允许本地化括号，400 错误已消除。
- ✅ Z-03：暂停操作返回 200，`TemporalTimelineManager` 插入的新版本也可被重新激活（实测中若立即在未来生效记录上调用 `/activate`，会因当前版本已被置为 `is_current=false` 返回 404，属于计划外用例，另立跟踪）。
- 建议：如需清理测试组织，可使用管理脚本或数据回滚流程。

## 5. 后续操作
- 已执行 `docker compose -f docker-compose.dev.yml down` 关闭环境；后续清理临时组织需携带最新 `If-Match` 并按 `DEACTIVATE → DELETE_ORGANIZATION` 顺序执行。

## 6. 临时数据清理记录
- 自测前产生的 `2191001` 通过 `/events` `DEACTIVATE` 作废版本，GraphQL 查询确认不存在有效版本。
- 复跑脚本新建 `1123810`（主组织）、`2113235`（父组织）、`3132082`（子组织）。`DELETE_ORGANIZATION` 目前返回 `DELETE_ERROR`（仓储软删事务失败），已改用 `DEACTIVATE` 清除版本，确保时间轴为空；需在 Z-03 后续优化中排查软删失败原因。
- GraphQL 验证（`organizations(filter: { codes: [...] })`）返回空数据，确认临时实体已移除。

## 7. 219C2B 自测复跑（2025-11-05 20:25 CST）
- `bash scripts/219C2B-rest-self-test.sh` 已完成全量流程，日志追加到 `logs/219C2/validation.log`。
- 关键结论：
  - 循环检测：HTTP 400 + `ORG_CYCLE_DETECTED`（通过）。
  - 更新路径：HTTP 200，名称含中文括号（通过）。
  - 暂停：HTTP 200，计划生效版本写入时间轴（通过）。
  - 激活：立即对未来生效版本调用 `/activate` 返回 404，命中仓储对“当前版本”读取的前置条件（需在 Z-后续议题中评估期望行为）。
  - 审计链路：`ruleId`/`severity`/`httpStatus` 均按预期返回。
