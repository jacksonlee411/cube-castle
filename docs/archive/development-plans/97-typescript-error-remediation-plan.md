# 97号文档：前端 TypeScript 编译错误修复计划（职位域）

**版本**: v1.1
**创建日期**: 2025-10-20
**最新更新**: 2025-10-20
**触发来源**: `npm run build` 在当前主干（commit `2bd3717`）下持续失败
**覆盖范围**: `frontend/`（职位管理、Temporal 组件、共享 hooks）
**维护团队**: 前端团队 · 职位域
**遵循原则**: CLAUDE.md 资源唯一性原则 · 前端 TypeScript 零错误标准

**修改记录**:
- v1.1 (2025-10-20): 补充完整错误清单、增加 Phase 0、调整时间估算、强化验收标准
- v1.0 (2025-10-20): 初始版本

---

## 1. 背景与目标

在推进 92 号计划收尾时执行 `npm run build`，发现大量 TypeScript 报错。确认这些报错与本次 Job Catalog 布局改造无直接关联，而是历史遗留问题。为完成 92 号计划的 T1 验收（TypeScript 零错误），需整理并修复这些编译失败项。本计划旨在：

1. 盘点当前编译错误清单，按模块/类型分组；
2. 给出修复策略与优先级；
3. 明确验收标准，确保后续编译通过并回归 92 号计划。

`npm run build` 命令会先执行 `tsc -b` 再运行 Vite 构建，下文列出的报错均源自该命令。

---

## 2. 当前错误清单（截至 2025-10-20）

**错误总计**：60+ 条 TypeScript 编译错误

### 2.1 Job Catalog 模块错误

| 文件 | 行号 | 错误概述 | 分类 |
|------|------|----------|------|
| `types.ts` | 28-29 | 枚举默认值使用字符串字面量 `"ACTIVE"`/`"INACTIVE"` | **类型定义错误** |
| `family-groups/JobFamilyGroupForm.tsx` | 21 | `'ACTIVE'` 字符串字面量不匹配 `JobCatalogStatus` 枚举 | **状态枚举使用错误** |
| `families/JobFamilyForm.tsx` | 23 | 同上 | **状态枚举使用错误** |
| `roles/JobRoleForm.tsx` | 23 | 同上 | **状态枚举使用错误** |
| `levels/JobLevelForm.tsx` | 24 | 同上 | **状态枚举使用错误** |
| `shared/CatalogVersionForm.tsx` | 31, 55 | 同上 | **状态枚举使用错误** |
| `family-groups/JobFamilyGroupList.tsx` | 98-99 | `CatalogTable` 泛型约束不兼容（需要 `Record<string, unknown>`） | **表格数据类型定义** |
| `families/JobFamilyList.tsx` | 131-132 | 同上 | **表格数据类型定义** |
| `roles/JobRoleList.tsx` | 160-161 | 同上 | **表格数据类型定义** |
| `levels/JobLevelList.tsx` | 194-195 | 同上 | **表格数据类型定义** |
| `levels/JobLevelList.tsx` | 155, 166 | Canvas `Select` 组件不接受 `disabled` prop | **Canvas Kit API 变更** |
| `roles/JobRoleList.tsx` | 133 | 同上 | **Canvas Kit API 变更** |
| `shared/CatalogFilters.tsx` | 32 | Box 组件 `flexDirection` 响应式值类型不兼容 | **Canvas Kit 响应式类型** |
| `shared/CatalogForm.tsx` | 58 | Dialog 组件不接受 `onClose` prop | **Canvas Kit API 变更** |

### 2.2 Position 模块错误

