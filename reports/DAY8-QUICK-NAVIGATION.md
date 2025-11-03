# 📍 Day8 数据一致性验证 - 快速导航

**最后更新**：2025-11-03
**状态**：✅ 规范已制定，等待 Day8 执行

---

## 🎯 我需要什么？

### Day8 即将到来，我该做什么？

| 角色 | 任务 | 时间 | 详见 |
|------|------|------|------|
| **QA** | 执行数据一致性脚本 | Day8 上午 | [规范二](#规范二执行命令) |
| **后端** | 执行单元测试 + 集成测试 | Day8 上午 | [规范二](#规范二执行命令) |
| **前端** | 执行 Lint + 单元测试 | Day8 上午 | [规范二](#规范二执行命令) |
| **所有人** | 汇总结果并登记 | Day8 中午 | [规范三](#规范三产出登记) |
| **架构师** | 若FAIL，启动异常处理 | Day8 下午 | [规范五](#规范五异常处理) |

---

## 📚 完整文档地图

```
cube-castle/
├── 📄 reports/
│   ├── ⭐ DAY8-DATA-CONSISTENCY-VERIFICATION-SPEC.md  ← 【官方规范】
│   ├── 📋 DAY8-VERIFICATION-SUMMARY.md                 ← 【交付总结】
│   ├── 📌 THIS FILE (快速导航)                         ← 【你在这里】
│   ├── phase1-regression.md                           ← 【执行结果登记】
│   └── consistency/                                   ← 【脚本输出目录】
├── 📄 docs/
│   └── development-plans/
│       └── 06-integrated-teams-progress-log.md        ← 【Plan 211全景】
└── 🔧 scripts/
    ├── tests/test-data-consistency.sh                 ← 【巡检脚本】
    └── data-consistency-check.sql                     ← 【SQL查询】
```

---

## 🚀 快速开始（3分钟）

### 第1步：验证环境
```bash
# 启动Docker + 迁移数据库
make docker-up && make db-migrate-all && sleep 30

# 验证服务就绪
curl http://localhost:9090/health
curl http://localhost:8090/health
```

### 第2步：干运行测试
```bash
scripts/tests/test-data-consistency.sh --dry-run
```

### 第3步：真正执行
```bash
scripts/tests/test-data-consistency.sh
```

### 第4步：检查结果
```bash
# 查看摘要报告（最新的）
ls -lt reports/consistency/data-consistency-summary-*.md | head -1
cat $(ls -t reports/consistency/data-consistency-summary-*.md | head -1)
```

**预期**：看到 ✅ PASS 或 ❌ FAIL

---

## 📖 规范文档导航

### 🔍 我想了解...

| 我想了解... | 跳转位置 | 行号 |
|-----------|--------|------|
| 环境如何准备 | [规范一](#规范一环境准备清单) | 1.1-1.3 |
| 具体执行哪些命令 | [规范二](#规范二执行命令详解) | 2.1-2.3 |
| 脚本会产出什么 | [规范三](#规范三产出登记) | 3.1-3.4 |
| PASS/FAIL 的判定规则 | [规范四](#规范四判定标准) | 4.1-4.3 |
| 出现 FAIL 怎么办 | [规范五](#规范五异常处理) | 5.1-5.3 |
| Day8 要完成哪些任务 | [规范六](#规范六day8-执行清单) | 6.1-6.3 |
| 常见问题答案 | [规范八](#规范八faq) | 8.1-8.5 |
| 完整的执行脚本 | [附录](#附录完整执行脚本示例) | — |

---

## 📋 规范一：环境准备清单

**位置**：`DAY8-DATA-CONSISTENCY-VERIFICATION-SPEC.md:一、环境准备清单`

**核心检查**：
- [ ] Docker 容器已启动（`make docker-up`）
- [ ] 数据库已迁移（`make db-migrate-all`）
- [ ] 命令服务健康（`curl http://localhost:9090/health` → 200）
- [ ] 查询服务健康（`curl http://localhost:8090/health` → 200）
- [ ] 数据库连接正常（`psql ... -c "SELECT version();"`）

**数据库连接**（选择其中一种）：
1. **推荐**：`export DATABASE_URL="postgres://cube:castle@localhost:5432/cubecastle?sslmode=disable"`
2. **自动**：脚本自动从 `.env` 加载
3. **手动**：`export PGHOST=localhost PGUSER=cube PGPASSWORD=castle PGDATABASE=cubecastle`

---

## 📋 规范二：执行命令详解

**位置**：`DAY8-DATA-CONSISTENCY-VERIFICATION-SPEC.md:二、执行命令详解`

**三个关键命令**：

### 1️⃣ 干运行（推荐先执行）
```bash
scripts/tests/test-data-consistency.sh --dry-run
```
**作用**：验证环境与脚本可用性，不实际访问数据库

### 2️⃣ 正式执行
```bash
scripts/tests/test-data-consistency.sh
```
**产出**：
- `reports/consistency/data-consistency-<timestamp>.csv`
- `reports/consistency/data-consistency-summary-<timestamp>.md`

### 3️⃣ 同步回归
```bash
# 三条并行轨道
go test ./... && make test && npm run lint
```

**执行顺序建议**：
1. 干运行 → 2. 环境启动 → 3. 迁移 → 4. 检查健康 → 5. 并行执行 → 6. 检查结果

---

## 📋 规范三：产出登记

**位置**：`DAY8-DATA-CONSISTENCY-VERIFICATION-SPEC.md:三、产出登记与文件组织`

**脚本会生成**：
```
reports/consistency/
├── data-consistency-20251103T143022Z.csv        ← 原始数据（CSV）
└── data-consistency-summary-20251103T143022Z.md ← 摘要报告（Markdown）
```

**需要做**：
1. 打开 `reports/phase1-regression.md`
2. 在表格中新增一行，填写以下字段：
   - 执行时间（UTC）
   - 环境（Dev/Test/Staging）
   - 判定（✅ PASS 或 ❌ FAIL）
   - 异常数（格式 M/C/IP/DC）
   - 审计日志数
   - 关键结论
   - 附件（指向摘要报告）

**示例行**：
```
| 2025-11-03T14:30:22Z | Dev | ✅ PASS | 所有检查通过，审计日志12847条 | consistency/data-consistency-summary-20251103T143022Z.md |
```

---

## 📋 规范四：判定标准

**位置**：`DAY8-DATA-CONSISTENCY-VERIFICATION-SPEC.md:四、判定标准`

### ✅ PASS 条件（全部满足）
```
1️⃣ MULTIPLE_CURRENT 计数 = 0       （无多版本冲突）
2️⃣ TEMPORAL_OVERLAP 计数 = 0        （无时态重叠）
3️⃣ INVALID_PARENT 计数 = 0          （无孤立子节点）
4️⃣ DELETED_BUT_CURRENT 计数 = 0     （无删除冲突）
5️⃣ AUDIT_RECENT 值 > 0              （有审计日志）
```

### ❌ FAIL 条件（任一成立）
```
❌ 任何异常计数 > 0
❌ 审计日志为 0
❌ 脚本执行异常
```

---

## 📋 规范五：异常处理

**位置**：`DAY8-DATA-CONSISTENCY-VERIFICATION-SPEC.md:五、异常处理流程`

### 若出现 ❌ FAIL，按以下步骤处理：

#### **步骤1：收集信息**
```bash
# 保存输出文件
cp reports/consistency/data-consistency-*.csv /tmp/backup/
cp reports/consistency/data-consistency-summary-*.md /tmp/backup/
```

#### **步骤2：识别根因**
根据异常类型，运行对应的 SQL 查询：

**MULTIPLE_CURRENT > 0**（多版本冲突）
```sql
SELECT code, COUNT(*) FROM organization_units
WHERE is_current = TRUE GROUP BY code HAVING COUNT(*) > 1;
```

**TEMPORAL_OVERLAP > 0**（时态重叠）
```sql
SELECT u1.code, u1.effective_date, u1.end_date,
       u2.effective_date, u2.end_date
FROM organization_units u1 JOIN organization_units u2
WHERE daterange(...) && daterange(...);
```

**INVALID_PARENT > 0**（孤立子节点）
```sql
SELECT c.code, c.parent_code FROM organization_units c
LEFT JOIN organization_units p ON p.code = c.parent_code
WHERE c.parent_code IS NOT NULL AND p.code IS NULL;
```

**DELETED_BUT_CURRENT > 0**（删除冲突）
```sql
UPDATE organization_units SET is_current = FALSE
WHERE status = 'DELETED' AND is_current = TRUE;
```

#### **步骤3-5：修复→验证→更新**
- 制定修复方案（参见规范五、5.1）
- 重新执行脚本验证
- 在回归记录中补充修复过程

**详见**：[规范五完整流程](#规范五异常处理流程)

---

## 📋 规范六：Day8 执行清单

**位置**：`DAY8-DATA-CONSISTENCY-VERIFICATION-SPEC.md:六、Day8 执行清单`

### 时间表
```
09:00-09:15  准备阶段（Docker启动、迁移、等待服务）
09:15-09:30  健康检查（验证所有服务）
09:30-10:30  并行执行（3条轨道同时进行）
10:30-11:00  结果汇总（检查是否有异常）
11:00-12:00  登记更新（记录结果、准备交付）
```

### 待办清单
- [ ] Day7下午：准备环境、验证脚本
- [ ] Day8上午：启动→检查→执行
- [ ] Day8中午：汇总→判定→登记
- [ ] Day8下午：若FAIL则推进异常处理
- [ ] Day9-10：延伸测试与最终交付

---

## 📋 规范八：FAQ

**位置**：`DAY8-DATA-CONSISTENCY-VERIFICATION-SPEC.md:八、常见问题`

**Q1: 如何确认脚本的 SQL 逻辑正确？**
A: 查看 `scripts/data-consistency-check.sql` 的注释，或在 psql 中单独执行。

**Q2: 审计日志数量为 0 时是否一定代表问题？**
A: 初始化环境可能为0，但在生产环境中7天内为0需要排查。

**Q3: 多次执行脚本会生成多份报告吗？**
A: 是的，每次执行生成新的时间戳文件，便于追踪历史。

**Q4: 如果 FAIL 后迅速修复，需要重新执行吗？**
A: 是的，修复后都需重新执行以验证问题已解决。

**Q5: 修复脚本是否需要通过标准 PR 流程？**
A: 是的，任何数据库修改都需完整的审查流程。

---

## 📋 附录：完整执行脚本示例

**位置**：`DAY8-DATA-CONSISTENCY-VERIFICATION-SPEC.md:附录`

完整的 Day8 执行脚本（可直接运行）：

```bash
#!/bin/bash
set -euo pipefail

# Step 1: 环境准备
make docker-up && make db-migrate-all && sleep 30

# Step 2: 健康检查
curl http://localhost:9090/health && echo "✅ command OK"
curl http://localhost:8090/health && echo "✅ query OK"

# Step 3: 干运行验证
scripts/tests/test-data-consistency.sh --dry-run

# Step 4: 并行执行
(scripts/tests/test-data-consistency.sh) &
(go test ./... && make test) &
(npm run lint) &
wait

# Step 5: 检查结果
cat reports/consistency/data-consistency-summary-*.md | tail -1

echo "✅ 执行完成，请手动登记至 reports/phase1-regression.md"
```

---

## 🔗 相关文件一览

| 文件 | 用途 | 优先级 |
|------|------|--------|
| `DAY8-DATA-CONSISTENCY-VERIFICATION-SPEC.md` | 完整规范（官方） | ⭐⭐⭐ |
| `phase1-regression.md` | 执行结果登记 | ⭐⭐⭐ |
| `06-integrated-teams-progress-log.md` | Plan 211全景 | ⭐⭐ |
| `temporal-consistency-implementation-report.md` | 修复参考 | ⭐⭐ |
| `DAY8-VERIFICATION-SUMMARY.md` | 交付总结 | ⭐ |

---

## ✅ 检查清单

开始 Day8 前，确保以下项已完成：

```
准备阶段（Day7 下午）
  [ ] 阅读本文档
  [ ] 阅读 DAY8-DATA-CONSISTENCY-VERIFICATION-SPEC.md（第一章）
  [ ] 验证环境可用：make docker-up && make db-migrate-all
  [ ] 测试脚本：scripts/tests/test-data-consistency.sh --dry-run

执行阶段（Day8 上午）
  [ ] 启动 Docker 环境
  [ ] 执行迁移
  [ ] 健康检查通过
  [ ] 执行脚本并记录结果
  [ ] 检查摘要报告（PASS/FAIL）

登记阶段（Day8 中午）
  [ ] 在 phase1-regression.md 新增运行记录
  [ ] 若 FAIL，按规范五流程推进
  [ ] PR 提交与审核

延伸阶段（Day9-10）
  [ ] REST/GraphQL 对照测试
  [ ] E2E 核心流程验证
  [ ] 补充回归记录
  [ ] 准备复盘
```

---

## 📞 需要帮助？

| 问题 | 查看 |
|------|------|
| 环境配置问题 | 规范一、1.1-1.3 |
| 命令执行问题 | 规范二、2.1-2.3 |
| 结果解读问题 | 规范四、4.1-4.3 |
| 脚本失败问题 | 规范五、5.2 |
| 常见疑问 | 规范八、FAQ |

---

**版本**：v1.0 Quick Navigation
**最后更新**：2025-11-03
**状态**：✅ 准备就绪，等待 Day8 执行
