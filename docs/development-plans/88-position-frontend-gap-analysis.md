# 88号文档：职位管理前端功能差距分析

**版本**: v1.1
**创建日期**: 2025-10-17
**分析方法**: 静态代码分析（MCP Browser认证问题回退）
**对比基准**: 组织架构模块（frontend/src/features/organizations）
**分析对象**: 职位管理模块（frontend/src/features/positions）
**状态**: P0 完成（2025-10-17），P1 版本差异视图上线（2025-10-18）
**维护团队**: 前端团队 · 架构组
**遵循原则**: CLAUDE.md 资源唯一性 · CQRS 分工 · API-First 契约

---

## 1. 背景与目标

### 1.1 分析背景

- **需求来源**：在评审86号文档（职位任职 Stage 4）时，发现职位管理前端功能与组织架构模块存在明显差距。
- **对比基准**：组织架构模块作为本项目的成熟参考实现，已完整实现 CRUD、时态管理、详情导航等核心功能。
- **分析目的**：识别职位管理前端的功能缺口，为后续 UI 完善提供清晰的待办清单与优先级参考。

### 1.2 分析方法

由于 MCP Browser 启动后遇到认证问题（JWT Token 过期，GraphQL 请求失败），分析方法调整为**静态代码分析**：

1. **路由配置对比**：检查 `frontend/src/App.tsx` 中两个模块的路由定义
2. **Dashboard 对比**：对比主页面文件行数、功能特性、CRUD 处理器
3. **组件结构对比**：对比组件数量、目录层次、职责划分
4. **CRUD 操作对比**：通过 grep 检索 CRUD 相关代码（Create/Update/Delete/Transfer）
5. **交互模式对比**：分析导航方式、表单展示方式、详情页模式

### 1.3 关键发现概览

| 差距类别 | 差距数量 | 严重程度 |
|---------|---------|---------|
| 路由与导航 | 2个路由缺失 | 🟡 中等 |
| CRUD操作 | 3类操作缺失 | 🔴 高 |
| 组件架构 | 层次化缺失 | 🟢 低 |
| 交互模式 | 详情页导航缺失 | 🟡 中等 |
| 时态功能 | 时态版本管理缺失 | 🔴 高 |

### 1.4 评审结论采纳

- 已采纳《88号文档评审报告》（2025-10-17）中的 P0/P1 整改意见。
- 本版更新：
  - 明确命令服务 REST API 均已就绪，仅前端 UI 缺失。
  - 已实现 `positionVersions` GraphQL 查询，并在建议 3 中更新依赖说明。
  - 调整差距表述与优先级矩阵，避免误导。
  - 将版本号更新至 v1.1，状态改为“已修订（评审意见已采纳）”。

---

## 2. 路由配置对比

### 2.1 证据：App.tsx 路由定义

**Organizations 路由（3条）**：

```typescript
// frontend/src/App.tsx:62-64
<Route path="/organizations" element={<OrganizationDashboard />} />
<Route path="/organizations/:code" element={<OrganizationTemporalPage />} />
<Route path="/organizations/:code/temporal" element={<OrganizationTemporalPage />} />
```

**Positions 路由（1条）**：

```typescript
// frontend/src/App.tsx:65
<Route path="/positions" element={<PositionDashboard />} />
```

### 2.2 差距分析

| 路由类型 | Organizations | Positions | 差距 |
|---------|--------------|-----------|------|
| 列表页 | ✅ `/organizations` | ✅ `/positions` | 无差距 |
| 详情页 | ✅ `/organizations/:code` | ❌ 缺失 | **缺少独立详情页** |
| 时态管理页 | ✅ `/organizations/:code/temporal` | ❌ 缺失 | **缺少时态版本管理页** |

### 2.3 影响

- **用户体验降级**：无法通过 URL 直接访问某个职位的详情（如 `/positions/POS00001`），不利于分享链接、书签收藏。
- **功能不完整**：缺少时态版本管理页面，无法在 UI 层查看历史版本、创建未来版本。
- **模块不对称**：违反"职位管理完全复用组织架构模式"的设计承诺（见80号文档184-187行）。

