# 92号文档：职位管理二级导航实施方案（Canvas Kit 官方组件）

**版本**: v2.3
**创建日期**: 2025-10-19
**最新更新**: 2025-10-19
**维护团队**: 前端团队 + 后端团队
**状态**: 技术方案确认 (推荐使用 Canvas Kit 官方组件) ✅
**关联计划**: 80号职位管理模块设计方案
**参考系统**: Workday Canvas Kit Expandable + Side Panel
**遵循原则**: CLAUDE.md 资源唯一性与跨层一致性原则（最高优先级）

---

## 0. Canvas Kit 官方方案研究结论 ⭐

### 0.1 技术选型对比

经过深入研究 Workday Canvas Kit 官方文档和组件库，确认以下**官方推荐方案**：

| 对比项 | 自定义方案（v1.0草案） | **Canvas Kit 官方方案（推荐）** |
|--------|----------------------|------------------------------|
| 折叠组件 | 自定义 NavigationItem | ✅ **Expandable** 复合组件 |
| 状态管理 | 手动 useState | ✅ **DisclosureModel** 自动管理 |
| 无障碍支持 | 手动添加 ARIA | ✅ **自动** aria-expanded/controls |
| 图标动画 | 自定义 CSS | ✅ **内置** chevron 旋转动画 |
| 组件复用性 | 低（项目特定） | ✅ **高**（Workday 标准） |
| 维护成本 | 高（需持续维护） | ✅ **低**（官方维护升级） |
| 设计一致性 | 需手动对齐 | ✅ **自动**符合 Canvas Kit 规范 |
| 包依赖 | 无需额外安装 | ✅ **已安装** `@workday/canvas-kit-react@13.2.15` |

**结论**：**强烈推荐使用 Canvas Kit 官方 Expandable 组件**，避免自定义实现。

### 0.2 官方组件可用性确认

✅ **已在项目中可用**（无需额外安装）：

```bash
# 项目当前已安装
@workday/canvas-kit-react@13.2.15

# 可用组件
node_modules/@workday/canvas-kit-react/expandable  ✅
node_modules/@workday/canvas-kit-react/side-panel  ✅
```

**导入路径**：
```typescript
import { Expandable } from '@workday/canvas-kit-react/expandable';
import { SidePanel } from '@workday/canvas-kit-react/side-panel';
```

### 0.3 官方组件 API 架构

**SidePanel + Expandable 组合模式**：
```typescript
<SidePanel
  open
  header="组件"
  openWidth={312}
  backgroundColor={SidePanel.BackgroundColor.Gray}
>
  <Box as="nav" display="flex" flexDirection="column" gap={space.xxs}>
    <NavigationButton active={false} onClick={() => navigate('/dashboard')}>
      <SystemIcon icon={dashboardIcon} size={20} />
      仪表板
    </NavigationButton>

    <Expandable initialVisibility="visible">
      <ExpandableTrigger headingLevel="h3" active>
        <Expandable.Icon iconPosition="start" />
        <SystemIcon icon={viewTeamIcon} size={20} />
        <Expandable.Title>职位管理</Expandable.Title>
      </ExpandableTrigger>
      <Expandable.Content>
        <Box
          as="ul"
          margin={space.zero}
          paddingLeft={space.zero}
          listStyle="none"
          gap={space.xxs}
        >
          <li>
            <SubNavigationButton active onClick={() => navigate('/positions')}>
              职位列表
            </SubNavigationButton>
          </li>
        </Box>
      </Expandable.Content>
    </Expandable>
  </Box>
</SidePanel>
```

> 说明：职位详情采用单一路由 `/positions/:code`，时态版本通过详情页内的页签/组件呈现；不再保留 `/positions/:code/temporal` 冗余地址，避免路由重复。

> 以上示例沿用 2.3.1 中定义的 `NavigationButton`、`ExpandableTrigger`、`SubNavigationButton` 以及 `space` tokens。

**核心特性**：
- ✅ `SidePanel` 提供官方浅灰背景、固定宽度与滚动节奏
- ✅ `Expandable.Target` 内建 `aria-expanded` 与键盘焦点管理
- ✅ `Expandable.Icon iconPosition="start"` 保证左侧 Chevron 与截图一致
- ✅ 导航按钮统一使用 Canvas tokens（space/soap/borderRadius）实现圆角高亮

### 0.4 Phase 0 技术验证清单（新增）

在 Phase 1 正式开发前新增 0.5 天的 POC 验证，确保官方组件扩展方式与项目环境完全兼容：

