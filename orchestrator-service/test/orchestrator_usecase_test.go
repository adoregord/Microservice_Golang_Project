package test

import (
	"context"
	"errors"
	"microservice_orchestrator/internal/domain"
	"microservice_orchestrator/internal/usecase"
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repository and producer
type MockOrchestratorRepo struct {
	mock.Mock
}

func (m *MockOrchestratorRepo) GetNextTopic(incoming_message *domain.Message, isRetry bool, kontek context.Context) string {
	args := m.Called(incoming_message, isRetry, kontek)
	return args.String(0)
}

func (m *MockOrchestratorRepo) GetRollbackTopic(incoming_message *domain.Message, kontek context.Context) string {
	args := m.Called(incoming_message, kontek)
	return args.String(0)
}

func (m *MockOrchestratorRepo) OrchesLog(message domain.Message, kontek context.Context) error {
	args := m.Called(message, kontek)
	return args.Error(0)
}

// Test ViewOrchesSteps - Positive Case
func TestViewOrchesSteps_Positive(t *testing.T) {
	mockRepo := new(MockOrchestratorRepo)
	mockProducer := new(MockProducer)
	uc := usecase.NewOrchestratorUsecase(mockProducer, mockRepo)

	incomingMsg := &sarama.ConsumerMessage{
		Value: []byte(`{"OrderID":1,"OrderType":"type1","UserId":"1","ItemID":1,"Amount":1000}`),
	}

	expectedMessage := domain.Message{
		OrderID:   1,
		OrderType: "type1",
		UserId:    "1",
		ItemID:    1,
		Amount:    1000,
	}

	mockRepo.On("GetNextTopic", &expectedMessage, false, mock.Anything).Return("next_topic")
	mockProducer.On("SendMessage", mock.Anything).Return(int32(0), int64(0), nil)
	mockRepo.On("OrchesLog", mock.Anything, mock.Anything).Return(nil)

	err := uc.ViewOrchesSteps(context.Background(), incomingMsg)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockProducer.AssertExpectations(t)
}

// Test ViewOrchesSteps - Negative Case (Retry Scenario)
func TestViewOrchesSteps_Retry(t *testing.T) {
	mockRepo := new(MockOrchestratorRepo)
	mockProducer := new(MockProducer)
	uc := usecase.NewOrchestratorUsecase(mockProducer, mockRepo)

	incomingMsg := &sarama.ConsumerMessage{
		Value: []byte(`{"OrderID":1,"OrderType":"type1","UserId":"1","ItemID":1,"Amount":1000,"RespCode":500,"Retry":2}`),
	}

	expectedMessage := domain.Message{
		OrderID:   1,
		OrderType: "type1",
		UserId:    "1",
		ItemID:    1,
		Amount:    1000,
		RespCode:  500,
		Retry:     3,
	}

	mockRepo.On("GetNextTopic", &expectedMessage, true, mock.Anything).Return("retry_topic")
	mockProducer.On("SendMessage", mock.Anything).Return(int32(0), int64(0), nil)
	mockRepo.On("OrchesLog", mock.Anything, mock.Anything).Return(nil)

	err := uc.ViewOrchesSteps(context.Background(), incomingMsg)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockProducer.AssertExpectations(t)
}

// Test ViewOrchesSteps - Negative Case (Rollback Scenario)
func TestViewOrchesSteps_Negative_Rollback(t *testing.T) {
	mockRepo := new(MockOrchestratorRepo)
	mockProducer := new(MockProducer)
	uc := usecase.NewOrchestratorUsecase(mockProducer, mockRepo)

	incomingMsg := &sarama.ConsumerMessage{
		Value: []byte(`{"OrderID":1,"OrderType":"type1","UserId":"1","ItemID":1,"Amount":1000,"RespCode":400}`),
	}

	expectedMessage := domain.Message{
		OrderID:   1,
		OrderType: "type1",
		UserId:    "1",
		ItemID:    1,
		Amount:    1000,
		RespCode:  400,
	}

	mockRepo.On("GetRollbackTopic", &expectedMessage, mock.Anything).Return("rollback_topic")
	mockProducer.On("SendMessage", mock.Anything).Return(int32(0), int64(0), nil)
	mockRepo.On("OrchesLog", mock.Anything, mock.Anything).Return(nil)

	err := uc.ViewOrchesSteps(context.Background(), incomingMsg)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockProducer.AssertExpectations(t)
}

// Test OrchesLog - Positive Case
func TestOrchesLog_Positive(t *testing.T) {
	mockRepo := new(MockOrchestratorRepo)
	mockProducer := new(MockProducer)
	uc := usecase.NewOrchestratorUsecase(mockProducer, mockRepo)

	message := domain.Message{
		OrderID:   1,
		OrderType: "type1",
		UserId:    "1",
		ItemID:    1,
		Amount:    1000,
	}

	mockRepo.On("OrchesLog", message, mock.Anything).Return(nil)

	err := uc.OrchesLog(message, context.Background())

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// Test OrchesLog - Negative Case
func TestOrchesLog_Negative(t *testing.T) {
	mockRepo := new(MockOrchestratorRepo)
	mockProducer := new(MockProducer)
	uc := usecase.NewOrchestratorUsecase(mockProducer, mockRepo)

	message := domain.Message{
		OrderID:   1,
		OrderType: "type1",
		UserId:    "1",
		ItemID:    1,
		Amount:    1000,
	}

	mockRepo.On("OrchesLog", message, mock.Anything).Return(errors.New("logging failed"))

	err := uc.OrchesLog(message, context.Background())

	assert.Error(t, err)
	assert.EqualError(t, err, "logging failed")
	mockRepo.AssertExpectations(t)
}
