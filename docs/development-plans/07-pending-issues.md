# 07 — GraphQL 路径字段增量迭代追踪

最后更新：2025-09-21  
责任团队：前端组（主责）+ 查询服务组（协同）

---

## 一、阶段目标
- 为 `Organization` GraphQL 类型新增可选 `path` 字段，并贯通前端展示链路。
- 确保 TypeScript 类型、数据转换、文档说明同步更新，避免临时占位字段。
- 在完整测试通过后，提交规范化交付物。

---

## 二、最新进展（已完成）
- **GraphQL 契约同步**：`docs/api/schema.graphql` 将 `Organization.path` 调整为可空，契约与实现对齐。
- **前端类型链路**：`OrganizationUnit`、`TemporalOrganizationUnit`、`TimelineVersion` 及 GraphQL 转换/守卫均改为接受 `string | null | undefined`，移除强制空字符串回填。
- **查询与展示逻辑**：时态详情页 GraphQL 查询显式请求 `path`，映射层与内联新版本表单均采用可空路径回退逻辑。
- **文档同步**：`docs/reference/02-IMPLEMENTATION-INVENTORY.md` 标注 `path` 字段可空化背景及返回约定。
- **质量校验**：`npm run lint` 通过。

相关改动：
- `docs/api/schema.graphql`  
- `docs/reference/02-IMPLEMENTATION-INVENTORY.md`  
- `frontend/src/shared/types/organization.ts`  
- `frontend/src/shared/types/temporal.ts`  
- `frontend/src/features/temporal/components/TemporalMasterDetailView.tsx`  
- `frontend/src/features/temporal/components/InlineNewVersionForm.tsx`  
- `frontend/src/features/temporal/components/TimelineComponent.tsx`  
- `frontend/src/shared/types/converters.ts`  
- `frontend/src/shared/api/type-guards.ts`  
- `frontend/src/shared/validation/schemas.ts`  
- `frontend/src/shared/hooks/useEnterpriseOrganizations.ts`

---

## 三、待解决事项（按优先级排序）
1. **TypeScript 编译错误修复（P0）**  
   - `npm run build` 当前在 `ParentOrganizationSelector.tsx`、`auth.ts`、`error-messages.ts` 等文件报错，需梳理 Canvas Kit 组件 API 与自定义类型定义，恢复可编译状态。  
   - 修复后补充必要的类型守卫或重构历史残留代码，防止再次阻塞构建。
2. **前端测试执行确认（P1）**  
   - `npm run test` 由于 `ERR_IPC_CHANNEL_CLOSED` 异常提前终止，需在修复编译错误后重跑，确认 Vitest 配置及 Node Worker 行为正常。  
   - 若问题持续，记录详细栈信息并评估是否与并行 worker 或限制作业环境相关。
3. **后端返回值确认（P1）**  
   - 验证查询服务在真实数据源上能回填 `path`；如存在空值，应评估数据补采或兜底策略，并更新前端展示提示。  
   - 若需后端补充映射/计算逻辑，另起任务跟踪。
4. **集成与验收（P2）**  
   - 完成上述修复后，执行 `npm run build`、`npm run test`、`make test`（必要时）以及核心用户路径自测。  
   - 根据结果更新交付说明，准备合并/发布材料。

---

## 四、下一步动作
- 优先排期 TypeScript 编译错误处理，必要时拆分任务指派到对应模块维护者。
- 修复完成后，复测 lint/test/build，并在此文档记录结论。
- 若触及后端逻辑调整，补充接口契约对齐说明与测试策略。
