package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

type ValidateUser interface {
	ValidateUser(incoming *domain.Message) error
}
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

func (uc UserUsecase) ValidateUser(incoming *domain.Message) error {
	type responseStk struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	payload := domain.User{
		Username: incoming.UserId,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return errors.New("FAILED at user validation")
	}

	// Make the POST request to hit outbound request
	log.Println("Now hitting outbound service to validate user")
	response, err := http.Post("https://7a70c146-33cd-4f4b-9b33-38cc4824afa0.mock.pstmn.io/chekuser", "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return errors.New("FAILED at user validation")
	}
	defer response.Body.Close()

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return errors.New("FAILED at user validation")
	}

	// Unmarshal the response JSON into responseStk struct
	var responseStruct responseStk
	err = json.Unmarshal(body, &responseStruct)
	if err != nil {
		return errors.New("FAILED at user validation")
	}

	if responseStruct.Status != "OK" {
		incoming.RespCode = 404
		incoming.RespStatus = responseStruct.Status
		incoming.RespMessage = fmt.Sprintf("FAILED %s", responseStruct.Message)
		return nil
	}
	// update total price
	incoming.RespCode = 200
	incoming.RespStatus = "OK"
	incoming.RespMessage = "SUCCESS user is verified"

	return nil
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

	err := uc.ValidateUser(&incoming)
	if err != nil {
		incoming.RespCode = 500
		incoming.RespStatus = "Internal Server Error"
		incoming.RespMessage = fmt.Sprintf("FAILED USER VERIFY: %v", err)
		return nil
	}

	return nil
}
