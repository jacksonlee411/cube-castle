package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/cube-castle/cmd/organization-command-server/internal/infrastructure/logging"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string                     `json:"status"`
	Timestamp time.Time                  `json:"timestamp"`
	Checks    map[string]ComponentHealth `json:"checks"`
	Version   string                     `json:"version"`
	Uptime    string                     `json:"uptime"`
}

// ComponentHealth represents the health of a specific component
type ComponentHealth struct {
	Status  string        `json:"status"`
	Latency time.Duration `json:"latency"`
	Error   string        `json:"error,omitempty"`
}

// HealthHandler handles health check requests
type HealthHandler struct {
	dbPool      *pgxpool.Pool
	kafkaAdmin  kafka.AdminClient
	logger      logging.Logger
	version     string
	startTime   time.Time
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(
	dbPool *pgxpool.Pool,
	kafkaAdmin kafka.AdminClient,
	logger logging.Logger,
	version string,
) *HealthHandler {
	return &HealthHandler{
		dbPool:     dbPool,
		kafkaAdmin: kafkaAdmin,
		logger:     logger,
		version:    version,
		startTime:  time.Now(),
	}
}

// CheckHealth handles GET /health
func (h *HealthHandler) CheckHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	checks := make(map[string]ComponentHealth)
	overall := "healthy"
	
	// Check database
	dbHealth := h.checkDatabase(ctx)
	checks["database"] = dbHealth
	if dbHealth.Status != "healthy" {
		overall = "unhealthy"
	}
	
	// Check Kafka
	kafkaHealth := h.checkKafka(ctx)
	checks["kafka"] = kafkaHealth
	if kafkaHealth.Status != "healthy" && overall != "unhealthy" {
		overall = "degraded"
	}
	
	response := HealthResponse{
		Status:    overall,
		Timestamp: time.Now(),
		Checks:    checks,
		Version:   h.version,
		Uptime:    time.Since(h.startTime).String(),
	}
	
	// Set appropriate HTTP status code
	statusCode := http.StatusOK
	switch overall {
	case "unhealthy":
		statusCode = http.StatusServiceUnavailable
	case "degraded":
		statusCode = http.StatusOK // Still accepting traffic but with warnings
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
	
	// Log health check
	h.logger.Debug("health check performed",
		"status", overall,
		"checks", len(checks),
	)
}

// checkDatabase checks database connectivity
func (h *HealthHandler) checkDatabase(ctx context.Context) ComponentHealth {
	start := time.Now()
	
	err := h.dbPool.Ping(ctx)
	latency := time.Since(start)
	
	if err != nil {
		h.logger.Warn("database health check failed", "error", err, "latency", latency)
		return ComponentHealth{
			Status:  "unhealthy",
			Latency: latency,
			Error:   err.Error(),
		}
	}
	
	return ComponentHealth{
		Status:  "healthy",
		Latency: latency,
	}
}

// checkKafka checks Kafka connectivity
func (h *HealthHandler) checkKafka(ctx context.Context) ComponentHealth {
	start := time.Now()
	
	// Simple metadata request to check Kafka connectivity
	metadata, err := h.kafkaAdmin.GetMetadata(nil, false, 5000) // 5 second timeout
	latency := time.Since(start)
	
	if err != nil {
		h.logger.Warn("kafka health check failed", "error", err, "latency", latency)
		return ComponentHealth{
			Status:  "unhealthy",
			Latency: latency,
			Error:   err.Error(),
		}
	}
	
	// Check if we have any brokers
	if len(metadata.Brokers) == 0 {
		return ComponentHealth{
			Status:  "unhealthy",
			Latency: latency,
			Error:   "no kafka brokers available",
		}
	}
	
	return ComponentHealth{
		Status:  "healthy",
		Latency: latency,
	}
}

// CheckReadiness handles GET /ready
func (h *HealthHandler) CheckReadiness(w http.ResponseWriter, r *http.Request) {
	// Simple readiness check - just ensure we can respond
	response := map[string]interface{}{
		"status":    "ready",
		"timestamp": time.Now(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CheckLiveness handles GET /live
func (h *HealthHandler) CheckLiveness(w http.ResponseWriter, r *http.Request) {
	// Simple liveness check - just ensure the service is running
	response := map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now(),
		"uptime":    time.Since(h.startTime).String(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}