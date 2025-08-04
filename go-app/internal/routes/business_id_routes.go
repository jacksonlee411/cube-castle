package routes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"github.com/gaogu/cube-castle/go-app/ent"
	"github.com/gaogu/cube-castle/go-app/internal/common"
	"github.com/gaogu/cube-castle/go-app/internal/corehr"
	"github.com/gaogu/cube-castle/go-app/internal/handler"
	businessIDMiddleware "github.com/gaogu/cube-castle/go-app/internal/middleware"
)

// BusinessIDEnabledRouter 支持业务ID的路由配置
type BusinessIDEnabledRouter struct {
	entClient     *ent.Client
	db            *sql.DB
	businessIDMgr *common.BusinessIDManager
}

// NewBusinessIDEnabledRouter 创建新的路由配置实例
func NewBusinessIDEnabledRouter(entClient *ent.Client, db *sql.DB) *BusinessIDEnabledRouter {
	// 创建业务ID服务和管理器
	businessIDService := common.NewBusinessIDService(db)
	businessIDConfig := common.DefaultBusinessIDManagerConfig()
	businessIDMgr := common.NewBusinessIDManager(businessIDService, businessIDConfig)

	return &BusinessIDEnabledRouter{
		entClient:     entClient,
		db:            db,
		businessIDMgr: businessIDMgr,
	}
}

// SetupRoutes 设置所有路由
func (router *BusinessIDEnabledRouter) SetupRoutes() chi.Router {
	r := chi.NewRouter()

	// 基础中间件
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// CORS中间件 (如果需要)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
			
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	})

	// API版本路由
	r.Route("/api/v1", func(r chi.Router) {
		// 健康检查
		r.Get("/health", router.healthCheck)
		
		// CoreHR 模块路由
		r.Route("/corehr", func(r chi.Router) {
			router.setupCoreHRRoutes(r)
		})
		
		// 业务ID管理路由
		router.setupBusinessIDRoutes(r)
	})

	return r
}

// setupCoreHRRoutes 设置CoreHR模块路由
func (router *BusinessIDEnabledRouter) setupCoreHRRoutes(r chi.Router) {
	// 创建服务实例
	employeeService := corehr.NewEmployeeService(router.entClient, router.businessIDMgr)
	
	// 创建处理器
	employeeHandler := handler.NewEmployeeHandlerBusinessID(employeeService)
	
	// 员工管理路由
	r.Route("/employees", func(r chi.Router) {
		// 列表和创建 (不需要ID验证)
		r.Get("/", employeeHandler.ListEmployees)
		r.Post("/", employeeHandler.CreateEmployee)
		
		// 统计信息
		r.Get("/statistics", employeeHandler.GetEmployeeStatistics)
		
		// 业务ID验证测试端点
		r.Get("/validate-business-id", employeeHandler.ValidateBusinessID)
		
		// 单个员工操作 (需要ID验证)
		r.Route("/{employee_id}", func(r chi.Router) {
			// 应用业务ID验证中间件
			r.Use(businessIDMiddleware.BusinessIDValidator(common.EntityTypeEmployee))
			
			r.Get("/", employeeHandler.GetEmployee)
			r.Put("/", employeeHandler.UpdateEmployee)
			r.Delete("/", employeeHandler.DeleteEmployee)
		})
	})

	// 组织管理路由 (待实现)
	r.Route("/organizations", func(r chi.Router) {
		r.Get("/", router.placeholderHandler("GET /organizations"))
		r.Post("/", router.placeholderHandler("POST /organizations"))
		
		r.Route("/{organization_id}", func(r chi.Router) {
			// 应用业务ID验证中间件
			r.Use(businessIDMiddleware.BusinessIDValidator(common.EntityTypeOrganization))
			
			r.Get("/", router.placeholderHandler("GET /organizations/{id}"))
			r.Put("/", router.placeholderHandler("PUT /organizations/{id}"))
			r.Delete("/", router.placeholderHandler("DELETE /organizations/{id}"))
		})
		
		// 组织树结构
		r.Get("/tree", router.placeholderHandler("GET /organizations/tree"))
	})

	// 职位管理路由 (待实现)
	r.Route("/positions", func(r chi.Router) {
		r.Get("/", router.placeholderHandler("GET /positions"))
		r.Post("/", router.placeholderHandler("POST /positions"))
		
		r.Route("/{position_id}", func(r chi.Router) {
			// 应用业务ID验证中间件
			r.Use(businessIDMiddleware.BusinessIDValidator(common.EntityTypePosition))
			
			r.Get("/", router.placeholderHandler("GET /positions/{id}"))
			r.Put("/", router.placeholderHandler("PUT /positions/{id}"))
			r.Delete("/", router.placeholderHandler("DELETE /positions/{id}"))
		})
	})
}

// setupBusinessIDRoutes 设置业务ID管理路由
func (router *BusinessIDEnabledRouter) setupBusinessIDRoutes(r chi.Router) {
	r.Route("/business-ids", func(r chi.Router) {
		// 生成业务ID
		r.Post("/generate", router.generateBusinessID)
		
		// 查询业务ID信息
		r.Get("/lookup", router.lookupBusinessID)
		
		// 验证业务ID格式
		r.Get("/validate", router.validateBusinessID)
		
		// 获取ID范围信息
		r.Get("/ranges", router.getBusinessIDRanges)
	})
}

