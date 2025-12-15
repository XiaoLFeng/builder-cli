package terminal

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/xiaolfeng/builder-cli/internal/types"
)

// LogEntry 日志条目
type LogEntry struct {
	TaskID   string
	TaskName string
	Line     string
	IsError  bool
}

// Model 终端组件模型 - 用于展示实时日志输出
type Model struct {
	viewport   viewport.Model
	width      int
	height     int
	logEntries []LogEntry
	maxEntries int
	ready      bool
	autoScroll bool
	taskNames  map[string]string // taskID -> taskName 映射
}

// New 创建新的终端组件
func New() Model {
	return Model{
		logEntries: make([]LogEntry, 0),
		maxEntries: 500, // 最多保留 500 行日志
		autoScroll: true,
		taskNames:  make(map[string]string),
	}
}

// SetSize 设置组件尺寸
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	// viewport 尺寸（减去边框）
	vpWidth := width - 4
	vpHeight := height - 2
	if vpWidth < 20 {
		vpWidth = 20
	}
	if vpHeight < 3 {
		vpHeight = 3
	}

	if !m.ready {
		m.viewport = viewport.New(vpWidth, vpHeight)
		m.viewport.Style = viewportStyle
		m.ready = true
	} else {
		m.viewport.Width = vpWidth
		m.viewport.Height = vpHeight
	}

	// 更新内容
	m.updateContent()
}

// RegisterTask 注册任务名称映射
func (m *Model) RegisterTask(taskID, taskName string) {
	m.taskNames[taskID] = taskName
}

// AppendLog 追加日志条目
func (m *Model) AppendLog(taskID, line string, isError bool) {
	taskName := m.taskNames[taskID]
	if taskName == "" {
		taskName = taskID
	}

	entry := LogEntry{
		TaskID:   taskID,
		TaskName: taskName,
		Line:     line,
		IsError:  isError,
	}

	m.logEntries = append(m.logEntries, entry)

	// 限制日志数量
	if len(m.logEntries) > m.maxEntries {
		m.logEntries = m.logEntries[len(m.logEntries)-m.maxEntries:]
	}

	// 更新 viewport 内容
	m.updateContent()
}

// updateContent 更新 viewport 内容
func (m *Model) updateContent() {
	if !m.ready {
		return
	}

	var lines []string
	for _, entry := range m.logEntries {
		line := m.formatLogEntry(entry)
		lines = append(lines, line)
	}

	m.viewport.SetContent(strings.Join(lines, "\n"))

	if m.autoScroll {
		m.viewport.GotoBottom()
	}
}

// formatLogEntry 格式化日志条目
func (m *Model) formatLogEntry(entry LogEntry) string {
	// 获取任务名简称（最多 12 字符）
	taskLabel := entry.TaskName
	if len(taskLabel) > 12 {
		taskLabel = taskLabel[:9] + "..."
	}

	// 格式化：[任务名] 日志内容
	prefix := taskLabelStyle.Render("[" + taskLabel + "]")

	content := entry.Line
	if entry.IsError {
		content = errorLineStyle.Render(content)
	} else {
		content = normalLineStyle.Render(content)
	}

	return prefix + " " + content
}

// ClearLogs 清空日志
func (m *Model) ClearLogs() {
	m.logEntries = make([]LogEntry, 0)
	if m.ready {
		m.viewport.SetContent("")
	}
}

// ToggleAutoScroll 切换自动滚动
func (m *Model) ToggleAutoScroll() {
	m.autoScroll = !m.autoScroll
}

// IsAutoScroll 是否自动滚动
func (m Model) IsAutoScroll() bool {
	return m.autoScroll
}

// GetLogCount 获取日志数量
func (m Model) GetLogCount() int {
	return len(m.logEntries)
}

// Init 实现 tea.Model 接口
func (m Model) Init() tea.Cmd {
	return nil
}

// Update 实现 tea.Model 接口
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)

	case types.OutputMsg:
		m.AppendLog(msg.TaskID, msg.Line, msg.IsError)

	case tea.KeyMsg:
		// 支持 viewport 滚动
		if m.ready {
			switch msg.String() {
			case "up", "k":
				m.autoScroll = false
				m.viewport.LineUp(1)
			case "down", "j":
				m.viewport.LineDown(1)
			case "pgup":
				m.autoScroll = false
				m.viewport.HalfViewUp()
			case "pgdown":
				m.viewport.HalfViewDown()
			case "home", "g":
				m.autoScroll = false
				m.viewport.GotoTop()
			case "end", "G":
				m.autoScroll = true
				m.viewport.GotoBottom()
			}
		}
	}

	// 更新 viewport
	if m.ready {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
