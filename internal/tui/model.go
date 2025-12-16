package tui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/xiaolfeng/builder-cli/internal/config"
	"github.com/xiaolfeng/builder-cli/internal/pipeline"
	"github.com/xiaolfeng/builder-cli/internal/tui/components/progressbar"
	"github.com/xiaolfeng/builder-cli/internal/tui/components/statusbar"
	"github.com/xiaolfeng/builder-cli/internal/tui/components/taskcard"
	"github.com/xiaolfeng/builder-cli/internal/tui/components/terminal"
	"github.com/xiaolfeng/builder-cli/internal/tui/components/todolist"
)

// State 应用状态
type State int

const (
	StateInit State = iota
	StateRunning
	StateCompleted
	StateFailed
)

// Model 主 TUI Model
type Model struct {
	// 窗口尺寸
	width  int
	height int

	// 子组件
	todoList    todolist.Model
	terminal    terminal.Model // 实时日志终端
	progressBar progressbar.Model
	taskCards   map[string]*taskcard.Model // taskID -> taskCard
	statusBar   statusbar.Model

	// 业务状态
	config       *config.Config
	pipeline     *pipeline.Pipeline
	state        State
	currentStage int
	spinnerIndex int

	// 上下文和取消
	ctx    context.Context
	cancel context.CancelFunc

	// tea.Program 引用（用于向 pipeline 传递）
	program *tea.Program

	// 快捷键
	keys KeyMap

	// 错误信息
	err          error
	failedTaskID string   // 失败任务的 ID
	failedOutput []string // 失败任务的输出日志

	// 退出标记
	quitting bool
}

// New 创建新的主 Model
func New(cfg *config.Config) Model {
	ctx, cancel := context.WithCancel(context.Background())

	// 创建 pipeline
	p := pipeline.New(cfg)

	// 创建任务列表
	tasks := make([]todolist.Task, 0)
	for _, task := range p.GetAllTasks() {
		tasks = append(tasks, todolist.Task{
			ID:     task.ID,
			Name:   task.Name,
			Status: StatusPending,
		})
	}

	// 创建 taskCards map
	taskCards := make(map[string]*taskcard.Model)

	// 创建 terminal 组件并注册所有任务名称
	term := terminal.New()
	for _, task := range p.GetAllTasks() {
		term.RegisterTask(task.ID, task.Name)
	}

	return Model{
		todoList:    todolist.New().WithTasks(tasks),
		terminal:    term,
		progressBar: progressbar.New(),
		taskCards:   taskCards,
		statusBar:   statusbar.New(),
		config:      cfg,
		pipeline:    p,
		state:       StateInit,
		ctx:         ctx,
		cancel:      cancel,
		keys:        DefaultKeyMap,
	}
}

// SetProgram 设置 tea.Program 引用
func (m *Model) SetProgram(p *tea.Program) {
	m.program = p
	m.pipeline.SetProgram(p)
}

// Init 实现 tea.Model 接口
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		m.tickCmd(),
		// 自动开始构建，不需要等待用户按 Enter
		func() tea.Msg { return startPipelineMsg{} },
	)
}

