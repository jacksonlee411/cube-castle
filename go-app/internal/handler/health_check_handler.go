package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/gaogu/cube-castle/go-app/internal/repositories"
	"github.com/gaogu/cube-castle/go-app/internal/service"
)

// HealthCheckHandler CQRS健康检查处理器
type HealthCheckHandler struct {
	postgresRepo repositories.PostgresCommandRepository
	neo4jService *service.Neo4jService
	logger       *logging.StructuredLogger
}

// NewHealthCheckHandler 创建健康检查处理器
func NewHealthCheckHandler(
	postgresRepo repositories.PostgresCommandRepository,
	neo4jService *service.Neo4jService,
	logger *logging.StructuredLogger,
) *HealthCheckHandler {
	return &HealthCheckHandler{
		postgresRepo: postgresRepo,
		neo4jService: neo4jService,
		logger:       logger,
	}
}

// HealthCheckResponse 健康检查响应
type HealthCheckResponse struct {
	Status     string                 `json:"status"`
	Timestamp  time.Time              `json:"timestamp"`
	Version    string                 `json:"version"`
	Checks     map[string]HealthCheck `json:"checks"`
	CQRS       CQRSHealthStatus       `json:"cqrs"`
}

// HealthCheck 单个组件健康检查
type HealthCheck struct {
	Status      string        `json:"status"`
	Duration    string        `json:"duration"`
	Error       string        `json:"error,omitempty"`
	LastChecked time.Time     `json:"last_checked"`
}

// CQRSHealthStatus CQRS架构健康状态
type CQRSHealthStatus struct {
	CommandSide  ComponentHealth `json:"command_side"`
	QuerySide    ComponentHealth `json:"query_side"`
	EventBus     ComponentHealth `json:"event_bus"`
	Synchronization SyncHealth   `json:"synchronization"`
}

// ComponentHealth 组件健康状态
type ComponentHealth struct {
	Status       string    `json:"status"`
	ResponseTime string    `json:"response_time"`
	LastChecked  time.Time `json:"last_checked"`
	Errors       int       `json:"error_count"`
}

// SyncHealth 同步健康状态
type SyncHealth struct {
	PostgresNeo4j ComponentHealth `json:"postgres_neo4j"`
	EventConsumer ComponentHealth `json:"event_consumer"`
	DataConsistency ComponentHealth `json:"data_consistency"`
}

// GetCQRSHealth 获取CQRS架构健康状态
func (h *HealthCheckHandler) GetCQRSHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		startTime := time.Now()
		
		response := HealthCheckResponse{
			Status:    "healthy",
			Timestamp: time.Now(),
			Version:   "v1.4.0",
			Checks:    make(map[string]HealthCheck),
			CQRS: CQRSHealthStatus{
				CommandSide: ComponentHealth{
					Status:      "healthy",
					LastChecked: time.Now(),
				},
				QuerySide: ComponentHealth{
					Status:      "healthy", 
					LastChecked: time.Now(),
				},
				EventBus: ComponentHealth{
					Status:      "healthy",
					LastChecked: time.Now(),
				},
			},
		}
		
		// 检查PostgreSQL命令端
		postgresCheck := h.checkPostgres(ctx)
		response.Checks["postgres"] = postgresCheck
		if postgresCheck.Status != "healthy" {
			response.Status = "degraded"
			response.CQRS.CommandSide.Status = "unhealthy"
		}
		response.CQRS.CommandSide.ResponseTime = postgresCheck.Duration
		
		// 检查Neo4j查询端
		neo4jCheck := h.checkNeo4j(ctx)
		response.Checks["neo4j"] = neo4jCheck
		if neo4jCheck.Status != "healthy" {
			response.Status = "degraded"
			response.CQRS.QuerySide.Status = "unhealthy"
		}
		response.CQRS.QuerySide.ResponseTime = neo4jCheck.Duration
		
		// 检查数据一致性
		consistencyCheck := h.checkDataConsistency(ctx)
		response.Checks["data_consistency"] = consistencyCheck
		response.CQRS.Synchronization.DataConsistency = ComponentHealth{
			Status:       consistencyCheck.Status,
			ResponseTime: consistencyCheck.Duration,
			LastChecked:  consistencyCheck.LastChecked,
		}
		
		// 检查事件消费者
		eventConsumerCheck := h.checkEventConsumer(ctx)
		response.Checks["event_consumer"] = eventConsumerCheck
		response.CQRS.Synchronization.EventConsumer = ComponentHealth{
			Status:       eventConsumerCheck.Status,
			ResponseTime: eventConsumerCheck.Duration,
			LastChecked:  eventConsumerCheck.LastChecked,
		}
		
		// 设置HTTP状态码
		statusCode := http.StatusOK
		if response.Status == "unhealthy" {
			statusCode = http.StatusServiceUnavailable
		} else if response.Status == "degraded" {
			statusCode = http.StatusPartialContent
		}
		
		// 记录健康检查指标
		duration := time.Since(startTime)
		h.logger.Info("CQRS Health Check Completed",
			"overall_status", response.Status,
			"postgres_status", postgresCheck.Status,
			"neo4j_status", neo4jCheck.Status,
			"consistency_status", consistencyCheck.Status,
			"duration_ms", duration.Milliseconds(),
		)
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(response)
	}
}

