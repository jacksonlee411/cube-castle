package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	RequestsPerMinute int           `json:"requestsPerMinute"`
	BurstSize         int           `json:"burstSize"`
	CleanupInterval   time.Duration `json:"cleanupInterval"`
	WhitelistIPs      []string      `json:"whitelistIPs"`
	BlockDuration     time.Duration `json:"blockDuration"`
}

// DefaultRateLimitConfig 默认限流配置
var DefaultRateLimitConfig = &RateLimitConfig{
	RequestsPerMinute: 100, // 每分钟100个请求
	BurstSize:         10,  // 允许10个突发请求
	CleanupInterval:   5 * time.Minute,
	WhitelistIPs:      []string{"127.0.0.1", "::1"},
	BlockDuration:     1 * time.Minute,
}

// ClientInfo 客户端信息
type ClientInfo struct {
	IP           string    `json:"ip"`
	RequestCount int       `json:"requestCount"`
	LastRequest  time.Time `json:"lastRequest"`
	BlockedUntil time.Time `json:"blockedUntil"`
	BurstCount   int       `json:"burstCount"`
	BurstStart   time.Time `json:"burstStart"`
}

// RateLimitMiddleware 限流中间件
type RateLimitMiddleware struct {
	config  *RateLimitConfig
	clients map[string]*ClientInfo
	mutex   sync.RWMutex
	logger  *log.Logger
	stats   *RateLimitStats
}

// RateLimitStats 限流统计
type RateLimitStats struct {
	TotalRequests   int64     `json:"totalRequests"`
	BlockedRequests int64     `json:"blockedRequests"`
	ActiveClients   int       `json:"activeClients"`
	LastReset       time.Time `json:"lastReset"`
	mutex           sync.RWMutex
}

// NewRateLimitMiddleware 创建限流中间件
func NewRateLimitMiddleware(config *RateLimitConfig, logger *log.Logger) *RateLimitMiddleware {
	if config == nil {
		config = DefaultRateLimitConfig
	}

	rlm := &RateLimitMiddleware{
		config:  config,
		clients: make(map[string]*ClientInfo),
		logger:  logger,
		stats: &RateLimitStats{
			LastReset: time.Now(),
		},
	}

	// 启动清理协程
	go rlm.cleanupRoutine()

	return rlm
}

// Middleware 限流中间件
func (rlm *RateLimitMiddleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)

			// 检查白名单
			if rlm.isWhitelisted(clientIP) {
				next.ServeHTTP(w, r)
				return
			}

			// 检查限流
			if !rlm.allowRequest(clientIP) {
				rlm.handleRateLimitExceeded(w, r, clientIP)
				return
			}

			// 更新统计信息
			rlm.updateStats(true)

			// 添加限流头部信息
			rlm.addRateLimitHeaders(w, clientIP)

			next.ServeHTTP(w, r)
		})
	}
}

// allowRequest 检查是否允许请求
func (rlm *RateLimitMiddleware) allowRequest(clientIP string) bool {
	rlm.mutex.Lock()
	defer rlm.mutex.Unlock()

	now := time.Now()
	client, exists := rlm.clients[clientIP]

	if !exists {
		// 新客户端
		rlm.clients[clientIP] = &ClientInfo{
			IP:           clientIP,
			RequestCount: 1,
			LastRequest:  now,
			BurstCount:   1,
			BurstStart:   now,
		}
		return true
	}

	// 检查是否被阻塞
	if now.Before(client.BlockedUntil) {
		return false
	}

	// 重置分钟计数器
	if now.Sub(client.LastRequest) > time.Minute {
		client.RequestCount = 0
		client.BurstCount = 0
		client.BurstStart = now
	}

	// 重置突发计数器
	if now.Sub(client.BurstStart) > 10*time.Second {
		client.BurstCount = 0
		client.BurstStart = now
	}

	// 检查每分钟限制
	if client.RequestCount >= rlm.config.RequestsPerMinute {
		client.BlockedUntil = now.Add(rlm.config.BlockDuration)
		rlm.logger.Printf("🚫 IP %s 超过每分钟请求限制 (%d), 阻塞 %v",
			clientIP, rlm.config.RequestsPerMinute, rlm.config.BlockDuration)
		return false
	}

	// 检查突发限制
	if client.BurstCount >= rlm.config.BurstSize {
		client.BlockedUntil = now.Add(rlm.config.BlockDuration / 2) // 突发阻塞时间较短
		rlm.logger.Printf("⚡ IP %s 超过突发请求限制 (%d), 短暂阻塞 %v",
			clientIP, rlm.config.BurstSize, rlm.config.BlockDuration/2)
		return false
	}

	// 更新客户端信息
	client.RequestCount++
	client.BurstCount++
	client.LastRequest = now

	return true
}

