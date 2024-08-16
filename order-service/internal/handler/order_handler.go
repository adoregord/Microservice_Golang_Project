package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	"order_microservice/internal/domain"
	jasonwebtoken "order_microservice/internal/middleware/jwt"
	"order_microservice/internal/usecase"
	utils "order_microservice/internal/util"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type OrderHandlerInterface interface {
	OrderInit
	OrderEditRetry
}
type OrderInit interface{ OrderInit(c *gin.Context) }
type OrderEditRetry interface{ OrderEditRetry(c *gin.Context) }

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
			utils.LogFailed(logMessage, c.Request.Method, kontek.Value("waktu").(time.Time), logStatus, logError)
		} else {
			utils.LogSuccess(logMessage, c.Request.Method, kontek.Value("waktu").(time.Time), logStatus)
		}
	}()

	// check if the method is post
	if c.Request.Method != "POST" {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"Message": "Method not allowed", "Status": http.StatusMethodNotAllowed})
		logError = errors.New("method not allowed")
		logMessage = "Create Order API Failed"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// validate the input
	if err := validate.Struct(orderReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Message": err.Error(), "Status": http.StatusBadRequest})
		logError = err
		logMessage = "Create Order API Failed"
		logStatus = http.StatusBadRequest
		return
	}
	// get the username of the person from their token
	userID := strconv.Itoa(claims.UserID)
	orderReq.UserID = userID
	log.Println("===============================================================================================")
	log.Printf("Order ID %d for order type '%s' is STARTED\n", orderReq.ID, orderReq.OrderType)
	log.Println("===============================================================================================")

	message, err := h.Usecase.OrderInit(&orderReq, context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Order placed successfully", "order": message})
}

func (h OrderHandler) OrderEditRetry(c *gin.Context) {
	authHeader := c.GetHeader("authorization")
	c.Writer.Header().Set("Content-Type", "application/json")
	kontek := context.WithValue(c.Request.Context(), "waktu", time.Now())

	var logError error
	var logMessage string
	var logStatus int

	defer func() {
		if logError != nil {
			utils.LogFailed(logMessage, c.Request.Method, kontek.Value("waktu").(time.Time), logStatus, logError)
		} else {
			utils.LogSuccess(logMessage, c.Request.Method, kontek.Value("waktu").(time.Time), logStatus)
		}
	}()

	// check if the method is post
	if c.Request.Method != "PATCH" {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"Message": "Method not allowed", "Status": http.StatusMethodNotAllowed})
		logError = errors.New("method not allowed")
		logMessage = "Edit and Retry Order API Failed"
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

	var orderReq domain.RetryOrder
	if err := c.ShouldBindJSON(&orderReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// validate the input
	if err := validate.Struct(orderReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Message": err.Error(), "Status": http.StatusBadRequest})
		logError = err
		logMessage = "Edit and Retry API Failed"
		logStatus = http.StatusBadRequest
		return
	}
	// get the username of the person from their token
	userID := strconv.Itoa(claims.UserID)
	orderReq.UserID = userID
	log.Println("===============================================================================================")
	log.Printf("Order ID %d for order type '%s' is STARTED\n", orderReq.OrderID, orderReq.OrderType)
	log.Println("===============================================================================================")

	message, err := h.Usecase.OrderRetry(&orderReq, context.Background())
	if err != nil {
		if err.Error() == "theres no order for this id or ure not allowed to retry" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order placed successfully", "order": message})
}
