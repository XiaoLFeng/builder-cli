package app

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xiaolfeng/builder-cli/internal/config"
	"github.com/xiaolfeng/builder-cli/internal/tui"
)

// BuildOptions 构建选项
type BuildOptions struct {
	ConfigFile   string   // 配置文件路径
	ValidateOnly bool     // 仅验证
	StageStart   int      // 开始阶段 (0-based)
	StageEnd     int      // 结束阶段 (0-based), -1 表示到最后
	OnlyTasks    []string // 仅执行指定名称的任务
	TargetServer string   // 仅部署到指定服务器（可选）
}

// RunBuild 运行构建
func RunBuild(opts BuildOptions) error {
	// 查找配置文件
	configPath := opts.ConfigFile
	if configPath == "" {
		var err error
		configPath, err = config.FindConfigFile()
		if err != nil {
			return fmt.Errorf("❌ %v", err)
		}
	} else {
		// 检查指定的配置文件是否存在
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			return fmt.Errorf("❌ 配置文件不存在: %s", configPath)
		}
	}

	fmt.Printf("📄 使用配置文件: %s\n", configPath)

	// 加载配置
	loader := config.NewLoader(configPath)
	cfg, err := loader.Load()
	if err != nil {
		return fmt.Errorf("❌ 加载配置失败: %v", err)
	}

	// 验证配置
	validator := config.NewValidator(cfg)
	if err := validator.Validate(); err != nil {
		return fmt.Errorf("❌ 配置验证失败:\n%v", err)
	}

	// 处理阶段范围
	totalStages := len(cfg.Pipeline)
	startIdx := opts.StageStart
	endIdx := opts.StageEnd

	// 验证阶段范围
	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx >= totalStages {
		return fmt.Errorf("❌ 开始阶段 %d 超出范围 (共 %d 个阶段)", startIdx+1, totalStages)
	}

	if endIdx < 0 || endIdx >= totalStages {
		endIdx = totalStages - 1
	}

	if startIdx > endIdx {
		return fmt.Errorf("❌ 开始阶段 (%d) 不能大于结束阶段 (%d)", startIdx+1, endIdx+1)
	}

	// 过滤阶段
	if startIdx > 0 || endIdx < totalStages-1 {
		cfg.Pipeline = cfg.Pipeline[startIdx : endIdx+1]
		fmt.Printf("🎯 执行阶段: %d-%d (共 %d 个阶段)\n", startIdx+1, endIdx+1, len(cfg.Pipeline))
	}

	// 过滤任务（--only 参数）
	if len(opts.OnlyTasks) > 0 {
		cfg.Pipeline = filterTasks(cfg.Pipeline, opts.OnlyTasks)
		if countTotalTasks(cfg.Pipeline) == 0 {
			return fmt.Errorf("❌ 没有找到匹配的任务: %v", opts.OnlyTasks)
		}
		fmt.Printf("🎯 仅执行任务: %v\n", opts.OnlyTasks)
	}

	// 过滤服务器（--server 参数，仅作用于 SSH 部署任务）
	if opts.TargetServer != "" {
		if _, ok := cfg.Servers[opts.TargetServer]; !ok {
			return fmt.Errorf("❌ 服务器不存在: %s", opts.TargetServer)
		}
		cfg.Pipeline = filterTasksByServer(cfg.Pipeline, opts.TargetServer)
		if countTotalTasks(cfg.Pipeline) == 0 {
			return fmt.Errorf("❌ 没有找到匹配服务器 [%s] 的任务", opts.TargetServer)
		}
		fmt.Printf("🎯 仅部署到服务器: %s\n", opts.TargetServer)
	}

	fmt.Printf("✅ 配置验证通过\n")
	fmt.Printf("📦 项目: %s\n", cfg.Project.Name)
	fmt.Printf("🔄 阶段数: %d\n\n", len(cfg.Pipeline))

	// 显示将要执行的阶段
	for i, stage := range cfg.Pipeline {
		fmt.Printf("   %d. %s\n", i+1, stage.Name)
	}
	fmt.Println()

	// 创建 TUI Model
	model := tui.New(cfg)

	// 创建 tea.Program
	p := tea.NewProgram(&model, tea.WithAltScreen())

	// 设置 program 引用，让 pipeline 可以发送消息
	model.SetProgram(p)

	// 运行 TUI
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("❌ TUI 运行失败: %v", err)
	}

	// 检查构建结果
	if m, ok := finalModel.(*tui.Model); ok && m.IsFailed() {
		// 显示美化的错误信息
		printBuildError(*m)
		return fmt.Errorf("构建失败")
	}

	// 成功完成
	fmt.Println()
	fmt.Println("✅ 构建成功完成！")

	return nil
}

