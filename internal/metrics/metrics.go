package metrics

import (
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    permissionCheckTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "graphql_permission_check_total",
            Help: "Total number of GraphQL permission checks",
        },
        []string{"query"},
    )

    permissionCheckSuccess = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "graphql_permission_check_success_total",
            Help: "Total number of successful GraphQL permission checks",
        },
        []string{"query"},
    )

    permissionCheckDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "graphql_permission_check_duration_seconds",
            Help:    "Duration of GraphQL permission checks in seconds",
            Buckets: []float64{0.0005, 0.001, 0.005, 0.01, 0.02, 0.05, 0.1},
        },
        []string{"query"},
    )
)

// RecordPermissionCheck 记录GraphQL权限检查指标
func RecordPermissionCheck(query string, success bool, d time.Duration) {
    permissionCheckTotal.WithLabelValues(query).Inc()
    permissionCheckDuration.WithLabelValues(query).Observe(d.Seconds())
    if success {
        permissionCheckSuccess.WithLabelValues(query).Inc()
    }
}

