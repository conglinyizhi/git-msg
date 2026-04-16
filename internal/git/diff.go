package git

import (
	"fmt"
	"gitmsg/internal/types"
)

// 获取差异，顺序尝试暂存区和工作区
func GetDiff(cmd types.CommandlineConfig) (string, bool, error) {
	var isFormDisk = false
	// 尝试获取暂存区的差异
	projectDiff, err := GetDiffInStaged(cmd)
	if err != nil {
		return "", isFormDisk, err
	}
	// 如果项目没有差异，尝试获取项目的差异
	if projectDiff == "" {
		isFormDisk = true
		stashDiff, err := GetDiffInDisk(cmd)
		if err != nil {
			return "", isFormDisk, err
		}
		if stashDiff != "" {
			projectDiff = stashDiff
		}
	}
	if len(projectDiff) < 1 {
		return "", isFormDisk, fmt.Errorf("git diff 没有捕获到有效信息，终止提交远程 API")
	}
	return projectDiff, isFormDisk, nil
}

// 获取工作区差异
func GetDiffInDisk(cmd types.CommandlineConfig) (string, error) {
	return RunCmdBase(cmd, "diff", "-U10")
}

// 获取暂存区的差异
func GetDiffInStaged(cmd types.CommandlineConfig) (string, error) {
	return RunCmdBase(cmd, "diff", "--staged", "-U10")
}