// Update 实现 tea.Model 接口
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateComponentSizes()

	case TickMsg:
		m.spinnerIndex = (m.spinnerIndex + 1) % len(IconSpinner)
		cmds = append(cmds, m.tickCmd())

	case OutputMsg:
		// 确保 taskCard 存在（首次输出时创建）
		m.ensureTaskCard(msg.TaskID)
		// 注意：实际的输出追加由 taskCard.Update 处理，避免重复

	case TaskStatusMsg:
		m.handleTaskStatusMsg(msg)

	case TaskProgressMsg:
		m.handleTaskProgressMsg(msg)

	case StageStartMsg:
		m.currentStage = msg.StageIndex
		m.statusBar.SetStage(msg.StageName, msg.StageIndex, len(m.pipeline.GetStages()))

	case StageCompleteMsg:
		// 阶段完成

	case PipelineCompleteMsg:
		if msg.Success {
			m.state = StateCompleted
			// 成功完成时，给用户一点时间看结果，然后可以按任意键退出
		} else {
			m.state = StateFailed
			m.err = msg.Error
			// 失败时自动退出，显示错误信息
			m.quitting = true
			return m, tea.Quit
		}

	case startPipelineMsg:
		m.state = StateRunning
		m.statusBar.Start()
		cmds = append(cmds, m.runPipeline())

	case pipelineErrorMsg:
		m.state = StateFailed
		m.err = msg.err
		// 错误时自动退出
		m.quitting = true
		return m, tea.Quit
	}

	// 更新子组件
	var cmd tea.Cmd

	m.todoList, cmd = m.todoList.Update(msg)
	cmds = append(cmds, cmd)

	m.terminal, cmd = m.terminal.Update(msg)
	cmds = append(cmds, cmd)

	m.progressBar, cmd = m.progressBar.Update(msg)
	cmds = append(cmds, cmd)

	m.statusBar, cmd = m.statusBar.Update(msg)
	cmds = append(cmds, cmd)

	// 更新所有 taskCards
	for id, card := range m.taskCards {
		newCard, cardCmd := card.Update(msg)
		m.taskCards[id] = &newCard
		cmds = append(cmds, cardCmd)
	}

	return m, tea.Batch(cmds...)
}

// handleKeyMsg 处理按键消息
func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		m.quitting = true
		if m.cancel != nil {
			m.cancel()
		}
		return m, tea.Quit

	case key.Matches(msg, m.keys.Enter):
		if m.state == StateInit {
			return m, func() tea.Msg { return startPipelineMsg{} }
		}
	}

	return m, nil
}

// ensureTaskCard 确保 taskCard 存在
func (m *Model) ensureTaskCard(taskID string) {
	if _, exists := m.taskCards[taskID]; !exists {
		// 查找任务名称
		var taskName string
		for _, task := range m.pipeline.GetAllTasks() {
			if task.ID == taskID {
				taskName = task.Name
				break
			}
		}
		newCard := taskcard.New(taskID, taskName)
		newCard.SetSize(m.getCardWidth(), m.getCardHeight())
		m.taskCards[taskID] = &newCard
	}
}

// handleTaskStatusMsg 处理任务状态消息
func (m *Model) handleTaskStatusMsg(msg TaskStatusMsg) {
	m.todoList.UpdateTaskStatus(msg.TaskID, msg.Status)

	// 更新进度
	completed, total := m.todoList.GetProgress()
	m.progressBar.SetProgress(completed, total)
	m.statusBar.SetTasks(completed, total)

	// 更新 taskCard 状态
	if card, exists := m.taskCards[msg.TaskID]; exists {
		card.SetStatus(msg.Status)

		// 如果任务失败，保存任务ID和输出日志
		if msg.Status == StatusFailed {
			m.failedTaskID = msg.TaskID
			m.failedOutput = card.GetOutputLines()
		}
	}
}

// handleTaskProgressMsg 处理任务进度消息
func (m *Model) handleTaskProgressMsg(msg TaskProgressMsg) {
	if card, exists := m.taskCards[msg.TaskID]; exists {
		card.SetProgress(msg.Current, msg.Total)
	}
}

// updateComponentSizes 更新组件尺寸
func (m *Model) updateComponentSizes() {
	// TodoList 高度（根据终端高度做自适应，避免挤压头部/底部区域）
	todoHeight := m.getTodoHeight()
	m.todoList.SetSize(m.width-4, todoHeight)

	// Terminal 终端区域（仅当需要显示时设置尺寸）
	terminalHeight := m.getTerminalHeight()
	if terminalHeight > 0 {
		m.terminal.SetSize(m.width-4, terminalHeight)
	}

	// 进度条
	m.progressBar.SetSize(m.width - 8)

	// 状态栏
	m.statusBar.SetSize(m.width)

	// TaskCards（保留但不再主要显示）
	cardWidth := m.getCardWidth()
	cardHeight := m.getCardHeight()
	for _, card := range m.taskCards {
		card.SetSize(cardWidth, cardHeight)
	}
}

