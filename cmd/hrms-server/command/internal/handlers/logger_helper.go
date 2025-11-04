package handlers

import (
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
