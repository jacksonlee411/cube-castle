/**
 * 组织时间线管理服务
 * 专门处理组织的时态事件、时间线查询和版本历史管理
 */
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ===== 时间线事件数据模型 =====

type TimelineEvent struct {
	ID             string                 `json:"id" db:"id"`
	OrganizationCode string               `json:"organization_code" db:"organization_code"`
	EventType      string                 `json:"event_type" db:"event_type"`
	EventDate      time.Time              `json:"event_date" db:"event_date"`
	EffectiveDate  *time.Time             `json:"effective_date,omitempty" db:"effective_date"`
	Status         string                 `json:"status" db:"status"`
	Title          string                 `json:"title" db:"title"`
	Description    string                 `json:"description,omitempty" db:"description"`
	Metadata       map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	PreviousValue  map[string]interface{} `json:"previous_value,omitempty" db:"previous_value"`
	NewValue       map[string]interface{} `json:"new_value,omitempty" db:"new_value"`
	AffectedFields []string               `json:"affected_fields,omitempty" db:"affected_fields"`
	TriggeredBy    *string                `json:"triggered_by,omitempty" db:"triggered_by"`
	ApprovedBy     *string                `json:"approved_by,omitempty" db:"approved_by"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
	TenantID       string                 `json:"tenant_id" db:"tenant_id"`
}

// 时间线查询选项
type TimelineQueryOptions struct {
	StartDate      *time.Time `json:"start_date,omitempty"`
	EndDate        *time.Time `json:"end_date,omitempty"`
	EventTypes     []string   `json:"event_types,omitempty"`
	Status         []string   `json:"status,omitempty"`
	Limit          int        `json:"limit,omitempty"`
	Offset         int        `json:"offset,omitempty"`
	OrderBy        string     `json:"order_by,omitempty"` // event_date, effective_date
	OrderDirection string     `json:"order_direction,omitempty"` // asc, desc
	IncludeMetadata bool      `json:"include_metadata,omitempty"`
}

// 时间线统计信息
type TimelineStats struct {
	TotalEvents       int                    `json:"total_events"`
	EventsByType      map[string]int         `json:"events_by_type"`
	EventsByStatus    map[string]int         `json:"events_by_status"`
	RecentEvents      []TimelineEvent        `json:"recent_events"`
	TimelineSpan      *TimelineSpan          `json:"timeline_span,omitempty"`
	MonthlyActivity   []MonthlyEventCount    `json:"monthly_activity,omitempty"`
}

type TimelineSpan struct {
	EarliestEvent time.Time `json:"earliest_event"`
	LatestEvent   time.Time `json:"latest_event"`
	SpanDays      int       `json:"span_days"`
}

type MonthlyEventCount struct {
	Month      string `json:"month"`      // YYYY-MM
	EventCount int    `json:"event_count"`
}

// 版本历史信息
type OrganizationVersion struct {
	ID              string                 `json:"id" db:"id"`
	OrganizationCode string                `json:"organization_code" db:"organization_code"`
	Version         int                    `json:"version" db:"version"`
	EffectiveFrom   time.Time              `json:"effective_from" db:"effective_from"`
	EffectiveTo     *time.Time             `json:"effective_to,omitempty" db:"effective_to"`
	SnapshotData    map[string]interface{} `json:"snapshot_data" db:"snapshot_data"`
	ChangeReason    string                 `json:"change_reason" db:"change_reason"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	TenantID        string                 `json:"tenant_id" db:"tenant_id"`
}

// 版本比较结果
type VersionComparison struct {
	FromVersion     int                         `json:"from_version"`
	ToVersion       int                         `json:"to_version"`
	ComparedAt      time.Time                   `json:"compared_at"`
	FieldChanges    []FieldChange               `json:"field_changes"`
	Summary         VersionComparisonSummary    `json:"summary"`
}

type FieldChange struct {
	Field         string      `json:"field"`
	OldValue      interface{} `json:"old_value"`
	NewValue      interface{} `json:"new_value"`
	ChangeType    string      `json:"change_type"` // added, removed, modified
}

type VersionComparisonSummary struct {
	TotalChanges    int `json:"total_changes"`
	AddedFields     int `json:"added_fields"`
	RemovedFields   int `json:"removed_fields"`
	ModifiedFields  int `json:"modified_fields"`
}

