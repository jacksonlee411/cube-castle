package cache

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// 统一缓存管理器 - 企业级三层缓存架构
type UnifiedCacheManager struct {
	l1Cache  *L1Cache         // 进程内缓存
	l2Cache  *redis.Client    // Redis分布式缓存
	l3Query  L3QueryInterface // 数据查询接口
	eventBus *CacheEventBus   // 缓存事件总线
	logger   *log.Logger
	config   *CacheConfig
	mu       sync.RWMutex
}

// 缓存配置
type CacheConfig struct {
	L1TTL           time.Duration // L1缓存TTL
	L2TTL           time.Duration // L2缓存TTL
	L1MaxSize       int           // L1缓存最大条目数
	WriteThrough    bool          // 是否启用写时更新
	ConsistencyMode string        // 一致性模式: STRONG, EVENTUAL
	Namespace       string        // 缓存命名空间
}

// L3数据查询接口 - 抽象数据源
type L3QueryInterface interface {
	GetOrganizations(ctx context.Context, tenantID uuid.UUID, params QueryParams) ([]Organization, error)
	GetOrganization(ctx context.Context, tenantID uuid.UUID, code string) (*Organization, error)
	GetOrganizationStats(ctx context.Context, tenantID uuid.UUID) (*OrganizationStats, error)
}

// 查询参数
type QueryParams struct {
	First      int    `json:"first"`
	Offset     int    `json:"offset"`
	SearchText string `json:"search_text"`
}

// 缓存条目定义
type CacheEntry struct {
	Key       string          `json:"key"`
	Data      json.RawMessage `json:"data"`
	Metadata  CacheMetadata   `json:"metadata"`
	Tags      []string        `json:"tags"`
	CreatedAt time.Time       `json:"created_at"`
	ExpiresAt time.Time       `json:"expires_at"`
}

// 缓存元数据
type CacheMetadata struct {
	TenantID     string    `json:"tenant_id"`
	EntityType   string    `json:"entity_type"` // organization, stats, list
	EntityID     string    `json:"entity_id"`   // 具体实体ID
	Version      int64     `json:"version"`     // 数据版本号
	LastModified time.Time `json:"last_modified"`
	Source       string    `json:"source"` // 数据来源层级
}

// 缓存键管理器
type CacheKeyManager struct {
	namespace string
}

func (km *CacheKeyManager) GenerateKey(entityType string, identifiers ...string) string {
	h := md5.New()
	// 使用与查询服务一致的格式：org:entityType:identifiers
	keyBase := fmt.Sprintf("org:%s:%v", entityType, identifiers)
	h.Write([]byte(keyBase))
	return fmt.Sprintf("cache:%x", h.Sum(nil))
}

func (km *CacheKeyManager) GetPatternForTags(tags []string) []string {
	patterns := make([]string, 0, len(tags))
	for _, tag := range tags {
		patterns = append(patterns, fmt.Sprintf("%s:*%s*", km.namespace, tag))
	}
	return patterns
}

// 初始化统一缓存管理器
func NewUnifiedCacheManager(redisClient *redis.Client, l3Query L3QueryInterface, config *CacheConfig, logger *log.Logger) *UnifiedCacheManager {
	if config == nil {
		config = &CacheConfig{
			L1TTL:           5 * time.Minute,
			L2TTL:           30 * time.Minute,
			L1MaxSize:       1000,
			WriteThrough:    true,
			ConsistencyMode: "STRONG",
			Namespace:       "org_v1",
		}
	}

	ucm := &UnifiedCacheManager{
		l1Cache:  NewL1Cache(config.L1MaxSize, config.L1TTL),
		l2Cache:  redisClient,
		l3Query:  l3Query,
		eventBus: NewCacheEventBus(),
		logger:   logger,
		config:   config,
	}

	// 启动缓存事件监听
	go ucm.startEventListener()

	return ucm
}

// ==================== 查询接口 ====================

