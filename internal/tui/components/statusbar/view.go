package statusbar

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/xiaolfeng/builder-cli/internal/styles"
)

// View 渲染状态栏
func (m Model) View() string {
	if m.width < 40 {
		return m.renderCompact()
	}

	// 时间
	elapsed := m.Elapsed()
	timeStr := formatDuration(elapsed)
	timeItem := m.renderItem("⏱", "用时", timeStr)

	// 阶段
	var stageStr string
	if m.totalStages > 0 {
		stageStr = fmt.Sprintf("%s (%d/%d)", m.stageName, m.stageIndex+1, m.totalStages)
	} else {
		stageStr = m.stageName
	}
	stageItem := m.renderItem("📦", "阶段", stageStr)

	// 任务
	taskStr := fmt.Sprintf("%d/%d", m.tasksDone, m.totalTasks)
	taskItem := m.renderItem("🔄", "任务", taskStr)

	// 状态
	var statusStr string
	var statusStyle lipgloss.Style
	if m.isRunning {
		statusStr = "运行中"
		statusStyle = styles.WarningTextStyle
	} else {
		statusStr = "已停止"
		statusStyle = lipgloss.NewStyle().Foreground(styles.MutedColor)
	}
	statusItem := m.renderItemStyled("💫", "状态", statusStr, statusStyle)

	// 组合
	items := []string{timeItem, stageItem, taskItem, statusItem}
	content := strings.Join(items, "  │  ")

	return statusBarStyle.Width(m.width).Render(content)
}

// renderCompact 紧凑模式渲染
func (m Model) renderCompact() string {
	elapsed := m.Elapsed()
	return statusBarStyle.Width(m.width).Render(
		fmt.Sprintf("⏱ %s  📦 %s  🔄 %d/%d",
			formatDuration(elapsed),
			m.stageName,
			m.tasksDone,
			m.totalTasks,
		),
	)
}

// renderItem 渲染单个状态项
func (m Model) renderItem(icon, key, value string) string {
	keyStyle := lipgloss.NewStyle().Foreground(styles.MutedColor)
	valueStyle := lipgloss.NewStyle().Foreground(styles.PrimaryColor).Bold(true)

	return fmt.Sprintf("%s %s: %s", icon, keyStyle.Render(key), valueStyle.Render(value))
}

// renderItemStyled 渲染带自定义样式的状态项
func (m Model) renderItemStyled(icon, key, value string, valueStyle lipgloss.Style) string {
	keyStyle := lipgloss.NewStyle().Foreground(styles.MutedColor)

	return fmt.Sprintf("%s %s: %s", icon, keyStyle.Render(key), valueStyle.Render(value))
}

// 状态栏样式
var statusBarStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("#1a1a2e")).
	Foreground(styles.TextColor).
	Padding(0, 1)

// formatDuration 格式化时间
func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}
