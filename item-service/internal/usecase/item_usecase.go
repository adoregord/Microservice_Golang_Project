package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/IBM/sarama"
	"item_microservice/internal/domain"
	"log"
	"net/http"
	"time"
)

type ItemUsecaseInterface interface {
	ValidateItemKey
	ValidateItem
	SendMessage
	RollbackItem
	DeductStock
	ReAddStock
	SetHTTPClient
}
type ValidateItemKey interface {
	ValidateItemKey(incoming *domain.Message) error
}
type ValidateItem interface {
	ValidateItem(incoming *domain.Message) error
}
type SendMessage interface {
	SendMessage(kontek context.Context, msg *sarama.ConsumerMessage) error
}
type RollbackItem interface {
	RollbackItem(kontek context.Context, msg *sarama.ConsumerMessage) error
}
type DeductStock interface {
	DeductStock(id int, amount int) error
}
type ReAddStock interface {
	ReAddStock(id int, amount int) error
}
type SetHTTPClient interface {
	SetHTTPClient(client *http.Client)
}

type ItemUsecase struct {
	Producer   sarama.SyncProducer
	httpClient *http.Client
}

func NewItemUsecase(producer sarama.SyncProducer) ItemUsecaseInterface {
	return &ItemUsecase{
		Producer:   producer,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (uc *ItemUsecase) SetHTTPClient(client *http.Client) {
	uc.httpClient = client
}

func (uc *ItemUsecase) validateItem(url string, incoming *domain.Message, itemType string) error {
	//client := &http.Client{Timeout: 10 * time.Second}
	//resp, err := client.Get(url)
	resp, err := uc.httpClient.Get(url)
	if err != nil {
		return errors.New("FAILED at validate item: " + err.Error())
	}
	defer resp.Body.Close()

	var response domain.Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return errors.New("FAILED to decode item response: " + err.Error())
	}

	incoming.RespCode = resp.StatusCode
	incoming.RespStatus = response.Status

	if resp.StatusCode != http.StatusOK {
		incoming.RespMessage = fmt.Sprintf("FAILED: %s", response.Message)
		return nil
	}

	if response.Data.Quantity < incoming.Amount {
		incoming.RespCode = http.StatusBadRequest
		incoming.RespStatus = "Bad Request"
		incoming.RespMessage = fmt.Sprintf("FAILED: %s stock is not enough", itemType)
		return nil
	}
	// send to outbound to deduct stock item
	uc.DeductStock(incoming.ItemID, incoming.Amount)

	incoming.Total = incoming.Amount * response.Data.Price
	incoming.RespMessage = fmt.Sprintf("SUCCESS: %s is Reserved", itemType)
	return nil
}

func (uc *ItemUsecase) ValidateItemKey(incoming *domain.Message) error {
	url := fmt.Sprintf("http://localhost:3001/item-key/?itemID=%d", incoming.ItemID)
	return uc.validateItem(url, incoming, "Key")
}

func (uc *ItemUsecase) ValidateItem(incoming *domain.Message) error {
	url := fmt.Sprintf("http://localhost:3001/item/?itemID=%d", incoming.ItemID)
	return uc.validateItem(url, incoming, "Item")
}

func (uc *ItemUsecase) DeductStock(id int, amount int) error {
	url := fmt.Sprintf("http://localhost:3001/item-deduct/?itemID=%d/?amount=%d", id, amount)
	client := &http.Client{Timeout: 10 * time.Second}
	resp2, err := client.Get(url)
	if err != nil {
		return errors.New("FAILED at validate item: " + err.Error())
	}
	defer resp2.Body.Close()
	var response2 domain.Response
	if err := json.NewDecoder(resp2.Body).Decode(&response2); err != nil {
		return errors.New("FAILED to decode item response: " + err.Error())
	}
	log.Println("Message: ", response2.Message)
	return nil
}
func (uc *ItemUsecase) ReAddStock(id int, amount int) error {
	url := fmt.Sprintf("http://localhost:3001/item-add/?itemID=%d/?amount=%d", id, amount)
	client := &http.Client{Timeout: 10 * time.Second}
	resp2, err := client.Get(url)
	if err != nil {
		return errors.New("FAILED at validate item: " + err.Error())
	}
	defer resp2.Body.Close()
	var response2 domain.Response
	if err := json.NewDecoder(resp2.Body).Decode(&response2); err != nil {
		return errors.New("FAILED to decode item response: " + err.Error())
	}
	log.Println("Message: ", response2.Message)
	return nil
}

func (uc *ItemUsecase) SendMessage(kontek context.Context, msg *sarama.ConsumerMessage) error {
	var incoming domain.Message
	if err := json.Unmarshal(msg.Value, &incoming); err != nil {
		return errors.New("FAILED to unmarshal incoming message: " + err.Error())
	}

	var validateFunc func(*domain.Message) error

	switch incoming.OrderType {
	case "Buy Activation Key":
		incoming.OrderService = "Reserve Key"
		validateFunc = uc.ValidateItemKey
	case "Buy Item":
		incoming.OrderService = "Reserve Item"
		validateFunc = uc.ValidateItem
	default:
		return errors.New("UNKNOWN order type")
	}

	if err := validateFunc(&incoming); err != nil {
		incoming.RespCode = http.StatusInternalServerError
		incoming.RespStatus = "Internal Server Error"
		incoming.RespMessage = fmt.Sprintf("FAILED CHECK ITEM: %v", err)
	}

	responseBytes, err := json.Marshal(incoming)
	if err != nil {
		return errors.New("FAILED to marshal response message: " + err.Error())
	}

	_, _, err = uc.Producer.SendMessage(&sarama.ProducerMessage{
		Topic: "orchestrator_topic",
		Key:   sarama.ByteEncoder(msg.Key),
		Value: sarama.ByteEncoder(responseBytes),
	})
	if err != nil {
		return errors.New("FAILED to send message to topic orchestrator_topic: " + err.Error())
	}

	log.Printf("Message sent to orchestrator_topic: %s\n", string(responseBytes))
	return nil
}

func (uc *ItemUsecase) RollbackItem(kontek context.Context, msg *sarama.ConsumerMessage) error {
	log.Println("Item is being rolled back")
	var incoming domain.Message
	if err := json.Unmarshal(msg.Value, &incoming); err != nil {
		return errors.New("FAILED to unmarshal incoming message: " + err.Error())
	}
	uc.ReAddStock(incoming.ItemID, incoming.Amount)
	return nil
}
