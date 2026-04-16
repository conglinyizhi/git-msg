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
//
// TODO: 这里也传入了 cmd 对象，就直接用就行，CommitMessageObject 可以直接连根删除，但这是大任务，暂时搁置
func RunGitCommit(cmd types.CommandlineConfig, msgObj types.CommitMessageObject) (string, error) {
	var msg string
	if len(msgObj.Body) > 0 {
		msg = fmt.Sprintf("%s\n\n%s", msgObj.Description, msgObj.Body)
	} else {
		msg = msgObj.Description
	}
	return RunCmdBase(cmd, "commit", "-m", msg)
}
