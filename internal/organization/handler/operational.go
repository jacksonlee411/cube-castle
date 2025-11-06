package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"cube-castle/internal/organization/middleware"
	scheduler "cube-castle/internal/organization/scheduler"
	pkglogger "cube-castle/pkg/logger"
	"github.com/go-chi/chi/v5"
)

// OperationalHandler 运维管理处理器
type OperationalHandler struct {
	monitor   *scheduler.TemporalMonitor
	scheduler *scheduler.OperationalScheduler
	logger    pkglogger.Logger
	rateLimit *middleware.RateLimitMiddleware
}

// NewOperationalHandler 创建运维管理处理器
func NewOperationalHandler(monitor *scheduler.TemporalMonitor, scheduler *scheduler.OperationalScheduler, rateLimit *middleware.RateLimitMiddleware, baseLogger pkglogger.Logger) *OperationalHandler {
	return &OperationalHandler{
		monitor:   monitor,
		scheduler: scheduler,
		rateLimit: rateLimit,
		logger:    scopedLogger(baseLogger, "operational", pkglogger.Fields{"module": "operational"}),
	}
}

func (h *OperationalHandler) requestLogger(r *http.Request, action string, extra pkglogger.Fields) pkglogger.Logger {
	return requestScopedLogger(h.logger, r, action, extra)
}

// SetupRoutes 设置运维管理路由
func (h *OperationalHandler) SetupRoutes(r chi.Router) {
	r.Route("/api/v1/operational", func(r chi.Router) {
		// 监控相关端点
		r.Get("/health", h.GetHealth)
		r.Get("/metrics", h.GetMetrics)
		r.Get("/alerts", h.GetAlerts)
		r.Get("/rate-limit/stats", h.GetRateLimitStats)

		// 任务调度相关端点
		r.Get("/tasks", h.GetTasks)
		r.Get("/tasks/status", h.GetTaskStatus)
		r.Post("/tasks/{taskName}/trigger", h.TriggerTask)

		// 系统操作端点
		r.Post("/cutover", h.TriggerCutover)
		r.Post("/consistency-check", h.TriggerConsistencyCheck)
	})
}

// GetRateLimitStats 获取限流统计（受PBAC保护）
func (h *OperationalHandler) GetRateLimitStats(w http.ResponseWriter, r *http.Request) {
	stats := h.rateLimit.GetStats()
	logger := h.requestLogger(r, "GetRateLimitStats", pkglogger.Fields{
		"totalRequests":   stats.TotalRequests,
		"blockedRequests": stats.BlockedRequests,
		"activeClients":   stats.ActiveClients,
	})
	response := map[string]interface{}{
		"success":   true,
		"timestamp": time.Now().Format(time.RFC3339),
		"data": map[string]interface{}{
			"totalRequests":   stats.TotalRequests,
			"blockedRequests": stats.BlockedRequests,
			"activeClients":   stats.ActiveClients,
			"lastReset":       stats.LastReset.Format(time.RFC3339),
			// 统一固定类型为字符串百分比，前端展示直用
			"blockRate": func() string {
				if stats.TotalRequests == 0 {
					return "0.00%"
				}
				return fmt.Sprintf("%.2f%%", float64(stats.BlockedRequests)/float64(stats.TotalRequests)*100)
			}(),
		},
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Error("encode rate limit stats response failed")
	}
}

// GetHealth 获取系统健康状态
func (h *OperationalHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	logger := h.requestLogger(r, "GetHealth", nil)

	metrics, err := h.monitor.CollectMetrics(ctx)
	if err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Error("collect health metrics failed")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	healthResponse := map[string]interface{}{
		"success":   true,
		"timestamp": time.Now().Format(time.RFC3339),
		"data": map[string]interface{}{
			"status":      metrics.AlertLevel,
			"healthScore": metrics.HealthScore,
			"summary": map[string]interface{}{
				"totalOrganizations": metrics.TotalOrganizations,
				"currentRecords":     metrics.CurrentRecords,
				"futureRecords":      metrics.FutureRecords,
				"historicalRecords":  metrics.HistoricalRecords,
			},
			"issues": map[string]interface{}{
				"duplicateCurrentCount": metrics.DuplicateCurrentCount,
				"missingCurrentCount":   metrics.MissingCurrentCount,
				"timelineOverlapCount":  metrics.TimelineOverlapCount,
				"inconsistentFlagCount": metrics.InconsistentFlagCount,
				"orphanRecordCount":     metrics.OrphanRecordCount,
			},
			"lastCheckTime": metrics.LastCheckTime.Format(time.RFC3339),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(healthResponse); err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Error("encode health response failed")
	}
}

// GetMetrics 获取详细监控指标
func (h *OperationalHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	logger := h.requestLogger(r, "GetMetrics", nil)

	metrics, err := h.monitor.CollectMetrics(ctx)
	if err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Error("collect monitoring metrics failed")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":   true,
		"timestamp": time.Now().Format(time.RFC3339),
		"data":      metrics,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Error("encode metrics response failed")
	}
}