- [x] **Expandable.Target 自定义验证**
  - [x] `headingLevel` 属性语义化输出正常
  - [x] `styled(Expandable.Target)` 包装后 `aria-expanded` / `aria-controls` 仍由组件自动注入
  - [x] `Expandable.Icon iconPosition="start"` 正确渲染左侧 Chevron
- [x] **SidePanel 集成验证**
  - [x] 将 `AppShell` 中现有 240px `Box` 容器替换为 `SidePanel`，确认主内容区在 312px 左侧栏宽度下布局正常（包括滚动条、padding、分隔线）
  - [x] 若 312px 不符合视觉稿，记录与设计确认结果并更新 `openWidth`
  - [x] `backgroundColor={SidePanel.BackgroundColor.Gray}` 呈现浅灰背景，同时与主内容区域分层清晰
- [x] **权限 Hook 可用性**
  - [x] 在 `frontend/src/shared/auth/context.ts` / `hooks.ts` 中扩展 `AuthContext`，新增 `hasPermission(permission: string): boolean` 与 `userPermissions: string[]`
  - [x] Phase 0 前提供最小实现（可基于现有认证状态 mock），并编写单元测试覆盖基础行为
  - [x] 完成后更新 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 登记 Hook 来源
- [x] **Canvas tokens 校验**
  - [x] `colors.soap200`、`borderRadius.l`、`space` token 在构建链路中无解析警告
  - [x] 选中态、Hover 态与 Canvas Design System 保持一致

> 2025-10-19 POC 结论：`AppShell` 已切换至 312px `SidePanel` 灰色背景布局，`NavigationItem` 通过 `Expandable` 实现语义化二级导航，`AuthContext` 暴露 `hasPermission`/`userPermissions` 并通过单测验证，满足 Phase 0 准入条件。

POC 结论需在完成后记录于本方案文档的变更历史区，并作为推进 Phase 1 的准入条件。

---

## 1. 背景与目标（更新）

### 1.1 业务背景

根据80号文档第8.4节设计要求，职位管理模块需要通过**二级导航**将 Job Catalog 四层体系（职类/职种/职务/职级）呈现为独立可管理的对象模型，便于运维按 Workday 规范管理主数据。

**当前问题**：
- ❌ Sidebar 仅支持一级导航，无法展示子菜单
- ❌ 缺少 Job Catalog 四层的独立页面和路由
- ❌ 用户无法直接访问和管理职类、职种、职务、职级数据

### 1.2 设计目标

1. **实现可折叠二级菜单**：扩展 Sidebar 组件，支持职位管理下的5个子菜单项
2. **复用组织管理模式**：借鉴 OrganizationDashboard 的表格、筛选器、表单组件
3. **遵循 Canvas Kit 规范**：使用 Workday 官方组件和视觉设计语言
4. **保持架构一致性**：CQRS 分离、权限控制、审计追踪

### 1.3 Workday 导航最佳实践

参考 Workday HCM 导航模式：
```
Setup (设置)
  └── Job Architecture (岗位架构)
      ├── Job Families (职种管理)
      ├── Job Profiles (岗位定义)
      └── Job Levels (职级管理)
```

本方案对应：
```
职位管理
  ├── 职位列表 (默认)
  ├── 职类管理 (Job Family Groups)
  ├── 职种管理 (Job Families)
  ├── 职务管理 (Job Roles)
  └── 职级管理 (Job Levels)
```

---

## 2. 核心设计

### 2.1 导航结构设计

#### 2.1.1 导航层级关系

```
一级导航（侧边栏）
├── 仪表板
├── 组织架构
├── 职位管理 ⬅️ 可展开/折叠
│   ├── 职位列表 (默认)          /positions
│   ├── 职类管理                /positions/catalog/family-groups
│   ├── 职种管理                /positions/catalog/families
│   ├── 职务管理                /positions/catalog/roles
│   └── 职级管理                /positions/catalog/levels
└── 契约测试
```

#### 2.1.2 交互行为

| 操作 | 行为 | 视觉反馈 |
|------|------|----------|
| 点击"职位管理" | 展开/折叠二级菜单 | 展开图标旋转90° |
| 首次访问 | 默认展开状态 | 高亮"职位列表" |
| 点击子菜单 | 导航到对应页面 | 高亮当前项，保持展开 |
| 访问其他一级菜单 | 折叠职位管理 | 恢复默认状态 |

### 2.2 路由规划

#### 2.2.1 路由表

