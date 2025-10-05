# Plan 18 本地端到端验证操作指南

**目的**：完成 Plan 18 启动前的 P0 必需操作（本地端到端验证）
**预计耗时**：约 20-30 分钟（含依赖启动 + Playwright 执行）
**前置条件**：已完成 `docker-compose.e2e.yml`、Playwright 套件与 RS256 工具链配置

---

## 快速操作步骤

# 步骤 1：启动依赖栈

```bash
make docker-up
make run-auth-rs256-sim
```

> 若本地已有运行中的服务，请先执行 `make dev-kill` 清理端口占用（9090 / 8090）。

完成后使用下列命令确认依赖健康：

```bash
curl -fsS http://localhost:9090/health
curl -fsS http://localhost:8090/health
```

两条命令均返回 200 即视为通过。

---

### 步骤 2：生成 JWT 并执行 Playwright 套件

```bash
# 生成 RS256 开发令牌
make jwt-dev-mint

# 设置环境变量
export PW_JWT=$(cat .cache/dev.jwt)
export PW_TENANT_ID=3b99930c-4dc6-4cc9-8e4d-7d960a931cb9

# 进入前端目录并运行完整 Playwright 套件
cd frontend
PW_JWT=$PW_JWT PW_TENANT_ID=$PW_TENANT_ID npm run test:e2e
```

**提示**：创建场景需要填写 `form-field-effective-date`、`form-field-name` 等字段，删除场景会触发 `temporal-delete-record-button` 与确认弹窗。

**预期结果**：
- Playwright 至少完成一次执行（允许存在失败用例）
- 生成 `frontend/playwright-report/` 与 `frontend/test-results/`

---

### 步骤 3：收集验证证据

| 证据项目 | 操作 |
|----------|------|
| Playwright 报告 | 打开 `frontend/playwright-report/index.html`，确认日志与截图存在 |
| 失败截图/视频 | 检查 `frontend/test-results/` 目录 |
| 命令输出 | 保存 `npm run test:e2e` 的终端输出摘要 |
| 环境信息 | 记录操作系统、Docker、Node.js、Go 的版本号 |

如需保留完整命令日志，可执行：

```bash
PW_JWT=$PW_JWT PW_TENANT_ID=$PW_TENANT_ID npm run test:e2e | tee ../reports/iig-guardian/plan18-local-validation.log
```

---

### 步骤 4：填写验证报告

1. 编辑 `reports/iig-guardian/plan18-ci-verification-20251002.md`
2. 在 “四、验证结果” 中填写：
   - 执行日期与时长
   - 健康检查结果
   - Playwright 执行结论（通过 / 部分通过 / 失败）
   - 关键日志或问题
3. 在 “五、结论与建议” 中确认是否满足启动条件。

范例（若验证通过）：

```markdown
**状态**：✅ 验证通过

- 健康检查：9090 / 8090 服务均返回 200
- JWT 生成：.cache/dev.jwt 已生成并载入
- Playwright：chromium / firefox 项目均完成执行，失败 0 项（或列出失败详情）
- 产物：frontend/playwright-report/index.html 已生成
```

---

---

### 步骤 5：更新 Plan 18 文档

完成验证后，在 `docs/development-plans/18-e2e-test-improvement-plan.md` 中更新：

```markdown
- ✅ 本地端到端验证 — 已执行 {YYYY-MM-DD}
  - 验证报告：reports/iig-guardian/plan18-ci-verification-20251002.md
  - Playwright 报告：frontend/playwright-report/index.html
```

---

## 常见问题

### Q1：`make run-auth-rs256-sim` 启动失败怎么办？

**A1**：
- 确认 Docker 服务已启动
- 检查 9090/8090 端口是否被占用，必要时执行 `make dev-kill`
- 查看 `run-dev*.log` 获取详细错误

---

### Q2：Playwright 有少量失败算通过吗？

**A2**：可以。验证的目标是确认环境可完整运行流程。若个别用例失败，请在报告中记录并附上截图或日志，后续在 Phase 1 修复。

---

### Q3：验证结束后需要保留哪些文件？

**A3**：
- `frontend/playwright-report/` 与 `frontend/test-results/`
- `reports/iig-guardian/plan18-ci-verification-20251002.md`
- 如有失败，保留控制台输出或截图用于分析

---

### Q4：如何加速验证耗时？

**A4**：
- 预先执行 `npm ci`、`npx playwright install --with-deps`
- 确保 Docker 镜像已构建过一次
- 运行 Playwright 时追加 `--project=chromium` 进行快速验证（留存完整流程需两种浏览器时再补跑）

---

## 总结

完成以上 5 个步骤后，Plan 18 的 P0 阻塞项将被解除，可以正式启动 Phase 1 任务。

**当前状态**：
- ⏳ 步骤 1-5：需在本地执行并记录结果

如需协助，请参考 `docs/development-plans/18-e2e-test-improvement-plan.md` 第七章或联系 QA 团队支持。
