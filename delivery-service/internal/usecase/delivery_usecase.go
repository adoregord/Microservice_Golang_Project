package usecase

import (
	"context"
	"delivery_microservice/internal/domain"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/IBM/sarama"
	"log"
	"net/http"
	"time"
)

type DeliveryUsecaseInterface interface {
	SendMessage
}
type SendMessage interface {
	SendMessage(kontek context.Context, msg *sarama.ConsumerMessage) error
}

type DeliveryUsecase struct {
	Producer sarama.SyncProducer
}

func NewDeliveryUsecase(producer sarama.SyncProducer) DeliveryUsecaseInterface {
	return &DeliveryUsecase{Producer: producer}
}

func (uc DeliveryUsecase) SendMessage(kontek context.Context, msg *sarama.ConsumerMessage) error {
	topic := "orchestrator_topic"

	// Parse the incoming message
	var incoming domain.Message
	if err := json.Unmarshal(msg.Value, &incoming); err != nil {
		return errors.New("FAILED to unmarshal incoming message")
	}

	// Set the service name for orchestration tracking
	incoming.OrderService = "Delivery Service"

	// Call the appropriate delivery function based on the order type
	if err := uc.HandleDelivery(&incoming); err != nil {
		incoming.RespCode = 500
		incoming.RespStatus = "Internal Server Error"
		incoming.RespMessage = fmt.Sprintf("FAILED DELIVERY: %v", err)
	}

	// Prepare and send the message to the Kafka topic
	responseBytes, err := json.Marshal(incoming)
	if err != nil {
		return errors.New("FAILED to marshal response message")
	}
	_, _, err = uc.Producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(msg.Key),
		Value: sarama.ByteEncoder(responseBytes),
	})
	if err != nil {
		return errors.New("FAILED to send message to topic")
	}

	log.Printf("Message sent to %s: %s\n", topic, string(responseBytes))
	return nil
}

func (uc DeliveryUsecase) HandleDelivery(incoming *domain.Message) error {
	// Define the URL based on order type
	var url string
	if incoming.OrderType == "Buy Activation Key" {
		url = fmt.Sprintf("https://delivery.free.beeceptor.com/delivery/?userID=%s", incoming.UserId)
	} else if incoming.OrderType == "Buy Item" {
		url = fmt.Sprintf("http://localhost:3001/deliveryitem/?userID=%s", incoming.UserId)
	} else {
		return errors.New("INVALID order type for delivery")
	}

	// Make the GET request with a timeout
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return errors.New("FAILED to make delivery request")
	}
	defer resp.Body.Close()

	// Read and parse the response body
	var response domain.Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return errors.New("FAILED to decode delivery response")
	}

	// Update the incoming message based on the response
	incoming.RespCode = 200
	incoming.RespStatus = "OK"
	incoming.RespMessage = "delivery SUCCESS"

	return nil
}
