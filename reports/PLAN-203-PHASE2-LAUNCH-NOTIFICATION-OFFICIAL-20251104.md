# Plan 203 Phase 2 启动通知 - 正式版

**发送日期**: 2025-11-04
**收件人**: 后端团队、前端团队、QA、架构组、DevOps、PM
**主题**: 🚀 Plan 203 Phase 2 启动准备 - 2025-11-13 正式启动

---

## 📢 核心通知

亲爱的各位团队成员，

**Plan 214 Phase1 基线萃取已于 2025-11-03 完成并签字**，比原计划提前 6 天。数据库基线与 Goose/Atlas 工具链现已稳定。为确保 **Plan 203 Phase 2 能在 2025-11-13 顺利启动**，请按照以下事项准确执行：

---

## 📋 启动准备事项

### 1️⃣ 环境与依赖 (DevOps & DBA 负责)

**使用统一的容器化环境**:
- PostgreSQL 16-alpine (已运行 30+ 小时稳定)
- Redis 7-alpine (已就绪)
- 无须额外迁移操作

**复用最新基线文件**:
- `database/schema.sql` (50 KB, 60 对象)
- `database/schema/schema-inspect.hcl` (Atlas 导出)
- `database/migrations/20251106000000_base_schema.sql` (Goose 基线迁移)

**增量迁移方案**:
- 若需生成新的迁移文件，使用仓库根目录 `bin/atlas` (离线编译版本)
- 详见: `docs/development-tools/atlas-offline-guide.md`

**验收清单**:
- [ ] Docker Postgres 容器运行正常
- [ ] Goose round-trip 测试通过 (`goose up → down → up`)
- [ ] `go test ./... -count=1` 无失败
- [ ] CI workflow 中 Goose & go test 已启用

---

### 2️⃣ 资源冻结窗口 (PM 负责)

**确认可用资源**:
- ✅ 后端开发 (1-2 名): 2025-11-10 ~ 2025-12-10 全量投入
- ✅ 前端开发 (1-2 名): 2025-11-10 ~ 2025-12-10 全量投入
- ✅ QA (1-2 名): 2025-11-15 前完成测试用例草稿，2025-11-20 前进入测试
- ✅ DevOps: 2025-11-12 前完成环境最后检查
- ✅ 架构师: 始终可用于技术决策

**人员确认表**:
| 岗位 | 姓名 | 2025-11-13 可用 | 备注 |
|------|------|-----------------|------|
| 后端 TL | Codex | [ ] | 全栈负责 Day1-10 |
| 前端 TL | [姓名] | [ ] | UI/组件开发 |
| QA | [姓名] | [ ] | 集成测试 |
| DevOps | 林浩 | [ ] | CI/CD & 基础设施 |
| 架构师 | 周楠 | [ ] | 技术决策支持 |

---

### 3️⃣ 任务准备 (各功能团队负责)

#### 后端团队

**workforce 模块 API 契约** (截止 2025-11-07):
```yaml
Commands (REST):
  - POST /api/v1/workforce/employees     # 新增员工
  - PUT /api/v1/workforce/employees/{id} # 编辑员工
  - DELETE /api/v1/workforce/employees/{id} # 删除员工

Queries (GraphQL):
  - query GetEmployee(id: ID!)
  - query ListEmployees(filter: EmployeeFilter)
  - query GetOrganizationHierarchy(code: String)
```

**实现计划**:
- Day 1-2: 数据模型与关键路径设计
- Day 3-5: Command 服务 CRUD 实现
- Day 6-7: Query 服务与 GraphQL 端点
- Day 8-9: 事务一致性与时间戳管理
- Day 10: 集成测试与性能验证

#### 前端团队

**预先准备工作**:
- [ ] 复核共享类型定义 (见 `docs/api/schema.graphql`)
- [ ] 确认静态资源方案 (CDN/本地缓存)
- [ ] 准备 workforce 表单组件库
- [ ] Storybook 组件预演

**开发计划**:
- Day 1-2: 表单设计与原型
- Day 3-5: 编辑器与选择器组件
- Day 6-8: 集成与样式优化
- Day 9-10: 端到端测试与性能调优

#### QA 团队

**测试计划** (截止 2025-11-15):
- [ ] 单元测试覆盖 (command/query 服务)
- [ ] 集成测试用例 (API 端点 + 数据一致性)
- [ ] E2E 测试流程 (UI 工作流 + 用户场景)
- [ ] 性能基线测试

#### DevOps

**CI 流程验证**:
- [ ] Goose round-trip 测试已启用
- [ ] `go test ./...` 已集成到 CI
- [ ] `npm run lint && npm run test` 已启用
- [ ] Atlas 增量迁移 dry-run 已可用

---

### 4️⃣ 会议安排 (PM 协调)

