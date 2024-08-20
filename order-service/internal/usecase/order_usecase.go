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
	OrderRetry
}
type OrderInit interface {
	OrderInit(orderReq *domain.OrderRequest, ctx context.Context) (*domain.Message, error)
}
type UpdateOrderDetails interface {
	UpdateOrderDetails(msg *sarama.ConsumerMessage, ctx context.Context) error
}
type OrderRetry interface {
	OrderRetry(orderRetry *domain.RetryOrder, ctx context.Context) (*domain.Message, error)
}

type OrderUsecase struct {
	repo     repository.OrderRepoInterface
	producer sarama.SyncProducer
}

func NewOrderUsecase(repo repository.OrderRepoInterface, producer sarama.SyncProducer) OrderUsecaseInterface {
	return &OrderUsecase{
		repo:     repo,
		producer: producer,
	}
}

func (uc *OrderUsecase) OrderInit(orderReq *domain.OrderRequest, ctx context.Context) (*domain.Message, error) {
	// Save to db first
	message, err := uc.repo.StoreOrderDetails(*orderReq, ctx)
	if err != nil {
		return nil, err
	}

	message.OrderService = "Order Init"

	// Send the message to Kafka
	if err := uc.sendMessageToKafka(message, orderReq.ID); err != nil {
		return nil, err
	}

	log.Printf("Message sent to orchestrator_topic with data %v", message)
	return message, nil
}

func (uc *OrderUsecase) UpdateOrderDetails(msg *sarama.ConsumerMessage, ctx context.Context) error {
	var incomingMessage domain.Message
	if err := json.Unmarshal(msg.Value, &incomingMessage); err != nil {
		return err
	}

	return uc.repo.UpdateOrderDetails(incomingMessage, ctx)
}

func (uc *OrderUsecase) OrderRetry(orderRetry *domain.RetryOrder, ctx context.Context) (*domain.Message, error) {
	// Save update the db first
	message, err := uc.repo.EditRetryOrder(orderRetry, ctx)
	if err != nil {
		return nil, err
	}

	message.OrderService = "Order Init"
	message.RespMessage = "EDIT and RETRY order"
	message.RespCode = 200

	// Send the message to Kafka
	if err := uc.sendMessageToKafka(message, orderRetry.OrderID); err != nil {
		return nil, err
	}

	log.Printf("Message sent to orchestrator_topic with data %v", message)
	return message, nil
}

func (uc *OrderUsecase) sendMessageToKafka(message *domain.Message, keyID int) error {
	topic := "orchestrator_topic"

	msgValue, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal message: %s\n", err)
		return err
	}

	key := strconv.Itoa(keyID)
	_, _, err = uc.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(msgValue),
	})

	if err != nil {
		log.Printf("Failed to send message to Kafka: %s\n", err)
		return err
	}

	return nil
}