| 路径 | 组件 | 说明 | 权限 |
|------|------|------|------|
| `/positions` | PositionDashboard | 职位列表（默认页） | position:read |
| `/positions/:code` | PositionTemporalPage | 职位详情 | position:read |
| `/positions/catalog/family-groups` | JobFamilyGroupList | 职类管理 | job-catalog:read |
| `/positions/catalog/family-groups/:code` | JobFamilyGroupDetail | 职类详情 | job-catalog:read |
| `/positions/catalog/families` | JobFamilyList | 职种管理 | job-catalog:read |
| `/positions/catalog/families/:code` | JobFamilyDetail | 职种详情 | job-catalog:read |
| `/positions/catalog/roles` | JobRoleList | 职务管理 | job-catalog:read |
| `/positions/catalog/roles/:code` | JobRoleDetail | 职务详情 | job-catalog:read |
| `/positions/catalog/levels` | JobLevelList | 职级管理 | job-catalog:read |
| `/positions/catalog/levels/:code` | JobLevelDetail | 职级详情 | job-catalog:read |

#### 2.2.2 路由配置示例

```typescript
// frontend/src/App.tsx
<Route path="/" element={<AppShell />}>
  {/* 职位管理模块 */}
  <Route path="/positions">
    {/* 职位列表和详情 */}
    <Route index element={<PositionDashboard />} />
    <Route path=":code" element={<PositionTemporalPage />} />
    {/* Job Catalog 四层 */}
    <Route path="catalog">
      <Route path="family-groups" element={<JobFamilyGroupList />} />
      <Route path="family-groups/:code" element={<JobFamilyGroupDetail />} />
      <Route path="families" element={<JobFamilyList />} />
      <Route path="families/:code" element={<JobFamilyDetail />} />
      <Route path="roles" element={<JobRoleList />} />
      <Route path="roles/:code" element={<JobRoleDetail />} />
      <Route path="levels" element={<JobLevelList />} />
      <Route path="levels/:code" element={<JobLevelDetail />} />
    </Route>
  </Route>
</Route>
```

### 2.3 组件架构设计（基于 Canvas Kit 官方组件）✅

#### 2.3.1 采用 SidePanel + Expandable 的二级导航实现

**核心思路**：以 Canvas Kit `SidePanel` 承载侧栏布局，使用原生 `Expandable.Target` 触发器与左侧 Chevron，复用官方导航间距和焦点态。

```typescript
// frontend/src/layout/NavigationItem.tsx
import React from 'react';
import styled from '@emotion/styled';
import {useNavigate, useLocation} from 'react-router-dom';
import {CanvasSystemIcon} from '@workday/design-assets-types';
import {Expandable} from '@workday/canvas-kit-react/expandable';
import {SystemIcon} from '@workday/canvas-kit-react/icon';
import {Box} from '@workday/canvas-kit-react/layout';
import {colors, space, borderRadius} from '@workday/canvas-kit-react/tokens';
import {TertiaryButton} from '@workday/canvas-kit-react/button';
import {useAuth} from '@/shared/hooks/useAuth';

interface SubMenuItem {
  label: string;
  path: string;
  permission?: string;
}

interface NavigationItemProps {
  label: string;
  path: string;
  icon: CanvasSystemIcon;
  subItems?: SubMenuItem[];
  permission?: string;
}

const NavigationButton = styled(TertiaryButton, {
  shouldForwardProp: prop => prop !== 'active',
})<{active: boolean}>`
  display: flex;
  align-items: center;
  gap: ${space.xs};
  width: 100%;
  border: 0;
  background: ${({active}) => (active ? colors.soap200 : 'transparent')};
  color: ${({active}) => (active ? colors.blueberry400 : colors.licorice500)};
  border-radius: ${borderRadius.l};
  padding: ${space.xs} ${space.s};
  font: inherit;
  cursor: pointer;
  text-align: left;
  &:focus-visible {
    outline: 2px solid ${colors.blueberry400};
    outline-offset: 2px;
  }
`;

const ExpandableTrigger = styled(Expandable.Target, {
  shouldForwardProp: prop => prop !== 'active',
})<{active: boolean}>`
  border-radius: ${borderRadius.l};
  background: ${({active}) => (active ? colors.soap200 : 'transparent')};
  color: ${({active}) => (active ? colors.blueberry400 : colors.licorice500)};
  gap: ${space.xs};
  padding: ${space.xs} ${space.s};
  display: flex;
  align-items: center;
`;

const SubNavigationButton = styled(NavigationButton)`
  padding: ${space.xxs} ${space.m};
  margin-left: ${space.m};
  gap: ${space.xxs};
