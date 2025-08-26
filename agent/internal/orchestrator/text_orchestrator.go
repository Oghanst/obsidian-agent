package orchestrator

import (
	"context"
	"sync"

	"github.com/obsidian-agent/internal/llm/client"
	"github.com/obsidian-agent/internal/transport"
)

type TextOrchestrator struct {
	llm client.BaseClient
	mu      sync.Mutex
	cancels map[string]context.CancelFunc
}

func BuildTextOrchestrator(llm client.BaseClient) *TextOrchestrator {
	return &TextOrchestrator{
		llm:     llm,
		cancels: make(map[string]context.CancelFunc),
	}
}

func (o *TextOrchestrator) Cancel(id string) {
	o.mu.Lock()
	if c, ok := o.cancels[id]; ok {
		c()
		delete(o.cancels, id)
	}
	o.mu.Unlock()
}

func (o *TextOrchestrator) Run(ctx context.Context, msg transport.Msg, sender transport.Sender) error {
	// 记录 cancel
	ctx, cancel := context.WithCancel(ctx)
	o.mu.Lock()
	o.cancels[msg.ID] = cancel
	o.mu.Unlock()
	defer func() { o.Cancel(msg.ID) }()

	onDelta := func(delta string) error {
		_ = sender.Send(delta)
		return nil
	}

	// 构建 messages：system + 历史 + 本轮 user
	messages := o.llm.BuildMessages(msg.Question)
	// 调用 LLM（流式）
	_, err := o.llm.StreamChatCompletion(ctx, messages, &client.StreamOptions{
		Temperature: 0.3,
		MaxTokens:   800,
	}, onDelta)
	if err != nil {
		_ = sender.Send(transport.Msg{
			ID:    msg.ID,
			Type:  "text/error",
			Text:  err.Error(),
		})
		return err
	}
	// 完成
	_ = sender.Send(transport.Msg{
		ID:   msg.ID,
		Type: "text/done",
	})
	return  nil
}