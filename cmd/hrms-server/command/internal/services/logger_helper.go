package services

import (
	pkglogger "cube-castle/pkg/logger"
)

func scopedLogger(base pkglogger.Logger, service string, extra pkglogger.Fields) pkglogger.Logger {
	if base == nil {
		base = pkglogger.NewNoopLogger()
	}

	fields := pkglogger.Fields{
		"component": "service",
	}
	if service != "" {
		fields["service"] = service
	}
	for k, v := range extra {
		fields[k] = v
	}
	return base.WithFields(fields)
}