// ===== 时间线仓储层 =====

type TimelineRepository struct {
	db *sql.DB
}

func NewTimelineRepository(db *sql.DB) *TimelineRepository {
	return &TimelineRepository{db: db}
}

// 获取组织的时间线事件
func (r *TimelineRepository) GetTimeline(ctx context.Context, tenantID uuid.UUID, orgCode string, opts *TimelineQueryOptions) ([]TimelineEvent, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	// 基础条件
	conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIndex))
	args = append(args, tenantID.String())
	argIndex++

	conditions = append(conditions, fmt.Sprintf("organization_code = $%d", argIndex))
	args = append(args, orgCode)
	argIndex++

	// 时间范围筛选
	if opts.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("event_date >= $%d", argIndex))
		args = append(args, *opts.StartDate)
		argIndex++
	}

	if opts.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("event_date <= $%d", argIndex))
		args = append(args, *opts.EndDate)
		argIndex++
	}

	// 事件类型筛选
	if len(opts.EventTypes) > 0 {
		placeholders := make([]string, len(opts.EventTypes))
		for i, eventType := range opts.EventTypes {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, eventType)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("event_type IN (%s)", strings.Join(placeholders, ",")))
	}

	// 状态筛选
	if len(opts.Status) > 0 {
		placeholders := make([]string, len(opts.Status))
		for i, status := range opts.Status {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, status)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("status IN (%s)", strings.Join(placeholders, ",")))
	}

	// 构建排序
	orderBy := "event_date"
	if opts.OrderBy == "effective_date" {
		orderBy = "effective_date"
	}
	
	orderDirection := "DESC"
	if opts.OrderDirection == "asc" {
		orderDirection = "ASC"
	}

	// 构建分页
	limit := 50
	if opts.Limit > 0 && opts.Limit <= 500 {
		limit = opts.Limit
	}

	offset := 0
	if opts.Offset > 0 {
		offset = opts.Offset
	}

	// 构建查询
	query := fmt.Sprintf(`
		SELECT id, organization_code, event_type, event_date, effective_date, status,
		       title, description, metadata, previous_value, new_value, affected_fields,
		       triggered_by, approved_by, created_at, updated_at, tenant_id
		FROM organization_timeline_events
		WHERE %s
		ORDER BY %s %s, id %s
		LIMIT %d OFFSET %d
	`, strings.Join(conditions, " AND "), orderBy, orderDirection, orderDirection, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询时间线事件失败: %w", err)
	}
	defer rows.Close()

	var events []TimelineEvent
	for rows.Next() {
		var event TimelineEvent
		var metadataBytes, previousBytes, newBytes, fieldsBytes []byte

		err := rows.Scan(
			&event.ID, &event.OrganizationCode, &event.EventType,
			&event.EventDate, &event.EffectiveDate, &event.Status,
			&event.Title, &event.Description, &metadataBytes,
			&previousBytes, &newBytes, &fieldsBytes,
			&event.TriggeredBy, &event.ApprovedBy,
			&event.CreatedAt, &event.UpdatedAt, &event.TenantID,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描时间线事件失败: %w", err)
		}

		// 解析JSON字段
		if len(metadataBytes) > 0 {
			json.Unmarshal(metadataBytes, &event.Metadata)
		}
		if len(previousBytes) > 0 {
			json.Unmarshal(previousBytes, &event.PreviousValue)
		}
		if len(newBytes) > 0 {
			json.Unmarshal(newBytes, &event.NewValue)
		}
		if len(fieldsBytes) > 0 {
			json.Unmarshal(fieldsBytes, &event.AffectedFields)
		}

		events = append(events, event)
	}

	return events, nil
}

