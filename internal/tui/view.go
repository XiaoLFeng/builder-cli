package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/xiaolfeng/builder-cli/internal/tui/components/taskcard"
	"github.com/xiaolfeng/builder-cli/pkg/version"
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

	// 帮助覆盖层（可折叠）
	if m.showHelp {
		b.WriteString("\n\n")
		b.WriteString(m.renderHelpOverlay())
	}

	return b.String()
}

// renderHeader 渲染标题栏
func (m *Model) renderHeader() string {
	// 左侧：应用名称
	title := AppTitleStyle.Render("⚡ xbuilder")
	ver := version.Version
	if ver == "" {
		ver = "dev"
	}
	if !strings.HasPrefix(ver, "v") {
		ver = "v" + ver
	}
	versionText := VersionStyle.Render(" " + ver)

	// 右侧：帮助提示
	help := HelpStyle.Render("[q] 退出  [?] 帮助")

	// 计算间距
	leftPart := title + versionText
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

	// 实时日志终端（仅当空间足够时显示）
	if m.shouldShowTerminal() {
		b.WriteString(m.terminal.RenderWithTitle("实时日志"))
		b.WriteString("\n\n")
	}

	// 进度条（始终显示）
	b.WriteString(m.progressBar.RenderWithTitle("Overall Progress"))
	b.WriteString("\n\n")

	// 任务列表（始终显示）
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

// renderHelpOverlay 渲染帮助面板
func (m *Model) renderHelpOverlay() string {
	width := m.width
	if width < 30 {
		width = 30
	}
	boxWidth := width - 6
	if boxWidth > 70 {
		boxWidth = 70
	}

	rows := []string{
		formatBinding("启动/退出", m.keys.Enter, m.keys.Quit),
		formatBinding("日志分页", m.keys.LogPrev, m.keys.LogNext),
		formatBinding("全部日志", m.keys.LogAll),
		formatBinding("恢复自动滚动", m.keys.LogResume),
		formatBinding("日志滚动", m.keys.Up, m.keys.Down, m.keys.PageUp, m.keys.PageDown, m.keys.Home, m.keys.End),
		formatBinding("任务列表滚动", m.keys.ScrollUp, m.keys.ScrollDown),
		formatBinding("帮助", m.keys.Help),
	}

	content := strings.Join(rows, "\n")
	hint := lipgloss.NewStyle().Foreground(MutedColor).Render("按 ? 关闭")
	content += "\n\n" + hint

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryColor).
		Padding(1, 2).
		Width(boxWidth).
		Align(lipgloss.Left)

	return CenterText(style.Render(content), m.width)
}

// formatBinding 将多组快捷键渲染成一行
func formatBinding(label string, bindings ...key.Binding) string {
	var keys []string
	for _, b := range bindings {
		for _, k := range b.Keys() {
			keys = append(keys, k)
		}
	}
	return fmt.Sprintf("%-10s %s", label, strings.Join(keys, " / "))
}
