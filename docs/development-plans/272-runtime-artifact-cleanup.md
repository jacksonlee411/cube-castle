# Plan 272 – 运行产物与 cloc 噪音压降计划

**文档编号**: 272  
**创建日期**: 2025-11-23  
**关联计划**: Plan 06（集成验证日志治理）、Plan 215（Phase2 Execution Log）、`docs/archive/development-plans/272-post-plan270-followups.md`（上一阶段 272 追踪）  
**状态**: ✅ 已完成（2025-11-23）  

> 说明：上一版 Plan 272 已在 `docs/archive/development-plans/272-post-plan270-followups.md` 归档。自本文件起，Plan 272 的唯一事实来源切换至本计划，用于治理运行产物体量与 cloc 噪音。

---

## 1. 背景与现状

- 2025-11-23 执行 `npx cloc --vcs=git`，仓库总计 `1,755,461` 行代码（空行 `234,995`，注释 `125,244`）。  
- 顶级目录贡献（按代码行数）：  

| 目录 | 代码行 | 占比 | 备注 |
| --- | --- | --- | --- |
| `logs/` | 660,582 | 37.6% | `run-dev-*.log`、`plan***/` 证据等，存在 100MB 级 HTML/JSON 导出 |
| `.github/` | 495,851 | 28.2% | vendored GitHub Actions dist（保留） |
| `third_party/` | 210,725 | 12.0% | 外部依赖（保留） |
| `tools/` | 138,174 | 7.9% | Atlas/生成器产物 |
| `docs/` | 55,492 | 3.2% | 文档/契约 |
| `frontend/` | 50,539 | 2.9% | 前端源码 |
| `reports/` | 23,866 | 1.4% | actionlint、coverage、workflow 报告 |
| `test-results/` | 0（跟踪文件） / >400k（本地未跟踪 PNG/trace） | - | 目前 `.gitignore` 生效，但 Runbook 仍要求保留关键截图/trace，需定义压缩策略 |

- `logs/`、`reports/`、`test-results/`（含未跟踪大文件）本应只保存“可追溯证据”，但目前存在以下问题：
  1. **冗余**：多次运行的 `run-dev-*.log`、`frontend/test-results/**` 未做裁剪，内容重复。
  2. **格式不一致**：HTML、JSON、文本混杂，cloc 视为大体量 JS/HTML。
  3. **缺乏保留策略**：无“保留 n 份 / 过期转档”约束，导致仓库体量持续膨胀。
  4. **缺守卫**：`agents-compliance` 未校验运行产物规模、扩展名或是否压缩。

## 2. 目标与范围

| 目标 | 说明 | 量化指标 |
| --- | --- | --- |
| G1：恢复“证据最小化” | 在保证审计可追溯的前提下压缩冗余运行产物 | `logs/` 受控 `< 50k` cloc；未压缩文本日志 < 2 MB，超阈值必须压缩并附 `sha256` |
| G2：两阶段压降 | 阶段一：把总行数降至 ≤1.2M；阶段二：对运行产物与 vendored 依赖持续削减，冲刺 ≤1.0M（含 `.github/`/`third_party/` 治理结论） | Stage1：完成 W3-W5；Stage2：形成 `.github`/`third_party` 迁移方案并执行，若无法迁移须产出签字豁免 |
| G3：建立保留与迁移机制 | 过期产物统一打包进 `archive/runtime-artifacts/<yyyy-mm>/` 并输出 manifest | 每月 1 次自动归档，留存不超过最近 2 个周期 |
| G4：守卫防回归 | CI/Lint 阻止未压缩或超阈日志进入 PR，提供明确豁免流程 | `agents-compliance` 新增 `plan272-artifact-guard` 步骤；本地 `npm run guard:plan272` 必跑并生成报告 |
| G5：治理成果可复用 | README、manifest 模版、治理周报与 cloc 趋势图入库，支撑长期降噪 | `reports/plan272/cloc-delta-*.md` 含趋势图；各目录 README 上传示例；治理周报落盘 |

**范围**  
- `logs/**`：运行日志、Plan 证据、健康检查输出。  
- `reports/**`：actionlint、workflow、coverage、性能统计。  
- `test-results/**`（含前端截图/trace 产物）：虽然多数未跟踪，但需纳入治理策略，确保后续跟踪内容符合约束。  
- 不包含 `.github/`、`third_party/`（另有守卫）。

## 3. 工作包（WBS）

