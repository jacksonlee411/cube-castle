# 开发计划文档目录

**建立时间**: 2025-08-23  
**用途**: 组织架构重构项目的计划/路线/验收统一入口  
**状态**: 活跃  
**排序**: 以阅读优先级与依赖顺序排列

## 目录边界与唯一事实来源
- 本目录仅存储“计划、路线、阶段性报告”，验收完成的文档需在 30 分钟内迁入 `../archive/development-plans/`，并在此 README 中保留索引。
- 规范/架构类的长期资料分别放在 `../reference/` 与 `../architecture/`；若发现描述冲突，以参考/架构目录为准并立即回写计划文档。
- 自 2025-11-21 起执行 **Plan 271 Guard**：`npm run lint:docs` 会阻断活跃与归档目录存在同名文件，提交前必须通过该守卫。
- 数据库、契约与权限均以迁移文件及 `docs/api/openapi.yaml`、`docs/api/schema.graphql` 为唯一事实来源，严禁引用第二事实来源。

## 维护流程（默认四步）
1. 启动新能力前先查 `../reference/02-IMPLEMENTATION-INVENTORY.md`、`docs/api/*`，确认不存在现成方案。
2. 在本目录新增或更新计划，写清范围、验收标准、回滚策略与事实来源链接。
3. 执行过程中同步更新依赖与风险，必要时创建子文档并保持互链。
4. 验收通过后将文件迁入归档目录，更新本 README 索引，并将变更在 30 分钟内推送远程。

## 活跃计划速览
- `00-README.md`（本文）—— 使用指南与索引原则。
- `02-technical-architecture-design.md`—— 当前技术架构说明，所有计划需引用其边界。
- `06-integrated-teams-progress-log.md`—— 团队协作依赖与跨域进展。
- `200-Go语言ERP系统最佳实践.md` / `201-Go实践对齐分析.md`—— ERP 实践基线及项目差距分析。
- `203-hrms-module-division-plan.md`—— **重点** 模块化单体边界（v2.0 95% 完成，协调 204/206）。
- `204-HRMS-Implementation-Roadmap.md`—— HRMS 实施路线图，对齐 203 的阶段目标。
- `206-Alignment-With-200-201.md`—— 200/201/203/204 的一致性校验与差异登记。
- `271-archive-integrity-guard.md`—— Plan 271，归档守卫、CI 自动化与 `lint:docs` 规则。
- `272-runtime-artifact-cleanup.md`—— Plan 272，运行产物与 cloc 噪音压降（Stage1 logs/reports/test-results，Stage2 vendored 依赖治理）。
- `400-standard-object-model-plan.md`—— 将组织/职位等对象抽象为统一 SOM，覆盖生命周期、契约、UI、迁移策略。

> 新增计划需遵循 `docs/development-plans/<id>-<slug>.md` 命名，并立即在此处补充一句描述与事实来源。

## 归档信息摘要
- **2025-11-23 最新归档**  
  - `../archive/development-plans/222-organization-verification.md` —— Phase2 验收完成，后续由 Plan 222A-D 承接。  
  - `../archive/development-plans/HRMS-DOCUMENTATION-INDEX.md` —— HRMS 文档索引完成导读合流后转入归档。
- **2025-11-22**  
  - `../archive/development-plans/270-workflow-contract-guardian-remediation.md` —— Workflow 契约守卫恢复。  
  - `../archive/development-plans/05-CI-LOCAL-AUTOMATION-GUIDE.md` —— 自托管 Runner 指南停用，所有 CI 均使用 GitHub 平台 Runner。
- **2025-11-16**  
  - `../archive/development-plans/250-modular-monolith-merge.md`、`256-contract-ssot-generation-pipeline.md`、`256-signoff-20251116.md` —— 模块化单体合流与契约 SSoT 上线，drift 监控已启用。
- **2025-11-06**  
  - Plan 210/211/212/213/214 全量迁入归档；Go 1.24 基线、共享架构复用、schema 萃取指标全部完成。
- **2025-11-04 及更早**  
  - Plan 205 系列、Plan 07-20、60-66、84-110 等文档均保留在归档目录，仅供历史追踪与经验复用。

## Phase1 成果摘要
- Plan 210-214 均“✅ 完成”，整体评分 9.1/10；数据库迁移、Go 基线、架构复用、Schema 萃取覆盖率达到 95%+。
- 复用经验时请引用具体归档章节，禁止复制粘贴以免破坏唯一事实来源。

## 阅读顺序建议
1. `../architecture/01-organization-units-api-specification.md` 与 `docs/api/openapi.yaml`（契约基线）。  
2. `02-technical-architecture-design.md`（支撑架构）。  
3. 当前重点计划（`203` → `204` → `206` → `06`）。  
4. 对应归档记录，确认无冲突后再执行实现。

## 兼容性提醒
- 所有计划需遵守 Docker Compose 环境、Go 1.24 基线与 GraphQL/REST 契约；发生变更时同步更新 `docs/reference/` 中的事实来源。
- 若使用 `// TODO-TEMPORARY` 临时方案，必须在计划中注明截止日期、责任人，并接受 `scripts/check-temporary-tags.sh` 与 CI 守卫校验。
