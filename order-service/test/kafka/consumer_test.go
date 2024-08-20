package test

import (
	"context"
	"errors"
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"order_microservice/internal/domain"
	"order_microservice/internal/kafka"
	"testing"
)

// Mock for OrderUsecaseInterface
type MockOrderUsecase struct {
	mock.Mock
}

func (m *MockOrderUsecase) OrderInit(orderReq *domain.OrderRequest, ctx context.Context) (*domain.Message, error) {
	args := m.Called(orderReq, ctx)
	return args.Get(0).(*domain.Message), args.Error(1)
}

func (m *MockOrderUsecase) UpdateOrderDetails(msg *sarama.ConsumerMessage, ctx context.Context) error {
	args := m.Called(msg, ctx)
	return args.Error(0)
}

func (m *MockOrderUsecase) OrderRetry(orderRetry *domain.RetryOrder, ctx context.Context) (*domain.Message, error) {
	args := m.Called(orderRetry, ctx)
	return args.Get(0).(*domain.Message), args.Error(1)
}

// Mock for Sarama SyncProducer
type MockProducer struct {
	mock.Mock
}

func (p *MockProducer) SendMessages(msgs []*sarama.ProducerMessage) error {
	//TODO implement me
	panic("implement me")
}

func (p *MockProducer) Close() error {
	//TODO implement me
	panic("implement me")
}

func (p *MockProducer) TxnStatus() sarama.ProducerTxnStatusFlag {
	//TODO implement me
	panic("implement me")
}

func (p *MockProducer) IsTransactional() bool {
	//TODO implement me
	panic("implement me")
}

func (p *MockProducer) BeginTxn() error {
	//TODO implement me
	panic("implement me")
}

func (p *MockProducer) CommitTxn() error {
	//TODO implement me
	panic("implement me")
}

func (p *MockProducer) AbortTxn() error {
	//TODO implement me
	panic("implement me")
}

func (p *MockProducer) AddOffsetsToTxn(offsets map[string][]*sarama.PartitionOffsetMetadata, groupId string) error {
	//TODO implement me
	panic("implement me")
}

func (p *MockProducer) AddMessageToTxn(msg *sarama.ConsumerMessage, groupId string, metadata *string) error {
	//TODO implement me
	panic("implement me")
}

func (p *MockProducer) SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	args := p.Called(msg)
	return args.Get(0).(int32), args.Get(1).(int64), args.Error(2)
}

// Mock for Sarama ConsumerGroupSession
type MockConsumerGroupSession struct {
	mock.Mock
}

func (s *MockConsumerGroupSession) Claims() map[string][]int32 {
	//TODO implement me
	panic("implement me")
}

func (s *MockConsumerGroupSession) MemberID() string {
	//TODO implement me
	panic("implement me")
}

func (s *MockConsumerGroupSession) GenerationID() int32 {
	//TODO implement me
	panic("implement me")
}

func (s *MockConsumerGroupSession) MarkOffset(topic string, partition int32, offset int64, metadata string) {
	//TODO implement me
	panic("implement me")
}

func (s *MockConsumerGroupSession) Commit() {
	//TODO implement me
	panic("implement me")
}

func (s *MockConsumerGroupSession) ResetOffset(topic string, partition int32, offset int64, metadata string) {
	//TODO implement me
	panic("implement me")
}

func (s *MockConsumerGroupSession) MarkMessage(msg *sarama.ConsumerMessage, metadata string) {
	s.Called(msg, metadata)
}

func (s *MockConsumerGroupSession) Context() context.Context {
	return context.TODO()
}

// Mock for Sarama ConsumerGroupClaim
type MockConsumerGroupClaim struct {
	mock.Mock
	messages []*sarama.ConsumerMessage
}

func (c *MockConsumerGroupClaim) Topic() string {
	//TODO implement me
	panic("implement me")
}

func (c *MockConsumerGroupClaim) Partition() int32 {
	//TODO implement me
	panic("implement me")
}

func (c *MockConsumerGroupClaim) InitialOffset() int64 {
	//TODO implement me
	panic("implement me")
}

func (c *MockConsumerGroupClaim) HighWaterMarkOffset() int64 {
	//TODO implement me
	panic("implement me")
}

func (c *MockConsumerGroupClaim) Messages() <-chan *sarama.ConsumerMessage {
	ch := make(chan *sarama.ConsumerMessage, len(c.messages))
	for _, msg := range c.messages {
		ch <- msg
	}
	close(ch)
	return ch
}

func TestMessageHandler_ConsumeClaim_Success(t *testing.T) {
	mockUsecase := new(MockOrderUsecase)
	mockProducer := new(MockProducer)

	// Mock the UpdateOrderDetails method to return nil (no error)
	mockUsecase.On("UpdateOrderDetails", mock.Anything, mock.Anything).Return(nil)

	handler := kafka.NewMessageHandler(mockProducer, mockUsecase)

	claim := &MockConsumerGroupClaim{
		messages: []*sarama.ConsumerMessage{
			{Value: []byte(`{"order_id": 1}`)},
		},
	}

	session := new(MockConsumerGroupSession)
	session.On("MarkMessage", mock.Anything, "").Return()

	err := handler.ConsumeClaim(session, claim)
	assert.NoError(t, err)
	mockUsecase.AssertExpectations(t)
}

func TestMessageHandler_ConsumeClaim_Error(t *testing.T) {
	mockUsecase := new(MockOrderUsecase)
	mockProducer := new(MockProducer)

	// Mock the UpdateOrderDetails method to return an error
	mockUsecase.On("UpdateOrderDetails", mock.Anything, mock.Anything).Return(errors.New("some error"))

	handler := kafka.NewMessageHandler(mockProducer, mockUsecase)

	claim := &MockConsumerGroupClaim{
		messages: []*sarama.ConsumerMessage{
			{Value: []byte(`{"order_id": 1}`)},
		},
	}

	session := new(MockConsumerGroupSession)
	session.On("MarkMessage", mock.Anything, "").Return()

	err := handler.ConsumeClaim(session, claim)
	assert.NoError(t, err)
	mockUsecase.AssertExpectations(t)
}

func TestNewKafkaConsumer_Success(t *testing.T) {
	brokers := []string{"localhost:29092"}
	groupID := "test-group"
	topics := []string{"test-topic"}
	mockHandler := new(MockOrderUsecase)
	mockProducer := new(MockProducer)
	msgHandler := kafka.NewMessageHandler(mockProducer, mockHandler)

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
	mockHandler := new(MockOrderUsecase)
	mockProducer := new(MockProducer)
	msgHandler := kafka.NewMessageHandler(mockProducer, mockHandler)

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

	// Mock Consume method to return no error
	mockConsumerGroup.On("Consume", ctx, topics, kc.Handler).Return(nil)

	// Cancel the context to simulate context cancellation
	cancel()

	err := kc.Consume(ctx)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
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
