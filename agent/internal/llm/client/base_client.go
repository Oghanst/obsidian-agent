package client

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
)


type BaseClient interface {
	GetClient() *openai.Client
	ChatCompletion(ctx context.Context,request openai.ChatCompletionRequest)(openai.ChatCompletionResponse, error)
}