// 获取时间线统计信息
func (r *TimelineRepository) GetTimelineStats(ctx context.Context, tenantID uuid.UUID, orgCode string) (*TimelineStats, error) {
	stats := &TimelineStats{
		EventsByType:   make(map[string]int),
		EventsByStatus: make(map[string]int),
	}

	// 1. 获取总事件数
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM organization_timeline_events WHERE tenant_id = $1 AND organization_code = $2",
		tenantID.String(), orgCode).Scan(&stats.TotalEvents)
	if err != nil {
		return nil, fmt.Errorf("获取事件总数失败: %w", err)
	}

	// 2. 按类型统计
	typeRows, err := r.db.QueryContext(ctx, `
		SELECT event_type, COUNT(*) 
		FROM organization_timeline_events 
		WHERE tenant_id = $1 AND organization_code = $2 
		GROUP BY event_type
	`, tenantID.String(), orgCode)
	if err != nil {
		return nil, fmt.Errorf("按类型统计失败: %w", err)
	}
	defer typeRows.Close()

	for typeRows.Next() {
		var eventType string
		var count int
		typeRows.Scan(&eventType, &count)
		stats.EventsByType[eventType] = count
	}

	// 3. 按状态统计
	statusRows, err := r.db.QueryContext(ctx, `
		SELECT status, COUNT(*) 
		FROM organization_timeline_events 
		WHERE tenant_id = $1 AND organization_code = $2 
		GROUP BY status
	`, tenantID.String(), orgCode)
	if err != nil {
		return nil, fmt.Errorf("按状态统计失败: %w", err)
	}
	defer statusRows.Close()

	for statusRows.Next() {
		var status string
		var count int
		statusRows.Scan(&status, &count)
		stats.EventsByStatus[status] = count
	}

	// 4. 获取最近事件
	recentEvents, err := r.GetTimeline(ctx, tenantID, orgCode, &TimelineQueryOptions{
		Limit: 10,
		OrderBy: "event_date",
		OrderDirection: "desc",
	})
	if err != nil {
		return nil, fmt.Errorf("获取最近事件失败: %w", err)
	}
	stats.RecentEvents = recentEvents

	// 5. 获取时间线跨度
	var earliest, latest sql.NullTime
	err = r.db.QueryRowContext(ctx, `
		SELECT MIN(event_date), MAX(event_date) 
		FROM organization_timeline_events 
		WHERE tenant_id = $1 AND organization_code = $2
	`, tenantID.String(), orgCode).Scan(&earliest, &latest)
	
	if err == nil && earliest.Valid && latest.Valid {
		stats.TimelineSpan = &TimelineSpan{
			EarliestEvent: earliest.Time,
			LatestEvent:   latest.Time,
			SpanDays:      int(latest.Time.Sub(earliest.Time).Hours() / 24),
		}
	}

	// 6. 获取月度活动统计
	monthlyRows, err := r.db.QueryContext(ctx, `
		SELECT TO_CHAR(event_date, 'YYYY-MM') as month, COUNT(*) as event_count
		FROM organization_timeline_events 
		WHERE tenant_id = $1 AND organization_code = $2
		  AND event_date >= NOW() - INTERVAL '12 months'
		GROUP BY TO_CHAR(event_date, 'YYYY-MM')
		ORDER BY month DESC
	`, tenantID.String(), orgCode)
	
	if err == nil {
		defer monthlyRows.Close()
		for monthlyRows.Next() {
			var month string
			var count int
			monthlyRows.Scan(&month, &count)
			stats.MonthlyActivity = append(stats.MonthlyActivity, MonthlyEventCount{
				Month: month,
				EventCount: count,
			})
		}
	}

	return stats, nil
}

