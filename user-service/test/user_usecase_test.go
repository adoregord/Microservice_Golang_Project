package test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"user_microservice/internal/domain"
	"user_microservice/internal/usecase"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockHTTPClient is a mock implementation of the HTTP client
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Post(url string) (*http.Response, error) {
	args := m.Called(url)
	return args.Get(0).(*http.Response), args.Error(1)
}

// Test for UserUsecase.SendMessage
func TestUserUsecase_SendMessage(t *testing.T) {
	incomingMsg := &sarama.ConsumerMessage{
		Value: []byte(`{"OrderID":1,"OrderType":"user_validation","UserId":"testuser","ItemID":1,"Amount":1000}`),
		Key:   []byte("1"),
	}

	expectedMessage := domain.Message{
		OrderID:   1,
		OrderType: "user_validation",
		UserId:    "testuser",
		ItemID:    1,
		Amount:    1000,
	}

	t.Run("success case - user validation", func(t *testing.T) {
		mockProducer := new(MockProducer)
		uc := usecase.NewUserUsecase(mockProducer)

		// Mock HTTP response
		responseStruct := struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		}{
			Status:  "OK",
			Message: "User is valid",
		}

		responseBody, _ := json.Marshal(responseStruct)
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(responseBody)
		}))
		defer mockServer.Close()

		uc.SetHTTPClient(mockServer.Client())

		expectedMessage.RespCode = 200
		expectedMessage.RespStatus = "OK"
		expectedMessage.RespMessage = "SUCCESS: user is verified"

		mockProducer.On("SendMessage", mock.Anything).Return(int32(0), int64(0), nil)

		err := uc.SendMessage(context.Background(), incomingMsg)
		assert.NoError(t, err)
		mockProducer.AssertExpectations(t)
	})

	t.Run("failure case - user not found", func(t *testing.T) {
		mockProducer := new(MockProducer)
		uc := usecase.NewUserUsecase(mockProducer)

		// Mock HTTP response
		responseStruct := struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		}{
			Status:  "FAIL",
			Message: "User not found",
		}

		responseBody, _ := json.Marshal(responseStruct)
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write(responseBody)
		}))
		defer mockServer.Close()

		uc.SetHTTPClient(mockServer.Client())

		expectedMessage.RespCode = 404
		expectedMessage.RespStatus = "FAIL"
		expectedMessage.RespMessage = "FAILED User not found"

		mockProducer.On("SendMessage", mock.Anything).Return(int32(0), int64(0), nil)

		err := uc.SendMessage(context.Background(), incomingMsg)
		assert.NoError(t, err)
		mockProducer.AssertExpectations(t)
	})

	t.Run("failure case - internal server error", func(t *testing.T) {
		mockProducer := new(MockProducer)
		uc := usecase.NewUserUsecase(mockProducer)

		// Mock HTTP error
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer mockServer.Close()

		uc.SetHTTPClient(mockServer.Client())

		expectedMessage.RespCode = 500
		expectedMessage.RespStatus = "Internal Server Error"
		expectedMessage.RespMessage = "FAILED USER VERIFY: error during POST request"

		mockProducer.On("SendMessage", mock.Anything).Return(int32(0), int64(0), nil)

		err := uc.SendMessage(context.Background(), incomingMsg)
		assert.NoError(t, err)
		mockProducer.AssertExpectations(t)
	})

	t.Run("failure case - JSON unmarshal error", func(t *testing.T) {
		mockProducer := new(MockProducer)
		uc := usecase.NewUserUsecase(mockProducer)

		invalidMsg := &sarama.ConsumerMessage{
			Value: []byte(`{invalid_json}`),
			Key:   []byte("1"),
		}

		err := uc.SendMessage(context.Background(), invalidMsg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid character")
		mockProducer.AssertExpectations(t)
	})
}