`;

export const NavigationItem: React.FC<NavigationItemProps> = ({
  label,
  path,
  icon,
  subItems,
  permission,
}) => {
  const navigate = useNavigate();
  const location = useLocation();
  const {hasPermission} = useAuth();

  if (permission && !hasPermission(permission)) {
    return null;
  }

  const isActive = location.pathname.startsWith(path);
  const hasSubItems = Boolean(subItems?.length);

  if (!hasSubItems) {
    return (
      <NavigationButton
        type="button"
        active={isActive}
        onClick={() => navigate(path)}
        aria-current={isActive ? 'page' : undefined}
      >
        <SystemIcon icon={icon} size={20} />
        {label}
      </NavigationButton>
    );
  }

  return (
    <Expandable initialVisibility={isActive ? 'visible' : 'hidden'}>
      <ExpandableTrigger headingLevel="h3" active={isActive}>
        <Expandable.Icon iconPosition="start" />
        <SystemIcon icon={icon} size={20} />
        <Expandable.Title>{label}</Expandable.Title>
      </ExpandableTrigger>

      <Expandable.Content>
        <Box
          as="ul"
          paddingTop={space.xxs}
          paddingLeft={space.zero}
          margin={space.zero}
          display="flex"
          flexDirection="column"
          gap={space.xxs}
          listStyle="none"
        >
          {subItems?.map(subItem => {
            if (subItem.permission && !hasPermission(subItem.permission)) {
              return null;
            }

            const isSubActive = location.pathname === subItem.path;

            return (
              <li key={subItem.path}>
                <SubNavigationButton
                  type="button"
                  active={isSubActive}
                  onClick={() => navigate(subItem.path)}
                  aria-current={isSubActive ? 'page' : undefined}
                >
                  {subItem.label}
                </SubNavigationButton>
              </li>
            );
          })}
        </Box>
      </Expandable.Content>
    </Expandable>
  );
};
```

**Canvas Kit 侧栏组合优势**：

| 特性 | 自动获得 | 无需手动实现 |
|------|---------|------------|
| ✅ SidePanel 布局 | 默认浅灰背景、滚动区域 | 继承 Workday 侧栏节奏 |
| ✅ Expandable.Target | `aria-expanded` 与键盘支持 | 无需再包裹按钮 |
| ✅ Chevron 在左侧 | `Expandable.Icon iconPosition="start"` | 与设计稿一致 |
| ✅ 选中态与圆角 | 依赖官方 tokens（soap 系列 + borderRadius.l） | 保持 UI 一致性 |

**Sidebar 结构（SidePanel 作为容器）**：

```typescript
// frontend/src/layout/Sidebar.tsx
import {SidePanel} from '@workday/canvas-kit-react/side-panel';
import {Box} from '@workday/canvas-kit-react/layout';
import {space} from '@workday/canvas-kit-react/tokens';

const navigationConfig = [
  {
    label: '仪表板',
    path: '/dashboard',
    icon: dashboardIcon,
  },
  {
    label: '组织架构',
    path: '/organizations',
    icon: homeIcon,
  },
  {
    label: '职位管理',
    path: '/positions',
    icon: viewTeamIcon,
    subItems: [
      {label: '职位列表', path: '/positions', permission: 'position:read'},
      {label: '职类管理', path: '/positions/catalog/family-groups', permission: 'job-catalog:read'},
      {label: '职种管理', path: '/positions/catalog/families', permission: 'job-catalog:read'},
      {label: '职务管理', path: '/positions/catalog/roles', permission: 'job-catalog:read'},
      {label: '职级管理', path: '/positions/catalog/levels', permission: 'job-catalog:read'},
    ],
  },
  {
    label: '契约测试',
    path: '/contract-testing',
    icon: checkIcon,
  },
];

export const Sidebar: React.FC = () => (
  <SidePanel
    open
    padding={space.m}
    openWidth={312}
    backgroundColor={SidePanel.BackgroundColor.Gray}
    aria-label="主导航"
  >
    <Box as="nav" display="flex" flexDirection="column" gap={space.xxs}>
      {navigationConfig.map(item => (
        <NavigationItem key={item.path} {...item} />
      ))}
    </Box>
  </SidePanel>
);
```

> 国际化前置：为减少后续多语言改造成本，实施时应将 `navigationConfig` 的 `label` 字段接入现有 i18n 方案（如 `t('nav.positions.list')`），并在 Phase 1 完成相关语言包登记。

#### 2.3.2 Job Catalog 页面组件

**目录结构**：

