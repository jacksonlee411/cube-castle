# 🏰 Cube-Castle 多租户用户管理与登录方案

## 📊 现状分析

**当前项目已具备的多租户基础：**
- ✅ 所有核心实体（Employee、OrganizationUnit）都包含 `tenant_id` 字段
- ✅ 数据库级别的多租户隔离索引已建立
- ✅ 基础的 Middleware 框架已存在（TenantContext、RBAC）
- ✅ CQRS + Event Sourcing 架构支持多租户事件流
- ✅ PostgreSQL RLS（行级安全）+ OPA 策略引擎

**缺失的用户管理组件：**
- ❌ 缺少 User/Account 实体模型
- ❌ 缺少 Tenant 实体和租户管理
- ❌ 缺少完整的身份认证和会话管理
- ❌ 中间件实现为占位符，需要完整实现

## 🎯 多租户用户管理数据模型设计

### 🏗️ 核心实体设计

#### 1. **Tenant** 实体（租户管理）
```go
// Tenant holds the schema definition for the Tenant entity.
type Tenant struct {
    ent.Schema
}

func (Tenant) Fields() []ent.Field {
    return []ent.Field{
        // Core Identity
        field.UUID("id", uuid.UUID{}).
            Default(uuid.New).
            Immutable(),
        
        // Business ID (User-friendly identifier)
        field.String("business_id").
            Unique().
            NotEmpty().
            MaxLen(8).
            Match(regexp.MustCompile(`^[A-Z0-9]{4,8}$`)),
        
        // Tenant Information
        field.String("name").
            NotEmpty().
            MaxLen(100),
        
        field.String("domain").
            Unique().
            NotEmpty().
            MaxLen(100),
        
        field.Enum("subscription_type").
            Values("FREE", "BASIC", "PROFESSIONAL", "ENTERPRISE"),
        
        field.Enum("status").
            Values("ACTIVE", "SUSPENDED", "TRIAL", "EXPIRED").
            Default("TRIAL"),
        
        // Tenant Configuration
        field.JSON("settings", map[string]interface{}{}).
            Optional(),
        
        // Billing & Limits
        field.Int("max_users").
            Default(10),
        
        field.Time("trial_expires_at").
            Optional().
            Nillable(),
        
        // Audit Trail
        field.Time("created_at").
            Default(time.Now).
            Immutable(),
        
        field.Time("updated_at").
            Default(time.Now).
            UpdateDefault(time.Now),
    }
}
```

