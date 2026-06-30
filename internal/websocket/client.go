package websocket

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"qisur-service/pkg/audit"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for the challenge
	},
}

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("WebSocket error", "error", err)
			}
			break
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func ServeWs(hub *Hub, c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("Failed to set websocket upgrade", "error", err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

// BroadcastEvent sends a structured WSEvent to all connected clients
func (h *Hub) BroadcastEvent(traceID, eventName string, data interface{}) {
	event := WSEvent{
		TraceID: traceID,
		Event:   eventName,
		Data:    data,
	}
	msg, err := json.Marshal(event)
	if err != nil {
		slog.Error("Failed to marshal WS event", "error", err)
		return
	}

	// Emit PUBLISHED_TO_BROKER
	audit.Emit(h.rmq, traceID, "PUBLISHED_TO_BROKER", eventName, "")

	if h.rmq != nil {
		err = h.rmq.Channel.Publish(
			"qisur_events", // exchange
			"",             // routing key (ignored for fanout)
			false,          // mandatory
			false,          // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        msg,
			})
		if err != nil {
			slog.Error("Failed to publish to RabbitMQ", "error", err)
		}
	} else {
		// Fallback to local broadcast if RabbitMQ is not connected
		h.broadcast <- msg
	}
}
