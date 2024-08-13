package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	"order_microservice/internal/domain"
	jasonwebtoken "order_microservice/internal/middleware/jwt"
	"order_microservice/internal/usecase"
	"order_microservice/internal/util"
	"time"

	"github.com/gin-gonic/gin"
)

type OrderHandlerInterface interface {
	OrderInit
}
type OrderInit interface{ OrderInit(c *gin.Context) }

type OrderHandler struct {
	Usecase usecase.OrderUsecaseInterface
}

func NewOrderHandler(usecase usecase.OrderUsecaseInterface) OrderHandlerInterface {
	return OrderHandler{
		Usecase: usecase,
	}
}

func (h OrderHandler) OrderInit(c *gin.Context) {
	authHeader := c.GetHeader("authorization")
	c.Writer.Header().Set("Content-Type", "application/json")
	kontek := context.WithValue(c.Request.Context(), "waktu", time.Now())

	var logError error
	var logMessage string
	var logStatus int

	defer func() {
		if logError != nil {
			util.LogFailed(logMessage, c.Request.Method, kontek.Value("waktu").(time.Time), logStatus, logError)
		} else {
			util.LogSuccess(logMessage, c.Request.Method, kontek.Value("waktu").(time.Time), logStatus)
		}
	}()

	// check if the method is post
	if c.Request.Method != "POST" {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"Message": "Method not allowed", "Status": http.StatusMethodNotAllowed})
		logError = errors.New("method not allowed")
		logMessage = "Create Event API Failed"
		logStatus = http.StatusMethodNotAllowed
		return
	}

	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
		return
	}

	claims, err := jasonwebtoken.VerifyJWT(authHeader)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var orderReq domain.OrderRequest
	if err := c.ShouldBindJSON(&orderReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	// validate the input
	if err := validate.Struct(orderReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Message": err.Error(), "Status": http.StatusBadRequest})
		logError = err
		logMessage = "Create Event API Failed"
		logStatus = http.StatusBadRequest
		return
	}
	// get the username of the person from their token
	orderReq.UserID = claims.Username

	log.Println("===============================================================================================")
	log.Printf("Order ID %d for order type '%s' is STARTED\n", orderReq.ID, orderReq.OrderType)
	log.Println("===============================================================================================")

	message, err := h.Usecase.OrderInit(&orderReq, context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order placed successfully", "order": message})
}