#### 2. **User** 实体（用户账户）
```go
// User holds the schema definition for the User entity.
type User struct {
    ent.Schema
}

func (User) Fields() []ent.Field {
    return []ent.Field{
        // Core Identity
        field.UUID("id", uuid.UUID{}).
            Default(uuid.New).
            Immutable(),
        
        // Business ID (User-friendly identifier)
        field.String("business_id").
            Unique().
            NotEmpty().
            MaxLen(8).
            Match(regexp.MustCompile(`^[U][0-9]{7}$`)), // U1234567
        
        // Multi-tenant Association
        field.UUID("tenant_id", uuid.UUID{}).
            Immutable(),
        
        // Authentication Credentials
        field.String("email").
            NotEmpty().
            MaxLen(255),
        
        field.String("password_hash").
            NotEmpty().
            Sensitive(), // 敏感字段标记
        
        field.String("salt").
            NotEmpty().
            Sensitive(),
        
        // User Profile
        field.String("first_name").
            NotEmpty().
            MaxLen(50),
        
        field.String("last_name").
            NotEmpty().
            MaxLen(50),
        
        field.String("display_name").
            Optional().
            MaxLen(100),
        
        field.String("avatar_url").
            Optional().
            MaxLen(500),
        
        // Account Status & Security
        field.Enum("status").
            Values("ACTIVE", "INACTIVE", "LOCKED", "PENDING_VERIFICATION").
            Default("PENDING_VERIFICATION"),
        
        field.Bool("email_verified").
            Default(false),
        
        field.Time("last_login_at").
            Optional().
            Nillable(),
        
        field.Int("failed_login_attempts").
            Default(0),
        
        field.Time("locked_until").
            Optional().
            Nillable(),
        
        // Two-Factor Authentication
        field.Bool("two_factor_enabled").
            Default(false),
        
        field.String("two_factor_secret").
            Optional().
            Sensitive(),
        
        // Preferences & Settings
        field.String("preferred_language").
            Default("zh-CN"),
        
        field.String("timezone").
            Default("Asia/Shanghai"),
        
        field.JSON("preferences", map[string]interface{}{}).
            Optional(),
        
        // Relationship to Employee (Optional)
        field.UUID("employee_id", uuid.UUID{}).
            Optional().
            Nillable(),
        
        // Audit Trail
        field.Time("created_at").
            Default(time.Now).
            Immutable(),
        
        field.Time("updated_at").
            Default(time.Now).
            UpdateDefault(time.Now),
    }
}

func (User) Edges() []ent.Edge {
    return []ent.Edge{
        // Tenant Relationship
        edge.From("tenant", Tenant.Type).
            Field("tenant_id").
            Ref("users").
            Unique().
            Required(),
        
        // Employee Relationship (Optional)
        edge.From("employee", Employee.Type).
            Field("employee_id").
            Ref("user_account").
            Unique(),
        
        // Role Assignments
        edge.To("role_assignments", UserRole.Type),
        
        // Sessions
        edge.To("sessions", UserSession.Type),
    }
}
```

#### 3. **Role** & **UserRole** 实体（角色权限）
```go
// Role holds the schema definition for the Role entity.
type Role struct {
    ent.Schema
}

func (Role) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).
            Default(uuid.New).
            Immutable(),
        
        field.UUID("tenant_id", uuid.UUID{}).
            Immutable(),
        
        field.String("name").
            NotEmpty().
            MaxLen(50),
        
        field.String("description").
            Optional().
            MaxLen(200),
        
        field.Enum("role_type").
            Values("SYSTEM", "TENANT", "CUSTOM").
            Default("CUSTOM"),
        
        field.JSON("permissions", []string{}).
            Comment("权限列表"),
        
        field.Bool("is_active").
            Default(true),
        
        field.Time("created_at").
            Default(time.Now).
            Immutable(),
    }
}

// UserRole holds the schema definition for the UserRole entity.
type UserRole struct {
    ent.Schema
}

func (UserRole) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).
            Default(uuid.New).
            Immutable(),
        
        field.UUID("user_id", uuid.UUID{}),
        field.UUID("role_id", uuid.UUID{}),
        field.UUID("tenant_id", uuid.UUID{}).
            Immutable(),
        
        // 角色分配范围（可选）
        field.UUID("scope_organization_id", uuid.UUID{}).
            Optional().
            Nillable().
            Comment("角色作用域：限定到特定组织单位"),
        
        field.Time("assigned_at").
            Default(time.Now).
            Immutable(),
        
        field.Time("expires_at").
            Optional().
            Nillable(),
    }
}
```

#### 4. **UserSession** 实体（会话管理）
```go
// UserSession holds the schema definition for the UserSession entity.
type UserSession struct {
    ent.Schema
}

func (UserSession) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).
            Default(uuid.New).
            Immutable(),
        
        field.UUID("user_id", uuid.UUID{}),
        field.UUID("tenant_id", uuid.UUID{}).
            Immutable(),
        
        // Session Token
        field.String("session_token").
            Unique().
            NotEmpty().
            Sensitive(),
        
        field.String("refresh_token").
            Optional().
            Sensitive(),
        
        // Session Metadata
        field.String("ip_address").
            Optional().
            MaxLen(45), // IPv6 support
        
        field.String("user_agent").
            Optional().
            MaxLen(500),
        
        field.String("device_info").
            Optional().
            MaxLen(200),
        
        // Session Lifecycle
        field.Time("created_at").
            Default(time.Now).
            Immutable(),
        
        field.Time("last_accessed_at").
            Default(time.Now).
            UpdateDefault(time.Now),
        
        field.Time("expires_at").
            Comment("会话过期时间"),
        
        field.Bool("is_active").
            Default(true),
        
        field.Enum("session_type").
            Values("WEB", "MOBILE", "API", "SSO").
            Default("WEB"),
    }
}
```

