package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

func sendReqCore(sys, user string, cfg Config, isStreamMode bool) (string, error) {
	ctx := context.Background()
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   cfg.api.MODEL_NAME,
		APIKey:  cfg.api.API_KEY,
		BaseURL: cfg.api.BASE_URL,
	})
	if err != nil {
		log.Println("创建 chat model 失败")
		return "", err
	}
	messageList := []*schema.Message{
		schema.SystemMessage(sys),
		schema.UserMessage(user),
	}
	if isStreamMode {
		sr, err := chatModel.Stream(ctx, messageList)
		if err != nil {
			log.Println("创建 stream 失败")
			return "", err
		}
		return reportStream(sr)
	} else {
		txt, err := chatModel.Generate(ctx, messageList)
		if err != nil {
			return "", err
		}
		return txt.Content, nil
	}
}

func reportStream(sr *schema.StreamReader[*schema.Message]) (string, error) {
	defer sr.Close()
	var strBuff strings.Builder
	for {
		message, err := sr.Recv()
		if err == io.EOF { // 流式输出结束
			fmt.Println()
			return strBuff.String(), nil
		}
		if err != nil {
			fmt.Println()
			log.Fatalf("recv failed: %v", err)
			return strBuff.String(), err
		}
		fmt.Print(message.Content)
		strBuff.WriteString(message.Content)
	}
}
