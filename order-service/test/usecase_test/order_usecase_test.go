package test

import (
	"context"
	"errors"
	"order_microservice/internal/domain"
	"order_microservice/internal/usecase"
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repository and producer
type MockOrderRepo struct {
	mock.Mock
}

func (m *MockOrderRepo) StoreOrderDetails(orderReq domain.OrderRequest, ctx context.Context) (*domain.Message, error) {
	args := m.Called(orderReq, ctx)
	return args.Get(0).(*domain.Message), args.Error(1)
}

func (m *MockOrderRepo) UpdateOrderDetails(msg domain.Message, ctx context.Context) error {
	args := m.Called(msg, ctx)
	return args.Error(0)
}

func (m *MockOrderRepo) EditRetryOrder(orderRetryReq *domain.RetryOrder, ctx context.Context) (*domain.Message, error) {
	args := m.Called(orderRetryReq, ctx)
	return args.Get(0).(*domain.Message), args.Error(1)
}

type MockProducer struct {
	mock.Mock
}

// IsTransactional implements sarama.SyncProducer.
func (m *MockProducer) IsTransactional() bool {
	panic("unimplemented")
}

func (m *MockProducer) SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	args := m.Called(msg)
	return args.Get(0).(int32), args.Get(1).(int64), args.Error(2)
}

func (m *MockProducer) SendMessages(msgs []*sarama.ProducerMessage) error {
	args := m.Called(msgs)
	return args.Error(0)
}

func (m *MockProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Additional required methods for sarama.SyncProducer
func (m *MockProducer) TxnStatus() sarama.ProducerTxnStatusFlag {
	args := m.Called()
	return args.Get(0).(sarama.ProducerTxnStatusFlag)
}

func (m *MockProducer) BeginTxn() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockProducer) CommitTxn() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockProducer) AbortTxn() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockProducer) AddOffsetsToTxn(offsets map[string][]*sarama.PartitionOffsetMetadata, groupId string) error {
	args := m.Called(offsets, groupId)
	return args.Error(0)
}

func (m *MockProducer) AddMessageToTxn(msg *sarama.ConsumerMessage, groupId string, metadata *string) error {
	args := m.Called(msg, groupId, metadata)
	return args.Error(0)
}

// Test OrderInit - Positive Case
func TestOrderInit_Positive(t *testing.T) {
	mockRepo := new(MockOrderRepo)
	mockProducer := new(MockProducer)
	uc := usecase.NewOrderUsecase(mockRepo, mockProducer)

	orderReq := &domain.OrderRequest{
		OrderType: "type1",
		UserID:    "1",
		ItemID:    1,
		Amount:    1000,
		ID:        1,
	}

	expectedMessage := &domain.Message{
		OrderID:      1,
		OrderType:    "type1",
		UserId:       "1",
		ItemID:       1,
		Amount:       1000,
		RespCode:     200,
		RespMessage:  "Order Success",
		OrderService: "Order Init",
	}

	mockRepo.On("StoreOrderDetails", *orderReq, mock.Anything).Return(expectedMessage, nil)
	mockProducer.On("SendMessage", mock.Anything).Return(int32(0), int64(0), nil)

	message, err := uc.OrderInit(orderReq, context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedMessage, message)
	mockRepo.AssertExpectations(t)
	mockProducer.AssertExpectations(t)
}

// Test OrderInit - Negative Case
func TestOrderInit_Negative(t *testing.T) {
	mockRepo := new(MockOrderRepo)
	mockProducer := new(MockProducer)
	uc := usecase.NewOrderUsecase(mockRepo, mockProducer)

	orderReq := &domain.OrderRequest{
		ID:        1,
		OrderType: "TestOrder",
		UserID:    "123",
		ItemID:    456,
		Amount:    789,
	}

	// Mocking the StoreOrderDetails to return nil *domain.Message and an error
	mockRepo.On("StoreOrderDetails", *orderReq, mock.Anything).Return((*domain.Message)(nil), errors.New("DB error"))

	result, err := uc.OrderInit(orderReq, context.Background())

	// Expecting an error and a nil result
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "DB error")
	mockRepo.AssertExpectations(t)
}

// Test UpdateOrderDetails - Positive Case
func TestUpdateOrderDetails_Positive(t *testing.T) {
	mockRepo := new(MockOrderRepo)
	usecase := usecase.NewOrderUsecase(mockRepo, nil)

	incomingMessage := domain.Message{
		OrderID:     1,
		RespCode:    200,
		RespMessage: "Update Success",
	}

	mockRepo.On("UpdateOrderDetails", incomingMessage, mock.Anything).Return(nil)

	err := usecase.UpdateOrderDetails(&sarama.ConsumerMessage{Value: []byte(`{"OrderID":1,"RespCode":200,"RespMessage":"Update Success"}`)}, context.Background())

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// Test UpdateOrderDetails - Negative Case
func TestUpdateOrderDetails_Negative(t *testing.T) {
	mockRepo := new(MockOrderRepo)
	usecase := usecase.NewOrderUsecase(mockRepo, nil)

	incomingMessage := domain.Message{
		OrderID:     1,
		RespCode:    200,
		RespMessage: "Update Failed",
	}

	mockRepo.On("UpdateOrderDetails", incomingMessage, mock.Anything).Return(errors.New("failed to update order"))

	err := usecase.UpdateOrderDetails(&sarama.ConsumerMessage{Value: []byte(`{"OrderID":1,"RespCode":200,"RespMessage":"Update Failed"}`)}, context.Background())

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

// Test OrderRetry - Positive Case
func TestOrderRetry_Positive(t *testing.T) {
	mockRepo := new(MockOrderRepo)
	mockProducer := new(MockProducer)
	usecase := usecase.NewOrderUsecase(mockRepo, mockProducer)

	orderRetry := &domain.RetryOrder{
		OrderID:   1,
		UserID:    "1",
		OrderType: "type1",
		ItemID:    1,
		Amount:    1000,
	}

	expectedMessage := &domain.Message{
		OrderID:      1,
		OrderType:    "type1",
		UserId:       "1",
		ItemID:       1,
		Amount:       1000,
		RespCode:     200,
		RespMessage:  "Order Retry Success",
		OrderService: "Order Init",
	}

	mockRepo.On("EditRetryOrder", orderRetry, mock.Anything).Return(expectedMessage, nil)
	mockProducer.On("SendMessage", mock.Anything).Return(int32(0), int64(0), nil)

	message, err := usecase.OrderRetry(orderRetry, context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedMessage, message)
	mockRepo.AssertExpectations(t)
	mockProducer.AssertExpectations(t)
}

// Test OrderRetry - Negative Case
func TestOrderRetry_Negative(t *testing.T) {
	mockRepo := new(MockOrderRepo)
	mockProducer := new(MockProducer)
	uc := usecase.NewOrderUsecase(mockRepo, mockProducer)

	orderRetry := &domain.RetryOrder{
		OrderID:   1,
		UserID:    "1",
		OrderType: "type1",
		ItemID:    1,
		Amount:    1000,
	}

	mockRepo.On("EditRetryOrder", orderRetry, mock.Anything).Return((*domain.Message)(nil), errors.New("failed to retry order"))

	msg, err := uc.OrderRetry(orderRetry, context.Background())
	if err != nil {
		return
	}

	assert.Error(t, err)
	assert.Nil(t, msg)
	mockRepo.AssertExpectations(t)
	mockProducer.AssertExpectations(t)
}
