package health

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	pkglogger "cube-castle/pkg/logger"
)

// AlertLevel 告警级别
type AlertLevel string

const (
	AlertLevelInfo     AlertLevel = "info"
	AlertLevelWarning  AlertLevel = "warning"
	AlertLevelCritical AlertLevel = "critical"
)

// Alert 告警信息
type Alert struct {
	ID         string                 `json:"id"`
	Service    string                 `json:"service"`
	Component  string                 `json:"component"`
	Level      AlertLevel             `json:"level"`
	Status     HealthStatus           `json:"status"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Resolved   bool                   `json:"resolved"`
	ResolvedAt *time.Time             `json:"resolvedAt,omitempty"`
}

// AlertRule 告警规则
type AlertRule struct {
	Name          string         `json:"name"`
	Component     string         `json:"component"`
	Condition     AlertCondition `json:"condition"`
	Level         AlertLevel     `json:"level"`
	Message       string         `json:"message"`
	Cooldown      time.Duration  `json:"cooldown"`
	MaxRetries    int            `json:"maxRetries"`
	EnabledBy     time.Time      `json:"enabledBy"`
	lastTriggered time.Time
	retryCount    int
}

// AlertCondition 告警条件
type AlertCondition struct {
	StatusEquals     *HealthStatus  `json:"statusEquals,omitempty"`
	ResponseTimeGT   *time.Duration `json:"responseTimeGt,omitempty"`
	ConsecutiveFails *int           `json:"consecutiveFails,omitempty"`
}

// AlertChannel 告警渠道接口
type AlertChannel interface {
	Send(ctx context.Context, alert Alert) error
	Name() string
}

type loggerAwareChannel interface {
	SetLogger(pkglogger.Logger)
}

// WebhookChannel Webhook告警渠道
type WebhookChannel struct {
	name    string
	url     string
	headers map[string]string
	timeout time.Duration
}

// NewWebhookChannel 创建Webhook告警渠道
func NewWebhookChannel(name, url string) *WebhookChannel {
	return &WebhookChannel{
		name:    name,
		url:     url,
		headers: make(map[string]string),
		timeout: 10 * time.Second,
	}
}

func (w *WebhookChannel) Name() string {
	return w.name
}

func (w *WebhookChannel) AddHeader(key, value string) {
	w.headers[key] = value
}

func (w *WebhookChannel) Send(ctx context.Context, alert Alert) error {
	payload := map[string]interface{}{
		"alert":     alert,
		"timestamp": time.Now(),
		"source":    "cube-castle-health-monitor",
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", w.url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "cube-castle-health-monitor/1.0")

	for key, value := range w.headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// SlackChannel Slack告警渠道
type SlackChannel struct {
	webhookURL string
	channel    string
	username   string
}

// NewSlackChannel 创建Slack告警渠道
func NewSlackChannel(webhookURL, channel, username string) *SlackChannel {
	return &SlackChannel{
		webhookURL: webhookURL,
		channel:    channel,
		username:   username,
	}
}

func (s *SlackChannel) Name() string {
	return "slack"
}

func (s *SlackChannel) Send(ctx context.Context, alert Alert) error {
	// 根据告警级别选择颜色和emoji
	var color, emoji string
	switch alert.Level {
	case AlertLevelCritical:
		color = "#FF0000"
		emoji = "🚨"
	case AlertLevelWarning:
		color = "#FFA500"
		emoji = "⚠️"
	case AlertLevelInfo:
		color = "#0000FF"
		emoji = "ℹ️"
	}

	statusEmoji := ""
	switch alert.Status {
	case StatusHealthy:
		statusEmoji = "✅"
	case StatusDegraded:
		statusEmoji = "🟡"
	case StatusUnhealthy:
		statusEmoji = "❌"
	}

	payload := map[string]interface{}{
		"channel":  s.channel,
		"username": s.username,
		"attachments": []map[string]interface{}{
			{
				"color": color,
				"title": fmt.Sprintf("%s Cube Castle 健康告警", emoji),
				"text":  alert.Message,
				"fields": []map[string]interface{}{
					{
						"title": "服务",
						"value": alert.Service,
						"short": true,
					},
					{
						"title": "组件",
						"value": alert.Component,
						"short": true,
					},
					{
						"title": "状态",
						"value": fmt.Sprintf("%s %s", statusEmoji, alert.Status),
						"short": true,
					},
					{
						"title": "级别",
						"value": string(alert.Level),
						"short": true,
					},
					{
						"title": "时间",
						"value": alert.Timestamp.Format("2006-01-02 15:04:05"),
						"short": true,
					},
				},
				"footer": "Cube Castle Health Monitor",
				"ts":     alert.Timestamp.Unix(),
			},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal slack payload: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", s.webhookURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create slack request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send slack webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// EmailChannel 邮件告警渠道 (简化实现)
type EmailChannel struct {
	smtpHost string
	smtpPort int
	username string
	password string
	from     string
	to       []string
	logger   pkglogger.Logger
}

func (e *EmailChannel) Name() string {
	return "email"
}

// SetLogger 允许注入结构化日志器
func (e *EmailChannel) SetLogger(logger pkglogger.Logger) {
	if logger != nil {
		e.logger = logger.WithFields(pkglogger.Fields{
			"channel": "email",
		})
	}
}

func (e *EmailChannel) Send(ctx context.Context, alert Alert) error {
	// 这里应该实现SMTP邮件发送
	// 为了简化，这里只是记录日志
	logger := e.logger
	if logger == nil {
		logger = pkglogger.NewNoopLogger()
	}
	logger.WithFields(pkglogger.Fields{
		"channel":   "email",
		"alertId":   alert.ID,
		"level":     alert.Level,
		"service":   alert.Service,
		"component": alert.Component,
	}).Infof("EMAIL ALERT: %s", alert.Message)
	return nil
}

// AlertManager 告警管理器
type AlertManager struct {
	serviceName      string
	rules            []AlertRule
	channels         []AlertChannel
	activeAlerts     map[string]*Alert
	alertHistory     []Alert
	mu               sync.RWMutex
	maxHistorySize   int
	healthStates     map[string]HealthStatus
	consecutiveFails map[string]int
	logger           pkglogger.Logger
}

// NewAlertManager 创建告警管理器
func NewAlertManager(serviceName string) *AlertManager {
	return &AlertManager{
		serviceName:      serviceName,
		rules:            make([]AlertRule, 0),
		channels:         make([]AlertChannel, 0),
		activeAlerts:     make(map[string]*Alert),
		alertHistory:     make([]Alert, 0),
		maxHistorySize:   1000,
		healthStates:     make(map[string]HealthStatus),
		consecutiveFails: make(map[string]int),
		logger: pkglogger.NewLogger(
			pkglogger.WithLevel(pkglogger.LevelInfo),
		).WithFields(pkglogger.Fields{
			"service":   serviceName,
			"component": "health-alerting",
		}),
	}
}

// WithLogger 允许注入结构化日志器
func (am *AlertManager) WithLogger(logger pkglogger.Logger) *AlertManager {
	if logger != nil {
		am.logger = logger.WithFields(pkglogger.Fields{
			"service":   am.serviceName,
			"component": "health-alerting",
		})
	}
	return am
}

// AddRule 添加告警规则
func (am *AlertManager) AddRule(rule AlertRule) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.rules = append(am.rules, rule)
}

// AddChannel 添加告警渠道
func (am *AlertManager) AddChannel(channel AlertChannel) {
	am.mu.Lock()
	defer am.mu.Unlock()
	if lc, ok := channel.(loggerAwareChannel); ok {
		lc.SetLogger(am.logger)
	}
	am.channels = append(am.channels, channel)
}

// ProcessHealthCheck 处理健康检查结果
func (am *AlertManager) ProcessHealthCheck(ctx context.Context, health ServiceHealth) {
	am.mu.Lock()
	defer am.mu.Unlock()

	// 更新健康状态历史
	for _, check := range health.Checks {
		componentKey := fmt.Sprintf("%s:%s", health.Service, check.Name)

		// 记录连续失败次数
		if check.Status == StatusUnhealthy {
			am.consecutiveFails[componentKey]++
		} else {
			am.consecutiveFails[componentKey] = 0
		}

		// 检查是否需要触发告警
		am.evaluateRules(ctx, health.Service, check)

		// 更新健康状态
		am.healthStates[componentKey] = check.Status
	}

	// 检查是否有告警需要解决
	am.checkResolvedAlerts(ctx, health)
}

// evaluateRules 评估告警规则
func (am *AlertManager) evaluateRules(ctx context.Context, serviceName string, check HealthCheck) {
	for _, rule := range am.rules {
		if rule.Component != "" && rule.Component != check.Name {
			continue
		}

		// 检查冷却时间
		if time.Since(rule.lastTriggered) < rule.Cooldown {
			continue
		}

		// 评估条件
		if am.evaluateCondition(rule.Condition, serviceName, check) {
			rule.lastTriggered = time.Now()
			am.triggerAlert(ctx, rule, serviceName, check)
		}
	}
}

// evaluateCondition 评估告警条件
func (am *AlertManager) evaluateCondition(condition AlertCondition, serviceName string, check HealthCheck) bool {
	// 检查状态条件
	if condition.StatusEquals != nil && check.Status == *condition.StatusEquals {
		return true
	}

	// 检查响应时间条件
	if condition.ResponseTimeGT != nil && check.Duration > *condition.ResponseTimeGT {
		return true
	}

	// 检查连续失败次数
	if condition.ConsecutiveFails != nil {
		componentKey := fmt.Sprintf("%s:%s", serviceName, check.Name)
		if am.consecutiveFails[componentKey] >= *condition.ConsecutiveFails {
			return true
		}
	}

	return false
}

// triggerAlert 触发告警
func (am *AlertManager) triggerAlert(ctx context.Context, rule AlertRule, serviceName string, check HealthCheck) {
	alertID := fmt.Sprintf("%s-%s-%d", serviceName, check.Name, time.Now().Unix())

	alert := Alert{
		ID:        alertID,
		Service:   serviceName,
		Component: check.Name,
		Level:     rule.Level,
		Status:    check.Status,
		Message:   fmt.Sprintf(rule.Message, check.Name, check.Status, check.Message),
		Details:   check.Details,
		Timestamp: time.Now(),
		Resolved:  false,
	}

	// 保存活跃告警
	am.activeAlerts[alertID] = &alert

	// 添加到历史记录
	am.addToHistory(alert)

	// 发送告警到所有渠道
	for _, channel := range am.channels {
		ch := channel
		go func(ch AlertChannel) {
			if err := ch.Send(ctx, alert); err != nil {
				am.logger.WithFields(pkglogger.Fields{
					"channel": ch.Name(),
					"alertId": alert.ID,
					"error":   err,
				}).Error("failed to dispatch alert")
			}
		}(ch)
	}

	am.logger.WithFields(pkglogger.Fields{
		"alertId":   alert.ID,
		"level":     alert.Level,
		"service":   alert.Service,
		"component": alert.Component,
	}).Info("alert triggered")
}

// checkResolvedAlerts 检查已解决的告警
func (am *AlertManager) checkResolvedAlerts(ctx context.Context, health ServiceHealth) {
	for alertID, alert := range am.activeAlerts {
		if alert.Resolved {
			continue
		}

		// 查找对应的健康检查
		for _, check := range health.Checks {
			if check.Name == alert.Component && check.Status == StatusHealthy {
				// 标记告警为已解决
				alert.Resolved = true
				now := time.Now()
				alert.ResolvedAt = &now

				// 发送解决通知
				resolvedAlert := *alert
				resolvedAlert.Message = fmt.Sprintf("✅ 告警已解决: %s 组件 %s 恢复正常", alert.Service, alert.Component)
				resolvedAlert.Level = AlertLevelInfo

				for _, channel := range am.channels {
					ch := channel
					go func(ch AlertChannel) {
						if err := ch.Send(ctx, resolvedAlert); err != nil {
							am.logger.WithFields(pkglogger.Fields{
								"channel": ch.Name(),
								"alertId": alert.ID,
								"error":   err,
							}).Error("failed to dispatch resolved alert")
						}
					}(ch)
				}

				am.logger.WithFields(pkglogger.Fields{
					"alertId":   alert.ID,
					"service":   alert.Service,
					"component": alert.Component,
				}).Info("alert resolved")
				delete(am.activeAlerts, alertID)
				break
			}
		}
	}
}

// addToHistory 添加到历史记录
func (am *AlertManager) addToHistory(alert Alert) {
	am.alertHistory = append(am.alertHistory, alert)

	// 保持历史记录大小限制
	if len(am.alertHistory) > am.maxHistorySize {
		am.alertHistory = am.alertHistory[1:]
	}
}

// GetActiveAlerts 获取活跃告警
func (am *AlertManager) GetActiveAlerts() []Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	alerts := make([]Alert, 0, len(am.activeAlerts))
	for _, alert := range am.activeAlerts {
		alerts = append(alerts, *alert)
	}

	return alerts
}

// GetAlertHistory 获取告警历史
func (am *AlertManager) GetAlertHistory(limit int) []Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	if limit <= 0 || limit > len(am.alertHistory) {
		limit = len(am.alertHistory)
	}

	start := len(am.alertHistory) - limit
	return am.alertHistory[start:]
}
