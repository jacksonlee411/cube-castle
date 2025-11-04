package logger

import (
	"log"
	"strings"
)

// NewStdLogger returns a standard-library logger that delegates output to the
// structured Logger. It is intended as a temporary bridge for legacy code that
// still depends on *log.Logger; Plan 218E will remove this adapter once all
// call sites are migrated.
func NewStdLogger(l Logger, level Level) *log.Logger {
	if l == nil {
		l = NewNoopLogger()
	}
	return log.New(&stdWriter{
		logger: l,
		level:  level,
	}, "", 0)
}

type stdWriter struct {
	logger Logger
	level  Level
}

func (w *stdWriter) Write(p []byte) (int, error) {
	if len(p) == 0 || w.logger == nil {
		return len(p), nil
	}

	msg := strings.TrimRight(string(p), "\n")
	switch w.level {
	case LevelDebug:
		w.logger.Debug(msg)
	case LevelWarn:
		w.logger.Warn(msg)
	case LevelError:
		w.logger.Error(msg)
	default:
		w.logger.Info(msg)
	}
	return len(p), nil
}