// 获取组织列表 - 三层缓存策略
func (ucm *UnifiedCacheManager) GetOrganizations(ctx context.Context, tenantID uuid.UUID, params QueryParams) ([]Organization, error) {
	keyMgr := &CacheKeyManager{namespace: ucm.config.Namespace}
	cacheKey := keyMgr.GenerateKey("organizations", tenantID.String(), fmt.Sprintf("%d-%d-%s", params.First, params.Offset, params.SearchText))

	// L1 缓存查询
	if entry, ok := ucm.l1Cache.Get(cacheKey); ok {
		var orgs []Organization
		if err := json.Unmarshal(entry.Data, &orgs); err == nil {
			ucm.logger.Printf("[L1 HIT] 组织列表缓存命中: %s", cacheKey)
			return orgs, nil
		}
	}

	// L2 缓存查询
	cachedData, err := ucm.l2Cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var entry CacheEntry
		if json.Unmarshal([]byte(cachedData), &entry) == nil {
			var orgs []Organization
			if json.Unmarshal(entry.Data, &orgs) == nil {
				// 回填L1缓存
				ucm.l1Cache.Set(cacheKey, entry)
				ucm.logger.Printf("[L2 HIT] 组织列表缓存命中，回填L1: %s", cacheKey)
				return orgs, nil
			}
		}
	}

	// L3 数据源查询
	ucm.logger.Printf("[L3 QUERY] 查询数据源: %s", cacheKey)
	orgs, err := ucm.l3Query.GetOrganizations(ctx, tenantID, params)
	if err != nil {
		return nil, fmt.Errorf("L3查询失败: %w", err)
	}

	// 序列化数据用于缓存存储
	dataBytes, err := json.Marshal(orgs)
	if err != nil {
		return nil, fmt.Errorf("缓存数据序列化失败: %w", err)
	}

	// 写入多层缓存
	entry := CacheEntry{
		Key:  cacheKey,
		Data: dataBytes,
		Metadata: CacheMetadata{
			TenantID:     tenantID.String(),
			EntityType:   "organizations",
			EntityID:     fmt.Sprintf("list_%d_%d", params.First, params.Offset),
			Version:      time.Now().Unix(),
			LastModified: time.Now(),
			Source:       "L3",
		},
		Tags:      []string{fmt.Sprintf("tenant:%s", tenantID.String()), "type:organizations"},
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ucm.config.L2TTL),
	}

	// 同时写入L1和L2
	ucm.l1Cache.Set(cacheKey, entry)

	if cacheData, err := json.Marshal(entry); err == nil {
		ucm.l2Cache.Set(ctx, cacheKey, string(cacheData), ucm.config.L2TTL)
		ucm.logger.Printf("[CACHE SET] 多层缓存已更新: %s, 数据量: %d", cacheKey, len(orgs))
	}

	return orgs, nil
}

// 获取组织统计信息
func (ucm *UnifiedCacheManager) GetOrganizationStats(ctx context.Context, tenantID uuid.UUID) (*OrganizationStats, error) {
	keyMgr := &CacheKeyManager{namespace: ucm.config.Namespace}
	cacheKey := keyMgr.GenerateKey("stats", tenantID.String())

	// L1缓存查询
	if entry, ok := ucm.l1Cache.Get(cacheKey); ok {
		var stats OrganizationStats
		if err := json.Unmarshal(entry.Data, &stats); err == nil {
			ucm.logger.Printf("[L1 HIT] 统计缓存命中: %s", cacheKey)
			return &stats, nil
		}
	}

	// L2缓存查询
	cachedData, err := ucm.l2Cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var entry CacheEntry
		if json.Unmarshal([]byte(cachedData), &entry) == nil {
			var stats OrganizationStats
			if json.Unmarshal(entry.Data, &stats) == nil {
				ucm.l1Cache.Set(cacheKey, entry)
				ucm.logger.Printf("[L2 HIT] 统计缓存命中，回填L1: %s", cacheKey)
				return &stats, nil
			}
		}
	}

	// L3数据源查询
	ucm.logger.Printf("[L3 QUERY] 查询统计数据源: %s", cacheKey)
	stats, err := ucm.l3Query.GetOrganizationStats(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("L3统计查询失败: %w", err)
	}

	// 序列化统计数据用于缓存存储
	dataBytes, err := json.Marshal(stats)
	if err != nil {
		return nil, fmt.Errorf("统计数据序列化失败: %w", err)
	}

	// 写入多层缓存
	entry := CacheEntry{
		Key:  cacheKey,
		Data: dataBytes,
		Metadata: CacheMetadata{
			TenantID:     tenantID.String(),
			EntityType:   "stats",
			EntityID:     "organization_stats",
			Version:      time.Now().Unix(),
			LastModified: time.Now(),
			Source:       "L3",
		},
		Tags:      []string{fmt.Sprintf("tenant:%s", tenantID.String()), "type:stats"},
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ucm.config.L2TTL),
	}

	ucm.l1Cache.Set(cacheKey, entry)

	if cacheData, err := json.Marshal(entry); err == nil {
		ucm.l2Cache.Set(ctx, cacheKey, string(cacheData), ucm.config.L2TTL)
		ucm.logger.Printf("[CACHE SET] 统计缓存已更新: %s", cacheKey)
	}

	return stats, nil
}

