# 13 — OrganizationForm 验证兼容层收尾方案

生成日期：2025-09-23
责任团队：前端组织域小队（FEO）
优先级：P0（过期临时实现）
当前状态：已完成并通过回归验证

---

## 1. 背景与问题定义
- `docs/development-plans/10-implementation-inventory-maintenance-report.md` 标记 `frontend/src/features/organizations/components/OrganizationForm/ValidationRules.ts` 为 2025-09-16 到期的临时兼容层。
- 当前表单仍从 `./ValidationRules` 引入 `validateForm`，导致过期临时实现继续存在，违背临时方案回收要求。
- `ValidationRules.ts` 仅转调 `shared/validation/schemas.ts` 中的 `ValidationUtils`，说明共享 Schema 已经具备全部能力。

## 2. 现状调研摘要
- `frontend/src/features/organizations/components/OrganizationForm/index.tsx:7` 继续引用 `./ValidationRules`。
- `ValidationRules.ts` 顶部 `// TODO-TEMPORARY` 注释已逾期，并声明改用 `shared/validation/schemas.ts`。
- `ValidationRules` 接口定义在 `FormTypes.ts:32`，用于描述旧版逐字段函数；实际渲染组件未再引用。
- Schema 版本 `shared/validation/schemas.ts` 提供 `validateForm`，但要求可选字段（如 `code`）传入 `undefined` 才能跳过正则校验。

## 3. 风险评估
- 长期保留临时文件会误导审计并拖累临时实现治理。
- 兼容层未移除，后续开发易重复调用旧接口；一旦 Schema 更新，壳层可能与真实逻辑偏离。
- 删除文件后若未补齐数据归一化，空字符串 `code` 会触发 7 位数字校验失败，需在提交前预处理表单数据。

## 4. 处理方案
1. **直接引用共享验证**：在 `OrganizationForm/index.tsx` 改为 `import { validateForm } from '../../../../shared/validation/schemas';`。
2. **提交前归一化**：在调用 `validateForm` 之前，将表单数据副本中的 `code` 为空字符串时置为 `undefined`，保证与 `CreateOrganizationInputSchema` 的可选字段约束一致。可顺带去除额外空白、类型转换等轻量清洗。
3. **移除兼容层文件**：删除 `frontend/src/features/organizations/components/OrganizationForm/ValidationRules.ts`。
4. **同步移除类型定义**：删除 `frontend/src/features/organizations/components/OrganizationForm/FormTypes.ts` 中未再使用的 `ValidationRules` 接口。
5. **验证与回归**：
   - 运行 `cd frontend && npm run test -- OrganizationForm`（或现有覆盖 OrganizationForm 的 Vitest 用例）。
   - 若缺少单测，新增一个验证“空编码创建时通过校验”的单元测试。

## 5. 完成标准
- 代码层面不再存在 `ValidationRules.ts`；相关导入全部指向共享 Schema。
- 表单提交流程能够正确处理空编码、时态字段等原有场景，测试通过。
- 在实现清单与审计报告中更新状态（由 IIG 团队负责），确认临时实现条目已关闭。

---

## 6. 行动项追踪
| 项次 | 动作 | 负责人 | 截止时间 | 状态 |
| ---- | ---- | ------ | -------- | ---- |
| A | 更新 `OrganizationForm/index.tsx` 引用并归一化表单数据 | FEO-前端 | 2025-09-23 | ✅ 已完成 |
| B | 删除 `ValidationRules.ts` 与冗余接口定义 | FEO-前端 | 2025-09-23 | ✅ 已完成 |
| C | 补充/执行相关测试并提交结果 (`npm run test -- OrganizationForm`, 新增单测：`validation.test.ts`) | FEO-前端 | 2025-09-23 | ✅ 已完成 |
| D | 通知 IIG 更新实现清单与守护清单 | FEO-前端 | 2025-09-26 | 待办 |

### 6.1 执行摘要（2025-09-23）
- 组件改为直接调用共享 `validateForm`，并整合归一化工具 `validation.ts`。
- 删除逾期兼容层 `ValidationRules.ts` 与冗余类型定义。
- 新增 `__tests__/validation.test.ts`，验证空编码场景归一化后的校验行为；测试命令：`cd frontend && npm run test -- OrganizationForm`。
- 余下动作：与 IIG 联动更新实现清单状态。

---

备注：执行完成后需同步 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 中的临时实现状态，确保仓库上下一致。
