# 13 — temporalValidation.ts 迁移执行方案

生成日期：2025-09-21
责任团队：前端架构组（主责）+ 时态体验小组
问题等级：紧急（P0）- 临时实现已过期
当前状态：方案待确认

---

## 1. 背景与现状
- `frontend/src/features/temporal/utils/temporalValidation.ts` 自 2025-09-16 起标记过期，仍是 `TemporalDatePicker` 等组件的默认校验入口（参见 `frontend/src/features/temporal/components/TemporalDatePicker.tsx:4`）。
- `shared/utils/temporal-converter.ts` 已提供覆盖性更强的统一 API，但缺乏自动迁移方案，导致遗留模块无法完成替换（参见 `frontend/src/shared/utils/temporal-converter.ts:1`）。
- `frontend/src/shared/validation/schemas.ts:189` 已声明替代路径，`frontend/src/shared/utils/index.ts:82` 仍对旧文件进行别名导出，拖慢迁移关闭节奏。
- 当前临时文件仅做薄封装，无法充分利用新工具类的异常处理与数据标准化能力，持续增加维护成本和合规风险。

---

## 2. 目标与验收标准
**总体目标**：在一次短迭代内完成所有直接引用的迁移与回归验证，随后安全删除 `temporalValidation.ts`。

| 序号 | 验收项 | 说明 |
|------|--------|------|
| G1 | 所有前端引用切换至 `shared/utils/temporal-converter.ts` 或 `shared/validation/schemas.ts` | 通过脚本 + 手动审查验证 `rg 'temporalValidation'` 为空 |
| G2 | 新增迁移脚本可重复执行 | 脚本支持 dry-run / execute 模式，并在 README 备案 |
| G3 | 临时文件删除无副作用 | Vitest、Playwright、ESLint 全量通过；关键页面手动验证无回归 |
| G4 | 文档与清单同步更新 | 更新实现清单、临时标签登记表，并在 IIG 报告中关闭条目 |

---

## 3. 解决方案概述
1. **引用梳理**：通过静态扫描锁定所有直接、间接导入（含别名导出、re-export）。
2. **自动迁移脚本**：在 `frontend/scripts/migrations/` 下新增 `20250921-replace-temporal-validation.ts`，使用 TypeScript AST（ts-morph）批量重写 import 与调用。
3. **手动兜底**：对脚本未覆盖的业务自定义逻辑进行代码评审与微调，确保逻辑等价。
4. **统一导出**：调整 `shared/utils/index.ts`，去除对旧文件的再导出，鼓励统一入口。
5. **验证 & 清理**：执行前端测试、快照更新，最终删除临时文件并更新 TODO 清单。

## 4. 项目原则对照
- **单一事实来源**：迁移后所有时态校验逻辑统一依赖 `shared/utils/temporal-converter.ts` 与实现清单记录，杜绝平行实现；删除旧文件前同步更新 `docs/reference/02-IMPLEMENTATION-INVENTORY.md`。
- **诚实与悲观谨慎**：脚本提供 `--check` 模式与手动审查清单，提前暴露潜在差异；风险矩阵覆盖脚本遗漏、容错差异等最坏场景。
- **健壮优先**：迁移闭环要求补充自动化测试（Vitest、Playwright）并执行命名/架构校验脚本，确保功能可靠性与文档同步。
- **中文沟通**：文档、脚本 README、测试报告均以中文输出，便于团队协同追踪。
- **契约优先**：本次仅调整前端实现，不触达契约文件；若发现 API 校验差异需先在 `docs/api/` 更新契约后再落地实现。

---

## 5. 行动拆解与负责人

### Phase A — 准备与映射（0.5 天）
- [ ] 负责人：Jinwei（前端架构组）
  - 任务 A1：在 `frontend/scripts/` 下创建 `audit/` 子目录，生成引用清单脚本 `frontend/scripts/audit/list-temporal-validation-usage.ts`（可沿用 `rg` 输出，与迁移脚本共享数据）。
  - 任务 A2：确认 `temporalValidation.ts` 每个方法在新工具中一一对应，补齐使用差异（如 `isFutureDate` 逻辑需迁移至 `TemporalUtils` + 业务自定义判定）。
  - 产出：方法映射表、潜在差异点列表。

### Phase B — 脚本开发（1 天）
- [ ] 负责人：Yuxi（时态体验小组）
  - 任务 B1：在 `frontend/scripts/` 下创建 `migrations/` 子目录，并新增 `20250921-replace-temporal-validation.ts`，支持以下能力：
    1. 支持 `--check`（dry-run）输出将变更的文件列表。
    2. 自动替换 `import { validateTemporalDate } from '../utils/temporalValidation';` 为新路径：
       - 基础校验函数 → `TemporalConverter.validateTemporalRecord`
       - 日期比较函数 → `TemporalUtils.*`
    3. 对命名空间调用（例如 `validateTemporalDate.isFutureDate`）进行安全重写。
  - 任务 B2：在 `frontend/package.json` 中新增脚本别名 `"migrate:temporal-validation"`，命令为 `"tsx scripts/migrations/20250921-replace-temporal-validation.ts"`，并在 `devDependencies` 中补充 `tsx` 与 `ts-morph`。
  - 任务 B3：脚本执行后自动移除不再需要的解构引用，保持 ESLint 通过。
  - 任务 B4：更新 `frontend/scripts/README.md` 文档，记录脚本使用方式、依赖安装步骤与回滚说明。

