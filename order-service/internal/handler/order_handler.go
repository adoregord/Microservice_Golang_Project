package handler

import (
	"context"
	"net/http"
	"order_microservice/internal/domain"
	jasonwebtoken "order_microservice/internal/middleware/jwt"
	"order_microservice/internal/usecase"
	"order_microservice/internal/util"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func OrderHandler(c *gin.Context) {
	authHeader := c.GetHeader("authorization")
	kontek := context.WithValue(c.Request.Context(), "waktu", time.Now())

	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
		return
	}

	claims, err := jasonwebtoken.VerifyJWT(authHeader)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

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

	var orderReq domain.OrderRequest
	if err := c.ShouldBindJSON(&orderReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// get the username of the person from their token
	orderReq.UserId = claims.Username

	log.Info().Msg("===============================================================================================")
	log.Info().Msgf("Transaction ID %s for order type '%s' is STARTED\n", orderReq.TransactionID, orderReq.OrderType)
	log.Info().Msg("===============================================================================================")

	_, _, err = usecase.OrderUsecase(&orderReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order placed successfully", "order": orderReq})
}
