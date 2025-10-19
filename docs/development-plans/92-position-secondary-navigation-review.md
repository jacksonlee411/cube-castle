# 92号计划评审报告

**评审日期**: 2025-10-19
**评审人**: Claude Code
**计划版本**: v2.1
**评审结论**: ⚠️ **有条件通过** - 需补充关键前置条件后方可进入实施

---

## 1. 总体评价

### 1.1 优点

✅ **技术选型合理**
- 优先使用 Canvas Kit 官方组件而非自定义实现，符合"降低维护成本"原则
- 对比表详实（自定义 vs 官方方案），决策依据充分
- 已确认组件可用性（v13.2.15 已安装）

✅ **架构设计清晰**
- 导航层级结构合理，符合 Workday HCM 最佳实践
- 路由规划遵循 RESTful 规范，层级清晰
- CQRS 分离原则明确（GraphQL 查询 + REST 命令）

✅ **实施计划完整**
- 新增 Phase 0 POC 验证，体现悲观谨慎原则
- 阶段划分合理（0/1/2/3/4），时间估算留有缓冲
- 验收标准具体可量化（功能、技术、文档三维度）

✅ **文档质量高**
- 结构完整，涵盖背景、设计、API、风险、验收
- 代码示例详细，可直接参考实施
- 变更历史追踪完善，便于理解演进过程

---

## 2. 关键问题与建议

### 2.1 🔴 高优先级问题（阻塞性）

#### 问题 1: 权限 Hook 不存在

**现状**:
- 计划文档假设存在 `useAuth()` Hook 和 `hasPermission()` 方法
- 实际检查 `frontend/src/shared/hooks/` 目录，**未找到该 Hook 实现**
- NavigationItem 组件代码依赖 `import {useAuth} from '@/shared/hooks/useAuth'`

**影响**:
- Phase 0 POC 验证清单中的"权限 Hook 可用性"无法通过
- Phase 1 导航基础架构无法完成（权限控制逻辑缺失）
- 功能验收 F3 无法达成

**建议**:
```typescript
// 需在 Phase 0 之前补充实现
// frontend/src/shared/hooks/useAuth.ts

export interface AuthContext {
  hasPermission: (permission: string) => boolean;
  userPermissions: string[];
  isAuthenticated: boolean;
}

export const useAuth = (): AuthContext => {
  // 方案1: 从现有的 Zustand store 读取（推荐）
  // 方案2: 从 React Context 读取
  // 方案3: 临时实现（Phase 0 验证用）
  return {
    hasPermission: (permission: string) => {
      // TODO: 实现真实权限检查逻辑
      return true; // 临时返回 true 用于 POC
    },
    userPermissions: [],
    isAuthenticated: true,
  };
};
```

**准入条件**: 必须先实现 `useAuth` Hook（即使是简化版本），否则 Phase 0 无法启动。

---

#### 问题 2: SidePanel 与现有布局集成验证缺失

**现状**:
- 当前 `AppShell.tsx` 使用固定宽度 240px 的 Box 作为侧栏容器
- 计划中使用 `SidePanel` 组件，设定 `openWidth={312}`（增加 72px）
- 未明确说明如何过渡：是替换现有 Box 还是共存

**影响**:
- 可能导致布局错位或双侧栏问题
- Phase 0 POC 验证清单中"与现有 Shell 布局共存"项可能失败

**建议**:
1. 在 Phase 0 POC 中明确测试：
   ```typescript
   // 测试场景1: 完全替换 Box
   <SidePanel open openWidth={312} ...>
     <NavigationItem ... />
   </SidePanel>

   // 测试场景2: 验证宽度变化是否影响主内容区布局
   // 从 240px -> 312px 的影响评估
   ```

2. 更新 `AppShell.tsx` 的布局代码，在计划文档中给出具体的修改示例

**准入条件**: Phase 0 必须包含 AppShell 集成测试，确认无布局冲突。

---

### 2.2 🟡 中优先级问题（设计改进）

#### 问题 3: 代码示例与现有实现不一致

**现状**:
- 计划中 NavigationButton 使用 `styled('button')` 原生 button
- 当前 Sidebar.tsx 使用 Canvas Kit `<PrimaryButton>` 组件

**影响**:
- 可能导致视觉风格不一致（按钮样式、间距、交互态）
- 复用性降低

**建议**:
```typescript
// 保持 Canvas Kit 组件体系一致性
import {TertiaryButton} from '@workday/canvas-kit-react/button';

// 方案1: 继续使用 Canvas Button 组件
const NavigationButton = styled(TertiaryButton)<{active: boolean}>`
  background: ${({active}) => (active ? colors.soap200 : 'transparent')};
  // 其他样式覆盖
