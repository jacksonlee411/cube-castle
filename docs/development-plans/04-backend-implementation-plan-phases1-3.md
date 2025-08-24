# 后端团队第1-3阶段实施方案

**文档版本**: v1.0  
**创建日期**: 2025-08-24  
**方案编号**: 04  
**实施团队**: 后端团队 (Go服务)  
**基于计划**: 03-api-compliance-intensive-refactoring-plan.md  
**开发方式**: 与前端团队并行开发

## 🎯 后端团队职责范围

**核心任务**: API服务架构完善和权限体系集成  
**团队规模**: 2-3名后端工程师  
**主要交付**: REST命令服务 + GraphQL权限 + 监控审计  
**技术栈**: Go 1.21+, PostgreSQL, Redis, Prometheus

### 📋 后端专属任务清单

```yaml
架构服务:
  - REST命令服务 (localhost:9090): CRUD操作和业务命令
  - GraphQL查询服务 (localhost:8090): 数据查询和权限验证
  - 权限验证中间件: OAuth 2.0 + PBAC集成
  - 审计监控体系: Prometheus + 结构化日志

数据层:
  - PostgreSQL优化: 时态查询、层级管理、索引优化
  - Redis缓存: 查询结果缓存、会话管理
  - 数据一致性: 单一数据源架构保证

基础设施:
  - Docker容器化: 多阶段构建、环境隔离
  - 监控告警: Prometheus指标、Grafana仪表板
  - 健康检查: 存活探针、就绪探针
```

## 🚀 第1阶段: 核心架构修复 (3-4天)

### Day 1-2: REST命令服务完善

#### 🎯 任务目标
修复localhost:9090服务响应，建立企业级REST API标准

#### 📋 详细任务清单

**1.1 服务启动修复** (2小时)
```bash
# 问题诊断
cd /home/shangmeilin/cube-castle/cmd/organization-command-service
go run main.go  # 检查启动错误

# 预期问题和解决方案
- 端口冲突: 检查9090端口占用，修改配置
- 数据库连接: 验证PostgreSQL连接字符串
- 依赖缺失: go mod tidy更新依赖
```

**1.2 HTTP方法和端点规范化** (4小时)
```go
// 修正文件: internal/handlers/organization.go

// ❌ 修正前 - 不符合REST规范
PUT /api/v1/organization-units/{id}/suspend
PUT /api/v1/organization-units/{id}/reactivate

// ✅ 修正后 - 符合业务操作语义
POST /api/v1/organization-units/{code}/suspend
POST /api/v1/organization-units/{code}/activate

// 实现代码结构
type OrganizationHandler struct {
    service OrganizationService
    logger  *log.Logger
}

func (h *OrganizationHandler) SuspendOrganization(w http.ResponseWriter, r *http.Request) {
    // 提取路径参数 {code}
    code := mux.Vars(r)["code"]
    
    // 业务逻辑调用
    result, err := h.service.SuspendOrganization(r.Context(), code)
    if err != nil {
        h.writeErrorResponse(w, err)
        return
    }
    
    // 标准响应信封
    h.writeSuccessResponse(w, result, "Organization suspended successfully")
}

func (h *OrganizationHandler) ActivateOrganization(w http.ResponseWriter, r *http.Request) {
    // 注意: 方法名从reactivateOrganization改为activateOrganization
    code := mux.Vars(r)["code"]
    
    result, err := h.service.ActivateOrganization(r.Context(), code)
    if err != nil {
        h.writeErrorResponse(w, err)
        return
    }
    
    h.writeSuccessResponse(w, result, "Organization activated successfully")
}
```

**1.3 企业级响应信封实现** (4小时)
```go
// 新建文件: internal/types/responses.go

// 企业级成功响应结构
type SuccessResponse struct {
    Success   bool        `json:"success"`
    Data      interface{} `json:"data"`
    Message   string      `json:"message"`
    Timestamp string      `json:"timestamp"`
    RequestID string      `json:"requestId"`
}

// 企业级错误响应结构
type ErrorResponse struct {
    Success   bool      `json:"success"`
    Error     ErrorInfo `json:"error"`
    Timestamp string    `json:"timestamp"`
    RequestID string    `json:"requestId"`
}

type ErrorInfo struct {
    Code    string      `json:"code"`
    Message string      `json:"message"`
    Details interface{} `json:"details,omitempty"`
}

// 响应写入工具方法
func WriteSuccessResponse(w http.ResponseWriter, data interface{}, message string, requestID string) {
    response := SuccessResponse{
        Success:   true,
        Data:      data,
        Message:   message,
        Timestamp: time.Now().UTC().Format(time.RFC3339),
        RequestID: requestID,
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}

func WriteErrorResponse(w http.ResponseWriter, code, message string, statusCode int, requestID string) {
    response := ErrorResponse{
        Success: false,
        Error: ErrorInfo{
            Code:    code,
            Message: message,
        },
        Timestamp: time.Now().UTC().Format(time.RFC3339),
        RequestID: requestID,
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(response)
}
```

**1.4 请求追踪中间件** (2小时)
```go
// 修改文件: main.go 或 internal/middleware/request.go

func RequestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 生成或提取请求ID
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = generateUUID() // 实现UUID生成
        }
        
        // 设置响应头
        w.Header().Set("X-Request-ID", requestID)
        
        // 添加到上下文
        ctx := context.WithValue(r.Context(), "requestID", requestID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// 在main.go中注册中间件
func main() {
    router := mux.NewRouter()
    
    // 注册请求追踪中间件
    router.Use(RequestIDMiddleware)
    router.Use(LoggingMiddleware)
    
    // 注册API路由
    api := router.PathPrefix("/api/v1").Subrouter()
    orgHandler := handlers.NewOrganizationHandler(orgService)
    api.HandleFunc("/organization-units/{code}/suspend", orgHandler.SuspendOrganization).Methods("POST")
    api.HandleFunc("/organization-units/{code}/activate", orgHandler.ActivateOrganization).Methods("POST")
    
    log.Println("REST Command Service starting on :9090")
    http.ListenAndServe(":9090", router)
}
```

#### ✅ Day 1-2 完成标准
- [ ] localhost:9090服务正常启动和响应
- [ ] suspend/activate端点使用POST方法
- [ ] 方法重命名: reactivateOrganization→activateOrganization
- [ ] 统一/api/v1前缀和{code}路径参数
- [ ] 企业级响应信封格式实现
- [ ] 请求追踪中间件集成

### Day 3-4: GraphQL权限集成

#### 🎯 任务目标
为GraphQL查询服务集成OAuth 2.0权限验证和PBAC权限模型

#### 📋 详细任务清单

**3.1 JWT验证中间件** (4小时)
```go
// 新建文件: internal/auth/jwt.go

type JWTMiddleware struct {
    secretKey []byte
    issuer    string
    audience  string
}

func NewJWTMiddleware(secretKey, issuer, audience string) *JWTMiddleware {
    return &JWTMiddleware{
        secretKey: []byte(secretKey),
        issuer:    issuer,
        audience:  audience,
    }
}

func (j *JWTMiddleware) ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        // 验证签名方法
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("invalid signing method")
        }
        return j.secretKey, nil
    })
    
    if err != nil {
        return nil, err
    }
    
    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        // 验证issuer和audience
        if claims["iss"] != j.issuer || claims["aud"] != j.audience {
            return nil, fmt.Errorf("invalid token claims")
        }
        
        return extractClaims(claims), nil
    }
    
    return nil, fmt.Errorf("invalid token")
}

type Claims struct {
    UserID    string   `json:"sub"`
    TenantID  string   `json:"tenant_id"`
    Roles     []string `json:"roles"`
    ExpiresAt int64    `json:"exp"`
}
```

