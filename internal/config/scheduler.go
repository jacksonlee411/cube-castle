package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	cron "github.com/robfig/cron/v3"
	"gopkg.in/yaml.v3"
)

// SchedulerConfigResult wraps the resolved configuration together with provenance metadata.
type SchedulerConfigResult struct {
	Config   *SchedulerConfig
	Metadata SchedulerConfigMetadata
}

// SchedulerConfigMetadata describes how the final configuration was produced.
type SchedulerConfigMetadata struct {
	Sources         []string          // e.g. ["defaults", "config/scheduler.yaml", "env:SCHEDULER_ENABLED"]
	Overrides       map[string]string // env var -> value applied
	ConfigFile      string            // resolved config file path (empty if none)
	ValidationError error             // non-nil when validation fails
}

// SchedulerConfig captures all tunable values for scheduler & temporal orchestration.
type SchedulerConfig struct {
	Enabled  bool
	Temporal TemporalSettings
	Worker   WorkerSettings
	Retry    RetryPolicy
	Cron     CronSettings
	Monitor  MonitorSettings
	Scripts  ScriptsSettings
}

// TemporalSettings describes Temporal/queue integration parameters.
type TemporalSettings struct {
	Endpoint  string
	Namespace string
	TaskQueue string
}

// WorkerSettings controls worker concurrency for Temporal task processing.
type WorkerSettings struct {
	Concurrency int
	PollerCount int
}

// RetryPolicy defines Temporal activity retry behaviour.
type RetryPolicy struct {
	MaxAttempts        int
	InitialInterval    time.Duration
	BackoffCoefficient float64
	MaxInterval        time.Duration
}

// CronSettings holds definitions for scheduled maintenance tasks.
type CronSettings struct {
	CheckInterval time.Duration
	Tasks         map[string]CronDefinition
}

// CronDefinition describes an individual scheduled task.
type CronDefinition struct {
	Name         string
	Description  string
	CronExpr     string
	Enabled      bool
	Script       string
	InitialDelay time.Duration
	Timeout      time.Duration
}

// MonitorSettings governs temporal monitoring routines.
type MonitorSettings struct {
	Enabled       bool
	CheckInterval time.Duration
	AlertRules    []AlertRuleConfig
}

// AlertRuleConfig describes alert thresholds for monitoring.
type AlertRuleConfig struct {
	Name        string
	Description string
	Threshold   int
	Level       string
}

// ScriptsSettings resolves the base directory for SQL/maintenance scripts.
type ScriptsSettings struct {
	Root string
}

var (
	schedulerConfigOnce sync.Once
	schedulerConfig     SchedulerConfigResult
)

// GetSchedulerConfig loads configuration from defaults, YAML (if present) and environment variables.
func GetSchedulerConfig() SchedulerConfigResult {
	schedulerConfigOnce.Do(func() {
		cfg := defaultSchedulerConfig()
		meta := SchedulerConfigMetadata{
			Sources:   []string{"defaults"},
			Overrides: map[string]string{},
		}

		configFile := resolveSchedulerConfigFile()
		if configFile != "" {
			if err := applySchedulerConfigFile(cfg, configFile); err == nil {
				meta.Sources = append(meta.Sources, configFile)
				meta.ConfigFile = configFile
			} else {
				meta.Sources = append(meta.Sources, fmt.Sprintf("%s (error: %v)", configFile, err))
			}
		}

		applySchedulerEnvOverrides(cfg, &meta)

		if err := ValidateSchedulerConfig(cfg); err != nil {
			meta.ValidationError = err
		}

		schedulerConfig = SchedulerConfigResult{
			Config:   cfg,
			Metadata: meta,
		}
	})

	return schedulerConfig
}

