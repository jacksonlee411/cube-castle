# 🏰 Cube Castle 项目 - 第三阶段开发计划

## 📋 项目现状分析总结

### 1. 已完成的核心功能

**✅ 架构基础**
- 城堡模型架构设计已完整实现
- 多语言持久化（PostgreSQL + Neo4j）已配置
- API优先设计（OpenAPI 3.0）已定义
- gRPC通信（Go-Python）已实现
- Docker容器化部署已配置

**✅ 数据库层**
- 完整的数据库初始化脚本（15+表）
- 多租户支持的表结构
- 事务性发件箱模式实现
- 性能优化索引和触发器

**✅ 核心模块**
- CoreHR模块：员工管理、组织架构、职位管理
- Intelligence Gateway：AI集成、意图识别、gRPC服务
- 通用模块：类型定义、数据库连接、健康检查

### 2. 第二阶段工程蓝图目标实现情况

根据第二阶段工程蓝图，以下是目标实现情况：

**❌ 未实现的关键功能：**

1. **Temporal.io 工作流引擎集成**
2. **嵌入式OPA授权系统**
3. **PostgreSQL RLS多租户隔离**
4. **Next.js前端应用**
5. **Redis对话状态管理**
6. **可观测性三大支柱**

## 🔍 优先级最高的三个问题

### 问题1：Mock数据模式 - 缺乏真实业务逻辑

**当前状态：**
```go
// go-app/internal/corehr/service.go
func (s *Service) ListEmployees(ctx context.Context, page, pageSize int, search string) (*openapi.EmployeeListResponse, error) {
    // 使用 mock 数据
    return s.listEmployeesMock(ctx, page, pageSize, search)
}
```

**问题分析：**
- 所有CoreHR服务都使用硬编码的Mock数据
- 与元合约要求的"记录系统"原则不符
- 缺乏真实的数据持久化和业务逻辑
- 无法支持真实的业务场景

**影响程度：** 🔴 高 - 阻碍真实业务功能实现

### 问题2：简化的AI交互 - 无状态管理

**当前状态：**
```python
# python-ai/main.py
def InterpretText(self, request: intelligence_pb2.InterpretRequest, context):
    # 简单的意图识别，无状态管理
    response = client.chat.completions.create(
        model="deepseek-chat",
        messages=[{"role": "user", "content": request.user_text}],
        tools=tools,
        tool_choice="auto",
    )
```

**问题分析：**
- 无对话状态管理
- 缺乏上下文理解能力
- 与元合约的"对话状态追踪"要求不符
- 无法支持多轮对话和复杂交互

**影响程度：** 🔴 高 - 限制AI功能的核心价值

### 问题3：基础错误处理 - 缺乏可观测性

**当前状态：**
```go
// go-app/cmd/server/main.go
if err != nil {
    log.Printf("⚠️  Warning: Failed to connect to databases: %v", err)
    log.Printf("📝  Running in mock mode - using in-memory data")
    db = nil
}
```

**问题分析：**
- 错误处理过于简单
- 缺乏结构化日志
- 无监控和告警机制
- 无法进行有效的故障排查和性能优化

**影响程度：** 🟡 中 - 影响运维和问题排查能力

## 🚀 调整后的开发方案

### 阶段一：解决核心问题（优先级最高 - 4-5周）

#### 1.1 实现真实业务逻辑（2周）

**目标：** 替换所有Mock数据，实现真实的数据库操作和业务逻辑

**任务清单：**

**1.1.1 实现CoreHR Repository层**
```go
// 新增文件：internal/corehr/repository.go
type Repository struct {
    db *pgxpool.Pool
}

func (r *Repository) ListEmployees(ctx context.Context, tenantID uuid.UUID, page, pageSize int, search string) ([]Employee, int, error) {
    query := `
        SELECT id, employee_number, first_name, last_name, email, hire_date, status, created_at, updated_at
        FROM corehr.employees 
        WHERE tenant_id = $1 
        AND ($4 = '' OR first_name ILIKE $4 OR last_name ILIKE $4 OR employee_number ILIKE $4)
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `
    // 实现真实的分页查询
}

func (r *Repository) CreateEmployee(ctx context.Context, tenantID uuid.UUID, employee *Employee) error {
    query := `
        INSERT INTO corehr.employees (tenant_id, employee_number, first_name, last_name, email, hire_date, status)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, created_at, updated_at
    `
    // 实现真实的创建逻辑
}
```

