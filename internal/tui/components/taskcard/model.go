package taskcard

import (
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/xiaolfeng/builder-cli/internal/styles"
	"github.com/xiaolfeng/builder-cli/internal/types"
)

// Model 任务卡片组件模型
type Model struct {
	taskID      string
	title       string
	status      types.TaskStatus
	viewport    viewport.Model
	progress    progress.Model
	width       int
	height      int
	outputLines []string
	maxLines    int
	current     int
	total       int
	ready       bool
}

// New 创建新的任务卡片
func New(taskID, title string) Model {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithoutPercentage(),
	)

	return Model{
		taskID:      taskID,
		title:       title,
		status:      types.StatusPending,
		progress:    p,
		outputLines: make([]string, 0),
		maxLines:    200, // 最多保留 200 行
	}
}

// SetSize 设置卡片尺寸
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	// viewport 尺寸（减去标题、进度条和边框）
	vpWidth := width - 6
	vpHeight := height - 6
	if vpWidth < 10 {
		vpWidth = 10
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

	// 进度条宽度
	m.progress.Width = width - 10
	if m.progress.Width < 10 {
		m.progress.Width = 10
	}
}

// SetStatus 设置任务状态
func (m *Model) SetStatus(status types.TaskStatus) {
	m.status = status
}

// SetProgress 设置进度
func (m *Model) SetProgress(current, total int) {
	m.current = current
	m.total = total
}

// AppendOutput 追加输出内容
func (m *Model) AppendOutput(line string, isError bool) {
	// 格式化输出行
	if isError {
		line = styles.ErrorTextStyle.Render(line)
	}

	m.outputLines = append(m.outputLines, line)

	// 限制行数
	if len(m.outputLines) > m.maxLines {
		m.outputLines = m.outputLines[len(m.outputLines)-m.maxLines:]
	}

	// 更新 viewport 内容
	m.viewport.SetContent(strings.Join(m.outputLines, "\n"))
	m.viewport.GotoBottom()
}

// AppendOutputs 批量追加输出内容
func (m *Model) AppendOutputs(lines []types.OutputLine) {
	if len(lines) == 0 {
		return
	}

	for _, l := range lines {
		line := l.Line
		if l.IsError {
			line = styles.ErrorTextStyle.Render(line)
		}
		m.outputLines = append(m.outputLines, line)
	}

	// 限制行数
	if len(m.outputLines) > m.maxLines {
		m.outputLines = m.outputLines[len(m.outputLines)-m.maxLines:]
	}

	// 更新 viewport 内容
	m.viewport.SetContent(strings.Join(m.outputLines, "\n"))
	m.viewport.GotoBottom()
}

// ClearOutput 清空输出
func (m *Model) ClearOutput() {
	m.outputLines = make([]string, 0)
	m.viewport.SetContent("")
}

// GetTaskID 获取任务 ID
func (m Model) GetTaskID() string {
	return m.taskID
}

// GetStatus 获取任务状态
func (m Model) GetStatus() types.TaskStatus {
	return m.status
}

// GetOutputLines 获取原始输出行（不含样式）
func (m Model) GetOutputLines() []string {
	return m.outputLines
}

// GetRawOutputLines 获取未格式化的输出行
func (m Model) GetRawOutputLines() []string {
	result := make([]string, len(m.outputLines))
	copy(result, m.outputLines)
	return result
}

// Init 实现 tea.Model 接口
func (m Model) Init() tea.Cmd {
	return nil
}

// Update 实现 tea.Model 接口
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case types.OutputMsg:
		if msg.TaskID == m.taskID {
			m.AppendOutput(msg.Line, msg.IsError)
		}

	case types.OutputBatchMsg:
		if msg.TaskID == m.taskID {
			m.AppendOutputs(msg.Lines)
		}

	case types.TaskStatusMsg:
		if msg.TaskID == m.taskID {
			m.SetStatus(msg.Status)
		}

	case types.TaskProgressMsg:
		if msg.TaskID == m.taskID {
			m.SetProgress(msg.Current, msg.Total)
		}
	}

	// 更新 viewport
	if m.ready {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	// 更新进度条
	progressModel, cmd := m.progress.Update(msg)
	m.progress = progressModel.(progress.Model)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}