// isWhitelisted 检查IP是否在白名单中
func (rlm *RateLimitMiddleware) isWhitelisted(ip string) bool {
	for _, whiteIP := range rlm.config.WhitelistIPs {
		if ip == whiteIP {
			return true
		}
	}
	return false
}

// handleRateLimitExceeded 处理限流超限
func (rlm *RateLimitMiddleware) handleRateLimitExceeded(w http.ResponseWriter, r *http.Request, clientIP string) {
	rlm.updateStats(false)

	requestID := GetRequestID(r.Context())

	// 设置限流头部
	w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rlm.config.RequestsPerMinute))
	w.Header().Set("X-RateLimit-Remaining", "0")
	w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Minute).Unix(), 10))
	w.Header().Set("Retry-After", strconv.Itoa(int(rlm.config.BlockDuration.Seconds())))

	// 返回限流错误
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)

	response := fmt.Sprintf(`{
		"success": false,
		"error": {
			"code": "RATE_LIMIT_EXCEEDED",
			"message": "请求频率超过限制，请稍后重试"
		},
		"timestamp": "%s",
		"requestId": "%s",
		"meta": {
			"rateLimit": {
				"limit": %d,
				"remaining": 0,
				"resetTime": "%s",
				"retryAfter": "%s"
			}
		}
	}`, time.Now().UTC().Format(time.RFC3339),
		requestID,
		rlm.config.RequestsPerMinute,
		time.Now().Add(time.Minute).Format(time.RFC3339),
		rlm.config.BlockDuration.String())

	if _, err := w.Write([]byte(response)); err != nil {
		rlm.logger.Printf("写入限流响应失败: %v", err)
	}

	rlm.logger.Printf("🚫 限流拦截: IP %s | Path: %s | RequestID: %s",
		clientIP, r.URL.Path, requestID)
}

// addRateLimitHeaders 添加限流相关头部
func (rlm *RateLimitMiddleware) addRateLimitHeaders(w http.ResponseWriter, clientIP string) {
	rlm.mutex.RLock()
	client, exists := rlm.clients[clientIP]
	rlm.mutex.RUnlock()

	if !exists {
		return
	}

	remaining := rlm.config.RequestsPerMinute - client.RequestCount
	if remaining < 0 {
		remaining = 0
	}

	w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rlm.config.RequestsPerMinute))
	w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
	w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(client.LastRequest.Add(time.Minute).Unix(), 10))
}

// updateStats 更新统计信息
func (rlm *RateLimitMiddleware) updateStats(allowed bool) {
	rlm.stats.mutex.Lock()
	defer rlm.stats.mutex.Unlock()

	rlm.stats.TotalRequests++
	if !allowed {
		rlm.stats.BlockedRequests++
	}

	rlm.mutex.RLock()
	rlm.stats.ActiveClients = len(rlm.clients)
	rlm.mutex.RUnlock()
}

// cleanupRoutine 清理过期客户端
func (rlm *RateLimitMiddleware) cleanupRoutine() {
	ticker := time.NewTicker(rlm.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rlm.cleanupExpiredClients()
	}
}

