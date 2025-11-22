# 开发计划文档归档记录

**更新日期**: 2025-11-03
**归档者**: Claude Code 自动化工具
**归档原因**: 执行203号HRMS系统模块化演进计划是下一步工作方向，对于历史性遗留任务文档进行集中归档

---

## 第一批归档（2025-11-03）

### 归档概览

本批次归档6个文档（包括之前的06号），都是在203号HRMS计划发起前的历史性工作文档。相关结论和成果已纳入其他主要文档或已完成，后续工作将在203号HRMS系统模块化演进计划框架下进行。

### 归档清单

#### 1. 70-temporal-timeline-lifecycle-investigation.md
- **编号**: 70
- **标题**: 组织时间轴全生命周期连贯性调查报告
- **创建日期**: 2025-10-17
- **状态**: 草稿（调查结论已形成）
- **类型**: 调查报告
- **原因**: 时间轴治理的调查结论已用于指导实施，后续工作在203号HRMS计划中进行
- **关联文档**:
  - `docs/architecture/temporal-timeline-consistency-guide.md`
  - `docs/architecture/temporal-consistency-implementation-report.md`

#### 2. 105-navigation-ui-alignment-fix.md
- **编号**: 105
- **标题**: 导航栏 UI 对齐与布局优化
- **创建日期**: 2025-10-20
- **完成日期**: 2025-10-20
- **状态**: ✅ 已完成
- **类型**: UI/UX 改进任务
- **优先级**: 中
- **原因**: UI 改进任务已完成，后续维护在203号计划中进行
- **主要工作**:
  - 下三角符号位置修正（Canvas Kit 规范）
  - 图标对齐问题处理
  - 其他UI细节优化

#### 3. 107-position-closeout-gap-report.md
- **编号**: 107
- **标题**: 职位管理收口差距核查报告
- **创建日期**: 2025-10-21
- **最终更新**: 2025-10-21 18:30 UTC
- **版本**: v2.0 归档确认版
- **状态**: ✅ 完成
- **类型**: 模块收口报告
- **编制团队**: 架构组、前端团队、后端团队、QA团队
- **原因**: 职位管理模块Stage 4已完成，归档材料齐备
- **关联计划**: 80号职位管理方案、86号任职Stage 4计划、99号收口顺序指引
- **主要成果**:
  - Go 单元测试覆盖率达标（auth 80.8%、cache 85.1%、graphql 87.0%）
  - 认证/缓存/GraphQL 核心模块补齐单测
  - 验收清单完成

#### 4. 109-position-audit-history-realignment.md
- **编号**: 109
- **标题**: 职位审计历史缺失整改计划
- **创建日期**: 2025-10-22
- **版本**: v1.0（执行中）
- **状态**: 方案已定
- **类型**: 整改计划
- **责任团队**: 后端查询服务组（主责）、职位领域前端组（协同）
- **原因**: 职位管理阶段性工作总结，后续在203号HRMS计划中推进
- **关键问题**:
  - 职位详情页审计历史显示缺失
  - AuditHistorySection 仅显示"暂无审计记录"
  - GraphQL 查询返回空数组

#### 5. 110-position-status-normalization.md
- **编号**: 110
- **标题**: 职位版本状态与"当前版本"标识异常整改
- **创建日期**: 2025-10-22
- **版本**: v1.0（执行中）
- **状态**: 方案已定
- **类型**: 整改计划
- **责任团队**: 职位领域后端组（主责）、职位前端组（协同）
- **原因**: 职位时态处理的历史遗留任务，后续在203号HRMS计划中统一处理
- **关键问题**:
  - 职位版本列表状态异常显示（均显示PLANNED）
  - "当前版本"标识混乱
  - 历史版本状态显示不准确

#### 6. 06-design-review-task-assessment.md（之前归档）
- **编号**: 06
- **标题**: 06号设计评审任务确认报告
- **创建日期**: 2025-10-20
- **完成日期**: 2025-10-20
- **状态**: ✅ 已完成
- **类型**: 设计评审完成报告
- **原因**: 设计评审任务已完成
- **内容**: Job Catalog 列表/详情在新导航下的设计评审结论

---

## 归档统计

| 指标 | 数值 |
|------|------|
| 本批次归档数量 | 5个 |
| 包含之前归档数量 | 1个（06号） |
| 总计归档数量（第一批） | 6个 |
| 完成状态占比 | 50% 已完成，50% 方案已定 |
| 主要类型 | 调查报告、改进任务、模块报告、整改计划 |

---

## 追加归档（2025-11-21 · Plan 271 守卫）

| 计划 | 路径 | 说明 |
|------|------|------|
| 264 | `docs/archive/development-plans/264-workflow-governance.md` | 2025-11-20 已完成并归档，因 2025-11-21 合并误将活跃副本带回，现依据 Plan 271 再次确认仅保留归档版本 |
| 265 | `docs/archive/development-plans/265-selfhosted-required-checks.md` | 同上，归档内容保留、活跃副本移除 |
| 266 | `docs/archive/development-plans/266-selfhosted-tracking.md` | 同上 |
| 267 | `docs/archive/development-plans/267-docker-network-stability.md` | 同上 |
| 268 | `docs/archive/development-plans/268-github-actions-vendoring.md` | 同上 |
| 269 | `docs/archive/development-plans/269-wsl-runner-deployment.md` | 同上 |

> 备注：本次追加归档由 Plan 271 文档治理守卫发起，记录了资源唯一性违规的调查与整改。归档动作在 30 分钟内完成 push，并通过 `npm run lint:docs` 自检；Agents Compliance workflow 已新增 Plan 271 Guard 作为门禁。

---

## 后续工作方向

### 203号HRMS系统模块化演进计划作为新工作焦点

归档这些历史文档的同时，确认了以下工作方向：

1. **核心系统架构**
   - 模块化单体架构设计
   - DDD 界定上下文
   - CQRS 模式实施

2. **技术栈升级**
   - 数据访问层演进（sqlc/Ent）
   - 事务性发件箱强制要求
   - 数据库迁移治理
   - 权限策略外部化

3. **模块开发**
   - Core HR 域（organization、workforce、contract）
   - Talent Management 域（recruitment、performance、development）
   - Compensation & Operations 域（compensation、payroll、attendance、compliance）

4. **质量保证**
   - 编译期类型安全
   - Docker 真实数据库测试
   - 连接池配置标准

---

## 归档维护说明

### 文件位置
- **原位置**: `docs/development-plans/`
- **新位置**: `docs/archive/development-plans/`

### 访问方式
所有归档文件可通过以下方式访问：
```bash
# 查看归档文件列表
ls docs/archive/development-plans/

# 查看特定文件
cat docs/archive/development-plans/70-temporal-timeline-lifecycle-investigation.md
```

### 引用方式
若需在其他文档中引用归档文件，使用相对路径：
```markdown
详见 `../archive/development-plans/70-temporal-timeline-lifecycle-investigation.md`
```

---

## 后续归档计划

- **下一批次**: 定于2025年11月底前，归档其他已完成的辅助性文档
- **定期检查**: 按季度检查 development-plans 目录中状态为"✅ 已完成"的文档
- **归档标准**: 仅当任务完全完成且成果已在主文档中集成时进行归档

---

**文档维护者**: 架构团队
**最后更新**: 2025-11-03
**下一次审视**: 2025-11-30