```
frontend/src/features/
├── positions/
│   ├── PositionDashboard.tsx
│   ├── PositionTemporalPage.tsx
│   └── components/
│       └── ...
└── job-catalog/                    ⬅️ 新增
    ├── family-groups/
    │   ├── JobFamilyGroupList.tsx
    │   ├── JobFamilyGroupDetail.tsx
    │   └── JobFamilyGroupForm.tsx
    ├── families/
    │   ├── JobFamilyList.tsx
    │   ├── JobFamilyDetail.tsx
    │   └── JobFamilyForm.tsx
    ├── roles/
    │   ├── JobRoleList.tsx
    │   ├── JobRoleDetail.tsx
    │   └── JobRoleForm.tsx
    ├── levels/
    │   ├── JobLevelList.tsx
    │   ├── JobLevelDetail.tsx
    │   └── JobLevelForm.tsx
    ├── shared/
    │   ├── CatalogTable.tsx       # 通用表格组件
    │   ├── CatalogFilters.tsx     # 通用筛选组件
    │   └── CatalogForm.tsx        # 通用表单组件
    └── types.ts
```

**通用列表页模板**：

```typescript
// frontend/src/features/job-catalog/family-groups/JobFamilyGroupList.tsx
import React from 'react';
import { Box } from '@workday/canvas-kit-react/layout';
import { Heading } from '@workday/canvas-kit-react/text';
import { PrimaryButton } from '@workday/canvas-kit-react/button';
import { useJobFamilyGroups } from '@/shared/hooks/useJobCatalog';
import { CatalogTable } from '../shared/CatalogTable';
import { CatalogFilters } from '../shared/CatalogFilters';

export const JobFamilyGroupList: React.FC = () => {
  const { data, isLoading } = useJobFamilyGroups();

  return (
    <Box padding="l">
      <Box marginBottom="l">
        <Heading size="large">职类管理 (Job Family Groups)</Heading>
        <PrimaryButton marginTop="m">新增职类</PrimaryButton>
      </Box>

      <CatalogFilters onFilterChange={() => {}} />

      <CatalogTable
        data={data?.jobFamilyGroups ?? []}
        columns={[
          { key: 'code', label: '职类编码' },
          { key: 'name', label: '职类名称' },
          { key: 'status', label: '状态' },
          { key: 'effectiveDate', label: '生效日期' }
        ]}
        isLoading={isLoading}
        onRowClick={(item) => navigate(`/positions/catalog/family-groups/${item.code}`)}
      />
    </Box>
  );
};
```

### 2.4 权限模型

#### 2.4.1 权限定义

| 权限 Scope | 说明 | 适用场景 |
|-----------|------|----------|
| `position:read` | 查看职位列表和详情 | 职位列表、详情页 |
| `position:create` | 创建职位 | 新增职位按钮、表单 |
| `position:update` | 更新职位信息 | 编辑职位按钮、表单 |
| `position:delete` | 删除职位 | 删除按钮 |
| `job-catalog:read` | 查看 Job Catalog 数据 | 四层列表和详情 |
| `job-catalog:create` | 创建 Job Catalog 条目 | 新增按钮、表单 |
| `job-catalog:update` | 更新 Job Catalog 条目 | 编辑按钮、表单 |
| `job-catalog:delete` | 删除 Job Catalog 条目 | 删除按钮 |

#### 2.4.2 权限控制示例

```typescript
// 菜单级权限控制
const { hasPermission } = useAuth();

// 二级菜单项权限检查
{subItems.map((subItem) => {
  if (subItem.permission && !hasPermission(subItem.permission)) {
    return null; // 无权限则不显示
  }
  return <SubMenuItem {...subItem} />;
})}

// 操作按钮权限控制
<PrimaryButton
  onClick={handleCreate}
  disabled={!hasPermission('job-catalog:create')}
>
  新增职类
</PrimaryButton>
```

---

## 3. API 契约映射

### 3.1 GraphQL 查询（端口 8090）

参考 `docs/api/schema.graphql`（统一使用 `includeInactive` 参数，与契约保持一致）：

