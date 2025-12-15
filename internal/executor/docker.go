package executor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xiaolfeng/builder-cli/internal/config"
)

// DockerBuildExecutor Docker 构建执行器
type DockerBuildExecutor struct {
	*BaseExecutor
	dockerfile string
	context    string
	imageName  string
	tag        string
	buildArgs  map[string]string
	platforms  []string // 多平台支持
}

// NewDockerBuildExecutor 创建 Docker 构建执行器
func NewDockerBuildExecutor(taskName string, cfg config.TaskConfig) *DockerBuildExecutor {
	e := &DockerBuildExecutor{
		BaseExecutor: NewBaseExecutor(taskName, TypeDockerBuild),
		dockerfile:   cfg.Dockerfile,
		context:      cfg.Context,
		imageName:    cfg.ImageName,
		tag:          cfg.Tag,
		buildArgs:    cfg.BuildArgs,
		platforms:    cfg.Platforms,
	}

	// 默认值
	if e.context == "" {
		e.context = "."
	}
	if e.tag == "" {
		e.tag = "latest"
	}

	// 设置超时
	if cfg.Timeout > 0 {
		e.SetTimeout(time.Duration(cfg.Timeout) * time.Second)
	} else {
		e.SetTimeout(30 * time.Minute) // Docker 构建默认 30 分钟
	}

	return e
}

// FullImageName 返回完整的镜像名称
func (e *DockerBuildExecutor) FullImageName() string {
	return fmt.Sprintf("%s:%s", e.imageName, e.tag)
}

// Execute 执行 Docker 构建
func (e *DockerBuildExecutor) Execute(ctx context.Context, handler OutputHandler) error {
	handler(fmt.Sprintf("🐳 构建 Docker 镜像: %s", e.FullImageName()), false)
	handler(fmt.Sprintf("📄 Dockerfile: %s", e.dockerfile), false)
	handler(fmt.Sprintf("📁 Context: %s", e.context), false)
	if len(e.platforms) > 0 {
		handler(fmt.Sprintf("🖥️  Platforms: %s", strings.Join(e.platforms, ", ")), false)
	}
	handler("", false)

	// 构建命令参数
	args := []string{"build"}

	// Dockerfile 路径
	if e.dockerfile != "" {
		args = append(args, "-f", e.dockerfile)
	}

	// 镜像标签
	args = append(args, "-t", e.FullImageName())

	// 构建参数
	for k, v := range e.buildArgs {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", k, v))
	}

	// 多平台支持
	if len(e.platforms) > 0 {
		// 使用 buildx 进行多平台构建
		args = []string{"buildx", "build", "--push"}

		// Dockerfile 路径
		if e.dockerfile != "" {
			args = append(args, "-f", e.dockerfile)
		}

		// 镜像标签
		args = append(args, "-t", e.FullImageName())

		// 构建参数
		for k, v := range e.buildArgs {
			args = append(args, "--build-arg", fmt.Sprintf("%s=%s", k, v))
		}

		// 平台列表
		args = append(args, "--platform", strings.Join(e.platforms, ","))
	}

	// Context
	args = append(args, e.context)

	// 构建命令字符串
	command := "docker " + strings.Join(args, " ")

	runner := NewCommandRunner(e.Name(), command)
	runner.SetWorkingDir(e.GetWorkingDir())
	runner.SetTimeout(e.GetTimeout())
	runner.SetEnv(e.GetEnv())

	return runner.Execute(ctx, handler)
}

// DockerPushExecutor Docker 推送执行器
type DockerPushExecutor struct {
	*BaseExecutor
	registry   *config.Registry
	images     []string
	pushLatest bool // 是否同时推送 latest 标签
}

// NewDockerPushExecutor 创建 Docker 推送执行器
func NewDockerPushExecutor(taskName string, cfg config.TaskConfig, registry *config.Registry) *DockerPushExecutor {
	e := &DockerPushExecutor{
		BaseExecutor: NewBaseExecutor(taskName, TypeDockerPush),
		registry:     registry,
		images:       cfg.Images,
		pushLatest:   cfg.PushLatest,
	}

	// 设置超时
	if cfg.Timeout > 0 {
		e.SetTimeout(time.Duration(cfg.Timeout) * time.Second)
	} else {
		e.SetTimeout(20 * time.Minute) // 推送默认 20 分钟
	}

	return e
}

// SetImages 设置要推送的镜像列表
func (e *DockerPushExecutor) SetImages(images []string) {
	e.images = images
}

