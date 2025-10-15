package dashboard

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for emulator
	},
}

// HandleWebSocket handles WebSocket connections
func (d *Dashboard) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		d.log.Error("WebSocket upgrade error: %v", err)
		return
	}

	d.RegisterWebSocketClient(conn)

	defer func() {
		d.UnregisterWebSocketClient(conn)
	}()

	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
