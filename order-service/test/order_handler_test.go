package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"order_microservice/internal/domain"
	"order_microservice/internal/handler"
	jasonwebtoken "order_microservice/internal/middleware/jwt"
	"testing"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockOrderUsecase struct {
	mock.Mock
}

func (m *MockOrderUsecase) UpdateOrderDetails(msg *sarama.ConsumerMessage, ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockOrderUsecase) OrderInit(orderReq *domain.OrderRequest, ctx context.Context) (*domain.Message, error) {
	args := m.Called(orderReq, ctx)
	return args.Get(0).(*domain.Message), args.Error(1)
}

func (m *MockOrderUsecase) OrderRetry(orderReq *domain.RetryOrder, ctx context.Context) (*domain.Message, error) {
	args := m.Called(orderReq, ctx)
	return args.Get(0).(*domain.Message), args.Error(1)
}

func TestOrderInit_Positive(t *testing.T) {
	mockUsecase := new(MockOrderUsecase)
	handler := handler.NewOrderHandler(mockUsecase)

	orderReq := domain.OrderRequest{
		ID:        1, // Jika ini dihasilkan oleh sistem, Anda bisa menyesuaikan nilainya
		OrderType: "new",
		ItemID:    1,
		Amount:    1,
		UserID:    "1",
	}
	orderReqJson, _ := json.Marshal(orderReq)

	expectedMessage := &domain.Message{RespMessage: "Order created successfully"}

	// Gunakan mock.MatchedBy untuk fleksibilitas
	mockUsecase.On("OrderInit", mock.MatchedBy(func(req *domain.OrderRequest) bool {
		return req.OrderType == "new" && req.UserID == "1"
	}), mock.Anything).Return(expectedMessage, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/order", bytes.NewBuffer(orderReqJson))
	c.Request.Header.Set("authorization", jasonwebtoken.CreateMockJWT())

	handler.OrderInit(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "Order created successfully")
}

func TestOrderInit_Negative_InvalidMethod(t *testing.T) {
	mockUsecase := new(MockOrderUsecase)
	handler := handler.NewOrderHandler(mockUsecase)

	router := gin.Default()
	router.POST("/order", handler.OrderInit)

	req, _ := http.NewRequest("GET", "/order", nil)
	req.Header.Set("authorization", jasonwebtoken.CreateMockJWT())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestOrderInit_Negative_InvalidInput(t *testing.T) {
	mockUsecase := new(MockOrderUsecase)
	handler := handler.NewOrderHandler(mockUsecase)

	router := gin.Default()
	router.POST("/order", handler.OrderInit)

	// Invalid input with missing fields
	orderReq := domain.OrderRequest{
		OrderType: "",
	}
	jsonValue, _ := json.Marshal(orderReq)
	req, _ := http.NewRequest("POST", "/order", bytes.NewBuffer(jsonValue))
	req.Header.Set("authorization", jasonwebtoken.CreateMockJWT())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrderEditRetry_Positive(t *testing.T) {
	mockUsecase := new(MockOrderUsecase)
	handler := handler.NewOrderHandler(mockUsecase)

	router := gin.Default()
	router.PATCH("/order/edit-retry", handler.OrderEditRetry)

	orderReq := domain.RetryOrder{
		OrderID:   1,
		OrderType: "1",
		ItemID:    1,
		UserID:    "1",
		Amount:    1,
	}

	// Adjust the mock expectation to match the context type correctly
	mockUsecase.On("OrderRetry", &orderReq, mock.Anything).Return(&domain.Message{OrderID: 1}, nil)

	jsonValue, _ := json.Marshal(orderReq)
	req, _ := http.NewRequest("PATCH", "/order/edit-retry", bytes.NewBuffer(jsonValue))
	req.Header.Set("authorization", jasonwebtoken.CreateMockJWT())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUsecase.AssertExpectations(t)
}

func TestOrderEditRetry_Negative_Method(t *testing.T) {
	mockUsecase := new(MockOrderUsecase)
	handler := handler.NewOrderHandler(mockUsecase)

	router := gin.Default()
	router.PATCH("/order/edit-retry", handler.OrderEditRetry)

	req, _ := http.NewRequest("GET", "/order/edit-retry", nil)
	req.Header.Set("authorization", "valid_token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestOrderEditRetry_Negative_InvalidInput(t *testing.T) {
	mockUsecase := new(MockOrderUsecase)
	handler := handler.NewOrderHandler(mockUsecase)

	router := gin.Default()
	router.PATCH("/order/edit-retry", handler.OrderEditRetry)

	// Invalid input with missing fields
	orderReq := domain.RetryOrder{
		OrderID:   0, // Missing OrderID
		OrderType: "",
	}
	jsonValue, _ := json.Marshal(orderReq)
	req, _ := http.NewRequest("PATCH", "/order/edit-retry", bytes.NewBuffer(jsonValue))
	req.Header.Set("authorization", jasonwebtoken.CreateMockJWT())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
