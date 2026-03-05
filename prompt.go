package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/erikgeiser/promptkit/selection"
)

const defaultPrompt = `你是拥有数年开发经验的专业全职软硬件开发者。现在请根据提供的差异信息，分析改动的原因，用这些信息构成一个提交记录，信息简略。这应该是 30 字内的一行文本，符合《约定式提交》。
type 要求使用英文，范围优先使用模块的英文名，取更短的名称，描述部分需用简体中文和少量英文与空格 ，注释会提示修改内容的含义，如有可供参考。
type & emoji 的关系是这样的：<emoji>test 🔬;style 🎨;chore 🧹;docs 📚;ci 🔄;build 🛠️;refactor ♻️;fix 🐛;feat ✨;perf 🚀;</emoji>
现在，请严格遵循模板内容完成变化总结，遵循上面提到的 emoji 列表，仅一行：	{{emoji}} <{{type}}>({{范围}}): {{描述}}
<范例>🔬 test(cli.go): 新增了命令行功能 -t 参数，用于指定具体分类。</范例>`

func genErrorAndUseDefaultPrompt(errMsg string, e error) string {
	fmt.Fprintln(os.Stdout, errMsg, "失败，使用默认提示词", e)
	return defaultPrompt
}

// LLM 提示词
func getPromptMain() string {
	skillDir, err := getConfigRootDir("./skill")
	if err != nil {
		return genErrorAndUseDefaultPrompt("定位 skill 目录", err)
	}
	pathFileList, err := os.ReadDir(skillDir)
	if err != nil {
		return genErrorAndUseDefaultPrompt("访问 skill 目录", err)
	}
	skillFile, skillFileName := getSkillFileList(pathFileList)
	var skillFileBody string
	skillFileLength := len(skillFile)
	if skillFileLength == 0 {
		return genErrorAndUseDefaultPrompt("查找有效 skill 文件", nil)
	}
	if skillFileLength == 1 {
		return tryReadSkillFile(skillFileName)
	}
	sp := selection.New("请选择要执行的 System Prompt", skillFileName)
	sp.PageSize = 10
	spResult, err := sp.RunPrompt()
	fullPath := filepath.Join(skillDir, spResult)
	skillFileBody, err = readfileToString(fullPath)
	if err != nil {
		return genErrorAndUseDefaultPrompt("读取"+fullPath+"文件", err)
	}
	return skillFileBody
}

func tryReadSkillFile(skillFileName []string) string {
	skillDir, err := getConfigRootDir("./skill")
	if err != nil {
		return genErrorAndUseDefaultPrompt("定位 skill 目录", err)
	}
	oneFileName := skillFileName[0]
	skillFileBody, err := readfileToString(filepath.Join(skillDir, oneFileName))
	if err != nil {
		return genErrorAndUseDefaultPrompt("读取文件", err)
	}
	return skillFileBody
}

// 获取所有的技能(skill)文件
func getSkillFileList(pathFileList []os.DirEntry) ([]os.DirEntry, []string) {
	isMarkdownByFileExt := func(file os.DirEntry) bool {
		fileName := file.Name()
		nameArraySplitDot := strings.Split(fileName, ".")
		splitDotLength := len(nameArraySplitDot)
		if splitDotLength < 2 {
			return false
		}
		lastNameSplit := strings.ToLower(nameArraySplitDot[len(nameArraySplitDot)-1])
		return strings.HasPrefix(lastNameSplit, "md")
	}
	skillFile := []os.DirEntry{}
	skillFileName := []string{}
	for _, file := range pathFileList {
		if isMarkdownByFileExt(file) {
			skillFile = append(skillFile, file)
			skillFileName = append(skillFileName, file.Name())
		}
	}
	return skillFile, skillFileName
}
