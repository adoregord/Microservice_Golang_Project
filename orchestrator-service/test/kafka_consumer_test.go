package test

import (
	"context"
	"errors"
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"microservice_orchestrator/internal/domain"
	"microservice_orchestrator/internal/kafka"
	"testing"
)

// Mock for OrchestratorUsecaseInterface
type MockOrchestratorUsecase struct {
	mock.Mock
}

func (m *MockOrchestratorUsecase) ViewOrchesSteps(ctx context.Context, msg *sarama.ConsumerMessage) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockOrchestratorUsecase) OrchesLog(message domain.Message, ctx context.Context) error {
	args := m.Called(message, ctx)
	return args.Error(0)
}

func TestMessageHandler_ConsumeClaim_Success(t *testing.T) {
	mockUsecase := new(MockOrchestratorUsecase)
	mockProducer := new(MockProducer)

	// Mock the ViewOrchesSteps method to return nil (no error)
	mockUsecase.On("ViewOrchesSteps", mock.Anything, mock.AnythingOfType("*sarama.ConsumerMessage")).Return(nil)

	handler := kafka.NewMessageHandler(mockProducer, mockUsecase)

	claim := &mockConsumerGroupClaim{
		messages: []*sarama.ConsumerMessage{
			{Value: []byte(`{"orderID": 1}`)},
		},
	}

	err := handler.ConsumeClaim(&mockConsumerGroupSession{}, claim)
	assert.NoError(t, err)
	mockUsecase.AssertExpectations(t)
}

func TestMessageHandler_ConsumeClaim_Error(t *testing.T) {
	mockUsecase := new(MockOrchestratorUsecase)
	mockProducer := new(MockProducer)

	// Mock the ViewOrchesSteps method to return an error
	mockUsecase.On("ViewOrchesSteps", mock.Anything, mock.AnythingOfType("*sarama.ConsumerMessage")).Return(errors.New("some error"))

	handler := kafka.NewMessageHandler(mockProducer, mockUsecase)

	claim := &mockConsumerGroupClaim{
		messages: []*sarama.ConsumerMessage{
			{Value: []byte(`{"orderID": 1}`)},
		},
	}

	err := handler.ConsumeClaim(&mockConsumerGroupSession{}, claim)
	assert.NoError(t, err)
	mockUsecase.AssertExpectations(t)
}

func TestNewKafkaConsumer_Success(t *testing.T) {
	brokers := []string{"localhost:29092"}
	groupID := "test-group"
	topics := []string{"test-topic"}
	mockUsecase := new(MockOrchestratorUsecase)
	mockProducer := new(MockProducer)
	msgHandler := kafka.NewMessageHandler(mockProducer, mockUsecase)

	consumer, err := kafka.NewKafkaConsumer(brokers, groupID, topics, msgHandler)
	assert.NoError(t, err)
	assert.NotNil(t, consumer)
	assert.Equal(t, topics, consumer.Topics)
	assert.Equal(t, msgHandler, consumer.Handler)
}

func TestNewKafkaConsumer_Error(t *testing.T) {
	brokers := []string{"invalid-broker"}
	groupID := "test-group"
	topics := []string{"test-topic"}
	mockUsecase := new(MockOrchestratorUsecase)
	mockProducer := new(MockProducer)
	msgHandler := kafka.NewMessageHandler(mockProducer, mockUsecase)

	consumer, err := kafka.NewKafkaConsumer(brokers, groupID, topics, msgHandler)
	assert.Error(t, err)
	assert.Nil(t, consumer)
}

func TestKafkaConsumer_Consume_Error(t *testing.T) {
	mockConsumerGroup := new(MockConsumerGroup)
	topics := []string{"test-topic"}
	ctx := context.TODO()

	kc := &kafka.KafkaConsumer{
		Consumer: mockConsumerGroup,
		Topics:   topics,
		Handler:  new(kafka.MessageHandler),
	}

	// Mock Consume method to return an error
	mockConsumerGroup.On("Consume", ctx, topics, kc.Handler).Return(errors.New("consume error"))

	err := kc.Consume(ctx)
	assert.Error(t, err)
	assert.EqualError(t, err, "consume error")
	mockConsumerGroup.AssertExpectations(t)
}

func TestKafkaConsumer_Consume_ContextCancel(t *testing.T) {
	mockConsumerGroup := new(MockConsumerGroup)
	topics := []string{"test-topic"}
	ctx, cancel := context.WithCancel(context.TODO())

	kc := &kafka.KafkaConsumer{
		Consumer: mockConsumerGroup,
		Topics:   topics,
		Handler:  new(kafka.MessageHandler),
	}

	// Mock Consume method to return nil
	mockConsumerGroup.On("Consume", ctx, topics, kc.Handler).Return(nil)

	// Cancel the context to simulate context cancellation
	cancel()

	err := kc.Consume(ctx)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	mockConsumerGroup.AssertExpectations(t)
}

// Mock for Sarama ConsumerGroup
type MockConsumerGroup struct {
	mock.Mock
}

func (c *MockConsumerGroup) Consume(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
	args := c.Called(ctx, topics, handler)
	return args.Error(0)
}

func (c *MockConsumerGroup) Close() error {
	args := c.Called()
	return args.Error(0)
}

func (c *MockConsumerGroup) Errors() <-chan error {
	return nil
}

func (c *MockConsumerGroup) Pause(partitions map[string][]int32) {}

func (c *MockConsumerGroup) Resume(partitions map[string][]int32) {}

func (c *MockConsumerGroup) PauseAll() {}

func (c *MockConsumerGroup) ResumeAll() {}

// Mocks for ConsumerGroupSession and ConsumerGroupClaim
type mockConsumerGroupSession struct{}

func (s *mockConsumerGroupSession) Commit() {}

func (s *mockConsumerGroupSession) Claims() map[string][]int32 {
	return nil
}

func (s *mockConsumerGroupSession) MemberID() string {
	return ""
}

func (s *mockConsumerGroupSession) GenerationID() int32 {
	return 0
}

func (s *mockConsumerGroupSession) MarkOffset(topic string, partition int32, offset int64, metadata string) {
}

func (s *mockConsumerGroupSession) ResetOffset(topic string, partition int32, offset int64, metadata string) {
}

func (s *mockConsumerGroupSession) MarkMessage(msg *sarama.ConsumerMessage, metadata string) {}

func (s *mockConsumerGroupSession) Context() context.Context {
	return context.TODO()
}

type mockConsumerGroupClaim struct {
	messages []*sarama.ConsumerMessage
}

func (c *mockConsumerGroupClaim) Topic() string {
	return "test-topic"
}

func (c *mockConsumerGroupClaim) Partition() int32 {
	return 0
}

func (c *mockConsumerGroupClaim) InitialOffset() int64 {
	return 0
}

func (c *mockConsumerGroupClaim) HighWaterMarkOffset() int64 {
	return 0
}

func (c *mockConsumerGroupClaim) Messages() <-chan *sarama.ConsumerMessage {
	ch := make(chan *sarama.ConsumerMessage, len(c.messages))
	for _, msg := range c.messages {
		ch <- msg
	}
	close(ch)
	return ch
}
