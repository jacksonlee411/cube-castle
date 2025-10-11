# 64号文档：Phase 3 验收草案（前端 API / Hooks / 配置整治）

**版本**: v0.1  
**创建日期**: 2025-10-12  
**维护人**: 全栈工程师（单人执行）  
**关联计划**: 60号总体计划、61号执行计划、63号前端整治计划、06号测试执行要求  

---

## 1. 执行摘要

Phase 3 聚焦前端查询客户端、Hooks 与运行配置的统一。本草案记录当前可交付成果与剩余工作：

- ✅ **统一查询客户端与 Hooks**：`frontend/src/shared/api/queryClient.ts`、`frontend/src/shared/hooks/*` 已完成迁移，测试覆盖率 ≥ 75%（参见 06 号文档第 3 节记录）。  
- ✅ **修复 JWKS 代理 SSL 告警**：确认根因来自 Vite Proxy 默认将命令服务视为 HTTPS，已在 `frontend/src/shared/config/ports.ts` 调整默认协议推断，DEV/TEST 环境回退为 HTTP。  
- ✅ **配置/QA 文档同步**：更新 `docs/development-plans/06-integrated-teams-progress-log.md` 常见故障与 TODO；在 `docs/reference/03-API-AND-TOOLS-GUIDE.md` 补充代理运行说明。  
- ⏳ **Bundle 体积目标**：`npm run build:analyze` 已恢复可用，需对比基线确认 ≥5% 优化或提供解释。  

---

## 2. 验收检查清单（当前状态）

| 项目 | 权威来源 | 状态 | 说明 |
|------|----------|------|------|
| 查询客户端与错误包装统一 | `shared/api/queryClient.ts`、63号文档 §4.1 | ✅ | 所有核心 Hooks 已迁移，Vitest 用例覆盖 |
| Hooks 重构完成 | `frontend/src/shared/hooks/`、63号文档 §4.2 | ✅ | 写操作 Hook 复用统一客户端，桥接层无需保留 |
| 配置助手更新 | `frontend/src/shared/config/` | ✅ | `environment.ts`/`ports.ts` 同步优化，支持显式协议覆盖 |
| JWKS 代理告警关闭 | 06号文档 §4、`ports.ts` | ✅ | DEV/TEST 默认 HTTP，提供 HTTPS 手册 |
| QA 文档同步 | 06号文档 §3-4、03-API-AND-TOOLS-GUIDE | ✅ | 常见故障表新增 JWKS 条目，指南记录代理说明 |
| Vitest 覆盖率 ≥ 75% | 06号文档 §3 | ✅ | 2025-10-11 覆盖率 84.1%（Phase 3 模块） |
| Playwright 冒烟通过 | 06号文档 §3 | ✅ | `npm run test:e2e:smoke` 完成，报告归档 |
| Bundle 体积下降 ≥ 5% 或解释 | 63号文档 §4.3 | ⏳ | `npm run build:analyze` 恢复，待对比基线输出 |

---

## 3. 验证步骤与运行说明

### 3.1 开发/QA 环境准备

```bash
make docker-up
make run-dev         # 命令服务 9090 / 查询服务 8090（RS256）
make frontend-dev    # Vite Dev Server 3000
```

- DEV/TEST 环境默认通过 HTTP 代理命令服务：无需额外配置即可避免 `/.well-known/jwks.json` EPROTO。  
- 若需要在 QA 环境启用 HTTPS，请：
  1. 为命令服务部署有效 TLS 证书；
  2. 在前端启动前设置 `VITE_SERVICE_PROTOCOL=https`；
  3. 验证 `curl https://<host>:9090/.well-known/jwks.json` 正常返回 JWKS。

### 3.2 运行时验证

1. **确认 Vite 代理输出**（Node 环境）  
   ```bash
   cd frontend
   npx tsx -e "import('./src/shared/config/ports.ts').then(m => console.log(m.CQRS_ENDPOINTS))"
   # 预期输出: http://localhost:9090 / http://localhost:8090
   ```
2. **冒烟测试**（详见 06 号文档表格）  
   ```bash
   npm run test:e2e:smoke
   ```
3. **覆盖率验证**  
   ```bash
   npx vitest run --coverage --run
   ```
4. **Bundle 分析**（待完成验收项）  
   ```bash
   npm run build:analyze
   # 输出与基线对比待记录
   ```

### 3.3 文档同步记录

- `docs/development-plans/06-integrated-teams-progress-log.md`：新增 JWKS EPROTO 处理条目并结案 TODO。  
- `docs/reference/03-API-AND-TOOLS-GUIDE.md`：在环境启动小节补充 Vite 代理与协议覆盖说明。  
- 本草案（64号文档）：汇总 Phase 3 当前验收状态并承接后续行动。

---

## 4. 风险与待办

| 风险/待办 | 影响 | 当前进度 | 缓解措施 |
|-----------|------|----------|----------|
| Bundle 体积优化目标尚未量化 | 中 | `build:analyze` 可执行，待比对 | 与 63 号文档协同，记录基线数据与优化策略 |
| QA 运行手册需补充 HTTPS 场景 | 低 | HTTP 场景已文档化 | 结合后续 HTTPS 验证结果追加示例 |
| Phase 3 验收报告定稿 | 中 | 草稿建立 | 完成待办后更新至 v0.2 并申请评审 |

---

## 5. 后续行动

1. 收集 `npm run build:analyze` 最新输出，与历史基线比对并更新 §2 状态。  
2. 根据是否启用 HTTPS 补全 QA 操作示例与风险说明。  
3. 汇总 Phase 3 交付成果，准备将本草案升级为正式验收报告并归档。  
4. 同步 60/63 号计划的进度看板，确保交付闭环。  

---

**当前状态**：v0.1 —— 主要目标完成，非功能性指标待补充  
**下一步**：完成 bundle 体积验收与 HTTPS QA 示例，更新为 v0.2 草稿后申请评审  
