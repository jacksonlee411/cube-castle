# Plan 272 – Vendored 依赖评估（2025-11-23）

**范围**：`.github/actions/**/dist`（GitHub 官方 Actions 的 vendored 版本）与 `third_party/`（Go 模块源码镜像、工具依赖）。

## 1. GitHub Actions dist 评估

- 2025-11-23：`document-sync.yml` 已全部切换至 `dorny/paths-filter@v3`、`actions/setup-node@v4`、`actions/upload-artifact@v4`、`actions/github-script@v7`，仓库中的 `.github/actions/{checkout,setup-node,upload-artifact,github-script,paths-filter}` 目录已删除。  
- 当前 `git grep 'uses: ./.github/actions' .github/workflows` 为空，说明没有 workflow 再依赖 vendored dist。若后续 workflow 新增第三方 Action，需遵循“直接引用官方 release + 固定版本”的原则；如确因监管要求需要 vendoring，应在 Plan 272 文档中立项并附豁免。

## 2. `third_party/` 评估

- 2025-11-23：清理前仅存在 `third_party/github.com/99designs/gqlgen` mirror；现已删除该目录，并在 `go.mod` 中移除 replace，构建时直接通过 `go install github.com/99designs/gqlgen@v0.17.45` 下载官方 release（CI 已具备网络与 GOPROXY）。  
- 当前 `third_party/` 目录为空；如后续必须引入镜像，需说明“无法从官方获取”的原因、引用 commit hash、回收计划，并在 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 登记。

## 3. 后续行动
1. **定期复核**：每次新增/调整 workflow 时必须运行 `git grep 'uses: ./.github/actions'`，若发现 vendoring 征兆立即在 Plan 272 文档记录原因与时限。
2. **镜像引入守卫**：在 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 中保留“运行产物治理”段落，要求新增 third_party 镜像前先完成 RFC & 签字。
3. **治理成果更新**：每次压降或豁免后，更新 `reports/plan272/vendor-audit-<ts>.md` 与 `reports/plan272/governance-kit-<ts>.tar.gz`，确保审计材料可追溯。

> 本报告将随 Stage 2 进展持续更新：若未来确需重新 vendoring，务必在此文件中补充“镜像来源、原因、回收计划”，并在 Plan 272 文档同步。
