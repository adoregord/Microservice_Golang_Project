package usecase

import (
	"context"
	"delivery_microservice/internal/domain"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"io"
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
	topic := "orchestrator_topic" // always send to this topic

	// Parse the incoming message
	var incoming domain.Message
	if err := json.Unmarshal(msg.Value, &incoming); err != nil {
		return err
	}

	// set the service so orches know where the message came from
	incoming.OrderService = "Delivery Service"

	defer func() {
		responseBytes, err := json.Marshal(incoming)
		if err != nil {
			log.Println(err)
		}
		_, _, err = uc.Producer.SendMessage(&sarama.ProducerMessage{
			Topic: topic,
			Key:   sarama.ByteEncoder(msg.Key),
			Value: sarama.ByteEncoder(responseBytes),
		})
		if err != nil {
			log.Println(err)
		}
		log.Printf("Message sent to %s: %s\n\n", topic, string(responseBytes))
	}()

	// call function to validate item
	err := uc.Delivery(&incoming)
	if err != nil {
		incoming.RespCode = 500
		incoming.RespStatus = "Internal Server Error"
		incoming.RespMessage = fmt.Sprintf("FAILED DELIVERY: %v", err)
		return nil
	}

	return nil
}

func (uc DeliveryUsecase) Delivery(incoming *domain.Message) error {

	url := fmt.Sprintf("https://delivery.free.beeceptor.com/delivery/?userID=%s", incoming.UserId)

	// Melakukan request GET
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	//Membaca response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Memetakan JSON ke struct
	var response domain.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	incoming.RespCode = 200
	incoming.RespStatus = "OK"
	incoming.RespMessage = "delivery SUCCESS"

	return nil
}
