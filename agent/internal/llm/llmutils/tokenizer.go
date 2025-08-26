package llmutils

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	openai "github.com/sashabaranov/go-openai"
	tiktoken "github.com/pkoukk/tiktoken-go"
)

// CountTokens 用于统计单个字符串在指定模型/分词器下的 token 数量。
func CountTokens(model, s string) (int, error) {
	enc, err := tiktoken.EncodingForModel(model)
	if err != nil {
		// 回退到一个广泛使用的基础编码
		enc, err = tiktoken.GetEncoding("cl100k_base")
		if err != nil {
			return 0, err
		}
	}
	return len(enc.Encode(s, nil, nil)), nil
}

// CountMessageTokens 用于估算单条聊天消息的 token 数量。
// 注意：不同模型的严格 token 规则（角色 token、工具调用等）有所不同。
// 我们保留一个小的角色开销以保证安全。
func CountMessageTokens(model string, msg openai.ChatCompletionMessage) (int, error) {
	roleOverhead := 4 // 每条消息的粗略开销（角色/元数据）；如有需要可调整
	n, err := CountTokens(model, msg.Content)
	if err != nil {
		return 0, err
	}
	return n + roleOverhead, nil
}

// CountMessagesTokens 用于统计一组消息的总 token 数量。
func CountMessagesTokens(model string, msgs []openai.ChatCompletionMessage) (int, error) {
	total := 0
	for _, m := range msgs {
		n, err := CountMessageTokens(model, m)
		if err != nil {
			return 0, err
		}
		total += n
	}
	return total, nil
}

// ClipMessagesToTokenLimit 用于在保证 maxPromptTokens 限制下裁剪消息，
// 保留首条 system 消息（如有）和最新的对话内容。
// 同时为模型的回复预留 responseTokensReserve 数量的 token。
func ClipMessagesToTokenLimit(
	ctx context.Context,
	model string,
	msgs []openai.ChatCompletionMessage,
	maxPromptTokens int,
	responseTokensReserve int,
) ([]openai.ChatCompletionMessage, error) {

	if maxPromptTokens <= 0 {
		return nil, errors.New("maxPromptTokens 必须大于 0")
	}
	if responseTokensReserve < 0 {
		responseTokensReserve = 0
	}
	budget := maxPromptTokens - responseTokensReserve
	if budget <= 0 {
		return nil, errors.New("预留后可用的 prompt token 不为正")
	}

	// 1) 提取首条 system 消息（如有）
	var system *openai.ChatCompletionMessage
	rest := make([]openai.ChatCompletionMessage, 0, len(msgs))
	for i := range msgs {
		if msgs[i].Role == openai.ChatMessageRoleSystem && system == nil {
			cp := msgs[i]
			system = &cp
			continue
		}
		rest = append(rest, msgs[i])
	}

	// 2) 从尾部（最新消息）到头部贪婪选取
	selected := make([]openai.ChatCompletionMessage, 0, len(rest)+1)

	total := 0
	if system != nil {
		n, err := CountMessageTokens(model, *system)
		if err != nil {
			return nil, err
		}
		if n > budget {
			// system 消息本身过大 -> 对 system 内容进行硬截断
			trunc := truncateByTokens(model, system.Content, budget-8) // 留出一些余量
			sys := *system
			sys.Content = trunc
			return []openai.ChatCompletionMessage{sys}, nil
		}
		total += n
		selected = append(selected, *system)
	}

	// 从末尾累加
	for i := len(rest) - 1; i >= 0; i-- {
		n, err := CountMessageTokens(model, rest[i])
		if err != nil {
			return nil, err
		}
		if total+n <= budget {
			// 前插以保持后续时间顺序
			selected = append(selected, openai.ChatCompletionMessage{}) // 插入占位
			copy(selected[2:], selected[1:])                           // 右移；如有 system 保持在索引 0
			selected[1] = rest[i]
			total += n
			continue
		}
		// 如果最后一条（最新）消息是 user 或 assistant，尝试截断
		if rest[i].Role == openai.ChatMessageRoleUser || rest[i].Role == openai.ChatMessageRoleAssistant {
			allow := budget - total
			if allow > 16 { // 只在剩余空间足够时才截断
				truncContent := truncateByTokens(model, rest[i].Content, allow-8)
				if strings.TrimSpace(truncContent) != "" {
					msg := rest[i]
					msg.Content = truncContent
					// 插入截断后的消息
					selected = append(selected, openai.ChatCompletionMessage{})
					copy(selected[2:], selected[1:])
					selected[1] = msg
					total = budget // 已用满预算
				}
			}
		}
		break // 已达到预算
	}

	// 保证时间顺序：system（如有）+ 较旧...较新。
	// 上述插入逻辑已保证顺序。
	return selected, nil
}

// truncateByTokens 用于将字符串 s 截断到最多 maxTokens 个 token（近似），
// 如果无法快速获得精确边界，则回退到按字符截断。
func truncateByTokens(model, s string, maxTokens int) string {
	if maxTokens <= 0 || len(s) == 0 {
		return ""
	}
	enc, err := tiktoken.EncodingForModel(model)
	if err != nil {
		enc, _ = tiktoken.GetEncoding("cl100k_base")
	}
	if enc == nil {
		// 回退：粗略按字符截断（非精确 token）
		return roughCutRunes(s, maxTokens*3/2)
	}
	ids := enc.Encode(s, nil, nil)
	if len(ids) <= maxTokens {
		return s
	}
	ids = ids[:maxTokens]
	out := enc.Decode(ids)
	// 防护：如果分词器解码结果不是有效 UTF-8，则回退
	if !utf8.ValidString(out) {
		return roughCutRunes(s, maxTokens*3/2)
	}
	return out + "…"
}

// roughCutRunes 用于按字符（rune）粗略截断字符串 s 到最多 n 个字符。
func roughCutRunes(s string, n int) string {
	if n <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}
