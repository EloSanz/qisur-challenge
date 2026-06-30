package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"qisur-service/pkg/audit"
)

// AuditTrace maps to the audit_traces table
type AuditTrace struct {
	ID         string    `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TraceID    string    `json:"trace_id"`
	Action     string    `json:"action"`
	EntityType string    `json:"entity_type"`
	EntityID   string    `json:"entity_id"`
	Timestamp  time.Time `json:"timestamp"`
}

func main() {
	slog.Info("Starting Qisur Tracker Service")

	// Connect to Database
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		dbHost, dbUser, dbPass, dbName, dbPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		slog.Error("Failed to connect to tracker DB", "error", err)
		os.Exit(1)
	}

	// Connect to RabbitMQ
	rmqURL := os.Getenv("RABBITMQ_URL")
	if rmqURL == "" {
		rmqURL = "amqp://user:password@localhost:5673/"
	}
	conn, err := amqp.Dial(rmqURL)
	if err != nil {
		slog.Error("Failed to connect to RabbitMQ", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		slog.Error("Failed to open a channel", "error", err)
		os.Exit(1)
	}
	defer ch.Close()

	// Ensure exchange exists
	err = ch.ExchangeDeclare(
		audit.AuditExchangeName,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		slog.Error("Failed to declare exchange", "error", err)
	}

	q, err := ch.QueueDeclare(
		"tracker_audit_queue", // name
		true,                  // durable
		false,                 // delete when unused
		false,                 // exclusive
		false,                 // no-wait
		nil,                   // arguments
	)
	if err != nil {
		slog.Error("Failed to declare queue", "error", err)
	}

	err = ch.QueueBind(
		q.Name,
		"", // routing key
		audit.AuditExchangeName,
		false,
		nil,
	)
	if err != nil {
		slog.Error("Failed to bind queue", "error", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		slog.Error("Failed to register consumer", "error", err)
	}

	// Start consuming in background
	go func() {
		for d := range msgs {
			var msg audit.AuditMessage
			if err := json.Unmarshal(d.Body, &msg); err == nil {
				trace := AuditTrace{
					TraceID:    msg.TraceID,
					Action:     msg.Action,
					EntityType: msg.EntityType,
					EntityID:   msg.EntityID,
					Timestamp:  msg.Timestamp,
				}
				db.Create(&trace)
				slog.Info("Saved trace", "trace_id", msg.TraceID, "action", msg.Action)
			}
		}
	}()

	// Setup API
	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/api/traces", func(c *gin.Context) {
		var traces []AuditTrace
		// Get unique traces
		db.Raw("SELECT DISTINCT ON (trace_id) id, trace_id, action, entity_type, entity_id, timestamp FROM audit_traces ORDER BY trace_id, timestamp DESC LIMIT 50").Scan(&traces)
		c.JSON(http.StatusOK, traces)
	})

	r.GET("/api/traces/:id", func(c *gin.Context) {
		id := c.Param("id")
		var traces []AuditTrace
		db.Where("trace_id = ?", id).Order("timestamp ASC").Find(&traces)
		c.JSON(http.StatusOK, traces)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8087"
	}
	slog.Info("Tracker running on port " + port)
	r.Run(":" + port)
}
