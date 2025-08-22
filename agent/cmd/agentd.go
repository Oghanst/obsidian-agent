package main

import (
	"context"

	"github.com/obsidian-agent/internal/client"
	"github.com/obsidian-agent/internal/logger"
	"github.com/obsidian-agent/internal/property"
	"github.com/sashabaranov/go-openai"
)

func main() {

	property.LoadConfig("/Users/jianghaojun/Projects/obsidian-agent/agent/config/config.json")
	config := property.GetConfig()
	logger, err := logger.New(config.LogDir + "/agent.log")
	if err != nil {
		panic(err)
	}
	logger.Info("Agent started with log directory: %s", config.LogDir)
	deepseekClient := client.NewDeepSeekClient(config.Apikey)
	resp, err := deepseekClient.Client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: "deepseek-chat", 
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "用 Go 写一个 quicksort 示例",
				},
			},
		},
	)

	if err != nil {
		logger.Error("ChatCompletion error: %v", err)
		return
	}
	logger.Info("ChatCompletion response: %v", resp.Choices[0].Message.Content)
}