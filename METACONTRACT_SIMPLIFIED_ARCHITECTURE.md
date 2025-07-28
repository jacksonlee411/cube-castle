# 元合约编译器简化架构说明

## 简化前后对比

### 🚨 简化前（过度设计）
```
cmd/metacontract-compiler/          # 独立CLI工具 ❌
docker-compose.editor-dev.yml       # 专用开发环境 ❌
start-visual-editor.sh              # 独立启动脚本 ❌
internal/localai/                   # AI增强功能 ❌
internal/intelligencegateway/       # 智能网关 ❌
internal/metacontracteditor/websocket.go  # 实时协作 ❌
generated/grpc/                     # gRPC服务 ❌
```

### ✅ 简化后（轻量级模块）
```
internal/metacontract/              # 核心编译逻辑 ✅
internal/codegen/                   # 代码生成器 ✅
internal/metacontracteditor/        # 简化的Web界面 ✅
cmd/server/main.go                  # 集成到主服务 ✅
```

## 新架构特点

### 1. 轻量级内部模块
- 元合约编译器作为城堡项目的一个内部模块
- 共享项目的基础设施（数据库、日志、认证等）
- 无需独立部署和维护

### 2. 简化的Web界面
- 移除WebSocket实时协作功能
- 使用RESTful API进行交互
- 集成到主服务的路由系统

### 3. 统一的开发环境
- 使用主项目的Docker配置
- 统一的日志和监控系统
- 共享的中间件和认证机制

## API端点

### 项目管理
```
GET    /api/v1/metacontract/projects          # 列出项目
POST   /api/v1/metacontract/projects          # 创建项目
GET    /api/v1/metacontract/projects/{id}     # 获取项目
PUT    /api/v1/metacontract/projects/{id}     # 更新项目
DELETE /api/v1/metacontract/projects/{id}     # 删除项目
POST   /api/v1/metacontract/projects/{id}/compile  # 编译项目
```

### 模板和设置
```
GET    /api/v1/metacontract/templates         # 获取模板
GET    /api/v1/metacontract/settings          # 用户设置
PUT    /api/v1/metacontract/settings          # 更新设置
```

## 核心组件

### 1. Compiler (`internal/metacontract/compiler.go`)
- 元合约解析和验证
- 代码生成协调
- 错误处理和报告

### 2. Code Generators (`internal/codegen/`)
- EntGenerator: 生成数据库schema
- APIGenerator: 生成API路由

### 3. Editor Service (`internal/metacontracteditor/service.go`)
- 项目管理业务逻辑
- 编译请求处理
- 用户设置管理

## 开发和部署

### 开发环境
```bash
# 启动整个项目（包含元合约编辑器）
cd go-app
go run cmd/server/main.go
```

### 访问界面
- 主服务: http://localhost:8080
- 元合约编辑器: http://localhost:8080/api/v1/metacontract/
- 健康检查: http://localhost:8080/health

### 编译测试
```bash
# 通过API编译元合约
curl -X POST http://localhost:8080/api/v1/metacontract/projects/123/compile \
  -H "Content-Type: application/json" \
  -d '{"content": "yaml content here", "preview": true}'
```

## 优势

1. **部署简单**: 不需要单独部署编译器服务
2. **维护成本低**: 减少配置文件和启动脚本
3. **资源效率**: 共享基础设施，减少资源占用
4. **开发体验**: 统一的开发环境和工具链
5. **集成度高**: 与主项目功能深度集成

## 迁移指南

### 如果之前使用独立CLI
```bash
# 之前
./metacontract-compiler -input contract.yaml -output ./generated

# 现在：通过Web API
curl -X POST /api/v1/metacontract/projects \
  -d '{"name": "my-contract", "content": "..."}'
curl -X POST /api/v1/metacontract/projects/{id}/compile
```

### 如果之前使用Docker Compose
```bash
# 之前
docker-compose -f docker-compose.editor-dev.yml up

# 现在：使用主项目配置
docker-compose up  # 使用主项目的docker-compose.yml
```

## 下一步

1. 实现数据库repository层
2. 完善Web界面处理函数
3. 添加前端React组件
4. 编写单元测试和集成测试