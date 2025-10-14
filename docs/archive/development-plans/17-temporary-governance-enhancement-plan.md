# 17. TODO-TEMPORARY 治理与 419 状态码决策计划

**文档类型**: 治理/一致性强化计划  
**创建日期**: 2025-09-29  
**优先级**: P0（到期治理 + 契约一致性）  
**负责团队**: IIG 守护代理（Owner） / 认证小组（Co-owner） / 平台工程团队（CI 支持）  
**关联文档**: `docs/reference/04-AUTH-ERROR-CODES-AND-FLOWS.md`、`docs/development-plans/10-implementation-inventory-maintenance-report.md`、`scripts/check-temporary-tags.sh`

---

## 1. 背景与触发

1. `docs/reference/04-AUTH-ERROR-CODES-AND-FLOWS.md` 在“错误码与 HTTP 状态”段落保留 `// TODO-TEMPORARY`，约定 **2025-09-30 前**决定是否启用 419 状态码（区分会话过期）；当前实现仍统一返回 401。  
2. `scripts/check-temporary-tags.sh` 仅扫描代码目录并排除 `*.md` 文件，导致参考文档中的临时条目无法纳入自动治理。  
3. `docs/development-plans/10-implementation-inventory-maintenance-report.md` 已将代码层临时实现清零，但明确要求扩展 `check-temporary-tags` 并回收文档级 `TODO-TEMPORARY`。

**最高优先级原则**：依据 `CLAUDE.md` 与 `AGENTS.md`，需立即保证资源唯一性与临时实现按期回收，避免文档事实漂移。

---

## 2. 单一事实来源与一致性校验

- **契约事实**: 认证错误码语义以 `docs/api/openapi.yaml` 为准；`docs/reference/04-AUTH-ERROR-CODES-AND-FLOWS.md` 记录当前实现统一返回 401，并标记临时事项。  
- **治理脚本**: `scripts/check-temporary-tags.sh` 是唯一的 TODO-TEMPORARY 审核工具，但目前排除了 Markdown。  
- **计划约束**: `docs/development-plans/10-implementation-inventory-maintenance-report.md` 指定需将文档 TODO 纳入巡检，并维持 IIG 报表与实现一致。

本计划所有输出必须以以上事实来源为基线；任何变更先更新契约/脚本，再同步文档。

---

## 3. 问题定义

1. **419 状态码决策缺失**：临时条目已到期，若未裁定可能造成前后端行为偏差；若决定保留 401，则需移除 TODO 并记录一致性说明。  
2. **TODO 治理盲区**：`check-temporary-tags` 未覆盖文档目录，CI 无法阻止超期文档条目，违背“临时实现必须按期回收”要求。  
3. **巡检机制缺口**：缺少固定节奏（周频）校验结果的落地流程，IIG 守护需在计划中明确责任人与输出。

---

## 4. 目标与验收标准

| 目标 | 验收标准 |
| --- | --- |
| 决定 419 策略 | 产出决策记录（保留 401 或引入 419）；若引入 419，需先更新 `docs/api/openapi.yaml`、实现与前端处理；若不引入，移除 TODO 并在文档说明原因。 |
| 扩展 TODO 审核 | `scripts/check-temporary-tags.sh` 支持扫描 Markdown（含 docs/），输出中标记文档路径；新增白名单机制以避免历史归档反复触发。 |
| CI 集成 | 在现有 CI 或 Git Hook 中调用扩展脚本，失败时阻断合并；提交 `agents-compliance` 工作流变更或补充说明。 |
| 巡检与归档 | 建立周度巡检日志模板，保存于 `reports/iig-guardian/`，记录扫描结果、处理人、剩余 TODO；计划完成后将本文件移入 `docs/archive/development-plans/`。 |

---

## 5. 实施计划

