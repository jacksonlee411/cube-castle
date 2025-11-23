# test-results/ 前端测试产物治理指南

`test-results/` 用于存放 Playwright/Vitest 的截图、trace、视频等证据，Plan 272 要求：

- **默认不追踪**：临时调试产物请保持在 `.gitignore` 外；只有“最新一次 PASS”且需写入 Runbook 的截图/trace 才能进入该目录。
- **归档要求**：一旦有新的 PASS 产物，需要执行 `make archive-run-artifacts` 将旧版本压缩到 `archive/runtime-artifacts/tests/<suite>/pass-<timestamp>.tar.zst`，并在 manifest 中注明 suites、浏览器、Run ID。
- **引用方式**：Runbook/计划文档只引用 README/manifest 或 `archive/runtime-artifacts` 中的压缩包，禁止直接链接历史未归档的截图。
- **守卫**：`plan272-artifact-guard` 会比对 README 中登记的产物列表与实际文件，发现遗漏或体量超标会阻塞提交。

若需调取历史截图，请按 README 指引从归档压缩包解出，并在审计记录中写明 `sha256` 校验值。EOF
