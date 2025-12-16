package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/xiaolfeng/builder-cli/resources"
)

// genCmd 父命令
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "生成配置文件和模板",
	Long: `生成 xbuilder 配置文件、Dockerfile、docker-compose 和 Makefile 模板。

子命令:
  config        生成完整配置文件
  dockerfile    生成 Dockerfile
  dockercompose 生成 docker-compose 文件
  makefile      生成 Makefile`,
	Example: `  xbuilder gen config              # 生成完整配置
  xbuilder gen dockerfile          # 生成 Go Dockerfile (默认)
  xbuilder gen dockercompose       # 生成全部三个环境的 compose 文件
  xbuilder gen makefile            # 生成 Makefile`,
}

func init() {
	rootCmd.AddCommand(genCmd)

	// 添加子命令
	genCmd.AddCommand(genConfigCmd)
	genCmd.AddCommand(genDockerfileCmd)
	genCmd.AddCommand(genDockerComposeCmd)
	genCmd.AddCommand(genMakefileCmd)
}

// ═══════════════════════════════════════════════════════════════════════════
// gen config 子命令
// ═══════════════════════════════════════════════════════════════════════════

var (
	configForce   bool
	configScripts bool
	configOutput  string
)

var genConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "生成完整配置文件",
	Long: `生成完整的 xbuilder.yaml 配置文件模板。

可选择同时生成示例脚本文件 (--scripts)。`,
	Example: `  xbuilder gen config            # 生成配置文件
  xbuilder gen config --scripts  # 同时生成脚本文件
  xbuilder gen config -f         # 强制覆盖`,
	RunE: runGenConfig,
}

func init() {
	genConfigCmd.Flags().BoolVarP(&configForce, "force", "f", false, "强制覆盖已存在的文件")
	genConfigCmd.Flags().BoolVar(&configScripts, "scripts", false, "同时生成示例脚本文件")
	genConfigCmd.Flags().StringVarP(&configOutput, "output", "o", "xbuilder.yaml", "输出文件路径")
}

func runGenConfig(cmd *cobra.Command, args []string) error {
	files := []struct {
		path     string
		template string
		desc     string
	}{
		{configOutput, "config/full.yaml", "完整配置文件"},
	}

	// 如果需要生成脚本
	if configScripts {
		files = append(files, []struct {
			path     string
			template string
			desc     string
		}{
			{"scripts/build.sh", "scripts/build.sh", "构建脚本"},
			{"scripts/deploy.sh", "scripts/deploy.sh", "部署脚本"},
			{"scripts/notify.sh", "scripts/notify.sh", "通知脚本"},
		}...)
	}

	return generateFiles(files, configForce)
}

// ═══════════════════════════════════════════════════════════════════════════
// gen dockerfile 子命令
// ═══════════════════════════════════════════════════════════════════════════

var (
	dockerfileForce  bool
	dockerfileLang   string
	dockerfileOutput string
)

var genDockerfileCmd = &cobra.Command{
	Use:     "dockerfile",
	Aliases: []string{"docker", "df"},
	Short:   "生成 Dockerfile",
	Long: `生成 Dockerfile 模板。

支持的语言:
  go    Go 应用 (默认，多阶段构建，scratch 基础镜像)
  java  Java/Spring Boot 应用`,
	Example: `  xbuilder gen dockerfile              # 生成 Go Dockerfile (默认)
  xbuilder gen dockerfile --lang java  # 生成 Java Dockerfile
  xbuilder gen dockerfile -o Dockerfile.prod  # 自定义输出路径`,
	RunE: runGenDockerfile,
}

func init() {
	genDockerfileCmd.Flags().BoolVarP(&dockerfileForce, "force", "f", false, "强制覆盖已存在的文件")
	genDockerfileCmd.Flags().StringVarP(&dockerfileLang, "lang", "l", "go", "语言类型 (go/java)")
	genDockerfileCmd.Flags().StringVarP(&dockerfileOutput, "output", "o", "Dockerfile", "输出文件路径")
}

