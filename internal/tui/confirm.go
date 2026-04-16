package tui

import (
	"fmt"
	"gitmsg/internal/git"
	"gitmsg/internal/types"
	"gitmsg/internal/utils"

	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/erikgeiser/promptkit/selection"
)

func CallCmd(cfg types.Config, commitMessage string, isNeedAdd bool) error {
	// 询问用户是否提交，如果需要，则提交
	goCommit, err := confirmation.New("一切准备就绪，发起提交吗?", confirmation.Yes).RunPrompt()
	if err != nil {
		return fmt.Errorf("交互命令出现异常: %w", err)
	}
	if !goCommit {
		return fmt.Errorf("用户取消提交")
	}
	// 是否需要添加文件到暂存区
	var goAdd = false
	const YesItem = "是(Yes)"
	const NoItem = "否(No)"
	const showGitStatusItem = "查看仓库状态"
	const exitSelectPromptItem = "退出"
	if isNeedAdd {
		selectPrompt := selection.New("本次操作使用暂存区外的文件差异，先添加到暂存区然后提交吗？", []string{YesItem, NoItem, showGitStatusItem, exitSelectPromptItem})
		selectPrompt.PageSize = 2
	loop:
		for {
			spResult, err := selectPrompt.RunPrompt()
			if err != nil {
				return fmt.Errorf("交互命令出现异常: %w", err)
			}
			switch spResult {
			case YesItem:
				goAdd = true
				break loop
			case NoItem:
				goAdd = false
				break loop
			case exitSelectPromptItem:
				return fmt.Errorf("用户选择退出")
			case showGitStatusItem:
				cmdResult, err := git.GetStatus(cfg.Cmd)
				if err != nil {
					return err
				}
				utils.PrintCommandOutput(cmdResult, "status -sb")
				fmt.Println("*咳咳*，所以……")
			}
		}
	}

	if goAdd {
		stdout, err := git.RunGitAdd(cfg.Cmd)
		if err != nil {
			return err
		}
		utils.PrintCommandOutput(stdout, "add")
	}
	if goCommit {
		stdout, err := git.RunGitCommit(cfg.Cmd, commitMessage)
		if err != nil {
			return err
		}
		utils.PrintCommandOutput(stdout, "commit -m")
	}
	return nil
}
