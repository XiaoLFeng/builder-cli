package statusbar

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Model 状态栏组件模型
type Model struct {
	width       int
	startTime   time.Time
	currentTime time.Time
	stageName   string
	stageIndex  int
	totalStages int
	tasksDone   int
	totalTasks  int
	isRunning   bool
}

// New 创建新的状态栏组件
func New() Model {
	return Model{
		startTime:   time.Now(),
		currentTime: time.Now(),
	}
}

// SetSize 设置状态栏宽度
func (m *Model) SetSize(width int) {
	m.width = width
}

// Start 开始计时
func (m *Model) Start() {
	m.startTime = time.Now()
	m.currentTime = time.Now()
	m.isRunning = true
}

// Stop 停止计时
func (m *Model) Stop() {
	m.currentTime = time.Now()
	m.isRunning = false
}

// SetStage 设置当前阶段
func (m *Model) SetStage(name string, index, total int) {
	m.stageName = name
	m.stageIndex = index
	m.totalStages = total
}

// SetTasks 设置任务进度
func (m *Model) SetTasks(done, total int) {
	m.tasksDone = done
	m.totalTasks = total
}

// Elapsed 返回已用时间
func (m Model) Elapsed() time.Duration {
	if m.isRunning {
		return time.Since(m.startTime)
	}
	return m.currentTime.Sub(m.startTime)
}

// Init 实现 tea.Model 接口
func (m Model) Init() tea.Cmd {
	return nil
}

// Update 实现 tea.Model 接口
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg.(type) {
	case tickMsg:
		if m.isRunning {
			m.currentTime = time.Now()
		}
	}
	return m, nil
}

// tickMsg 内部计时消息
type tickMsg time.Time

// TickCmd 返回定时器命令
func TickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
