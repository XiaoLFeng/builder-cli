package executor

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/xiaolfeng/builder-cli/internal/config"
)

// ShellExecutor Shell 命令执行器
type ShellExecutor struct {
	*BaseExecutor
	command string
	script  string
}

// NewShellExecutor 创建 Shell 执行器
func NewShellExecutor(taskName string, cfg config.TaskConfig) *ShellExecutor {
	e := &ShellExecutor{
		BaseExecutor: NewBaseExecutor(taskName, TypeShell),
		command:      cfg.Command,
		script:       cfg.Script,
	}

	// 设置工作目录
	if cfg.WorkingDir != "" {
		e.SetWorkingDir(cfg.WorkingDir)
	}

	// 设置超时（默认 10 分钟）
	if cfg.Timeout > 0 {
		e.SetTimeout(time.Duration(cfg.Timeout) * time.Second)
	} else {
		e.SetTimeout(10 * time.Minute)
	}

	return e
}

// Execute 执行 Shell 命令
func (e *ShellExecutor) Execute(ctx context.Context, handler OutputHandler) error {
	// 优先使用脚本模式
	if e.script != "" {
		return e.executeScript(ctx, handler)
	}

	// 使用命令模式
	if e.command != "" {
		return e.executeCommand(ctx, handler)
	}

	return fmt.Errorf("shell 任务必须指定 command 或 script")
}

// executeCommand 执行 Shell 命令
func (e *ShellExecutor) executeCommand(ctx context.Context, handler OutputHandler) error {
	handler("🐚 执行 Shell 命令", false)
	handler(fmt.Sprintf("📁 工作目录: %s", e.getWorkDir()), false)
	handler("", false)

	runner := NewCommandRunner(e.Name(), e.command)
	runner.SetWorkingDir(e.getWorkDir())
	runner.SetTimeout(e.GetTimeout())
	runner.SetEnv(e.GetEnv())

	return runner.Execute(ctx, handler)
}

// executeScript 执行脚本文件
func (e *ShellExecutor) executeScript(ctx context.Context, handler OutputHandler) error {
	scriptPath := e.script
	if e.workingDir != "" && !isAbsPath(scriptPath) {
		scriptPath = e.workingDir + "/" + scriptPath
	}

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("脚本文件不存在: %s", scriptPath)
	}

	handler(fmt.Sprintf("🐚 执行 Shell 脚本: %s", e.script), false)
	handler(fmt.Sprintf("📁 工作目录: %s", e.getWorkDir()), false)
	handler("", false)

	runner := NewScriptRunner(e.Name(), scriptPath)
	runner.SetWorkingDir(e.getWorkDir())
	runner.SetTimeout(e.GetTimeout())
	runner.SetEnv(e.GetEnv())

	return runner.Execute(ctx, handler)
}

// getWorkDir 获取工作目录
func (e *ShellExecutor) getWorkDir() string {
	if e.workingDir != "" {
		return e.workingDir
	}
	dir, _ := os.Getwd()
	return dir
}