func defaultSchedulerConfig() *SchedulerConfig {
	return &SchedulerConfig{
		Enabled: false,
		Temporal: TemporalSettings{
			Endpoint:  "",
			Namespace: "",
			TaskQueue: "",
		},
		Worker: WorkerSettings{
			Concurrency: 4,
			PollerCount: 2,
		},
		Retry: RetryPolicy{
			MaxAttempts:        3,
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaxInterval:        time.Minute,
		},
		Cron: CronSettings{
			CheckInterval: time.Minute,
			Tasks: map[string]CronDefinition{
				"daily_cutover": {
					Name:        "daily_cutover",
					Description: "每日cutover任务：维护时态数据一致性",
					CronExpr:    "0 2 * * *",
					Enabled:     true,
					Script:      "daily-cutover.sql",
				},
				"acting_assignment_auto_revert": {
					Name:        "acting_assignment_auto_revert",
					Description: "自动结束到期的代理任职",
					CronExpr:    "15 2 * * *",
					Enabled:     true,
				},
				"data_consistency_check": {
					Name:        "data_consistency_check",
					Description: "数据一致性检查",
					CronExpr:    "30 2 * * *",
					Enabled:     true,
					Script:      "data-consistency-check.sql",
				},
				"system_monitoring": {
					Name:        "system_monitoring",
					Description: "系统健康监控",
					CronExpr:    "0 * * * *",
					Enabled:     true,
				},
			},
		},
		Monitor: MonitorSettings{
			Enabled:       true,
			CheckInterval: 5 * time.Minute,
			AlertRules: []AlertRuleConfig{
				{
					Name:        "DUPLICATE_CURRENT_RECORDS",
					Description: "重复的当前记录数量超过阈值",
					Threshold:   0,
					Level:       "CRITICAL",
				},
				{
					Name:        "MISSING_CURRENT_RECORDS",
					Description: "缺失当前记录的组织数量超过阈值",
					Threshold:   0,
					Level:       "CRITICAL",
				},
				{
					Name:        "TIMELINE_OVERLAPS",
					Description: "时间线重叠记录数量超过阈值",
					Threshold:   0,
					Level:       "CRITICAL",
				},
				{
					Name:        "INCONSISTENT_FLAGS",
					Description: "is_current/is_future标志不一致记录数量超过阈值",
					Threshold:   5,
					Level:       "WARNING",
				},
				{
					Name:        "ORPHAN_RECORDS",
					Description: "孤立记录（父级不存在）数量超过阈值",
					Threshold:   10,
					Level:       "WARNING",
				},
				{
					Name:        "HEALTH_SCORE",
					Description: "系统健康分数低于阈值",
					Threshold:   85,
					Level:       "WARNING",
				},
			},
		},
		Scripts: ScriptsSettings{
			Root: "./scripts",
		},
	}
}

func resolveSchedulerConfigFile() string {
	if v := strings.TrimSpace(os.Getenv("SCHEDULER_CONFIG_FILE")); v != "" {
		return v
	}
	defaultPath := filepath.Join("config", "scheduler.yaml")
	if _, err := os.Stat(defaultPath); err == nil {
		return defaultPath
	}
	return ""
}

type schedulerYAML struct {
	Enabled  *bool `yaml:"enabled"`
	Temporal struct {
		Endpoint  string `yaml:"endpoint"`
		Namespace string `yaml:"namespace"`
		TaskQueue string `yaml:"taskQueue"`
	} `yaml:"temporal"`
	Worker struct {
		Concurrency *int `yaml:"concurrency"`
		PollerCount *int `yaml:"pollerCount"`
	} `yaml:"worker"`
	Retry struct {
		MaxAttempts        *int     `yaml:"maxAttempts"`
		InitialInterval    string   `yaml:"initialInterval"`
		BackoffCoefficient *float64 `yaml:"backoffCoefficient"`
		MaxInterval        string   `yaml:"maxInterval"`
	} `yaml:"retry"`
	Cron struct {
		CheckInterval string `yaml:"checkInterval"`
		Tasks         map[string]struct {
			Description  string `yaml:"description"`
			CronExpr     string `yaml:"cron"`
			Enabled      *bool  `yaml:"enabled"`
			Script       string `yaml:"script"`
			InitialDelay string `yaml:"initialDelay"`
			Timeout      string `yaml:"timeout"`
		} `yaml:"tasks"`
	} `yaml:"cron"`
	Monitor struct {
		Enabled       *bool  `yaml:"enabled"`
		CheckInterval string `yaml:"checkInterval"`
		AlertRules    []struct {
			Name        string `yaml:"name"`
			Description string `yaml:"description"`
			Threshold   *int   `yaml:"threshold"`
			Level       string `yaml:"level"`
		} `yaml:"alertRules"`
	} `yaml:"monitor"`
	Scripts struct {
		Root string `yaml:"root"`
	} `yaml:"scripts"`
}

