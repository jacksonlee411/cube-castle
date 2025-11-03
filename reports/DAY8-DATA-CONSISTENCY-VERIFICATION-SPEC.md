# Day8 数据一致性验证规范（官方版）

**文档版本**：v1.0
**生效日期**：2025-11-03
**维护方**：QA & Architecture Team
**关联计划**：Plan 211 Phase1 & 06号全面评审

---

## 执行概览

本规范定义了 Plan 211 Phase1 **Day8** 的数据一致性验证流程，旨在确保模块统一化过程中的数据完整性、时态一致性与审计追溯。通过标准化脚本执行、产出登记、判定标准与异常处理，消除旧信息混淆，确保事实唯一来源。

**核心目标**：
- ✅ 验证命令写入 → 查询读取的数据一致性
- ✅ 检查时态时间表中是否存在重叠、多版本冲突、无效父节点
- ✅ 确保软删除记录不被标记为当前版本
- ✅ 追踪审计日志完整性（覆盖≥7日）

---

## 一、环境准备清单

### 1.1 前置条件

在执行脚本前，确保以下环境就绪：

| 项目 | 检查命令 | 预期结果 | 备注 |
|------|---------|---------|------|
| Docker 容器启动 | `make docker-up` | 返回成功，PostgreSQL/Redis/Temporal 容器运行中 | 必须容器化，不得在宿主机直接安装 |
| 数据库迁移 | `make db-migrate-all` | 返回成功，无报错 | 确保最新的数据库架构已加载 |
| 命令服务健康 | `curl http://localhost:9090/health` | HTTP 200 + JSON 响应 | command 服务就绪 |
| 查询服务健康 | `curl http://localhost:8090/health` | HTTP 200 + JSON 响应 | query 服务就绪 |
| 数据库连接 | `psql $DATABASE_URL -c "SELECT version();"` | 显示 PostgreSQL 版本信息 | 确保客户端可连接 |

### 1.2 数据库连接配置

根据环境选择以下方式之一配置连接参数：

**方式A：显式 DATABASE_URL（推荐）**
```bash
export DATABASE_URL="postgres://cube:castle@localhost:5432/cubecastle?sslmode=disable"
```

**方式B：从 .env 文件加载（自动）**
```bash
# 脚本会自动检测 .env，无需手动设置
# .env 中需包含 DATABASE_URL 或 PGHOST/PGUSER/PGPASSWORD/PGDATABASE
```

**方式C：个别 PostgreSQL 环境变量**
```bash
export PGHOST=localhost
export PGPORT=5432
export PGUSER=cube
export PGPASSWORD=castle
export PGDATABASE=cubecastle
```

### 1.3 依赖工具检查

```bash
# 检查 psql 可用
which psql

# 查看版本（建议 ≥ PostgreSQL 12）
psql --version

# 验证脚本文件存在
ls -l scripts/tests/test-data-consistency.sh scripts/data-consistency-check.sql
```

---

## 二、执行命令详解

### 2.1 干运行（推荐先执行）

在真正执行前，使用 `--dry-run` 演练流程，验证环境与脚本可用性：

```bash
# 基础干运行
scripts/tests/test-data-consistency.sh --dry-run

# 预期输出
# [data-consistency] Dry-run 模式：
#   - SQL 文件：.../scripts/data-consistency-check.sql
#   - 输出目录：.../reports/consistency
#   - 使用命令：psql (需要 DATABASE_URL / PGHOST 等连接信息)
#   - 产物：原始 CSV + Markdown 摘要
```

### 2.2 正式执行

满足前置条件后，执行真实的数据一致性巡检：

```bash
# 使用默认输出目录（reports/consistency）
scripts/tests/test-data-consistency.sh

# 或指定自定义输出目录
scripts/tests/test-data-consistency.sh --output /tmp/my-reports

# 或两者结合
DATABASE_URL="postgres://..." scripts/tests/test-data-consistency.sh --output /tmp/my-reports
```

