package transport

// ChatMessage 表示一条对话消息
type ChatMessage struct {
	Role    string `json:"role"`    // "system" | "user" | "assistant"
	Content string `json:"content"` // 文本内容
}

// MsgRequest 前端 -> 后端
type MsgRequest struct {
	Type string `json:"type"` // 消息类型: agent/run, agent/cancel, agent/confirm

	ID string `json:"id,omitempty"` // 前端生成的唯一请求 ID，后端会回传
	Question string `json:"question,omitempty"` // 用户输入（兼容旧字段，不推荐）
	Intent   string `json:"intent,omitempty"`   // 用户意图: qa|write|scaffold|brainstorm 等

	Reserve    int            `json:"reserve,omitempty"`    // 提示后端预留多少 token 空间
	AllowTools bool           `json:"allowTools,omitempty"` // 是否允许调用工具
	Context    map[string]any `json:"context,omitempty"`    // 上下文: 笔记名、光标位置、时间戳等
	Messages   []ChatMessage  `json:"messages,omitempty"`   // 对话历史 (system+user+assistant)

	ConfirmToken string `json:"confirmToken,omitempty"` // 鉴权/确认用 token
}

// MsgResponse 后端 -> 前端
type MsgResponse struct {
	Type string `json:"type"` // 消息类型: agent/preview.delta, agent/full.delta, agent/done, agent/error...

	ID  string `json:"id,omitempty"`  // 对应请求的 ID
	Seq int    `json:"seq,omitempty"` // 流式分片序号，从 1 开始递增

	Text   string         `json:"text,omitempty"`   // 流式输出文本
	Result map[string]any `json:"result,omitempty"` // 工具调用结果

	ConfirmToken string `json:"confirmToken,omitempty"` // 后端要求确认时返回的 token

	ErrorCode string `json:"code,omitempty"`    // 错误码: invalid_token | tool_failed | llm_timeout 等
	ErrorMsg  string `json:"message,omitempty"` // 错误描述
}
