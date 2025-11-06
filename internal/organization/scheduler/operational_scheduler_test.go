package scheduler

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	configpkg "cube-castle/internal/config"
	pkglogger "cube-castle/pkg/logger"
	"github.com/DATA-DOG/go-sqlmock"
)

func TestOperationalScheduler_RunTaskExecutesScript(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	tempDir := t.TempDir()
	scriptName := "test.sql"
	scriptPath := filepath.Join(tempDir, scriptName)
	if err := os.WriteFile(scriptPath, []byte("SELECT 1;"), 0o600); err != nil {
		t.Fatalf("write script file: %v", err)
	}

	cfg := &configpkg.SchedulerConfig{
		Enabled: true,
		Scripts: configpkg.ScriptsSettings{
			Root: tempDir,
		},
		Cron: configpkg.CronSettings{
			CheckInterval: 10 * time.Millisecond,
			Tasks: map[string]configpkg.CronDefinition{
				"test_task": {
					Name:        "test_task",
					Description: "integration smoke task",
					CronExpr:    "*/1 * * * *",
					Enabled:     true,
					Script:      scriptName,
				},
			},
		},
	}

	logger := pkglogger.NewLogger(pkglogger.WithWriter(io.Discard))
	s := NewOperationalScheduler(db, logger, nil, nil, cfg)

	mock.ExpectExec("SELECT 1;").WillReturnResult(sqlmock.NewResult(0, 0))

	if err := s.RunTask(context.Background(), "test_task"); err != nil {
		t.Fatalf("run task: %v", err)
	}

	task, ok := s.tasks["test_task"]
	if !ok {
		t.Fatalf("task not registered in scheduler map")
	}
	if task.Running {
		t.Fatalf("expected task to finish execution")
	}
	if task.LastRun == nil {
		t.Fatalf("expected task last run timestamp to be recorded")
	}
	if task.NextRun.Before(time.Now()) {
		t.Fatalf("expected next run to be scheduled in the future")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations not met: %v", err)
	}
}

func TestOperationalScheduler_RunTaskDisabledScheduler(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	cfg := &configpkg.SchedulerConfig{
		Enabled: false,
		Cron: configpkg.CronSettings{
			CheckInterval: time.Minute,
			Tasks: map[string]configpkg.CronDefinition{
				"noop": {
					Name:     "noop",
					CronExpr: "*/1 * * * *",
					Enabled:  true,
				},
			},
		},
	}

	logger := pkglogger.NewLogger(pkglogger.WithWriter(io.Discard))
	s := NewOperationalScheduler(db, logger, nil, nil, cfg)

	if err := s.RunTask(context.Background(), "noop"); err == nil {
		t.Fatalf("expected scheduler disabled error")
	}
}

func TestOperationalScheduler_RunTaskUnknown(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	cfg := &configpkg.SchedulerConfig{
		Enabled: true,
		Cron: configpkg.CronSettings{
			CheckInterval: time.Minute,
			Tasks:         map[string]configpkg.CronDefinition{},
		},
	}

	logger := pkglogger.NewLogger(pkglogger.WithWriter(io.Discard))
	s := NewOperationalScheduler(db, logger, nil, nil, cfg)

	if err := s.RunTask(context.Background(), "missing_task"); err == nil {
		t.Fatalf("expected error for unknown task")
	}
}
