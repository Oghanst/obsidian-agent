package client

import (
	openai "github.com/sashabaranov/go-openai"
)

// DeepSeekClient is a client for interacting with the DeepSeek API.
type DeepSeekClient struct {
	Client *openai.Client
}

func NewDeepSeekClient(apiKey string) *DeepSeekClient {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.deepseek.com/v1"
	client := openai.NewClientWithConfig(config)
	return &DeepSeekClient{
		Client: client,
	}
}