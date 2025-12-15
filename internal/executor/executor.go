package executor

import (
	"context"
	"time"
)

// OutputHandler 输出处理函数类型
// line: 输出行内容
// isError: 是否为错误输出（stderr）
type OutputHandler func(line string, isError bool)

// Executor 执行器接口
type Executor interface {
	// Execute 执行任务
	// ctx: 上下文，用于取消操作
	// handler: 输出处理函数，用于实时接收输出
	Execute(ctx context.Context, handler OutputHandler) error

	// Name 返回执行器名称
	Name() string

	// Type 返回执行器类型
	Type() string
}

// Result 执行结果
type Result struct {
	Success   bool          // 是否成功
	Duration  time.Duration // 执行耗时
	Output    string        // 输出内容
	Error     error         // 错误信息
	ExitCode  int           // 退出码
	StartTime time.Time     // 开始时间
	EndTime   time.Time     // 结束时间
}

// ExecutorType 执行器类型常量
const (
	TypeMaven       = "maven"
	TypeDockerBuild = "docker-build"
	TypeDockerPush  = "docker-push"
	TypeSSH         = "ssh"
	TypeCommand     = "command"
	TypeGoBuild     = "go-build"
)

// BaseExecutor 基础执行器（可嵌入其他执行器）
type BaseExecutor struct {
	name       string
	execType   string
	workingDir string
	timeout    time.Duration
	env        []string
}

// NewBaseExecutor 创建基础执行器
func NewBaseExecutor(name, execType string) *BaseExecutor {
	return &BaseExecutor{
		name:     name,
		execType: execType,
		timeout:  10 * time.Minute, // 默认 10 分钟超时
	}
}

// Name 返回执行器名称
func (e *BaseExecutor) Name() string {
	return e.name
}

// Type 返回执行器类型
func (e *BaseExecutor) Type() string {
	return e.execType
}

// SetWorkingDir 设置工作目录
func (e *BaseExecutor) SetWorkingDir(dir string) {
	e.workingDir = dir
}

// SetTimeout 设置超时时间
func (e *BaseExecutor) SetTimeout(timeout time.Duration) {
	e.timeout = timeout
}

// SetEnv 设置环境变量
func (e *BaseExecutor) SetEnv(env []string) {
	e.env = env
}

// AddEnv 添加环境变量
func (e *BaseExecutor) AddEnv(key, value string) {
	e.env = append(e.env, key+"="+value)
}

// GetWorkingDir 获取工作目录
func (e *BaseExecutor) GetWorkingDir() string {
	return e.workingDir
}

// GetTimeout 获取超时时间
func (e *BaseExecutor) GetTimeout() time.Duration {
	return e.timeout
}

// GetEnv 获取环境变量
func (e *BaseExecutor) GetEnv() []string {
	return e.env
}
