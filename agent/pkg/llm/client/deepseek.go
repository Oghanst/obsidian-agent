package client

import (
	"context"
	"errors"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

const (
	DEFAULT_MODEL              = "deepseek-chat"
	DEFAULT_HISTORY_SIZE_LIMIT = 100_000 // 先当“字符预算”用；后续可换 token 计数
)

type DeepSeekClient struct {
	Client           *openai.Client
	Model            string
	HistorySizeLimit int

	systemPrompt string
}

func NewDeepSeekClient(apiKey string) *DeepSeekClient {
	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = "https://api.deepseek.com/v1"

	return &DeepSeekClient{
		Client:           openai.NewClientWithConfig(cfg),
		Model:            DEFAULT_MODEL,
		HistorySizeLimit: DEFAULT_HISTORY_SIZE_LIMIT,
	}
}

func (d *DeepSeekClient) GetClient() *openai.Client { return d.Client }

func (d *DeepSeekClient) SetSystemPrompt(p string) { d.systemPrompt = p }
func (d *DeepSeekClient) GetSystemPrompt() string  { return d.systemPrompt }

// BuildMessages 将系统 prompt 组装到用户的上下文尾部
func (d *DeepSeekClient) BuildMessages(llmContext []openai.ChatCompletionMessage) []openai.ChatCompletionMessage {
	var msgs []openai.ChatCompletionMessage

	if len(llmContext) > 0 {
		msgs = append(msgs, llmContext...)
	}

	if strings.TrimSpace(d.systemPrompt) != "" {
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: d.systemPrompt,
		})
	}
	return msgs
}

// ---- 非流式 ----
func (d *DeepSeekClient) ChatCompletion(
		ctx context.Context,
		messages []openai.ChatCompletionMessage,
		opts *ChatOptions,
) (response openai.ChatCompletionResponse, err error){
	// 组装请求
	req := openai.ChatCompletionRequest{
		Model:       d.Model,
		Messages:    messages,
		Temperature: opts.Temperature,
		MaxTokens:   opts.MaxTokens,
		Stop:        opts.Stop,
	}
	return d.Client.CreateChatCompletion(ctx, req)
}

// StreamChatCompletion：基于现有 messages 发起流式请求，返回完整结果
// 支持 onDelta 回调实时处理每个增量片段
func (d *DeepSeekClient) StreamChatCompletion(
    ctx context.Context,
    messages []openai.ChatCompletionMessage,
    opts *ChatOptions,
    onDelta StreamHandler,
) (StreamResult, error) {
    var out StreamResult
    if len(messages) == 0 {
        return out, errors.New("messages is empty")
    }
    if opts == nil {
        opts = &ChatOptions{Temperature: 0.3, MaxTokens: 512}
    }

    // 构造请求，开启流式模式
    req := openai.ChatCompletionRequest{
        Model:       d.Model,
        Messages:    messages,
        Temperature: opts.Temperature,
        MaxTokens:   opts.MaxTokens,
        Stop:        opts.Stop,
        Stream:      true,
        // 关键：请求服务端在最后一个 chunk 中包含 usage
        StreamOptions: &openai.StreamOptions{IncludeUsage: true},
    }

    stream, err := d.Client.CreateChatCompletionStream(ctx, req)
    if err != nil {
        return out, err
    }
    defer stream.Close()

    var b strings.Builder

    for {
        // 每次接收一个流式分片
        resp, recvErr := stream.Recv()
        if recvErr != nil {
            // EOF 或上下文取消：返回已收集的内容
            if b.Len() > 0 && (errors.Is(recvErr, context.Canceled) || strings.Contains(recvErr.Error(), "EOF")) {
                out.Text = b.String()
                return out, nil
            }
            // 其他错误：也返回已收集的内容并报错
            out.Text = b.String()
            return out, recvErr
        }

        // 捕获元数据（可能只会在最后一个 chunk 出现）
        if resp.Model != "" {
            out.Model = resp.Model
        }
        if resp.SystemFingerprint != "" {
            out.SystemFingerprint = resp.SystemFingerprint
        }
        if resp.Usage != nil { // IncludeUsage 开启时，最终 chunk 才会带 usage
            out.Usage = resp.Usage
        }

        // 处理每个 choice（通常只有一个）
        for _, ch := range resp.Choices {
            if frag := ch.Delta.Content; frag != "" {
                // 拼接文本
                b.WriteString(frag)
                // 如果上层传了回调，增量片段交给回调
                if onDelta != nil {
                    if cbErr := onDelta(frag); cbErr != nil {
                        out.Text = b.String()
                        return out, cbErr // 上层要求中断
                    }
                }
            }
            // finish_reason 通常只在最后一个分片里出现
            if ch.FinishReason != "" {
                out.FinishReason = ch.FinishReason
            }
        }

        // （可选）保存原始 choices 便于调试
        out.RawChoices = resp.Choices
    }
}
