package config

import "time"

// Config 根配置结构
type Config struct {
	Version    string              `yaml:"version"`
	Project    ProjectConfig       `yaml:"project"`
	Variables  map[string]string   `yaml:"variables"`
	Registries map[string]Registry `yaml:"registries"`
	Servers    map[string]Server   `yaml:"servers"`
	Pipeline   []Stage             `yaml:"pipeline"`
	Hooks      *Hooks              `yaml:"hooks,omitempty"`
}

// ProjectConfig 项目基本信息
type ProjectConfig struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
}

// Registry Docker Registry 配置
type Registry struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Server SSH 服务器配置
type Server struct {
	Host     string     `yaml:"host"`
	Port     int        `yaml:"port"`
	Username string     `yaml:"username"`
	Auth     ServerAuth `yaml:"auth"`
}

// ServerAuth SSH 认证配置
type ServerAuth struct {
	Type     string `yaml:"type"` // "password" | "key"
	Password string `yaml:"password,omitempty"`
	KeyPath  string `yaml:"key_path,omitempty"`
}

// Stage 流水线阶段
type Stage struct {
	Stage    string `yaml:"stage"`
	Name     string `yaml:"name"`
	Parallel bool   `yaml:"parallel,omitempty"`
	Tasks    []Task `yaml:"tasks"`
}

// Task 任务配置
type Task struct {
	Name   string     `yaml:"name"`
	Type   string     `yaml:"type"` // "maven" | "docker-build" | "docker-push" | "ssh"
	Config TaskConfig `yaml:"config"`
}

// TaskConfig 任务具体配置
type TaskConfig struct {
	// 通用配置
	WorkingDir string `yaml:"working_dir,omitempty"`
	Timeout    int    `yaml:"timeout,omitempty"` // 超时时间（秒）

	// Maven 配置
	Command string `yaml:"command,omitempty"`
	Script  string `yaml:"script,omitempty"`

	// Docker Build 配置
	Dockerfile        string            `yaml:"dockerfile,omitempty"`
	Context           string            `yaml:"context,omitempty"`
	ImageName         string            `yaml:"image_name,omitempty"`
	Tag               string            `yaml:"tag,omitempty"`
	BuildArgs         map[string]string `yaml:"build_args,omitempty"`
	Platforms         []string          `yaml:"platforms,omitempty"`            // 多平台构建，如 ["linux/amd64", "linux/arm64"]
	PushOnBuild       *bool             `yaml:"push_on_build,omitempty"`        // 多平台构建时是否自动推送 (默认 true)
	PushLatestOnBuild bool              `yaml:"push_latest_on_build,omitempty"` // 多平台构建时是否同时推送 latest 标签
	AutoScan          *AutoScanConfig   `yaml:"auto_scan,omitempty"`

	// Docker Push 配置
	Registry   string   `yaml:"registry,omitempty"`
	Images     []string `yaml:"images,omitempty"`
	Auto       bool     `yaml:"auto,omitempty"`
	PushLatest bool     `yaml:"push_latest,omitempty"` // 同时推送 latest 标签

	// SSH 配置
	Server      string   `yaml:"server,omitempty"`
	Commands    []string `yaml:"commands,omitempty"`
	LocalScript string   `yaml:"local_script,omitempty"`

	// Go Build 配置
	GoCommand  string `yaml:"go_command,omitempty"`  // go 子命令: build/test/generate (默认 build)
	GoOS       string `yaml:"goos,omitempty"`        // 目标操作系统
	GoArch     string `yaml:"goarch,omitempty"`      // 目标架构
	Output     string `yaml:"output,omitempty"`      // 输出文件路径 (-o)
	LDFlags    string `yaml:"ldflags,omitempty"`     // 链接标志 (-ldflags)
	BuildTags  string `yaml:"tags,omitempty"`        // 构建标签 (-tags)
	CGOEnabled *bool  `yaml:"cgo_enabled,omitempty"` // CGO 开关 (指针区分未设置和 false)
	GoPrivate  string `yaml:"goprivate,omitempty"`   // GOPRIVATE 环境变量
	GoProxy    string `yaml:"goproxy,omitempty"`     // GOPROXY 环境变量
	Race       bool   `yaml:"race,omitempty"`        // 竞态检测 (-race)
	Trimpath   bool   `yaml:"trimpath,omitempty"`    // 移除路径 (-trimpath)
	Mod        string `yaml:"mod,omitempty"`         // 模块模式 (-mod=vendor/readonly/mod)
	Packages   string `yaml:"packages,omitempty"`    // 目标包 (默认 .)
	GoVerbose  bool   `yaml:"go_verbose,omitempty"`  // 详细输出 (-v)
}

// GetTimeoutDuration 返回超时时间的 Duration 格式
func (c TaskConfig) GetTimeoutDuration() time.Duration {
	if c.Timeout <= 0 {
		return 5 * time.Minute // 默认 5 分钟
	}
	return time.Duration(c.Timeout) * time.Second
}

// AutoScanConfig Dockerfile 自动扫描配置
type AutoScanConfig struct {
	Enabled     bool     `yaml:"enabled"`
	Pattern     string   `yaml:"pattern"`
	Exclude     []string `yaml:"exclude,omitempty"`
	ImagePrefix string   `yaml:"image_prefix,omitempty"`
	Tag         string   `yaml:"tag,omitempty"`
	Platforms   []string `yaml:"platforms,omitempty"` // 多平台构建，如 ["linux/amd64", "linux/arm64"]
}

// Hooks 钩子配置
type Hooks struct {
	PreBuild  []string `yaml:"pre_build,omitempty"`
	PostBuild []string `yaml:"post_build,omitempty"`
	OnFailure []string `yaml:"on_failure,omitempty"`
}

// TaskType 任务类型常量
const (
	TaskTypeMaven       = "maven"
	TaskTypeDockerBuild = "docker-build"
	TaskTypeDockerPush  = "docker-push"
	TaskTypeSSH         = "ssh"
	TaskTypeGoBuild     = "go-build"
)
