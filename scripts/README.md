# Scripts Directory - 标准化脚本管理

## 🚨 CLAUDE.md 第10条资源唯一性合规（最高优先级）

经过P1级别清理，脚本数量已从87个降至**5个核心脚本**，完全符合CLAUDE.md资源唯一性原则；任何新增脚本若破坏唯一事实来源或跨层一致性，均视为最高优先级阻断项。

## 📋 核心脚本列表 (仅5个)

### 1. 基础设施启动 
- **`start-infrastructure.sh`** - 启动数据库等基础设施
- 用途: Docker基础设施管理
- 使用: `./start-infrastructure.sh`

### 2. 开发环境启动
- **`dev-start-simple.sh`** - 简化版开发环境启动  
- 用途: 简化开发环境快速启动
- 使用: `./dev-start-simple.sh`
- 补充: Makefile `make run-dev` 的简化版本

### 3. 监控栈管理
- **`start-monitoring.sh`** - 启动Prometheus/Grafana/AlertManager
- 用途: 启动完整监控栈
- 使用: `./start-monitoring.sh` 或 `make monitoring-up`

### 4. 监控验证
- **`test-monitoring.sh`** - 验证监控栈运行状况
- 用途: 监控系统健康检查
- 使用: `./test-monitoring.sh` 或 `make monitoring-test`

### 5. 测试执行
- **`run-tests.sh`** - 综合测试执行
- 用途: 扩展Makefile测试功能的综合测试
- 使用: `./run-tests.sh`

## 📂 目录结构

```
scripts/
├── README.md                     # 本文档
├── start-infrastructure.sh       # 基础设施启动
├── dev-start-simple.sh          # 开发环境启动
├── start-monitoring.sh          # 监控栈启动
├── test-monitoring.sh           # 监控验证  
├── run-tests.sh                 # 测试执行
├── tests/                       # 所有测试脚本统一管理
├── ci/                          # CI/CD相关脚本
├── quality/                     # 代码质量相关脚本
└── codebase-maintenance/        # 代码库维护脚本
```

## 🎯 优先使用 Makefile

**强烈建议优先使用 Makefile** 而不是直接调用脚本：

```bash
# 推荐使用 Makefile
make run-dev          # 启动开发环境
make frontend-dev     # 启动前端开发  
make monitoring-up    # 启动监控栈
make test            # 运行测试

# 仅在Makefile不支持时使用脚本
./scripts/dev-start-simple.sh
```

## 🚨 脚本使用原则

### 资源唯一性原则（最高优先级）
- **绝对禁止**: 创建功能重复的启动脚本
- **严格控制**: 核心脚本数量保持在3-5个
- **定期清理**: 每月审查并清理过时脚本
- **一致性复核**: 发布脚本必须记录所依赖契约，确保与 Makefile、文档描述保持一致

### 命名规范
- **启动脚本**: `start-{purpose}.sh` 
- **测试脚本**: `test-{purpose}.sh`
- **工具脚本**: `{action}-{target}.sh`
- **禁止后缀**: `-final`, `-v2`, `-fix`, `-uuid` 等二义性后缀

### 维护要求
- **生命周期**: 临时脚本必须设定删除时间
- **文档要求**: 新脚本必须更新此README
- **审批要求**: 新增脚本需要明确业务场景和用户审批

## 📊 清理成果统计

- **清理前**: 87个脚本文件 (违规率1740%)
- **清理后**: 5个核心脚本 (完全合规)
- **清理内容**:
  - ✅ 删除 deprecated-scripts 目录 (14个过时脚本)
  - ✅ 删除重复启动脚本 (start.sh, quick_start.sh等)
  - ✅ 整理测试脚本到 tests/ 子目录
  - ✅ 删除功能重复的管理脚本
  - ✅ 备份所有原始脚本到 backup/script-cleanup-{timestamp}/

## 🔧 维护机制

### 自动化检查
- **每月执行**: `find . -name "*.sh" | wc -l` 检查脚本数量
- **违规警告**: 超过10个脚本时自动警告
- **强制清理**: 超过20个脚本时阻止新增

### 长期防护
- **代码审查**: 新增脚本必须通过严格代码审查
- **生命周期管理**: 临时脚本必须标注删除时间
- **定期审计**: 每季度全面审计脚本必要性

---

**维护负责人**: 开发团队  
**最后更新**: 2025-09-15  
**下次审查**: 2025-10-08

## 🛡️ 审计一致性门禁与本地校验（新增）

为解决“审计记录倍增/错配/触发器连锁”问题，新增标准化校验与门禁脚本，已接入 CI。

### 关键脚本
- `scripts/validate-audit-recordid-consistency.sql`（报告版）：
  - 输出汇总：`EMPTY_UPDATES`、`MISMATCHED_RECORD_ID`、`OU_TRIGGERS_PRESENT`
  - 列出错配样本与“changes 为空但 before!=after”的 UPDATE 样本
- `scripts/validate-audit-recordid-consistency-assert.sql`（断言版）：
  - 断言：空 UPDATE=0、recordId 与载荷一致；目标触发器不存在（审计/时态/软删标志四项）
- `scripts/apply-audit-fixes.sh`（一键执行）：
  - 报告版校验：默认执行
  - 断言版校验：设置 `ENFORCE=1` 启用
  - 修复/回填可选：`APPLY_FIXES=1` 先修复再校验；CI 默认 `APPLY_FIXES=0` 仅校验

### 本地等效命令

1) 仅校验（不改动数据）
```bash
export DATABASE_URL="postgres://user:password@localhost:5432/cubecastle?sslmode=disable"
ENFORCE=1 APPLY_FIXES=0 bash scripts/apply-audit-fixes.sh
```

2) 修复+校验（本地修复流程）
```bash
export DATABASE_URL="postgres://user:password@localhost:5432/cubecastle?sslmode=disable"
# 建议先应用关键迁移（仅值变更更新 + 移除目标触发器）
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f database/migrations/021_audit_and_temporal_sane_updates.sql
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f database/migrations/022_remove_db_triggers_and_functions.sql

ENFORCE=1 APPLY_FIXES=1 bash scripts/apply-audit-fixes.sh
```

### CI 工作流
- `.github/workflows/audit-consistency.yml`：
  - 应用 021→022 后，`ENFORCE=1 APPLY_FIXES=0` 执行强制校验
- `.github/workflows/consistency-guard.yml`：
  - 新增 `audit` 任务，流程同上

### 备注
- 可通过 `APP_ASSERT_TRIGGERS_ZERO=0` 暂时跳过“目标触发器为 0”的断言（例如执行 022 之前），仅用于过渡开发场景。
