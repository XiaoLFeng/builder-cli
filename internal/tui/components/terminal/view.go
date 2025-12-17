package terminal

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View 渲染终端组件
func (m Model) View() string {
	if !m.ready {
		return ""
	}

	return m.viewport.View()
}

// RenderWithTitle 带标题渲染
func (m Model) RenderWithTitle(title string) string {
	if !m.ready {
		return ""
	}

	var b strings.Builder

	// 标题行
	titleIcon := "💻"
	taskLabel, pageIdx := m.currentTaskLabel()
	pageInfo := ""
	if pageIdx > 0 {
		pageInfo = titleStyle.Render(fmt.Sprintf(" [%d/%d]", pageIdx, len(m.tasksOrder)))
	}
	titleText := titleStyle.Render(titleIcon + " " + title + " · " + taskLabel)

	// 日志数量和滚动状态指示
	logCountText := countStyle.Render(fmt.Sprintf("[%d lines]", m.GetLogCount())) + pageInfo

	scrollIndicator := ""
	if m.autoScroll {
		scrollIndicator = autoScrollStyle.Render(" ⬇ AUTO")
	} else {
		scrollIndicator = manualScrollStyle.Render(" ⏸ MANUAL")
	}

	// 计算间距
	leftPart := titleText
	rightPart := logCountText + scrollIndicator
	spacing := m.width - lipgloss.Width(leftPart) - lipgloss.Width(rightPart) - 4

	if spacing < 0 {
		spacing = 0
	}

	b.WriteString(leftPart + strings.Repeat(" ", spacing) + rightPart)
	b.WriteString("\n")

	// 终端内容框
	content := m.viewport.View()
	if content == "" {
		content = emptyHintStyle.Render("  等待日志输出...")
	}

	terminalBox := terminalBoxStyle.Width(m.width - 2).Render(content)
	b.WriteString(terminalBox)

	return b.String()
}

// RenderCompact 紧凑模式渲染（只显示最后几行）
func (m Model) RenderCompact(lines int) string {
	if len(m.logEntries) == 0 {
		return emptyHintStyle.Render("  等待日志输出...")
	}

	// 获取最后 n 行
	startIdx := 0
	if len(m.logEntries) > lines {
		startIdx = len(m.logEntries) - lines
	}

	var result []string
	for i := startIdx; i < len(m.logEntries); i++ {
		line := m.formatLogEntry(m.logEntries[i])
		result = append(result, line)
	}

	return strings.Join(result, "\n")
}
