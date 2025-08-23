# 📝 Cube Castle API 规范变更日志

本文件记录Cube Castle API规范文件 (`openapi.yaml` 和 `schema.graphql`) 的所有重要变更。

## 格式说明

- **新增** - 新功能
- **变更** - 现有功能的修改  
- **废弃** - 即将移除的功能
- **移除** - 已删除的功能
- **修复** - 错误修正
- **安全** - 安全相关的修正

---

## [v4.2.1] - 2025-08-23 🔧 **文档完善和增强**

### 🔧 改进
- **profile字段增强**: 为不同unitType提供详细的配置结构说明
  - DEPARTMENT: budget, managerPositionCode, costCenterCode, headCountLimit, establishedDate
  - COMPANY: legalName, registrationNumber, taxId, industry, incorporationDate  
  - PROJECT_TEAM: projectCode, projectManager, startDate, endDate, budget
  - ORGANIZATION_UNIT: function, region, parentType
- **枚举类型补充**: 在OpenAPI规范中添加缺失的枚举定义
  - AnalysisType: SUMMARY, DETAILED
  - ConsistencyCheckMode: FAST, DEEP, TARGETED
  - SearchField: NAME, DESCRIPTION, CODE_PATH, NAME_PATH
  - SortField: CODE, NAME, CREATED_AT, UPDATED_AT, EFFECTIVE_DATE, LEVEL, SORT_ORDER
  - SortOrder: ASC, DESC

### 📈 符合度改进
- **100%端点覆盖**: 确认batch-refresh-hierarchy端点完整实现
- **配置文档完善**: profile字段的类型特定配置结构说明
- **跨协议一致性**: OpenAPI和GraphQL枚举定义完全对应

## [v4.2.0] - 2025-08-23 🚀 **API规范正式发布**

### 🆕 新增
- **OpenAPI 3.0.3规范**: 创建完整的REST API规范文档 (`openapi.yaml`)
  - 11个REST端点的完整定义 (CRUD + 业务操作 + 运维工具)
  - OAuth 2.0 Client Credentials Flow认证
  - 17个细粒度PBAC权限模型
  - 统一企业级响应信封结构
  - 完整的错误处理体系 (401/403/400/500)

- **GraphQL Schema定义**: 创建完整的GraphQL Schema (`schema.graphql`)
  - 10个GraphQL查询的完整类型定义
  - 时态查询支持 (asOfDate参数)
  - 层级查询优化 (17级深度支持)
  - 审计和分析查询功能
  - 性能优化的输入/输出类型

- **Single Source of Truth机制**: 建立API规范的权威来源机制
  - 严格的变更管理流程
  - 版本控制标准
  - 自动化验证机制

### 🔧 特性
- **严格CQRS架构合规**: 查询操作仅GraphQL，命令操作仅REST
- **企业级安全模型**: OAuth 2.0 + JWT + PBAC权限体系
- **时态数据支持**: 完整的历史版本管理和未来生效计划
- **智能层级管理**: 17级深度 + 双路径系统 + 自动级联更新
- **性能优化**: 26个专用索引，响应时间 < 200ms目标

### 📋 端点总览

#### REST API端点 (11个)
- **标准CRUD** (4个): POST, PUT, PATCH, DELETE `/api/v1/organization-units`
- **业务操作** (3个): suspend, activate, validate
- **运维工具** (2个): refresh-hierarchy, batch-refresh-hierarchy  
- **CoreHR兼容** (2个): 兼容性创建端点

#### GraphQL查询 (10个)
- **基础查询** (3个): organizations, organization, organizationStats
- **层级查询** (3个): organizationHierarchy, organizationSubtree, hierarchyStatistics
- **审计查询** (3个): organizationAuditHistory, auditLog, organizationChangeAnalysis
- **运维查询** (1个): hierarchyConsistencyCheck

### 🔒 安全特性
- **OAuth 2.0 Client Credentials Flow**: 企业级机器对机器认证
- **JWT标准载荷**: 权限、租户、审计信息
- **17个核心权限**: org:read, org:create, org:update, org:delete, org:suspend, org:reactivate, 等
- **多租户隔离**: 严格的租户数据边界
- **审计追踪**: 完整的操作记录和责任追溯

### 📊 数据模型
- **OrganizationUnit**: 40+字段的完整组织模型
- **时态字段**: effectiveDate, endDate, isCurrent, isFuture
- **审计字段**: operationType, operatedBy, operationReason, recordId
- **层级字段**: level, hierarchyDepth, codePath, namePath
- **配置字段**: profile (JSONB), 支持动态配置

### 🎯 性能目标
- **GraphQL查询**: < 200ms (实际1.5-8ms)
- **REST创建**: < 300ms
- **REST更新**: < 200ms  
- **层级刷新**: < 2000ms
- **并发支持**: 1000+ QPS (查询), 100+ TPS (命令)

---

## 📈 版本统计

- **当前版本**: v4.2.0
- **总端点数**: 21个 (11 REST + 10 GraphQL)
- **权限数量**: 17个细粒度权限
- **数据模型**: 25+核心类型定义
- **文档页面**: 1000+ 行完整规范

---

## 🔄 兼容性说明

### v4.2.0 兼容性
- **初始发布**: 无兼容性问题
- **命名规范**: 统一camelCase字段命名
- **协议分离**: 严格CQRS架构实施
- **错误格式**: 统一企业级错误响应结构

### 未来版本策略
- **主版本** (v5.0): 破坏性API变更
- **次版本** (v4.3): 新功能添加，向后兼容
- **修订版本** (v4.2.1): Bug修复和性能优化

---

## 📞 变更支持

### 如何提交变更
1. **创建Issue**: 描述API变更需求和业务场景
2. **规范设计**: 先在规范文件中设计变更
3. **社区讨论**: 通过PR进行技术讨论
4. **测试验证**: 确保变更不破坏现有功能
5. **文档更新**: 同步更新相关文档

### 联系方式
- **API支持**: api-support@yourcompany.com
- **技术讨论**: 企业内部技术委员会
- **紧急问题**: 通过内部Issue跟踪系统

---

**📍 说明**: 本CHANGELOG遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/) 格式，采用 [语义化版本](https://semver.org/lang/zh-CN/) 规范。

**🔄 最后更新**: 2025-08-23  
**📋 下次版本**: v4.3.0 (计划2025-11-23)