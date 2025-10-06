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

#### **07 - 待修复问题**
- `07-audit-history-load-failure-fix-plan.md` - 审计历史页签“加载审计历史失败”分析与修复计划（P1）

#### **12 - 时态命令契约缺口（归档复测）**
- `../archive/development-plans/12-temporal-command-contract-gap-remediation.md` - `/organization-units/temporal` 契约缺失专项治理计划（P3 复测，核心整改已归档）

#### **16 - 代码异味治理**
- `16-code-smell-analysis-and-improvement-plan.md` - Go 工程实践为导向的代码异味分析与改进计划（P2）




### **归档文档** (docs/archive/development-plans/)

> **文档归档时间**: 2025-09-13  
> **归档原因**: 项目完成企业级生产就绪状态，部分计划文档已执行完成  
> **访问方式**: `../archive/development-plans/` 子目录中可继续访问所有归档文档

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
- `../archive/development-plans/17-index-audit-and-optimization-plan.md` - 索引审计与优化计划
- `../archive/development-plans/17-temporary-governance-enhancement-plan.md` - TODO-TEMPORARY 治理与 419 状态码决策计划
- `../archive/development-plans/18-duplicate-code-elimination-plan.md` - ⭐ **重复代码消除计划** (S级全面完成)
- `../archive/development-plans/18-e2e-test-improvement-plan.md` - E2E 测试完善计划（Phase 1.3 完成）
- `../archive/development-plans/19-code-smells-and-remediation-plan.md` - 代码异味和修复计划
- `../archive/development-plans/19-phase0-workload-review.md` - Plan 19：Plan 16 Phase 0 工作量复核纪要（证据归档）
- `../archive/development-plans/20-existing-resource-analysis-guide.md` - 现有资源分析开发指南
- `../archive/development-plans/20-eslint-exception-strategy-and-zero-warning-plan.md` - ESLint 例外策略与零告警方案（完成）

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

### **当前项目状态 (2025-09-13)**  
**当前阶段**: ✅ **企业级生产就绪** - 核心功能完成，质量门禁生效，架构成熟  
**主要成就**: PostgreSQL原生CQRS架构、契约测试自动化体系、重复代码消除系统  
**文档归档**: 已完成开发计划文档归档，保留核心活跃文档  
**主要参考**: `06-integrated-teams-progress-log.md` - 最新项目进展和状态  
**归档文档**: 14个已完成的计划和指南文档已移至 `archived/` 目录

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
