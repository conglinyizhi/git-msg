package core

import (
	"fmt"
	"gitmsg/internal/llm"
	"gitmsg/internal/types"
	"os"
)

func SubcommandPing(cfg types.RemoteAPIConfig) int {
	const sys = "测试"
	const user = "请返回且只返回OK"
	configObject := types.Config{
		Cmd: types.CommandlineConfig{},
		Api: cfg,
	}
	if str, err := llm.SendReqCore(sys, user, configObject, false); err != nil {
		fmt.Fprintln(os.Stdout, "测试失败：", err)
		return 1
	} else {
		fmt.Fprintln(os.Stdout, "测试通过，LLM API 回复：", str)
		return 0
	}
}
