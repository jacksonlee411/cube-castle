# Plan 245 – Temporal Entity 类型 & 契约统一

**关联主计划**: Plan 242（T3）  
**目标窗口**: Day 8-12  
**范围**: 统一 TS 类型、GraphQL Operation、Hook 命名

## 背景
- `PositionRecord`、`OrganizationUnit`、`PositionDetailQuery` 等命名造成多套事实来源  
- Plan 215/219 要求模块接口标准化，需在类型/契约层彻底抽象

## 工作内容
1. **Shared Types**：新增 `TemporalEntityRecord`, `TemporalEntityTimelineEntry`, `TemporalEntityStatus`，组织/职位改为类型别名。  
2. **GraphQL Operation**：统一 Query/Mutation 命名（`TemporalEntityDetailQuery` 等），同步更新 `docs/api/schema.graphql` 与 `docs/api/openapi.yaml`。  
3. **React Query**：实现 `useTemporalEntityDetail` Hook（含 QueryKey、Suspense 支持），`usePositionDetail`/`useOrganizationDetail` 成为薄封装。  
4. **Codemod & 验证**：编写脚本批量替换类型/查询名，运行 `node scripts/generate-implementation-inventory.js` 校验。  
5. **测试**：新增 Hook/Vitest 覆盖、GraphQL 层契约测试。

## 里程碑 & 验收
- Day 10：完成类型/Hook 重构 MR + 单测  
- Day 12：GraphQL 文档/契约更新、inventory 校验通过  
- 验收标准：仓库不再出现旧类型或 Query 名称；`useTemporalEntityDetail` 覆盖所有消费端。

## 汇报
- 更新 `215-phase2-execution-log.md`、Plan 242 文档；日志存放 `logs/plan242/t3/`.
