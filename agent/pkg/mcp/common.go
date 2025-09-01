package mcp

import (
	"context"
	"encoding/json"
	"sync"

    // 可选：使用 JSON Schema 校验库（需要时再取消注释其中一个）
	// "github.com/santhosh-tekuri/jsonschema/v5"
	// "github.com/xeipuuv/gojsonschema"
)

// ---- 对外的工具“说明书”定义 ----

// ToolDef 表示一个工具的元数据和入/出参的 JSON-Schema（原始 JSON 文本）。
// 注意：这里不用承载任何“可执行”内容，仅用于对外暴露（如 tools/list）或持久化。
type ToolDef struct {
	Name         string          `json:"name"`
	Title        string          `json:"title,omitempty"`
	Description  string          `json:"description,omitempty"`
	InputSchema  json.RawMessage `json:"inputSchema"`
	OutputSchema json.RawMessage `json:"outputSchema,omitempty"`
}

// ToolCallRequest 表示一次工具调用的请求体（由客户端发起）。
type ToolCallRequest struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// ContentPart 是工具返回的人类可读内容（最常用是 text）。
type ContentPart struct {
	Type string `json:"type"`           // 例如 "text"；未来可扩展图片、音频等
	Text string `json:"text,omitempty"` // 当 type=text 时的正文
}

// ToolCallResult 是工具调用的标准返回。
// - Content：人类可读的内容（必须存在，哪怕只是一行文本）；
// - StructuredContent：结构化结果（可选；若提供 outputSchema，建议返回并对齐）；
// - IsError / ErrorMessage：业务错误（注意：协议级错误请走 JSON-RPC 的 error）。
type ToolCallResult struct {
	Content           []ContentPart `json:"content"`
	StructuredContent any           `json:"structuredContent,omitempty"`
	IsError           bool          `json:"isError,omitempty"`
	ErrorMessage      string        `json:"errorMessage,omitempty"`
}

// ---- 可执行工具的绑定 ----

// ToolHandler 是与某个 ToolDef 绑定的可执行函数。
// 入参为已反序列化的 arguments，返回 ToolCallResult 或错误。
// 约定：返回 error 表示“业务执行报错”，上层会转成 IsError=true 的成功响应；
// 若要返回协议级错误（如方法不存在），请在更上层（RPC 层）处理。
type ToolHandler func(ctx context.Context, args map[string]any) (ToolCallResult, error)

// toolEntry 代表一个已注册的工具，包含：说明书、可执行函数、（可选）编译后的 schema。
// 把编译后的 schema 放在这里有三点好处：性能（编译一次复用）、并发安全、避免对外暴露内部实现。
type toolEntry struct {
	def     *ToolDef
	handler ToolHandler

	// ---- 可选：编译后的 JSON-Schema（取消注释并引入库即可）----
	// inSchema  *jsonschema.Schema
	// outSchema *jsonschema.Schema

	// 懒编译控制（只在第一次需要时编译，避免注册时阻塞）
	compileOnce sync.Once
	compileErr  error
}
