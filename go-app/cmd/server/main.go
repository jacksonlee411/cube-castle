package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/joho/godotenv"

	pb "github.com/gaogu/cube-castle/go-app/generated/grpc/intelligence"
	"github.com/gaogu/cube-castle/go-app/generated/openapi"
	"github.com/gaogu/cube-castle/go-app/internal/common"
	"github.com/gaogu/cube-castle/go-app/internal/corehr"
	"github.com/gaogu/cube-castle/go-app/internal/outbox"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Server struct {
	db             *common.Database
	intelSvcClient pb.IntelligenceServiceClient
	corehrService  *corehr.Service
	outboxService  *outbox.Service
}



// (POST /api/v1/interpret)
func (s *Server) InterpretQuery(w http.ResponseWriter, r *http.Request) {
	var req openapi.InterpretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("JSON decode error: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 添加调试信息 - 解引用指针后打印
	var queryStr, userIdStr string = "nil", "nil"
	if req.Query != nil {
		queryStr = *req.Query
	}
	if req.UserId != nil {
		userIdStr = req.UserId.String()
	}
	log.Printf("Received request: Query=%s, UserId=%s", queryStr, userIdStr)

	// --- 这里是修改点 ---
	// 对 UserId 和 Query 两个指针类型进行nil检查
	if req.UserId == nil || req.Query == nil {
		log.Printf("Missing required fields: Query=%v, UserId=%v", req.Query, req.UserId)
		http.Error(w, "user_id and query are required", http.StatusBadRequest)
		return
	}
	// --------------------

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*15)
	defer cancel()

	// --- 这里是修改点 ---
	// 使用 * 解引用操作符，获取指针指向的真实string值
	grpcRes, err := s.intelSvcClient.InterpretText(ctx, &pb.InterpretRequest{
		UserText:  *req.Query,
		SessionId: req.UserId.String(),
	})
	// --------------------

	if err != nil {
		http.Error(w, fmt.Sprintf("gRPC call failed: %v", err), http.StatusInternalServerError)
		return
	}

	switch grpcRes.Intent {
	case "update_phone_number":
		var params struct {
			EmployeeID     uuid.UUID `json:"employee_id"`
			NewPhoneNumber string    `json:"new_phone_number"`
		}
		if err := json.Unmarshal([]byte(grpcRes.StructuredDataJson), &params); err != nil {
			http.Error(w, "Failed to parse AI parameters", http.StatusInternalServerError)
			return
		}
		err := s.processPhoneNumberUpdate(r.Context(), params.EmployeeID, params.NewPhoneNumber)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		msg := "Phone number update event accepted."
		json.NewEncoder(w).Encode(openapi.GeneralResponse{Message: &msg})

	case "no_intent_detected":
		// 对于未识别的意图，返回通用响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		msg := fmt.Sprintf("I understand you said: '%s'. This is a general response as the specific intent was not detected.", *req.Query)
		json.NewEncoder(w).Encode(openapi.GeneralResponse{Message: &msg})
	
	default:
		// 对于其他未处理的意图，返回通用响应而不是错误
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		msg := fmt.Sprintf("Intent '%s' received but not yet implemented. Your query was: '%s'", grpcRes.Intent, *req.Query)
		json.NewEncoder(w).Encode(openapi.GeneralResponse{Message: &msg})
	}
}

