package client

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/obsidian-agent-cli/internal/constant"
	"github.com/obsidian-agent-cli/internal/proto"
	"github.com/obsidian-agent-cli/internal/utils"
)

type WSClient struct {
	conn *websocket.Conn
}

func NewWSClient(url, token string) (*WSClient, error) {
	wsURL := url + "?token=" + utils.UrlEscape(token)
	d := &websocket.Dialer{HandshakeTimeout: 8 * time.Second}
	c, resp, err := d.Dial(wsURL, nil)
	if err != nil {
		if resp != nil {
			return nil, fmt.Errorf("dial failed: %w (status %s)", err, resp.Status)
		}
		return nil, fmt.Errorf("dial failed: %w", err)
	}
	log.Printf("%s[connected]%s %s\n", constant.COLOR_CYAN, constant.COLOR_RESET, wsURL)
	return &WSClient{conn: c}, nil
}

func (w *WSClient) Close() error {
	if w.conn != nil {
		return w.conn.Close()
	}
	return nil
}

func (w *WSClient) SendJSON(v any) error {
	return w.conn.WriteJSON(v)
}

func (w *WSClient) ReadOne(msg *proto.Msg) error {
	_, data, err := w.conn.ReadMessage()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, msg)
}
