package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/joho/godotenv"
)

const LLM_API = "https://open.bigmodel.cn/api/paas/v4/chat/completions"

const prompt = "请你作为一个专业的开发人员的角度，分析一下这个工程有什么改动，并且考虑为什么会有这种改动，用这些信息构成一个提交记录。要求控制文本必须用一行，最好在 30 字内，让他看起来像是一个版本控制里面的修改记录的说明消息文本，并且用中文输出。只输出纯文本提示而不是在两侧添加各种标点符号。"

// 定义结构体来解析 JSON 数据
type Event struct {
	ID      string `json:"id"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int    `json:"index"`
		FinishReason string `json:"finish_reason"`
		Delta        struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

func getToken() (string, error) {
	err := godotenv.Load()
	if err != nil {
		return "", err
	}
	token := os.Getenv("BIGMODEL_TOKEN")
	if token == "" {
		return "", fmt.Errorf("BIGMODEL_TOKEN 是空的，请通过 .env 填写")
	}
	return token, nil
}

func getDiffInDisk() (string, error) {
	commandObject, err := exec.Command("git", []string{"diff"}...).Output()
	if err != nil {
		return "", err
	}
	return string(commandObject), nil
}

// 获取暂存区的差异
func getDiffInStaged() (string, error) {
	commandObject, err := exec.Command("git", []string{"diff", "--staged"}...).Output()
	if err != nil {
		return "", err
	}
	return string(commandObject), nil
}

func getDiff() (string, bool, error) {
	var b = true
	// 获取项目的差异
	projectDiff, err := getDiffInDisk()
	if err != nil {
		return "", false, err
	}
	if projectDiff == "" {
		// 如果项目没有差异，尝试获取暂存区外的差异
		stashDiff, err := getDiffInStaged()
		b = false
		if err != nil {
			return "", false, err
		}
		if stashDiff != "" {
			projectDiff = stashDiff
		}
	}
	return projectDiff, b, nil
}

func main() {
	token, err := getToken()
	if err != nil {
		fmt.Fprintln(os.Stdout, []any{"获取大模型 key 失败：", err}...)
		return
	}
	commandObject, isNeedAddCommand, err := getDiff()
	if err != nil {
		fmt.Fprintln(os.Stdout, []any{"执行命令失败，原因：", err}...)
		return
	}

	req, err := http.NewRequest("POST", LLM_API, nil)
	if err != nil {
		fmt.Fprintln(os.Stdout, []any{"构建请求失败，原因：", err}...)
		return
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)
	jsonObjectMap := map[string]interface{}{
		"model": "glm-4-flash",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": prompt,
			},
			{
				"role":    "user",
				"content": string(commandObject),
			},
		},
		"type":   "text",
		"stream": true,
	}
	jsonObject, err := json.Marshal(jsonObjectMap)
	if err != nil {
		fmt.Fprintln(os.Stdout, []any{"json.Marshal JSON解析失败，原因：", err}...)
		return
	}

	// 根据字符串准备一个Reader
	bytesReader := bytes.NewReader(jsonObject)

	req.Body = io.NopCloser(bytesReader)
	req.ContentLength = int64(len(jsonObject))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stdout, []any{"HTTP 请求失败，原因：", err}...)
		return
	}
	defer resp.Body.Close()

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
			continue
		}
		// 提取并打印 delta.content
		if len(event.Choices) > 0 && event.Choices[0].Delta.Content != "" {
			content := event.Choices[0].Delta.Content
			fmt.Print(content)
		}
	}

	// 询问用户是否提交，如果需要，则提交
	goCommit, err := confirmation.New("need commit?", confirmation.Yes).RunPrompt()
	if err != nil {
		fmt.Fprintln(os.Stdout, []any{"交互命令出现异常：", err}...)
		return
	}
	// 是否需要添加文件到暂存区
	var goAdd = false
	if isNeedAddCommand {
		goAdd, err = confirmation.New("检测到暂存区外的文件差异，是否需要添加到暂存区？", confirmation.Yes).RunPrompt()
		if err != nil {
			fmt.Fprintln(os.Stdout, []any{"交互命令出现异常：", err}...)
			return
		}
	}

	if goAdd {
		fmt.Println("\n TODO: 添加文件到暂存区")
	}

	if goCommit {
		fmt.Println("\n TODO: 提交")
	}
}
