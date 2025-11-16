# 301 · 字段可扩展标准（标准化装配能力）

状态：提案（可实施）  
最后更新：2025-11-15  
对齐：300 平台化 UI、202（254/256/257/258）、203（DDD 边界）、215（证据登记）

---

## 1. 目标
- 将“新增/复制字段”的全链路（契约→迁移→服务→前端→门禁→回归）标准化，支持配置化与扩展化落地。
- 保证“单一事实来源”：字段事实仅出现在 `docs/api/*`；预设/脚手架仅作为生成器，不承载业务事实。

---

## 2. 能力清单
- 字段预设库 Field Presets（生成器）
  - 如：`text.like(description)`、`code.like(organizationCode)`、`date.like(effectiveDate)` 等。
- 字段生成 CLI（SSoT 驱动）
  - 命令：`npm run fields:add -- --entity OrganizationUnit --field alias --preset text.like(description) --persist --nullable true --bcId organization --ui.group "基本信息" --ui.order 45`
  - 产物：建议补丁（GraphQL/OpenAPI）、迁移 SQL（Goose Up/Down）、前端生成触发说明（表单/列表/片段）。
- 前端装配器（零/低代码）
  - 表单：OpenAPI→Zod + React Hook Form + Canvas 映射。
  - 列表：GraphQL 自省→列定义 + TanStack Table（headless）+ Canvas 皮肤。
  - 详情：属性网格由 FieldDescriptor 渲染；页签/区块通过 Manifest/插槽注册（按 requiredScopes 显隐）。
- 领域门面强制（202:257）：页面与扩展仅经“前端领域 API 门面”访问数据，禁直连客户端。
- 单基址与契约门禁（202:254/258）：端点收敛与契约漂移校验纳入 preflight。
- DDD 边界（203）：`bcId` 锚定 Bounded Context；清单/插槽按 BC 校验注册。

---

## 3. 标准流程
1) 契约先行  
   - GraphQL（查询字段）：更新 `docs/api/schema.graphql`，类型/可空性/描述。  
   - OpenAPI（命令字段）：更新 `docs/api/openapi.yaml` 对应请求/响应模型（可空/format/enum/description/x-scopes）。
2) 迁移（如需存储）  
   - 生成 Goose 迁移 `database/migrations/yyyymmddHHMMSS_add_<entity>_<field>.sql`（Up/Down 对称）。
3) 服务层  
   - 查询：gqlgen 生成，仓储 SELECT/Scan/DTO/Resolver 增量。  
   - 命令：DTO/校验/持久化/审计/Outbox；如增端点补 PBAC 映射。
4) 前端装配  
   - 表单：OpenAPI→Zod/RHF 自动渲染；控件映射来自注册表。  
   - 列表：自省→列定义；默认排序/过滤取自预设。  
   - 详情：属性网格；如需页签，通过 Manifest 插槽注册。
5) 门禁与回归  
   - 端点收敛、契约漂移、禁直连门面、清单/插槽结构校验。  
   - E2E 覆盖创建/查看/筛选；215 执行日志登记证据。

---

## 4. CLI 与预设（最小形态）
- 入口：`scripts/fields/add.js`（Node，无外部依赖）
- 参数：`--entity`、`--field`、`--preset | --copy-from`、`--persist`、`--nullable`、`--bcId`、`--ui.group`、`--ui.order`
- 预设：`scripts/fields/presets.json`（仅作为生成器加速，不是事实来源）
- 产物：  
  - 迁移 SQL：实际可直接执行的 Up/Down。  
  - 建议补丁：`scripts/fields/out/<entity>.<field>.graphql.patch` 与 `.openapi.patch`（供审阅并人工合并至契约）。  
  - 说明：前端生成链与 E2E 提示。

---

## 5. 门禁（preflight）
- `scripts/quality/preflight-field-standard.js`：
  - 检查新增迁移 Up/Down 对称；  
  - 校验 `bcId` 是否落在 203 约定集合；  
  - 提示存在“待合并契约补丁”，引导审阅（不阻断）。
- 与 300/202 接线：在根 `quality:preflight` 中串行执行。

---

## 6. 验收
- 复制同类字段后，无需手改页面骨架即可：  
  - 表单显示与校验正确；  
  - 详情属性网格展示；  
  - 列表可选显示并支持排序/过滤；  
  - 门禁通过，E2E 证据登记于 215。

---

## 7. 边界
- 事实来源仅 `docs/api/*`；预设与 UI hints 不得携带权限/业务事实。  
- 枚举/复杂引用字段可先走“建议补丁 + 人审”模式，确保契约与实现一致。

---

## 8. 可视化与 UI 内创建（面向管理员）

8.1 Studio 流程（开发/预发环境）
- 入口：`/admin/studio/fields`（ADMIN/CONFIG_MANAGER）
- 操作：选择实体→选择“复制字段”或“预设”→配置可空/默认/索引/分组/顺序→生成
- 结果：落盘到
  - `scripts/fields/out/<Entity>.<field>.graphql.patch`
  - `scripts/fields/out/<Entity>.<field>.openapi.patch`
  - `database/migrations/<ts>_add_<table>_<field>.sql`
  - `frontend/src/shared/extension/layout-patches/*.json`（布局 hints）
- 推广：走 PR 审阅与 CI 门禁（preflight），215 登记；禁止在 Prod 直接改契约。

8.2 FLS/FOV 可视化
- 以矩阵（角色×字段×读/写）管理 OpenAPI `x-fls`，生成补丁与前端消费配置；前端表单/详情/列表自动应用显隐/只读。
- 后端 PBAC 一致校验，拒绝越权操作。

8.3 规则与显隐（可视化规则编辑器）
- 限定型 JSON 规则（必填/长度/模式/范围/条件显隐/跨字段依赖）；Studio 编辑→生成补丁→映射到 Zod/RHF 与后端校验。
- 规则存储位置：OpenAPI `x-rules` 或 `docs/reference/rules/*.json`（统一由生成器合入）。

---

## 9. 运行时扩展通道（Flexfield/EAV 兼容）

9.1 设计
- 核心实体增加 `customAttributes JSONB` 与 `attributes_meta`（字段定义/类型/约束/可见性）两表；迁移归档，配置作为数据管理。
- GraphQL 暴露 `customAttributes: JSON`；REST 暴露 `extensions` 对象。

9.2 约束
- 配置数据走迁移与 215 审计；限制字段数量/索引/查询，避免性能风险。
- 此通道不取代核心字段；仅适用租户轻量自定义；要沉淀为核心字段，走第 3 章流程。

9.3 前端装配
- FieldDescriptor 由元数据生成→自动渲染；列表/过滤遵循类型与索引策略；Layout/规则/FLS 同样生效。

---

## 10. 打包与推广（Packaging）
- Studio 支持导出“配置包”（字段/布局/规则/FLS 元数据 JSON + 迁移 SQL 草案 + 契约补丁）。
- Dev→Staging→Prod：PR 合并、迁移执行、契约合入、preflight 门禁、215 登记、可回滚（含 Down SQL 与反向补丁）。
- P2：为包与远程功能包引入版本/签名/依赖与白名单加载策略。

---

## 11. 安全与审计
- 权限：仅 ADMIN/CONFIG_MANAGER 可访问 Studio；敏感操作审计落盘（不含私密数据）。
- 门禁：端点收敛/契约漂移/禁直连/清单结构校验必过；拒绝产生第二事实来源。
- 回滚：生成 Down SQL 与反向契约补丁；保留配置包与审计记录。
