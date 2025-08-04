# 业务ID系统真实数据库测试数据创建方案

## 概述

为解决测试报告中指出的"测试环境限制（缺乏真实数据库）"问题，本方案将在真实PostgreSQL数据库中创建完整的测试数据，以提升测试覆盖率从42.7%到80%以上。

## 目标

1. **提升数据库相关函数测试覆盖率**：从0%提升到80%以上
2. **验证业务ID系统完整生命周期**：创建、查询、验证、生成
3. **测试真实数据库约束和关系**：外键、唯一性、索引
4. **压力测试业务ID生成性能**：并发生成、冲突处理

## 数据库环境分析

### 现有表结构
- **employees**: 员工表（带重复列问题）
- **organization_units**: 组织单元表
- **positions**: 职位表（带重复列问题）

### 业务ID设计
- **员工ID**: 1-99999 (5位数)
- **组织ID**: 100000-999999 (6位数)  
- **职位ID**: 1000000-9999999 (7位数)

## Phase 1: 数据库结构优化

### 1.1 添加业务ID字段

```sql
-- 为employees表添加business_id字段
ALTER TABLE employees ADD COLUMN IF NOT EXISTS business_id VARCHAR(5) UNIQUE;
CREATE INDEX IF NOT EXISTS idx_employees_business_id ON employees(business_id);

-- 为organization_units表添加business_id字段  
ALTER TABLE organization_units ADD COLUMN IF NOT EXISTS business_id VARCHAR(6) UNIQUE;
CREATE INDEX IF NOT EXISTS idx_organization_units_business_id ON organization_units(business_id);

-- 为positions表添加business_id字段
ALTER TABLE positions ADD COLUMN IF NOT EXISTS business_id VARCHAR(7) UNIQUE;
CREATE INDEX IF NOT EXISTS idx_positions_business_id ON positions(business_id);
```

### 1.2 创建业务ID序列

```sql
-- 创建业务ID生成序列
CREATE SEQUENCE IF NOT EXISTS employee_business_id_seq START 1 INCREMENT 1;
CREATE SEQUENCE IF NOT EXISTS org_business_id_seq START 1 INCREMENT 1;  
CREATE SEQUENCE IF NOT EXISTS position_business_id_seq START 1 INCREMENT 1;
```

## Phase 2: 测试数据创建策略

### 2.1 分层数据创建

#### 组织结构数据（100条）
```sql
-- 创建根组织单元
INSERT INTO organization_units (id, tenant_id, unit_type, name, description, parent_unit_id, 
                               status, level, employee_count, is_active, business_id, 
                               created_at, updated_at)
VALUES 
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000000', 'DEPARTMENT', '技术部',
     '负责产品技术研发', NULL, 'ACTIVE', 1, 0, true, '100000',
     NOW(), NOW()),
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000000', 'DEPARTMENT', '产品部', 
     '负责产品设计规划', NULL, 'ACTIVE', 1, 0, true, '100001',
     NOW(), NOW());
```

#### 职位数据（200条）
```sql
-- 创建技术类职位
INSERT INTO positions (id, tenant_id, position_type, title, code, job_profile_id,
                      department_id, status, budgeted_fte, business_id,
                      created_at, updated_at)
SELECT 
    gen_random_uuid(),
    '00000000-0000-0000-0000-000000000000',
    'TECHNICAL',
    CASE 
        WHEN s % 4 = 0 THEN '高级软件工程师'
        WHEN s % 4 = 1 THEN '软件工程师' 
        WHEN s % 4 = 2 THEN '测试工程师'
        ELSE '架构师'
    END,
    'POS' || LPAD(s::text, 4, '0'),
    gen_random_uuid(),
    (SELECT id FROM organization_units WHERE business_id = '100000' LIMIT 1),
    'ACTIVE',
    1.0,
    (1000000 + s)::varchar,
    NOW(),
    NOW()
FROM generate_series(0, 199) s;
```

