# 开发计划文档目录

**目录创建**: 2025-08-23  
**用途**: 组织架构重构项目开发计划集中管理  
**状态**: 活跃使用中  
**文档排序**: 按阅读顺序和重要程度排序

## 📌 目录边界声明与交叉链接

- 本目录仅承载“计划/路线/进展/阶段报告”等时效性内容；完成项请移入 `../archive/development-plans/`。
- 规范性、长期稳定的参考资料（快速参考、实现清单、API 使用与质量手册）放在 `../reference/`，避免混淆权威参考与计划。
- 建议流程：
  - 开始新功能前 → 先查 [实现清单](../reference/02-IMPLEMENTATION-INVENTORY.md) 与 [API与质量工具指南](../reference/03-API-AND-TOOLS-GUIDE.md)
  - 确认需要新增能力 → 在本目录新增/更新对应计划，并按阶段完成后归档到 `../archive/development-plans/`。

## 📁 文档结构

### **活跃文档**

#### **00 - 使用指南**
- `00-README.md` - 本文档，开发计划目录使用指南

#### **01 - API规范基础**
- `../architecture/01-organization-units-api-specification.md` - 组织单元API规范 v4.2 (企业级标准)

#### **02 - 技术架构**
- `02-technical-architecture-design.md` - 技术架构设计文档

#### **06 - 团队协作进展**
- `06-integrated-teams-progress-log.md` - 集成团队协作进展日志

#### **210-213 - Phase1 执行计划**
- `210-database-baseline-reset-plan.md` - 数据库基线重置与验收计划
- `211-phase1-module-unification-plan.md` - Phase1 模块统一化实施方案
- `212-shared-architecture-alignment-plan.md` - Day6-7 架构复用决策与目录归属计划
- `213-go-toolchain-baseline-plan.md` - Go 1.24 工具链基线评审计划

#### **12 - 时态命令契约缺口（归档复测）**
- `../archive/development-plans/12-temporal-command-contract-gap-remediation.md` - `/organization-units/temporal` 契约缺失专项治理计划（P3 复测，核心整改已归档）

#### **200-206 - HRMS 系统模块化演进计划**
- `200-Go语言ERP系统最佳实践.md` - Go ERP 系统最佳实践参考
- `201-Go实践对齐分析.md` - 项目与最佳实践的对齐分析
- `203-hrms-module-division-plan.md` - **🌟 当前重点** HRMS 系统模块化演进与领域划分（v2.0，对齐度 95%+）
- `204-HRMS-Implementation-Roadmap.md` - HRMS 系统实施路线图与时间表
- `206-Alignment-With-200-201.md` - HRMS 计划与 200、201 文档的对齐分析
- `HRMS-DOCUMENTATION-INDEX.md` - HRMS 文档体系索引与快速导航

#### **205 - HRMS 系统过渡方案（已归档）** 📁
- `../archive/development-plans/205-HRMS-Transition-Plan.md` - HRMS 系统过渡方案详解（2025-11-04 归档）
- `../archive/development-plans/205-Plan-Alignment-Assessment.md` - Plan 205 对标评估报告（2025-11-04 归档）

### **归档文档** (docs/archive/development-plans/)

> **最新归档时间**: 2025-11-04
> **最新归档原因**: Plan 205（HRMS 系统过渡方案）已完成并按规范归档，执行细节由 Plan 210/211/212 推进；历史项目任务文档继续归档
> **访问方式**: `../archive/development-plans/` 子目录中可继续访问所有归档文档

#### **最近新增归档（2025-11-04）**
- `../archive/development-plans/205-HRMS-Transition-Plan.md` - HRMS 系统过渡方案详解（设计决议与参考）
- `../archive/development-plans/205-Plan-Alignment-Assessment.md` - Plan 205 对标评估报告（与 Plan 210/211/212 对齐分析）

#### **历史归档（2025-11-03）**
- `../archive/development-plans/06-design-review-task-assessment.md` - 06号设计评审任务确认报告（Job Catalog 设计评审完成）
- `../archive/development-plans/70-temporal-timeline-lifecycle-investigation.md` - 组织时间轴全生命周期连贯性调查报告
- `../archive/development-plans/105-navigation-ui-alignment-fix.md` - 导航栏 UI 对齐与布局优化（已完成）
- `../archive/development-plans/107-position-closeout-gap-report.md` - 职位管理收口差距核查报告（v2.0 归档确认版）
- `../archive/development-plans/109-position-audit-history-realignment.md` - 职位审计历史缺失整改计划
- `../archive/development-plans/110-position-status-normalization.md` - 职位版本状态与"当前版本"标识异常整改

