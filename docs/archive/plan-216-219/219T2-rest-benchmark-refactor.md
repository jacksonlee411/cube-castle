# 219T2 – REST 性能脚本重构子方案

## 1. 背景
- 现行 `scripts/perf/rest-benchmark.sh` 对 `/api/v1/organization-units` 发送固定 payload 与常量 `Idempotency-Key`，导致 99% 请求命中 429/409，无法反映真实延迟。
- 219T 报告建议重写脚本，增加随机化与退避逻辑，并记录新的 P50/P95/P99 作为 219E 绩效指标。

## 2. 目标
1. 让性能脚本可配置 idempotency key 与 payload 生成策略，避免重复请求。
2. 引入速率控制/退避（例如批次 sleep 或并发上限），以便测量成功请求的延迟分布。
3. 输出结构化日志（JSON 或 CSV），便于后续比较。

## 3. 工作包
| 编号 | 任务 | 说明 |
| --- | --- | --- |
| P1 | 脚本重构 | 使用 `jq`/`python` 生成唯一 payload（随机组织名、时间），并为每次请求添加 `X-Idempotency-Key=$(uuid)` |
| P2 | 速率控制 | 支持 `RATE_LIMIT` 配置；对 `hey` 或自研脚本添加 `sleep`/burst 限制，减少 429 |
| P3 | 结果解析 | 将 `hey` 输出转换为 JSON，记录 P50/P95/P99、成功率、429 比例 |
| P4 | 文档更新 | 在 `docs/development-plans/219T-e2e-validation-report.md` 与 `docs/reference/03-API-AND-TOOLS-GUIDE.md` 登记新脚本用法 |

## 4. 依赖
- `go install github.com/rakyll/hey@latest` 已可用。
- 需要可写 `logs/219E/` 目录以存放新日志。

## 5. 验收标准
1. 在默认配置下，脚本成功率 ≥90%，P95/P99 可计算。
2. `logs/219E/perf-rest-*.log` 新版包含 JSON summary。
3. 219T 报告更新性能章节，附新数据。

## 6. 进展纪要（2025-11-07）
- `scripts/perf/rest-benchmark.sh` 默认切换为 Node 驱动，支持 `REQUEST_COUNT`、`THROTTLE_DELAY_MS`、`REQUEST_TIMEOUT_MS`、`NAME_PREFIX`、`IDEMPOTENCY_PREFIX` 等参数，并为每次请求生成唯一 `code` 与 `X-Idempotency-Key`。
- 运行示例：`LOAD_DRIVER=node REQUEST_COUNT=30 CONCURRENCY=5 THROTTLE_DELAY_MS=50 scripts/perf/rest-benchmark.sh`，日志 `logs/219E/perf-rest-20251107-091914.log`，成功率 100%，P95 ≈ 123ms。
- `docs/reference/03-API-AND-TOOLS-GUIDE.md` 与 219T 报告已更新使用说明及场景 C 数据；若需回退旧行为，可设置 `LOAD_DRIVER=hey`。
- 新增 `scripts/diagnostics/check-rest-benchmark-summary.sh` 并在 `make status` 中展示最新 JSON Summary，CI 可通过 `STRICT=1` 强制校验成功率阈值。

---

> 唯一事实来源：`scripts/perf/rest-benchmark.sh`、`logs/219E/`。  
> 更新时间：2025-11-07。
