package backend

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	wsUpgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	wsHub = &hub{clients: map[*websocket.Conn]struct{}{}}
)

type hub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]struct{}
}

func (h *hub) register(conn *websocket.Conn) {
	h.mu.Lock()
	h.clients[conn] = struct{}{}
	h.mu.Unlock()
}

func (h *hub) unregister(conn *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, conn)
	h.mu.Unlock()
}

func Broadcast() {
	wsHub.mu.Lock()
	defer wsHub.mu.Unlock()
	for conn := range wsHub.clients {
		conn.WriteMessage(websocket.TextMessage, []byte("update"))
	}
}

func WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	wsHub.register(conn)
	defer wsHub.unregister(conn)

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