func applySchedulerConfigFile(cfg *SchedulerConfig, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var raw schedulerYAML
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return err
	}

	if raw.Enabled != nil {
		cfg.Enabled = *raw.Enabled
	}
	if raw.Temporal.Endpoint != "" {
		cfg.Temporal.Endpoint = raw.Temporal.Endpoint
	}
	if raw.Temporal.Namespace != "" {
		cfg.Temporal.Namespace = raw.Temporal.Namespace
	}
	if raw.Temporal.TaskQueue != "" {
		cfg.Temporal.TaskQueue = raw.Temporal.TaskQueue
	}
	if raw.Worker.Concurrency != nil {
		cfg.Worker.Concurrency = *raw.Worker.Concurrency
	}
	if raw.Worker.PollerCount != nil {
		cfg.Worker.PollerCount = *raw.Worker.PollerCount
	}
	if raw.Retry.MaxAttempts != nil {
		cfg.Retry.MaxAttempts = *raw.Retry.MaxAttempts
	}
	if raw.Retry.InitialInterval != "" {
		if d, err := time.ParseDuration(raw.Retry.InitialInterval); err == nil {
			cfg.Retry.InitialInterval = d
		}
	}
	if raw.Retry.BackoffCoefficient != nil {
		cfg.Retry.BackoffCoefficient = *raw.Retry.BackoffCoefficient
	}
	if raw.Retry.MaxInterval != "" {
		if d, err := time.ParseDuration(raw.Retry.MaxInterval); err == nil {
			cfg.Retry.MaxInterval = d
		}
	}

	if raw.Cron.CheckInterval != "" {
		if d, err := time.ParseDuration(raw.Cron.CheckInterval); err == nil {
			cfg.Cron.CheckInterval = d
		}
	}
	if raw.Cron.Tasks != nil {
		cfg.Cron.Tasks = map[string]CronDefinition{}
		for name, task := range raw.Cron.Tasks {
			def := CronDefinition{
				Name:        name,
				Description: task.Description,
				CronExpr:    task.CronExpr,
				Script:      task.Script,
			}
			if task.Enabled != nil {
				def.Enabled = *task.Enabled
			} else {
				def.Enabled = true
			}
			if task.InitialDelay != "" {
				if d, err := time.ParseDuration(task.InitialDelay); err == nil {
					def.InitialDelay = d
				}
			}
			if task.Timeout != "" {
				if d, err := time.ParseDuration(task.Timeout); err == nil {
					def.Timeout = d
				}
			}
			cfg.Cron.Tasks[name] = def
		}
	}

	if raw.Monitor.Enabled != nil {
		cfg.Monitor.Enabled = *raw.Monitor.Enabled
	}
	if raw.Monitor.CheckInterval != "" {
		if d, err := time.ParseDuration(raw.Monitor.CheckInterval); err == nil {
			cfg.Monitor.CheckInterval = d
		}
	}
	if len(raw.Monitor.AlertRules) > 0 {
		cfg.Monitor.AlertRules = make([]AlertRuleConfig, 0, len(raw.Monitor.AlertRules))
		for _, item := range raw.Monitor.AlertRules {
			rule := AlertRuleConfig{
				Name:        item.Name,
				Description: item.Description,
				Level:       item.Level,
			}
			if item.Threshold != nil {
				rule.Threshold = *item.Threshold
			}
			cfg.Monitor.AlertRules = append(cfg.Monitor.AlertRules, rule)
		}
	}

	if raw.Scripts.Root != "" {
		cfg.Scripts.Root = raw.Scripts.Root
	}

	return nil
}

