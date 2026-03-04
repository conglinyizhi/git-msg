package main

import (
	"bufio"
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/erikgeiser/promptkit/selection"
)

const appName = "git-msg"

// 当调用 LLM 接口后程序后处理报错时回退
func afterRemoteCallRollback(msg string) {
	tmpFilePath := filepath.Join(os.TempDir(), "git-commit-latest.txt")
	err := os.WriteFile(tmpFilePath, []byte(msg), 0666)
	if err != nil {
		fmt.Fprintln(os.Stdout, []any{"[回退]失败，原因：", err}...)
		return
	}
	println("[回退]大模型输出结果将保存到", tmpFilePath)
}

func callRemoteURL(diff string, config RemoteAPIConfig) (string, error) {
	req, err := http.NewRequest("POST", config.BASE_URL, nil)
	if err != nil {
		return "", fmt.Errorf("请求构建失败，详情：%w", err)
	}
	prompt := getPromptMain()

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+config.API_KEY)
	jsonObjectMap := map[string]interface{}{
		"model": config.MODEL_NAME,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": prompt,
			}, {
				"role":    "user",
				"content": diff,
			},
		},
		"type":   "text",
		"stream": true,
	}
	jsonObject, err := json.Marshal(jsonObjectMap)
	if err != nil {
		return "", fmt.Errorf("json.Marshal JSON解析失败，原因：%w", err)
	}

	// 根据字符串准备一个Reader
	bytesReader := bytes.NewReader(jsonObject)

	req.Body = io.NopCloser(bytesReader)
	req.ContentLength = int64(len(jsonObject))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP 请求失败，原因：%w", err)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("状态码不在预期内：%d", resp.StatusCode)
	}
	defer resp.Body.Close()
	var commitMessage strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		length := len(line)
		// 忽略空行
		if length == 0 {
			continue
		}

		// 解析 "data: " 前缀
		if length > 5 && line[:5] == "data:" {
			line = line[5:]
		}

		// "[DONE]" 标记结束
		if strings.TrimSpace(line) == "[DONE]" {
			break
		}

		var event Event
		// 解析 JSON 数据
		err := json.Unmarshal([]byte(line), &event)
		if err != nil {
			fmt.Println("解析 JSON 出错，原因:", err)
			fmt.Println("原始 JSON 数据:", line)
			continue
		}
		// 提取并打印 delta.content
		if len(event.Choices) > 0 && event.Choices[0].Delta.Content != "" {
			content := event.Choices[0].Delta.Content
			fmt.Print(content)
			commitMessage.WriteString(content)
		}
	}
	// 打印一个空行，避免大模型输出之后和后续内容写在一行内
	fmt.Println()
	return commitMessage.String(), nil
}

func callcmd(cmd CommandlineConfig, commitMessage string, isNeedAdd bool) error {
	// 询问用户是否提交，如果需要，则提交
	goCommit, err := confirmation.New("一切准备就绪，发起提交吗?", confirmation.Yes).RunPrompt()
	if err != nil {
		fmt.Fprintln(os.Stdout, []any{"交互命令出现异常：", err}...)
		return err
	}
	if !goCommit {
		return fmt.Errorf("用户取消提交")
	}
	// 是否需要添加文件到暂存区
	var goAdd = false
	const showGitStatusItem = "查看仓库状态"
	const exitSelectPromptItem = "退出"
	if isNeedAdd {
		selectPrompt := selection.New("本次操作使用暂存区外的文件差异，先添加到暂存区然后提交吗？", []string{"Yes", "No", showGitStatusItem, exitSelectPromptItem})
		selectPrompt.PageSize = 2
		for {
			spResult, err := selectPrompt.RunPrompt()
			if err != nil {
				fmt.Fprintln(os.Stdout, []any{"交互命令出现异常：", err}...)
				return err
			}
			if spResult == "Yes" {
				goAdd = true
				break
			}
			if spResult == exitSelectPromptItem {
				return fmt.Errorf("用户选择退出")
			}
			if spResult == showGitStatusItem {
				cmdResult, err := getStatus(cmd)
				if err != nil {
					fmt.Fprintln(os.Stdout, []any{"执行命令失败，原因：", err}...)
					return err
				}
				printCommandOutput(cmdResult, "status -sb")
				fmt.Println("*咳咳*，所以……")
			}
		}

	}

	if goAdd {
		stdout, err := runGitAdd(cmd)
		if err != nil {
			fmt.Fprintln(os.Stdout, []any{"执行命令失败，原因：", err}...)
			return err
		}
		printCommandOutput(stdout, "add")
	}
	if goCommit {
		stdout, err := runGitCommit(cmd, commitMessage)
		if err != nil {
			fmt.Fprintln(os.Stdout, []any{"执行命令失败，原因：", err}...)
			return err
		}
		printCommandOutput(stdout, "commit -m")
	}
	return nil
}

// 主函数
func main() {
	cmdConfig := parseCommandLineExData()
	if cmdConfig.init {
		rootDir, err := getConfigRootDir("")
		if err != nil {
			fmt.Fprintln(os.Stdout, []any{"定位配置目录失败：", err}...)
			os.Exit(1)
		}
		initConfigDir(rootDir)
		initSkillDir(rootDir)
		os.Exit(0)
	}
	config, err := getConfigValue()
	if err != nil {
		fmt.Fprintln(os.Stdout, []any{"获取大模型配置信息失败：", err}...)
		os.Exit(1)
	}

	diff, isNeedAddCommand, err := getDiff(cmdConfig)
	if err != nil {
		fmt.Fprintln(os.Stdout, []any{"获取差异信息失败，原因：", err}...)
		os.Exit(1)
	}

	commitMessage, err := callRemoteURL(diff, config)
	if err != nil {
		fmt.Fprintln(os.Stdout, []any{"调用远程大模型失败，原因：", err}...)
		os.Exit(1)
	}
	err = callcmd(cmdConfig, commitMessage, isNeedAddCommand)
	if err != nil {
		fmt.Fprintln(os.Stdout, []any{"运行指令的过程中出现错误，详情：", err}...)
		afterRemoteCallRollback(commitMessage)
		os.Exit(1)
	}
}

func initConfigDir(rootDir string) error {
	if err := os.MkdirAll(filepath.Join(rootDir), 0755); err != nil {
		return err
	}
	return initNewTomlFile(RemoteAPIConfig{})
}

//go:embed skill/*
var skillFilesEmbed embed.FS

func initSkillDir(rootDir string) error {
	if err := os.MkdirAll(filepath.Join(rootDir, "skill"), 0755); err != nil {
		return err
	}
	skillFiles, err := skillFilesEmbed.ReadDir("skill")
	if err != nil {
		return err
	}
	for _, skill := range skillFiles {
		if skill.IsDir() {
			continue
		}
		data, err := skillFilesEmbed.ReadFile(filepath.Join("skill", skill.Name()))
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(rootDir, "skill", skill.Name()), data, 0644); err != nil {
			return err
		}
	}
	return nil
}
