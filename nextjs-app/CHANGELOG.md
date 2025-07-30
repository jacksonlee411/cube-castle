# Changelog

All notable changes to the Cube Castle Next.js Frontend will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.5.0] - 2025-07-30

### ✨ Added
- **完整CRUD功能恢复**: 所有核心管理页面从临时模式恢复为完整功能
- **职位管理入口**: 主页新增职位管理模块导航卡片
- **员工管理完整功能**: 创建、读取、更新、删除员工记录，支持高级筛选和分页
- **组织架构管理**: 树形视图和表格视图切换，层级关系管理，统计仪表板
- **职位管理系统**: FTE预算管理，利用率统计，状态跟踪，完整CRUD操作
- **员工职位历史**: 时间线视图，工作流历史跟踪，职位变更记录管理

### 🔧 Fixed
- **GraphQL依赖错误**: 员工职位历史页面移除Apollo Client依赖，使用独立实现
- **ES模块兼容性**: 完全解决Next.js和Ant Design之间的ES模块兼容性问题
- **临时模式问题**: 所有核心页面从"临时模式"恢复为正常功能模式
- **导航缺失问题**: 主页添加职位管理入口，改善用户导航体验

### 🚀 Improved
- **UAT测试就绪**: 所有核心功能现在可进行完整的用户验收测试
- **代码质量**: 移除未使用的GraphQL依赖，提高代码维护性
- **用户体验**: 优化页面加载性能，统一界面设计风格
- **数据管理**: 完善类型定义，加强数据验证和错误处理

### 📚 Documentation
- **UAT测试文档**: 更新测试计划反映CRUD功能恢复状态
- **README更新**: 添加最新功能说明和项目结构更新
- **CHANGELOG创建**: 建立版本变更记录跟踪

### 🔒 Technical
- **版本稳定性**: 使用经过验证的稳定版本组合
- **依赖管理**: 优化依赖结构，移除不必要的GraphQL相关包
- **类型安全**: 完善TypeScript类型定义，加强类型检查

## [1.4.0] - 2025-07-29

### Added
- ES模块兼容性修复
- Ant Design组件库集成
- 基础页面架构搭建

### Fixed
- 模块导入错误
- 构建过程问题

## [1.3.0] - 2025-07-28

### Added
- 初始项目架构
- 基础配置文件
- 开发环境设置

---

## 版本说明

- **Major** (x.0.0): 不兼容的API变更
- **Minor** (0.x.0): 新增功能，向后兼容
- **Patch** (0.0.x): 问题修复，向后兼容

## 贡献指南

请在提交PR时更新此CHANGELOG文件，描述你的更改。

## 支持

如有问题，请访问 [GitHub Issues](../../issues)。