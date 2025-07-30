# 元合约编辑器架构调整方案

## 🏗️ 基于城堡蓝图的单体架构调整

根据城堡蓝图，元合约编辑器将采用"雄伟单体"架构，作为Core HR城堡的一个独立"塔楼"模块。

## 🎯 调整后的核心特性

### 1. 单体集成方案
```go
// 元合约编辑器作为城堡内的一个模块
type MetaContractEditorTower struct {
    // 符合城堡模型的模块定义
    name        string  // "MetaContractEditor"
    boundaries  API     // 明确的API边界
    governance  OPA     // 嵌入式策略引擎
}

// 模块API定义（城墙与门禁）
type EditorAPI interface {
    // 编辑器核心API
    CreateProject(tenant_id string, config ProjectConfig) (*Project, error)
    EditContract(project_id string, contract MetaContract) error
    CompileContract(project_id string) (*CompileResult, error)
    
    // 模板和协作API
    ListTemplates(category string) ([]Template, error)
    ShareProject(project_id string, users []User) error
}
```

### 2. 进程内实时编译
```go
// 嵌入式编译器（进程内，零网络延迟）
type EmbeddedCompiler struct {
    core          *metacontract.Compiler  // 现有编译器
    changeTracker *ChangeTracker          // 变更跟踪
    eventBus      *InProcessEventBus      // 进程内事件总线
}

// 实时编译（进程内调用）
func (ec *EmbeddedCompiler) CompileInProcess(
    changes []MetaContractChange,
) *CompilationResult {
    // 1. 进程内增量编译
    result := ec.core.CompilePartial(changes)
    
    // 2. 进程内事件广播
    ec.eventBus.Publish("COMPILATION_COMPLETE", result)
    
    return result
}
```

### 3. 本地WebSocket通信
```go
// 本地WebSocket服务（同进程）
type LocalWebSocketService struct {
    hub        *websocket.Hub
    compiler   *EmbeddedCompiler
    storage    *LocalStorage
}

// 本地实时同步
func (ws *LocalWebSocketService) HandleEditorChanges(
    conn *websocket.Conn, 
    changes []EditorChange,
) {
    // 1. 本地存储更新
    ws.storage.SaveChanges(changes)
    
    // 2. 进程内编译
    result := ws.compiler.CompileInProcess(changes)
    
    // 3. 本地广播结果
    ws.hub.BroadcastToRoom(conn.ProjectID, result)
}
```

## 🔧 本地部署架构

### 部署拓扑
```yaml
LocalDeployment:
  single_container: "cube-castle-monolith"
  components:
    - name: "go_application"
      description: "包含所有城堡模块的单体应用"
      ports: [8080, 8081] # HTTP + WebSocket
      
    - name: "postgresql"
      description: "本地数据库实例"
      ports: [5432]
      storage: "/data/postgres"
      
    - name: "redis"
      description: "本地缓存实例"  
      ports: [6379]
      storage: "/data/redis"
      
    - name: "nginx"
      description: "反向代理和静态文件服务"
      ports: [80, 443]
      config: "/etc/nginx/cube-castle.conf"

  networking:
    type: "bridge"
    internal_dns: "enabled"
    external_access: "nginx_proxy"
```

### Docker Compose配置
```yaml
# docker-compose.yml
version: '3.8'
services:
  cube-castle:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - redis
    environment:
      - DATABASE_URL=postgres://user:pass@postgres:5432/cubecastle
      - REDIS_URL=redis://redis:6379
    volumes:
      - ./data:/app/data
      
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: cubecastle
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
    volumes:
      - postgres_data:/var/lib/postgresql/data
      
  redis:
    image: redis:7
    volumes:
      - redis_data:/data
      
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - cube-castle

volumes:
  postgres_data:
  redis_data:
```

## 📦 简化的CI/CD流程

