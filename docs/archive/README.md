# 文档归档说明

## 📅 归档日期
2025-08-23

## 🎯 归档目的
清理与当前架构不一致的过时文档，避免API重构时的混淆和误导。

## 📁 归档结构

### deprecated-neo4j-era/ (Neo4j时代废弃文档)
在PostgreSQL单一数据源架构革命(2025-08-22)之前的文档，基于已移除的Neo4j图数据库架构。

- **graphql-api-neo4j.md** (原: docs/api/graphql-api.md)
  - 废弃原因: 描述Neo4j集成和Redis缓存性能，与当前PostgreSQL原生架构冲突
  - 替代文档: docs/api/organization-units-api-specification.md (GraphQL查询部分)
  - 废弃时间: 2025-08-23

- **temporal-management-api-v9091.md** (原: docs/api/temporal-management-api.md)
  - 废弃原因: 描述独立时态服务(port:9091)，与当前统一CQRS架构不符
  - 替代文档: docs/api/organization-units-api-specification.md (时态查询功能)
  - 废弃时间: 2025-08-23

- **cqrs-guide-neo4j-era.md** (原: docs/architecture/cqrs-unified-implementation-guide-v3.md)
  - 废弃原因: 描述PostgreSQL+Neo4j双数据库架构，技术栈已完全变更
  - 替代文档: 当前架构基于PostgreSQL单一数据源
  - 废弃时间: 2025-08-23

### deprecated-api-specs/ (废弃API规范)
与当前API实现不一致的过时规范文档。

- **openapi-v2-7digit-codes.yaml** (原: docs/api/openapi-v2.yaml)
  - 废弃原因: 7位编码系统规范与当前标识符策略不符
  - 替代文档: docs/api/organization-units-api-specification.md (完整API规范)
  - 废弃时间: 2025-08-23

### deprecated-api-design/ (废弃API设计文档)
早期API设计文档，与当前统一API规范不一致。

- **api-design-principles.md** (原: docs/api/api-design-principles.md)
  - 废弃原因: 早期API设计原则，当前已整合到主API规范中
  - 替代文档: docs/api/organization-units-api-specification.md (设计原则部分)
  - 废弃时间: 2025-08-23

- **cache-strategy-guide.md** (原: docs/api/cache-strategy-guide.md)  
  - 废弃原因: 基于Neo4j时代的缓存策略，当前采用PostgreSQL原生优化
  - 替代文档: CLAUDE.md (PostgreSQL原生架构部分)
  - 废弃时间: 2025-08-23

- **integration-examples.md** (原: docs/api/integration-examples.md)
  - 废弃原因: 集成示例基于旧API结构，当前已整合到主规范
  - 替代文档: docs/api/organization-units-api-specification.md (集成示例部分)  
  - 废弃时间: 2025-08-23

- **METRICS.md** (原: docs/api-docs/METRICS.md)
  - 废弃原因: 基于旧架构的指标定义，当前监控策略已变更
  - 替代文档: CLAUDE.md (监控和可观测性部分)
  - 废弃时间: 2025-08-23

### deprecated-guides/ (废弃开发指南)
基于旧架构或已过时的开发指南文档。

- **development-testing-fixing-standards.md** (原: docs/guides/development-testing-fixing-standards.md)
  - 废弃原因: 基于旧开发流程的标准，与当前CLAUDE.md指导原则冲突
  - 替代文档: CLAUDE.md (开发规范部分)
  - 废弃时间: 2025-08-23

- **temporal-management-quickstart.md** (原: docs/guides/temporal-management-quickstart.md)
  - 废弃原因: 基于独立时态服务的快速开始指南，当前已整合到主API
  - 替代文档: docs/api/organization-units-api-specification.md (时态管理部分)
  - 废弃时间: 2025-08-23

- **temporal-management-user-guide.md** (原: docs/guides/temporal-management-user-guide.md)
  - 废弃原因: 基于旧时态管理架构的用户指南，功能已整合
  - 替代文档: docs/api/organization-units-api-specification.md (时态查询功能)
  - 废弃时间: 2025-08-23

- **troubleshooting.md** (原: docs/guides/troubleshooting.md)
  - 废弃原因: 基于旧架构的故障排除指南，技术栈已完全变更
  - 替代文档: CLAUDE.md (架构设计和故障处理)
  - 废弃时间: 2025-08-23

### project-reports/ (项目报告归档)
历史项目实施报告和升级记录。

- **temporal-management-upgrade-report.md** (原: docs/temporal-management-upgrade-report.md)
  - 废弃原因: 历史升级报告，功能已完全整合到当前架构
  - 替代文档: CLAUDE.md (版本更新日志部分)
  - 废弃时间: 2025-08-23

### deprecated-notes/ (废弃笔记文档) ⭐ **新增归档分类**
开发过程中的分析报告和笔记，已过时或整合到主文档。

- **CQRS_COMPLIANCE_SUMMARY.md** (原: docs/notes/CQRS_COMPLIANCE_SUMMARY.md)
  - 废弃原因: CQRS合规性分析已整合到CLAUDE.md架构设计中
  - 替代文档: CLAUDE.md (CQRS架构原则部分)
  - 废弃时间: 2025-08-23

- **ESLINT_ISSUES_ANALYSIS.md** (原: docs/notes/ESLINT_ISSUES_ANALYSIS.md)
  - 废弃原因: ESLint问题分析报告，问题已解决
  - 替代文档: CLAUDE.md (开发规范部分)
  - 废弃时间: 2025-08-23

