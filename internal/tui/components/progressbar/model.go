package progressbar

import (
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

// Model 进度条组件模型
type Model struct {
	progress progress.Model
	current  int
	total    int
	width    int
	message  string
}

// New 创建新的进度条组件
func New() Model {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithoutPercentage(),
	)
	return Model{
		progress: p,
		total:    1, // 避免除零
	}
}

// SetSize 设置组件宽度
func (m *Model) SetSize(width int) {
	m.width = width
	m.progress.Width = width - 20 // 留出百分比和消息的空间
	if m.progress.Width < 20 {
		m.progress.Width = 20
	}
}

// SetProgress 设置进度
func (m *Model) SetProgress(current, total int) {
	m.current = current
	m.total = total
	if m.total <= 0 {
		m.total = 1
	}
}

// SetMessage 设置进度消息
func (m *Model) SetMessage(msg string) {
	m.message = msg
}

// GetPercent 获取百分比
func (m Model) GetPercent() float64 {
	if m.total == 0 {
		return 0
	}
	return float64(m.current) / float64(m.total)
}

// Init 实现 tea.Model 接口
func (m Model) Init() tea.Cmd {
	return nil
}

// Update 实现 tea.Model 接口
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width)
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	}
	return m, nil
}
