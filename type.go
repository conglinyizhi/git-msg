package main

// 定义结构体来解析 JSON 数据
type Event struct {
	ID         string `json:"id"`
	Created    int64  `json:"created"`
	MODEL_NAME string `json:"model"`
	Choices    []struct {
		Index        int    `json:"index"`
		FinishReason string `json:"finish_reason"`
		Delta        struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

type RemoteAPIConfig struct {
	API_KEY    string
	BASE_URL   string
	MODEL_NAME string
}

type CommandlineConfig struct {
	git  string
	loop int
	init bool
	ping bool
}

type CommitChan struct {
	data  string
	err   error
	index int
}
