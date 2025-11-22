package health

import (
	"context"
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
	"time"
)

// StatusReporter 状态报告生成器
type StatusReporter struct {
	healthManager *HealthManager
	baseURL       string
}

// NewStatusReporter 创建状态报告生成器
func NewStatusReporter(hm *HealthManager, baseURL string) *StatusReporter {
	return &StatusReporter{
		healthManager: hm,
		baseURL:       baseURL,
	}
}

// ServiceDashboard 服务仪表板数据
type ServiceDashboard struct {
	Service      string             `json:"service"`
	Version      string             `json:"version"`
	Status       HealthStatus       `json:"status"`
	Timestamp    time.Time          `json:"timestamp"`
	Uptime       time.Duration      `json:"uptime"`
	Summary      Summary            `json:"summary"`
	Checks       []HealthCheck      `json:"checks"`
	Metrics      ServiceMetrics     `json:"metrics"`
	Environment  EnvironmentInfo    `json:"environment"`
	Dependencies []DependencyStatus `json:"dependencies"`
}

// ServiceMetrics 服务指标
type ServiceMetrics struct {
	ResponseTime  time.Duration `json:"responseTime"`
	RequestCount  int64         `json:"requestCount"`
	ErrorRate     float64       `json:"errorRate"`
	MemoryUsage   string        `json:"memoryUsage"`
	CPUUsage      string        `json:"cpuUsage"`
	DatabaseConns int           `json:"databaseConnections"`
	CacheHitRate  float64       `json:"cacheHitRate"`
}

// EnvironmentInfo 环境信息
type EnvironmentInfo struct {
	Hostname    string            `json:"hostname"`
	Platform    string            `json:"platform"`
	GoVersion   string            `json:"goVersion"`
	Environment string            `json:"environment"`
	Region      string            `json:"region"`
	Config      map[string]string `json:"config"`
}

// DependencyStatus 依赖状态
type DependencyStatus struct {
	Name         string        `json:"name"`
	Status       HealthStatus  `json:"status"`
	LastChecked  time.Time     `json:"lastChecked"`
	ResponseTime time.Duration `json:"responseTime"`
	Version      string        `json:"version,omitempty"`
	URL          string        `json:"url,omitempty"`
}

// StatusPage 状态页面模板数据
type StatusPage struct {
	Title         string           `json:"title"`
	LastUpdated   time.Time        `json:"lastUpdated"`
	OverallStatus HealthStatus     `json:"overallStatus"`
	Services      []ServiceSummary `json:"services"`
	Incidents     []Incident       `json:"incidents"`
	Metrics       SystemMetrics    `json:"metrics"`
}

// ServiceSummary 服务摘要
type ServiceSummary struct {
	Name         string        `json:"name"`
	Status       HealthStatus  `json:"status"`
	Uptime       string        `json:"uptime"`
	ResponseTime time.Duration `json:"responseTime"`
	LastIncident *time.Time    `json:"lastIncident,omitempty"`
}

// Incident 事件记录
type Incident struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"` // investigating, monitoring, resolved
	StartTime   time.Time  `json:"startTime"`
	EndTime     *time.Time `json:"endTime,omitempty"`
	Severity    string     `json:"severity"` // low, medium, high, critical
	Services    []string   `json:"affectedServices"`
}

// SystemMetrics 系统指标
type SystemMetrics struct {
	TotalServices    int    `json:"totalServices"`
	HealthyServices  int    `json:"healthyServices"`
	DegradedServices int    `json:"degradedServices"`
	FailedServices   int    `json:"failedServices"`
	OverallUptime    string `json:"overallUptime"`
	AvgResponseTime  string `json:"avgResponseTime"`
}

// GenerateDashboard 生成服务仪表板
func (sr *StatusReporter) GenerateDashboard(ctx context.Context) ServiceDashboard {
	health := sr.healthManager.Check(ctx)

	// 计算指标
	metrics := sr.calculateMetrics(health)

	// 获取环境信息
	env := sr.getEnvironmentInfo()

	// 获取依赖状态
	deps := sr.getDependencyStatus(health)

	return ServiceDashboard{
		Service:      health.Service,
		Version:      health.Version,
		Status:       health.Status,
		Timestamp:    health.Timestamp,
		Uptime:       health.Uptime,
		Summary:      health.Summary,
		Checks:       health.Checks,
		Metrics:      metrics,
		Environment:  env,
		Dependencies: deps,
	}
}

