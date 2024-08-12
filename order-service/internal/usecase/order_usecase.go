package usecase

import (
	"encoding/json"
	"order_microservice/internal/domain"
	"order_microservice/internal/kafka"

	"log"
)

func OrderUsecase(orderReq *domain.OrderRequest) (int32, int64, error) {
	kafkaConfig := domain.KafkaConfig{
		Brokers: []string{"127.0.0.1:29092"},
		Topic:   "order_topic1",
	}

	// Create an IncomingMessage struct
	incomingMsg := domain.IncomingMessage{
		OrderType:     orderReq.OrderType,
		OrderService:  "Order Init",
		TransactionId: orderReq.TransactionID,
		UserId:        orderReq.UserId,
		PackageId:     orderReq.PackageId,
		RespStatus:    "On-Going",
		RespMessage:   "",
		RespCode:      0,
	}

	log.Println("Incoming message: ", incomingMsg)

	// Marshall the incoming message struct to JSON
	msgValue, err := json.Marshal(incomingMsg)
	if err != nil {
		log.Printf("Failed to marshal incoming message: %s\n", err)
		return 0, 0, err
	}

	// Create a new producer for Kafka
	producer, err := kafka.NewKafkaProducer(kafkaConfig.Brokers)
	if err != nil {
		log.Printf("Failed to produce new producer for Kafka: %s\n", err)
		return 0, 0, err
	}
	defer producer.Close()

	// Send the message to Kafka
	partition, offset, err := kafka.SendMessage(producer, kafkaConfig.Topic, []byte(orderReq.TransactionID), msgValue)
	if err != nil {
		log.Printf("Failed to send message: %s\n", err)
		return 0, 0, err
	}

	return partition, offset, nil
}