## 🔐 用户登录认证流程方案

### 🚀 三层认证架构

#### **第一层：租户识别 (Tenant Discovery)**
```go
// 1. 域名识别方式
// https://acme-corp.cubecastle.com -> tenant: acme-corp
// https://cubecastle.com/login?tenant=acme-corp -> tenant: acme-corp

// 2. 子域名路由中间件
func TenantDiscoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        var tenantIdentifier string
        
        // 方式1：从子域名提取
        host := r.Host
        if subdomain := extractSubdomain(host); subdomain != "" {
            tenantIdentifier = subdomain
        }
        
        // 方式2：从查询参数提取
        if tenant := r.URL.Query().Get("tenant"); tenant != "" {
            tenantIdentifier = tenant
        }
        
        // 方式3：从路径提取 /t/{tenant}/...
        if tenant := extractTenantFromPath(r.URL.Path); tenant != "" {
            tenantIdentifier = tenant
        }
        
        if tenantIdentifier == "" {
            http.Error(w, "Tenant identifier required", http.StatusBadRequest)
            return
        }
        
        // 验证租户是否存在且活跃
        tenant, err := validateTenant(tenantIdentifier)
        if err != nil {
            http.Error(w, "Invalid tenant", http.StatusNotFound)
            return
        }
        
        // 将租户信息注入上下文
        ctx := context.WithValue(r.Context(), "tenant", tenant)
        ctx = context.WithValue(ctx, "tenant_id", tenant.ID)
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

#### **第二层：用户认证 (User Authentication)**
```go
// 登录流程
type AuthService struct {
    userRepo    *repository.UserRepository
    sessionRepo *repository.SessionRepository
    jwtService  *jwt.Service
    logger      *zap.Logger
}

type LoginRequest struct {
    Email       string `json:"email" validate:"required,email"`
    Password    string `json:"password" validate:"required,min=8"`
    TenantID    string `json:"tenant_id" validate:"required"`
    DeviceInfo  string `json:"device_info,omitempty"`
    RememberMe  bool   `json:"remember_me"`
    TwoFactorCode string `json:"two_factor_code,omitempty"`
}

type LoginResponse struct {
    AccessToken  string    `json:"access_token"`
    RefreshToken string    `json:"refresh_token"`
    ExpiresAt    time.Time `json:"expires_at"`
    User         UserInfo  `json:"user"`
    Permissions  []string  `json:"permissions"`
}

