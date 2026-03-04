package main

import (
	"fmt"
	"os"
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
