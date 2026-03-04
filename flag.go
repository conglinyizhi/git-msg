package main

import "github.com/spf13/pflag"

func parseCommandLineExData() CommandlineConfig {
	cmdConfig := CommandlineConfig{}
	cmdConfig.git = *pflag.StringP("git", "g", "git", "Git 指令替换，比如某些情况下用于替换为 yadm 等 Git Like 项目")
	cmdConfig.init = *pflag.Bool("init", false, "尝试对当前版本的 git-msg 所需环境进行最大的初始化")
	pflag.Parse()
	return cmdConfig
}
