package repository

import (
	pkglogger "cube-castle/pkg/logger"
)

func scopedLogger(base pkglogger.Logger, module, name string, extra pkglogger.Fields) pkglogger.Logger {
	if base == nil {
		base = pkglogger.NewNoopLogger()
	}

	fields := pkglogger.Fields{
		"component": "repository",
	}
	if module != "" {
		fields["module"] = module
	}
	if name != "" {
		fields["repository"] = name
	}
	for k, v := range extra {
		fields[k] = v
	}
	return base.WithFields(fields)
}
