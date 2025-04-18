package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// 版本信息
var (
	Version   = "1.0.0"
	BuildDate = "未知"
	GitCommit = "未知"
)

// versionCmd 表示version命令
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Long:  `显示应用程序的版本、构建日期和Git提交信息。`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("假OpenAPI服务器 %s\n", Version)
		fmt.Printf("构建日期: %s\n", BuildDate)
		fmt.Printf("Git提交: %s\n", GitCommit)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
