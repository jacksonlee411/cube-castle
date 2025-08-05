# DOCS2 - 规范化文档中心

**创建时间**: 2025-08-04  
**目的**: 职位API路由规范化处理过渡期文档管理  
**状态**: 进行中

## 📁 文档结构

### `/api-specifications/` - API规范文档
- `positions-api-specification.md` - 职位管理API完整规范
- `api-design-principles.md` - API设计原则和标准
- `openapi-schemas/` - OpenAPI规范文件

### `/architecture-decisions/` - 架构决策记录
- `ADR-001-positions-api-architecture.md` - 职位API架构选择决策
- `ADR-002-route-standardization.md` - 路由标准化决策

### `/implementation-guides/` - 实施指南
- `frontend-api-integration.md` - 前端API集成指南
- `backend-api-implementation.md` - 后端API实现指南
- `testing-guidelines.md` - API测试指南

### `/standards/` - 技术标准
- `coding-standards.md` - 代码规范
- `documentation-standards.md` - 文档规范
- `api-versioning-policy.md` - API版本管理策略

## 🎯 规范化目标

1. **统一职位API路由**: 标准化使用 `/api/v1/positions`
2. **消除架构混淆**: 清理 `/api/v1/corehr/positions` 相关引用
3. **提供清晰指导**: 为开发者提供明确的使用规范
4. **建立长期标准**: 确保未来开发的一致性

## 📋 当前状态

- ✅ 文档结构创建完成
- 🔄 正在创建核心规范文档
- ⏳ 计划更新现有文档
- ⏳ 计划清理代码注释

## 🔗 相关链接

- [原始API设计文档](../docs/employee-model-implementation/API接口设计与集成规范.md)
- [职位处理器实现](../go-app/internal/handler/position_handler.go)
- [前端API客户端](../nextjs-app/src/lib/api/positions.ts)

## 📞 联系信息

如有疑问，请参考相关文档或联系开发团队。

---
*此文档是Cube Castle项目职位API规范化处理的一部分*