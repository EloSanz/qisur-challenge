package bootstrap

import (
	"qisur-service/internal/api"
	"qisur-service/internal/auth"
	"qisur-service/internal/category"
	"qisur-service/internal/product"
	"qisur-service/internal/search"
	"qisur-service/internal/websocket"

	"qisur-service/pkg/rabbitmq"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log/slog"
	"os"
)

// InitApp assembles the dependencies and returns the Gin router
func InitApp(db *gorm.DB, rmq *rabbitmq.Client) *gin.Engine {

	hub := websocket.NewHub(rmq)
	go hub.Run()
	if rmq != nil {
		go hub.ListenRabbitMQ()
	}


	categoryRepo := category.NewRepository(db)
	productRepo := product.NewRepository(db)

	categorySvc := category.NewService(categoryRepo, hub, rmq)
	productSvc := product.NewService(productRepo, hub, rmq)

	authMiddleware, err := auth.SetupJWTMiddleware()
	if err != nil {
		slog.Error("Failed to init JWT middleware", "error", err)
		os.Exit(1)
	}

	categoryHandler := category.NewHandler(categorySvc)
	productHandler := product.NewHandler(productSvc)
	searchHandler := search.NewHandler(db)


	router := api.InitRouter(
		authMiddleware,
		categoryHandler,
		productHandler,
		searchHandler,
		hub,
	)

	return router
}
