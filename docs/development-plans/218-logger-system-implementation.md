# Plan 218 - `pkg/logger/` 日志系统实现

**文档编号**: 218
**标题**: 结构化日志系统 - 统一实现
**创建日期**: 2025-11-04
**分支**: `feature/204-phase2-infrastructure`
**版本**: v1.0
**关联计划**: Plan 216（eventbus）、Plan 217（database）、Plan 215（Phase2 执行日志）

---

## 1. 概述

### 1.1 目标

实现统一的结构化日志系统（pkg/logger），为所有模块提供：
- 结构化日志记录（JSON 格式）
- 日志级别控制（Debug, Info, Warn, Error）
- 性能监控集成
- Prometheus 指标暴露

**关键成果**:
- ✅ Logger 接口定义
- ✅ 结构化日志实现
- ✅ 日志级别控制
- ✅ Prometheus 集成
- ✅ 单元测试（覆盖率 > 80%）

### 1.2 为什么需要统一的日志系统

- **可观测性** - 所有模块统一格式，便于日志聚合分析
- **性能追踪** - 记录关键操作的响应时间
- **错误诊断** - 结构化日志便于快速定位问题
- **审计日志** - 记录业务操作的完整链路

### 1.3 时间计划

- **计划完成**: Week 3 Day 2 (Day 13)
- **交付周期**: 1 天
- **负责人**: 基础设施团队

---

## 2. 需求分析

### 2.1 功能需求

#### 需求 1: Logger 接口定义

```go
// Logger 定义日志系统的标准接口
type Logger interface {
    // Debug 级别日志
    Debug(msg string)
    Debugf(format string, args ...interface{})

    // Info 级别日志
    Info(msg string)
    Infof(format string, args ...interface{})

    // Warn 级别日志
    Warn(msg string)
    Warnf(format string, args ...interface{})

    // Error 级别日志
    Error(msg string)
    Errorf(format string, args ...interface{})

    // WithFields 添加结构化字段
    WithFields(fields map[string]interface{}) Logger
}
```

#### 需求 2: 结构化日志输出

日志输出为 JSON 格式，包含：
- timestamp：日志时间
- level：日志级别
- message：日志消息
- fields：自定义字段
- caller：调用位置（文件:行号）

示例：
```json
{
  "timestamp": "2025-11-04T10:30:45.123Z",
  "level": "INFO",
  "message": "organization created",
  "fields": {
    "organizationID": "org-123",
    "module": "organization"
  },
  "caller": "organization/service.go:42"
}
```

#### 需求 3: 日志级别控制

```go
// LogLevel 定义日志级别
const (
    DebugLevel LogLevel = iota
    InfoLevel
    WarnLevel
    ErrorLevel
)
```

应该支持通过环境变量 `LOG_LEVEL` 动态设置。

### 2.2 非功能需求

| 需求 | 标准 | 说明 |
|------|------|------|
| **性能** | < 1ms 日志写入 | P99 延迟 |
| **可观测性** | 与 Prometheus 集成 | 暴露日志相关指标 |
| **测试覆盖率** | > 80% | 单元测试 |

---

## 3. 架构设计

### 3.1 模块结构

```
pkg/logger/
├── logger.go         # 接口定义和实现
├── formatter.go      # 日志格式化
├── metrics.go        # Prometheus 指标
├── logger_test.go    # 单元测试
└── README.md         # 使用说明
```

### 3.2 关键设计决策

**决策 1**: 为什么不使用第三方库（如 logrus、zap）

虽然这些库功能强大，但为了：
- 降低依赖复杂度
- 更好地控制日志格式
- 与项目的内部标准对齐
- 易于定制和扩展

Phase2 实现简单的内部日志系统，未来可升级到 zap。

---

## 4. 详细实现

### 4.1 logger.go - 日志系统实现

