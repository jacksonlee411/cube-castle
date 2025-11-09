# Plan 246 – Temporal Entity Selector & Fixture 统一

**关联主计划**: Plan 242（T4）  
**目标窗口**: Day 12-15  
**范围**: 统一 Playwright selector/testid、测试 fixtures 与 E2E 工具

## 背景
- 现有 E2E 用例使用 `organization-*`、`position-*` testid 与 fixtures（frontend/tests/e2e/position-tabs.spec.ts 等）  
- Plan 242 需要 `temporalEntitySelectors` 作为唯一事实来源，减少 selector 漂移

## 工作内容
1. **temporalEntitySelectors**：在 `frontend/src/shared/testing/temporalEntitySelectors.ts` 集中定义 testid。  
2. **E2E 替换**：所有 Playwright 用例（职位/组织）使用新 selector，提供 codemod 与短期 alias。  
3. **Fixture 合并**：创建 `frontend/tests/e2e/utils/temporalEntityFixtures.ts`，以 `entityType` 生成 GraphQL/REST mock。  
4. **工具更新**：`waitPatterns`, `auth-setup`, `positionFixtures` 等 util 改写为中性命名；旧文件标记废弃。  
5. **验证**：Chromium/Firefox 连续 3 次运行 `position-tabs.spec.ts`、`organization-create.spec.ts`、`temporal-management-integration.spec.ts`。

## 里程碑 & 验收
- Day 14：codemod、selector、fixtures MR + Playwright 绿灯  
- Day 15：alias 清理 & 文档更新  
- 验收标准：仓库仅引用 `temporalEntitySelectors`；`positionFixtures.ts` 删除或仅 re-export 并带弃用告警。

## 汇报
- 更新 `215-phase2-execution-log.md` 并归档执行日志至 `logs/plan242/t4/`.
