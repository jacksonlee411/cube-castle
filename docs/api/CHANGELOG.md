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

## [v4.6.0] - 2025-08-30 🎯 **审计系统精确化重构**

### 🚨 重大变更 - 审计系统架构调整
- **变更查询结构**: 移除`organizationAuditHistory`查询，新增`auditHistory(recordId: UUID!)`查询
  - **理由**: 审计记录现在精确追踪每个时态版本(record_id)，而不是组织代码级别
  - **影响**: 前端需要传递具体的recordId而不是组织代码来获取审计信息
- **变更AuditLogDetail类型**: 移除`businessEntityId`字段，保留`recordId`字段作为唯一标识
  - **理由**: 简化审计模型，每个审计记录直接关联到具体的时态版本记录
  - **影响**: 审计记录现在只返回recordId，不再返回组织代码

### 📋 审计查询优化
- **新增查询**: `auditHistory(recordId: UUID!, startDate: Date, endDate: Date, operation: OperationType, userId: UUID, limit: Int)`
  - **功能**: 获取特定temporal版本的完整审计历史
  - **性能**: 优化索引设计，查询响应时间<50ms
- **移除查询**: `organizationAuditHistory(code: String!)`
  - **理由**: 替换为更精确的record_id级别查询
- **保留查询**: `auditLog(auditId: String!)`单条审计记录查询保持不变

### 🗑️ 清理废弃类型
- **移除类型**: `OrganizationAuditHistory`, `AuditTimelineEntry`, `AuditHistoryMeta`, `ChangesSummary`
  - **理由**: 这些类型支持已废弃的组织级别审计查询
- **简化响应**: `auditHistory`查询直接返回`[AuditLogDetail!]!`数组
  - **理由**: 减少嵌套结构，简化前端集成

### 📈 版本升级指南
- **前端修改**: 将`AuditAPI.getOrganizationAuditHistory(code)`替换为`AuditAPI.getRecordAuditHistory(recordId)`
- **查询调整**: 在组织详情页面传递当前选中版本的recordId而不是组织代码
- **类型更新**: 更新TypeScript接口，移除`businessEntityId`相关字段

---

## [v4.5.0] - 2025-08-30 📋 **审计功能实用化简化** 

### 🗑️ 移除 - 完全删除organizationChangeAnalysis
- **删除GraphQL查询**: 完全移除`organizationChangeAnalysis`查询端点
  - **理由**: 遵循API优先原则，删除未实现且过度设计的功能
- **删除类型定义**: 移除`OrganizationChangeAnalysis`、`VersionRange`、`AnalysisType`类型
  - **理由**: 清理不再使用的类型定义，减少API表面积
- **更新契约测试**: 调整前端测试预期查询数量从10个减少到9个
  - **理由**: 保持测试与实际API规范的一致性

### 📋 专注 - 核心审计功能
- **保留核心查询**: `organizationAuditHistory`和`auditLog`查询
  - **理由**: 这些提供实际业务价值的基础审计功能
- **简化API表面积**: GraphQL查询从10个减少到9个
  - **理由**: 更小的API表面积更易维护和理解

---

## [v4.4.0] - 2025-08-30 🗑️ **完全移除过度设计功能**

### 🎯 移除 - 完全删除organizationChangeAnalysis
- **删除GraphQL查询**: 完全移除`organizationChangeAnalysis`查询端点
  - **理由**: 遵循API优先原则，删除未实现且过度设计的功能
- **删除类型定义**: 移除`OrganizationChangeAnalysis`、`VersionRange`、`AnalysisType`类型
  - **理由**: 清理不再使用的类型定义，减少API表面积
- **更新契约测试**: 调整前端测试预期查询数量从10个减少到9个
  - **理由**: 保持测试与实际API规范的一致性

### 📋 专注 - 核心审计功能
- **保留核心查询**: `organizationAuditHistory`和`auditLog`查询
  - **理由**: 这些提供实际业务价值的基础审计功能
- **简化API表面积**: GraphQL查询从10个减少到9个
  - **理由**: 更小的API表面积更易维护和理解

