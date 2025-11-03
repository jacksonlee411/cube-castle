# 215 - Phase2 执行日志与进度跟踪

**文档编号**: 215
**标题**: Phase2 - Core HR 模块化建设执行日志
**创建日期**: 2025-11-04
**分支**: `feature/205-phase2-core-hr-modules`
**版本**: v1.0

---

## 概述

本文档跟踪 Phase2 的实施进展，包括：
- `workforce` 模块建设（员工档案与生命周期）
- `contract` 模块建设（劳动合同管理）
- 完成 Core HR 域三个模块的建立（organization ✅、workforce 🔨、contract 🔨）

---

## 阶段时间表（Week 3-4）

| 周 | 周一 | 周二 | 周三 | 周四 | 周五 |
|-----|------|------|------|------|------|
| **W3** | 需求梳理 | DB 设计 | API 设计 | 迁移脚本 | 迁移脚本 |
| **W4** | 模块开发 | 模块开发 | 单元测试 | 集成测试 | 部署 & 回归 |

---

## 进度记录

### Week 3 Day 1 - 需求梳理

**计划行动**:
- [ ] 根据 79 号文档分析 workforce 业务需求
- [ ] 分析 contract 业务需求
- [ ] 输出需求文档

**负责人**: 产品 + 架构师
**状态**: ⏳ 待启动

---

### Week 3 Day 2 - 数据库设计

**计划行动**:
- [ ] 设计 employees 表结构
- [ ] 设计 labor_contracts 表结构
- [ ] 输出 ER 图

**负责人**: 后端 TL + DBA
**状态**: ⏳ 待启动

---

### Week 3 Day 3 - API 设计

**计划行动**:
- [ ] 补充 OpenAPI 中的 workforce 端点
- [ ] 补充 GraphQL schema 中的 employees 查询
- [ ] 补充 contract 端点与权限声明

**负责人**: 架构师
**状态**: ⏳ 待启动

---

### Week 3 Day 4-5 - 迁移脚本创建

**计划行动**:
- [ ] 创建 workforce 迁移脚本
- [ ] 创建 contract 迁移脚本
- [ ] 验证 up/down 循环

**负责人**: DevOps
**状态**: ⏳ 待启动

---

### Week 4 Day 1-2 - 模块开发

**计划行动**:
- [ ] 创建 `internal/workforce/` 模块结构
- [ ] 创建 `internal/contract/` 模块结构
- [ ] 实现 models.go、repository.go、service.go
- [ ] 实现 handler.go、resolver.go

**负责人**: 后端团队（2-3人）
**状态**: ⏳ 待启动

---

### Week 4 Day 3-4 - 测试

**计划行动**:
- [ ] 单元测试（workforce）- 覆盖率 ≥80%
- [ ] 单元测试（contract）- 覆盖率 ≥80%
- [ ] 集成测试
- [ ] 契约测试

**负责人**: QA + 后端
**状态**: ⏳ 待启动

---

### Week 4 Day 5 - 部署与验收

**计划行动**:
- [ ] E2E 流程验证（创建 → 合同签署 → 变更 → 离职）
- [ ] 性能基线验收
- [ ] 代码审查与合并
- [ ] 文档更新

**负责人**: QA + 后端 TL + 文档支持
**状态**: ⏳ 待启动

---

## 关键检查点

### API 契约检查点

- [ ] 所有新增端点都在 `docs/api/openapi.yaml` 中声明
- [ ] 所有新增查询都在 `docs/api/schema.graphql` 中声明
- [ ] 字段命名一律 camelCase
- [ ] 路径参数统一 `{code}`
- [ ] 权限 scopes 在 OpenAPI 中明确声明

### 代码质量检查点

- [ ] `go fmt ./...` 通过
- [ ] `go test ./...` 通过，覆盖率 ≥80%
- [ ] `npm run lint` 通过（如有前端变更）
- [ ] 无循环依赖
- [ ] CQRS 边界清晰（command ↔ query 不直接调用）

### 数据库检查点

- [ ] 迁移脚本包含 `-- +goose Down` 回滚逻辑
- [ ] 审计字段一致性（created_at、updated_at、deleted_at、is_deleted）
- [ ] 外键约束正确性
- [ ] up/down 循环验证通过

---

## 风险与应对

| 风险 | 影响 | 预防措施 |
|------|------|--------|
| 需求不清导致返工 | 中 | Week3-D1 充分梳理，架构师 sign-off |
| DB 设计缺陷 | 高 | DBA 审查，迁移脚本验证 |
| API 契约不一致 | 中 | 契约优先，自动化测试 |
| 模块边界混淆 | 中 | 明确职责，架构审查 |
| 测试覆盖不足 | 中 | 追踪覆盖率，目标 ≥80% |

---

## 相关文档

- `06-integrated-teams-progress-log.md` - Phase2 启动指导
- `203-hrms-module-division-plan.md` - HRMS 模块划分蓝图
- `204-HRMS-Implementation-Roadmap.md` - 实施路线图
- `docs/api/openapi.yaml` - REST API 契约
- `docs/api/schema.graphql` - GraphQL 契约

---

## 提交记录

| 日期 | 提交 | 描述 |
|------|------|------|
| 2025-11-04 | 7370316b | docs: sync Plan 211 completion and launch Phase2 |
| 2025-11-04 | - | chore: initialize Phase2 branch and execution log |

---

**维护者**: Codex（AI 助手）
**最后更新**: 2025-11-04