// healthCheck 健康检查端点
func (router *BusinessIDEnabledRouter) healthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "1.0.0",
		"features": map[string]bool{
			"business_id_support": true,
			"uuid_compatibility":  true,
			"validation_enabled":  true,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// generateBusinessID 生成业务ID
func (router *BusinessIDEnabledRouter) generateBusinessID(w http.ResponseWriter, r *http.Request) {
	var req struct {
		EntityType string `json:"entity_type"`
		Count      int    `json:"count,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		router.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.Count == 0 {
		req.Count = 1
	}

	if req.Count > 10 {
		router.writeError(w, http.StatusBadRequest, "INVALID_COUNT", "Cannot generate more than 10 IDs at once")
		return
	}

	entityType := common.EntityType(req.EntityType)
	if entityType != common.EntityTypeEmployee && 
	   entityType != common.EntityTypeOrganization && 
	   entityType != common.EntityTypePosition {
		router.writeError(w, http.StatusBadRequest, "INVALID_ENTITY_TYPE", "Invalid entity type")
		return
	}

	generatedIDs := make([]string, req.Count)
	for i := 0; i < req.Count; i++ {
		businessID, err := router.businessIDMgr.GenerateUniqueBusinessID(r.Context(), entityType)
		if err != nil {
			router.writeError(w, http.StatusInternalServerError, "GENERATION_FAILED", "Failed to generate business ID")
			return
		}
		generatedIDs[i] = businessID
	}

	response := map[string]interface{}{
		"generated_ids": generatedIDs,
		"entity_type":   req.EntityType,
		"count":         req.Count,
		"range":         common.GetBusinessIDRange(entityType),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// lookupBusinessID 查询业务ID信息
func (router *BusinessIDEnabledRouter) lookupBusinessID(w http.ResponseWriter, r *http.Request) {
	businessID := r.URL.Query().Get("business_id")
	entityTypeStr := r.URL.Query().Get("entity_type")

	if businessID == "" || entityTypeStr == "" {
		router.writeError(w, http.StatusBadRequest, "MISSING_PARAMETERS", "business_id and entity_type are required")
		return
	}

	entityType := common.EntityType(entityTypeStr)
	businessIDService := common.NewBusinessIDService(router.db)
	
	result, err := businessIDService.LookupByBusinessID(r.Context(), entityType, businessID)
	if err != nil {
		router.writeError(w, http.StatusBadRequest, "LOOKUP_FAILED", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// validateBusinessID 验证业务ID格式
func (router *BusinessIDEnabledRouter) validateBusinessID(w http.ResponseWriter, r *http.Request) {
	businessID := r.URL.Query().Get("business_id")
	entityTypeStr := r.URL.Query().Get("entity_type")

	if businessID == "" || entityTypeStr == "" {
		router.writeError(w, http.StatusBadRequest, "MISSING_PARAMETERS", "business_id and entity_type are required")
		return
	}

	entityType := common.EntityType(entityTypeStr)
	err := common.ValidateBusinessID(entityType, businessID)

	result := map[string]interface{}{
		"business_id":  businessID,
		"entity_type":  entityTypeStr,
		"valid":        err == nil,
	}

	if err != nil {
		result["error"] = err.Error()
		result["range"] = common.GetBusinessIDRange(entityType)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// getBusinessIDRanges 获取所有实体类型的业务ID范围信息
func (router *BusinessIDEnabledRouter) getBusinessIDRanges(w http.ResponseWriter, r *http.Request) {
	ranges := map[string]interface{}{
		"employee": map[string]interface{}{
			"range":       "1-99999999",
			"description": "Employee business ID range",
			"pattern":     "^[1-9][0-9]{0,7}$",
			"example":     "1",
		},
		"organization": map[string]interface{}{
			"range":       "100000-999999",
			"description": "Organization business ID range",
			"pattern":     "^[1-9][0-9]{5}$",
			"example":     "100000",
		},
		"position": map[string]interface{}{
			"range":       "1000000-9999999",
			"description": "Position business ID range",
			"pattern":     "^[1-9][0-9]{6}$",
			"example":     "1000000",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ranges)
}

// placeholderHandler 占位符处理器 (用于未实现的端点)
func (router *BusinessIDEnabledRouter) placeholderHandler(endpoint string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"message":  "Endpoint not yet implemented",
			"endpoint": endpoint,
			"status":   "coming_soon",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		json.NewEncoder(w).Encode(response)
	}
}

// writeError 写入错误响应
func (router *BusinessIDEnabledRouter) writeError(w http.ResponseWriter, statusCode int, errorCode, message string) {
	errorResp := businessIDMiddleware.ErrorResponse{
		Error:     errorCode,
		Message:   message,
		Timestamp: time.Now(),
		RequestID: fmt.Sprintf("req_%s", uuid.New().String()[:8]),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResp)
}