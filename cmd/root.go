package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xiaolfeng/builder-cli/pkg/version"
)

var (
	cfgFile string
)

// rootCmd 根命令
var rootCmd = &cobra.Command{
	Use:   "xbuilder",
	Short: "xbuilder - 美观的 TUI 构建工具",
	Long: `xbuilder 是一个基于 Bubble Tea + Lipgloss 的美观 TUI 构建工具，
支持 Maven 构建、Go 构建、Docker 镜像构建推送、SSH 远程部署。

作者: 筱锋 (xiao_lfeng)
项目: https://github.com/XiaoLFeng/builder-cli

使用示例:
  xbuilder init                    # 初始化最小配置文件
  xbuilder gen                     # 生成完整模板和 scripts
  xbuilder build                   # 运行全部构建流程
  xbuilder build 2                 # 只运行第 2 个阶段
  xbuilder build 1-3               # 运行第 1 到第 3 个阶段
  xbuilder build 2-                # 从第 2 个阶段运行到最后`,
	Version: version.Version,
}

// Execute 执行根命令
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// 全局 flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "配置文件路径 (默认: xbuilder.yaml)")

	// 版本信息格式
	rootCmd.SetVersionTemplate(fmt.Sprintf(`xbuilder version %s
Author: 筱锋 (xiao_lfeng)
GitHub: https://github.com/XiaoLFeng/builder-cli
`, version.Version))
}

// GetConfigFile 获取配置文件路径
func GetConfigFile() string {
	return cfgFile
}
