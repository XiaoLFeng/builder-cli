package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/xiaolfeng/builder-cli/resources"
)

var (
	initForce bool
)

// initCmd init 命令
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化最小配置文件",
	Long: `初始化一个最小的 xbuilder.yaml 配置文件。

该配置文件只包含最基础的 Maven 构建和 Docker 构建配置，
适合快速开始使用。如需完整模板，请使用 'xbuilder gen config' 命令。`,
	Example: `  xbuilder init          # 在当前目录创建 xbuilder.yaml
  xbuilder init -f       # 强制覆盖已存在的配置文件`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "强制覆盖已存在的配置文件")
}

func runInit(cmd *cobra.Command, args []string) error {
	configPath := "xbuilder.yaml"

	// 检查文件是否已存在
	if _, err := os.Stat(configPath); err == nil {
		if !initForce {
			return fmt.Errorf("配置文件 %s 已存在，使用 -f 参数强制覆盖", configPath)
		}
		fmt.Printf("⚠️  覆盖已存在的配置文件: %s\n", configPath)
	}

	// 从 embed 读取模板
	content, err := resources.GetTemplate("config/minimal.yaml")
	if err != nil {
		return fmt.Errorf("读取模板失败: %w", err)
	}

	// 写入配置文件
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		return fmt.Errorf("创建配置文件失败: %w", err)
	}

	absPath, _ := filepath.Abs(configPath)
	fmt.Printf("✅ 配置文件已创建: %s\n", absPath)
	fmt.Println()
	fmt.Println("下一步:")
	fmt.Println("  1. 编辑 xbuilder.yaml 配置你的构建流程")
	fmt.Println("  2. 运行 'xbuilder build' 开始构建")
	fmt.Println()
	fmt.Println("提示: 运行 'xbuilder gen config' 可生成完整的配置模板和示例脚本")

	return nil
}