| 文件 | 行号 | 错误概述 | 分类 |
|------|------|----------|------|
| `PositionTemporalPage.tsx` | 324, 332 | `PrimaryButton` `variant` 不接受 `"primary"/"secondary"` | **Canvas Kit API 变更** |
| `components/PositionForm/FormFields.tsx` | 58 | FormField `error` prop 类型不匹配（`string` vs `"error"\|"alert"`） | **Canvas Kit API 变更** |
| `components/PositionForm/FormFields.tsx` | 73 (多处) | FormField 子组件（Error/HelperText）不存在 | **Canvas Kit 组件架构变更** |
| `components/PositionForm/FormFields.tsx` | 109, 120, 129... (10+ 处) | TextInput `error` prop 类型为 `boolean`，应为 `ErrorType` | **Canvas Kit API 变更** |
| `components/PositionForm/FormFields.tsx` | 113, 134, 194... (10+ 处) | Flex `flexDirection` 响应式对象不匹配类型定义 | **Canvas Kit 响应式类型** |
| `components/PositionForm/FormFields.tsx` | 284 | TextInput 缺少 `as` prop 或 `label` prop 不兼容 | **Canvas Kit 组件约束** |
| `components/PositionForm/FormFields.tsx` | 320 | TextArea `error` prop 类型为 `boolean` | **Canvas Kit API 变更** |
| `PositionTemporalPage.test.tsx` | 3 | `React` 变量未使用 | **简单 lint 警告** |

### 2.3 Temporal 模块错误

| 文件 | 行号 | 错误概述 | 分类 |
|------|------|----------|------|
| `components/hooks/temporalMasterDetailSubmissions.ts` | 172 | `lifecycleStatus` 枚举包含 `"INACTIVE"` 等额外值 | **Temporal 枚举映射** |
| `components/hooks/temporalMasterDetailSubmissions.ts` | 179 | setState 类型不兼容（`lifecycleStatus` 类型冲突） | **Temporal 枚举映射** |

### 2.4 Shared Hooks 错误

| 文件 | 行号 | 错误概述 | 分类 |
|------|------|----------|------|
| `useEnterprisePositions.ts` | 671-882 | GraphQL 变量（`sorting`、`filter`）与 `JsonValue` 类型不匹配 | **GraphQL 变量类型** |
| `useJobCatalog.ts` | 163-271 | GraphQL 变量类型 `Record<string, unknown>` 与 `JsonValue` 不兼容 | **GraphQL 变量类型** |
| `usePositionMutations.ts` | 62, 200 | `logger.mutation` 调用参数不足 | **工具函数签名** |

### 2.5 Storybook/测试错误

| 文件 | 行号 | 错误概述 | 分类 |
|------|------|----------|------|
| `components/PositionForm/PositionFormFields.stories.tsx` | 1, 31 | 缺失 `@storybook/react` 类型声明 | **依赖缺失** |
| `__tests__/useJobCatalogMutations.test.tsx` | 78, 117, 157, 197 | 测试中使用字符串字面量 `"ACTIVE"` | **状态枚举使用错误** |

### 2.6 错误优先级分类

| 优先级 | 分类 | 错误数量 | 影响范围 |
|--------|------|----------|----------|
| **P0** | Canvas Kit 响应式类型系统 | 15+ | FormFields、CatalogFilters 等多处布局组件 |
| **P1** | 状态枚举使用错误 | 10+ | Job Catalog 表单、测试、类型定义 |
| **P1** | Canvas Kit API 变更（FormField/Button） | 15+ | Position 表单组件 |
| **P2** | 表格泛型类型定义 | 8 | Job Catalog 列表页 |
| **P2** | GraphQL 变量类型 | 200+ 行 | Shared hooks |
| **P3** | Temporal 枚举映射 | 2 | Temporal 组件 |
| **P3** | 工具函数签名、依赖缺失 | 5 | 零散问题 |

> **说明**：运行 `npm run build` 输出约 60 条错误，但部分错误在多个文件重复出现（如 `CatalogTable` 泛型问题在 4 个列表页重复）。优先级基于影响范围和修复复杂度。

---

## 3. 修复策略与分阶段计划

**总体策略**：按优先级分阶段修复，每阶段独立提交验证，确保无回归。重点关注 Canvas Kit API 变更研究，避免后续返工。

**预计总工期**：6-9 工作日（建议按 9 天规划，留缓冲时间）

---

### Phase 0：准备阶段（0.5 天）

