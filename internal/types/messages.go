package types

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

// 状态图标颜色（与 styles 包同步）
var (
	mutedColor   = lipgloss.Color("#626262")
	warningColor = lipgloss.Color("#EDFF82")
	successColor = lipgloss.Color("#73F59F")
	errorColor   = lipgloss.Color("#FF6B6B")
)

// TaskStatus 任务状态
type TaskStatus int

const (
	StatusPending TaskStatus = iota
	StatusRunning
	StatusSuccess
	StatusFailed
	StatusSkipped
	StatusCancelled
)

// String 返回状态字符串
func (s TaskStatus) String() string {
	switch s {
	case StatusPending:
		return "等待中"
	case StatusRunning:
		return "进行中"
	case StatusSuccess:
		return "完成"
	case StatusFailed:
		return "失败"
	case StatusSkipped:
		return "跳过"
	case StatusCancelled:
		return "已取消"
	default:
		return "未知"
	}
}

// Icon 返回状态图标
func (s TaskStatus) Icon() string {
	switch s {
	case StatusPending:
		return lipgloss.NewStyle().Foreground(mutedColor).Render("○")
	case StatusRunning:
		return lipgloss.NewStyle().Foreground(warningColor).Render("●")
	case StatusSuccess:
		return lipgloss.NewStyle().Foreground(successColor).Render("✓")
	case StatusFailed:
		return lipgloss.NewStyle().Foreground(errorColor).Render("✗")
	case StatusSkipped:
		return lipgloss.NewStyle().Foreground(mutedColor).Render("⊘")
	case StatusCancelled:
		return lipgloss.NewStyle().Foreground(mutedColor).Render("⊗")
	default:
		return lipgloss.NewStyle().Foreground(mutedColor).Render("○")
	}
}

// ─────────────────────────────────────────────────────────────────────
// 自定义消息类型
// ─────────────────────────────────────────────────────────────────────

// OutputMsg 任务输出消息（实时流）
type OutputMsg struct {
	TaskID  string
	Line    string
	IsError bool
}

// OutputLine 批量输出中的单行
type OutputLine struct {
	Line    string
	IsError bool
}

// OutputBatchMsg 任务输出消息（批量，用于降低刷新频率）
type OutputBatchMsg struct {
	TaskID string
	Lines  []OutputLine
}

// TaskStatusMsg 任务状态变更消息
type TaskStatusMsg struct {
	TaskID string
	Status TaskStatus
}

// TaskProgressMsg 任务进度消息
type TaskProgressMsg struct {
	TaskID  string
	Current int
	Total   int
	Message string
}

// StageStartMsg 阶段开始消息
type StageStartMsg struct {
	StageIndex int
	StageName  string
}

// StageCompleteMsg 阶段完成消息
type StageCompleteMsg struct {
	StageIndex int
	StageName  string
	Success    bool
	Duration   time.Duration
}

// PipelineCompleteMsg 流水线完成消息
type PipelineCompleteMsg struct {
	Success  bool
	Duration time.Duration
	Error    error
}

// ErrorMsg 错误消息
type ErrorMsg struct {
	TaskID  string
	Error   error
	Message string
}

// TickMsg 定时器消息
type TickMsg time.Time

// ─────────────────────────────────────────────────────────────────────
// 消息构造函数
// ─────────────────────────────────────────────────────────────────────

// NewOutputMsg 创建输出消息
func NewOutputMsg(taskID, line string, isError bool) OutputMsg {
	return OutputMsg{TaskID: taskID, Line: line, IsError: isError}
}

// NewOutputBatchMsg 创建批量输出消息
func NewOutputBatchMsg(taskID string, lines []OutputLine) OutputBatchMsg {
	return OutputBatchMsg{TaskID: taskID, Lines: lines}
}

// NewTaskStatusMsg 创建任务状态消息
func NewTaskStatusMsg(taskID string, status TaskStatus) TaskStatusMsg {
	return TaskStatusMsg{TaskID: taskID, Status: status}
}

// NewTaskProgressMsg 创建任务进度消息
func NewTaskProgressMsg(taskID string, current, total int, message string) TaskProgressMsg {
	return TaskProgressMsg{TaskID: taskID, Current: current, Total: total, Message: message}
}

// NewStageStartMsg 创建阶段开始消息
func NewStageStartMsg(index int, name string) StageStartMsg {
	return StageStartMsg{StageIndex: index, StageName: name}
}

// NewStageCompleteMsg 创建阶段完成消息
func NewStageCompleteMsg(index int, name string, success bool, duration time.Duration) StageCompleteMsg {
	return StageCompleteMsg{StageIndex: index, StageName: name, Success: success, Duration: duration}
}

// NewPipelineCompleteMsg 创建流水线完成消息
func NewPipelineCompleteMsg(success bool, duration time.Duration, err error) PipelineCompleteMsg {
	return PipelineCompleteMsg{Success: success, Duration: duration, Error: err}
}

// NewErrorMsg 创建错误消息
func NewErrorMsg(taskID string, err error, message string) ErrorMsg {
	return ErrorMsg{TaskID: taskID, Error: err, Message: message}
}
