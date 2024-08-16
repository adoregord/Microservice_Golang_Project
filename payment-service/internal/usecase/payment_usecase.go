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
}
type SendMessage interface {
	SendMessage(kontek context.Context, msg *sarama.ConsumerMessage) error
}

type PaymentUsecase struct {
	Producer sarama.SyncProducer
}

func NewPaymentUsecase(producer sarama.SyncProducer) PaymentUsecaseInterface {
	return &PaymentUsecase{
		Producer: producer,
	}
}
func (uc PaymentUsecase) SendMessage(kontek context.Context, msg *sarama.ConsumerMessage) error {
	topic := "orchestrator_topic" // always send to this topic

	// Parse the incoming message
	var incoming domain.Message
	if err := json.Unmarshal(msg.Value, &incoming); err != nil {
		return err
	}

	// set the service so orches know where the message came from

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
	err := uc.Payment(&incoming)
	if err != nil {
		incoming.RespCode = 500
		incoming.RespStatus = "Internal Server Error"
		incoming.RespMessage = fmt.Sprintf("FAILED PAYMENT: %v", err)
		return nil
	}

	incoming.OrderService = "Payment Service"
	return nil
}

func (uc PaymentUsecase) Payment(incoming *domain.Message) error {

	url := fmt.Sprintf("https://payment.free.beeceptor.com/?userID=%s", incoming.UserId)

	// Melakukan request GET
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return errors.New("FAILED at payment")
	}
	defer resp.Body.Close()

	//Membaca response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New("FAILED at payment")
	}

	// Memetakan JSON ke struct
	var response domain.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return errors.New("FAILED at payment")
	}

	if response.Status != "OK" {
		incoming.RespCode = 404
		incoming.RespStatus = response.Status
		incoming.RespMessage = fmt.Sprintf("FAILED %s", response.Message)
		return nil
	} else if response.Data.Balance < incoming.Total {
		incoming.RespCode = 400
		incoming.RespStatus = "Bad Request"
		incoming.RespMessage = "FAILED user balance is not enough"
		return nil
	}
	incoming.RespCode = 200
	incoming.RespStatus = "OK"
	incoming.RespMessage = "payment SUCCESS"

	return nil
}
