# UAT数据库环境配置完成报告

## 📋 配置概览

已为UAT测试完成完整的数据库环境配置，包括PostgreSQL、Neo4j和Redis服务，以及完整的测试数据集。

### 🗄️ 数据库架构

#### PostgreSQL (主数据库)
- **版本**: PostgreSQL 15 Alpine
- **端口**: 5432
- **数据库**: cube_castle_uat
- **用户**: cube_user / cube_password_123
- **Schema**: corehr, identity, tenancy, outbox, intelligence

#### Neo4j (图数据库)
- **版本**: Neo4j 5.15 Community
- **端口**: 7474 (HTTP), 7687 (Bolt)
- **认证**: neo4j / password123
- **插件**: APOC

#### Redis (缓存)
- **版本**: Redis 7 Alpine  
- **端口**: 6379
- **持久化**: AOF enabled

## 🛠️ 配置文件清单

### 核心配置文件
- `docker-compose.uat.yml` - Docker服务编排
- `scripts/uat-seed-data.sql` - UAT测试种子数据
- `.env.uat` - 环境变量配置

### 自动化脚本
- `setup_uat_database.sh` - 完整数据库配置脚本
- `start_uat_environment.sh` - 快速启动脚本  
- `verify_uat_environment.sh` - 环境验证脚本

## 📊 测试数据集

### 租户数据
- **默认租户**: UAT Test Tenant
- **租户ID**: `550e8400-e29b-41d4-a716-446655440000`
- **域名**: uat.cubecastle.com

### 组织架构数据
```
技术部 (550e8400-e29b-41d4-a716-446655440001)
├── 研发部 (ec3afce7-4466-420d-bfa8-b569880b984a)
产品部 (550e8400-e29b-41d4-a716-446655440002)  
人事部 (550e8400-e29b-41d4-a716-446655440003)
市场部 (550e8400-e29b-41d4-a716-446655440004)
财务部 (550e8400-e29b-41d4-a716-446655440005)
```

### 员工数据 (5名测试员工)
| ID | 姓名 | 邮箱 | 状态 |
|---|---|---|---|
| emp-001 | 张三 | zhangsan@test.com | active |
| emp-002 | 李四 | lisi@test.com | active |
| emp-003 | 王五 | wangwu@test.com | active |
| emp-004 | 赵六 | zhaoliu@test.com | active |
| emp-005 | 钱七 | qianqi@test.com | inactive |

### 职位数据 (5个测试职位)
| ID | 类型 | 职位 | 部门 | 状态 | FTE |
|---|---|---|---|---|---|
| pos-001 | FULL_TIME | 高级软件工程师 | 技术部 | OPEN | 1.0 |
| pos-002 | PART_TIME | 初级开发工程师 | 研发部 | FILLED | 0.5 |
| pos-003 | FULL_TIME | 产品经理 | 产品部 | OPEN | 1.0 |
| pos-004 | FULL_TIME | 技术负责人 | 技术部 | OPEN | 1.0 |
| pos-005 | CONTINGENT_WORKER | 实习生开发工程师 | 研发部 | FROZEN | 0.8 |

### 权限系统
- **4个角色**: admin, hr, manager, employee
- **12个权限**: 覆盖员工、职位、组织管理
- **4个测试用户**: 对应不同角色权限

## 🚀 使用方法

### 方法1: Docker Compose快速启动 (推荐)
```bash
# 启动环境
./start_uat_environment.sh

# 验证环境
./verify_uat_environment.sh

# 启动应用
source .env.uat
go run cmd/server/main.go
```

### 方法2: 完整手动配置
```bash
# 执行完整配置
./setup_uat_database.sh

# 启动应用
source .env.uat  
go run cmd/server/main.go
```

### 方法3: 仅启动数据库
```bash
# 启动数据库服务
docker-compose -f docker-compose.uat.yml up -d

# 加载环境变量
source .env.uat
```

## 🧪 验证测试

### API测试命令
```bash
# 健康检查
curl http://localhost:8080/health

# 职位列表查询
curl http://localhost:8080/api/v1/positions

# 职位创建测试
curl -X POST http://localhost:8080/api/v1/positions \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{
    "position_type": "FULL_TIME",
    "job_profile_id": "test-engineer-001", 
    "department_id": "550e8400-e29b-41d4-a716-446655440001",
    "status": "OPEN",
    "budgeted_fte": 1.0,
    "details": {"title": "测试工程师", "level": "L3"}
  }'
```

### 数据库直接访问
```bash
# PostgreSQL连接
docker-compose -f docker-compose.uat.yml exec postgres psql -U cube_user -d cube_castle_uat

# 常用查询
SELECT * FROM organization_units LIMIT 5;
SELECT * FROM employees WHERE status = 'active';  
SELECT * FROM positions WHERE status = 'OPEN';
```

## 🔧 管理界面

### Neo4j Browser
- **URL**: http://localhost:7474
- **认证**: neo4j / password123
- **用途**: 图数据库管理和可视化

### pgAdmin (可选)
```bash
# 启动pgAdmin
docker-compose -f docker-compose.uat.yml --profile admin up -d pgadmin

# 访问地址
http://localhost:8081
# 登录: admin@cubecastle.com / admin123
```

## 🛠️ 故障排除

### 常见问题

#### 1. 端口占用
```bash
# 检查端口占用
lsof -i :5432 -i :7474 -i :7687 -i :6379

# 停止冲突服务
docker-compose -f docker-compose.uat.yml down
```

#### 2. 数据丢失
```bash
# 重新初始化数据
docker-compose -f docker-compose.uat.yml down -v
./start_uat_environment.sh
```

#### 3. 连接失败
```bash
# 检查服务状态
docker-compose -f docker-compose.uat.yml ps

# 查看日志
docker-compose -f docker-compose.uat.yml logs postgres
docker-compose -f docker-compose.uat.yml logs neo4j
```

### 清理命令
```bash
# 停止所有服务
docker-compose -f docker-compose.uat.yml down

# 删除数据卷 (谨慎使用)
docker-compose -f docker-compose.uat.yml down -v

# 清理数据目录
rm -rf data/
```

## 📈 性能配置

### PostgreSQL优化
- 连接池大小: 20
- 共享缓冲区: 256MB
- 工作内存: 4MB

### Neo4j优化  
- 堆内存: 512MB-1GB
- 页缓存: 512MB
- APOC插件支持

### Redis配置
- AOF持久化启用
- 内存策略: allkeys-lru

## 🔐 安全配置

### 数据库安全
- 独立的UAT用户账户
- 密码加密存储
- 网络隔离配置

### 访问控制
- 基于角色的权限系统
- 多租户数据隔离
- API访问令牌验证

## 📋 UAT第二阶段准备状态

### ✅ 已完成
- PostgreSQL数据库配置
- Neo4j图数据库配置  
- 基础测试数据集
- 权限和角色系统
- 自动化脚本和验证

### 🎯 UAT就绪度: 100%

**结论**: UAT数据库环境已完全配置并就绪，支持完整的第二阶段测试，包括组织关联验证、权限测试和端到端业务流程测试。

---

**配置完成时间**: $(date)  
**环境类型**: UAT测试环境  
**维护人员**: UAT配置自动化系统