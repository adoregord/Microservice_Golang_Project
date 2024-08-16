package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/IBM/sarama"
	"io"
	"item_microservice/internal/domain"
	"log"
	"net/http"
	"time"
)

type ItemUsecaseInterface interface {
	ValidateItem
	SendMessage
}
type ValidateItem interface {
	ValidateItem(incoming *domain.Message) error
}
type SendMessage interface {
	SendMessage(kontek context.Context, msg *sarama.ConsumerMessage) error
}

type ItemUsecase struct {
	Producer sarama.SyncProducer
}

func NewItemUsecase(producer sarama.SyncProducer) ItemUsecaseInterface {
	return &ItemUsecase{
		Producer: producer,
	}
}

func (uc ItemUsecase) ValidateItem(incoming *domain.Message) error {

	url := fmt.Sprintf("https://item-validation.free.beeceptor.com/?itemID=%d", incoming.ItemID)

	// Melakukan request GET
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return errors.New("FAILED at validate item")
	}
	defer resp.Body.Close()

	//Membaca response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New("FAILED at validate item")
	}

	// Memetakan JSON ke struct
	var response domain.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	if resp.Status != "200 OK" {
		incoming.RespCode = 404
		incoming.RespStatus = response.Status
		incoming.RespMessage = fmt.Sprintf("FAILED %s", response.Message)
		return nil
	} else if response.Data.Amount < incoming.Amount {
		incoming.RespCode = 400
		incoming.RespStatus = "Bad Request"
		incoming.RespMessage = "FAILED item stock is not enough"
		return nil
	}
	// update total price
	incoming.Total = incoming.Amount * response.Data.Price
	incoming.RespCode = 200
	incoming.RespStatus = "OK"
	incoming.RespMessage = "SUCCESS Key is Reserved"

	return nil
}

func (uc ItemUsecase) SendMessage(kontek context.Context, msg *sarama.ConsumerMessage) error {
	topic := "orchestrator_topic" // always send to this topic

	// Parse the incoming message
	var incoming domain.Message
	if err := json.Unmarshal(msg.Value, &incoming); err != nil {
		return err
	}

	// set the service so orches know where the message came from
	incoming.OrderService = "Reserve Key"

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
	err := uc.ValidateItem(&incoming)
	if err != nil {
		incoming.RespCode = 500
		incoming.RespStatus = "Internal Server Error"
		incoming.RespMessage = fmt.Sprintf("FAILED CHECK ITEM: %v", err)
		return nil
	}

	return nil
}
