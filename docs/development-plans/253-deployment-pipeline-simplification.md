# Plan 253 - 部署流水线简化

文档编号: 253  
标题: 部署流水线简化（来源：202 计划拆分）  
创建日期: 2025-11-15  
版本: v1.1  
关联计划: 202、203、221（Docker 基座）、CI

---

## 1. 目标
- 用最少的步骤完成构建、测试、验收与发布，保持环境一致性；
- 固化镜像标签、预拉取策略与测试基座在流水线中的复用。

## 2. 交付物
- CI 任务清单与依赖（只引用 .github/workflows/*）；
- 发布流程说明（镜像标签、回滚策略）；
- 证据：logs/plan253/*（流水线执行节选与耗时统计）。

## 3. 验收标准
- 构建/测试/验收任务可重入、可复用（冷/热启动下稳定）；
- 端到端执行时间达标（以 204 时间表为基准）；
- 关键步骤产物可追溯（报告、日志）。

---

维护者: DevOps（与 QA/后端协作）

---

## 4. 门禁（CI）
- 端口映射门禁：扫描 docker-compose*.yml，检测 5432/6379/9090/8090 容器端口映射变更；发现变更即阻断（须按 AGENTS 卸载宿主服务）
- 镜像标签门禁：PostgreSQL/Redis 镜像标签禁止使用 `latest`；必须为固定版本标签
- 预拉取与冷启动计时：记录镜像预拉取后冷启动<10s、数据库就绪<15s（首次 CI 运行登记至 215）

---

## 5. 发布流程（版本与回滚）
- 版本标签策略：容器镜像以 `ghcr.io/<org>/hrms:<semver>-<yyyymmddHHMM>` 标记；`latest` 禁止用于 PostgreSQL/Redis，应用镜像仅作为“易用标签”，不参与回滚判定。
- 变更登记：所有镜像版本升级需在 `CHANGELOG.md` 添加条目并附 CI 运行链接；如为基础镜像（PostgreSQL/Redis），需附“兼容性评估与回退计划”小节。
- 回滚策略：通过部署清单回退到上一个通过验收的 `<semver>-<timestamp>`；数据库迁移遵循“迁移即真源”，严格配套 `goose down`，并在 221/253 基座上演练。
- 证据要求：每次发布/回滚在工作流产物中附 `reports/publish-summary.txt`（包含镜像摘要、校验和、触发者、CI 运行链接）。

---

## 6. 工作流与脚本（实现对齐 AGENTS）
- 工作流：`.github/workflows/plan-253-gates.yml`
  - `compose-ports-and-images`（阻断）：调用 `scripts/quality/gates-253-compose-ports-and-images.sh`
    - 冻结端口映射：禁止将 5432/6379/8090/9090 映射到非同号主机端口
    - 固定镜像标签：`image: postgres:*`/`redis:*` 禁止 `latest`
  - `coldstart-metrics`（记录）：调用 `scripts/quality/gates-253-coldstart.sh` 预拉取后测量冷启动/数据库就绪时间，产物落盘 `logs/plan253/` 并通过 `upload-artifact` 上传
- 证据落盘：`logs/plan253/*`（CI 自动写入，作为 215 索引的唯一事实来源）

---

## 7. 证据与对齐
- 门禁日志：`logs/plan253/compose-ports-and-images.log`
- 冷启动与健康就绪：`logs/plan253/coldstart-*.log`（包含 compose 文件名、耗时、时间戳）
- 215 登记：在 `docs/development-plans/215-phase2-execution-log.md`“25x 启动登记”处仅做索引，不重复细节，避免第二事实来源。
