package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// 时态组织结构
type TemporalOrganization struct {
	TenantID      string     `json:"tenant_id"`
	Code          string     `json:"code"`
	ParentCode    *string    `json:"parent_code,omitempty"`
	Name          string     `json:"name"`
	UnitType      string     `json:"unit_type"`
	Status        string     `json:"status"`
	Level         int        `json:"level"`
	Path          string     `json:"path"`
	SortOrder     int        `json:"sort_order"`
	Description   string     `json:"description"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	EffectiveDate *time.Time `json:"effective_date,omitempty"`
	EndDate       *time.Time `json:"end_date,omitempty"`
	ChangeReason  *string    `json:"change_reason,omitempty"`
	IsCurrent     *bool      `json:"is_current,omitempty"`
}

// 时态查询选项
type TemporalQueryOptions struct {
	AsOfDate         *time.Time `json:"as_of_date,omitempty"`
	EffectiveDate    *time.Time `json:"effective_date,omitempty"`
	EndDate          *time.Time `json:"end_date,omitempty"`
	IncludeHistory   bool       `json:"include_history,omitempty"`
	IncludeFuture    bool       `json:"include_future,omitempty"`
	IncludeDissolved bool       `json:"include_dissolved,omitempty"`
	MaxRecords       int        `json:"max_records,omitempty"`
}

// 组织变更事件
type OrganizationChangeEvent struct {
	EventType     string                 `json:"event_type"`
	EffectiveDate time.Time              `json:"effective_date"`
	EndDate       *time.Time             `json:"end_date,omitempty"`
	ChangeData    map[string]interface{} `json:"change_data"`
	ChangeReason  string                 `json:"change_reason"`
}

// 模拟时态数据存储
var temporalData = make(map[string][]*TemporalOrganization)

// 初始化测试数据
func initTestData() {
	tenantID := "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
	now := time.Now()
	
	// 测试组织: 1000056 - 有多个时态版本
	effective1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	effective2 := time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC)
	effective3 := time.Date(2025, 8, 10, 0, 0, 0, 0, time.UTC)
	end1 := time.Date(2025, 7, 31, 23, 59, 59, 0, time.UTC)
	end2 := time.Date(2025, 8, 9, 23, 59, 59, 0, time.UTC)
	
	reason1 := "初始创建"
	reason2 := "部门重组"
	reason3 := "缓存同步修复"
	
	isCurrentFalse := false
	isCurrentTrue := true
	
	temporalData["1000056"] = []*TemporalOrganization{
		// 历史版本1 (2025-01-01 到 2025-07-31)
		{
			TenantID: tenantID, Code: "1000056", Name: "技术部", UnitType: "DEPARTMENT",
			Status: "ACTIVE", Level: 1, Path: "/1000056", SortOrder: 1,
			Description: "负责技术研发", CreatedAt: now, UpdatedAt: now,
			EffectiveDate: &effective1, EndDate: &end1,
			ChangeReason: &reason1, IsCurrent: &isCurrentFalse,
		},
		// 历史版本2 (2025-08-01 到 2025-08-09) 
		{
			TenantID: tenantID, Code: "1000056", Name: "技术研发部", UnitType: "DEPARTMENT",
			Status: "ACTIVE", Level: 1, Path: "/1000056", SortOrder: 1,
			Description: "技术研发和创新", CreatedAt: now, UpdatedAt: now,
			EffectiveDate: &effective2, EndDate: &end2,
			ChangeReason: &reason2, IsCurrent: &isCurrentFalse,
		},
		// 当前版本 (2025-08-10 至今)
		{
			TenantID: tenantID, Code: "1000056", Name: "测试更新缓存_同步修复", UnitType: "DEPARTMENT",
			Status: "ACTIVE", Level: 1, Path: "/1000056", SortOrder: 1,
			Description: "测试时态管理功能", CreatedAt: now, UpdatedAt: now,
			EffectiveDate: &effective3, EndDate: nil,
			ChangeReason: &reason3, IsCurrent: &isCurrentTrue,
		},
	}
	
	// 测试组织: 1000057 - 单一当前版本
	effective4 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	reason4 := "部门设立"
	
	temporalData["1000057"] = []*TemporalOrganization{
		{
			TenantID: tenantID, Code: "1000057", Name: "人力资源部", UnitType: "DEPARTMENT",
			Status: "ACTIVE", Level: 2, Path: "/1000056/1000057", SortOrder: 1,
			Description: "人力资源管理", CreatedAt: now, UpdatedAt: now,
			EffectiveDate: &effective4, EndDate: nil,
			ChangeReason: &reason4, IsCurrent: &isCurrentTrue,
		},
	}
	
	// 测试组织: 1000059 - 计划中的组织（未来生效）
	effective5 := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
	reason5 := "新项目筹备"
	
	temporalData["1000059"] = []*TemporalOrganization{
		{
			TenantID: tenantID, Code: "1000059", Name: "计划项目组", UnitType: "PROJECT_TEAM",
			Status: "PLANNED", Level: 3, Path: "/1000056/1000057/1000059", SortOrder: 1,
			Description: "计划中的项目团队", CreatedAt: now, UpdatedAt: now,
			EffectiveDate: &effective5, EndDate: nil,
			ChangeReason: &reason5, IsCurrent: &isCurrentTrue,
		},
	}
}

// 时态查询实现
func queryTemporal(code string, opts *TemporalQueryOptions) []*TemporalOrganization {
	records, exists := temporalData[code]
	if !exists {
		return nil
	}
	
	var result []*TemporalOrganization
	for _, record := range records {
		shouldInclude := true
		
		// 时间点查询：查询在指定日期有效的记录
		if opts.AsOfDate != nil {
			// 记录必须在指定日期之前或当天开始生效
			if record.EffectiveDate != nil && record.EffectiveDate.After(*opts.AsOfDate) {
				shouldInclude = false
			}
			// 记录必须在指定日期之后或当天结束，或者没有结束日期
			if record.EndDate != nil && !record.EndDate.After(*opts.AsOfDate) {
				shouldInclude = false
			}
		}
		
		// 日期范围查询
		if opts.EffectiveDate != nil && record.EffectiveDate != nil {
			if record.EffectiveDate.Before(*opts.EffectiveDate) {
				shouldInclude = false
			}
		}
		if opts.EndDate != nil && record.EndDate != nil {
			if record.EndDate.After(*opts.EndDate) {
				shouldInclude = false
			}
		}
		
		// 历史记录过滤 - 如果没有特殊指定，默认包含历史记录
		if !opts.IncludeHistory && opts.AsOfDate == nil && record.IsCurrent != nil && !*record.IsCurrent {
			shouldInclude = false
		}
		
		// 未来记录过滤
		if !opts.IncludeFuture && record.EffectiveDate != nil {
			if record.EffectiveDate.After(time.Now()) {
				shouldInclude = false
			}
		}
		
		if shouldInclude {
			result = append(result, record)
		}
	}
	
	return result
}

// 时态事件处理
func processTemporalEvent(code string, event *OrganizationChangeEvent) error {
	records, exists := temporalData[code]
	if !exists {
		return fmt.Errorf("组织 %s 不存在", code)
	}
	
	// 获取当前记录
	var currentRecord *TemporalOrganization
	for _, record := range records {
		if record.IsCurrent != nil && *record.IsCurrent {
			currentRecord = record
			break
		}
	}
	
	if currentRecord == nil {
		return fmt.Errorf("未找到组织 %s 的当前记录", code)
	}
	
	switch event.EventType {
	case "UPDATE":
		return processUpdateEvent(code, currentRecord, event)
	case "RESTRUCTURE":
		return processRestructureEvent(code, currentRecord, event)
	case "DISSOLVE":
		return processDissolveEvent(code, currentRecord, event)
	default:
		return fmt.Errorf("不支持的事件类型: %s", event.EventType)
	}
}

func processUpdateEvent(code string, currentRecord *TemporalOrganization, event *OrganizationChangeEvent) error {
	// 设置当前记录结束日期
	endDate := event.EffectiveDate.Add(-time.Second)
	currentRecord.EndDate = &endDate
	isCurrentFalse := false
	currentRecord.IsCurrent = &isCurrentFalse
	
	// 创建新记录
	newRecord := *currentRecord
	newRecord.EffectiveDate = &event.EffectiveDate
	newRecord.EndDate = event.EndDate
	reason := event.ChangeReason
	newRecord.ChangeReason = &reason
	isCurrentTrue := true
	newRecord.IsCurrent = &isCurrentTrue
	newRecord.UpdatedAt = time.Now()
	
	// 应用变更
	for field, value := range event.ChangeData {
		switch field {
		case "name":
			if name, ok := value.(string); ok {
				newRecord.Name = name
			}
		case "unit_type":
			if unitType, ok := value.(string); ok {
				newRecord.UnitType = unitType
			}
		case "status":
			if status, ok := value.(string); ok {
				newRecord.Status = status
			}
		case "description":
			if desc, ok := value.(string); ok {
				newRecord.Description = desc
			}
		}
	}
	
	// 添加新记录
	temporalData[code] = append(temporalData[code], &newRecord)
	
	log.Printf("✅ UPDATE事件处理完成: %s -> %s (生效日期: %s)", 
		code, newRecord.Name, event.EffectiveDate.Format("2006-01-02"))
	
	return nil
}

func processRestructureEvent(code string, currentRecord *TemporalOrganization, event *OrganizationChangeEvent) error {
	// 类似UPDATE，但重点处理层级结构变更
	return processUpdateEvent(code, currentRecord, event)
}

func processDissolveEvent(code string, currentRecord *TemporalOrganization, event *OrganizationChangeEvent) error {
	// 设置结束日期和状态
	endDate := event.EffectiveDate
	if event.EndDate != nil {
		endDate = *event.EndDate
	}
	
	currentRecord.EndDate = &endDate
	currentRecord.Status = "INACTIVE"
	isCurrentFalse := false
	currentRecord.IsCurrent = &isCurrentFalse
	
	log.Printf("✅ DISSOLVE事件处理完成: %s 已解散 (结束日期: %s)", 
		code, endDate.Format("2006-01-02"))
	
	return nil
}

// HTTP处理器
func temporalQueryHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/organization-units/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "temporal" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	
	code := parts[0]
	
	// 解析查询参数
	opts := &TemporalQueryOptions{}
	if asOfStr := r.URL.Query().Get("as_of_date"); asOfStr != "" {
		if asOfDate, err := time.Parse("2006-01-02", asOfStr); err == nil {
			opts.AsOfDate = &asOfDate
		}
	}
	
	opts.IncludeHistory = r.URL.Query().Get("include_history") == "true"
	opts.IncludeFuture = r.URL.Query().Get("include_future") == "true"
	
	// 查询数据
	results := queryTemporal(code, opts)
	if results == nil {
		http.Error(w, "Organization not found", http.StatusNotFound)
		return
	}
	
	response := map[string]interface{}{
		"organizations": results,
		"query_options": opts,
		"result_count":  len(results),
		"queried_at":    time.Now().Format(time.RFC3339),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func temporalEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/organization-units/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "events" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	
	code := parts[0]
	
	var event OrganizationChangeEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if err := processTemporalEvent(code, &event); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"event_type":     event.EventType,
		"organization":   code,
		"effective_date": event.EffectiveDate,
		"status":         "processed",
		"processed_at":   time.Now().Format(time.RFC3339),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"service":   "temporal-function-test",
		"timestamp": time.Now().Format(time.RFC3339),
		"data_count": len(temporalData),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	initTestData()
	
	http.HandleFunc("/api/v1/organization-units/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/temporal") {
			temporalQueryHandler(w, r)
		} else if strings.HasSuffix(r.URL.Path, "/events") {
			temporalEventHandler(w, r)
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})
	
	http.HandleFunc("/health", healthHandler)
	
	log.Println("🚀 时态功能测试服务启动在端口 9091")
	log.Println("📋 测试数据已初始化:")
	log.Println("  - 1000056: 3个时态版本 (2025-01-01, 2025-08-01, 2025-08-10)")
	log.Println("  - 1000057: 1个当前版本 (2025-01-01)")
	log.Println("  - 1000059: 1个未来版本 (2025-09-01)")
	log.Println("")
	log.Println("🧪 测试端点:")
	log.Println("  - GET /health")
	log.Println("  - GET /api/v1/organization-units/{code}/temporal")
	log.Println("  - POST /api/v1/organization-units/{code}/events")
	
	if err := http.ListenAndServe(":9091", nil); err != nil {
		log.Fatal("服务启动失败:", err)
	}
}