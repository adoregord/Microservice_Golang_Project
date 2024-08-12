package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"user_microservice/internal/domain"
	"user_microservice/internal/usecase"

	"github.com/IBM/sarama"
)

type KafkaConsumer struct {
	consumer sarama.ConsumerGroup
	handler  sarama.ConsumerGroupHandler
	topics   []string
}

type MessageHandler struct {
	Producer    sarama.SyncProducer
	UserUsecase *usecase.UserUsecase
}

func NewMessageHandler(producer sarama.SyncProducer, userUsecase *usecase.UserUsecase) *MessageHandler {
	return &MessageHandler{
		Producer:    producer,
		UserUsecase: userUsecase,
	}
}

func (h MessageHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h MessageHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h MessageHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	if h.Producer == nil {
		log.Println("Producer is nil! Make sure it's initialized properly.")
		return errors.New("kafka producer is not initialized")
	}

	for msg := range claim.Messages() {
		sess.MarkMessage(msg, "")

		// Parse the incoming message
		var incoming domain.IncomingMessage
		if err := json.Unmarshal(msg.Value, &incoming); err != nil {
			log.Printf("Error parsing message: %v\n\n", err)
			continue
		}
		switch incoming.OrderType {
		case "Buy Package":
			// set the service so orches know where the message came from
			incoming.OrderService = "Verify User"
			// call usecase to validate user
			ok := h.UserUsecase.ValidateUser(incoming.UserId)
			if !ok {
				incoming.RespCode = 400
				incoming.RespStatus = "Bad Request"
				incoming.RespMessage = "User is not Valid"

				responseBytes, err := json.Marshal(incoming)
				if err != nil {
					log.Printf("Failed to marshal message: %v\n\n", err)
					continue
				}

				_, _, err = h.Producer.SendMessage(&sarama.ProducerMessage{
					Topic: "order_topic1",
					Key:   sarama.ByteEncoder(msg.Key),
					Value: sarama.ByteEncoder(responseBytes),
				})
				if err != nil {
					log.Printf("Error writing message to topic_order: %v\n\n", err)
					continue
				}
				log.Printf("Message sent to topic_validate_user: %s\n\n", string(responseBytes))

			} else {
				incoming.RespCode = 200
				incoming.RespStatus = "ok"
				incoming.RespMessage = "User is Valid"

				responseBytes, err := json.Marshal(incoming)
				if err != nil {
					log.Printf("Failed to marshal message: %v\n\n", err)
					continue
				}

				_, _, err = h.Producer.SendMessage(&sarama.ProducerMessage{
					Topic: "order_topic1",
					Key:   sarama.ByteEncoder(msg.Key),
					Value: sarama.ByteEncoder(responseBytes),
				})
				if err != nil {
					log.Printf("Error writing message to topic_validateUser: %v\n\n", err)
					continue
				}
				log.Printf("Message sent to topic_validate_user: %s\n\n", string(responseBytes))
			}

		default:
			fmt.Println("salah awokwok")
		}
		
		fmt.Printf("Received message: %+v\n", incoming)
		sess.MarkMessage(msg, "")
	}
	return nil
}

func NewKafkaConsumer(brokers []string, groupID string, topics []string, msg *MessageHandler) (*KafkaConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &KafkaConsumer{
		consumer: consumer,
		topics:   topics,
		handler:  msg,
	}, nil
}

func (kc *KafkaConsumer) Consume(ctx context.Context) error {
	for {
		err := kc.consumer.Consume(ctx, kc.topics, kc.handler)
		if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

func (kc *KafkaConsumer) Close() error {
	return kc.consumer.Close()
}

func (kc *KafkaConsumer) SetHandler(handler sarama.ConsumerGroupHandler) {
	kc.handler = handler
}
