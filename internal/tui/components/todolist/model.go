package todolist

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xiaolfeng/builder-cli/internal/types"
)

// Task 任务项
type Task struct {
	ID          string
	Name        string
	Description string
	Status      types.TaskStatus
	StartTime   time.Time
	EndTime     time.Time
}

// Duration 返回任务耗时
func (t Task) Duration() time.Duration {
	if t.StartTime.IsZero() {
		return 0
	}
	if t.EndTime.IsZero() {
		return time.Since(t.StartTime)
	}
	return t.EndTime.Sub(t.StartTime)
}

// Model Todo List 组件模型
type Model struct {
	tasks       []Task
	width       int
	height      int
	scrollIndex int
	maxVisible  int
}

// New 创建新的 Todo List 组件
func New() Model {
	return Model{
		tasks:      make([]Task, 0),
		maxVisible: 8, // 默认显示 8 个任务
	}
}

// WithTasks 设置任务列表
func (m Model) WithTasks(tasks []Task) Model {
	m.tasks = tasks
	return m
}

// SetSize 设置组件尺寸
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	// 根据高度计算可显示的任务数（减去标题和边框）
	// 保底至少 1 行，避免小窗口时 maxVisible 维持默认值导致内容溢出
	m.maxVisible = height - 4
	if m.maxVisible < 1 {
		m.maxVisible = 1
	}
}

// GetTasks 获取任务列表
func (m Model) GetTasks() []Task {
	return m.tasks
}

// AddTask 添加任务
func (m *Model) AddTask(task Task) {
	m.tasks = append(m.tasks, task)
}

// UpdateTaskStatus 更新任务状态
func (m *Model) UpdateTaskStatus(taskID string, status types.TaskStatus) {
	for i := range m.tasks {
		if m.tasks[i].ID == taskID {
			m.tasks[i].Status = status
			if status == types.StatusRunning && m.tasks[i].StartTime.IsZero() {
				m.tasks[i].StartTime = time.Now()
			}
			if status == types.StatusSuccess || status == types.StatusFailed {
				m.tasks[i].EndTime = time.Now()
			}
			break
		}
	}
}

// GetProgress 获取完成进度
func (m Model) GetProgress() (completed, total int) {
	total = len(m.tasks)
	for _, task := range m.tasks {
		if task.Status == types.StatusSuccess || task.Status == types.StatusSkipped {
			completed++
		}
	}
	return
}

// Init 实现 tea.Model 接口
func (m Model) Init() tea.Cmd {
	return nil
}

// Update 实现 tea.Model 接口
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case types.TaskStatusMsg:
		m.UpdateTaskStatus(msg.TaskID, msg.Status)
	}

	return m, nil
}
