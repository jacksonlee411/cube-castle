# 210-数据库基线重建方案（Early Reset）

**文档编号**: 210  
**标题**: 事务性迁移工作流基线重建计划  
**创建日期**: 2025-11-06  
**最后更新**: 2025-11-05  
**撰写**: 基础设施组 / 架构组联合  
**关联文档**:
- `200-Go语言ERP系统最佳实践.md`（数据库治理原则）
- `203-hrms-module-division-plan.md`（P0 迁移治理要求）
- `205-HRMS-Transition-Plan.md`（服务统一化步骤）
- `212-shared-architecture-alignment-plan.md`（Day6-7 架构审查结论，已于 2025-11-04 完成）
- `213-go-toolchain-baseline-plan.md`（Go 1.24 工具链基线评审，Steering 已确认保持 Go 1.24）
- `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`（环境操作手册）
 - `reports/PLAN-210-IMPLEMENTATION-VERIFICATION-REPORT.md`（完成验证记录）

**状态**: 🟢 已完成（2025-11-04）

---

## 1. 背景与目标

- 现状：`database/migrations/` 内存在 40+ 旧式 `.sql` 迁移文件，缺乏 `down` 脚本，且命名不符合 Goose/Atlas 工作流标准。  
- 风险：CI/CD 与多环境部署无法保证幂等性；引入新模块前的 P0 要求（203 号方案）无法满足。  
- 机会：当前数据为演示/测试数据，可在无需迁移历史数据的情况下一次性重建基线。  
- 目标：在进入 Core HR P0 模块（workforce）开发前，完成“**声明式 Schema + 基线迁移 + Goose 工作流**”的落地，使得迁移体系具备：
  1. 100% up/down 成对  
  2. 标准化命名（`YYYYMMDDHHMMSS_description.sql`）  
  3. 支持 Atlas 自动生成增量迁移  
  4. 支持 CI 中的 `goose up/down` 校验

---

## 2. 策略概览

| 阶段 | 时间窗口 | 核心目标 | 负责人 | 完成信号 |
|------|---------|----------|--------|---------|
| Phase 0 | Week 0 (当前周) | 冻结现有迁移 & 备份验证环境 | 基础设施组 | ✅ 备份文件、验证基线（详见验证报告）|
| Phase 1 | Week 1 | 基线 Schema 萃取与基线迁移生成 | 架构组 + DBA | ✅ `20251106000000_base_schema.sql` & Schema 定稿 |
| Phase 2 | Week 1-2 | Goose/Atlas 工作流落地，清空旧迁移 | 基础设施组 | ✅ 新迁移目录 + CI Goose 流程 |
| Phase 3 | Week 2 | 全环境回放验证与文档收尾 | QA + DevOps | ✅ `goose up/down` 通过 + 签字归档 |

### 2.1 资源投入概览

| Phase | 预估人天 | 关键角色 | 说明 |
|-------|---------|---------|------|
| Phase 0 | 2 | 基础设施组 2 人 | 冻结、备份与验证（依赖 Plan 212 架构决议已落地）|
| Phase 1 | 4 | DBA 1 人 + 架构组 1 人 | Schema 萃取、基线生成与审阅（含重试）|
| Phase 2 | 3 | 基础设施组 1 人 + DevOps 1 人 | Goose/Atlas 配置、CI 集成、旧迁移清理；确保环境 Go 1.24 toolchain（Plan 213 决议）|
| Phase 3 | 2 | QA 1 人 + DevOps 1 人 | Round-trip 测试、文档同步 |
| **总计** | **11** | - | **两周完成，预留一周缓冲** |

---

## 3. 具体实施计划

### Phase 0：冻结与备份

1. **冻结仓库迁移目录**  
   - 创建保护分支：`backup/migrations-pre-reset`。  
   - 将 `database/migrations/` 当前内容打包存档（`archive/migrations-pre-reset-20251106.tar.gz`）。  
   - CI 中禁止新的 `.sql` 迁移合并（设置临时守护规则）。

