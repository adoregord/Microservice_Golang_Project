package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
	"user_microservice/internal/domain"

	"log"

	"github.com/IBM/sarama"
)

type UserUsecaseInterface interface {
	ValidateUser
	SendMessage
	SetHTTPClient
}
type ValidateUser interface {
	ValidateUser(incoming *domain.Message) error
}
type SendMessage interface {
	SendMessage(ctx context.Context, msg *sarama.ConsumerMessage) error
}
type SetHTTPClient interface {
	SetHTTPClient(client *http.Client)
}

type UserUsecase struct {
	producer   sarama.SyncProducer
	httpClient *http.Client
}

func NewUserUsecase(producer sarama.SyncProducer) UserUsecaseInterface {
	return &UserUsecase{
		producer:   producer,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (uc *UserUsecase) SetHTTPClient(client *http.Client) {
	uc.httpClient = client
}

func (uc *UserUsecase) ValidateUser(incoming *domain.Message) error {
	url := "https://7a70c146-33cd-4f4b-9b33-38cc4824afa0.mock.pstmn.io/chekuser"

	payload := domain.User{
		Username: incoming.UserId,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return errors.New("FAILED at user validation: error marshaling payload")
	}

	log.Println("Now hitting outbound service to validate user")
	//response, err := http.Post(url, "application/json", bytes.NewBuffer(payloadBytes))
	response, err := uc.httpClient.Post(url, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return errors.New("FAILED at user validation: error during POST request")
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return errors.New("FAILED at user validation: error reading response")
	}

	var responseStruct struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &responseStruct); err != nil {
		return errors.New("FAILED at user validation: error unmarshaling response")
	}

	if responseStruct.Status != "OK" {
		incoming.RespCode = 404
		incoming.RespStatus = responseStruct.Status
		incoming.RespMessage = fmt.Sprintf("FAILED %s", responseStruct.Message)
		return nil
	}

	incoming.RespCode = 200
	incoming.RespStatus = "OK"
	incoming.RespMessage = "SUCCESS: user is verified"

	return nil
}

func (uc *UserUsecase) SendMessage(ctx context.Context, msg *sarama.ConsumerMessage) error {
	topic := "orchestrator_topic"

	var incoming domain.Message
	if err := json.Unmarshal(msg.Value, &incoming); err != nil {
		return err
	}

	incoming.OrderService = "Verify User"

	err := uc.ValidateUser(&incoming)
	if err != nil {
		incoming.RespCode = 500
		incoming.RespStatus = "Internal Server Error"
		incoming.RespMessage = fmt.Sprintf("FAILED USER VERIFY: %v", err)
	}

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
