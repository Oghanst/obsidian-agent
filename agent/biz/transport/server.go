package transport

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/obsidian-agent/pkg/logger"
)

type Orchestrator interface {
	Run(ctx context.Context, msg MsgRequest, sender Sender) error
	Cancel(id string)
}

type Sender interface {
	Send(v any) error
}

type WsSender struct {
	c  *websocket.Conn
	mu sync.Mutex
}

func (s *WsSender) Send(v any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.c.WriteJSON(v)
}

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // 仅 127.0.0.1 使用，生产请收紧
}

var wsLogger *logger.Logger
var globalHandlerMap map[string]http.HandlerFunc

func init() {
	var err error
	wsLogger, err = logger.New("logs/ws_server.log")
	if err != nil {
		panic(err)
	}
	globalHandlerMap = InitHandlerMap()
}

func Serve(addr string, orch Orchestrator) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("token") == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}
		conn, err := Upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		sender := &WsSender{c: conn}

		// 心跳
		go func() {
			t := time.NewTicker(15 * time.Second)
			defer t.Stop()
			for range t.C {
				sender.mu.Lock()
				_ = conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(5*time.Second))
				sender.mu.Unlock()
			}
		}()

		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			var msg MsgRequest
			if err := json.Unmarshal(data, &msg); err != nil {
				continue
			}

			switch msg.Type {
			case "agent/run":
				// 为每个 run 开 goroutine
				go func(m MsgRequest) {
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					_ = orch.Run(ctx, m, sender)
				}(msg)
			case "agent/cancel":
				orch.Cancel(msg.ID)
				// 也可以扩展 tools/call 等其它 type
			}
		}
	})

	return http.ListenAndServe(addr, mux)
}

func InitHandlerMap() map[string]http.HandlerFunc {
	handlerMap := make(map[string]http.HandlerFunc)

	// handlerMap["/test"] = wsHandler

	return handlerMap
}

func registerHandlers() {
	for pattern, handler := range globalHandlerMap {
		http.HandleFunc(pattern, handler)
	}
}

func StartWebSocketServer(addr string) {
	registerHandlers()
	wsLogger.Info("WebSocket server starting on %s", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		wsLogger.Error("WebSocket server error: %v", err)
	}
}