func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
    tenantID := GetTenantID(ctx)
    
    // 1. 查找用户（租户作用域内）
    user, err := s.userRepo.FindByEmailAndTenant(ctx, req.Email, tenantID)
    if err != nil {
        s.logFailedAttempt(req.Email, tenantID, "user_not_found")
        return nil, ErrInvalidCredentials
    }
    
    // 2. 检查账户状态
    if user.Status != "ACTIVE" {
        return nil, ErrAccountLocked
    }
    
    // 3. 验证密码
    if !s.verifyPassword(req.Password, user.PasswordHash, user.Salt) {
        s.handleFailedLogin(user)
        return nil, ErrInvalidCredentials
    }
    
    // 4. 二次认证验证（如果启用）
    if user.TwoFactorEnabled {
        if req.TwoFactorCode == "" {
            return &LoginResponse{RequiresTwoFactor: true}, nil
        }
        if !s.verifyTwoFactor(req.TwoFactorCode, user.TwoFactorSecret) {
            return nil, ErrInvalidTwoFactor
        }
    }
    
    // 5. 创建会话
    session, err := s.createUserSession(ctx, user, req)
    if err != nil {
        return nil, err
    }
    
    // 6. 生成JWT令牌
    accessToken, err := s.jwtService.GenerateAccessToken(user, session)
    refreshToken, err := s.jwtService.GenerateRefreshToken(session)
    
    // 7. 更新用户最后登录时间
    s.userRepo.UpdateLastLogin(ctx, user.ID)
    
    // 8. 加载用户权限
    permissions, err := s.loadUserPermissions(ctx, user.ID, tenantID)
    
    return &LoginResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresAt:    session.ExpiresAt,
        User:         s.buildUserInfo(user),
        Permissions:  permissions,
    }, nil
}
```

#### **第三层：会话管理 (Session Management)**
```go
// JWT 令牌结构
type JWTClaims struct {
    UserID     string    `json:"user_id"`
    TenantID   string    `json:"tenant_id"`
    SessionID  string    `json:"session_id"`
    Email      string    `json:"email"`
    Roles      []string  `json:"roles"`
    Permissions []string `json:"permissions"`
    IssuedAt   time.Time `json:"iat"`
    ExpiresAt  time.Time `json:"exp"`
    jwt.StandardClaims
}