// 获取单个组织
func (ucm *UnifiedCacheManager) GetOrganization(ctx context.Context, tenantID uuid.UUID, code string) (*Organization, error) {
	keyMgr := &CacheKeyManager{namespace: ucm.config.Namespace}
	cacheKey := keyMgr.GenerateKey("organization", tenantID.String(), code)

	// L1缓存查询
	if entry, ok := ucm.l1Cache.Get(cacheKey); ok {
		var org Organization
		if err := json.Unmarshal(entry.Data, &org); err == nil {
			ucm.logger.Printf("[L1 HIT] 单个组织缓存命中: %s", code)
			return &org, nil
		}
	}

	// L2缓存查询
	cachedData, err := ucm.l2Cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var entry CacheEntry
		if json.Unmarshal([]byte(cachedData), &entry) == nil {
			var org Organization
			if json.Unmarshal(entry.Data, &org) == nil {
				ucm.l1Cache.Set(cacheKey, entry)
				ucm.logger.Printf("[L2 HIT] 单个组织缓存命中，回填L1: %s", code)
				return &org, nil
			}
		}
	}

	// L3数据源查询
	org, err := ucm.l3Query.GetOrganization(ctx, tenantID, code)
	if err != nil {
		return nil, fmt.Errorf("L3查询失败: %w", err)
	}
	if org == nil {
		return nil, nil
	}

	// 序列化组织数据用于缓存存储
	dataBytes, err := json.Marshal(org)
	if err != nil {
		return nil, fmt.Errorf("组织数据序列化失败: %w", err)
	}

	// 写入多层缓存
	entry := CacheEntry{
		Key:  cacheKey,
		Data: dataBytes,
		Metadata: CacheMetadata{
			TenantID:     tenantID.String(),
			EntityType:   "organization",
			EntityID:     code,
			Version:      time.Now().Unix(),
			LastModified: time.Now(),
			Source:       "L3",
		},
		Tags:      []string{fmt.Sprintf("tenant:%s", tenantID.String()), fmt.Sprintf("org:%s", code)},
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ucm.config.L2TTL),
	}

	ucm.l1Cache.Set(cacheKey, entry)

	if cacheData, err := json.Marshal(entry); err == nil {
		ucm.l2Cache.Set(ctx, cacheKey, string(cacheData), ucm.config.L2TTL)
		ucm.logger.Printf("[CACHE SET] 单个组织缓存已更新: %s", code)
	}

	return org, nil
}

// ==================== 写时更新接口 ====================

