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

	"github.com/joho/godotenv"
)

const LLM_API = "https://open.bigmodel.cn/api/paas/v4/chat/completions"

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
		return "", fmt.Errorf("BIGMODEL_TOKEN is empty")
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

func getDiff() (string, error) {
	// 获取项目的差异
	projectDiff, err := getDiffInDisk()
	if err != nil {
		return "", err
	}
	if projectDiff == "" {
		// 如果项目没有差异，尝试获取暂存区的差异
		stashDiff, err := getDiffInStaged()
		if err != nil {
			return "", err
		}
		if stashDiff != "" {
			projectDiff = stashDiff
		} else {
			return "", fmt.Errorf("no diff infomation")
		}
	}
	return projectDiff, nil
}

func main() {
	token, err := getToken()
	if err != nil {
		fmt.Fprintln(os.Stdout, []any{"getToken failed:", err}...)
		return
	}
	req, err := http.NewRequest("POST", LLM_API, nil)
	if err != nil {
		fmt.Fprintln(os.Stdout, []any{"NewRequest failed:", err}...)
		return
	}
	commandObject, err := getDiff()
	if err != nil {
		fmt.Fprintln(os.Stdout, []any{"exec.Command failed:", err}...)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)
	talkListMap := []map[string]string{
		{
			"role":    "system",
			"content": "帮我分析一下这个工程有什么改动，控制文本必须用一行，最好在 30 字内，让他看起来像是一个版本控制里面的修改记录的说明消息文本，并且用中文输出",
		},
		{
			"role":    "user",
			"content": string(commandObject),
		},
	}
	jsonObjectMap := map[string]interface{}{
		"model":    "glm-4-flash",
		"messages": talkListMap,
		"type":     "text",
		"stream":   true,
	}
	jsonObject, err := json.Marshal(jsonObjectMap)
	if err != nil {
		fmt.Fprintln(os.Stdout, []any{"json.Marshal failed,%v", err}...)
		return
	}

	// 根据字符串准备一个Reader
	bytesReader := bytes.NewReader(jsonObject)
	bytesReaderCloser := io.NopCloser(bytesReader)

	req.Body = bytesReaderCloser
	req.ContentLength = int64(len(jsonObject))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stdout, []any{"HTTP request failed,%v", err}...)
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
			fmt.Println("Error parsing JSON:", err)
			continue
		}
		// 提取并打印 delta.content
		if len(event.Choices) > 0 && event.Choices[0].Delta.Content != "" {
			content := event.Choices[0].Delta.Content
			fmt.Print(content)
		}
	}
}
