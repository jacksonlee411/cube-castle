# Plan 219C3 – 文档、测试与验收对齐

**文档编号**: 219C3  
**上级计划**: [219C – Audit & Validator 规则收敛](./219C-audit-validator.md)  
**目标周期**: Week 4 Day 24-25  
**负责人**: 后端团队 + 技术写作组  

---

## 1. 目标

1. 汇总 219C1/219C2 的实现结果，更新 README、参考文档、计划索引，并明确登记 REST 命令自测与 Job Level 验证补测，确保唯一事实来源一致。
2. 完成必要的单元/集成测试，覆盖审计与 validator 关键路径，同时补充 REST 命令自测与 Job Level Update/Version 验证，用于计划验收证据。
3. 在 `docs/development-plans/` 与 `docs/reference/` 之间更新引用关系，归档执行证据，满足审计可追溯性。

---

## 2. 范围

| 模块 | 内容 |
|---|---|
| 文档 | `internal/organization/README.md`、`docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`、`docs/reference/02-IMPLEMENTATION-INVENTORY.md`、`docs/development-plans/219C-audit-validator.md` 同步更新并互相引用；强调 GraphQL 仅承载查询，命令验证与补测统一走 REST。 |
| 测试 | `go test ./internal/organization/audit ./internal/organization/validator`，以及必要的 service 层回归测试；补测 Job Level Update/Version 请求验证与 REST 命令端到端脚本。 |
| 计划/归档 | 将阶段性结果追加到 `docs/development-plans/219C-audit-validator.md`，并准备归档（完成后移动至 `docs/archive/development-plans/`）。 |
| 验收表 | 输出勾选清单：审计操作全覆盖、validator 规则矩阵落地、README 更新完成。 |

---

## 3. 详细任务

### 3.1 文档同步
- [x] 更新 `internal/organization/README.md`：新增 `## 审计规范（219C1）` / `## Validators` 小节并登记 219C3 自测脚本（参见 `internal/organization/README.md:20-133`）。
- [x] `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 增加“审计执行检查”“REST 命令自测（219C3）”等条目，明确脚本与输出路径（参见 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md:112-184`）。
- [x] `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 同步记录 219C3 交付与 219C3A 修复，保持与 README/代码一致（参见 `docs/reference/02-IMPLEMENTATION-INVENTORY.md:73-87`）。
- [x] 在上述文档中补充 CQRS 边界提醒，强调命令验证统一走 REST（参见 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md:364-372`）。
- [x] `docs/development-plans/219C-audit-validator.md` 更新 219C3 进度与凭证链接，确保主计划可追溯（最新状态见 `docs/development-plans/219C-audit-validator.md:96-103`）。

### 3.2 测试与验证
- [x] 运行 `go test ./internal/organization/audit ./internal/organization/validator`，结果写入 `logs/219C3/validation.log:2893-2896`，作为验收凭证。
- [x] 使用 `scripts/219C3-rest-self-test.sh` 覆盖 Position/Assignment/Job Level 关键场景并复核返回（脚本流程见 `scripts/219C3-rest-self-test.sh:194-350`，执行产物见 `logs/219C3/validation.log:2628-2889` 与 `logs/219C3/report.json:2-38`）。
- [x] 通过 `verify_audit` 与 `psql` 输出确认审计字段与 README 描述一致（参考 `logs/219C3/validation.log:26-30`）。
- [x] 将脚本执行与测试命令的终端输出统一保存在 `logs/219C3/validation.log`（含 REST 场景与 go test 记录）。
- [x] Job Level `Update` / `CreateVersion` 请求校验补测：handler 层验证覆盖见 `internal/organization/handler/job_catalog_handler_test.go:43-205`，仓储版本逻辑验证见 `internal/organization/repository/job_catalog_repository_test.go:16-199`，相关执行输出包含在 `logs/219C3/validation.log:2768-2877`。

### 3.3 验收记录
- [x] 本文已追加验收章节，登记文档同步、REST 自测、Job Level 验证与审计校验的日志与请求 ID。
- [x] 计划归档副本已生成：`docs/archive/development-plans/219C3-20251106.md`，记录归档日期与证据指引。

---

## 4. 交付物

- 更新后的 README/参考文档。
- 完整的测试结果（CLI 输出或截图引用）。
- 219C 主计划的验收标记与回溯信息。

---

## 5. 风险与缓解

| 风险 | 影响 | 缓解 |
|---|---|---|
| 文档与实现不一致 | 高 | 以代码为准同步更新 README，必要时拉 Reviewer 联合检查。 |
| 测试覆盖不足 | 中 | 在 219C2 完成后统计未覆盖路径，纳入本阶段补测。 |
| 计划归档遗漏 | 低 | 在 PR 模板中添加检查项，确保归档操作可追踪。 |

---

## 6. 验收记录（2025-11-06）

- REST 自测脚本执行成功：`logs/219C3/validation.log:2628-2889` 展示 Position 填充/关闭与 Job Level 版本创建/冲突的返回体，`logs/219C3/report.json:2-38` 标记三大场景全部 `passed`。
- 审计字段核查：`logs/219C3/validation.log:26-29` 中的 `psql` 输出确认 `tenant_id`、`resource_type`、`request_id` 与 `correlation` 字段落盘。
- 单元测试回归：`logs/219C3/validation.log:2893-2896` 记录 `go test ./internal/organization/audit ./internal/organization/validator` 通过。
- Job Level 版本验证：`logs/219C3/validation.log:2802-2875` 展示 `jobLevel.version` 成功与冲突场景分别返回 `201` / `400 JOB_CATALOG_TEMPORAL_CONFLICT`。
