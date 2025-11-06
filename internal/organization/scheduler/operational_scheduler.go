package scheduler

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	configpkg "cube-castle/internal/config"
	"cube-castle/internal/organization/service"
	"cube-castle/internal/types"
	pkglogger "cube-castle/pkg/logger"
	cron "github.com/robfig/cron/v3"
)

// OperationalScheduler 根据集中化配置调度后台维护任务。
type OperationalScheduler struct {
	db              *sql.DB
	logger          pkglogger.Logger
	monitor         *TemporalMonitor
	positions       *service.PositionService
	scriptsPath     string
	config          *configpkg.SchedulerConfig
	tasks           map[string]*ScheduledTask
	tickInterval    time.Duration
	stopCh          chan struct{}
	running         bool
	monitorEnabled  bool
	monitorInterval time.Duration
	mu              sync.RWMutex
}

// ScheduledTask 描述单个任务的运行时状态。
type ScheduledTask struct {
	Name             string        `json:"name"`
	Description      string        `json:"description"`
	CronExpr         string        `json:"cron"`
	Enabled          bool          `json:"enabled"`
	ScriptFile       string        `json:"scriptFile"`
	InitialDelay     time.Duration `json:"initialDelay"`
	Timeout          time.Duration `json:"timeout"`
	LastRun          *time.Time    `json:"lastRun,omitempty"`
	NextRun          time.Time     `json:"nextRun"`
	usesInitialDelay bool
	cronSchedule     cron.Schedule
	Running          bool `json:"running"`
	mu               sync.Mutex
}

// NewOperationalScheduler 创建运维任务调度器。
func NewOperationalScheduler(
	db *sql.DB,
	baseLogger pkglogger.Logger,
	monitor *TemporalMonitor,
	positions *service.PositionService,
	cfg *configpkg.SchedulerConfig,
) *OperationalScheduler {
	logger := scopedLogger(baseLogger, "operationalScheduler", nil)

	if cfg == nil {
		cfg = configpkg.GetSchedulerConfig().Config
	}

	scriptsPath := cfg.Scripts.Root
	if scriptsPath == "" {
		scriptsPath = "./scripts"
	}
	if !filepath.IsAbs(scriptsPath) {
		if pwd := os.Getenv("PWD"); pwd != "" {
			scriptsPath = filepath.Join(pwd, scriptsPath)
		}
	}

	tick := cfg.Cron.CheckInterval
	if tick <= 0 {
		tick = time.Minute
	}

	taskMap := buildScheduledTasks(cfg, logger)

	return &OperationalScheduler{
		db:              db,
		logger:          logger,
		monitor:         monitor,
		positions:       positions,
		scriptsPath:     scriptsPath,
		config:          cfg,
		tasks:           taskMap,
		tickInterval:    tick,
		stopCh:          make(chan struct{}),
		monitorEnabled:  cfg.Monitor.Enabled,
		monitorInterval: cfg.Monitor.CheckInterval,
	}
}

func buildScheduledTasks(cfg *configpkg.SchedulerConfig, logger pkglogger.Logger) map[string]*ScheduledTask {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	result := make(map[string]*ScheduledTask, len(cfg.Cron.Tasks))
	now := time.Now()

	for name, def := range cfg.Cron.Tasks {
		cronExpr := strings.TrimSpace(def.CronExpr)
		if cronExpr == "" {
			logger.WithFields(pkglogger.Fields{
				"task": name,
			}).Warn("跳过任务：缺少 cron 表达式")
			continue
		}

		schedule, err := parser.Parse(cronExpr)
		if err != nil {
			logger.WithFields(pkglogger.Fields{
				"task": name,
				"cron": cronExpr,
				"err":  err,
			}).Error("解析 cron 表达式失败，任务将被跳过")
			continue
		}

		taskName := def.Name
		if taskName == "" {
			taskName = name
		}

		task := &ScheduledTask{
			Name:             taskName,
			Description:      def.Description,
			CronExpr:         cronExpr,
			Enabled:          def.Enabled,
			ScriptFile:       def.Script,
			InitialDelay:     def.InitialDelay,
			Timeout:          def.Timeout,
			cronSchedule:     schedule,
			usesInitialDelay: def.InitialDelay > 0,
		}

		if task.Enabled {
			if def.InitialDelay > 0 {
				task.NextRun = now.Add(def.InitialDelay)
			} else {
				task.NextRun = schedule.Next(now)
			}
		} else {
			task.NextRun = schedule.Next(now)
		}

		result[name] = task
	}

	return result
}

