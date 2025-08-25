package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// 缓存事件总线 - 用于处理CDC事件和缓存更新
type CacheEventBus struct {
	subscribers []chan CDCEvent
	mu          sync.RWMutex
	closed      bool
}

// CDC事件定义
type CDCEvent struct {
	EventID    string                 `json:"event_id"`
	Operation  string                 `json:"operation"`   // CREATE, UPDATE, DELETE
	EntityType string                 `json:"entity_type"` // organization
	EntityID   string                 `json:"entity_id"`   // 组织代码
	TenantID   string                 `json:"tenant_id"`
	Before     map[string]interface{} `json:"before,omitempty"`
	After      map[string]interface{} `json:"after,omitempty"`
	Timestamp  int64                  `json:"timestamp"`
	Source     string                 `json:"source"` // debezium, domain_event
}

// 组织模型
type Organization struct {
	Code        string    `json:"code"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	UnitType    string    `json:"unit_type"`
	Status      string    `json:"status"`
	Level       int       `json:"level"`
	Path        string    `json:"path"`
	SortOrder   int       `json:"sort_order"`
	Description string    `json:"description"`
	ParentCode  string    `json:"parent_code"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 组织统计模型
type OrganizationStats struct {
	TotalCount int           `json:"total_count"`
	ByType     []TypeCount   `json:"by_type"`
	ByStatus   []StatusCount `json:"by_status"`
	ByLevel    []LevelCount  `json:"by_level"`
}

type TypeCount struct {
	UnitType string `json:"unit_type"`
	Count    int    `json:"count"`
}

type StatusCount struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

type LevelCount struct {
	Level string `json:"level"`
	Count int    `json:"count"`
}

// 缓存统计
type CacheStats struct {
	L1Stats         L1Stats `json:"l1_stats"`
	L2Connected     bool    `json:"l2_connected"`
	WriteThrough    bool    `json:"write_through"`
	ConsistencyMode string  `json:"consistency_mode"`
}

type L1Stats struct {
	HitCount  int64   `json:"hit_count"`
	MissCount int64   `json:"miss_count"`
	Size      int     `json:"size"`
	HitRate   float64 `json:"hit_rate"`
}

// 创建事件总线
func NewCacheEventBus() *CacheEventBus {
	return &CacheEventBus{
		subscribers: make([]chan CDCEvent, 0),
	}
}

// 订阅事件
func (bus *CacheEventBus) Subscribe() <-chan CDCEvent {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	if bus.closed {
		ch := make(chan CDCEvent)
		close(ch)
		return ch
	}

	ch := make(chan CDCEvent, 100) // 带缓冲的通道
	bus.subscribers = append(bus.subscribers, ch)
	return ch
}

// 发布事件
func (bus *CacheEventBus) Publish(event CDCEvent) {
	bus.mu.RLock()
	defer bus.mu.RUnlock()

	if bus.closed {
		return
	}

	for _, ch := range bus.subscribers {
		select {
		case ch <- event:
		default:
			// 如果通道满了，跳过该订阅者，避免阻塞
			log.Printf("事件总线订阅者通道满，跳过事件: %s", event.EventID)
		}
	}
}

// 关闭事件总线
func (bus *CacheEventBus) Close() {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	if bus.closed {
		return
	}

	bus.closed = true
	for _, ch := range bus.subscribers {
		close(ch)
	}
	bus.subscribers = nil
}

// 将CDC事件转换为组织对象
func (event CDCEvent) ToOrganization() Organization {
	var org Organization

	// 根据操作类型选择数据源
	var data map[string]interface{}
	if event.After != nil {
		data = event.After
	} else if event.Before != nil {
		data = event.Before
	}

	if data == nil {
		return org
	}

	// 安全地提取字段
	org.Code = getStringFromMap(data, "code")
	org.TenantID = event.TenantID
	org.Name = getStringFromMap(data, "name")
	org.UnitType = getStringFromMap(data, "unit_type")
	org.Status = getStringFromMap(data, "status")
	org.Level = getIntFromMap(data, "level")
	org.Path = getStringFromMap(data, "path")
	org.SortOrder = getIntFromMap(data, "sort_order")
	org.Description = getStringFromMap(data, "description")
	org.ParentCode = getStringFromMap(data, "parent_code")

	// 时间字段处理
	if createdAt := getStringFromMap(data, "created_at"); createdAt != "" {
		if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
			org.CreatedAt = t
		}
	}

	if updatedAt := getStringFromMap(data, "updated_at"); updatedAt != "" {
		if t, err := time.Parse(time.RFC3339, updatedAt); err == nil {
			org.UpdatedAt = t
		}
	}

	return org
}

