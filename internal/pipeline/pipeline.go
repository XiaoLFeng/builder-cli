package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xiaolfeng/builder-cli/internal/config"
	"github.com/xiaolfeng/builder-cli/internal/executor"
	"github.com/xiaolfeng/builder-cli/internal/types"
)

// Pipeline 流水线编排器
type Pipeline struct {
	config       *config.Config
	stages       []*Stage
	program      *tea.Program    // 用于向 TUI 发送消息
	builtImages  []string        // 记录已构建的镜像
	pushedImages map[string]bool // 记录已推送的镜像
	mu           sync.RWMutex
}

// New 创建新的流水线
func New(cfg *config.Config) *Pipeline {
	p := &Pipeline{
		config:       cfg,
		stages:       make([]*Stage, 0, len(cfg.Pipeline)),
		builtImages:  make([]string, 0),
		pushedImages: make(map[string]bool),
	}

	// 创建阶段
	for i, stageCfg := range cfg.Pipeline {
		stage := NewStage(i, stageCfg, cfg)
		p.stages = append(p.stages, stage)
	}

	return p
}

// SetProgram 设置 tea.Program 用于发送消息
func (p *Pipeline) SetProgram(program *tea.Program) {
	p.program = program
}

// GetStages 获取所有阶段
func (p *Pipeline) GetStages() []*Stage {
	return p.stages
}

// GetAllTasks 获取所有任务（扁平化）
func (p *Pipeline) GetAllTasks() []*Task {
	var tasks []*Task
	for _, stage := range p.stages {
		tasks = append(tasks, stage.Tasks...)
	}
	return tasks
}

// Run 运行流水线
func (p *Pipeline) Run(ctx context.Context) error {
	startTime := time.Now()

	// 发送流水线开始消息
	p.sendMsg(pipelineStartMsg{})

	for i, stage := range p.stages {
		// 发送阶段开始消息
		p.sendMsg(types.NewStageStartMsg(i, stage.Name))

		stageStart := time.Now()
		var err error

		if stage.Parallel {
			err = p.runStageParallel(ctx, stage)
		} else {
			err = p.runStageSequential(ctx, stage)
		}

		stageDuration := time.Since(stageStart)

		if err != nil {
			// 发送阶段失败消息
			p.sendMsg(types.NewStageCompleteMsg(i, stage.Name, false, stageDuration))
			// 发送流水线失败消息
			p.sendMsg(types.NewPipelineCompleteMsg(false, time.Since(startTime), err))
			return fmt.Errorf("阶段 [%s] 执行失败: %w", stage.Name, err)
		}

		// 发送阶段完成消息
		p.sendMsg(types.NewStageCompleteMsg(i, stage.Name, true, stageDuration))
	}

	// 发送流水线完成消息
	p.sendMsg(types.NewPipelineCompleteMsg(true, time.Since(startTime), nil))

	return nil
}

// runStageSequential 顺序执行阶段中的任务
func (p *Pipeline) runStageSequential(ctx context.Context, stage *Stage) error {
	for _, task := range stage.Tasks {
		if err := p.runTask(ctx, task); err != nil {
			return err
		}
	}
	return nil
}

// runStageParallel 并行执行阶段中的任务
func (p *Pipeline) runStageParallel(ctx context.Context, stage *Stage) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(stage.Tasks))

	for _, task := range stage.Tasks {
		wg.Add(1)
		go func(t *Task) {
			defer wg.Done()
			if err := p.runTask(ctx, t); err != nil {
				errChan <- err
			}
		}(task)
	}

	wg.Wait()
	close(errChan)

	// 收集所有错误
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("%d 个任务失败: %v", len(errs), errs[0])
	}
	return nil
}

// runTask 运行单个任务
func (p *Pipeline) runTask(ctx context.Context, task *Task) error {
	// 发送任务开始消息
	p.sendMsg(types.NewTaskStatusMsg(task.ID, types.StatusRunning))

	// 创建输出处理器
	handler := func(line string, isError bool) {
		p.sendMsg(types.NewOutputMsg(task.ID, line, isError))
	}

	// 获取执行器
	exec, err := p.createExecutor(task)
	if err != nil {
		p.sendMsg(types.NewTaskStatusMsg(task.ID, types.StatusFailed))
		p.sendMsg(types.NewErrorMsg(task.ID, err, "创建执行器失败"))
		return err
	}

	// 执行任务
	if err := exec.Execute(ctx, handler); err != nil {
		p.sendMsg(types.NewTaskStatusMsg(task.ID, types.StatusFailed))
		p.sendMsg(types.NewErrorMsg(task.ID, err, "任务执行失败"))
		return err
	}

	// 记录构建的镜像（用于 docker push）
	if task.Type == config.TaskTypeDockerBuild {
		if dockerExec, ok := exec.(*executor.DockerBuildExecutor); ok {
			imageName := dockerExec.FullImageName()
			p.mu.Lock()
			p.builtImages = append(p.builtImages, imageName)
			// 记录是否已在构建阶段推送
			if dockerExec.IsPushed() {
				p.pushedImages[imageName] = true
			}
			p.mu.Unlock()
		}
	}

	// 发送任务完成消息
	p.sendMsg(types.NewTaskStatusMsg(task.ID, types.StatusSuccess))

	return nil
}

// createExecutor 根据任务类型创建执行器
func (p *Pipeline) createExecutor(task *Task) (executor.Executor, error) {
	switch task.Type {
	case config.TaskTypeMaven:
		return executor.NewMavenExecutor(task.Name, task.Config), nil

	case config.TaskTypeDockerBuild:
		return executor.NewDockerBuildExecutor(task.Name, task.Config), nil

	case config.TaskTypeDockerPush:
		// 获取 registry 配置
		reg, ok := p.config.Registries[task.Config.Registry]
		if !ok {
			return nil, fmt.Errorf("Registry 不存在: %s", task.Config.Registry)
		}

		exec := executor.NewDockerPushExecutor(task.Name, task.Config, &reg)

		// 如果启用了 auto，使用已构建的镜像
		if task.Config.Auto {
			p.mu.RLock()
			exec.SetImages(p.builtImages)
			// 传递已推送的镜像状态，让 push 执行器跳过这些镜像
			exec.SetSkipPushedImages(p.pushedImages)
			p.mu.RUnlock()
		}

		return exec, nil

	case config.TaskTypeSSH:
		// 获取服务器配置
		server, ok := p.config.Servers[task.Config.Server]
		if !ok {
			return nil, fmt.Errorf("服务器不存在: %s", task.Config.Server)
		}
		return executor.NewSSHExecutor(task.Name, task.Config, &server)

	case config.TaskTypeGoBuild:
		return executor.NewGoBuildExecutor(task.Name, task.Config), nil

	default:
		return nil, fmt.Errorf("不支持的任务类型: %s", task.Type)
	}
}

// sendMsg 发送消息到 TUI
func (p *Pipeline) sendMsg(msg tea.Msg) {
	if p.program != nil {
		p.program.Send(msg)
	}
}

// GetBuiltImages 获取已构建的镜像列表
func (p *Pipeline) GetBuiltImages() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	result := make([]string, len(p.builtImages))
	copy(result, p.builtImages)
	return result
}

// 内部消息类型
type pipelineStartMsg struct{}
