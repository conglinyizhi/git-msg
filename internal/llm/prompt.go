package llm

import (
	"errors"
	"fmt"
	"gitmsg/embed"
	"gitmsg/internal/utils"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/erikgeiser/promptkit"
	"github.com/erikgeiser/promptkit/selection"
)

func GenErrorAndUseDefaultPrompt(errMsg string, err error) (string, error) {
	log.Println(errMsg, "失败，使用默认提示词", err)
	if errors.Is(err, promptkit.ErrAborted) {
		return "", fmt.Errorf("用户取消提交，预期内错误:%w", err)
	} else {
		return embed.DefaultPrompt, fmt.Errorf("提示词获取逻辑失败，错误详情：%w，将会使用默认提示词", err)
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
