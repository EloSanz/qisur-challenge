package api

import (
	"net/http"

	"qisur-service/internal/auth"
	"qisur-service/internal/category"
	"qisur-service/internal/product"
	"qisur-service/internal/search"
	"qisur-service/internal/websocket"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "qisur-service/docs" // Swagger docs

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func InitRouter(
	authMiddleware *jwt.GinJWTMiddleware,
	categoryHandler *category.Handler,
	productHandler *product.Handler,
	searchHandler *search.Handler,
	wsHub *websocket.Hub,
) *gin.Engine {
	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})


	apiGroup := r.Group("/api")
	{
		apiGroup.POST("/auth/login", authMiddleware.LoginHandler)
		apiGroup.GET("/search", searchHandler.Search)
	}


	r.GET("/ws", func(c *gin.Context) {
		websocket.ServeWs(wsHub, c)
	})


	publicProducts := apiGroup.Group("/products")
	{
		publicProducts.GET("", productHandler.GetAll)
		publicProducts.GET("/:id", productHandler.GetByID)
		publicProducts.GET("/:id/history", productHandler.GetHistory)
	}

	publicCategories := apiGroup.Group("/categories")
	{
		publicCategories.GET("", categoryHandler.GetAll)
		publicCategories.GET("/:id", categoryHandler.GetByID)
	}

	protected := r.Group("/api")
	protected.Use(authMiddleware.MiddlewareFunc())
	{
		products := protected.Group("/products")
		{
			products.POST("", auth.RequireRole("admin"), productHandler.Create)
			products.PUT("/:id", auth.RequireRole("admin"), productHandler.Update)
			products.DELETE("/:id", auth.RequireRole("admin"), productHandler.Delete)
		}

		categories := protected.Group("/categories")
		{
			categories.POST("", auth.RequireRole("admin"), categoryHandler.Create)
			categories.PUT("/:id", auth.RequireRole("admin"), categoryHandler.Update)
			categories.DELETE("/:id", auth.RequireRole("admin"), categoryHandler.Delete)
		}
	}

	return r
}