**执行流程（由脚本内部完成）**：
1. 加载数据库连接参数（.env 或环境变量）
2. 校验 SQL 文件 `scripts/data-consistency-check.sql` 存在
3. 创建输出目录 `reports/consistency`
4. 通过 `psql` 执行 5 个 SQL 查询：
   - 检查多个 `is_current=true` 冲突
   - 检查时态区间重叠
   - 检查无效父节点
   - 检查软删除但仍标记为当前
   - 统计近 7 日审计日志
5. 生成原始 CSV 与 Markdown 摘要
6. 输出判定结果（✅ PASS 或 ❌ FAIL）

### 2.3 同步执行其他验证

Day8 不仅需要执行数据一致性脚本，还需同时执行基线回归以确保完整的构建链路无回归：

```bash
# 1. 编译与单元测试（Go）
go test ./... -count=1

# 2. 集成测试（若配置）
go test ./... -tags=integration -count=1

# 3. 通用测试命令
make test

# 4. 前端 lint 检查
npm run lint

# 5. 前端单元测试（如需）
cd frontend && npm run test -- --runInBand
```

**执行顺序建议**：
```bash
# Step 1: 干运行验证环境
scripts/tests/test-data-consistency.sh --dry-run

# Step 2: 启动容器化环境
make docker-up

# Step 3: 迁移数据库
make db-migrate-all

# Step 4: 等待服务启动（≈30s）
sleep 30

# Step 5: 验证服务健康
curl http://localhost:9090/health && echo "✅ command OK" || echo "❌ command FAIL"
curl http://localhost:8090/health && echo "✅ query OK" || echo "❌ query FAIL"

# Step 6: 并行执行验证
(scripts/tests/test-data-consistency.sh && echo "✅ Data consistency OK") &
(go test ./... && echo "✅ Tests OK") &
(npm run lint && echo "✅ Lint OK") &

# Step 7: 等待所有后台任务完成
wait

# Step 8: 检查结果
ls -la reports/consistency/data-consistency-*.md
```

---

## 三、产出登记与文件组织

### 3.1 脚本产出物

执行脚本后，以下文件将在 `reports/consistency/` 目录生成：

```
reports/consistency/
├── data-consistency-20251103T143022Z.csv          # 原始巡检结果（CSV 格式）
├── data-consistency-summary-20251103T143022Z.md   # 摘要报告（Markdown）
├── data-consistency-20251103T144511Z.csv
├── data-consistency-summary-20251103T144511Z.md
└── ...（后续执行产生的时间戳文件）
```

**文件命名规则**：
- 时间戳格式：`YYYYMMDDTHHmmSSZ`（UTC 时区，ISO 8601）
- 示例：`2025-11-03T14:30:22Z` → `20251103T143022Z`

### 3.2 CSV 文件格式

原始 CSV 文件包含 SQL 查询的直接输出，每行一条问题记录：

```csv
MULTIPLE_CURRENT,code_abc,2
TEMPORAL_OVERLAP,code_def,2025-01-01,2025-01-31,2025-01-15,2025-02-15
INVALID_PARENT,child_code_xyz,parent_code_missing
DELETED_BUT_CURRENT,code_del
AUDIT_RECENT,12847
```

**列含义**：
- 第1列：问题类别（MULTIPLE_CURRENT | TEMPORAL_OVERLAP | INVALID_PARENT | DELETED_BUT_CURRENT | AUDIT_RECENT）
- 第2-5列：问题详情（格式由问题类型决定）

### 3.3 Markdown 摘要报告

摘要报告为人类可读的格式，示例：

```markdown
# 数据一致性巡检报告 (20251103T143022Z)

| 检查项 | 异常数量 |
|--------|----------|
| 多个 is_current 版本冲突 | 0 |
| 时态区间重叠 | 0 |
| 无效父节点 | 0 |
| 软删除仍标记为当前 | 0 |
| 最近 7 天审计日志数 | 12847 |
| 结论 | ✅ PASS |

- SQL 来源：`scripts/data-consistency-check.sql`
- 原始输出：`data-consistency-20251103T143022Z.csv`
- 生成时间（UTC）：20251103T143022Z

如发现异常，请参考 `docs/architecture/temporal-consistency-implementation-report.md` 制定修复计划...
```

