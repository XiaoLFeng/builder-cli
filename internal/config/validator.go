package config

import (
	"fmt"
	"os"
	"strings"
)

// ValidationError 验证错误
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors 多个验证错误
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "\n")
}

// Validator 配置验证器
type Validator struct {
	config *Config
	errors ValidationErrors
}

// NewValidator 创建验证器
func NewValidator(cfg *Config) *Validator {
	return &Validator{config: cfg}
}

// Validate 执行完整验证
func (v *Validator) Validate() error {
	v.errors = nil

	v.validateProject()
	v.validateRegistries()
	v.validateServers()
	v.validatePipeline()

	if len(v.errors) > 0 {
		return v.errors
	}
	return nil
}

// validateProject 验证项目配置
func (v *Validator) validateProject() {
	if v.config.Project.Name == "" {
		v.addError("project.name", "项目名称不能为空")
	}
}

// validateRegistries 验证 Registry 配置
func (v *Validator) validateRegistries() {
	for name, reg := range v.config.Registries {
		if reg.URL == "" {
			v.addError(fmt.Sprintf("registries.%s.url", name), "Registry URL 不能为空")
		}
	}
}

// validateServers 验证服务器配置
func (v *Validator) validateServers() {
	for name, srv := range v.config.Servers {
		if srv.Host == "" {
			v.addError(fmt.Sprintf("servers.%s.host", name), "服务器地址不能为空")
		}
		if srv.Port <= 0 {
			v.addError(fmt.Sprintf("servers.%s.port", name), "端口号必须大于 0")
		}
		if srv.Username == "" {
			v.addError(fmt.Sprintf("servers.%s.username", name), "用户名不能为空")
		}

		// 验证认证配置
		switch srv.Auth.Type {
		case "password":
			if srv.Auth.Password == "" {
				v.addError(fmt.Sprintf("servers.%s.auth.password", name), "密码不能为空")
			}
		case "key":
			if srv.Auth.KeyPath == "" {
				v.addError(fmt.Sprintf("servers.%s.auth.key_path", name), "密钥路径不能为空")
			} else {
				// 检查密钥文件是否存在
				keyPath := expandHomePath(srv.Auth.KeyPath)
				if _, err := os.Stat(keyPath); os.IsNotExist(err) {
					v.addError(fmt.Sprintf("servers.%s.auth.key_path", name),
						fmt.Sprintf("密钥文件不存在: %s", keyPath))
				}
			}
		default:
			v.addError(fmt.Sprintf("servers.%s.auth.type", name),
				fmt.Sprintf("无效的认证类型: %s (支持: password, key)", srv.Auth.Type))
		}
	}
}

// validatePipeline 验证流水线配置
func (v *Validator) validatePipeline() {
	if len(v.config.Pipeline) == 0 {
		v.addError("pipeline", "流水线不能为空")
		return
	}

	stageNames := make(map[string]bool)
	for i, stage := range v.config.Pipeline {
		// 检查阶段名称唯一性
		if stageNames[stage.Stage] {
			v.addError(fmt.Sprintf("pipeline[%d].stage", i),
				fmt.Sprintf("阶段名称重复: %s", stage.Stage))
		}
		stageNames[stage.Stage] = true

		if stage.Name == "" {
			v.addError(fmt.Sprintf("pipeline[%d].name", i), "阶段显示名称不能为空")
		}

		if len(stage.Tasks) == 0 {
			v.addError(fmt.Sprintf("pipeline[%d].tasks", i), "阶段任务不能为空")
			continue
		}

		// 验证每个任务
		for j, task := range stage.Tasks {
			v.validateTask(fmt.Sprintf("pipeline[%d].tasks[%d]", i, j), task)
		}
	}
}

// validateTask 验证任务配置
func (v *Validator) validateTask(path string, task Task) {
	if task.Name == "" {
		v.addError(path+".name", "任务名称不能为空")
	}

	switch task.Type {
	case TaskTypeMaven:
		v.validateMavenTask(path, task)
	case TaskTypeDockerBuild:
		v.validateDockerBuildTask(path, task)
	case TaskTypeDockerPush:
		v.validateDockerPushTask(path, task)
	case TaskTypeSSH:
		v.validateSSHTask(path, task)
	case TaskTypeGoBuild:
		// Go 构建任务可接受默认参数；若使用 go test/generate，可按需补充校验
	case TaskTypeShell:
		v.validateShellTask(path, task)
	default:
		v.addError(path+".type",
			fmt.Sprintf("无效的任务类型: %s (支持: maven, docker-build, docker-push, ssh, go-build, shell)", task.Type))
	}
}

// validateMavenTask 验证 Maven 任务
func (v *Validator) validateMavenTask(path string, task Task) {
	if task.Config.Command == "" && task.Config.Script == "" {
		v.addError(path+".config", "Maven 任务必须指定 command 或 script")
	}
}

// validateDockerBuildTask 验证 Docker Build 任务
func (v *Validator) validateDockerBuildTask(path string, task Task) {
	// 如果启用了自动扫描，则不需要其他配置
	if task.Config.AutoScan != nil && task.Config.AutoScan.Enabled {
		if task.Config.AutoScan.Pattern == "" {
			v.addError(path+".config.auto_scan.pattern", "自动扫描模式不能为空")
		}
		return
	}

	if task.Config.Dockerfile == "" {
		v.addError(path+".config.dockerfile", "Dockerfile 路径不能为空")
	}
	if task.Config.ImageName == "" {
		v.addError(path+".config.image_name", "镜像名称不能为空")
	}
}

// validateDockerPushTask 验证 Docker Push 任务
func (v *Validator) validateDockerPushTask(path string, task Task) {
	if task.Config.Registry == "" {
		v.addError(path+".config.registry", "Registry 名称不能为空")
	} else if _, ok := v.config.Registries[task.Config.Registry]; !ok {
		v.addError(path+".config.registry",
			fmt.Sprintf("Registry 不存在: %s", task.Config.Registry))
	}

	if len(task.Config.Images) == 0 && !task.Config.Auto {
		v.addError(path+".config", "必须指定 images 列表或设置 auto: true")
	}
}

// validateSSHTask 验证 SSH 任务
func (v *Validator) validateSSHTask(path string, task Task) {
	if task.Config.Server == "" {
		v.addError(path+".config.server", "服务器名称不能为空")
	} else if _, ok := v.config.Servers[task.Config.Server]; !ok {
		v.addError(path+".config.server",
			fmt.Sprintf("服务器不存在: %s", task.Config.Server))
	}

	if len(task.Config.Commands) == 0 && task.Config.Script == "" && task.Config.LocalScript == "" {
		v.addError(path+".config", "必须指定 commands、script 或 local_script")
	}
}

// validateShellTask 验证 Shell 任务
func (v *Validator) validateShellTask(path string, task Task) {
	if task.Config.Command == "" && task.Config.Script == "" {
		v.addError(path+".config", "Shell 任务必须指定 command 或 script")
	}
}

// addError 添加验证错误
func (v *Validator) addError(field, message string) {
	v.errors = append(v.errors, ValidationError{Field: field, Message: message})
}

// expandHomePath 展开 ~ 为 home 目录
func expandHomePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return home + path[1:]
		}
	}
	return path
}
