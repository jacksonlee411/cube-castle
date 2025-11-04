package auth

import (
	"net/http"
	"strings"

	"cube-castle/internal/middleware"
	pkglogger "cube-castle/pkg/logger"
)

func scopedLogger(base pkglogger.Logger, component string, extra pkglogger.Fields) pkglogger.Logger {
	if base == nil {
		base = pkglogger.NewNoopLogger()
	}
	fields := pkglogger.Fields{
		"component": component,
	}
	for k, v := range extra {
		fields[k] = v
	}
	return base.WithFields(fields)
}

func requestLogger(base pkglogger.Logger, r *http.Request, action string, extra pkglogger.Fields) pkglogger.Logger {
	if base == nil {
		base = pkglogger.NewNoopLogger()
	}
	fields := pkglogger.Fields{}
	for k, v := range extra {
		fields[k] = v
	}
	if action != "" {
		fields["action"] = action
	}
	if r != nil {
		fields["method"] = r.Method
		fields["path"] = r.URL.Path
		fields["requestId"] = middleware.GetRequestID(r.Context())
		if tenant := strings.TrimSpace(r.Header.Get("X-Tenant-ID")); tenant != "" {
			fields["tenantId"] = tenant
		}
	}
	return base.WithFields(fields)
}
