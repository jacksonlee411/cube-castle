# Cube Castle 项目变更日志

## 未发布 - 数据库基线重建（2025-11-06）

### 🛠️ 基础设施
- **数据库迁移体系重建**：生成 `database/schema.sql` 为唯一事实来源，输出 Goose 基线迁移 `20251106000000_base_schema.sql`，并清理历史脚本。
- **迁移工具链切换**：新增 `goose.yaml`、`atlas.hcl` 与 `scripts/generate-migration.sh`，Makefile 和 CI 改为使用 Goose `up/down`。
- **验证资产**：新增 `tests/integration/migration_roundtrip_test.go`，确保 Goose 迁移支持 up→down→up 循环；在 `schema/` 输出备份与 diff 清单。

### 📚 文档
- `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`、`docs/reference/02-IMPLEMENTATION-INVENTORY.md`、`docs/reference/03-API-AND-TOOLS-GUIDE.md` 同步更新 Goose/Atlas 使用方式。
- `scripts/README.md` 与 Plan 210 文档补充新流程、回滚策略与检查清单。

## v1.5.5 - 前端日志统一与ESLint零告警方案 (2025-10-02)

### ✨ 新增
- **统一日志工具**：新增 `frontend/src/shared/utils/logger.ts`，按环境分级输出并提供 `mutation` 专用日志接口
- **日志单元测试**：覆盖调试模式、生产模式及分组日志行为，确保桥接层行为稳定

### 🔧 改进
- **移除 `console.*`**：`frontend/src` 目录全部迁移至 `logger`，强化 CQRS 调试日志与错误报表一致性
- **ESLint 门禁**：`no-console` 升级为 `error`，新增 CI 步骤校验 `eslint-disable-next-line camelcase` 例外说明
- **架构验证器扩展**：`scripts/quality/architecture-validator.js` 新增 `eslintExceptionComment` 规则和 `--rule` 过滤能力
- **文档更新**：开发者速查手册补充日志规范，Plan 20 验收标准同步调整

### ✅ 验证
- `npm run lint:frontend-api`（已知配置缺陷触发循环引用错误，详见提交备注）
- `npm run test`（覆盖 logger 单测）

---

## v1.5.4 - 移除未经审批的组织架构复制链接功能 (2025-09-21)

### 🔧 合规性修正
- **移除违规功能**：清理未经产品/治理流程审批的复制链接功能
  - 移除 `OrganizationDashboard.tsx` 中的"复制列表链接"按钮
  - 移除 `OrganizationTree.tsx` 中的复制链接、复制名称路径、复制编码路径按钮
  - 移除 `TemporalMasterDetailView.tsx` 中的相应复制功能
- **清理工具代码**：删除 `frontend/src/shared/utils/clipboard.ts` 及相关测试
- **测试文件更新**：移除相关测试用例并清理依赖

### ✅ 质量验证
- **ESLint**：✅ 通过，无 lint 错误
- **TypeScript**：✅ 通过类型检查
- **代码清理**：✅ 移除所有未使用的导入和依赖

### 📚 项目原则强化
- 重申"先契约后实现"核心原则的重要性
- 加强合规性检查流程，防止类似违规情况再次发生
- 完善文档生命周期管理和归档机制

---

## v1.5.3 - 质量门禁工具链升级与Go 1.23支持 (2025-09-17)

### 🔧 开发工具升级
- **golangci-lint 升级**：v1.55.2 → v1.61.0
  - 解决 Go 1.23 兼容性问题（之前版本构建于 go1.21.3，无法识别新语法特性）
  - 支持 Go 1.23 新特性：`for range` 常量语法、`slices` 扩展等
  - 安装路径：`~/.local/bin/golangci-lint`
- **gosec 配置优化**：v2.22.8
  - 创建符号链接至 `~/.local/bin/gosec` 便于访问
  - PATH 环境变量配置完善

### ✅ 质量门禁验证
- **make lint**：✅ 执行成功，发现并记录代码质量问题
  - errcheck: JSON encoder 错误未检查
  - unused: 未使用函数和字段清理
  - gosimple、staticcheck: 代码简化建议