// getTodoHeight 获取任务队列区域高度（外框高度）
func (m Model) getTodoHeight() int {
	// 期望高度：任务数 + 标题/边框占用（最多 8 行）
	desired := min(len(m.todoList.GetTasks())+2, 8)

	// 运行态 UI（不含实时日志）固定占用：
	// 标题栏+分隔线+空行(3) + 进度条(2) + 间距(1) + 底部分隔线+状态栏(2) = 8
	maxFit := m.height - 8
	if maxFit < 0 {
		maxFit = 0
	}
	if desired > maxFit {
		desired = maxFit
	}

	// 保底高度：尽量保证能显示标题+至少一行内容（边框 2 + 标题 1 + 内容 1 = 4）
	if desired > 0 && desired < 4 {
		desired = min(4, maxFit)
	}

	return desired
}

// getTerminalHeight 获取终端区域高度
// 返回 0 表示空间不足，应隐藏实时日志区域
func (m Model) getTerminalHeight() int {
	todoHeight := m.getTodoHeight()

	// 固定占用（不含日志框本体）：
	// 标题栏+分隔线+空行(3)
	// + 日志标题(1)
	// + 日志与进度条间距(1)
	// + 进度条(2)
	// + 进度条与任务队列间距(1)
	// + 任务队列(todoHeight)
	// + 底部分隔线+状态栏(2)
	fixed := 3 + 1 + 1 + 2 + 1 + todoHeight + 2
	availableHeight := m.height - fixed

	// 低于 8 行时隐藏（避免 viewport 最小高度修正导致反向溢出）
	if availableHeight < 8 {
		return 0
	}

	return availableHeight
}

// shouldShowTerminal 判断是否应该显示实时日志区域
func (m Model) shouldShowTerminal() bool {
	return m.getTerminalHeight() > 0
}

// getCardWidth 获取卡片宽度
func (m Model) getCardWidth() int {
	// 根据运行中的任务数量决定宽度
	runningCount := m.getRunningTaskCount()
	if runningCount <= 1 {
		return m.width - 4
	}
	return (m.width - 6) / min(runningCount, 2)
}

// getCardHeight 获取卡片高度
func (m Model) getCardHeight() int {
	return max(m.height-25, 8)
}

// getRunningTaskCount 获取运行中的任务数量
func (m Model) getRunningTaskCount() int {
	count := 0
	for _, card := range m.taskCards {
		if card.GetStatus() == StatusRunning {
			count++
		}
	}
	return count
}

// runPipeline 运行流水线
func (m Model) runPipeline() tea.Cmd {
	return func() tea.Msg {
		// 设置 program
		// 注意：这里需要从外部设置 program
		if err := m.pipeline.Run(m.ctx); err != nil {
			return pipelineErrorMsg{err}
		}
		return nil
	}
}

// tickCmd 返回定时器命令
func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// 内部消息类型
type startPipelineMsg struct{}
type pipelineErrorMsg struct{ err error }

// helper 函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// GetError 获取错误信息
func (m Model) GetError() error {
	return m.err
}

// GetFailedTaskID 获取失败任务的 ID
func (m Model) GetFailedTaskID() string {
	return m.failedTaskID
}

// GetFailedOutput 获取失败任务的输出日志
func (m Model) GetFailedOutput() []string {
	return m.failedOutput
}

// GetFailedTaskName 获取失败任务的名称
func (m Model) GetFailedTaskName() string {
	if m.failedTaskID == "" {
		return ""
	}
	for _, task := range m.pipeline.GetAllTasks() {
		if task.ID == m.failedTaskID {
			return task.Name
		}
	}
	return m.failedTaskID
}

// IsFailed 检查是否失败
func (m Model) IsFailed() bool {
	return m.state == StateFailed
}
