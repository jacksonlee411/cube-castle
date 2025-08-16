# 📚 我的文档导航中心

> **极简设计原则**: 一个目录，一个导航，够用就好！

## 🔥 常用文档 (一键直达)

### 🚀 立即开始
- [生产部署指南](setup/deployment-guide.md) - 完整的生产环境部署流程
- [API快速参考](api/README.md) - 所有API接口总览
- [故障排除手册](guides/troubleshooting.md) - 常见问题快速解决

### 🎯 核心功能
- [时态管理快速开始](guides/temporal-management-quickstart.md) - 时态功能使用指南
- [时态管理用户手册](guides/temporal-management-user-guide.md) - 详细使用说明
- [GraphQL API文档](api/graphql-api.md) - GraphQL接口完整参考
- [REST API文档](api/temporal-management-api.md) - REST接口参考

### 🏗️ 架构文档
- [元合约v6.0规范](architecture/metacontract-v6.0-specification.md) - 核心设计合约
- [城堡蓝图](architecture/castle-blueprint.md) - 系统架构蓝图
- [CQRS统一实施指南](architecture/cqrs-unified-implementation-guide-v3.md) - CQRS架构指南

## 📂 目录结构

```
docs/
├── 🚀 setup/          # 环境配置与部署
├── 📖 guides/         # 使用指南与最佳实践  
├── 🔌 api/            # API文档与示例
├── 🏗️ architecture/   # 架构设计文档
├── 📝 notes/          # 个人笔记与临时文档
└── 📁 archive/        # 不常用的历史文档
```

### 🚀 setup/ - 环境配置
快速配置开发和生产环境的所有指南
- 开发环境配置
- 生产部署指南  
- Docker配置
- 依赖管理

### 📖 guides/ - 使用指南
日常使用的操作指南和最佳实践
- 用户使用指南
- 故障排除手册
- 最佳实践总结
- 维护操作指南

### 🔌 api/ - API文档
完整的API参考文档和使用示例
- REST API参考
- GraphQL API参考
- API使用示例
- 集成指南

### 🏗️ architecture/ - 架构设计
系统核心架构文档和设计规范
- 元合约v6.0规范
- 城堡蓝图架构
- CQRS实施指南
- 设计决策记录

### 📝 notes/ - 个人工作区
开发过程中的笔记、想法和临时文档
- `todo.md` - 待办事项
- `ideas.md` - 想法记录  
- `debugging.md` - 调试记录
- `temp/` - 临时文件夹

## 🔍 快速搜索技巧

```bash
# 在所有文档中搜索关键词
grep -r "关键词" docs/

# 只搜索API文档
grep -r "关键词" docs/api/

# 搜索指南文档
grep -r "关键词" docs/guides/

# 查找特定文件类型
find docs/ -name "*.md" | grep "关键词"
```

## 📝 个人使用习惯

### 日常工作流
1. **查文档** → 先看这个README找到对应链接
2. **写笔记** → 直接放到`notes/`对应文件
3. **临时内容** → 放到`notes/temp/`，定期清理
4. **更新文档** → 就地编辑，无需复杂流程

### 简单维护规则
- ✅ **常用内容** → 放在对应的功能目录
- ✅ **临时内容** → 放在`notes/`
- ✅ **过时内容** → 移到`archive/`  
- ✅ **每月清理** → 整理`notes/temp/`

## 🎯 文档哲学

> **够用就好，简单就是美**
> 
> - 不追求完美的分类
> - 不制定复杂的规则  
> - 专注于快速查找和使用
> - 保持足够的灵活性

---

*最后更新: 2025-08-16*  
*维护者: 单人团队，极简高效* 🚀