# Plan 06 – 集成测试行动记录（2025-11-07 更新）

> 唯一事实来源：`docs/development-plans/219E-e2e-validation.md`、`logs/219E/`。  
> 本记录聚焦当前阻塞与分阶段推进路径。

## 当前状态

- **219D1~219D5**：已完成，Scheduler 配置/监控/测试资料可直接复用
- **219E 端到端与性能验证**：因Docker socket权限问题暂停（详见 `logs/219E/BLOCKERS-2025-11-06.md`）
- **新脚本入库**：已准备就绪，等待Docker环境恢复
  - `scripts/e2e/org-lifecycle-smoke.sh`（组织/部门生命周期冒烟）
  - `scripts/perf/rest-benchmark.sh`（REST P99 基准）

---

## 分阶段推进路径

### 第一阶段：紧迫路径（并行进行，不依赖Docker）

在Docker权限问题解决前，以下工作可立即推进：

1. **代码审查与单元测试**
   - 审查 219E 相关代码更改（特别是Scheduler、Assignment、缓存刷新逻辑）
   - 本地运行单元测试：`go test ./...`（无需Docker）
   - 输出结果至 `logs/219E/unit-tests-*.log`

2. **文档完善**
   - 补齐 219E 验收清单：`docs/development-plans/219E-e2e-validation.md` § 测试范围表格
   - 性能基准表格初稿：`docs/reference/03-API-AND-TOOLS-GUIDE.md` § 性能部分
   - 预期完成：3-5个工作日

3. **Playwright脚本编写与预审**
   - 基于 `frontend/tests/e2e/*.spec.ts` 编写或扩展测试用例
   - 代码审查通过但不执行（缺Docker）
   - 预期完成：2-3个工作日

**预期时间**：1-2周  
**产出**：代码审查完毕、文档就绪、脚本预审通过

---

### 第二阶段：恢复Docker访问（阻塞解除条件）

#### 问题诊断与修复

执行 Docker 权限诊断与修复：

**快速诊断**
```bash
bash -c 'echo "=== Docker 权限诊断 ==="; echo "1. Docker daemon:"; docker ps 2>&1 | head -2; echo "2. socket信息:"; ls -la /var/run/docker.sock 2>/dev/null || echo "不存在"; echo "3. 当前用户:"; whoami; echo "4. 用户所在组:"; groups'
```

**修复选项**（详见 `docs/troubleshooting/docker-socket-permission-fix.md`）

| 选项 | 适用场景 | 时间 |
|------|--------|------|
| **方案A**：用户组授权（推荐） | 多数Linux/WSL2环境 | 5分钟 |
| **方案B**：Socket权限修改 | 无sudo权限或用户组失效 | 10分钟 |
| **方案C**：CI/CD runner | 本地权限无法获取 | 在GitHub Actions中执行 |

**验证Docker访问已恢复**
```bash
# 运行以下命令，应无 permission denied
docker ps
docker images | grep cube-castle

# 启动全栈服务
make run-dev
# 或
docker compose -f docker-compose.dev.yml up -d

# 验证关键服务运行
docker ps | grep -E 'postgres|redis|rest-service|graphql-service'
```

**优先级**：⚠️ 阻塞路径，需立即推进（或在CI/CD中绕过）

---

### 第三阶段：执行E2E验证（Docker恢复后）

#### 3.1 执行组织生命周期冒烟脚本

在服务就绪后运行：

```bash
# 配置变量（按需自定义）
export COMMAND_API="http://localhost:9090"   # 与 .env.example / docker-compose.dev.yml 对齐
export TENANT_ID="3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"  # DEFAULT_TENANT_ID 唯一事实来源
export JWT_TOKEN="your-token"

# 执行脚本
scripts/e2e/org-lifecycle-smoke.sh

# 检查输出日志
ls -la logs/219E/org-lifecycle-*.log
```

**验证项**：
- ✅ REST API 组织创建/更新/删除返回预期状态码
- ✅ GraphQL 查询返回正确的组织结构
- ✅ 部门生命周期变更正确同步

