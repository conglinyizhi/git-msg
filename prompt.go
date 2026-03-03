package main

import (
	"fmt"
	"os"
)

const defaultPrompt = `你是拥有数年开发经验的专业全职软硬件开发者。现在请根据提供的差异信息，分析改动的原因，用这些信息构成一个提交记录，信息简略。这应该是 30 字内的一行文本，符合《约定式提交》。
type 要求使用英文，范围优先使用模块的英文名，取更短的名称，描述部分需用简体中文和少量英文与空格 ，注释会提示修改内容的含义，如有可供参考。
type & emoji 的关系是这样的：<emoji>test 🔬;style 🎨;chore 🧹;docs 📚;ci 🔄;build 🛠️;refactor ♻️;fix 🐛;feat ✨;perf 🚀;</emoji>
现在，请严格遵循模板内容完成变化总结，遵循上面提到的 emoji 列表，仅一行：	{{emoji}} <{{type}}>({{范围}}): {{描述}}
<范例>✨ test(cli.go): 新增了命令行功能 -t 参数，用于指定具体分类。</范例>`

// LLM 提示词
func getPrompt() string {
	skillBuff, err := os.ReadFile("./clyzhi-git.md")
	if err != nil {
		fmt.Fprintln(os.Stdout, []any{"读取文件失败，使用默认提示词", err}...)
	}
	return string(skillBuff)
}
