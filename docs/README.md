# 📚 我的文档导航中心

返回根 README: [../README.md](../README.md)

> **极简设计原则**: 一个目录，一个导航，够用就好！

## 🧭 重要导航（Reference vs Plans）
- Reference（长期稳定参考）：`docs/reference/` — 开发者快速参考、实现清单、API 使用与质量手册
  - [开发者快速参考](reference/01-DEVELOPER-QUICK-REFERENCE.md) · [实现清单](reference/02-IMPLEMENTATION-INVENTORY.md) · [API 使用指南](reference/03-API-USAGE-GUIDE.md)
- Development Plans（阶段性计划与进展）：`docs/development-plans/`
  - [使用指南](development-plans/00-README.md) · [技术架构设计](development-plans/02-technical-architecture-design.md) · [团队进展日志](development-plans/06-integrated-teams-progress-log.md)

边界约定：参考只收“长期稳定、可依赖”的材料；计划目录只收“计划/路线/进展/阶段报告（含 archived/）”。

## 🔥 常用文档 (一键直达)

### 🚀 立即开始
- [API快速参考](api/README.md) - 所有API接口总览
- [实现清单（Implementation Inventory）](reference/02-IMPLEMENTATION-INVENTORY.md) - 已实现的API/函数/接口索引（中英）
- **大规模清理完成**: 95%文档已归档 → archive/ ⭐ **极简状态**

### 🎯 核心功能  
- [**组织单元API规范**](architecture/01-organization-units-api-specification.md) - **主要API文档** ⭐ 完整的GraphQL/REST规范  
- [CLAUDE.md项目记忆](../CLAUDE.md) - 项目指导原则和架构设计 ⭐ 必读
- [设计开发标准](development-guides/design-development-standards.md) - 前端 UI/组件规范补充

### 🏗️ 架构文档
- [元合约v6.0规范](architecture/metacontract-v6.0-specification.md) - 核心设计合约
- [城堡蓝图](architecture/castle-blueprint.md) - 系统架构蓝图
- **已归档**: CQRS实施指南等其他架构文档 → archive/

## 📂 目录结构

```
docs/
├── reference/           # 长期稳定的权威参考（查询、指南、清单）
├── development-plans/   # 开发计划、路线与进展（活跃）
├── architecture/        # 架构设计与说明（非契约）
├── development-guides/  # 开发指南（前端 UI/组件规范等）
├── development-tools/   # 开发工具文档
├── api/                 # API 契约（OpenAPI/GraphQL）
└── archive/             # 历史文档归档区
    ├── development-plans/            # 开发计划归档（已完成/历史）
    ├── deprecated-neo4j-era/
    ├── deprecated-api-specs/
    ├── deprecated-api-design/
    ├── deprecated-guides/
    ├── deprecated-notes/
    ├── deprecated-setup/
    ├── project-reports/
    └── frontend-ux-optimization-deprecated/
```

### 📖 development-guides/ - 开发指南
- **design-development-standards.md** - 前端 UI/组件规范补充（以 CLAUDE.md 为准）
- 历史 guides 已归档：`archive/deprecated-guides/`

### 🔌 api/ - 核心API文档
精简到唯一权威API文档，其他已全部归档
- **organization-units-api-specification.md** - 唯一权威API规范
- **已全部归档**: 其他5份API文档 → archive/deprecated-api-design/

### 🏗️ architecture/ - 核心架构文档
保留2份核心架构设计文档
- **metacontract-v6.0-specification.md** - 核心设计合约
- **castle-blueprint.md** - 系统架构蓝图
- **已归档**: CQRS实施指南等其他架构文档 → archive/deprecated-neo4j-era/

## 🔍 快速搜索技巧

```bash
# 在所有文档中搜索关键词
grep -r "关键词" docs/

# 只搜索API文档
grep -r "关键词" docs/api/

# 搜索开发指南文档
grep -r "关键词" docs/development-guides/

# 查找特定文件类型
find docs/ -name "*.md" | grep "关键词"
```

## 📝 个人使用习惯

### 日常工作流
1. **查文档** → 先看这个README找到对应链接
2. **写笔记** → 直接放到`notes/`对应文件
3. **临时内容** → 放到`notes/temp/`，定期清理
4. **更新文档** → 就地编辑，无需复杂流程

### 极简维护规则
- ✅ **核心文档** → 已精简到仅7份文档(包含README)
- ✅ **过时内容** → 已全部移到`archive/` ⭐ 97%史诗级清理完成
- ✅ **新增文档** → 严格控制，避免文档膨胀
- ✅ **归档原则** → 保持极简状态，定期清理

## 🎯 文档哲学

> **够用就好，简单就是美**
> 
> - 不追求完美的分类
> - 不制定复杂的规则  
> - 专注于快速查找和使用
> - 保持足够的灵活性

---

*最后更新: 2025-08-23*  
*维护者: 单人团队，极简高效* 🚀  
*史诗级更新: 97%文档大规模归档清理 - 从约30份文档精简到7份核心文档* ⭐⭐⭐
