# 职位管理模块导航结构图

**版本**: v1.0
**创建日期**: 2025-10-21
**维护团队**: 前端团队 + UX团队
**关联计划**: 80号职位管理方案 · 107号收口差距核查报告

---

## 1. 系统级导航层次（3级结构）

```mermaid
graph TB
    A[Cube Castle 主导航] --> B[组织管理]
    A --> C[职位管理]
    A --> D[职位目录]

    B --> B1[组织列表]
    B --> B2[组织详情]
    B --> B3[组织层级树]

    C --> C1[职位仪表板]
    C --> C2[职位详情]
    C --> C3[编制统计]

    D --> D1[职类管理]
    D --> D2[职种管理]
    D --> D3[职务管理]
    D --> D4[职级管理]

    style C fill:#e1f5ff,stroke:#0078d4
    style C1 fill:#fff4e6,stroke:#ff8c00
    style C2 fill:#fff4e6,stroke:#ff8c00
    style C3 fill:#fff4e6,stroke:#ff8c00
```

**权限控制**:
- **职位管理** 整体需要 `position:read` 权限
- **组织管理** 需要 `org:read` 权限
- **职位目录** 需要 `job-catalog:read` 权限

---

## 2. 职位管理模块导航结构（L2: 页面级）

```mermaid
graph LR
    A[职位仪表板<br>/positions] --> B[职位详情<br>/positions/:code]
    A --> C[编制统计<br>/positions/headcount]

    B --> B1[详情Tab]
    B --> B2[版本历史Tab]
    B --> B3[任职记录Tab]
    B --> B4[转移记录Tab]

    B1 --> D1[填充职位]
    B1 --> D2[空缺职位]
    B1 --> D3[编辑职位]

    B2 --> E1[创建新版本]
    B2 --> E2[查看历史版本]

    B3 --> F1[查看任职历史]
    B3 --> F2[导出任职记录]

    B4 --> G1[发起转移]
    B4 --> G2[查看转移历史]

    style A fill:#0078d4,color:#fff
    style B fill:#0078d4,color:#fff
    style C fill:#0078d4,color:#fff
    style B1 fill:#106ebe,color:#fff
    style B2 fill:#106ebe,color:#fff
    style B3 fill:#106ebe,color:#fff
    style B4 fill:#106ebe,color:#fff
```

**路由权限映射**:
| 路由 | 组件 | 权限 | 降级行为 |
|------|------|------|---------|
| `/positions` | PositionDashboard | `position:read` | 显示空状态 |
| `/positions/:code` | PositionTemporalPage | `position:read` | 404错误 |
| `/positions/headcount` | PositionHeadcountDashboard | `position:read:stats` | 显示"无权限"提示 |

---

## 3. 职位仪表板页面结构（L3: 组件级）

```mermaid
graph TB
    Dashboard[职位仪表板<br>PositionDashboard.tsx]

    Dashboard --> Header[页面头部]
    Dashboard --> Stats[统计卡片区]
    Dashboard --> Actions[操作区]
    Dashboard --> List[列表区]

    Header --> H1[面包屑导航<br>首页 > 职位管理]
    Header --> H2[标题<br>职位管理]

    Stats --> S1[总职位数<br>PositionSummaryCards]
    Stats --> S2[已填充职位]
    Stats --> S3[空缺职位]
    Stats --> S4[计划中职位]

    Actions --> A1[创建职位按钮<br>position:create]
    Actions --> A2[批量导入<br>position:batch-create]
    Actions --> A3[导出列表<br>position:read]
    Actions --> A4[高级筛选]

    List --> L1[筛选条件栏<br>组织/状态/职级]
    List --> L2[职位列表表格<br>PositionList.tsx]
    List --> L3[分页组件]

    L2 --> L2A[列:编码]
    L2 --> L2B[列:标题]
    L2 --> L2C[列:组织]
    L2 --> L2D[列:职级]
    L2 --> L2E[列:编制]
    L2 --> L2F[列:状态]
    L2 --> L2G[列:操作]

    L2G --> Op1[查看详情]
    L2G --> Op2[填充<br>position:fill]
    L2G --> Op3[空缺<br>position:vacate]
    L2G --> Op4[编辑<br>position:update]

    style Dashboard fill:#0078d4,color:#fff,stroke:#005a9e,stroke-width:3px
    style Header fill:#e1f5ff
    style Stats fill:#fff4e6
    style Actions fill:#e6f7e6
    style List fill:#f3e6ff
```

**布局说明**:
- 页面宽度：固定容器（max-width: 1200px）
- 网格系统：12列栅格
- 统计卡片：4列布局（每卡3格）
- 列表区：全宽12列

---

## 4. 职位详情页面结构（L3: Tab级）

