package executor

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/xiaolfeng/builder-cli/internal/config"
)

// GoBuildExecutor Go 构建执行器
type GoBuildExecutor struct {
	*BaseExecutor
	goCommand  string // build/test/generate
	command    string // 自定义命令
	script     string // 自定义脚本
	goos       string
	goarch     string
	output     string
	ldflags    string
	tags       string
	cgoEnabled *bool
	goPrivate  string
	goProxy    string
	race       bool
	trimpath   bool
	mod        string
	packages   string
	verbose    bool
}

// NewGoBuildExecutor 创建 Go 构建执行器
func NewGoBuildExecutor(taskName string, cfg config.TaskConfig) *GoBuildExecutor {
	e := &GoBuildExecutor{
		BaseExecutor: NewBaseExecutor(taskName, TypeGoBuild),
		goCommand:    cfg.GoCommand,
		command:      cfg.Command,
		script:       cfg.Script,
		goos:         cfg.GoOS,
		goarch:       cfg.GoArch,
		output:       cfg.Output,
		ldflags:      cfg.LDFlags,
		tags:         cfg.BuildTags,
		cgoEnabled:   cfg.CGOEnabled,
		goPrivate:    cfg.GoPrivate,
		goProxy:      cfg.GoProxy,
		race:         cfg.Race,
		trimpath:     cfg.Trimpath,
		mod:          cfg.Mod,
		packages:     cfg.Packages,
		verbose:      cfg.GoVerbose,
	}

	// 设置工作目录
	if cfg.WorkingDir != "" {
		e.SetWorkingDir(cfg.WorkingDir)
	}

	// 设置超时（默认 15 分钟）
	if cfg.Timeout > 0 {
		e.SetTimeout(time.Duration(cfg.Timeout) * time.Second)
	} else {
		e.SetTimeout(15 * time.Minute)
	}

	// 设置 Go 相关环境变量
	e.setupGoEnv()

	return e
}

// setupGoEnv 设置 Go 环境变量
func (e *GoBuildExecutor) setupGoEnv() {
	if e.goos != "" {
		e.AddEnv("GOOS", e.goos)
	}
	if e.goarch != "" {
		e.AddEnv("GOARCH", e.goarch)
	}
	if e.cgoEnabled != nil {
		if *e.cgoEnabled {
			e.AddEnv("CGO_ENABLED", "1")
		} else {
			e.AddEnv("CGO_ENABLED", "0")
		}
	}
	if e.goPrivate != "" {
		e.AddEnv("GOPRIVATE", e.goPrivate)
	}
	if e.goProxy != "" {
		e.AddEnv("GOPROXY", e.goProxy)
	}
}

// Execute 执行 Go 构建
func (e *GoBuildExecutor) Execute(ctx context.Context, handler OutputHandler) error {
	// 优先使用脚本模式
	if e.script != "" {
		return e.executeScript(ctx, handler)
	}

	// 使用命令模式
	if e.command != "" {
		return e.executeCommand(ctx, handler)
	}

	// 默认构建命令
	return e.executeDefaultBuild(ctx, handler)
}

// executeCommand 执行自定义命令
func (e *GoBuildExecutor) executeCommand(ctx context.Context, handler OutputHandler) error {
	handler(fmt.Sprintf("🔨 执行 Go 命令: %s", e.command), false)
	handler(fmt.Sprintf("📁 工作目录: %s", e.getWorkDir()), false)
	e.printEnvInfo(handler)
	handler("", false)

	runner := NewCommandRunner(e.Name(), e.command)
	runner.SetWorkingDir(e.getWorkDir())
	runner.SetTimeout(e.GetTimeout())
	runner.SetEnv(e.GetEnv())

	return runner.Execute(ctx, handler)
}

// executeScript 执行构建脚本
func (e *GoBuildExecutor) executeScript(ctx context.Context, handler OutputHandler) error {
	scriptPath := e.script
	if e.workingDir != "" && !isAbsPath(scriptPath) {
		scriptPath = e.workingDir + "/" + scriptPath
	}

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("构建脚本不存在: %s", scriptPath)
	}

	handler(fmt.Sprintf("🔨 执行构建脚本: %s", e.script), false)
	handler(fmt.Sprintf("📁 工作目录: %s", e.getWorkDir()), false)
	e.printEnvInfo(handler)
	handler("", false)

	runner := NewScriptRunner(e.Name(), scriptPath)
	runner.SetWorkingDir(e.getWorkDir())
	runner.SetTimeout(e.GetTimeout())
	runner.SetEnv(e.GetEnv())

	return runner.Execute(ctx, handler)
}

