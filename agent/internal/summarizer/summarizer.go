package summarizer

import (
	"context"

	"github.com/obsidian-agent/pkg/llm/client"
	"github.com/sashabaranov/go-openai"
)

type Summarizer struct {
	llmClient     client.BaseClient

	SummaryPrompts map[string]string // 对应 summary、review、judge 等不同场景的 prompt
	Options       map[string]*client.ChatOptions
}
func NewSummarizer(llmClient client.BaseClient) *Summarizer {
	summarizer := &Summarizer{
		llmClient:     llmClient,
		SummaryPrompts: make(map[string]string),
		Options:       make(map[string]*client.ChatOptions),
	}
	defaultPrompt := "请帮我总结以上内容的要点，要求简洁明了，适合快速阅读：\n\n"
	defaultOptions := &client.ChatOptions{
		Temperature: 0.7,
		MaxTokens: 1024,
	}
	summarizer.SetDefaultScene(defaultPrompt, defaultOptions)
	return summarizer
}

func (s *Summarizer) GetDefaultScene() (string, *client.ChatOptions) {
	return s.SummaryPrompts["default"], s.Options["default"]
}

func (s *Summarizer) SetDefaultScene(prompt string, opts *client.ChatOptions) {
	s.SummaryPrompts["default"] = prompt
	s.Options["default"] = opts
}

func (s *Summarizer) SetScene(scene string, prompt string, opts *client.ChatOptions) {
	s.SummaryPrompts[scene] = prompt
	s.Options[scene] = opts
}

func (s *Summarizer) GetScene(scene string) (string, *client.ChatOptions) {
	prompt, opts := s.GetDefaultScene()
	if p, ok := s.SummaryPrompts[scene]; ok {
		prompt = p
	} 
	if o, ok := s.Options[scene]; ok{
		opts = o
	}
	return prompt, opts
}

// Summary 负责总结一段对话
func (s *Summarizer) Summary(ctx context.Context,messages []openai.ChatCompletionMessage) (openai.ChatCompletionResponse, error) {
	prompt, opts := s.GetScene("summary")
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	})
	return  s.llmClient.ChatCompletion(ctx,messages,opts)
}

// Judge 负责根据用户的提问/对话进行判断
func (s *Summarizer) Judge(ctx context.Context,messages []openai.ChatCompletionMessage) (openai.ChatCompletionResponse, error) {
	prompt, opts := s.GetScene("judge")
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	})
	return  s.llmClient.ChatCompletion(ctx,messages,opts)
}