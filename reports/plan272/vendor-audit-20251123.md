# Plan 272 – Vendored 依赖评估（2025-11-23）

**范围**：`.github/actions/**/dist`（GitHub 官方 Actions 的 vendored 版本）与 `third_party/`（Go 模块源码镜像、工具依赖）。

## 1. GitHub Actions dist 评估

| Action | 当前引用方式 | 官方发布版是否可用 | 风险/理由 | 处置建议 |
| --- | --- | --- | --- | --- |
| `actions/checkout` (`.github/actions/checkout/dist/**`) | workflow 中通过 `uses: ./.github/actions/checkout` | ✅ v4 发布可直接 `uses: actions/checkout@v4` | 早期为兼容 WSL/self-hosted runner、可审计 dist；目前已切换 GitHub 托管 Runner，可考虑回迁 | 建议 Stage2 建立 PoC：在 `workflow-lint` 分支上改为官方引用，验证权限/代理设置后移除 dist；若遇阻需记录豁免与回收计划 |
| `actions/setup-node` | `uses: ./.github/actions/setup-node` | ✅ 官方 `actions/setup-node@v4` | 目前 vendored 版本包含内置缓存补丁；官方版本已支持缓存策略 | 与上同，先验证 `workflow-lint`/`agents-compliance` 任务在官方 action 下是否行为一致；结果写入下一版 `vendor-audit-*.md` |
| `actions/upload-artifact` | `uses: ./.github/actions/upload-artifact` | ✅ 官方 `v4` 可用 | Vendoring 主要用于离线/可审计；转用官方版本可减少 19 万行 JS | 计划与 `actions/download-artifact` 一并评估；若确认无额外 patch，Stage2 直接移除 dist |
| `actions/github-script` | `uses: ./.github/actions/github-script` | ✅ 官方 `v7` | 依赖 GH Token 权限，与 vendored 版本行为一致 | 建议直接切换至 `uses: actions/github-script@v7` 并删除 dist；需要在 `agents-compliance` 先做一次试跑 |
| `dorny/paths-filter` (`.github/actions/paths-filter/dist/index.js`) | 自定义引用 | ✅ 官方 `dorny/paths-filter@v3` | 自定义 dist 用于锁定 SHA；可改为 `@v3` 并在 workflow 固定 SHA | Stage2 评估是否需要 fork；若无 patch，回迁至官方 |

> 结论：`.github/` 下 vendored Actions 无自定义业务逻辑，仅为历史审计/离线可用。Stage2 默认目标是“改回官方发布版”，若实际存在差异（如特殊 patch）需在下一次 audit 中记录，并在 Plan 272 文档中附豁免。

## 2. `third_party/` 评估

| 模块/目录 | 用途 | 是否可替换为 go mod / npm 安装 | 风险/理由 | 处置建议 |
| --- | --- | --- | --- | --- |
| `third_party/github.com/99designs/gqlgen` 及相关 | GraphQL 代码生成器源码镜像 | ✅ 官方 tag 可通过 `go install github.com/99designs/gqlgen@<tag>` 获取；当前镜像以 pinned 方式存在 | 镜像用于避免 `go install` 过程中被墙或版本删除；本地/CI 已配备缓存 | Stage2 讨论是否可通过 `go toolchain` + `GOPROXY` 固定版本；如可行则移除本地镜像改用 `go install`，否则保留并将 hash 写入 README |
| `third_party/github.com/urfave/cli/v2` 等 | Go CLI 依赖镜像 | ✅（同上） | 同样出于可复现目的 | 推荐先对 `go.sum` 中引用的模块执行 `go env -w GOPROXY` 固定，然后逐步清理不再需要的 mirror |
| `third_party/tools/atlaslib` | 数据建模工具的内嵌网站 | ❌（需要可视化，官方发布获取复杂） | 包含演示网站与文档；CI 构建 rely on local copy | 保留，README 中说明“无法直接引用官方 release” |
| `third_party/opa` / `third_party/jsonschema`（若存在） | 静态工具/校验脚本 | ❓ | 需进一步扫描 `go.mod`、`package.json`，确认是否仍引用；未使用的可直接清理 | Stage2 行动项：运行 `node scripts/generate-implementation-inventory.js` + `go env GOPATH` 检查是否仍使用，否则删除 |

## 3. 后续行动
1. **PoC 切换**：在临时分支上把 `actions/checkout`, `setup-node`, `upload-artifact`, `github-script`, `paths-filter` 切换到官方引用，运行 `agents-compliance` 和 `workflow-lint`，记录结果。
2. **第三方镜像清单**：借助 `find third_party -maxdepth 2 -mindepth 2 -type d` + `go list -m all` 生成模块对照表，标记出“可移除”“必须保留”的条目。
3. **文档更新**：在 Plan 272 文档 W8/W9 小节添加“vendored 依赖决议”链接，并在 README 中写明如何获取官方 action 版本。
4. **治理成果包**：将本报告与 cloc 报告、守卫日志一并纳入 `reports/plan272/governance-kit-20251123.tar.zst`（下一步执行）。

> 本报告为 Stage2 起点，后续每次变更（切换/豁免）均需更新此文件或追加新版本（`vendor-audit-<ts>.md`），并在 OPS 会议记录中说明原因。
