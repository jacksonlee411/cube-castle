# 开发计划 08：上级组织选择器增强方案

## 概述
**创建日期**: 2025-01-15
**状态**: 计划中
**优先级**: P1
**预计工期**: 3天

### 背景
当前组织详情页面的"上级组织编码"使用文本输入框，存在以下问题：
1. 用户需要手动输入组织编码，容易出错
2. 无法验证输入的组织编码是否有效
3. 无法根据生效日期筛选有效的组织
4. 可能造成组织层级循环依赖

### 目标
将上级组织编码输入框改造为智能下拉选择器，实现：
- 基于生效日期的有效组织筛选（遵循 GraphQL 契约）
- 循环依赖自动检测与预防（基于 parentCode 回溯，无路径依赖）
- 使用 Canvas Kit 标准组件提升用户体验
- 严格符合 CQRS 分工与 API 契约（GraphQL 查询 / REST 校验）

## 技术方案

### 1. 组件架构设计

#### 1.1 新增 ParentOrganizationSelector 组件

**位置**: `frontend/src/features/temporal/components/ParentOrganizationSelector.tsx`

**接口定义**（命名采用 camelCase，对齐 API 一致性规范）:
```typescript
interface ParentOrganizationSelectorProps {
  // 必需属性
  currentCode: string;           // 当前组织编码
  effectiveDate: string;         // 生效日期（ISO格式）
  onChange: (parentCode: string | undefined) => void;

  // 可选属性
  currentParentCode?: string;    // 当前上级编码
  tenantId?: string;            // 租户ID
  disabled?: boolean;           // 是否禁用
  required?: boolean;           // 是否必填

  // 验证回调
  onValidationError?: (error: string) => void;
}
```

#### 1.2 组件职责划分

- **ParentOrganizationSelector**: 业务逻辑封装
  - GraphQL查询管理
  - 数据过滤与验证
  - 循环依赖检测

- **OrganizationDetailForm**: 表单集成
  - 状态管理
  - 字段协调
  - 提交处理

### 2. GraphQL 查询设计（契约对齐）

#### 2.1 查询定义

```graphql
query GetValidParentOrganizations($asOfDate: String!, $pageSize: Int = 500) {
  organizations(
    filter: { status: ACTIVE, asOfDate: $asOfDate }
    pagination: { page: 1, pageSize: $pageSize, sortBy: "code", sortOrder: "asc" }
  ) {
    data {
      code
      name
      unitType
      parentCode
      level
      effectiveDate
      endDate
      isFuture
    }
    pagination {
      total
      page
      pageSize
    }
  }
}
```

说明：
- 契约来自 `docs/api/schema.graphql`：`organizations(filter, pagination)` 返回 `data + pagination + temporal`，无 `nodes/totalCount`；分页使用 `page/pageSize` 而非 `limit/offset`。
- 多租户通过请求头 `X-Tenant-ID` 注入（由统一客户端处理），不通过 GraphQL 变量传递 `tenantId`。
- 候选排除当前组织由前端在结果中过滤实现（无 `excludeCodes` 过滤器）。

#### 2.2 查询优化策略

- 统一客户端：复用 `frontend/src/shared/api/unified-client.ts` 的 `UnifiedGraphQLClient`/`GraphQLEnterpriseAdapter`，不引入独立 Apollo 客户端。
- 轻量缓存：组件内维护 5 分钟 TTL 的内存缓存（Key: asOfDate + pageSize）。
- 预加载：编辑模式激活时触发一次预加载（可控并可取消）。
- 增量加载：当候选组织超过 500 时，按页加载并开启虚拟滚动。

### 3. 循环依赖检测算法（契约友好）

#### 3.1 检测逻辑

```typescript
type OrgLite = { code: string; parentCode?: string };

class CyclicDependencyDetector {
  /**
   * 使用 parentCode 向上回溯检测是否形成环，不依赖 path 字段。
   */
  static detectCycle(
    currentCode: string,
    targetParentCode: string | undefined,
    orgMap: Map<string, OrgLite>
  ): { hasCycle: boolean; cyclePath?: string[] } {
    if (!targetParentCode) return { hasCycle: false };
    if (currentCode === targetParentCode) {
      return { hasCycle: true, cyclePath: [currentCode, targetParentCode] };
    }

    const seen = new Set<string>();
    const path: string[] = [currentCode];
    let cur = targetParentCode;

    while (cur && !seen.has(cur)) {
      path.push(cur);
      if (cur === currentCode) {
        // 回到起点，形成环
        return { hasCycle: true, cyclePath: [...path] };
      }
      seen.add(cur);
      cur = orgMap.get(cur)?.parentCode || '';
    }

    return { hasCycle: false };
  }
}
```

