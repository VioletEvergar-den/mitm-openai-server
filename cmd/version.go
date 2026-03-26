package cmd

import (
	"fmt"

	"github.com/llm-sec/mitm-openai-server/pkg/version"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Long:  `显示应用程序的版本、构建日期和Git提交信息。`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("MITM OpenAI Server %s\n", version.Version)
		fmt.Printf("构建日期: %s\n", version.BuildDate)
		fmt.Printf("Git提交: %s\n", version.GitCommit)
	},
}