// 会话验证中间件
func JWTAuthMiddleware(jwtService *jwt.Service) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 1. 提取 Token
            token := extractTokenFromHeader(r)
            if token == "" {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            
            // 2. 验证 JWT 签名
            claims, err := jwtService.ValidateToken(token)
            if err != nil {
                http.Error(w, "Invalid token", http.StatusUnauthorized)
                return
            }
            
            // 3. 验证会话是否仍然有效
            session, err := validateSession(claims.SessionID)
            if err != nil || !session.IsActive {
                http.Error(w, "Session expired", http.StatusUnauthorized)
                return
            }
            
            // 4. 更新会话最后访问时间
            updateSessionLastAccessed(claims.SessionID)
            
            // 5. 注入用户信息到上下文
            ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
            ctx = context.WithValue(ctx, "tenant_id", claims.TenantID)
            ctx = context.WithValue(ctx, "session_id", claims.SessionID)
            ctx = context.WithValue(ctx, "permissions", claims.Permissions)
            
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

## 🛡️ 权限管理和数据隔离策略

### 🔒 四层权限控制模型

#### **Layer 1: 租户隔离层 (Tenant Isolation)**
```sql
-- PostgreSQL 行级安全策略 (RLS)
-- 自动为所有表启用租户隔离

-- 用户表 RLS 策略
CREATE POLICY tenant_isolation_users ON users 
    FOR ALL TO application_role 
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

-- 员工表 RLS 策略  
CREATE POLICY tenant_isolation_employees ON employees 
    FOR ALL TO application_role 
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

-- 组织单位表 RLS 策略
CREATE POLICY tenant_isolation_organization_units ON organization_units 
    FOR ALL TO application_role 
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);
```

#### **Layer 2: RBAC 角色层 (Role-Based Access Control)**
```go
// 预定义系统角色
const (
    // 系统级角色
    RoleSuperAdmin     = "SUPER_ADMIN"      // 跨租户系统管理员
    RoleSystemAnalyst  = "SYSTEM_ANALYST"   // 系统分析师
    
    // 租户级角色
    RoleTenantAdmin    = "TENANT_ADMIN"     // 租户管理员
    RoleHRManager      = "HR_MANAGER"       // 人力资源经理
    RoleHRSpecialist   = "HR_SPECIALIST"    // 人力资源专员
    RoleLineManager    = "LINE_MANAGER"     // 直线经理
    RoleEmployee       = "EMPLOYEE"         // 普通员工
    RoleGuest          = "GUEST"            // 访客用户
)

// 权限定义
const (
    // 用户管理权限
    PermUserCreate     = "user:create"
    PermUserRead       = "user:read"
    PermUserUpdate     = "user:update"
    PermUserDelete     = "user:delete"
    PermUserManageRole = "user:manage_role"
    
    // 员工管理权限
    PermEmployeeCreate = "employee:create"
    PermEmployeeRead   = "employee:read"
    PermEmployeeUpdate = "employee:update"
    PermEmployeeDelete = "employee:delete"
    
    // 组织管理权限
    PermOrgCreate      = "organization:create"
    PermOrgRead        = "organization:read"
    PermOrgUpdate      = "organization:update"
    PermOrgDelete      = "organization:delete"
    
    // 系统管理权限
    PermSystemConfig   = "system:config"
    PermSystemAudit    = "system:audit"
    PermTenantManage   = "tenant:manage"
)

// 角色权限矩阵
var RolePermissions = map[string][]string{
    RoleSuperAdmin: {
        PermTenantManage, PermSystemConfig, PermSystemAudit,
        PermUserCreate, PermUserRead, PermUserUpdate, PermUserDelete, PermUserManageRole,
    },
    RoleTenantAdmin: {
        PermUserCreate, PermUserRead, PermUserUpdate, PermUserDelete, PermUserManageRole,
        PermEmployeeCreate, PermEmployeeRead, PermEmployeeUpdate, PermEmployeeDelete,
        PermOrgCreate, PermOrgRead, PermOrgUpdate, PermOrgDelete,
    },
    RoleHRManager: {
        PermUserRead, PermUserUpdate,
        PermEmployeeCreate, PermEmployeeRead, PermEmployeeUpdate, PermEmployeeDelete,
        PermOrgRead, PermOrgUpdate,
    },
    RoleHRSpecialist: {
        PermEmployeeCreate, PermEmployeeRead, PermEmployeeUpdate,
        PermOrgRead,
    },
    RoleLineManager: {
        PermEmployeeRead, PermEmployeeUpdate, // 限定范围内
        PermOrgRead,
    },
    RoleEmployee: {
        PermEmployeeRead, // 仅自己的记录
    },
}
```

#### **Layer 3: 资源级权限 (Resource-Level Authorization)**
```go
// 基于 OPA (Open Policy Agent) 的细粒度权限控制
// policy/authorization.rego

package authorization

import rego.v1

# 默认拒绝所有请求
default allow := false

# 租户隔离：用户只能访问自己租户的数据
tenant_isolated if {
    input.user.tenant_id == input.resource.tenant_id
}

# 管理员权限：租户管理员可以访问租户内所有资源
admin_access if {
    "TENANT_ADMIN" in input.user.roles
    tenant_isolated
}

# HR 经理权限：可以管理所有员工和组织数据
hr_manager_access if {
    "HR_MANAGER" in input.user.roles
    tenant_isolated
    input.resource.type in ["employee", "organization"]
}

# 直线经理权限：只能管理自己部门的员工
line_manager_access if {
    "LINE_MANAGER" in input.user.roles
    tenant_isolated
    input.resource.type == "employee"
    input.resource.department_id in input.user.managed_departments
}

# 员工自我访问权限：员工只能访问自己的记录
self_access if {
    input.user.employee_id == input.resource.employee_id
    input.action in ["read", "update_profile"]
}

# 综合权限判断
allow if admin_access
allow if hr_manager_access  
allow if line_manager_access
allow if self_access

# 特殊权限：系统管理员跨租户访问
allow if {
    "SUPER_ADMIN" in input.user.roles
    input.action != "delete_tenant"  # 即使超管也不能删除租户
}
```

#### **Layer 4: 字段级权限 (Field-Level Security)**
```go
// 敏感字段访问控制
type FieldAccessControl struct {
    userRoles []string
    tenantID  string
}

// 字段访问权限映射
var FieldPermissions = map[string]map[string][]string{
    "employee": {
        "salary":           {"HR_MANAGER", "TENANT_ADMIN"},
        "personal_email":   {"HR_MANAGER", "TENANT_ADMIN", "SELF"},
        "phone_number":     {"HR_MANAGER", "HR_SPECIALIST", "LINE_MANAGER", "SELF"},
        "hire_date":        {"HR_MANAGER", "HR_SPECIALIST", "TENANT_ADMIN"},
        "termination_date": {"HR_MANAGER", "TENANT_ADMIN"},
    },
    "user": {
        "password_hash":    {"SUPER_ADMIN"}, // 极度敏感
        "two_factor_secret": {"SUPER_ADMIN"},
        "last_login_at":    {"TENANT_ADMIN", "SELF"},
        "failed_login_attempts": {"TENANT_ADMIN", "SUPER_ADMIN"},
    },
}

// 动态字段过滤
func (f *FieldAccessControl) FilterFields(data map[string]interface{}, resourceType string) map[string]interface{} {
    result := make(map[string]interface{})
    
    for field, value := range data {
        if f.hasFieldAccess(resourceType, field) {
            result[field] = value
        }
    }
    
    return result
}
```

### 🔄 数据同步与一致性保证

#### **CQRS 多租户事件流**
```go
// 多租户事件总线
type MultiTenantEventBus struct {
    kafkaProducer *kafka.Producer
    topics        map[string]string // tenant_id -> topic_name
}

func (bus *MultiTenantEventBus) PublishEvent(tenantID string, event domain.Event) error {
    // 1. 租户事件隔离：每个租户使用独立的 Kafka Topic
    topic := fmt.Sprintf("tenant-%s-events", tenantID)
    
    // 2. 事件序列化（包含租户信息）
    eventData := EventEnvelope{
        TenantID:  tenantID,
        EventID:   uuid.New(),
        EventType: event.GetType(),
        Payload:   event,
        Timestamp: time.Now(),
        Metadata: map[string]string{
            "source":  "cube-castle-backend",
            "version": "v1.0",
        },
    }
    
    // 3. 发布到租户专用主题
    return bus.kafkaProducer.Produce(&kafka.Message{
        TopicPartition: kafka.TopicPartition{
            Topic:     &topic,
            Partition: kafka.PartitionAny,
        },
        Key:   []byte(tenantID),
        Value: jsonEncode(eventData),
        Headers: []kafka.Header{
            {Key: "tenant_id", Value: []byte(tenantID)},
            {Key: "event_type", Value: []byte(event.GetType())},
        },
    }, nil)
}

// Neo4j 多租户同步消费者
func (c *Neo4jConsumer) ProcessEvent(msg *kafka.Message) error {
    // 1. 提取租户信息
    tenantID := string(getHeaderValue(msg.Headers, "tenant_id"))
    
    // 2. 租户数据隔离：使用租户标签
    session := c.driver.NewSession(neo4j.SessionConfig{
        DatabaseName: "neo4j",
        AccessMode:   neo4j.AccessModeWrite,
    })
    defer session.Close()
    
    // 3. 执行租户隔离的 Cypher 查询
    _, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
        query := `
            MERGE (e:Employee {id: $employee_id, tenant_id: $tenant_id})
            SET e.name = $name,
                e.email = $email,
                e.updated_at = datetime()
            RETURN e
        `
        
        return tx.Run(query, map[string]interface{}{
            "employee_id": event.AggregateID,
            "tenant_id":   tenantID,
            "name":        event.Name,
            "email":       event.Email,
        })
    })
    
    return err
}
```

## 📋 完整实施方案总结

### 🚀 **实施路线图**

#### **Phase 1: 核心基础设施 (2-3周)**
1. **数据模型实施**
   - 创建 Tenant、User、Role、UserRole、UserSession 实体
   - 更新现有 Employee 实体，添加 user_id 关联
   - 实施数据库迁移脚本

2. **基础中间件完善**  
   - 完善 TenantContext 中间件实现
   - 实现 JWT 认证中间件
   - 集成 OPA 策略引擎

#### **Phase 2: 认证服务 (2-3周)**
1. **用户认证服务**
   - 实现登录/注册 API
   - 集成双因子认证 (2FA)
   - 实现会话管理

2. **权限管理服务**
   - RBAC 权限框架
   - 字段级权限控制
   - 权限缓存机制

#### **Phase 3: 安全强化 (1-2周)**
1. **数据隔离加固**
   - PostgreSQL RLS 策略部署
   - Neo4j 租户标签隔离
   - Kafka 租户主题隔离

2. **安全监控**
   - 审计日志系统
   - 异常行为检测
   - 安全警报机制

### 🎯 **核心优势**

#### **1. 完全的租户隔离**
- ✅ 数据库级别的 RLS 保证数据安全
- ✅ 应用级别的多层权限控制
- ✅ 事件流的租户隔离

#### **2. 灵活的权限模型**
- ✅ 基于角色的权限控制 (RBAC)
- ✅ 细粒度的字段级权限
- ✅ 动态的权限策略 (OPA)

#### **3. 企业级安全性**
- ✅ JWT + 会话双重认证
- ✅ 双因子认证支持
- ✅ 完整的审计日志

#### **4. 高性能架构**
- ✅ 基于现有 CQRS 架构
- ✅ 缓存友好的权限设计
- ✅ 分布式会话管理

### 🔧 **技术集成点**

#### **与现有架构的兼容性**
1. **无缝集成现有 Ent Schema**
   - User 实体与 Employee 实体通过外键关联
   - 保持现有业务 ID 系统的一致性
   - 向后兼容现有 API

2. **增强现有 CQRS 流程**
   - 事件中自动注入租户信息
   - Neo4j 查询添加租户过滤
   - Kafka 消息添加租户路由

3. **前端集成支持**
   - JWT Token 包含完整权限信息
   - REST API 返回用户权限列表
   - 支持前端路由级权限控制

### 📊 **性能考虑**

#### **缓存策略**
```go
// Redis 缓存层次
type CacheStrategy struct {
    // L1: 用户会话缓存 (TTL: 30分钟)
    UserSessions map[string]*UserSession
    
    // L2: 用户权限缓存 (TTL: 15分钟) 
    UserPermissions map[string][]string
    
    // L3: 租户配置缓存 (TTL: 1小时)
    TenantConfigs map[string]*TenantConfig
}
```

#### **数据库性能优化**
- 所有多租户查询都包含 `tenant_id` 索引
- 用户认证采用复合索引 `(tenant_id, email)`
- 会话查询优化 `(user_id, is_active, expires_at)`

### 🔒 **安全合规**

#### **数据保护措施**
1. **敏感数据加密**
   - 密码使用 bcrypt + salt
   - 双因子密钥使用 AES-256 加密
   - 会话令牌使用 HMAC-SHA256 签名

2. **审计合规**
   - 所有用户操作记录审计日志
   - 权限变更事件追踪
   - 数据访问日志记录

#### **灾难恢复**
- 用户数据多副本备份
- 会话状态 Redis 集群部署  
- 权限配置版本化管理

---

## 📝 **总结**

这套方案充分利用了 cube-castle 现有的 CQRS + Event Sourcing 架构，在保持系统高性能的同时，提供了企业级的多租户用户管理和安全控制能力。建议按照分阶段实施，确保与现有业务的平滑集成。

**文档版本：** v1.0  
**创建日期：** 2025-01-05  
**最后更新：** 2025-01-05  
**作者：** Architecture Agent  
**状态：** Draft - 待评审