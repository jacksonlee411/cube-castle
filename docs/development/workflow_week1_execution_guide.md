# 工作流系统第一周开发任务执行指南

## 🎯 **第一周目标：建立核心基础设施**

基于已批准的工作流实施方案，完成数据模型建立和核心服务框架。

## 📋 **任务清单**

### **Day 1-2: 数据模型创建**

#### **任务1.1：创建Ent Schema文件**

1. **创建业务流程事件模型**
```bash
# 创建schema文件
touch go-app/ent/schema/business_process_event.go
```

2. **创建事务性发件箱模型**
```bash
touch go-app/ent/schema/outbox_event.go
```

3. **创建工作流实例模型**
```bash
touch go-app/ent/schema/workflow_instance.go
```

4. **创建工作流步骤模型**
```bash
touch go-app/ent/schema/workflow_step.go
```

#### **任务1.2：生成Ent代码和数据库迁移**
```bash
cd go-app
go generate ./ent
go run cmd/migrate/main.go
```

#### **任务1.3：验证数据库结构**
```bash
# 检查生成的SQL文件
cat ent/migrate/schema.sql

# 验证数据库连接和表创建
go run cmd/server/main.go
```

### **Day 3-4: 核心服务实现**

#### **任务2.1：业务流程事件服务**
创建文件：`go-app/internal/service/business_process_event_service.go`

核心功能：
- [x] PublishEvent() - 发布事件到数据库和发件箱
- [x] GetEventsByEntity() - 查询实体相关事件
- [x] 事务性保证

#### **任务2.2：事务性发件箱处理器**
创建文件：`go-app/internal/service/outbox_processor.go`

核心功能：
- [x] Start() - 启动后台处理线程
- [x] processUnprocessedEvents() - 批量处理未处理事件
- [x] 重试机制和错误处理

#### **任务2.3：工作流引擎基础框架**
创建文件：`go-app/internal/service/workflow_engine.go`

核心功能：
- [x] StartWorkflow() - 启动工作流实例
- [x] registerDefaultWorkflows() - 注册默认工作流定义
- [x] 员工入职和岗位变更工作流定义

### **Day 5: 服务集成和基础测试**

#### **任务3.1：服务集成到主应用**
修改文件：`go-app/cmd/server/main.go`

集成步骤：
1. 初始化新服务
2. 启动发件箱处理器
3. 添加优雅关闭处理

#### **任务3.2：基础单元测试**
创建测试文件：
- `go-app/internal/service/business_process_event_service_test.go`
- `go-app/internal/service/outbox_processor_test.go`
- `go-app/internal/service/workflow_engine_test.go`

测试覆盖：
- 基本CRUD操作
- 事务性验证
- 错误处理

## 🛠️ **具体实施步骤**

### **Step 1: 设置开发环境**
```bash
# 确保当前在正确的目录
cd /home/shangmeilin/cube-castle/go-app

# 检查现有依赖
go mod tidy

# 添加新依赖（如果需要）
go get github.com/google/uuid@latest
```

### **Step 2: 创建Schema文件**
按照实施方案中提供的完整代码创建所有schema文件。

### **Step 3: 生成和验证**
```bash
# 生成Ent代码
go generate ./ent

# 检查生成结果
ls -la ent/
ls -la ent/migrate/

# 运行数据库迁移
go run cmd/migrate/main.go
```

### **Step 4: 实现核心服务**
按照实施方案中的服务代码，实现三个核心服务类。

### **Step 5: 集成测试**
```bash
# 运行现有测试确保没有破坏
go test ./...

# 运行新的单元测试
go test ./internal/service/...

# 启动服务器验证集成
go run cmd/server/main.go
```

## ✅ **验收标准**

### **第一周结束时应该达到：**

1. **数据模型完成**
   - ✅ 4个新的Ent schema正确生成
   - ✅ 数据库迁移成功执行
   - ✅ 表结构符合设计要求

2. **核心服务实现**
   - ✅ BusinessProcessEventService可以发布事件
   - ✅ OutboxProcessor可以处理发件箱事件
   - ✅ WorkflowEngine可以启动基础工作流

3. **集成验证**
   - ✅ 服务正常启动无报错
   - ✅ 基础单元测试通过
   - ✅ 端到端事件流可以工作

4. **代码质量**
   - ✅ 符合现有代码规范
   - ✅ 适当的错误处理
   - ✅ 结构化日志记录

## 🚨 **注意事项**

1. **数据安全**
   - 确保所有新表都包含tenant_id字段
   - 验证RLS策略正确应用
   - 测试跨租户数据隔离

2. **性能考虑**
   - 为高频查询字段添加索引
   - 发件箱处理批量大小适中
   - 避免N+1查询问题

3. **向后兼容**
   - 不要修改现有API接口
   - 新增功能以增量方式实现
   - 保持现有测试通过

## 📞 **问题升级**

遇到以下情况时请及时反馈：
- 数据库迁移失败
- Ent代码生成错误
- 现有测试失败
- 性能问题

## 📝 **日报模板**

每日结束时更新进度：
```
# 第X天开发日报
## 完成的任务
- [ ] 任务描述

## 遇到的问题
- 问题描述及解决方案

## 明日计划
- [ ] 待完成任务

## 需要支持
- 需要的帮助或资源
```

## 🎯 **下周预览**

第一周完成后，第二周将重点开发：
- 工作流状态机逻辑
- 工作流步骤管理
- API控制器实现
- 更完善的测试覆盖

---
**创建时间**：2025-07-29  
**执行状态**：准备开始  
**预计完成**：2025-08-05