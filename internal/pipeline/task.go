package pipeline

import (
	"fmt"
	"time"

	"github.com/xiaolfeng/builder-cli/internal/config"
	"github.com/xiaolfeng/builder-cli/internal/types"
)

// Task 流水线任务
type Task struct {
	ID         string
	Name       string
	Type       string
	Config     config.TaskConfig
	Status     types.TaskStatus
	StartTime  time.Time
	EndTime    time.Time
	Error      error
	StageIndex int
	TaskIndex  int
}

// NewTask 创建新的任务
func NewTask(stageIndex, taskIndex int, cfg config.Task) *Task {
	return &Task{
		ID:         fmt.Sprintf("task-%d-%d", stageIndex, taskIndex),
		Name:       cfg.Name,
		Type:       cfg.Type,
		Config:     cfg.Config,
		Status:     types.StatusPending,
		StageIndex: stageIndex,
		TaskIndex:  taskIndex,
	}
}

// Start 开始任务
func (t *Task) Start() {
	t.Status = types.StatusRunning
	t.StartTime = time.Now()
}

// Complete 完成任务
func (t *Task) Complete() {
	t.Status = types.StatusSuccess
	t.EndTime = time.Now()
}

// Fail 任务失败
func (t *Task) Fail(err error) {
	t.Status = types.StatusFailed
	t.EndTime = time.Now()
	t.Error = err
}

// Skip 跳过任务
func (t *Task) Skip() {
	t.Status = types.StatusSkipped
}

// IsCompleted 检查任务是否已完成（成功或跳过）
func (t *Task) IsCompleted() bool {
	return t.Status == types.StatusSuccess || t.Status == types.StatusSkipped
}

// IsFailed 检查任务是否失败
func (t *Task) IsFailed() bool {
	return t.Status == types.StatusFailed
}

// IsRunning 检查任务是否正在运行
func (t *Task) IsRunning() bool {
	return t.Status == types.StatusRunning
}

// IsPending 检查任务是否等待中
func (t *Task) IsPending() bool {
	return t.Status == types.StatusPending
}

// Duration 返回任务耗时
func (t *Task) Duration() time.Duration {
	if t.StartTime.IsZero() {
		return 0
	}
	if t.EndTime.IsZero() {
		return time.Since(t.StartTime)
	}
	return t.EndTime.Sub(t.StartTime)
}

// ToTodoListTask 转换为 TodoList 任务
func (t *Task) ToTodoListTask() interface{} {
	return struct {
		ID          string
		Name        string
		Description string
		Status      types.TaskStatus
		StartTime   time.Time
		EndTime     time.Time
	}{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Type,
		Status:      t.Status,
		StartTime:   t.StartTime,
		EndTime:     t.EndTime,
	}
}