**3.2 GraphQL权限装饰器** (4小时)
```go
// 新建文件: internal/auth/graphql_middleware.go

type GraphQLPermissionMiddleware struct {
    jwtMiddleware *JWTMiddleware
    permissionDB  PermissionRepository
}

func (g *GraphQLPermissionMiddleware) Middleware() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        // 提取Authorization头
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(401, gin.H{
                "success": false,
                "error": gin.H{
                    "code":    "UNAUTHORIZED",
                    "message": "Authorization header required",
                },
            })
            c.Abort()
            return
        }
        
        // 验证Bearer Token
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        claims, err := g.jwtMiddleware.ValidateToken(tokenString)
        if err != nil {
            c.JSON(401, gin.H{
                "success": false,
                "error": gin.H{
                    "code":    "INVALID_TOKEN",
                    "message": err.Error(),
                },
            })
            c.Abort()
            return
        }
        
        // 设置用户上下文
        c.Set("user_id", claims.UserID)
        c.Set("tenant_id", claims.TenantID)
        c.Set("user_roles", claims.Roles)
        
        c.Next()
    })
}

// GraphQL查询级权限检查
func (g *GraphQLPermissionMiddleware) CheckQueryPermission(ctx context.Context, queryName string) error {
    userID := ctx.Value("user_id").(string)
    tenantID := ctx.Value("tenant_id").(string)
    roles := ctx.Value("user_roles").([]string)
    
    // 检查查询权限
    hasPermission := g.permissionDB.CheckPermission(tenantID, userID, roles, queryName)
    if !hasPermission {
        return fmt.Errorf("insufficient permissions for query: %s", queryName)
    }
    
    return nil
}
```

**3.3 PBAC权限模型实现** (4小时)
```go
// 新建文件: internal/auth/pbac.go

type PBACPermissionChecker struct {
    db *sql.DB
}

// 权限检查主方法
func (p *PBACPermissionChecker) CheckPermission(tenantID, userID string, roles []string, resource string) bool {
    // 1. 检查直接用户权限
    if p.checkUserPermission(tenantID, userID, resource) {
        return true
    }
    
    // 2. 检查角色权限
    for _, role := range roles {
        if p.checkRolePermission(tenantID, role, resource) {
            return true
        }
    }
    
    // 3. 检查继承权限 (基于组织层级)
    if p.checkInheritedPermission(tenantID, userID, resource) {
        return true
    }
    
    return false
}

// GraphQL查询权限映射表
var GraphQLQueryPermissions = map[string]string{
    "organizations":         "READ_ORGANIZATION",
    "organization":          "READ_ORGANIZATION",
    "organizationHistory":   "READ_ORGANIZATION_HISTORY",
    "organizationHierarchy": "READ_ORGANIZATION_HIERARCHY",
    // 添加更多查询映射
}

func (p *PBACPermissionChecker) CheckGraphQLQuery(ctx context.Context, queryName string) error {
    tenantID := ctx.Value("tenant_id").(string)
    userID := ctx.Value("user_id").(string)
    roles := ctx.Value("user_roles").([]string)
    
    // 获取查询所需权限
    requiredPermission, exists := GraphQLQueryPermissions[queryName]
    if !exists {
        return fmt.Errorf("unknown query: %s", queryName)
    }
    
    // 执行权限检查
    if !p.CheckPermission(tenantID, userID, roles, requiredPermission) {
        return fmt.Errorf("access denied for query: %s", queryName)
    }
    
    return nil
}
```

**3.4 GraphQL服务集成** (2小时)
```go
// 修改文件: cmd/organization-query-service/main.go

func main() {
    // 初始化权限中间件
    jwtMiddleware := auth.NewJWTMiddleware(
        os.Getenv("JWT_SECRET"),
        os.Getenv("JWT_ISSUER"),
        os.Getenv("JWT_AUDIENCE"),
    )
    
    permissionChecker := auth.NewPBACPermissionChecker(db)
    graphqlMiddleware := auth.NewGraphQLPermissionMiddleware(jwtMiddleware, permissionChecker)
    
    // 设置路由
    router := gin.Default()
    
    // 应用权限中间件
    authorized := router.Group("/")
    authorized.Use(graphqlMiddleware.Middleware())
    {
        authorized.POST("/graphql", graphqlHandler)
        authorized.GET("/graphiql", graphiqlHandler)
    }
    
    log.Println("GraphQL Query Service starting on :8090 with JWT authentication")
    router.Run(":8090")
}
```

#### ✅ Day 3-4 完成标准
- [ ] JWT Token验证中间件实现
- [ ] OAuth服务Token验证集成
- [ ] GraphQL权限装饰器开发
- [ ] PBAC权限模型实现
- [ ] 权限映射表定义
- [ ] 租户隔离验证机制

## 🚀 第2阶段: 业务逻辑完善 (4-5天)

### Day 5-6: 智能层级管理实现

#### 🎯 任务目标
实现PostgreSQL递归查询的智能层级管理和级联更新机制

#### 📋 详细任务清单

**5.1 PostgreSQL递归CTE查询** (4小时)
```go
// 新建文件: internal/repository/hierarchy.go

type HierarchyRepository struct {
    db *sql.DB
}

// 获取组织层级结构 (递归查询)
func (h *HierarchyRepository) GetOrganizationHierarchy(ctx context.Context, rootCode string, tenantID string) ([]OrganizationNode, error) {
    query := `
    WITH RECURSIVE org_tree AS (
        -- 递归基准: 根组织
        SELECT 
            code, parent_code, name, level, code_path, name_path,
            effective_date, end_date, is_current,
            0 as depth
        FROM organization_units 
        WHERE code = $1 AND tenant_id = $2 AND is_current = true
        
        UNION ALL
        
        -- 递归部分: 子组织
        SELECT 
            ou.code, ou.parent_code, ou.name, ou.level, ou.code_path, ou.name_path,
            ou.effective_date, ou.end_date, ou.is_current,
            ot.depth + 1
        FROM organization_units ou
        INNER JOIN org_tree ot ON ou.parent_code = ot.code
        WHERE ou.tenant_id = $2 AND ou.is_current = true AND ot.depth < 17
    )
    SELECT * FROM org_tree ORDER BY depth, code;
    `
    
    rows, err := h.db.QueryContext(ctx, query, rootCode, tenantID)
    if err != nil {
        return nil, fmt.Errorf("failed to query organization hierarchy: %w", err)
    }
    defer rows.Close()
    
    var nodes []OrganizationNode
    for rows.Next() {
        var node OrganizationNode
        err := rows.Scan(
            &node.Code, &node.ParentCode, &node.Name, &node.Level,
            &node.CodePath, &node.NamePath, &node.EffectiveDate,
            &node.EndDate, &node.IsCurrent, &node.Depth,
        )
        if err != nil {
            return nil, err
        }
        nodes = append(nodes, node)
    }
    
    return nodes, nil
}

// 计算层级路径 (code_path, name_path)
func (h *HierarchyRepository) UpdateHierarchyPaths(ctx context.Context, parentCode string, tenantID string) error {
    // 获取父组织路径
    var parentCodePath, parentNamePath string
    err := h.db.QueryRowContext(ctx, `
        SELECT COALESCE(code_path, ''), COALESCE(name_path, '')
        FROM organization_units 
        WHERE code = $1 AND tenant_id = $2 AND is_current = true
    `, parentCode, tenantID).Scan(&parentCodePath, &parentNamePath)
    
    if err != nil && err != sql.ErrNoRows {
        return fmt.Errorf("failed to get parent paths: %w", err)
    }
    
    // 批量更新子组织路径
    updateQuery := `
    UPDATE organization_units SET
        code_path = CASE 
            WHEN $1 = '' THEN code
            ELSE $1 || '/' || code
        END,
        name_path = CASE
            WHEN $2 = '' THEN name
            ELSE $2 || '/' || name  
        END,
        level = CASE
            WHEN $1 = '' THEN 1
            ELSE array_length(string_to_array($1, '/'), 1) + 1
        END,
        updated_at = NOW()
    WHERE parent_code = $3 AND tenant_id = $4 AND is_current = true;
    `
    
    _, err = h.db.ExecContext(ctx, updateQuery, parentCodePath, parentNamePath, parentCode, tenantID)
    return err
}
```

