package test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"payment_microservice/internal/domain"
	"payment_microservice/internal/usecase"
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProducer is a mock implementation of sarama.SyncProducer
type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) SendMessages(msgs []*sarama.ProducerMessage) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockProducer) TxnStatus() sarama.ProducerTxnStatusFlag {
	//TODO implement me
	panic("implement me")
}

func (m *MockProducer) IsTransactional() bool {
	//TODO implement me
	panic("implement me")
}

func (m *MockProducer) BeginTxn() error {
	//TODO implement me
	panic("implement me")
}

func (m *MockProducer) CommitTxn() error {
	//TODO implement me
	panic("implement me")
}

func (m *MockProducer) AbortTxn() error {
	//TODO implement me
	panic("implement me")
}

func (m *MockProducer) AddOffsetsToTxn(offsets map[string][]*sarama.PartitionOffsetMetadata, groupId string) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockProducer) AddMessageToTxn(msg *sarama.ConsumerMessage, groupId string, metadata *string) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockProducer) SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	args := m.Called(msg)
	return args.Get(0).(int32), args.Get(1).(int64), args.Error(2)
}

func (m *MockProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockHTTPClient is a mock implementation of the HTTP client
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Get(url string) (*http.Response, error) {
	args := m.Called(url)
	return args.Get(0).(*http.Response), args.Error(1)
}

// Test for PaymentUsecase.SendMessage
func TestPaymentUsecase_SendMessage(t *testing.T) {
	incomingMsg := &sarama.ConsumerMessage{
		Value: []byte(`{"OrderID":1,"OrderType":"payment","UserId":"1","ItemID":1,"Amount":1000,"Total":500}`),
		Key:   []byte("1"),
	}

	expectedMessage := domain.Message{
		OrderID:   1,
		OrderType: "payment",
		UserId:    "1",
		ItemID:    1,
		Amount:    1000,
		Total:     500,
	}

	t.Run("success case", func(t *testing.T) {
		mockProducer := new(MockProducer)
		uc := usecase.NewPaymentUsecase(mockProducer)
		// Mock HTTP response
		response := domain.Response{
			Status: "OK",
			Data: domain.Data{
				ID:      1,
				Name:    "User1",
				Balance: 1000,
			},
		}

		responseBody, _ := json.Marshal(response)
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(responseBody)
		}))
		defer mockServer.Close()

		uc.SetHTTPClient(mockServer.Client())

		expectedMessage.RespCode = 200
		expectedMessage.RespStatus = "OK"
		expectedMessage.RespMessage = "payment SUCCESS"

		mockProducer.On("SendMessage", mock.Anything).Return(int32(0), int64(0), nil)

		err := uc.SendMessage(context.Background(), incomingMsg)
		assert.NoError(t, err)
		mockProducer.AssertExpectations(t)
	})

	t.Run("failure case - insufficient balance", func(t *testing.T) {
		mockProducer := new(MockProducer)
		uc := usecase.NewPaymentUsecase(mockProducer)
		// Mock HTTP response
		response := domain.Response{
			Status: "OK",
			Data: domain.Data{
				ID:      1,
				Name:    "User1",
				Balance: 400,
			},
		}

		responseBody, _ := json.Marshal(response)
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(responseBody)
		}))
		defer mockServer.Close()

		uc.SetHTTPClient(mockServer.Client())

		expectedMessage.RespCode = 400
		expectedMessage.RespStatus = "Bad Request"
		expectedMessage.RespMessage = "FAILED: user balance is not enough"

		mockProducer.On("SendMessage", mock.Anything).Return(int32(0), int64(0), nil)

		err := uc.SendMessage(context.Background(), incomingMsg)
		assert.NoError(t, err)
		mockProducer.AssertExpectations(t)
	})

	t.Run("failure case - internal server error", func(t *testing.T) {
		mockProducer := new(MockProducer)
		uc := usecase.NewPaymentUsecase(mockProducer)
		// Mock HTTP error
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer mockServer.Close()

		uc.SetHTTPClient(mockServer.Client())

		expectedMessage.RespCode = 500
		expectedMessage.RespStatus = "Internal Server Error"
		expectedMessage.RespMessage = "FAILED PAYMENT: FAILED to make payment request"

		mockProducer.On("SendMessage", mock.Anything).Return(int32(0), int64(0), nil)

		err := uc.SendMessage(context.Background(), incomingMsg)
		assert.NoError(t, err)
		mockProducer.AssertExpectations(t)
	})

	t.Run("failure case - JSON unmarshal error", func(t *testing.T) {
		mockProducer := new(MockProducer)
		uc := usecase.NewPaymentUsecase(mockProducer)
		invalidMsg := &sarama.ConsumerMessage{
			Value: []byte(`{invalid_json}`),
			Key:   []byte("1"),
		}

		err := uc.SendMessage(context.Background(), invalidMsg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid character")
		mockProducer.AssertExpectations(t)
	})

	t.Run("failure case - JSON unmarshal error", func(t *testing.T) {
		mockProducer := new(MockProducer)
		uc := usecase.NewPaymentUsecase(mockProducer)

		// Create an invalid JSON string that can't be unmarshaled into Message struct
		invalidJSON := `{
        "orderID": "not_an_integer",
        "orderType": "payment",
        "userId": "user123",
        "itemID": "not_an_integer",
        "amount": "not_an_integer",
        "retry": "not_an_integer"
    }`

		// Remove the mock expectation for SendMessage
		// mockProducer.On("SendMessage", mock.Anything).Return(int32(0), int64(0), nil)

		err := uc.SendMessage(context.Background(), &sarama.ConsumerMessage{Value: []byte(invalidJSON), Key: []byte("1")})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "json: cannot unmarshal")

		// Assert that SendMessage was not called on the mock producer
		mockProducer.AssertNotCalled(t, "SendMessage", mock.Anything)
	})

	t.Run("failure case - Kafka send message error", func(t *testing.T) {
		mockProducer := new(MockProducer)
		uc := usecase.NewPaymentUsecase(mockProducer)
		// Mock HTTP response
		response := domain.Response{
			Status: "OK",
			Data: domain.Data{
				ID:      1,
				Name:    "User1",
				Balance: 1000,
			},
		}

		responseBody, _ := json.Marshal(response)
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
