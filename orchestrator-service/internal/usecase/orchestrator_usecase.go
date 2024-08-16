package usecase

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"log"
	"microservice_orchestrator/internal/domain"
	"microservice_orchestrator/internal/repository"
	"strings"
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
	log.Println("ini Topik: ", msg.Topic)
	if err := json.Unmarshal(msg.Value, &incoming_message); err != nil {
		return err
	}

	defer func() {
		err := uc.OrchesLog(incoming_message, kontek)
		if err != nil {
			log.Println(err)
		}
	}()

	var topic string
	if strings.Contains(incoming_message.RespMessage, "FAILED") {
		if incoming_message.Retry < 3 {
			incoming_message.Retry += 1
			topic = uc.Repo.ViewOrchesFailStep(&incoming_message, kontek)

		} else {
			topic = "order_topic"
		}
		responseBytes, err := json.Marshal(incoming_message)
		if err != nil {
			return err
		}
		_, _, err = uc.Producer.SendMessage(&sarama.ProducerMessage{
			Topic: topic,
			Key:   sarama.ByteEncoder(msg.Key),
			Value: sarama.ByteEncoder(responseBytes),
		})
		if err != nil {
			return err
		}
		log.Printf("Message sent to %s: %s\n\n", topic, string(responseBytes))
		return nil
	}
	topic = uc.Repo.ViewOrchesSteps(&incoming_message, kontek)
	// check if incoming message contains FAILED then it will be automatically go back to order

	responseBytes, err := json.Marshal(incoming_message)
	if err != nil {
		return err
	}

	_, _, err = uc.Producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(msg.Key),
		Value: sarama.ByteEncoder(responseBytes),
	})
	if err != nil {
		return err
	}
	log.Printf("Message sent to %s: %s\n\n", topic, string(responseBytes))

	return nil
}

func (uc OrchestratorUsecase) OrchesLog(message domain.Message, kontek context.Context) error {
	return uc.Repo.OrchesLog(message, kontek)
}