// cleanupExpiredClients 清理过期客户端
func (rlm *RateLimitMiddleware) cleanupExpiredClients() {
	rlm.mutex.Lock()
	defer rlm.mutex.Unlock()

	now := time.Now()
	expiredCount := 0

	for ip, client := range rlm.clients {
		// 清理5分钟内没有请求的客户端
		if now.Sub(client.LastRequest) > 5*time.Minute {
			delete(rlm.clients, ip)
			expiredCount++
		}
	}

	if expiredCount > 0 {
		rlm.logger.Printf("🧹 限流清理: 清理了 %d 个过期客户端，当前活跃: %d",
			expiredCount, len(rlm.clients))
	}
}

// GetStats 获取限流统计信息
func (rlm *RateLimitMiddleware) GetStats() *RateLimitStats {
	rlm.stats.mutex.RLock()
	defer rlm.stats.mutex.RUnlock()

	// 返回副本避免并发问题
	return &RateLimitStats{
		TotalRequests:   rlm.stats.TotalRequests,
		BlockedRequests: rlm.stats.BlockedRequests,
		ActiveClients:   rlm.stats.ActiveClients,
		LastReset:       rlm.stats.LastReset,
	}
}

// GetClientInfo 获取客户端信息
func (rlm *RateLimitMiddleware) GetClientInfo(ip string) *ClientInfo {
	rlm.mutex.RLock()
	defer rlm.mutex.RUnlock()

	if client, exists := rlm.clients[ip]; exists {
		// 返回副本避免并发问题
		return &ClientInfo{
			IP:           client.IP,
			RequestCount: client.RequestCount,
			LastRequest:  client.LastRequest,
			BlockedUntil: client.BlockedUntil,
			BurstCount:   client.BurstCount,
			BurstStart:   client.BurstStart,
		}
	}
	return nil
}

// ResetStats 重置统计信息
func (rlm *RateLimitMiddleware) ResetStats() {
	rlm.stats.mutex.Lock()
	defer rlm.stats.mutex.Unlock()

	rlm.stats.TotalRequests = 0
	rlm.stats.BlockedRequests = 0
	rlm.stats.LastReset = time.Now()

	rlm.logger.Printf("📊 限流统计信息已重置")
}

// UpdateConfig 更新限流配置
func (rlm *RateLimitMiddleware) UpdateConfig(config *RateLimitConfig) {
	rlm.mutex.Lock()
	defer rlm.mutex.Unlock()

	rlm.config = config
	rlm.logger.Printf("⚙️ 限流配置已更新: %+v", config)
}

// GetActiveClients 获取活跃客户端列表
func (rlm *RateLimitMiddleware) GetActiveClients() map[string]*ClientInfo {
	rlm.mutex.RLock()
	defer rlm.mutex.RUnlock()

	clients := make(map[string]*ClientInfo)
	for ip, client := range rlm.clients {
		clients[ip] = &ClientInfo{
			IP:           client.IP,
			RequestCount: client.RequestCount,
			LastRequest:  client.LastRequest,
			BlockedUntil: client.BlockedUntil,
			BurstCount:   client.BurstCount,
			BurstStart:   client.BurstStart,
		}
	}
	return clients
}

// BlockIP 手动阻塞IP
func (rlm *RateLimitMiddleware) BlockIP(ip string, duration time.Duration) {
	rlm.mutex.Lock()
	defer rlm.mutex.Unlock()

	now := time.Now()
	client, exists := rlm.clients[ip]

	if !exists {
		rlm.clients[ip] = &ClientInfo{
			IP:           ip,
			LastRequest:  now,
			BlockedUntil: now.Add(duration),
		}
	} else {
		client.BlockedUntil = now.Add(duration)
	}

	rlm.logger.Printf("🚫 手动阻塞IP: %s, 时长: %v", ip, duration)
}

// UnblockIP 解除IP阻塞
func (rlm *RateLimitMiddleware) UnblockIP(ip string) {
	rlm.mutex.Lock()
	defer rlm.mutex.Unlock()

	if client, exists := rlm.clients[ip]; exists {
		client.BlockedUntil = time.Time{}
		rlm.logger.Printf("✅ 解除IP阻塞: %s", ip)
	}
}
