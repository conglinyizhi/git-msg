package main

import (
	"fmt"
	"os/exec"
)

// 获取差异，顺序尝试暂存区和工作区
func getDiff(cmd CommandlineConfig) (string, bool, error) {
	var isStagedDiff = false
	// 尝试获取暂存区的差异
	projectDiff, err := getDiffInStaged(cmd)
	if err != nil {
		return "", isStagedDiff, err
	}
	// 如果项目没有差异，尝试获取项目的差异
	if projectDiff == "" {
		isStagedDiff = true
		stashDiff, err := getDiffInDisk(cmd)
		if err != nil {
			return "", isStagedDiff, err
		}
		if stashDiff != "" {
			projectDiff = stashDiff
		}
	}
	if len(projectDiff) < 1 {
		return "", isStagedDiff, fmt.Errorf("git diff 没有捕获到有效信息，终止提交远程 API")
	}
	return projectDiff, isStagedDiff, nil
}

// 获取工作区差异
func getDiffInDisk(cmd CommandlineConfig) (string, error) {
	commandObject, err := exec.Command(cmd.git, []string{"diff", "-U10"}...).Output()
	if err != nil {
		return "", err
	}
	return string(commandObject), nil
}

// 获取暂存区的差异
func getDiffInStaged(cmd CommandlineConfig) (string, error) {
	commandObject, err := exec.Command(cmd.git, []string{"diff", "--staged", "-U10"}...).Output()
	if err != nil {
		return "", err
	}
	return string(commandObject), nil
}

// 获取仓库当前状态
func getStatus(cmd CommandlineConfig) (string, error) {
	commandObject, err := exec.Command(cmd.git, []string{"status", "-sb"}...).Output()
	if err != nil {
		return "", err
	}
	return string(commandObject), nil
}

// 将当前路径添加到 git 仓库
func runGitAdd(cmd CommandlineConfig) (string, error) {
	commandObject, err := exec.Command(cmd.git, []string{"add", "."}...).Output()
	if err != nil {
		return "", err
	}
	return string(commandObject), nil
}

// 发起提交并携带对应的信息
func runGitCommit(cmd CommandlineConfig, msg string) (string, error) {
	commandObject, err := exec.Command(cmd.git, []string{"commit", "-m", msg}...).Output()
	if err != nil {
		return "", err
	}
	return string(commandObject), nil
}
