package git

import (
	"fmt"
	"gitmsg/internal/types"
	"os/exec"
	"strings"
)

// 获取仓库当前状态
func GetStatus(cmd types.CommandlineConfig) (string, error) {
	return RunCmdBase(cmd, "status", "-sb")
}

func RunCmdBase(cmd types.CommandlineConfig, args ...string) (string, error) {
	stdText, err := exec.Command(cmd.Git, args...).CombinedOutput()
	if err != nil {
		fmt.Println("runCmdBase: 运行 ", strings.Join(args, " "), "时出现了错误，完整情况如下")
		fmt.Println(string(stdText))
	}
	return string(stdText), err
}