#### 员工数据（1000条）
```sql
-- 创建员工数据
INSERT INTO employees (id, tenant_id, employee_number, employee_type, first_name, last_name,
                      email, position_id, status, hire_date, employment_status,
                      business_id, created_at, updated_at)
SELECT 
    gen_random_uuid(),
    '00000000-0000-0000-0000-000000000000',
    'EMP' || LPAD(s::text, 6, '0'),
    'FULL_TIME',
    CASE s % 10
        WHEN 0 THEN '张'
        WHEN 1 THEN '李'
        WHEN 2 THEN '王'
        WHEN 3 THEN '刘'
        WHEN 4 THEN '陈'
        WHEN 5 THEN '杨'
        WHEN 6 THEN '赵'
        WHEN 7 THEN '黄'
        WHEN 8 THEN '周'
        ELSE '吴'
    END || (s % 100 + 1)::text,
    CASE s % 5
        WHEN 0 THEN '伟'
        WHEN 1 THEN '芳'
        WHEN 2 THEN '娜'
        WHEN 3 THEN '秀英'
        ELSE '敏'
    END,
    'employee' || s || '@company.com',
    (SELECT id FROM positions WHERE business_id = (1000000 + (s % 200))::varchar LIMIT 1),
    CASE s % 10 WHEN 9 THEN 'INACTIVE' ELSE 'ACTIVE' END,
    CURRENT_DATE - (s % 2000)::int,
    'ACTIVE',
    (s + 1)::varchar,
    NOW(),
    NOW()
FROM generate_series(0, 999) s;
```

### 2.2 边界条件测试数据

#### 边界值数据
```sql
-- 员工ID边界值
INSERT INTO employees (id, tenant_id, employee_number, employee_type, first_name, last_name,
                      email, status, hire_date, employment_status, business_id, 
                      created_at, updated_at)
VALUES 
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000000', 'EMP_MIN', 'FULL_TIME',
     '边界', '测试1', 'boundary1@test.com', 'ACTIVE', CURRENT_DATE, 'ACTIVE', '1',
     NOW(), NOW()),
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000000', 'EMP_MAX', 'FULL_TIME', 
     '边界', '测试2', 'boundary2@test.com', 'ACTIVE', CURRENT_DATE, 'ACTIVE', '99999',
     NOW(), NOW());

-- 组织ID边界值
INSERT INTO organization_units (id, tenant_id, unit_type, name, status, level, 
                               employee_count, is_active, business_id, created_at, updated_at)
VALUES 
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000000', 'DEPARTMENT', 
     '边界组织最小', 'ACTIVE', 1, 0, true, '100000', NOW(), NOW()),
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000000', 'DEPARTMENT',
     '边界组织最大', 'ACTIVE', 1, 0, true, '999999', NOW(), NOW());

-- 职位ID边界值  
INSERT INTO positions (id, tenant_id, position_type, title, code, job_profile_id,
                      status, budgeted_fte, business_id, created_at, updated_at)
VALUES 
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000000', 'TECHNICAL',
     '边界职位最小', 'POS_MIN', gen_random_uuid(), 'ACTIVE', 1.0, '1000000',
     NOW(), NOW()),
     (gen_random_uuid(), '00000000-0000-0000-0000-000000000000', 'TECHNICAL',
     '边界职位最大', 'POS_MAX', gen_random_uuid(), 'ACTIVE', 1.0, '9999999', 
     NOW(), NOW());
```

## Phase 3: 测试用例扩展

### 3.1 数据库函数测试

#### LookupByBusinessID测试
```go
func TestBusinessIDService_LookupByBusinessID_WithRealDB(t *testing.T) {
    db := setupRealDBConnection(t)
    service := NewBusinessIDService(db)
    
    testCases := []struct {
        entityType EntityType
        businessID string
        expectFound bool
    }{
        {EntityTypeEmployee, "1", true},        // 边界最小值
        {EntityTypeEmployee, "99999", true},    // 边界最大值
        {EntityTypeEmployee, "100000", false},  // 超出范围
        {EntityTypeOrganization, "100000", true}, // 组织最小值
        {EntityTypePosition, "1000000", true},    // 职位最小值
    }
    
    for _, tc := range testCases {
        result, err := service.LookupByBusinessID(context.Background(), tc.entityType, tc.businessID)
        assert.NoError(t, err)
        assert.Equal(t, tc.expectFound, result.Found)
    }
}
```