**1.1.2 实现事务性发件箱模式**
```go
// 新增文件：internal/outbox/processor.go
type OutboxProcessor struct {
    db *pgxpool.Pool
}

func (p *OutboxProcessor) ProcessEvents(ctx context.Context) error {
    query := `
        SELECT id, aggregate_id, aggregate_type, event_type, payload, metadata
        FROM outbox.events 
        WHERE processed_at IS NULL
        ORDER BY created_at ASC
        LIMIT 100
    `
    // 实现事件处理逻辑
}
```

**1.1.3 实现组织架构管理**
```go
// 新增文件：internal/corehr/organization_repository.go
func (r *Repository) GetOrganizationTree(ctx context.Context, tenantID uuid.UUID) (*OrganizationTree, error) {
    query := `
        WITH RECURSIVE org_tree AS (
            SELECT id, name, code, level, parent_id, 0 as depth
            FROM corehr.organizations 
            WHERE tenant_id = $1 AND parent_id IS NULL
            UNION ALL
            SELECT o.id, o.name, o.code, o.level, o.parent_id, ot.depth + 1
            FROM corehr.organizations o
            JOIN org_tree ot ON o.parent_id = ot.id
            WHERE o.tenant_id = $1
        )
        SELECT * FROM org_tree ORDER BY depth, level
    `
    // 实现递归查询组织树
}
```

**交付物：**
- [ ] 完整的Repository层实现
- [ ] 事务性发件箱处理器
- [ ] 组织架构递归查询
- [ ] 单元测试覆盖

#### 1.2 实现Redis对话状态管理（1.5周）

**目标：** 为AI服务添加状态管理，支持多轮对话

**任务清单：**

**1.2.1 集成Redis客户端**
```python
# 修改：python-ai/main.py
import redis
import json
from typing import Dict, List, Optional

class DialogueStateManager:
    def __init__(self, redis_host='localhost', redis_port=6379, session_ttl=900):
        self.redis_client = redis.Redis(
            host=redis_host, 
            port=redis_port, 
            decode_responses=True
        )
        self.session_ttl = session_ttl
    
    def get_state(self, session_id: str) -> Dict:
        history_key = f"session:{session_id}:history"
        state_key = f"session:{session_id}:state"
        
        pipeline = self.redis_client.pipeline()
        pipeline.lrange(history_key, 0, -1)
        pipeline.hgetall(state_key)
        results = pipeline.execute()
        
        history = [json.loads(msg) for msg in results[0]]
        state = results[1]
        
        return {"history": history, "state": state}
    
    def update_state(self, session_id: str, user_message: dict, assistant_message: dict, new_state: dict):
        history_key = f"session:{session_id}:history"
        state_key = f"session:{session_id}:state"
        
        pipeline = self.redis_client.pipeline()
        pipeline.rpush(history_key, json.dumps(user_message))
        pipeline.rpush(history_key, json.dumps(assistant_message))
        if new_state:
            pipeline.hset(state_key, mapping=new_state)
        
        pipeline.expire(history_key, self.session_ttl)
        pipeline.expire(state_key, self.session_ttl)
        pipeline.execute()
```

**1.2.2 增强AI服务状态管理**
```python
# 修改：python-ai/main.py
class IntelligenceServiceImpl(intelligence_pb2_grpc.IntelligenceServiceServicer):
    def __init__(self):
        self.state_manager = DialogueStateManager()
    
    def InterpretText(self, request: intelligence_pb2.InterpretRequest, context):
        # 获取对话状态
        state_data = self.state_manager.get_state(request.session_id)
        conversation_history = state_data["history"]
        current_state = state_data["state"]
        
        # 构建上下文消息
        messages = []
        for msg in conversation_history[-10:]:  # 保留最近10轮对话
            messages.append({"role": msg["role"], "content": msg["content"]})
        messages.append({"role": "user", "content": request.user_text})
        
        # 调用AI服务
        response = client.chat.completions.create(
            model="deepseek-chat",
            messages=messages,
            tools=tools,
            tool_choice="auto",
        )
        
        # 更新对话状态
        user_message = {"role": "user", "content": request.user_text}
        assistant_message = {"role": "assistant", "content": response.choices[0].message.content}
        new_state = {"last_intent": response.choices[0].message.tool_calls[0].function.name if response.choices[0].message.tool_calls else "no_intent"}
        
        self.state_manager.update_state(request.session_id, user_message, assistant_message, new_state)
        
        return intelligence_pb2.InterpretResponse(
            intent=response.choices[0].message.tool_calls[0].function.name if response.choices[0].message.tool_calls else "no_intent_detected",
            structured_data_json=response.choices[0].message.tool_calls[0].function.arguments if response.choices[0].message.tool_calls else "{}"
        )
```

