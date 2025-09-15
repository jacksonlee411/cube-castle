# 最小回归用例：isTemporal 与 endDate 同步一致性

目标
- 验证全链重算在所有写路径后，`isTemporal` 必然与 `endDate` 一致：
  - endDate IS NULL  → isTemporal=false（尾部开放，当前或未来）
  - endDate IS NOT NULL → isTemporal=true（历史版本）

前置
- 后端已合入：重算路径同步维护 `is_temporal`；插入初值统一 `is_temporal=false`。
- 数据库执行过一次“对齐修复”脚本：`sql/maintenance/fix_is_temporal_alignment.sql`（可选）。

测试租户与样例 code
- 使用 `.env` 的 `DEFAULT_TENANT_ID`，新建 code：`TST9001`。

用例1：向前插入（中间版本）
1) 创建首条版本（2025-09-08）
   - 预期：endDate=NULL，isCurrent=true，isTemporal=false
2) 向前插入历史版本（2025-09-06）
   - 预期：
     - 2025-09-06：endDate=2025-09-07，isCurrent=false，isTemporal=true
     - 2025-09-08：endDate=NULL，isCurrent=true，isTemporal=false

用例2：修改生效日期（等价“删旧+插新”）
1) 将 2025-09-06 版改为 2025-09-05
   - 预期：
     - 新 2025-09-05：endDate=2025-09-07，isTemporal=true
     - 2025-09-08：endDate=NULL，isTemporal=false

用例3：停用与重新启用（即时/计划生效均可）
1) 即时停用（effectiveDate=今天）
   - 预期：
     - 新停用版本：若成当前，则其后的版本 endDate=NULL 且 isTemporal=false；历史均 isTemporal=true
2) 计划启用（effectiveDate=明天）
   - 预期：
     - 未来版本：endDate=NULL，isCurrent=false（未来），isTemporal=false

验证方式
- 数据库直接校验（推荐）
```sql
SELECT code, effective_date, end_date, is_current, is_temporal
FROM organization_units
WHERE code = 'TST9001'
ORDER BY effective_date;
```

- REST/GraphQL 响应校验（若查询层已改为派生 isTemporal）
  - REST 列表/详情：响应 `isTemporal` 与 DB 计算保持一致
  - GraphQL 查询：`endDate != null` → `isTemporal=true`

回归合格标准
- 所有尾部开放记录 `endDate IS NULL` 的 `isTemporal` 必为 false
- 所有历史记录 `endDate IS NOT NULL` 的 `isTemporal` 必为 true
- 写操作后无需额外修复脚本，重算即可保证一致性

