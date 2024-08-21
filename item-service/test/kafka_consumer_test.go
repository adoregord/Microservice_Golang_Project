package test

import (
	"context"
	"errors"
	"item_microservice/internal/domain"
	"item_microservice/internal/kafka"
	"net/http"
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock interfaces
type MockItemUsecase struct {
	mock.Mock
}

func (m *MockItemUsecase) SetHTTPClient(client *http.Client) {
	//TODO implement me
	panic("implement me")
}

// DeductStock implements usecase.ItemUsecaseInterface.
func (m *MockItemUsecase) DeductStock(id int, amount int) error {
	panic("unimplemented")
}

// ReAddStock implements usecase.ItemUsecaseInterface.
func (m *MockItemUsecase) ReAddStock(id int, amount int) error {
	panic("unimplemented")
}

// ValidateItem implements usecase.ItemUsecaseInterface.
func (m *MockItemUsecase) ValidateItem(incoming *domain.Message) error {
	panic("unimplemented")
}

// ValidateItemKey implements usecase.ItemUsecaseInterface.
func (m *MockItemUsecase) ValidateItemKey(incoming *domain.Message) error {
	panic("unimplemented")
}

func (m *MockItemUsecase) SendMessage(ctx context.Context, msg *sarama.ConsumerMessage) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockItemUsecase) RollbackItem(ctx context.Context, msg *sarama.ConsumerMessage) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

// Mock ConsumerGroupSession and ConsumerGroupClaim
type mockConsumerGroupSession struct{}

// Claims implements sarama.ConsumerGroupSession.
func (m *mockConsumerGroupSession) Claims() map[string][]int32 {
	panic("unimplemented")
}

// GenerationID implements sarama.ConsumerGroupSession.
func (m *mockConsumerGroupSession) GenerationID() int32 {
	panic("unimplemented")
}

// MemberID implements sarama.ConsumerGroupSession.
func (m *mockConsumerGroupSession) MemberID() string {
	panic("unimplemented")
}

func (m *mockConsumerGroupSession) MarkMessage(msg *sarama.ConsumerMessage, metadata string) {}
func (m *mockConsumerGroupSession) Context() context.Context                                 { return context.Background() }
func (m *mockConsumerGroupSession) Commit()                                                  {}
func (m *mockConsumerGroupSession) ResetOffset(topic string, partition int32, offset int64, metadata string) {
}
func (m *mockConsumerGroupSession) MarkOffset(topic string, partition int32, offset int64, metadata string) {
}

type mockConsumerGroupClaim struct {
	messages []*sarama.ConsumerMessage
}

func (m *mockConsumerGroupClaim) Topic() string              { return "mock_topic" }
func (m *mockConsumerGroupClaim) Partition() int32           { return 0 }
func (m *mockConsumerGroupClaim) InitialOffset() int64       { return 0 }
func (m *mockConsumerGroupClaim) HighWaterMarkOffset() int64 { return 0 }
func (m *mockConsumerGroupClaim) Messages() <-chan *sarama.ConsumerMessage {
	msgChan := make(chan *sarama.ConsumerMessage, len(m.messages))
	for _, msg := range m.messages {
		msgChan <- msg
	}
	close(msgChan)
	return msgChan
}

// Test cases for ConsumeClaim
func TestMessageHandler_ConsumeClaim_Success_SendMessage(t *testing.T) {
	mockUsecase := new(MockItemUsecase)
	mockProducer := new(MockProducer)

	// Mock the SendMessage method to return nil (no error)
	mockUsecase.On("SendMessage", mock.Anything, mock.AnythingOfType("*sarama.ConsumerMessage")).Return(nil)

	handler := kafka.NewMessageHandler(mockProducer, mockUsecase)

	claim := &mockConsumerGroupClaim{
		messages: []*sarama.ConsumerMessage{
			{Topic: "item_reserve", Value: []byte(`{"itemID": 1, "amount": 2}`)},
		},
	}

	err := handler.ConsumeClaim(&mockConsumerGroupSession{}, claim)
	assert.NoError(t, err)
	mockUsecase.AssertExpectations(t)
}

func TestMessageHandler_ConsumeClaim_Error_SendMessage(t *testing.T) {
	mockUsecase := new(MockItemUsecase)
	mockProducer := new(MockProducer)

	// Mock the SendMessage method to return an error
	mockUsecase.On("SendMessage", mock.Anything, mock.AnythingOfType("*sarama.ConsumerMessage")).Return(errors.New("some error"))

	handler := kafka.NewMessageHandler(mockProducer, mockUsecase)

	claim := &mockConsumerGroupClaim{
		messages: []*sarama.ConsumerMessage{
			{Topic: "item_reserve", Value: []byte(`{"itemID": 1, "amount": 2}`)},
		},
	}

	err := handler.ConsumeClaim(&mockConsumerGroupSession{}, claim)
	assert.NoError(t, err)
	mockUsecase.AssertExpectations(t)
}

func TestMessageHandler_ConsumeClaim_Success_RollbackItem(t *testing.T) {
	mockUsecase := new(MockItemUsecase)
	mockProducer := new(MockProducer)

	// Mock the RollbackItem method to return nil (no error)
	mockUsecase.On("RollbackItem", mock.Anything, mock.AnythingOfType("*sarama.ConsumerMessage")).Return(nil)

	handler := kafka.NewMessageHandler(mockProducer, mockUsecase)

	claim := &mockConsumerGroupClaim{
		messages: []*sarama.ConsumerMessage{
			{Topic: "item_rollback", Value: []byte(`{"itemID": 1, "amount": 2}`)},
		},
	}

	err := handler.ConsumeClaim(&mockConsumerGroupSession{}, claim)
	assert.NoError(t, err)
	mockUsecase.AssertExpectations(t)
}

func TestNewKafkaConsumer_Success(t *testing.T) {
	brokers := []string{"localhost:29092"}
	groupID := "item-group"
	topics := []string{"item_reserve"}
	mockUsecase := new(MockItemUsecase)
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
	groupID := "item-group"
	topics := []string{"item_reserve"}
	mockUsecase := new(MockItemUsecase)
	mockProducer := new(MockProducer)
	msgHandler := kafka.NewMessageHandler(mockProducer, mockUsecase)

	consumer, err := kafka.NewKafkaConsumer(brokers, groupID, topics, msgHandler)
	assert.Error(t, err)
	assert.Nil(t, consumer)
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

// Test KafkaConsumer Consume method
func TestKafkaConsumer_Consume_Error(t *testing.T) {
	mockConsumerGroup := new(MockConsumerGroup)
	topics := []string{"user-topic"}
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
	topics := []string{"user-topic"}
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
