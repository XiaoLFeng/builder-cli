package taskcard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/xiaolfeng/builder-cli/internal/styles"
	"github.com/xiaolfeng/builder-cli/internal/types"
)

// View 渲染任务卡片
func (m Model) View() string {
	if !m.ready {
		return ""
	}

	var b strings.Builder

	// 标题行
	titleLine := m.renderTitle()
	b.WriteString(titleLine)
	b.WriteString("\n")

	// 输出内容
	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	// 进度条
	if m.total > 0 {
		percent := float64(m.current) / float64(m.total)
		b.WriteString(m.progress.ViewAs(percent))
		b.WriteString(fmt.Sprintf(" %d%%", int(percent*100)))
	}

	// 根据状态选择边框样式
	var cardStyle lipgloss.Style
	switch m.status {
	case types.StatusRunning:
		cardStyle = runningCardStyle
	case types.StatusSuccess:
		cardStyle = successCardStyle
	case types.StatusFailed:
		cardStyle = errorCardStyle
	default:
		cardStyle = defaultCardStyle
	}

	return cardStyle.Width(m.width - 2).Render(b.String())
}

// renderTitle 渲染标题行
func (m Model) renderTitle() string {
	icon := m.status.Icon()
	title := m.title

	// 截断过长的标题
	maxWidth := m.width - 15
	if maxWidth < 10 {
		maxWidth = 10
	}
	if len(title) > maxWidth {
		title = title[:maxWidth-3] + "..."
	}

	// 标题样式
	var titleStyle lipgloss.Style
	switch m.status {
	case types.StatusRunning:
		titleStyle = lipgloss.NewStyle().Foreground(styles.WarningColor).Bold(true)
	case types.StatusSuccess:
		titleStyle = lipgloss.NewStyle().Foreground(styles.SuccessColor).Bold(true)
	case types.StatusFailed:
		titleStyle = lipgloss.NewStyle().Foreground(styles.ErrorColor).Bold(true)
	default:
		titleStyle = lipgloss.NewStyle().Foreground(styles.TextColor).Bold(true)
	}

	return fmt.Sprintf("%s %s", icon, titleStyle.Render(title))
}

// RenderCompact 紧凑模式渲染（不显示输出内容）
func (m Model) RenderCompact() string {
	icon := m.status.Icon()
	title := m.title

	// 截断标题
	maxWidth := m.width - 20
	if len(title) > maxWidth {
		title = title[:maxWidth-3] + "..."
	}

	// 状态文本
	statusText := m.status.String()

	// 进度
	var progressText string
	if m.total > 0 {
		percent := float64(m.current) / float64(m.total) * 100
		progressText = fmt.Sprintf(" %d%%", int(percent))
	}

	line := fmt.Sprintf("%s %s%s  %s", icon, title, progressText, statusText)

	// 根据状态选择样式
	var style lipgloss.Style
	switch m.status {
	case types.StatusRunning:
		style = lipgloss.NewStyle().Foreground(styles.WarningColor)
	case types.StatusSuccess:
		style = lipgloss.NewStyle().Foreground(styles.SuccessColor)
	case types.StatusFailed:
		style = lipgloss.NewStyle().Foreground(styles.ErrorColor)
	default:
		style = lipgloss.NewStyle().Foreground(styles.TextColor)
	}

	return style.Render(line)
}
