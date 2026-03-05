package main

import "github.com/spf13/pflag"

func parseCommandLineExData() CommandlineConfig {
	cmdConfig := CommandlineConfig{}
	pflag.StringVarP(&cmdConfig.git, "git", "g", "git", "Git 指令替换，比如某些情况下用于替换为 yadm 等 Git Like 项目")
	pflag.BoolVar(&cmdConfig.init, "init", false, "尝试对当前版本的 git-msg 所需环境进行最大的初始化")
	pflag.BoolVar(&cmdConfig.ping, "ping", false, "尝试对配置中的远程大模型发起测试请求，最小提示词")
	pflag.Parse()
	return cmdConfig
}