```graphql
# 职类查询
query GetJobFamilyGroups($includeInactive: Boolean = false, $asOfDate: Date) {
  jobFamilyGroups(includeInactive: $includeInactive, asOfDate: $asOfDate) {
    code
    name
    description
    status
    effectiveDate
    endDate
    isCurrent
  }
}

# 职种查询（需指定父级职类）
query GetJobFamilies($groupCode: JobFamilyGroupCode!, $includeInactive: Boolean = false, $asOfDate: Date) {
  jobFamilies(groupCode: $groupCode, includeInactive: $includeInactive, asOfDate: $asOfDate) {
    code
    name
    groupCode
    status
    effectiveDate
    endDate
    isCurrent
  }
}

# 职务查询（需指定父级职种）
query GetJobRoles($familyCode: JobFamilyCode!, $includeInactive: Boolean = false, $asOfDate: Date) {
  jobRoles(familyCode: $familyCode, includeInactive: $includeInactive, asOfDate: $asOfDate) {
    code
    name
    familyCode
    status
    effectiveDate
    endDate
  }
}

# 职级查询（需指定父级职务）
query GetJobLevels($roleCode: JobRoleCode!, $includeInactive: Boolean = false, $asOfDate: Date) {
  jobLevels(roleCode: $roleCode, includeInactive: $includeInactive, asOfDate: $asOfDate) {
    code
    name
    roleCode
    status
    effectiveDate
    endDate
  }
}
```

### 3.2 REST 命令（端口 9090）

参考 `docs/api/openapi.yaml`：

```yaml
# 职类管理
POST   /api/v1/job-family-groups              # 创建职类
PUT    /api/v1/job-family-groups/{code}       # 更新职类
POST   /api/v1/job-family-groups/{code}/versions  # 创建职类版本

# 职种管理（类似结构）
POST   /api/v1/job-families
PUT    /api/v1/job-families/{code}
POST   /api/v1/job-families/{code}/versions

# 职务管理
POST   /api/v1/job-roles
PUT    /api/v1/job-roles/{code}
POST   /api/v1/job-roles/{code}/versions

# 职级管理
POST   /api/v1/job-levels
PUT    /api/v1/job-levels/{code}
POST   /api/v1/job-levels/{code}/versions
```

### 3.3 前端 Hooks 设计

```typescript
// frontend/src/shared/hooks/useJobCatalog.ts

// 职类 Hooks
export const useJobFamilyGroups = (options?: QueryOptions) => {
  return useQuery({
    queryKey: ['jobFamilyGroups', options],
    queryFn: () => unifiedGraphQLClient.query(GET_JOB_FAMILY_GROUPS, options)
  });
};

export const useCreateJobFamilyGroup = () => {
  return useMutation({
    mutationFn: (data: CreateJobFamilyGroupInput) =>
      unifiedRESTClient.post('/job-family-groups', data),
    onSuccess: () => {
      queryClient.invalidateQueries(['jobFamilyGroups']);
    }
  });
};

// 职种 Hooks
export const useJobFamilies = (groupCode: string, options?: QueryOptions) => {
  return useQuery({
    queryKey: ['jobFamilies', groupCode, options],
    queryFn: () => unifiedGraphQLClient.query(GET_JOB_FAMILIES, { groupCode, ...options })
  });
};

// 职务和职级 Hooks（类似结构）
```

---

## 4. 实施计划

### 4.1 阶段划分

#### **Phase 0: 技术验证 POC（0.5天）**

- [ ] 复现并验证第0.4节技术清单
  - [ ] Expandable.Target 自定义包装兼容性
  - [ ] 替换 `AppShell` 左侧容器为 `SidePanel`，验证 312px 宽度、滚动与分隔线
  - [ ] 权限 Hook 与 tokens 可用性（含 `hasPermission` 实装与单测）
- [ ] 输出 POC 报告并在本文档变更历史登记

#### **Phase 1: 导航基础架构（1-2天）**

- [x] 创建92号文档
- [x] 重构 Sidebar 支持二级菜单
  - [x] 新增 `NavigationItem` 组件
  - [x] 实现折叠/展开动画
  - [x] 添加权限控制逻辑（接入 Phase 0 完成的 `useAuth.hasPermission`）
- [x] Shell 集成
  - [x] 将 `AppShell` 中的固定宽度 `Box` 替换为 `SidePanel`
  - [x] 调整主内容区 padding/分隔线以匹配视觉稿，保留 312px 宽度（或记录与设计沟通后的最终宽度）
- [x] 配置路由结构
  - [x] 扩展 `App.tsx` 路由配置
  - [x] 添加路由守卫（权限检查）
- [x] 单元测试
  - [x] `NavigationItem.test.tsx`
  - [x] Sidebar 交互测试

#### **Phase 2: Job Catalog 页面开发（3-5天）**

- [x] 创建通用组件
  - [x] `CatalogTable` - 复用组织管理表格逻辑
  - [x] `CatalogFilters` - 状态/日期筛选器
  - [x] `CatalogForm` - 通用CRUD表单
- [x] 职类管理页面
  - [x] `JobFamilyGroupList`
  - [x] `JobFamilyGroupDetail`
  - [x] `JobFamilyGroupForm`
