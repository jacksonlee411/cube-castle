# Plan 268 - 本地化 GitHub Actions 依赖与门禁缓存治理

**文档编号**: 268  
**创建日期**: 2025-11-19  
**关联计划**: Plan 262/265（自托管 Runner）、Plan 266（执行追踪）  
**状态**: ⚠️ 搁置（2025-11-20）——自托管 runner 网络未稳定前，Actions vendoring 无法完整验证，待 Plan 265/267 恢复后再继续。

---

## 1. 背景与目标

- 自托管 Runner 每次执行 `document-sync`/`api-compliance` 等门禁时，都会重新从 GitHub 下载 `actions/checkout`、`actions/setup-node`、`actions/upload-artifact` 等依赖。日志显示重复的 `Download action repository ...`，甚至因网络抖动出现 `HttpClient.Timeout`、`GnuTLS recv error (-110)`，延长排队时间并造成失败。
- Plan 266/267 当前已聚焦网络与 TLS 稳定性，但即便网络恢复良好，频繁拉取 actions tarball 仍会耗费时间与带宽，同时与“Docker 强制 + 资源唯一性”原则相违背（workflow 的事实来源分散在 GitHub 官方仓库）。
- Plan 268 的目标：将所有门禁依赖的第三方 action 固化在仓库 `.github/actions/` 下，并建立更新脚本、版本索引与 workflow 规范，确保自托管 runner 只需读取本仓库内容即可完成门禁，减少外部依赖。

---

## 2. 范围与交付物

| 项目 | 交付物 | 说明 |
|------|--------|------|
| Actions 盘点 | `docs/reference/actions-inventory.md` | 列出当前 workflow 使用的所有 `uses:` 引用（仓库内/外部），标注来源、版本、负责团队 |
| 本地镜像 | `.github/actions/<name>/` | 按 action 名称落库，包含原始源码（tarball 解包）与 `VERSION`（记录 `repo` + `commit`） |
| 更新脚本 | `scripts/ci/actions/vendor-action.sh` | 输入 repo/name 与 commit，自动下载、校验 `sha256`、写入 `VERSION`，供未来升级使用 |
| Workflow 调整 | `.github/workflows/*.yml` | 将引用切换为相对路径（例如 `./.github/actions/checkout`），并在需要时通过 `with:` 传参保持行为不变 |
| 治理文档 | `docs/reference/github-actions-vendoring-guide.md` | 描述引入/升级流程、合规要求、如何验证 vendored action、回滚策略 |

不在本计划范围：创建新的自定义 action、修改上游 action 功能、在 `.github/actions` 下执行构建（需保持上游产物原样）。

---

## 3. 实施步骤

### 3.1 Actions 盘点
1. 编写脚本 `scripts/ci/actions/list-workflow-actions.js`（或使用 `rg`) 扫描 `.github/workflows/**/*.yml` 的 `uses:` 字段，列出所有外部依赖。
2. 输出 `docs/reference/actions-inventory.md`，包含 action 名称、版本/commit、使用的 workflow、是否已本地化。
3. 在 Plan 265/266 文档中引用本计划，说明门禁依赖被集中治理。

### 3.2 初始 vendoring（document-sync pilot）
1. 对 `document-sync.yml` 使用的 `actions/checkout@v4`、`dorny/paths-filter@v3`、`actions/setup-node@v4`、`actions/upload-artifact@v4`、`actions/github-script@v7`，通过 `vendor-action.sh` 下载到 `.github/actions/<name>/` 并记录 `VERSION`（已在当前会话完成，Plan 268 需将流程固化）。
2. 修改 workflow `uses:` 为相对路径，并验证 `workflow_dispatch` 在 ubuntu/selfhosted 均可成功。
3. 在 `docs/development-plans/266-selfhosted-tracking.md` 中补充“Actions vendoring”进展。

### 3.3 扩展至所有 workflow
1. 根据 3.1 的 inventory 逐个处理剩余 workflow（`api-compliance.yml`、`consistency-guard.yml`、`iig-guardian.yml`、`agents-compliance.yml` 等），优先级：自托管 job > 必跑门禁 > 其它。
2. 对于多 workflow 复用的 action，应共用同一目录（如 `.github/actions/checkout`），避免重复。
3. 更新完成后，运行关键 workflow 的 `workflow_dispatch` 以确认行为一致。