```mermaid
graph TB
    Detail[职位详情页<br>PositionTemporalPage.tsx]

    Detail --> Header[页面头部]
    Detail --> Tabs[Tab导航栏]
    Detail --> TabContent[Tab内容区]

    Header --> H1[面包屑<br>首页 > 职位管理 > P1000001]
    Header --> H2[职位标题 + 状态徽章]
    Header --> H3[快捷操作<br>填充/空缺/转移]

    Tabs --> Tab1[详情Tab]
    Tabs --> Tab2[版本历史Tab<br>position:read:history]
    Tabs --> Tab3[任职记录Tab]
    Tabs --> Tab4[转移记录Tab<br>position:read:history]

    TabContent --> TC1[详情内容<br>PositionDetails.tsx]
    TabContent --> TC2[版本列表<br>VersionList.tsx]
    TabContent --> TC3[任职历史<br>PositionAssignmentHistory.tsx]
    TabContent --> TC4[转移历史表格]

    TC1 --> D1[基本信息卡片<br>标题/组织/职位体系]
    TC1 --> D2[编制信息卡片<br>容量/已用/可用FTE]
    TC1 --> D3[时态信息卡片<br>生效日期/结束日期]
    TC1 --> D4[操作记录卡片<br>创建人/时间/原因]

    TC2 --> V1[版本工具栏<br>VersionToolbar.tsx]
    TC2 --> V2[版本时间轴<br>当前/历史/未来]
    TC2 --> V3[版本详情展开]

    TC3 --> A1[任职时间线]
    TC3 --> A2[任职事件列表<br>填充/空缺/转移]
    TC3 --> A3[导出CSV按钮]

    style Detail fill:#0078d4,color:#fff,stroke:#005a9e,stroke-width:3px
    style Tab1 fill:#106ebe,color:#fff
    style Tab2 fill:#106ebe,color:#fff
    style Tab3 fill:#106ebe,color:#fff
    style Tab4 fill:#106ebe,color:#fff
```

**Tab切换逻辑**:
- 默认Tab: 详情Tab
- Tab状态: URL hash控制 (`#details`, `#versions`, `#assignments`, `#transfers`)
- 权限缺失: Tab自动隐藏（如无 `position:read:history` 则隐藏版本Tab）

---

## 5. 职位表单对话框结构（L3: 表单级）

```mermaid
graph TB
    Form[职位表单对话框<br>PositionForm]

    Form --> Header[对话框头部<br>创建职位/编辑职位]
    Form --> Body[表单主体]
    Form --> Footer[对话框底部]

    Body --> Section1[基本信息区]
    Body --> Section2[职位体系区]
    Body --> Section3[组织归属区]
    Body --> Section4[编制信息区]
    Body --> Section5[时态信息区]

    Section1 --> F1[职位标题<br>必填 TextField]
    Section1 --> F2[职位类型<br>必填 Select<br>REGULAR/TEMPORARY/CONTRACT]
    Section1 --> F3[雇佣类型<br>必填 Select<br>FULL_TIME/PART_TIME/INTERN]

    Section2 --> F4[职类<br>必填 Select<br>来自 jobFamilyGroups]
    Section2 --> F5[职种<br>必填 级联Select<br>过滤职类]
    Section2 --> F6[职务<br>必填 级联Select<br>过滤职种]
    Section2 --> F7[职级<br>必填 级联Select<br>过滤职务]

    Section3 --> F8[归属组织<br>必填 OrganizationSelect]
    Section3 --> F9[汇报职位<br>可选 PositionSelect]

    Section4 --> F10[编制容量<br>必填 Number<br>默认1.0 FTE]
    Section4 --> F11[成本中心<br>可选 TextField]

    Section5 --> F12[生效日期<br>必填 DatePicker<br>默认今天]
    Section5 --> F13[结束日期<br>可选 DatePicker]

    Footer --> B1[取消按钮]
    Footer --> B2[保存按钮<br>触发validation]

    B2 --> V1{表单验证}
    V1 -->|通过| API1[POST /api/v1/positions<br>或<br>PUT /api/v1/positions/:code]
    V1 -->|失败| Error[显示验证错误]

    API1 -->|成功| Success[关闭对话框+刷新列表]
    API1 -->|失败| ErrorMsg[显示API错误]

    style Form fill:#0078d4,color:#fff,stroke:#005a9e,stroke-width:3px
    style Section1 fill:#e1f5ff
    style Section2 fill:#fff4e6
    style Section3 fill:#e6f7e6
    style Section4 fill:#f3e6ff
    style Section5 fill:#ffe6e6
```

