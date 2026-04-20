package llm

import (
	"errors"
	"fmt"
	"gitmsg/internal/utils"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/erikgeiser/promptkit"
	"github.com/erikgeiser/promptkit/selection"
)

const defaultPrompt = `你是拥有数年开发经验的专业全职软硬件开发者。现在请根据提供的差异信息，分析改动的原因，用这些信息构成一个提交记录，信息简略。这应该是 30 字内的一行文本，符合《约定式提交》。
type 要求使用英文，范围优先使用模块的英文名，取更短的名称，描述部分需用简体中文和少量英文与空格 ，注释会提示修改内容的含义，如有可供参考。
type & emoji 的关系是这样的：<emoji>test 🔬;style 🎨;chore 🧹;docs 📚;ci 🔄;build 🛠️;refactor ♻️;fix 🐛;feat ✨;perf 🚀;</emoji>
现在，请严格遵循模板内容完成变化总结，遵循上面提到的 emoji 列表，仅一行：	{{emoji}} <{{type}}>({{范围}}): {{描述}}
<范例>🔬 test(cli.go): 新增了命令行功能 -t 参数，用于指定具体分类。</范例>`

func GenErrorAndUseDefaultPrompt(errMsg string, err error) (string, error) {
	log.Println(errMsg, "失败，使用默认提示词", err)
	if errors.Is(err, promptkit.ErrAborted) {
		return "", fmt.Errorf("用户取消提交，预期内错误:%w", err)
	} else {
		return "", fmt.Errorf("提示词获取逻辑失败，错误详情：%w", err)
	}
}

// LLM 提示词
func GetPromptMain() (string, error) {
	skillDir, err := utils.GetConfigRootDir("./skill")
	if err != nil {
		return GenErrorAndUseDefaultPrompt("定位 skill 目录", err)
	}
	pathFileList, err := os.ReadDir(skillDir)
	if err != nil {
		return GenErrorAndUseDefaultPrompt("访问 skill 目录", err)
	}
	skillFile, skillFileNames := GetSkillFileList(pathFileList)
	var skillFileBody string
	skillFileLength := len(skillFile)
	if skillFileLength == 0 {
		return GenErrorAndUseDefaultPrompt("查找有效 skill 文件", nil)
	}
	if skillFileLength == 1 {
		return TryReadSkillFile(skillFileNames)
	}
	return SelectSkillFile(skillDir, skillFileBody, skillFileNames)
}

func SelectSkillFile(skillDir, skillFileBody string, skillFileNames []string) (string, error) {
	sp := selection.New("请选择要执行的 System Prompt", skillFileNames)
	sp.PageSize = 10
	spResult, err := sp.RunPrompt()
	fullPath := filepath.Join(skillDir, spResult)
	skillFileBody, err = utils.ReadfileToString(fullPath)
	if err != nil {
		return GenErrorAndUseDefaultPrompt("读取"+fullPath+"文件", err)
	}
	return skillFileBody, nil
}

func TryReadSkillFile(skillFileName []string) (string, error) {
	skillDir, err := utils.GetConfigRootDir("./skill")
	if err != nil {
		return GenErrorAndUseDefaultPrompt("定位 skill 目录", err)
	}
	oneFileName := skillFileName[0]
	skillFileBody, err := utils.ReadfileToString(filepath.Join(skillDir, oneFileName))
	if err != nil {
		return GenErrorAndUseDefaultPrompt("读取文件", err)
	}
	return skillFileBody, nil
}

// 获取所有的技能(skill)文件
func GetSkillFileList(pathFileList []os.DirEntry) ([]os.DirEntry, []string) {
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
