package proto

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Msg struct {
	Type         string                 `json:"type"`
	ID           string                 `json:"id,omitempty"`
	Seq          int                    `json:"seq,omitempty"`
	Text         string                 `json:"text,omitempty"`
	Question     string                 `json:"question,omitempty"`
	Intent       string                 `json:"intent,omitempty"`
	Reserve      int                    `json:"reserve,omitempty"`
	AllowTools   bool                   `json:"allowTools,omitempty"`
	ConfirmToken string                 `json:"confirmToken,omitempty"`
	Context      map[string]any         `json:"context,omitempty"`
	Messages     []ChatMessage          `json:"messages,omitempty"` // 新增：可选历史
	ErrorCode    string                 `json:"code,omitempty"`
	ErrorMsg     string                 `json:"message,omitempty"`
	Result       map[string]any         `json:"result,omitempty"`
}