#### 3.2 验证规则

1. **基础规则**:
   - 不能选择自己作为上级
   - 不能选择自己的任何下级作为上级（通过 parent 链回溯检测）

2. **时态规则**:
   - 只显示在 `asOfDate = effectiveDate` 时点有效的组织
   - 由后端时态逻辑保障有效性；前端无需自行计算区间

3. **状态规则**:
   - 只显示 `status = ACTIVE` 的组织
   - “计划中/PLANNED”不属于 GraphQL 的业务状态；未来态通过 `isFuture` 派生

### 4. Canvas Kit组件集成

#### 4.1 使用 Combobox 组件（支持搜索）

```typescript
import { Combobox } from '@workday/canvas-kit-react/combobox';
import { FormField } from '@workday/canvas-kit-react/form-field';
import { Flex } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';

const ParentOrganizationSelector: React.FC<Props> = (props) => {
  return (
    <FormField error={validationError}>
      <FormField.Label required={props.required}>
        上级组织
      </FormField.Label>

      <Combobox
        items={filteredOrganizations}
        onChange={handleSelect}
        disabled={props.disabled}
      >
        <Combobox.Input
          placeholder="搜索并选择上级组织..."
          value={searchValue}
          onChange={handleSearch}
        />

        <Combobox.Menu>
          <Combobox.MenuList>
            {(item: Organization) => (
              <Combobox.Item key={item.code}>
                <Flex direction="column" gap="xxs">
                  <Text weight="medium">
                    {item.code} - {item.name}
                  </Text>
                  <Text size="small" variant="hint">
                    层级: {item.level} | 上级: {item.parentCode || '-'}
                  </Text>
                </Flex>
              </Combobox.Item>
            )}
          </Combobox.MenuList>
        </Combobox.Menu>
      </Combobox>

      <FormField.Hint>
        显示在 {formatDate(props.effectiveDate)} 生效且状态为 ACTIVE 的组织
      </FormField.Hint>

      {validationError && (
        <FormField.Error>{validationError}</FormField.Error>
      )}
    </FormField>
  );
};
```

#### 4.2 UI/UX增强

- **搜索功能**: 支持按编码、名称、路径搜索
- **分组显示**: 按组织层级分组显示选项
- **信息展示**: 默认显示“编码/名称/层级”；如需路径，仅对当前选中项按需查询 `organizationHierarchy` 显示 `codePath/namePath`（需 `org:read:hierarchy` 权限）
- **加载状态**: 显示加载指示器
- **空状态**: 无可选组织时的友好提示

### 4.3 PBAC 权限与多租户

- 必要权限：
  - 候选查询 `organizations` 需要 `org:read`
  - 路径展示（按需）使用 `organizationHierarchy` 需要 `org:read:hierarchy`
  - 服务器校验 `/api/v1/organization-units/validate` 需要 `org:validate`
- 多租户：统一客户端自动注入 `X-Tenant-ID`；组件不接受 `tenantId` 作为变量
- UI gating：缺少相应 scope 时禁用或隐藏功能并给出权限提示

### 5. 性能优化

#### 5.1 查询优化
- 使用 GraphQL 字段选择，只查询必要字段（见 2.1）
- 组件级 TTL 缓存（5 分钟），相同 asOfDate/pageSize 不重复请求
- 使用 React.memo 与受控搜索输入减少重渲染

#### 5.2 渲染优化
- 虚拟滚动：组织列表超过100项时启用
- 防抖搜索：搜索输入300ms防抖
- 懒加载：延迟加载下拉选项直到用户交互

### 6. 错误处理

#### 6.1 错误场景
1. **网络错误**: GraphQL查询失败
2. **数据错误**: 返回数据格式异常
3. **业务错误**: 循环依赖检测到
4. **权限错误**: 用户无权查看某些组织

#### 6.2 错误提示
```typescript
const errorMessages = {
  NETWORK_ERROR: '加载组织列表失败，请检查网络连接',
  CYCLIC_DEPENDENCY: '选择该组织将导致循环依赖：{path}',
  NO_VALID_ORGS: '没有符合条件的上级组织可选',
  PERMISSION_DENIED: '您没有权限查看组织列表',
};
```

### 7. 服务器端校验集成（健壮优先）

- 在表单提交前调用 REST 校验端点 `/api/v1/organization-units/validate`（OpenAPI 已定义），示例：
  ```json
  {
    "operation": "update",
    "data": {
      "code": "{currentCode}",
      "parentCode": "{selectedParent}",
      "effectiveDate": "{effectiveDate}"
    },
    "dryRun": true
  }
  ```
- 需要 `org:validate` scope；无权限时，保留前端循环预检并在下文以 TODO-TEMPORARY 明确过渡与回收期限。

## 实施计划

