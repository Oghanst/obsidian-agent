package ws

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/obsidian-agent/pkg/logger"
	"github.com/obsidian-agent/biz/transport"
)

var testLogger *logger.Logger

func init() {
	testLogger, _ = logger.New("logs/ws/test.log")
}

func TestHandler(writer http.ResponseWriter, request *http.Request) {
	conn, err := transport.Upgrader.Upgrade(writer, request, nil)
	if err != nil {
		http.Error(writer, "Failed to upgrade to WebSocket", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	clientIP := request.RemoteAddr
	testLogger.Info("New WebSocket connection from %s", clientIP)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			testLogger.Error("Read error from %s: %v", clientIP, err)
			break
		}
		testLogger.Info("Received message from %s: %s", clientIP, string(message))

		err = conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			testLogger.Error("Write error to %s: %v", clientIP, err)
			break
		}
	}
}
