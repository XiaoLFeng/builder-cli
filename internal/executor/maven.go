package executor

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/xiaolfeng/builder-cli/internal/config"
)

// MavenExecutor Maven 构建执行器
type MavenExecutor struct {
	*BaseExecutor
	command string
	script  string
}

// NewMavenExecutor 创建 Maven 执行器
func NewMavenExecutor(taskName string, cfg config.TaskConfig) *MavenExecutor {
	e := &MavenExecutor{
		BaseExecutor: NewBaseExecutor(taskName, TypeMaven),
		command:      cfg.Command,
		script:       cfg.Script,
	}

	// 设置工作目录
	if cfg.WorkingDir != "" {
		e.SetWorkingDir(cfg.WorkingDir)
	}

	// 设置超时
	if cfg.Timeout > 0 {
		e.SetTimeout(time.Duration(cfg.Timeout) * time.Second)
	} else {
		e.SetTimeout(30 * time.Minute) // Maven 默认 30 分钟
	}

	return e
}

// Execute 执行 Maven 构建
func (e *MavenExecutor) Execute(ctx context.Context, handler OutputHandler) error {
	// 优先使用脚本
	if e.script != "" {
		return e.executeScript(ctx, handler)
	}

	// 使用命令
	if e.command != "" {
		return e.executeCommand(ctx, handler)
	}

	// 默认命令
	return e.executeCommand(ctx, handler)
}

// executeCommand 执行 Maven 命令
func (e *MavenExecutor) executeCommand(ctx context.Context, handler OutputHandler) error {
	command := e.command
	if command == "" {
		command = "mvn clean package -DskipTests"
	}

	handler(fmt.Sprintf("🔨 执行 Maven 命令: %s", command), false)
	handler(fmt.Sprintf("📁 工作目录: %s", e.getWorkDir()), false)
	handler("", false)

	runner := NewCommandRunner(e.Name(), command)
	runner.SetWorkingDir(e.getWorkDir())
	runner.SetTimeout(e.GetTimeout())
	runner.SetEnv(e.GetEnv())

	return runner.Execute(ctx, handler)
}

// executeScript 执行构建脚本
func (e *MavenExecutor) executeScript(ctx context.Context, handler OutputHandler) error {
	// 检查脚本是否存在
	scriptPath := e.script
	if e.workingDir != "" && !isAbsPath(scriptPath) {
		scriptPath = e.workingDir + "/" + scriptPath
	}

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("构建脚本不存在: %s", scriptPath)
	}

	handler(fmt.Sprintf("🔨 执行构建脚本: %s", e.script), false)
	handler(fmt.Sprintf("📁 工作目录: %s", e.getWorkDir()), false)
	handler("", false)

	runner := NewScriptRunner(e.Name(), scriptPath)
	runner.SetWorkingDir(e.getWorkDir())
	runner.SetTimeout(e.GetTimeout())
	runner.SetEnv(e.GetEnv())

	return runner.Execute(ctx, handler)
}

// getWorkDir 获取工作目录
func (e *MavenExecutor) getWorkDir() string {
	if e.workingDir != "" {
		return e.workingDir
	}
	dir, _ := os.Getwd()
	return dir
}

// isAbsPath 检查是否为绝对路径
func isAbsPath(path string) bool {
	return len(path) > 0 && path[0] == '/'
}