// (POST /api/v1/internal/corehr/employee-events/phone-number-update)
// 这是符合 "want" 要求的正确函数签名
func (s *Server) PostPhoneNumberUpdateEvent(w http.ResponseWriter, r *http.Request) {
	var req openapi.PhoneNumberUpdateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 直接使用值，不再需要nil检查和解引用
	err := s.processPhoneNumberUpdate(r.Context(), req.EmployeeId, req.NewPhoneNumber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	msg := "Phone number update event accepted."
	json.NewEncoder(w).Encode(openapi.GeneralResponse{Message: &msg})
}

// 内部核心业务逻辑
func (s *Server) processPhoneNumberUpdate(ctx context.Context, employeeId uuid.UUID, newPhoneNumber string) error {
	tx, err := s.db.PostgreSQL.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	updateCmd, err := tx.Exec(ctx,
		"UPDATE corehr.employees SET phone_number = $1 WHERE id = $2",
		newPhoneNumber, employeeId,
	)
	if err != nil {
		return fmt.Errorf("failed to update employee: %w", err)
	}
	if updateCmd.RowsAffected() == 0 {
		return fmt.Errorf("employee with id %s not found", employeeId)
	}

	eventPayload, _ := json.Marshal(map[string]any{
		"employee_id":      employeeId,
		"new_phone_number": newPhoneNumber,
	})
	_, err = tx.Exec(ctx,
		`INSERT INTO outbox.events (aggregate_id, aggregate_type, event_type, payload)
		 VALUES ($1, 'employee', 'phone_number_updated', $2)`,
		employeeId, eventPayload,
	)
	if err != nil {
		return fmt.Errorf("failed to write to outbox: %w", err)
	}

	return tx.Commit(ctx)
}

func main() {
	// 加载环境变量 - 首先尝试当前目录，然后尝试上级目录
	var err error
	err = godotenv.Load(".env")
	if err != nil {
		// 如果当前目录没有.env文件，尝试上级目录
		err = godotenv.Load("../.env")
		if err != nil {
			log.Printf("Warning: Error loading .env file: %v", err)
		}
	}

	// 检查命令行参数
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init-db":
			initDatabase()
			return
		case "seed-data":
			seedDatabase()
			return
		}
	}

	// 初始化数据库连接
	dbConfig := common.NewDatabaseConfig()
	var db *common.Database
	
	// 尝试连接数据库，如果失败则使用 mock 模式
	db, err = common.Connect(dbConfig)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to connect to databases: %v", err)
		log.Printf("📝  Running in mock mode - using in-memory data")
		db = nil
	} else {
		defer db.Close()
		log.Printf("✅ Connected to databases successfully")
	}

	// 连接 gRPC 服务
	grpcTarget := os.Getenv("INTELLIGENCE_SERVICE_GRPC_TARGET")
	if grpcTarget == "" {
		grpcTarget = "localhost:50051"
	}
	
	var conn *grpc.ClientConn
	conn, err = grpc.Dial(grpcTarget, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Did not connect to gRPC server: %v", err)
	}
	defer conn.Close()
	intelClient := pb.NewIntelligenceServiceClient(conn)
	log.Printf("✅ Connected to gRPC server at %s.", grpcTarget)

	// 创建服务器实例
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	
	router := chi.NewRouter()
	
	// 初始化发件箱服务
	var outboxService *outbox.Service
	if db != nil && db.PostgreSQL != nil {
		outboxService = outbox.NewService(db.PostgreSQL)
		log.Printf("✅ Outbox service initialized")
	} else {
		log.Printf("📝 Outbox service not initialized (mock mode)")
	}

	// 初始化 CoreHR 服务
	var corehrService *corehr.Service
	if db != nil && db.PostgreSQL != nil {
		corehrRepo := corehr.NewRepository(db.PostgreSQL)
		corehrService = corehr.NewService(corehrRepo, outboxService)
		log.Printf("✅ CoreHR service initialized with database and outbox")
	} else {
		// 使用 mock 服务
		corehrService = corehr.NewMockService()
		log.Printf("📝 CoreHR service initialized in mock mode")
	}
	
	server := &Server{
		db:             db,
		intelSvcClient: intelClient,
		corehrService:  corehrService,
		outboxService:  outboxService,
	}

	// 注册事件处理器（如果可用）
	if outboxService != nil && db != nil && db.PostgreSQL != nil {
		// 创建适配器
		corehrRepo := corehr.NewRepository(db.PostgreSQL)
		outboxAdapter := corehr.NewOutboxAdapter(corehrRepo)
		
		// 注册事件处理器
		outboxService.RegisterHandler(outbox.NewEmployeeEventHandler(outboxAdapter))
		outboxService.RegisterHandler(outbox.NewEmployeeUpdatedEventHandler(outboxAdapter))
		outboxService.RegisterHandler(outbox.NewEmployeePhoneUpdatedEventHandler(outboxAdapter))
		outboxService.RegisterHandler(outbox.NewOrganizationEventHandler(outboxAdapter))
		outboxService.RegisterHandler(outbox.NewLeaveRequestEventHandler(outboxAdapter))
		outboxService.RegisterHandler(outbox.NewLeaveRequestApprovedEventHandler(outboxAdapter))
		outboxService.RegisterHandler(outbox.NewLeaveRequestRejectedEventHandler(outboxAdapter))
		outboxService.RegisterHandler(outbox.NewNotificationEventHandler())
		
		log.Printf("✅ Event handlers registered")
	}

	// 启动发件箱服务（如果可用）
	if outboxService != nil {
		go func() {
			ctx := context.Background()
			if err := outboxService.Start(ctx); err != nil {
				log.Printf("❌ Failed to start outbox service: %v", err)
			}
		}()
		log.Printf("✅ Outbox service started in background")
	}

	// 添加中间件
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.AllowContentType("application/json"))
	
	// CORS 中间件
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-Requested-With")
			
			// 处理预检请求
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	})

	// 健康检查端点
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	router.Get("/health/db", func(w http.ResponseWriter, r *http.Request) {
		if err := db.HealthCheck(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	// 注册 CoreHR API 路由
	router.Route("/api/v1/corehr", func(r chi.Router) {
		r.Get("/employees", func(w http.ResponseWriter, r *http.Request) {
			// 解析查询参数
			page := 1
			pageSize := 20
			search := ""
			
			if pageStr := r.URL.Query().Get("page"); pageStr != "" {
				if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
					page = p
				}
			}
			if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
				if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
					pageSize = ps
				}
			}
			if searchStr := r.URL.Query().Get("search"); searchStr != "" {
				search = searchStr
			}
			
			tenantID := server.getDefaultTenantID()
			response, err := server.corehrService.ListEmployees(r.Context(), tenantID, page, pageSize, search)
			if err != nil {
				server.handleError(w, err, "Failed to list employees")
				return
			}
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		})
		
		r.Post("/employees", server.CreateEmployee)
		
		r.Get("/employees/{employee_id}", func(w http.ResponseWriter, r *http.Request) {
			employeeIDStr := chi.URLParam(r, "employee_id")
			employeeID, err := uuid.Parse(employeeIDStr)
			if err != nil {
				server.sendErrorResponse(w, "Invalid employee ID", http.StatusBadRequest)
				return
			}
			
			tenantID := server.getDefaultTenantID()
			employee, err := server.corehrService.GetEmployee(r.Context(), tenantID, employeeID)
			if err != nil {
				server.handleError(w, err, "Failed to get employee")
				return
			}
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(employee)
		})
		
		r.Put("/employees/{employee_id}", func(w http.ResponseWriter, r *http.Request) {
			employeeIDStr := chi.URLParam(r, "employee_id")
			employeeID, err := uuid.Parse(employeeIDStr)
			if err != nil {
				server.sendErrorResponse(w, "Invalid employee ID", http.StatusBadRequest)
				return
			}
			
			var req openapi.UpdateEmployeeRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				server.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
				return
			}
			
			tenantID := server.getDefaultTenantID()
			employee, err := server.corehrService.UpdateEmployee(r.Context(), tenantID, employeeID, &req)
			if err != nil {
				server.handleError(w, err, "Failed to update employee")
				return
			}
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(employee)
		})
		
		r.Delete("/employees/{employee_id}", func(w http.ResponseWriter, r *http.Request) {
			employeeIDStr := chi.URLParam(r, "employee_id")
			employeeID, err := uuid.Parse(employeeIDStr)
			if err != nil {
				server.sendErrorResponse(w, "Invalid employee ID", http.StatusBadRequest)
				return
			}
			
			tenantID := server.getDefaultTenantID()
			err = server.corehrService.DeleteEmployee(r.Context(), tenantID, employeeID)
			if err != nil {
				server.handleError(w, err, "Failed to delete employee")
				return
			}
			
			w.WriteHeader(http.StatusNoContent)
		})
		
		r.Get("/organizations", server.ListOrganizations)
		r.Get("/organizations/tree", server.GetOrganizationTree)
	})

	// 静态文件服务（用于提供测试页面）
	router.Get("/test.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../test.html")
	})
	
	router.Get("/verify_1.1.1.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./verify_1.1.1.html")
	})

	// 注册 AI 服务路由
	router.Route("/api/v1/intelligence", func(r chi.Router) {
		r.Post("/interpret", server.InterpretQuery)
	})
	
	// 注册发件箱管理路由
	if outboxService != nil {
		router.Route("/api/v1/outbox", func(r chi.Router) {
			r.Get("/stats", server.GetOutboxStats)
			r.Get("/events", server.GetOutboxEvents)
			r.Post("/replay", server.ReplayEvents)
			r.Get("/unprocessed", server.GetUnprocessedEvents)
		})
	}
	
	// 注册其他 API 路由（OpenAPI 生成的路由）
	// router.Mount("/api/v1/openapi", openapi.Handler(server)) // 暂时注释掉，因为我们使用 Chi 而不是 Echo

	// 调试路由 - 显示所有注册的路由
	router.Get("/debug/routes", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "CoreHR API routes registered",
			"endpoints": []string{
				"GET /api/v1/corehr/employees",
				"POST /api/v1/corehr/employees",
				"GET /api/v1/corehr/employees/{id}",
				"PUT /api/v1/corehr/employees/{id}",
				"DELETE /api/v1/corehr/employees/{id}",
				"GET /api/v1/corehr/organizations",
				"GET /api/v1/corehr/organizations/tree",
				"POST /api/v1/intelligence/interpret",
				"GET /test.html",
				"GET /health",
				"GET /health/db",
			},
		})
	})

	fmt.Printf("🏰 Go Monolith 'The Keep' is listening on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

// initDatabase 初始化数据库
func initDatabase() {
	dbConfig := common.NewDatabaseConfig()
	db, err := common.Connect(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to databases: %v", err)
	}
	defer db.Close()

	if err := common.InitDatabase(db); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
}

// seedDatabase 插入种子数据
func seedDatabase() {
	dbConfig := common.NewDatabaseConfig()
	db, err := common.Connect(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to databases: %v", err)
	}
	defer db.Close()

	if err := common.SeedData(db); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}
}

