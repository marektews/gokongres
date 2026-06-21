package ws

import "sync"

// Hub utrzymuje subskrypcje klientów pogrupowane po tematach (topic).
// Mapowanie: topic -> zbiór klientów zapisanych na ten temat.
type Hub struct {
	mu     sync.RWMutex
	topics map[string]map[*Client]bool
}

// Default to globalny hub używany przez serwer.
var Default = NewHub()

func NewHub() *Hub {
	return &Hub{topics: make(map[string]map[*Client]bool)}
}

// subscribe zapisuje klienta na temat.
func (h *Hub) subscribe(c *Client, topic string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	subs := h.topics[topic]
	if subs == nil {
		subs = make(map[*Client]bool)
		h.topics[topic] = subs
	}
	subs[c] = true
}

// unsubscribe wypisuje klienta z tematu.
func (h *Hub) unsubscribe(c *Client, topic string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if subs := h.topics[topic]; subs != nil {
		delete(subs, c)
		if len(subs) == 0 {
			delete(h.topics, topic)
		}
	}
}

// remove usuwa klienta ze wszystkich jego tematów (wołane przy rozłączeniu).
func (h *Hub) remove(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for t := range c.topics {
		if subs := h.topics[t]; subs != nil {
			delete(subs, c)
			if len(subs) == 0 {
				delete(h.topics, t)
			}
		}
	}
}

// Publish rozsyła payload do wszystkich klientów zapisanych na temat.
// Wysyłka jest nieblokująca; klient z przepełnionym buforem (np. zawieszony)
// jest ubijany, a sprzątanie następuje w jego pompach.
func (h *Hub) Publish(topic string, payload []byte) {
	h.mu.RLock()
	var slow []*Client
	for c := range h.topics[topic] {
		select {
		case c.send <- payload:
		default:
			slow = append(slow, c)
		}
	}
	h.mu.RUnlock()

	for _, c := range slow {
		c.kill()
	}
}
