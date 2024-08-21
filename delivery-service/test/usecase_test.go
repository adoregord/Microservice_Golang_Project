package test

import (
	"context"
	"delivery_microservice/internal/domain"
	"delivery_microservice/internal/usecase"
	"encoding/json"
	"errors"
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) SendMessages(msgs []*sarama.ProducerMessage) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockProducer) Close() error {
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

func TestDeliveryUsecase_SendMessage_Success(t *testing.T) {
	// Arrange
	producer := new(MockProducer)
	uc := usecase.NewDeliveryUsecase(producer)

	mockMessage := domain.Message{
		OrderType: "Buy Item",
		UserId:    "123",
	}

	mockMessageBytes, _ := json.Marshal(mockMessage)
	consumerMessage := &sarama.ConsumerMessage{Value: mockMessageBytes, Key: []byte("test-key")}

	// Mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/deliveryitem/?userID=123", r.URL.String())
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	}))
	defer server.Close()

	// Override URL in the usecase
	originalHttpClient := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: &mockTransport{server.URL}}

	// Mock SendMessage response
	producer.On("SendMessage", mock.Anything).Return(int32(1), int64(1), nil)

	// Act
	err := uc.SendMessage(context.Background(), consumerMessage)

	// Assert
	assert.NoError(t, err)
	producer.AssertExpectations(t)

	// Restore original HTTP client
	http.DefaultClient = originalHttpClient
}

//func TestDeliveryUsecase_SendMessage_InvalidOrderType(t *testing.T) {
//	// Arrange
//	producer := new(MockProducer)
//	uc := usecase.NewDeliveryUsecase(producer)
//
//	mockMessage := domain.Message{
//		OrderType: "Invalid Type",
//		UserId:    "123",
//	}
//
//	mockMessageBytes, _ := json.Marshal(mockMessage)
//	consumerMessage := &sarama.ConsumerMessage{Value: mockMessageBytes, Key: []byte("test-key")}
//
//	// Set up the expected behavior for the mock producer
//	producer.On("SendMessage", mock.Anything).Return(nil, &sarama.ProducerError{Err: nil, Msg: &sarama.ProducerMessage{Partition: 0, Offset: 0}})
//
//	// Act
//	err := uc.SendMessage(context.Background(), consumerMessage)
//
//	// Assert
//	assert.Error(t, err)
//	assert.Contains(t, err.Error(), "INVALID order type for delivery")
//
//	// Ensure SendMessage was called once with any ProducerMessage
//	producer.AssertCalled(t, "SendMessage", mock.Anything)
//}

func TestDeliveryUsecase_SendMessage_FailedToSendMessage(t *testing.T) {
	// Arrange
	producer := new(MockProducer)
	uc := usecase.NewDeliveryUsecase(producer)

	mockMessage := domain.Message{
		OrderType: "Buy Item",
		UserId:    "123",
	}

	mockMessageBytes, _ := json.Marshal(mockMessage)
	consumerMessage := &sarama.ConsumerMessage{Value: mockMessageBytes, Key: []byte("test-key")}

	// Mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	}))
	defer server.Close()

	// Override URL in the usecase
	originalHttpClient := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: &mockTransport{server.URL}}

	// Mock SendMessage response to fail
	producer.On("SendMessage", mock.Anything).Return(int32(0), int64(0), errors.New("FAILED to send message"))

	// Act
	err := uc.SendMessage(context.Background(), consumerMessage)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "FAILED to send message")
	producer.AssertExpectations(t)

	// Restore original HTTP client
	http.DefaultClient = originalHttpClient
}

type mockTransport struct {
	url string
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp := httptest.NewRecorder()
	if req.URL.String() == "/deliveryitem/?userID=123" {
		resp.WriteHeader(http.StatusOK)
		resp.Write([]byte(`{"status":"success"}`))
		return resp.Result(), nil
	}
	return nil, errors.New("unexpected request")
}
