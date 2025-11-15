# 🏰 Cube Castle Frontend - 企业级React应用

> 说明：前端文档遵循仓库根目录 `AGENTS.md` 为唯一事实来源；如本文件与 `AGENTS.md` 或 `docs/reference/*` 存在不一致，以 `AGENTS.md` 为准。

## 🚀 统一配置架构 ⭐ **S级架构成果 (2025-09-07)**

基于React 19 + Canvas Kit v13 + TypeScript的现代化前端应用，采用统一配置管理和企业级架构标准。

### ✅ **企业级端口配置管理**
**权威配置源**: `src/shared/config/ports.ts`
```typescript
export const SERVICE_PORTS = {
  FRONTEND_DEV: 3000,           // 开发服务器
  FRONTEND_PREVIEW: 3001,       // 预览服务器  
  REST_COMMAND_SERVICE: 9090,   // CQRS命令服务
  GRAPHQL_QUERY_SERVICE: 8090,  // CQRS查询服务
  POSTGRESQL: 5432,
  REDIS: 6379
} as const;
```

### ✅ **重复代码消除完成**
- **Hook统一**: 7→2个Hook实现 (71%重复消除)
- **API客户端统一**: 6→1个客户端 (83%重复消除)  
- **类型系统重构**: 90+→8个核心接口 (80%+重复消除)
- **端口配置集中**: 15+文件→1个统一配置 (95%+硬编码消除)

### ✅ **统一 API 客户端 / Hook 使用示例**
```typescript
// 查询：统一使用共享 TanStack Query 客户端
import { useOrganizationsQuery } from '@/shared/hooks/useOrganizationsQuery';

export const OrganizationList = () => {
  const { data, isLoading, error } = useOrganizationsQuery({ status: 'ACTIVE' });

  if (isLoading) return <Spinner />;
  if (error) return <InlineError message={error.message} />;

  return (
    <ul>
      {data?.organizations?.data?.map(item => (
        <li key={item.code}>{item.name}</li>
      ))}
    </ul>
  );
};

// 命令：统一复用 useOrganizationMutations
import { useOrganizationMutations } from '@/shared/hooks/useOrganizationMutations';

export const CreateButton = () => {
  const { createOrganization, isCreating } = useOrganizationMutations();

  const handleCreate = () => {
    createOrganization.mutate({
      name: '新部门',
      unitType: 'DEPARTMENT',
      effectiveDate: new Date().toISOString().slice(0, 10),
    });
  };

  return (
    <PrimaryButton onClick={handleCreate} loading={isCreating}>
      新建组织
    </PrimaryButton>
  );
};
```

### 🔧 技术栈
- **构建工具**: Vite 7.0+ (统一配置支持) — 版本: Vite 7.0.4
- **UI框架**: React — 版本: React 19.1.0；TypeScript — 版本: TypeScript 5.8.3；Canvas Kit v13
- **状态管理**: TanStack Query + Zustand
- **测试**: Playwright + Vitest
- **质量保证**: P3企业级防控系统 ⭐ **新集成**

## 🛡️ 开发防控流程 ⭐ **P3系统集成**

### 🚀 开发前检查（本地自检，CI 同步执行）
```bash
# 1. 重复代码检测
bash ../scripts/quality/duplicate-detection.sh -s frontend

# 2. 架构一致性验证
node ../scripts/quality/architecture-validator.js --scope frontend

# 3. 文档同步检查
node ../scripts/quality/document-sync.js
```

#### ⚠️ Mock 模式只读提醒
- 默认配置在 `frontend/.env` / `.env.local` 中设置 `VITE_POSITIONS_MOCK_MODE=false`，确保职位模块直接连接真实 GraphQL/REST 服务。
- 当临时开启 Mock 模式（`VITE_POSITIONS_MOCK_MODE=true`）时，`PositionDashboard` 与职位详情视图（Temporal Entity 页面）会显示醒目的只读提示并禁用创建/编辑/版本操作；QA 验收必须在真实模式下执行完整 CRUD 流程。
- Playwright/CI 运行前请确认 `PW_REQUIRE_LIVE_BACKEND=1` 与 Mock 变量关闭，防止演示数据掩盖真实故障。

### ✅ 提交验证（CI 强制，本地建议执行）
CI 将在 PR 上强制执行以下检查；本地建议在提交前手动执行对应脚本保持一致：
- **Pre-commit Hook**: 架构一致性验证
- **CQRS守护**: 禁止前端REST查询，强制GraphQL
- **端口配置**: 检测硬编码端口，强制统一配置
- **API契约**: camelCase字段命名，废弃字段检查

### 📊 质量指标
- 重复代码率、架构违规、契约测试通过数等请以 `../reports/` 下最新报告为准（CI 会产出），避免在文档中固化具体数值导致事实漂移。

### 🔧 质量修复命令
```bash
# 自动修复重复代码
bash ../scripts/quality/duplicate-detection.sh --fix

# 自动修复文档同步
node ../scripts/quality/document-sync.js --auto-sync

# 查看详细违规报告
cat ../reports/architecture/architecture-validation.json
```

## ESLint 配置
- 项目已在 `frontend/eslint.config.js` 定义前端规则（含“禁止硬编码端口/REST 查询”等架构守护）；请以该文件为准进行调整。
- 如需新增/收紧规则，请在 PR 中说明依据并链接至 `AGENTS.md` 或 `docs/reference/*`，以保持唯一事实来源与一致性。