// calculateMetrics 计算服务指标
func (sr *StatusReporter) calculateMetrics(health ServiceHealth) ServiceMetrics {
	// 计算平均响应时间
	var totalDuration time.Duration
	for _, check := range health.Checks {
		totalDuration += check.Duration
	}

	avgResponseTime := time.Duration(0)
	if len(health.Checks) > 0 {
		avgResponseTime = totalDuration / time.Duration(len(health.Checks))
	}

	// 计算错误率
	errorRate := 0.0
	if health.Summary.Total > 0 {
		errorRate = float64(health.Summary.Failed) / float64(health.Summary.Total) * 100
	}

	return ServiceMetrics{
		ResponseTime:  avgResponseTime,
		RequestCount:  0, // 需要从实际指标系统获取
		ErrorRate:     errorRate,
		MemoryUsage:   "N/A", // 需要从系统获取
		CPUUsage:      "N/A", // 需要从系统获取
		DatabaseConns: 0,     // 需要从数据库获取
		CacheHitRate:  0.0,   // 需要从缓存系统获取
	}
}

// getEnvironmentInfo 获取环境信息
func (sr *StatusReporter) getEnvironmentInfo() EnvironmentInfo {
	return EnvironmentInfo{
		Hostname:    "localhost", // 应该从系统获取
		Platform:    "docker",
		GoVersion:   "1.22",
		Environment: "development",
		Region:      "local",
		Config: map[string]string{
			"debug_mode": "true",
			"log_level":  "info",
		},
	}
}

// getDependencyStatus 获取依赖状态
func (sr *StatusReporter) getDependencyStatus(health ServiceHealth) []DependencyStatus {
	var deps []DependencyStatus

	for _, check := range health.Checks {
		if strings.Contains(check.Name, "service") || strings.Contains(check.Name, "dependency") {
			deps = append(deps, DependencyStatus{
				Name:         check.Name,
				Status:       check.Status,
				LastChecked:  time.Now(),
				ResponseTime: check.Duration,
			})
		}
	}

	return deps
}

// DashboardHandler 仪表板HTTP处理器
func (sr *StatusReporter) DashboardHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		dashboard := sr.GenerateDashboard(ctx)

		// 检查请求的格式
		accept := r.Header.Get("Accept")
		if strings.Contains(accept, "application/json") || r.URL.Query().Get("format") == "json" {
			// 返回JSON格式
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(dashboard); err != nil {
				http.Error(w, "failed to encode dashboard", http.StatusInternalServerError)
			}
			return
		}

		// 返回HTML格式
		w.Header().Set("Content-Type", "text/html")
		sr.renderHTMLDashboard(w, dashboard)
	}
}