#### GenerateBusinessID压力测试
```go
func TestBusinessIDService_GenerateBusinessID_Concurrent(t *testing.T) {
    db := setupRealDBConnection(t) 
    service := NewBusinessIDService(db)
    
    var wg sync.WaitGroup
    const numWorkers = 10
    const numPerWorker = 100
    
    results := make(chan string, numWorkers*numPerWorker)
    
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < numPerWorker; j++ {
                id, err := service.GenerateBusinessID(context.Background(), EntityTypeEmployee)
                assert.NoError(t, err)
                results <- id
            }
        }()
    }
    
    wg.Wait()
    close(results)
    
    // 验证所有生成的ID都是唯一的
    seen := make(map[string]bool)
    for id := range results {
        assert.False(t, seen[id], "发现重复的业务ID: %s", id)
        seen[id] = true
    }
}
```

### 3.2 真实数据库集成测试

#### 完整生命周期测试
```go
func TestBusinessIDSystem_FullLifecycle_WithRealDB(t *testing.T) {
    db := setupRealDBConnection(t)
    service := NewBusinessIDService(db)
    manager := NewBusinessIDManager(service, DefaultBusinessIDManagerConfig())
    
    // 1. 生成新的业务ID
    businessID, err := manager.GenerateUniqueBusinessID(context.Background(), EntityTypeEmployee)
    assert.NoError(t, err)
    assert.NotEmpty(t, businessID)
    
    // 2. 验证ID格式
    err = ValidateBusinessID(EntityTypeEmployee, businessID)
    assert.NoError(t, err)
    
    // 3. 创建员工记录
    employeeUUID := uuid.New()
    _, err = db.Exec(`
        INSERT INTO employees (id, tenant_id, employee_number, employee_type, 
                              first_name, last_name, email, status, hire_date, 
                              employment_status, business_id, created_at, updated_at) 
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
        employeeUUID, "00000000-0000-0000-0000-000000000000", "EMP_TEST", 
        "FULL_TIME", "测试", "员工", "test@company.com", "ACTIVE", 
        time.Now(), "ACTIVE", businessID, time.Now(), time.Now())
    assert.NoError(t, err)
    
    // 4. 通过业务ID查找UUID
    result, err := service.LookupByBusinessID(context.Background(), EntityTypeEmployee, businessID)
    assert.NoError(t, err)
    assert.True(t, result.Found)
    assert.Equal(t, employeeUUID, result.UUID)
    
    // 5. 通过UUID查找业务ID
    result2, err := service.LookupByUUID(context.Background(), EntityTypeEmployee, employeeUUID)
    assert.NoError(t, err)
    assert.True(t, result2.Found)
    assert.Equal(t, businessID, result2.BusinessID)
}
```

## Phase 4: 执行计划

### 4.1 脚本化执行

创建数据库初始化脚本：

```bash
#!/bin/bash
# setup_business_id_test_data.sh

echo "🚀 开始创建业务ID系统测试数据..."

# 1. 添加业务ID字段
echo "📋 添加业务ID字段..."
PGPASSWORD=password psql -h localhost -p 5432 -U user -d cubecastle -f add_business_id_fields.sql

# 2. 创建序列
echo "🔢 创建业务ID序列..."
PGPASSWORD=password psql -h localhost -p 5432 -U user -d cubecastle -f create_sequences.sql

# 3. 插入测试数据
echo "📊 插入测试数据..."
PGPASSWORD=password psql -h localhost -p 5432 -U user -d cubecastle -f insert_test_data.sql

# 4. 创建边界条件数据
echo "🎯 创建边界测试数据..."
PGPASSWORD=password psql -h localhost -p 5432 -U user -d cubecastle -f insert_boundary_data.sql