### 3.4 登记至 Phase1 回归记录

执行完成后，需立即将结果登记至 `reports/phase1-regression.md`：

**操作步骤**：

1. **定位表格**：打开 `reports/phase1-regression.md`，找到"运行记录"表格（约第30行）
2. **新增行**：在表格中添加一行记录：
   ```markdown
   | 2025-11-03T14:30:22Z | Dev | ✅ PASS | 所有一致性检查通过，审计日志12847条 | `data-consistency-summary-20251103T143022Z.md` |
   ```
3. **填写字段**：
   - **日期**：脚本生成的 UTC 时间戳
   - **环境**：执行环境（Dev/Test/Staging/Prod）
   - **判定**：✅ PASS 或 ❌ FAIL
   - **关键结论**：简述发现的问题或"无异常"
   - **附件**：指向摘要报告文件相对路径
4. **提交 PR**：将更新的 `phase1-regression.md` 作为 Day8 执行的证据提交

---

## 四、判定标准（PASS/FAIL）

### 4.1 PASS 的充要条件

数据一致性验证判定为 ✅ **PASS**，当且仅当以下条件**全部满足**：

1. **无多版本冲突**：`MULTIPLE_CURRENT` 计数 = 0
   - 含义：每个 code 最多只有一个 `is_current=true` 的记录
2. **无时态重叠**：`TEMPORAL_OVERLAP` 计数 = 0
   - 含义：同一 code 的不同版本时间区间不重叠
3. **无孤立子节点**：`INVALID_PARENT` 计数 = 0
   - 含义：所有子节点的父节点都存在且为当前版本
4. **无删除冲突**：`DELETED_BUT_CURRENT` 计数 = 0
   - 含义：软删除的记录不被标记为当前版本
5. **审计完整性**：`AUDIT_RECENT` 值 > 0  
   - 含义：过去 7 天内至少有 1 条审计日志  
   - **豁免说明**：若确认过去 7 天确无业务/操作事件触发审计日志，且已由 QA + 架构共同记录豁免结论，可视为满足此条件。豁免需在 `reports/phase1-regression.md` 及 06 号文档中注明。

**摘要报告输出**：
```
结论 | ✅ PASS
```

### 4.2 FAIL 的触发条件

数据一致性验证判定为 ❌ **FAIL**，当以下任一条件成立：

1. 任何异常计数 > 0：
   - `MULTIPLE_CURRENT` > 0
   - `TEMPORAL_OVERLAP` > 0
   - `INVALID_PARENT` > 0
   - `DELETED_BUT_CURRENT` > 0
2. 审计日志缺失：`AUDIT_RECENT` = 0（除非已有按 4.1 豁免说明备案）
3. 脚本执行异常（连接失败、SQL 错误）

**摘要报告输出**：
```
结论 | ❌ FAIL
```

**脚本退出码**：
- PASS：`exit 0`
- FAIL：`exit 2`

### 4.3 判定示例

**示例1：完全健康**
```
多个 is_current 版本冲突：0
时态区间重叠：0
无效父节点：0
软删除仍标记为当前：0
最近 7 天审计日志数：15234
结论：✅ PASS
```
→ 判定：**✅ PASS** ✓

**示例2：发现多版本冲突**
```
多个 is_current 版本冲突：3
时态区间重叠：0
无效父节点：0
软删除仍标记为当前：0
最近 7 天审计日志数：8001
结论：❌ FAIL
```
→ 判定：**❌ FAIL** ✗（需要修复）

**示例3：审计日志缺失**
```
多个 is_current 版本冲突：0
时态区间重叠：0
无效父节点：0
软删除仍标记为当前：0
最近 7 天审计日志数：0
结论：❌ FAIL
```
→ 判定：**❌ FAIL** ✗（需要补充审计数据或调查日志丢失原因）