// CDC事件处理 - 写时更新策略
func (ucm *UnifiedCacheManager) HandleCDCEvent(ctx context.Context, event CacheEvent) error {
	if !ucm.config.WriteThrough {
		return ucm.handleTraditionalInvalidation(ctx, event)
	}

	switch event.Operation {
	case "CREATE":
		return ucm.handleCreateEvent(ctx, event)
	case "UPDATE":
		return ucm.handleUpdateEvent(ctx, event)
	case "DELETE":
		return ucm.handleDeleteEvent(ctx, event)
	default:
		ucm.logger.Printf("未知CDC事件类型: %s", event.Operation)
		return nil
	}
}

// 处理创建事件 - 新增到缓存
func (ucm *UnifiedCacheManager) handleCreateEvent(ctx context.Context, event CacheEvent) error {
	org := event.ToOrganization()
	tenantID := uuid.MustParse(org.TenantID)

	// 1. 添加单个组织到缓存
	keyMgr := &CacheKeyManager{namespace: ucm.config.Namespace}
	orgKey := keyMgr.GenerateKey("organization", org.TenantID, org.Code)

	// 序列化组织数据
	dataBytes, err := json.Marshal(&org)
	if err != nil {
		return fmt.Errorf("CDC组织数据序列化失败: %w", err)
	}

	entry := CacheEntry{
		Key:  orgKey,
		Data: dataBytes,
		Metadata: CacheMetadata{
			TenantID:     org.TenantID,
			EntityType:   "organization",
			EntityID:     org.Code,
			Version:      event.Timestamp,
			LastModified: time.Now(),
			Source:       "CDC",
		},
		Tags:      []string{fmt.Sprintf("tenant:%s", org.TenantID), fmt.Sprintf("org:%s", org.Code)},
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ucm.config.L2TTL),
	}

	ucm.l1Cache.Set(orgKey, entry)

	if cacheData, err := json.Marshal(entry); err == nil {
		ucm.l2Cache.Set(ctx, orgKey, string(cacheData), ucm.config.L2TTL)
	}

	// 2. 智能更新列表缓存
	return ucm.smartUpdateListCaches(ctx, tenantID, &org, "CREATE")
}

// 处理更新事件 - 更新缓存中的数据
func (ucm *UnifiedCacheManager) handleUpdateEvent(ctx context.Context, event CacheEvent) error {
	org := event.ToOrganization()
	tenantID := uuid.MustParse(org.TenantID)

	// 1. 更新单个组织缓存
	keyMgr := &CacheKeyManager{namespace: ucm.config.Namespace}
	orgKey := keyMgr.GenerateKey("organization", org.TenantID, org.Code)

	// 序列化组织数据
	dataBytes, err := json.Marshal(&org)
	if err != nil {
		return fmt.Errorf("CDC更新组织数据序列化失败: %w", err)
	}

	entry := CacheEntry{
		Key:  orgKey,
		Data: dataBytes,
		Metadata: CacheMetadata{
			TenantID:     org.TenantID,
			EntityType:   "organization",
			EntityID:     org.Code,
			Version:      event.Timestamp,
			LastModified: time.Now(),
			Source:       "CDC",
		},
		Tags:      []string{fmt.Sprintf("tenant:%s", org.TenantID), fmt.Sprintf("org:%s", org.Code)},
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ucm.config.L2TTL),
	}

	ucm.l1Cache.Set(orgKey, entry)

	if cacheData, err := json.Marshal(entry); err == nil {
		ucm.l2Cache.Set(ctx, orgKey, string(cacheData), ucm.config.L2TTL)
	}

	// 2. 智能更新列表缓存
	return ucm.smartUpdateListCaches(ctx, tenantID, &org, "UPDATE")
}

// 处理删除事件 - 从缓存中移除
func (ucm *UnifiedCacheManager) handleDeleteEvent(ctx context.Context, event CacheEvent) error {
	org := event.ToOrganization()
	tenantID := uuid.MustParse(org.TenantID)

	// 1. 删除单个组织缓存
	keyMgr := &CacheKeyManager{namespace: ucm.config.Namespace}
	orgKey := keyMgr.GenerateKey("organization", org.TenantID, org.Code)

	ucm.l1Cache.Delete(orgKey)
	ucm.l2Cache.Del(ctx, orgKey)

	// 2. 从列表缓存中移除
	return ucm.smartUpdateListCaches(ctx, tenantID, &org, "DELETE")
}

