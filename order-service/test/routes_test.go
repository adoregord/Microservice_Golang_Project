package test

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"order_microservice/internal/provider/routes"
	"testing"
)

// MockOrderHandler implements OrderHandlerInterface for testing purposes
type MockOrderHandler struct {
	OrderInitFunc      func(c *gin.Context)
	OrderEditRetryFunc func(c *gin.Context)
}

func (m *MockOrderHandler) OrderInit(c *gin.Context) {
	m.OrderInitFunc(c)
}

func (m *MockOrderHandler) OrderEditRetry(c *gin.Context) {
	m.OrderEditRetryFunc(c)
}

func TestSetupRoutes_OrderInit_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockHandler := &MockOrderHandler{
		OrderInitFunc: func(c *gin.Context) {
			c.JSON(http.StatusCreated, gin.H{"message": "Order placed successfully"})
		},
	}

	router := routes.SetupRoutes(mockHandler)

	w := httptest.NewRecorder()
	reqBody := bytes.NewBufferString(`{"some":"json"}`)
	req, _ := http.NewRequest("POST", "/order/", reqBody)
	req.Header.Set("authorization", "Bearer valid_token")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "Order placed successfully")
}

func TestSetupRoutes_OrderInit_MethodNotAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockHandler := &MockOrderHandler{}

	router := routes.SetupRoutes(mockHandler)

	w := httptest.NewRecorder()
	reqBody := bytes.NewBufferString(`{"some":"json"}`)
	req, _ := http.NewRequest("GET", "/order/", reqBody)
	req.Header.Set("authorization", "Bearer valid_token")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "page not found")
}