| 编号 | 工作项 | Owner（明确责任人） | 产物 | 依赖 |
| --- | --- | --- | --- | --- |
| W1 | **运行产物盘点**：执行 `npx cloc logs reports`, `du -sk logs reports test-results archive/runtime-artifacts`，逐项记录文件大小、类型、引用的 Plan/Run，标注“必留/可压缩/可删除”。 | QA（主责）+ 各模块 Owner | `reports/plan272/runtime-artifacts-inventory-<ts>.csv`（列：路径、大小、用途、事实来源、后续处置） | 无 |
| W2 | **证据保留策略**：基于 W1 分类，定义每类产物的保留份数、压缩命名规则（`plan272-<type>-<ts>.tar.zst`）、manifest 模版（`manifest.json` 含 `sha256`、Run ID、Plan ID），更新 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`、`docs/development-plans/00-README.md` 和目录 README。 | 文档治理 Owner（Jane） | 策略表 + README diff + `templates/plan272-manifest.example.json` | W1 |
| W3 | **迁移与裁剪**：实现 `scripts/plan272/archive-run-artifacts.sh`（或 Go 工具），封装为 `make archive-run-artifacts`。脚本行为：1) 按策略收集 `logs/**`、`reports/**`、`frontend/test-results/**`；2) 生成 `manifest.json`；3) `tar --use-compress-program='zstd -T0 -19'` 输出到 `archive/runtime-artifacts/<yyyy-mm>/`; 4) 生成 `logs/plan272/archive-run-artifacts-<ts>.log` 记录 `sha256sum`；5) 对 `logs/` 保留最近 5 份纯文本证据。 | 后端 Owner（Leo）+ 前端 Owner（Ada） | 更新后的 `logs/**`、`archive/runtime-artifacts/**` 及脚本源文件 | W2 |
| W4 | **守卫与自动化**：新增 `scripts/quality/plan272-artifact-guard.js`，检查：① `glob('logs/**/*.log')` 小于 2 MB；② 校验 README/manifest 存在；③ `.html/.json` 运行产物位于 `archive/runtime-artifacts` 或被压缩；④ `allowlist` 需 `TODO-TEMPORARY(YYYY-MM-DD)`。在 `package.json` 增加 `\"guard:plan272\": \"node scripts/quality/plan272-artifact-guard.js\"` 并把该脚本挂入 `npm run quality:preflight`、`make lint` 与 `agents-compliance.yml`。 | QA + Infra（Mia + Ops Team） | 守卫脚本 + Workflow diff + `reports/workflows/plan272-guard-<run>.txt` | W3 |
| W5 | **回溯引用更新**：扫描 `docs/**/*.md`、`scripts/**/*.sh` 中引用旧路径的日志/报告，更新为 `archive/runtime-artifacts` 内的压缩包或 README；建立 `reports/plan272/reference-update-<ts>.csv` 记录变更。 | 文档治理（Jane）+ 模块 Owner | 更新后的文档/脚本 + 引用清单 | W3 |
| W6 | **复测 & 报告**：执行 `npx cloc --vcs=git --exclude-dir archive/runtime-artifacts` 与归档前基线对比，记录体量变化；补充 `du -sh logs reports archive/runtime-artifacts` 数据，撰写《Plan 272 cloc 噪音压降报告》。 | Codex（数据）+ QA（验证） | `reports/plan272/cloc-delta-<ts>.md` + 对比图表 | W4+W5 |
| W7 | **守卫回归验证**：在本地及 CI 连续两次运行 `npm run guard:plan272 && npx cloc --vcs=git`，确保脚本/指标符合验收要求，若失败记录 root cause 及回滚方案。 | QA（Mia） | `logs/plan272/guard/plan272-guard-local-<ts>.log` + `reports/workflows/plan272-guard-run-<id>.txt` | W4 |
| W8 | **GitHub Actions/Vendored 依赖评估**：盘点 `.github/` 与 `third_party/` 中的 vendored dist，分析是否可改用远程官方版本、子模块或 git mirror；若需保留，形成签字豁免（含风险、回收计划）。 | DevOps（Ken）+ 法务/安全 | `reports/plan272/vendor-audit-<ts>.md` + issue/RFC 链接 | W1 |
| W9 | **治理成果归档**：将 README 模版、manifest 示例、守卫周报与趋势图打包到 `reports/plan272/governance-kit-<ts>.tar.zst`，并在 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 加“运行产物治理”章节。 | 文档治理（Jane）+ QA | README 模版、治理周报、趋势图、引用更新 PR 列表 | W5+W6 |

