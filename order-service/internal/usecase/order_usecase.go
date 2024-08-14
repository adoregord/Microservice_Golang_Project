package usecase

import (
	"context"
	"encoding/json"
	"order_microservice/internal/domain"
	"order_microservice/internal/repository"
	"strconv"

	"log"

	"github.com/IBM/sarama"
)

type OrderUsecaseInterface interface {
	OrderInit
	UpdateOrderDetails
}

type OrderInit interface {
	OrderInit(orderReq *domain.OrderRequest, kontek context.Context) (*domain.Message, error)
}
type UpdateOrderDetails interface {
	UpdateOrderDetails(msg *sarama.ConsumerMessage, kontek context.Context) error
}

type OrderUsecase struct {
	repo     repository.OrderRepoInterface
	Producer sarama.SyncProducer
}

func NewOrderUsecase(Repo repository.OrderRepoInterface, producer sarama.SyncProducer) OrderUsecaseInterface {
	return OrderUsecase{
		repo:     Repo,
		Producer: producer,
	}
}

func (uc OrderUsecase) OrderInit(orderReq *domain.OrderRequest, kontek context.Context) (*domain.Message, error) {
	topic := "orchestrator_topic" // will send to this topic always

	// save to db first
	message, err := uc.repo.StoreOrderDetails(*orderReq, kontek)
	if err != nil {
		return nil, err
	}
	// set message order service to order init
	message.OrderService = "Order Init"

	// Marshall the incoming message struct to JSON
	msgValue, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal incoming message: %s\n", err)
		return nil, err
	}

	//change to string
	key := strconv.Itoa(orderReq.ID)
	// Send the message to Kafka
	_, _, err = uc.Producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(msgValue),
	})

	if err != nil {
		log.Printf("Failed to send message: %s\n", err)
		return nil, err
	}

	log.Printf("Message sent to %s with data %v", topic, message)
	return message, nil
}

func (uc OrderUsecase) UpdateOrderDetails(msg *sarama.ConsumerMessage, kontek context.Context) error {
	// Parse the incoming message
	var incoming_message domain.Message
	if err := json.Unmarshal(msg.Value, &incoming_message); err != nil {
		return err
	}

	return uc.repo.UpdateOrderDetails(incoming_message, kontek)
}