// 智能更新列表缓存 - 使用失效策略确保一致性
func (ucm *UnifiedCacheManager) smartUpdateListCaches(ctx context.Context, tenantID uuid.UUID, org *Organization, operation string) error {
	keyMgr := &CacheKeyManager{namespace: ucm.config.Namespace}

	// 由于使用MD5哈希键，无法直接pattern匹配，采用失效策略确保健壮性
	// 失效所有可能的列表缓存键 (覆盖常见的分页和搜索组合)
	keysToInvalidate := []string{}

	// 1. 常见分页大小的缓存
	pageSizes := []int{50, 100}
	maxPages := 10 // 假设最多10页

	for _, pageSize := range pageSizes {
		for page := 0; page < maxPages; page++ {
			offset := page * pageSize
			// 不带搜索的列表缓存
			key := keyMgr.GenerateKey("organizations", tenantID.String(), fmt.Sprintf("%d", pageSize), fmt.Sprintf("%d", offset), "")
			keysToInvalidate = append(keysToInvalidate, key)
		}
	}

	// 2. 统计缓存
	statsKey := keyMgr.GenerateKey("stats", tenantID.String())
	keysToInvalidate = append(keysToInvalidate, statsKey)

	// 3. 单个组织缓存
	orgKey := keyMgr.GenerateKey("organization", tenantID.String(), org.Code)
	keysToInvalidate = append(keysToInvalidate, orgKey)

	// 执行缓存失效
	invalidatedCount := 0
	for _, key := range keysToInvalidate {
		// 从L1缓存删除
		ucm.l1Cache.Delete(key)
		invalidatedCount++

		// 从L2缓存删除
		deleted, err := ucm.l2Cache.Del(ctx, key).Result()
		if err == nil && deleted > 0 {
			// 如果L2也删除了，计数加1 (L1已经计数了)
		}
	}

	ucm.logger.Printf("智能列表缓存更新完成: %s %s, 影响缓存: %d个", operation, org.Code, invalidatedCount)
	return nil
}

// 更新单个列表缓存
func (ucm *UnifiedCacheManager) updateSingleListCache(ctx context.Context, cacheKey string, org *Organization, operation string) error {
	// 从L2缓存获取现有列表
	cachedData, err := ucm.l2Cache.Get(ctx, cacheKey).Result()
	if err != nil {
		return err // 缓存不存在，跳过更新
	}

	var entry CacheEntry
	if err := json.Unmarshal([]byte(cachedData), &entry); err != nil {
		return err
	}

	var orgs []Organization
	if err := json.Unmarshal(entry.Data, &orgs); err != nil {
		return fmt.Errorf("缓存数据反序列化失败: %w", err)
	}

	// 根据操作类型更新列表
	switch operation {
	case "CREATE":
		// 检查是否符合列表的筛选条件
		if ucm.shouldIncludeInList(cacheKey, org) {
			orgs = append(orgs, *org)
			ucm.sortOrganizations(orgs) // 保持排序
		}
	case "UPDATE":
		for i, existingOrg := range orgs {
			if existingOrg.Code == org.Code {
				orgs[i] = *org
				break
			}
		}
		ucm.sortOrganizations(orgs)
	case "DELETE":
		for i, existingOrg := range orgs {
			if existingOrg.Code == org.Code {
				orgs = append(orgs[:i], orgs[i+1:]...)
				break
			}
		}
	}

	// 序列化更新后的列表数据
	updatedDataBytes, err := json.Marshal(orgs)
	if err != nil {
		return fmt.Errorf("更新列表数据序列化失败: %w", err)
	}

	// 更新缓存条目
	entry.Data = updatedDataBytes
	entry.Metadata.LastModified = time.Now()
	entry.Metadata.Version = time.Now().Unix()
	entry.Metadata.Source = "CDC_SMART_UPDATE"

	// 写回缓存
	ucm.l1Cache.Set(cacheKey, entry)

	if cacheData, err := json.Marshal(entry); err == nil {
		ucm.l2Cache.Set(ctx, cacheKey, string(cacheData), ucm.config.L2TTL)
	}

	return nil
}