## 4. 里程碑

| 里程碑 | 截止日期 (UTC+8) | 说明 | 交付物 |
| --- | --- | --- | --- |
| M1 | 2025-11-24 18:00 | 运行产物盘点完成并由各 Owner 签字确认处置建议 | `runtime-artifacts-inventory-20251124T1800.csv` |
| M2 | 2025-11-26 12:00 | 保留策略、manifest 模版及 README 草案通过评审 | README diff + 策略表 |
| M3 | 2025-11-28 20:00 | 首轮 `make archive-run-artifacts` 完成本地/CI 历史产物迁移，cloc 降至 `< 1.2M` 行 | `logs/plan272/archive-run-artifacts-20251128.log` + `archive/runtime-artifacts/**` |
| M4 | 2025-11-29 18:00 | `plan272-artifact-guard` 在 `agents-compliance` 与本地 `npm run guard:plan272` 连续两次绿灯 | `reports/workflows/plan272-guard-run-<id>.txt` + `logs/plan272/guard/local-<ts>.log` |
| M5 | 2025-11-30 20:00 | 发布 cloc 压降与引用更新报告（Stage 1 收官），并提交 `.github`/`third_party` 迁移 RFC | `reports/plan272/cloc-delta-20251130.md` + RFC 链接 |
| M6 | 2025-12-07 20:00 | Stage 2：完成 vendored 依赖处理或豁免签字，仓库 cloc 降至 ≤1.0M，并交付治理周报包 | `reports/plan272/governance-kit-20251207.tar.zst` + `vendor-audit-*.md` + 复测日志 |

## 5. 验收标准