func applySchedulerEnvOverrides(cfg *SchedulerConfig, meta *SchedulerConfigMetadata) {
	if set, ok := lookupEnvBool("SCHEDULER_ENABLED"); ok {
		cfg.Enabled = set
		meta.Sources = append(meta.Sources, "env:SCHEDULER_ENABLED")
		meta.Overrides["SCHEDULER_ENABLED"] = strconv.FormatBool(set)
	}
	if str, ok := os.LookupEnv("SCHEDULER_TEMPORAL_ENDPOINT"); ok {
		cfg.Temporal.Endpoint = strings.TrimSpace(str)
		meta.Sources = append(meta.Sources, "env:SCHEDULER_TEMPORAL_ENDPOINT")
		meta.Overrides["SCHEDULER_TEMPORAL_ENDPOINT"] = cfg.Temporal.Endpoint
	}
	if str, ok := os.LookupEnv("SCHEDULER_NAMESPACE"); ok {
		cfg.Temporal.Namespace = strings.TrimSpace(str)
		meta.Sources = append(meta.Sources, "env:SCHEDULER_NAMESPACE")
		meta.Overrides["SCHEDULER_NAMESPACE"] = cfg.Temporal.Namespace
	}
	if str, ok := os.LookupEnv("SCHEDULER_TASK_QUEUE"); ok {
		cfg.Temporal.TaskQueue = strings.TrimSpace(str)
		meta.Sources = append(meta.Sources, "env:SCHEDULER_TASK_QUEUE")
		meta.Overrides["SCHEDULER_TASK_QUEUE"] = cfg.Temporal.TaskQueue
	}
	if v, ok := lookupEnvInt("SCHEDULER_WORKER_CONCURRENCY"); ok {
		cfg.Worker.Concurrency = v
		meta.Sources = append(meta.Sources, "env:SCHEDULER_WORKER_CONCURRENCY")
		meta.Overrides["SCHEDULER_WORKER_CONCURRENCY"] = strconv.Itoa(v)
	}
	if v, ok := lookupEnvInt("SCHEDULER_WORKER_POLLER_COUNT"); ok {
		cfg.Worker.PollerCount = v
		meta.Sources = append(meta.Sources, "env:SCHEDULER_WORKER_POLLER_COUNT")
		meta.Overrides["SCHEDULER_WORKER_POLLER_COUNT"] = strconv.Itoa(v)
	}
	if v, ok := lookupEnvInt("SCHEDULER_RETRY_MAX_ATTEMPTS"); ok {
		cfg.Retry.MaxAttempts = v
		meta.Sources = append(meta.Sources, "env:SCHEDULER_RETRY_MAX_ATTEMPTS")
		meta.Overrides["SCHEDULER_RETRY_MAX_ATTEMPTS"] = strconv.Itoa(v)
	}
	if d, ok := lookupEnvDuration("SCHEDULER_RETRY_INITIAL_INTERVAL"); ok {
		cfg.Retry.InitialInterval = d
		meta.Sources = append(meta.Sources, "env:SCHEDULER_RETRY_INITIAL_INTERVAL")
		meta.Overrides["SCHEDULER_RETRY_INITIAL_INTERVAL"] = d.String()
	}
	if f, ok := lookupEnvFloat("SCHEDULER_RETRY_BACKOFF_COEFFICIENT"); ok {
		cfg.Retry.BackoffCoefficient = f
		meta.Sources = append(meta.Sources, "env:SCHEDULER_RETRY_BACKOFF_COEFFICIENT")
		meta.Overrides["SCHEDULER_RETRY_BACKOFF_COEFFICIENT"] = strconv.FormatFloat(f, 'f', -1, 64)
	}
	if d, ok := lookupEnvDuration("SCHEDULER_RETRY_MAX_INTERVAL"); ok {
		cfg.Retry.MaxInterval = d
		meta.Sources = append(meta.Sources, "env:SCHEDULER_RETRY_MAX_INTERVAL")
		meta.Overrides["SCHEDULER_RETRY_MAX_INTERVAL"] = d.String()
	}
	if d, ok := lookupEnvDuration("SCHEDULER_CRON_CHECK_INTERVAL"); ok {
		cfg.Cron.CheckInterval = d
		meta.Sources = append(meta.Sources, "env:SCHEDULER_CRON_CHECK_INTERVAL")
		meta.Overrides["SCHEDULER_CRON_CHECK_INTERVAL"] = d.String()
	}
	if set, ok := lookupEnvBool("SCHEDULER_MONITOR_ENABLED"); ok {
		cfg.Monitor.Enabled = set
		meta.Sources = append(meta.Sources, "env:SCHEDULER_MONITOR_ENABLED")
		meta.Overrides["SCHEDULER_MONITOR_ENABLED"] = strconv.FormatBool(set)
	}
	if d, ok := lookupEnvDuration("SCHEDULER_MONITOR_CHECK_INTERVAL"); ok {
		cfg.Monitor.CheckInterval = d
		meta.Sources = append(meta.Sources, "env:SCHEDULER_MONITOR_CHECK_INTERVAL")
		meta.Overrides["SCHEDULER_MONITOR_CHECK_INTERVAL"] = d.String()
	}
	if root, ok := os.LookupEnv("SCHEDULER_SCRIPTS_ROOT"); ok {
		cfg.Scripts.Root = strings.TrimSpace(root)
		meta.Sources = append(meta.Sources, "env:SCHEDULER_SCRIPTS_ROOT")
		meta.Overrides["SCHEDULER_SCRIPTS_ROOT"] = cfg.Scripts.Root
	}

	for name, task := range cfg.Cron.Tasks {
		envPrefix := fmt.Sprintf("SCHEDULER_TASK_%s_", strings.ToUpper(strings.ReplaceAll(name, "-", "_")))

		if str, ok := os.LookupEnv(envPrefix + "CRON"); ok {
			task.CronExpr = strings.TrimSpace(str)
			meta.Sources = append(meta.Sources, "env:"+envPrefix+"CRON")
			meta.Overrides[envPrefix+"CRON"] = task.CronExpr
		}
		if str, ok := os.LookupEnv(envPrefix + "DESCRIPTION"); ok {
			task.Description = strings.TrimSpace(str)
			meta.Sources = append(meta.Sources, "env:"+envPrefix+"DESCRIPTION")
			meta.Overrides[envPrefix+"DESCRIPTION"] = task.Description
		}
		if str, ok := os.LookupEnv(envPrefix + "SCRIPT"); ok {
			task.Script = strings.TrimSpace(str)
			meta.Sources = append(meta.Sources, "env:"+envPrefix+"SCRIPT")
			meta.Overrides[envPrefix+"SCRIPT"] = task.Script
		}
		if val, ok := lookupEnvBool(envPrefix + "ENABLED"); ok {
			task.Enabled = val
			meta.Sources = append(meta.Sources, "env:"+envPrefix+"ENABLED")
			meta.Overrides[envPrefix+"ENABLED"] = strconv.FormatBool(val)
		}
		if d, ok := lookupEnvDuration(envPrefix + "INITIAL_DELAY"); ok {
			task.InitialDelay = d
			meta.Sources = append(meta.Sources, "env:"+envPrefix+"INITIAL_DELAY")
			meta.Overrides[envPrefix+"INITIAL_DELAY"] = d.String()
		}
		if d, ok := lookupEnvDuration(envPrefix + "TIMEOUT"); ok {
			task.Timeout = d
			meta.Sources = append(meta.Sources, "env:"+envPrefix+"TIMEOUT")
			meta.Overrides[envPrefix+"TIMEOUT"] = d.String()
		}

		cfg.Cron.Tasks[name] = task
	}
}

