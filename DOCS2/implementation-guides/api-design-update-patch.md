# API接口设计更新补丁

**日期**: 2025-08-04  
**类型**: 路由标准化更新  
**影响**: 职位管理API相关设计

## 🔄 更新内容

### 已废弃的设计概念
以下API设计已确认不会实现，从标准规范中移除：

```go
// ❌ 不再实现 - CoreHR模块职位路由
r.Route("/api/v1/corehr", func(r chi.Router) {
    r.Route("/positions", func(r chi.Router) {
        r.Get("/", h.GetCoreHRPositions)      // 不实现
        r.Post("/", h.CreateCoreHRPosition)   // 不实现
    })
})

// ❌ 不再实现 - 员工职位历史子资源
r.Route("/employees/{employeeID}", func(r chi.Router) {
    r.Route("/positions", func(r chi.Router) {
        r.Get("/", h.GetPositionHistory)         // 不实现
        r.Post("/", h.CreatePositionChange)      // 不实现
        r.Get("/current", h.GetCurrentPosition)  // 不实现
    })
})

// ❌ 不再实现 - 组织级职位管理
r.Route("/organization", func(r chi.Router) {
    r.Get("/positions", h.GetPositions)  // 不实现
})
```

### 标准化的实际实现
确认以下为正式的职位管理API规范：

```go
// ✅ 标准实现 - 独立职位资源API
r.Route("/api/v1/positions", func(r chi.Router) {
    r.Get("/", h.ListPositions)        // 获取职位列表
    r.Post("/", h.CreatePosition)      // 创建新职位
    r.Route("/{id}", func(r chi.Router) {
        r.Get("/", h.GetPosition)      // 获取特定职位
        r.Put("/", h.UpdatePosition)   // 更新职位信息
        r.Delete("/", h.DeletePosition) // 删除职位
    })
})
```

## 📋 标准路由规范

### 核心业务模块路由
```yaml
职位管理 (独立资源):
  - GET    /api/v1/positions
  - POST   /api/v1/positions  
  - GET    /api/v1/positions/{id}
  - PUT    /api/v1/positions/{id}
  - DELETE /api/v1/positions/{id}

员工管理 (CoreHR模块):
  - GET    /api/v1/corehr/employees
  - POST   /api/v1/corehr/employees
  - GET    /api/v1/corehr/employees/{id}
  - PUT    /api/v1/corehr/employees/{id}
  - DELETE /api/v1/corehr/employees/{id}

组织管理 (CoreHR模块):
  - GET    /api/v1/corehr/organizations
  - POST   /api/v1/corehr/organizations
  - GET    /api/v1/corehr/organizations/{id}
  - PUT    /api/v1/corehr/organizations/{id}
  - DELETE /api/v1/corehr/organizations/{id}
```

### 设计决策说明
1. **职位独立化**: 职位管理采用独立资源模式，便于多模块共享使用
2. **CoreHR集成**: 员工和组织保持在CoreHR模块内，体现业务关联性
3. **路由一致性**: 所有CRUD操作遵循统一的RESTful设计模式

## 🔗 相关文档链接
- [完整API规范](../DOCS2/api-specifications/positions-api-specification.md)
- [架构决策记录](../DOCS2/architecture-decisions/ADR-001-positions-api-architecture.md)
- [前端集成指南](../DOCS2/implementation-guides/frontend-api-integration.md)

## ⚠️ 迁移指导
对于引用了废弃设计的代码或文档：
1. 将所有职位相关API调用更新为 `/api/v1/positions`
2. 移除对 `/api/v1/corehr/positions` 的引用
3. 更新相关测试用例和文档

---
*此补丁是职位API路由标准化工作的一部分*