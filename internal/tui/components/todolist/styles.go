package todolist

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/xiaolfeng/builder-cli/internal/styles"
)

// 组件样式
var (
	// 容器样式
	containerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.PrimaryColor).
			Padding(0, 1)

	// 标题样式
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(styles.TextColor)
)

// RenderWithTitle 带标题渲染
func (m Model) RenderWithTitle(title string, width int) string {
	m.width = width - 4 // 减去边框和内边距

	content := titleStyle.Render("📋 " + title)
	content += "\n"
	content += m.View()

	style := containerStyle.Width(width - 2)
	if m.height > 0 {
		style = style.Height(m.height)
	}
	return style.Render(content)
}