**5.2 异步级联更新机制** (4小时)
```go
// 新建文件: internal/services/cascade.go

type CascadeUpdateService struct {
    repo      *HierarchyRepository
    taskQueue chan CascadeTask
    workers   int
}

type CascadeTask struct {
    Type      string
    Code      string
    TenantID  string
    UserID    string
    Context   context.Context
}

func NewCascadeUpdateService(repo *HierarchyRepository, workers int) *CascadeUpdateService {
    service := &CascadeUpdateService{
        repo:      repo,
        taskQueue: make(chan CascadeTask, 1000),
        workers:   workers,
    }
    
    // 启动工作协程
    for i := 0; i < workers; i++ {
        go service.worker()
    }
    
    return service
}

func (c *CascadeUpdateService) worker() {
    for task := range c.taskQueue {
        switch task.Type {
        case "UPDATE_HIERARCHY":
            c.processHierarchyUpdate(task)
        case "UPDATE_STATUS":
            c.processStatusUpdate(task)
        case "UPDATE_PATHS":
            c.processPathUpdate(task)
        }
    }
}

// 处理层级结构变更
func (c *CascadeUpdateService) processHierarchyUpdate(task CascadeTask) {
    ctx := task.Context
    
    // 获取所有子组织
    children, err := c.repo.GetDirectChildren(ctx, task.Code, task.TenantID)
    if err != nil {
        log.Printf("Failed to get children for %s: %v", task.Code, err)
        return
    }
    
    // 递归更新所有子组织的路径
    for _, child := range children {
        err := c.repo.UpdateHierarchyPaths(ctx, child.Code, task.TenantID)
        if err != nil {
            log.Printf("Failed to update paths for %s: %v", child.Code, err)
            continue
        }
        
        // 继续级联到下一层
        c.ScheduleHierarchyUpdate(child.Code, task.TenantID, task.UserID, ctx)
    }
}

// 调度层级更新任务
func (c *CascadeUpdateService) ScheduleHierarchyUpdate(code, tenantID, userID string, ctx context.Context) {
    task := CascadeTask{
        Type:     "UPDATE_HIERARCHY",
        Code:     code,
        TenantID: tenantID,
        UserID:   userID,
        Context:  ctx,
    }
    
    select {
    case c.taskQueue <- task:
        log.Printf("Scheduled hierarchy update for %s", code)
    default:
        log.Printf("Task queue full, dropping task for %s", code)
    }
}
```

**5.3 业务规则验证器** (4小时)
```go
// 新建文件: internal/validators/business.go

type BusinessRuleValidator struct {
    repo *HierarchyRepository
}

// 验证层级深度限制 (最大17级)
func (v *BusinessRuleValidator) ValidateDepthLimit(ctx context.Context, parentCode, tenantID string) error {
    if parentCode == "" {
        return nil // 根组织无深度限制
    }
    
    depth, err := v.repo.GetOrganizationDepth(ctx, parentCode, tenantID)
    if err != nil {
        return fmt.Errorf("failed to get organization depth: %w", err)
    }
    
    if depth >= 17 {
        return fmt.Errorf("maximum organization depth (17 levels) exceeded")
    }
    
    return nil
}

// 检测循环引用
func (v *BusinessRuleValidator) ValidateCircularReference(ctx context.Context, code, parentCode, tenantID string) error {
    if parentCode == "" {
        return nil // 根组织无循环引用风险
    }
    
    // 向上遍历检查是否存在循环
    currentParent := parentCode
    visited := make(map[string]bool)
    
    for currentParent != "" {
        if visited[currentParent] {
            return fmt.Errorf("circular reference detected")
        }
        
        if currentParent == code {
            return fmt.Errorf("organization cannot be parent of itself")
        }
        
        visited[currentParent] = true
        
        // 获取父组织的父组织
        nextParent, err := v.repo.GetParentCode(ctx, currentParent, tenantID)
        if err != nil {
            return fmt.Errorf("failed to validate hierarchy: %w", err)
        }
        
        currentParent = nextParent
    }
    
    return nil
}

// 层级一致性验证
func (v *BusinessRuleValidator) ValidateHierarchyConsistency(ctx context.Context, code, tenantID string) error {
    org, err := v.repo.GetOrganization(ctx, code, tenantID)
    if err != nil {
        return err
    }
    
    // 验证code_path一致性
    expectedCodePath, err := v.calculateCodePath(ctx, org.ParentCode, tenantID)
    if err != nil {
        return err
    }
    
    expectedCodePath += "/" + org.Code
    if org.CodePath != expectedCodePath {
        return fmt.Errorf("code_path inconsistency detected: expected %s, got %s", 
            expectedCodePath, org.CodePath)
    }
    
    return nil
}

// 综合业务规则验证
func (v *BusinessRuleValidator) ValidateOrganizationChange(ctx context.Context, req *UpdateOrganizationRequest) error {
    // 深度限制检查
    if err := v.ValidateDepthLimit(ctx, req.ParentCode, req.TenantID); err != nil {
        return err
    }
    
    // 循环引用检查  
    if err := v.ValidateCircularReference(ctx, req.Code, req.ParentCode, req.TenantID); err != nil {
        return err
    }
    
    // 层级一致性检查
    if err := v.ValidateHierarchyConsistency(ctx, req.Code, req.TenantID); err != nil {
        return err
    }
    
    return nil
}
```

#### ✅ Day 5-6 完成标准
- [ ] PostgreSQL递归CTE查询实现
- [ ] 层级路径自动计算和更新
- [ ] 异步级联处理机制
- [ ] 17级深度限制检查
- [ ] 循环引用检测算法  
- [ ] 层级一致性验证器

### Day 7-8: 审计监控体系

#### 🎯 任务目标
实现完整的操作审计日志和性能监控集成

#### 📋 详细任务清单

