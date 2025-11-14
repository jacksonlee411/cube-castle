# 245T – OpenAPI no-$ref-siblings 修复方案（独立任务/PR 草案）

目标：修复 `docs/api/openapi.yaml` 中的 `$ref` 旁挂（siblings）校验错误，满足 Spectral 规则 `no-$ref-siblings`，保持与实现一致且不引入破坏性改动。

— 

现状与错误明细
- 规则来源：`.spectral.yml`（启用 `no-$ref-siblings`）
- 最新校验结果：`logs/plan242/t3/51-openapi-lint.log`
- 错误（2 处）：
  - `components.schemas.PositionResource.properties.currentAssignment.nullable`
  - `components.schemas.PositionResource.properties.currentAssignment.description`
  - 说明：当前 `currentAssignment` 属性同时含有 `$ref` 与其他键（`nullable`、`description`），违反规范

— 

修复策略（最小且标准做法）
- 原始（不合法）：
  ```yaml
  currentAssignment:
    $ref: '#/components/schemas/PositionAssignmentResource'
    nullable: true
    description: Current assignment snapshot
  ```
- 建议（合法）：
  ```yaml
  currentAssignment:
    allOf:
      - $ref: '#/components/schemas/PositionAssignmentResource'
    nullable: true
    description: Current assignment snapshot
  ```
- 原因：`no-$ref-siblings` 规则要求 `$ref` 不与 sibling keys 并列；用 `allOf` 包裹 `$ref` 后，可安全添加 `nullable`/`description`

— 

影响评估
- OpenAPI 层面：语义不变；仅结构化方式调整
- 生成客户端：各语言常见生成器（OpenAPI Generator/Swagger Codegen）对 `allOf`+`$ref` 的支持良好，一般不会引入破坏性更改（需在 MR 中附 codegen PoC，如 TS/Java）
- 实现侧：不需要改动后端/前端代码

— 

实施步骤
1. 编辑 `docs/api/openapi.yaml`：
   - 定位 `components.schemas.PositionResource.properties.currentAssignment`
   - 将 `$ref` 改为 `allOf: - $ref: ...`，保留 `nullable` 与 `description`
2. 本地校验：
   - `npm run lint:api`，Spectral 0 error
3. 补充验证：
   - 运行 `node scripts/generate-implementation-inventory.js`，确认文档变更未引入 Inventory 漂移
4. MR 内容：
   - 说明变更背景（规则/位置/原因）
   - 附 `openapi.yaml` diff 与 `spectral` 日志
   - 若仓库包含 OpenAPI 生成物（目前无强制生成），附 TS/Java 生成 PoC（可作为附件，不入库）

— 

回滚方案
- 若发现客户端生成异常（不预期）：直接回滚到前一版本（保留问题，再以更保守的调整方案评估）

— 

完成标准
- `npm run lint:api` 显示 `0 errors`（仅允许既有 warnings）
- CI `api-compliance.yml` 通过
- 运行 `node scripts/generate-implementation-inventory.js` 通过

