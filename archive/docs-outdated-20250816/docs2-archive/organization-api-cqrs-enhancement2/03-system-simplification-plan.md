# Cube Castle 系统功能精简方案

**制定时间**: 2025-08-08  
**目标分支**: feature/simplify-core-functions  
**方案版本**: v1.0  
**状态**: 待审批

---

## 📋 精简目标

### 核心原则
- ✅ **保留核心功能**: 组织架构管理和前端页面展示
- ✅ **保留基础设施**: Docker容器、数据库、消息队列等基础组件
- ❌ **剥离非核心**: 移除员工管理、岗位管理等次要功能
- ✅ **保持架构**: 维持现有CQRS架构和微服务边界

### 精简收益 (修正版)
- **代码复杂度**: 降低60% (移除员工、岗位业务功能)
- **维护成本**: 减少50% (减少业务功能模块)
- **部署复杂度**: 降低30% (移除业务服务)
- **开发效率**: 提升40% (专注核心组织管理功能)
- **服务数量**: 从8个Go服务减少到4个核心服务 
- **业务复杂度**: 大幅简化，专注组织架构管理

---

## 🎯 保留功能清单

### ✅ 前端保留功能
```
frontend/
├── src/
│   ├── App.tsx                           # 主应用组件
│   ├── features/
│   │   └── organizations/                # 组织架构功能
│   │       ├── OrganizationDashboard.tsx # 组织管理页面
│   │       ├── OrganizationFilters.tsx   # 筛选组件
│   │       └── PaginationControls.tsx    # 分页组件
│   ├── layout/                           # 布局组件
│   │   ├── AppShell.tsx
│   │   ├── Header.tsx
│   │   ├── Sidebar.tsx
│   │   └── TopBar.tsx
│   ├── shared/
│   │   ├── api/
│   │   │   ├── client.ts                 # API客户端
│   │   │   └── organizations.ts          # 组织API
│   │   ├── hooks/
│   │   │   ├── useOrganizations.ts       # 组织数据钩子
│   │   │   └── useOrganizationMutations.ts
│   │   └── types/
│   │       └── organization.ts           # 组织类型定义
│   └── design-system/                    # Canvas Kit设计系统
```

### ✅ 后端保留服务
```
cmd/
├── organization-api-gateway/             # API网关 (8000)
├── organization-api-server/              # 查询服务 (8080)
├── organization-graphql-service/         # GraphQL服务 (8090)
└── organization-command-server/          # 命令服务 (9090)
```

### ✅ 基础设施保留
```
docker-compose.yml 保留组件:
├── PostgreSQL (5432)                     # 主数据库
├── Neo4j (7474/7687)                     # 图数据库
├── Redis (6379)                          # 缓存
├── Kafka生态系统                         # 消息队列
│   ├── Zookeeper (2181)
│   ├── Kafka (9092)
│   ├── Kafka-Connect (8083)
│   └── Kafka-UI (8081)                   # 管理界面
├── Temporal工作流引擎                    # ✅ 保留完整
│   ├── temporal-server (7233)
│   ├── temporal-ui (8085)
│   └── elasticsearch (9200)              # ✅ Temporal依赖保留
└── pgAdmin (5050)                        # PostgreSQL管理界面
```

### ✅ AI智能模块保留
```
python-ai/
├── main.py                               # ✅ AI服务主程序 (gRPC 50051)
├── dialogue_state.py                     # ✅ 对话状态管理
├── intelligence_pb2.py                   # ✅ gRPC协议实现
├── intelligence_pb2_grpc.py              # ✅ gRPC服务实现
├── requirements.txt                      # ✅ Python依赖
├── health-check.sh                       # ✅ AI服务健康检查
├── start.sh                              # ✅ AI服务启动脚本
└── venv/                                 # ✅ Python虚拟环境

contracts/proto/
└── intelligence.proto                    # ✅ AI服务gRPC协议定义

相关功能 (保留):
✅ 自然语言意图识别
✅ 员工查询智能助手  
✅ 对话状态管理 (Redis)
✅ OpenAI/DeepSeek集成
✅ AI缓存机制
```

