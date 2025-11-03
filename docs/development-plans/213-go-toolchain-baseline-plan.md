# 213-Go Toolchain Baseline Alignment Plan

**编号**: 213  
**标题**: Go 1.24 基线评审与回退预案  
**创建日期**: 2025-11-03  
**最近更新**: 2025-11-04  
**状态**: 🟢 已完成（Steering 确认 Go 1.24 基线）  
**关联文档**:
- `211-phase1-module-unification-plan.md`（现阶段默认 Go 1.24.0）
- `06-integrated-teams-progress-log.md`（管理层风险记录）
- `204-HRMS-Implementation-Roadmap.md`（阶段路线图）
- `PLAN200 系列`（若需回落 1.22.x 将触发的变更流程）

---

## 1. 背景与目标

- **背景**：Plan 211 执行过程中，根模块已统一至 `go 1.24.0`，满足 `pgx/v5` 等依赖最低要求；但 Steering 尚未正式确认是否将 1.24 作为项目长期基线。
- **目标**：评估 Go 1.24 的兼容性、CI/部署影响以及回退成本，向 Steering 提交决议材料；若需回落 1.22.x，准备 PLAN200 变更包。

---

## 2. 关键交付物

| 编号 | 交付物 | 内容 | 责任人 | 归档位置 |
|------|--------|------|--------|----------|
| D1 | 基线评审报告 | 依赖清单、CI 验证、风险评估、推荐结论 | Codex + DevOps | `reports/phase1-regression.md` 补充章节 |
| D2 | Steering 决议摘要 | 会议纪要或邮件决议 | PM | `docs/development-plans/06-integrated-teams-progress-log.md` |
| D3 | 回退预案（如需） | PLAN200 变更概要、时间窗口、验收项 | Codex | `docs/development-plans/213-go-toolchain-baseline-plan.md` 附录 |

---

## 3. 工作分解

| 阶段 | 任务 | 描述 | 输出 |
|------|------|------|------|
| 评估准备 | 依赖盘点 | 列出强制要求 Go ≥1.24 的依赖（pgx、golang.org/x/* 等） | 依赖表 |
| 评估执行 | CI/生产验证 | 确认 CI、docker build、运行时日志等在 1.24 下无回归 | 测试记录（`reports/phase1-acceptance-summary-*.md`） |
| 决议材料 | 风险总结 | 收集第三方库兼容性、工具链支持、团队开发环境现状 | 评审报告草稿 |
| Steering 审议 | 决策会议 | 与 Steering 确认是否采纳 1.24 或回落 | 纪要 |
| 回退预案 | 若需回落 | 准备 PLAN200 变更范围、回退步骤与风险提示 | 预案草稿 |

---

## 4. 时间线（建议）

| 日期 | 行动 | 负责人 |
|------|------|--------|
| Day6 | 完成依赖与 CI 验证收集 | Codex + DevOps |
| Day7 | 形成评审草稿并同步 Steering 会前材料 | Codex |
| Day8 | Steering 审议并给出决议 | PM（主持） |
| Day9 | 如需回落，启动 PLAN200 流程；否则更新文档与 go.mod 注释 | Codex |

---

## 5. 风险与应对

| 风险 | 等级 | 说明 | 应对 |
|------|------|------|------|
| 第三方库兼容性未充分验证 | 中 | 某些依赖在 1.24 下行为变化 | 补充 `go test ./... -count=1`、`go test -race`、必要集成测试 |
| 开发环境未升级 | 中 | 团队成员本地 Go 版本低于 1.24 | 发布升级指引，保留 1.22 编译说明 |
| 回落操作影响 Plan 211 进度 | 中 | 回退需大规模替换依赖 | 预备 PLAN200 草案，明确时间窗口与责任人 |

---

## 6. 结论记录（待填）

| 项目 | 状态 | 说明 |
|------|------|------|
| 推荐方案 | ✅ 采纳 Go 1.24.0（toolchain go1.24.9），保留向下兼容测试 | 依赖盘点显示无第三方强制要求 1.24+；`go test ./... -count=1`、`go env GOVERSION` 验证通过 |
| 回退计划 | 暂不启用 PLAN200；若未来依赖回退需由 Steering 重新立项 | 保留原 1.22 工具链说明，必要时恢复 `toolchain` 指令并执行依赖调降 |
| 文档更新 | ✅ `reports/phase1-regression.md` 新增“Go 1.24 基线评审”章节；`docs/development-plans/06-integrated-teams-progress-log.md` 补充状态 | 同步完成 |

---

**版本历史**  
- v1.0 (2025-11-03)：建立评审计划，等待 Steering 决策。  
- v1.1 (2025-11-04)：完成评审，确认保持 Go 1.24 基线，无须回退预案。
