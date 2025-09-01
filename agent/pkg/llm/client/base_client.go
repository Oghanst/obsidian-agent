package client

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
)

type BaseClient interface {
	GetClient() *openai.Client

	// ChatCompletion 基于用户输入和上下文生成回复
	ChatCompletion(
		ctx context.Context,
		messages []openai.ChatCompletionMessage,
		opts *ChatOptions,
	) (response openai.ChatCompletionResponse, err error)

	StreamChatCompletion(
		ctx context.Context,
		messages []openai.ChatCompletionMessage,
		opts *ChatOptions,
		onDelta StreamHandler,
	) (fullText StreamResult, err error)

	// BuildMessages 组装消息
	BuildMessages(llmContext []openai.ChatCompletionMessage) []openai.ChatCompletionMessage 
}

// ChatOptions 用于控制温度、maxTokens 等
type ChatOptions struct {
	Temperature float32
	MaxTokens   int
	Stop        []string
}

// StreamResult 用于保存流式结果，避免丢失元数据
type StreamResult struct {
    Text              string                        // 模型生成的完整文本
    Model             string                        // 使用的模型名称
    FinishReason      openai.FinishReason           // 结束原因（stop/length/...）
    SystemFingerprint string                        // 模型快照指纹，便于复现
    Usage             *openai.Usage                 // Token 使用情况（输入/输出/总数等）
    RawChoices        []openai.ChatCompletionStreamChoice // 原始返回的分片，便于调试
    Headers           map[string]string             // 可选：HTTP 响应头
}
// StreamHandler 每次收到增量时调用；
// 返回 error 可中止流（例如上层发现用户取消）。
type StreamHandler func(delta string) error
