package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Loader 配置加载器
type Loader struct {
	configPath string
}

// NewLoader 创建配置加载器
func NewLoader(configPath string) *Loader {
	return &Loader{configPath: configPath}
}

// Load 加载并解析配置文件
func (l *Loader) Load() (*Config, error) {
	data, err := os.ReadFile(l.configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 替换变量
	l.replaceVariables(&cfg)

	return &cfg, nil
}

// replaceVariables 替换配置中的变量引用
func (l *Loader) replaceVariables(cfg *Config) {
	// 构建变量映射（包含环境变量）
	vars := make(map[string]string)
	for k, v := range cfg.Variables {
		vars[k] = v
	}

	// 递归替换所有字符串字段
	l.replaceInRegistries(cfg.Registries, vars)
	l.replaceInServers(cfg.Servers, vars)
	l.replaceInPipeline(cfg.Pipeline, vars)
}

// replaceInRegistries 替换 Registry 配置中的变量
func (l *Loader) replaceInRegistries(registries map[string]Registry, vars map[string]string) {
	for name, reg := range registries {
		reg.URL = l.expandVars(reg.URL, vars)
		reg.Username = l.expandVars(reg.Username, vars)
		reg.Password = l.expandVars(reg.Password, vars)
		registries[name] = reg
	}
}

// replaceInServers 替换 Server 配置中的变量
func (l *Loader) replaceInServers(servers map[string]Server, vars map[string]string) {
	for name, srv := range servers {
		srv.Host = l.expandVars(srv.Host, vars)
		srv.Username = l.expandVars(srv.Username, vars)
		srv.Auth.Password = l.expandVars(srv.Auth.Password, vars)
		srv.Auth.KeyPath = l.expandVars(srv.Auth.KeyPath, vars)
		servers[name] = srv
	}
}

// replaceInPipeline 替换 Pipeline 配置中的变量
func (l *Loader) replaceInPipeline(stages []Stage, vars map[string]string) {
	for i := range stages {
		for j := range stages[i].Tasks {
			task := &stages[i].Tasks[j]
			task.Config.WorkingDir = l.expandVars(task.Config.WorkingDir, vars)
			task.Config.Command = l.expandVars(task.Config.Command, vars)
			task.Config.Script = l.expandVars(task.Config.Script, vars)
			task.Config.Dockerfile = l.expandVars(task.Config.Dockerfile, vars)
			task.Config.Context = l.expandVars(task.Config.Context, vars)
			task.Config.ImageName = l.expandVars(task.Config.ImageName, vars)
			task.Config.Tag = l.expandVars(task.Config.Tag, vars)
			task.Config.Registry = l.expandVars(task.Config.Registry, vars)
			task.Config.LocalScript = l.expandVars(task.Config.LocalScript, vars)
			task.Config.GoCommand = l.expandVars(task.Config.GoCommand, vars)
			task.Config.GoOS = l.expandVars(task.Config.GoOS, vars)
			task.Config.GoArch = l.expandVars(task.Config.GoArch, vars)
			task.Config.Output = l.expandVars(task.Config.Output, vars)
			task.Config.LDFlags = l.expandVars(task.Config.LDFlags, vars)
			task.Config.BuildTags = l.expandVars(task.Config.BuildTags, vars)
			task.Config.GoPrivate = l.expandVars(task.Config.GoPrivate, vars)
			task.Config.GoProxy = l.expandVars(task.Config.GoProxy, vars)
			task.Config.Mod = l.expandVars(task.Config.Mod, vars)
			task.Config.Packages = l.expandVars(task.Config.Packages, vars)

			// 替换 images 数组
			for k := range task.Config.Images {
				task.Config.Images[k] = l.expandVars(task.Config.Images[k], vars)
			}

			// 替换 commands 数组
			for k := range task.Config.Commands {
				task.Config.Commands[k] = l.expandVars(task.Config.Commands[k], vars)
			}

			// 替换 platforms 数组
			for k := range task.Config.Platforms {
				task.Config.Platforms[k] = l.expandVars(task.Config.Platforms[k], vars)
			}

			// 替换 build_args
			for k, v := range task.Config.BuildArgs {
				task.Config.BuildArgs[k] = l.expandVars(v, vars)
			}

			// 替换 auto_scan 配置
			if task.Config.AutoScan != nil {
				task.Config.AutoScan.Pattern = l.expandVars(task.Config.AutoScan.Pattern, vars)
				task.Config.AutoScan.ImagePrefix = l.expandVars(task.Config.AutoScan.ImagePrefix, vars)
				task.Config.AutoScan.Tag = l.expandVars(task.Config.AutoScan.Tag, vars)
			}
		}
	}
}

// expandVars 展开字符串中的变量引用
// 支持格式: ${VAR_NAME} 和 $VAR_NAME
func (l *Loader) expandVars(s string, vars map[string]string) string {
	if s == "" {
		return s
	}

	// 匹配 ${VAR_NAME} 格式
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	result := re.ReplaceAllStringFunc(s, func(match string) string {
		varName := match[2 : len(match)-1] // 提取变量名
		if val, ok := vars[varName]; ok {
			return val
		}
		// 尝试从环境变量获取
		if val := os.Getenv(varName); val != "" {
			return val
		}
		return match // 保持原样
	})

	// 匹配 $VAR_NAME 格式（仅字母数字下划线）
	re2 := regexp.MustCompile(`\$([A-Za-z_][A-Za-z0-9_]*)`)
	result = re2.ReplaceAllStringFunc(result, func(match string) string {
		varName := match[1:] // 提取变量名
		if val, ok := vars[varName]; ok {
			return val
		}
		if val := os.Getenv(varName); val != "" {
			return val
		}
		return match
	})

	return result
}

// FindConfigFile 在当前目录及父目录中查找配置文件
func FindConfigFile() (string, error) {
	names := []string{"xbuilder.yaml", "xbuilder.yml", ".xbuilder.yaml", ".xbuilder.yml"}

	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		for _, name := range names {
			path := dir + "/" + name
			if _, err := os.Stat(path); err == nil {
				return path, nil
			}
		}

		// 向上查找
		parent := dir[:strings.LastIndex(dir, "/")]
		if parent == dir || parent == "" {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("未找到配置文件 (xbuilder.yaml)")
}
