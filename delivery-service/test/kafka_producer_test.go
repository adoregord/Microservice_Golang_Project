package test

import (
	"context"
	"delivery_microservice/internal/domain"
	"delivery_microservice/internal/kafka"
	"errors"
	"github.com/IBM/sarama"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock for UserUsecaseInterface
type MockUserUsecase struct {
	mock.Mock
}

func (m *MockUserUsecase) ValidateUser(incoming *domain.Message) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockUserUsecase) SetHTTPClient(client *http.Client) {
	//TODO implement me
	panic("implement me")
}

func (m *MockUserUsecase) SendMessage(ctx context.Context, msg *sarama.ConsumerMessage) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

// Test NewKafkaProducer
func TestNewKafkaProducer_Success(t *testing.T) {
	brokers := []string{"localhost:29092"}

	producer, err := kafka.NewKafkaProducer(brokers)
	assert.NoError(t, err)
	assert.NotNil(t, producer)
}

func TestNewKafkaProducer_Error(t *testing.T) {
	invalidBrokers := []string{"invalid-broker"}

	producer, err := kafka.NewKafkaProducer(invalidBrokers)
	assert.Error(t, err)
	assert.Nil(t, producer)
}

// Test SendMessage
func TestSendMessage_Success(t *testing.T) {
	mockProducer := new(MockProducer)
	topic := "test-topic"
	key := []byte("key")
	message := []byte("message")

	// Mock the SendMessage method to return a valid partition and offset
	mockProducer.On("SendMessage", mock.AnythingOfType("*sarama.ProducerMessage")).Return(int32(0), int64(0), nil)

	partition, offset, err := kafka.SendMessage(mockProducer, topic, key, message)
	assert.NoError(t, err)
	assert.Equal(t, int32(0), partition)
	assert.Equal(t, int64(0), offset)
	mockProducer.AssertExpectations(t)
}

func TestSendMessage_Error(t *testing.T) {
	mockProducer := new(MockProducer)
	topic := "test-topic"
	key := []byte("key")
	message := []byte("message")

	// Mock the SendMessage method to return an error
	mockProducer.On("SendMessage", mock.AnythingOfType("*sarama.ProducerMessage")).Return(int32(0), int64(0), errors.New("some error"))

	partition, offset, err := kafka.SendMessage(mockProducer, topic, key, message)
	assert.Error(t, err)
	assert.Equal(t, int32(0), partition)
	assert.Equal(t, int64(0), offset)
	mockProducer.AssertExpectations(t)
}