// 辅助函数：从map中安全提取字符串
func getStringFromMap(data map[string]interface{}, key string) string {
	if value, ok := data[key]; ok && value != nil {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// 辅助函数：从map中安全提取整数
func getIntFromMap(data map[string]interface{}, key string) int {
	if value, ok := data[key]; ok && value != nil {
		switch v := value.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return 0
}

// 智能缓存更新器 - 负责高效更新列表缓存
type SmartCacheUpdater struct {
	logger *log.Logger
}

func NewSmartCacheUpdater(logger *log.Logger) *SmartCacheUpdater {
	return &SmartCacheUpdater{
		logger: logger,
	}
}

// 高效列表更新算法
func (updater *SmartCacheUpdater) UpdateListCache(
	existingList []Organization,
	updatedOrg *Organization,
	operation string,
	queryParams QueryParams,
) ([]Organization, bool) {

	switch operation {
	case "CREATE":
		return updater.handleCreate(existingList, updatedOrg, queryParams)
	case "UPDATE":
		return updater.handleUpdate(existingList, updatedOrg, queryParams)
	case "DELETE":
		return updater.handleDelete(existingList, updatedOrg, queryParams)
	}

	return existingList, false
}

// 处理创建操作
func (updater *SmartCacheUpdater) handleCreate(
	existingList []Organization,
	newOrg *Organization,
	queryParams QueryParams,
) ([]Organization, bool) {

	// 检查新组织是否符合查询条件
	if !updater.matchesQueryParams(newOrg, queryParams) {
		return existingList, false
	}

	// 添加到列表并排序
	updatedList := make([]Organization, len(existingList)+1)
	copy(updatedList, existingList)
	updatedList[len(existingList)] = *newOrg

	updater.sortOrganizations(updatedList)
	return updatedList, true
}

// 处理更新操作
func (updater *SmartCacheUpdater) handleUpdate(
	existingList []Organization,
	updatedOrg *Organization,
	queryParams QueryParams,
) ([]Organization, bool) {

	updated := false
	updatedList := make([]Organization, 0, len(existingList))

	for _, org := range existingList {
		if org.Code == updatedOrg.Code {
			// 检查更新后的组织是否仍符合查询条件
			if updater.matchesQueryParams(updatedOrg, queryParams) {
				updatedList = append(updatedList, *updatedOrg)
			}
			updated = true
		} else {
			updatedList = append(updatedList, org)
		}
	}

	// 如果原来不在列表中，但现在符合条件，则添加
	if !updated && updater.matchesQueryParams(updatedOrg, queryParams) {
		updatedList = append(updatedList, *updatedOrg)
		updated = true
	}

	if updated {
		updater.sortOrganizations(updatedList)
	}

	return updatedList, updated
}

// 处理删除操作
func (updater *SmartCacheUpdater) handleDelete(
	existingList []Organization,
	deletedOrg *Organization,
	queryParams QueryParams,
) ([]Organization, bool) {

	updatedList := make([]Organization, 0, len(existingList))
	found := false

	for _, org := range existingList {
		if org.Code != deletedOrg.Code {
			updatedList = append(updatedList, org)
		} else {
			found = true
		}
	}

	return updatedList, found
}

// 检查组织是否匹配查询参数
func (updater *SmartCacheUpdater) matchesQueryParams(org *Organization, params QueryParams) bool {
	if params.SearchText == "" {
		return true
	}

	searchText := params.SearchText
	return contains(org.Name, searchText) || contains(org.Code, searchText)
}

// 组织排序
func (updater *SmartCacheUpdater) sortOrganizations(orgs []Organization) {
	// 实现排序逻辑：先按sort_order，再按code
	// 这里使用简化的冒泡排序，生产环境建议使用sort包
	for i := 0; i < len(orgs)-1; i++ {
		for j := 0; j < len(orgs)-i-1; j++ {
			if shouldSwap(orgs[j], orgs[j+1]) {
				orgs[j], orgs[j+1] = orgs[j+1], orgs[j]
			}
		}
	}
}

// 排序比较函数
func shouldSwap(a, b Organization) bool {
	if a.SortOrder != b.SortOrder {
		return a.SortOrder > b.SortOrder
	}
	return a.Code > b.Code
}

// 字符串包含检查（忽略大小写）
func contains(str, substr string) bool {
	return len(str) >= len(substr) &&
		findInString(str, substr) >= 0
}

// 简化的字符串查找
func findInString(str, substr string) int {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// 缓存一致性检查器
type ConsistencyChecker struct {
	l1Cache *L1Cache
	l2Cache interface{} // Redis客户端接口
	logger  *log.Logger
}

func NewConsistencyChecker(l1Cache *L1Cache, l2Cache interface{}, logger *log.Logger) *ConsistencyChecker {
	return &ConsistencyChecker{
		l1Cache: l1Cache,
		l2Cache: l2Cache,
		logger:  logger,
	}
}

// 执行一致性检查
func (checker *ConsistencyChecker) CheckConsistency(ctx context.Context, sampleKeys []string) ConsistencyReport {
	report := ConsistencyReport{
		TotalChecked:    len(sampleKeys),
		Inconsistencies: []InconsistencyRecord{},
		CheckedAt:       time.Now(),
	}

	for _, key := range sampleKeys {
		if inconsistency := checker.checkSingleKey(ctx, key); inconsistency != nil {
			report.Inconsistencies = append(report.Inconsistencies, *inconsistency)
		}
	}

	return report
}

// 检查单个键的一致性
func (checker *ConsistencyChecker) checkSingleKey(ctx context.Context, key string) *InconsistencyRecord {
	// L1缓存获取
	l1Entry, l1Exists := checker.l1Cache.Get(key)

	// L2缓存获取（这里需要根据实际Redis客户端实现）
	// 简化实现，实际需要根据具体的Redis客户端调用
	l2Data, l2Err := checker.getFromL2(ctx, key)
	l2Exists := l2Err == nil && l2Data != ""

	// 比较结果
	if l1Exists != l2Exists {
		return &InconsistencyRecord{
			Key:        key,
			Issue:      "存在性不一致",
			L1Exists:   l1Exists,
			L2Exists:   l2Exists,
			DetectedAt: time.Now(),
		}
	}

	if l1Exists && l2Exists {
		// 比较数据内容
		l1Hash := checker.hashContent(l1Entry.Data)
		l2Hash := checker.hashContent(l2Data)

		if l1Hash != l2Hash {
			return &InconsistencyRecord{
				Key:        key,
				Issue:      "数据内容不一致",
				L1Hash:     l1Hash,
				L2Hash:     l2Hash,
				DetectedAt: time.Now(),
			}
		}
	}

	return nil
}

// 从L2缓存获取数据（需要根据实际实现）
func (checker *ConsistencyChecker) getFromL2(ctx context.Context, key string) (string, error) {
	// 这里需要根据实际的Redis客户端实现
	return "", fmt.Errorf("需要实现L2缓存获取逻辑")
}

// 计算内容哈希
func (checker *ConsistencyChecker) hashContent(data interface{}) string {
	if jsonData, err := json.Marshal(data); err == nil {
		return fmt.Sprintf("%x", jsonData)[:8] // 简化的哈希
	}
	return "error"
}

// 一致性检查报告
type ConsistencyReport struct {
	TotalChecked    int                   `json:"total_checked"`
	Inconsistencies []InconsistencyRecord `json:"inconsistencies"`
	CheckedAt       time.Time             `json:"checked_at"`
}

// 不一致记录
type InconsistencyRecord struct {
	Key        string    `json:"key"`
	Issue      string    `json:"issue"`
	L1Exists   bool      `json:"l1_exists"`
	L2Exists   bool      `json:"l2_exists"`
	L1Hash     string    `json:"l1_hash,omitempty"`
	L2Hash     string    `json:"l2_hash,omitempty"`
	DetectedAt time.Time `json:"detected_at"`
}
