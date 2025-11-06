package scheduler

import (
	pkglogger "cube-castle/pkg/logger"
)

func scopedLogger(base pkglogger.Logger, name string, extra pkglogger.Fields) pkglogger.Logger {
	if base == nil {
		base = pkglogger.NewNoopLogger()
	}

	fields := pkglogger.Fields{
		"component": "scheduler",
	}
	if name != "" {
		fields["service"] = name
	}
	if extra != nil {
		for k, v := range extra {
			fields[k] = v
		}
	}

	return base.WithFields(fields)
}
