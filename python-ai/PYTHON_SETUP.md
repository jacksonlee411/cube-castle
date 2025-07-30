# Cube Castle AI Service - Python环境设置指南

## 📋 概述

Cube Castle AI Service是一个基于Python gRPC的智能服务，提供HR系统的自然语言处理和意图识别功能。

## 🛠️ 系统要求

- **Python**: 3.8+（推荐3.12）
- **Redis**: 6.0+（用于对话状态管理）
- **内存**: 最少512MB，推荐1GB+
- **网络**: 需要访问OpenAI API

## 🚀 快速开始

### 1. 安装依赖

```bash
# 克隆项目（如果需要）
cd /path/to/cube-castle/python-ai

# 运行安装脚本
./install.sh

# 或者安装开发环境
./install.sh --dev
```

### 2. 配置环境

创建或编辑 `.env` 文件：

```bash
# OpenAI配置
OPENAI_API_KEY=your_openai_api_key_here
OPENAI_API_BASE_URL=https://api.openai.com/v1

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379

# 服务配置（可选）
AI_PORT=50051
AI_HOST=0.0.0.0
LOG_LEVEL=INFO
```

### 3. 启动服务

```bash
# 生产模式
./start.sh

# 开发模式（详细日志）
./start.sh --dev

# 自定义端口
./start.sh --port 50052
```

### 4. 健康检查

```bash
# 检查服务状态
./health-check.sh

# 检查特定端口
./health-check.sh --port 50052
```

## 📦 依赖包详情

### 生产环境 (requirements.txt)

| 包名 | 版本 | 用途 |
|------|------|------|
| grpcio | ≥1.59.0 | gRPC核心库 |
| grpcio-tools | ≥1.59.0 | gRPC开发工具 |
| grpcio-health-checking | ≥1.59.0 | gRPC健康检查 |
| grpcio-status | ≥1.59.0 | gRPC状态码 |
| openai | ≥1.3.0 | OpenAI API客户端 |
| redis | 4.5.0-5.0.0 | Redis客户端 |
| python-dotenv | ≥1.0.0 | 环境变量加载 |
| fastapi | ≥0.104.0 | Web框架（管理接口） |
| structlog | ≥23.2.0 | 结构化日志 |
| pydantic | ≥2.0.0 | 数据验证 |

### 开发环境 (requirements-dev.txt)

包含生产环境所有依赖，外加：

- **测试框架**: pytest, pytest-asyncio, pytest-grpc
- **代码质量**: black, flake8, isort, mypy
- **开发工具**: ipython, jupyter
- **性能测试**: locust
- **安全扫描**: bandit, safety

## 🔧 故障排除

### 常见问题

#### 1. ModuleNotFoundError: No module named 'grpc_health'

**原因**: 缺少 `grpcio-health-checking` 包

**解决**:
```bash
source venv/bin/activate
pip install grpcio-health-checking>=1.59.0
```

#### 2. Redis连接失败

**原因**: Redis服务未启动或连接配置错误

**解决**:
```bash
# 启动Redis（Ubuntu/Debian）
sudo systemctl start redis-server

# 或使用Docker
docker run -d -p 6379:6379 redis:7-alpine

# 检查连接
redis-cli ping
```

#### 3. OpenAI API调用失败

**原因**: API密钥或基础URL配置错误

**解决**:
1. 检查 `.env` 文件中的配置
2. 验证API密钥有效性
3. 确认网络连接正常

#### 4. gRPC端口被占用

**原因**: 端口50051已被其他服务占用

**解决**:
```bash
# 检查端口占用
lsof -i :50051

# 使用其他端口
./start.sh --port 50052
```

### 日志分析

服务日志包含以下关键信息：

- `✅ OpenAI客户端初始化成功`: OpenAI连接正常
- `✅ DialogueStateManager initialized successfully`: Redis连接正常
- `✅ AI Service successfully started`: 服务启动成功
- `📥 Received signal`: 收到停止信号

## 🧪 测试验证

### 基础功能测试

```bash
# 激活虚拟环境
source venv/bin/activate

# 验证导入
python -c "
import grpc
from grpc_health.v1 import health
import openai
import redis
print('✅ 所有依赖导入正常')
"

# 运行测试套件（如果有）
python -m pytest tests/ -v
```

### 性能测试

```bash
# 使用内置性能测试
python comprehensive_performance_test.py

# 或使用locust（开发环境）
locust -f performance_test.py --host=http://localhost:50051
```

## 📈 监控和维护

### 健康检查端点

服务提供标准gRPC健康检查：

- **整体服务**: `service=""`
- **AI智能服务**: `service="intelligence"`

### 日志管理

- **生产环境**: 使用 `INFO` 级别
- **开发环境**: 使用 `DEBUG` 级别
- **日志格式**: 结构化JSON格式（通过structlog）

### 内存管理

- **Redis内存**: 监控Redis内存使用情况
- **Python内存**: 使用内存分析工具检查内存泄漏
- **对话历史**: 自动清理过期会话

## 🔐 安全建议

1. **API密钥**: 使用环境变量存储，不提交到版本控制
2. **网络安全**: 在生产环境中使用TLS加密
3. **依赖管理**: 定期更新依赖包，扫描安全漏洞
4. **访问控制**: 限制Redis和gRPC端口的网络访问

## 📞 技术支持

如遇到问题，请：

1. 检查日志文件 `ai.log`
2. 运行健康检查 `./health-check.sh`
3. 查看Redis连接状态
4. 验证环境变量配置

---

## 更新历史

### v1.1.0 (2024-07-27)
- ✅ 修复 `grpc-health-checking` 依赖问题
- ✅ 完善依赖版本管理
- ✅ 添加自动化安装和启动脚本
- ✅ 增强错误处理和日志记录
- ✅ 完善文档和故障排除指南