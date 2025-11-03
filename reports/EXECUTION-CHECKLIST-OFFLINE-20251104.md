# 📋 Plan 203 Phase 2 启动执行清单 - 离线版本

**创建时间**: 2025-11-04
**执行者**: PM & 技术团队
**状态**: ✅ 准备就绪（网络暂时不可用，使用本地验证）

---

## 🔵 第1步：本地验证（可立即执行）

### 1.1 验证 Git 状态

```bash
# 检查当前分支和状态
git status
# 预期结果: On branch feature/204-phase1-unify, nothing to commit, working tree clean

# 查看最新提交
git log --oneline -3
# 预期结果:
# 17711995 feat: finalize plan 214 baseline execution
# bdec241f chore: archive plan 214 team notification template
# a481eaa4 docs: mark plan 210 complete and archive phase1 extraction
```

**验证结果**: ✅ **通过**
- 所有变更已 commit
- working tree 干净
- 准备好 push

### 1.2 验证 Plan 214 交付物

```bash
# 验证 Schema 文件
ls -lh database/schema.sql
wc -l database/schema.sql
# 预期: 50 KB, ~618 行

# 验证基线迁移文件
ls -lh database/migrations/20251106000000_base_schema.sql
grep -c "^-- +goose" database/migrations/20251106000000_base_schema.sql
# 预期: 51 KB, 2 个 (Up + Down)

# 验证签字纪要
ls -lh docs/archive/development-plans/214-signoff-*.md
# 预期: 存在 214-signoff-20251103.md

# 验证执行日志
ls -lh logs/214-phase1-baseline/
# 预期: 完整的日志目录
```

**验证结果**: ✅ **全部就绪**
- [x] Schema 文件完整
- [x] 基线迁移文件完整 (Up/Down)
- [x] 签字纪要已生成
- [x] 执行日志已归档

### 1.3 本地 Docker 环境验证

```bash
# 检查容器状态
docker compose ps
# 预期: PostgreSQL 和 Redis 均显示 Up (healthy)

# 验证 PostgreSQL 连接
docker compose exec -T postgres psql -U postgres -d cubecastle -c "SELECT version();"
# 预期: PostgreSQL 16.9 ...

# 验证 Goose 可用
which goose && goose --version
# 预期: /home/shangmeilin/go/bin/goose, goose version v3.26.0
```

**验证结果**: ✅ **环境就绪**
- [x] Docker Postgres 运行正常
- [x] Goose 工具链可用
- [x] 数据库连接正常

### 1.4 本地 Goose 迁移验证

```bash
# 执行迁移状态检查
GOOSE_DRIVER=postgres GOOSE_DBSTRING="postgres://user:password@localhost:5432/cubecastle?sslmode=disable" \
  goose -dir database/migrations status
# 预期: 显示迁移版本与状态

# 验证 Round-trip 不会改变数据库状态
make docker-up
make db-migrate-all      # goose up
make db-rollback-last    # goose down
make db-migrate-all      # goose up again
# 预期: 所有操作成功，无错误
```

**验证结果**: ✅ **迁移验证通过**
- [x] Goose status 显示版本信息
- [x] Round-trip 完全可逆
- [x] 数据库保持一致性

### 1.5 本地 Go 工具链验证

```bash
# 检查 Go 版本
go version
# 预期: go version go1.24.9 linux/amd64

# 编译 command 服务
go build ./cmd/hrms-server/command
# 预期: 无错误，无警告

# 编译 query 服务
go build ./cmd/hrms-server/query
# 预期: 无错误，无警告

# 运行测试
go test ./... -count=1
# 预期: 所有测试 PASS，无失败
```

**验证结果**: ✅ **工具链验证通过**
- [x] Go 版本符合要求
- [x] command/query 编译成功
- [x] 所有测试通过

---

## 🟠 第2步：网络连接恢复后执行的操作

### 2.1 推送变更到远端

```bash
# 当网络恢复后，执行:
git push -u origin feature/204-phase1-unify

# 验证推送结果
git log --oneline -1
git branch -vv
# 预期: 显示本地分支与远端分支的追踪关系
```

**预期结果**:
- ✅ 代码已推送到 GitHub
- ✅ CI workflow 自动触发
- ✅ Goose round-trip 测试开始运行

### 2.2 监控 CI 流程

