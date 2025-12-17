package progressbar

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/xiaolfeng/builder-cli/internal/styles"
)

// View 渲染进度条
func (m Model) View() string {
	percent := m.GetPercent()

	// 进度条
	bar := m.progress.View()

	// 百分比文本
	percentText := fmt.Sprintf("%3d%%", int(percent*100))
	percentStyle := lipgloss.NewStyle().
		Foreground(styles.TextColor).
		Bold(true)

	// 进度数字
	countText := fmt.Sprintf("%d/%d", m.current, m.total)
	countStyle := lipgloss.NewStyle().
		Foreground(styles.MutedColor)

	// 组合显示
	return fmt.Sprintf("%s %s (%s)",
		bar,
		percentStyle.Render(percentText),
		countStyle.Render(countText),
	)
}

// RenderWithTitle 带标题渲染
func (m Model) RenderWithTitle(title string) string {
	var b strings.Builder

	// 标题
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.TextColor)
	b.WriteString(titleStyle.Render("📊 " + title))
	b.WriteString("\n")

	// 进度条
	b.WriteString(m.View())

	// 消息
	if m.message != "" {
		b.WriteString("\n")
		msgStyle := lipgloss.NewStyle().
			Foreground(styles.MutedColor).
			Italic(true)
		b.WriteString(msgStyle.Render("   " + m.message))
	}

	return b.String()
}
