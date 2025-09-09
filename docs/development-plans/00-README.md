# 开发计划文档目录

**目录创建**: 2025-08-23  
**用途**: 组织架构重构项目开发计划集中管理  
**状态**: 活跃使用中  
**文档排序**: 按阅读顺序和重要程度排序

## 📁 文档结构 (按序号排序)

### **00 - 使用指南**
- `00-README.md` - 本文档，开发计划目录使用指南

### **01 - API规范基础**
- `../architecture/01-organization-units-api-specification.md` - 组织单元API规范 v4.2 (企业级标准)

### **02 - 技术架构**  
- `02-technical-architecture-design.md` - 技术架构设计文档

### **03 - API符合度重构**
- `03-api-compliance-intensive-refactoring-plan.md` - API规范符合度集中重构计划

### **04 - 后端实施计划**
- `04-backend-implementation-plan-phases1-3.md` - 后端实施计划(阶段1-3)

### **05 - 前端团队计划**
- `05-frontend-team-implementation-plan.md` - 前端团队实施计划

### **06 - 团队协作进展**
- `06-integrated-teams-progress-log.md` - 集成团队协作进展日志

### **07 - 契约测试自动化**
- `07-contract-testing-automation-system.md` - 契约测试自动化验证体系

### **08 - 前端API标准规范**
- `08-frontend-api-standards.md` - 前端API调用规范文档和开发标准

### **09 - 代码审查规范**
- `09-code-review-checklist.md` - 前端代码审查检查清单和质量标准

### **10 - 代码库清理维护**
- `10-codebase-cleanup-maintenance-plan.md` - 冗余文件和技术债务清理计划

### **11 - API权限映射验证**
- `11-api-permissions-mapping.md` - API权限映射完整性验证指南和权限体系文档

### **12 - 时态一致性实施计划**
- `12-temporal-consistency-implementation-plan.md` - 时态时间轴连贯性简化方案具体实施计划

### **13 - 权限系统实施计划**
- `13-permission-system-implementation-plan.md` - 权限系统实施计划

### **14 - API实现对齐计划**
- `14-api-implementation-alignment-plan.md` - API实现对齐计划文档

### **15 - 数据库触发器诊断**
- `15-database-triggers-diagnostic-report.md` - 数据库触发器混乱问题诊断分析和优化方案

### **20 - 现有资源分析开发指南**
- `20-existing-resource-analysis-guide.md` - 现有资源分析开发指南，强制执行优先使用现有API和组件原则

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
1. `20-existing-resource-analysis-guide.md` - **强制阅读** - 掌握现有资源分析原则和开发指导
2. `01-organization-units-api-specification.md` - 了解API规范和业务需求
3. `02-technical-architecture-design.md` - 理解技术架构和设计决策  
4. `11-api-permissions-mapping.md` - 掌握API权限体系和映射关系
5. `12-temporal-consistency-implementation-plan.md` - 理解时态数据管理方案
6. `03-api-compliance-intensive-refactoring-plan.md` - 掌握项目整体规划和进度
7. `04-backend-implementation-plan-phases1-3.md` - 了解后端实施策略
8. `05-frontend-team-implementation-plan.md` - 了解前端开发计划

### **当前开发重点 (2025-09-06)**  
**当前阶段**: 时态一致性架构优化期  
**重点任务**: 时态时间轴连贯性简化方案实施  
**实施方案**: 5阶段渐进式实施，预计2周完成  
**主要文档**: `12-temporal-consistency-implementation-plan.md`  
**技术重点**: PostgreSQL原生时态查询 + 应用层事务控制 + 性能优化  
**风险控制**: 分阶段实施 + 完整回滚方案 + 数据一致性验证

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