---

## 3. 组件结构对比

### 3.1 证据：组件目录树

**Organizations 组件（6个，层次化结构）**：

```
frontend/src/features/organizations/components/
├── OrganizationForm/           (子目录)
│   ├── index.tsx               (346行，完整CRUD表单)
│   ├── FormFields.tsx
│   ├── FormTypes.ts
│   └── validation.ts
├── OrganizationTable/          (子目录)
│   └── ...
├── OrganizationTree.tsx
└── index.ts
```

**Positions 组件（7个，扁平结构）**：

```
frontend/src/features/positions/components/
├── PositionDetails.tsx         (256行，只读详情展示)
├── PositionHeadcountDashboard.tsx
├── PositionList.tsx
├── PositionSummaryCards.tsx
├── PositionTransferDialog.tsx  (200行，唯一的写操作组件)
├── PositionVacancyBoard.tsx
└── SimpleStack.tsx
```

### 3.2 差距分析

| 维度 | Organizations | Positions | 差距 |
|-----|--------------|-----------|------|
| 组件数量 | 6个 | 7个 | 无差距 |
| 目录层次 | ✅ 层次化（Form/、Table/子目录） | ❌ 扁平结构 | **缺少层次化组织** |
| Form组件 | ✅ `OrganizationForm/`（346行） | ❌ 无 | **缺少CRUD表单组件** |
| 详情组件 | ✅ 详情页 + Modal双模式 | ✅ `PositionDetails.tsx`（256行，只读） | Modal编辑模式缺失 |

### 3.3 影响

- **可维护性降低**：扁平结构在组件数量增加后难以维护（当前7个组件已接近扁平结构上限）。
- **职责不清晰**：缺少独立的 PositionForm 组件，导致创建/编辑逻辑无处安放（目前仅有 Transfer 操作）。
- **复用性差**：PositionTransferDialog 是特定操作的对话框，无法复用为通用的创建/编辑表单。

---

## 4. CRUD 操作对比

### 4.1 证据：操作处理器检索

**Organizations 操作处理器**：

```typescript
// frontend/src/features/organizations/OrganizationDashboard.tsx:200
const handleCreateOrganization = () => {
  navigate('/organizations/new');
};

// frontend/src/features/organizations/components/OrganizationForm/index.tsx:84-266
const handleSubmit = async (e: React.FormEvent) => {
  // ...
  if (isEditing) {
    if (normalizedFormData.isTemporal) {
      await createVersionMutation.mutateAsync({...}); // 时态版本
    } else {
      await updateMutation.mutateAsync(updateData);   // 常规更新
    }
  } else {
    await createMutation.mutateAsync(createData);     // 创建
  }
  // ...
};
```

**Positions 操作处理器**：

```bash
# grep 结果：无 handleCreate、handleEdit、handleDelete 处理器

# 唯一的写操作：
// frontend/src/features/positions/components/PositionTransferDialog.tsx:82-94
const handleSubmit = async (event: React.FormEvent) => {
  await transferAsync({
    code: position.code,
    targetOrganizationCode,
    effectiveDate,
    operationReason: operationReason.trim(),
    reassignReports,
  });
};
```

### 4.2 差距分析

| 操作类型 | Organizations | Positions | 差距 |
|---------|--------------|-----------|------|
| **Create（创建）** | ✅ `handleCreateOrganization` + `OrganizationForm` | ❌ 无 | **缺少创建 UI（REST 已就绪）** |
| **Read（读取）** | ✅ 详情页 + Dashboard | ✅ Dashboard内嵌详情 | 无差距 |
| **Update（编辑）** | ✅ `OrganizationForm`（isEditing模式） | ❌ 无 | **缺少编辑 UI（REST 已就绪）** |
| **Delete（删除）** | ❌ 无明确删除操作 | ❌ 无 | 双方均无（可能通过状态修改代替） |
| **时态版本** | ✅ `createVersionMutation` | ❌ 无 | **缺少时态版本创建 UI** |
| **Transfer（转移）** | N/A | ✅ `PositionTransferDialog` | Positions有，Organizations无 |