```go
package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// LogLevel 定义日志级别
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// levelNames 日志级别名称映射
var levelNames = map[LogLevel]string{
	DebugLevel: "DEBUG",
	InfoLevel:  "INFO",
	WarnLevel:  "WARN",
	ErrorLevel: "ERROR",
}

// Logger 定义日志系统的标准接口
type Logger interface {
	Debug(msg string)
	Debugf(format string, args ...interface{})

	Info(msg string)
	Infof(format string, args ...interface{})

	Warn(msg string)
	Warnf(format string, args ...interface{})

	Error(msg string)
	Errorf(format string, args ...interface{})

	WithFields(fields map[string]interface{}) Logger
}

// LogEntry 定义日志条目
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Caller    string                 `json:"caller"`
}

// StandardLogger 实现 Logger 接口
type StandardLogger struct {
	level      LogLevel
	writer     io.Writer
	fields     map[string]interface{}
	mu         sync.Mutex
	metricsCollector MetricsCollector
}

// MetricsCollector 用于收集日志指标
type MetricsCollector interface {
	RecordLog(level LogLevel, duration int64)
}

// NewLogger 创建新的日志记录器
func NewLogger() *StandardLogger {
	level := parseLoglevel(os.Getenv("LOG_LEVEL"))
	return &StandardLogger{
		level:  level,
		writer: os.Stdout,
		fields: make(map[string]interface{}),
	}
}

// parseLoglevel 从字符串解析日志级别
func parseLoglevel(levelStr string) LogLevel {
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		return DebugLevel
	case "WARN":
		return WarnLevel
	case "ERROR":
		return ErrorLevel
	default:
		return InfoLevel
	}
}

// WithFields 添加结构化字段
func (l *StandardLogger) WithFields(fields map[string]interface{}) Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newFields := make(map[string]interface{})
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}

	return &StandardLogger{
		level:      l.level,
		writer:     l.writer,
		fields:     newFields,
		metricsCollector: l.metricsCollector,
	}
}

// 内部日志记录方法
func (l *StandardLogger) log(level LogLevel, msg string) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	caller := getCaller(3)
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     levelNames[level],
		Message:   msg,
		Fields:    l.fields,
		Caller:    caller,
	}

	jsonData, _ := json.Marshal(entry)
	fmt.Fprintln(l.writer, string(jsonData))

	// 记录指标
	if l.metricsCollector != nil {
		l.metricsCollector.RecordLog(level, 0)
	}
}

// Debug 级别日志
func (l *StandardLogger) Debug(msg string) {
	l.log(DebugLevel, msg)
}

func (l *StandardLogger) Debugf(format string, args ...interface{}) {
	l.log(DebugLevel, fmt.Sprintf(format, args...))
}

// Info 级别日志
func (l *StandardLogger) Info(msg string) {
	l.log(InfoLevel, msg)
}

func (l *StandardLogger) Infof(format string, args ...interface{}) {
	l.log(InfoLevel, fmt.Sprintf(format, args...))
}

// Warn 级别日志
func (l *StandardLogger) Warn(msg string) {
	l.log(WarnLevel, msg)
}

func (l *StandardLogger) Warnf(format string, args ...interface{}) {
	l.log(WarnLevel, fmt.Sprintf(format, args...))
}

// Error 级别日志
func (l *StandardLogger) Error(msg string) {
	l.log(ErrorLevel, msg)
}

func (l *StandardLogger) Errorf(format string, args ...interface{}) {
	l.log(ErrorLevel, fmt.Sprintf(format, args...))
}

// getCaller 获取调用位置信息
func getCaller(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}

	// 只保留相对路径
	file = filepath.Base(filepath.Dir(file)) + "/" + filepath.Base(file)

	return fmt.Sprintf("%s:%d", file, line)
}

// NoopLogger 实现 Logger 接口，但不输出任何内容（用于测试）
type NoopLogger struct{}

func (n *NoopLogger) Debug(msg string)                          {}
func (n *NoopLogger) Debugf(format string, args ...interface{}) {}
func (n *NoopLogger) Info(msg string)                           {}
func (n *NoopLogger) Infof(format string, args ...interface{})  {}
func (n *NoopLogger) Warn(msg string)                           {}
func (n *NoopLogger) Warnf(format string, args ...interface{})  {}
func (n *NoopLogger) Error(msg string)                          {}
func (n *NoopLogger) Errorf(format string, args ...interface{}) {}
func (n *NoopLogger) WithFields(fields map[string]interface{}) Logger {
	return n
}

// NewNoopLogger 创建 noop logger（用于测试）
func NewNoopLogger() Logger {
	return &NoopLogger{}
}
```

### 4.2 metrics.go - Prometheus 指标

