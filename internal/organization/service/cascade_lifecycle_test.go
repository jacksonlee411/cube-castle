package service

import (
	"testing"
)

func TestCascadeLifecycle_StartStopAndScheduleGuard(t *testing.T) {
	svc := NewCascadeUpdateService(nil, 2, nil)

	// 初始状态
	if q, w, running := svc.GetQueueStats(); q != 0 || w != 2 || running {
		t.Fatalf("unexpected initial stats: q=%d w=%d running=%v", q, w, running)
	}

	// 未启动时调度应返回 false
	ok := svc.ScheduleTask(CascadeTask{Type: TaskTypeUpdatePaths, Code: "1000001"})
	if ok {
		t.Fatalf("expected ScheduleTask=false when service not running")
	}

	// 启动 -> running=true
	svc.Start()
	if _, _, running := svc.GetQueueStats(); !running {
		t.Fatalf("expected running=true after Start()")
	}

	// 停止 -> running=false
	svc.Stop()
	if _, _, running := svc.GetQueueStats(); running {
		t.Fatalf("expected running=false after Stop()")
	}
}
