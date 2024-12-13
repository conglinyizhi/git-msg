package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"

	"github.com/joho/godotenv"
)

const LLM_API = "https://open.bigmodel.cn/api/paas/v4/chat/completions"

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Fprintln(os.Stdout, []any{"Load .env file failed,%v", err}...)
		return
	}
	token := os.Getenv("BIGMODEL_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stdout, []any{"BIGMODEL_TOKEN is empty"}...)
		return
	}
	req, err := http.NewRequest("POST", LLM_API, nil)
	if err != nil {
		fmt.Fprintln(os.Stdout, []any{"NewRequest failed,%v", err}...)
		return
	}
	commandObject, err := exec.Command("git", []string{"diff"}...).Output()
	if err != nil {
		fmt.Fprintln(os.Stdout, []any{"exec.Command failed,%v", err}...)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)
	talkListMap := []map[string]string{
		{
			"role":    "system",
			"content": "帮我分析一下这个工程有什么改动，简单说说。",
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
	}
	jsonObject, err := json.Marshal(jsonObjectMap)
	fmt.Println(string(jsonObject))
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintln(os.Stdout, []any{"Read response failed,%v", err}...)
		return
	}

	fmt.Println(string(body))
}
