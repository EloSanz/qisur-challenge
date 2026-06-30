package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"qisur-service/internal/bootstrap"
	"qisur-service/internal/database"
	"qisur-service/pkg/rabbitmq"

	"github.com/joho/godotenv"
)

// @title Qisur API
// @version 1.0
// @description REST API and WebSockets for Qisur Challenge products management.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.qisur.com.ar
// @contact.email support@qisur.com.ar

// @BasePath /qisur
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	_ = godotenv.Load()

	db, err := database.InitDB()
	if err != nil {
		slog.Error("Database connection failed", "error", err)
		os.Exit(1)
	}
	slog.Info("Connected to PostgreSQL database successfully")

	rmqUrl := os.Getenv("RABBITMQ_URL")
	var rmqClient *rabbitmq.Client
	if rmqUrl != "" {
		rmqClient, err = rabbitmq.Connect(rmqUrl)
		if err != nil {
			slog.Warn("Could not connect to RabbitMQ, scaling disabled", "error", err)
		} else {
			defer rmqClient.Close()
		}
	}

	router := bootstrap.InitApp(db, rmqClient)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8086"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		slog.Info("Starting Qisur Service", "port", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("Server exiting")
}
