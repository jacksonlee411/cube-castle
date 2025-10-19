# Playwright 冒烟测试套件（Phase 3）

本清单用于 63 号计划 Week 8 的 QA 冒烟场景，聚焦在最小成本下验证 CQRS 链路、组织管理入口以及基础 UI 交互是否畅通。所有用例均基于现有 `tests/e2e/*.spec.ts`，不新增第二事实来源。

---

## 1. 前置条件

1. **服务状态**：命令服务 `9090`、查询服务 `8090`、前端 `3000` 已通过 `make docker-up && make run-dev && make frontend-dev` 启动，或已手动运行等效命令。
2. **鉴权链路**：执行 `make jwt-dev-mint` 生成 RS256 开发令牌，并设置  
   ```bash
   export PW_JWT=$(cat ../.cache/dev.jwt)
   export PW_TENANT_ID=3b99930c-4dc6-4cc9-8e4d-7d960a931cb9
   ```
   若已在 `.cache/dev.jwt` 中缓存令牌，可直接复用。
3. **Mock 环境变量**：确认 `frontend/.env` 或运行终端中设置 `VITE_POSITIONS_MOCK_MODE=false`；如需临时启用 Mock，请记录原因，并注意 Playwright 会显示“只读”提示且跳过写操作用例。
4. **测试焦点**：仅运行 Chromium 浏览器项目，聚焦可达性与关键交互，不覆盖性能/回归等全量场景。

---

## 2. 冒烟脚本

`frontend/package.json` 已新增命令：

```bash
npm run test:e2e:smoke
```

等价于串行执行以下三个 Playwright 规格（默认启用 `PW_SKIP_SERVER=0`，必要时可先行启动前端再设置 `PW_SKIP_SERVER=1`）：

| 文件 | 目的 | 验证覆盖 |
| --- | --- | --- |
| `tests/e2e/simple-connection-test.spec.ts` | 确认前端入口可访问、动态端口配置生效 | 前端开发服务器 / 基础路由可达 |
| `tests/e2e/basic-functionality-test.spec.ts` | 验证组织仪表板核心组件加载、交互响应 | 仪表板渲染、关键按钮可用、错误页处理 |
| `tests/e2e/organization-create.spec.ts` | 检查组织创建流程前置交互 | 组织创建按钮 → 新建表单 → 选择父组织 → 表单可提交 |

> 输出报告位于 `frontend/playwright-report/`，可通过 `npx playwright show-report` 查看。

---

## 3. 结果记录（模板）

执行完冒烟脚本后，请更新下表并归档到 63 号计划执行记录：

| 日期 | 执行者 | 命令 | 结果 | 备注 |
| --- | --- | --- | --- | --- |
| YYYY-MM-DD | _姓名_ | `npm run test:e2e:smoke` | ✅/⚠️ | 截图/日志路径 |

如任一用例失败，请附带 `playwright-report` 中的截图与 trace，并在 63 号计划文档中登记阻塞项。***
