package test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"item_microservice/internal/domain"
	"item_microservice/internal/usecase"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestItemUsecase_SendMessage(t *testing.T) {
	incomingMsg := &sarama.ConsumerMessage{
		Value: []byte(`{"OrderID":1,"OrderType":"Buy Item","UserId":"testuser","ItemID":1,"Amount":1000}`),
		Key:   []byte("1"),
	}

	expectedMessage := domain.Message{
		OrderID:   1,
		OrderType: "Buy Item",
		UserId:    "testuser",
		ItemID:    1,
		Amount:    1000,
	}

	t.Run("success case - valid item", func(t *testing.T) {
		mockProducer := new(MockProducer)
		uc := usecase.NewItemUsecase(mockProducer)

		// Mock HTTP response
		responseStruct := domain.Response{
			Status: "OK",
			Data: domain.Data{
				ID:       1,
				Quantity: 2000,
				Price:    100,
			},
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
		expectedMessage.Total = 100000 // 1000 * 100
		expectedMessage.RespMessage = "SUCCESS: Item is Reserved"

		mockProducer.On("SendMessage", mock.Anything).Return(int32(0), int64(0), nil)

		err := uc.SendMessage(context.Background(), incomingMsg)
		assert.NoError(t, err)

		mockProducer.AssertExpectations(t)

		// Verify that the message sent to Kafka contains the correct data
		//var sentMsg domain.Message
		//mockProducer.AssertCalled(t, "SendMessage", mock.MatchedBy(func(msg *sarama.ProducerMessage) bool {
		//	json.Unmarshal(msg.Value.([]byte), &sentMsg)
		//	return assert.Equal(t, expectedMessage, sentMsg)
		//}))
	})

	t.Run("failure case - HTTP timeout", func(t *testing.T) {
		mockProducer := new(MockProducer)
		uc := usecase.NewItemUsecase(mockProducer)

		// Mock HTTP timeout
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * time.Second)
		}))
		defer mockServer.Close()

		uc.SetHTTPClient(&http.Client{Timeout: 1 * time.Second}) // Set shorter timeout

		expectedMessage.RespCode = 500
		expectedMessage.RespStatus = "Internal Server Error"
		expectedMessage.RespMessage = "FAILED CHECK ITEM: Get \"http://127.0.0.1:xxxxx/item/?itemID=1\": context deadline exceeded (Client.Timeout exceeded while awaiting headers)"

		mockProducer.On("SendMessage", mock.Anything).Return(int32(0), int64(0), nil)

		err := uc.SendMessage(context.Background(), incomingMsg)
		assert.NoError(t, err)
		mockProducer.AssertExpectations(t)
	})

	// Include other test cases here for DeductStock, ReAddStock, etc.

	// ...

	t.Run("failure case - Kafka send message error", func(t *testing.T) {
		mockProducer := new(MockProducer)
		uc := usecase.NewItemUsecase(mockProducer)

		// Mock HTTP response
		responseStruct := domain.Response{
			Status: "OK",
			Data: domain.Data{
				ID:       1,
				Quantity: 2000,
				Price:    100,
			},
		}

		responseBody, _ := json.Marshal(responseStruct)
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(responseBody)
		}))
		defer mockServer.Close()

		uc.SetHTTPClient(mockServer.Client())

		mockProducer.On("SendMessage", mock.Anything).Return(int32(0), int64(0), errors.New("kafka error"))

		err := uc.SendMessage(context.Background(), incomingMsg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "kafka error")
		mockProducer.AssertExpectations(t)
	})
}

func TestItemUsecase_ValidateItemKey(t *testing.T) {
	mockProducer := new(MockProducer)
	uc := usecase.NewItemUsecase(mockProducer)

	t.Run("success case - valid item key", func(t *testing.T) {
		responseStruct := domain.Response{
			Status: "OK",
			Data: domain.Data{
				ID:       1,
				Quantity: 1,
				Price:    1,
			},
		}

		responseBody, _ := json.Marshal(responseStruct)
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(responseBody)
		}))
		defer mockServer.Close()

		uc.SetHTTPClient(mockServer.Client())

		incomingMessage := &domain.Message{
			ItemID: 1,
			Amount: 1,
		}

		err := uc.ValidateItemKey(incomingMessage)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, incomingMessage.RespCode)
		assert.Equal(t, "OK", incomingMessage.RespStatus)
		assert.Equal(t, "SUCCESS: Key is Reserved", incomingMessage.RespMessage)
		assert.Equal(t, 5300, incomingMessage.Total)
	})

	t.Run("failure case - insufficient stock", func(t *testing.T) {
		responseStruct := domain.Response{
			Status: "OK",
			Data: domain.Data{
				ID:       1,
				Quantity: 100,
				Price:    100,
			},
		}

		responseBody, _ := json.Marshal(responseStruct)
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(responseBody)
		}))
		defer mockServer.Close()

		uc.SetHTTPClient(mockServer.Client())

		incomingMessage := &domain.Message{
			ItemID: 1,
			Amount: 500,
		}

		err := uc.ValidateItemKey(incomingMessage)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, incomingMessage.RespCode)
		assert.Equal(t, "Bad Request", incomingMessage.RespStatus)
		assert.Equal(t, "FAILED: Key stock is not enough", incomingMessage.RespMessage)
		assert.Equal(t, 0, incomingMessage.Total)
	})
}

func TestItemUsecase_DeductStock(t *testing.T) {
	mockProducer := new(MockProducer)
	uc := usecase.NewItemUsecase(mockProducer)

	t.Run("success case - stock deducted", func(t *testing.T) {
		responseStruct := domain.Response{
			Status: "OK",
		}

		responseBody, _ := json.Marshal(responseStruct)
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(responseBody)
		}))
		defer mockServer.Close()

		uc.SetHTTPClient(mockServer.Client())

		err := uc.DeductStock(1, 100)
		assert.NoError(t, err)
	})

}

func TestItemUsecase_ReAddStock(t *testing.T) {
	mockProducer := new(MockProducer)
	uc := usecase.NewItemUsecase(mockProducer)

	t.Run("success case - stock re-added", func(t *testing.T) {
		responseStruct := domain.Response{
			Status: "OK",
		}

		responseBody, _ := json.Marshal(responseStruct)
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(responseBody)
		}))
		defer mockServer.Close()

		uc.SetHTTPClient(mockServer.Client())

		err := uc.ReAddStock(1, 100)
		assert.NoError(t, err)
	})

}

func TestItemUsecase_RollbackItem(t *testing.T) {
	mockProducer := new(MockProducer)
	uc := usecase.NewItemUsecase(mockProducer)

	t.Run("success case - rollback", func(t *testing.T) {
		incomingMsg := &sarama.ConsumerMessage{
			Value: []byte(`{"OrderID":1,"OrderType":"Buy Item","UserId":"testuser","ItemID":1,"Amount":1000}`),
			Key:   []byte("1"),
		}

		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer mockServer.Close()

		uc.SetHTTPClient(mockServer.Client())

		err := uc.RollbackItem(context.Background(), incomingMsg)
		assert.NoError(t, err)
	})
}
