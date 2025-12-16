package tui

import (
	"github.com/xiaolfeng/builder-cli/internal/types"
)

// 重新导出 types 包中的类型，保持向后兼容
type (
	TaskStatus          = types.TaskStatus
	OutputMsg           = types.OutputMsg
	OutputLine          = types.OutputLine
	OutputBatchMsg      = types.OutputBatchMsg
	TaskStatusMsg       = types.TaskStatusMsg
	TaskProgressMsg     = types.TaskProgressMsg
	StageStartMsg       = types.StageStartMsg
	StageCompleteMsg    = types.StageCompleteMsg
	PipelineCompleteMsg = types.PipelineCompleteMsg
	ErrorMsg            = types.ErrorMsg
	TickMsg             = types.TickMsg
)

// 重新导出状态常量
const (
	StatusPending   = types.StatusPending
	StatusRunning   = types.StatusRunning
	StatusSuccess   = types.StatusSuccess
	StatusFailed    = types.StatusFailed
	StatusSkipped   = types.StatusSkipped
	StatusCancelled = types.StatusCancelled
)

// Icon 返回状态图标 (需要在 tui 包中实现，因为依赖 styles)
func StatusIcon(s types.TaskStatus) string {
	switch s {
	case types.StatusPending:
		return IconPending
	case types.StatusRunning:
		return IconRunning
	case types.StatusSuccess:
		return IconSuccess
	case types.StatusFailed:
		return IconFailed
	case types.StatusSkipped:
		return IconSkipped
	case types.StatusCancelled:
		return IconFailed
	default:
		return IconPending
	}
}

// 重新导出构造函数
var (
	NewOutputMsg           = types.NewOutputMsg
	NewOutputBatchMsg      = types.NewOutputBatchMsg
	NewTaskStatusMsg       = types.NewTaskStatusMsg
	NewTaskProgressMsg     = types.NewTaskProgressMsg
	NewStageStartMsg       = types.NewStageStartMsg
	NewStageCompleteMsg    = types.NewStageCompleteMsg
	NewPipelineCompleteMsg = types.NewPipelineCompleteMsg
	NewErrorMsg            = types.NewErrorMsg
)
