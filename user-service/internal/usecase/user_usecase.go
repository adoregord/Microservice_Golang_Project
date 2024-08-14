package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"user_microservice/internal/domain"

	"log"

	"github.com/IBM/sarama"
)

type UserUsecaseInterface interface {
	ValidateUser
	SendMessage
}

type ValidateUser interface{ ValidateUser(username string) bool }
type SendMessage interface {
	SendMessage(kontek context.Context, msg *sarama.ConsumerMessage) error
}

type UserUsecase struct {
	Producer sarama.SyncProducer
}

func NewUserUsecase(producer sarama.SyncProducer) UserUsecaseInterface {
	return &UserUsecase{
		Producer: producer,
	}
}

func (uc UserUsecase) ValidateUser(username string) bool {
	type responseStk struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	payload := domain.User{
		Username: username,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return false
	}

	// Make the POST request to hit outbound request
	log.Println("Now hitting outbound service to validate user")
	response, err := http.Post("https://7a70c146-33cd-4f4b-9b33-38cc4824afa0.mock.pstmn.io/chekuser", "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return false
	}
	defer response.Body.Close()

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return false
	}

	// Unmarshal the response JSON into responseStk struct
	var responseStruct responseStk
	err = json.Unmarshal(body, &responseStruct)
	if err != nil {
		return false
	}

	// Check if the status is success
	if responseStruct.Status == "ok" {
		return true
	}

	return false
}

func (uc UserUsecase) SendMessage(kontek context.Context, msg *sarama.ConsumerMessage) error {

	topic := "orchestrator_topic" // always send to this topic

	// Parse the incoming message
	var incoming domain.Message
	if err := json.Unmarshal(msg.Value, &incoming); err != nil {
		return err
	}

	// set the service so orches know where the message came from
	incoming.OrderService = "Verify User"

	// call usecase to validate user
	ok := uc.ValidateUser(incoming.UserId)
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

	if !ok {
		incoming.RespCode = 401
		incoming.RespStatus = "Unauthorized "
		incoming.RespMessage = "FAILED User is not Valid"
	} else {
		incoming.RespCode = 200
		incoming.RespStatus = "ok"
		incoming.RespMessage = "SUCCESS User is Valid"
	}

	return nil
}
