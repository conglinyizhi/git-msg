package tui

import (
	"fmt"

	"github.com/erikgeiser/promptkit/selection"
)

func SelectPrompt(list []string) (string, error) {
	const allNo = "都不行"
	selectPrompt := selection.New("请选择提示词", append(list, allNo))
	selectPrompt.PageSize = len(list) + 1
	commitMsg, err := selectPrompt.RunPrompt()
	if err != nil {
		return "", err
	}
	if commitMsg == allNo {
		return "", fmt.Errorf("用户取消提交")
	}
	return commitMsg, nil
}
