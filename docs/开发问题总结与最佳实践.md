# Cube Castle 开发问题总结与最佳实践

## 📋 项目概述

本文档总结了Cube Castle项目开发过程中遇到的关键问题、解决方案和最佳实践，旨在为后续开发提供经验参考，避免重蹈覆辙。

## 🚨 开发过程中遇到的问题总结

### 1. **环境配置问题**

#### 问题描述：
- WSL环境中Python命令不可用，需要使用`python3`
- 缺少Python虚拟环境和依赖包
- WSL代理配置问题影响网络连接

#### 解决方案：
```bash
# 创建虚拟环境
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

#### 注意事项：
- **WSL环境检查**：确认Python版本和命令别名
- **虚拟环境管理**：始终使用虚拟环境隔离依赖
- **网络代理**：WSL NAT模式下localhost代理配置问题

### 2. **Go模块和依赖问题**

#### 问题描述：
- Go模块路径配置错误
- 生成的gRPC代码路径混乱
- 依赖版本冲突

#### 解决方案：
```bash
# 清理并重新初始化Go模块
rm -rf go.sum go.mod
go mod init github.com/gaogu/cube-castle/go-app
go mod tidy
```

#### 注意事项：
- **模块路径规划**：提前规划好Go模块的命名空间
- **代码生成路径**：确保gRPC代码生成到正确位置
- **依赖管理**：定期运行`go mod tidy`清理无用依赖

### 3. **gRPC服务连接问题**

#### 问题描述：
- Go服务无法连接到Python gRPC服务
- 连接超时和拒绝连接错误
- 服务启动顺序问题

#### 解决方案：
```go
// 添加重试机制和超时配置
conn, err := grpc.Dial("localhost:50051", 
    grpc.WithInsecure(),
    grpc.WithBlock(),
    grpc.WithTimeout(5*time.Second))
```

#### 注意事项：
- **服务启动顺序**：先启动Python AI服务，再启动Go服务
- **端口冲突检查**：确保端口50051未被占用
- **连接超时设置**：合理设置gRPC连接超时时间

### 4. **HTTP路由注册问题**

#### 问题描述：
- Chi路由器路由注册冲突
- 404错误：`/api/v1/interpret`端点未找到
- 中间件配置问题

#### 解决方案：
```go
// 明确的路由注册顺序
r.Route("/api/v1", func(r chi.Router) {
    r.Post("/interpret", interpretHandler)
    r.Route("/corehr", func(r chi.Router) {
        // CoreHR routes
    })
})
```

#### 注意事项：
- **路由注册顺序**：确保路由注册的顺序正确
- **中间件应用**：注意中间件的应用范围和顺序
- **路由冲突检查**：避免重复注册相同路径

### 5. **AI服务集成问题**

#### 问题描述：
- OpenAI API连接失败
- 模型配置错误
- 环境变量读取问题

#### 解决方案：
```python
# 环境变量配置
import os
from dotenv import load_dotenv
load_dotenv()

api_key = os.getenv("OPENAI_API_KEY")
base_url = os.getenv("OPENAI_BASE_URL")
```

#### 注意事项：
- **环境变量管理**：使用`.env`文件管理敏感配置
- **API密钥安全**：不要硬编码API密钥
- **模型配置验证**：确认模型名称和API端点正确

### 6. **数据库连接问题**

#### 问题描述：
- 数据库连接字符串配置错误
- 表结构不匹配
- 事务处理问题

#### 解决方案：
```go
// 数据库连接配置
dsn := "host=localhost user=postgres password=password dbname=cube_castle port=5432 sslmode=disable"
db, err := sql.Open("postgres", dsn)
```

#### 注意事项：
- **连接字符串格式**：确保PostgreSQL连接字符串格式正确
- **数据库初始化**：提供数据库初始化脚本
- **连接池配置**：合理配置数据库连接池参数

## 📋 开发最佳实践总结

### 1. **项目结构规范**
```
project/
├── go-app/           # Go服务
├── python-ai/        # Python AI服务
├── contracts/        # API契约
├── docs/            # 文档
├── scripts/         # 脚本
└── docker-compose.yml
```

### 2. **环境管理**
- 使用Docker Compose统一管理服务
- 环境变量集中管理
- 虚拟环境隔离依赖

### 3. **服务通信**
- 使用gRPC进行服务间通信
- 实现健康检查机制
- 添加重试和超时机制

### 4. **错误处理**
- 统一的错误响应格式
- 详细的日志记录
- 优雅的错误恢复

### 5. **测试策略**
- 单元测试覆盖核心逻辑
- 集成测试验证服务通信
- 端到端测试验证完整流程

## 🛠️ 开发工具和脚本

### 1. **脚本开发规范** ⭐ **重要规则**
- **只使用Bash脚本**：项目中的所有脚本都使用Bash编写
- **不创建PowerShell脚本**：避免编码问题和跨平台兼容性问题
- **脚本命名**：使用`.sh`后缀，如`test_api.sh`、`start.sh`
- **编码格式**：使用UTF-8编码，确保在WSL/Linux环境中正常运行

#### 脚本开发原则：
```bash
#!/bin/bash
# 脚本头部必须包含shebang
set -e  # 遇到错误立即退出
set -u  # 使用未定义变量时报错

