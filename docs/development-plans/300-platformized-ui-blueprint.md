# 平台化 UI 蓝图（文档编号：300）

状态：提案（RFC）  
最后更新：2025-11-15  
适用范围：前端（React/TS）、Query 服务（GraphQL/gqlgen）、Command 服务（OpenAPI/REST）  
上游事实来源（SSoT）：`docs/api/schema.graphql`（查询）、`docs/api/openapi.yaml`（命令+权限 scopes）

---

## 1. 背景与目标

在不引入第二事实来源的前提下，补齐“平台化 UI”的关键能力，使前端具备接近 Odoo 的“安装/装配即用”体验，同时保留现有工程治理优势（契约先行、CQRS、门禁与回归能力）。

核心目标
- P0（本期）：提供“页面插槽 + 功能清单（Manifest）”能力，实现菜单/路由/页签/区块的清单化装配，并与 PBAC scope 一致化控制可见性。
- P1（次期）：打通“OpenAPI→表单 Schema 生成”“GraphQL 自省→列表列定义生成”，构建开发提效链路和统一的元数据消费层；引入脚手架。
- P2（远景）：探索“远程功能包/按需装配”的渐进方案（构建期/运行时），在不破坏 SSOT 的情况下实现更强的增量装配能力。

非目标
- 不构建“另一个后台系统”；不在前端重新定义字段/权限/状态等事实，全部从 OpenAPI/GraphQL 契约生成或映射。

---

## 2. 约束与一致性

- 单一事实来源：字段/枚举/权限 scopes 仅取自 `docs/api/schema.graphql` 与 `docs/api/openapi.yaml`。任何 UI 可见性（按钮/菜单/页签）依据 OpenAPI 中的 scopes。
- CQRS 不变：命令→REST（OpenAPI）、查询→GraphQL（schema）。前端数据访问继续通过统一客户端与 codegen 类型。
- Docker 强制：开发/测试/集成依赖容器（PostgreSQL/Redis），禁止宿主安装以免环境漂移。
- 门禁延续：禁硬编码 data-testid；新增能力需通过 ESLint/脚本门禁纳入 `npm run quality:preflight`。
- 领域 API 门面强制（对齐 202:257）：Manifest/Slot 中的业务代码禁止直接调用 GraphQL/REST 客户端，必须经“前端领域 API 门面”访问领域数据；以 ESLint 规则与 preflight 守卫强制。
- 单基址端点收敛（对齐 202:254）：路由装配与数据请求必须遵循“单基址代理”配置，GraphQL `/graphql` 与 REST `/api/v1` 共享基址，预防多源配置与环境漂移。

---

## 3. 顶层设计（组件与数据流）

文本视图（组件→数据→权限→契约）：
```
   [Feature Manifest] --注册--> [Extension Registry] --挂载点解析--> [Slot Renderer]
           |                                                |
     (scopes/menus/routes)                             (PBAC 显隐)
           |                                                |
   [OpenAPI/GraphQL 契约] ------------------> [生成/校验：表单Schema、列表列、类型]
                                              |
                                         [UI 组件渲染]
```

- Extension Registry：运行时注册表，管理“菜单、路由、插槽内容”的清单。由功能清单（Manifest）静态/动态注册。
- Slot Renderer：在页面骨架（如 TemporalEntity 页面）渲染具体插槽（tabs/sidepanels/sections）。
- PBAC 显隐：前端根据用户 JWT scopes（来自 OpenAPI 权限契约）决定清单项/插槽是否可见；后端仍进行强校验。
- 生成层：OpenAPI→表单 Schema、GraphQL 自省→列表列定义，作为“只读元数据”，避免重复事实来源。

---

## 4. 分阶段路线图与交付物