**目标**：建立基准、隔离环境、准备文档

1. **创建专用分支**
   - 从当前主干创建分支 `fix/typescript-errors-remediation`
   - 验收：分支创建成功，工作目录干净

2. **建立错误基准**
   - 备份 `npm run build` 完整输出至 `docs/development-plans/97-build-errors-baseline.txt`
   - 记录当前测试通过率：`npm run test` 输出截图
   - 验收：基准文件已创建

3. **环境检查**
   - 确认 Canvas Kit 版本：`npm list @workday/canvas-kit-react`（应为 v13.2.15）
   - 检查 TypeScript 版本：`npx tsc --version`
   - 验收：依赖版本符合预期

4. **文档准备**
   - 在 06 号进展日志中记录修复启动
   - 创建修复进度追踪表（错误数量 vs 阶段）
   - 验收：日志已更新

---

### Phase 1：类型定义与枚举修复（1.5-2 天）

**目标**：修复所有状态枚举错误、类型定义问题

1. **types.ts 枚举定义修正**（0.5天）
   - 修改 `src/features/job-catalog/types.ts:28-29`，将默认值改为 `JobCatalogStatus.ACTIVE`
   - 验收：types.ts 编译通过，无 TS2820 错误

2. **表单与测试枚举替换**（1天）
   - 替换所有 `'ACTIVE'`/`'INACTIVE'` 字符串为 `JobCatalogStatus` 枚举值
   - 覆盖文件：
     - `family-groups/JobFamilyGroupForm.tsx`
     - `families/JobFamilyForm.tsx`
     - `roles/JobRoleForm.tsx`
     - `levels/JobLevelForm.tsx`
     - `shared/CatalogVersionForm.tsx`
     - `__tests__/useJobCatalogMutations.test.tsx`
   - 验收：`npm run test -- --run src/features/job-catalog` 通过

3. **CatalogTable 泛型调整**（0.5天）
   - 修改 `CatalogTable` 组件泛型约束，支持 `Record<string, unknown>`
   - 或在各数据节点类型上扩展索引签名：`[key: string]: unknown`
   - 验收：Job Catalog 列表页编译通过，无 TS2322 错误

**阶段验收**：
- [ ] 枚举相关错误清零（约 10+ 条）
- [ ] `npm run test -- src/features/job-catalog` 通过
- [ ] 错误数量从 60+ 降至 50 以下

---

### Phase 2：Canvas Kit API 研究与修复（2.5-3 天）

**目标**：深入研究 Canvas Kit v13.2.x API 变更，系统性修复组件用法

1. **Canvas Kit 官方文档研究**（0.5天）
   - 阅读 Canvas Kit v13.x 迁移指南（Breaking Changes）
   - 研究 FormField、Button、Select、Dialog、Flex/Box 组件 API
   - 重点：响应式 props 类型系统（`flexDirection` 等）
   - 产出：技术调研笔记（记录至 `docs/development-plans/97-canvas-kit-api-notes.md`）
   - 验收：形成修复策略文档

2. **FormField 组件迁移**（1-1.5天）
   - 修复 FormField 子组件引用方式（Error/HelperText）
   - 统一 `error` prop 类型：`boolean` → `"error" | "alert" | undefined`
   - 覆盖文件：`components/PositionForm/FormFields.tsx` (20+ 处错误)
   - 验收：FormFields.tsx 编译通过

3. **响应式类型修复**（0.5-1天）
   - 修复 Flex/Box 组件 `flexDirection` 响应式值类型
   - 策略选项：
     - 方案 A：使用类型断言 `as FlexDirection`
     - 方案 B：创建类型辅助函数统一处理
     - 方案 C：调整为非响应式值
   - 覆盖文件：`FormFields.tsx`、`CatalogFilters.tsx`
   - 验收：布局组件编译通过

4. **Button/Select/Dialog 组件修复**（0.5天）
   - PrimaryButton `variant` 属性调整
   - Select 组件 `disabled` prop 处理（封装或替换）
   - Dialog `onClose` 改为使用 model events
   - 验收：相关组件编译通过

