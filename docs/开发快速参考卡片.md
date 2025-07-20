# 🚀 Cube Castle 开发快速参考卡片

## 📋 常用命令

### 启动服务
```bash
# 启动Python AI服务
cd python-ai && source venv/bin/activate && python main.py

# 启动Go API服务
cd go-app && go run cmd/server/main.go

# 一键启动（使用脚本）
./start_complete.sh
```

### 测试服务
```bash
# 测试AI服务
./go-app/test_ai.sh

# 测试所有API路由
./go-app/test_all_routes.sh

# 测试特定端点
curl -X POST http://localhost:8080/api/v1/interpret \
  -H "Content-Type: application/json" \
  -d '{"query": "test", "user_id": "test-user"}'
```

### 环境检查
```bash
# 检查服务状态
netstat -tlnp | grep :8080  # Go服务
netstat -tlnp | grep :50051 # Python AI服务

# 检查进程
ps aux | grep python
ps aux | grep server
```

## ⚠️ 常见问题快速解决

### 1. Python命令不可用
```bash
python3 --version  # 检查Python3
alias python=python3  # 创建别名
```

### 2. gRPC连接失败
```bash
# 检查AI服务是否运行
ps aux | grep main.py
# 重启AI服务
cd python-ai && source venv/bin/activate && python main.py
```

### 3. Go模块错误
```bash
cd go-app
rm go.mod go.sum
go mod init github.com/gaogu/cube-castle/go-app
go mod tidy
```

### 4. 数据库连接失败
```bash
sudo systemctl status postgresql
sudo -u postgres psql -c "\l"
```

### 5. 路由404错误
- 检查路由注册顺序
- 确认中间件配置
- 验证Chi路由器设置

## 🔧 开发工具

### 代码生成
```bash
# 生成gRPC代码
protoc --go_out=. --go-grpc_out=. contracts/proto/intelligence.proto
```

### 依赖管理
```bash
# Go依赖
go mod tidy
go mod download

# Python依赖
pip install -r requirements.txt
```

### 代码格式化
```bash
# Go格式化
go fmt ./...

# Python格式化
black python-ai/
```

## 📊 服务端口

| 服务 | 端口 | 说明 |
|------|------|------|
| Go API服务 | 8080 | HTTP API服务 |
| Python AI服务 | 50051 | gRPC AI服务 |
| PostgreSQL | 5432 | 数据库服务 |

## 🎯 开发检查清单

### 新功能开发
- [ ] 创建功能分支
- [ ] 实现核心逻辑
- [ ] 添加错误处理
- [ ] 编写测试用例
- [ ] 更新文档
- [ ] 提交代码

### 问题排查
- [ ] 检查服务状态
- [ ] 查看错误日志
- [ ] 验证环境配置
- [ ] 测试服务通信
- [ ] 查阅相关文档

## 📚 重要文件

| 文件 | 用途 |
|------|------|
| `go-app/cmd/server/main.go` | Go服务主入口 |
| `python-ai/main.py` | Python AI服务 |
| `.env.example` | 环境变量模板 |
| `docker-compose.yml` | 容器编排配置 |
| `docs/开发问题总结与最佳实践.md` | 详细问题解决方案 |

## 🚨 紧急情况

### 服务完全无法启动
1. 检查所有依赖是否安装
2. 验证环境变量配置
3. 查看系统资源使用情况
4. 重启相关服务

### 数据丢失
1. 检查数据库连接
2. 查看数据库日志
3. 恢复备份数据
4. 验证数据完整性

---

**最后更新**: 2025年7月20日  
**版本**: v1.0 