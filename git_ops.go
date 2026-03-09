package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// 获取差异，顺序尝试暂存区和工作区
func getDiff(cmd CommandlineConfig) (string, bool, error) {
	var isFormDisk = false
	// 尝试获取暂存区的差异
	projectDiff, err := getDiffInStaged(cmd)
	if err != nil {
		return "", isFormDisk, err
	}
	// 如果项目没有差异，尝试获取项目的差异
	if projectDiff == "" {
		isFormDisk = true
		stashDiff, err := getDiffInDisk(cmd)
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
func getDiffInDisk(cmd CommandlineConfig) (string, error) {
	return runCmdBase(cmd, "diff", "-U10")
}

// 获取暂存区的差异
func getDiffInStaged(cmd CommandlineConfig) (string, error) {
	return runCmdBase(cmd, "diff", "--staged", "-U10")
}

// 获取仓库当前状态
func getStatus(cmd CommandlineConfig) (string, error) {
	return runCmdBase(cmd, "status", "-sb")
}

// 将当前路径添加到 git 仓库
func runGitAdd(cmd CommandlineConfig) (string, error) {
	return runCmdBase(cmd, "add", ".")
}

// 发起提交并携带对应的信息
func runGitCommit(cmd CommandlineConfig, msg string) (string, error) {
	return runCmdBase(cmd, "commit", "-m", msg)
}

func runCmdBase(cmd CommandlineConfig, args ...string) (string, error) {
	stdText, err := exec.Command(cmd.git, args...).CombinedOutput()
	if err != nil {
		fmt.Println("runCmdBase: 运行 ", strings.Join(args, " "), "时出现了错误，完整情况如下")
		fmt.Println(string(stdText))
	}
	return string(stdText), err
}
