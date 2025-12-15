package terminal

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/xiaolfeng/builder-cli/internal/styles"
)

// 终端样式
var (
	// 终端框样式 - 模拟真实终端外观
	terminalBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#444444")).
				Background(lipgloss.Color("#1a1a1a")).
				Padding(0, 1)

	// viewport 内容样式
	viewportStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1a1a1a"))

	// 标题样式
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(styles.TextColor)

	// 日志数量样式
	countStyle = lipgloss.NewStyle().
			Foreground(styles.MutedColor).
			Italic(true)

	// 自动滚动指示样式
	autoScrollStyle = lipgloss.NewStyle().
			Foreground(styles.SuccessColor).
			Bold(true)

	// 手动滚动指示样式
	manualScrollStyle = lipgloss.NewStyle().
				Foreground(styles.WarningColor).
				Bold(true)

	// 任务标签样式
	taskLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5B8DEF")).
			Bold(true)

	// 普通日志行样式
	normalLineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCCCCC"))

	// 错误日志行样式
	errorLineStyle = lipgloss.NewStyle().
			Foreground(styles.ErrorColor)

	// 空内容提示样式
	emptyHintStyle = lipgloss.NewStyle().
			Foreground(styles.MutedColor).
			Italic(true)
)