### 4.3 影响

- **功能严重不完整**：职位管理无法在前端创建、编辑职位，仅能通过后端 API 或数据库操作。
- **用户体验极差**：用户无法自主管理职位数据，严重影响系统可用性。
- **业务流程阻塞**：创建职位→填充→转移的完整流程无法闭环（缺少创建和编辑环节）。

---

## 5. 交互模式对比

### 5.1 证据：导航与交互方式

**Organizations 交互模式**：

```typescript
// frontend/src/features/organizations/OrganizationDashboard.tsx:200-202
const handleCreateOrganization = () => {
  navigate('/organizations/new');  // 导航到新页面
};

// frontend/src/features/organizations/OrganizationDashboard.tsx:204-206
const handleTemporalManage = (code: string) => {
  navigate(`/organizations/${code}/temporal`);  // 导航到时态管理页
};

// OrganizationForm 既支持 Modal 模式，也支持嵌入页面模式
```

**Positions 交互模式**：

```typescript
// frontend/src/features/positions/PositionDashboard.tsx:142-152
const [selectedCode, setSelectedCode] = useState<string>();

useEffect(() => {
  if (filteredPositions.length === 0) {
    setSelectedCode(undefined);
    return;
  }
  setSelectedCode(prev =>
    prev && filteredPositions.some(item => item.code === prev) ? prev : filteredPositions[0].code,
  );
}, [filteredPositions]);

// 详情在 Dashboard 内嵌展示：
// frontend/src/features/positions/PositionDashboard.tsx:250-258
<PositionDetails
  position={detailPosition}
  timeline={timeline}
  currentAssignment={currentAssignment ?? undefined}
  assignments={assignments}
  transfers={transfers}
  isLoading={!useMockData && detailQuery.isLoading}
  dataSource={useMockData ? 'mock' : 'api'}
/>
```

### 5.2 差距分析

| 交互方式 | Organizations | Positions | 差距 |
|---------|--------------|-----------|------|
| **列表页操作** | 点击创建按钮 → 导航到新页面 | ❌ 无创建按钮 | **无创建入口** |
| **详情展示** | 独立详情页（支持URL访问） | Dashboard内嵌（仅支持交互选择） | **缺少独立详情页** |
| **编辑模式** | Modal表单（支持创建/编辑/时态版本） | ❌ 无编辑入口 | **无编辑模式** |
| **时态管理** | 独立时态管理页（版本列表+创建） | ❌ 无 | **无时态版本管理UI** |
| **操作反馈** | Modal关闭 + 列表自动刷新 | Transfer对话框关闭 + 手动刷新 | 基本一致 |

### 5.3 影响

- **信息架构混乱**：组织架构采用"列表+独立详情页"，职位管理采用"列表+内嵌详情"，用户认知负担增加。
- **操作效率降低**：无法快速跳转到职位详情页（如从通知链接直达）。
- **移动端不友好**：Dashboard内嵌详情在小屏幕上体验差，独立详情页更适合响应式设计。

---

## 6. 时态功能对比

### 6.1 证据：时态版本管理

**Organizations 时态功能**：

```typescript
// frontend/src/features/organizations/OrganizationTemporalPage.tsx
// 完整的时态版本管理页面，支持：
// 1. 查看历史版本列表
// 2. 创建未来版本（计划组织）
// 4. 时间线可视化

// frontend/src/features/organizations/components/OrganizationForm/index.tsx:168-183
if (isEditing) {
  if (normalizedFormData.isTemporal) {
    await createVersionMutation.mutateAsync({
      code: organization!.code,
      name: nameValue,
      effectiveDate: TemporalConverter.dateToDateString(...),
      ...(normalizedFormData.effectiveTo ? { endDate: ... } : {}),
    });
  }
}
```

