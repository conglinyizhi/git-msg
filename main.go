package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
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

// 获取配置文件
func getToken() (string, string, string, error) {
	errorMessageBuild := func(message string) error {
		fmt.Println("提示：可以通过 .env 文件填写 BIGMODEL_TOKEN、LLM_API_URL、MODEL 三个参数")
		return fmt.Errorf("%s", message+"没有填写")
	}
	err := godotenv.Load()
	if err != nil {
		return "", "", "", err
	}
	TOKEN := os.Getenv("BIGMODEL_TOKEN")
	if TOKEN == "" {
		return "", "", "", errorMessageBuild("BIGMODEL_TOKEN")
	}
	LLM_API_URL := os.Getenv("LLM_API_URL")
	if LLM_API_URL == "" {
		return "", "", "", errorMessageBuild("LLM_API_URL")
	}
	MODEL := os.Getenv("MODEL")
	if MODEL == "" {
		return "", "", "", errorMessageBuild("MODEL")
	}
	return TOKEN, LLM_API_URL, MODEL, nil
}

// 获取工作区差异
func getDiffInDisk(gitCommand string) (string, error) {
	commandObject, err := exec.Command(gitCommand, []string{"diff", "-U10"}...).Output()
	if err != nil {
		return "", err
	}
	return string(commandObject), nil
}

// 获取暂存区的差异
func getDiffInStaged(gitCommand string) (string, error) {
	commandObject, err := exec.Command(gitCommand, []string{"diff", "--staged", "-U10"}...).Output()
	if err != nil {
		return "", err
	}
	return string(commandObject), nil
}

// 获取差异，顺序尝试暂存区和工作区
func getDiff(gitCommand string) (string, bool, error) {
	var isStagedDiff = false
	// 尝试获取暂存区的差异
	projectDiff, err := getDiffInStaged(gitCommand)
	if err != nil {
		return "", isStagedDiff, err
	}
	isStagedDiff = true
	if projectDiff == "" {
		// 如果项目没有差异，尝试获取项目的差异
		stashDiff, err := getDiffInDisk(gitCommand)
		if err != nil {
			return "", isStagedDiff, err
		}
		if stashDiff != "" {
			projectDiff = stashDiff
		}
	}
	if len(projectDiff) < 1 {
		return "", isStagedDiff, fmt.Errorf("git diff 没有捕获到有效信息，终止提交远程 API")
	}
	return projectDiff, isStagedDiff, nil
}

func callRemoteURL(diff string, TOKEN string, LLM_API_URL string, MODEL string) (string, error) {
	req, err := http.NewRequest("POST", LLM_API_URL, nil)
	if err != nil {
		return "", fmt.Errorf("请求构建失败，详情：%w", err)
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
	defer resp.Body.Close()
	var commitMessage strings.Builder
	fmt.Println("接口返回状态：", resp.StatusCode)
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
	return commitMessage.String(), nil
}

// 主函数
func main() {
	var gitCommand = flag.String("git", "git", "Git 指令替换，比如某些情况下用于替换为 yadm 等 Git Like 项目")

	TOKEN, LLM_API_URL, MODEL, err := getToken()
	if err != nil {
		fmt.Fprintln(os.Stdout, []any{"获取大模型 key 失败：", err}...)
		return
	}
	diff, isNeedAddCommand, err := getDiff(*gitCommand)
	if err != nil {
		fmt.Fprintln(os.Stdout, []any{"执行命令失败，原因：", err}...)
		return
	}
	commitMessage, err := callRemoteURL(diff, TOKEN, LLM_API_URL, MODEL)
	if err != nil {
		fmt.Fprintln(os.Stdout, []any{"callRemoteURL 函数报错，原因：", err}...)
		return
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
		stdout, err := exec.Command("git", []string{"add", "."}...).Output()
		if err != nil {
			fmt.Fprintln(os.Stdout, []any{"执行命令失败，原因：", err}...)
			return
		}
		fmt.Println("add . 输出的内容")
		fmt.Println(string(stdout))
	}
	if goCommit {
		stdout, err := exec.Command("git", []string{"commit", "-m", commitMessage}...).Output()
		if err != nil {
			fmt.Fprintln(os.Stdout, []any{"执行命令失败，原因：", err}...)
			return
		}
		fmt.Println("commit -m ... 输出的内容")
		fmt.Println(string(stdout))
	}
}