**表单验证规则**:
| 字段 | 规则 | 错误提示 |
|------|------|---------|
| title | 非空, ≤255字符 | "职位标题为必填项，最多255字符" |
| organizationCode | 非空, 7位数字格式 | "归属组织为必填项" |
| job*Code | 非空, 符合编码规则 | "职位体系分类为必填项" |
| headcountCapacity | 非空, ≥0, ≤999.99 | "编制容量必须为0-999.99之间的数字" |
| effectiveDate | 非空, ≥今天（创建时） | "生效日期为必填项" |
| endDate | 可空, >effectiveDate | "结束日期必须晚于生效日期" |

---

## 6. 用户交互流程图（E2E场景）

### 6.1 创建职位流程

```mermaid
sequenceDiagram
    actor User
    participant Dashboard as 职位仪表板
    participant Dialog as 表单对话框
    participant REST as REST API
    participant GraphQL as GraphQL API
    participant List as 列表组件

    User->>Dashboard: 点击"创建职位"按钮
    Dashboard->>Dialog: 打开空表单对话框
    User->>Dialog: 填写职位信息
    User->>Dialog: 选择职位体系（级联）
    User->>Dialog: 选择归属组织
    User->>Dialog: 点击"保存"
    Dialog->>Dialog: 执行前端验证
    alt 验证失败
        Dialog-->>User: 显示验证错误
    else 验证通过
        Dialog->>REST: POST /api/v1/positions
        REST-->>Dialog: 201 Created + Position数据
        Dialog->>List: 触发查询刷新
        List->>GraphQL: 重新查询positions
        GraphQL-->>List: 返回新列表（含新职位）
        Dialog-->>User: 关闭对话框 + 成功提示
        List-->>User: 列表更新显示新职位
    end
```

### 6.2 填充职位流程

```mermaid
sequenceDiagram
    actor User
    participant Detail as 职位详情页
    participant FillDialog as 填充对话框
    participant REST as REST API
    participant GraphQL as GraphQL API

    User->>Detail: 进入职位详情页
    User->>Detail: 点击"填充职位"按钮
    Detail->>FillDialog: 打开填充对话框
    User->>FillDialog: 选择员工
    User->>FillDialog: 填写任职类型（PRIMARY/SECONDARY）
    User->>FillDialog: 设置生效日期
    User->>FillDialog: 点击"确认填充"
    FillDialog->>REST: POST /api/v1/positions/:code/fill
    REST-->>FillDialog: 200 OK + Assignment数据
    FillDialog->>Detail: 触发刷新
    Detail->>GraphQL: 重新查询position + assignments
    GraphQL-->>Detail: 返回更新后数据
    FillDialog-->>User: 关闭对话框 + 成功提示
    Detail-->>User: 详情页更新（状态→FILLED）
    Detail-->>User: 任职记录Tab显示新记录
```

### 6.3 创建版本流程

```mermaid
sequenceDiagram
    actor User
    participant VersionTab as 版本历史Tab
    participant VersionDialog as 版本对话框
    participant REST as REST API
    participant GraphQL as GraphQL API

    User->>VersionTab: 切换到"版本历史"Tab
    User->>VersionTab: 点击"创建新版本"
    VersionTab->>VersionDialog: 打开版本对话框（预填当前数据）
    User->>VersionDialog: 修改字段（如职级提升）
    User->>VersionDialog: 设置新生效日期（未来日期）
    User->>VersionDialog: 点击"保存"
    VersionDialog->>REST: POST /api/v1/positions/:code/versions
    REST-->>VersionDialog: 201 Created + 新版本数据
    VersionDialog->>VersionTab: 触发刷新
    VersionTab->>GraphQL: 重新查询positionVersions
    GraphQL-->>VersionTab: 返回版本列表（含新版本）
    VersionDialog-->>User: 关闭对话框 + 成功提示
    VersionTab-->>User: 时间轴显示新的未来版本
```

---

## 7. 面包屑导航路径

| 页面 | 面包屑路径 | 可点击节点 |
|------|-----------|-----------|
| 职位仪表板 | 首页 > 职位管理 | 首页 |
| 职位详情 | 首页 > 职位管理 > P1000001 | 首页, 职位管理 |
| 编制统计 | 首页 > 职位管理 > 编制统计 | 首页, 职位管理 |
| 职位详情-版本Tab | 首页 > 职位管理 > P1000001 > 版本历史 | 首页, 职位管理, P1000001 |
| 职位详情-任职Tab | 首页 > 职位管理 > P1000001 > 任职记录 | 首页, 职位管理, P1000001 |

**实现方式**: 使用 Canvas Kit 的 `Breadcrumbs` 组件 + React Router 的 `useLocation` Hook

---

## 8. 移动端响应式导航（未实施）

> **注意**: 当前版本（Stage 4）仅支持桌面端（≥1024px），移动端自适应计划在后续阶段实施。

**规划中的移动端导航**:
- 汉堡菜单（Hamburger Menu）替代顶部导航
- 底部Tab Bar替代详情页多Tab
- 抽屉式筛选面板
- 简化版列表卡片