// 获取组织版本历史
func (r *TimelineRepository) GetVersionHistory(ctx context.Context, tenantID uuid.UUID, orgCode string, limit int) ([]OrganizationVersion, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := `
		SELECT id, organization_code, version, effective_from, effective_to,
		       snapshot_data, change_reason, created_at, tenant_id
		FROM organization_unit_versions
		WHERE tenant_id = $1 AND organization_code = $2
		ORDER BY version DESC
		LIMIT $3
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID.String(), orgCode, limit)
	if err != nil {
		return nil, fmt.Errorf("查询版本历史失败: %w", err)
	}
	defer rows.Close()

	var versions []OrganizationVersion
	for rows.Next() {
		var version OrganizationVersion
		var snapshotBytes []byte

		err := rows.Scan(
			&version.ID, &version.OrganizationCode, &version.Version,
			&version.EffectiveFrom, &version.EffectiveTo,
			&snapshotBytes, &version.ChangeReason,
			&version.CreatedAt, &version.TenantID,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描版本历史失败: %w", err)
		}

		// 解析快照数据
		if len(snapshotBytes) > 0 {
			json.Unmarshal(snapshotBytes, &version.SnapshotData)
		}

		versions = append(versions, version)
	}

	return versions, nil
}

// 比较两个版本
func (r *TimelineRepository) CompareVersions(ctx context.Context, tenantID uuid.UUID, orgCode string, fromVersion, toVersion int) (*VersionComparison, error) {
	// 获取两个版本的快照数据
	query := `
		SELECT version, snapshot_data
		FROM organization_unit_versions
		WHERE tenant_id = $1 AND organization_code = $2 AND version IN ($3, $4)
		ORDER BY version
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID.String(), orgCode, fromVersion, toVersion)
	if err != nil {
		return nil, fmt.Errorf("查询版本快照失败: %w", err)
	}
	defer rows.Close()

	var fromSnapshot, toSnapshot map[string]interface{}
	versionsFound := 0

	for rows.Next() {
		var version int
		var snapshotBytes []byte
		rows.Scan(&version, &snapshotBytes)

		var snapshot map[string]interface{}
		if len(snapshotBytes) > 0 {
			json.Unmarshal(snapshotBytes, &snapshot)
		}

		if version == fromVersion {
			fromSnapshot = snapshot
		} else if version == toVersion {
			toSnapshot = snapshot
		}
		versionsFound++
	}

	if versionsFound < 2 {
		return nil, fmt.Errorf("无法找到指定的版本进行比较")
	}

	// 执行版本比较
	comparison := &VersionComparison{
		FromVersion: fromVersion,
		ToVersion:   toVersion,
		ComparedAt:  time.Now(),
	}

	// 找出所有字段
	allFields := make(map[string]bool)
	for field := range fromSnapshot {
		allFields[field] = true
	}
	for field := range toSnapshot {
		allFields[field] = true
	}

	// 比较每个字段
	for field := range allFields {
		oldValue, oldExists := fromSnapshot[field]
		newValue, newExists := toSnapshot[field]

		var change FieldChange
		change.Field = field

		if !oldExists && newExists {
			// 字段被添加
			change.ChangeType = "added"
			change.NewValue = newValue
			comparison.Summary.AddedFields++
		} else if oldExists && !newExists {
			// 字段被移除
			change.ChangeType = "removed"
			change.OldValue = oldValue
			comparison.Summary.RemovedFields++
		} else if oldExists && newExists {
			// 检查字段是否被修改
			oldJSON, _ := json.Marshal(oldValue)
			newJSON, _ := json.Marshal(newValue)
			if string(oldJSON) != string(newJSON) {
				change.ChangeType = "modified"
				change.OldValue = oldValue
				change.NewValue = newValue
				comparison.Summary.ModifiedFields++
			} else {
				// 字段未改变，跳过
				continue
			}
		}

		comparison.FieldChanges = append(comparison.FieldChanges, change)
		comparison.Summary.TotalChanges++
	}

	return comparison, nil
}

// ===== HTTP处理器 =====

type TimelineHandler struct {
	repo *TimelineRepository
}

func NewTimelineHandler(db *sql.DB) *TimelineHandler {
	return &TimelineHandler{
		repo: NewTimelineRepository(db),
	}
}

// Prometheus指标
var (
	timelineRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "timeline_requests_total",
			Help: "Total number of timeline requests",
		},
		[]string{"operation", "status"},
	)
	timelineRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "timeline_request_duration_seconds",
			Help: "Timeline request duration in seconds",
		},
		[]string{"operation"},
	)
)

func init() {
	prometheus.MustRegister(timelineRequestsTotal)
	prometheus.MustRegister(timelineRequestDuration)
}

func (h *TimelineHandler) getTenantID(r *http.Request) uuid.UUID {
	tenantHeader := r.Header.Get("X-Tenant-ID")
	if tenantHeader != "" {
		if tenantID, err := uuid.Parse(tenantHeader); err == nil {
			return tenantID
		}
	}
	// 返回默认租户ID
	return uuid.MustParse("3b99930c-4dc6-4cc9-8e4d-7d960a931cb9")
}

