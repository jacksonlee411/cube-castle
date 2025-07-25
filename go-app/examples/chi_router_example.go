package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
}

// APIResponse 通用API响应
type APIResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func main() {
	// 创建Chi路由器
	r := chi.NewRouter()

	// 添加中间件
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))

	// 健康检查端点
	r.Get("/health", healthHandler)

	// API v1 路由组
	r.Route("/api/v1", func(r chi.Router) {
		// CoreHR 模块路由
		r.Route("/corehr", func(r chi.Router) {
			r.Get("/employees", listEmployeesHandler)
			r.Post("/employees", createEmployeeHandler)
			r.Route("/employees/{employeeID}", func(r chi.Router) {
				r.Get("/", getEmployeeHandler)
				r.Put("/", updateEmployeeHandler)
				r.Delete("/", deleteEmployeeHandler)
			})
		})

		// Intelligence Gateway 路由
		r.Route("/intelligence", func(r chi.Router) {
			r.Post("/query", interpretQueryHandler)
			r.Post("/batch", batchQueryHandler)
		})

		// 监控模块路由
		r.Route("/monitoring", func(r chi.Router) {
			r.Get("/metrics", metricsHandler)
			r.Get("/health/detailed", detailedHealthHandler)
		})

		// 工作流引擎路由
		r.Route("/workflow", func(r chi.Router) {
			r.Post("/start", startWorkflowHandler)
			r.Get("/status/{workflowID}", getWorkflowStatusHandler)
		})
	})

	// 启动服务器
	fmt.Println("🚀 Cube Castle API Server starting on :8080")
	fmt.Println("📊 Using Chi v5.2.2 Router")
	fmt.Println("🔗 Health check: http://localhost:8080/health")
	fmt.Println("📖 API docs: http://localhost:8080/api/v1")

	log.Fatal(http.ListenAndServe(":8080", r))
}

// healthHandler 健康检查处理器
func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Service:   "cube-castle-api",
		Version:   "v1.2.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CoreHR 处理器示例
func listEmployeesHandler(w http.ResponseWriter, r *http.Request) {
	response := APIResponse{
		Message: "List employees endpoint",
		Data: map[string]interface{}{
			"framework": "Chi v5.2.2",
			"module":    "CoreHR",
			"action":    "list_employees",
		},
	}
	respondJSON(w, http.StatusOK, response)
}

func createEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	response := APIResponse{
		Message: "Create employee endpoint",
		Data: map[string]interface{}{
			"framework": "Chi v5.2.2",
			"module":    "CoreHR",
			"action":    "create_employee",
		},
	}
	respondJSON(w, http.StatusCreated, response)
}

func getEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	employeeID := chi.URLParam(r, "employeeID")
	response := APIResponse{
		Message: "Get employee endpoint",
		Data: map[string]interface{}{
			"framework":   "Chi v5.2.2",
			"module":      "CoreHR",
			"action":      "get_employee",
			"employee_id": employeeID,
		},
	}
	respondJSON(w, http.StatusOK, response)
}

func updateEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	employeeID := chi.URLParam(r, "employeeID")
	response := APIResponse{
		Message: "Update employee endpoint",
		Data: map[string]interface{}{
			"framework":   "Chi v5.2.2",
			"module":      "CoreHR",
			"action":      "update_employee",
			"employee_id": employeeID,
		},
	}
	respondJSON(w, http.StatusOK, response)
}

func deleteEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	employeeID := chi.URLParam(r, "employeeID")
	response := APIResponse{
		Message: "Delete employee endpoint",
		Data: map[string]interface{}{
			"framework":   "Chi v5.2.2",
			"module":      "CoreHR",
			"action":      "delete_employee",
			"employee_id": employeeID,
		},
	}
	respondJSON(w, http.StatusOK, response)
}

// Intelligence Gateway 处理器示例
func interpretQueryHandler(w http.ResponseWriter, r *http.Request) {
	response := APIResponse{
		Message: "Interpret query endpoint",
		Data: map[string]interface{}{
			"framework": "Chi v5.2.2",
			"module":    "Intelligence Gateway",
			"action":    "interpret_query",
		},
	}
	respondJSON(w, http.StatusOK, response)
}

func batchQueryHandler(w http.ResponseWriter, r *http.Request) {
	response := APIResponse{
		Message: "Batch query endpoint",
		Data: map[string]interface{}{
			"framework": "Chi v5.2.2",
			"module":    "Intelligence Gateway",
			"action":    "batch_query",
		},
	}
	respondJSON(w, http.StatusOK, response)
}

// 监控模块处理器示例
func metricsHandler(w http.ResponseWriter, r *http.Request) {
	response := APIResponse{
		Message: "Metrics endpoint",
		Data: map[string]interface{}{
			"framework": "Chi v5.2.2",
			"module":    "Monitoring",
			"action":    "get_metrics",
		},
	}
	respondJSON(w, http.StatusOK, response)
}

func detailedHealthHandler(w http.ResponseWriter, r *http.Request) {
	response := APIResponse{
		Message: "Detailed health check endpoint",
		Data: map[string]interface{}{
			"framework": "Chi v5.2.2",
			"module":    "Monitoring",
			"action":    "detailed_health",
		},
	}
	respondJSON(w, http.StatusOK, response)
}

// 工作流引擎处理器示例
func startWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	response := APIResponse{
		Message: "Start workflow endpoint",
		Data: map[string]interface{}{
			"framework": "Chi v5.2.2",
			"module":    "Workflow Engine",
			"action":    "start_workflow",
		},
	}
	respondJSON(w, http.StatusOK, response)
}

func getWorkflowStatusHandler(w http.ResponseWriter, r *http.Request) {
	workflowID := chi.URLParam(r, "workflowID")
	response := APIResponse{
		Message: "Get workflow status endpoint",
		Data: map[string]interface{}{
			"framework":   "Chi v5.2.2",
			"module":      "Workflow Engine",
			"action":      "get_workflow_status",
			"workflow_id": workflowID,
		},
	}
	respondJSON(w, http.StatusOK, response)
}

// respondJSON 统一JSON响应函数
func respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}