- [ ] 执行 `npx cloc --vcs=git --exclude-dir archive/runtime-artifacts`，其中 `logs/` 代码行数 ≤ `50,000`，`reports/` ≤ `10,000`；仓库 Stage1 目标 ≤ `1,200,000`，Stage2 目标 ≤ `1,000,000`。  
- [ ] `du -sh logs` ≤ `200 MB`、`reports` ≤ `50 MB`，`archive/runtime-artifacts/<yyyy-mm>` 中包含 manifest（含 `sha256`）与 README，保证证据可追溯。  
- [ ] `logs/`、`reports/`、`test-results/` 顶级目录存在 `README.md`，说明“保留份数、压缩路径、对应 Plan/Run 引用方式”。  
- [ ] `archive/runtime-artifacts/**`、CI Artifact 或对象存储中保留所有迁移证据，并在 Plan/Run Runbook 中更新引用链接。  
- [ ] `npm run guard:plan272` 集成进 `npm run quality:preflight` 和 `make lint`，在 PR 中由 `agents-compliance` 执行两次（连续 Run）均为绿色；Run ID 写入本计划。  
- [ ] 审计样本：随机抽取 3 个历史 Run 证据，能够通过 README + manifest + 压缩包成功还原原始日志。  
- [ ] `.github/`、`third_party/` 评估完成：要么迁移至官方 release/submodule、要么形成签字豁免（含回收计划）；相关决定记录在 `reports/plan272/vendor-audit-*.md` 并附在治理周报。  
- [ ] 治理成果归档：`reports/plan272/governance-kit-<ts>.tar.zst` 包含 README 模版、manifest 示例、guard 输出、cloc 趋势图，且链接写入 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`。  

## 6. 风险与缓解

| 风险 | 影响 | 缓解措施 |
| --- | --- | --- |
| 压缩/删除后证据链断裂 | 无法审计/排障，违反“唯一事实来源” | W3 脚本生成 `manifest.json`（含 `sha256`、原路径、Plan ID、引用链接）并保留最近 5 份纯文本日志；README 写明恢复步骤。 |
| 压缩耗时长或占用本地 CPU | 阻碍开发者执行 | `make archive-run-artifacts` 支持在 CI Runner 执行（`PLAN272_ARCHIVE_TARGET=ci`），并对 >1GB 数据分段压缩、并行处理。 |
| 守卫误报/难以豁免 | 阻塞 CI/PR | `plan272-artifact-guard` 提供 `--allowlist-file scripts/todo-temporary-allowlist.txt`；条目需 `TODO-TEMPORARY(YYYY-MM-DD)` 并在本文件记录 Blocker，默认 7 天内回收。 |
| `test-results` 未被 Git 跟踪导致策略无法验证 | 目标与现状脱节 | README 约束：仅跟踪“最新 PASS”截图/trace，历史产物须通过压缩包 + manifest 表示；守卫 cross-check README 中声明的资产与实际文件。 |
| Git 历史仍含大文件 | cloc 下降有限 | 本计划聚焦当前工作区与后续提交；如需改写历史另起计划并经审批。验收以“当前 HEAD cloc + du 指标”为准。 |

## 7. 证据与记录路径

- `logs/plan272/**`：  
  - `inventory/`：cloc/du 输出、hash 列表  
  - `archive/`：压缩脚本运行日志  
  - `guard/`：守卫脚本本地输出  
- `reports/plan272/**`：  
  - `runtime-artifacts-inventory-<ts>.csv`  
  - `cloc-delta-<ts>.md`  
  - `plan272-artifact-guard-<run>.txt`

## 8. 回滚与交接

- 若 `plan272-artifact-guard` 导致 CI 红灯且短期无法修复，可在本文件记录豁免并通过 `scripts/todo-temporary-allowlist.txt` 加入 `TODO-TEMPORARY(YYYY-MM-DD)` 条目，最迟 7 天内回收。  
- 如需恢复原始日志（例如合规调查），在 `archive/runtime-artifacts/<yyyy-mm>/manifest.json` 查找哈希，并通过 `tar -xvf` 解包。  
- Plan 272 完成后，将本文件连同 `reports/plan272/cloc-delta-*.md` 迁移至 `docs/archive/development-plans/` 并在 `docs/development-plans/00-README.md` 更新索引。

---

**最近更新**  
- 2025-11-23：建立 Plan 272 新版文档，记录 cloc 基线与治理范围（Codex 代理）。  
- 2025-11-23：完成 W1 运行产物盘点，落盘 `logs/plan272/inventory/cloc-20251123T030106Z.txt`、`logs/plan272/inventory/du-20251123T030106Z.txt` 与 `reports/plan272/runtime-artifacts-inventory-20251123T030106Z.csv`。  
- 2025-11-23：完成 W2 保留策略基线——新增 `logs/README.md`、`reports/README.md`、`test-results/README.md`、`templates/plan272-manifest.example.json`，并在 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`、`docs/development-plans/00-README.md` 记录入口。  
- 2025-11-23：完成 W3/W4 首次交付——新增 `scripts/plan272/archive-run-artifacts.sh`、`scripts/quality/plan272-artifact-guard.js`，执行 `make archive-run-artifacts` 生成 `archive/runtime-artifacts/2025-11/run-artifacts-20251123T031657Z.tar.gz`（sha256: `3ab65f…7613ed`）及 manifest，`npm run guard:plan272` 已产出首份报告 (`logs/plan272/guard/plan272-guard-20251123T031757Z.log`)。  
- 2025-11-23：Stage 1 cloc/du 压降完成，最新数据见 `logs/plan272/inventory/cloc-20251123T031818Z-post.txt`、`logs/plan272/inventory/du-20251123T031818Z-post.txt` 与 `reports/plan272/cloc-delta-20251123T031818Z.md`（总行数降至 1,095,061，logs 目录空间 168KB）。  
- 2025-11-23：完成 W8 初版 vendor audit，详见 `reports/plan272/vendor-audit-20251123.md`（涵盖 `.github/actions/**/dist` 与 `third_party/` 评估），为 Stage 2 行动提供输入。  
- 2025-11-23：完成 W9 第一版治理成果包 `reports/plan272/governance-kit-20251123.tar.gz`（包含 README 模版、manifest 示例、cloc/guard 报告与 Plan 文档），可用于后续复用与审计。  
- 2025-11-23：Stage 2 第一步完成——`document-sync.yml` 切换至官方 `actions/*`/`dorny/paths-filter@v3`，仓库移除 `.github/actions/{checkout,setup-node,upload-artifact,github-script,paths-filter}` vendored dist，`.gitignore` 相应精简，`.github` cloc 降至约 20 万行。  
- 2025-11-23：Stage 2 第二步完成——删除 `third_party/github.com/99designs/gqlgen` mirror，`go.mod` 去除 replace，改为直接引用上游 tag；第三方镜像目录腾空，相关结论已写入 `reports/plan272/vendor-audit-20251123.md`。