**预期完成**：1-2天

---

#### 3.2 采集REST性能基准

安装 `hey` 工具并执行性能测试：

```bash
# 安装 hey
go install github.com/rakyll/hey@latest

# 执行性能基准
scripts/perf/rest-benchmark.sh

# 收集性能数据
ls -la logs/219E/perf-rest-*.log
```

**收集指标**：P50、P95、P99 响应时间（ms），吞吐量（req/s）  
**对标基线**：与219D历史性能对比，记录差异原因  
**登记位置**：`docs/reference/03-API-AND-TOOLS-GUIDE.md` § 性能部分

**预期完成**：1天

---

#### 3.3 Playwright 前端E2E验证

复用并扩展前端测试：

```bash
# 运行现有E2E测试
cd frontend
npm run test:e2e

# 重点验证场景：
# - Assignment 功能完整性
# - Outbox → Dispatcher → 缓存刷新路径
# - 性能场景（大批量操作）
```

**失败用例处理**：
- 整理至 `logs/219E/playwright-failures-*.log`
- 在 219E 文档"测试范围"表格更新状态（PASS/FAIL）
- 失败原因追踪至代码（特别注意缓存一致性）

**预期完成**：2-3天

---

### 第四阶段：回退演练（验证完毕后）

依据 `internal/organization/README.md#Scheduler / Temporal（219D）` 的回退指引，演练一次完整回退：

```bash
# 1. 禁用Scheduler
export SCHEDULER_ENABLED=false

# 2. 恢复至219D1目录结构（保留数据）
# 依据文档指引，恢复相关配置

# 3. 验证业务连续性
make run-dev
scripts/e2e/org-lifecycle-smoke.sh

# 4. 记录结果
# logs/219E/rollback-test-*.log
```

**输出**：回退步骤文档、验证日志、风险评估

**预期完成**：1-2天

---

## 参考资料

### 核心文档
- **Docker权限诊断与修复**：`docs/troubleshooting/docker-socket-permission-fix.md`
- **阻塞说明**：`logs/219E/BLOCKERS-2025-11-06.md`
- **219E验收标准**：`docs/development-plans/219E-e2e-validation.md`

### 支撑资料
- **服务日志**：`logs/219D2/`、`logs/219D3/`、`logs/219D4/`
- **监控配置**：`docs/reference/monitoring/`
- **开发者速查**：`docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`
- **Docker Compose配置**：`docker-compose.dev.yml`
- **Makefile目标**：`make run-dev`、`make stop-dev`

---

## 时间预估与关键路径

```
第一阶段（并行）
├─ 代码审查 & 单元测试      [3-5天]
├─ 文档完善                [3-5天]
└─ Playwright脚本预审      [2-3天]
   → 预期完成：1-2周

第二阶段（阻塞解除）
└─ Docker权限诊断与修复     [5-30分钟 或 CI/CD绕过]
   → 预期完成：同日或使用CI/CD

第三阶段（顺序执行）
├─ 冒烟脚本                [1-2天]
├─ 性能基准                [1天]
├─ Playwright验证          [2-3天]
└─ 问题修复与迭代          [按需]
   → 预期完成：1-2周

第四阶段（验证后）
└─ 回退演练                [1-2天]
   → 预期完成：1-2天

总耗时：3-4周（Docker恢复后加速）
关键路径：Docker权限 → E2E验证 → 回退测试
```

---

## 即时行动项

**今天可做**：
- [ ] 运行Docker诊断脚本，确认问题类型
- [ ] 开始第一阶段：代码审查、文档编写

**本周必做**：
- [ ] 执行Docker权限修复（方案A/B/C选一）
- [ ] 完成第一阶段所有工作

**恢复Docker后**：
- [ ] 执行第二、三、四阶段（按时间预估推进）

---

> **更新说明**：本文档于2025-11-07简化重构，重点突出分阶段推进路径和Docker权限问题的具体解决方案。如有遗漏或需更新，请参考单一事实来源文档：219E验收标准、阻塞清单、troubleshooting指南。
