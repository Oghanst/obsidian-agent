package ws

import (
	"net/http"

	"github.com/obsidian-agent/biz/transport"
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("token") == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}
	conn, err := transport.Upgrader.Upgrade(w, r, nil)
	if err != nil { return }
	defer conn.Close()
}