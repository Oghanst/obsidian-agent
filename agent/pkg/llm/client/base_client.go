package client

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
)
type BaseClient interface {
	GetClient() *openai.Client
	ChatCompletion(ctx context.Context,request openai.ChatCompletionRequest)(openai.ChatCompletionResponse, error)
	StreamAsk(
		ctx context.Context,
		userPrompt string,
		opts *StreamOptions,
		onDelta StreamHandler,
	) (answer string, err error)
	StreamChatCompletion(
		ctx context.Context,
		messages []openai.ChatCompletionMessage,
		opts *StreamOptions,
		onDelta StreamHandler,
	) (fullText string, err error)
	BuildMessages(userPrompt string, extra ...openai.ChatCompletionMessage) []openai.ChatCompletionMessage
}
// ---- 流式接口 ----
// StreamOptions 用于控制温度、maxTokens 等
type StreamOptions struct {
	Temperature float32
	MaxTokens   int
	Stop        []string
	// 如果你有“预览/全量双轨”的需求，可以上层用两个回调分别处理
}

// StreamHandler 每次收到增量时调用；
// 返回 error 可中止流（例如上层发现用户取消）。
type StreamHandler func(delta string) error

