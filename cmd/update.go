package cmd

import (
	"fmt"
	"os"

	"github.com/llm-sec/mitm-openai-server/pkg/updater"
	"github.com/llm-sec/mitm-openai-server/pkg/version"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().Bool("check", false, "仅检查是否有更新")
	updateCmd.Flags().Bool("git", false, "通过 Git 拉取源码并编译")
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "更新到最新版本",
	Long: `检查并更新到最新版本。

支持两种更新方式:
1. 下载预编译的二进制文件 (默认)
2. 通过 Git 拉取源码并编译 (--git)

示例:
  mitm-openai-server update          # 下载预编译版本
  mitm-openai-server update --git    # Git 拉取并编译
  mitm-openai-server update --check  # 仅检查是否有更新`,
	Run: func(cmd *cobra.Command, args []string) {
		checkOnly, _ := cmd.Flags().GetBool("check")
		useGit, _ := cmd.Flags().GetBool("git")

		fmt.Printf("当前版本: %s\n", version.Version)
		fmt.Printf("构建日期: %s\n", version.BuildDate)
		fmt.Printf("Git提交: %s\n\n", version.GitCommit)

		if useGit {
			if checkOnly {
				fmt.Println("--check 参数与 --git 参数不兼容")
				return
			}

			if err := updater.UpdateViaGit(); err != nil {
				fmt.Printf("更新失败: %v\n", err)
				os.Exit(1)
			}
			return
		}

		fmt.Println("正在检查更新...")

		hasUpdate, latestVersion, err := updater.CheckForUpdate(version.Version)
		if err != nil {
			fmt.Printf("检查更新失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("最新版本: %s\n", latestVersion)

		if !hasUpdate {
			fmt.Println("\n已是最新版本，无需更新。")
			return
		}

		fmt.Println("\n发现新版本!")

		if checkOnly {
			fmt.Printf("下载地址: %s\n", updater.GitHubReleasesURL)
			return
		}

		if err := updater.PerformUpdate(version.Version); err != nil {
			fmt.Printf("更新失败: %v\n", err)
			os.Exit(1)
		}
	},
}
