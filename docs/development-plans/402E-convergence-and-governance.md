# 402E · 收敛与治理（可选）

**关联计划**：Plan 400、Plan 401、Plan 215、Plan 272、Plan 403  
**状态**：待启动（依赖 402D 完成并稳定运行至少一个迭代）  
**范围**：旧表/代码清理、文档与索引更新、监控/OBS 完善、未来模块接入指南  
**日志要求**：`logs/plan402/cleanup/*.log`、`logs/plan402/capability/*.log`（更新记录）、`logs/plan402/metrics/*.log`

> 目标：在 SOM 切换稳定后，清理历史实现、完善文档/监控/质量守卫，并沉淀未来模块接入指南。

---

## 1. 目标

1. 清除 `organization_units` 等旧表、触发器与遗留仓储，只保留受控视图或归档。
2. 更新参考文档（Plan 400/401、开发者速查、Plan 00-README 等），确保 SOM 成为唯一事实来源。
3. 完善监控/OBS/质量守卫，例如 capability-contract-check、schema hash 校验、Time Constraint（TC1/TC2/TC3）与事务时间（transaction_lag）漂移告警。
4. 产出《SOM 接入指南》，指导 payroll/workforce 等未来模块复用 Standard Object 并正确处理双时态。

---

## 2. 工作项

-### E1 · 数据与代码清理
- 删除/归档旧表及触发器（必要时提供 Goose Down 脚本），保留只读视图以兼容历史查询；清理任何 residual `effective_from/effective_to` 字段。
- 清理 `internal/organization` 等目录下不再使用的仓储/DTO，并确保所有生产代码仅调用 `internal/standardobject` Port，且读取/写入均使用 `validity_range/transaction_range` API。
- 更新 `database/migrations`，附执行日志 `logs/plan402/cleanup/*.log`。

### E2 · 文档与索引
- 更新 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`、Plan 400/401、Plan 403 等文档，声明 SOM 为唯一事实来源。
- 在 `docs/development-plans/00-README.md`、`docs/reference/standard-object-evidence-guide.md` 等索引中添加最新链接。
- 将 Plan 402 主文档与子计划按规范归档到 `docs/archive/development-plans/`，保留验收标准与证据索引。
- 发布《能力契约与视点维护指南》，并在 `scripts/quality/capability-contract-check.js` 或架构验证器中纳入检查。

### E3 · 监控 / OBS / 接入指南
- 为 SOM 相关操作配置指标与告警，将日志采集纳入 Plan 272 的运行产物治理脚本；指标需包含 `transaction_lag`、双时态回放耗时、TC1/TC2 违规计数。
- 编写《SOM 接入指南》，说明如何使用 Schema Registry、Port、Manifest/Slot、能力契约，并给出双时态（valid/transaction）参数使用规范与回滚/日志要求。
- 在 `scripts/quality/` 中新增守卫（如验证仓库不再引用旧表、Schema hash 校验、capability contract 检查、`timeConstraint`/`transaction_range` 监控脚本），确保 TC1/TC2 违规和事务时间倒退能被 CI 阻断。

---

## 3. 交付物

- 数据/代码清理脚本、执行日志 (`logs/plan402/cleanup/*.log`)。
- 更新后的 Plan 400/401/403、开发者速查、参考指南，以及归档版本。
- 监控/OBS 配置（含 Time Constraint 违规告警）、SOM 接入指南、质量守卫脚本。
- `logs/plan402/capability/*.log` 中的维护记录。

---

## 4. 验收标准

1. 生产代码中不再存在对 `organization_units` 的直接引用（除只读视图或归档），CI 守卫能够检测该情况。
2. 文档索引与开发者速查全部指向 SOM，并在 `docs/development-plans/00-README.md` 登记。
3. 监控/OBS 指标上线，Plan 272 的运行产物治理脚本记录最新产物（含 `transaction_lag`、TC 漂移、双时态回放指标）。
4. 质量守卫（capability contract、schema hash、Time Constraint、`transaction_range`、旧表引用检测等）可阻止回退到旧实现。

---

## 5. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 遗留引用遗漏 | 旧逻辑悄然复活 | 在 `scripts/quality/architecture-validator.js` 增加检查；运行 `rg organization_units` 并制定白名单 |
| 文档不同步 | 团队引用旧资料 | 建立文档清单与审核流程，更新 `docs/reference/standard-object-evidence-guide.md` |
| 监控/守卫缺失 | 回归无法及时发现 | 在 Plan 272/255 守卫中加入 SOM 专项脚本，强制在 PR/CI 运行 |

---

402E 完成后，Plan 402 可归档；后续模块接入需直接引用 Plan 400/401/403、能力契约与证据指南。