P0（1–2 周，最小可用）
1) 页面插槽机制（Tabs/Sections）
   - 新增：`frontend/src/shared/extension/slots.ts`
     - 能力：`registerTab(entity, id, component, { order, requiredScopes })`
     - Slot ID 约定：`temporal:organization:tabs` / `temporal:position:tabs`
   - 接入：在时态实体页面骨架渲染插槽（`features/temporal/layout/TemporalEntityLayout.tsx` 附近），不改变现有 DOM/testid。
   - PBAC：渲染前按 scopes 过滤；scopes 来自 OpenAPI 权限契约。
2) 功能清单（Manifest）与路由/菜单装配
   - 新增：`frontend/src/shared/extension/manifest.d.ts`（类型）、`frontend/src/shared/extension/registry.ts`（注册表）
   - Manifest 字段：
     ```
     id, routes[{ path, element, scopes? }], menus[{ id, label, path, order?, scopes? }],
     slots[{ target, component, order?, scopes? }], enabled?(tenant, scopes)
     ```
   - 接入：`App.tsx` 启动时扫描/注册清单，装配菜单与路由。
3) 质量与门禁
   - 选择器守卫适配插槽/清单：禁止硬编码 testid，统一从 `shared/testids` 导出。
   - 在 `package.json` 的 `quality:preflight` 增加 manifest/slot 的基本校验脚本（文件存在性、重复 ID、必填字段）。

P1（2–3 周，实用提效）
4) OpenAPI → 表单 Schema 生成（命令/表单）
   - 新增脚本：`scripts/generate-forms-from-openapi.ts`
     - 输入：`docs/api/openapi.yaml`
     - 输出：`frontend/src/generated/forms/*.ts`（Zod/JSON Schema + UI hints）
     - 映射：请求体验证规则→表单校验；枚举/字段描述→控件与占位提示；权限 scope→提交按钮显隐。
   - 前端：统一 Form 引擎消费 schema，减少手写表单。
5) GraphQL 自省 → 列表列定义生成（查询/列表）
   - 新增脚本：`scripts/generate-columns-from-graphql.ts`
     - 输入：`docs/api/schema.graphql`
     - 输出：`frontend/src/generated/columns/*.ts`（字段、类型、可排序/过滤）
   - 前端：统一 Table 组件消费列定义，驱动排序/筛选 UI。
6) 脚手架与清单模板
   - 新增命令：`npm run g:feature`
   - 产物：功能包骨架（manifest、路由、插槽样板、测试、选择器），接入 `quality:preflight`。
7) 租户覆盖与主题化
   - 扩展 `shared/config/tenant.ts`：支持租户远端配置拉取（字段可见性/文案覆盖/菜单开关），前端仅做渲染逻辑，不产生第二事实来源。

P2（2 周，前瞻探索）
8) 渐进式远程功能包
   - 方案 A：构建期“外部功能包仓库”静态合并（最稳健，风险低）。
   - 方案 B：Vite Federation/远程模块（需严格门禁：不得引入第二事实来源；仅前端 UI，数据契约仍来自 docs/api）。
   - 统一回滚策略：Manifest 层级禁用 + 基线路由/页面兜底。

---

## 5. 接口与类型设计（建议稿）

5.1 插槽注册表
```ts
// frontend/src/shared/extension/slots.ts
export type SlotTarget =
  | 'temporal:organization:tabs'
  | 'temporal:position:tabs';

export interface SlotItem {
  bcId: 'organization' | 'workforce' | 'contract' | 'performance' | 'compensation' | 'payroll' | 'attendance' | 'recruitment' | 'development' | 'compliance' | string;
  id: string;
  target: SlotTarget;
  component: React.LazyExoticComponent<React.ComponentType<any>> | React.ComponentType<any>;
  order?: number;
  requiredScopes?: string[];
}

export function registerSlot(item: SlotItem): void
export function getSlots(target: SlotTarget, userScopes: string[], bcId?: SlotItem['bcId']): SlotItem[]
```

