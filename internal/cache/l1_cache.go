package cache

import (
	"container/list"
	"sync"
	"time"
)

// L1Cache 是基于 LRU 的进程内缓存实现。
type L1Cache struct {
	maxSize int
	ttl     time.Duration
	items   map[string]*list.Element
	lru     *list.List
	mu      sync.RWMutex
	stats   L1CacheStats
}

// L1缓存条目
type l1CacheItem struct {
	key       string
	entry     CacheEntry
	createdAt time.Time
	expiresAt time.Time
}

// L1CacheStats 汇总缓存命中率与容量信息。
type L1CacheStats struct {
	HitCount  int64   `json:"hit_count"`
	MissCount int64   `json:"miss_count"`
	Size      int     `json:"size"`
	HitRate   float64 `json:"hit_rate"`
}

// NewL1Cache 创建指定容量与 TTL 的缓存。
func NewL1Cache(maxSize int, ttl time.Duration) *L1Cache {
	cache := &L1Cache{
		maxSize: maxSize,
		ttl:     ttl,
		items:   make(map[string]*list.Element),
		lru:     list.New(),
		stats:   L1CacheStats{},
	}

	// 启动过期清理goroutine
	go cache.startEvictionWorker()

	return cache
}

// Get 返回缓存项及命中标记。
func (c *L1Cache) Get(key string) (CacheEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	element, exists := c.items[key]
	if !exists {
		c.stats.MissCount++
		c.updateHitRate()
		return CacheEntry{}, false
	}

	item := element.Value.(*l1CacheItem)

	// 检查过期
	if time.Now().After(item.expiresAt) {
		c.removeElement(element)
		c.stats.MissCount++
		c.updateHitRate()
		return CacheEntry{}, false
	}

	// 更新LRU位置
	c.lru.MoveToFront(element)
	c.stats.HitCount++
	c.updateHitRate()

	return item.entry, true
}

// Set 写入缓存条目。
func (c *L1Cache) Set(key string, entry CacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	item := &l1CacheItem{
		key:       key,
		entry:     entry,
		createdAt: now,
		expiresAt: now.Add(c.ttl),
	}

	// 如果键已存在，更新
	if element, exists := c.items[key]; exists {
		element.Value = item
		c.lru.MoveToFront(element)
		return
	}

	// 添加新项
	element := c.lru.PushFront(item)
	c.items[key] = element

	// 检查容量限制
	if c.lru.Len() > c.maxSize {
		c.evictOldest()
	}
}

// Delete 移除指定键。
func (c *L1Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, exists := c.items[key]; exists {
		c.removeElement(element)
	}
}

// 移除元素
func (c *L1Cache) removeElement(element *list.Element) {
	item := element.Value.(*l1CacheItem)
	delete(c.items, item.key)
	c.lru.Remove(element)
}

// 驱逐最老的项
func (c *L1Cache) evictOldest() {
	if c.lru.Len() > 0 {
		oldest := c.lru.Back()
		if oldest != nil {
			c.removeElement(oldest)
		}
	}
}

// 更新命中率
func (c *L1Cache) updateHitRate() {
	total := c.stats.HitCount + c.stats.MissCount
	if total > 0 {
		c.stats.HitRate = float64(c.stats.HitCount) / float64(total)
	}
}

// GetStats 返回当前缓存统计数据。
func (c *L1Cache) GetStats() L1CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := c.stats
	stats.Size = len(c.items)
	return stats
}

// Clear 清空缓存内容。
func (c *L1Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*list.Element)
	c.lru = list.New()
}

// 启动过期清理工作器
func (c *L1Cache) startEvictionWorker() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.evictExpired()
	}
}

// 清理过期项
func (c *L1Cache) evictExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	var toRemove []*list.Element

	for element := c.lru.Back(); element != nil; element = element.Prev() {
		item := element.Value.(*l1CacheItem)
		if now.After(item.expiresAt) {
			toRemove = append(toRemove, element)
		} else {
			break // 由于按时间排序，后面的都不会过期
		}
	}

	for _, element := range toRemove {
		c.removeElement(element)
	}
}
