package styles

import "github.com/charmbracelet/lipgloss"

// 颜色定义
var (
	PrimaryColor   = lipgloss.Color("#7D56F4") // 主色调 - 紫色
	SecondaryColor = lipgloss.Color("#5B8DEF") // 次要色 - 蓝色
	SuccessColor   = lipgloss.Color("#73F59F") // 成功 - 绿色
	WarningColor   = lipgloss.Color("#EDFF82") // 警告 - 黄色
	ErrorColor     = lipgloss.Color("#FF6B6B") // 错误 - 红色
	SubtleColor    = lipgloss.Color("#383838") // 微妙色 - 深灰
	TextColor      = lipgloss.Color("#EEEEEE") // 文本色 - 浅灰
	MutedColor     = lipgloss.Color("#626262") // 静音色 - 中灰
)

// 通用样式
var (
	// 应用标题样式
	AppTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(PrimaryColor).
			Background(lipgloss.Color("#1a1a2e")).
			Padding(0, 1)

	// 版本号样式
	VersionStyle = lipgloss.NewStyle().
			Foreground(MutedColor).
			Italic(true)

	// 帮助提示样式
	HelpStyle = lipgloss.NewStyle().
			Foreground(MutedColor)

	// 分隔线样式
	DividerStyle = lipgloss.NewStyle().
			Foreground(SubtleColor)
)

// 卡片与边框样式
var (
	// 普通卡片边框
	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryColor).
			Padding(1, 2)

	// 成功状态卡片
	SuccessCardStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(SuccessColor).
				Padding(1, 2)

	// 错误状态卡片
	ErrorCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ErrorColor).
			Padding(1, 2)

	// 运行中状态卡片
	RunningCardStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(WarningColor).
				Padding(1, 2)
)

// 文本样式
var (
	// 标题样式
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(TextColor).
			MarginBottom(1)

	// 副标题样式
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(MutedColor)

	// 成功文本
	SuccessTextStyle = lipgloss.NewStyle().
				Foreground(SuccessColor)

	// 错误文本
	ErrorTextStyle = lipgloss.NewStyle().
			Foreground(ErrorColor)

	// 警告文本
	WarningTextStyle = lipgloss.NewStyle().
				Foreground(WarningColor)

	// 高亮文本
	HighlightStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)
)

// 状态图标
var (
	IconPending = lipgloss.NewStyle().Foreground(MutedColor).Render("○")
	IconRunning = lipgloss.NewStyle().Foreground(WarningColor).Render("●")
	IconSuccess = lipgloss.NewStyle().Foreground(SuccessColor).Render("✓")
	IconFailed  = lipgloss.NewStyle().Foreground(ErrorColor).Render("✗")
	IconSkipped = lipgloss.NewStyle().Foreground(MutedColor).Render("⊘")
	IconArrow   = lipgloss.NewStyle().Foreground(PrimaryColor).Render("→")
	IconBullet  = lipgloss.NewStyle().Foreground(SecondaryColor).Render("•")
	IconSpinner = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
)

// 状态栏样式
var (
	StatusBarStyle = lipgloss.NewStyle().
			Background(SubtleColor).
			Foreground(TextColor).
			Padding(0, 1)

	StatusItemStyle = lipgloss.NewStyle().
			Foreground(TextColor).
			Padding(0, 1)

	StatusKeyStyle = lipgloss.NewStyle().
			Foreground(MutedColor)

	StatusValueStyle = lipgloss.NewStyle().
				Foreground(PrimaryColor).
				Bold(true)
)

// 进度条样式
var (
	ProgressFilledStyle = lipgloss.NewStyle().
				Foreground(PrimaryColor)

	ProgressEmptyStyle = lipgloss.NewStyle().
				Foreground(SubtleColor)

	ProgressPercentStyle = lipgloss.NewStyle().
				Foreground(TextColor).
				Width(6).
				Align(lipgloss.Right)
)

// 布局辅助函数
func RenderDivider(width int) string {
	return DividerStyle.Render(lipgloss.NewStyle().
		Width(width).
		Render("─────────────────────────────────────────────────────────────────────"))
}

// RenderBox 渲染带边框的盒子
func RenderBox(content string, width int, style lipgloss.Style) string {
	return style.Width(width).Render(content)
}

// CenterText 居中文本
func CenterText(text string, width int) string {
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(text)
}

// TruncateText 截断文本（带省略号）
func TruncateText(text string, maxWidth int) string {
	if len(text) <= maxWidth {
		return text
	}
	if maxWidth <= 3 {
		return text[:maxWidth]
	}
	return text[:maxWidth-3] + "..."
}