### 📚 文档同步
- **清理API规范**: 移除API表格和权限映射中的相关引用
- **更新开发计划**: 清理过时的实现计划和进度记录
- **版本升级**: Schema版本从v4.3.0升级至v4.4.0

### 🎯 设计原则强化
- **API优先原则**: 严格遵循"先定义API契约，后实现代码"
- **实用主义**: 专注解决实际业务问题，避免过度工程化
- **可维护性**: 保持小而专注的API表面积

---

## [v4.3.0] - 2025-08-30 ✂️ **过度设计简化重构**

### 🎯 移除 - 过度设计的审计分析功能
- **移除复杂趋势分析**: 从`OrganizationChangeAnalysis`中移除`TrendAnalysis`类型
  - 移除字段: `changeFrequency`, `stabilityScore`, `riskTrend`
  - **理由**: 需要6-9个月开发周期，超出实际业务需求，缺乏行业标准基准
- **移除自动影响评估**: 从`OrganizationChangeAnalysis`中移除`ImpactAssessment`类型
  - 移除字段: `overallImpact`, `affectedSystems`, `mitigationRequired`
  - **理由**: 需要复杂的跨系统集成分析，当前数据基础不足
- **移除智能建议生成**: 从`OrganizationChangeAnalysis`中移除`recommendations`字段
  - **理由**: 需要领域专家知识库和业务规则引擎，实现复杂度高但用户信任度低

### 📋 保留 - 核心业务价值功能
- **保留变更摘要**: `ChangesSummary`类型完整保留
  - 包含: `operationSummary`, `totalChanges`, `keyChanges`
  - **理由**: 提供实用的审计信息，满足实际HR管理需求
- **保留时间范围**: `VersionRange`类型完整保留
  - **理由**: 基础时态查询必需信息

### 🎨 优化 - API文档更新
- **简化功能描述**: 更新`organizationChangeAnalysis`查询说明
- **新增设计原则**: 明确"专注核心业务价值，避免过度工程化"
- **版本升级**: Schema版本从v4.2.1升级至v4.3.0

### 📊 影响评估
- **开发成本降低**: 从6-9个月减少至2-4周实现周期  
- **维护成本降低**: 移除复杂算法逻辑，降低长期维护负担
- **用户价值提升**: 专注实际需求，提供清晰的审计历史信息

---

## [v4.2.2] - 2025-08-27 🏗️ **单表时态架构API优化**

### 🔧 简化 - GraphQL Schema优化
- **移除versionSequence字段**: 从`AuditLogDetail`和`AuditTimelineEntry`类型中移除不必要的版本序列概念
  - **理由**: 单表时态架构使用`recordId: UUID`已足够标识版本，`versionSequence`引入不必要复杂性
  - **影响**: 简化客户端集成逻辑，直接使用数据库原生UUID标识符
- **优化VersionRange类型**: 重构为基于日期的范围查询
  - **变更**: `fromVersion/toVersion: Int` → `fromDate/toDate: Date`
  - **变更**: `totalVersions: Int` → `totalRecords: Int` 
  - **理由**: 与单表时态架构的`effective_date`字段直接对应，避免版本号映射复杂性
- **简化organizationChangeAnalysis查询**: 参数从版本号改为日期范围
  - **变更**: 查询参数从`fromVersion/toVersion`改为`fromDate/toDate`
  - **性能提升**: 利用26个专用时态索引，查询时间从300ms降至100ms
  - **理由**: 直接利用PostgreSQL时态索引，避免版本号查找的额外开销

### 📋 架构对齐
- **API文档与数据库架构一致性**: 移除多表概念引入的复杂性，直接反映单表时态架构的简洁性
- **性能文档更新**: 更新查询性能预期，基于26个专用索引的实际表现
- **术语标准化**: 统一使用基于日期的时态查询语义

### ⚠️ 破坏性变更
- 客户端需要更新时态查询逻辑，从版本号切换到日期范围
- `AuditLogDetail`和`AuditTimelineEntry`响应结构中移除`versionSequence`字段
- `organizationChangeAnalysis`查询参数接口变更

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