**阶段验收**：
- [ ] Canvas Kit 相关错误清零（约 30+ 条）
- [ ] `npm run test -- src/features/positions` 通过
- [ ] 形成 Canvas Kit 封装组件（如需要）

---

### Phase 3：Temporal 枚举与 GraphQL 类型（2-3 天）

**目标**：修复 Temporal 模块枚举映射、GraphQL 变量类型问题

1. **Temporal 枚举映射**（1天）
   - 与后端确认 `TimelineVersion.lifecycleStatus` 实际值范围
   - 对比 GraphQL Schema，扩展内部枚举定义或添加类型守卫
   - 修复 `temporalMasterDetailSubmissions.ts` 类型冲突
   - 验收：`npm run test -- src/features/temporal` 通过

2. **GraphQL 变量类型建模**（1-2天）
   - 引入或定义 `JsonValue` 类型
   - 重构 `useEnterprisePositions` 和 `useJobCatalog` 的 `filter`/`sorting` 变量类型
   - 确保与 GraphQL 客户端类型约束一致
   - 验收：`npm run test -- src/shared/hooks` 通过，GraphQL 类型错误清零

3. **logger 工具函数修复**（0.5天）
   - 检查 `logger.mutation` 签名，补齐缺失参数
   - 或更新 `shared/utils/logger.ts` 类型定义
   - 验收：`usePositionMutations.ts` 编译通过

**阶段验收**：
- [ ] Temporal + GraphQL 错误清零（约 15+ 条）
- [ ] `npm run build` 错误数降至 10 以下
- [ ] 所有 shared hooks 测试通过

---

### Phase 4：依赖整理与收尾（0.5-1 天）

**目标**：清理依赖、移除 lint 警告、最终验证

1. **Storybook 类型声明**（0.3天）
   - 安装 `@storybook/react` 类型依赖：`npm install -D @storybook/react`
   - 或修改 `tsconfig.app.json` 排除 `.stories.tsx` 文件
   - 验收：Storybook 相关错误清零

2. **清理未使用变量**（0.2天）
   - 移除 `PositionTemporalPage.test.tsx` 中未使用的 `React` 引入
   - 运行 `npm run lint -- --fix` 自动修复其他 lint 警告
   - 验收：`npm run lint` 通过

3. **最终验证与文档**（0.5天）
   - 运行完整测试套件：`npm run test`
   - 运行生产构建：`npm run build`
   - 对比错误基准，确认所有错误已修复
   - 更新 92 号文档 T1 验收条目
   - 更新 06 号进展日志
   - 将本文档归档至 `docs/archive/development-plans/`
   - 验收：所有命令通过，文档已同步

**最终验收**：
- [ ] `npm run build` 零 TypeScript 错误（从 60+ → 0）
- [ ] `npm run lint` 零警告
- [ ] `npm run test` 全部通过，覆盖率不下降
- [ ] 关键 E2E 场景手动验证通过
- [ ] 相关文档已更新

---

## 4. 依赖与风险

### 4.1 关键依赖

| 依赖项 | 当前版本 | 说明 | 验证方式 |
|--------|---------|------|----------|
| Canvas Kit | v13.2.15 | API 变更是主要错误来源 | `npm list @workday/canvas-kit-react` |
| TypeScript | ~5.x | 编译器版本影响类型推断 | `npx tsc --version` |
| GraphQL Schema | 最新 | 后端契约需保持同步 | 对比 `docs/api/schema.graphql` |
| Vite | ~5.x | 构建工具，影响类型检查流程 | `npm list vite` |

### 4.2 风险评估与缓解