```go
package logger

import "github.com/prometheus/client_golang/prometheus"

var (
	// 按级别统计的日志计数
	logsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "logs_total",
			Help: "Total number of logs recorded by level",
		},
		[]string{"level"},
	)

	// 日志写入延迟
	logWriteDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "log_write_duration_seconds",
			Help:    "Log write duration in seconds",
			Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01},
		},
		[]string{"level"},
	)
)

// PrometheusMetricsCollector 实现 MetricsCollector 接口
type PrometheusMetricsCollector struct{}

func (p *PrometheusMetricsCollector) RecordLog(level LogLevel, duration int64) {
	levelStr := levelNames[level]
	logsTotal.WithLabelValues(levelStr).Inc()

	if duration > 0 {
		logWriteDuration.WithLabelValues(levelStr).Observe(float64(duration) / 1e9)
	}
}

// RegisterMetrics 注册 Prometheus 指标
func RegisterMetrics() {
	prometheus.MustRegister(logsTotal)
	prometheus.MustRegister(logWriteDuration)
}
```

---

## 5. 单元测试

### 5.1 logger_test.go - 测试套件

```go
package logger

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestLogDebug(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := &StandardLogger{
		level:  DebugLevel,
		writer: buf,
		fields: make(map[string]interface{}),
	}

	logger.Debug("test debug message")

	var entry LogEntry
	json.Unmarshal(buf.Bytes(), &entry)

	if entry.Level != "DEBUG" {
		t.Errorf("expected DEBUG, got %s", entry.Level)
	}
	if entry.Message != "test debug message" {
		t.Errorf("expected 'test debug message', got %s", entry.Message)
	}
}

func TestLogWithFields(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := &StandardLogger{
		level:  InfoLevel,
		writer: buf,
		fields: make(map[string]interface{}),
	}

	newLogger := logger.WithFields(map[string]interface{}{
		"userID":   "user-123",
		"module":   "organization",
	})

	newLogger.Info("test message")

	var entry LogEntry
	json.Unmarshal(buf.Bytes(), &entry)

	if userID, ok := entry.Fields["userID"]; !ok || userID != "user-123" {
		t.Error("field userID not found or incorrect")
	}
}

func TestLogLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := &StandardLogger{
		level:  WarnLevel,
		writer: buf,
		fields: make(map[string]interface{}),
	}

	logger.Debug("should not appear")
	logger.Info("should not appear")
	logger.Warn("should appear")

	if !strings.Contains(buf.String(), "should appear") {
		t.Error("warn message not recorded")
	}
	if strings.Contains(buf.String(), "should not appear") {
		t.Error("debug/info messages should not be recorded")
	}
}

func TestLogFormatting(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := &StandardLogger{
		level:  InfoLevel,
		writer: buf,
		fields: make(map[string]interface{}),
	}

	logger.Infof("user %s created with status %s", "john", "active")

	var entry LogEntry
	json.Unmarshal(buf.Bytes(), &entry)

	if entry.Message != "user john created with status active" {
		t.Errorf("formatting failed: %s", entry.Message)
	}
}
```

---

## 6. 验收标准

### 6.1 功能验收

- [ ] Logger 接口定义完整
- [ ] 结构化日志输出为 JSON
- [ ] 支持五个日志级别（Debug, Info, Warn, Error）
- [ ] 字段添加正常工作
- [ ] 调用位置信息准确
- [ ] NoopLogger 实现正确

### 6.2 质量验收

- [ ] 单元测试覆盖率 > 80%
- [ ] 所有测试通过
- [ ] 代码通过 `go fmt` 检查
- [ ] 代码通过 `go vet` 检查

### 6.3 集成验收

- [ ] 可在 Plan 216 (eventbus) 中使用
- [ ] 可在 Plan 217 (database) 中使用
- [ ] 支持 Prometheus 指标记录
- [ ] 日志性能达标（< 1ms）

---

## 7. 交付物清单

- ✅ `pkg/logger/logger.go`
- ✅ `pkg/logger/metrics.go`
- ✅ `pkg/logger/logger_test.go`
- ✅ `pkg/logger/README.md`
- ✅ 本计划文档（218）

---

**维护者**: Codex（AI 助手）
**最后更新**: 2025-11-04
**计划完成日期**: Week 3 Day 2 (Day 13)
