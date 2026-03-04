package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

// readfileToString 读取文件内容并返回字符串
func readfileToString(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	return string(data), err
}

// printCommandOutput 格式化打印命令执行结果
func printCommandOutput(stdout string, command string) {
	const printLine = "---"
	if len(stdout) < 1 {
		fmt.Println(command + " 执行完成，空输出")
	} else {
		fmt.Println(printLine + command + " 输出的内容" + printLine)
		fmt.Println(stdout)
	}
}

// getConfigRootDir 获取系统规范的软件配置目录
func getConfigRootDir(fp string) (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	switch runtime.GOOS {
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = filepath.Join(usr.HomeDir, "AppData", "Local")
		}
		return filepath.Join(localAppData, appName, "Config", fp), nil
	case "darwin":
		return filepath.Join(usr.HomeDir, "Library", "Preferences", appName, fp), nil
	}
	// Linux 和其他系统
	return filepath.Join(usr.HomeDir, ".config", appName, fp), nil
}