func runGenDockerfile(cmd *cobra.Command, args []string) error {
	var templatePath string
	switch dockerfileLang {
	case "go", "golang":
		templatePath = "dockerfile/go.Dockerfile"
	case "java", "spring", "springboot":
		templatePath = "dockerfile/java.Dockerfile"
	default:
		return fmt.Errorf("不支持的语言类型: %s (支持: go, java)", dockerfileLang)
	}

	files := []struct {
		path     string
		template string
		desc     string
	}{
		{dockerfileOutput, templatePath, fmt.Sprintf("Dockerfile (%s)", dockerfileLang)},
	}

	return generateFiles(files, dockerfileForce)
}

// ═══════════════════════════════════════════════════════════════════════════
// gen dockercompose 子命令
// ═══════════════════════════════════════════════════════════════════════════

var (
	composeForce bool
	composeScope string
)

var genDockerComposeCmd = &cobra.Command{
	Use:     "dockercompose",
	Aliases: []string{"compose", "dc"},
	Short:   "生成 docker-compose 文件",
	Long: `生成 docker-compose 配置文件。

默认行为: 生成全部三个环境的配置文件
  - docker-compose.dev.yaml   开发环境
  - docker-compose.test.yaml  测试环境
  - docker-compose.yaml       生产环境

使用 --scope 指定只生成单个环境的配置文件。`,
	Example: `  xbuilder gen dockercompose              # 生成全部三个文件 (默认)
  xbuilder gen dockercompose --scope dev  # 仅生成开发环境配置
  xbuilder gen dc -s prod                 # 仅生成生产环境配置
  xbuilder gen dc -s test                 # 仅生成测试环境配置`,
	RunE: runGenDockerCompose,
}

func init() {
	genDockerComposeCmd.Flags().BoolVarP(&composeForce, "force", "f", false, "强制覆盖已存在的文件")
	genDockerComposeCmd.Flags().StringVarP(&composeScope, "scope", "s", "", "环境范围 (dev/test/prod/all)，默认生成全部")
}

func runGenDockerCompose(cmd *cobra.Command, args []string) error {
	type composeFile struct {
		path     string
		template string
		desc     string
	}

	allFiles := []composeFile{
		{"docker-compose.dev.yaml", "dockercompose/dev.yaml", "开发环境 docker-compose"},
		{"docker-compose.test.yaml", "dockercompose/test.yaml", "测试环境 docker-compose"},
		{"docker-compose.yaml", "dockercompose/prod.yaml", "生产环境 docker-compose"},
	}

	var files []struct {
		path     string
		template string
		desc     string
	}

	switch composeScope {
	case "", "all":
		// 默认生成全部三个文件
		for _, f := range allFiles {
			files = append(files, struct {
				path     string
				template string
				desc     string
			}{f.path, f.template, f.desc})
		}
	case "dev", "development":
		files = append(files, struct {
			path     string
			template string
			desc     string
		}{"docker-compose.dev.yaml", "dockercompose/dev.yaml", "开发环境 docker-compose"})
	case "test", "testing":
		files = append(files, struct {
			path     string
			template string
			desc     string
		}{"docker-compose.test.yaml", "dockercompose/test.yaml", "测试环境 docker-compose"})
	case "prod", "production":
		files = append(files, struct {
			path     string
			template string
			desc     string
		}{"docker-compose.yaml", "dockercompose/prod.yaml", "生产环境 docker-compose"})
	default:
		return fmt.Errorf("不支持的环境范围: %s (支持: dev, test, prod, all)", composeScope)
	}

	return generateFiles(files, composeForce)
}

// ═══════════════════════════════════════════════════════════════════════════
// gen makefile 子命令
// ═══════════════════════════════════════════════════════════════════════════

var (
	makefileForce    bool
	makefileOutput   string
	makefileProject  string
	makefileRegistry string
)

