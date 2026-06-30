package rabbitmq

import (
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

const ExchangeName = "qisur_events"

type Client struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

func Connect(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	// Declare fanout exchange
	err = ch.ExchangeDeclare(
		ExchangeName, // name
		"fanout",     // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	slog.Info("Successfully connected to RabbitMQ and declared exchange", "exchange", ExchangeName)

	return &Client{
		Conn:    conn,
		Channel: ch,
	}, nil
}

func (c *Client) Close() {
	if c.Channel != nil {
		if err := c.Channel.Close(); err != nil {
			slog.Error("Failed to close RabbitMQ channel", "error", err)
		}
	}
	if c.Conn != nil {
		if err := c.Conn.Close(); err != nil {
			slog.Error("Failed to close RabbitMQ connection", "error", err)
		}
	}
}
