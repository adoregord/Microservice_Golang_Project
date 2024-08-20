package usecase

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"log"
	"microservice_orchestrator/internal/domain"
	"microservice_orchestrator/internal/repository"
)

type OrchestratorUsecaseInterface interface {
	ViewOrchesSteps
	OrchesLog
}

type ViewOrchesSteps interface {
	ViewOrchesSteps(kontek context.Context, msg *sarama.ConsumerMessage) error
}
type OrchesLog interface {
	OrchesLog(message domain.Message, kontek context.Context) error
}

type OrchestratorUsecase struct {
	Producer sarama.SyncProducer
	Repo     repository.OrchestratorRepoInterface
}

func NewOrchestratorUsecase(producer sarama.SyncProducer, repo repository.OrchestratorRepoInterface) OrchestratorUsecaseInterface {
	return &OrchestratorUsecase{
		Producer: producer,
		Repo:     repo,
	}
}

func (uc OrchestratorUsecase) ViewOrchesSteps(kontek context.Context, msg *sarama.ConsumerMessage) error {
	// Parse the incoming message
	var incoming_message domain.Message
	if err := json.Unmarshal(msg.Value, &incoming_message); err != nil {

		return err
	}

	// log the orches steps
	defer func() {
		err := uc.OrchesLog(incoming_message, kontek)
		if err != nil {
			log.Println(err)
		}
	}()

	var topic string
	rollbackTriggered := false

	switch {
	// Condition for 5xx errors: server errors
	case incoming_message.RespCode >= 500:
		if incoming_message.Retry < 3 {
			log.Println("Retrying message for order: ", incoming_message.OrderID)
			incoming_message.Retry++
			topic = uc.Repo.GetNextTopic(&incoming_message, true, kontek)
		} else {
			// Rollback after maximum retries
			log.Println("Maximum retries reached for order: ", incoming_message.OrderID)
			log.Println("Triggering rollback")
			topic = uc.Repo.GetRollbackTopic(&incoming_message, kontek)
			rollbackTriggered = true
		}

	// Condition for 4xx errors: client errors
	case incoming_message.RespCode >= 400 && incoming_message.RespCode < 500:
		// Immediately trigger rollback without retry
		log.Printf("Order Failed for order: %d, because: %s", incoming_message.OrderID, incoming_message.RespMessage)
		topic = uc.Repo.GetRollbackTopic(&incoming_message, kontek)
		if topic != "-" {
			rollbackTriggered = true
		}

	// Normal processing for successful responses
	default:
		topic = uc.Repo.GetNextTopic(&incoming_message, false, kontek)
	}

	// Send message to the determined topic
	if topic != "-" {
		err := uc.sendMessage(kontek, topic, msg.Key, &incoming_message)
		if err != nil {
			return err
		}
	}

	// If rollback was triggered, ensure we send a message to order_topic as well
	if rollbackTriggered || topic == "order_topic" || incoming_message.RespCode >= 400 {
		if err := uc.sendMessage(kontek, "order_topic", msg.Key, &incoming_message); err != nil {
			return err
		}
	}
	return nil
}

func (uc OrchestratorUsecase) OrchesLog(message domain.Message, kontek context.Context) error {
	return uc.Repo.OrchesLog(message, kontek)
}

func (uc OrchestratorUsecase) sendMessage(kontek context.Context, topic string, key []byte, incoming_message *domain.Message) error {
	responseBytes, err := json.Marshal(incoming_message)
	if err != nil {
		return err
	}

	_, _, err = uc.Producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(responseBytes),
	})
	if err != nil {
		return err
	}

	log.Printf("Message sent to %s: %s\n\n", topic, string(responseBytes))
	return nil
}
