package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xiaolfeng/builder-cli/internal/app"
	"github.com/xiaolfeng/builder-cli/internal/config"
)

var (
	buildValidate bool
	buildOnly     []string // 仅执行指定名称的任务
)

// StageRange 阶段范围
type StageRange struct {
	Start int // 开始阶段 (1-based)
	End   int // 结束阶段 (1-based), -1 表示到最后
}

// buildCmd build 命令
var buildCmd = &cobra.Command{
	Use:   "build [stage]",
	Short: "运行构建流程",
	Long: `运行构建流程，支持指定执行的阶段范围。

阶段参数格式:
  (空)     运行全部阶段
  N        只运行第 N 个阶段
  N-M      运行第 N 到第 M 个阶段
  N-       从第 N 个阶段运行到最后
  -M       从第一个阶段运行到第 M 个

阶段编号从 1 开始。

使用 --only 参数可以只执行指定名称的任务（支持多个）:
  --only "用户服务镜像"           只执行名为"用户服务镜像"的任务
  --only "用户服务" --only "订单服务"  同时执行两个指定的任务`,
	Example: `  xbuilder build              # 运行全部阶段
  xbuilder build 2            # 只运行第 2 个阶段
  xbuilder build 1-3          # 运行第 1 到第 3 个阶段
  xbuilder build 2-           # 从第 2 个阶段运行到最后
  xbuilder build -3           # 运行第 1 到第 3 个阶段
  xbuilder build -v           # 先验证配置，再运行
  xbuilder build --only "用户服务镜像"  # 只执行指定任务
  xbuilder build 2 --only "用户服务"   # 在第 2 阶段中只执行指定任务`,
	Args:              cobra.MaximumNArgs(1),
	RunE:              runBuild,
	ValidArgsFunction: completeBuildStages,
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().BoolVarP(&buildValidate, "validate", "v", false, "构建前先验证配置文件")
	buildCmd.Flags().StringArrayVarP(&buildOnly, "only", "o", nil, "只执行指定名称的任务（可多次使用）")

	// 注册 --only 参数的补全函数
	_ = buildCmd.RegisterFlagCompletionFunc("only", completeTaskNames)
}

// completeBuildStages 为 build 命令提供阶段补全
func completeBuildStages(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// 如果已经有参数了，不再补全
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// 尝试加载配置文件获取阶段列表
	configFile := GetConfigFile()
	if configFile == "" {
		configFile, _ = config.FindConfigFile()
	}

	if configFile == "" {
		// 没有配置文件，返回示例格式
		return []string{
			"1\t只运行第 1 个阶段",
			"2\t只运行第 2 个阶段",
			"1-2\t运行第 1 到第 2 个阶段",
			"2-\t从第 2 个阶段运行到最后",
		}, cobra.ShellCompDirectiveNoFileComp
	}

	// 加载配置
	loader := config.NewLoader(configFile)
	cfg, err := loader.Load()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// 生成补全建议
	var completions []string
	for i, stage := range cfg.Pipeline {
		stageNum := i + 1
		// 单个阶段
		completions = append(completions, fmt.Sprintf("%d\t%s", stageNum, stage.Name))
	}

	// 添加范围示例
	total := len(cfg.Pipeline)
	if total > 1 {
		completions = append(completions, fmt.Sprintf("1-%d\t运行全部 %d 个阶段", total, total))
		completions = append(completions, "2-\t从第 2 个阶段运行到最后")
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// completeTaskNames 为 --only 参数提供任务名称补全
func completeTaskNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// 尝试加载配置文件获取任务列表
	configFile := GetConfigFile()
	if configFile == "" {
		configFile, _ = config.FindConfigFile()
	}

	if configFile == "" {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// 加载配置
	loader := config.NewLoader(configFile)
	cfg, err := loader.Load()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// 收集所有任务名称
	var completions []string
	for _, stage := range cfg.Pipeline {
		for _, task := range stage.Tasks {
			// 添加任务名称和所属阶段作为描述
			completions = append(completions, fmt.Sprintf("%s\t[%s] %s", task.Name, stage.Name, task.Type))
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

func runBuild(cmd *cobra.Command, args []string) error {
	// 解析阶段范围
	var stageRange *StageRange
	if len(args) > 0 {
		var err error
		stageRange, err = parseStageRange(args[0])
		if err != nil {
			return fmt.Errorf("无效的阶段参数: %w", err)
		}
	}

	// 获取配置文件路径
	configFile := GetConfigFile()

	// 创建构建选项
	opts := app.BuildOptions{
		ConfigFile:   configFile,
		ValidateOnly: false,
		StageStart:   0,
		StageEnd:     -1,
		OnlyTasks:    buildOnly, // 仅执行指定任务
	}

	if stageRange != nil {
		opts.StageStart = stageRange.Start - 1 // 转换为 0-based
		if stageRange.End > 0 {
			opts.StageEnd = stageRange.End - 1
		} else {
			opts.StageEnd = -1 // -1 表示到最后
		}
	}

	// 如果需要先验证
	if buildValidate {
		fmt.Println("🔍 验证配置文件...")
		if err := app.ValidateConfig(configFile); err != nil {
			return err
		}
		fmt.Println("✅ 配置验证通过")
		fmt.Println()
	}

	// 运行构建
	return app.RunBuild(opts)
}

// parseStageRange 解析阶段范围参数
func parseStageRange(arg string) (*StageRange, error) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		return nil, nil
	}

	// 检查是否包含 "-"
	if strings.Contains(arg, "-") {
		parts := strings.SplitN(arg, "-", 2)

		var start, end int
		var err error

		// 解析开始
		if parts[0] == "" {
			start = 1 // 从第一个开始
		} else {
			start, err = strconv.Atoi(parts[0])
			if err != nil || start < 1 {
				return nil, fmt.Errorf("无效的开始阶段: %s", parts[0])
			}
		}

		// 解析结束
		if parts[1] == "" {
			end = -1 // 到最后
		} else {
			end, err = strconv.Atoi(parts[1])
			if err != nil || end < 1 {
				return nil, fmt.Errorf("无效的结束阶段: %s", parts[1])
			}
		}

		// 验证范围
		if end > 0 && start > end {
			return nil, fmt.Errorf("开始阶段 (%d) 不能大于结束阶段 (%d)", start, end)
		}

		return &StageRange{Start: start, End: end}, nil
	}

	// 单个数字
	n, err := strconv.Atoi(arg)
	if err != nil || n < 1 {
		return nil, fmt.Errorf("无效的阶段编号: %s", arg)
	}

	return &StageRange{Start: n, End: n}, nil
}