| 风险级别 | 风险描述 | 影响 | 概率 | 缓解措施 | 负责人 |
|---------|---------|------|------|----------|--------|
| **高** | Canvas Kit 响应式类型系统理解不足 | 修复方案可能需要多次返工 | 中 | Phase 2.1 专门研究官方文档，形成技术笔记；优先测试多种方案 | 前端团队 |
| **高** | FormField 组件架构重大变更 | 子组件引用方式可能完全不同 | 中 | 查阅官方迁移指南；考虑降级或自定义封装 | 前端团队 |
| **中** | 与 92 号计划代码冲突 | Job Catalog 代码同时被两个计划修改 | 高 | 建立专用分支隔离；先完成 97 号再合并 92 号 | 前端团队 |
| **中** | GraphQL 变量类型修复触发连锁变更 | 可能影响其他未发现的 hooks | 中 | 每次修改后运行完整测试套件；建立类型辅助函数 | 前端团队 |
| **中** | E2E 测试失败 | 大量表单组件修改可能破坏用户流程 | 中 | 每个 Phase 完成后立即运行 E2E；优先修复关键路径 | 前端团队 |
| **低** | Temporal 枚举值后端不确定 | 需要后端团队确认实际使用范围 | 低 | Phase 3 前与后端同步；建立类型守卫兜底 | 前端+后端 |
| **低** | Storybook 依赖安装问题 | 可能影响开发体验 | 低 | 优先排除编译目标；确实需要再安装 | 前端团队 |
| **低** | 测试覆盖率下降 | 修改过程中可能遗漏边界case | 低 | 对比 Phase 0 建立的基准；补充缺失测试 | 前端团队 |

### 4.3 外部依赖

| 依赖事项 | 依赖方 | 时间点 | 状态 |
|---------|--------|--------|------|
| GraphQL Schema 确认 | 后端团队 | Phase 3 前 | 待确认 |
| Temporal 枚举值范围 | 后端团队 | Phase 3 前 | 待确认 |
| 92 号计划完成状态 | 前端团队 | 本计划启动前 | 待确认 |
| Canvas Kit 官方文档访问 | Workday GitHub | 持续 | 正常 |

### 4.4 技术债务

**本次修复可能引入的技术债务**：
- 如采用类型断言（`as` 强制转换）解决响应式类型问题，需标注 `// TODO-TEMPORARY` 并在 Canvas Kit 升级后重新评估
- 如创建封装组件，需在 `frontend/README.md` 中记录维护责任
- 如 Temporal 枚举采用宽松类型守卫，需在后端明确枚举定义后回收

**避免技术债务的原则**：
1. 优先查阅官方文档，避免猜测性修复
2. 所有临时方案必须标注 `// TODO-TEMPORARY` 并说明原因
3. 封装组件需有明确的测试覆盖
4. 类型修改需考虑向后兼容性

---

## 5. 验收标准

### 5.1 定量指标（必须全部满足）

#### T1：编译与静态检查
- [ ] **TypeScript 编译零错误**
  - 基准：当前 60+ 条错误
- [ ] **`npm run lint` 零警告**
  - 基准：当前 `TextInput`、`Select` 等组件报错
- [ ] **`npm run test` 全部通过**
  - 基准：当前 REST Mutation + GraphQL hooks 均报错
- [ ] **CI/CD 通过**
  - 基准：CI pipeline 卡在 typecheck 阶段

#### T2：可维护性
- [ ] Canvas Kit 用法符合官方文档
- [ ] Shared Hooks 类型定义与 GraphQL 契约一致
- [ ] 枚举、常量统一集中维护
- [ ] 所有新增封装组件有对应单测

#### T3：文档与流程
- [ ] 92 号计划文档已更新 T1 验收条目
- [ ] 06 号日志记录修复经过
- [ ] 新增技术笔记归档

### 5.2 定性指标（建议满足）
- [ ] 表单体验保持一致（视觉、交互无明显退化）
- [ ] 关键链路 E2E 测试覆盖（职位创建、版本创建、更新）
- [ ] 组件封装具备扩展性（如 Select 可复用）

---

## 6. 资源计划

### 6.1 人力投入

| 角色 | 人天估算 | 职责 |
|------|----------|------|
| 前端负责人 | 3 | 拟定方案、代码审查、对齐原则 |
| 模块开发者 A | 2 | Job Catalog 表单与列表修复 |
| 模块开发者 B | 2 | Position 表单与 Temporal 页面修复 |
| 测试工程师 | 1 | 编写/更新单元测试、回归验证 |
| 文档维护 | 0.5 | 更新进展日志与计划归档 |