**7.1 结构化审计日志系统** (4小时)
```go
// 新建文件: internal/audit/logger.go

type AuditLogger struct {
    db     *sql.DB
    logger *log.Logger
}

type AuditRecord struct {
    AuditID         string                 `json:"auditId"`
    OperationType   string                 `json:"operationType"`
    OperatedBy      OperatedByInfo         `json:"operatedBy"`
    BusinessEntityID string                `json:"businessEntityId"`
    ChangesSummary  map[string]interface{} `json:"changesSummary"`
    OperationReason string                 `json:"operationReason"`
    TenantID        string                 `json:"tenantId"`
    Timestamp       time.Time              `json:"timestamp"`
    RequestID       string                 `json:"requestId"`
}

type OperatedByInfo struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

// 记录API操作审计
func (a *AuditLogger) LogAPIOperation(ctx context.Context, req *AuditRequest) error {
    record := AuditRecord{
        AuditID:          generateUUID(),
        OperationType:    req.OperationType,
        OperatedBy:       req.OperatedBy,
        BusinessEntityID: req.BusinessEntityID,
        ChangesSummary:   req.ChangesSummary,
        OperationReason:  req.OperationReason,
        TenantID:         req.TenantID,
        Timestamp:        time.Now().UTC(),
        RequestID:        getRequestID(ctx),
    }
    
    // 数据库存储
    err := a.saveToDatabase(ctx, record)
    if err != nil {
        a.logger.Printf("Failed to save audit record to database: %v", err)
    }
    
    // 结构化日志输出
    a.logToFile(record)
    
    return err
}

// PostgreSQL审计表存储
func (a *AuditLogger) saveToDatabase(ctx context.Context, record AuditRecord) error {
    query := `
    INSERT INTO audit_logs (
        audit_id, operation_type, operated_by_id, operated_by_name,
        business_entity_id, changes_summary, operation_reason,
        tenant_id, timestamp, request_id
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    `
    
    changesSummaryJSON, _ := json.Marshal(record.ChangesSummary)
    
    _, err := a.db.ExecContext(ctx, query,
        record.AuditID, record.OperationType,
        record.OperatedBy.ID, record.OperatedBy.Name,
        record.BusinessEntityID, string(changesSummaryJSON),
        record.OperationReason, record.TenantID,
        record.Timestamp, record.RequestID,
    )
    
    return err
}

// 结构化日志输出
func (a *AuditLogger) logToFile(record AuditRecord) {
    logData := map[string]interface{}{
        "level":             "INFO",
        "type":              "AUDIT",
        "audit_id":          record.AuditID,
        "operation_type":    record.OperationType,
        "operated_by":       record.OperatedBy,
        "business_entity_id": record.BusinessEntityID,
        "changes_summary":   record.ChangesSummary,
        "operation_reason":  record.OperationReason,
        "tenant_id":         record.TenantID,
        "timestamp":         record.Timestamp.Format(time.RFC3339),
        "request_id":        record.RequestID,
    }
    
    jsonData, _ := json.Marshal(logData)
    a.logger.Println(string(jsonData))
}
```

**7.2 审计数据库表结构** (2小时)
```sql
-- 新建文件: database/migrations/audit_schema.sql

-- 审计日志主表
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    audit_id VARCHAR(36) NOT NULL UNIQUE,
    operation_type VARCHAR(50) NOT NULL,
    operated_by_id VARCHAR(36) NOT NULL,
    operated_by_name VARCHAR(255) NOT NULL,
    business_entity_id VARCHAR(50) NOT NULL,
    changes_summary JSONB,
    operation_reason TEXT,
    tenant_id VARCHAR(36) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    request_id VARCHAR(36),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 索引优化
CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_timestamp 
ON audit_logs(tenant_id, timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_audit_logs_operation_type 
ON audit_logs(operation_type);

CREATE INDEX IF NOT EXISTS idx_audit_logs_business_entity 
ON audit_logs(business_entity_id);

CREATE INDEX IF NOT EXISTS idx_audit_logs_operated_by 
ON audit_logs(operated_by_id);

CREATE INDEX IF NOT EXISTS idx_audit_logs_request_id 
ON audit_logs(request_id);

-- GIN索引支持JSON查询
CREATE INDEX IF NOT EXISTS idx_audit_logs_changes_gin 
ON audit_logs USING GIN(changes_summary);
```

**7.3 Prometheus指标收集** (4小时)
```go
// 新建文件: internal/metrics/prometheus.go

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

type PrometheusMetrics struct {
    // HTTP请求指标
    HTTPRequestsTotal *prometheus.CounterVec
    HTTPDuration     *prometheus.HistogramVec
    
    // 业务操作指标
    OrganizationOperations *prometheus.CounterVec
    DatabaseConnections    prometheus.Gauge
    
    // 审计和错误指标
    AuditLogsTotal    *prometheus.CounterVec
    ErrorsTotal       *prometheus.CounterVec
}

func NewPrometheusMetrics() *PrometheusMetrics {
    return &PrometheusMetrics{
        HTTPRequestsTotal: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "cube_castle_http_requests_total",
                Help: "Total number of HTTP requests",
            },
            []string{"method", "endpoint", "status_code", "tenant_id"},
        ),
        
        HTTPDuration: promauto.NewHistogramVec(
            prometheus.HistogramOpts{
                Name:    "cube_castle_http_request_duration_seconds",
                Help:    "HTTP request duration in seconds",
                Buckets: []float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20},
            },
            []string{"method", "endpoint", "tenant_id"},
        ),
        
        OrganizationOperations: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "cube_castle_organization_operations_total",
                Help: "Total organization operations by type",
            },
            []string{"operation_type", "tenant_id", "status"},
        ),
        
        DatabaseConnections: promauto.NewGauge(
            prometheus.GaugeOpts{
                Name: "cube_castle_database_connections",
                Help: "Current database connections",
            },
        ),
        
        AuditLogsTotal: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "cube_castle_audit_logs_total",
                Help: "Total audit logs by operation type",
            },
            []string{"operation_type", "tenant_id"},
        ),
        
        ErrorsTotal: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "cube_castle_errors_total",
                Help: "Total errors by type and service",
            },
            []string{"error_type", "service", "tenant_id"},
        ),
    }
}

// HTTP中间件集成
func (m *PrometheusMetrics) HTTPMetricsMiddleware() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start).Seconds()
        statusCode := strconv.Itoa(c.Writer.Status())
        tenantID := c.GetString("tenant_id")
        
        m.HTTPRequestsTotal.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
            statusCode,
            tenantID,
        ).Inc()
        
        m.HTTPDuration.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
            tenantID,
        ).Observe(duration)
    })
}

// 业务操作指标记录
func (m *PrometheusMetrics) RecordOrganizationOperation(operationType, tenantID, status string) {
    m.OrganizationOperations.WithLabelValues(operationType, tenantID, status).Inc()
}

// 审计日志指标记录
func (m *PrometheusMetrics) RecordAuditLog(operationType, tenantID string) {
    m.AuditLogsTotal.WithLabelValues(operationType, tenantID).Inc()
}
```

**7.4 服务集成和启动配置** (2小时)
```go
// 修改文件: main.go (命令服务和查询服务)

func main() {
    // 初始化Prometheus指标
    metrics := metrics.NewPrometheusMetrics()
    
    // 初始化审计日志
    auditLogger := audit.NewAuditLogger(db, logger)
    
    // 设置路由
    router := gin.Default()
    
    // 中间件注册
    router.Use(RequestIDMiddleware())
    router.Use(metrics.HTTPMetricsMiddleware())
    router.Use(AuditMiddleware(auditLogger))
    
    // Prometheus metrics端点
    router.GET("/metrics", gin.WrapH(promhttp.Handler()))
    
    // 健康检查端点
    router.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "status": "healthy",
            "service": "cube-castle-api",
            "version": "1.0.0",
            "timestamp": time.Now().UTC().Format(time.RFC3339),
        })
    })
    
    // 业务路由
    api := router.Group("/api/v1")
    api.Use(JWTAuthMiddleware())
    {
        // 组织管理端点
        api.POST("/organization-units", createOrganization)
        api.PUT("/organization-units/:code", updateOrganization)
        api.POST("/organization-units/:code/suspend", suspendOrganization)
        api.POST("/organization-units/:code/activate", activateOrganization)
        api.DELETE("/organization-units/:code", deleteOrganization)
    }
    
    log.Printf("Starting service on :9090 with monitoring enabled")
    router.Run(":9090")
}
```

#### ✅ Day 7-8 完成标准
- [ ] 结构化审计日志系统实现
- [ ] PostgreSQL审计表结构创建
- [ ] operationType/operatedBy标准化实现
- [ ] Prometheus指标收集集成
- [ ] HTTP请求和业务操作指标
- [ ] 自定义业务指标定义
- [ ] /metrics和/health端点暴露

## 🚀 第3阶段: 集成测试与验证 (2-3天)

### Day 9-10: 端到端测试

#### 🎯 任务目标
建立完整的API规范符合性测试、安全认证测试和性能基准验证

#### 📋 详细任务清单

**9.1 API规范符合性自动化测试** (4小时)
```go
// 新建文件: tests/integration/api_compliance_test.go

package integration

import (
    "testing"
    "net/http"
    "encoding/json"
    "bytes"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
)

type APIComplianceTestSuite struct {
    suite.Suite
    baseURL     string
    authToken   string
    testTenantID string
}

func (suite *APIComplianceTestSuite) SetupSuite() {
    suite.baseURL = "http://localhost:9090/api/v1"
    suite.authToken = suite.getTestAuthToken()
    suite.testTenantID = "test-tenant-123"
}

// 测试REST端点HTTP方法符合性
func (suite *APIComplianceTestSuite) TestHTTPMethodCompliance() {
    tests := []struct {
        name       string
        method     string
        endpoint   string
        expectCode int
    }{
        {"Suspend Organization", "POST", "/organization-units/TEST001/suspend", 200},
        {"Activate Organization", "POST", "/organization-units/TEST001/activate", 200},
        {"Create Organization", "POST", "/organization-units", 201},
        {"Update Organization", "PUT", "/organization-units/TEST001", 200},
        {"Delete Organization", "DELETE", "/organization-units/TEST001", 200},
    }
    
    for _, test := range tests {
        suite.Run(test.name, func() {
            req := suite.createRequest(test.method, test.endpoint, nil)
            resp, err := http.DefaultClient.Do(req)
            assert.NoError(suite.T(), err)
            defer resp.Body.Close()
            
            // 验证HTTP方法被正确支持
            assert.NotEqual(suite.T(), http.StatusMethodNotAllowed, resp.StatusCode,
                "Endpoint %s should support %s method", test.endpoint, test.method)
        })
    }
}

// 测试企业级响应信封格式
func (suite *APIComplianceTestSuite) TestEnterpriseResponseEnvelope() {
    // 测试成功响应格式
    suite.Run("Success Response Format", func() {
        resp := suite.makeAuthenticatedRequest("GET", "/organization-units/TEST001", nil)
        defer resp.Body.Close()
        
        var response map[string]interface{}
        err := json.NewDecoder(resp.Body).Decode(&response)
        assert.NoError(suite.T(), err)
        
        // 验证企业级信封必需字段
        assert.Contains(suite.T(), response, "success")
        assert.Contains(suite.T(), response, "data")
        assert.Contains(suite.T(), response, "message")
        assert.Contains(suite.T(), response, "timestamp")
        assert.Contains(suite.T(), response, "requestId")
        
        assert.Equal(suite.T(), true, response["success"])
        assert.IsType(suite.T(), "", response["timestamp"])
        assert.IsType(suite.T(), "", response["requestId"])
    })
    
    // 测试错误响应格式
    suite.Run("Error Response Format", func() {
        resp := suite.makeAuthenticatedRequest("GET", "/organization-units/NONEXISTENT", nil)
        defer resp.Body.Close()
        
        var response map[string]interface{}
        json.NewDecoder(resp.Body).Decode(&response)
        
        assert.Equal(suite.T(), false, response["success"])
        assert.Contains(suite.T(), response, "error")
        assert.Contains(suite.T(), response, "timestamp")
        assert.Contains(suite.T(), response, "requestId")
        
        errorInfo := response["error"].(map[string]interface{})
        assert.Contains(suite.T(), errorInfo, "code")
        assert.Contains(suite.T(), errorInfo, "message")
    })
}

// 测试camelCase字段命名
func (suite *APIComplianceTestSuite) TestCamelCaseNaming() {
    resp := suite.makeAuthenticatedRequest("GET", "/organization-units/TEST001", nil)
    defer resp.Body.Close()
    
    var response map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&response)
    
    data := response["data"].(map[string]interface{})
    
    // 验证关键字段使用camelCase
    requiredCamelCaseFields := []string{
        "parentCode", "unitType", "createdAt", "updatedAt",
        "effectiveDate", "endDate", "operationType", "operatedBy",
    }
    
    for _, field := range requiredCamelCaseFields {
        assert.Contains(suite.T(), data, field, 
            "Field %s should use camelCase naming", field)
    }
    
    // 验证不存在snake_case字段
    forbiddenSnakeCaseFields := []string{
        "parent_code", "unit_type", "created_at", "updated_at",
        "effective_date", "end_date", "operation_type", "operated_by",
    }
    
    for _, field := range forbiddenSnakeCaseFields {
        assert.NotContains(suite.T(), data, field,
            "Field %s should not use snake_case naming", field)
    }
}
```

**9.2 安全认证集成测试** (4小时)
```go
// 新建文件: tests/security/oauth_pbac_test.go

type SecurityTestSuite struct {
    suite.Suite
    oauthServer *MockOAuthServer
    apiBaseURL  string
}

// OAuth 2.0流程端到端测试
func (suite *SecurityTestSuite) TestOAuth2Flow() {
    suite.Run("Valid JWT Token Authentication", func() {
        token := suite.generateValidJWT()
        
        req := suite.createRequestWithToken("GET", "/graphql", token, `{
            organizations(first: 10) {
                nodes { code name }
            }
        }`)
        
        resp, err := http.DefaultClient.Do(req)
        assert.NoError(suite.T(), err)
        assert.Equal(suite.T(), 200, resp.StatusCode)
    })
    
    suite.Run("Invalid JWT Token Rejection", func() {
        invalidToken := "invalid.jwt.token"
        
        req := suite.createRequestWithToken("GET", "/graphql", invalidToken, `{
            organizations { code }
        }`)
        
        resp, err := http.DefaultClient.Do(req)
        assert.NoError(suite.T(), err)
        assert.Equal(suite.T(), 401, resp.StatusCode)
    })
    
    suite.Run("Expired JWT Token Handling", func() {
        expiredToken := suite.generateExpiredJWT()
        
        req := suite.createRequestWithToken("GET", "/graphql", expiredToken, `{
            organizations { code }
        }`)
        
        resp, err := http.DefaultClient.Do(req)
        assert.NoError(suite.T(), err)
        assert.Equal(suite.T(), 401, resp.StatusCode)
    })
}

// PBAC权限矩阵验证
func (suite *SecurityTestSuite) TestPBACPermissionMatrix() {
    permissionTests := []struct {
        userRole     string
        operation    string
        resource     string
        expectAccess bool
    }{
        {"ADMIN", "READ_ORGANIZATION", "organizations", true},
        {"MANAGER", "READ_ORGANIZATION", "organizations", true},
        {"EMPLOYEE", "READ_ORGANIZATION", "organizations", true},
        {"ADMIN", "WRITE_ORGANIZATION", "organization-units", true},
        {"MANAGER", "WRITE_ORGANIZATION", "organization-units", false},
        {"EMPLOYEE", "WRITE_ORGANIZATION", "organization-units", false},
        {"ADMIN", "READ_ORGANIZATION_HISTORY", "organizationHistory", true},
        {"MANAGER", "READ_ORGANIZATION_HISTORY", "organizationHistory", true},
        {"EMPLOYEE", "READ_ORGANIZATION_HISTORY", "organizationHistory", false},
    }
    
    for _, test := range permissionTests {
        suite.Run(fmt.Sprintf("%s_%s_%s", test.userRole, test.operation, test.resource), func() {
            token := suite.generateJWTWithRole(test.userRole)
            
            var endpoint, method, body string
            if strings.HasPrefix(test.resource, "organization-units") {
                method = "POST"
                endpoint = "/api/v1/organization-units"
                body = `{"code": "TEST", "name": "Test"}`
            } else {
                method = "POST"
                endpoint = "/graphql"
                body = fmt.Sprintf(`{"query": "%s { code }", test.resource)
            }
            
            req := suite.createRequestWithToken(method, endpoint, token, body)
            resp, err := http.DefaultClient.Do(req)
            assert.NoError(suite.T(), err)
            
            if test.expectAccess {
                assert.NotEqual(suite.T(), 403, resp.StatusCode,
                    "User with role %s should have access to %s", test.userRole, test.operation)
            } else {
                assert.Equal(suite.T(), 403, resp.StatusCode,
                    "User with role %s should NOT have access to %s", test.userRole, test.operation)
            }
        })
    }
}
```

**9.3 性能基准验证测试** (4小时)
```go
// 新建文件: tests/performance/benchmark_test.go

