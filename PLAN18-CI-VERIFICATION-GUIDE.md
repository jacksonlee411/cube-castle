# Plan 18 CI 验证操作指南

**目的**：完成 Plan 18 启动前的 P0 必需操作（CI 环境验证）
**预计耗时**：10-15 分钟（设置）+ 30-45 分钟（workflow 运行）
**前置条件**：✅ 已完成（workflow 已提交并推送）

---

## 快速操作步骤

### 步骤 1：创建验证 PR（触发 CI）

#### 方式 A：通过浏览器创建 PR（推荐，无需本地工具）

1. **点击以下链接**：
   ```
   https://github.com/jacksonlee411/cube-castle/pull/new/plan16-code-smell-implementation
   ```

2. **填写 PR 信息**：
   - **标题**：`[CI验证] Plan 18 E2E 环境测试`
   - **描述**（复制以下内容）：
     ```markdown
     ## 目的
     验证 Plan 18 的 GitHub Actions E2E workflows 在 CI 环境可用性

     ## 验证范围
     - ✅ e2e-tests.yml workflow 完整运行
     - ✅ Docker Compose E2E 栈启动成功
     - ✅ 服务健康检查通过
     - ✅ JWT mint 流程正常
     - ✅ Playwright 测试执行（不要求全部通过）
     - ✅ Artifact 上传成功

     ## 验证报告
     结果将记录到 reports/iig-guardian/plan18-ci-verification-20251002.md

     ## 注意
     本 PR 为验证用途，不应合并到主分支。验证完成后将关闭此 PR。
     ```

3. **选择**：✅ **Create draft pull request**（草稿状态）

4. **点击**："Create pull request"

5. **等待**：GitHub Actions 自动触发，预计 30-45 分钟完成

---

#### 方式 B：安装 gh CLI 后手动触发（可选）

```bash
# 1. 安装 gh CLI（如未安装）
# Ubuntu/Debian:
sudo apt install gh

# macOS:
brew install gh

# 2. 认证
gh auth login

# 3. 手动触发 workflow
gh workflow run e2e-tests.yml --ref plan16-code-smell-implementation

# 4. 监控运行
gh run watch
```

---

### 步骤 2：监控 Workflow 运行

**监控入口**（任选一种）：

- **PR Checks 页签**（如使用方式 A）
- **GitHub Actions 页面**：https://github.com/jacksonlee411/cube-castle/actions

**关注要点**：

| 步骤名称 | 验收标准 | 预计耗时 |
|---------|---------|---------|
| Set up Node.js | ✅ 成功 | <1 分钟 |
| Set up Go | ✅ 成功 | <1 分钟 |
| Install frontend dependencies | ✅ 成功 | 2-3 分钟 |
| Start E2E stack | ✅ Docker 镜像构建成功 | 5-10 分钟 |
| Wait for services | ✅ 所有服务健康检查通过 | 1-2 分钟 |
| Mint dev JWT | ✅ JWT 生成成功 | <30 秒 |
| Run Playwright E2E suite | ⚠️ 执行即可（允许部分测试失败） | 10-20 分钟 |
| Upload Playwright report | ✅ Artifact 上传成功 | 1-2 分钟 |

**成功标准**：
- ✅ **核心成功**：前 7 步全部通过（到 "Mint dev JWT"）
- ⚠️ **可接受**：Playwright 测试有失败，但 Artifact 成功上传
- ❌ **需修复**：Docker 栈启动失败或服务健康检查超时

---

### 步骤 3：下载验证证据

**如 Workflow 运行完成**：

1. **进入 Workflow 运行页面**
   - 点击 PR 的 "Checks" 页签
   - 或访问 GitHub Actions 页面找到最新运行

2. **复制 Workflow URL**
   - 格式：`https://github.com/jacksonlee411/cube-castle/actions/runs/{run_id}`
   - 保存此 URL，稍后填入验证报告

3. **下载 Artifacts**
   - 滚动到页面底部 "Artifacts" 区域
   - 下载 `playwright-report`（如存在）
   - 解压并打开 `index.html` 查看测试报告

---

### 步骤 4：填写验证报告

**编辑文件**：`reports/iig-guardian/plan18-ci-verification-20251002.md`

**需要填写的关键信息**：

