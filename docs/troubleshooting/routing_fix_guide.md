# 🔧 路由修复说明

## 🚨 问题描述

访问以下路由时出现 404 错误：
- http://localhost:8080/debug/routes
- http://localhost:8080/test.html  
- http://localhost:8080/api/v1/corehr/employees

## 🔍 问题原因

1. **路由冲突**: OpenAPI 生成的路由被挂载到根路径 `/`，覆盖了手动注册的 CoreHR 路由
2. **静态文件路径错误**: test.html 文件路径不正确
3. **路由注册顺序问题**: 路由注册顺序导致冲突

## ✅ 已实施的修复

### 1. 修复路由冲突
```go
// 修复前
router.Mount("/", openapi.Handler(server))

// 修复后  
router.Mount("/api/v1", openapi.Handler(server))
```

### 2. 修复静态文件路径
```go
// 修复前
http.ServeFile(w, r, "test.html")

// 修复后
http.ServeFile(w, r, "../test.html")
```

### 3. 调整路由注册顺序
- CoreHR 路由先注册
- 静态文件服务其次
- OpenAPI 路由最后注册

## 🛠️ 修复后的路由结构

```
GET  /health                    - 健康检查
GET  /debug/routes             - 调试路由
GET  /test.html                - 测试页面
GET  /api/v1/corehr/employees  - 员工列表
POST /api/v1/corehr/employees  - 创建员工
GET  /api/v1/corehr/employees/{id} - 获取员工
PUT  /api/v1/corehr/employees/{id} - 更新员工
DELETE /api/v1/corehr/employees/{id} - 删除员工
GET  /api/v1/corehr/organizations - 组织列表
GET  /api/v1/corehr/organizations/tree - 组织树
```

## 🚀 启动方法

### 方法 1: 使用修复后的启动脚本
```bash
wsl bash -c "cd /home/shangmeilin/cube-castle/go-app && chmod +x start_fixed.sh && ./start_fixed.sh"
```

### 方法 2: 手动启动
```bash
cd go-app
rm -f go.sum
export APP_PORT=8080
export INTELLIGENCE_SERVICE_GRPC_TARGET=localhost:50051
go run cmd/server/main.go
```

## 🧪 测试方法

### 方法 1: 使用测试脚本
```bash
# Linux/macOS
chmod +x test_routes.sh
./test_routes.sh

# Windows PowerShell
.\test_routes.ps1
```

### 方法 2: 手动测试
```bash
# 健康检查
curl http://localhost:8080/health

# 调试路由
curl http://localhost:8080/debug/routes

# 测试页面
curl http://localhost:8080/test.html

# 员工列表
curl http://localhost:8080/api/v1/corehr/employees
```

## 📋 验证步骤

启动服务器后，请验证以下端点：

1. **健康检查**: http://localhost:8080/health
   - 预期: `{"status":"healthy"}`

2. **调试路由**: http://localhost:8080/debug/routes
   - 预期: 显示所有注册的路由列表

3. **测试页面**: http://localhost:8080/test.html
   - 预期: 显示 CoreHR API 测试页面

4. **员工列表**: http://localhost:8080/api/v1/corehr/employees
   - 预期: 返回员工列表 JSON 数据

## 🎯 预期结果

修复后，所有路由都应该正常响应：

- ✅ 健康检查返回 200 状态码
- ✅ 调试路由显示所有注册的端点
- ✅ 测试页面正常加载
- ✅ CoreHR API 返回正确的 JSON 数据

## 🔧 故障排除

如果仍然遇到问题：

1. **检查服务器日志**: 查看是否有错误信息
2. **验证端口占用**: 确保 8080 端口没有被其他程序占用
3. **检查文件权限**: 确保脚本有执行权限
4. **清理缓存**: 删除 go.sum 文件并重新启动

## 📞 获取帮助

如果问题仍然存在，请提供：
1. 服务器启动日志
2. 具体的错误信息
3. 浏览器开发者工具的网络请求信息 