#### **07-20 - 已完成计划和指南**
- `../archive/development-plans/07-contract-testing-automation-system.md` - ⭐ **契约测试自动化验证体系** (S级成功完成)
- `../archive/development-plans/08-parent-organization-selector-enhancement.md` - 上级组织选择器增强方案（Canvas Kit回归 + 测试同步完成）
- `../archive/development-plans/08-frontend-api-standards.md` - 前端API调用规范文档和开发标准
- `../archive/development-plans/09-code-review-checklist.md` - 前端代码审查检查清单和质量标准
- `../archive/development-plans/10-codebase-cleanup-maintenance-plan.md` - 代码库清理维护计划
- `../archive/development-plans/11-api-permissions-mapping.md` - API权限映射完整性验证指南和权限体系文档
- `../archive/development-plans/12-organization-1000001-temporal-analysis-report.md` - 组织1000001时态分析报告
- `../archive/development-plans/13-permission-system-implementation-plan.md` - 权限系统实施计划
- `../archive/development-plans/14-api-implementation-alignment-plan.md` - API实现对齐计划文档
- `../archive/development-plans/15-database-triggers-diagnostic-report.md` - 数据库触发器混乱问题诊断分析和优化方案
- `../archive/development-plans/16-trigger-optimization-action-plan.md` - 触发器优化行动计划
- `../archive/development-plans/16-code-smell-analysis-and-improvement-plan.md` - Go 工程实践为导向的代码异味分析与改进计划（P0-P3 完成）
- `../archive/development-plans/16-REVIEW-SUMMARY.md` - Plan16 评审摘要（Phase 0-3 + Plan24 验收记录）
- `../archive/development-plans/17-index-audit-and-optimization-plan.md` - 索引审计与优化计划
- `../archive/development-plans/17-temporary-governance-enhancement-plan.md` - TODO-TEMPORARY 治理与 419 状态码决策计划
- `../archive/development-plans/18-duplicate-code-elimination-plan.md` - ⭐ **重复代码消除计划** (S级全面完成)
- `../archive/development-plans/18-e2e-test-improvement-plan.md` - E2E 测试完善计划（Phase 1.3 完成）
- `../archive/development-plans/19-code-smells-and-remediation-plan.md` - 代码异味和修复计划
- `../archive/development-plans/19-phase0-workload-review.md` - Plan 19：Plan 16 Phase 0 工作量复核纪要（证据归档）
- `../archive/development-plans/20-existing-resource-analysis-guide.md` - 现有资源分析开发指南
- `../archive/development-plans/20-eslint-exception-strategy-and-zero-warning-plan.md` - ESLint 例外策略与零告警方案（完成）
- `../archive/development-plans/07-audit-history-load-failure-fix-plan.md` - 审计历史加载失败修复计划（P1，2025-10-07 验收归档）
- `../archive/development-plans/23-plan16-p0-stabilization.md` - Plan16 P0 稳定化方案（E2E ≥90% 验收完成）
- `../archive/development-plans/60-system-wide-quality-refactor-plan.md` - 系统级质量整合与重构计划（四阶段全部完成）
- `../archive/development-plans/61-system-quality-refactor-execution-plan.md` - 60号计划执行落地指南（Phase 4 验收后归档）
- `../archive/development-plans/63-front-end-query-plan.md` - Phase 3 前端 API/Hooks/配置整治计划（2025-10-12 验收归档）
- `../archive/development-plans/64-phase-3-acceptance-report.md` - Phase 3 验收报告（2025-10-12 通过）
- `../archive/development-plans/65-tooling-validation-consolidation-plan.md` - Phase 4 工具与验证巩固计划（2025-10-12 完成）
- `../archive/development-plans/66-phase-4-acceptance-draft.md` - Phase 4 验收草案
- `../archive/development-plans/84-position-lifecycle-stage2-implementation-plan.md` - 职位生命周期 Stage 2 实施计划（方案B）
- `../archive/development-plans/90-organization-stats-null-handling-report.md` - 组织列表 GraphQL 错误复盘与修复建议
- `../archive/development-plans/106-postgres-image-tags-investigation.md` - PostgreSQL 镜像标签一致性调查报告（完成）

