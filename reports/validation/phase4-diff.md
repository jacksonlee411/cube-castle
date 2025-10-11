# Phase 4 Validation Diff (Backend vs Frontend vs Contract)

**生成日期**: 2025-10-12  
**范围**: 对比第一阶段契约生成物 (`shared/contracts/organization.json` / `contract_gen.go|ts`) 与当前实现（后端 `internal/utils/validation.go`、前端 `src/shared/validation/schemas.ts`）的主要差异。  
**目的**: 为 Phase 4 统一校验任务提供基线，指导后续整改。

---

## 1. 枚举取值

| 枚举 | 契约 `organization.json` | 前端 `schemas.ts` | 后端 `validation.go` | 差异说明 |
|------|---------------------------|-------------------|----------------------|----------|
| unitType | `DEPARTMENT` / `ORGANIZATION_UNIT` / `COMPANY` / `PROJECT_TEAM` | ✅ 同契约（通过 `OrganizationUnitTypeEnumValues`） | ✅ 同契约（`validUnitTypes` 包含四类） | - |
| status | `ACTIVE` / `INACTIVE` / `PLANNED` / `DELETED` | ✅ 同契约（`statusValues`） | ✅ 同契约 | - |
| operationType | `CREATE` / `UPDATE` / `SUSPEND` / `REACTIVATE` / `DEACTIVATE` / `DELETE` | 未使用 | 未使用 | 契约枚举完整，当前层未直接使用 |

---

## 2. 字段约束

| 字段 | 契约约束 (`OrganizationConstraints`) | 前端实现 | 后端实现 | 差异 / 备注 |
|------|--------------------------------------|-----------|-----------|-------------|
| code | 正则 `^[1-9][0-9]{6}$`（7 位、不含前导 0） | ✅ 使用契约正则 `codePattern` | ✅ 使用契约正则 `OrganizationCodePattern` | - |
| parentCode | 正则 `^(0|[1-9][0-9]{6})$` | ✅ 使用契约正则 `parentCodePattern` | ✅ 使用契约正则 `OrganizationParentCodePattern` | - |
| name | `maxLength:255` | `max(255)`、`min(1)` | `max(255)`、非空校验 | - |
| description | `maxLength:1000` | `max(1000)` | `<=1000` | - |
| level | `min:1`、`max:17` | `min(1)`、`max(17)` | 未直接校验（依赖其他逻辑） | 后端待显式校验或文档说明 |
| sortOrder | 默认 0 | `min(0)`、默认 0 | `0 <= sortOrder <= 9999` | 契约仅声明默认值，无上限；保留后端上限或在契约中记录 |
| operationReason | `maxLength:500` | `max(500)` + 最小长度 5 的业务规则 | `<=500`（最小长度 5） | 已统一；最小长度为额外业务规则 |
| effectiveDate | 格式 `date` | 自定义校验（校验 YYYY-MM-DD） | 比较 `EffectiveDate` / `EndDate` | 校验方式一致，契约无额外限制 |
| endDate | 契约未定义（可选） | 与 `effectiveDate` 同步校验 | 与 `EffectiveDate` 比较 | 契约侧无专门约束 |

---

## 3. 工具函数与额外校验

- **前端**：`schemas.ts` 中引入自定义 Zod Schema（如 `TemporalFormSchema`）以及 `validateTemporalDate`、`FutureDateSchema` 等；部分逻辑（例如 `Past/Future` 判断）未体现于契约，需要确认是否保留。
- **后端**：`NormalizeParentCodePointer`、`ValidateTemporalParentAvailability` 等业务验证并非纯粹格式限制，属于业务层逻辑，可保留。

---

## 4. 建议与后续行动

1. **复用契约生成物**  
   - （已执行）前端以 `OrganizationConstraints` / `OrganizationUnitTypeEnumValues` 等常量驱动 Zod Schema。  
   - 后端若维持 3-10 位代码逻辑，需要确认契约是否需要同步（或调整后端以契约为准）。

2. **决定统一标准**  
   - 后端目前未显式校验 `level` 上限（依赖其他逻辑），如需落地可考虑引用契约常量或在文档注明。  
   - 对仍保留的附加业务规则（如最小长度、排序上限等）需在文档中说明原因，避免与契约产生认知偏差。

3. **记录决定**  
   - 上述差异与处理结果请更新至 Phase 4 验收草案（66 号文档）以及 06 号进展日志，确保后续阶段可追踪。

---

**备注**：本文件仅作为对比基线；实际执行中若发现新的差异，请追加记录并更新处理结论。