type PerformanceTestSuite struct {
    suite.Suite
    baseURL   string
    authToken string
}

// GraphQL查询性能测试 (目标 <200ms)
func (suite *PerformanceTestSuite) TestGraphQLQueryPerformance() {
    queries := []struct {
        name        string
        query       string
        targetMS    int
        iterations  int
    }{
        {
            "Simple Organization List",
            `{ organizations(first: 10) { nodes { code name } } }`,
            200, 100,
        },
        {
            "Organization with History",
            `{ organization(code: "ROOT") { code name history(first: 5) { nodes { effectiveDate } } } }`,
            200, 50,
        },
        {
            "Organization Hierarchy",
            `{ organizationHierarchy(rootCode: "ROOT", maxDepth: 5) { code level children { code } } }`,
            300, 30,
        },
    }
    
    for _, test := range queries {
        suite.Run(test.name, func() {
            durations := make([]time.Duration, test.iterations)
            
            for i := 0; i < test.iterations; i++ {
                start := time.Now()
                
                resp := suite.makeGraphQLRequest(test.query)
                assert.Equal(suite.T(), 200, resp.StatusCode)
                resp.Body.Close()
                
                durations[i] = time.Since(start)
            }
            
            // 计算统计数据
            avgDuration := suite.calculateAverage(durations)
            p95Duration := suite.calculatePercentile(durations, 95)
            p99Duration := suite.calculatePercentile(durations, 99)
            
            suite.T().Logf("Query: %s", test.name)
            suite.T().Logf("Average: %.2fms", avgDuration.Seconds()*1000)
            suite.T().Logf("P95: %.2fms", p95Duration.Seconds()*1000)  
            suite.T().Logf("P99: %.2fms", p99Duration.Seconds()*1000)
            
            // 性能断言
            assert.Less(suite.T(), avgDuration.Milliseconds(), int64(test.targetMS),
                "Query %s average duration should be less than %dms", test.name, test.targetMS)
                
            assert.Less(suite.T(), p95Duration.Milliseconds(), int64(test.targetMS*2),
                "Query %s P95 duration should be less than %dms", test.name, test.targetMS*2)
        })
    }
}