# 使用颜色输出提高可读性
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

echo -e "${GREEN}✅ 操作成功${NC}"
echo -e "${RED}❌ 操作失败${NC}"
```

### 2. **启动脚本**
```bash
#!/bin/bash
# 启动完整系统
cd python-ai && source venv/bin/activate && python main.py &
cd go-app && go run cmd/server/main.go
```

### 3. **测试脚本**
```bash
#!/bin/bash
# 测试API端点
curl -X POST http://localhost:8080/api/v1/interpret \
  -H "Content-Type: application/json" \
  -d '{"query": "test", "user_id": "test-user"}'
```

### 4. **调试脚本**
```bash
#!/bin/bash
# 检查服务状态
netstat -tlnp | grep :8080
netstat -tlnp | grep :50051
```

### 5. **验证脚本示例**
```bash
#!/bin/bash
# 验证实现状态的脚本示例
set -e

echo "🔍 开始验证实现状态..."

# 检查文件是否存在
if [ -f "internal/corehr/repository.go" ]; then
    echo "✅ Repository层文件存在"
else
    echo "❌ Repository层文件不存在"
    exit 1
fi

# 检查关键方法
if grep -q "CreateEmployee" internal/corehr/repository.go; then
    echo "✅ CreateEmployee方法已实现"
else
    echo "❌ CreateEmployee方法未实现"
fi

echo "🎉 验证完成！"
```

## 🚀 后续开发建议

### 1. **代码质量**
- 使用linter和formatter保持代码风格一致
- 添加单元测试提高代码覆盖率
- 使用CI/CD自动化构建和测试

### 2. **监控和日志**
- 集成结构化日志系统
- 添加性能监控和指标收集
- 实现分布式追踪

### 3. **安全性**
- 实现身份认证和授权
- 添加API限流和防护
- 定期更新依赖包

### 4. **可扩展性**
- 设计微服务架构
- 实现服务发现和负载均衡
- 添加缓存层提高性能

## ⚠️ 关键经验教训

1. **环境一致性**：确保开发、测试、生产环境的一致性
2. **配置管理**：集中管理配置，避免硬编码
3. **错误处理**：实现完善的错误处理和日志记录
4. **服务依赖**：明确服务启动顺序和依赖关系
5. **测试覆盖**：编写全面的测试用例
6. **文档维护**：及时更新技术文档和使用说明

## 🔧 常见问题快速解决

### 问题1：Python命令不可用
```bash
# 解决方案
python3 --version  # 检查Python3是否可用
alias python=python3  # 创建别名
```

### 问题2：gRPC连接失败
```bash
# 检查服务状态
ps aux | grep python
ps aux | grep server
netstat -tlnp | grep 50051
```

### 问题3：Go模块错误
```bash
# 重新初始化模块
cd go-app
rm go.mod go.sum
go mod init github.com/gaogu/cube-castle/go-app
go mod tidy
```

### 问题4：数据库连接失败
```bash
# 检查PostgreSQL服务
sudo systemctl status postgresql
sudo -u postgres psql -c "\l"  # 列出数据库
```

## 📚 相关文档链接

- [项目架构文档](./城堡蓝图.md)
- [API文档](./元合约v6.0.md)
- [部署指南](./Cube%20Castle%20项目%20-%20第二阶段工程蓝图.md)

---

**最后更新**: 2025年7月20日  
**版本**: v1.0  
**维护者**: Cube Castle开发团队 