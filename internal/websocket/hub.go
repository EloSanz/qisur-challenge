package websocket

import (
	"encoding/json"
	"log/slog"
	"sync"

	"qisur-service/pkg/audit"
	"qisur-service/pkg/rabbitmq"
)

// WSEvent represents the structure of messages sent over WebSocket
type WSEvent struct {
	TraceID string      `json:"trace_id,omitempty"`
	Event   string      `json:"event"`
	Data    interface{} `json:"data"`
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
	rmq        *rabbitmq.Client
}

func NewHub(rmq *rabbitmq.Client) *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		rmq:        rmq,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			slog.Info("WebSocket client connected")
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			slog.Info("WebSocket client disconnected")
		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- message:
					// Parse the trace ID just to emit DELIVERED_TO_WS if needed, 
					// but doing it for every client might be too much. 
					// We'll emit it once here per broadcast locally.
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) ListenRabbitMQ() {
	if h.rmq == nil {
		slog.Warn("No RabbitMQ client provided to Hub, ListenRabbitMQ aborted")
		return
	}

	q, err := h.rmq.Channel.QueueDeclare(
		"",    // empty name generates a random unique name
		false, // non-durable
		true,  // auto-delete when unused
		true,  // exclusive to this connection
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		slog.Error("Failed to declare exclusive queue", "error", err)
		return
	}

	err = h.rmq.Channel.QueueBind(
		q.Name,                // queue name
		"",                    // routing key
		rabbitmq.ExchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		slog.Error("Failed to bind queue", "error", err)
		return
	}

	msgs, err := h.rmq.Channel.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		slog.Error("Failed to register consumer", "error", err)
		return
	}

	slog.Info("Hub listening for RabbitMQ events on exclusive queue", "queue", q.Name)

	for d := range msgs {
		slog.Info("Received message from RabbitMQ", "body_size", len(d.Body))
		var evt WSEvent
		json.Unmarshal(d.Body, &evt)
		if evt.TraceID != "" {
			audit.Emit(h.rmq, evt.TraceID, "CONSUMED_BY_HUB", evt.Event, "")
		}
		
		h.broadcast <- d.Body
		
		if evt.TraceID != "" {
			audit.Emit(h.rmq, evt.TraceID, "DELIVERED_TO_WS", evt.Event, "")
		}
	}
}
