package database

import (
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	dbConnectionsInUse = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "db_connections_in_use",
			Help: "Number of database connections currently in use",
		},
		[]string{"service"},
	)

	dbConnectionsIdle = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "db_connections_idle",
			Help: "Number of idle database connections",
		},
		[]string{"service"},
	)

	dbQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
		},
		[]string{"service", "query_type"},
	)

	metricsOnce sync.Once
)

// RegisterMetrics 将当前包的指标注册到给定的 Registerer。
func RegisterMetrics(reg prometheus.Registerer) {
	if reg == nil {
		return
	}

	metricsOnce.Do(func() {
		reg.MustRegister(dbConnectionsInUse, dbConnectionsIdle, dbQueryDuration)
	})
}

// RecordConnectionStats 将当前连接池状态写入指标。
func (d *Database) RecordConnectionStats(serviceName string) {
	if d == nil || d.db == nil {
		return
	}

	label := resolveServiceLabel(serviceName, d.config.ServiceName)
	stats := d.db.Stats()

	dbConnectionsInUse.WithLabelValues(label).Set(float64(stats.InUse))
	dbConnectionsIdle.WithLabelValues(label).Set(float64(stats.Idle))
}

func recordQueryDuration(config ConnectionConfig, query string, duration time.Duration) {
	ObserveQueryDuration(resolveServiceLabel(config.ServiceName), extractQueryType(query), duration)
}

func resolveServiceLabel(names ...string) string {
	for _, name := range names {
		if strings.TrimSpace(name) != "" {
			return strings.TrimSpace(name)
		}
	}
	return "default"
}

func extractQueryType(query string) string {
	query = strings.TrimSpace(query)
	if query == "" {
		return ""
	}

	for len(query) > 0 && query[0] == '(' {
		query = query[1:]
	}

	query = strings.TrimSpace(query)
	if query == "" {
		return ""
	}

	parts := strings.Fields(query)
	if len(parts) == 0 {
		return ""
	}

	return strings.ToLower(parts[0])
}

// ObserveQueryDuration 对外开放的查询耗时记录入口，供其他组件复用。
func ObserveQueryDuration(serviceLabel, queryType string, duration time.Duration) {
	if queryType == "" {
		queryType = "unknown"
	}
	dbQueryDuration.WithLabelValues(resolveServiceLabel(serviceLabel), strings.ToLower(queryType)).Observe(duration.Seconds())
}