### 3.4 工具与治理
1. **vendor-action.sh**：支持参数 `--repo owner/name --ref <commit/tag> --dest .github/actions/<name>`，自动：
   - 下载 tarball → `sha256sum` 验证（可选）
   - 解包到目标目录（清空旧文件）
   - 写入/更新 `VERSION`
   - 生成 `NOTICE`（记录来源许可证）
2. **版本锁定**：在 `docs/reference/github-actions-vendoring-guide.md` 中规定：
   - 升级 action 必须运行 `vendor-action.sh`
   - 提交 PR 时在描述中注明 action 版本变化
   - 附带测试或 workflow 运行截图
3. **自动校验**：新增脚本 `scripts/ci/actions/validate-vendoring.sh`：
   - 扫描 `.github/workflows`，若 `uses:` 指向 GitHub 官方 action（`actions/*`, `dorny/*`, etc.）但没有对应 `.github/actions/<name>`，即失败
   - 检查 `.github/actions/*/VERSION` 是否存在
   - 在 `agents-compliance.yml` 或 `document-sync.yml` 中加入该脚本，确保 CI 阶段阻止遗漏

### 3.5 文档与交接
1. `docs/reference/github-actions-vendoring-guide.md` 包含：
   - 目的与收益
   - 如何新增 action
   - vendor-action.sh 使用示例
   - 常见问题（如 action 依赖 npm install、dist/ 文件夹等）
2. 在 `docs/reference/05-CI-LOCAL-AUTOMATION-GUIDE.md` 增加“Action vendoring”章节，指向指南与脚本。

---

## 4. 验收标准

- [ ] `docs/reference/actions-inventory.md` 发布，列出所有 workflow 的 action 依赖及其本地化状态。
- [ ] `.github/workflows/document-sync.yml`、`api-compliance.yml`、`consistency-guard.yml`、`iig-guardian.yml` 等必跑门禁均使用本地 action，`git grep 'uses: actions/' .github/workflows` 结果为空（除非 action 已在 `.github/actions` 下）。
- [ ] `scripts/ci/actions/vendor-action.sh` 能自动下载/更新 action，`VERSION` 记录完整，`scripts/ci/actions/validate-vendoring.sh` 在 CI 中执行通过。
- [ ] 至少一次关键 workflow（document-sync 自托管 + ubuntu）使用本地 action 成功运行，Run ID 记录于 Plan 265。
- [ ] `docs/reference/github-actions-vendoring-guide.md` 完成，AGENTS.md/CI 指南中已指向该文档。

---

## 5. 风险与回滚

| 风险 | 描述 | 缓解/回滚 |
|------|------|-----------|
| action 升级遗漏 | 新版本发布后未及时同步，导致安全修复缺失 | 在 `actions-inventory.md` 记录负责人与跟新周期，并在 `scripts/ci/actions/list-workflow-actions.js` 中检测版本差异 |
| vendored 代码受损 | 手动修改导致与上游不一致 | 禁止直接编辑 `.github/actions/<name>`；如需修复，重新运行 `vendor-action.sh` 并覆盖 |
| 仓库体积增加 | 复制多个 action 导致提交体积上升 | 仅保留必要 action，定期清理未使用目录，并在 `.gitignore` 中避免缓存大文件 |
| 合规问题 | Action 许可证与仓库不兼容 | `vendor-action.sh` 获取 `LICENSE`，在 `actions-inventory.md` 中记录并由法务/安全确认 |

回滚：若某 action 本地化后出现问题，可在 workflow 中临时改回远端 `uses: owner/action@ref`，同时在 Plan 268 中登记原因与回滚窗口，修复后再改回本地路径。

---

## 6. 里程碑

- **M1（2025-11-20）**：`document-sync` 使用本地 action 运行成功，`actions-inventory.md` 首版完成。
- **M2（2025-11-22）**：所有 Required workflow 完成本地化；`vendor-action.sh`/`validate-vendoring.sh` 入库。
- **M3（2025-11-25）**：CI/文档更新完成，Plan 265 记录三大 workflow 的自托管 run ID，Plan 268 转入维护。

---

## 7. 参考资料

- AGENTS.md：资源唯一性、Docker 强制
- Plan 265/266：自托管 Runner 目标与运行记录
- `docs/reference/05-CI-LOCAL-AUTOMATION-GUIDE.md`：门禁+本地开发指南
- GitHub 官方文档：Reusable actions, Composite actions
