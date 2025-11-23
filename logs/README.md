# logs/ 运行产物治理指南

本目录仅保存运行日志与计划执行证据（Plan 06/215/242 等），自 Plan 272 起执行“证据最小化”策略：

- **保留份数**：仅存最近 5 份纯文本日志（含 `run-dev-*.log`、`plan***/` 关键证据），其余必须通过 `make archive-run-artifacts` 压缩并迁移至 `archive/runtime-artifacts/<yyyy-mm>/`。
- **压缩与 manifest**：归档脚本会生成 `manifest.json`（含 `sha256`、Plan ID、源路径），供 Runbook/审计引用；manifest 模版见 `templates/plan272-manifest.example.json`。
- **命名与结构**：所有新证据按 `logs/plan272/<type>/<timestamp>.<ext>` 分类存放，并在 `docs/development-plans/272-runtime-artifact-cleanup.md` 中记录来源。
- **守卫要求**：提交前需运行 `npm run guard:plan272`，超过 2 MB 的纯文本日志或缺少 README/manifest 会触发失败；CI 的 `plan272-artifact-guard` 使用同一规则。

需要查询历史日志时，请先查阅 `archive/runtime-artifacts/` 或 CI Artifact，并按照 manifest 中的 `sha256` 进行校验。严禁在未运行归档脚本的情况下直接删除或移动本目录文件。EOF