---

## 五、异常处理流程

### 5.1 FAIL 场景处理

当数据一致性验证返回 ❌ **FAIL** 时，按以下流程处理：

#### 步骤1：收集问题信息

```bash
# 1. 保存原始 CSV 与摘要报告
cp reports/consistency/data-consistency-*.csv /tmp/backup/
cp reports/consistency/data-consistency-summary-*.md /tmp/backup/

# 2. 记录环境信息
echo "Environment snapshot:" > /tmp/issue-context.txt
uname -a >> /tmp/issue-context.txt
go version >> /tmp/issue-context.txt
docker ps >> /tmp/issue-context.txt

# 3. 导出问题详情（详见 5.2）
```

#### 步骤2：识别问题根因

根据异常类型，查询具体的问题记录：

**问题：MULTIPLE_CURRENT > 0**
```bash
# 查询多版本冲突的 codes
psql $DATABASE_URL -c "
  SELECT code, COUNT(*) as current_count
    FROM organization_units
   WHERE is_current = TRUE
   GROUP BY code
  HAVING COUNT(*) > 1
  ORDER BY current_count DESC;"

# 预期输出示例：
# code | current_count
# ----------+---------------
# ORG001   | 2
# ORG005   | 3
```

**问题：TEMPORAL_OVERLAP > 0**
```bash
# 查询时态重叠的详细记录
psql $DATABASE_URL -c "
  SELECT u1.code, u1.record_id, u1.effective_date, u1.end_date,
         u2.record_id, u2.effective_date, u2.end_date
    FROM organization_units u1
    JOIN organization_units u2
      ON u1.code = u2.code
     AND u1.record_id < u2.record_id
   WHERE daterange(u1.effective_date, COALESCE(u1.end_date, 'infinity'::date), '[]') &&
         daterange(u2.effective_date, COALESCE(u2.end_date, 'infinity'::date), '[]')
     AND NOT (u1.end_date = u2.effective_date OR u2.end_date = u1.effective_date);"
```

**问题：INVALID_PARENT > 0**
```bash
# 查询指向无效父节点的记录
psql $DATABASE_URL -c "
  SELECT c.code, c.parent_code, c.record_id
    FROM organization_units c
    LEFT JOIN organization_units p
      ON p.code = c.parent_code
     AND p.is_current = TRUE
   WHERE c.parent_code IS NOT NULL
     AND p.code IS NULL;"
```

**问题：DELETED_BUT_CURRENT > 0**
```bash
# 查询软删除但仍为当前的记录
psql $DATABASE_URL -c "
  SELECT code, status, is_current, updated_at
    FROM organization_units
   WHERE status = 'DELETED'
     AND is_current = TRUE
   ORDER BY updated_at DESC;"
```

#### 步骤3：制定修复方案

参考 `docs/architecture/temporal-consistency-implementation-report.md` 中的修复指南，根据问题类型选择相应修复脚本或代码更改。

**常见修复场景**：

| 问题 | 修复策略 | 参考文档 |
|------|--------|--------|
| MULTIPLE_CURRENT | 通过时间线重算将旧版本标记为非当前 | `temporal-consistency-implementation-report.md:阶段2` |
| TEMPORAL_OVERLAP | 调整 effective_date/end_date，确保时间不重叠 | `temporal-timeline-consistency-guide.md` |
| INVALID_PARENT | 更正 parent_code 或标记为孤立节点 | `organization-hierarchy-consistency.md` |
| DELETED_BUT_CURRENT | 将 is_current 标记为 FALSE | 见下方脚本示例 |

**修复脚本示例（DELETED_BUT_CURRENT）**：
```sql
-- 临时脚本（需经过完整 code review）
-- 标记软删除记录为非当前
UPDATE organization_units
   SET is_current = FALSE,
       updated_at = NOW()
 WHERE status = 'DELETED'
   AND is_current = TRUE;

-- 验证修复
SELECT COUNT(*) as remaining_issues
  FROM organization_units
 WHERE status = 'DELETED'
   AND is_current = TRUE;
```