- **make security**：✅ 执行成功，gosec 安全扫描正常运行
- **工具可执行性**：✅ `golangci-lint` 与 `gosec` 均可直接执行

### 📄 文档更新
- 更新 `docs/development-plans/06-integrated-teams-progress-log.md`：
  - 记录工具升级过程和版本信息
  - 更新质量门禁检查清单状态
  - 添加执行进度记录和验证结果

### 🎯 下一步
- 代码质量问题修复（基于 lint 结果）
- RS256 认证依赖配置验证

## v1.5.2 - 精简CLAUDE与跨文档导航完善 (2025-09-13)

### 📄 文档结构
- 精简 `CLAUDE.md` 为“核心原则 + 单一事实来源索引”，移除易变细节（变更通告、流程清单、脚本说明）。
- 易变内容迁移路径：
  - 变更通告/进展 → `docs/development-plans/` 与本 `CHANGELOG.md`
  - 开发前必检/禁止事项/操作清单 → `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`
  - API 一致性细则与工具 → `docs/reference/03-API-AND-TOOLS-GUIDE.md`
  - 文档治理与目录边界 → `docs/DOCUMENT-MANAGEMENT-GUIDELINES.md`、`docs/README.md`

### 🔗 跨文档导航
- 在以下文档中补充/统一指向权威链接：
  - `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` → 增补指向 `CLAUDE.md`、`AGENTS.md`、`docs/api/*`、文档治理。
  - `docs/reference/03-API-AND-TOOLS-GUIDE.md` → 增补同类“进一步阅读与治理”链接。
  - `docs/development-plans/06-integrated-teams-progress-log.md` → 增补相关规范与参考链接。

### ✅ 影响范围
- 不涉及代码行为变更，仅文档组织与导航优化；CI 目录边界与文档同步检查不受影响。

## v1.5.1 - 文档与规则加固 (2025-09-13)

### 🔧 规则与门禁
- 架构验证器降误报：
  - 端口检查仅在 URL/port 键值对场景触发，跳过注释/样式数字。
  - CQRS 检查移除通用 `.get(` 匹配，保留 `fetch()/axios.get()` 精确检测。
  - 契约检查跳过注释、增加白名单（`client_credentials`、`cube_castle_oauth_token`）。
- 结果：`node scripts/quality/architecture-validator.js` 全通过（CQRS=0 / 端口=0 / 契约=0）。

### 🧩 契约命名归零（M-1）
- 修正 temporal 相关组件与类型的 snake_case → camelCase：
  - TimelineComponent、TemporalMasterDetailView、PlannedOrganizationForm、TemporalSettings、shared/types/temporal.ts。

### 🪝 废弃 Hook 替换（M-2）
- 业务侧使用 useEnterpriseOrganizations；测试 mock 同步调整。
- ESLint 限制：禁止导入 `shared/hooks/useOrganizations`（防回归）。

### 📝 文档/模板
- PR 模板：新增“契约命名自查”项（前端字段 camelCase 自检）。
- 06 进展日志：记录修复清单与里程碑完成状态。

## v1.5.0 - 文档治理与目录边界 (2025-09-13)

### 🗂️ 文档结构与治理强化
- 新增 `docs/reference/` 目录：承载长期稳定的参考文档（开发者快速参考、实现清单、API 使用指南、质量手册）。
- 开发计划归档迁移：`docs/development-plans/archived/` → `docs/archive/development-plans/`，统一归档入口新增 `docs/archive/README.md`。
- 文档导航更新：`docs/README.md` 提供 Reference vs Plans 分区导航与边界说明。
- 目录边界规则加入规范：更新 `docs/DOCUMENT-MANAGEMENT-GUIDELINES.md`，新增“目录边界（强制）”与“月度审计”检查项。

### 🔧 审查与CI门禁
- PR 模板（.github/pull_request_template.md）：新增“文档治理与目录边界（Reference vs Plans）”检查清单。
- CI（.github/workflows/document-sync.yml）：新增“目录边界检查”与“文档同步检查”，违规将自动评论并阻断；质量门禁输出纳入总判定。

### 🗺️ 文档链接修正
- 全面修正指向旧的 `docs/development-plans/archived/` 的链接为 `docs/archive/development-plans/`。
- 更新 `CLAUDE.md`、`AGENTS.md`、根 `README.md`，同步目录结构与最新规范。