**1.2.3 扩展业务场景支持**
```python
# 新增：python-ai/business_functions.py
def approve_leave_request(request_id: str, decision: str, comment: str = "") -> dict:
    """批准或拒绝休假申请"""
    return {
        "request_id": request_id,
        "decision": decision,
        "comment": comment,
        "timestamp": datetime.now().isoformat()
    }

def get_employee_details(employee_id: str) -> dict:
    """获取员工详细信息"""
    return {
        "employee_id": employee_id,
        "name": "张三",
        "department": "技术部",
        "position": "软件工程师",
        "manager": "李四"
    }

# 扩展tools列表
tools = [
    {
        "type": "function",
        "function": {
            "name": "update_phone_number",
            "description": "Update an employee's phone number",
            "parameters": {
                "type": "object",
                "properties": {
                    "employee_id": {"type": "string"},
                    "new_phone_number": {"type": "string"}
                },
                "required": ["employee_id", "new_phone_number"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "approve_leave_request",
            "description": "Approve or reject a leave request",
            "parameters": {
                "type": "object",
                "properties": {
                    "request_id": {"type": "string"},
                    "decision": {"type": "string", "enum": ["approve", "reject"]},
                    "comment": {"type": "string"}
                },
                "required": ["request_id", "decision"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "get_employee_details",
            "description": "Get detailed information about an employee",
            "parameters": {
                "type": "object",
                "properties": {
                    "employee_id": {"type": "string"}
                },
                "required": ["employee_id"]
            }
        }
    }
]
```

**交付物：**
- [ ] Redis对话状态管理器
- [ ] 多轮对话支持
- [ ] 扩展的业务场景
- [ ] 对话历史持久化

#### 1.3 实现结构化日志和监控（1.5周）

**目标：** 建立完整的可观测性体系

**任务清单：**

**1.3.1 实现结构化日志**
```go
// 新增文件：internal/logging/logger.go
package logging

import (
    "context"
    "log/slog"
    "os"
    "time"
)

type Logger struct {
    *slog.Logger
}

func NewLogger() *Logger {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
        AddSource: true,
    }))
    return &Logger{logger}
}

func (l *Logger) WithRequestID(requestID string) *Logger {
    return &Logger{l.With("request_id", requestID)}
}

func (l *Logger) WithUserID(userID string) *Logger {
    return &Logger{l.With("user_id", userID)}
}

func (l *Logger) WithTenantID(tenantID string) *Logger {
    return &Logger{l.With("tenant_id", tenantID)}
}

// 请求日志中间件
func LoggingMiddleware(logger *Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            requestID := uuid.New().String()
            
            // 创建请求上下文日志器
            reqLogger := logger.WithRequestID(requestID)
            
            // 记录请求开始
            reqLogger.Info("HTTP request started",
                "method", r.Method,
                "path", r.URL.Path,
                "remote_addr", r.RemoteAddr,
                "user_agent", r.UserAgent(),
            )
            
            // 包装ResponseWriter以捕获状态码
            wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: 200}
            
            // 处理请求
            next.ServeHTTP(wrappedWriter, r)
            
            // 记录请求完成
            duration := time.Since(start)
            reqLogger.Info("HTTP request completed",
                "status_code", wrappedWriter.statusCode,
                "duration_ms", duration.Milliseconds(),
            )
        })
    }
}

type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}
```

**1.3.2 集成Prometheus监控**
```go
// 新增文件：internal/metrics/prometheus.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "net/http"
)

var (
    // HTTP请求指标
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )
    
    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )
    
    // 业务指标
    employeesCreatedTotal = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "employees_created_total",
            Help: "Total number of employees created",
        },
    )
    
    leaveRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "leave_requests_total",
            Help: "Total number of leave requests",
        },
        []string{"status"},
    )
    
    aiRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "ai_requests_total",
            Help: "Total number of AI requests",
        },
        []string{"intent", "status"},
    )
)

// 监控中间件
func PrometheusMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // 包装ResponseWriter
        wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: 200}
        
        next.ServeHTTP(wrappedWriter, r)
        
        // 记录指标
        duration := time.Since(start).Seconds()
        httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(wrappedWriter.statusCode)).Inc()
        httpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
    })
}

// 指标端点
func MetricsHandler() http.Handler {
    return promhttp.Handler()
}
```

