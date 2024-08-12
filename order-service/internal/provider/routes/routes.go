package routes

import (
	"order_microservice/internal/handler"

	"github.com/gin-gonic/gin"
)

func SetupRoutes() *gin.Engine {
	// gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	orderRoutes := router.Group("/order")
	{
		orderRoutes.POST("/", handler.OrderHandler)
	}

	return router
}
