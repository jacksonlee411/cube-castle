# Plan 256 - 契约单一事实来源（SSoT）与生成链路

文档编号: 256  
标题: 契约单一事实来源与生成链路（来源：202 计划拆分）  
创建日期: 2025-11-15  
版本: v1.0  
关联计划: 202、API 契约（docs/api/openapi.yaml、docs/api/schema.graphql）、脚本链路

---

## 1. 目标
- 明确契约 SSoT 与单向生成的边界，避免“第二事实来源”；
- 固化“契约→生成（types/clients）→实现校验”的链路（禁止“实现→契约”反生成）；
- 前后端一致性检查（字段命名 camelCase、path 参数 {code}、错误码等）。

## 2. 交付物
- SSoT 边界声明与门禁配置；
- 生成/校验脚本与执行顺序（只索引现有脚本与校验配置）；
- 失败示例与修复建议（登记，不复制正文）；
- 证据：logs/plan256/*。

## 3. 验收标准
- 契约 SSoT 边界清晰：仅以 docs/api/openapi.yaml 与 docs/api/schema.graphql 为契约真源；
- 仅允许“契约→生成代码/类型”，禁止“代码/实现→生成契约”；
- 合同校验（spectral 等）与实现对齐校验（脚本）通过；
- 生成脚本在 CI 与本地可重复执行，产物标注“generated, do not edit”。

---

维护者: 合同治理小组（前/后端与平台）

---

## 4. SSoT 边界与单向生成
- 唯一事实来源（SSoT）：`docs/api/openapi.yaml` 与 `docs/api/schema.graphql`（契约先行；禁止从实现生成契约）
- 单向生成：
  - 从契约生成 Go/TS 类型与客户端，仅落地至生成产物路径（带文件头禁止手改）
  - 生成产物示例（按现有脚本链路，具体以仓库为准）：
    - Go：`cmd/**/internal/types/contract_gen.go`
    - TS：`frontend/src/shared/types/contract_gen.ts`
- 修改契约的流程：
  - 修改 docs/api/* → 通过 spectral/schema 校验 → 运行生成脚本 → 提交生成产物 → 通过契约漂移门禁（见 Plan 258）

---

## 5. 门禁与工作流（CI）
- contract-sync（阻断型）：
  - 步骤：spectral 校验 → GraphQL schema 校验 → 运行生成脚本 → 检查工作树 clean
  - 失败条件：契约语义错误、生成后存在未提交变更、生成产物被手改
- drift-check（与 Plan 258 协同）：
  - 比对 OpenAPI 与 GraphQL 字段/类型/可空/描述的一致性（允许经白名单豁免，见 258）
- 实现对齐校验：
  - 扫描 REST 响应字段蛇形命名与路径参数不合规；扫描 GraphQL 枚举/字段硬编码（AGENTS 禁止项）

---

## 6. 执行顺序（参考）
1) 更新契约 docs/api/*（必须引用 Issue/PR 说明，避免第二事实来源）
2) 运行生成脚本（`scripts/contract/*`、`make generate-contracts`）
3) 本地合同校验与实现对齐校验通过
4) 提交生成产物与契约
5) CI contract-sync 与 drift-check 全绿

---

## 7. 备注
- 与 AGENTS 对齐：契约先于实现；契约与实现/文档之间只有“单向生成 + 校验”，不得反向生成契约
- 与 Plan 258：差异白名单需要 `// TODO-TEMPORARY(YYYY-MM-DD):` 且一个迭代内收敛；脚本对接 `scripts/check-temporary-tags.sh`