// executeDefaultBuild 执行默认构建
func (e *GoBuildExecutor) executeDefaultBuild(ctx context.Context, handler OutputHandler) error {
	args := e.buildDefaultArgs()
	logLine := strings.Join(args, " ")

	handler(fmt.Sprintf("🔨 执行 Go %s: %s", e.getGoCommand(), logLine), false)
	handler(fmt.Sprintf("📁 工作目录: %s", e.getWorkDir()), false)
	e.printEnvInfo(handler)
	handler("", false)

	runner := NewCommandRunnerWithArgs(e.Name(), args[0], args[1:])
	runner.SetShell(false)
	runner.SetWorkingDir(e.getWorkDir())
	runner.SetTimeout(e.GetTimeout())
	runner.SetEnv(e.GetEnv())

	return runner.Execute(ctx, handler)
}

// buildDefaultArgs 根据 goCommand 构建默认命令参数
func (e *GoBuildExecutor) buildDefaultArgs() []string {
	goCmd := e.getGoCommand()

	switch goCmd {
	case "test":
		return e.buildTestArgs()
	case "generate":
		return e.buildGenerateArgs()
	default:
		return e.buildBuildArgs()
	}
}

// buildBuildArgs 构建 go build 命令参数
func (e *GoBuildExecutor) buildBuildArgs() []string {
	args := []string{"go", "build"}

	// -v 详细输出
	if e.verbose {
		args = append(args, "-v")
	}

	// -o 输出路径
	if e.output != "" {
		args = append(args, "-o", e.output)
	}

	// -ldflags
	if e.ldflags != "" {
		args = append(args, "-ldflags", e.ldflags)
	}

	// -tags
	if e.tags != "" {
		args = append(args, "-tags", e.tags)
	}

	// -race
	if e.race {
		args = append(args, "-race")
	}

	// -trimpath
	if e.trimpath {
		args = append(args, "-trimpath")
	}

	// -mod
	if e.mod != "" {
		args = append(args, "-mod="+e.mod)
	}

	// 目标包
	args = append(args, e.getPackages())

	return args
}

// buildTestArgs 构建 go test 命令参数
func (e *GoBuildExecutor) buildTestArgs() []string {
	args := []string{"go", "test"}

	// -v 详细输出
	if e.verbose {
		args = append(args, "-v")
	}

	// -race
	if e.race {
		args = append(args, "-race")
	}

	// -mod
	if e.mod != "" {
		args = append(args, "-mod="+e.mod)
	}

	// 目标包
	args = append(args, e.getPackages())

	return args
}

// buildGenerateArgs 构建 go generate 命令参数
func (e *GoBuildExecutor) buildGenerateArgs() []string {
	args := []string{"go", "generate"}

	// -v 详细输出
	if e.verbose {
		args = append(args, "-v")
	}

	// 目标包
	args = append(args, e.getPackages())

	return args
}

// printEnvInfo 打印环境变量信息
func (e *GoBuildExecutor) printEnvInfo(handler OutputHandler) {
	if e.goos != "" || e.goarch != "" {
		handler(fmt.Sprintf("🎯 目标平台: %s/%s", e.getGOOS(), e.getGOARCH()), false)
	}
	if e.cgoEnabled != nil {
		status := "禁用"
		if *e.cgoEnabled {
			status = "启用"
		}
		handler(fmt.Sprintf("⚙️  CGO: %s", status), false)
	}
}

// getGoCommand 获取 Go 子命令
func (e *GoBuildExecutor) getGoCommand() string {
	if e.goCommand != "" {
		return e.goCommand
	}
	return "build"
}

// getPackages 获取目标包
func (e *GoBuildExecutor) getPackages() string {
	if e.packages != "" {
		return e.packages
	}
	return "."
}

// getGOOS 获取 GOOS 值
func (e *GoBuildExecutor) getGOOS() string {
	if e.goos != "" {
		return e.goos
	}
	return "native"
}

// getGOARCH 获取 GOARCH 值
func (e *GoBuildExecutor) getGOARCH() string {
	if e.goarch != "" {
		return e.goarch
	}
	return "native"
}

// getWorkDir 获取工作目录
func (e *GoBuildExecutor) getWorkDir() string {
	if e.workingDir != "" {
		return e.workingDir
	}
	dir, _ := os.Getwd()
	return dir
}
