// internal/orchestrator/orchestrator.go
package orchestrator

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/obsidian-agent/internal/llm/client"
	"github.com/obsidian-agent/internal/transport"
)

type Orchestrator struct {
	llm     *client.DeepSeekClient
	mu      sync.Mutex
	cancels map[string]context.CancelFunc
}

func New(llm *client.DeepSeekClient) *Orchestrator {
	return &Orchestrator{
		llm:     llm,
		cancels: make(map[string]context.CancelFunc),
	}
}

func (o *Orchestrator) Cancel(id string) {
	o.mu.Lock()
	if c, ok := o.cancels[id]; ok {
		c()
		delete(o.cancels, id)
	}
	o.mu.Unlock()
}

func (o *Orchestrator) Run(ctx context.Context, req transport.Msg, sink transport.Sender) error {
	// 记录 cancel
	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	o.mu.Lock(); o.cancels[req.ID] = cancel; o.mu.Unlock()
	defer func() { o.Cancel(req.ID) }()

	// 构建 messages：system + 历史 + 本轮 user
	messages := o.llm.BuildMessages(req.Question)

	// 预览策略：首句/首段只发一次
	previewSent := false
	seq := 0
	var previewBuf strings.Builder
	previewDeadline := time.NewTimer(300 * time.Millisecond)
	defer previewDeadline.Stop()

	onDelta := func(delta string) error {
		seq++

		// 累积到 previewBuf，满足条件就发 preview.delta（只发一次）
		if !previewSent {
			previewBuf.WriteString(delta)

			// 条件一：遇到句子终止符
			if i := strings.IndexAny(previewBuf.String(), "。.!?"); i >= 0 {
				text := previewBuf.String()[:i+1]
				_ = sink.Send(transport.Msg{Type: "agent/preview.delta", ID: req.ID, Seq: 1, Text: text})
				previewSent = true
			}
		}
		// 条件二：超时仍未到句末，也先发
		select {
		case <-previewDeadline.C:
			if !previewSent && previewBuf.Len() > 0 {
				_ = sink.Send(transport.Msg{Type: "agent/preview.delta", ID: req.ID, Seq: 1, Text: previewBuf.String()})
				previewSent = true
			}
		default:
		}

		// 全量永远流
		_ = sink.Send(transport.Msg{Type: "agent/full.delta", ID: req.ID, Seq: seq, Text: delta})
		return nil
	}

	// 调用 LLM（流式）
	_, err := o.llm.StreamChatCompletion(ctx, messages, &client.StreamOptions{
		Temperature: 0.3,
		MaxTokens:   800,
	}, onDelta)

	if err != nil {
		_ = sink.Send(transport.Msg{Type: "agent/error", ID: req.ID, ErrorCode: "LLM_ERROR", ErrorMsg: err.Error()})
		return err
	}

	_ = sink.Send(transport.Msg{Type: "agent/done", ID: req.ID})
	return nil
}