func (h *TimelineHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, errorCode, message string, details error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := map[string]interface{}{
		"error_code": errorCode,
		"message":    message,
		"timestamp":  time.Now().Format(time.RFC3339),
	}
	
	if details != nil {
		response["details"] = details.Error()
	}
	
	json.NewEncoder(w).Encode(response)
}

// 解析时间线查询参数
func (h *TimelineHandler) parseQueryOptions(r *http.Request) *TimelineQueryOptions {
	opts := &TimelineQueryOptions{}

	if startStr := r.URL.Query().Get("start_date"); startStr != "" {
		if start, err := time.Parse("2006-01-02", startStr); err == nil {
			opts.StartDate = &start
		}
	}

	if endStr := r.URL.Query().Get("end_date"); endStr != "" {
		if end, err := time.Parse("2006-01-02", endStr); err == nil {
			opts.EndDate = &end
		}
	}

	if eventTypesStr := r.URL.Query().Get("event_types"); eventTypesStr != "" {
		opts.EventTypes = strings.Split(eventTypesStr, ",")
	}

	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		opts.Status = strings.Split(statusStr, ",")
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			opts.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			opts.Offset = offset
		}
	}

	opts.OrderBy = r.URL.Query().Get("order_by")
	opts.OrderDirection = r.URL.Query().Get("order_direction")
	opts.IncludeMetadata = r.URL.Query().Get("include_metadata") == "true"

	return opts
}

