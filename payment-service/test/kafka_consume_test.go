package test

import (
	"context"
	"errors"
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"payment_microservice/internal/kafka"
	"testing"
)

// Mock for PaymentUsecaseInterface
type MockPaymentUsecase struct {
	mock.Mock
}

func (m *MockPaymentUsecase) SetHTTPClient(client *http.Client) {
	//TODO implement me
	panic("implement me")
}

func (m *MockPaymentUsecase) SendMessage(ctx context.Context, msg *sarama.ConsumerMessage) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

// Test cases for ConsumeClaim
func TestMessageHandler_ConsumeClaim_Success(t *testing.T) {
	mockUsecase := new(MockPaymentUsecase)
	mockProducer := new(MockProducer)

	// Mock the SendMessage method to return nil (no error)
	mockUsecase.On("SendMessage", mock.Anything, mock.AnythingOfType("*sarama.ConsumerMessage")).Return(nil)

	handler := kafka.NewMessageHandler(mockProducer, mockUsecase)

	claim := &mockConsumerGroupClaim{
		messages: []*sarama.ConsumerMessage{
			{Value: []byte(`{"payment_id": 1}`)},
		},
	}

	err := handler.ConsumeClaim(&mockConsumerGroupSession{}, claim)
	assert.NoError(t, err)
	mockUsecase.AssertExpectations(t)
}

func TestMessageHandler_ConsumeClaim_Error(t *testing.T) {
	mockUsecase := new(MockPaymentUsecase)
	mockProducer := new(MockProducer)

	// Mock the SendMessage method to return an error
	mockUsecase.On("SendMessage", mock.Anything, mock.AnythingOfType("*sarama.ConsumerMessage")).Return(errors.New("some error"))

	handler := kafka.NewMessageHandler(mockProducer, mockUsecase)

	claim := &mockConsumerGroupClaim{
		messages: []*sarama.ConsumerMessage{
			{Value: []byte(`{"payment_id": 1}`)},
		},
	}

	err := handler.ConsumeClaim(&mockConsumerGroupSession{}, claim)
	assert.NoError(t, err)
	mockUsecase.AssertExpectations(t)
}

// Test NewKafkaConsumer
func TestNewKafkaConsumer_Success(t *testing.T) {
	brokers := []string{"localhost:29092"}
	groupID := "test-group"
	topics := []string{"test-topic"}
	msgHandler := new(kafka.MessageHandler)

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
	msgHandler := new(kafka.MessageHandler)

	consumer, err := kafka.NewKafkaConsumer(brokers, groupID, topics, msgHandler)
	assert.Error(t, err)
	assert.Nil(t, consumer)
}

// Test KafkaConsumer Consume method
//func TestKafkaConsumer_Consume_Success(t *testing.T) {
//	mockConsumerGroup := new(MockConsumerGroup)
//	topics := []string{"test-topic"}
//
//	// Use a context with a timeout to break the loop after a short duration
//	ctx := context.TODO()
//
//	// Create a mock message handler
//	mockProducer := new(MockProducer)
//	mockUsecase := new(MockPaymentUsecase)
//	handler := kafka.NewMessageHandler(mockProducer, mockUsecase)
//
//	kc := &kafka.KafkaConsumer{
//		Consumer: mockConsumerGroup,
//		Topics:   topics,
//		Handler:  handler,
//	}
//
//	// Mock Consume method to simulate successful consumption
//	mockConsumerGroup.On("Consume", ctx, topics, kc.Handler).Return(nil)
//
//	err := kc.Consume(ctx)
//	assert.NoError(t, err)
//	mockConsumerGroup.AssertExpectations(t)
//}

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

// Test KafkaConsumer Close method
func TestKafkaConsumer_Close(t *testing.T) {
	mockConsumerGroup := new(MockConsumerGroup)

	kc := &kafka.KafkaConsumer{
		Consumer: mockConsumerGroup,
	}

	// Mock Close method to return no error
	mockConsumerGroup.On("Close").Return(nil)

	err := kc.Close()
	assert.NoError(t, err)
	mockConsumerGroup.AssertExpectations(t)
}

func TestKafkaConsumer_Close_Error(t *testing.T) {
	mockConsumerGroup := new(MockConsumerGroup)

	kc := &kafka.KafkaConsumer{
		Consumer: mockConsumerGroup,
	}

	// Mock Close method to return an error
	mockConsumerGroup.On("Close").Return(errors.New("close error"))

	err := kc.Close()
	assert.Error(t, err)
	assert.EqualError(t, err, "close error")
	mockConsumerGroup.AssertExpectations(t)
}

// Mocks for ConsumerGroupSession and ConsumerGroupClaim
type mockConsumerGroupSession struct{}

func (s *mockConsumerGroupSession) Commit() {
	//TODO implement me
	panic("implement me")
}

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

// Mock for Sarama ConsumerGroup
type MockConsumerGroup struct {
	mock.Mock
}

func (c *MockConsumerGroup) Consume(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
	args := c.Called(ctx, topics, handler)
	return args.Error(0)
}

func (c *MockConsumerGroup) Errors() <-chan error {
	return nil
}

func (c *MockConsumerGroup) Close() error {
	args := c.Called()
	return args.Error(0)
}

func (c *MockConsumerGroup) Pause(partitions map[string][]int32) {}

func (c *MockConsumerGroup) Resume(partitions map[string][]int32) {}

func (c *MockConsumerGroup) PauseAll() {}

func (c *MockConsumerGroup) ResumeAll() {}
