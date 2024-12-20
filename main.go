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

func getToken() (string, string, string, error) {
	err := godotenv.Load()
	if err != nil {
		return "", "", "", err
	}
	TOKEN := os.Getenv("BIGMODEL_TOKEN")
	if TOKEN == "" {
		return "", "", "", fmt.Errorf("BIGMODEL_TOKEN 是空的，请通过 .env 填写")
	}
	LLM_API_URL := os.Getenv("LLM_API_URL")
	if LLM_API_URL == "" {
		return "", "", "", fmt.Errorf("LLM_API_URL 是空的，请通过.env 填写")
	}
	MODEL := os.Getenv("MODEL")
	if MODEL == "" {
		return "", "", "", fmt.Errorf("MODEL 是空的，请通过.env 填写")
	}
	return TOKEN, LLM_API_URL, MODEL, nil
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
	TOKEN, LLM_API_URL, MODEL, err := getToken()
	if err != nil {
		fmt.Fprintln(os.Stdout, []any{"获取大模型 key 失败：", err}...)
		return
	}
	// promptByte, err := os.ReadFile("./prompt.txt")
	// if err != nil {
	// 	fmt.Fprintln(os.Stdout, []any{"读取 prompt.txt 文件失败，原因：", err}...)
	// 	return
	// }
	diff, isNeedAddCommand, err := getDiff()
	if err != nil {
		fmt.Fprintln(os.Stdout, []any{"执行命令失败，原因：", err}...)
		return
	}

	req, err := http.NewRequest("POST", LLM_API_URL, nil)
	if err != nil {
		fmt.Fprintln(os.Stdout, []any{"构建请求失败，原因：", err}...)
		return
	}
	prompt := getPrompt()

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+TOKEN)
	jsonObjectMap := map[string]interface{}{
		"model": MODEL,
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
	commitMessage := ""
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
			commitMessage += content
		}
	}
	fmt.Println()
	// 询问用户是否提交，如果需要，则提交
	goCommit, err := confirmation.New("一切准备就绪，发起提交吗?", confirmation.Yes).RunPrompt()
	if err != nil {
		fmt.Fprintln(os.Stdout, []any{"交互命令出现异常：", err}...)
		return
	}
	if !goCommit {
		fmt.Println("取消提交")
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
