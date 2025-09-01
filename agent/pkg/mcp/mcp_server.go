package mcp

import (
	"context"
	"errors"
	"fmt"
	"sync"

	// 可选：使用 JSON Schema 校验库（需要时再取消注释其中一个）
	// "github.com/santhosh-tekuri/jsonschema/v5"
	// "github.com/xeipuuv/gojsonschema"
)

// MCPServer 维护一个工具注册表，对外提供“列出工具/调用工具”的能力。
type MCPServer struct {
	mu    sync.RWMutex
	tools map[string]*toolEntry
}

// NewMCPServer 创建一个空的工具注册表。
func NewMCPServer() *MCPServer {
	return &MCPServer{tools: make(map[string]*toolEntry)}
}

// RegisterTool 注册一个工具（说明书 + 执行函数）。
// 你可以选择：
//   1) 立即编译 schema（把编译逻辑放在这里）；
//   2) 懒编译（首次调用再编译，下方 CallTool 中会触发 compileSchemasOnce）。
func (s *MCPServer) RegisterTool(def *ToolDef, handler ToolHandler) error {
	if def == nil || handler == nil {
		return errors.New("tool definition and handler must not be nil")
	}
	if def.Name == "" {
		return errors.New("tool name must not be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.tools[def.Name]; exists {
		return fmt.Errorf("tool %q already registered", def.Name)
	}
	s.tools[def.Name] = &toolEntry{
		def:     def,
		handler: handler,
	}
	return nil
}

// ListRegisteredTools 返回对外的工具说明书列表（用于 tools/list），不暴露内部 handler/编译器等实现细节。
func (s *MCPServer) ListRegisteredTools() []*ToolDef {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*ToolDef, 0, len(s.tools))
	for _, te := range s.tools {
		out = append(out, te.def)
	}
	return out
}

// CallTool 根据名称执行工具：可选参数校验 → 执行 → 可选返回校验。
// - 若工具不存在：返回协议级错误（error）；
// - 若业务执行失败：以 IsError=true 的成功响应返回（JSON-RPC 200）。
func (s *MCPServer) CallTool(ctx context.Context, name string, args map[string]any) (ToolCallResult, error) {
	s.mu.RLock()
	te, ok := s.tools[name]
	s.mu.RUnlock()
	if !ok {
		return ToolCallResult{}, fmt.Errorf("unknown tool: %s", name)
	}

	// 懒编译 schema（需要时再编译；如果你选择“注册时编译”，可删除这段）
	if err := te.compileSchemasOnce(); err != nil {
		// 编译失败属于服务端配置问题，建议作为协议级错误抛出
		return ToolCallResult{}, fmt.Errorf("schema compile failed: %w", err)
	}

	// ---- 可选：对入参进行 JSON-Schema 校验 ----
	/*
		if te.inSchema != nil {
			raw, _ := json.Marshal(args)
			if err := te.inSchema.Validate(bytes.NewReader(raw)); err != nil {
				// 入参校验失败：通常作为“业务错误”返回，以便客户端收到 200 + isError=true + 文本提示
				return ToolCallResult{IsError: true, ErrorMessage: "input validation failed: " + err.Error()}, nil
			}
		}
	*/

	// 执行工具逻辑
	res, err := te.handler(ctx, args)
	if err != nil {
		// 业务错误：按 MCP 惯例，以成功响应 + isError=true 形式返回
		return ToolCallResult{
			IsError:          true,
			ErrorMessage:     err.Error(),
			Content:          res.Content,
			StructuredContent: res.StructuredContent,
		}, nil
	}

	// ---- 可选：对结构化返回进行 JSON-Schema 校验 ----
	/*
		if te.outSchema != nil && res.StructuredContent != nil {
			raw, _ := json.Marshal(res.StructuredContent)
			if err := te.outSchema.Validate(bytes.NewReader(raw)); err != nil {
				return ToolCallResult{IsError: true, ErrorMessage: "output validation failed: " + err.Error()}, nil
			}
		}
	*/

	return res, nil
}

// compileSchemasOnce 在需要时编译一次 schema，并缓存结果。
// 若你不需要 schema 校验，可直接让该函数空实现（返回 nil）。
func (te *toolEntry) compileSchemasOnce() error {
	te.compileOnce.Do(func() {
		// 在这里执行一次性编译，并把结果放到 te.inSchema / te.outSchema。
		// 例如使用 santhosh-tekuri/jsonschema：
		/*
			compiler := jsonschema.NewCompiler()
			if len(te.def.InputSchema) > 0 {
				if err := compiler.AddResource("input.json", bytes.NewReader(te.def.InputSchema)); err != nil {
					te.compileErr = fmt.Errorf("add input schema: %w", err)
					return
				}
			}
			inSch, err := compiler.Compile("input.json")
			if err != nil {
				te.compileErr = fmt.Errorf("compile input schema: %w", err)
				return
			}
			te.inSchema = inSch

			if len(te.def.OutputSchema) > 0 {
				if err := compiler.AddResource("output.json", bytes.NewReader(te.def.OutputSchema)); err != nil {
					te.compileErr = fmt.Errorf("add output schema: %w", err)
					return
				}
				outSch, err := compiler.Compile("output.json")
				if err != nil {
					te.compileErr = fmt.Errorf("compile output schema: %w", err)
					return
				}
				te.outSchema = outSch
			}
		*/
		// 如果你当前不想做校验，这里什么都不做即可（te.compileErr 默认为 nil）
	})
	return te.compileErr
}