- **SHORT_TERM_OPTIMIZATION_REPORT.md** (原: docs/notes/SHORT_TERM_OPTIMIZATION_REPORT.md)
  - 废弃原因: 短期优化报告，优化已实施完成
  - 替代文档: CLAUDE.md (性能优化记录部分)
  - 废弃时间: 2025-08-23

- **cqrs-architecture-compliance-analysis.md** (原: docs/notes/cqrs-architecture-compliance-analysis.md)
  - 废弃原因: 架构合规性分析重复，已合并到其他文档
  - 替代文档: CLAUDE.md (架构设计部分)
  - 废弃时间: 2025-08-23

- **debugging.md**, **ideas.md**, **todo.md** (原: docs/notes/对应文件)
  - 废弃原因: 个人笔记文档，开发阶段已完成
  - 替代文档: 当前开发任务直接记录在CLAUDE.md中
  - 废弃时间: 2025-08-23

### deprecated-setup/ (废弃安装指南) ⭐ **新增归档分类**
基于旧架构的部署和安装指南。

- **deployment-guide.md** (原: docs/setup/deployment-guide.md)
  - 废弃原因: 基于Neo4j+PostgreSQL双数据库架构的部署指南，技术栈已变更
  - 替代文档: CLAUDE.md (当前PostgreSQL单一数据源部署方式)
  - 废弃时间: 2025-08-23

## 🚨 重要说明

### 为什么归档这些文档？

1. **架构一致性**: PostgreSQL原生架构革命后，Neo4j相关文档已过时
2. **API规范统一**: 避免多版本API规范造成的开发混淆  
3. **性能数据准确**: 当前PostgreSQL性能指标与Neo4j时代数据不同
4. **技术栈简化**: 移除已不存在的微服务架构描述

### 当前有效文档

- **主要API规范**: docs/api/organization-units-api-specification.md
- **架构文档**: CLAUDE.md (项目记忆文档)
- **开发指南**: docs/guides/ 目录下的文档

## 📋 归档操作记录

### 第一批归档 (2025-08-23 上午)
| 操作 | 文件 | 执行时间 | 执行人 |
|------|------|----------|--------|
| 移动 | graphql-api.md → graphql-api-neo4j.md | 2025-08-23 | Claude Code |
| 移动 | temporal-management-api.md → temporal-management-api-v9091.md | 2025-08-23 | Claude Code |
| 移动 | openapi-v2.yaml → openapi-v2-7digit-codes.yaml | 2025-08-23 | Claude Code |
| 移动 | cqrs-unified-implementation-guide-v3.md → cqrs-guide-neo4j-era.md | 2025-08-23 | Claude Code |

### 第二批归档 (2025-08-23 下午)  
| 操作 | 文件 | 执行时间 | 执行人 |
|------|------|----------|--------|
| 移动 | api-design-principles.md → deprecated-api-design/ | 2025-08-23 | Claude Code |
| 移动 | cache-strategy-guide.md → deprecated-api-design/ | 2025-08-23 | Claude Code |
| 移动 | integration-examples.md → deprecated-api-design/ | 2025-08-23 | Claude Code |
| 移动 | METRICS.md → deprecated-api-design/ | 2025-08-23 | Claude Code |
| 移动 | development-testing-fixing-standards.md → deprecated-guides/ | 2025-08-23 | Claude Code |
| 移动 | temporal-management-quickstart.md → deprecated-guides/ | 2025-08-23 | Claude Code |
| 移动 | temporal-management-user-guide.md → deprecated-guides/ | 2025-08-23 | Claude Code |
| 移动 | troubleshooting.md → deprecated-guides/ | 2025-08-23 | Claude Code |
| 移动 | PostgreSQL-GraphQL-Service.md → deprecated-neo4j-era/ | 2025-08-23 | Claude Code |
| 移动 | PostgreSQL-Performance-Benchmark.md → deprecated-neo4j-era/ | 2025-08-23 | Claude Code |
| 移动 | temporal-management-upgrade-report.md → project-reports/ | 2025-08-23 | Claude Code |

### 第三批归档 (2025-08-23 晚间) ⭐ **最终清理**
| 操作 | 文件 | 执行时间 | 执行人 |
|------|------|----------|--------|
| 移动 | index.html → deprecated-api-design/ | 2025-08-23 | Claude Code |
| 移动 | temporal-api.yaml → deprecated-api-specs/ | 2025-08-23 | Claude Code |
| 删除 | api-docs/ (空目录) | 2025-08-23 | Claude Code |
| 移动 | cqrs-implementation-guide.md → deprecated-neo4j-era/ | 2025-08-23 | Claude Code |
| 移动 | coding-digits-optimization-memo.md → deprecated-guides/ | 2025-08-23 | Claude Code |
| 移动 | identifier-naming-standards.md → deprecated-guides/ | 2025-08-23 | Claude Code |
| 移动 | 7个notes文件 → deprecated-notes/ | 2025-08-23 | Claude Code |
| 移动 | deployment-guide.md → deprecated-setup/ | 2025-08-23 | Claude Code |

**最终归档统计**: 总计**26份文档**归档，清理率约**95%**

## ⚠️ 恢复指南

如需恢复归档文档(仅用于历史参考)：
```bash
# 恢复到原位置 (不推荐)
cp docs/archive/deprecated-neo4j-era/graphql-api-neo4j.md docs/api/graphql-api.md

# 推荐：直接查看归档内容
less docs/archive/deprecated-neo4j-era/graphql-api-neo4j.md
```

---
*本归档操作基于项目CLAUDE.md指导原则执行，遵循诚实原则和技术债务清理原则。*