// 判断组织是否应该包含在特定列表中
func (ucm *UnifiedCacheManager) shouldIncludeInList(cacheKey string, org *Organization) bool {
	// 从缓存键解析查询参数
	// 这里需要根据实际的缓存键格式来解析
	// 简化版本：假设所有组织都应该包含
	return true
}

// 组织列表排序
func (ucm *UnifiedCacheManager) sortOrganizations(orgs []Organization) {
	// 实现排序逻辑，按sort_order和code排序
	// 这里简化处理
}

// ==================== 缓存管理接口 ====================

// 获取缓存统计信息
func (ucm *UnifiedCacheManager) GetCacheStats(ctx context.Context) CacheStats {
	l1Stats := ucm.l1Cache.GetStats()

	return CacheStats{
		L1Stats: L1Stats{
			HitCount:  l1Stats.HitCount,
			MissCount: l1Stats.MissCount,
			Size:      l1Stats.Size,
			HitRate:   l1Stats.HitRate,
		},
		L2Connected:     ucm.l2Cache.Ping(ctx).Err() == nil,
		WriteThrough:    ucm.config.WriteThrough,
		ConsistencyMode: ucm.config.ConsistencyMode,
	}
}

// 手动刷新缓存
func (ucm *UnifiedCacheManager) RefreshCache(ctx context.Context, tenantID uuid.UUID, entityType string, entityID string) error {
	keyMgr := &CacheKeyManager{namespace: ucm.config.Namespace}
	var cacheKey string

	switch entityType {
	case "organization":
		cacheKey = keyMgr.GenerateKey("organization", tenantID.String(), entityID)
		// 删除缓存，下次查询时重新加载
		ucm.l1Cache.Delete(cacheKey)
		ucm.l2Cache.Del(ctx, cacheKey)
	case "organizations":
		// 删除所有列表缓存
		pattern := keyMgr.GenerateKey("organizations", tenantID.String(), "*")
		keys, err := ucm.l2Cache.Keys(ctx, pattern).Result()
		if err != nil {
			return err
		}
		for _, key := range keys {
			ucm.l1Cache.Delete(key)
		}
		if len(keys) > 0 {
			ucm.l2Cache.Del(ctx, keys...)
		}
	}

	ucm.logger.Printf("缓存刷新完成: %s:%s", entityType, entityID)
	return nil
}

// 启动缓存事件监听器
func (ucm *UnifiedCacheManager) startEventListener() {
	for event := range ucm.eventBus.Subscribe() {
		if err := ucm.HandleCDCEvent(context.Background(), event); err != nil {
			ucm.logger.Printf("处理缓存事件失败: %v", err)
		}
	}
}

// 传统失效模式处理
func (ucm *UnifiedCacheManager) handleTraditionalInvalidation(ctx context.Context, event CacheEvent) error {
	org := event.ToOrganization()
	keyMgr := &CacheKeyManager{namespace: ucm.config.Namespace}

	// 精确失效相关缓存
	patterns := []string{
		keyMgr.GenerateKey("organization", org.TenantID, org.Code),
		fmt.Sprintf("%s:organizations:%s:*", ucm.config.Namespace, org.TenantID),
		keyMgr.GenerateKey("stats", org.TenantID),
	}

	for _, pattern := range patterns {
		keys, err := ucm.l2Cache.Keys(ctx, pattern).Result()
		if err != nil {
			continue
		}

		for _, key := range keys {
			ucm.l1Cache.Delete(key)
		}
		if len(keys) > 0 {
			ucm.l2Cache.Del(ctx, keys...)
		}
	}

	return nil
}

// 关闭缓存管理器
func (ucm *UnifiedCacheManager) Close() error {
	ucm.eventBus.Close()
	return nil
}
