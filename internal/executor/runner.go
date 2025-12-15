package executor

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// CommandRunner 通用命令运行器，支持实时输出流
type CommandRunner struct {
	*BaseExecutor
	command string
	args    []string
	shell   bool // 是否使用 shell 执行
}

// NewCommandRunner 创建命令运行器
func NewCommandRunner(name, command string) *CommandRunner {
	return &CommandRunner{
		BaseExecutor: NewBaseExecutor(name, TypeCommand),
		command:      command,
		shell:        true, // 默认使用 shell
	}
}

// NewCommandRunnerWithArgs 创建带参数的命令运行器
func NewCommandRunnerWithArgs(name, cmd string, args []string) *CommandRunner {
	return &CommandRunner{
		BaseExecutor: NewBaseExecutor(name, TypeCommand),
		command:      cmd,
		args:         args,
		shell:        false,
	}
}

// SetShell 设置是否使用 shell 执行
func (r *CommandRunner) SetShell(shell bool) {
	r.shell = shell
}

// Execute 执行命令并实时流式输出
func (r *CommandRunner) Execute(ctx context.Context, handler OutputHandler) error {
	// 创建带超时的上下文
	execCtx, cancel := context.WithTimeout(ctx, r.GetTimeout())
	defer cancel()

	// 构建命令
	var cmd *exec.Cmd
	if r.shell {
		// 使用 shell 执行
		cmd = exec.CommandContext(execCtx, "sh", "-c", r.command)
	} else {
		cmd = exec.CommandContext(execCtx, r.command, r.args...)
	}

	// 设置工作目录
	if r.workingDir != "" {
		cmd.Dir = r.workingDir
	}

	// 设置环境变量
	cmd.Env = append(os.Environ(), r.env...)

	// 获取 stdout 和 stderr 管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("获取 stdout 管道失败: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("获取 stderr 管道失败: %w", err)
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动命令失败: %w", err)
	}

	// 使用 WaitGroup 等待输出读取完成
	var wg sync.WaitGroup

	// 异步读取 stdout
	wg.Add(1)
	go func() {
		defer wg.Done()
		r.readOutput(stdout, handler, false)
	}()

	// 异步读取 stderr
	wg.Add(1)
	go func() {
		defer wg.Done()
		r.readOutput(stderr, handler, true)
	}()

	// 等待输出读取完成
	wg.Wait()

	// 等待命令结束
	if err := cmd.Wait(); err != nil {
		// 检查是否是超时
		if execCtx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("命令执行超时 (%v)", r.GetTimeout())
		}
		// 检查是否是取消
		if execCtx.Err() == context.Canceled {
			return fmt.Errorf("命令被取消")
		}
		return fmt.Errorf("命令执行失败: %w", err)
	}

	return nil
}

// readOutput 读取输出流并调用 handler
func (r *CommandRunner) readOutput(reader io.Reader, handler OutputHandler, isError bool) {
	scanner := bufio.NewScanner(reader)
	// 增加缓冲区大小以处理长行
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if handler != nil {
			handler(line, isError)
		}
	}
}

// ScriptRunner 脚本运行器
type ScriptRunner struct {
	*CommandRunner
	scriptPath string
}

// NewScriptRunner 创建脚本运行器
func NewScriptRunner(name, scriptPath string) *ScriptRunner {
	return &ScriptRunner{
		CommandRunner: &CommandRunner{
			BaseExecutor: NewBaseExecutor(name, TypeCommand),
			command:      scriptPath,
			shell:        false,
		},
		scriptPath: scriptPath,
	}
}

// Execute 执行脚本
func (r *ScriptRunner) Execute(ctx context.Context, handler OutputHandler) error {
	// 检查脚本是否存在
	if _, err := os.Stat(r.scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("脚本文件不存在: %s", r.scriptPath)
	}

	// 使用 bash 执行脚本
	r.command = "bash"
	r.args = []string{r.scriptPath}
	r.shell = false

	return r.CommandRunner.Execute(ctx, handler)
}

// RunCommand 便捷函数：运行单个命令
func RunCommand(ctx context.Context, command string, handler OutputHandler) error {
	runner := NewCommandRunner("command", command)
	return runner.Execute(ctx, handler)
}

// RunCommandInDir 便捷函数：在指定目录运行命令
func RunCommandInDir(ctx context.Context, command, dir string, handler OutputHandler) error {
	runner := NewCommandRunner("command", command)
	runner.SetWorkingDir(dir)
	return runner.Execute(ctx, handler)
}

// RunScript 便捷函数：运行脚本
func RunScript(ctx context.Context, scriptPath string, handler OutputHandler) error {
	runner := NewScriptRunner("script", scriptPath)
	return runner.Execute(ctx, handler)
}

// ParseCommand 解析命令字符串为命令和参数
func ParseCommand(command string) (string, []string) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", nil
	}
	if len(parts) == 1 {
		return parts[0], nil
	}
	return parts[0], parts[1:]
}

// FormatDuration 格式化执行时间
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm%ds", m, s)
}
