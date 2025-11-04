package middleware

import (
	"net/http"

	pkglogger "cube-castle/pkg/logger"
)

func scopedLogger(base pkglogger.Logger, name string, extra pkglogger.Fields) pkglogger.Logger {
	if base == nil {
		base = pkglogger.NewNoopLogger()
	}
	fields := pkglogger.Fields{
		"component": "middleware",
	}
	if name != "" {
		fields["middleware"] = name
	}
	for k, v := range extra {
		fields[k] = v
	}
	return base.WithFields(fields)
}

func withRequestLogger(base pkglogger.Logger, r *http.Request) pkglogger.Logger {
	if r == nil {
		return base
	}
	return base.WithFields(pkglogger.Fields{
		"method":    r.Method,
		"path":      r.URL.Path,
		"requestId": GetRequestID(r.Context()),
	})
}
