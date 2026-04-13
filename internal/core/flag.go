package core

import (
	"gitmsg/internal/types"

	"github.com/spf13/pflag"
)

func ParseCommandLineExData() types.CommandlineConfig {
	cmdConfig := types.CommandlineConfig{}
	pflag.StringVarP(&cmdConfig.Git, "git", "g", "git", "Git 指令替换，比如某些情况下用于替换为 yadm 等 Git Like 项目")
	pflag.IntVarP(&cmdConfig.Loop, "loop", "l", 1, "同时发起几个请求，然后使用交互界面选择一个更合适的")
	pflag.BoolVar(&cmdConfig.Init, "init", false, "尝试对当前版本的 git-msg 所需环境进行最大的初始化")
	pflag.BoolVar(&cmdConfig.Ping, "ping", false, "尝试对配置中的远程大模型发起测试请求，最小提示词")
	pflag.Parse()
	return cmdConfig
}
