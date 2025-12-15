package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/xiaolfeng/builder-cli/internal/tui/components/taskcard"
)

// View 实现 tea.Model 接口
func (m *Model) View() string {
	if m.quitting {
		return m.renderQuitMessage()
	}

	var b strings.Builder

	// 标题栏
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// 分隔线
	b.WriteString(RenderDivider(m.width))
	b.WriteString("\n\n")

	// 主内容区
	switch m.state {
	case StateInit:
		b.WriteString(m.renderInitView())
	case StateRunning:
		b.WriteString(m.renderRunningView())
	case StateCompleted:
		b.WriteString(m.renderCompletedView())
	case StateFailed:
		b.WriteString(m.renderFailedView())
	}

	// 分隔线
	b.WriteString("\n")
	b.WriteString(RenderDivider(m.width))
	b.WriteString("\n")

	// 状态栏
	b.WriteString(m.statusBar.View())

	return b.String()
}

// renderHeader 渲染标题栏
func (m *Model) renderHeader() string {
	// 左侧：应用名称
	title := AppTitleStyle.Render("⚡ xbuilder")
	version := VersionStyle.Render(" v1.0.0")

	// 右侧：帮助提示
	help := HelpStyle.Render("[q] 退出  [?] 帮助")

	// 计算间距
	leftPart := title + version
	rightPart := help
	spacing := m.width - lipgloss.Width(leftPart) - lipgloss.Width(rightPart) - 2

	if spacing < 0 {
		spacing = 0
	}

	return leftPart + strings.Repeat(" ", spacing) + rightPart
}

// renderInitView 渲染初始化视图
func (m *Model) renderInitView() string {
	var b strings.Builder

	// 项目信息
	projectName := m.config.Project.Name
	if projectName == "" {
		projectName = "未命名项目"
	}

	b.WriteString(TitleStyle.Render("📦 项目: " + projectName))
	b.WriteString("\n\n")

	// 任务列表预览
	b.WriteString(m.todoList.RenderWithTitle("任务队列", m.width-2))
	b.WriteString("\n\n")

	// 启动提示
	hint := lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true).
		Render("⏳ 正在启动构建...")

	b.WriteString(CenterText(hint, m.width))

	return b.String()
}

// renderRunningView 渲染运行中视图
func (m *Model) renderRunningView() string {
	var b strings.Builder

	// 实时日志终端（顶部）
	b.WriteString(m.terminal.RenderWithTitle("实时日志"))
	b.WriteString("\n\n")

	// 进度条
	b.WriteString(m.progressBar.RenderWithTitle("Overall Progress"))
	b.WriteString("\n\n")

	// 任务列表（底部）
	b.WriteString(m.todoList.RenderWithTitle("任务队列", m.width-2))

	return b.String()
}

// renderCurrentTasks 渲染当前任务卡片区
func (m *Model) renderCurrentTasks() string {
	var b strings.Builder

	// 标题
	spinner := IconSpinner[m.spinnerIndex]
	title := fmt.Sprintf("%s 当前任务", WarningTextStyle.Render(spinner))
	b.WriteString(TitleStyle.Render("🔧 " + title))
	b.WriteString("\n\n")

	// 收集运行中的任务卡片
	var runningCards []*taskcard.Model
	for _, card := range m.taskCards {
		if card.GetStatus() == StatusRunning {
			runningCards = append(runningCards, card)
		}
	}

	if len(runningCards) == 0 {
		b.WriteString(lipgloss.NewStyle().
			Foreground(MutedColor).
			Italic(true).
			Render("  等待任务启动..."))
		return b.String()
	}

	// 渲染卡片（最多显示 2 个）
	displayCount := min(len(runningCards), 2)
	cardWidth := (m.width - 6) / displayCount

	var cardViews []string
	for i := 0; i < displayCount; i++ {
		runningCards[i].SetSize(cardWidth, m.getCardHeight())
		cardViews = append(cardViews, runningCards[i].View())
	}

	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, cardViews...))

	return b.String()
}

// renderCompletedView 渲染完成视图
func (m *Model) renderCompletedView() string {
	var b strings.Builder

	// 成功消息
	successBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(SuccessColor).
		Padding(1, 3).
		Width(m.width - 4)

	content := SuccessTextStyle.Bold(true).Render("✅ 构建成功完成！")
	content += "\n\n"

	// 统计信息
	completed, total := m.todoList.GetProgress()
	content += fmt.Sprintf("完成任务: %d/%d", completed, total)

	b.WriteString(successBox.Render(content))
	b.WriteString("\n\n")

	// 任务列表
	b.WriteString(m.todoList.RenderWithTitle("任务队列", m.width-2))

	return b.String()
}

// renderFailedView 渲染失败视图
func (m *Model) renderFailedView() string {
	var b strings.Builder

	// 失败消息
	failBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ErrorColor).
		Padding(1, 3).
		Width(m.width - 4)

	content := ErrorTextStyle.Bold(true).Render("❌ 构建失败！")
	if m.err != nil {
		content += "\n\n"
		content += ErrorTextStyle.Render("错误: " + m.err.Error())
	}

	b.WriteString(failBox.Render(content))
	b.WriteString("\n\n")

	// 任务列表
	b.WriteString(m.todoList.RenderWithTitle("任务队列", m.width-2))

	return b.String()
}

// renderQuitMessage 渲染退出消息
func (m *Model) renderQuitMessage() string {
	return lipgloss.NewStyle().
		Foreground(MutedColor).
		Render("👋 再见！")
}