`;

// 方案2: 如果必须用原生 button，统一所有导航项风格
// 需在 Phase 0 评估两种方案的视觉效果
```

---

#### 问题 4: GraphQL Schema 参数命名不一致

**现状**:
- ✅ 所有查询已存在于 `docs/api/schema.graphql`（jobFamilyGroups/jobFamilies/jobRoles/jobLevels）
- ⚠️ 但参数命名存在差异：
  - 计划文档使用: `includeHistorical: Boolean`
  - 实际 Schema 使用: `includeInactive: Boolean`

**影响**:
- Phase 3 API 集成时 GraphQL 查询变量名错误，导致调试时间浪费
- 前端开发者可能混淆两种命名的语义差异

**建议**:
1. 统一使用 Schema 中的实际参数名 `includeInactive`
2. 更新计划文档第3.1节查询示例：
   ```graphql
   # 修正前
   query GetJobFamilyGroups($includeHistorical: Boolean) {
     jobFamilyGroups(includeHistorical: $includeHistorical) { ... }
   }

   # 修正后
   query GetJobFamilyGroups($includeInactive: Boolean) {
     jobFamilyGroups(includeInactive: $includeInactive) { ... }
   }
   ```

3. 如果语义上确实需要区分"历史版本"与"非活跃状态"，需要：
   - 在 Schema 中新增 `includeHistorical` 参数，或
   - 在业务逻辑层明确两者的映射关系

**准入条件**: Phase 1 开始前，更新计划文档中的所有查询示例，与 schema.graphql 保持一致。

---

#### 问题 5: 职位详情页路由设计冗余

**现状**:
```typescript
// 路由表中存在重复
<Route path=":code" element={<PositionTemporalPage />} />
<Route path=":code/temporal" element={<PositionTemporalPage />} />
```

**问题**:
- `/positions/:code` 和 `/positions/:code/temporal` 都指向同一组件
- 语义不清：是否存在非时态的详情页？

**建议**:
```typescript
// 方案1: 统一为时态路由（推荐）
<Route path=":code" element={<PositionTemporalPage />} />

// 方案2: 如需区分，明确两种视图的差异
<Route path=":code" element={<PositionDetailPage />} />      // 当前版本
<Route path=":code/history" element={<PositionTemporalPage />} /> // 历史版本
```

**准入条件**: Phase 1 路由配置前，明确详情页的视图策略。

---

### 2.3 🟢 低优先级建议（优化项）

#### 建议 1: 补充无障碍测试清单

**现状**:
- Phase 4 E2E 测试中未包含无障碍测试
- Canvas Kit Expandable 自动提供 ARIA 支持，但需验证集成后是否仍正常

**建议**:
在 Phase 4 验收标准中增加：
```markdown
- [ ] **A11y - 无障碍支持**
  - [ ] 键盘导航：Tab/Enter/Space/Arrow 键功能正常
  - [ ] 屏幕阅读器：aria-expanded/aria-controls 属性正确
  - [ ] 焦点管理：展开/折叠后焦点保持合理位置
  - [ ] 对比度：选中态/悬停态符合 WCAG AA 标准
```

---

#### 建议 2: 考虑国际化（i18n）前置

**现状**:
- 硬编码中文标签（"职位列表"、"职类管理"等）
- 长期规划（7.2）中提到多语言支持，但短期未考虑

**建议**:
```typescript
// 即使暂不实现多语言，也建议使用 i18n key 结构
const navigationConfig = [
  {
    label: t('nav.positions.list'), // 而非 '职位列表'
    path: '/positions',
    // ...
  }
];
```

**好处**: 后续国际化改造成本降低 80%

---

#### 建议 3: 性能优化考虑

**现状**:
- 计划中提到"懒加载优化"，但未给出具体方案

**建议**:
```typescript
// 路由懒加载（Phase 2 实施）
const JobFamilyGroupList = lazy(() =>
  import('@/features/job-catalog/family-groups/JobFamilyGroupList')
);

// 配合 Suspense
<Route
  path="catalog/family-groups"
  element={
    <Suspense fallback={<LoadingSpinner />}>
      <JobFamilyGroupList />
    </Suspense>
  }
/>
```

**验收标准更新**: T3 性能指标增加"懒加载模块首次打开 < 300ms"

---

## 3. 与项目原则的符合性评估

### 3.1 ✅ 符合项

| 原则 | 符合情况 | 证据 |
|------|---------|------|
| 资源唯一性 | ✅ | API 契约映射明确（schema.graphql + openapi.yaml） |
| Docker 容器化 | ✅ | 前端方案不涉及服务部署，无违反 |
| 诚实原则 | ✅ | 风险评估现实（Canvas Kit 不支持嵌套导航风险中级） |
| 悲观谨慎 | ✅ | 新增 Phase 0 POC，时间缓冲充足（9-14天） |
| 先契约后实现 | ✅ | 第3节明确 API 契约映射，Phase 3 前要求契约完整 |
| PostgreSQL 原生 CQRS | ✅ | GraphQL 查询 + REST 命令，分工明确 |

