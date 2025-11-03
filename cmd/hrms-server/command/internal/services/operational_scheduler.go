package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cube-castle/internal/types"
)

// OperationalScheduler 运维任务调度器
type OperationalScheduler struct {
	db          *sql.DB
	logger      *log.Logger
	monitor     *TemporalMonitor
	positions   *PositionService
	scriptsPath string
	running     bool
	stopCh      chan struct{}
}

// NewOperationalScheduler 创建运维任务调度器
func NewOperationalScheduler(db *sql.DB, logger *log.Logger, monitor *TemporalMonitor, positions *PositionService) *OperationalScheduler {
	// 获取脚本目录路径
	scriptsPath := filepath.Join(os.Getenv("PWD"), "scripts")
	if _, err := os.Stat(scriptsPath); os.IsNotExist(err) {
		// 如果当前目录下没有scripts目录，尝试相对路径
		scriptsPath = "./scripts"
	}

	return &OperationalScheduler{
		db:          db,
		logger:      logger,
		monitor:     monitor,
		positions:   positions,
		scriptsPath: scriptsPath,
		stopCh:      make(chan struct{}),
	}
}

// ScheduledTask 定时任务配置
type ScheduledTask struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Schedule    string        `json:"schedule"`   // "daily", "hourly", "custom"
	Interval    time.Duration `json:"interval"`   // 自定义间隔
	ScriptFile  string        `json:"scriptFile"` // SQL脚本文件路径
	Enabled     bool          `json:"enabled"`
	LastRun     *time.Time    `json:"lastRun,omitempty"`
	NextRun     time.Time     `json:"nextRun"`
}

// GetDefaultTasks 获取默认任务配置
func (s *OperationalScheduler) GetDefaultTasks() []ScheduledTask {
	now := time.Now()

	// 计算明天凌晨2点的时间
	tomorrow2AM := time.Date(now.Year(), now.Month(), now.Day()+1, 2, 0, 0, 0, now.Location())

	// 计算下一个整点时间
	nextHour := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, now.Location())

	return []ScheduledTask{
		{
			Name:        "daily_cutover",
			Description: "每日cutover任务：维护时态数据一致性",
			Schedule:    "daily",
			Interval:    24 * time.Hour,
			ScriptFile:  "daily-cutover.sql",
			Enabled:     true,
			NextRun:     tomorrow2AM,
		},
		{
			Name:        "acting_assignment_auto_revert",
			Description: "自动结束到期的代理任职",
			Schedule:    "daily",
			Interval:    24 * time.Hour,
			ScriptFile:  "",
			Enabled:     true,
			NextRun:     tomorrow2AM.Add(15 * time.Minute),
		},
		{
			Name:        "data_consistency_check",
			Description: "数据一致性检查",
			Schedule:    "daily",
			Interval:    24 * time.Hour,
			ScriptFile:  "data-consistency-check.sql",
			Enabled:     true,
			NextRun:     tomorrow2AM.Add(30 * time.Minute), // 在cutover任务30分钟后执行
		},
		{
			Name:        "system_monitoring",
			Description: "系统健康监控",
			Schedule:    "hourly",
			Interval:    time.Hour,
			ScriptFile:  "", // 使用代码而非SQL脚本
			Enabled:     true,
			NextRun:     nextHour,
		},
	}
}

// Start 启动调度器
func (s *OperationalScheduler) Start(ctx context.Context) {
	if s.running {
		s.logger.Println("运维任务调度器已在运行中")
		return
	}

	s.running = true
	s.logger.Println("🚀 启动运维任务调度器...")

	// 启动定期监控
	s.monitor.StartPeriodicMonitoring(ctx, 5*time.Minute)

	// 启动任务调度循环
	go s.schedulingLoop(ctx)

	s.logger.Println("✅ 运维任务调度器已启动")
}

// Stop 停止调度器
func (s *OperationalScheduler) Stop() {
	if !s.running {
		return
	}

	s.logger.Println("🛑 正在停止运维任务调度器...")
	close(s.stopCh)
	s.running = false
	s.logger.Println("✅ 运维任务调度器已停止")
}

// schedulingLoop 任务调度循环
func (s *OperationalScheduler) schedulingLoop(ctx context.Context) {
	tasks := s.GetDefaultTasks()
	ticker := time.NewTicker(1 * time.Minute) // 每分钟检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case now := <-ticker.C:
			for i := range tasks {
				task := &tasks[i]
				if task.Enabled && now.After(task.NextRun) {
					go s.executeTask(ctx, task)
					s.scheduleNextRun(task)
				}
			}
		}
	}
}

