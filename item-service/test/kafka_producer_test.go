package test

import (
	"errors"
	"github.com/IBM/sarama"
	"item_microservice/internal/kafka"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func (m *MockProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockProducer to simulate Kafka producer behavior
type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	args := m.Called(msg)
	return args.Get(0).(int32), args.Get(1).(int64), args.Error(2)
}

// MockSaramaProducer is a mock of the Sarama SyncProducer
type MockSaramaProducer struct {
	mock.Mock
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