// REST命令性能测试 (目标 <300ms)
func (suite *PerformanceTestSuite) TestRESTCommandPerformance() {
    commands := []struct {
        name       string
        method     string
        endpoint   string
        body       string
        targetMS   int
        iterations int
    }{
        {
            "Create Organization",
            "POST", "/organization-units",
            `{"code": "PERF001", "name": "Performance Test", "unitType": "DEPARTMENT"}`,
            300, 50,
        },
        {
            "Update Organization",
            "PUT", "/organization-units/PERF001",
            `{"name": "Updated Performance Test", "description": "Updated"}`,
            300, 50,
        },
        {
            "Suspend Organization",
            "POST", "/organization-units/PERF001/suspend",
            `{"reason": "Performance testing"}`,
            300, 50,
        },
    }
    
    for _, test := range commands {
        suite.Run(test.name, func() {
            durations := make([]time.Duration, test.iterations)
            
            for i := 0; i < test.iterations; i++ {
                start := time.Now()
                
                resp := suite.makeRESTRequest(test.method, test.endpoint, test.body)
                durations[i] = time.Since(start)
                
                // 验证响应状态
                assert.True(suite.T(), resp.StatusCode >= 200 && resp.StatusCode < 300,
                    "Command should return success status")
                resp.Body.Close()
            }
            
            avgDuration := suite.calculateAverage(durations)
            
            suite.T().Logf("Command: %s", test.name)
            suite.T().Logf("Average: %.2fms", avgDuration.Seconds()*1000)
            
            assert.Less(suite.T(), avgDuration.Milliseconds(), int64(test.targetMS),
                "Command %s duration should be less than %dms", test.name, test.targetMS)
        })
    }
}

// 并发负载测试
func (suite *PerformanceTestSuite) TestConcurrentLoad() {
    concurrencyLevels := []int{10, 50, 100}
    
    for _, concurrency := range concurrencyLevels {
        suite.Run(fmt.Sprintf("Concurrency_%d", concurrency), func() {
            var wg sync.WaitGroup
            durations := make(chan time.Duration, concurrency)
            
            for i := 0; i < concurrency; i++ {
                wg.Add(1)
                go func() {
                    defer wg.Done()
                    
                    start := time.Now()
                    resp := suite.makeGraphQLRequest(`{ organizations(first: 5) { nodes { code } } }`)
                    duration := time.Since(start)
                    
                    durations <- duration
                    assert.Equal(suite.T(), 200, resp.StatusCode)
                    resp.Body.Close()
                }()
            }
            
            wg.Wait()
            close(durations)
            
            var allDurations []time.Duration
            for d := range durations {
                allDurations = append(allDurations, d)
            }
            
            avgDuration := suite.calculateAverage(allDurations)
            suite.T().Logf("Concurrency %d - Average: %.2fms", concurrency, avgDuration.Seconds()*1000)
            
            // 并发下性能不应严重退化 (允许2倍延迟)
            assert.Less(suite.T(), avgDuration.Milliseconds(), int64(400),
                "Concurrent requests should complete within reasonable time")
        })
    }
}
```

#### ✅ Day 9-10 完成标准
- [ ] OpenAPI规范验证自动化测试
- [ ] GraphQL Schema一致性检查
- [ ] 企业级响应信封格式验证
- [ ] camelCase字段命名符合性测试
- [ ] OAuth 2.0完整流程端到端测试
- [ ] PBAC权限矩阵全面验证
- [ ] JWT Token生命周期测试
- [ ] GraphQL查询性能<200ms验证
- [ ] REST命令性能<300ms验证  
- [ ] 并发负载测试和性能基线

### Day 11-12: 部署配置完善

#### 🎯 任务目标
建立生产就绪的Docker配置、监控告警体系和运维脚本

#### 📋 详细任务清单

**11.1 生产环境Docker配置** (4小时)
```dockerfile
# 新建文件: docker-compose.production.yml