// executeTask 执行任务
func (s *OperationalScheduler) executeTask(ctx context.Context, task *ScheduledTask) {
	startTime := time.Now()
	s.logger.Printf("🔄 开始执行任务: %s", task.Description)

	var err error

	if task.ScriptFile != "" {
		// 执行SQL脚本
		err = s.executeScript(ctx, task.ScriptFile)
	} else if task.Name == "acting_assignment_auto_revert" {
		err = s.runActingAssignmentAutoRevert(ctx)
	} else if task.Name == "system_monitoring" {
		// 执行监控检查
		err = s.executeMonitoring(ctx)
	}

	duration := time.Since(startTime)
	now := time.Now()
	task.LastRun = &now

	if err != nil {
		s.logger.Printf("❌ 任务执行失败: %s, 耗时: %v, 错误: %v", task.Description, duration, err)

		// 记录失败到审计日志
		s.recordTaskExecution(ctx, task, "FAILED", err.Error(), duration)
	} else {
		s.logger.Printf("✅ 任务执行成功: %s, 耗时: %v", task.Description, duration)

		// 记录成功到审计日志
		s.recordTaskExecution(ctx, task, "SUCCESS", "", duration)
	}
}

// executeScript 执行SQL脚本文件
func (s *OperationalScheduler) executeScript(ctx context.Context, scriptFile string) error {
	basePath, err := filepath.Abs(s.scriptsPath)
	if err != nil {
		return fmt.Errorf("解析脚本目录失败: %w", err)
	}

	cleanFile := filepath.Clean(scriptFile)
	scriptPath := filepath.Join(basePath, cleanFile)
	if scriptPath != basePath && !strings.HasPrefix(scriptPath, basePath+string(os.PathSeparator)) {
		return fmt.Errorf("拒绝执行目录外脚本: %s", scriptFile)
	}

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("脚本文件不存在: %s", scriptPath)
	}

	// #nosec G304 -- scriptPath 已限定在脚本目录下
	sqlContent, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("读取脚本文件失败: %w", err)
	}

	_, err = s.db.ExecContext(ctx, string(sqlContent))
	if err != nil {
		return fmt.Errorf("执行SQL脚本失败: %w", err)
	}

	return nil
}

// executeMonitoring 执行监控检查
func (s *OperationalScheduler) executeMonitoring(ctx context.Context) error {
	alerts, err := s.monitor.CheckAlerts(ctx)
	if err != nil {
		return fmt.Errorf("监控检查失败: %w", err)
	}

	if len(alerts) > 0 {
		s.logger.Printf("🚨 监控发现 %d 个告警:", len(alerts))
		for _, alert := range alerts {
			s.logger.Printf("  - %s", alert)
		}
	}

	return nil
}

// scheduleNextRun 安排下次运行时间
func (s *OperationalScheduler) scheduleNextRun(task *ScheduledTask) {
	now := time.Now()

	switch task.Schedule {
	case "daily":
		// 安排到明天的同一时间
		task.NextRun = task.NextRun.Add(24 * time.Hour)

		// 如果计算的下次运行时间还在过去，继续加一天直到未来
		for task.NextRun.Before(now) {
			task.NextRun = task.NextRun.Add(24 * time.Hour)
		}

	case "hourly":
		// 安排到下一个小时
		task.NextRun = task.NextRun.Add(time.Hour)

		// 如果计算的下次运行时间还在过去，继续加一小时直到未来
		for task.NextRun.Before(now) {
			task.NextRun = task.NextRun.Add(time.Hour)
		}

	case "custom":
		// 使用自定义间隔
		task.NextRun = now.Add(task.Interval)

	default:
		// 默认使用间隔时间
		task.NextRun = now.Add(task.Interval)
	}
}

// recordTaskExecution 记录任务执行结果
func (s *OperationalScheduler) recordTaskExecution(ctx context.Context, task *ScheduledTask, status, errorMsg string, duration time.Duration) {
	summary := fmt.Sprintf("任务: %s, 状态: %s, 耗时: %v", task.Name, status, duration)
	if errorMsg != "" {
		summary += fmt.Sprintf(", 错误: %s", errorMsg)
	}

	// 任务执行结果记录到应用日志而非审计表
	// 系统任务日志不属于业务操作审计范围，应该单独处理
	s.logger.Printf("📋 任务执行完成: %s - %s", task.Name, summary)
}

// GetTaskStatus 获取任务状态
func (s *OperationalScheduler) GetTaskStatus() []ScheduledTask {
	return s.GetDefaultTasks()
}

// IsRunning 检查调度器是否运行中
func (s *OperationalScheduler) IsRunning() bool {
	return s.running
}

func (s *OperationalScheduler) runActingAssignmentAutoRevert(ctx context.Context) error {
	if s.positions == nil {
		return fmt.Errorf("position service not configured for auto revert task")
	}

	operator := types.OperatedByInfo{ID: "", Name: "auto-revert-scheduler"}
	processed, err := s.positions.ProcessAutoReverts(ctx, types.DefaultTenantID, time.Now().UTC(), 200, operator)
	if err != nil {
		return err
	}

	if len(processed) > 0 {
		s.logger.Printf("[AUTO-REVERT] 成功自动结束 %d 条代理任职", len(processed))
	} else {
		s.logger.Println("[AUTO-REVERT] 无代理任职需要自动结束")
	}

	return nil
}