**1.3.3 集成OpenTelemetry追踪**
```go
// 新增文件：internal/tracing/otel.go
package tracing

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/jaeger"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
    "go.opentelemetry.io/otel/trace"
)

func InitTracing(serviceName string) (*sdktrace.TracerProvider, error) {
    // 创建Jaeger导出器
    exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://localhost:14268/api/traces")))
    if err != nil {
        return nil, err
    }
    
    // 创建资源
    res, err := resource.New(context.Background(),
        resource.WithAttributes(semconv.ServiceName(serviceName)),
    )
    if err != nil {
        return nil, err
    }
    
    // 创建TracerProvider
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exp),
        sdktrace.WithResource(res),
    )
    
    // 设置全局TracerProvider
    otel.SetTracerProvider(tp)
    
    return tp, nil
}

// 追踪中间件
func TracingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        tracer := otel.Tracer("http")
        
        ctx, span := tracer.Start(ctx, "HTTP "+r.Method+" "+r.URL.Path)
        defer span.End()
        
        // 添加请求信息到span
        span.SetAttributes(
            semconv.HTTPMethod(r.Method),
            semconv.HTTPURL(r.URL.String()),
            semconv.HTTPUserAgent(r.UserAgent()),
        )
        
        // 将span上下文传递给下一个处理器
        r = r.WithContext(ctx)
        next.ServeHTTP(w, r)
    })
}
```

**交付物：**
- [ ] 结构化日志系统
- [ ] Prometheus指标收集
- [ ] OpenTelemetry追踪
- [ ] 监控仪表板配置

### 阶段二：架构增强（3-4周）

#### 2.1 实现嵌入式OPA授权系统
#### 2.2 实现PostgreSQL RLS多租户隔离
#### 2.3 集成Temporal.io工作流引擎

### 阶段三：前端应用开发（3-4周）

#### 3.1 搭建Next.js项目
#### 3.2 实现休假审批界面

## 📅 调整后的实施计划

| 阶段 | 时间 | 主要交付物 | 关键里程碑 |
|------|------|------------|------------|
| 阶段一 | 4-5周 | 真实业务逻辑、AI状态管理、可观测性 | 核心问题解决 |
| 阶段二 | 3-4周 | OPA授权、RLS、Temporal | 架构增强完成 |
| 阶段三 | 3-4周 | Next.js前端、审批界面 | 用户界面完成 |

**总计：10-13周**

## 🎯 阶段一成功标准

### 技术指标
- [ ] 所有Mock数据替换为真实数据库操作
- [ ] AI服务支持多轮对话和状态管理
- [ ] 完整的结构化日志和监控体系
- [ ] API响应时间 < 200ms

### 功能指标
- [ ] 员工CRUD操作完全基于数据库
- [ ] 组织架构支持递归查询
- [ ] AI对话保持上下文连续性
- [ ] 完整的错误追踪和监控

### 质量指标
- [ ] 代码覆盖率 > 80%
- [ ] 所有关键路径有日志记录
- [ ] 业务指标可监控
- [ ] 错误率 < 1%

## 🚨 风险与缓解措施

### 高风险项
1. **数据库性能问题** - 缓解：早期性能测试和索引优化
2. **Redis连接稳定性** - 缓解：实现连接池和重试机制
3. **日志性能影响** - 缓解：异步日志写入和采样策略

### 中风险项
1. **数据迁移复杂性** - 缓解：渐进式迁移和回滚计划
2. **监控系统开销** - 缓解：合理的采样率和聚合策略
3. **代码重构风险** - 缓解：充分的测试覆盖和渐进式重构

---

**这份开发方案将解决当前最关键的三个问题作为最高优先级，确保项目有坚实的业务逻辑基础、完善的AI交互能力和可靠的可观测性体系，为后续的架构增强和前端开发奠定基础。**

**最后更新**: 2025年1月  
**项目状态**: 开发中  
**负责人**: 开发团队 