5.2 功能清单（Manifest）
```ts
// frontend/src/shared/extension/manifest.d.ts
export interface FeatureManifest {
  id: string;
  bcId: 'organization' | 'workforce' | 'contract' | 'performance' | 'compensation' | 'payroll' | 'attendance' | 'recruitment' | 'development' | 'compliance' | string;
  routes?: Array<{ path: string; element: () => Promise<{ default: React.ComponentType<any> }>; scopes?: string[] }>;
  menus?: Array<{ id: string; label: string; path: string; order?: number; scopes?: string[] }>;
  slots?: Array<{ target: 'temporal:organization:tabs'|'temporal:position:tabs'; component: () => Promise<{ default: React.ComponentType<any> }>; order?: number; scopes?: string[]; bcId?: FeatureManifest['bcId'] }>;
  enabled?: (tenantId: string, scopes: string[]) => boolean;
}
```

5.3 PBAC 显隐（前端）
```ts
export function hasAnyScope(userScopes: string[], required?: string[]): boolean {
  return !required || required.length === 0 || required.some(s => userScopes.includes(s));
}
```
说明：scope 名称由 OpenAPI 提供，禁止在前端自造或变形。

---

## 6. 生成工具规范

- OpenAPI→表单 Schema
  - SSOT：`docs/api/openapi.yaml`
  - 生成内容：字段类型/必填/枚举、校验规则（Zod/JSON Schema）、UI hints（description、example）
  - 目录：`frontend/src/generated/forms/`（只读；PR 中禁止手改）
  - 门禁：在 `quality:preflight` 中校验生成结果是否最新

- GraphQL 自省→列表列定义
  - SSOT：`docs/api/schema.graphql`
  - 生成内容：列 key、label（优先取 schema 描述/指令）、排序/过滤能力，类型安全
  - 目录：`frontend/src/generated/columns/`
  - 门禁：生成与引用一致性检查
  - 推荐消费层：TanStack Table（headless）+ Canvas Kit 组件映射；避免引入与 Canvas 冲突的重型表格框架
- 契约漂移门禁（对齐 202:258）
  - preflight 集成 OpenAPI ↔ GraphQL 漂移检查：字段名、可空、描述、枚举。
  - 生成链路与差异报告绑定：若生成物未更新或契约不一致，PR 阻断并输出差异摘要。
- 端点收敛校验（对齐 202:254）
  - preflight 检查 GraphQL/REST 客户端基址与代理配置是否一致，禁止多基址散落。

---

## 7. 安全、权限与可观测性

- 权限来源：仅以 OpenAPI scopes 为准（服务端 PBAC 同步校验）；前端仅做“显隐控制+UX 反馈”，不视作安全边界。
- JWT/租户：沿用现有统一客户端与 header 注入方式；清单/插槽渲染时读取用户 claims 的 scopes。
- 可观测性：为清单加载、插槽渲染、表单/列表生成注入 `performance.mark` 与 `[OBS]` 事件；在 `logs/` 归档测试证据（由测试/CI 采集）。

---

## 8. 测试与门禁

- 单测：注册表/清单/插槽选择逻辑（scopes 过滤、排序、重复 ID 拦截）。
- 合规：禁硬编码 testid（复用 `shared/testids`），新增 ESlint 规则/脚本检查 Manifest 与 Slot 的基本结构。
- 门面守卫（对齐 202:257）：ESLint/脚本禁止 Manifest/Slot 代码直接引用统一客户端/`fetch/axios`；要求通过“领域 API 门面”（建议路径 `frontend/src/shared/api/facade/*`）访问数据。
- E2E：Playwright 覆盖
  - “具备 scope → 菜单/页签可见”与“缺失 scope → 隐藏”
  - 表单由 OpenAPI 生成时，基础校验与提交流程可跑通
- CI 门禁：`npm run quality:preflight` 包含 manifest/slot 校验 + 生成物一致性校验。
  - 追加：契约漂移门禁、端点收敛校验、门面禁直连校验。
  - 追加（对齐 240）：复跑职位/组织关键 E2E 套件，确保插槽注入不改变 DOM/testid 与可观测事件命名。

