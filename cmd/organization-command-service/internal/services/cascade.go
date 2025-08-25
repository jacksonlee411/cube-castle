package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"organization-command-service/internal/repository"
)

// CascadeUpdateService 异步级联更新服务
type CascadeUpdateService struct {
	hierarchyRepo *repository.HierarchyRepository
	taskQueue     chan CascadeTask
	workers       int
	logger        *log.Logger
	wg            sync.WaitGroup
	shutdown      chan struct{}
	running       bool
	mu            sync.RWMutex
}

// CascadeTask 级联任务
type CascadeTask struct {
	Type      string    `json:"type"`
	Code      string    `json:"code"`
	TenantID  uuid.UUID `json:"tenantId"`
	UserID    string    `json:"userId"`
	Context   context.Context
	Timestamp time.Time `json:"timestamp"`
	Priority  int       `json:"priority"` // 1=高优先级, 2=中优先级, 3=低优先级
}

// 任务类型常量
const (
	TaskTypeUpdateHierarchy = "UPDATE_HIERARCHY"
	TaskTypeUpdatePaths     = "UPDATE_PATHS"
	TaskTypeUpdateStatus    = "UPDATE_STATUS"
	TaskTypeValidateRules   = "VALIDATE_RULES"
)

func NewCascadeUpdateService(hierarchyRepo *repository.HierarchyRepository, workers int, logger *log.Logger) *CascadeUpdateService {
	if workers <= 0 {
		workers = 4 // 默认4个工作协程
	}

	service := &CascadeUpdateService{
		hierarchyRepo: hierarchyRepo,
		taskQueue:     make(chan CascadeTask, 1000), // 任务队列缓冲区
		workers:       workers,
		logger:        logger,
		shutdown:      make(chan struct{}),
		running:       false,
	}

	return service
}

// Start 启动级联更新服务
func (c *CascadeUpdateService) Start() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		c.logger.Println("⚠️ 级联更新服务已在运行")
		return
	}

	c.running = true
	c.logger.Printf("🚀 启动级联更新服务，工作协程数: %d", c.workers)

	// 启动工作协程池
	for i := 0; i < c.workers; i++ {
		c.wg.Add(1)
		go c.worker(i)
	}
}

// Stop 停止级联更新服务
func (c *CascadeUpdateService) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return
	}

	c.logger.Println("🛑 正在停止级联更新服务...")
	c.running = false
	close(c.shutdown)
	close(c.taskQueue)

	c.wg.Wait()
	c.logger.Println("✅ 级联更新服务已停止")
}

// worker 工作协程
func (c *CascadeUpdateService) worker(workerID int) {
	defer c.wg.Done()

	c.logger.Printf("🔧 工作协程 %d 已启动", workerID)

	for {
		select {
		case task, ok := <-c.taskQueue:
			if !ok {
				c.logger.Printf("📤 工作协程 %d 退出 (任务队列已关闭)", workerID)
				return
			}
			c.processTask(workerID, task)

		case <-c.shutdown:
			c.logger.Printf("🔻 工作协程 %d 退出 (收到停止信号)", workerID)
			return
		}
	}
}

// processTask 处理任务
func (c *CascadeUpdateService) processTask(workerID int, task CascadeTask) {
	start := time.Now()
	c.logger.Printf("⚡ 工作协程 %d 开始处理任务: %s (组织: %s, 优先级: %d)",
		workerID, task.Type, task.Code, task.Priority)

	var err error
	switch task.Type {
	case TaskTypeUpdateHierarchy:
		err = c.processHierarchyUpdate(task)
	case TaskTypeUpdatePaths:
		err = c.processPathUpdate(task)
	case TaskTypeUpdateStatus:
		err = c.processStatusUpdate(task)
	case TaskTypeValidateRules:
		err = c.processValidateRules(task)
	default:
		c.logger.Printf("❌ 未知任务类型: %s", task.Type)
		return
	}

	duration := time.Since(start)
	if err != nil {
		c.logger.Printf("❌ 工作协程 %d 任务处理失败: %s (组织: %s, 耗时: %v, 错误: %v)",
			workerID, task.Type, task.Code, duration, err)
	} else {
		c.logger.Printf("✅ 工作协程 %d 任务处理成功: %s (组织: %s, 耗时: %v)",
			workerID, task.Type, task.Code, duration)
	}
}