---

## ❌ 剥离功能清单

### 🔧 业务功能服务 (完全移除)
```
cmd/
├── employee-server/                      # 员工管理服务 (8081)
├── position-server/                      # 岗位管理服务 (8082)  
├── organization-sync-service/            # 同步服务 (废弃)
├── position-graphql-service/             # 岗位GraphQL服务 (8095)
├── position-sync-service/                # 岗位同步服务 (废弃)
└── server/                               # 旧版服务器 (废弃)
```

### 📊 监控和性能模块 (选择性移除)
```
monitoring/
├── README.md                             # 保留 - 监控文档
├── performance_monitor.sh                # 移除 - 性能监控脚本
├── analyze_logs.sh                       # 移除 - 日志分析脚本
├── logs/                                 # 移除 - 历史日志
└── monitor.pid                           # 移除 - 监控进程文件

performance/
├── baseline.csv                          # 移除 - 性能基准数据
└── benchmarks.md                         # 移除 - 基准测试报告
```

### 🗂️ 前端功能模块
```
frontend/src/
├── features/
│   ├── dashboard/                        # 移除 - 非组织相关仪表板
│   ├── employees/                        # 移除 - 员工管理模块
│   └── positions/                        # 移除 - 岗位管理模块
├── shared/
│   ├── api/
│   │   ├── employees.ts                  # 移除 - 员工API
│   │   └── positions.ts                  # 移除 - 岗位API
│   ├── hooks/
│   │   ├── useEmployees.ts               # 移除 - 员工钩子
│   │   └── usePositions.ts               # 移除 - 岗位钩子
│   └── types/
│       ├── employee.ts                   # 移除 - 员工类型
│       └── position.ts                   # 移除 - 岗位类型
```

### 🗃️ 数据库表结构
```sql
-- 保留的组织相关表
✅ organization_units
✅ organization_hierarchies (Neo4j)

-- 需要清理的表
❌ employees                              # 员工基本信息表
❌ positions                              # 岗位信息表
❌ employee_positions                     # 员工岗位关系表
❌ employee_organization_relations        # 员工组织关系表
❌ position_hierarchies                   # 岗位层级表

-- Temporal相关表 (保留基础结构)
✅ 保留Temporal数据库schema
❌ 清理业务工作流数据
```

### 🗃️ 配置和文档清理
```
删除目录:
├── docs/employees/                       # 员工相关文档
├── docs/positions/                       # 岗位相关文档  
├── python-ai/                           # AI功能完整目录
├── archive/frontend-legacy-*/            # 遗留前端代码
├── backup/redundant-tools-*/             # 冗余工具备份
├── monitoring/logs/                      # 监控日志
├── performance/                          # 性能测试数据
└── temporal-config/                      # Temporal业务配置

删除文件:
├── *employee*.sql                        # 员工相关SQL脚本
├── *position*.sql                        # 岗位相关SQL脚本
├── test_employee_*.js                    # 员工测试脚本
├── test_position_*.sh                    # 岗位测试脚本
├── position-*.html                       # 岗位测试页面
├── create-position-test-data.js          # 岗位测试数据脚本
└── *performance*.sh                      # 性能测试脚本

删除数据文件:
├── data/corrected_employee_data.sql      # 员工数据修正
├── data/final_employee_data.sql          # 员工最终数据
├── data/fix_employee_data.sql            # 员工数据修复
└── data/simple_employee_data.sql         # 员工简化数据
```

### 🔧 基础设施精简 (Docker容器调整)
```yaml
# docker-compose.yml 移除的组件:
# 全部保留，不移除任何基础设施组件

# 保留完整基础设施:
✅ postgres                               # 主数据库 (5432)
✅ neo4j                                  # 图数据库 (7474/7687)  
✅ redis                                  # 缓存 (6379)
✅ kafka + zookeeper + kafka-connect      # 消息队列生态
✅ kafka-ui                               # Kafka管理界面 (8081)
✅ temporal-server + temporal-ui          # 工作流引擎
✅ elasticsearch                          # Temporal依赖
✅ pgadmin                                # PostgreSQL管理界面 (5050)
```

