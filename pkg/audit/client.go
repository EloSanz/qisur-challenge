package audit

import (
	"encoding/json"
	"log/slog"
	"time"

	"qisur-service/pkg/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
)

const AuditExchangeName = "audit_events"

type AuditMessage struct {
	TraceID    string    `json:"trace_id"`
	Action     string    `json:"action"`
	EntityType string    `json:"entity_type"`
	EntityID   string    `json:"entity_id"`
	Timestamp  time.Time `json:"timestamp"`
}

// Emit sends an audit event to RabbitMQ asynchronously
func Emit(rmq *rabbitmq.Client, traceID, action, entityType, entityID string) {
	if rmq == nil {
		return
	}

	msg := AuditMessage{
		TraceID:    traceID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Timestamp:  time.Now(),
	}

	body, err := json.Marshal(msg)
	if err != nil {
		slog.Error("Failed to marshal audit message", "error", err)
		return
	}

	// Make sure the exchange exists
	err = rmq.Channel.ExchangeDeclare(
		AuditExchangeName,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		slog.Error("Failed to declare audit exchange", "error", err)
		return
	}

	err = rmq.Channel.Publish(
		AuditExchangeName,
		"", // routing key
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		slog.Error("Failed to publish audit event", "error", err)
	}
}
