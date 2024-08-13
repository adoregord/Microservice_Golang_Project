package routes

import (
	"order_microservice/internal/handler"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(h handler.OrderHandlerInterface) *gin.Engine {
	// gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	orderRoutes := router.Group("/order")
	{
		orderRoutes.POST("/", h.OrderInit)
	}

	return router
}