echo "✅ 测试数据创建完成！"
echo "📈 数据统计："
PGPASSWORD=password psql -h localhost -p 5432 -U user -d cubecastle -c "
SELECT 
    'employees' as table_name, 
    COUNT(*) as total_records,
    COUNT(business_id) as with_business_id
FROM employees
UNION ALL
SELECT 
    'organization_units', 
    COUNT(*), 
    COUNT(business_id) 
FROM organization_units
UNION ALL  
SELECT 
    'positions', 
    COUNT(*), 
    COUNT(business_id) 
FROM positions;"
```

### 4.2 测试执行

```bash
#!/bin/bash
# run_business_id_tests_with_real_db.sh

echo "🧪 开始真实数据库测试..."

# 设置测试环境变量
export TEST_WITH_REAL_DB=true
export DB_URL="postgres://user:password@localhost:5432/cubecastle"

# 运行扩展测试
echo "🔍 运行数据库相关函数测试..."
go test -v ./internal/common -run TestBusinessIDService.*WithRealDB -cover

echo "🔄 运行完整生命周期测试..."
go test -v ./internal/common -run TestBusinessIDSystem_FullLifecycle -cover

echo "⚡ 运行并发压力测试..."
go test -v ./internal/common -run TestBusinessIDService.*Concurrent -cover

echo "📊 生成覆盖率报告..."
go test -coverprofile=coverage_real_db.out ./internal/common
go tool cover -html=coverage_real_db.out -o coverage_real_db.html

echo "✅ 真实数据库测试完成！"
```

## Phase 5: 预期收益

### 5.1 覆盖率提升预期

| 函数名 | 当前覆盖率 | 预期覆盖率 | 提升幅度 |
|--------|-----------|-----------|----------|
| `LookupByBusinessID` | 0% | 90% | +90% |
| `LookupByUUID` | 0% | 90% | +90% |
| `GenerateBusinessID` | 47.4% | 85% | +37.6% |
| `HealthCheck` | 0% | 80% | +80% |
| `InitDatabase` | 0% | 75% | +75% |
| **总体覆盖率** | **42.7%** | **85%** | **+42.3%** |

### 5.2 质量提升

1. **真实约束验证** - 数据库级别的唯一性、外键约束
2. **性能基准测试** - 真实环境下的业务ID生成性能
3. **并发安全验证** - 多用户同时操作的数据一致性
4. **错误处理完善** - 真实数据库错误场景的处理

## Phase 6: 清理和维护

### 6.1 测试数据清理

```sql
-- 清理测试数据
DELETE FROM employees WHERE email LIKE '%@test.com' OR email LIKE '%@company.com';
DELETE FROM organization_units WHERE name LIKE '边界%' OR name LIKE '测试%';
DELETE FROM positions WHERE title LIKE '边界%' OR title LIKE '测试%';

-- 重置序列
ALTER SEQUENCE employee_business_id_seq RESTART 1;
ALTER SEQUENCE org_business_id_seq RESTART 1;
ALTER SEQUENCE position_business_id_seq RESTART 1;
```

### 6.2 持续集成

```yaml
# .github/workflows/test-with-real-db.yml
name: Real Database Tests
on:
  push:
    branches: [main, develop]
  
jobs:
  test-with-real-db:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:13
        env:
          POSTGRES_PASSWORD: password
          POSTGRES_USER: user
          POSTGRES_DB: cubecastle
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: 1.23
    
    - name: Setup test database
      run: ./scripts/setup_business_id_test_data.sh
      
    - name: Run real database tests
      run: ./scripts/run_business_id_tests_with_real_db.sh
      
    - name: Upload coverage reports
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage_real_db.out
```

## 总结

通过此方案，我们将：

1. **解决测试环境限制** - 建立真实数据库测试环境
2. **大幅提升覆盖率** - 从42.7%提升到85%以上
3. **验证真实场景** - 1000+条员工数据的真实业务场景
4. **建立CI/CD流程** - 自动化的真实环境测试
5. **提供性能基准** - 并发环境下的性能测试数据

这将彻底解决测试报告中指出的数据库环境限制问题，为业务ID系统提供完整的质量保障。