## 📊 文档关系

```
阅读流程建议:
01-API规范 → 02-技术架构 → 20-现有资源分析指南 → 03-API重构计划 → 04-后端实施 → 05-前端团队 → 06-团队协作 → 07-契约测试 → 08-前端API标准 → 09-代码审查规范 → 10-代码清理 → 11-权限映射验证 → 12-时态一致性实施 → 13-权限系统实施 → 14-API实现对齐 → 15-数据库触发器诊断

项目层级关系:
03-api-compliance-intensive-refactoring-plan.md (当前重构方案)
├── 阶段1: 基础架构搭建 (90%完成) ✅ Main文件冲突已解决
├── 阶段2: 核心功能实现 ⏳ 进行中  
├── 阶段3: 高级功能和优化
└── 阶段4: 企业级特性和监控

技术支撑文档:
01-organization-units-api-specification.md (API规范基础)
└── 定义所有端点、数据模型、业务规则
02-technical-architecture-design.md (架构设计)  
└── 支撑所有开发阶段的技术架构决策

实施策略文档:
04-backend-implementation-plan-phases1-3.md (后端实施策略)
└── 后端服务阶段1-3的具体实施指导
05-frontend-team-implementation-plan.md (前端团队策略)
└── 前端团队的开发计划和实施指导
```

## 🎯 使用说明

### **新团队成员入门 (推荐阅读顺序)**
1. `../architecture/01-organization-units-api-specification.md` - 了解API规范和业务需求
2. `02-technical-architecture-design.md` - 理解技术架构和设计决策  
3. `06-integrated-teams-progress-log.md` - 了解项目当前状态和进展

### **归档文档参考** (仅作历史参考)
4. `../archive/development-plans/20-existing-resource-analysis-guide.md` - **历史参考** - 现有资源分析原则和开发指导
5. `../archive/development-plans/07-contract-testing-automation-system.md` - 契约测试自动化验证体系 (已完成)
6. `../archive/development-plans/18-duplicate-code-elimination-plan.md` - 重复代码消除计划 (已完成)

### **当前项目状态 (2025-11-03)**
**当前阶段**: 🚀 **HRMS 系统模块化演进** - 从职位管理单模块扩展至完整的企业级人力资源管理系统
**主要焦点**: 203号计划 - HRMS 系统模块化演进与领域划分（v2.0）
**对齐度**: 与 200/201 最佳实践对齐度 **95%+**
**主要计划**:
- Core HR 域（organization、workforce、contract）
- Talent Management 域（recruitment、performance、development）
- Compensation & Operations 域（compensation、payroll、attendance、compliance）
**版本演进**: v4.7.0（organization）→ v5.2.0（完整 Core HRMS）
**文档入口**:
  - 🌟 `203-hrms-module-division-plan.md` - 主计划（强烈推荐）
  - `HRMS-DOCUMENTATION-INDEX.md` - 文档体系索引
**历史文档**: 职位管理等历史阶段性工作已归档，后续工作统一在 203 号计划框架下推进

## 📋 API契约文档快速访问

### **契约优先开发 - Single Source of Truth**
- 🔧 **REST API规范**: [../api/openapi.yaml](../api/openapi.yaml) - 命令操作完整规范
- 🚀 **GraphQL Schema**: [../api/schema.graphql](../api/schema.graphql) - 查询操作完整Schema  
- 📚 **API文档入口**: [../api/README.md](../api/README.md) - API规范使用指南
- 📋 **版本变更历史**: [../api/CHANGELOG.md](../api/CHANGELOG.md) - API演进记录

> **契约驱动开发原则**: "先改契约，再写代码" - 所有API变更必须先更新`../api/`目录下的契约文件，后修改实现代码

## 📋 文档维护

- 所有开发计划文档统一存放在此目录
- 定期更新项目进度和状态
- 新增功能开发计划应在此目录创建对应文档
- 过时文档移入 `docs/archive/` 目录

---

*此目录替代原有的 `docs/note/organization-units-refactoring-2025-08/` 避免文档分散和遗漏*