1. **四、验证结果 > 4.1 执行摘要**
   ```markdown
   **状态**：✅ 验证通过 / ⚠️ 部分通过 / ❌ 验证失败

   **Workflow 运行信息**：
   - 运行 ID：{从 URL 提取}
   - 运行 URL：https://github.com/jacksonlee411/cube-castle/actions/runs/{run_id}
   - 触发方式：PR 创建
   - 运行时间：2025-10-02 {具体时间}
   - 总耗时：{X} 分钟

   **结果**：
   - [x] ✅ 验证通过：所有关键指标满足验收标准
   ```

2. **四、验证结果 > 4.2 详细指标验证结果**
   - 将表格中的 `⏳ 待验证` 更新为 `✅ 通过` 或 `❌ 失败`
   - 填写"说明"列（如有失败或异常）

3. **四、验证结果 > 4.3 发现的问题**
   - 如有失败步骤，记录问题详情

4. **五、结论与建议 > 5.1 验证结论**
   ```markdown
   **✅ 验证通过**

   CI 环境配置完整可用，满足 Plan 18 启动条件：
   - Docker Compose E2E 栈在 GitHub Actions 成功启动
   - 所有服务健康检查通过
   - JWT mint 流程正常
   - Playwright 测试套件成功执行
   - Artifact 上传功能正常

   **可以解除 Plan 18 的 P0 阻塞项，启动 Phase 1 任务。**
   ```

---

### 步骤 5：更新 Plan 18 文档

**编辑文件**：`docs/development-plans/18-e2e-test-improvement-plan.md`

**更新 1：📊 执行摘要**
```markdown
**状态**：✅ **就绪，可启动**

**关键阻塞项**（优先级 P0）：
- ~~CI 环境未实际验证~~ → ✅ 已完成验证
```

**更新 2：6.2 依赖项**
```markdown
- ✅ CI 环境配置（GitHub Actions）— **已验证可用**
  - ✅ docker-compose.e2e.yml 已创建并提交（ff383eb0）
  - ✅ frontend-e2e.yml workflow 已创建并提交（7f1644eb）
  - ✅ e2e-smoke.yml workflow 已创建并提交（ff383eb0）
  - ✅ e2e-tests.yml workflow 已创建并提交（7cbba95d）
  - ✅ **首次成功运行验证**：{Workflow URL}
  - 📄 验证报告：reports/iig-guardian/plan18-ci-verification-20251002.md
```

---

## 常见问题

### Q1：Workflow 运行失败怎么办？

**A1**：根据失败步骤定位问题：

- **Docker 构建失败**：检查 `docker-compose.e2e.yml` 配置和镜像拉取
- **健康检查超时**：增加 `wait for services` 步骤的重试时间
- **JWT mint 失败**：确认 `make jwt-dev-mint` 依赖的工具（jq、curl）已安装
- **Playwright 失败**：查看 Artifact 中的错误日志和截图

**修复后重新触发**：
- 推送修复代码到同一分支
- PR 会自动重新运行 CI
- 或手动点击 "Re-run jobs"

---

### Q2：需要所有 Playwright 测试都通过吗？

**A2**：**不需要**。验证目标是 **CI 环境可用性**，而非测试通过率。

**验收标准**：
- ✅ **必须通过**：Docker 栈启动、服务健康检查、JWT mint
- ⚠️ **允许失败**：Playwright 测试（已知有 3 个失败，见 Plan 18 第二章）
- ✅ **必须成功**：Artifact 上传（即使测试失败也要能下载报告）

---

### Q3：验证通过后需要合并 PR 吗？

**A3**：**不需要**。此 PR 仅用于验证，应：
1. ✅ 保持 draft 状态
2. ✅ 添加评论说明验证结果
3. ✅ 关闭 PR（不合并）

实际功能开发的 PR 会在 Phase 1 任务中另行创建。

---

### Q4：验证耗时超过 45 分钟怎么办？

**A4**：Workflow 会自动超时失败。可能原因：
- Docker 镜像构建过慢（网络问题）
- 服务启动卡住
- Playwright 测试执行过久

**解决方案**：
- 检查 Workflow 日志找到卡顿步骤
- 调整超时配置或优化执行速度
- 考虑拆分测试（仅运行核心测试）

---

## 总结

完成以上 5 个步骤后，Plan 18 的 P0 阻塞项将被解除，可以正式启动 Phase 1 任务。

**当前状态**：
- ✅ 步骤 1：workflow 已提交并推送
- ⏳ 步骤 2-5：等待您执行

**下一步行动**：
👉 **点击创建 PR**：https://github.com/jacksonlee411/cube-castle/pull/new/plan16-code-smell-implementation
