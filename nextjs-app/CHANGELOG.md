# Changelog

All notable changes to the Cube Castle Next.js Frontend will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0-alpha.1] - 2025-07-31

### 🎯 Breaking Changes
- **完全移除Ant Design**: 彻底移除antd和@ant-design/icons依赖
- **现代化UI架构**: 全面迁移至shadcn/ui + Radix UI + Tailwind CSS技术栈

### ✨ Added
- **现代化组件基础**: 建立基于Radix UI的现代组件系统
- **测试UI页面**: 新增`/test-ui`页面展示现代UI组件库
- **临时页面占位**: 为重构期间的复杂页面提供用户友好的占位界面

### 🔧 Technical Infrastructure
- **构建系统优化**: 修复ESLint配置，支持现代React + TypeScript工作流
- **依赖清理**: 移除所有antd相关导入，清理17个核心文件
- **类型安全**: 优化TypeScript配置，提升类型检查严格性
- **组件接口**: 建立统一的UI组件接口标准

### 🚧 In Progress (Temporary State)
以下页面已替换为占位符，等待Phase 2-3重构：
- `src/pages/workflows/demo.tsx` - 工作流演示页面
- `src/pages/workflows/[id].tsx` - 工作流详情页面  
- `src/pages/admin/graph-sync.tsx` - 图数据同步页面
- `src/pages/positions/index.tsx` - 职位管理页面
- `src/pages/employees/positions/[id].tsx` - 员工职位历史页面
- `src/pages/organization/chart.tsx` - 组织架构图页面

### 📚 Documentation
- **实施方案更新**: UI组件库标准化实施方案增加详细进度记录
- **架构决策记录**: 记录从Ant Design迁移的技术决策和实施细节

### ⚠️ Important Notes
- 当前版本为重构过渡状态，6个核心页面功能暂不可用
- 建议仅在开发环境使用，等待Phase 2-3完成后部署生产环境
- 所有现有功能将在后续版本中以现代化形式恢复

---

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