### 6.2 时间排期（建议）

| 日期 | 阶段 | 说明 |
|------|------|------|
| 10-20 上午 | Phase 0 | 建立基线、确认依赖、沟通范围 |
| 10-20 下午 ~ 10-21 | Phase 1 | 完成 Job Catalog 枚举与类型修复 |
| 10-22 ~ 10-24 | Phase 2 | Canvas Kit API 迁移、组件封装 |
| 10-27 ~ 10-29 | Phase 3 | Temporal & GraphQL 类型治理 |
| 10-30 | Phase 4 | 收尾、测试、文档归档 |

> 说明：若 Phase 2 耗时延长，可适当压缩 Phase 3（前提是 GraphQL 类型可在 Phase 2 并行推进）。

---

## 7. 风险与应对

### 7.1 主要风险

| 风险 | 描述 | 影响 | 缓解 |
|------|------|------|------|
| Canvas Kit 升级理解不足 | 文档碎片化，使用方式差异大 | 高 | Phase 2.1 先验证 Demo，沉淀笔记再改造 |
| GraphQL 类型变更 | `JsonValue` 限制导致 hooks 需要重构 | 中 | 与后端确认 Schema，定义统一类型辅助 |
| Temporal 枚举不一致 | 前端/后端枚举源不一致 | 中 | Phase 3 前同步架构师，必要时新增映射函数 |
| 测试覆盖不足 | 改动范围大，单测未覆盖 | 中 | Phase 2/3 完成后立即补测 |
| 排期压缩 | 与 92 号计划并行 | 中 | 明确先后顺序，避免并行修改同文件 |

### 7.2 升级与回滚策略

**升级步骤**：
1. Phase 0：完成基线记录
2. Phase 1~4：按阶段提交并占位 Review
3. 每个阶段结束运行 `npm run build`
4. 最终合并前触发 CI/CD 全流程

**回滚策略**：
1. 触发条件：
   - 单个 Phase 超时 1 天以上
   - 发现需要大规模重构（>500 行代码）
   - Canvas Kit API 无官方文档支持
2. 升级路径：团队负责人 → 技术委员会
3. 决策选项：调整策略、寻求外部支持、延期或降级目标

### 7.3 质量门禁

**提交前检查**（每次 Git commit）
- [ ] `npm run lint` 通过
- [ ] 相关模块测试通过
- [ ] 无新增 `// @ts-ignore` 或 `as any`（除非标注 TODO-TEMPORARY）
- [ ] 提交信息清晰（格式：`fix(frontend): Phase X - 修复 XXX 类型错误`）

**合并前检查**（每个 Phase）
- [ ] 阶段验收清单全部完成
- [ ] CI/CD 流水线通过
- [ ] Code Review 通过（至少 1 人审查）
- [ ] 无未解决的 TODO 或 FIXME（除非标注计划日期）

**最终发布检查**（Phase 4 完成后）
- [ ] 5.3 节最终验收清单全部完成
- [ ] 验收会议通过
- [ ] 相关文档已同步
- [ ] 92 号计划可恢复

### 7.4 回滚预案

**触发条件**：
- 发现修复引入严重 bug（P0/P1 级别）
- E2E 测试失败率 >20%
- 生产环境关键功能受影响

**回滚步骤**：
1. 立即停止当前 Phase
2. 切换回主干分支
3. 分析根因：技术方案问题 vs 执行问题
4. 重新评估：调整策略或分阶段合并
5. 记录教训：更新风险评估表

---

## 8. 参考资料

### 8.1 项目文档

- **CLAUDE.md**：项目核心原则与唯一事实来源索引
- **AGENTS.md**：临时方案管控规则（`// TODO-TEMPORARY` 标注要求）
- **开发者速查手册**：`docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`
- **实现清单**：`docs/reference/02-IMPLEMENTATION-INVENTORY.md`
- **API 工具指南**：`docs/reference/03-API-AND-TOOLS-GUIDE.md`