---

## 9. 风险与回滚

- 风险
  - 插槽/清单引入运行时耦合：通过类型与门禁限制；加载失败兜底为“基础页面”。
  - 元数据生成失配：锁定契约版本，生成脚本幂等；增量字段先改契约再生成。
  - 第三方表单/列表框架与 Canvas Kit 主题不一致：优先选择 headless 或自行映射，避免视觉/交互割裂。
- 回滚
  - Feature flag/租户开关；Manifest 层级禁用；删除功能包不影响核心流程。

---

## 10. 里程碑与验收标准

P0 验收
- 能在组织/职位详情页看到通过清单注册的“额外页签”；具备 scope 时显示、缺失 scope 隐藏。
- 菜单/路由由 Manifest 驱动生成；无任何 testid 硬编码违规；CI 通过新增校验。
- 门面禁直连生效：Manifest/Slot 代码禁直连 GraphQL/REST 客户端的门禁通过（对齐 202:257）。
- 端点收敛：单基址代理检查通过（对齐 202:254）。
- 与 240 对齐：复跑 240 的职位/组织相关 E2E 套件，DOM/testid 不变；OBS 事件名称遵循既有规范。

P1 验收
- 至少 2 个命令表单由 OpenAPI 生成（创建/替换/挂起组织三选二），端到端提交流畅，错误提示与校验一致。
- 至少 1 个列表由 GraphQL 列定义生成，排序/筛选可用。
- 脚手架产物可直接集成、通过门禁。
- 契约漂移门禁通过：OpenAPI ↔ GraphQL 差异报告为零或白名单内（对齐 202:258）。

P2 验收
- 完成“构建期合并”的外部功能包 PoC；Manifest 可启用/禁用；失败可回滚到基础路由。

---

## 11. 变更与文档治理

- 本蓝图仅描述原则与索引，不复述契约字段。实际字段/权限以 `docs/api/*` 为唯一事实来源。
- 任何实现落地后，应在 `CHANGELOG.md` 登记，并将执行证据落盘路径登记至 `docs/development-plans/215-phase2-execution-log.md`。

---

## 附录：与现有代码的接入点参考

- 页面骨架与统一 Hook：`frontend/src/features/temporal/layout/TemporalEntityLayout.tsx`、`frontend/src/shared/hooks/useTemporalEntityDetail.ts`
- 状态元数据注册表：`frontend/src/features/temporal/entity/statusMeta.ts`
- 统一 GraphQL 客户端/企业信封适配：`frontend/src/shared/api/graphql-enterprise-adapter.ts`
- 选择器守卫：`frontend/src/shared/testids/temporalEntity.ts`

---

## 12. 框架选型评估（基于现有栈）

现有前端基线：React + React Router + React Query + Zod 校验 + GraphQL codegen + Canvas Kit（设计系统）。

12.1 P0 插槽/清单（Manifest）
- 推荐：自研轻量“注册表 + 清单”实现（~200–300 行），无外部依赖，最贴合现有路由与 Canvas 组件。
- 不推荐：single-spa/通用微前端容器（过重，调试与门禁接入复杂），会破坏现有工程门禁与一致性。

12.2 P1 表单引擎（OpenAPI → 表单）
- 首选：React Hook Form + @hookform/resolvers + Zod
  - 理由：与现有 Zod 校验一致；性能优；能直接映射到 Canvas Kit 表单控件；类型安全。
  - 生成链：OpenAPI → Zod schema + 轻量 field-descriptor（label/placeholder/控件类型）→ RHF 动态表单。
- 备选：react-jsonschema-form（rjsf）
  - 优点：可直接消费 JSON Schema（与 OpenAPI 接近），上手快。
  - 风险与约束：
    - 需要自研 Canvas Kit 主题（FieldTemplate/Widgets/Array/Object 全套），适配成本高。
    - UI Schema 仅限表现层，不得携带业务事实（避免第二事实来源）；必须由门禁脚本强制。