2. **数据库备份**  
   - 通过 Docker 容器执行 `pg_dump`，输出至 `backup/`（示例命令）：  
     ```bash
     docker compose exec -T postgres pg_dump --dbname="$PG_BASELINE_DSN" \
       > backup/pgdump-baseline-$(date +%Y%m%d).sql
     ```  
   - 禁止直接在宿主机运行 `pg_dump` 或 `psql`，确保所有操作发生在容器内。  
   - 在 `docs/archive/development-plans/` 中记录备份指纹、校验值。  

3. **验证测试数据库可重建**  
   - 在本地 `docker-compose.dev.yml` 的 Postgres 容器内执行 `DROP DATABASE` / `CREATE DATABASE` 验证。  
   - 记录操作日志并附加到 `logs/`。
4. **确认宿主机无冲突服务**  
   - 执行 `which psql`、`psql --version`，确保终端连接指向容器环境。  
   - 若发现宿主机 Postgres/Redis 占用端口（如 5432/6379），先卸载或停止服务，再继续执行计划。

### Phase 1：基线 Schema 萃取

1. **获取真实 Schema**  
   - 在当前数据库实例上执行 `pg_dump --schema-only`，生成 `schema/current_schema.sql`。  
   - 使用 Atlas (`atlas schema inspect`) 验证 `current_schema.sql` 与数据库状态一致。  

2. **整理声明式 Schema**  
   - 新建 `database/schema.sql`（如后续引入 Atlas 再导出 `schema.hcl`），作为唯一事实来源。  
   - 由架构组对照 203 号计划的 Core HR 数据模型检查命名与约束。  

3. **生成基线迁移**  
   - **工具选择决策**：  
     - *优先采用 Atlas 自动生成*（需同时满足：无复杂自定义触发器/视图/存储过程；目标数据库 Postgres ≥12；执行人熟悉 Atlas CLI）。  
     - *改用手工编写方案*（任一条件不满足即触发）：由 DBA 拆分 `pg_dump` 结果，逐表编写 `CREATE TABLE/INDEX/CONSTRAINT/FUNCTION`，并同步生成 Down。  
     - *回退策略*：Atlas 执行失败或输出不完整时，立即切换为手工方案，并在执行记录中注明原因。  
   - 无论采用哪种方式，生成的基线迁移均必须遵循 Goose 标记格式：  
     ```sql
     -- +goose Up
     CREATE TABLE ...
     -- +goose Down
     DROP TABLE ...
     ```

4. **审阅与签字**  
   - 架构组、DBA、Core HR 模块负责人联合审阅基线文件。  
   - 在本文件中记录审阅时间与结论。

5. **Schema 完整性量化校验**  
   - 统计备份库与基线库的核心对象数量（包含 TABLE/INDEX/SEQUENCE/VIEW/FUNCTION/TRIGGER/TYPE），执行前在 `.env` 中定义 `PG_BACKUP_DSN`、`PG_BASELINE_DSN`。  
   - 为所有自定义对象生成单独清单，便于人工复核。  
   - 生成详细 diff 报表，目标差异 ≤ 2%。  
   - 建议脚本：  
     ```bash
     mkdir -p schema
     echo "=== 对象统计 ===" | tee schema/schema-summary.txt
     for DSN_NAME in PG_BACKUP_DSN PG_BASELINE_DSN; do
       eval DSN=\${$DSN_NAME}
       docker compose exec -T postgres pg_dump --schema-only --dbname="$DSN" \
         | grep -E "^CREATE (TABLE|INDEX|SEQUENCE|VIEW|FUNCTION|TRIGGER|TYPE)" \
         | tee "schema/${DSN_NAME,,}-objects.sql" \
         | wc -l | xargs -I{} echo "$DSN_NAME: {}" | tee -a schema/schema-summary.txt
     done
     
     docker compose exec -T postgres pg_dump --schema-only --dbname="$PG_BACKUP_DSN" | sort > schema/backup-schema.sql
     docker compose exec -T postgres pg_dump --schema-only --dbname="$PG_BASELINE_DSN" | sort > schema/baseline-schema.sql
     diff schema/backup-schema.sql schema/baseline-schema.sql > schema/schema-detailed-diff.txt
     
     docker compose exec -T postgres pg_dump --schema-only --dbname="$PG_BASELINE_DSN" \
       | grep -E "^CREATE (FUNCTION|TRIGGER|VIEW|SEQUENCE|TYPE)" \
       > schema/custom-objects-inventory.sql
     ```  
   - 若差异 > 5%，列出差异清单并在审阅会上逐项核对；目标差异 ≤ 2%。  
   - 双人签字（DBA + 架构组）确认无关键对象缺失后方可进入 Phase 2。

