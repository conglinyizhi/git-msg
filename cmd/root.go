package cmd

import (
	"gitmsg/internal/core"
	"gitmsg/internal/types"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd 表示在没有子命令时调用的基础命令
var rootCmd = &cobra.Command{
	Use:   "gitmsg",
	Short: "你的应用的简短描述",
	Long: `一个更长的描述，可以跨越多行，并且可能包含
使用你的应用的示例和用法。例如：

Cobra 是一个用于 Go 的 CLI 库，可以增强应用的能力。
这个应用是一个工具，用于生成所需的文件
以快速创建一个 Cobra 应用。`,
	// 如果你的裸应用有关联的操作，请取消下面这行的注释：
	Run: func(cmd *cobra.Command, args []string) {
		core.BootloaderMain()
	},
}

// Execute 将所有子命令添加到根命令并正确设置标志。
// 这由 main.main() 调用。对 rootCmd 只需要发生一次。
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// 在这里定义你的标志和配置设置。
	// Cobra 支持持久化标志，如果在这里定义，
	// 它们将对整个应用全局可用。

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件（默认为 $HOME/.gitmsg.yaml）")

	// Cobra 也支持本地标志，它们只在
	// 直接调用此操作时运行。
	// rootCmd.Flags().BoolP("toggle", "t", false, "toggle 的帮助信息")
	pflag := rootCmd.Flags()
	cmdConfig := types.CommandlineConfig{}
	pflag.StringVarP(&cmdConfig.Git, "git", "g", "git", "Git 指令替换，比如某些情况下用于替换为 yadm 等 Git Like 项目")
	pflag.IntVarP(&cmdConfig.Loop, "loop", "l", 1, "同时发起几个请求，然后使用交互界面选择一个更合适的")
	pflag.BoolVar(&cmdConfig.Init, "init", false, "尝试对当前版本的 git-msg 所需环境进行最大的初始化")
	pflag.BoolVar(&cmdConfig.Ping, "ping", false, "尝试对配置中的远程大模型发起测试请求，最小提示词")
}
