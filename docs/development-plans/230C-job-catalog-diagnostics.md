# Plan 230C – Job Catalog 自检脚本与 `make status` 集成

**文档编号**: 230C  
**母计划**: Plan 230 – Position CRUD 参考数据修复计划  
**前置计划**: 230B（数据修复完成后方可验证）  
**负责人**: DevEx / 平台工程

---

## 1. 背景与目标

- 219T/219E 的职位链路因 Job Catalog 缺失而被阻塞，缺口并未在常规 `make status` 中暴露。  
- 230C 将交付可复用的诊断脚本 `scripts/diagnostics/check-job-catalog.sh`，并在 `Makefile status` 目标中强制执行，一旦 `OPER` / 未来扩展的 Job Catalog 不符合预期即失败。  
- 目标是让开发者在跑服务前即可发现参考数据缺失，避免再次发生“Position CRUD 首步就 422”。

---

## 2. 范围

1. 新增 Bash 脚本：默认检查 `OPER` 栈，支持 `JOB_CATALOG_CODES=OPER,FINANCE` 扩展。  
2. 脚本逻辑：  
   - 使用 `docker compose ... exec postgres psql -Atq` 查询 `job_roles`、`job_levels` 等表；  
   - 若任一 code 不存在或 `status != 'ACTIVE'`，输出错误提示并 `exit 1`；  
   - 当环境缺少 Docker 或数据库未运行时，需要给出清晰提示并 `exit 2`。  
3. 在 `Makefile` 的 `status` 目标中新增该脚本调用，保持与现有健康检查顺序一致。  
4. 为脚本编写短 README/注释，说明如何扩展 `JOB_CATALOG_CODES`、如何在 CI 中使用。

---

## 3. 任务清单

| 步骤 | 描述 | 输出 |
| --- | --- | --- |
| C1 | 根据 230B 的迁移结果列出必须存在的实体（group/family/role/level），写入脚本常量或 JSON | `scripts/diagnostics/check-job-catalog.sh` |
| C2 | 实现脚本，示例伪码：<br>```bash
#!/usr/bin/env bash
set -euo pipefail
CODES=${JOB_CATALOG_CODES:-OPER}
for code in ${CODES//,/ }; do
  docker compose -f docker-compose.dev.yml exec -T postgres psql -U user -d cubecastle \
    -Atq "SELECT COUNT(*) FROM job_roles WHERE code LIKE '${code}% AND status='ACTIVE''" | ...
done
``` | 脚本 + 内嵌帮助信息 |
| C3 | 修改 `Makefile`：在 `status:` 规则里追加 `bash scripts/diagnostics/check-job-catalog.sh`，确保失败会终止整体命令 | `Makefile` diff |
| C4 | 记录脚本运行结果，附加到 `logs/230/job-catalog-check-YYYYMMDD.log`，并在 PR 说明如何使用 | 日志 + PR 描述 |

---

## 4. 依赖

- 230B 确保数据状态为 `ACTIVE`，以便脚本可以验证成功路径。  
- Docker Compose 环境必须按照 `AGENTS.md` 约束运行；脚本中不可直接访问宿主 PostgreSQL。  
- 需要读取 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 以获取 Job Catalog 结构。

---

## 5. 验收标准

1. `make status` 在数据缺失时会失败并给出“请执行 230B 迁移”提示；在数据完整时通过。  
2. 脚本可接受多个 `JOB_CATALOG_CODES`，并在 README/注释中示例如何调用。  
3. `logs/230/job-catalog-check-YYYYMMDD.log` 记录一次成功与一次失败（可通过手动删除记录或使用事务回滚模拟）以证明脚本有效。  
4. PR 包含脚本文件、Makefile 更新以及使用说明。  
5. 230C 的输出被 230D/219E 作为运行前置引用。

---

## 6. 交付记录（2025-11-08）

- **脚本与 Make 目标**：`scripts/diagnostics/check-job-catalog.sh` 现已被 `make status` 调用，脚本默认使用 Docker Compose 访问 `postgres` 容器，并支持 `JOB_CATALOG_CODES` / `JOB_CATALOG_LEVELS` 扩展。  
- **运行日志**：`logs/230/job-catalog-check-20251108T093645.log` 记录 `OPER` 检查通过；同日志可附在 PR/CI 供复验。  
- **配套说明**：脚本在 `docker compose ps` 或容器未就绪时会返回非零并输出指引，确保 230D/219E 在执行前即可发现数据缺口。

---

> 唯一事实来源：`scripts/diagnostics/check-job-catalog.sh`、`database/migrations/20251107123000_230_job_catalog_oper_fix.sql`、`logs/230/job-catalog-check-20251108T093645.log`。  
> 更新时间：2025-11-08。