| 阶段 | 目标 | 主要任务 | Owner | 截止 |
| --- | --- | --- | --- | --- |
| Phase 0 | 基线确认 | 复核 `docs/reference/04-AUTH-ERROR-CODES-AND-FLOWS.md` 与脚本输出，生成基线报告 `reports/iig-guardian/todo-temporary-baseline-20250929.md` | IIG 守护代理 | 2025-09-29 |
| Phase 1 | 419 决策 | 召开认证工作组评审会，出具决策结论；如保留 401，更新 Reference 文档并移除 TODO；如启用 419，同步 OpenAPI、命令服务实现与前端提示 | 认证小组 | 2025-09-30 |
| Phase 2 | 脚本扩展 | 修改 `scripts/check-temporary-tags.sh` 支持扫描 `docs/`、允许配置白名单文件，输出格式需包含 “文档/代码” 标记；补充自测记录 | 平台工程团队 | 2025-10-02 |
| Phase 3 | CI 集成 | 更新 `scripts/check-temporary-tags.sh` 调用链（`npm run lint` 前或 `agents-compliance` 工作流）；于 `reports/iig-guardian/todo-temporary-ci-verification-20251003.md` 附日志 | 平台工程团队 | 2025-10-03 |
| Phase 4 | 巡检机制 | 制定周度巡检模板，建立负责人轮值表；首次巡检报告提交到 `reports/iig-guardian/todo-temporary-digest-20251004.md` | IIG 守护代理 | 2025-10-04 |
| Phase 5 | 归档 | 验收各阶段交付物，更新实现清单统计（TS 导出数量如有变化），将本计划移至 `docs/archive/development-plans/` | IIG 守护代理 | 2025-10-05 |

---

## 6. 风险与缓解

| 风险 | 影响 | 对策 |
| --- | --- | --- |
| 419 决策延迟 | 超期 TODO 继续存在，破坏一致性 | 制定截至 2025-09-30 的评审会议纪要，若无法上线 419，立即确认继续使用 401 并记录原因。 |
| 脚本覆盖过宽导致噪音 | 大量历史文档被提示 | 为归档目录提供白名单配置，并在计划中说明豁免范围（仅限 `docs/archive/`）。 |
| CI 修改阻断其他工作 | 合并队列受影响 | 在单独分支验证脚本变更，附带回滚方案，确保 CI 失败可快速回退。 |
| 巡检流于形式 | 结果不落地 | 周报模板必须包含“违规项列表/整改人/ETA”，并在 `docs/development-plans/06-integrated-teams-progress-log.md` 例会中通报。 |

---

## 7. 验收清单（完成）

- ✅ `reports/iig-guardian/todo-temporary-baseline-20250929.md` 已提交并引用事实来源。  
- ✅ 认证评审结论发布，Reference 文档同步更新并无 TODO 残留（2025-09-29 决议继续使用 401）。  
- ✅ 扩展版 `scripts/check-temporary-tags.sh` 合入主干，自测脚本输出包含文档检查（新增白名单与文档输出标签）。  
- ✅ CI/工作流日志附于 `reports/iig-guardian/todo-temporary-ci-verification-20251003.md`，失败案例能阻断合并。  
- ✅ 周度巡检模板启用，首份周报完成并共享（模板：`reports/iig-guardian/todo-temporary-weekly-template.md`，首报：`reports/iig-guardian/todo-temporary-weekly-2025-09-23_2025-09-29.md`）。  
- ✅ 本计划归档并在 `docs/development-plans/06-integrated-teams-progress-log.md` 中标记完成。

---

**归档时间**: 2025-09-29  
**归档事由**: TODO 治理到期事项处理完成，CI/周报机制上线；证据与模板已归档于 `reports/iig-guardian/`。  
**后续提醒**: 持续执行周度巡检；若出现新的 `TODO-TEMPORARY` 条目需重新立项。

- 2025-10-14 登记：职位编制临时端点 `/api/v1/positions/{code}/fill|vacate|transfer` 已按 Stage1 实施计划添加 `// TODO-TEMPORARY`（Owner: 命令服务组，截止: 2025-11-15）。计划接入统一 assignments 模块后替换，届时需回收本地实现并在 IIG 周报中记录关闭。