func lookupEnvBool(key string) (bool, bool) {
	val, ok := os.LookupEnv(key)
	if !ok {
		return false, false
	}
	val = strings.TrimSpace(strings.ToLower(val))
	switch val {
	case "1", "true", "yes", "on":
		return true, true
	case "0", "false", "no", "off":
		return false, true
	default:
		return false, false
	}
}

func lookupEnvInt(key string) (int, bool) {
	val, ok := os.LookupEnv(key)
	if !ok {
		return 0, false
	}
	v, err := strconv.Atoi(strings.TrimSpace(val))
	if err != nil {
		return 0, false
	}
	return v, true
}

func lookupEnvFloat(key string) (float64, bool) {
	val, ok := os.LookupEnv(key)
	if !ok {
		return 0, false
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func lookupEnvDuration(key string) (time.Duration, bool) {
	val, ok := os.LookupEnv(key)
	if !ok {
		return 0, false
	}
	d, err := time.ParseDuration(strings.TrimSpace(val))
	if err != nil {
		return 0, false
	}
	return d, true
}

// ResetSchedulerConfig is exposed for testability, allowing config reload between tests.
func ResetSchedulerConfig() {
	schedulerConfigOnce = sync.Once{}
	schedulerConfig = SchedulerConfigResult{}
}

// ValidateSchedulerConfig performs a light structural check; full validation lives in scheduler_validator.go.
func ValidateSchedulerConfig(cfg *SchedulerConfig) error {
	if cfg == nil {
		return errors.New("scheduler config cannot be nil")
	}
	var validationErrors []string

	// 工作流引擎已清退：不再强制校验 Temporal 相关字段
	if cfg.Worker.Concurrency <= 0 {
		validationErrors = append(validationErrors, "worker concurrency must be positive")
	}
	if cfg.Worker.PollerCount <= 0 {
		validationErrors = append(validationErrors, "worker poller count must be positive")
	}
	if cfg.Retry.MaxAttempts <= 0 {
		validationErrors = append(validationErrors, "retry maxAttempts must be positive")
	}
	if cfg.Retry.InitialInterval <= 0 {
		validationErrors = append(validationErrors, "retry initialInterval must be positive")
	}
	if cfg.Retry.BackoffCoefficient < 1.0 {
		validationErrors = append(validationErrors, "retry backoffCoefficient must be >= 1.0")
	}
	if cfg.Retry.MaxInterval < cfg.Retry.InitialInterval {
		validationErrors = append(validationErrors, "retry maxInterval must be greater than or equal to initialInterval")
	}
	if cfg.Cron.CheckInterval <= 0 {
		validationErrors = append(validationErrors, "cron check interval must be positive")
	}
	if cfg.Monitor.CheckInterval <= 0 {
		validationErrors = append(validationErrors, "monitor check interval must be positive")
	}
	if cfg.Scripts.Root == "" {
		validationErrors = append(validationErrors, "scripts root must not be empty")
	}

	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	for name, task := range cfg.Cron.Tasks {
		if strings.TrimSpace(task.CronExpr) == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("task %s cron expression must not be empty", name))
			continue
		}
		if _, err := parser.Parse(task.CronExpr); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("task %s has invalid cron expression: %v", name, err))
		}
		if task.Enabled && task.CronExpr == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("task %s enabled but missing cron expression", name))
		}
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("scheduler config validation failed: %s", strings.Join(validationErrors, "; "))
	}
	return nil
}
