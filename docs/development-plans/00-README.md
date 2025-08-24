# 开发计划文档目录

**目录创建**: 2025-08-23  
**用途**: 组织架构重构项目开发计划集中管理  
**状态**: 活跃使用中  
**文档排序**: 按阅读顺序和重要程度排序

## 📁 文档结构 (按序号排序)

### **00 - 使用指南**
- `00-README.md` - 本文档，开发计划目录使用指南

### **01 - API规范基础**
- `01-organization-units-api-specification.md` - 组织单元API规范 v4.2 (企业级标准)

### **02 - 技术架构**  
- `02-technical-architecture-design.md` - 技术架构设计文档

### **03 - API符合度重构**
- `03-api-compliance-intensive-refactoring-plan.md` - API规范符合度集中重构计划

### **04 - 早期实施策略**
- `04-early-stage-implementation-strategy.md` - 早期实施策略

### **05 - 核心功能开发**
- `05-core-features-development-plan.md` - 阶段2核心功能开发详细计划

## 📊 文档关系

```
阅读流程建议:
01-API规范 → 02-技术架构 → 03-API重构计划 → 04-早期策略 → 05-核心功能开发

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
04-early-stage-implementation-strategy.md (策略指导)
└── 项目早期阶段的具体实施指导
```

## 🎯 使用说明

### **新团队成员入门 (推荐阅读顺序)**
1. `01-organization-units-api-specification.md` - 了解API规范和业务需求
2. `02-technical-architecture-design.md` - 理解技术架构和设计决策  
3. `03-implementation-plan.md` - 掌握项目整体规划和进度
4. `04-early-stage-implementation-strategy.md` - 了解早期实施策略

### **当前开发重点 (2025-08-24)**  
**当前阶段**: API符合度提升 - 集中重构期  
**符合度现状**: 70%符合API规范，需要集中重构提升至95%  
**重构方案**: 采用"方式2: 集中重构"，1-2周完成质量提升  
**主要文档**: `03-api-compliance-intensive-refactoring-plan.md`  
**重构重点**: REST命令服务完善 + OAuth权限集成 + 企业级功能补全  
**风险控制**: Git分支保护 + 每日备份 + 功能开关机制

## 📋 文档维护

- 所有开发计划文档统一存放在此目录
- 定期更新项目进度和状态
- 新增功能开发计划应在此目录创建对应文档
- 过时文档移入 `docs/archive/` 目录

---

*此目录替代原有的 `docs/note/organization-units-refactoring-2025-08/` 避免文档分散和遗漏*