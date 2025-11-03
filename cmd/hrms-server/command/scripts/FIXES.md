# 已识别问题修复方案

## 问题1: audit_logs表结构不匹配

### 问题详情
SQL脚本中使用的字段名与实际数据库表结构不匹配：

| 脚本字段名 | 实际表字段名 | 类型差异 |
|------------|-------------|----------|
| `operation_type` | `event_type` | VARCHAR(20) |
| `business_entity_type` | `resource_type` | VARCHAR(50) |  
| `business_entity_id` | `resource_id` | UUID (而非字符串) |
| `operated_by` | `actor_id` | VARCHAR(100) |
| `changes_summary` | `business_context` | JSONB (而非纯文本) |

### 必须添加的字段
- `tenant_id` (UUID, NOT NULL) - 必须字段
- `actor_type` (VARCHAR, NOT NULL) - 必须字段，值为 'SYSTEM'

### 修复方案
更新audit_logs插入语句：

```sql
-- 原错误版本
INSERT INTO audit_logs (
    operation_type,
    business_entity_type, 
    business_entity_id,
    operated_by,
    operation_reason,
    changes_summary,
    created_at
) VALUES (...)

-- 修复版本  
INSERT INTO audit_logs (
    tenant_id,
    event_type,
    resource_type,
    actor_id,
    actor_type,
    action_name,
    request_id,
    operation_reason,
    business_context,
    timestamp
) VALUES (
    '00000000-0000-0000-0000-000000000000'::uuid,  -- 系统租户ID
    'UPDATE',
    'SYSTEM', 
    'DAILY_CUTOVER_SYSTEM',
    'SYSTEM',
    'TEMPORAL_MAINTENANCE',
    'daily-cutover-' || to_char(NOW(), 'YYYY-MM-DD-HH24-MI-SS'),
    '每日cutover任务：维护时态数据一致性',
    '{"task": "daily_cutover", "status": "completed"}'::jsonb,
    NOW()
);
```

### 影响的文件
- `/scripts/daily-cutover.sql` - 第98-114行
- `/scripts/data-consistency-check.sql` - 第267-287行
- Go代码中所有audit_logs插入操作

## 问题2: setup-cron.sh脚本语法错误

### 问题详情  
第185行包含损坏的中文字符，导致bash语法错误：
```
185:echo "M-fM-^IM-^KM-eM-^JM-(M-fM-5M-^KM-hM-/M-^UM-dM-;M-;M-eM-^JM-!:"^M$
```

这些是UTF-8编码损坏后的控制字符。

### 根本原因
- 文件编码问题或复制粘贴时的字符损坏
- 可能是在不同编辑器间转换时发生的编码问题

### 修复方案
替换损坏行：
```bash
# 原损坏行 (第185行)
echo "M-fM-^IM-^KM-eM-^JM-(M-fM-5M-^KM-hM-/M-^UM-dM-;M-;M-eM-^JM-!:"

# 修复为 (推测原意)
echo "手动测试任务:"
```

### 验证修复
```bash
# 验证语法
bash -n scripts/setup-cron.sh

# 清理所有控制字符
sed -i 's/\r$//' scripts/setup-cron.sh  # 移除Windows回车符
sed -i 's/[[:cntrl:]]//g' scripts/setup-cron.sh  # 移除控制字符
```

## 修复优先级

### 高优先级 (影响功能)
1. **audit_logs字段映射** - 导致审计日志记录失败
2. **bash脚本语法错误** - 导致cron设置脚本无法执行

### 中优先级 (影响体验) 
3. Go代码中audit_logs插入操作的一致性更新
4. 数据一致性检查脚本的类似问题修复

### 低优先级 (优化)
5. 统一所有脚本的日志记录格式
6. 添加更详细的错误处理

## 测试验证计划

### audit_logs修复验证
```sql
-- 测试修复后的插入语句
INSERT INTO audit_logs (
    tenant_id, event_type, resource_type, actor_id, actor_type,
    action_name, request_id, operation_reason, business_context, timestamp
) VALUES (
    '00000000-0000-0000-0000-000000000000'::uuid,
    'UPDATE', 'SYSTEM', 'TEST_SYSTEM', 'SYSTEM',
    'TEST_ACTION', 'test-' || extract(epoch from now()),
    '测试audit_logs修复', '{"test": true}'::jsonb, NOW()
);
```

### bash脚本修复验证
```bash
# 语法检查
bash -n scripts/setup-cron-fixed.sh

# 运行测试 (非root环境下的模拟测试)
DRY_RUN=true bash scripts/setup-cron-fixed.sh
```

## 预期修复结果

修复完成后应达到：
1. ✅ daily-cutover.sql 完全执行成功，包括audit_logs记录
2. ✅ setup-cron.sh 语法检查通过，可正常执行
3. ✅ 所有运维脚本的审计日志功能正常工作
4. ✅ E2E测试无错误或警告