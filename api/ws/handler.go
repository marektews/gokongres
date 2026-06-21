package ws

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// upgrader przekształca połączenie HTTP w WebSocket.
// CheckOrigin jest permisywne — endpoint jest publiczny (parytet z /states i /notify/*),
// a ekrany operatorskie działają w zamkniętej sieci (m.in. za proxy dev Vite).
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// HandleWS obsługuje handshake /ws/odprawa i uruchamia pompy klienta.
func HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws: upgrade error: %v", err)
		return
	}

	c := &Client{
		hub:    Default,
		conn:   conn,
		send:   make(chan []byte, sendBuffer),
		topics: make(map[string]bool),
		dead:   make(chan struct{}),
	}

	go c.writePump()
	go c.readPump()

	c.trySend(mustJSON(map[string]any{
		"type": "hello",
		"ts":   time.Now().Format("02.01.2006 15:04:05"),
	}))
}
