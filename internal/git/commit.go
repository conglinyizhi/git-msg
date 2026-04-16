package git

import "gitmsg/internal/types"

// 将当前路径添加到 git 仓库
func RunGitAdd(cmd types.CommandlineConfig) (string, error) {
	return RunCmdBase(cmd, "add", ".")
}

// 发起提交并携带对应的信息
func RunGitCommit(cmd types.CommandlineConfig, msg string) (string, error) {
	return RunCmdBase(cmd, "commit", "-m", msg)
}