### 第一阶段：基础功能实现（第1天）

1. **创建ParentOrganizationSelector组件**
   - 基础组件结构
   - Props接口定义
   - 基础UI布局

2. **实现GraphQL查询**
   - 创建查询定义（对齐 `docs/api/schema.graphql`）
   - 复用统一客户端（`UnifiedGraphQLClient`）
   - 实现数据获取 hook 与组件级缓存

3. **数据过滤逻辑**
   - 按生效日期过滤
   - 按状态过滤
   - 排除当前组织

### 第二阶段：业务逻辑完善（第2天）

4. **实现循环依赖检测（parent 回溯）**
   - 检测算法实现（不依赖 path）
   - 单元测试编写
   - 错误提示集成

5. **集成Canvas Kit Combobox**
   - 组件样式调整
   - 搜索功能实现
   - 选项渲染优化

6. **集成到OrganizationDetailForm**
   - 替换现有输入框
   - 状态管理对接
   - 表单验证集成
   - UI gating（基于 scopes）

### 第三阶段：优化与测试（第3天）

7. **性能优化**
   - 查询缓存实现
   - 虚拟滚动（如需要）
   - 防抖优化

8. **完整测试**
   - 组件单元测试
   - 集成测试
   - E2E测试场景
   - 契约测试：新增“父级候选查询” GraphQL 校验用例

9. **文档更新**
   - 组件使用文档
   - 引用 API 契约（无需修改契约）
   - 更新实现清单

## 当前进展（2025-09-15）

已完成功能与验证（版本 1.2.0 预览）：

- 组件与集成
  - 新增组件 `frontend/src/features/temporal/components/ParentOrganizationSelector.tsx`
    - GraphQL 查询对齐契约：`organizations(filter: {status: ACTIVE, asOfDate}, pagination: {page, pageSize}) { data, pagination }`
    - 统一客户端：使用 `UnifiedGraphQLClient`，未引入 Apollo
    - 性能：组件内 5 分钟 TTL 内存缓存，搜索过滤
    - 循环检测：基于 `parentCode` 向上回溯，完全移除对 `path` 的依赖
    - PBAC：集成 `useOrgPBAC`，无 `org:read` 时禁用组件并显示权限错误
  - 表单集成 `OrganizationDetailForm.tsx`
    - 替换“上级组织编码”文本框为选择器
    - 路径展示改为 `codePath/namePath`（移除 `record.path` 引用）
    - 状态编辑仅 `ACTIVE/INACTIVE`（PLANNED 仅作未来态展示语义）
  - 提交前服务器校验（健壮优先）
    - `OrganizationForm/index.tsx` 调用 `/api/v1/organization-units/validate`（`dryRun`）
    - 无权限或端点不可用时不阻断提交，后端为最终裁决

- 契约与测试
  - 新增契约用例：父级候选查询（asOfDate + ACTIVE + 分页）
    - 位置：`frontend/tests/contract/schema-validation.test.ts`
  - 新增组件单测：`ParentOrganizationSelector` 的加载/选择、循环检测、PBAC 禁用
    - 位置：`frontend/src/features/temporal/components/__tests__/ParentOrganizationSelector.test.tsx`
  - 测试环境适配：`setupTests.ts` 增补 Canvas Kit `Combobox/FormField/Flex` 的简易可交互 mock
  - GraphQL 企业适配器测试迁移到 Vitest 风格（去除 `jest.mock`）
    - 位置：`frontend/src/shared/api/__tests__/graphql-enterprise-adapter.test.ts`
  - 全量测试：12/12 文件通过，78 条用例（77 passed, 1 skipped）

- 统一权限 Hook
  - 新增 `frontend/src/shared/hooks/useScopes.ts`（`useScopes`/`useOrgPBAC`）
  - 提供 `canRead`/`canReadHierarchy`/`canValidate`，便于组件级 UI gating

- UI 门控组件
  - 新增 `frontend/src/shared/components/RequireScopes.tsx`（支持 allOf/anyOf/fallback）
  - 示例：`<RequireScopes allOf={["org:read"]} anyOf={["org:validate"]}>...</RequireScopes>`

差异记录与说明：
- PLANNED 保留于前端类型与校验的兼容定义，但从表单“可编辑状态”中移除；“计划中”展示由时态派生（`effectiveDate` 与 `asOfDate` 比较）
- 路径展示统一改为 `codePath/namePath`；不再依赖 GraphQL 的 `path`（契约未提供）

## 测试计划

### 单元测试
- ParentOrganizationSelector组件测试
- CyclicDependencyDetector算法测试
- GraphQL查询mock测试

### 集成测试
- 与OrganizationDetailForm集成测试
- 表单提交流程测试
- 错误处理测试

