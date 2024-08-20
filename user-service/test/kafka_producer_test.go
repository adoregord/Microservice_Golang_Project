package test

import (
	"context"
	"errors"
	"github.com/IBM/sarama"
	"net/http"
	"testing"
	"user_microservice/internal/domain"
	"user_microservice/internal/kafka"

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
