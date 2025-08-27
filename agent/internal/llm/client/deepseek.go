package client

import (
	"context"
	"errors"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

const (
	DEFAULT_MODEL            = "deepseek-chat"
	DEFAULT_HISTORY_SIZE_LIMIT = 100_000 // 先当“字符预算”用；后续可换 token 计数
)

type DeepSeekClient struct {
	Client           *openai.Client
	Model            string
	HistorySizeLimit int

	systemPrompt string
	messagePool  []openai.ChatCompletionMessage // user/assistant 的缓冲池
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

func (d *DeepSeekClient) SetSystemPrompt(p string)  { d.systemPrompt = p }
func (d *DeepSeekClient) GetSystemPrompt() string   { return d.systemPrompt }

// ---- 会话管理 ----

func (d *DeepSeekClient) ResetMessages() {
	d.messagePool = nil
}

func (d *DeepSeekClient) AppendUser(content string) {
	d.append(openai.ChatMessageRoleUser, content)
}

func (d *DeepSeekClient) AppendAssistant(content string) {
	d.append(openai.ChatMessageRoleAssistant, content)
}

func (d *DeepSeekClient) append(role, content string) {
	msg := openai.ChatCompletionMessage{Role: role, Content: content}
	d.messagePool = append(d.messagePool, msg)

}

// BuildMessages 组装 system + 历史 + 本轮用户输入
func (d *DeepSeekClient) BuildMessages(userPrompt string, extra ...openai.ChatCompletionMessage) []openai.ChatCompletionMessage {
	var msgs []openai.ChatCompletionMessage
	if strings.TrimSpace(d.systemPrompt) != "" {
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: d.systemPrompt,
		})
	}
	msgs = append(msgs, d.messagePool...)
	if userPrompt != "" {
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: userPrompt,
		})
	}
	if len(extra) > 0 {
		msgs = append(msgs, extra...)
	}
	return msgs
}

// ---- 非流式（保留原来的） ----

func (d *DeepSeekClient) ChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	return d.Client.CreateChatCompletion(ctx, req)
}



// StreamChatCompletion：给“现成 messages”直接流式
func (d *DeepSeekClient) StreamChatCompletion(
	ctx context.Context,
	messages []openai.ChatCompletionMessage,
	opts *StreamOptions,
	onDelta StreamHandler,
) (fullText string, err error) {

	if len(messages) == 0 {
		return "", errors.New("messages is empty")
	}
	if opts == nil { opts = &StreamOptions{Temperature: 0.3, MaxTokens: 512} }

	req := openai.ChatCompletionRequest{
		Model:       d.Model,
		Messages:    messages,
		Temperature: opts.Temperature,
		MaxTokens:   opts.MaxTokens,
		Stop:        opts.Stop,
		Stream:      true, // 关键
	}

	stream, err := d.Client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return "", err
	}
	defer stream.Close()

	var b strings.Builder
	for {
		resp, recvErr := stream.Recv()
		if recvErr != nil {
			// io.EOF 或 context 取消都会在这里返回
			if respErr := stream.Close(); respErr != nil {
				// 忽略 close 错误
			}
			if b.Len() > 0 && (errors.Is(recvErr, context.Canceled) || strings.Contains(recvErr.Error(), "EOF")) {
				// 半途取消/EOF 也返回已累计文本
				return b.String(), nil
			}
			return b.String(), recvErr
		}
		// DeepSeek / OpenAI 兼容：取 delta 内容
		for _, ch := range resp.Choices {
			frag := ch.Delta.Content
			if frag == "" {
				continue
			}
			b.WriteString(frag)
			if onDelta != nil {
				if cbErr := onDelta(frag); cbErr != nil {
					// 上层要求中断
					return b.String(), cbErr
				}
			}
		}
	}
}

// StreamAsk：给“单轮 user 提问”用，同时自动维护历史
func (d *DeepSeekClient) StreamAsk(
	ctx context.Context,
	userPrompt string,
	opts *StreamOptions,
	onDelta StreamHandler,
) (answer string, err error) {

	msgs := d.BuildMessages(userPrompt)
	ans, err := d.StreamChatCompletion(ctx, msgs, opts, onDelta)
	if err != nil {
		return "", err
	}
	// 回写到历史
	d.AppendUser(userPrompt)
	d.AppendAssistant(ans)
	return ans, nil
}

// 示例：带“超时”的 ctx
func WithTimeout(ctx context.Context, dur time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, dur)
}
