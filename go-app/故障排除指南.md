# CoreHR API 故障排除指南

## 🚨 常见问题及解决方案

### 1. Go 模块锁定错误

**错误信息：**
```
go: RLock \\wsl.localhost\Ubuntu\home\shangmeilin\cube-castle\go-app\go.mod: Incorrect function.
```

**解决方案：**

#### 方法 1: 使用修复脚本（推荐）
```bash
# Linux/macOS
./fix_go_modules.sh

# Windows PowerShell
.\fix_go_modules.ps1
```

#### 方法 2: 手动修复
```bash
# 1. 删除 go.sum 文件
rm go.sum

# 2. 清理 Go 模块缓存
go clean -modcache

# 3. 重新初始化模块
go mod tidy
```

#### 方法 3: 使用简化启动脚本
```bash
# Linux/macOS
./start_simple.sh

# Windows PowerShell
.\start_simple.ps1
```

### 2. Go 版本不兼容

**错误信息：**
```
go: go.mod requires go >= 1.23.0 but running go 1.21.x
```

**解决方案：**

#### 方法 1: 更新 Go 版本
```bash
# 下载并安装 Go 1.21 或更高版本
# 访问 https://golang.org/dl/
```

#### 方法 2: 修改 go.mod 文件（已修复）
go.mod 文件已修改为使用 Go 1.21，兼容性更好。

### 3. WSL 环境问题

**问题：** 在 WSL 环境下可能出现文件系统锁定问题。

**解决方案：**

#### 方法 1: 使用 WSL 终端
```bash
# 在 WSL 终端中运行
wsl
cd /home/shangmeilin/cube-castle/go-app
./start_simple.sh
```

#### 方法 2: 使用 Windows 终端
```powershell
# 在 Windows PowerShell 中运行
cd go-app
.\start_simple.ps1
```

### 4. 端口被占用

**错误信息：**
```
listen tcp :8080: bind: address already in use
```

**解决方案：**
```bash
# 查找占用端口的进程
netstat -ano | findstr :8080

# 终止进程（替换 PID 为实际进程 ID）
taskkill /PID <PID> /F

# 或者使用不同的端口
set APP_PORT=8081
go run cmd/server/main.go
```

### 5. 数据库连接失败

**错误信息：**
```
Failed to connect to databases
```

**解决方案：**

#### 方法 1: 使用 Mock 模式（推荐）
服务器会自动切换到 mock 模式，无需数据库连接。

#### 方法 2: 设置数据库环境变量
```bash
# 创建 .env 文件
echo "DATABASE_URL=postgresql://username:password@localhost:5432/dbname" > .env
echo "NEO4J_URI=bolt://localhost:7687" >> .env
echo "NEO4J_USER=neo4j" >> .env
echo "NEO4J_PASSWORD=password" >> .env
```

## 🛠️ 快速修复步骤

### 步骤 1: 清理环境
```bash
cd go-app
rm -f go.sum
go clean -modcache
```

### 步骤 2: 重新初始化
```bash
go mod tidy
go mod verify
```

### 步骤 3: 启动服务器
```bash
# 使用简化启动脚本
./start_simple.sh
```

## 📋 验证步骤

启动成功后，验证以下端点：

1. **健康检查**: http://localhost:8080/health
2. **调试路由**: http://localhost:8080/debug/routes
3. **测试页面**: http://localhost:8080/test.html
4. **CoreHR API**: http://localhost:8080/api/v1/corehr/employees

## 🔧 调试技巧

### 1. 检查 Go 环境
```bash
go version
go env
```

### 2. 检查模块状态
```bash
go mod why github.com/go-chi/chi/v5
go list -m all
```

### 3. 启用详细日志
```bash
go run -v cmd/server/main.go
```

### 4. 使用 Go 工作区（如果有多模块）
```bash
go work init
go work use .
```

## 📞 获取帮助

如果问题仍然存在，请提供以下信息：

1. **操作系统**: Windows/Linux/macOS
2. **Go 版本**: `go version`
3. **错误信息**: 完整的错误日志
4. **环境**: WSL/原生/虚拟机
5. **网络**: 是否有代理设置

## 🎯 最佳实践

1. **使用简化启动脚本**: 避免手动设置环境变量
2. **定期清理缓存**: 防止模块锁定问题
3. **使用 Mock 模式**: 开发阶段无需数据库
4. **检查端口占用**: 避免端口冲突
5. **使用 WSL 终端**: 避免 Windows 路径问题 