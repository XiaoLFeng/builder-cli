package todolist

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/xiaolfeng/builder-cli/internal/styles"
	"github.com/xiaolfeng/builder-cli/internal/types"
)

// View 渲染 Todo List 组件
func (m Model) View() string {
	if len(m.tasks) == 0 {
		return m.renderEmpty()
	}

	var b strings.Builder

	// 计算可显示的任务范围
	start := m.scrollIndex
	end := start + m.maxVisible
	if end > len(m.tasks) {
		end = len(m.tasks)
	}

	// 渲染任务列表
	for i := start; i < end; i++ {
		task := m.tasks[i]
		b.WriteString(m.renderTask(task, i))
		if i < end-1 {
			b.WriteString("\n")
		}
	}

	// 如果有更多任务，显示滚动提示
	if len(m.tasks) > m.maxVisible {
		scrollInfo := fmt.Sprintf("  %s 显示 %d-%d / %d",
			styles.IconBullet, start+1, end, len(m.tasks))
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Foreground(styles.MutedColor).Render(scrollInfo))
	}

	return b.String()
}

// renderTask 渲染单个任务
func (m Model) renderTask(task Task, index int) string {
	icon := statusIcon(task.Status)
	name := task.Name

	// 根据状态设置名称样式
	var nameStyle lipgloss.Style
	switch task.Status {
	case types.StatusSuccess:
		nameStyle = styles.SuccessTextStyle
	case types.StatusFailed:
		nameStyle = styles.ErrorTextStyle
	case types.StatusRunning:
		nameStyle = styles.WarningTextStyle
	default:
		nameStyle = lipgloss.NewStyle().Foreground(styles.TextColor)
	}

	// 截断过长的名称
	maxNameWidth := m.width - 30 // 留出状态和时间的空间
	if maxNameWidth < 20 {
		maxNameWidth = 20
	}
	if len(name) > maxNameWidth {
		name = name[:maxNameWidth-3] + "..."
	}

	// 格式化状态文本
	statusText := m.formatStatus(task)

	// 构建任务行
	line := fmt.Sprintf("  %s %s", icon, nameStyle.Render(name))

	// 添加状态文本（右对齐）
	padding := m.width - lipgloss.Width(line) - lipgloss.Width(statusText) - 4
	if padding < 1 {
		padding = 1
	}

	return line + strings.Repeat(" ", padding) + statusText
}

// statusIcon 返回状态图标
func statusIcon(s types.TaskStatus) string {
	switch s {
	case types.StatusPending:
		return styles.IconPending
	case types.StatusRunning:
		return styles.IconRunning
	case types.StatusSuccess:
		return styles.IconSuccess
	case types.StatusFailed:
		return styles.IconFailed
	case types.StatusSkipped:
		return styles.IconSkipped
	default:
		return styles.IconPending
	}
}

// formatStatus 格式化状态文本
func (m Model) formatStatus(task Task) string {
	var statusStyle lipgloss.Style
	var statusText string

	switch task.Status {
	case types.StatusPending:
		statusStyle = lipgloss.NewStyle().Foreground(styles.MutedColor)
		statusText = "等待中"
	case types.StatusRunning:
		statusStyle = styles.WarningTextStyle
		// 显示运行时间
		duration := task.Duration()
		statusText = fmt.Sprintf("进行中 %s", formatDuration(duration))
	case types.StatusSuccess:
		statusStyle = styles.SuccessTextStyle
		statusText = fmt.Sprintf("完成 %s", formatDuration(task.Duration()))
	case types.StatusFailed:
		statusStyle = styles.ErrorTextStyle
		statusText = "失败"
	case types.StatusSkipped:
		statusStyle = lipgloss.NewStyle().Foreground(styles.MutedColor)
		statusText = "跳过"
	default:
		statusStyle = lipgloss.NewStyle().Foreground(styles.MutedColor)
		statusText = "未知"
	}

	return statusStyle.Render(statusText)
}

// renderEmpty 渲染空状态
func (m Model) renderEmpty() string {
	return lipgloss.NewStyle().
		Foreground(styles.MutedColor).
		Italic(true).
		Render("  暂无任务")
}

// formatDuration 格式化时间
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return "<1s"
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		m := int(d.Minutes())
		s := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm%ds", m, s)
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh%dm", h, m)
}
