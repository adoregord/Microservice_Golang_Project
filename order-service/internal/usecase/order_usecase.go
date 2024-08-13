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
}

type OrderInit interface {
	OrderInit(orderReq *domain.OrderRequest, kontek context.Context) (*domain.Message, error)
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

// Create a new producer for Kafka
// producer, err := kafka.NewKafkaProducer(kafkaConfig.Brokers)
// if err != nil {
// 	log.Printf("Failed to produce new producer for Kafka: %s\n", err)
// 	return 0, 0, err
// }
// defer producer.Close()