### 单体构建流程
```bash
#!/bin/bash
# build.sh - 单体应用构建脚本

# 1. 前端构建
cd frontend
npm run build
cd ..

# 2. 后端构建（包含前端资源）
go build -o cube-castle \
  -ldflags "-X main.Version=${VERSION}" \
  ./cmd/server

# 3. Docker镜像构建
docker build -t cube-castle:${VERSION} .

# 4. 本地部署
docker-compose up -d
```

### 元合约验证流程
```go
// 构建时验证（CI/CD集成）
func ValidateMetaContractIntegrity() error {
    // 1. 加载所有元合约文件
    contracts, err := LoadAllMetaContracts("./contracts/")
    
    // 2. 验证模块API一致性
    for _, contract := range contracts {
        if err := ValidateModuleAPI(contract); err != nil {
            return fmt.Errorf("API validation failed: %w", err)
        }
    }
    
    // 3. 验证编译器生成代码
    for _, contract := range contracts {
        result := compiler.Compile(contract)
        if !result.Success {
            return fmt.Errorf("compilation failed: %v", result.Errors)
        }
    }
    
    return nil
}
```

## 🛡️ 嵌入式治理

### 进程内OPA集成
```go
// 嵌入式策略引擎
type EmbeddedGovernance struct {
    opa     *opa.OPA           // 嵌入式OPA库
    policies map[string]string  // 策略文件映射
    cache   *PolicyCache       // 策略结果缓存
}

// 进程内策略检查（零延迟）
func (eg *EmbeddedGovernance) CheckPolicy(
    user User, 
    action string, 
    resource string,
) (*PolicyResult, error) {
    // 1. 构建上下文
    ctx := map[string]interface{}{
        "user":     user,
        "action":   action,
        "resource": resource,
        "tenant":   user.TenantID,
    }
    
    // 2. 进程内策略评估
    result, err := eg.opa.Eval(ctx, "data.authz.allow")
    if err != nil {
        return nil, err
    }
    
    // 3. 缓存结果
    eg.cache.Store(ctx, result)
    
    return &PolicyResult{
        Allowed: result.Allowed(),
        Reason:  result.Reason(),
    }, nil
}
```

## 🎯 简化的功能范围

### 第一阶段：核心编辑器（4周）
1. **Week 1**: 单体应用框架 + 嵌入式编译器
2. **Week 2**: React编辑器 + 本地WebSocket
3. **Week 3**: 实时预览 + 错误展示
4. **Week 4**: 本地部署 + Docker化

### 第二阶段：模板和协作（4周）
1. **Week 5**: 本地模板系统
2. **Week 6**: 项目管理和用户权限
3. **Week 7**: 版本控制（本地Git集成）
4. **Week 8**: 测试和优化

## 💡 城堡蓝图的优势体现

### 1. 开发效率最大化
- **零网络延迟**: 所有模块间调用都是进程内函数调用
- **统一数据库**: 无需处理分布式事务和数据一致性
- **简化调试**: 单进程调试，完整的调用栈跟踪

### 2. 运维复杂度最小化  
- **单一部署单元**: 一个Docker容器搞定所有
- **本地数据库**: 无需管理云数据库和网络连接
- **零依赖服务**: 无需Kafka、Consul等分布式组件

### 3. 未来演进能力
- **清晰模块边界**: 每个"塔楼"都有明确的API边界
- **绞杀者就绪**: 未来可无缝拆分为微服务
- **元合约保障**: 架构纪律通过元合约强制执行

## 🚀 立即开始的行动项

### 第一周任务清单
1. **环境搭建**: 创建单体应用基础框架
2. **模块定义**: 定义MetaContractEditor塔楼的API边界
3. **编译器集成**: 将现有编译器嵌入到单体应用中
4. **本地WebSocket**: 实现进程内实时通信机制

这个调整后的方案完全符合城堡蓝图的哲学，将复杂度降到最低，同时保持了所有核心功能和未来扩展能力。