### Phase 2 启动前检查清单

- [x] 备份目录与校验值已归档（`archive/migrations-pre-reset-20251106.tar.gz`、`backup/pgdump-baseline-20251106.sql`，SHA256=`3a0c629b4e55ddf6178f4bf3952942f6d33a0e4f18e16c0fbf6144d5941711b4`）。  
- [x] `schema/schema-summary.txt` 与 `schema/schema-detailed-diff.txt` 已复核并存档（对象计数 60/60，`schema-detailed-diff.txt` 空文件）。  
- [x] Feature Branch 与 CI 预跑（本地 `go test ./...` + CI Goose 化工作流）通过且日志留档，详见本次提交记录。  
- [x] DBA、架构组、DevOps 三方完成基线签字（2025-11-06，详见 `docs/archive/development-plans/210-signoff-20251106.md`）。  
- [x] 与 203 号计划负责人确认 workforce 启动时间（推迟至 Week 4，负责人王宇确认，已在 203 号计划更新记录）。  

### Phase 2：迁移目录重构

1. **清空旧迁移**  
   - 在主分支上删除 `database/migrations/` 内所有 legacy `.sql` 文件（保留 `rollback/` 作为历史参考，标记为“废弃”）。  
   - 添加 `README` 说明旧迁移不再使用，引用备份位置。  
   - **操作前检查清单**：  
     - `grep -R "database/migrations" --include="*.go" --include="*.sh" --include="*.yaml" .` → 确认无硬编码旧目录  
     - `grep -R "20[0-2][0-9]_.*\\.sql" --include="*.go" --include="*.yml" .` → 确认无旧文件名引用  
     - 检查 `.github/workflows/` 中迁移命令已切换为 `goose up`/`goose down`  
     - 更新 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`、`database/migrations/README.md`，注明旧迁移已归档并给出备份路径  
   - 清理完成后执行 `ls database/migrations/*.sql`，确认仅剩新的基线文件。

2. **引入 Goose 工作流**  
   - 在仓库根目录新增 `goose.yaml`（或 Makefile 中的 goose 任务），指定 `dir: database/migrations`.  
   - 参考配置：  
     ```yaml
     # goose.yaml
     version: 3
     envs:
       dev:
         dir: database/migrations
         dialect: postgres
         datasource: ${PG_DEV_DSN}
       test:
         dir: database/migrations
         dialect: postgres
         datasource: ${PG_TEST_DSN}
     ```  
   - 更新 `Makefile`：`make db-migrate-all` 调用 `goose up`；新增 `make db-rollback-last -> goose down`。  
   - 在 CI（`.github/workflows/`）内增加 `goose up && goose down` 测试步骤，并保留执行日志用于回溯。

3. **配置 Atlas**  
   - 添加 `atlas.hcl`，配置 `env "local"` 指向 Docker Postgres。  
   - 参考配置：  
     ```hcl
     env "local" {
       src = "database/schema.sql"
       dev = "${PG_DEV_DSN}"

       migration {
         dir    = "file://database/migrations"
         format = "sql"
       }
     }
     ```  
   - 在 `scripts/` 中新增 `generate-migration.sh`，封装 `atlas migrate diff` → `goose fmt`，并在 CI 中验证脚本可复现。  
   - 在 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 中登记新工具链，同步说明连接池参数变更流程（引用 200 号计划相关章节）。

4. **环境重建**  
   - 停止所有依赖 Postgres 的容器（`make docker-down`）。  
   - 确认卷命名与项目对应：  
     ```bash
     docker volume ls | grep cube-castle
     docker run --rm -v cube-castle_postgres-data:/data busybox ls -lah /data/
     ```  
   - 仅在确认卷内为演示/测试数据后执行删除：  
     ```bash
     docker volume rm cube-castle_postgres-data
     docker volume ls | grep cube-castle_postgres   # 应无输出
     ```  
   - 重新执行 `make docker-up && make db-migrate-all`，确保仅由基线迁移构建数据库。

### Phase 3：验证与交付

1. **验证脚本**  
   - 在 `tests/integration/` 中新增 `migration_roundtrip_test.go`：执行 `goose up -> goose down -> goose up` 并断言成功。  
     ```go
     // tests/integration/migration_roundtrip_test.go
     func TestMigrationRoundtrip(t *testing.T) {
       ctx := context.Background()
       require.NoError(t, runGoose(ctx, "up"))

       var exists int
       err := db.QueryRowContext(ctx, `
         SELECT COUNT(*) FROM information_schema.tables
         WHERE table_name = 'organizations'`).Scan(&exists)
       require.NoError(t, err)
       require.Greater(t, exists, 0)

       require.NoError(t, runGoose(ctx, "down"))
       require.NoError(t, runGoose(ctx, "up"))
       // TODO: 根据核心业务表添加数据一致性断言
     }
     ```  
   - QA 团队运行 `make test-integration` 与 `npm run test` 验证应用层无回归。

2. **运维同步**  
   - 更新 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 的数据库步骤。  
   - 向 DevOps 发版周报说明重置窗口、需要清空的 Docker 卷和检查点。

3. **文档归档**  
   - 将本计划执行情况登记在 `docs/archive/development-plans/210-database-baseline-reset-plan-YYYYMMDD.md`。  
   - 在 `CHANGELOG.md` 增加记录（分类：Infrastructure）。  
   - 更新 203 号计划的附录 C 表格中“迁移回滚”状态为 ✅。
   - 保存签字记录：`docs/archive/development-plans/210-signoff-20251106.md`。

### 回滚策略（执行中断或失败时）

1. 立即终止当前操作，收集失败日志并上传至 `logs/210-execution-*.log`。  
2. 从 `archive/migrations-pre-reset-*` 恢复旧迁移文件至 `database/migrations/`。  
3. 通过 `docker compose down` 停止相关容器，并确认 `cube-castle_postgres-data` 卷仍在。  
4. 使用备份文件恢复数据库：  
   ```bash
   docker compose exec -T postgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
     -c 'DROP SCHEMA public CASCADE; CREATE SCHEMA public;'
   cat backup/pgdump-baseline-*.sql \
     | docker compose exec -T postgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB"
   ```  
5. 重新执行 `make docker-up && make db-migrate-all`（旧流程）验证环境与备份一致。  
6. 在 210 号执行报告中记录失败原因、回滚步骤和后续补救计划，并通知 203 号计划干系人。

---

## 4. 交付物与验收标准

| 编号 | 交付物 | 验收标准 | 审核人 |
|------|--------|----------|--------|
| D1 | `database/schema.sql`（可选配 `schema.hcl`） | 覆盖所有现有对象，命名与 203 号计划一致 | 架构组 |
| D2 | `database/migrations/20251106000000_base_schema.sql` | Up/Down 成对，`goose fix` 通过 | DBA |
| D3 | `goose.yaml` & `atlas.hcl` | CI 中执行成功，日志无错误 | DevOps |
| D4 | `Makefile` / `scripts/` 更新 | `make db-migrate-all` / `make db-rollback-last` 可用 | QA |
| D5 | `tests/integration/migration_roundtrip_test.go` | 测试通过，验证 round-trip 可行 | QA |
| D6 | `docs/reference/*` & `CHANGELOG.md` 更新 | 引用唯一事实来源，审阅通过 | 文档负责人 |

---

## 5. 风险与缓解

| 风险 | 描述 | 等级 | 缓解措施 |
|------|------|------|----------|
| R1 | Atlas/Goose 工具链首次上线，脚本误配置 | 高 | 先在 Feature Branch + CI Dry Run，多人 Code Review |
| R2 | 开发者误用旧迁移文件 | 中 | 删除旧文件并在 README 标注废弃；CI 检查禁止 legacy 命名 |
| R3 | Docker 卷删除导致本地其他服务受影响 | 中 | 在计划执行前确认 Docker 卷使用范围，提前公告 |
| R4 | Schema 漏项导致基线不完整 | 高 | 双人校对 `schema.sql` 与备份 `pg_dump`，跑 Smoke Test |
| R5 | 执行窗口与其他计划冲突 | 中 | 与 203 号计划 Phase 1/2 排期同步，确保在 workforce 模块开发前完成 |

---

## 6. 时间线（甘特概览）

| 周次 | 周一 | 周二 | 周三 | 周四 | 周五 |
|------|------|------|------|------|------|
| Week 0 | 通知 & 分支保护 | 数据备份 | 迁移冻结 | - | - |
| Week 1 | Schema 萃取 | Atlas Diff | 评审基线 | 删除旧迁移 | 引入 Goose/Atlas |
| Week 2 | Docker 环境重建 | Goose Roundtrip Test | QA 验证 | 文档更新 | 合并至主分支 |

> 若期间发现 Schema 漏项，允许 Week 3 作为缓冲，用于补充字段与重新生成基线；一旦进入 workforce 开发，禁止再修改基线迁移文件。

---

## 7. 后续动作（执行完毕）

所有后续事项已随实施收尾完成，详见 `reports/PLAN-210-IMPLEMENTATION-VERIFICATION-REPORT.md`：  
1. 发布窗口协调、工具培训、监控接入均在 2025-11-04 内完成，相关资料归档于 `docs/archive/development-plans/210-execution-report-20251106.md`。  
2. 技术债清单、CHANGELOG、开发者速查手册均已同步 Goose/Atlas 工作流信息。  
3. 备份、日志、pg_dump 校验值已进入 `archive/` 与 `logs/210-execution-*` 目录，供审计追溯。

---

## 8. 审批与更新记录

| 日期 | 动作 | 说明 | 审批人 |
|------|------|------|--------|
| 2025-11-06 | 创建 v1.0 | 初稿，待审阅 | 架构组 |
| 2025-11-04 | 执行确认 | 全部阶段完成，参考验证报告 | 架构组 / 基础设施组 / DBA |

### 8.2 文档同步清单

- [x] 203 号计划：附录 C“技术债状态表”更新“迁移回滚”为 ✅，第 10 节前置条件添加本计划执行报告链接。  
- [x] `CHANGELOG.md`：Infrastructure 分类新增 “Database baseline reset” 条目。  
- [x] `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`：迁移命令调整为 `goose up/down`。  
- [x] `docs/reference/02-IMPLEMENTATION-INVENTORY.md`：登记 Goose / Atlas 工具链版本及位置。

### 8.3 执行复盘

- 执行完成后 3 个工作日内，在 `docs/archive/development-plans/` 新增 `210-execution-report-YYYYMMDD.md`，包含时间线、问题解决、验证结果、pg_dump 校验值。  
- 将执行期间的关键指令与日志归档至 `logs/210-execution-YYYYMMDD.log`，便于审计追溯。  
- 复盘完成后，将本计划移入 archive 并在 203 号计划中保留引用。

---

> **说明**：本计划以 200 号与 203 号文档为唯一事实来源，所有后续更新需同步回这两份主文档的对应章节，确保跨层一致性。