### 3.2 ⚠️ 需完善项

| 原则 | 问题 | 改进措施 |
|------|------|---------|
| 唯一事实来源 | NavigationItem 代码在计划文档中，未指明实际存放位置 | Phase 1 交付时需更新 `02-IMPLEMENTATION-INVENTORY.md` |
| 中文沟通 | 代码注释部分使用英文（如 `// Job Catalog 四层`） | 统一为中文注释或明确英文注释的使用场景 |

---

## 4. 风险评估补充

### 4.1 新增技术风险

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| React Router v7.7.1 兼容性 | 中 | 低 | Phase 0 验证嵌套路由在新版本下的行为 |
| Canvas Kit 主题冲突 | 中 | 低 | 确认 AppShell 未覆盖 SidePanel 默认样式 |
| 职位与组织架构耦合度 | 高 | 中 | 明确职位-组织关系的数据流向（查询时是否需要联表） |

### 4.2 业务风险补充

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| 四层体系与现有岗位数据不一致 | 高 | 高 | Phase 2 前完成数据建模验证，确认字段映射关系 |
| 用户混淆"职类/职种/职务/职级" | 中 | 高 | 每个页面增加 Tooltip 说明 + 关联 Workday 概念的文档链接 |

---

## 5. 准入条件清单（Phase 0 前）

- [ ] **P0 - 权限 Hook 实现**
  - [ ] 创建 `useAuth.ts` Hook（即使是简化版本）
  - [ ] 实现 `hasPermission` 方法（可临时 mock）
  - [ ] 编写单元测试验证基本功能

- [ ] **P1 - GraphQL Schema 参数一致性修正**
  - [ ] 更新计划文档第3.1节，所有查询示例使用 `includeInactive` 而非 `includeHistorical`
  - [ ] 确认前端 Hooks（useJobFamilyGroups 等）使用正确的参数名
  - [ ] 运行 `npm run validate:schema` 确认语法正确

- [ ] **P1 - 路由策略明确**
  - [ ] 决定 `/positions/:code` 与 `/positions/:code/temporal` 的关系
  - [ ] 更新路由表，消除冗余或歧义

- [ ] **P1 - AppShell 集成方案确认**
  - [ ] 决定是替换现有 Box 还是调整布局结构
  - [ ] 评估宽度从 240px 到 312px 的影响
  - [ ] 更新 AppShell.tsx 示例代码到计划文档

---

## 6. 评审结论

### 6.1 总体评分

| 维度 | 得分 | 满分 | 说明 |
|------|------|------|------|
| 技术方案合理性 | 8 | 10 | Canvas Kit 选型优秀，但缺少依赖项验证 |
| 实施可行性 | 7 | 10 | 时间估算合理，但前置条件缺失影响启动 |
| 文档完整性 | 9 | 10 | 结构清晰，代码示例丰富，变更历史完善 |
| 风险控制 | 7 | 10 | 技术风险覆盖全面，业务风险需补充数据验证 |
| 原则符合性 | 8 | 10 | 基本符合 CLAUDE.md 要求，细节需完善 |
| **总分** | **78** | **100** | **良好，需完善后实施** |

---

### 6.2 最终建议

**评审结论**: ⚠️ **有条件通过**

**建议行动**:
1. **立即执行**（1-2天）:
   - 实现 `useAuth` Hook（简化版本）
   - 补充 GraphQL Schema 契约定义
   - 明确路由策略和布局集成方案

2. **Phase 0 执行时**（0.5天）:
   - 严格按照第0.4节清单逐项验证
   - POC 失败立即中止，不进入 Phase 1
   - POC 结果记录于计划文档变更历史

3. **Phase 1-4 执行时**:
   - 每个阶段完成后更新 `02-IMPLEMENTATION-INVENTORY.md`
   - Phase 3 前确认所有后端 Resolver 已实现
   - Phase 4 增加无障碍测试

4. **文档维护**:
   - 本评审报告纳入计划文档参考资料
   - 所有发现的问题在实施中解决后，更新变更历史

---

## 7. 后续跟进

**下次审查**: 2025-10-26（与计划文档保持一致）
**审查重点**:
- Phase 0 POC 结果是否满足准入条件
- 准入条件清单是否全部完成
- 是否有新的阻塞性问题

**联系人**: 前端团队 Lead + 后端团队 Lead（需协同完成 GraphQL Schema）

---

**评审人签名**: Claude Code
**评审日期**: 2025-10-19
