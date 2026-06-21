package ws

import (
	"encoding/json"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second // limit czasu na pojedynczy zapis
	pongWait       = 60 * time.Second // brak pong w tym czasie → rozłączenie
	pingPeriod     = 30 * time.Second // okres wysyłania ping (musi być < pongWait)
	maxMessageSize = 4096             // limit rozmiaru ramki od klienta
	sendBuffer     = 32               // bufor kanału wysyłkowego klienta
)

// Client reprezentuje pojedyncze połączenie WebSocket.
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	topics   map[string]bool // tematy, na które klient jest zapisany (dotyka tylko readPump)
	dead     chan struct{}
	killOnce sync.Once
}

// inbound to ramka przychodząca od klienta.
type inbound struct {
	Type   string   `json:"type"`
	Topics []string `json:"topics,omitempty"`
}

// kill sygnalizuje zamknięcie połączenia (idempotentne).
func (c *Client) kill() {
	c.killOnce.Do(func() { close(c.dead) })
}

// trySend wysyła payload nieblokująco; przepełnienie bufora ubija klienta.
func (c *Client) trySend(payload []byte) {
	select {
	case c.send <- payload:
	default:
		c.kill()
	}
}

// readPump czyta ramki od klienta i obsługuje subscribe/unsubscribe/ping.
func (c *Client) readPump() {
	defer func() {
		c.hub.remove(c)
		c.kill()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		c.handleMessage(raw)
	}
}

// writePump drenuje kanał wysyłkowy i utrzymuje połączenie pingami.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case msg := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.dead:
			return
		}
	}
}

func (c *Client) handleMessage(raw []byte) {
	var msg inbound
	if err := json.Unmarshal(raw, &msg); err != nil {
		c.trySend(mustJSON(map[string]string{"type": "error", "message": "invalid frame"}))
		return
	}

	switch msg.Type {
	case "subscribe":
		for _, t := range c.expandTopics(msg.Topics) {
			if !c.topics[t] {
				c.topics[t] = true
				c.hub.subscribe(c, t)
			}
		}
		c.trySend(mustJSON(map[string]any{"type": "subscribed", "topics": c.topicList()}))

	case "unsubscribe":
		for _, t := range c.expandTopics(msg.Topics) {
			if c.topics[t] {
				delete(c.topics, t)
				c.hub.unsubscribe(c, t)
			}
		}
		c.trySend(mustJSON(map[string]any{"type": "subscribed", "topics": c.topicList()}))

	case "ping":
		c.trySend(mustJSON(map[string]string{"type": "pong"}))

	default:
		c.trySend(mustJSON(map[string]string{"type": "error", "message": "unknown type"}))
	}
}

// expandTopics rozwija tematy buffer:<name> na zbiór sector:<sid> danego terminala.
// Pozostałe tematy są przepuszczane bez zmian.
func (c *Client) expandTopics(topics []string) []string {
	out := make([]string, 0, len(topics))
	for _, t := range topics {
		if name, ok := strings.CutPrefix(t, "buffer:"); ok {
			for _, sid := range sectorIDsOfTerminal(name) {
				out = append(out, "sector:"+sid)
			}
			continue
		}
		out = append(out, t)
	}
	return out
}

// topicList zwraca posortowaną listę aktywnych tematów klienta.
func (c *Client) topicList() []string {
	out := make([]string, 0, len(c.topics))
	for t := range c.topics {
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}