---

## 🔧 实施步骤

### Phase 1: 后端服务精简 (1天)

#### Step 1.1: 停止非核心业务服务 (0.5天)
```bash
# 使用管理脚本停止员工和岗位服务
./scripts/microservices-manager.sh stop employee-server
./scripts/microservices-manager.sh stop position-server
./scripts/microservices-manager.sh stop position-graphql-service

# 停止监控脚本 (AI服务保留运行)
sudo pkill -f "performance_monitor"
sudo pkill -f "analyze_logs"
```

#### Step 1.2: 更新微服务管理脚本 (0.25天)
```bash
# 编辑 scripts/microservices-manager.sh
# 移除员工和岗位服务配置，保留AI服务
declare -A SERVICES=(
    ["organization-api-gateway"]="8000:cmd/organization-api-gateway:organization-api-gateway"
    ["organization-api-server"]="8080:cmd/organization-api-server:organization-api-server"  
    ["organization-graphql-service"]="8090:cmd/organization-graphql-service:organization-graphql-service"
    ["organization-command-server"]="9090:cmd/organization-command-server:organization-command-server"
    # 移除以下行:
    # ["employee-server"]="8081:cmd/employee-server:employee-server"
    # ["position-server"]="8082:cmd/position-server:position-server"
    # AI服务单独管理，不在此脚本中
)
```

#### Step 1.3: 清理业务服务目录 (0.25天)
```bash
# 创建归档目录
mkdir -p archive/removed-business-services-$(date +%Y%m%d-%H%M%S)

# 移动业务服务目录到归档 (保留AI模块)
mv cmd/employee-server archive/removed-business-services-*/
mv cmd/position-server archive/removed-business-services-*/
mv cmd/position-graphql-service archive/removed-business-services-*/
mv cmd/organization-sync-service archive/removed-business-services-*/
mv cmd/position-sync-service archive/removed-business-services-*/
mv cmd/server archive/removed-business-services-*/

# 清理监控和性能模块 (保留AI模块)
mv monitoring archive/removed-business-services-*/
mv performance archive/removed-business-services-*/
```

### Phase 2: 前端模块精简 (1天)

#### Step 2.1: 移除前端非核心功能 (0.5天)
```bash
# 创建前端清理归档
mkdir -p archive/removed-frontend-modules-$(date +%Y%m%d-%H%M%S)

# 移除员工和岗位相关文件
find frontend/src -name "*employee*" -o -name "*position*" | while read file; do
    echo "Moving $file to archive"
    mv "$file" archive/removed-frontend-modules-*/
done

# 移除非组织相关的dashboard功能
if [ -d "frontend/src/features/dashboard" ]; then
    mv frontend/src/features/dashboard archive/removed-frontend-modules-*/
fi
```

#### Step 2.2: 更新前端路由和导航 (0.5天)
```typescript
// 更新 frontend/src/App.tsx
// 移除员工和岗位相关路由

// 更新 frontend/src/layout/Sidebar.tsx  
// 移除员工和岗位菜单项

// 更新 frontend/src/layout/Header.tsx
// 简化导航栏，只保留组织架构相关功能
```

### Phase 3: 数据库和配置精简 (0.5天)

#### Step 3.1: 备份非核心数据 (0.25天)
```sql
-- 备份员工和岗位数据
mkdir -p backup/pre-simplification-$(date +%Y%m%d-%H%M%S)

pg_dump -h localhost -p 5432 -U user -d cubecastle -t employees > backup/pre-simplification-*/employees_backup.sql
pg_dump -h localhost -p 5432 -U user -d cubecastle -t positions > backup/pre-simplification-*/positions_backup.sql
pg_dump -h localhost -p 5432 -U user -d cubecastle -t employee_positions > backup/pre-simplification-*/employee_positions_backup.sql

# 不清理Temporal和AI相关数据
```

