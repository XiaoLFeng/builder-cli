package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xiaolfeng/builder-cli/internal/app"
)

// validateCmd validate 命令
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "验证配置文件",
	Long:  `验证 xbuilder.yaml 配置文件的语法和内容是否正确。`,
	Example: `  xbuilder validate                 # 验证默认配置文件
  xbuilder validate -c custom.yaml  # 验证指定配置文件`,
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) error {
	configFile := GetConfigFile()

	if err := app.ValidateConfig(configFile); err != nil {
		return err
	}

	fmt.Println("✅ 配置文件验证通过")
	return nil
}
