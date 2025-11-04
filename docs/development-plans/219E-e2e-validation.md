# Plan 219E – 端到端测试与性能验收

**文档编号**: 219E  
**关联路线图**: Plan 219  
**依赖子计划**: 219A~219D 全部完成  
**目标周期**: Week 5 Day 23-25（Day 26 作为缓冲、对齐 Plan 204 行动 2.9-2.10 后续验收）  
**负责人**: QA 团队 + 后端团队

---

## 1. 目标

1. 执行组织聚合的端到端回归（REST/GraphQL/Temporal/Audit/Validator 等全链路）。
2. 完成性能基准对比（重构前 vs. 重构后），确保 P99 延迟不退化。
3. 验证回退策略（rollback），确保在异常情况下可恢复。

---

## 2. 测试范围

### 2.1 端到端场景（至少覆盖以下 5 组）
1. **Organization + Department 生命周期**：创建 → 子部门 → 更新 → 层级移动 → 删除；验证 REST、GraphQL、Audit。
2. **Position + Assignment 流程**：创建职位 → Fill → Transfer → Vacate → 删除；验证 Temporal timeline 与 Assignment 查询。
3. **Job Catalog 导入/导出**：导入 → 校验版本 → 导出 → 对比 checksum；验证 validator 拦截冲突。
4. **Outbox → Dispatcher → Query 缓存**：更新组织数据，检查 outbox、dispatcher 指标、Query 缓存刷新日志。
5. **故障恢复**：模拟事务失败/dispatcher 失败/Temporal 中断，验证 retry 与报警。

### 2.2 性能测试
- REST P99：重点接口 `/api/v1/organizations`, `/api/v1/positions`, `/api/v1/job-family-groups`。目标：P99 不超过基线 +10%。
- GraphQL P95/P99：`organizations`, `positions`, `assignmentHistory`。目标：不退化。
- 资源消耗：CPU、内存、DB 连接数。

### 2.3 回退验证
- 定义回退步骤清单（恢复旧目录/适配层、切回旧配置）。
- 在测试环境演练一次回退 → 再恢复新结构。

---

## 3. 工具与脚本

- API 测试：Postman collection / curl / scripts in `tests/organization/`。
- GraphQL：GraphQL Playground 脚本或 CLI（`graphql-client`）。
- 性能：`hey` 或 `ab`；必要时使用 k6。
  - 例：`hey -n 1000 -c 20 -H "Authorization: Bearer $JWT" http://localhost:9090/api/v1/organizations`
  - 例：`graphql-client --endpoint=http://localhost:8090/graphql --query-file tests/organization/perf/positions.graphql`
  - 对比脚本：`scripts/perf/compare-benchmark.sh baseline.json current.json`（需输出差异报告）
- Temporal：`tctl workflow describe` 验证 workflow 结果。
- 监控：Prometheus/Grafana 面板（记录 P99、CPU、DB 连接）。

---

## 4. 验收标准

- [ ] 端到端场景执行完毕，结果记录于测试报告。
- [ ] 性能基准与重构前对比，P99 差异在可接受范围内；若超出，提供优化方案或评估。
- [ ] 回退演练完成并记录步骤。
- [ ] 测试脚本/工具更新入库（`tests/organization/` 或相关目录），并在 `internal/organization/README.md` 的“测试与验收”小节记录引用。
- [ ] 发布测试结论（通过/阻塞、剩余风险）。

---

## 5. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 端到端场景覆盖不足 | 高 | 与业务/架构确认清单；回顾历史缺陷 |
| 性能指标退化 | 高 | 预留性能优化时间；收集指标分析瓶颈 |
| 回退流程复杂 | 中 | 提前拟定脚本；演练前确认恢复点 |

---

## 6. 交付物

- 端到端测试报告（包含场景、结果、日志链接，统一归档至 `docs/reference/03-API-AND-TOOLS-GUIDE.md` 性能与验收章节）
- 性能基准报告（对比表 + 图表，同步更新至上述章节）
- 回退演练记录（补充到 `internal/organization/README.md` 的回退小节）
- 阶段结论（供上线/合并决策）