// Start 按配置启动调度器。
func (s *OperationalScheduler) Start(ctx context.Context) {
	if s == nil {
		return
	}
	s.mu.Lock()
	if s.config != nil && !s.config.Enabled {
		s.mu.Unlock()
		s.logger.Info("运维任务调度器未启动（配置禁用）")
		return
	}
	if s.running {
		s.mu.Unlock()
		s.logger.Warn("运维任务调度器已在运行中")
		return
	}
	if len(s.tasks) == 0 {
		s.mu.Unlock()
		s.logger.Warn("无任务配置，跳过运维任务调度器启动")
		return
	}

	s.logger.Infof("启动运维任务调度器 (tick=%v, tasks=%d)", s.tickInterval, len(s.tasks))

	if s.monitor != nil && s.monitorEnabled {
		interval := s.monitorInterval
		if interval <= 0 {
			interval = 5 * time.Minute
		}
		s.monitor.StartPeriodicMonitoring(ctx, interval)
		s.logger.Infof("时态监控已启用 (interval=%v)", interval)
	} else {
		s.logger.Info("时态监控已禁用或未配置")
	}

	s.stopCh = make(chan struct{})
	s.running = true
	s.mu.Unlock()

	go s.schedulingLoop(ctx)
}

// Stop 停止调度器。
func (s *OperationalScheduler) Stop() {
	if s == nil {
		return
	}
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}

	s.logger.Warn("正在停止运维任务调度器...")
	close(s.stopCh)
	s.running = false
	s.mu.Unlock()
	s.logger.Info("运维任务调度器已停止")
}

func (s *OperationalScheduler) schedulingLoop(ctx context.Context) {
	ticker := time.NewTicker(s.tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("接收到上游上下文取消，停止调度循环")
			return
		case <-s.stopCh:
			s.logger.Info("收到停止信号，结束调度循环")
			return
		case now := <-ticker.C:
			for _, task := range s.tasks {
				task.mu.Lock()
				if !task.Enabled {
					task.mu.Unlock()
					continue
				}

				nextRun := task.NextRun
				if nextRun.IsZero() {
					nextRun = task.cronSchedule.Next(now)
					task.NextRun = nextRun
				}
				if now.Before(nextRun) || task.Running {
					task.mu.Unlock()
					continue
				}

				scheduledAt := nextRun
				task.NextRun = task.cronSchedule.Next(now)
				if task.usesInitialDelay {
					task.usesInitialDelay = false
				}
				task.Running = true
				task.mu.Unlock()

				go s.executeTask(ctx, task, scheduledAt)
			}
		}
	}
}

func (s *OperationalScheduler) executeTask(ctx context.Context, task *ScheduledTask, scheduledAt time.Time) {
	startTime := time.Now()
	s.logger.WithFields(pkglogger.Fields{
		"task":      task.Name,
		"cron":      task.CronExpr,
		"scheduled": scheduledAt.Format(time.RFC3339),
	}).Info("开始执行任务")

	var err error

	switch task.Name {
	case "acting_assignment_auto_revert":
		err = s.runActingAssignmentAutoRevert(ctx)
	case "system_monitoring":
		err = s.executeMonitoring(ctx)
	default:
		if task.ScriptFile != "" {
			err = s.executeScript(ctx, task.ScriptFile)
		} else {
			err = fmt.Errorf("任务 %s 缺少脚本或实现", task.Name)
		}
	}

	completed := time.Now()
	task.mu.Lock()
	if task.LastRun == nil {
		task.LastRun = &completed
	} else {
		*task.LastRun = completed
	}
	task.Running = false
	task.mu.Unlock()

	if err != nil {
		s.logger.WithFields(pkglogger.Fields{
			"task": task.Name,
			"err":  err,
		}).Errorf("任务执行失败，耗时: %v", time.Since(startTime))
		s.recordTaskExecution(task, "FAILED", err.Error(), time.Since(startTime))
	} else {
		s.logger.WithFields(pkglogger.Fields{
			"task": task.Name,
		}).Infof("任务执行成功，耗时: %v", time.Since(startTime))
		s.recordTaskExecution(task, "SUCCESS", "", time.Since(startTime))
	}
}

