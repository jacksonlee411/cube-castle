# logs/plan402/mapping

402A 阶段要求将映射、契约与评审证据集中落在该目录。日志类型：

| 文件名 | 内容 | 触发命令 |
|--------|------|----------|
| `spec-review.log` | 《402A 标准对象映射规格》评审会议纪要、结论、责任人 | 手工记录 |
| `dec-gap.log` | `node scripts/quality/architecture-validator.js --rule capabilityContracts` 输出的 DEC/OCL 缺口 JSON | 构契约守卫 |
| `api-contract.log` | `node scripts/quality/contract-checker.js` + `npm --prefix frontend run contract:generate` 结果 | 契约同步 |
| `naming-check.log` | `scripts/quality/architecture-validator.js` 命名守卫摘要 | 守卫运行 |
| `time-constraint.log` | `pkg/temporal/constraints` 或辅助脚本输出的 TC1/TC2/TC3 声明与巡检结果 | 时间约束巡检 |

> 命名、契约及 DEC/OCL 缺口一经确认需同步更新 `docs/development-plans/402A-standard-object-mapping-spec.md` 与 `schema-registry.json`。日志保留 UTC 时间戳并遵循 `AGENTS.md` 的“唯一事实来源”约束。