- [x] 职种管理页面（复用模板）
- [x] 职务管理页面（复用模板）
- [x] 职级管理页面（复用模板）

#### **Phase 3: API 集成与联调（2-3天）**

- [x] 实现 GraphQL Hooks
  - [x] `useJobFamilyGroups`
  - [x] `useJobFamilies`
  - [x] `useJobRoles`
  - [x] `useJobLevels`
- [x] 实现 REST Mutation Hooks
  - [x] `useCreateJobFamilyGroup` 等
  - [x] `useUpdateJobFamilyGroup` 等
  - [x] 复用 `CatalogVersionForm` 实现“编辑当前版本”对话框
- [x] 单元测试补充
  - [x] `jobCatalogPages.test.tsx` 覆盖“编辑当前版本”校验与权限控制
- [ ] 前后端联调
  - [ ] 查询测试
  - [ ] 命令测试
  - [ ] 权限验证

#### **Phase 4: E2E 测试与优化（2-3天）**

- [ ] Playwright E2E 测试
  - [ ] 二级菜单展开/折叠
  - [ ] 职类 CRUD 流程
  - [ ] 权限控制验证
- [ ] 性能优化
  - [ ] 懒加载优化
  - [ ] 缓存策略
- [ ] 文档更新
  - [ ] 更新开发者快速参考
  - [ ] 更新实现清单

### 4.2 时间估算

| 阶段 | 工作量 | 交付物 |
|------|--------|--------|
| Phase 0 | 0.5天 | POC 验证报告、准入结论 |
| Phase 1 | 1-2天 | 二级导航组件、路由配置 |
| Phase 2 | 3-5天 | Job Catalog 四层页面 |
| Phase 3 | 2-3天 | API 集成、Hooks |
| Phase 4 | 2-3天 | E2E 测试、文档 |
| **总计** | **9-14天** | 完整功能上线（含 POC 与缓冲） |

---

## 5. 验收标准

### 5.1 功能验收

- [ ] **F1 - 导航交互**
  - [ ] 点击"职位管理"可展开/折叠二级菜单
  - [ ] 二级菜单包含5个子项且顺序正确
  - [ ] 当前激活页面高亮显示
  - [ ] 切换到其他一级菜单时自动折叠

- [ ] **F2 - 路由导航**
  - [ ] 所有11个路由均可正常访问
  - [ ] URL 与页面标题匹配
  - [ ] 浏览器前进/后退功能正常

- [ ] **F3 - 权限控制**
  - [ ] 无 `job-catalog:read` 权限时，四层子菜单不显示
  - [ ] 无 `job-catalog:create` 权限时，新增按钮禁用
  - [ ] 权限检查在前端和后端均生效

- [ ] **F4 - Job Catalog 页面**
  - [ ] 职类列表展示正确（表格、筛选器）
  - [ ] 职种列表展示正确（级联查询）
  - [ ] 职务列表展示正确
  - [ ] 职级列表展示正确
  - [ ] 所有页面支持时态查询（历史/未来）

- [ ] **F5 - CRUD 操作**
  - [ ] 创建职类成功并刷新列表
  - [ ] 更新职类成功并显示最新数据
  - [ ] 创建版本成功（时态管理）
  - [ ] 审计日志正确记录所有操作

### 5.2 技术验收

- [ ] **T1 - 代码质量**
  - [ ] TypeScript 零错误编译
  - [ ] ESLint 零告警（允许合理例外）
  - [ ] 所有组件有 PropTypes/Interface 定义
  - [ ] 复用率 ≥80%（通用组件）

- [ ] **T2 - 测试覆盖**
  - [ ] 单元测试覆盖率 ≥80%
  - [ ] E2E 测试通过率 100%
  - [ ] 关键路径测试完整（创建→编辑→查询）

- [ ] **T3 - 性能指标**
  - [ ] 页面初次加载 < 1秒
  - [ ] 路由切换 < 200ms
  - [ ] 表格渲染100行数据 < 500ms

- [ ] **T4 - 契约一致性**
  - [ ] 字段命名严格 camelCase
  - [ ] API 调用遵循 CQRS 分离
  - [ ] GraphQL Schema 与类型定义同步

- [ ] **T5 - 无障碍验证**
  - [ ] 使用键盘 (Tab/Enter/Space/Arrow) 可完成展开、选中与导航
  - [ ] 屏幕阅读器朗读 `aria-expanded` / `aria-controls` 信息准确
  - [ ] 焦点状态、hover/active 对比度符合 WCAG AA
  - [ ] 无额外 console 无障碍警告

### 5.3 文档验收

