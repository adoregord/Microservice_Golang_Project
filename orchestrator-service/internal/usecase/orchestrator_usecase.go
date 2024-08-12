package usecase

import (
	"context"
	"encoding/json"
	"log"
	"microservice_orchestrator/internal/domain"
	"microservice_orchestrator/internal/repository"

	"github.com/IBM/sarama"
)

type OrchestratorUsecaseInterface interface {
	ViewOrchesSteps
}

type ViewOrchesSteps interface {
	ViewOrchesSteps(kontek context.Context, msg *sarama.ConsumerMessage) error
}

type OrchestratorUsecase struct{
	Producer sarama.SyncProducer
	OrchestratorRepo repository.OrchestratorRepoInterface
}

func NewOrchestratorUsecase(producer sarama.SyncProducer, orchestratorRepo repository.OrchestratorRepoInterface) OrchestratorUsecaseInterface{
	return &OrchestratorUsecase{
		Producer: producer,
		OrchestratorRepo: orchestratorRepo,
	}
}

func (uc OrchestratorUsecase) ViewOrchesSteps(kontek context.Context, msg *sarama.ConsumerMessage) error {
	// Parse the incoming message
	var incoming_message domain.IncomingMessage
	if err := json.Unmarshal(msg.Value, &incoming_message); err != nil {
		return err
	}
	log.Printf("order type: %s order service %s", incoming_message.OrderType, incoming_message.OrderService)

	topic, err := uc.OrchestratorRepo.ViewOrchesSteps(incoming_message.OrderType, incoming_message.OrderService, kontek)
	log.Printf("topicnya yang dituju ini : %s", topic)
	if err != nil {
		return err
	}

	responseBytes, err := json.Marshal(incoming_message)
	if err != nil {
		log.Printf("Failed to marshal message: %v\n\n", err)
		return err
	}

	_, _, err = uc.Producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(msg.Key),
		Value: sarama.ByteEncoder(responseBytes),
	})
	if err != nil {
		log.Printf("Error writing message to %s: %v\n", topic, err)
		return err
	}
	log.Printf("Message sent to %s: %s\n\n", topic, string(responseBytes))

	return nil
}
