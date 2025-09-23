# 14 — API 类型临时导出回收方案

文档类型：整改计划
创建日期：2025-09-22
责任团队：前端平台组（主责）＋ 实现清单守护代理（监督）
优先级：P0（已过期临时实现）
当前状态：进行中（临时导出已回收，待报告同步与流程守卫验收）

---

## 1. 背景
- `docs/development-plans/10-implementation-inventory-maintenance-report.md` 已在 2025-09-21 指出 “API 类型临时导出” 逾期未回收。
- 新的统一错误处理体系已落地于 `frontend/src/shared/api/error-handling.ts` 与 `frontend/src/shared/api/type-guards.ts`，但旧入口 `frontend/src/shared/types/api.ts` 仍保留临时别名，违反“单一事实来源”原则。
- 过期的 `// TODO-TEMPORARY` 导致实现清单、CI 报表长期处于红灯状态，阻塞 IIG 关闭该缺口。

---

## 2. 现状评估
### 2.1 临时导出处理记录
- ✅ 2025-09-22（Codex）已删除 `frontend/src/shared/types/api.ts` 中的 `APIError`、`ValidationError`、`isAPIError`、`isValidationError` 临时导出，并新增指引注释。
- ✅ 同步更新 `frontend/src/shared/types/index.ts`，提示错误类型需改从 `shared/api/*` 导入，消除聚合出口扩散风险。

### 2.2 引用分析
- `rg "APIError" frontend/src` 显示主流代码已迁移至 `shared/api/error-handling.ts` 与 `shared/api/type-guards.ts`。
- 未发现仍直接从 `shared/types/api` 导入上述符号的业务代码，但聚合出口存在回归风险，且实现清单与代理报表持续报错。

### 2.3 风险评估
- **契约漂移**：双入口易引入分叉实现，破坏统一错误模型。
- **治理压力**：IIG 报告持续标红，影响迭代验收。
- **复归成本**：后续若再有模块引用旧路径，会在运行时表现不一致，难以及时察觉。

---

## 3. 目标
1. 彻底移除 `shared/types/api.ts` 中与错误处理相关的临时别名导出。
2. 确保所有代码仅通过 `shared/api/error-handling.ts` 与 `shared/api/type-guards.ts` 获取错误类型与守卫。
3. 更新监控与文档，保证 IIG 报告、实现清单与脚本检查恢复绿色状态。

---

## 4. 解决方案
### 4.1 代码层面
1. **临时导出回收（D0）** — ✅ 2025-09-22 完成
   - 删除 `frontend/src/shared/types/api.ts` 中的 `_APIError/_ValidationError` 临时别名导出及 `isAPIError/isValidationError` 透传。
   - 在文件首部增加显式指引，引导开发者改用 `shared/api/error-handling` 与 `shared/api/type-guards`。
2. **聚合出口收敛（D0）**
   - 保留 `APIResponse`、`PaginatedResponse` 等真正的类型定义。
   - 在 `frontend/src/shared/types/index.ts` 中新增说明，提醒错误类型不得从此出口导出；如有必要，可改为只导出类型接口（不再透出运行时守卫）。
3. **引用确认（D0）**
   - 运行 `rg "from '@/shared/types'" -g"*.ts?(x)" frontend/src` 与 `rg "from '../types/api'" frontend/src`，确认不存在对已删除符号的引用；若发现遗留，逐一改为从 `shared/api/error-handling` 或 `shared/api/type-guards` 导入。
4. **实现清单同步（D0）**
   - 更新 `reports/implementation-inventory.json` 与 `reports/iig-guardian/iig-guardian-report.json`，移除对应临时实现条目。
   - 若自动生成，执行生成脚本或记录待批处理任务。

### 4.2 流程与守卫
1. **CI 守卫（D1）**
   - 修改 `scripts/check-temporary-tags.sh`（责任人：Codex）：在遍历结果中一旦检测到 `frontend/src/shared/types/api.ts` 仍存在 `TODO-TEMPORARY`，立即打印错误并返回非零退出码；同步在脚本输出中附带指导信息，提示改用 `shared/api/error-handling`。
   - 在 `frontend/scripts/validate-field-naming*.js` 或新增脚本中检查 `shared/types` 是否导出运行时守卫函数，避免类型/运行时混用。
2. **知识同步（D1）**
   - 在 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 的错误处理章节记录入口调整。
   - 在团队例会或公告中提醒“错误处理 = error-handling, 类型守卫 = type-guards”。

---

## 5. 里程碑与负责
| 里程碑 | 截止日期 | 负责人 | 验收标准 |
| --- | --- | --- | --- |
| D0-代码清理 | 2025-09-23 | 前端平台组 | ✅ 2025-09-22 完成：`shared/types/api.ts` 无临时导出，所有引用已替换 |
| D0-报告同步 | 2025-09-24 | IIG | 实现清单与 IIG 报告移除对应告警 |
| D1-流程守卫 | 2025-09-27 | 前端平台组 + DevOps（执行人：Codex） | `scripts/check-temporary-tags.sh` 启用专门检测并已在 CI 中验证可阻断回归 |

---

## 6. 验证步骤
1. `cd frontend && npm run lint && npm run test`。
2. 运行 `frontend/scripts/validate-field-naming.js` 与 `node scripts/quality/architecture-validator.js`，确保无新增告警。
3. 手动执行 `rg "APIError" frontend/src`，确认仅存在于 `shared/api/*` 与测试中。
4. 执行 `scripts/check-temporary-tags.sh`，验证脚本会在检测到 `shared/types/api.ts` 残留临时导出时立即失败。
5. 如果前端提供 Storybook / 开发环境，执行一次主要表单提交流程，确认错误处理 UI 行为未变。

---

## 7. 风险与应对
- **隐藏引用遗漏**：若某特性分支仍引用旧路径，可在 CI 守卫中加入“运行时导出”检测并在 MR 评审清单中强调。
- **类型/运行时耦合**：确保 `shared/types` 仅导出 TypeScript 类型；若确需运行时守卫，统一迁移到 `shared/api/type-guards`。
- **下游文档延迟更新**：提前通知文档维护者，避免本次清理后 Reference 文档出现过期示例。

---

## 8. 完成判定
- `frontend/src/shared/types/api.ts` 中不再存在 `TODO-TEMPORARY` 与错误相关的临时导出。
- 所有错误处理导入均指向 `shared/api/error-handling` 或 `shared/api/type-guards`。
- 实现清单、IIG 报告及相关 CI 检查均恢复绿色状态。