// RunTask 手动触发指定任务。
func (s *OperationalScheduler) RunTask(ctx context.Context, name string) error {
	if s == nil {
		return fmt.Errorf("scheduler 已禁用")
	}
	s.mu.RLock()
	cfgEnabled := s.config == nil || s.config.Enabled
	task, ok := s.tasks[name]
	s.mu.RUnlock()
	if !cfgEnabled {
		return fmt.Errorf("scheduler 已禁用")
	}
	if !ok {
		return fmt.Errorf("任务 %s 未配置", name)
	}

	task.mu.Lock()
	if task.Running {
		task.mu.Unlock()
		return fmt.Errorf("任务 %s 正在执行中", name)
	}
	scheduledAt := time.Now()
	task.NextRun = task.cronSchedule.Next(scheduledAt)
	task.usesInitialDelay = false
	task.Running = true
	task.mu.Unlock()

	s.executeTask(ctx, task, scheduledAt)

	return nil
}
func (s *OperationalScheduler) executeScript(ctx context.Context, scriptFile string) error {
	basePath, err := filepath.Abs(s.scriptsPath)
	if err != nil {
		return fmt.Errorf("解析脚本目录失败: %w", err)
	}

	cleanFile := filepath.Clean(scriptFile)
	target := filepath.Join(basePath, cleanFile)

	if !strings.HasPrefix(target, basePath) {
		return fmt.Errorf("拒绝执行目录外脚本: %s", scriptFile)
	}

	if _, err := os.Stat(target); os.IsNotExist(err) {
		return fmt.Errorf("脚本文件不存在: %s", target)
	}

	sqlContent, err := os.ReadFile(target) // #nosec G304
	if err != nil {
		return fmt.Errorf("读取脚本文件失败: %w", err)
	}

	if _, err := s.db.ExecContext(ctx, string(sqlContent)); err != nil {
		return fmt.Errorf("执行SQL脚本失败: %w", err)
	}

	return nil
}

func (s *OperationalScheduler) executeMonitoring(ctx context.Context) error {
	if s.monitor == nil {
		return fmt.Errorf("monitor 未配置")
	}
	alerts, err := s.monitor.CheckAlerts(ctx)
	if err != nil {
		return fmt.Errorf("监控检查失败: %w", err)
	}
	if len(alerts) > 0 {
		s.logger.Warnf("监控发现 %d 个告警", len(alerts))
		for _, alert := range alerts {
			s.logger.Warnf("告警详情: %s", alert)
		}
	}
	return nil
}

func (s *OperationalScheduler) recordTaskExecution(task *ScheduledTask, status, message string, duration time.Duration) {
	fields := pkglogger.Fields{
		"task":     task.Name,
		"status":   status,
		"duration": duration.String(),
	}
	if message != "" {
		fields["error"] = message
	}
	s.logger.WithFields(fields).Info("任务执行完成")
}

// ListTasks 返回当前任务状态副本。
func (s *OperationalScheduler) ListTasks() []ScheduledTask {
	results := make([]ScheduledTask, 0, len(s.tasks))
	for _, task := range s.tasks {
		task.mu.Lock()
		clone := *task
		task.mu.Unlock()
		results = append(results, clone)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	return results
}

// IsRunning 返回调度器运行状态。
func (s *OperationalScheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

func (s *OperationalScheduler) runActingAssignmentAutoRevert(ctx context.Context) error {
	if s.positions == nil {
		return fmt.Errorf("position service 未配置")
	}

	operator := types.OperatedByInfo{ID: "", Name: "auto-revert-scheduler"}
	processed, err := s.positions.ProcessAutoReverts(ctx, types.DefaultTenantID, time.Now().UTC(), 200, operator)
	if err != nil {
		return err
	}

	if len(processed) > 0 {
		s.logger.Infof("[AUTO-REVERT] 成功自动结束 %d 条代理任职", len(processed))
	} else {
		s.logger.Info("[AUTO-REVERT] 无代理任职需要自动结束")
	}

	return nil
}