// processHierarchyUpdate 处理层级结构更新
func (c *CascadeUpdateService) processHierarchyUpdate(task CascadeTask) error {
	ctx := task.Context
	if ctx == nil {
		ctx = context.Background()
	}

	// 获取所有直接子组织
	children, err := c.hierarchyRepo.GetDirectChildren(ctx, task.Code, task.TenantID)
	if err != nil {
		return fmt.Errorf("获取子组织失败: %w", err)
	}

	c.logger.Printf("🔄 层级更新: 组织 %s 有 %d 个直接子组织", task.Code, len(children))

	// 更新路径信息
	err = c.hierarchyRepo.UpdateHierarchyPaths(ctx, task.Code, task.TenantID)
	if err != nil {
		return fmt.Errorf("更新层级路径失败: %w", err)
	}

	// 为每个子组织调度路径更新任务
	for _, child := range children {
		childTask := CascadeTask{
			Type:      TaskTypeUpdatePaths,
			Code:      child.Code,
			TenantID:  task.TenantID,
			UserID:    task.UserID,
			Context:   ctx,
			Timestamp: time.Now(),
			Priority:  task.Priority + 1, // 子任务优先级递减
		}

		// 异步调度子任务
		go c.ScheduleTask(childTask)
	}

	return nil
}

// processPathUpdate 处理路径更新
func (c *CascadeUpdateService) processPathUpdate(task CascadeTask) error {
	ctx := task.Context
	if ctx == nil {
		ctx = context.Background()
	}

	// 更新当前组织的路径
	err := c.hierarchyRepo.UpdateHierarchyPaths(ctx, task.Code, task.TenantID)
	if err != nil {
		return fmt.Errorf("更新路径失败: %w", err)
	}

	// 检查是否有子组织需要继续级联更新
	children, err := c.hierarchyRepo.GetDirectChildren(ctx, task.Code, task.TenantID)
	if err != nil {
		return fmt.Errorf("检查子组织失败: %w", err)
	}

	// 为每个子组织继续调度更新任务
	for _, child := range children {
		if task.Priority < 5 { // 限制递归深度，防止无限循环
			childTask := CascadeTask{
				Type:      TaskTypeUpdatePaths,
				Code:      child.Code,
				TenantID:  task.TenantID,
				UserID:    task.UserID,
				Context:   ctx,
				Timestamp: time.Now(),
				Priority:  task.Priority + 1,
			}
			go c.ScheduleTask(childTask)
		}
	}

	return nil
}

// processStatusUpdate 处理状态更新级联
func (c *CascadeUpdateService) processStatusUpdate(task CascadeTask) error {
	ctx := task.Context
	if ctx == nil {
		ctx = context.Background()
	}

	// 获取组织当前状态
	org, err := c.hierarchyRepo.GetOrganization(ctx, task.Code, task.TenantID)
	if err != nil {
		return fmt.Errorf("获取组织状态失败: %w", err)
	}

	// 如果是停用状态，需要检查对子组织的影响
	if org.Status == "INACTIVE" {
		children, err := c.hierarchyRepo.GetDirectChildren(ctx, task.Code, task.TenantID)
		if err != nil {
			return fmt.Errorf("获取子组织失败: %w", err)
		}

		c.logger.Printf("📊 状态级联检查: 组织 %s 状态为 %s, 影响 %d 个子组织",
			task.Code, org.Status, len(children))

		// 这里可以实现具体的状态级联逻辑
		// 例如：父组织停用时，是否自动停用子组织
		for _, child := range children {
			c.logger.Printf("⚠️ 子组织 %s 受父组织状态变化影响", child.Code)
			// 可以在这里实现具体的业务逻辑
		}
	}

	return nil
}