**Positions 时态功能**：

```typescript
// frontend/src/features/positions/PositionDashboard.tsx:169-173
const timeline: PositionTimelineEvent[] = useMockData
  ? selectedCode
    ? mockTimelineMap.get(selectedCode) ?? []
    : []
  : detailQuery.data?.timeline ?? [];

// ✅ 有时间线展示：
// frontend/src/features/positions/components/PositionDetails.tsx:238-244
<Heading size="small">职位时间线</Heading>
{timeline.length === 0 ? (
  <Text color={colors.licorice400}>暂无时间线记录</Text>
) : (
  timeline.map(item => <TimelineItem key={item.id} event={item} />)
)}

// ❌ 无时态版本管理UI（无法创建未来版本、无法查看历史版本详情）
```

### 6.2 差距分析

| 时态功能 | Organizations | Positions | 差距 |
|---------|--------------|-----------|------|
| **时间线展示** | ✅ 有 | ✅ 有（PositionDetails） | 无差距 |
| **历史版本列表** | ✅ OrganizationTemporalPage | ❌ 无 | **缺少历史版本列表** |
| **创建未来版本** | ✅ OrganizationForm（isTemporal模式） | ❌ 无 | **缺少计划版本创建** |
| **版本详情页签** | ✅ 时态管理页支持（版本历史 Tab） | ❌ 无 | **缺少版本详情页签** |
| **GraphQL查询支持** | ✅ `organizationVersion` | ❌ 待确认（需检查schema.graphql） | 需进一步验证 |

### 6.3 影响

- **无法规划未来**：用户无法在前端创建"计划中的职位"（PLANNED状态），破坏了时态管理的完整性。
- **历史追溯困难**：虽然有时间线展示，但无法查看某个历史版本的完整快照数据。
- **业务场景受限**：组织重组、岗位调整等需要提前规划的场景无法在前端完成。

---

## 7. 综合差距评估

### 7.1 差距总览表

| 差距类别 | 具体差距 | 严重程度 | 业务影响 | 技术难度 |
|---------|---------|---------|---------|---------|
| **路由导航** | 缺少独立详情页路由 | 🟡 中 | 用户无法分享链接、书签收藏 | 🟢 低（1天） |
| **路由导航** | 缺少时态管理页路由 | 🟡 中 | 无法访问时态版本管理UI | 🟢 低（1天） |
| **CRUD操作** | 前端缺少创建职位 UI | 🔴 高 | 用户无法自主创建职位 | 🟡 中（3-5天） |
| **CRUD操作** | 前端缺少编辑职位 UI | 🔴 高 | 用户无法修改职位基本信息 | 🟡 中（3-5天） |
| **CRUD操作** | 前端缺少时态版本创建 UI | 🔴 高 | 无法规划未来职位 | 🟡 中（3-5天） |
| **组件架构** | 扁平组件结构 | 🟢 低 | 可维护性降低（未来隐患） | 🟢 低（2天） |
| **交互模式** | Dashboard内嵌详情 | 🟡 中 | 移动端体验差、信息架构混乱 | 🟡 中（3天） |
| **时态功能** | 缺少历史版本列表 | 🔴 高 | 历史追溯困难 | 🟡 中（3天） |
| **时态功能** | 缺少版本详情页签 | 🟡 中 | 历史信息展示不足 | 🟡 中（3天） |

### 7.2 差距评分（5分制）