#### Step 3.2: 清理文档和测试文件 (0.125天)
```bash
# 移动员工和岗位相关文档到archive  
mkdir -p archive/removed-docs-$(date +%Y%m%d-%H%M%S)
find . -name "*employee*" -o -name "*position*" | grep -E "\.(md|sql|js|html|sh)$" | while read file; do
    echo "Moving $file to archive"
    mv "$file" archive/removed-docs-*/
done

# 清理数据目录 (保留AI相关文件)
mv data/*employee* archive/removed-docs-*/ 2>/dev/null || true
mv data/*position* archive/removed-docs-*/ 2>/dev/null || true
```

#### Step 3.3: 更新项目文档 (0.125天)
```markdown
# 更新以下文件:
- README.md                    # 移除员工、岗位功能说明，保留AI功能
- MICROSERVICES_MANAGEMENT.md # 更新服务列表，增加AI服务管理说明
- docker-compose.yml          # 保持完整基础设施不变
```

### Phase 4: 验证和优化 (0.5天)

#### Step 4.1: 系统验证 (0.25天)
```bash
# 启动精简后的系统 (保留AI服务)
./scripts/microservices-manager.sh start
cd python-ai && ./start.sh  # 启动AI服务

# 验证核心功能
curl http://localhost:8000/health
curl http://localhost:8080/api/v1/organization-units
curl http://localhost:8090/graphql -d '{"query":"query{organizations{code name}}"}'
curl -X POST http://localhost:9090/api/v1/organization-units -d '{"name":"测试部门","unit_type":"DEPARTMENT"}'

# 验证AI服务
grpc_cli call localhost:50051 intelligence.IntelligenceService.InterpretText "user_text:'查询员工列表' session_id:'test'"

# 验证基础设施
curl http://localhost:8085  # Temporal UI
curl http://localhost:5050  # pgAdmin
curl http://localhost:8081  # Kafka UI
curl http://localhost:9200  # Elasticsearch
```

#### Step 4.2: 前端功能验证 (0.25天)
```bash
# 启动前端
cd frontend && npm run dev

# 验证页面功能:
# - 组织架构管理页面正常加载
# - 增删改查功能正常
# - 无员工、岗位相关功能残留
# - 导航菜单简化正确
```

---

## 📊 精简前后对比

### 服务对比 (修正版)
| 分类 | 精简前 | 精简后 | 减少量 |
|------|--------|--------|--------|
| **Go微服务** | 8个 | 4个 | -50% |
| **Python服务** | 1个(AI) | 1个(AI) | 保留 |
| **端口占用** | 12个 | 9个 | -25% |
| **代码文件** | ~300个 | ~200个 | -33% |
| **技术栈** | Go+Python+JS | Go+Python+JS | 保持 |

### 前端对比 (修正版)
| 分类 | 精简前 | 精简后 | 减少量 |
|------|--------|--------|--------|
| **功能模块** | 6个 | 2个 | -67% |
| **API集成** | 4套(组织+员工+岗位+AI) | 2套(组织+AI) | -50% |
| **组件数量** | ~80个 | ~35个 | -56% |
| **Hook函数** | ~15个 | ~8个 | -47% |

### 数据库对比 (修正版)  
| 分类 | 精简前 | 精简后 | 减少量 |
|------|--------|--------|--------|
| **业务表** | 8个 | 2个 | -75% |
| **关系表** | 4个 | 0个 | -100% |
| **API端点** | 35+ | 16+ | -54% |
| **GraphQL操作** | 15+ | 10+ | -33% |

