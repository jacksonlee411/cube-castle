package outbox

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config 收敛 dispatcher 的运行参数，并统一处理环境变量覆盖逻辑。
type Config struct {
	PollInterval    time.Duration
	BatchSize       int
	MaxRetry        int
	BackoffBase     time.Duration
	MetricNamespace string
}

const (
	defaultInterval    = 5 * time.Second
	defaultBatchSize   = 50
	defaultMaxRetry    = 10
	defaultBackoffBase = 5 * time.Second
	defaultMetricNS    = "outbox_dispatch"
)

// LoadConfig 读取环境变量并填充默认值；若存在非法配置则返回错误。
func LoadConfig() (Config, error) {
	cfg := Config{
		PollInterval:    defaultInterval,
		BatchSize:       defaultBatchSize,
		MaxRetry:        defaultMaxRetry,
		BackoffBase:     defaultBackoffBase,
		MetricNamespace: defaultMetricNS,
	}

	var err error
	if s := os.Getenv("OUTBOX_DISPATCH_INTERVAL"); s != "" {
		if cfg.PollInterval, err = time.ParseDuration(s); err != nil || cfg.PollInterval <= 0 {
			return Config{}, fmt.Errorf("invalid OUTBOX_DISPATCH_INTERVAL: %q", s)
		}
	}

	if s := os.Getenv("OUTBOX_DISPATCH_BATCH_SIZE"); s != "" {
		if cfg.BatchSize, err = strconv.Atoi(s); err != nil || cfg.BatchSize <= 0 {
			return Config{}, fmt.Errorf("invalid OUTBOX_DISPATCH_BATCH_SIZE: %q", s)
		}
	}

	if s := os.Getenv("OUTBOX_DISPATCH_MAX_RETRY"); s != "" {
		if cfg.MaxRetry, err = strconv.Atoi(s); err != nil || cfg.MaxRetry <= 0 {
			return Config{}, fmt.Errorf("invalid OUTBOX_DISPATCH_MAX_RETRY: %q", s)
		}
	}

	if s := os.Getenv("OUTBOX_DISPATCH_BACKOFF_BASE"); s != "" {
		if cfg.BackoffBase, err = time.ParseDuration(s); err != nil || cfg.BackoffBase <= 0 {
			return Config{}, fmt.Errorf("invalid OUTBOX_DISPATCH_BACKOFF_BASE: %q", s)
		}
	}

	if s := os.Getenv("OUTBOX_DISPATCH_METRIC_PREFIX"); s != "" {
		cfg.MetricNamespace = s
	}

	return cfg, nil
}