> ⚠️ 任何数据库修改都需要：
> 1. 在本地环境先验证
> 2. 提交至 PR 供审查
> 3. 在 PR 描述中说明问题根因与修复理由
> 4. 经过架构师 + QA 批准后才能合并

#### 步骤4：重新执行验证

修复后，重新运行脚本验证问题是否已解决：

```bash
# 重新执行脚本
scripts/tests/test-data-consistency.sh --output reports/consistency

# 查看新的摘要报告
cat reports/consistency/data-consistency-summary-*.md | tail -1

# 验证结论
# 预期：✅ PASS
```

#### 步骤5：更新记录

在 `reports/phase1-regression.md` 中补充修复说明：

```markdown
| 日期 | 环境 | 判定 | 关键结论 | 附件 |
|------|------|------|----------|------|
| 2025-11-03T14:30Z | Dev | ❌ FAIL | 发现 MULTIPLE_CURRENT x3，已在 PR#2847 修复，见 `fix-temporal.md` | `data-consistency-summary-20251103T143022Z.md` |
| 2025-11-03T16:45Z | Dev | ✅ PASS | 修复后重新验证通过，所有异常清除 | `data-consistency-summary-20251103T164511Z.md` |
```

### 5.2 脚本执行失败的排查

若脚本本身执行异常（非数据问题），按以下步骤排查：

#### 错误：数据库连接失败

**症状**：
```
[data-consistency] 未检测到数据库连接信息 (DATABASE_URL 或 PGHOST)
```

**排查步骤**：
```bash
# 1. 检查环境变量
echo "DATABASE_URL=$DATABASE_URL"
echo "PGHOST=$PGHOST"

# 2. 检查 .env 文件
cat .env | grep -i database

# 3. 验证 Docker 容器运行
docker ps | grep postgres

# 4. 测试连接
psql -h localhost -U cube -d cubecastle -c "SELECT 1"

# 5. 若仍失败，查看容器日志
docker logs <postgres-container-id>
```

**解决方案**：
- 确保 `make docker-up` 已成功
- 设置 `DATABASE_URL` 或 PostgreSQL 环境变量
- 检查防火墙是否阻止 5432 端口

#### 错误：SQL 文件不存在

**症状**：
```
[data-consistency] 未找到 SQL 文件: .../scripts/data-consistency-check.sql
```

**排查步骤**：
```bash
# 1. 验证脚本文件存在
ls -l scripts/data-consistency-check.sql

# 2. 检查工作目录
pwd

# 3. 从项目根目录执行
cd /path/to/cube-castle
scripts/tests/test-data-consistency.sh
```

#### 错误：psql 命令不可用

**症状**：
```
[data-consistency] 未找到 psql，请确认已安装 PostgreSQL 客户端或设置 PSQL_BIN
```

**解决方案**：
```bash
# Linux（Ubuntu/Debian）
sudo apt install postgresql-client

# macOS
brew install postgresql

# 或显式设置 PSQL_BIN
export PSQL_BIN=/usr/lib/postgresql/15/bin/psql
```

#### 错误：权限不足

**症状**：
```
psql: error: FATAL:  password authentication failed for user "cube"
```

**解决方案**：
```bash
# 1. 验证 .env 中的凭证正确
cat .env | grep -E "(PGUSER|PGPASSWORD)"

# 2. 检查 Docker 容器中的用户创建脚本
docker exec <postgres-id> psql -U postgres -c "\du"

# 3. 若用户不存在，通过迁移补充
make db-migrate-all
```

### 5.3 部分异常的容忍策略

某些场景下，个别异常可能由外部原因引起，需要判断是否为真正的一致性问题：