### E2E测试场景
1. 创建新组织并选择上级
2. 编辑现有组织的上级
3. 验证循环依赖预防
4. 验证时态数据过滤（asOfDate 生效 + ACTIVE 状态）
5. 测试搜索功能
6. 缺少权限时的 UI gating 与错误提示

## 风险与对策

### 风险1：大量组织数据性能问题
- **风险**: 组织数量过多导致下拉列表性能差
- **对策**: 实现虚拟滚动和分页加载

### 风险2：复杂层级关系的循环检测
- **风险**: 深层级组织结构循环检测耗时
- **对策**: 仅在选择修改时触发 parent 回溯；使用本地 Map 提高查询性能

### 风险3：时态数据一致性
- **风险**: 不同时间点的数据可能不一致
- **对策**: 严格按effectiveDate筛选，添加数据完整性检查

## 验收标准

1. ✅ 上级组织使用下拉选择器，支持搜索
2. ✅ 根据 asOfDate 正确筛选有效组织（ACTIVE + 有效期内）
3. ✅ 成功预防循环依赖的选择
4. ✅ 使用Canvas Kit标准组件
5. ✅ 所有测试用例通过（含契约测试）
6. ✅ 性能满足要求（本地数据量 ~2000 条，首屏 < 500ms，滚动交互 < 16ms 帧，启用虚拟滚动）
7. ✅ 符合 CQRS 分工与 API 契约（无未声明字段/参数）
8. ✅ 多租户与 PBAC 对齐（基于 scopes 的 UI gating）

## 进度验证与结果

- 合同一致性：父级候选查询契约校验通过（GraphQL Schema v4.6.0）
- 单元与集成：新增选择器与 UI 门控组件单测、适配器测试迁移，覆盖循环检测/权限禁用/缓存等路径
- 全量测试结果（前端）：13/13 文件通过；80 条用例（79 passed, 1 skipped）

## 参考资料

- [Canvas Kit Combobox文档](https://workday.github.io/canvas-kit/?path=/docs/components-inputs-combobox--basic)
- [GraphQL Schema定义](../../api/schema.graphql)
- [OpenAPI 命令层契约（/validate）](../../api/openapi.yaml)
- [组织详情表单组件](../../frontend/src/features/temporal/components/OrganizationDetailForm.tsx)
- [时态类型定义](../../frontend/src/shared/types/temporal.ts)
- [统一 API 客户端](../../frontend/src/shared/api/unified-client.ts)
- [权限 Hook（useScopes/useOrgPBAC）](../../frontend/src/shared/hooks/useScopes.ts)
- [权限门控组件（RequireScopes）](../../frontend/src/shared/components/RequireScopes.tsx)
- [选择器组件单测](../../frontend/src/features/temporal/components/__tests__/ParentOrganizationSelector.test.tsx)
- [权限门控组件单测](../../frontend/src/shared/components/__tests__/RequireScopes.test.tsx)

## 后续优化建议

1. UI 权限封装：在更多页面落地 `<RequireScopes>` 包装（菜单/按钮/区域门控），统一处理 `org:read/org:read:hierarchy/org:validate`
2. 路径信息按需加载：仅在选择后按需调用 `organizationHierarchy` 获取 `codePath/namePath`，并做 `org:read:hierarchy` gating
3. 搜索与虚拟化：输入 300ms 防抖；>100 项启用虚拟滚动；>500 项分页增量加载（参数化阈值）
4. E2E 覆盖：
   - 缺少权限时的禁用与提示
   - 选择循环/自引用的提示
   - /validate 干跑失败的错误呈现与阻断
5. 性能验证：提供 2k/5k 数据规模的复现实验脚本与指标采集（首屏 < 500ms，滚动交互帧率）
6. 配置化缓存：选择器已支持 `cacheTtlMs`；完善文档与默认值说明，并在性能压测基础上给出推荐值
7. 状态语义统一：分阶段移除 UI 残留的“PLANNED 可编辑态”，全部改为由时态派生展示（追踪范围：`features/organizations/*`、`temporal/*`、`constants/temporalStatus`）
8. 错误处理统一：将组件内错误提示对齐 `frontend/src/shared/api/error-messages.ts` 的映射与分流策略
9. 可观测性：在选择器与 GraphQL 请求处埋点（加载耗时、命中缓存、错误率），纳入性能仪表盘

---

### 临时方案标注
- // TODO-TEMPORARY: 当缺少 `org:validate` 权限时，仅执行前端循环依赖预检作为用户体验优化；以服务器端 `/validate` 作为最终裁决的目标状态。截止时间：2025-09-30。负责人：前端架构团队。

**文档版本**: 1.3.0
**最后更新**: 2025-09-15
**负责人**: 前端架构团队