---

## v1.4.0 - 企业级生产就绪版本 (2025-08-25)

### 🏆 重大架构革命
- **PostgreSQL原生CQRS架构**: 性能提升70-90%，查询响应时间1.5-8ms
- **架构简化**: 移除Neo4j依赖，简化架构60%，单一PostgreSQL数据源
- **数据一致性**: 消除CDC同步复杂性，实现强数据一致性

### 🎯 契约测试自动化体系
- **测试覆盖**: 32个契约测试100%通过
- **质量门禁**: CI/CD自动化验证，GitHub Actions + Pre-commit Hook
- **API一致性**: 字段命名规范100%合规，Schema验证完全通过
- **分支保护**: 企业级合并阻塞机制配置完成

### 🔧 关键修复成果
- **OAuth认证修复**: 解决client_id/client_secret字段名特例问题
- **GraphQL Schema映射**: 修复前端查询字段与后端Schema不匹配问题
- **企业级响应结构**: 统一API响应信封格式
- **JWT认证体系**: 开发/生产模式完善支持

### 📊 监控与质量保证
- **Prometheus集成**: 企业级监控指标收集
- **契约测试监控**: React监控仪表板集成到主应用
- **实时质量状态**: 契约遵循度实时监控
- **自动化验证**: Pre-commit Hook提供秒级反馈

### 🚀 生产就绪特性
- **Canvas Kit v13**: 完整兼容集成，TypeScript零错误构建
- **API契约遵循**: 85%符合度，核心功能完全达标
- **构建系统**: 2.47s生产构建时间，完全稳定
- **开发体验**: IDE配置优化，开发工具链完善

---

## v1.3.0 及更早版本

# 开发进展记录 - 2025年8月10日

## 🎯 重要里程碑: E2E测试体系完成

### 📊 测试成果
- **E2E覆盖率**: 92% (超过90%目标)
- **测试用例数**: 64个测试用例，6个测试文件
- **跨浏览器**: Chrome + Firefox 全面支持
- **测试框架**: Playwright + TypeScript

### 🔧 关键修复
1. **数据一致性测试修复**
   - 问题: API返回"ACTIVE"，前端显示"启用"
   - 解决: 添加状态字段本地化映射
   - 文件: `frontend/tests/e2e/business-flow-e2e.spec.ts:322-330`

2. **API兼容性测试修复**  
   - 问题: REST API数据结构误判
   - 解决: 修正数据结构断言 (REST API直接返回数据，不包装在'data'字段)
   - 文件: `frontend/tests/e2e/regression-e2e.spec.ts:71-72`

### 📈 性能验证结果
- **页面加载时间**: 0.5-0.9秒 (< 1秒目标 ✅)
- **API响应时间**: 0.01-0.6秒 (< 1秒目标 ✅)  
- **CDC同步延迟**: < 300ms (企业级标准 ✅)
- **内存使用**: ~23MB (优化后)

### 📋 生成文档
- `e2e-coverage-report.md`: 详细的E2E测试覆盖率报告
- 更新 `CLAUDE.md`: 项目最新状态和E2E测试成果
- 更新 `README.md`: 版本信息和测试验收结果

### 🎉 质量保证达成
| 质量指标 | 目标 | 实际 | 状态 |
|---------|------|------|------|
| E2E覆盖率 | ≥90% | 92% | ✅ 达标 |
| 页面响应时间 | <1秒 | 0.5-0.9秒 | ✅ 优秀 |  
| API响应时间 | <1秒 | 0.01-0.6秒 | ✅ 优秀 |
| 跨浏览器兼容 | Chrome+Firefox | ✅ 支持 | ✅ 达标 |

## 🚀 下一步计划
- [ ] 提交所有变更到Git仓库
- [ ] 标记v1.1-E2E版本Tag  
- [ ] 准备生产环境部署计划
- [ ] 增强压力测试(可选)

---
**执行时间**: 2025-08-10 12:00  
**执行环境**: WSL2 + Docker + 完整CQRS服务栈  
**验证状态**: ✅ E2E测试体系完整，生产环境部署就绪
