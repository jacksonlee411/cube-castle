# Plan 222E – 文档同步与关单收口

编号: 222E  
上游: Plan 222（验证与文档更新）；依赖 222A-D 完成本体任务  
状态: 草案（待 222A-D 完成后启动）

---

## 目标
- 将 Plan 222 的所有验收项与证据同步到文档/报告/索引，确保唯一事实来源一致。
- 更新 `docs/archive/development-plans/222-organization-verification.md`、`reports/phase2-execution-report.md`、`docs/development-plans/215-phase2-execution-log.md` 中的勾选状态。
- 输出最终验收结论（从阶段性通过 → ✅ 完成），并记录回滚/风险状态。

## 范围
- 文档：`docs/archive/development-plans/222-organization-verification.md`、`reports/phase2-execution-report.md`、`docs/development-plans/215-phase2-execution-log.md`、`docs/reference/02-IMPLEMENTATION-INVENTORY.md`（若需）。  
- 索引/摘要：`docs/development-plans/HRMS-DOCUMENTATION-INDEX.md`, `logs/plan222/ACCEPTANCE-SUMMARY-*.md`。
- 变更只涉及文档与报告，不改代码。

## 任务清单
1) 证据核对  
   - 汇总 222A-D 落盘的日志/报告；更新 `logs/plan222/ACCEPTANCE-SUMMARY-*.md`。  
   - 确认所有勾选项有对应证据链接（health/jwks/REST/GraphQL/coverage/perf/E2E）。
2) 文档更新  
   - `../archive/development-plans/222-organization-verification.md`：补充最终覆盖率/性能/E2E 结果，勾选剩余任务。  
   - `reports/phase2-execution-report.md`：将 Plan 222 状态更新为 ✅，同步核心结论。  
   - `215-phase2-execution-log.md`：勾选相关 checklist（测试、文档）。  
   - 若 implementation inventory 需要更新“Plan 222 完成”，同步 `docs/reference/02-IMPLEMENTATION-INVENTORY.md`。
3) 索引与档案  
   - `HRMS-DOCUMENTATION-INDEX.md` 中 Plan 222 状态更新。  
   - 确保 `docs/archive/development-plans/222-closure-pr.md` 与本子计划一致（如需 PR 文案）。
4) 验收声明  
   - 在 `reports/phase2-execution-report.md` “风险”部分确认无未解事项或列出 residual risks。  
   - 输出最终验收结论（包含回滚路径、日志索引）。

## 验收标准
- 上述文档全部更新并引用最新证据；状态从 ⏳ 转为 ✅。  
- 相关勾选项（REST 契约、GraphQL 契约、E2E Live、性能、覆盖率）均在文档中勾选。  
- 有统一的关单摘要（`logs/plan222/ACCEPTANCE-SUMMARY-final.md` 或同类）。  
- 无新的临时项（TODO-TEMPORARY）遗留；若存在，列出回收计划。

## 产物与落盘
- 更新后的文档（见上）  
- `logs/plan222/ACCEPTANCE-SUMMARY-final.md`  
- 若开 PR，附 `222-closure-pr.md` 的更新内容。

## 回滚策略
- 文档改动如需回滚，可按文件粒度 revert；确保不会覆盖他人未合并的文档修改。  
- 若发现证据缺失，应先补齐日志再更新勾选，避免“先勾选后补证据”的不一致。

---

维护者: Codex（AI 助手）  
目标完成: 222A-D 结束后 Day 1 内完成  
最后更新: 2025-11-16 (草案)
