package cmd

import (
	"gitmsg/internal/core"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化项目配置",
	Long:  `将内嵌的 markdown 文件和 llm 配置初始化放到配置目录中`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return core.SubcommandInit()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
