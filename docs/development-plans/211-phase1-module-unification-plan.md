# 211-Phase1 模块统一化实施方案

**文档编号**: 211  
**标题**: Phase1 模块统一化实施方案（go.mod 与核心代码迁移）  
**创建日期**: 2025-11-03  
**最后更新**: 2025-11-03  
**关联文档**:  
- `204-HRMS-Implementation-Roadmap.md`（阶段整体路线图）  
- `203-hrms-module-division-plan.md`（模块划分）  
- `CLAUDE.md`（核心原则）  
- `AGENTS.md`（代理执行规范）

---

## 1. 范围与目标

- 范围：执行 Plan 204 第一阶段（Week 1-2）所有模块统一化任务，不引入新业务功能或契约变更。
- 目标：将组织域的 command/query 服务统一迁移至单一 `module cube-castle`，完成目录结构标准化与共享代码抽取，保持 REST/GraphQL 行为一致。
- 成功标准（阶段输出）：`go.mod` 唯一，`cmd/hrms-server/{command,query}` 入口规范，`internal/*` 与 `pkg/*` 分类清晰，编译/测试/部署全链路绿灯。

## 2. 前置条件与准备

- **契约与参考资料**：已对照 `docs/api/openapi.yaml` 与 `docs/api/schema.graphql`，确认当前接口与字段命名无漂移；复核 `docs/reference/02-IMPLEMENTATION-INVENTORY.md`，确保迁移不重复实现。
- **环境基线**：执行 `make docker-up`，确保 PostgreSQL/Redis/Temporal 均在 Docker 容器；宿主机无冲突端口（5432/6379/7233）。若发现宿主服务占用，按 `CLAUDE.md` 要求卸载。
- **代码基线**：
  - `main` 分支同步最新提交，合并前完成 `make fmt`、`npm install`（如依赖更新）。
  - Go 版本与工具链与 `go.mod` 对齐（Plan 204 默认 go1.22.x，如需调整应先提交 PLAN200 系列变更）。
- **CI/CD 约束**：阅读 `.github/workflows/agents-compliance.yml` 与 `.github/workflows/document-sync.yml`，确保目录、命名合规。
- **数据状态**：执行 `goose status`（或团队自建等效脚本）确认迁移基线一致；清理临时测试数据，避免影响 Week 1-2 回归。
- **人员排期**：架构师、后端 TL、QA、DevOps、文档支持均确认两周内可投入，避免与其他关键发布冲突。

## 3. 关键角色与职责

| 角色 | 主要职责 | 交付物 |
|------|----------|--------|
| 架构师 | 定义模块命名与目录蓝图；评估结构调整风险；把关资源唯一性 | Kick-off 纪要、结构审查记录 |
| 后端 TL | 统筹迁移分支、拆分任务、审阅关键 PR；解决依赖冲突 | 迁移 PR、依赖调整说明 |
| 后端开发 | 执行代码迁移、路径调整、共享代码抽取；修复编译/测试问题 | 更新后的代码、变更说明、临时脚本（含 `// TODO-TEMPORARY:`） |
| QA | 制定并执行回归计划（单测/集成/性能）；记录测试结果 | 回归报告 `reports/phase1-regression.md` |
| DevOps | 更新 Makefile/Docker 流程；完成测试环境部署与监控 | 部署日志、环境基线文件 |
| 文档支持 | 更新 README、开发者速查等受影响文档；保持事实唯一性 | 文档 PR、变更记录 |

## 4. 工作分解与时间表（Week 1-2）

| 日期 | 行动项 | 描述 | 负责人 | 产出 |
|------|--------|------|--------|------|
| Day1 | 启动会 | 对齐范围、风险、冻结窗口；确认沟通节奏 | 架构师 + TL | Kick-off 纪要 |
| Day1 | 资产盘点 | 运行 `node scripts/generate-implementation-inventory.js` 并确认差异 | 全员 | 更新后的差异清单 |
| Day2 | 1.1 模块命名 | 审核并锁定 `module cube-castle`；若需路径前缀调整提交方案 | 架构师 | 命名确认记录 |
| Day2 | 分支准备 | 建立 `feature/204-phase1-unify` 分支，配置 CI | 后端 TL | 分支策略说明 |
| Day3 | 1.2 go.mod 合并 | 将子模块依赖合并至根 `go.mod`，移除 `go.work`，确保 `go list ./...` 成功 | 后端 TL | 更新后的 go.mod/go.sum |
| Day3 | 依赖审计 | 记录新增/移除依赖、私有仓库凭证需求 | 架构师 | 依赖审计表 |
| Day4 | 1.3 command 迁移 | 调整源文件至 `cmd/hrms-server/command`；修复 import 与构建脚本 | 后端开发 | 迁移 PR（command） |
| Day4 | 预编译 | `go build ./cmd/hrms-server/command` 预检 | QA | 编译日志 |
| Day5 | 1.4 query 迁移 | 同步迁移 GraphQL 入口至 `cmd/hrms-server/query`；更新生成脚本 | 后端开发 | 迁移 PR（query） |
| Day5 | 构建配置 | 更新 Makefile、Dockerfile、docker-compose 引用路径 | DevOps | 配置变更说明 |
| Day6-7 | 1.5 共享代码抽取 | 分类整理 `/internal` 与 `/pkg`；确保 CQRS 边界；同步 README | 后端团队 + 文档 | 分类清单、文档更新 |
| Day6-7 | 架构审查 | 检查抽取是否破坏模块边界，必要时报回滚 | 架构师 | 审查意见 |
| Day8 | 1.6 编译与测试 | 执行 `go test ./...`、`make test`、`npm run lint`；修复失败 | QA + 后端 | 测试记录 |
| Day9 | 1.7 部署测试环境 | 重新构建镜像，部署至测试环境，更新基线日志 | DevOps | 部署日志、`baseline-*.log` |
| Day9 | 健康检查 | `curl` 健康端点，确认服务就绪 | QA | 健康检查记录 |
| Day10 | 1.8 回归与性能 | 按历史指标执行性能对比；完成功能回归 | QA | `reports/phase1-regression.md` |
| Day10 | 复盘 | 团队复盘输出问题清单、后续注意事项 | 全员 | 复盘记录 |