// printBuildError 打印美化的构建错误信息
func printBuildError(m tui.Model) {
	// 错误样式定义
	errorTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF6B6B"))

	errorBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF6B6B")).
		Padding(0, 1).
		MarginTop(1)

	taskNameStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFE66D"))

	errorMsgStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B"))

	logHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#4ECDC4")).
		MarginTop(1)

	logBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#555555")).
		Padding(0, 0).
		Foreground(lipgloss.Color("#AAAAAA"))

	fmt.Println()

	// 错误标题
	fmt.Println(errorTitleStyle.Render("❌ 构建失败"))

	// 错误详情框
	var errorContent strings.Builder
	errorContent.WriteString("任务: ")
	errorContent.WriteString(taskNameStyle.Render(m.GetFailedTaskName()))
	errorContent.WriteString("\n")

	if err := m.GetError(); err != nil {
		errorContent.WriteString("错误: ")
		errorContent.WriteString(errorMsgStyle.Render(err.Error()))
	}

	fmt.Println(errorBoxStyle.Render(errorContent.String()))

	// 输出日志
	output := m.GetFailedOutput()
	if len(output) > 0 {
		fmt.Println(logHeaderStyle.Render("📋 任务输出日志 (最后 20 行):"))

		// 只显示最后 20 行
		startIdx := 0
		if len(output) > 20 {
			startIdx = len(output) - 20
		}

		var logContent strings.Builder
		for i := startIdx; i < len(output); i++ {
			// 移除已有的 ANSI 颜色码，避免嵌套样式问题
			line := stripAnsi(output[i])
			if i > startIdx {
				logContent.WriteString("\n")
			}
			logContent.WriteString(line)
		}

		fmt.Println(logBoxStyle.Render(logContent.String()))
	}

	fmt.Println()
}

// stripAnsi 移除 ANSI 转义序列
func stripAnsi(s string) string {
	var result strings.Builder
	inEscape := false

	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (s[i] >= 'a' && s[i] <= 'z') || (s[i] >= 'A' && s[i] <= 'Z') {
				inEscape = false
			}
			continue
		}
		result.WriteByte(s[i])
	}

	return result.String()
}

// filterTasks 根据任务名称过滤任务
func filterTasks(pipeline []config.Stage, onlyTasks []string) []config.Stage {
	// 创建任务名称集合，方便查找
	taskSet := make(map[string]bool)
	for _, name := range onlyTasks {
		taskSet[name] = true
	}

	// 过滤每个阶段中的任务
	var result []config.Stage
	for _, stage := range pipeline {
		var filteredTasks []config.Task
		for _, task := range stage.Tasks {
			if taskSet[task.Name] {
				filteredTasks = append(filteredTasks, task)
			}
		}

		// 只保留有任务的阶段
		if len(filteredTasks) > 0 {
			newStage := stage
			newStage.Tasks = filteredTasks
			result = append(result, newStage)
		}
	}

	return result
}

// filterTasksByServer 仅保留目标服务器的 SSH 任务，其他类型任务保留
func filterTasksByServer(pipeline []config.Stage, server string) []config.Stage {
	var result []config.Stage
	for _, stage := range pipeline {
		var filtered []config.Task
		for _, task := range stage.Tasks {
			if task.Type != config.TaskTypeSSH {
				filtered = append(filtered, task)
				continue
			}
			if task.Config.Server == server {
				filtered = append(filtered, task)
			}
		}
		if len(filtered) > 0 {
			newStage := stage
			newStage.Tasks = filtered
			result = append(result, newStage)
		}
	}
	return result
}

// countTotalTasks 统计总任务数
func countTotalTasks(pipeline []config.Stage) int {
	count := 0
	for _, stage := range pipeline {
		count += len(stage.Tasks)
	}
	return count
}

// ValidateConfig 验证配置文件
func ValidateConfig(configPath string) error {
	// 查找配置文件
	if configPath == "" {
		var err error
		configPath, err = config.FindConfigFile()
		if err != nil {
			return fmt.Errorf("❌ %v", err)
		}
	} else {
		// 检查指定的配置文件是否存在
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			return fmt.Errorf("❌ 配置文件不存在: %s", configPath)
		}
	}

	fmt.Printf("🔍 验证配置文件: %s\n", configPath)

	// 加载配置
	loader := config.NewLoader(configPath)
	cfg, err := loader.Load()
	if err != nil {
		return fmt.Errorf("❌ 加载配置失败: %v", err)
	}

	// 验证配置
	validator := config.NewValidator(cfg)
	if err := validator.Validate(); err != nil {
		return fmt.Errorf("❌ 配置验证失败:\n%v", err)
	}

	return nil
}

// App 应用程序 (保留向后兼容)
type App struct {
	config     *config.Config
	configPath string
}

// New 创建新的应用程序
func New() *App {
	return &App{}
}

// Run 运行应用程序
func (a *App) Run() error {
	return RunBuild(BuildOptions{})
}

// RunWithConfig 使用指定配置文件运行
func (a *App) RunWithConfig(configPath string) error {
	return RunBuild(BuildOptions{ConfigFile: configPath})
}

// ValidateConfigLegacy 仅验证配置（保留向后兼容）
func (a *App) ValidateConfigLegacy(configPath string) error {
	return ValidateConfig(configPath)
}
