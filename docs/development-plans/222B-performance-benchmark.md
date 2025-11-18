# Plan 222B – 性能完整基准复跑

编号: 222B  
上游: Plan 222（验证与文档更新）  
依赖: Docker 基座（`make run-dev`）可用；无 232 依赖  
状态: 草案（待启动）

---

## 目标
- 依据 Plan 222 性能门槛，完成组织模块 REST 基准的完整复跑，覆盖查询/列表/创建与并发/限流场景。
- 产出可审计的 JSON Summary，并对比 Phase1/前一轮样本确认无退化。

## 范围
- 场景脚本：`scripts/perf/rest-benchmark.sh`（或等价 go/Node 驱动基准），目标接口限于 organization 相关 REST。
- 运行环境：Docker Compose（`make run-dev`），端口与配置按 SSoT（禁止修改端口映射）。
- 日志落盘：`logs/219E/perf-rest-*.log`（沿用现有目录，保持可追溯）。

## 任务清单
1) 基准参数确定  
   - 请求类型：单个查询、列表（分页）、创建。  
   - 指标门槛（参考 Plan 222）：单查 P99<50ms；列表 P99<200ms；创建 P99<100ms；并发100吞吐>100 req/s；记录速率限制行为。  
   - 配置：concurrency/requests 毫秒级超时统一，禁用随机端口/绕过限流。
2) 执行与落盘  
   - 启动环境：`make run-dev`（含数据库迁移）。  
   - 运行基准：`scripts/perf/rest-benchmark.sh`（传入 concurrency/requests 参数并写 LOG）。  
   - 落盘：`logs/219E/perf-rest-YYYYMMDD-HHMMSS.log`，记录输入参数、成功/失败计数、p95/p99、限流比例。
3) 结果分析  
   - 对比上一轮样本（如 `perf-rest-20251116-094327.log`, `perf-rest-20251116-100552.log`），标注是否达标/无退化。  
   - 如不达标，登记瓶颈与回滚方案（不在本子计划直接优化业务逻辑，可开后续计划）。
4) 文档同步  
   - 在 `222-organization-verification.md` 性能章节更新结论与日志索引。  
   - 如有风险，记录到 `reports/phase2-execution-report.md` 风险区。

## 验收标准
- 至少一轮完整基准执行并落盘 JSON Summary（含参数、分位值、限流/错误分布）。
- 指标达到 Plan 222 门槛或明确说明差异与后续处置。
- 文档更新完成并引用最新日志。

## 产物与落盘
- 基准日志：`logs/219E/perf-rest-*.log`（新增样本）
- 文档：`docs/development-plans/222-organization-verification.md` 性能章节更新；必要时更新 `reports/phase2-execution-report.md`

## 安全与回滚
- 仅运行基准与分析；不调整端口映射或宿主服务。  
- 若基准影响本地数据，可在基准前后执行 `make db-rollback-last && make db-migrate-all`（需遵守迁移真源原则）。

---

维护者: Codex（AI 助手）  
目标完成: Day 2（相对 222 收口节奏）  
最后更新: 2025-11-16 (草案)