| 异常 | 容忍条件 | 处理方式 |
|------|---------|--------|
| AUDIT_RECENT = 0 | 首次执行或数据库清空 | 允许在初始化环境，需在后续验证中补充审计日志 |
| TEMPORAL_OVERLAP（边界相接） | `end_date` = 下一版本的 `effective_date` | 无误，为连续时间线 |
| INVALID_PARENT（虚拟根） | 存在允许 `parent_code=NULL` 的根节点 | 符合设计，无需修复 |

> 若不确定异常的合理性，请在 PR 描述中详细说明，并邀请架构师审核。

---

## 六、Day8 执行清单

### 6.1 Day8 前一天（Day7 下午）

- [ ] 确认 Day6-7 架构审查已完成，共享代码抽取已合并至 `feature/204-phase1-unify`
- [ ] 验证环境可用：`make docker-up && make db-migrate-all`
- [ ] 测试脚本：`scripts/tests/test-data-consistency.sh --dry-run`
- [ ] 通知QA、测试人员Day8执行时间与责任分工

### 6.2 Day8 当天

**预计执行时间**：上午 09:00-12:00（3小时）

#### 09:00-09:15 准备阶段
- [ ] 所有人员到位
- [ ] 启动 Docker 环境：`make docker-up`
- [ ] 迁移数据库：`make db-migrate-all`
- [ ] 等待服务启动（≈30s）

#### 09:15-09:30 健康检查
- [ ] 验证命令服务：`curl http://localhost:9090/health`
- [ ] 验证查询服务：`curl http://localhost:8090/health`
- [ ] 验证数据库连接：`psql $DATABASE_URL -c "SELECT version();"`

#### 09:30-10:30 并行验证
- [ ] **QA线程**：执行数据一致性脚本
  ```bash
  scripts/tests/test-data-consistency.sh
  # 预期产出：CSV + 摘要报告
  # 预期结果：✅ PASS（或记录任何 ❌ FAIL）
  ```
- [ ] **后端线程**：执行单元 + 集成测试
  ```bash
  go test ./... -count=1
  ```
- [ ] **前端线程**：执行 lint + 单元测试
  ```bash
  npm run lint
  cd frontend && npm run test -- --runInBand
  ```

#### 10:30-11:00 结果汇总
- [ ] QA 收集所有执行日志
- [ ] 检查是否有 FAIL 或报错
- [ ] 若全部通过，准备登记记录

#### 11:00-12:00 登记与文档
- [ ] 在 `reports/phase1-regression.md` 中新增运行记录行
- [ ] 若有 FAIL，按 5.1 异常处理流程启动修复
- [ ] 生成 Day8 执行摘要并提交 PR
- [ ] 更新 06 号文档中的进度记录

### 6.3 Day8 后续（Day9-10）

- [ ] 对 Day8 验证结果进行 E2E + REST/GraphQL 对照测试
- [ ] 确认部署至测试环境后功能正常
- [ ] 在 `reports/phase1-regression.md` 中添加 Day9/10 延伸测试结果
- [ ] 准备 Day10 复盘与最终交付

---

## 七、事实来源与文档同步

本规范定义的所有文件路径、命令、判定标准均以以下源为唯一真源：

| 组件 | 唯一事实来源 | 维护者 |
|------|------------|--------|
| 脚本 | `scripts/tests/test-data-consistency.sh` | QA/DevOps |
| SQL | `scripts/data-consistency-check.sql` | Architecture |
| 回归模板 | `reports/phase1-regression.md` | QA |
| 本规范 | `reports/DAY8-DATA-CONSISTENCY-VERIFICATION-SPEC.md` | Architecture & QA |
| 参考文档 | `docs/architecture/temporal-consistency-implementation-report.md` | Architecture |
| 计划文档 | `docs/development-plans/06-integrated-teams-progress-log.md` | PM & Architecture |

**同步机制**：
- 若发现文档与实际脚本/SQL 不符，优先修正脚本/SQL，然后同步更新文档
- 每周检查事实来源一致性，记录在计划文档中

---

## 八、常见问题 (FAQ)

