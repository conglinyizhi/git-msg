package git

import (
	"fmt"
	"gitmsg/internal/types"
)

// 将当前路径添加到 git 仓库
func RunGitAdd(cmd types.CommandlineConfig) (string, error) {
	return RunCmdBase(cmd, "add", ".")
}

// 发起提交并携带对应的信息
func RunGitCommit(cmd types.CommandlineConfig, msgObj string) (string, error) {
	var cmdArg string
	description := msgObj
	body := cmd.Note
	if len(cmd.Note) > 0 {
		cmdArg = fmt.Sprintf("%s\n\n%s", description, body)
	} else {
		cmdArg = description
	}
	return RunCmdBase(cmd, "commit", "-m", cmdArg)
}
