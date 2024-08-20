package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/IBM/sarama"
	"io"
	"log"
	"net/http"
	"payment_microservice/internal/domain"
	"time"
)

type PaymentUsecaseInterface interface {
	SendMessage
	SetHTTPClient
}
type SendMessage interface {
	SendMessage(ctx context.Context, msg *sarama.ConsumerMessage) error
}
type SetHTTPClient interface {
	SetHTTPClient(client *http.Client)
}

type PaymentUsecase struct {
	producer   sarama.SyncProducer
	httpClient *http.Client
}

func NewPaymentUsecase(producer sarama.SyncProducer) PaymentUsecaseInterface {
	return &PaymentUsecase{
		producer:   producer,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (uc *PaymentUsecase) SetHTTPClient(client *http.Client) {
	uc.httpClient = client
}

func (uc *PaymentUsecase) SendMessage(ctx context.Context, msg *sarama.ConsumerMessage) error {
	topic := "orchestrator_topic"

	var incoming domain.Message
	if err := json.Unmarshal(msg.Value, &incoming); err != nil {
		return err
	}

	// Perform the payment validation
	err := uc.Payment(&incoming)
	if err != nil {
		// Set internal server error response
		incoming.RespCode = 500
		incoming.RespStatus = "Internal Server Error"
		incoming.RespMessage = fmt.Sprintf("FAILED PAYMENT: %v", err)
		incoming.OrderService = "Payment Service"
	}
	incoming.OrderService = "Payment Service"

	// Send the message back to Kafka
	responseBytes, err := json.Marshal(incoming)
	if err != nil {
		log.Printf("Failed to marshal message: %v\n", err)
		return err
	}

	_, _, err = uc.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(msg.Key),
		Value: sarama.ByteEncoder(responseBytes),
	})
	if err != nil {
		log.Printf("Failed to send message to Kafka: %v\n", err)
		return err
	}

	log.Printf("Message sent to %s: %s\n\n", topic, string(responseBytes))
	return nil
}

func (uc *PaymentUsecase) Payment(incoming *domain.Message) error {
	url := fmt.Sprintf("http://localhost:3001/payment/?userID=%s", incoming.UserId)

	// Melakukan request GET
	resp, err := uc.httpClient.Get(url)
	if err != nil {
		return errors.New("FAILED to make payment request")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New("FAILED to read payment response")
	}

	var response domain.Response
	if err := json.Unmarshal(body, &response); err != nil {
		return errors.New("FAILED to unmarshal payment response")
	}

	if response.Status != "OK" {
		incoming.RespCode = 404
		incoming.RespStatus = response.Status
		incoming.RespMessage = fmt.Sprintf("FAILED %s", response.Message)
		return nil
	}

	if response.Data.Balance < incoming.Total {
		incoming.RespCode = 400
		incoming.RespStatus = "Bad Request"
		incoming.RespMessage = "FAILED: user balance is not enough"
		return nil
	}

	incoming.RespCode = 200
	incoming.RespStatus = "OK"
	incoming.RespMessage = "payment SUCCESS"

	return nil
}