### Q1: 如何确认脚本的 SQL 逻辑正确？

**A**: 每个 SQL 查询都对应特定的一致性规则，详见 `scripts/data-consistency-check.sql` 中的注释。若需验证某条查询的逻辑，可在 psql 中单独执行：

```bash
psql $DATABASE_URL -f scripts/data-consistency-check.sql -q
```

### Q2: 审计日志数量为 0 时是否一定代表问题？

**A**: 在初始化环境或数据库清空后，审计日志可能为 0，这不一定是问题。但在生产或持续运行的环境中，若 7 天内为 0，说明要么审计机制失效，要么环境未正常工作。需要排查审计日志表是否包含数据。

### Q3: 多次执行脚本会生成多份报告吗？

**A**: 是的。每次执行生成新的时间戳文件（`data-consistency-20251103T143022Z.csv` 等），因此可以追踪多次执行的历史。建议定期清理旧文件：

```bash
find reports/consistency -mtime +30 -delete  # 删除 30 天前的报告
```

### Q4: 如果 FAIL 后迅速修复，需要重新执行吗？

**A**: 是的。每次修复后都需要重新执行脚本以验证问题已解决。在 `reports/phase1-regression.md` 中同时记录修复前后的两次执行结果。

### Q5: 修复脚本是否需要通过标准 PR 流程？

**A**: 是的。任何数据库修改（即使仅用于修复一致性问题）都必须：
1. 在本地完全验证
2. 通过 PR 供审查
3. 取得架构师 + QA 批准
4. 运行完整的测试套件

禁止直接在生产环境执行数据库修改脚本。

---

## 附录：完整执行示例

```bash
#!/bin/bash
set -euo pipefail

# ========== Day8 数据一致性验证完整流程 ==========

echo "=== Step 1: 环境准备 ==="
cd /path/to/cube-castle

echo "启动 Docker..."
make docker-up

echo "执行数据库迁移..."
make db-migrate-all

echo "等待服务启动..."
sleep 30

echo "=== Step 2: 健康检查 ==="
echo "检查命令服务..."
curl -s http://localhost:9090/health | jq . || echo "❌ command service unhealthy"

echo "检查查询服务..."
curl -s http://localhost:8090/health | jq . || echo "❌ query service unhealthy"

echo "检查数据库..."
export DATABASE_URL="${DATABASE_URL:-postgres://cube:castle@localhost:5432/cubecastle?sslmode=disable}"
psql $DATABASE_URL -c "SELECT version();" || echo "❌ database connection failed"

echo "=== Step 3: 干运行验证 ==="
scripts/tests/test-data-consistency.sh --dry-run

echo "=== Step 4: 并行执行验证 ==="

# 启动三个后台任务
(
  echo ">>> 数据一致性验证..."
  scripts/tests/test-data-consistency.sh
  echo "✅ 数据一致性验证完成"
) &
pid1=$!

(
  echo ">>> 后端测试..."
  go test ./... -count=1
  echo "✅ 后端测试完成"
) &
pid2=$!

(
  echo ">>> 前端 Lint..."
  npm run lint
  echo "✅ 前端 Lint 完成"
) &
pid3=$!

# 等待所有任务
wait $pid1 $pid2 $pid3
echo ""
echo "=== Step 5: 结果汇总 ==="
ls -lah reports/consistency/data-consistency-summary-*.md | tail -1
cat reports/consistency/data-consistency-summary-*.md | tail -1

echo ""
echo "=== Step 6: 记录登记 ==="
echo "请手动在 reports/phase1-regression.md 中新增以下行："
timestamp=$(date -u +%Y-%m-%dT%H:%M:%SZ)
echo "| $timestamp | Dev | ✅ PASS | 所有验证通过 | reports/consistency/data-consistency-summary-*.md |"

echo ""
echo "========== 验证完成 =========="
```

---

**文档变更历史**：
- v1.0 (2025-11-03)：初始发版，明确环境、命令、产出、判定、异常处理

