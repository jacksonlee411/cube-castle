# Plan 219D2 – 调度配置集中化与启动流程更新

**文档编号**: 219D2  
**关联路线图**: Plan 219 → 219D  
**依赖子计划**: 219D1 完成代码迁移；219A 目录结构约束  
**目标周期**: Week 4 Day 21（219D 并行阶段）  
**负责人**: 后端团队（配置 Owner）

---

## 1. 目标

1. 汇总调度相关配置（cron 表达式、队列名称、重试策略、worker 并发度），统一落在配置包与 `.env`/YAML。
2. 更新启动链路（Makefile、`make run-dev`、Docker Compose env）以确保 Scheduler 默认启用且可调试。
3. 建立配置变更的校验流程（默认值说明 + 变更 checklist）。

---

## 2. 范围

| 模块 | 内容 |
|------|------|
| 配置项归集 | 抓取所有 scheduler/Temporal 相关配置，统一声明并去重 |
| 启动流程 | Makefile、`cmd/hrms-server/command/main.go`、`config/` 中的初始化参数 |
| 检查机制 | 添加配置校验/日志，记录默认值及覆盖来源 |

不包含：代码迁移（219D1）、指标埋点（219D3）、测试与文档编写（219D4/219D5）。

---

## 3. 详细任务

1. **配置盘点**
   - 遍历 `config/`、`cmd/hrms-server/command/internal`、`.env*`、`docker-compose.*` 搜索 `cron`、`temporal`、`scheduler` 关键字。
   - 形成表格：参数名称、当前位置、默认值、用途、依赖调用方。

2. **集中化实现**
   - 在 `internal/config/`（或现有 config 包）新增 Scheduler 段，导出结构体/函数读取配置。
   - 更新应用启动逻辑：统一从配置结构体获取 cron/队列参数。
   - 若需迁移环境变量，更新 `.env.example` 与 Docker Compose env。

3. **启动流程对齐**
   - 调整 Makefile 目标（如 `make run-dev`、`make run-scheduler`），确保 scheduler 自动加载。
   - 补充说明如何在本地禁用或覆盖参数，避免调试冲突。

4. **配置校验与回滚**
   - 为关键参数添加启动时校验（空值、格式、范围）。
   - 记录旧值与回滚方法，纳入配置变更 checklist。

---

## 4. 验收标准

- [ ] 所有调度相关参数集中在单一配置入口，并在 `.env.example` 或配置文档中列明默认值。
- [ ] `make run-dev`、Docker Compose 启动后 Scheduler 正常读取新配置，日志无报错。
- [ ] 配置校验机制可在参数缺失或格式错误时阻止启动并给出指引。

---

## 5. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 遗漏散落配置项 | 中 | 使用脚本抓取关键字并与 219D1 迁移清单交叉验证 |
| 配置改动影响已有环境 | 中 | 升级前通知平台团队，提供回滚默认值；在 sandbox 先行验证 |
| 启动脚本修改引入回归 | 中 | 运行 `make run-dev`、`make docker-up`、`make test` 全量验证 |

---

## 6. 交付物

- 更新后的配置文件与 `.env.example`。
- 配置盘点表与检查清单（可附于 PR 或文档附录）。
- 经验证的启动日志与失败示例（校验机制截图/日志片段）。
