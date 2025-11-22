# Plan 222A – 覆盖率提档与稳健性验证

编号: 222A  
上游: Plan 222（organization 验证与文档更新）  
依赖: Plan 219 完成（重构已合并）；与 Plan 232 无直接依赖  
状态: 草案（待启动）

---

## 目标
- 将 organization 模块组合覆盖率从 ~31% 提升：阶段目标≥55%，终态≥80%，保持顶层关键包 ≥80% 不回退。
- 跑通 `-race` 全量与基础内存泄漏检查，确保无数据竞争与泄漏迹象。
- 保持现有功能/契约不变，新增用例仅覆盖已实现逻辑与负路径。

## 范围
- 代码：`internal/organization/**`（repository/service/handler/cascade/devtools 等），不改契约。
- 工具与脚本：复用 `scripts/plan222/collect-coverage-org.sh` 采集覆盖率；必要时补充整洁的 helper/fixture。
- 证据落盘：`logs/plan222/coverage-org-*.{out,txt,html}`、`logs/plan222/race-org-*.log`。

## 不做
- 不引入新的接口/契约变更。
- 不调整 Docker 端口映射或基础设施配置（遵循 AGENTS.md）。

## 任务清单
1) 执行流程（每轮固定顺序）  
   - （可选）`make test-db-up` 以启用本地依赖；完成后 `make test-db-down`。  
   - `bash scripts/plan222/collect-coverage-org.sh` → 生成 `coverage-org-*.{out,txt,html}`。  
   - `go test -v -race ./internal/organization/... | tee logs/plan222/race-org-$(date +%Y%m%d-%H%M%S).log`。  
   - 如需内存观察：`GODEBUG=madvdontneed=1 MALLOC_TRACE=1 bash scripts/perf/rest-benchmark.sh`（小样本）后比对 RSS；记录到 `logs/plan222/memory-check-*.md`。  
   - 若出现错误/数据竞争/异常增长，先修复再继续下一步。
2) 覆盖率提升迭代  
   - 优先补齐 repository/service/handler 高频路径与错误分支：  
     - repository: `organizations_list.go`（filters/scan 错误）、`hierarchy_repository.go`（UpdateHierarchyPaths/hierarchy build）、`organization_repository.go`（GetOrganization*/stats/subtree/positions）。  
     - service: `cascade` 生命周期 start/stop、幂等/错误分支。  
     - handler: devtools `/dev/database-status`、认证头解析（tenant/operator/If-Match）。  
   - 阶段门与动作：  
     - ≥40%：首轮完成，若未达标，产出 gap 列表（函数+用例思路）到 `logs/plan222/coverage-gap-*.md`。  
     - ≥55%：阶段目标；未达标需更新 gap 列表并排期下一轮。  
     - ≥70%、≥80%：冲刺与终态，未达标同样落盘 gap 并调整优先级。
3) 稳健性验证  
   - `-race` 每轮覆盖率执行后跑一次；发现数据竞争需修复并复验，若无法立刻修复，按 `// TODO-TEMPORARY(YYYY-MM-DD)` 标注且给出回收日期（≤1迭代）。  
   - 内存观察：基于短压测/基准的 RSS 快照，出现异常需记录原因与处置（如缓存清理、连接池上限调整）。
4) 回归守护  
   - 新增用例需覆盖负路径（错误码、事务回滚、空结果等），避免只测正向。  
   - 覆盖率报告与 `222-organization-verification.md` 覆盖率章节同步更新，仅在证据落盘且指标达成后勾选。

## 验收标准
- 组合覆盖率报告（以最新 `coverage-org-YYYYMMDD-HHMMSS.txt` 标注的样本）达到 ≥55%（阶段完成），终态 ≥80%；顶层关键包持续 ≥80%。  
- `-race` 日志最新一份无数据竞争；若曾出现，已修复且复验记录清晰。  
- 内存检查记录无异常增长或已有处置说明。  
- gap 列表（如有）与处置计划已落盘；`222-organization-verification.md` 已同步勾选并引用对应时间戳的报告。

## 产物与落盘
- 覆盖率：`logs/plan222/coverage-org-*.{out,txt,html}`（最新：`coverage-org-20251118-001054`，33.9%，差距登记：`logs/plan222/coverage-gap-20251118-000105.md`）
- 稳健性：`logs/plan222/race-org-*.log`、必要的内存观察记录（同目录）
- 文档同步：`docs/archive/development-plans/222-organization-verification.md` 覆盖率章节更新

## 最新进展（2025-11-18）
- ✅ 新增单测：`internal/organization/events/outbox_test.go`、`internal/organization/dto/scalars_test.go`、`internal/organization/service/position_service_more_test.go`，覆盖事件 payload、GraphQL 标量与 PositionService 辅助方法。
- ✅ repository sqlmock 覆盖：`position_repository_more_test.go`、`position_assignment_repository_more_test.go`，覆盖查询/插入/关闭任职分支。
- ✅ `go test -race ./internal/organization/...` 通过；证据：命令输出（2025-11-18 00:11 UTC）。
- 🔁 当前组合覆盖率 33.9%（`coverage-org-20251118-001054.txt`），仍低于 40% 阶段门，差距与后续切入点记录于 `logs/plan222/coverage-gap-20251118-000105.md`。

## 回滚/安全
- 仅新增测试与轻量 helper，无运行时行为改动；如用例引入脆弱依赖，可按文件粒度 revert 新增测试并保留覆盖率记录，避免影响主功能。

---

维护者: Codex（AI 助手）  
目标完成: 阶段目标 Day 2，终态 Day 4（相对 222 收口节奏）  
最后更新: 2025-11-16 (草案)