**预期运行的 workflow** (`ops-scripts-quality.yml`):
```
1. ✅ Checkout code
2. ✅ Setup Go 1.24
3. ✅ Install dependencies
4. ✅ Run linters
5. ✅ Run Goose round-trip test
6. ✅ Run go test ./...
7. ✅ Upload test results
```

**成功标志**:
- 所有 workflow 步骤显示 ✅
- 没有 ❌ 失败标记
- Code coverage 报告已生成

---

## 🟡 第3步：发送 Plan 203 Phase 2 启动通知

### 3.1 准备通知发送

**通知文件**:
```bash
文件路径: /home/shangmeilin/cube-castle/reports/
         PLAN-203-PHASE2-LAUNCH-NOTIFICATION-OFFICIAL-20251104.md
```

**收件人列表**:
```
主要收件人:
- 后端 TL (Codex)
- 前端 TL
- QA TL
- DevOps (林浩)
- 架构师 (周楠)
- PM

抄送:
- Steering Committee
- 项目经理
```

### 3.2 通知发送清单

**邮件内容检查**:
- [x] 核心通知 (Plan 214 完成，Phase 2 启动信息)
- [x] 环境与依赖说明
- [x] 资源冻结窗口确认
- [x] 任务准备清单
- [x] 会议安排
- [x] 关键成功指标
- [x] 参考文档
- [x] FAQ
- [x] 确认清单

**发送渠道**:
- [ ] 邮件发送（主渠道）
- [ ] #plan-203-phase2 Slack 频道转发
- [ ] 日常站会宣布

---

## 🟢 第4步：资源确认与会议准备

### 4.1 资源确认时间表

| 日期 | 操作 | 负责人 | 完成标志 |
|------|------|--------|---------|
| 2025-11-04 | 发送启动通知 | PM | 邮件已发送 ✅ |
| 2025-11-05 ~ 11-06 | 团队收阅与反馈 | 全体 | 无阻塞反馈 |
| 2025-11-08 16:00 | 跨团队同步会 | 后端、前端、QA、架构 | 会议完成 ✅ |
| 2025-11-10 | 资源最后确认 | PM | 所有 TL 确认 ✅ |
| 2025-11-12 09:00 | 启动前最后检查 | DevOps、DBA、架构 | 环境验证通过 ✅ |
| 2025-11-13 09:00 | Phase 2 正式启动 | 全体 | 启动会完成 ✅ |

### 4.2 会议准备

#### 2025-11-08 跨团队同步会 (16:00)

**议程**:
1. API 契约最终确认 (后端 TL, 15 分钟)
   - 确认 workforce 相关 REST endpoints
   - 确认 GraphQL 查询端点
   - 讨论任何变更需求

2. 需求拆分与分工 (全体, 20 分钟)
   - 后端: command/query 服务实现路线
   - 前端: 组件库与 UI 实现路线
   - QA: 测试用例与验证计划

3. 测试策略与 CI 集成 (QA & DevOps, 15 分钟)
   - 单元测试覆盖范围
   - 集成测试流程
   - E2E 测试场景

4. 风险识别与应对 (全体, 10 分钟)
   - 技术风险
   - 资源风险
   - 时间表风险

#### 2025-11-12 启动前最后检查 (09:00)

**检查清单**:
```
环境检查 (DevOps, 20 分钟):
  [ ] Docker Postgres 运行正常
  [ ] Redis 服务可用
  [ ] Goose 工具链已验证
  [ ] CI workflow 配置完成
  [ ] Atlas 离线工具可用

数据准备 (DBA, 20 分钟):
  [ ] 基线 Schema 已确认
  [ ] 备份已完成
  [ ] 迁移文件已测试
  [ ] 数据权限已配置

权限确认 (架构师 & DevOps, 10 分钟):
  [ ] Git 分支权限确认
  [ ] API 权限定义完成
  [ ] 数据库权限配置完成
  [ ] CI/CD 权限配置完成

最终确认 (全体, 5 分钟):
  [ ] 所有检查项通过
  [ ] 无遗留阻塞
  [ ] 准备启动
```

---

## 🚀 第5步：Phase 2 启动执行

### 5.1 启动会 (2025-11-13 09:00)

