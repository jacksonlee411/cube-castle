# 59号文档：工具 & 验证模块质量分析

## 背景与唯一事实来源
- 本文聚焦 `frontend/src/shared/utils/` 与 `frontend/src/shared/validation/` 目录下的 9 个工具/验证文件：`colorTokens.ts`、`organization-helpers.ts`、`organizationPermissions.ts`、`statusUtils.ts`、`temporal-converter.ts`、`temporal-validation-adapter.ts`、`index.ts`、`validation/index.ts`、`validation/schemas.ts`。
- 所有结论均直接来源于上述源码，并与运行手册中关于 Canvas Token 复用、`docs/api/openapi.yaml` 的字段约束交叉核对，确保不引入第二事实来源。

## 代码质量评估
- `colorTokens.ts`：提供 base/status/legacy 三层映射，便于兼容历史样式，但全部颜色值以硬编码字符串维护，与 Canvas 官方 token 已有的数据重复。
- `organization-helpers.ts`：父级编码归一化逻辑集中，`coerceOrganizationLevel`/`getDisplayLevel` 避免 NaN，但 `normalizeParentCode` 对根节点的注释与返回值不符，且未校验非根输入是否满足 7 位数字合同。
- `organizationPermissions.ts`：角色与 scope 两套判定函数共存，基础判断清晰；然而删除权限阻断逻辑重复且 `reason` 字段可能被二次覆盖。
- `statusUtils.ts`：状态元数据集中维护，API 一致性良好，但 `colors.soap600` 等键未在 tokens 中声明类型，且 `getAvailableActions` 返回的操作枚举未与现有命令服务契约对齐。
- `temporal-converter.ts`：时态转换函数覆盖全面，不过 `dateToIso` 等函数对字符串输入直接返回原值，未真正标准化为 ISO 8601，`normalizeTemporalFields` 对无效字符串抛错但缺少调用侧的降噪策略。
- `temporal-validation-adapter.ts`：封装 `TemporalConverter` 能力，便于遗留调用，但多数函数只是简单转调，缺乏差异化逻辑。
- `shared/utils/index.ts`：集中导出 Temporal 工具，并额外提供 `DateUtils`、`TEMPORAL_UTILS_INFO`；导出结构清晰，但信息常量与迁移注释冗长。
- `shared/validation/index.ts`：统一导出验证核心，便于消费层调用；`VALIDATION_SYSTEM_INFO` 冗余输入输出描述，与代码执行无关。
- `shared/validation/schemas.ts`：整合 Zod Schema 与工具函数，整体结构完整；但若干枚举未与 OpenAPI 契约同步，并在 `ValidationUtils.temporal` 中重复调用被标记为弃用的适配器。

## 主要问题
- **契约缺失值**：`OrganizationUnitSchema` 与 `CreateOrganizationInputSchema` `unitType` 枚举缺少契约中存在的 `COMPANY`，`status` 未包含查询服务 GraphQL 响应使用的 `DELETED`，与 `statusUtils.ts` 的配置不一致，容易导致前端解析失败。
- **根节点编码语义混乱**：`normalizeParentCode.forAPI` 注释写明“确保根组织使用 "0"”，实际返回 `ROOT_PARENT_CODE`（`0000000`）；`isRootParentCode` 对空字符串直接视为根，放大数据漂移风险。
- **错误信息泄露风险**：`TemporalConverter.dateToIso` 在处理字符串失败时直接抛出原字符串内容，配合 `TemporalUtils` 在多处被直接暴露给表单，缺乏统一的用户提示封装。
- **循环依赖倾向**：`validation/schemas.ts` 重新导出 `ValidationUtils.temporal`，内部又调用 `validateTemporalDate`（来自 `temporal-validation-adapter.ts`），与 `shared/utils/index.ts` 声明的“统一来源”理念相悖，迁移状态不明确。

## 过度设计分析
- `colorTokens.ts` 维护 base/status/legacy 多套结构，并附带大量注释强调“安全颜色”，实际上等价于 Canvas token 的直接引用，可视为重复抽象。
- `shared/utils/index.ts`、`shared/validation/index.ts` 内的 `*_INFO` 常量记录迁移状态、性能优化等元信息，但未被代码或文档消费，保留意义有限。
- `TemporalConverter` 与 `TemporalUtils`、`validateTemporalDate` 三层封装提供相同能力，造成调用方难以区分使用场景，属“虚拟层”过多。

## 重复造轮子情况
- 颜色 token 与状态颜色：`colorTokens.ts` 中的硬编码值与 `@workday/canvas-kit-react/tokens` 中的官方 token 重复。
- 权限判定：`getOperationPermissions` 与 `getOperationPermissionsByScopes` 对 `childCount`、删除限制逻辑重复维护，缺乏共享校验。
- 时间/日期工具：`TemporalConverter`、`TemporalUtils`、`validateTemporalDate`、`ValidationUtils.temporal` 四处封装相似函数，相当于在原生 Date API 上重复造轮子。
- 验证错误聚合：`ValidationUtils.validateForm` 重新构造错误 Map，与 `ValidationError` 结构及 `ValidationUtils.validateCreateInput` 已输出的错误列表重复。

## 改进建议
- **对齐契约枚举**：同步 `COMPANY`、`DELETED` 等 Enum 至所有 Schema/状态工具，确保 GraphQL/REST 响应不会被错误拒绝；补充对应单元测试。
- **收敛根节点处理**：统一规定根级组织编码（建议使用 `0000000`），更新注释与 `isRootParentCode` 判断，避免将空字符串/`null` 直接视为合法值；对非根输入引入 7 位校验。
- **折叠时态工具栈**：以 `TemporalConverter` 为唯一实现，视需求保留轻量导出（如 `TemporalUtils`），移除 `validateTemporalDate` 与 `ValidationUtils.temporal` 的重复实现，集中处理错误信息与格式化策略。
- **复用 Canvas Token**：将状态/颜色引用统一改为官方 token，保留必要的兼容映射（如 legacy key），并在 Storybook/主题文件中验证颜色一致性。
- **精简信息常量**：将 `TEMPORAL_UTILS_INFO`、`VALIDATION_SYSTEM_INFO` 迁移到文档或 README，代码内仅保留必要导出，降低 bundle 体积。
- **统一权限判定**：提取共享的 `canDelete`/`reason` 逻辑，明确 role 与 scope 版本的迁移路径，避免双维护。

## 验收标准
- [ ] Schema、状态工具与权限逻辑涵盖契约枚举（含 `COMPANY`、`DELETED`），新增测试覆盖非法值与边界输入。
- [ ] `normalizeParentCode` 与 `isRootParentCode` 的行为、注释、契约完全一致，空字符串不再被视为有效根值。
- [ ] 时态工具仅保留一套核心实现，其他层对其进行透明 re-export，不再出现循环引用或重复调用；错误信息以用户友好文本暴露。
- [ ] 颜色配置改为引用 Canvas token，保留必要映射；Storybook 或单元测试验证主流程颜色引用正确。
- [ ] `*_INFO` 常量与迁移状态记录移出运行时代码，改在说明文档中维护，确保打包产物与诊断输出精简。

## 一致性校验说明
- 全部问题与建议均在源码层复核，并与 `docs/api/openapi.yaml`、设计系统约定（Canvas Tokens）比对，确保唯一事实来源链路。
- 文档存放于 `docs/development-plans/`，后续整改完成后请按流程归档至 `docs/archive/development-plans/`，并在验收记录中引用本文，以维持一致性闭环。