// 获取组织时间线
func (h *TimelineHandler) GetOrganizationTimeline(w http.ResponseWriter, r *http.Request) {
	timer := prometheus.NewTimer(timelineRequestDuration.WithLabelValues("get_timeline"))
	defer timer.ObserveDuration()

	orgCode := chi.URLParam(r, "code")
	if orgCode == "" {
		timelineRequestsTotal.WithLabelValues("get_timeline", "failed").Inc()
		h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_CODE", "缺少组织代码", nil)
		return
	}

	tenantID := h.getTenantID(r)
	opts := h.parseQueryOptions(r)

	timeline, err := h.repo.GetTimeline(r.Context(), tenantID, orgCode, opts)
	if err != nil {
		timelineRequestsTotal.WithLabelValues("get_timeline", "failed").Inc()
		h.writeErrorResponse(w, http.StatusInternalServerError, "TIMELINE_QUERY_ERROR", "获取时间线失败", err)
		return
	}

	response := map[string]interface{}{
		"organization_code": orgCode,
		"timeline":          timeline,
		"query_options":     opts,
		"result_count":      len(timeline),
		"queried_at":        time.Now().Format(time.RFC3339),
	}

	timelineRequestsTotal.WithLabelValues("get_timeline", "success").Inc()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 获取时间线统计
func (h *TimelineHandler) GetTimelineStats(w http.ResponseWriter, r *http.Request) {
	timer := prometheus.NewTimer(timelineRequestDuration.WithLabelValues("get_stats"))
	defer timer.ObserveDuration()

	orgCode := chi.URLParam(r, "code")
	if orgCode == "" {
		timelineRequestsTotal.WithLabelValues("get_stats", "failed").Inc()
		h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_CODE", "缺少组织代码", nil)
		return
	}

	tenantID := h.getTenantID(r)

	stats, err := h.repo.GetTimelineStats(r.Context(), tenantID, orgCode)
	if err != nil {
		timelineRequestsTotal.WithLabelValues("get_stats", "failed").Inc()
		h.writeErrorResponse(w, http.StatusInternalServerError, "STATS_QUERY_ERROR", "获取统计信息失败", err)
		return
	}

	timelineRequestsTotal.WithLabelValues("get_stats", "success").Inc()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// 获取版本历史
func (h *TimelineHandler) GetVersionHistory(w http.ResponseWriter, r *http.Request) {
	timer := prometheus.NewTimer(timelineRequestDuration.WithLabelValues("get_versions"))
	defer timer.ObserveDuration()

	orgCode := chi.URLParam(r, "code")
	if orgCode == "" {
		timelineRequestsTotal.WithLabelValues("get_versions", "failed").Inc()
		h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_CODE", "缺少组织代码", nil)
		return
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	tenantID := h.getTenantID(r)

	versions, err := h.repo.GetVersionHistory(r.Context(), tenantID, orgCode, limit)
	if err != nil {
		timelineRequestsTotal.WithLabelValues("get_versions", "failed").Inc()
		h.writeErrorResponse(w, http.StatusInternalServerError, "VERSION_QUERY_ERROR", "获取版本历史失败", err)
		return
	}

	response := map[string]interface{}{
		"organization_code": orgCode,
		"versions":          versions,
		"version_count":     len(versions),
		"queried_at":        time.Now().Format(time.RFC3339),
	}

	timelineRequestsTotal.WithLabelValues("get_versions", "success").Inc()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 版本比较
func (h *TimelineHandler) CompareVersions(w http.ResponseWriter, r *http.Request) {
	timer := prometheus.NewTimer(timelineRequestDuration.WithLabelValues("compare_versions"))
	defer timer.ObserveDuration()

	orgCode := chi.URLParam(r, "code")
	if orgCode == "" {
		timelineRequestsTotal.WithLabelValues("compare_versions", "failed").Inc()
		h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_CODE", "缺少组织代码", nil)
		return
	}

	fromVersionStr := r.URL.Query().Get("from_version")
	toVersionStr := r.URL.Query().Get("to_version")

	if fromVersionStr == "" || toVersionStr == "" {
		timelineRequestsTotal.WithLabelValues("compare_versions", "failed").Inc()
		h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_VERSIONS", "缺少版本参数", nil)
		return
	}

	fromVersion, err := strconv.Atoi(fromVersionStr)
	if err != nil {
		timelineRequestsTotal.WithLabelValues("compare_versions", "failed").Inc()
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_FROM_VERSION", "无效的起始版本", err)
		return
	}

	toVersion, err := strconv.Atoi(toVersionStr)
	if err != nil {
		timelineRequestsTotal.WithLabelValues("compare_versions", "failed").Inc()
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_TO_VERSION", "无效的目标版本", err)
		return
	}

	tenantID := h.getTenantID(r)

	comparison, err := h.repo.CompareVersions(r.Context(), tenantID, orgCode, fromVersion, toVersion)
	if err != nil {
		timelineRequestsTotal.WithLabelValues("compare_versions", "failed").Inc()
		h.writeErrorResponse(w, http.StatusInternalServerError, "COMPARISON_ERROR", "版本比较失败", err)
		return
	}

	timelineRequestsTotal.WithLabelValues("compare_versions", "success").Inc()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comparison)
}

// ===== 主程序 =====

func main() {
	// 数据库连接
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/cubecastle?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("数据库连接测试失败:", err)
	}

	log.Println("✅ 数据库连接成功")

	// 创建处理器
	handler := NewTimelineHandler(db)

	// 设置路由
	r := chi.NewRouter()

	// 中间件
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Tenant-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// 健康检查
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"service":   "organization-timeline-service",
			"version":   "1.0.0",
			"timestamp": time.Now().Format(time.RFC3339),
			"features": []string{
				"timeline-events", "version-history", "version-comparison",
				"timeline-stats", "event-filtering", "prometheus-metrics",
			},
		})
	})

	// 监控指标
	r.Handle("/metrics", promhttp.Handler())

	// API路由
	r.Route("/api/v1/organization-units/{code}", func(r chi.Router) {
		r.Get("/timeline", handler.GetOrganizationTimeline)          // 获取时间线
		r.Get("/timeline/stats", handler.GetTimelineStats)          // 时间线统计
		r.Get("/versions", handler.GetVersionHistory)               // 版本历史
		r.Get("/versions/compare", handler.CompareVersions)         // 版本比较
	})

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "9092" // 使用9092端口避免冲突
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// 优雅关闭
	go func() {
		log.Printf("🚀 组织时间线管理服务启动在端口 %s", port)
		log.Println("📋 支持的功能:")
		log.Println("  - 时间线事件查询 (日期范围、事件类型、状态筛选)")
		log.Println("  - 时间线统计分析 (事件分布、月度活动、时间跨度)")
		log.Println("  - 版本历史管理 (快照查询、历史追踪)")
		log.Println("  - 版本对比分析 (字段变更、差异检测)")
		log.Println("  - Prometheus监控指标")
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("服务器启动失败:", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭时间线服务...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("服务器强制关闭:", err)
	}

	log.Println("时间线服务已关闭")
}