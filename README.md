# 🏰 Cube Castle - HR SaaS 平台

Cube Castle 是一个基于"城堡模型"架构的现代化 HR SaaS 平台，采用模块化单体设计，集成了人工智能驱动的自然语言交互能力。

## 🏗️ 架构概览

### 城堡模型 (Castle Model)

Cube Castle 采用独特的"城堡模型"架构，将整个系统构想为一个由以下部分组成的有机整体：

- **主堡 (The Keep)**: CoreHR 模块 - 核心人力资源管理功能
- **塔楼 (The Towers)**: 独立的功能模块
  - Intelligence Gateway Tower: AI 智能交互
  - Identity Access Tower: 用户认证授权
  - Tenancy Management Tower: 租户管理
- **城墙与门禁 (The Walls & Gates)**: 模块间的 API 接口

### 技术栈

- **后端**: Go 1.23
- **数据库**: PostgreSQL (记录系统) + Neo4j (洞察系统)
- **AI 服务**: Python + gRPC
- **API**: OpenAPI 3.0 + Chi Router
- **容器化**: Docker + Docker Compose

## 🚀 快速开始

### 环境要求

- Go 1.23+
- Python 3.12+
- Docker & Docker Compose
- PostgreSQL 16+
- Neo4j 5+

### 1. 克隆项目

```bash
git clone <repository-url>
cd cube-castle
```

### 2. 环境配置

```bash
# 复制环境变量模板
cp env.example .env

# 编辑环境变量（推荐使用 VSCode 或 vim）
vim .env
```

#### 关键环境变量说明
- `DATABASE_URL`：PostgreSQL 连接字符串，格式为 `postgresql://user:password@localhost:5432/cubecastle?sslmode=disable`
- `NEO4J_URI`、`NEO4J_USER`、`NEO4J_PASSWORD`：Neo4j 图数据库连接配置
- `INTELLIGENCE_SERVICE_GRPC_TARGET`：Python AI 服务 gRPC 地址，默认 `localhost:50051`
- `OPENAI_API_KEY`、`OPENAI_API_BASE_URL`：如需调用 OpenAI，可在此配置密钥和 API 地址
- `APP_PORT`：Go 主服务监听端口，默认 8080
- `JWT_SECRET`：JWT 签名密钥，务必妥善保管
- 其余变量详见 `env.example`

> **安全建议：** 切勿将 `.env` 文件提交到 Git 仓库，API 密钥和 JWT 密钥请妥善保管。

### 3. 依赖安装与虚拟环境

#### Python 依赖

```bash
cd python-ai
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

#### Go 依赖

```bash
cd go-app
go mod tidy
```

### 4. 常见问题排查

- **Python 依赖未安装/找不到 grpc 等模块**：请确保已激活虚拟环境并执行 `pip install -r requirements.txt`
- **Go 依赖报错/go.mod not found**：请在 go-app 目录下执行 `go mod tidy`
- **端口占用/服务无法启动**：检查 8080/50051 端口是否被占用，可用 `lsof -i:8080` 查找并 kill 进程
- **数据库连接失败**：确认 Docker 中的 PostgreSQL/Neo4j 已启动，且 .env 配置正确
- **WSL/VSCode 环境问题**：推荐使用 WSL2 + VSCode Remote，确保文件权限和路径一致

### 5. 启动基础设施

```bash
# 启动数据库服务
docker-compose up -d postgres neo4j

# 等待服务启动完成
docker-compose ps
```

### 6. 初始化数据库

```bash
# 进入 Go 应用目录
cd go-app

# 运行数据库初始化
go run cmd/server/main.go init-db
```

### 7. 启动服务

```bash
# 启动 Python AI 服务
cd python-ai
python main.py

# 新终端启动 Go 主服务
cd go-app
go run cmd/server/main.go
```

## 📁 项目结构

```
cube-castle/
├── contracts/                 # API 合约定义
│   ├── openapi.yaml          # OpenAPI 规范
│   └── proto/                # gRPC 协议定义
│       └── intelligence.proto
├── go-app/                   # Go 主应用
│   ├── cmd/server/           # 应用入口
│   ├── internal/             # 内部模块
│   │   ├── common/           # 通用组件
│   │   ├── corehr/           # 核心 HR 模块
│   │   └── intelligencegateway/ # 智能网关模块
│   ├── generated/            # 生成的代码
│   └── scripts/              # 数据库脚本
├── python-ai/                # Python AI 服务
│   ├── main.py              # AI 服务入口
│   └── requirements.txt     # Python 依赖
├── docs/                     # 项目文档
├── docker-compose.yml        # 容器编排
└── README.md                # 项目说明
```

## 🔧 核心功能

### 1. 员工管理 (CoreHR)

- 员工信息管理
- 组织架构管理
- 职位管理
- 汇报关系管理

### 2. 智能交互 (Intelligence Gateway)

- 自然语言理解
- 意图识别
- 实体提取
- 智能对话

### 3. 多租户支持

- 租户隔离
- 配置管理
- 权限控制

## 🛠️ 开发指南

### 脚本开发规范 ⭐ **重要规则**

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

### API 开发

1. **定义 API 合约**: 在 `contracts/openapi.yaml` 中定义 API 规范
2. **生成代码**: 使用 oapi-codegen 生成 Go 代码
3. **实现接口**: 在对应的模块中实现 API 接口

### 数据库操作

1. **添加表结构**: 在 `go-app/scripts/init-db.sql` 中添加表定义
2. **创建模型**: 在模块的 `models.go` 中定义数据结构
3. **实现 Repository**: 在 `repository.go` 中实现数据访问逻辑

### AI 功能扩展

1. **定义意图**: 在 Python AI 服务中添加新的意图定义
2. **实现处理逻辑**: 在 Go 服务中添加对应的业务逻辑
3. **更新合约**: 同步更新 gRPC 协议定义

## 🧪 测试

```bash
# 运行单元测试
cd go-app
go test ./...

# 运行集成测试
go test -tags=integration ./...

# 运行 API 测试
go test ./cmd/server/...
```

## 📊 监控与日志

### 健康检查

```bash
# 检查服务健康状态
curl http://localhost:8080/health

# 检查数据库连接
curl http://localhost:8080/health/db
```

### 日志查看

```bash
# 查看 Go 服务日志
docker-compose logs -f go-app

# 查看 AI 服务日志
docker-compose logs -f python-ai
```

## 🛡️ 安全与最佳实践

- 所有敏感信息（API 密钥、JWT 密钥等）请仅配置在 `.env` 文件中，切勿硬编码或提交到仓库
- 推荐定期更换密钥，生产环境请使用更强的密码策略
- 数据库、AI 服务等均需配置访问控制，避免外部未授权访问

## 📈 部署

### 开发环境

```bash
# 使用 Docker Compose 启动完整环境
docker-compose up -d

# 查看服务状态
docker-compose ps
```

### 生产环境

1. **构建镜像**:
   ```bash
   docker build -t cube-castle:latest .
   ```

2. **配置环境变量**:
   ```bash
   export DATABASE_URL="postgresql://..."
   export NEO4J_URI="bolt://..."
   ```

3. **启动服务**:
   ```bash
   docker run -d --name cube-castle cube-castle:latest
   ```

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🆘 支持

- 📧 邮箱: support@cubecastle.com
- 📖 文档: [docs/](docs/)
- 🐛 问题反馈: [Issues](../../issues)

## 🏆 致谢

感谢所有为 Cube Castle 项目做出贡献的开发者和用户！

---

**🏰 让 HR 管理变得简单而智能！** 