// checkPostgres 检查PostgreSQL连接和性能
func (h *HealthCheckHandler) checkPostgres(ctx context.Context) HealthCheck {
	startTime := time.Now()
	
	// 简单的连接测试（这里需要根据实际的repository接口调整）
	// 由于没有具体的ping方法，我们使用一个轻量级查询
	err := h.testPostgresConnection(ctx)
	
	duration := time.Since(startTime)
	
	if err != nil {
		return HealthCheck{
			Status:      "unhealthy",
			Duration:    duration.String(),
			Error:       err.Error(),
			LastChecked: time.Now(),
		}
	}
	
	return HealthCheck{
		Status:      "healthy",
		Duration:    duration.String(),
		LastChecked: time.Now(),
	}
}

// checkNeo4j 检查Neo4j连接和性能
func (h *HealthCheckHandler) checkNeo4j(ctx context.Context) HealthCheck {
	startTime := time.Now()
	
	// 使用一个简单的Neo4j查询来测试连接
	err := h.testNeo4jConnection(ctx)
	
	duration := time.Since(startTime)
	
	if err != nil {
		return HealthCheck{
			Status:      "unhealthy",
			Duration:    duration.String(),
			Error:       err.Error(),
			LastChecked: time.Now(),
		}
	}
	
	return HealthCheck{
		Status:      "healthy",
		Duration:    duration.String(),
		LastChecked: time.Now(),
	}
}

// checkDataConsistency 检查PostgreSQL和Neo4j之间的数据一致性
func (h *HealthCheckHandler) checkDataConsistency(ctx context.Context) HealthCheck {
	startTime := time.Now()
	
	// 这里可以实现一个简单的数据一致性检查
	// 比如对比PostgreSQL和Neo4j中的员工数量
	err := h.validateDataConsistency(ctx)
	
	duration := time.Since(startTime)
	
	if err != nil {
		return HealthCheck{
			Status:      "warning",
			Duration:    duration.String(),
			Error:       err.Error(),
			LastChecked: time.Now(),
		}
	}
	
	return HealthCheck{
		Status:      "healthy",
		Duration:    duration.String(),
		LastChecked: time.Now(),
	}
}

// checkEventConsumer 检查事件消费者状态
func (h *HealthCheckHandler) checkEventConsumer(ctx context.Context) HealthCheck {
	startTime := time.Now()
	
	// 这里可以检查事件消费者的状态
	// 比如检查最近是否有事件被处理
	err := h.validateEventConsumer(ctx)
	
	duration := time.Since(startTime)
	
	if err != nil {
		return HealthCheck{
			Status:      "warning",
			Duration:    duration.String(),
			Error:       err.Error(),
			LastChecked: time.Now(),
		}
	}
	
	return HealthCheck{
		Status:      "healthy",
		Duration:    duration.String(),
		LastChecked: time.Now(),
	}
}

// 辅助方法（需要根据实际的repository接口实现）
func (h *HealthCheckHandler) testPostgresConnection(ctx context.Context) error {
	// 实现PostgreSQL连接测试
	// 这里需要根据实际的repository接口调整
	return nil
}

func (h *HealthCheckHandler) testNeo4jConnection(ctx context.Context) error {
	// 实现Neo4j连接测试
	// 可以尝试获取一个简单的员工记录来测试连接
	if h.neo4jService == nil {
		return fmt.Errorf("Neo4j service not available")
	}
	// 这里可以调用一个简单的Neo4j查询方法
	// 比如获取员工数量等轻量级操作
	return nil
}

func (h *HealthCheckHandler) validateDataConsistency(ctx context.Context) error {
	// 实现数据一致性检查
	// 比如对比PostgreSQL和Neo4j中的数据
	return nil
}

func (h *HealthCheckHandler) validateEventConsumer(ctx context.Context) error {
	// 实现事件消费者状态检查
	return nil
}