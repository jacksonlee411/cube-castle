# Plan 244 – Temporal Entity Timeline & Status 抽象

**关联主计划**: Plan 242（T2）  
**目标窗口**: Day 3-8  
**范围**: 统一 Timeline Adapter 与状态元数据命名

## 背景
- `frontend/src/features/positions/timelineAdapter.ts`、`statusMeta.ts` 仅覆盖职位  
- 组织模块使用 `shared/utils/statusUtils.ts` 与独立 Timeline 结构  
- Phase2 强调组织模块标准化（Plan 219），需统一命名与类型

## 工作内容
1. **TemporalEntityTimelineAdapter**：抽象 `createTemporalTimelineAdapter`，覆盖组织/职位，提供 default mapper + entity override。  
2. **TemporalEntityStatusMeta**：集中 `statusConfig`，提供 `position`、`organization` 配置及扩展点。  
3. **引用更新**：Position/Organization 组件、Storybook、Vitest、GraphQL loader 全面替换。  
4. **Lint 规则**：新增 lint 以阻止引用旧 `timelineAdapter`/`statusMeta`。  
5. **回归**：Storybook 截图、Vitest/Playwright 基线更新。

## 里程碑 & 验收
- Day 6：完成 Adapter/Status 重构 MR + 单测  
- Day 8：完成引用替换与 lint 规则  
- 验收标准：`rg "timelineAdapter"|statusMeta` 仅在新命名空间出现；组织/职位 UI 行为一致。

## 汇报
- 在 `215-phase2-execution-log.md` 记录阶段性节点；日志输出 `logs/plan242/t2/`.