### Phase C — 执行迁移（1 天）
- [ ] 负责人：Jinwei
  - 任务 C1：在开发分支运行 `npm run migrate:temporal-validation -- --check`，确认输出与预期一致后执行正式迁移。
  - 任务 C2：对以下关键文件进行手动审查，确保逻辑等价：
    - `frontend/src/features/temporal/components/TemporalDatePicker.tsx`
    - `frontend/src/features/temporal/components/TemporalMasterDetailView.tsx`
    - 其他列表页、表单页涉及时态校验的文件。
  - 任务 C3：更新相关单元测试 / 快照，确保覆盖边界条件（未来日期、有效区间等）。

### Phase D — 清理与验证（0.5 天）
- [ ] 负责人：Lina（质量保障）
  - 任务 D1：执行 `npm run lint && npm run test && npm run test:e2e` 验证无回归。
  - 任务 D2：删除 `frontend/src/features/temporal/utils/temporalValidation.ts` 与相关 TODO 标记，更新实现清单 `docs/reference/02-IMPLEMENTATION-INVENTORY.md`。
  - 任务 D3：在 IIG 周报与 `docs/development-plans/07-pending-issues.md` 中标记任务关闭；如有延期改为记录新的截止信息。

---

## 6. 验证步骤
1. **静态检查**：`rg "temporalValidation" frontend/src` 返回空。
2. **命名与架构校验**：执行 `node scripts/quality/architecture-validator.js` 以及 `npm run validate:field-naming`，确认命名规范与分层约束未被破坏。
3. **类型检查与构建**：`cd frontend && npm run lint && npm run build` 必须通过。
4. **单元测试**：重点关注涉及日期边界的 Vitest 用例，新增覆盖未来日期与闭区间校验。
5. **E2E 场景**：运行 Playwright：
   - Temporal 版本创建流程
   - Temporal 历史编辑流程
   - 优先关注有效期校验 toast 提示、表单禁用等交互。
6. **手动验收**：
   - TemporalDatePicker 正常展示当前日期提示。
   - 错误提示信息与之前保持一致或更友好。

---

## 7. 风险与缓解措施
| 风险 | 描述 | 缓解措施 |
|------|------|----------|
| R1 | 脚本替换遗漏动态引用或别名导出 | 结合 `git diff` + 手动代码审查；保留脚本 dry-run 模式；对 `shared/utils/index.ts` 的导出改动单独评审 |
| R2 | 新工具类抛出的异常导致运行时错误 | 在迁移前补充非 try-catch 调用的错误处理；必要时在调用处增加防御性代码 |
| R3 | 时间紧迫导致测试覆盖不足 | 协调 QA 介入，锁定关键用户路径；优先保障时态创建/编辑功能 |
| R4 | 业务自定义逻辑依赖旧实现特性（例如容错行为） | 在 Phase A 列出差异并通过 Feature flag 或选项参数兼容 |
| R5 | 文档或清单未同步，破坏单一事实来源 | Phase D 强制校验实现清单、IIG 报告与脚本 README 已更新，必要时安排二次评审 |

---

## 8. 资源与依赖
- 工具依赖：新增 `tsx`、`ts-morph` 作为 `devDependencies` 并更新锁文件；若脚本需要 AST 额外能力，可评估 `ts-morph` 插件或自定义 helper。
- 人力依赖：前端架构负责人、时态小组核心开发、QA 人员各 1 名。
- 文档依赖：更新实现清单、IIG 报告、脚本 README。
- 时间窗口：建议在 2025-09-23 前完成，以便纳入下周发布窗口。

---

## 9. 输出与交付物
1. `frontend/scripts/migrations/20250921-replace-temporal-validation.ts`
2. 更新后的前端源码（所有引用使用 `TemporalConverter` / `TemporalUtils`）
3. 删除临时文件与 TODO 记录
4. 验证报告：测试日志、关键页面截图
5. 文档更新：实现清单、未决问题列表、IIG 报告

---

## 10. 后续跟踪
- 将脚本加入月度临时标签巡检流程，避免类似迁移任务重复堆积。
- 在 `scripts/check-temporary-tags.sh` 中新增规则，若检测到 `temporalValidation.ts` 或类似文件再现即阻断 CI。
- 回顾迁移执行情况，评估统一工具类在业务侧的扩展需求，必要时规划后续 API 优化。

---

**结论**：通过脚本 + 流程化验证的组合方案，可在短期内完成引用迁移并删除过期临时实现，降低 IIG 审计风险，提升时态功能代码质量与一致性。