---

## 9. 键盘导航与无障碍访问

| 快捷键 | 功能 | 页面 |
|--------|------|------|
| `Ctrl+K` | 全局搜索职位 | 全局 |
| `Ctrl+N` | 创建新职位 | 仪表板 |
| `Ctrl+E` | 编辑当前职位 | 详情页 |
| `Ctrl+S` | 保存表单 | 表单对话框 |
| `Esc` | 关闭对话框 | 所有对话框 |
| `Tab` | 焦点移动 | 所有页面 |
| `1/2/3/4` | 切换Tab | 详情页（焦点在Tab栏时） |

**无障碍支持**:
- ✅ ARIA 标签完整性
- ✅ 键盘焦点管理
- ✅ 屏幕阅读器支持
- ⏳ 高对比度主题（计划中）

---

## 10. 导航性能优化

### 10.1 代码分割（Code Splitting）

```typescript
// 懒加载职位模块
const PositionDashboard = lazy(() => import('./features/positions/PositionDashboard'));
const PositionTemporalPage = lazy(() => import('./features/positions/PositionTemporalPage'));

// 路由配置
<Route path="/positions" element={<Suspense fallback={<LoadingSpinner />}><PositionDashboard /></Suspense>} />
```

### 10.2 预加载策略

- **关键路径**: 职位仪表板在首屏加载时预加载
- **非关键路径**: 详情页在用户hover列表行时预加载
- **低优先级**: 编制统计页按需加载

### 10.3 缓存策略

| 数据类型 | 缓存时间 | 失效条件 |
|---------|---------|---------|
| 职位列表 | 5分钟 | 创建/更新/删除操作 |
| 职位详情 | 10分钟 | 编辑操作 |
| 职位体系选项 | 1小时 | 手动刷新 |
| 编制统计 | 15分钟 | 填充/空缺操作 |

---

## 11. 错误处理与降级方案

### 11.1 网络错误

```mermaid
graph TD
    A[API请求] -->|成功| B[正常渲染]
    A -->|失败| C{错误类型}
    C -->|401| D[跳转登录页]
    C -->|403| E[显示权限不足提示]
    C -->|404| F[显示资源不存在]
    C -->|500/网络错误| G[显示重试按钮]
    G -->|点击重试| A
```

### 11.2 权限降级

- **缺少 `position:read`**: 显示空状态 + "无权限访问"提示
- **缺少 `position:create`**: 隐藏"创建职位"按钮
- **缺少 `position:read:history`**: 隐藏"版本历史"Tab

---

## 12. 导航埋点与分析

### 12.1 关键埋点事件

| 事件名称 | 触发时机 | 参数 |
|---------|---------|------|
| `position_list_view` | 进入职位仪表板 | userId, tenantId, timestamp |
| `position_detail_view` | 进入职位详情 | positionCode, userId |
| `position_create_click` | 点击创建按钮 | userId |
| `position_fill_success` | 填充职位成功 | positionCode, assignmentId |
| `position_tab_switch` | 切换详情页Tab | positionCode, tabName |

### 12.2 性能监控指标

- **页面加载时间**: 目标 < 2s（P95）
- **Tab切换延迟**: 目标 < 300ms
- **表单提交响应**: 目标 < 1s

---

## 13. 与80号方案对应关系

| 80号方案章节 | 导航结构对应 | 完成状态 |
|-------------|-------------|---------|
| §7.0 Stage 0 | 页面布局设计 | ✅ 已验收 |
| §7.2 Stage 1 | 核心CRUD路由 | ✅ 已完成 |
| §7.3 Stage 2 | 职位生命周期流程图 | ✅ 已完成 |
| §7.4 Stage 3 | 编制统计页面 | ✅ 已完成 |
| §7.5 Stage 4 | 任职记录Tab+流程 | ✅ 已完成（86号计划） |

---

## 14. 改进建议（基于107号报告）

| 改进点 | 当前问题 | 建议方案 | 优先级 |
|--------|---------|---------|--------|
| E2E测试覆盖 | 仅有只读场景 | 补充完整CRUD生命周期脚本 | P0 |
| 性能基线 | 无P95数据 | 执行压力测试并记录 | P0 |
| 移动端支持 | 仅支持桌面端 | 实施响应式导航 | P2 |
| 离线模式 | 不支持离线 | Service Worker缓存 | P3 |

---

## 15. 版本变更记录

| 版本 | 日期 | 说明 |
|------|------|------|
| v1.0 | 2025-10-21 | 初版：根据107号报告要求补充导航结构图 |

---

**维护说明**:
- 此文档为80号计划§7.4的设计物料交付，满足107号报告§4.1要求
- 导航变更时请同步更新Mermaid图和路由映射表
- 新增页面时请更新§2的导航结构图
