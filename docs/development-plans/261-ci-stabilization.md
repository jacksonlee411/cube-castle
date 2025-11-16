# Plan 261 — GitHub Actions 稳定化与降噪（临时措施回收期至 2025-11-22）

文档编号: 261  
负责人: Codex（AI）  
目标: 快速降低远程 CI 红灯，提升诊断可读性与通过率；与 AGENTS.md 原则一致（Docker Compose 强制、SSoT），并为 11-22 前的全面回收打基础。

---

## 1. 背景与问题
- 近 50 次运行：31 failure / 19 success，失败集中在重复代码检测（jscpd 扫描 node_modules/产物导致超阈值）、E2E 等待不稳定、前端质量门禁过严、文档同步/一致性检查在历史/索引下的误报。
- 已启用“docs/ci-only 短路”（临时），但部分工作流仍未覆盖或证据采集薄弱，导致红灯定位效率低。

## 2. 变更范围（本次补丁）
- 重复代码检测（duplicate-code-detection.yml）
  - 为 jscpd 增加忽略清单：node_modules、dist/build/.cache/coverage/logs/reports、*.min.js 等；降低误报，保留 PR 阶段阻断。
  - 默认扫描范围仍保留；后续以“基线+缓冲”（每周 -1%）收紧阈值（由研发负责人设定）。
- 前端 E2E（frontend-e2e.yml、e2e-smoke.yml）
  - 启动后端改为 docker compose up -d --wait（利用 healthcheck）；保留 curl 兜底并延长等待时长。
  - 失败时强制落盘 docker compose logs 并作为工件上传，便于定位。
- 文档同步（document-sync.yml）
  - 补充 docs/ci-only 短路（临时，11-22 回收）：文档/工作流类变更直接快速通过，避免在文档类 PR 中引入重门禁阻断。

以上变更均为“临时稳态”优化，不改变业务口径与契约；待 232/252/255/256 达标后统一回收临时短路。

## 3. 验收标准
- 连续 3 次主干 push：上述工作流不再因第三方/构建产物/等待抖动导致红灯。
- 失败时：日志/证据完整（包含 docker compose logs / Playwright 报告），定位耗时显著下降。
- 文档/工作流类 PR：在 2025-11-22 之前默认快速通过，不触发重门禁。

## 4. 回滚与回收
- 回滚：任一工作流文件均可按 Git 版本回滚；不涉及数据迁移。
- 回收（11-22）：移除 docs/ci-only 短路，恢复严格门禁；按覆盖率/契约/端到端稳定化结果同步收紧阈值。

## 5. 关联事实来源
- 原则与约束：AGENTS.md
- 构建/运行：Makefile、docker-compose*.yml
- 计划联动：232（E2E 稳定化）、252（权限对齐）、255/256（质量与契约基线）

