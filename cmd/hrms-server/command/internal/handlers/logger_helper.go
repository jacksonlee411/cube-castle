package handlers

import (
	"net/http"

	"cube-castle/cmd/hrms-server/command/internal/middleware"
	pkglogger "cube-castle/pkg/logger"
)

func scopedLogger(base pkglogger.Logger, handler string, extra pkglogger.Fields) pkglogger.Logger {
	if base == nil {
		base = pkglogger.NewNoopLogger()
	}

	fields := pkglogger.Fields{
		"component": "handler",
	}
	if handler != "" {
		fields["handler"] = handler
	}
	for k, v := range extra {
		fields[k] = v
	}
	return base.WithFields(fields)
}

func requestScopedLogger(base pkglogger.Logger, r *http.Request, action string, extra pkglogger.Fields) pkglogger.Logger {
	if base == nil {
		base = pkglogger.NewNoopLogger()
	}
	fields := cloneFields(extra)
	if fields == nil {
		fields = pkglogger.Fields{}
	}
	if action != "" {
		fields["action"] = action
	}
	if r != nil {
		fields["method"] = r.Method
		fields["path"] = r.URL.Path
		fields["requestId"] = middleware.GetRequestID(r.Context())
	}
	return base.WithFields(fields)
}

func cloneFields(src pkglogger.Fields) pkglogger.Fields {
	if len(src) == 0 {
		return nil
	}
	dup := make(pkglogger.Fields, len(src))
	for k, v := range src {
		dup[k] = v
	}
	return dup
}
