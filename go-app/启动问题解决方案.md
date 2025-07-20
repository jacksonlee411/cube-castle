# 🚀 CoreHR API 启动问题解决方案

## 🚨 问题描述

在 WSL 环境下启动 Go 服务器时遇到以下问题：
1. Go 模块锁定错误：`go: RLock \\wsl.localhost\Ubuntu\home\shangmeilin\cube-castle\go-app\go.mod: Incorrect function.`
2. Go 版本兼容性问题
3. 代码语法错误

## ✅ 已实施的修复

### 1. 修复代码语法错误
- 修复了 `service.go` 文件中的语法错误
- 修正了缩进和括号匹配问题

### 2. 更新依赖版本
- 将 Go 版本从 1.23.0 降级到 1.21
- 更新所有依赖包到兼容版本：
  - `kin-openapi`: v0.120.0
  - `chi`: v5.0.10
  - `grpc`: v1.60.1
  - `protobuf`: v1.32.0

### 3. 创建多个启动脚本
- `quick_start.sh` - 快速启动脚本
- `start_minimal.sh` - 最小化启动脚本
- `final_fix.sh` - 完整修复脚本

## 🛠️ 解决方案

### 方案 1: 使用快速启动脚本（推荐）

```bash
# 在 WSL 终端中运行
cd /home/shangmeilin/cube-castle/go-app
chmod +x quick_start.sh
./quick_start.sh
```

### 方案 2: 手动修复

```bash
# 1. 进入项目目录
cd /home/shangmeilin/cube-castle/go-app

# 2. 删除 go.sum 文件
rm -f go.sum

# 3. 清理 Go 缓存
go clean -modcache

# 4. 重新初始化模块
go mod tidy

# 5. 设置环境变量
export APP_PORT=8080
export INTELLIGENCE_SERVICE_GRPC_TARGET=localhost:50051

# 6. 启动服务器
go run cmd/server/main.go
```

### 方案 3: 使用 Windows PowerShell

```powershell
# 在 Windows PowerShell 中运行
cd go-app
wsl bash -c "cd /home/shangmeilin/cube-castle/go-app && ./quick_start.sh"
```

## 🔧 故障排除

### 如果仍然遇到问题：

1. **检查 Go 版本**
   ```bash
   go version
   ```

2. **完全清理环境**
   ```bash
   rm -f go.sum
   rm -rf vendor/
   go clean -modcache
   go clean -cache
   go mod download
   go mod tidy
   ```

3. **验证模块**
   ```bash
   go mod verify
   ```

4. **编译测试**
   ```bash
   go build cmd/server/main.go
   ```

## 📋 验证步骤

启动成功后，验证以下端点：

1. **健康检查**: http://localhost:8080/health
2. **调试路由**: http://localhost:8080/debug/routes
3. **测试页面**: http://localhost:8080/test.html
4. **CoreHR API**: http://localhost:8080/api/v1/corehr/employees

## 🎯 最佳实践

1. **使用 WSL 终端**: 避免 Windows 路径问题
2. **定期清理缓存**: 防止模块锁定问题
3. **使用兼容版本**: 确保依赖包版本兼容
4. **使用启动脚本**: 避免手动设置环境变量

## 📞 获取帮助

如果问题仍然存在，请提供：
1. Go 版本信息
2. 完整的错误日志
3. 操作系统环境信息
4. 网络代理设置

## 🚀 快速启动命令

```bash
# 一键启动（推荐）
wsl bash -c "cd /home/shangmeilin/cube-castle/go-app && chmod +x quick_start.sh && ./quick_start.sh"
```

现在应该可以正常启动服务器了！ 