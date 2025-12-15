package pipeline

import (
	"github.com/xiaolfeng/builder-cli/internal/config"
)

// Stage 流水线阶段
type Stage struct {
	Index    int
	ID       string
	Name     string
	Parallel bool
	Tasks    []*Task
}

// NewStage 创建新的阶段
func NewStage(index int, cfg config.Stage, fullCfg *config.Config) *Stage {
	s := &Stage{
		Index:    index,
		ID:       cfg.Stage,
		Name:     cfg.Name,
		Parallel: cfg.Parallel,
		Tasks:    make([]*Task, 0, len(cfg.Tasks)),
	}

	// 创建任务
	for i, taskCfg := range cfg.Tasks {
		task := NewTask(index, i, taskCfg)
		s.Tasks = append(s.Tasks, task)
	}

	return s
}

// GetTaskCount 获取任务数量
func (s *Stage) GetTaskCount() int {
	return len(s.Tasks)
}

// GetCompletedCount 获取已完成的任务数量
func (s *Stage) GetCompletedCount() int {
	count := 0
	for _, task := range s.Tasks {
		if task.IsCompleted() {
			count++
		}
	}
	return count
}

// IsCompleted 检查阶段是否已完成
func (s *Stage) IsCompleted() bool {
	return s.GetCompletedCount() == len(s.Tasks)
}

// HasFailed 检查阶段是否有失败的任务
func (s *Stage) HasFailed() bool {
	for _, task := range s.Tasks {
		if task.IsFailed() {
			return true
		}
	}
	return false
}
