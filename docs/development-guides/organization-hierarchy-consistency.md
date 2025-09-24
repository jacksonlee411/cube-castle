# 组织层级一致性巡检与回收手册

本手册说明如何执行组织层级字段（`path`/`code_path`/`name_path`/`level`）的一致性巡检，以及在检测到异常后如何回收历史数据。

## 巡检脚本

- **脚本位置**：`scripts/maintenance/run-hierarchy-consistency-check.sh`
- **依赖**：`psql` 命令行客户端、`sql/hierarchy-consistency-check.sql`
- **输出**：将巡检结果导出为 `reports/hierarchy-consistency/hierarchy_anomalies_*.csv`

### 执行步骤

1. 设置数据库连接（任选其一）：
   ```bash
   export DATABASE_URL="postgres://user:password@host:5432/command_service"
   # 或者
   export PGHOST=localhost
   export PGPORT=5432
   export PGUSER=postgres
   export PGPASSWORD=secret
   export PGDATABASE=command_service
   ```
2. 运行巡检脚本：
   ```bash
   ./scripts/maintenance/run-hierarchy-consistency-check.sh
   ```
3. 若无异常，脚本直接退出；若检测到异常，脚本会生成 CSV 并返回非零退出码。

### CSV 字段说明

| 字段 | 含义 |
| --- | --- |
| `tenant_id` | 受影响租户 |
| `code` | 组织代码 |
| `parent_code` | 父组织代码（如有） |
| `level` | 当前记录的层级值 |
| `status` | 组织状态 |
| `anomaly_type` | 异常分类（如 `missing_code_path`、`parent_path_mismatch` 等） |
| `anomaly_detail` | 异常描述 |
| `path` / `code_path` / `name_path` | 当前存储的层级字段值 |

## CI 守卫

- **脚本位置**：`scripts/quality/hierarchy-consistency-guard.sh`
- **行为**：在 CI 或定时任务中运行，若检测到异常直接返回非零退出码以阻断流水线。
- **跳过条件**：环境未配置数据库连接或缺少 `psql` 时会打印警告并自动跳过。

建议在日常巡检或回归流水线中增加如下命令：

```bash
DATABASE_URL=... scripts/quality/hierarchy-consistency-guard.sh
```

## 异常回收流程

1. **导出异常列表**：使用巡检脚本生成最新的 CSV。
2. **确认影响范围**：人工核对 `anomaly_detail` 与上下文，区分历史数据与最新写路径。
3. **批量修复**：
   - 调用 `scripts/advanced-hierarchy-management.sql` 中的 `recalculate_hierarchy_cascade` 或相关函数，按租户和组织执行回收。
   - 如需一次性修复指定组织，可在 psql 中执行：
     ```sql
     SELECT recalculate_hierarchy_cascade('1000057', '<<tenant-uuid>>');
     ```
4. **复验**：修复后再次运行巡检脚本，确保 CSV 为空并删除旧文件。
5. **记录**：将修复摘要写入运营/事故记录，更新 `docs/development-plans/12-organization-update-inconsistency.md`。

## 常见问题

| 问题 | 处理建议 |
| --- | --- |
| `parent_missing` | 检查父组织是否被误删或尚未回填，必要时回滚或补齐父记录。 |
| `code_tail_mismatch` | 手工更新导致 `code_path` 尾段与 `code` 不一致，需重新回算层级。 |
| `depth_level_mismatch` | 可能源于重复迁移，执行 `recalculate_hierarchy_cascade` 后确认。 |

## 培训与评审要求

- Pull Request 涉及组织写入逻辑时，需说明如何确保调用 `ComputeHierarchyForNew`。
- 代码评审清单新增项目：**层级字段是否由仓储统一计算，是否提供巡检/复验结果**。
- 运维团队需在每个迭代收尾运行一次巡检脚本，并归档 CSV 结果。

如有新的修复工具或自动化 Pipeline，请同步至本手册并更新引用脚本路径。