- 备选：JsonForms / Uniforms
  - 优点：DSL 完备、功能强。
  - 风险：与 Canvas 主题适配重、引入 DSL 学习曲线；建议在 rjsf 遇阻且需求强烈时再评估。

12.3 P1 列表（GraphQL 自省 → 列定义）
- 首选：TanStack Table（@tanstack/react-table，headless）
  - 理由：将“列定义/排序/过滤”从 UI 解耦，恰好适配“由自省生成列元数据”；用 Canvas Kit 来实现表头、行、分页等皮肤。
- 备选：自研基础表格
  - 风险：排序/过滤/虚拟化等能力需要逐步补齐，长期总成本可能高于 TanStack Table。

12.4 P2 远程功能包（仅 UI 层）
- 首选：@originjs/vite-plugin-federation（Vite 生态）
  - 约束：仅承载 UI 组件与 Manifest；禁止携带数据模型/权限映射；SSOT 仍是 docs/api/*。
  - 回滚：Manifest 开关 + 基础路由兜底。
- 不推荐：运行时全托管微前端外壳（复杂度与门禁耦合高，不利于现有 CI/证据采集）。

12.5 推荐组合小结
- P0：自研插槽/清单注册表（PBAC 显隐）+ 门禁校验脚本。
- P1（表单）：RHF + Zod + Canvas 控件映射（必要时临时引入 rjsf 做 PoC，但需先完成 Canvas 主题层与门禁）。
- P1（列表）：TanStack Table + 自省生成列定义。
- P2：Vite Federation PoC（严格限制为 UI 层）。

---

## 13. 落地节奏建议（执行清单）

- 本周（P0）
  - 实现插槽注册表与功能清单，接入 App.tsx；组织详情新增一个“自定义页签”样例（按 scope 显隐）。
  - 新增 manifest/slot 校验脚本并挂到 `npm run quality:preflight`；补充最小 E2E。
- 下周–两周（P1）
  - 表单链路：引入 RHF + Zod resolver；完成 2 个命令表单（如创建/替换/挂起组织）从 OpenAPI 生成→渲染→提交的端到端闭环。
  - 列表链路：引入 TanStack Table；完成一个组织列表的“自省→列定义→渲染（排序/过滤）”PoC。
  - 脚手架：`npm run g:feature` 产出 manifest/slot/路由/测试骨架，纳入门禁。
- 后续（P2）
  - 远程功能包 PoC（仅 UI）：演练启用/禁用与回滚；记录性能与稳定性观测点；确保不引入第二事实来源。

---

## 14. 与 202/203/240/200-201 的对齐

- 对齐 202（CQRS 混合架构执行指引）
  - 202:254 端点收敛：本蓝图在“约束与一致性/生成工具规范/门禁”处明确单基址代理要求与校验。
  - 202:256 契约 SSoT：本蓝图的生成链（OpenAPI→表单、GraphQL 自省→列）严格以 `docs/api/*` 为 SSOT。
  - 202:257 前端领域 API 门面：本蓝图将“门面强制/禁直连”纳入约束与门禁，并作为 P0 验收项。
  - 202:258 契约漂移门禁：本蓝图在生成链与 preflight 中接入差异报告与阻断。
- 对齐 203（模块化单体与 DDD 边界）
  - Manifest/Slot 增加 `bcId` 字段；注册/校验以 BC 为边界，禁止跨 BC 未授权覆盖。
  - 清单与插槽放置/命名约定与 203 的 BC 目录结构一致，避免“以菜单驱动”而破坏领域边界。
- 对齐 240（职位管理页面重构与稳定化）
  - 插槽注入遵循“最小侵入”：不改变 DOM/testid；复跑 240 的 E2E 作为 P0 验收一部分。
  - OBS 事件命名与 240D 的约定保持一致，证据由 E2E/CI 落盘采集。
- 对齐 200/201（Go/工程最佳实践）
  - 避免重型黑盒框架，优先 headless + 自研映射（TanStack Table + Canvas；RHF+Zod）。
  - 生成链与门禁减少手工粘连，提升长期稳定性与可回归性；以单一事实来源驱动 UI 元数据，杜绝逻辑分叉。

---

## 15. 可视化与 UI 内创建（Studio）

目标：对标主流平台（Salesforce/SF、SuccessFactors、ServiceNow、Odoo Studio），在不破坏 SSOT/门禁前提下，提供“管理员可视化配置与 UI 内字段创建”的能力。分层实现，优先表现层与元数据编辑，契约与持久化仍走规范化生成/审阅流程。

15.1 Studio 模块（管理端 UI）
- 入口：`/admin/studio`（仅 ADMIN/CONFIG_MANAGER scope 可见）
- 能力：
  - 字段管理：创建/复制字段（选择实体、预设/复制来源、可空/默认/索引、分组/顺序）；生成“契约建议补丁 + 迁移脚本”，不直接改 SSOT。
  - 布局编辑：所见即所得（分组/页签/顺序）；生成 Manifest 布局 hints 片段，提交 PR 审阅并合入。
  - 规则与显隐：可视化规则编辑器（必填/范围/条件显隐/跨字段依赖）；生成受限 JSON 规则，映射到 Zod/RHF 与后端校验。
  - 字段级权限（FLS/FOV）：基于 OpenAPI `x-fls` 的矩阵（角色×字段×读/写）；前端消费实现显隐/只读；后端 PBAC 强校验。
  - 预览与模拟：按角色/租户/场景模拟页面布局、表单校验与列表列显示。
- 变更流：
  - Dev/Staging：Studio 生成文件落盘到 `scripts/fields/out/` 与 `frontend/src/shared/extension/layout-patches/`；由 PR 合并，CI 门禁与 215 登记。
  - Prod：禁直接改 SSOT；仅允许导出配置包（见 15.4），走迁移/灰度发布。

15.2 规则 DSL（限制型 JSON）
- 存放：OpenAPI `x-rules` 或 `docs/reference/rules/*.json`（SSOT 通过生成/审阅合入）
- 能力：必填/长度/模式/范围/条件显隐/跨字段依赖（表达式受限，避免脚本注入），统一映射到：
  - 前端：Zod schema 与 UI 显隐
  - 后端：请求体验证（返回一致的错误码/消息）
- 门禁：规则变更纳入 preflight 与契约漂移检查；表达式白名单校验。

15.3 运行时扩展通道（Flexfield/EAV 兼容）
- 设计：为核心实体提供 “customAttributes JSONB + metadata 表” 的扩展通道（SSOT 为 GraphQL 的 `customAttributes: JSON` 与 REST `extensions` 字段），适用于租户级快速字段。
- 约束：扩展元数据视作“配置数据”，由迁移/审计治理（215 登记）；不可取代核心字段；对性能/索引设定限额。
- 前端：生成 FieldDescriptor → 自动渲染；列表/筛选受限（明确何种类型可被索引/过滤）。

15.4 配置包与推广（Packaging）
- 导出：Studio 支持导出“配置包”（字段/布局/规则/FLS 元数据 JSON + 迁移 SQL 草案 + 契约建议补丁）
- 推广：Dev → Staging → Prod，通过 PR 合并、迁移执行与 215 登记；支持回滚（包内含反向补丁与 Down SQL）。
- 签名与版本（P2）：为远程功能包与配置包引入签名/版本/依赖/回滚策略；白名单加载。

15.5 安全与审计
- 权限：仅 ADMIN/CONFIG_MANAGER 可访问 Studio；所有操作审计落盘（不含密钥/隐私数据）。
- 校验：Studio 变更需通过 preflight（端点收敛/契约漂移/禁直连/清单结构）与本地模拟；拒绝产生第二事实来源。