### 8.2 Canvas Kit 资源

- **官方文档首页**：https://workday.github.io/canvas-kit/
- **v13.0 Breaking Changes**：https://github.com/Workday/canvas-kit/releases/tag/v13.0.0
- **FormField 组件文档**：https://workday.github.io/canvas-kit/?path=/docs/preview-form-field--basic
- **响应式 Props**：https://workday.github.io/canvas-kit/?path=/docs/styling-responsive-styles--docs
- **迁移指南**：https://github.com/Workday/canvas-kit/blob/master/MIGRATION_GUIDE.md

### 8.3 TypeScript 参考

- **TypeScript Handbook**：https://www.typescriptlang.org/docs/handbook/intro.html
- **类型推断**：https://www.typescriptlang.org/docs/handbook/type-inference.html
- **泛型约束**：https://www.typescriptlang.org/docs/handbook/2/generics.html#generic-constraints
- **高级类型**：https://www.typescriptlang.org/docs/handbook/2/types-from-types.html

### 8.4 工具与命令

| 命令 | 用途 | 备注 |
|------|------|------|
| `npm run build` | 完整编译检查 | 包含 `tsc -b` + Vite 构建 |
| `npm run lint` | ESLint 检查 | 可加 `-- --fix` 自动修复 |
| `npm run test` | 运行测试套件 | 可加 `-- --run <path>` 指定模块 |
| `npx tsc -b --verbose` | 详细类型检查 | 查看编译过程 |
| `npm list @workday/canvas-kit-react` | 检查 Canvas Kit 版本 | 应为 v13.2.15 |
| `git diff --stat` | 查看修改统计 | 评估修改范围 |

### 8.5 相关 Issue 与讨论

- Canvas Kit `flexDirection` 响应式类型：https://github.com/Workday/canvas-kit/issues/XXXX（待补充）
- FormField API 变更讨论：https://github.com/Workday/canvas-kit/discussions/XXXX（待补充）
- 项目内部技术债务跟踪：`docs/development-plans/tech-debt-tracker.md`（如存在）

---

## 9. 附录

### 9.1 常见问题

**Q1：为什么不直接降级 Canvas Kit 版本？**
A1：降级会失去新功能和安全修复，且可能与其他依赖冲突。优先适配新版本 API，确保长期可维护性。

**Q2：如果 Phase 2 Canvas Kit 研究后发现无法修复怎么办？**
A2：启动回滚预案，考虑：(1) 自定义封装组件；(2) 降级到 v13.0.0 并锁定版本；(3) 向 Canvas Kit 提 Issue/PR。

**Q3：GraphQL 变量类型修复会影响后端吗？**
A3：不会。GraphQL Schema 是契约，前端只修复变量类型以符合 Schema，不改变实际请求内容。

**Q4：修复过程中是否可以继续开发其他功能？**
A4：建议暂停 Job Catalog 和 Position 相关功能开发，避免代码冲突。其他模块可正常开发。

### 9.2 术语表

| 术语 | 全称 | 说明 |
|------|------|------|
| Canvas Kit | Workday Canvas Kit | Workday 开源的 React 组件库 |
| CQRS | Command Query Responsibility Segregation | 命令查询职责分离架构模式 |
| P0/P1/P2/P3 | Priority 0/1/2/3 | 优先级分级（0 最高，3 最低） |
| T1/T2/T3 | Tier 1/2/3 | 验收标准分级（1 必须，2 重要，3 建议） |
| TODO-TEMPORARY | Temporary TODO comment | 临时方案标注（需说明原因） |
| E2E | End-to-End | 端到端测试 |

---

> **文档说明**：本计划聚焦 TypeScript 编译错误修复，为 92 号计划 T1 验收的前置条件。性能优化、无障碍深度验收等事项继续在 92 号计划中跟踪。修复完成后，本文档将归档至 `docs/archive/development-plans/`。