// processValidateRules 处理业务规则验证
func (c *CascadeUpdateService) processValidateRules(task CascadeTask) error {
	ctx := task.Context
	if ctx == nil {
		ctx = context.Background()
	}

	// 获取组织层级信息
	hierarchy, err := c.hierarchyRepo.GetOrganizationHierarchy(ctx, task.Code, task.TenantID, 17)
	if err != nil {
		return fmt.Errorf("获取层级结构失败: %w", err)
	}

	// 验证层级深度
	maxDepth := 0
	for _, node := range hierarchy {
		if node.Depth > maxDepth {
			maxDepth = node.Depth
		}
	}

	if maxDepth > 17 {
		c.logger.Printf("⚠️ 业务规则违规: 组织 %s 层级深度 %d 超过限制 (17级)", task.Code, maxDepth)
		// 这里可以实现具体的违规处理逻辑
	}

	// 检查循环引用 (通过祖先链长度判断)
	ancestors, err := c.hierarchyRepo.GetAncestorChain(ctx, task.Code, task.TenantID)
	if err != nil {
		return fmt.Errorf("获取祖先链失败: %w", err)
	}

	if len(ancestors) > 20 { // 异常长度的祖先链可能表示循环引用
		c.logger.Printf("⚠️ 可能的循环引用: 组织 %s 祖先链长度 %d", task.Code, len(ancestors))
	}

	c.logger.Printf("✅ 业务规则验证完成: 组织 %s, 层级深度 %d, 祖先链长度 %d",
		task.Code, maxDepth, len(ancestors))

	return nil
}

// ScheduleTask 调度任务
func (c *CascadeUpdateService) ScheduleTask(task CascadeTask) bool {
	c.mu.RLock()
	running := c.running
	c.mu.RUnlock()

	if !running {
		c.logger.Printf("⚠️ 级联更新服务未运行，任务被丢弃: %s (组织: %s)", task.Type, task.Code)
		return false
	}

	select {
	case c.taskQueue <- task:
		c.logger.Printf("📝 任务已调度: %s (组织: %s, 优先级: %d)", task.Type, task.Code, task.Priority)
		return true
	default:
		c.logger.Printf("⚠️ 任务队列已满，任务被丢弃: %s (组织: %s)", task.Type, task.Code)
		return false
	}
}

// ScheduleHierarchyUpdate 调度层级更新任务 (便捷方法)
func (c *CascadeUpdateService) ScheduleHierarchyUpdate(code string, tenantID uuid.UUID, userID string, ctx context.Context) bool {
	task := CascadeTask{
		Type:      TaskTypeUpdateHierarchy,
		Code:      code,
		TenantID:  tenantID,
		UserID:    userID,
		Context:   ctx,
		Timestamp: time.Now(),
		Priority:  1, // 高优先级
	}

	return c.ScheduleTask(task)
}

// SchedulePathUpdate 调度路径更新任务 (便捷方法)
func (c *CascadeUpdateService) SchedulePathUpdate(code string, tenantID uuid.UUID, userID string, ctx context.Context) bool {
	task := CascadeTask{
		Type:      TaskTypeUpdatePaths,
		Code:      code,
		TenantID:  tenantID,
		UserID:    userID,
		Context:   ctx,
		Timestamp: time.Now(),
		Priority:  2, // 中优先级
	}

	return c.ScheduleTask(task)
}

// ScheduleStatusUpdate 调度状态更新任务 (便捷方法)
func (c *CascadeUpdateService) ScheduleStatusUpdate(code string, tenantID uuid.UUID, userID string, ctx context.Context) bool {
	task := CascadeTask{
		Type:      TaskTypeUpdateStatus,
		Code:      code,
		TenantID:  tenantID,
		UserID:    userID,
		Context:   ctx,
		Timestamp: time.Now(),
		Priority:  2, // 中优先级
	}

	return c.ScheduleTask(task)
}

// GetQueueStats 获取队列统计信息
func (c *CascadeUpdateService) GetQueueStats() (queueSize int, workers int, running bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.taskQueue), c.workers, c.running
}