| 时间 | 会议 | 参与人 | 议题 |
|------|------|--------|------|
| **2025-11-08 16:00** | 跨团队同步会 | 后端 TL, 前端 TL, QA, 架构师 | API 契约最终确认、需求拆分、测试策略 |
| **2025-11-12 09:00** | 启动前最后检查 | DevOps, DBA, 架构师 | 环境验证、数据准备、权限确认 |
| **2025-11-13 09:00** | Phase 2 启动会 | 全体参与者 | 工作分配、流程确认、日常站会开始 |

---

## 🎯 关键成功指标

**Phase 2 成功完成** (预期 2025-12-10):
- ✅ command 服务完成度: 100% (CRUD operations)
- ✅ query 服务完成度: 100% (GraphQL 端点)
- ✅ 前端组件完成度: 100% (workforce UI)
- ✅ E2E 测试覆盖: ≥ 80%
- ✅ 数据一致性: 无异常 (Round-trip 迁移 + 审计)
- ✅ CI/CD 绿灯: 100% (所有工作流通过)

---

## 📖 参考文档

**技术文档**:
- `docs/api/openapi.yaml` — REST API 契约
- `docs/api/schema.graphql` — GraphQL 查询
- `database/schema.sql` — 数据库基线
- `docs/development-tools/atlas-offline-guide.md` — 增量迁移指南

**执行文档**:
- `reports/PLAN-203-PHASE2-START-NOTIFICATION-20251103.md` (本文件)
- `reports/STATUS-UPDATE-PLAN-214-COMPLETE-20251104.md` (完整状态)
- `reports/EXECUTION-PATH-SUMMARY-20251104.md` (执行路径)

**计划文档**:
- `docs/development-plans/210-database-baseline-reset-plan.md` — Plan 210 (已完成)
- `docs/development-plans/214-phase1-baseline-extraction-plan.md` — Plan 214 (已完成)
- `docs/development-plans/06-integrated-teams-progress-log.md` — 项目进度日志

---

## ❓ 常见问题 (FAQ)

**Q: 如果遇到数据库问题，如何处理？**
A: 使用 `bin/atlas` 生成增量迁移。详见 `atlas-offline-guide.md`。若遇到问题，联系 DBA (李倩) 或 DevOps (林浩)。

**Q: Go 工具链版本要求？**
A: Go 1.24.0 及以上 (当前验证版本 1.24.9)。本地版本检查: `go version`。

**Q: 前端依赖共享代码如何处理？**
A: 见 Plan 212 决议，共享代码位置: `pkg/shared/` 与 `internal/shared/`。

**Q: 如何验证基线迁移是否可用？**
A: 执行 `make db-migrate-all && make db-rollback-last && make db-migrate-all`，应全部成功。

---

## 📞 沟通与支持

**日常同步**:
- 每日 16:00 站会 (5-10 分钟快速同步)
- 遇到阻塞 (> 2 小时) 立即升级给 PM

**技术支持**:
- DBA 相关: 李倩 (备份、Schema、数据一致性)
- DevOps 相关: 林浩 (环境、CI/CD、部署)
- 架构相关: 周楠 (设计决策、模块划分)
- 项目管理: PM (资源、时间表、风险)

**反馈渠道**:
- 即时反馈: #plan-203-phase2 频道
- 周报: 每周五 17:00 (回顾本周进度)

---

## ✅ 确认清单

请各团队 TL 在 **2025-11-10 前** 确认以下事项：

| 事项 | 负责人 | 确认状态 | 备注 |
|------|--------|---------|------|
| 资源冻结窗口确认 | PM | [ ] | 2025-11-10 ~ 2025-12-10 |
| API 契约初稿完成 | 后端 TL | [ ] | 截止 2025-11-07 |
| 前端准备工作完成 | 前端 TL | [ ] | 组件库、类型定义 |
| QA 测试计划完成 | QA TL | [ ] | 单元、集成、E2E |
| 环境验证完成 | DevOps | [ ] | Docker, Goose, CI |
| 架构最后评审 | 架构师 | [ ] | 2025-11-12 前完成 |

---

## 🎊 结语

**Plan 203 Phase 2 是继 Phase 1 模块统一化之后的核心功能开发阶段**。通过 Plan 214 的顺利完成，我们已经建立了稳定的数据库基线和完善的迁移工具链。

现在的挑战是**在保持代码质量与可维护性的同时，高效地实现 workforce 模块的 CRUD 功能和用户界面**。

让我们一起在 2025-11-13 开启新的征程！🚀

---

**邮件发送日期**: 2025-11-04
**接收截止日期**: 2025-11-10 (资源确认)
**启动日期**: 2025-11-13

**期待你们的确认与支持！**

Best regards,
PM Office
---

*CC: 架构师 (周楠)、项目经理、Steering Committee*

