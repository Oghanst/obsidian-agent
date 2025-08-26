package transport

type Msg struct {
	Type        string                 `json:"type"`
	ID          string                 `json:"id,omitempty"`
	Seq         int                    `json:"seq,omitempty"`
	Text        string                 `json:"text,omitempty"`
	Question    string                 `json:"question,omitempty"`
	Intent      string                 `json:"intent,omitempty"`
	Reserve     int                    `json:"reserve,omitempty"`
	AllowTools  bool                   `json:"allowTools,omitempty"`
	ConfirmToken string                `json:"confirmToken,omitempty"`
	Context     map[string]any         `json:"context,omitempty"`
	ErrorCode   string                 `json:"code,omitempty"`
	ErrorMsg    string                 `json:"message,omitempty"`
	Result      map[string]any         `json:"result,omitempty"`
}