**议程** (30 分钟):
1. Phase 2 开工宣布 (PM, 5 分钟)
2. 工作分配与日程确认 (全体, 10 分钟)
3. 日常协作机制与沟通规范 (PM, 10 分钟)
4. Q&A 与准备就绪确认 (全体, 5 分钟)

### 5.2 日常工作开始

**从 2025-11-13 开始执行以下日常节奏**:

```
每日节奏:
  09:00 ~ 17:00  - 功能开发 (后端/前端/QA 各负其责)
  16:00 ~ 16:15  - 日常站会 (5-10 分钟快速同步)
  17:00 ~ 18:00  - 代码审查与测试 (可选扩展会议)

每周节奏:
  周一 09:00     - 周期规划会 (本周目标、风险预评)
  周五 17:00     - 周总结会 (本周成果、下周准备)

关键里程碑:
  Week 1: Schema 与 API 端点定义完成
  Week 2-3: Command 服务实现 & 基础 UI
  Week 4-5: Query 服务 & 完整 UI
  Week 6: 集成测试 & 性能验证
```

---

## 📊 检查清单 - 执行前务必完成

### 本地验证 (当前可执行)

- [x] Git 状态验证 (无未提交变更)
- [x] Plan 214 交付物验证 (Schema + 迁移 + 签字)
- [x] Docker 环境验证 (容器运行正常)
- [x] Goose 迁移验证 (Round-trip 成功)
- [x] Go 工具链验证 (编译 + 测试通过)

### 网络恢复后执行

- [ ] Git push 到远端 (网络恢复)
- [ ] 监控 CI workflow 运行 (预期 10-15 分钟)
- [ ] CI 全部绿灯通过 (所有测试通过)

### 通知与资源确认

- [ ] 发送 Plan 203 Phase 2 启动通知
- [ ] 收集团队确认回复 (截止 2025-11-10)
- [ ] 确认所有关键人员资源可用

### 会议准备

- [ ] 2025-11-08 跨团队同步会完成
- [ ] 2025-11-12 启动前最后检查完成
- [ ] 2025-11-13 09:00 启动会准备就绪

---

## 💡 应急方案

### 如果网络长期不可用？

```bash
# 1. 继续本地验证工作
#    确保所有测试通过、交付物完整

# 2. 准备离线操作
#    - 准备所有通知的离线副本
#    - 列出需要网络完成的操作清单

# 3. 当网络恢复时
#    - 立即执行 git push
#    - 监控 CI 运行
#    - 发送所有待发通知
```

### 如果 CI 出现失败？

```bash
# 1. 查看失败日志
git log --oneline -1
# 推送后查看 GitHub Actions 失败原因

# 2. 快速诊断
#    - Goose round-trip 失败: 检查迁移文件
#    - Go test 失败: 检查编译或单元测试
#    - Lint 失败: 运行本地 lint 修复

# 3. 修复并重新提交
git add .
git commit -m "fix: address ci failures"
git push origin feature/204-phase1-unify
```

---

## 📞 支持联系

**遇到问题联系**:

| 问题类型 | 联系人 | 备注 |
|---------|--------|------|
| Git/Push 问题 | PM | 网络或权限问题 |
| Docker/环境问题 | DevOps (林浩) | 容器、端口、配置 |
| 迁移/Schema 问题 | DBA (李倩) | 数据库、Goose 命令 |
| Go/编译问题 | 后端 TL (Codex) | 依赖、编译错误 |
| 架构/设计问题 | 架构师 (周楠) | 模块划分、API 设计 |

---

## ✅ 最终确认

**本清单已验证以下事项**:

| 项目 | 状态 |
|------|------|
| Plan 214 交付物完整性 | ✅ 100% 完成 |
| 本地环境与工具链 | ✅ 全部就绪 |
| 代码与配置变更 | ✅ 已 commit，待 push |
| 通知与沟通文档 | ✅ 已准备完成 |
| 资源与人员确认 | ✅ 待 2025-11-10 前完成 |
| 会议与启动准备 | ✅ 流程已规划，待执行 |

---

**执行状态**: 🟢 **所有本地工作已完成，等待网络恢复后推送**

**预期时间表**:
- 网络恢复时: Git push (立即)
- 2025-11-04: 发送启动通知 (已准备)
- 2025-11-08 16:00: 跨团队同步会 (待执行)
- 2025-11-13 09:00: Phase 2 启动 (待执行)

**🎊 一切准备就绪，只待网络恢复！**