// GetAlerts 获取当前告警
func (h *OperationalHandler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	logger := h.requestLogger(r, "GetAlerts", nil)

	alerts, err := h.monitor.CheckAlerts(ctx)
	if err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Error("fetch alerts failed")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":   true,
		"timestamp": time.Now().Format(time.RFC3339),
		"data": map[string]interface{}{
			"alertCount": len(alerts),
			"alerts":     alerts,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Error("encode alerts response failed")
	}
}

// GetTasks 获取任务配置
func (h *OperationalHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	logger := h.requestLogger(r, "GetTasks", nil)
	if h.scheduler == nil {
		http.Error(w, "Scheduler module disabled", http.StatusServiceUnavailable)
		return
	}
	tasks := h.scheduler.ListTasks()

	response := map[string]interface{}{
		"success":   true,
		"timestamp": time.Now().Format(time.RFC3339),
		"data": map[string]interface{}{
			"tasks":            tasks,
			"schedulerRunning": h.scheduler.IsRunning(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Error("encode tasks response failed")
	}
}

// GetTaskStatus 获取任务运行状态
func (h *OperationalHandler) GetTaskStatus(w http.ResponseWriter, r *http.Request) {
	logger := h.requestLogger(r, "GetTaskStatus", nil)
	if h.scheduler == nil {
		http.Error(w, "Scheduler module disabled", http.StatusServiceUnavailable)
		return
	}
	tasks := h.scheduler.ListTasks()

	// 计算任务统计
	var enabledCount, runningCount int
	for _, task := range tasks {
		if task.Enabled {
			enabledCount++
		}
		if task.Running {
			runningCount++
		}
	}

	response := map[string]interface{}{
		"success":   true,
		"timestamp": time.Now().Format(time.RFC3339),
		"data": map[string]interface{}{
			"totalTasks":       len(tasks),
			"enabledTasks":     enabledCount,
			"runningTasks":     runningCount,
			"schedulerRunning": h.scheduler.IsRunning(),
			"tasks":            tasks,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Error("encode task status response failed")
	}
}

// TriggerTask 手动触发任务
func (h *OperationalHandler) TriggerTask(w http.ResponseWriter, r *http.Request) {
	taskName := chi.URLParam(r, "taskName")
	if taskName == "" {
		http.Error(w, "Task name is required", http.StatusBadRequest)
		return
	}
	if h.scheduler == nil {
		http.Error(w, "Scheduler module disabled", http.StatusServiceUnavailable)
		return
	}
	logger := h.requestLogger(r, "TriggerTask", pkglogger.Fields{"taskName": taskName})

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	if err := h.scheduler.RunTask(ctx, taskName); err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Error("manual task execution failed")
		response := map[string]interface{}{
			"success":   false,
			"timestamp": time.Now().Format(time.RFC3339),
			"error": map[string]interface{}{
				"message": "Task execution failed",
				"details": err.Error(),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.WithFields(pkglogger.Fields{"error": err}).Error("encode manual task error response failed")
		}
		return
	}

	response := map[string]interface{}{
		"success":   true,
		"timestamp": time.Now().Format(time.RFC3339),
		"data": map[string]interface{}{
			"taskName": taskName,
			"message":  fmt.Sprintf("%s triggered successfully", taskName),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Error("encode manual task success response failed")
	}
}

// TriggerCutover 手动触发cutover操作
func (h *OperationalHandler) TriggerCutover(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	logger := h.requestLogger(r, "TriggerCutover", nil)

	if h.scheduler == nil {
		http.Error(w, "Scheduler module disabled", http.StatusServiceUnavailable)
		return
	}

	err := h.scheduler.RunTask(ctx, "daily_cutover")
	if err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Error("manual cutover failed")
		response := map[string]interface{}{
			"success":   false,
			"timestamp": time.Now().Format(time.RFC3339),
			"error": map[string]interface{}{
				"message": "Cutover operation failed",
				"details": err.Error(),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.WithFields(pkglogger.Fields{"error": err}).Error("encode cutover error response failed")
		}
		return
	}

	response := map[string]interface{}{
		"success":   true,
		"timestamp": time.Now().Format(time.RFC3339),
		"data": map[string]interface{}{
			"message": "Cutover操作已完成",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Error("encode cutover success response failed")
	}
}

// TriggerConsistencyCheck 手动触发一致性检查
func (h *OperationalHandler) TriggerConsistencyCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	logger := h.requestLogger(r, "TriggerConsistencyCheck", nil)

	if h.scheduler == nil {
		http.Error(w, "Scheduler module disabled", http.StatusServiceUnavailable)
		return
	}

	err := h.scheduler.RunTask(ctx, "data_consistency_check")
	if err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Error("manual consistency check failed")
		response := map[string]interface{}{
			"success":   false,
			"timestamp": time.Now().Format(time.RFC3339),
			"error": map[string]interface{}{
				"message": "Consistency check failed",
				"details": err.Error(),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.WithFields(pkglogger.Fields{"error": err}).Error("encode consistency check error response failed")
		}
		return
	}

	response := map[string]interface{}{
		"success":   true,
		"timestamp": time.Now().Format(time.RFC3339),
		"data": map[string]interface{}{
			"message": "数据一致性检查已完成",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Error("encode consistency check success response failed")
	}
}

// 辅助方法已移除，统一通过 scheduler.RunTask 管理任务触发
