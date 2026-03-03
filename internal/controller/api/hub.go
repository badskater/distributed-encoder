package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	wsWriteWait  = 10 * time.Second
	wsPongWait   = 60 * time.Second
	wsPingPeriod = (wsPongWait * 9) / 10
	wsMaxMsgSize = 512
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // CORS handled by middleware
}

// HubEvent is a message broadcast to all connected WebSocket clients.
type HubEvent struct {
	Type    string `json:"type"`    // e.g. "job.progress", "task.completed"
	Payload any    `json:"payload"` // arbitrary JSON
}

// Hub manages WebSocket connections and broadcasts events.
type Hub struct {
	clients   map[*wsClient]struct{}
	mu        sync.Mutex
	broadcast chan HubEvent
	logger    *slog.Logger
}

// wsClient wraps a single WebSocket connection.
type wsClient struct {
	conn *websocket.Conn
	send chan HubEvent
}

// NewHub creates a Hub ready to accept connections and broadcast events.
func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		clients:   make(map[*wsClient]struct{}),
		broadcast: make(chan HubEvent, 256),
		logger:    logger,
	}
}

// Run starts the broadcast loop. It blocks until ctx is cancelled.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// Drain remaining events so senders don't block.
			for {
				select {
				case <-h.broadcast:
				default:
					return
				}
			}
		case evt := <-h.broadcast:
			h.mu.Lock()
			for c := range h.clients {
				select {
				case c.send <- evt:
				default:
					// Slow client — drop and disconnect.
					close(c.send)
					delete(h.clients, c)
				}
			}
			h.mu.Unlock()
		}
	}
}

// Publish enqueues an event for broadcast. Non-blocking; drops if the
// broadcast channel is full.
func (h *Hub) Publish(e HubEvent) {
	select {
	case h.broadcast <- e:
	default:
	}
}

// ServeWS upgrades an HTTP connection to WebSocket and registers the client.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("ws upgrade", "err", err)
		return
	}

	client := &wsClient{
		conn: conn,
		send: make(chan HubEvent, 64),
	}

	h.mu.Lock()
	h.clients[client] = struct{}{}
	h.mu.Unlock()

	go h.writePump(client)
	go h.readPump(client)
}

// writePump reads events from the client's send channel and writes them to
// the WebSocket connection. It also sends periodic pings.
func (h *Hub) writePump(c *wsClient) {
	ticker := time.NewTicker(wsPingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case evt, ok := <-c.send:
			if !ok {
				// Channel was closed — send close frame.
				_ = c.conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}
			_ = c.conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
			data, err := json.Marshal(evt)
			if err != nil {
				h.logger.Error("ws marshal", "err", err)
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump reads from the WebSocket connection to handle pongs and detect
// disconnects. Incoming messages are discarded.
func (h *Hub) readPump(c *wsClient) {
	defer func() {
		h.mu.Lock()
		delete(h.clients, c)
		h.mu.Unlock()
		c.conn.Close()
	}()

	c.conn.SetReadLimit(wsMaxMsgSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(wsPongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(wsPongWait))
		return nil
	})

	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}
