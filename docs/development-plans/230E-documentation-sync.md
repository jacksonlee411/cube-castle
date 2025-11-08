# Plan 230E – 219T/219E 文档与报告同步

**文档编号**: 230E  
**母计划**: Plan 230 – Position CRUD 参考数据修复计划  
**前置计划**: 230B（修复完成）、230D（E2E 通过）  
**负责人**: QA 团队 + 文档维护者

---

## 1. 背景与目标

- 219T 报告与 219E 计划中，Position CRUD 条目目前标记为阻塞。  
- 完成 230B/230D 后，需要及时在 `docs/development-plans/219T-e2e-validation-report.md`、`docs/development-plans/219E-e2e-validation.md`、`docs/development-plans/06-integrated-teams-progress-log.md` 等文档中更新状态，确保唯一事实来源一致。  
- 230E 专注文档与证据回填，避免信息漂移。

---

## 2. 范围

1. 更新 `docs/development-plans/219T-e2e-validation-report.md`：  
   - Position CRUD Section 添加修复脚本版本、Playwright 执行时间、产物路径、RequestId。  
   - 标记阻塞解除日期，引用 `logs/230/position-crud-playwright-*.log`。  
2. 更新 `docs/development-plans/219E-e2e-validation.md`：  
   - 在 2.4 “重启前置条件”表中，将 “Position + Assignment 数据链路恢复” 条目改为 ✅，并附证据链接。  
   - 在 2.5 P0 Playwright 阻塞列表中关闭与 Job Catalog 数据相关的项。  
3. 在 `docs/development-plans/06-integrated-teams-progress-log.md` 增加一条更新记录，说明 230 子计划完成及依赖关系。  
4. 若有额外报告（如 `frontend/test-results/README`、`logs/219E/*.log` 索引），同步更新引用。

---

## 3. 任务

| 步骤 | 描述 | 输出 |
| --- | --- | --- |
| E1 | 收集 230D 产物路径、RequestId、时间戳 | 资料清单 |
| E2 | 修改 219T 报告相应段落，确保引用链接可点击（使用仓库相对路径） | `docs/development-plans/219T-e2e-validation-report.md` diff |
| E3 | 更新 219E 计划的状态表与 Playwright 阻塞列表，注明“数据缺口已由 230D 解除” | `docs/development-plans/219E-e2e-validation.md` diff |
| E4 | 在 `06-integrated-teams-progress-log.md` 新增记录，包含时间、完成的子计划、证据路径 | 日志条目 |

---

## 4. 依赖

- 230D 已提供可引用的日志/产物。  
- 需要访问 219T/219E 文档的最新版本。  
- 所有引用的日志路径必须真实存在于仓库（或在 PR 中同步添加）。

---

## 5. 验收标准

1. 三份文档（219T、219E、06-log）均更新，并通过 Review；其中 Position CRUD 相关条目引用新的日志与测试目录。  
2. 文档中不再出现 “阻塞: Job Catalog 缺失” 等旧描述，取而代之的是“已由 Plan 230 完成”。  
3. 每处引用都遵循唯一事实来源：日志路径或测试目录必须存在且与 230D 输出一致。  
4. 若有未完成事项（例如 Firefox 仍失败），需在文档中注明后续计划，不得留空。

---

> 唯一事实来源：`logs/230/position-crud-playwright-*.log`、`frontend/test-results/position-crud-full-lifecyc-<commit>-chromium/`。  
> 更新时间：2025-11-07。
