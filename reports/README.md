# reports/ 运行报告治理指南

`reports/` 存放 actionlint、workflow、coverage、性能基线等 CI 报告。为减少 cloc 噪音并保持证据可追溯，遵循以下策略：

- **保留策略**：每类报告仅留最近 2 次（最新 + 上一次），更早的版本通过 `make archive-run-artifacts` 打包至 `archive/runtime-artifacts/<yyyy-mm>/reports/`，并在 manifest 中登记。
- **命名规范**：新增报告需使用 `reports/<domain>/<artifact>-<timestamp>.<ext>`，并在上游 Plan 文档或 README 中写明来源。
- **治理输出**：Plan 272 要求定期生成 `reports/plan272/*.csv|*.md`（如 `runtime-artifacts-inventory-*.csv`、`cloc-delta-*.md`），这些文件同时是守卫输入，严禁擅自删除。
- **守卫**：`npm run guard:plan272` 会校验报告数量和 README/manifest 是否存在，若超出保留份数需先执行归档再提交。

需要引用历史报告时，请查 `archive/runtime-artifacts/<yyyy-mm>/reports/` 或 GitHub Actions Artifact，并使用 manifest 中记录的 `sha256` 校验完整性。EOF