// renderHTMLDashboard 渲染HTML仪表板
func (sr *StatusReporter) renderHTMLDashboard(w http.ResponseWriter, dashboard ServiceDashboard) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>{{.Service}} - 服务健康仪表板</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
        .status-card { background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .status-healthy { border-left: 4px solid #4CAF50; }
        .status-degraded { border-left: 4px solid #FF9800; }
        .status-unhealthy { border-left: 4px solid #F44336; }
        .metrics-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 20px; }
        .metric-item { background: white; padding: 15px; border-radius: 8px; text-align: center; }
        .metric-value { font-size: 2em; font-weight: bold; color: #333; }
        .metric-label { color: #666; margin-top: 5px; }
        .checks-table { width: 100%; border-collapse: collapse; }
        .checks-table th, .checks-table td { padding: 10px; text-align: left; border-bottom: 1px solid #ddd; }
        .checks-table th { background-color: #f8f9fa; }
        .status-badge { padding: 4px 8px; border-radius: 4px; color: white; font-size: 0.8em; }
        .badge-healthy { background-color: #4CAF50; }
        .badge-degraded { background-color: #FF9800; }
        .badge-unhealthy { background-color: #F44336; }
        .refresh-btn { background: #007bff; color: white; padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; }
        .refresh-btn:hover { background: #0056b3; }
    </style>
    <script>
        function refreshData() {
            location.reload();
        }
        
        // 自动刷新每30秒
        setInterval(refreshData, 30000);
    </script>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🏰 {{.Service}} 健康仪表板</h1>
            <p>版本: {{.Version}} | 状态: <span class="status-badge badge-{{.Status}}">{{.Status}}</span> | 运行时间: {{.Uptime}}</p>
            <p>最后更新: {{.Timestamp.Format "2006-01-02 15:04:05"}}</p>
            <button class="refresh-btn" onclick="refreshData()">🔄 刷新</button>
        </div>
        
        <div class="metrics-grid">
            <div class="metric-item">
                <div class="metric-value">{{.Summary.Total}}</div>
                <div class="metric-label">总检查项</div>
            </div>
            <div class="metric-item">
                <div class="metric-value" style="color: #4CAF50;">{{.Summary.Healthy}}</div>
                <div class="metric-label">健康</div>
            </div>
            <div class="metric-item">
                <div class="metric-value" style="color: #FF9800;">{{.Summary.Degraded}}</div>
                <div class="metric-label">降级</div>
            </div>
            <div class="metric-item">
                <div class="metric-value" style="color: #F44336;">{{.Summary.Failed}}</div>
                <div class="metric-label">失败</div>
            </div>
        </div>
        
        <div class="status-card status-{{.Status}}">
            <h2>📊 健康检查详情</h2>
            <table class="checks-table">
                <thead>
                    <tr>
                        <th>组件</th>
                        <th>状态</th>
                        <th>响应时间</th>
                        <th>消息</th>
                        <th>详情</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Checks}}
                    <tr>
                        <td>{{.Name}}</td>
                        <td><span class="status-badge badge-{{.Status}}">{{.Status}}</span></td>
                        <td>{{.Duration}}</td>
                        <td>{{.Message}}</td>
                        <td>{{if .Details}}📋 有详情{{else}}-{{end}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </div>
        
        <div class="status-card">
            <h2>🔗 服务依赖</h2>
            {{if .Dependencies}}
            <table class="checks-table">
                <thead>
                    <tr>
                        <th>依赖服务</th>
                        <th>状态</th>
                        <th>响应时间</th>
                        <th>最后检查</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Dependencies}}
                    <tr>
                        <td>{{.Name}}</td>
                        <td><span class="status-badge badge-{{.Status}}">{{.Status}}</span></td>
                        <td>{{.ResponseTime}}</td>
                        <td>{{.LastChecked.Format "15:04:05"}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
            {{else}}
            <p>暂无依赖服务</p>
            {{end}}
        </div>
        
        <div class="status-card">
            <h2>⚙️ 环境信息</h2>
            <div class="metrics-grid">
                <div class="metric-item">
                    <div class="metric-value" style="font-size: 1.2em;">{{.Environment.Platform}}</div>
                    <div class="metric-label">平台</div>
                </div>
                <div class="metric-item">
                    <div class="metric-value" style="font-size: 1.2em;">{{.Environment.GoVersion}}</div>
                    <div class="metric-label">Go版本</div>
                </div>
                <div class="metric-item">
                    <div class="metric-value" style="font-size: 1.2em;">{{.Environment.Environment}}</div>
                    <div class="metric-label">环境</div>
                </div>
                <div class="metric-item">
                    <div class="metric-value" style="font-size: 1.2em;">{{.Environment.Hostname}}</div>
                    <div class="metric-label">主机</div>
                </div>
            </div>
        </div>
    </div>
</body>
</html>
`

	t, err := template.New("dashboard").Parse(tmpl)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, dashboard)
	if err != nil {
		http.Error(w, "Render error", http.StatusInternalServerError)
		return
	}
}

// StatusPageHandler 状态页面处理器
func (sr *StatusReporter) StatusPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		// 这里可以实现一个公共状态页面
		// 显示所有服务的整体状态
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"message":   "Status page coming soon",
			"timestamp": time.Now(),
		}); err != nil {
			http.Error(w, "failed to encode status page", http.StatusInternalServerError)
		}
	}
}