| 维度 | Organizations | Positions | 差值 |
|-----|--------------|-----------|------|
| 路由完整性 | ⭐⭐⭐⭐⭐ (5/5) | ⭐⭐ (2/5) | -3 |
| CRUD操作完整性 | ⭐⭐⭐⭐ (4/5，无删除） | ⭐⭐ (2/5，仅Transfer) | -2 |
| 组件架构合理性 | ⭐⭐⭐⭐⭐ (5/5) | ⭐⭐⭐ (3/5) | -2 |
| 交互模式一致性 | ⭐⭐⭐⭐⭐ (5/5) | ⭐⭐⭐ (3/5) | -2 |
| 时态功能完整性 | ⭐⭐⭐⭐⭐ (5/5) | ⭐⭐ (2/5，仅时间线） | -3 |
| **综合评分** | **⭐⭐⭐⭐⭐ (24/25)** | **⭐⭐ (12/25)** | **-12分** |

---

## 8. 补齐建议与优先级

### 8.1 高优先级（P0）- 核心功能缺失

**建议1：实现职位创建与编辑功能**

- **工作项**：
  1. 创建 `PositionForm/` 组件目录（参考 OrganizationForm 结构）
  2. 实现 `PositionForm/index.tsx`（支持创建/编辑/时态版本三种模式）
  3. 实现 `PositionForm/FormFields.tsx`（包含职类/职种/职务/职级等字段）
  4. 在 PositionDashboard 添加"创建职位"按钮 → 导航到 `/positions/new`
  5. 在 PositionDetails 添加"编辑"按钮 → 打开 PositionForm Modal
  6. 接入现有 REST 接口（`useCreatePosition`, `useUpdatePosition`）
- **技术难点**：
  - 职位创建依赖岗位目录（Job Catalog）选择，需要级联下拉框（职类组→职类→职种→职务→职级）
  - 编制容量（headcountCapacity）字段校验逻辑复杂
- **验收标准**：
  - 用户可以在前端创建职位并提交到 REST API（POST /api/v1/positions）
  - 用户可以编辑职位基本信息并提交（PATCH /api/v1/positions/{code}）
  - 表单支持时态版本创建（isTemporal 模式，指定 effectiveDate）
- **工作量预估**：5-8天（前端工程师1人）
- **依赖**：REST API 已就绪（详见第10节验证）
- **状态**：√ 已完成（2025-10-17，参考 docs/development-plans/06-integrated-teams-progress-log.md）

**建议2：补齐独立详情页路由**

- **工作项**：
  1. 创建 `PositionTemporalPage.tsx`（参考 OrganizationTemporalPage）
  2. 在 App.tsx 添加路由：`<Route path="/positions/:code" element={<PositionTemporalPage />} />`
  3. 在 PositionList 添加点击跳转逻辑：`navigate(\`/positions/\${position.code}\`)`
  4. 在 PositionTemporalPage 集成 PositionDetails、PositionForm、时间线展示
- **验收标准**：
  - 用户可以通过 URL `/positions/POS00001` 直接访问职位详情页
  - 详情页包含完整的职位信息、时间线、任职列表、操作按钮
- **工作量预估**：2-3天
- **依赖**：无（可独立完成）
- **状态**：√ 已完成（2025-10-17，路由已合并，详见 06 号日志）

### 8.2 中优先级（P1）- 功能增强

**建议3：实现时态版本管理页面**

- **工作项**：
  1. 扩展 PositionTemporalPage，添加"版本列表"Tab
  2. 集成 GraphQL 查询 `positionVersions(code: String!): [PositionVersion]`
  3. 实现版本列表展示（类似 OrganizationTemporalPage）
  4. 添加"创建未来版本"按钮 → 打开 PositionForm（isTemporal=true）
- **验收标准**：
  - 用户可以查看某个职位的所有历史版本（含 effectiveDate、endDate、isCurrent）
  - 用户可以创建未来版本（PLANNED 状态）
- **工作量预估**：3-5天
- **依赖**：后端需补充 `positionVersions` GraphQL 查询（待命令/查询服务排期）
- **完成说明**：`docs/api/schema.graphql` 新增 `positionVersions`，查询服务实现 `GetPositionVersions`，前端通过 `usePositionDetail` 拉取并渲染 `PositionVersionList`，并新增差异对比视图、CSV 导出与 `includeDeleted` 切换。
- **状态**：√ 已完成（2025-10-18，含 Vitest 覆盖 `PositionTemporalPage`）

**建议4：统一交互模式 - 采用"列表+独立详情页"架构**

- **工作项**：
  1. 移除 PositionDashboard 内嵌的 PositionDetails 组件
  2. 将 PositionDetails 集成到 PositionTemporalPage
  3. 修改 PositionList 的交互逻辑：点击职位 → `navigate(\`/positions/\${code}\`)`
  4. 确保响应式设计（移动端友好）
- **验收标准**：
  - 职位模块的交互模式与组织架构模块一致
  - 用户认知负担降低，操作流程更清晰
- **工作量预估**：2天
- **依赖**：建议2（独立详情页路由）完成后执行
- **状态**：√ 已完成（2025-10-17，PositionDashboard 已改为列表+跳转）

### 8.3 低优先级（P2）- 架构优化

**建议5：重构组件结构为层次化架构**

- **工作项**：
  1. 创建 `components/PositionForm/` 子目录（建议1已包含）
  2. 考虑创建 `components/PositionTable/` 子目录（如需支持表格视图）
  3. 整理 `components/` 根目录，保留通用组件（PositionSummaryCards、PositionVacancyBoard等）
- **验收标准**：
  - 组件目录结构清晰，职责明确
  - 代码可维护性提升
- **工作量预估**：1-2天（重构工作）
- **依赖**：无（可独立完成，但建议在建议1后执行）
- **状态**：□ 待实施

### 8.4 优先级决策矩阵

| 建议编号 | 建议名称 | 优先级 | 业务价值 | 技术难度 | 工作量 | 依赖项 | 建议开始时间 |
|---------|---------|-------|---------|---------|-------|-------|-------------|
| 建议1 | 创建与编辑功能 | 🔴 P0 | ⭐⭐⭐⭐⭐ | 🟡 中 | 5-8天 | REST API验证 | 立即 |
| 建议2 | 独立详情页路由 | 🔴 P0 | ⭐⭐⭐⭐ | 🟢 低 | 2-3天 | 无 | 立即 |
| 建议3 | 时态版本管理 | 🟡 P1 | ⭐⭐⭐⭐ | 🟡 中 | 3-5天 | GraphQL查询验证 | Week 2 |
| 建议4 | 统一交互模式 | 🟡 P1 | ⭐⭐⭐ | 🟢 低 | 2天 | 建议2完成 | Week 2 |
| 建议5 | 层次化架构 | 🟢 P2 | ⭐⭐ | 🟢 低 | 1-2天 | 建议1完成 | Week 3 |

---

## 9. 工作量与时间线预估

### 9.1 总工作量

- **前端开发工时**：13-20天（按1名前端工程师全职工作计算）
- **后端验证工时**：1-2天（确认 REST API 与 GraphQL 查询完整性）
- **测试工时**：3-5天（单元测试 + Playwright E2E 测试）
- **文档更新工时**：1天
- **总计**：18-28天（约3.5-5.5周）

### 9.2 建议时间线（3周迭代）

**Week 1（高优先级 P0）**：

- Day 1-3：实现独立详情页路由（建议2）
- Day 4-5：验证 REST API 与 GraphQL 查询完整性（后端团队协同）
- Day 6-10：实现职位创建与编辑功能（建议1）

**Week 2（中优先级 P1）**：

- Day 11-13：统一交互模式（建议4）
- Day 14-18：若后端交付 `positionVersions`，实现时态版本管理页面（建议3）；否则转为设计预研

**Week 3（低优先级 P2 + 测试）**：

- Day 19-20：重构组件结构为层次化架构（建议5）
- Day 21-23：编写单元测试与 Playwright E2E 测试
- Day 24-25：测试修复与文档更新

### 9.3 风险与缓解

| 风险 | 影响 | 概率 | 缓解措施 |
|-----|------|------|---------|
| REST API 不完整 | 高 | 中 | Week 1 Day 4-5 提前验证，发现问题立即提单给后端团队 |
| GraphQL 查询缺失 | 中 | 中 | 建议3可延后，先完成建议1-2-4 |
| 岗位目录级联选择器复杂 | 中 | 高 | 参考现有组织架构的父级选择器，或使用第三方组件库（如 react-select） |
| 时态版本管理逻辑复杂 | 中 | 中 | 复用 OrganizationTemporalPage 的实现模式，避免重复造轮子 |
| 测试覆盖不足 | 中 | 中 | 提前编写 E2E 测试场景清单，测试驱动开发（TDD） |

### 9.4 下一步行动（P1 进行中）

- [x] 补齐 GraphQL 契约：`docs/api/schema.graphql` + 查询服务 `GetPositionVersions` 已合并，PBAC 映射补全。
- [x] 扩展 `PositionTemporalPage`：版本列表组件 `PositionVersionList` 与 `PositionForm` 集成，GraphQL 数据链路打通。
- [x] 增补测试：新增 `frontend/src/features/positions/__tests__/PositionTemporalPage.test.tsx` 覆盖版本列表渲染与编码校验。
- [x] 阶段增强：差异视图、CSV 导出及 `includeDeleted` 切换已交付（PositionTemporalPage + PositionVersionDiff）。
- [ ] 下一步：补充差异摘要、字段分组及 CSV 字段映射（纳入 P1 后续迭代）。

---

## 10. 契约与API依赖验证清单

在开始前端开发前，必须验证以下后端 API 是否已完整实现：

### 10.1 REST API（命令服务，端口9090）

| 端点 | 方法 | 状态 | 验证方式 |
|-----|------|------|---------|
| `/api/v1/positions` | POST | ✅ 已实现 | openapi.yaml + position_service.go:CreatePosition |
| `/api/v1/positions/{code}` | PUT | ✅ 已实现 | openapi.yaml + position_service.go:ReplacePosition |
| `/api/v1/positions/{code}` | GET | ✅ 已实现 | openapi.yaml + position_repository:GetPositionByCode |
| `/api/v1/positions/{code}/versions` | POST | ✅ 已实现 | openapi.yaml + position_service.go:CreatePositionVersion |
| `/api/v1/positions/{code}/fill` | POST | ✅ 已实现 | Stage 2 已交付（见84号文档） |
| `/api/v1/positions/{code}/vacate` | POST | ✅ 已实现 | Stage 2 已交付 |
| `/api/v1/positions/{code}/transfer` | POST | ✅ 已实现 | Stage 3 已交付（见85号文档） |

### 10.2 GraphQL 查询（查询服务，端口8090）

| 查询 | 返回类型 | 状态 | 验证方式 |
|-----|---------|------|---------|
| `positions` | `[PositionRecord]` | ✅ 已实现 | 前端已使用（PositionDashboard） |
| `position` | `Position` | ✅ 已实现 | 前端已使用（usePositionDetail） |
| `positionTimeline` | `[PositionTimelineEntry]` | ✅ 已实现 | 前端已使用 |
| `positionAssignments` | `PositionAssignmentConnection` | ✅ 已实现 | 前端已使用 |
| `positionHeadcountStats` | `HeadcountStats` | ✅ 已实现 | 前端已使用 |
| `positionVersions` | `[PositionVersion]` | ⚠️ 待后端实现 | schema.graphql 未定义（需对齐 organizationVersions） |

### 10.3 验证命令

```bash
cd /home/shangmeilin/cube-castle

# 1. 检查 openapi.yaml 中的 positions 端点定义
grep -A20 "/positions" docs/api/openapi.yaml

# 2. 检查 schema.graphql 中的 positionVersions 查询
grep -A10 "positionVersions" docs/api/schema.graphql

# 3. 验证命令服务实现
grep -r "CreatePosition\|UpdatePosition" cmd/organization-command-service/internal

# 4. 验证查询服务实现
grep -r "positionVersions" cmd/organization-query-service/internal
```

---

## 11. 关联文档

- **`docs/development-plans/80-position-management-with-temporal-tracking.md`**
  职位管理总方案，Line 184-187 承诺"完全复用组织架构模式"

- **`docs/development-plans/86-position-assignment-stage4-plan.md`**
  职位任职 Stage 4 增量计划（v0.2），本次差距分析的触发来源

- **`docs/development-plans/06-integrated-teams-progress-log.md`**
  集成团队进展日志，包含86号文档评审结论

- **`frontend/src/App.tsx`**
  前端路由配置（Line 62-65），路由对比的证据来源

- **`frontend/src/features/organizations/OrganizationDashboard.tsx`**
  组织架构模块参考实现（326行）

- **`frontend/src/features/positions/PositionDashboard.tsx`**
  职位管理模块当前实现（294行）

- **`frontend/src/features/organizations/components/OrganizationForm/index.tsx`**
  组织架构表单组件参考实现（346行），PositionForm 的重要参考

- **`docs/api/openapi.yaml`**
  REST API 契约（命令服务）

- **`docs/api/schema.graphql`**
  GraphQL 查询契约（查询服务）

---

## 12. 决策与跟踪

### 12.1 待决策事项

- [ ] **决策1**：是否在本迭代（3周）完成所有P0+P1建议？
  - 选项A：完整实施（3周），达到组织架构模块同等水平
  - 选项B：仅实施P0（2周），P1延后到下一迭代
  - 选项C：拆分为两个2周迭代（P0 → P1 → P2）
  - **建议**：选项A（完整实施），确保模块一致性

- [ ] **决策2**：是否重构现有 PositionDashboard 的交互模式？
  - 选项A：重构（建议4），统一为"列表+独立详情页"
  - 选项B：保持现状（Dashboard内嵌详情），仅补齐CRUD
  - **建议**：选项A，避免长期的交互模式不一致

- [ ] **决策3**：剩余增强项排期（REST/GraphQL 已就绪）
  - 选项A：立即启动建议3（时态版本页面）并在本迭代完成
  - 选项B：建议3 放入下个迭代，与 P2 架构优化一并执行
  - **建议**：选项A——现已具备依赖，建议立即排期实施

### 12.2 跟踪清单

- [ ] 召开前后端团队对齐会议（讨论本文档内容）
- [ ] 执行"契约与API依赖验证清单"（第10节）
- [ ] 根据决策1-3选择实施方案，确定迭代范围
- [ ] 创建前端任务看板（Jira/GitHub Issues）
- [ ] 分配开发人员（前端工程师1-2人）
- [ ] 定期同步进展（每周三/五在06号日志更新）

---

## 13. 附录：代码证据索引

### 13.1 Organizations 关键代码位置

```yaml
路由定义:
  - frontend/src/App.tsx:62-64

Dashboard:
  - frontend/src/features/organizations/OrganizationDashboard.tsx:1-326
  - 创建处理器: :200-202
  - 时态管理处理器: :204-206

Form组件:
  - frontend/src/features/organizations/components/OrganizationForm/index.tsx:1-346
  - handleSubmit: :84-266
  - 时态版本创建: :168-183

时态管理页:
  - frontend/src/features/organizations/OrganizationTemporalPage.tsx:1-XXX
```

### 13.2 Positions 关键代码位置

```yaml
路由定义:
  - frontend/src/App.tsx:65

Dashboard:
  - frontend/src/features/positions/PositionDashboard.tsx:1-294
  - 选择逻辑: :142-152
  - 详情展示: :250-258

详情组件:
  - frontend/src/features/positions/components/PositionDetails.tsx:1-256
  - 时间线展示: :238-244

Transfer对话框:
  - frontend/src/features/positions/components/PositionTransferDialog.tsx:1-200
  - handleSubmit: :82-94
```

---

**文档完成**：2025-10-17
**下次更新**：决策完成后更新第12节，实施开始后记录进展到06号日志
