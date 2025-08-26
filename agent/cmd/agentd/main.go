package main

import (
	"context"

	"github.com/obsidian-agent/internal/llm/client"
	"github.com/obsidian-agent/internal/logger"
	"github.com/obsidian-agent/internal/orchestrator"
	"github.com/obsidian-agent/internal/property"
	"github.com/obsidian-agent/internal/transport"
	"github.com/sashabaranov/go-openai"
)

var mainLogger *logger.Logger

func main(){
	property.LoadConfig("/Users/jianghaojun/Projects/obsidian-agent/agent/config/config.json")
	config := property.GetConfig()
	agentLogger, err := logger.New(config.LogDir + "/agent.log")
	if err != nil {
		panic(err)
	}
	mainLogger = agentLogger
	TestWsServer()
}

func TestWsServer() {
	config := property.GetConfig()
	llm := client.NewDeepSeekClient(config.Apikey)
	llm.SetSystemPrompt(`You are an Obsidian writing companion. Be concise, helpful.`)
	orch := orchestrator.New(llm)
	mainLogger.Info("Starting WebSocket server on %s", config.ServerAddr)
	if err := transport.Serve(config.ServerAddr, orch); err != nil {
		mainLogger.Error("Failed to start WebSocket server: %v", err)
	}
	select {}
}


func TestDeepSeekClient() {
	config := property.GetConfig()
	mainLogger.Info("Agent started with log directory: %s", config.LogDir)
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
		mainLogger.Error("ChatCompletion error: %v", err)
		return
	}
	mainLogger.Info("ChatCompletion response: %v", resp.Choices[0].Message.Content)
}