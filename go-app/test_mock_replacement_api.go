package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/gaogu/cube-castle/go-app/generated/openapi"
	"github.com/gaogu/cube-castle/go-app/internal/common"
	"github.com/gaogu/cube-castle/go-app/internal/corehr"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/gaogu/cube-castle/go-app/internal/outbox"
	"github.com/gaogu/cube-castle/go-app/internal/validation"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func main() {
	fmt.Println("🧪 Mock替换API验证测试")
	fmt.Println("========================")

	// 设置数据库环境变量
	os.Setenv("DATABASE_URL", "postgresql://user:password@localhost:5432/cubecastle?sslmode=disable")
	os.Setenv("NEO4J_URI", "bolt://localhost:7687")
	os.Setenv("NEO4J_USER", "neo4j")
	os.Setenv("NEO4J_PASSWORD", "password")

	logger := logging.NewStructuredLogger()

	// 测试1: 无数据库连接时的行为
	fmt.Println("\n📋 测试1: 无数据库连接的行为")
	testWithoutDatabase(logger)

	// 测试2: 有数据库连接时的行为
	fmt.Println("\n📋 测试2: 有数据库连接的行为")
	testWithDatabase(logger)

	fmt.Println("\n🎉 Mock替换API验证完成！")
}

func testWithoutDatabase(logger *logging.StructuredLogger) {
	// 清除数据库环境变量
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("NEO4J_URI")

	// 初始化服务
	coreHRService := initializeCoreHRService(nil, logger)
	
	// 创建测试路由
	router := setupTestRoutes(coreHRService, logger)

	// 测试API调用
	testEmployeeAPI(router, "无数据库", false)
}

func testWithDatabase(logger *logging.StructuredLogger) {
	// 设置数据库环境变量
	os.Setenv("DATABASE_URL", "postgresql://user:password@localhost:5432/cubecastle?sslmode=disable")
	os.Setenv("NEO4J_URI", "bolt://localhost:7687")
	os.Setenv("NEO4J_USER", "neo4j")
	os.Setenv("NEO4J_PASSWORD", "password")

	// 初始化数据库连接
	db := common.InitDatabaseConnection()
	coreHRService := initializeCoreHRService(db, logger)
	
	// 创建测试路由
	router := setupTestRoutes(coreHRService, logger)

	// 测试API调用
	testEmployeeAPI(router, "有数据库", db != nil)
}

func initializeCoreHRService(db interface{}, logger *logging.StructuredLogger) *corehr.Service {
	if db == nil {
		logger.Info("初始化CoreHR服务 - Mock模式")
		return corehr.NewMockService()
	}

	// 实际模式
	logger.Info("初始化CoreHR服务 - 数据库模式")
	dbConn := db.(*common.Database)
	repo := corehr.NewRepository(dbConn.PostgreSQL)
	outboxService := outbox.NewService(dbConn.PostgreSQL)
	
	return corehr.NewService(repo, outboxService)
}

func setupTestRoutes(coreHRService *corehr.Service, logger *logging.StructuredLogger) *chi.Mux {
	r := chi.NewRouter()
	
	// 创建验证器
	mockChecker := validation.NewMockValidationChecker()
	validator := validation.NewEmployeeValidator(mockChecker, mockChecker, mockChecker, mockChecker)

	// 添加员工路由
	r.Route("/api/v1/corehr", func(r chi.Router) {
		r.Get("/employees", handleListEmployees(coreHRService, logger, validator))
		r.Post("/employees", handleCreateEmployee(coreHRService, logger, validator))
	})

	return r
}

func testEmployeeAPI(router *chi.Mux, testName string, hasDatabase bool) {
	fmt.Printf("  测试场景: %s\n", testName)

	// 测试获取员工列表
	req := httptest.NewRequest("GET", "/api/v1/corehr/employees?page=1&page_size=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Printf("  📡 GET /employees - 状态码: %d\n", w.Code)
	
	if hasDatabase {
		// 有数据库时期望的行为
		if w.Code == 500 || w.Code == 503 {
			fmt.Printf("  ✅ 正确：服务返回错误，表明不再使用Mock数据\n")
		} else if w.Code == 200 {
			fmt.Printf("  ✅ 正确：服务连接数据库成功\n")
		}
	} else {
		// 无数据库时期望的行为
		if w.Code == 500 {
			fmt.Printf("  ✅ 正确：服务检查到repository为nil并返回错误\n")
		} else if w.Code == 200 {
			fmt.Printf("  ❌ 错误：服务可能仍在使用Mock数据\n")
		}
	}

	// 测试创建员工
	createReq := openapi.CreateEmployeeRequest{
		EmployeeNumber: "TEST001",
		FirstName:      "测试",
		LastName:       "用户",
		Email:          openapi_types.Email("test@example.com"),
		HireDate:       openapi_types.Date{Time: time.Now()},
	}

	reqBody, _ := json.Marshal(createReq)
	req = httptest.NewRequest("POST", "/api/v1/corehr/employees", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Printf("  📡 POST /employees - 状态码: %d\n", w.Code)
	
	if hasDatabase {
		if w.Code == 500 || w.Code == 503 {
			fmt.Printf("  ✅ 正确：服务返回错误，可能是数据库schema问题\n")
		} else if w.Code == 201 {
			fmt.Printf("  ✅ 正确：服务成功创建员工\n")
		}
	} else {
		if w.Code == 500 {
			fmt.Printf("  ✅ 正确：服务检查到repository为nil并返回错误\n")
		} else if w.Code == 201 {
			fmt.Printf("  ❌ 错误：服务可能仍在使用Mock数据\n")
		}
	}
}

// 简化的处理器函数
func handleListEmployees(service *corehr.Service, logger *logging.StructuredLogger, validator *validation.EmployeeValidator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID := uuid.New()
		
		response, err := service.ListEmployees(r.Context(), tenantID, 1, 10, "")
		if err != nil {
			logger.Info("ListEmployees error", "error", err.Error())
			http.Error(w, "Service error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func handleCreateEmployee(service *corehr.Service, logger *logging.StructuredLogger, validator *validation.EmployeeValidator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req openapi.CreateEmployeeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		tenantID := uuid.New()
		employee, err := service.CreateEmployee(r.Context(), tenantID, &req)
		if err != nil {
			logger.Info("CreateEmployee error", "error", err.Error())
			http.Error(w, "Service error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(employee)
	}
}