var genMakefileCmd = &cobra.Command{
	Use:     "makefile",
	Aliases: []string{"make", "mk"},
	Short:   "生成 Makefile",
	Long: `生成项目 Makefile 模板。

Makefile 包含常用目标: init, gen, validate, build, docker-build, docker-push, deploy-*, clean, help`,
	Example: `  xbuilder gen makefile
  xbuilder gen makefile --project myapp --registry docker.io/myuser
  xbuilder gen mk -p myapp -r ghcr.io/myorg`,
	RunE: runGenMakefile,
}

func init() {
	genMakefileCmd.Flags().BoolVarP(&makefileForce, "force", "f", false, "强制覆盖已存在的文件")
	genMakefileCmd.Flags().StringVarP(&makefileOutput, "output", "o", "Makefile", "输出文件路径")
	genMakefileCmd.Flags().StringVarP(&makefileProject, "project", "p", "my-project", "项目名称")
	genMakefileCmd.Flags().StringVarP(&makefileRegistry, "registry", "r", "docker.io/myuser", "镜像仓库地址")
}

func runGenMakefile(cmd *cobra.Command, args []string) error {
	fmt.Println("📦 生成 xbuilder 模板文件...")
	fmt.Println()

	// 使用模板渲染
	content, err := resources.ExecuteTemplate("makefile/Makefile.tmpl", map[string]string{
		"ProjectName": makefileProject,
		"Registry":    makefileRegistry,
	})
	if err != nil {
		return fmt.Errorf("渲染 Makefile 模板失败: %w", err)
	}

	// 检查文件是否存在
	if _, err := os.Stat(makefileOutput); err == nil {
		if !makefileForce {
			fmt.Printf("  ⏭️  跳过 %s (已存在)\n", makefileOutput)
			fmt.Println()
			fmt.Println("完成! 创建 0 个文件, 跳过 1 个文件")
			return nil
		}
	}

	// 写入文件
	if err := os.WriteFile(makefileOutput, []byte(content), 0644); err != nil {
		return fmt.Errorf("创建文件 %s 失败: %w", makefileOutput, err)
	}

	fmt.Printf("  ✅ 创建 %s (Makefile)\n", makefileOutput)
	fmt.Println()
	fmt.Println("完成! 创建 1 个文件")
	return nil
}

// ═══════════════════════════════════════════════════════════════════════════
// 公共函数
// ═══════════════════════════════════════════════════════════════════════════

func generateFiles(files []struct {
	path     string
	template string
	desc     string
}, force bool) error {
	fmt.Println("📦 生成 xbuilder 模板文件...")
	fmt.Println()

	createdCount := 0
	skippedCount := 0

	for _, f := range files {
		// 创建目录
		dir := filepath.Dir(f.path)
		if dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("创建目录 %s 失败: %w", dir, err)
			}
		}

		// 检查文件是否已存在
		if _, err := os.Stat(f.path); err == nil {
			if !force {
				fmt.Printf("  ⏭️  跳过 %s (已存在)\n", f.path)
				skippedCount++
				continue
			}
		}

		// 读取模板内容
		content, err := resources.GetTemplate(f.template)
		if err != nil {
			return fmt.Errorf("读取模板 %s 失败: %w", f.template, err)
		}

		// 设置文件权限
		perm := os.FileMode(0644)
		if filepath.Ext(f.path) == ".sh" {
			perm = 0755 // 脚本文件添加执行权限
		}

		// 写入文件
		if err := os.WriteFile(f.path, content, perm); err != nil {
			return fmt.Errorf("创建文件 %s 失败: %w", f.path, err)
		}

		fmt.Printf("  ✅ 创建 %s (%s)\n", f.path, f.desc)
		createdCount++
	}

	fmt.Println()
	fmt.Printf("完成! 创建 %d 个文件", createdCount)
	if skippedCount > 0 {
		fmt.Printf(", 跳过 %d 个文件", skippedCount)
	}
	fmt.Println()

	return nil
}
