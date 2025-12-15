package taskcard

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/xiaolfeng/builder-cli/internal/styles"
)

// 卡片样式
var (
	// 默认卡片样式
	defaultCardStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(styles.PrimaryColor).
				Padding(0, 1)

	// 运行中卡片样式
	runningCardStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(styles.WarningColor).
				Padding(0, 1)

	// 成功卡片样式
	successCardStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(styles.SuccessColor).
				Padding(0, 1)

	// 失败卡片样式
	errorCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.ErrorColor).
			Padding(0, 1)

	// viewport 样式
	viewportStyle = lipgloss.NewStyle().
			Foreground(styles.TextColor)
)
