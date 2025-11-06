package config

import (
	"testing"
	"time"
)

func TestValidateSchedulerConfigValid(t *testing.T) {
	cfg := &SchedulerConfig{
		Enabled: true,
		Temporal: TemporalSettings{
			Endpoint:  "temporal:7233",
			Namespace: "cube-castle",
			TaskQueue: "organization-maintenance",
		},
		Worker: WorkerSettings{
			Concurrency: 2,
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
				"sample_task": {
					Name:     "sample_task",
					CronExpr: "0 * * * *",
					Enabled:  true,
				},
			},
		},
		Monitor: MonitorSettings{
			Enabled:       true,
			CheckInterval: time.Minute,
		},
		Scripts: ScriptsSettings{
			Root: "./scripts",
		},
	}

	if err := ValidateSchedulerConfig(cfg); err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}
}

func TestValidateSchedulerConfigInvalidCron(t *testing.T) {
	cfg := &SchedulerConfig{
		Enabled: true,
		Temporal: TemporalSettings{
			Endpoint:  "temporal:7233",
			Namespace: "cube-castle",
			TaskQueue: "organization-maintenance",
		},
		Worker: WorkerSettings{
			Concurrency: 1,
			PollerCount: 1,
		},
		Retry: RetryPolicy{
			MaxAttempts:        1,
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaxInterval:        time.Minute,
		},
		Cron: CronSettings{
			CheckInterval: time.Minute,
			Tasks: map[string]CronDefinition{
				"broken": {
					Name:     "broken",
					CronExpr: "invalid cron",
					Enabled:  true,
				},
			},
		},
		Monitor: MonitorSettings{
			Enabled:       true,
			CheckInterval: time.Minute,
		},
		Scripts: ScriptsSettings{
			Root: "./scripts",
		},
	}

	if err := ValidateSchedulerConfig(cfg); err == nil {
		t.Fatalf("expected error for invalid cron expression")
	}
}

func TestGetSchedulerConfigEnvOverride(t *testing.T) {
	ResetSchedulerConfig()
	t.Cleanup(ResetSchedulerConfig)

	t.Setenv("SCHEDULER_ENABLED", "false")
	t.Setenv("SCHEDULER_TASK_DAILY_CUTOVER_CRON", "*/5 * * * *")

	result := GetSchedulerConfig()
	if result.Config == nil {
		t.Fatal("expected config to be loaded")
	}
	if result.Config.Enabled {
		t.Fatalf("expected scheduler to be disabled via env override")
	}
	task, ok := result.Config.Cron.Tasks["daily_cutover"]
	if !ok {
		t.Fatalf("expected daily_cutover task to exist")
	}
	if task.CronExpr != "*/5 * * * *" {
		t.Fatalf("expected cron override, got %s", task.CronExpr)
	}
	if result.Metadata.ValidationError != nil {
		t.Fatalf("expected config to be valid, got %v", result.Metadata.ValidationError)
	}
}
