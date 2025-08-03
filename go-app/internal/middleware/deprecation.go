package middleware

import (
	"net/http"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
)

// DeprecationInfo 废弃信息结构
type DeprecationInfo struct {
	Endpoint    string
	Replacement string
	RemovalDate string
	Message     string
}

// DeprecationMiddleware 废弃端点中间件
func DeprecationMiddleware(info DeprecationInfo, logger *logging.StructuredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 添加废弃标头
			w.Header().Set("Deprecation", "true")
			w.Header().Set("Sunset", info.RemovalDate)
			w.Header().Set("Link", `<`+info.Replacement+`>; rel="successor-version"`)
			
			// 记录废弃端点使用情况
			logger.Warn("Deprecated endpoint accessed",
				"endpoint", info.Endpoint,
				"client_ip", r.RemoteAddr,
				"user_agent", r.UserAgent(),
				"replacement", info.Replacement,
				"removal_date", info.RemovalDate,
				"method", r.Method,
				"path", r.URL.Path,
			)
			
			// 继续处理请求
			next.ServeHTTP(w, r)
		})
	}
}

// WrapDeprecatedHandler 包装废弃的处理器
func WrapDeprecatedHandler(handler http.HandlerFunc, info DeprecationInfo, logger *logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 添加废弃标头
		w.Header().Set("Deprecation", "true")
		w.Header().Set("Sunset", info.RemovalDate)
		w.Header().Set("Link", `<`+info.Replacement+`>; rel="successor-version"`)
		
		// 记录废弃端点使用情况
		logger.Warn("Deprecated endpoint accessed",
			"endpoint", info.Endpoint,
			"client_ip", r.RemoteAddr,
			"user_agent", r.UserAgent(),
			"replacement", info.Replacement,
			"removal_date", info.RemovalDate,
			"method", r.Method,
			"path", r.URL.Path,
		)
		
		// 执行原处理器
		handler(w, r)
	}
}

// EmployeeDeprecationInfo 员工端点废弃信息
var EmployeeDeprecationInfo = map[string]DeprecationInfo{
	"list": {
		Endpoint:    "/api/v1/employees",
		Replacement: "/api/v1/queries/employees",
		RemovalDate: "2024-12-31",
		Message:     "Please migrate to CQRS queries endpoint: /api/v1/queries/employees",
	},
	"create": {
		Endpoint:    "/api/v1/employees",
		Replacement: "/api/v1/commands/hire-employee",
		RemovalDate: "2024-12-31",
		Message:     "Please migrate to CQRS commands endpoint: /api/v1/commands/hire-employee",
	},
	"get": {
		Endpoint:    "/api/v1/employees/{id}",
		Replacement: "/api/v1/queries/employees/{id}",
		RemovalDate: "2024-12-31",
		Message:     "Please migrate to CQRS queries endpoint: /api/v1/queries/employees/{id}",
	},
	"update": {
		Endpoint:    "/api/v1/employees/{id}",
		Replacement: "/api/v1/commands/update-employee",
		RemovalDate: "2024-12-31",
		Message:     "Please migrate to CQRS commands endpoint: /api/v1/commands/update-employee",
	},
	"delete": {
		Endpoint:    "/api/v1/employees/{id}",
		Replacement: "/api/v1/commands/terminate-employee",
		RemovalDate: "2024-12-31",
		Message:     "Please migrate to CQRS commands endpoint: /api/v1/commands/terminate-employee",
	},
	"assign_position": {
		Endpoint:    "/api/v1/employees/{id}/assign-position",
		Replacement: "/api/v1/commands/assign-employee-position",
		RemovalDate: "2024-12-31",
		Message:     "Please migrate to CQRS commands endpoint: /api/v1/commands/assign-employee-position",
	},
	"position_history": {
		Endpoint:    "/api/v1/employees/{id}/position-history",
		Replacement: "/api/v1/queries/reporting-hierarchy/{manager_id}",
		RemovalDate: "2024-12-31",
		Message:     "Please migrate to CQRS queries endpoint for position history",
	},
}