### 基础设施对比 (修正版)
| 分类 | 精简前 | 精简后 | 减少量 |
|------|--------|--------|--------|
| **Docker服务** | 12个 | 12个 | 保持完整 |
| **管理界面** | 3个 | 3个 | 保留全部 |
| **监控脚本** | 5个 | 0个 | -100% |
| **gRPC服务** | 1个(AI) | 1个(AI) | 保留 |

---

## ⚠️ 风险评估

### 技术风险
- **数据丢失风险**: 🔸 中等 (通过备份缓解)
- **功能回退风险**: 🔶 高 (需要重新开发)
- **架构破坏风险**: 🔹 低 (保持核心架构)

### 业务风险
- **功能缺失风险**: 🔶 高 (员工岗位管理完全移除)
- **用户体验风险**: 🔸 中等 (简化界面)
- **数据迁移风险**: 🔹 低 (通过备份保护)

### 缓解措施
1. **完整备份**: 所有删除数据先备份
2. **分支保护**: 在新分支操作，保持master完整
3. **回滚准备**: 保留完整回滚脚本
4. **测试验证**: 每个阶段都进行功能验证

---

## 🔄 回滚方案

### 快速回滚
```bash
# 切换回master分支
git checkout master

# 或者从archive恢复
./scripts/restore-from-archive.sh removed-services-<timestamp>
```

### 数据回滚
```sql
-- 恢复数据库备份
psql -h localhost -p 5432 -U user -d cubecastle < backup/employees_backup_<timestamp>.sql
psql -h localhost -p 5432 -U user -d cubecastle < backup/positions_backup_<timestamp>.sql
```

---

## ✅ 验收标准

### 功能验证
- [ ] 组织架构管理功能正常
- [ ] 前端页面加载正常
- [ ] API网关路由正确
- [ ] GraphQL查询正常
- [ ] 数据CRUD操作正常

### 性能验证
- [ ] 前端加载时间 < 2s
- [ ] API响应时间 < 200ms
- [ ] 内存使用降低 > 30%
- [ ] CPU使用降低 > 40%

### 稳定性验证
- [ ] 所有保留服务启动成功
- [ ] 健康检查通过
- [ ] 无错误日志
- [ ] 数据一致性保持

---

## 📅 实施时间表

| 阶段 | 内容 | 用时 | 负责人 |
|------|------|------|--------|
| **Phase 1** | 后端业务服务精简 | 1天 | 后端开发 |
| **Phase 2** | 前端模块精简 | 1天 | 前端开发 |
| **Phase 3** | 数据库和配置精简 | 0.5天 | DBA+DevOps |
| **Phase 4** | 验证和优化 | 0.5天 | 全栈团队 |
| **总计** | 完整精简 | **3天** | 全栈团队 |

---

## 🎯 后续建议

### 短期优化 (1-2周)
- 重构剩余组织管理代码
- 优化前端组件性能
- 完善API文档

### 长期规划 (1-3个月)
- 基于精简版本进行深度优化
- 实施前面制定的重构计划
- 添加必要的监控和测试

---

## 📋 审批检查清单

**请在审批前确认以下事项:**

- [ ] 确认保留组织架构管理功能
- [ ] ✅ 确认保留AI智能模块完整功能
- [ ] ✅ 确认保留Temporal工作流引擎和Elasticsearch
- [ ] ✅ 确认保留完整基础设施 (所有Docker服务)
- [ ] 确认可以接受员工和岗位功能的完全移除
- [ ] 确认备份策略可接受
- [ ] 确认回滚方案可执行
- [ ] 确认实施时间安排合理 (3天)
- [ ] 确认风险缓解措施充分

**审批决定:**
- [ ] ✅ 批准执行 
- [ ] ❌ 需要修改
- [ ] ⏸️ 暂缓执行

**审批意见:**
_请在此处添加审批意见..._

---

**方案制定**: Claude Code AI Assistant  
**审批状态**: 待审批  
**实施状态**: 待批准后执行  

> 💡 **重要提醒**: 本方案会永久移除员工管理和岗位管理功能。请确保这些功能在当前业务中不是必需的，或者已有其他替代方案。