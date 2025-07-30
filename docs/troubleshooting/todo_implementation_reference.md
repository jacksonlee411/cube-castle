# TODO实现快速参考

**文档类型**: 开发参考  
**用途**: 实施过程中的快速查询和操作指南  
**最后更新**: 2025-07-29  
**Phase 1状态**: ✅ 已完成

## 🎯 核心目标文件

### ✅ Phase 1 已完成文件
```go
📁 go-app/internal/workflow/
├── employee_lifecycle_activities.go  ✅ 已完成 (4/4个函数)
    ├── CreateCandidateActivity()           ✅ 2025-07-29
    ├── InitializeOnboardingActivity()      ✅ 2025-07-29  
    ├── CompleteOnboardingStepActivity()    ✅ 2025-07-29
    └── FinalizeOnboardingActivity()        ✅ 2025-07-29
```

### 🔄 Phase 2 待实现文件
```go
📁 go-app/internal/workflow/
├── employee_lifecycle_activities.go  🟡 部分完成 (1个函数)
    └── UpdateEmployeeInformationActivity() 🔲 待实现
├── position_change_activities.go     🟡 重要 (2-3个函数)

📁 go-app/internal/service/
├── enhanced_temporal_query_service.go 🟢 查询
├── temporal_metrics_collector.go      🟢 监控

📁 go-app/cmd/server/
├── main.go                           🟡 启动优化

📁 go-app/internal/graphql/resolvers/
├── position_history_resolver.go      🟡 API层
```

## ⚡ 快速开始指南

### Step 1: 立即开始 (15分钟设置)
```bash
# 1. 定位核心文件
cd /home/shangmeilin/cube-castle/go-app/internal/workflow
grep -n "TODO" employee_lifecycle_activities.go

# 2. 查看现有结构
head -50 employee_lifecycle_activities.go

# 3. 运行现有测试
go test ./internal/workflow -v
```

### Step 2: 实现模板
```go
// 标准实现模板
func (a *EmployeeLifecycleActivities) YourActivity(
    ctx context.Context,
    req YourRequest,
) (*YourResult, error) {
    // 1. 日志记录
    a.logger.LogInfo(ctx, "开始执行活动", map[string]interface{}{
        "function": "YourActivity",
        "request": req,
    })
    
    // 2. 输入验证
    if err := validateRequest(req); err != nil {
        return nil, fmt.Errorf("请求验证失败: %w", err)
    }
    
    // 3. 核心业务逻辑
    result := &YourResult{}
    
    // TODO: 实现具体逻辑
    
    // 4. 日志记录结果
    a.logger.LogInfo(ctx, "活动执行完成", map[string]interface{}{
        "function": "YourActivity", 
        "result": result,
    })
    
    return result, nil
}
```

## 🔧 常用代码片段

### 数据库操作 (Ent ORM)
```go
// 创建记录
employee, err := a.entClient.Employee.
    Create().
    SetName(req.Name).
    SetEmail(req.Email).
    Save(ctx)

// 查询记录  
employee, err := a.entClient.Employee.
    Query().
    Where(employee.ID(req.EmployeeID)).
    Only(ctx)

// 更新记录
err := a.entClient.Employee.
    UpdateOneID(req.EmployeeID).
    SetStatus("ACTIVE").
    Exec(ctx)
```

### UUID生成
```go
import "github.com/google/uuid"

// 生成新ID
newID := uuid.New()

// 转换字符串
idString := newID.String()
```

### 错误处理
```go
// 标准错误处理
if err != nil {
    a.logger.LogError(ctx, "操作失败", map[string]interface{}{
        "error": err.Error(),
        "function": "YourActivity",
    })
    return nil, fmt.Errorf("操作失败: %w", err)
}
```

## 📋 验证检查清单

### 每个函数实现后检查
- [ ] **编译通过**: `go build ./internal/workflow`
- [ ] **测试通过**: `go test ./internal/workflow -run TestYourActivity`
- [ ] **日志完整**: 关键步骤有日志记录
- [ ] **错误处理**: 异常情况能正确处理
- [ ] **类型安全**: 没有类型转换错误
- [ ] **空值检查**: 防止空指针异常

### API验证
```bash
# 启动服务
go run cmd/server/main.go

# 测试API调用
curl -X POST http://localhost:8080/api/employees \
  -H "Content-Type: application/json" \
  -d '{"name":"测试员工","email":"test@example.com"}'
```

## 🐛 常见问题解决

### 问题1: Ent ORM使用
```go
// ❌ 错误方式
employee := &ent.Employee{
    Name: req.Name,
}

// ✅ 正确方式
employee, err := a.entClient.Employee.
    Create().
    SetName(req.Name).
    Save(ctx)
```

### 问题2: Temporal活动注册
```go
// 确保活动已注册 (检查 cmd/server/main.go)
worker.RegisterActivity(activities.CreateCandidateActivity)
```

### 问题3: 日志格式
```go
// ❌ 简单日志
log.Printf("创建员工: %s", name)

// ✅ 结构化日志
a.logger.LogInfo(ctx, "创建员工", map[string]interface{}{
    "employee_name": name,
    "operation": "create_employee",
})
```

## 🎯 实施优先级提醒

### 🔴 Phase 1 - 必须优先实现
1. `CreateCandidateActivity` - 候选人创建
2. `InitializeOnboardingActivity` - 入职初始化
3. `CompleteOnboardingStepActivity` - 步骤完成
4. `FinalizeOnboardingActivity` - 入职完成

### 🟡 Phase 2 - 核心功能
5. `UpdateEmployeeInformationActivity` - 信息更新
6. 选择1-2个职位变更函数

### 🟢 Phase 3 - 系统完善
7. 查询服务功能
8. 基础监控功能

## ⏰ 时间预估参考

### 单个函数实现时间
- **简单CRUD**: 2-4小时
- **带业务逻辑**: 4-6小时  
- **复杂流程**: 6-8小时
- **测试验证**: 1-2小时

### 每日进度预期
- **熟练开发**: 2-3个函数/天
- **学习过程**: 1-2个函数/天
- **包含测试**: 1个完整功能/天

## 📞 问题升级

### 遇到技术问题时
1. **检查现有代码**: 寻找类似实现模式
2. **查看测试用例**: 了解预期行为
3. **检查日志输出**: 确认执行路径
4. **分步验证**: 逐步确认每个环节

### 需要帮助的场景
- Ent ORM复杂查询
- Temporal工作流状态管理
- 错误处理策略
- 性能优化点

---

**使用提示**: 开发时将此文档保持打开，随时参考代码模板和检查清单