// Execute 执行 Docker 推送
func (e *DockerPushExecutor) Execute(ctx context.Context, handler OutputHandler) error {
	// 登录 Registry
	if e.registry != nil && e.registry.Username != "" {
		if err := e.login(ctx, handler); err != nil {
			return fmt.Errorf("Registry 登录失败: %w", err)
		}
	}

	// 推送每个镜像
	for _, image := range e.images {
		// 推送原始标签
		handler(fmt.Sprintf("📤 推送镜像: %s", image), false)

		command := fmt.Sprintf("docker push %s", image)
		runner := NewCommandRunner(e.Name(), command)
		runner.SetTimeout(e.GetTimeout())

		if err := runner.Execute(ctx, handler); err != nil {
			return fmt.Errorf("推送镜像失败 [%s]: %w", image, err)
		}

		handler(fmt.Sprintf("✅ 镜像推送成功: %s", image), false)

		// 如果需要同时推送 latest 标签
		if e.pushLatest {
			latestImage, needsPush := e.getLatestTagImage(image)
			if needsPush {
				handler("", false)
				handler(fmt.Sprintf("🏷️  标记为 latest: %s", latestImage), false)

				// 先 tag 为 latest
				tagCmd := fmt.Sprintf("docker tag %s %s", image, latestImage)
				tagRunner := NewCommandRunner(e.Name()+"-tag", tagCmd)
				tagRunner.SetTimeout(30 * time.Second)

				if err := tagRunner.Execute(ctx, handler); err != nil {
					return fmt.Errorf("标记 latest 失败 [%s]: %w", image, err)
				}

				// 推送 latest
				handler(fmt.Sprintf("📤 推送镜像: %s", latestImage), false)
				pushCmd := fmt.Sprintf("docker push %s", latestImage)
				pushRunner := NewCommandRunner(e.Name()+"-push-latest", pushCmd)
				pushRunner.SetTimeout(e.GetTimeout())

				if err := pushRunner.Execute(ctx, handler); err != nil {
					return fmt.Errorf("推送 latest 失败 [%s]: %w", latestImage, err)
				}

				handler(fmt.Sprintf("✅ latest 推送成功: %s", latestImage), false)
			}
		}

		handler("", false)
	}

	return nil
}

// getLatestTagImage 获取 latest 标签版本的镜像名
// 返回 latest 版本的镜像名和是否需要推送（如果原本就是 latest 则不需要）
func (e *DockerPushExecutor) getLatestTagImage(image string) (string, bool) {
	// 解析镜像名和标签
	// 格式: registry/name:tag 或 name:tag
	lastColon := strings.LastIndex(image, ":")

	// 检查是否有端口号（如 registry:5000/name:tag）
	lastSlash := strings.LastIndex(image, "/")
	if lastColon != -1 && lastColon > lastSlash {
		// 有标签
		baseName := image[:lastColon]
		tag := image[lastColon+1:]

		// 如果已经是 latest，不需要再推送
		if tag == "latest" {
			return "", false
		}

		return baseName + ":latest", true
	}

	// 没有标签，默认就是 latest，不需要推送
	return "", false
}

// login 登录 Docker Registry
func (e *DockerPushExecutor) login(ctx context.Context, handler OutputHandler) error {
	handler(fmt.Sprintf("🔐 登录 Registry: %s", e.registry.URL), false)

	// 使用 --password-stdin 更安全
	command := fmt.Sprintf("echo '%s' | docker login %s -u '%s' --password-stdin",
		e.registry.Password, e.registry.URL, e.registry.Username)

	runner := NewCommandRunner("docker-login", command)
	runner.SetTimeout(30 * time.Second)

	return runner.Execute(ctx, handler)
}

// DockerScanner Dockerfile 扫描器
type DockerScanner struct {
	rootDir     string
	pattern     string
	exclude     []string
	imagePrefix string
	tag         string
	platforms   []string // 多平台构建支持
}

// NewDockerScanner 创建 Dockerfile 扫描器
func NewDockerScanner(rootDir string, cfg *config.AutoScanConfig) *DockerScanner {
	s := &DockerScanner{
		rootDir:     rootDir,
		pattern:     cfg.Pattern,
		exclude:     cfg.Exclude,
		imagePrefix: cfg.ImagePrefix,
		tag:         cfg.Tag,
		platforms:   cfg.Platforms,
	}

	if s.pattern == "" {
		s.pattern = "**/Dockerfile"
	}
	if s.tag == "" {
		s.tag = "latest"
	}

	return s
}

// Scan 扫描 Dockerfile 并返回构建执行器列表
func (s *DockerScanner) Scan() ([]*DockerBuildExecutor, error) {
	var executors []*DockerBuildExecutor

	err := filepath.Walk(s.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			// 检查是否在排除列表中
			for _, exclude := range s.exclude {
				if matched, _ := filepath.Match(exclude, path); matched {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// 检查是否匹配 Dockerfile 模式
		if !s.matchPattern(path) {
			return nil
		}

		// 检查是否在排除列表中
		for _, exclude := range s.exclude {
			if matched, _ := filepath.Match(exclude, path); matched {
				return nil
			}
		}

		// 创建执行器
		executor := s.createExecutor(path)
		executors = append(executors, executor)

		return nil
	})

	return executors, err
}

// matchPattern 检查路径是否匹配 Dockerfile 模式
func (s *DockerScanner) matchPattern(path string) bool {
	base := filepath.Base(path)
	return base == "Dockerfile" || strings.HasPrefix(base, "Dockerfile.")
}

// createExecutor 为扫描到的 Dockerfile 创建执行器
func (s *DockerScanner) createExecutor(dockerfilePath string) *DockerBuildExecutor {
	// 获取上下文目录（Dockerfile 所在目录）
	contextDir := filepath.Dir(dockerfilePath)

	// 根据目录名生成镜像名
	dirName := filepath.Base(contextDir)
	imageName := s.imagePrefix
	if imageName != "" && !strings.HasSuffix(imageName, "/") {
		imageName += "/"
	}
	imageName += dirName

	cfg := config.TaskConfig{
		Dockerfile: dockerfilePath,
		Context:    contextDir,
		ImageName:  imageName,
		Tag:        s.tag,
		Platforms:  s.platforms,
	}

	return NewDockerBuildExecutor(fmt.Sprintf("build-%s", dirName), cfg)
}
