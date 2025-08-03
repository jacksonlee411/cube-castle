package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/internal/cqrs/handlers"
)

// SetupOrganizationAdapterRoutes 设置组织管理适配路由
// 这些路由将前端的CoreHR API调用适配到CQRS端点
func SetupOrganizationAdapterRoutes(
	r chi.Router, 
	cmdHandler *handlers.CommandHandler, 
	queryHandler *handlers.QueryHandler,
) {
	// CoreHR API适配器 - 适配前端API调用到CQRS架构
	r.Route("/api/v1/corehr", func(r chi.Router) {
		// 组织管理适配器
		r.Route("/organizations", func(r chi.Router) {
			// GET /api/v1/corehr/organizations -> CQRS查询端点
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				// 适配查询参数
				adaptedRequest := r.Clone(r.Context())
				
				// 转发到CQRS查询端点
				queryHandler.ListOrganizations(w, adaptedRequest)
			})
			
			// GET /api/v1/corehr/organizations/{id} -> CQRS查询端点
			r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
				queryHandler.GetOrganization(w, r)
			})
			
			// POST /api/v1/corehr/organizations -> CQRS命令端点
			r.Post("/", func(w http.ResponseWriter, r *http.Request) {
				// 适配请求体格式
				var frontendRequest struct {
					UnitType     string                 `json:"unit_type"`
					Name         string                 `json:"name"`
					Description  *string                `json:"description"`
					ParentUnitID *string                `json:"parent_unit_id"`
					Status       string                 `json:"status"`
					Profile      map[string]interface{} `json:"profile"`
				}
				
				if err := json.NewDecoder(r.Body).Decode(&frontendRequest); err != nil {
					http.Error(w, "Invalid request body", http.StatusBadRequest)
					return
				}
				
				// 转换为CQRS命令格式
				cqrsRequest := map[string]interface{}{
					"unit_type":   frontendRequest.UnitType,
					"name":        frontendRequest.Name,
					"description": frontendRequest.Description,
					"status":      frontendRequest.Status,
					"profile":     frontendRequest.Profile,
				}
				
				if frontendRequest.ParentUnitID != nil && *frontendRequest.ParentUnitID != "" {
					if parentID, err := uuid.Parse(*frontendRequest.ParentUnitID); err == nil {
						cqrsRequest["parent_unit_id"] = parentID
					}
				}
				
				// 重新编码请求体
				reqBody, _ := json.Marshal(cqrsRequest)
				adaptedRequest := r.Clone(r.Context())
				adaptedRequest.Body = http.NoBody
				adaptedRequest.ContentLength = int64(len(reqBody))
				
				// 转发到CQRS命令端点
				cmdHandler.CreateOrganization(w, adaptedRequest)
			})
			
			// PUT /api/v1/corehr/organizations/{id} -> CQRS命令端点
			r.Put("/{id}", func(w http.ResponseWriter, r *http.Request) {
				cmdHandler.UpdateOrganization(w, r)
			})
			
			// DELETE /api/v1/corehr/organizations/{id} -> CQRS命令端点
			r.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
				cmdHandler.DeleteOrganization(w, r)
			})
			
			// GET /api/v1/corehr/organizations/stats -> 组织统计
			r.Get("/stats", func(w http.ResponseWriter, r *http.Request) {
				queryHandler.GetOrganizationStats(w, r)
			})
		})
	})
	
	// CQRS直接端点 (用于高级客户端)
	r.Route("/api/v1", func(r chi.Router) {
		SetupCQRSRoutes(r, cmdHandler, queryHandler)
	})
}

// 辅助函数：转换前端查询参数到CQRS格式
func convertFrontendParams(r *http.Request) map[string]string {
	params := make(map[string]string)
	
	// 基础分页参数
	if page := r.URL.Query().Get("page"); page != "" {
		params["page"] = page
	}
	if pageSize := r.URL.Query().Get("pageSize"); pageSize != "" {
		params["page_size"] = pageSize
	}
	
	// 过滤参数
	if search := r.URL.Query().Get("search"); search != "" {
		params["search"] = search
	}
	if unitType := r.URL.Query().Get("unit_type"); unitType != "" {
		params["unit_type"] = unitType
	}
	if status := r.URL.Query().Get("status"); status != "" {
		params["status"] = status
	}
	if parentUnitID := r.URL.Query().Get("parent_unit_id"); parentUnitID != "" {
		params["parent_unit_id"] = parentUnitID
	}
	
	return params
}

// 辅助函数：转换CQRS响应到前端格式
func convertToFrontendResponse(cqrsResponse map[string]interface{}) map[string]interface{} {
	// 如果需要格式转换，在这里进行
	// 目前保持相同格式
	return cqrsResponse
}