// CoreHR API 实现

// getDefaultTenantID 获取默认租户ID（临时实现，后续会从JWT或请求头中获取）
func (s *Server) getDefaultTenantID() uuid.UUID {
	// 临时使用硬编码的默认租户ID，后续会从JWT或请求头中获取
	defaultTenantID, _ := uuid.Parse("00000000-0000-0000-0000-000000000000")
	return defaultTenantID
}

// ListEmployees - 获取员工列表
func (s *Server) ListEmployees(w http.ResponseWriter, r *http.Request, params openapi.ListEmployeesParams) {
	page := 1
	if params.Page != nil {
		page = *params.Page
	}
	
	pageSize := 20
	if params.PageSize != nil {
		pageSize = *params.PageSize
	}
	
	search := ""
	if params.Search != nil {
		search = *params.Search
	}
	
	tenantID := s.getDefaultTenantID()
	response, err := s.corehrService.ListEmployees(r.Context(), tenantID, page, pageSize, search)
	if err != nil {
		s.handleError(w, err, "Failed to list employees")
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CreateEmployee - 创建员工
func (s *Server) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	var req openapi.CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	tenantID := s.getDefaultTenantID()
	employee, err := s.corehrService.CreateEmployee(r.Context(), tenantID, &req)
	if err != nil {
		s.handleError(w, err, "Failed to create employee")
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(employee)
}

// GetEmployee - 获取员工详情
func (s *Server) GetEmployee(w http.ResponseWriter, r *http.Request, employeeId uuid.UUID) {
	tenantID := s.getDefaultTenantID()
	employee, err := s.corehrService.GetEmployee(r.Context(), tenantID, employeeId)
	if err != nil {
		s.handleError(w, err, "Failed to get employee")
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(employee)
}

// UpdateEmployee - 更新员工
func (s *Server) UpdateEmployee(w http.ResponseWriter, r *http.Request, employeeId uuid.UUID) {
	var req openapi.UpdateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	tenantID := s.getDefaultTenantID()
	employee, err := s.corehrService.UpdateEmployee(r.Context(), tenantID, employeeId, &req)
	if err != nil {
		s.handleError(w, err, "Failed to update employee")
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(employee)
}

// DeleteEmployee - 删除员工
func (s *Server) DeleteEmployee(w http.ResponseWriter, r *http.Request, employeeId uuid.UUID) {
	tenantID := s.getDefaultTenantID()
	err := s.corehrService.DeleteEmployee(r.Context(), tenantID, employeeId)
	if err != nil {
		s.handleError(w, err, "Failed to delete employee")
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// ListOrganizations - 获取组织列表
func (s *Server) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	tenantID := s.getDefaultTenantID()
	response, err := s.corehrService.ListOrganizations(r.Context(), tenantID)
	if err != nil {
		s.handleError(w, err, "Failed to list organizations")
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetOrganizationTree - 获取组织树
func (s *Server) GetOrganizationTree(w http.ResponseWriter, r *http.Request) {
	tenantID := s.getDefaultTenantID()
	response, err := s.corehrService.GetOrganizationTree(r.Context(), tenantID)
	if err != nil {
		s.handleError(w, err, "Failed to get organization tree")
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetManagerByEmployeeId - 根据员工ID获取经理
func (s *Server) GetManagerByEmployeeId(w http.ResponseWriter, r *http.Request, employeeId uuid.UUID) {
	tenantID := s.getDefaultTenantID()
	manager, err := s.corehrService.GetManagerByEmployeeId(r.Context(), tenantID, employeeId)
	if err != nil {
		s.handleError(w, err, "Failed to get manager")
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(manager)
}

// 发件箱管理API

// GetOutboxEvents - 获取发件箱事件列表
func (s *Server) GetOutboxEvents(w http.ResponseWriter, r *http.Request) {
	if s.outboxService == nil {
		s.sendErrorResponse(w, "Outbox service not available", http.StatusServiceUnavailable)
		return
	}
	
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	events, err := s.outboxService.GetEvents(r.Context(), limit)
	if err != nil {
		s.handleError(w, err, "Failed to get outbox events")
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events": events,
		"count":  len(events),
		"limit":  limit,
	})
}

// GetOutboxStats - 获取发件箱统计信息
func (s *Server) GetOutboxStats(w http.ResponseWriter, r *http.Request) {
	if s.outboxService == nil {
		s.sendErrorResponse(w, "Outbox service not available", http.StatusServiceUnavailable)
		return
	}
	
	stats, err := s.outboxService.GetStats(r.Context())
	if err != nil {
		s.handleError(w, err, "Failed to get outbox stats")
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stats)
}

// ReplayEvents - 重放事件
func (s *Server) ReplayEvents(w http.ResponseWriter, r *http.Request) {
	if s.outboxService == nil {
		s.sendErrorResponse(w, "Outbox service not available", http.StatusServiceUnavailable)
		return
	}
	
	// 从请求体中获取aggregate_id
	var req struct {
		AggregateID string `json:"aggregate_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if req.AggregateID == "" {
		s.sendErrorResponse(w, "Missing aggregate_id", http.StatusBadRequest)
		return
	}
	
	aggregateID, err := uuid.Parse(req.AggregateID)
	if err != nil {
		s.sendErrorResponse(w, "Invalid aggregate ID", http.StatusBadRequest)
		return
	}
	
	err = s.outboxService.ReplayEvents(r.Context(), aggregateID)
	if err != nil {
		s.handleError(w, err, "Failed to replay events")
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Events replayed successfully",
		"aggregate_id": aggregateID.String(),
	})
}

// GetUnprocessedEvents - 获取未处理事件
func (s *Server) GetUnprocessedEvents(w http.ResponseWriter, r *http.Request) {
	if s.outboxService == nil {
		s.sendErrorResponse(w, "Outbox service not available", http.StatusServiceUnavailable)
		return
	}
	
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	events, err := s.outboxService.GetUnprocessedEvents(r.Context(), limit)
	if err != nil {
		s.handleError(w, err, "Failed to get unprocessed events")
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events": events,
		"count":  len(events),
		"limit":  limit,
	})
}

// 错误处理辅助方法

// handleError 统一处理错误响应
func (s *Server) handleError(w http.ResponseWriter, err error, defaultMessage string) {
	errorMessage := err.Error()
	
	// 根据错误类型返回不同的状态码
	switch {
	case strings.Contains(errorMessage, "not found"):
		s.sendErrorResponse(w, errorMessage, http.StatusNotFound)
	case strings.Contains(errorMessage, "already exists"):
		s.sendErrorResponse(w, errorMessage, http.StatusConflict)
	case strings.Contains(errorMessage, "validation failed"):
		s.sendErrorResponse(w, errorMessage, http.StatusBadRequest)
	case strings.Contains(errorMessage, "invalid"):
		s.sendErrorResponse(w, errorMessage, http.StatusBadRequest)
	default:
		// 对于内部错误，不暴露详细错误信息给客户端
		log.Printf("Internal error: %v", err)
		s.sendErrorResponse(w, defaultMessage, http.StatusInternalServerError)
	}
}

// sendErrorResponse 发送标准化的错误响应
func (s *Server) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"message": message,
			"status":  statusCode,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}
	
	json.NewEncoder(w).Encode(errorResponse)
}