- [ ] **D1 - 技术文档**
  - [ ] 更新 `02-IMPLEMENTATION-INVENTORY.md`
  - [ ] 更新 `01-DEVELOPER-QUICK-REFERENCE.md`
  - [ ] 组件使用说明完整

- [ ] **D2 - 用户文档**
  - [ ] 导航使用说明（截图）
  - [ ] Job Catalog 管理指南
  - [ ] 权限配置说明

---

## 6. 风险与缓解措施

### 6.1 技术风险

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| Canvas Kit 不支持嵌套导航 | 高 | 中 | Phase 0 POC 先验证；若失败则回退为自定义 Accordion |
| 权限逻辑复杂导致渲染问题 | 中 | 低 | 统一封装 `usePermission` Hook |
| 四层级联查询性能问题 | 中 | 中 | 添加 GraphQL DataLoader，后端批量查询优化 |
| 路由嵌套层级过深 | 低 | 低 | 扁平化路由设计（已采用） |

### 6.2 业务风险

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| 用户不理解四层体系 | 中 | 高 | 提供引导教程、Tooltip 说明 |
| 历史数据迁移不完整 | 高 | 中 | 分阶段上线，先只读模式验证 |
| 权限配置错误导致数据泄露 | 高 | 低 | 后端强制校验，前端仅辅助 |

---

## 7. 后续扩展

### 7.1 短期优化（1-2个月）

- [ ] 增加批量操作功能（批量导入/导出）
- [ ] 支持拖拽排序（职类、职种优先级）
- [ ] 添加数据校验规则（编码格式、层级关系）
- [ ] 集成搜索高亮功能

### 7.2 长期规划（3-6个月）

- [ ] 支持多语言（国际化）
- [ ] 与外部系统同步（Workday、SAP）
- [ ] AI 辅助分类建议
- [ ] 组织架构与职位体系联动可视化

---

## 8. 参考资料

### 8.1 内部文档

- [80号文档：职位管理模块设计方案](./80-position-management-with-temporal-tracking.md)
- [CLAUDE.md：项目指导原则](../../CLAUDE.md)
- [01-DEVELOPER-QUICK-REFERENCE.md](../reference/01-DEVELOPER-QUICK-REFERENCE.md)
- [OpenAPI 规范](../api/openapi.yaml)
- [GraphQL Schema](../api/schema.graphql)

### 8.2 外部参考

- [Workday Canvas Kit - Navigation Patterns](https://workday.github.io/canvas-kit/)
- [React Router v6 - Nested Routes](https://reactrouter.com/en/main/start/tutorial#nested-routes)
- [Workday HCM - Job Architecture Setup Guide](https://community.workday.com/)

---

## 9. 变更历史

| 版本 | 日期 | 变更内容 | 作者 |
|------|------|----------|------|
| v1.0 | 2025-10-19 | 初始版本，自定义 NavigationItem 方案 | Claude Code |
| v2.0 | 2025-10-19 | ⭐ **重大更新**：改用 Canvas Kit 官方 Expandable 组件，废弃自定义方案 | Claude Code |
| v2.1 | 2025-10-19 | 新增 Phase 0 技术验证清单与 POC 先决条件，更新时间估算 | Claude Code |
| v2.2 | 2025-10-19 | Phase 0 POC 验证完成：SidePanel 集成、二级导航权限 Hook、Canvas tokens 自检与单测落地 | ChatGPT |
| v2.3 | 2025-10-19 | Phase 3 更新链路：补齐更新 Mutations、详情页编辑（复用 CatalogVersionForm）、单元测试覆盖 | ChatGPT |

**v2.0 关键变更**：
- ✅ 新增第0章：Canvas Kit 官方方案研究结论
- ✅ 替换 NavigationItem 为 Canvas Kit Expandable 组件
- ✅ 确认组件可用性（已安装 v13.2.15）
- ✅ 增加技术选型对比表（官方 vs 自定义）
- ✅ 强调 ARIA、无障碍、维护成本优势
- ✅ 更新状态为"技术方案确认"

**v2.1 关键补充**：
- ✅ 增设 Phase 0 POC 验证清单，作为进入 Phase 1 的准入条件
- ✅ 调整实施计划为 9-14 天，并将 Phase 4 缓冲到 2-3 天
- ✅ 明确 POC 结果需回填至变更历史便于追踪

---

**遵循原则**: CLAUDE.md 资源唯一性与跨层一致性原则、先契约后实现原则
**下次审查**: 2025-10-26
**审批状态**: ✅ **推荐使用 Canvas Kit 官方组件** - 待实施验证