## 5. 交付物与验收标准

- **交付物**
  - 统一的 `go.mod` / `go.sum`、更新后的目录结构：`cmd/hrms-server/{command,query}`, `internal/*`, `pkg/*`。
  - 更新后的构建脚本、Make 目标、Docker 配置及部署日志。
  - `reports/phase1-module-unification.md`（结构调整说明）与 `reports/phase1-regression.md`（测试结果）。
  - 如涉及文档变更，更新 `README.md`、`docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 等。
- **验收标准**
  - `go build ./cmd/hrms-server`、`go test ./...`、`make test` 全部通过并归档日志。
  - `npm run lint`、`cd frontend && npm run test`（如前端依赖受影响）全部通过。
  - REST/GraphQL 接口对比原版本响应一致（QA 对照脚本输出）。
  - 测试环境部署稳定，健康检查 200 返回，关键日志无新增错误。
  - 变更 PR 获得架构师、后端 TL、QA 审核通过，CI 状态全部绿灯；必要的 `CHANGELOG.md` 更新已提交。

## 6. 验证流程与工具

1. **静态检查**：`make fmt`、`go fmt ./...`、`golangci-lint run ./...`；`node scripts/quality/architecture-validator.js` 确认目录合规。  
2. **自动化测试**：`go test ./... -count=1`，如需集成测试加 `-tags=integration`；前端执行 `npm run lint`、`cd frontend && npm run test -- --runInBand`。  
3. **契约校验**：手动对比 `docs/api/openapi.yaml`、`docs/api/schema.graphql` 与生成代码；如有生成任务需重新运行并审查差异。  
4. **部署验证**：`docker compose -f docker-compose.dev.yml build hrms-server`、`make status`、`curl http://localhost:{9090,8090}/health`，并查看 `run-dev*.log` 确认无新错误。  
5. **性能对比**：使用历史 `baseline-ports.log`、`baseline-processes.log` 与新的指标文件对比；必要时在测试环境运行 `wrk`/`ab` 短暂压测并记录。

## 7. 风险与应对措施

| 风险 | 等级 | 说明 | 应对策略 |
|------|------|------|---------|
| 隐性依赖缺失 | 高 | 合并 go.mod 时遗漏 replace 或私有仓库配置 | Day3 完成 `go list ./...` 与 `go env GOPRIVATE` 校验；记录凭证需求并由 DevOps 配置 |
| CQRS 边界破坏 | 高 | 共享代码抽取导致 command/query 相互耦合 | 架构师 Day6 进行审查，必要时拆回；QA 通过依赖分析脚本验证 |
| 构建/部署失败 | 中 | Docker 构建上下文变更导致 CI/CD 失败 | Day5 预演构建流程；提交前在 CI 中运行 `make build`、`docker compose build` |
| 功能回归 | 中 | 路由或配置遗漏 | QA 制定对照测试脚本，Day8/Day10 执行并记录差异 |
| 时间超期 | 低 | 其他计划冲突或问题修复耗时 | 项目经理在 Day1 冻结窗口，关键风险立即升级 |

## 8. 沟通与状态同步机制

- 每日下午 16:00 站会（≤15 分钟），汇报进展、阻塞与次日计划。
- Day3、Day7、Day10 向 Steering Committee 提交阶段更新（邮件 + `reports/phase1-module-unification.md` 更新日志）。
- 所有文档或脚本变更完成后 24h 内同步至对应事实来源，并在本计划中打勾。
- 若发现资源唯一性或 Docker 约束违规，立即中止相关操作并在 2h 内报告架构师与项目经理。

## 9. 附注与合规要求

- 严格遵守 `CLAUDE.md` 中资源唯一性、Docker 强制、中文沟通与“先契约后实现”原则。任何临时方案须使用 `// TODO-TEMPORARY:` 标注原因、计划、截止日期（≤1 个迭代），并登记在跟踪清单。
- 禁止在宿主机安装 PostgreSQL/Redis/Temporal；如需释放端口需卸载宿主服务，并在 `reports/environment-adjustments.md`（若不存在需新建并纳入事实来源）记录。
- 阶段结束时，在 `docs/development-plans/204-HRMS-Implementation-Roadmap.md` 第一阶段总结中引用本计划成果，并更新 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 中的模块结构记录。

---

**版本历史**  
- v1.0 (2025-11-03): 初始版本