version: '3.8'

services:
  # REST命令服务
  cube-castle-command:
    build:
      context: .
      dockerfile: cmd/organization-command-service/Dockerfile
      target: production
    ports:
      - "9090:9090"
    environment:
      - ENV=production
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=${POSTGRES_DB}
      - DB_USER=${POSTGRES_USER}
      - DB_PASSWORD=${POSTGRES_PASSWORD}
      - REDIS_URL=redis://redis:6379
      - JWT_SECRET=${JWT_SECRET}
      - JWT_ISSUER=${JWT_ISSUER}
      - JWT_AUDIENCE=${JWT_AUDIENCE}
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9090/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    deploy:
      replicas: 2
      resources:
        limits:
          memory: 512M
          cpus: '0.5'
        reservations:
          memory: 256M
          cpus: '0.25'
    restart: unless-stopped
    logging:
      driver: json-file
      options:
        max-size: "10m"
        max-file: "3"
        labels: "service=cube-castle-command"

  # GraphQL查询服务  
  cube-castle-query:
    build:
      context: .
      dockerfile: cmd/organization-query-service/Dockerfile
      target: production
    ports:
      - "8090:8090"
    environment:
      - ENV=production
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=${POSTGRES_DB}
      - DB_USER=${POSTGRES_USER}
      - DB_PASSWORD=${POSTGRES_PASSWORD}
      - REDIS_URL=redis://redis:6379
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8090/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    deploy:
      replicas: 2
      resources:
        limits:
          memory: 512M
          cpus: '0.5'
    restart: unless-stopped

  # PostgreSQL数据库
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: ${POSTGRES_DB}
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./database/migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER} -d ${POSTGRES_DB}"]
      interval: 30s
      timeout: 10s
      retries: 5
    deploy:
      resources:
        limits:
          memory: 1G
          cpus: '1.0'
    restart: unless-stopped

  # Redis缓存
  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes --maxmemory 256mb --maxmemory-policy allkeys-lru
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3
    restart: unless-stopped

  # Prometheus监控
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9091:9090"
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--storage.tsdb.retention.time=168h'
      - '--web.enable-lifecycle'
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - ./monitoring/rules:/etc/prometheus/rules
      - prometheus_data:/prometheus
    restart: unless-stopped

  # Grafana仪表板
  grafana:
    image: grafana/grafana:latest
    ports:
      - "3001:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_ADMIN_PASSWORD}
      - GF_USERS_ALLOW_SIGN_UP=false
    volumes:
      - grafana_data:/var/lib/grafana
      - ./monitoring/grafana/dashboards:/var/lib/grafana/dashboards
      - ./monitoring/grafana/provisioning:/etc/grafana/provisioning
    depends_on:
      - prometheus
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
  prometheus_data:
  grafana_data:

networks:
  default:
    driver: bridge
```

**11.2 多阶段Docker构建优化** (2小时)
```dockerfile
# 新建文件: cmd/organization-command-service/Dockerfile

# 构建阶段
FROM golang:1.21-alpine AS builder
WORKDIR /app

# 安装依赖
RUN apk add --no-cache git ca-certificates tzdata

# 复制go.mod和go.sum
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建二进制文件
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main ./cmd/organization-command-service

# 生产阶段
FROM alpine:latest AS production
RUN apk --no-cache add ca-certificates curl
WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup
USER appuser

# 暴露端口
EXPOSE 9090

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:9090/health || exit 1

# 启动服务
CMD ["./main"]
```

**11.3 Prometheus告警规则配置** (3小时)
```yaml
# 新建文件: monitoring/prometheus-rules.yml

groups:
- name: cube-castle-alerts
  rules:
  # HTTP错误率告警
  - alert: HighHTTPErrorRate
    expr: rate(cube_castle_http_requests_total{status_code=~"5.."}[5m]) > 0.1
    for: 2m
    labels:
      severity: critical
      service: cube-castle
    annotations:
      summary: "High HTTP error rate detected"
      description: "HTTP error rate is {{ $value | humanizePercentage }} for service {{ $labels.service }}"

  # 响应时间告警
  - alert: HighResponseTime
    expr: histogram_quantile(0.95, rate(cube_castle_http_request_duration_seconds_bucket[5m])) > 0.5
    for: 3m
    labels:
      severity: warning
      service: cube-castle
    annotations:
      summary: "High response time detected"
      description: "95th percentile response time is {{ $value }}s for {{ $labels.method }} {{ $labels.endpoint }}"

  # 数据库连接告警
  - alert: DatabaseConnectionIssue
    expr: cube_castle_database_connections < 1
    for: 1m
    labels:
      severity: critical
      service: cube-castle
    annotations:
      summary: "Database connection issue"
      description: "No active database connections available"

  # 服务可用性告警
  - alert: ServiceDown
    expr: up{job=~"cube-castle-.*"} == 0
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "{{ $labels.job }} service is down"
      description: "{{ $labels.job }} service has been down for more than 1 minute"

  # 审计日志异常告警
  - alert: AuditLogFailure
    expr: increase(cube_castle_errors_total{error_type="audit_failure"}[5m]) > 5
    for: 2m
    labels:
      severity: warning
      service: cube-castle
    annotations:
      summary: "Audit log failures detected"
      description: "{{ $value }} audit log failures in the last 5 minutes"

  # 权限验证失败告警
  - alert: AuthenticationFailures
    expr: increase(cube_castle_errors_total{error_type="auth_failure"}[5m]) > 10
    for: 2m
    labels:
      severity: warning
      service: cube-castle-auth
    annotations:
      summary: "High authentication failure rate"
      description: "{{ $value }} authentication failures in the last 5 minutes"

  # 内存使用告警
  - alert: HighMemoryUsage
    expr: (container_memory_usage_bytes / container_spec_memory_limit_bytes) * 100 > 80
    for: 3m
    labels:
      severity: warning
    annotations:
      summary: "High memory usage"
      description: "Memory usage is {{ $value | humanizePercentage }} for container {{ $labels.name }}"

  # CPU使用告警  
  - alert: HighCPUUsage
    expr: rate(container_cpu_usage_seconds_total[5m]) * 100 > 80
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High CPU usage"
      description: "CPU usage is {{ $value | humanizePercentage }} for container {{ $labels.name }}"
```

**11.4 Grafana仪表板模板** (3小时)
```json
// 新建文件: monitoring/grafana-dashboards.json

{
  "dashboard": {
    "id": null,
    "title": "Cube Castle - API Performance Dashboard",
    "tags": ["cube-castle", "api", "performance"],
    "timezone": "browser",
    "refresh": "30s",
    "panels": [
      {
        "id": 1,
        "title": "HTTP Request Rate",
        "type": "graph",
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 0},
        "targets": [
          {
            "expr": "rate(cube_castle_http_requests_total[5m])",
            "legendFormat": "{{ method }} {{ endpoint }}",
            "refId": "A"
          }
        ],
        "yAxes": [
          {
            "label": "Requests/sec",
            "min": 0
          }
        ]
      },
      {
        "id": 2,
        "title": "Response Time Percentiles",
        "type": "graph", 
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 0},
        "targets": [
          {
            "expr": "histogram_quantile(0.50, rate(cube_castle_http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "P50",
            "refId": "A"
          },
          {
            "expr": "histogram_quantile(0.95, rate(cube_castle_http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "P95", 
            "refId": "B"
          },
          {
            "expr": "histogram_quantile(0.99, rate(cube_castle_http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "P99",
            "refId": "C"
          }
        ]
      },
      {
        "id": 3,
        "title": "Error Rate by Status Code",
        "type": "graph",
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 8},
        "targets": [
          {
            "expr": "rate(cube_castle_http_requests_total{status_code=~\"4..\"}[5m])",
            "legendFormat": "4xx Errors",
            "refId": "A"
          },
          {
            "expr": "rate(cube_castle_http_requests_total{status_code=~\"5..\"}[5m])",
            "legendFormat": "5xx Errors", 
            "refId": "B"
          }
        ]
      },
      {
        "id": 4,
        "title": "Database Connections",
        "type": "singlestat",
        "gridPos": {"h": 4, "w": 6, "x": 12, "y": 8},
        "targets": [
          {
            "expr": "cube_castle_database_connections",
            "refId": "A"
          }
        ],
        "thresholds": "5,10",
        "colors": ["#d44a3a", "#e24d42", "#299c46"]
      },
      {
        "id": 5,
        "title": "Organization Operations",
        "type": "graph",
        "gridPos": {"h": 8, "w": 24, "x": 0, "y": 16},
        "targets": [
          {
            "expr": "rate(cube_castle_organization_operations_total[5m])",
            "legendFormat": "{{ operation_type }} - {{ status }}",
            "refId": "A"
          }
        ]
      },
      {
        "id": 6,
        "title": "Audit Logs Volume",
        "type": "graph",
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 24},
        "targets": [
          {
            "expr": "rate(cube_castle_audit_logs_total[5m])",
            "legendFormat": "{{ operation_type }}",
            "refId": "A"
          }
        ]
      },
      {
        "id": 7,
        "title": "System Resources",
        "type": "graph",
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 24},
        "targets": [
          {
            "expr": "rate(container_cpu_usage_seconds_total{name=~\"cube-castle.*\"}[5m]) * 100",
            "legendFormat": "CPU % - {{ name }}",
            "refId": "A"
          },
          {
            "expr": "(container_memory_usage_bytes{name=~\"cube-castle.*\"} / container_spec_memory_limit_bytes) * 100",
            "legendFormat": "Memory % - {{ name }}",
            "refId": "B"
          }
        ]
      }
    ]
  }
}
```

#### ✅ Day 11-12 完成标准
- [ ] 生产环境Docker Compose配置
- [ ] 多阶段构建优化和镜像体积压缩
- [ ] 环境变量标准化管理
- [ ] 健康检查和存活探针配置
- [ ] Prometheus告警规则完整配置
- [ ] Grafana仪表板模板创建
- [ ] 日志聚合和分析配置
- [ ] 服务发现和负载均衡配置

## 📊 后端团队成功指标

### 技术指标达成标准
```yaml
API服务指标:
  - REST命令服务可用性: >99.5%
  - GraphQL查询服务响应时间: <200ms (P95)
  - 企业级响应信封实现: 100%端点
  - JWT权限验证覆盖: 100%受保护端点

数据层指标:
  - PostgreSQL查询优化: 时态查询<150ms
  - 层级管理性能: 17级深度递归<300ms
  - 审计日志完整性: 100%操作记录
  - 数据一致性保证: 零数据丢失事件

监控指标:
  - Prometheus指标收集: 15+核心业务指标
  - 告警规则覆盖: API/数据库/资源/安全四大类
  - 日志结构化率: 100%业务操作日志
  - 性能基准建立: GraphQL/REST/层级管理基线
```

### 质量保证标准
```yaml
代码质量:
  - 单元测试覆盖率: >85% (核心业务逻辑)
  - 集成测试通过率: 100% (API规范符合性)
  - 安全测试通过: OAuth 2.0 + PBAC完整验证
  - 性能测试达标: 100%端点满足SLA要求

架构质量:
  - CQRS架构完整性: 查询/命令完全分离
  - PostgreSQL单一数据源: 零数据同步延迟
  - 企业级标准实现: 响应信封/审计/权限齐全
  - 生产就绪配置: Docker/监控/告警完整部署
```

### 协作成果指标
```yaml
与前端团队协作:
  - API规范一致性: 100%字段命名和响应格式统一
  - 接口集成顺畅度: 零API不兼容阻塞问题
  - 权限验证联调: 前后端JWT流程100%打通
  - 错误处理统一: 企业级错误信息标准化

文档和交付:
  - API文档完整性: 100%端点文档化
  - 部署脚本就绪: 一键生产环境部署
  - 监控配置齐全: 告警/仪表板/日志分析完整
  - 运维手册提供: 故障排查和维护指南
```

## 🎯 执行策略和风险控制

### 每日执行节奏
```yaml
每日站会 (9:00-9:15):
  - 前一日完成进度汇报
  - 当日具体任务分工
  - 技术难点和阻塞问题讨论
  - 与前端团队协作点确认

技术评审 (17:00-17:30):
  - 代码质量和架构一致性检查
  - API规范符合度验证
  - 安全和性能问题识别
  - 下一日任务优先级调整
```

### 风险预防措施
```yaml
技术风险:
  - PostgreSQL性能调优: 预先进行查询计划分析
  - JWT集成复杂性: 提前验证OAuth服务兼容性
  - 监控配置错误: 使用成熟的Prometheus规则模板
  - 并发性能问题: 早期进行负载测试验证

协作风险:
  - 前后端API不兼容: 每日验证企业级响应信封格式
  - 权限验证联调失败: 提前mock JWT验证流程
  - 字段命名不一致: 建立自动化camelCase检查
  - 进度不同步: 使用共享看板实时跟踪进度
```

### 质量保证机制
```yaml
代码质量:
  - 每个功能模块完成后立即进行单元测试
  - 每日代码提交前运行完整测试套件
  - 使用Go静态分析工具检查代码质量
  - 所有API端点必须通过集成测试

架构质量:
  - 每个服务启动后验证健康检查端点
  - API响应格式必须符合企业级信封标准
  - 权限验证逻辑必须通过PBAC矩阵测试
  - 审计日志必须包含标准化字段结构
```

---

**制定者**: 后端架构师  
**执行团队**: 后端开发团队  
**协作团队**: 前端开发团队  
**执行时间**: 2025-08-24 开始  
**预计完成**: 2025-09-06 (12个工作日)