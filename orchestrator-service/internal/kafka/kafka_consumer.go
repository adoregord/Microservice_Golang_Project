package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"microservice_orchestrator/internal/domain"

	"github.com/rs/zerolog/log"

	"github.com/IBM/sarama"
)

type KafkaConsumer struct {
	consumer sarama.ConsumerGroup
	handler  sarama.ConsumerGroupHandler
	topics   []string
}

type MessageHandler struct {
	Producer sarama.SyncProducer
}

func NewMessageHandler(producer sarama.SyncProducer) *MessageHandler {
	return &MessageHandler{
		Producer: producer,
	}
}

func (h MessageHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h MessageHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h MessageHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	if h.Producer == nil {
		log.Info().Msg("Producer is nil! Make sure it's initialized properly.")
		return errors.New("kafka producer is not initialized")
	}

	for msg := range claim.Messages() {
		sess.MarkMessage(msg, "")

		// Parse the incoming message
		var incoming domain.IncomingMessage
		if err := json.Unmarshal(msg.Value, &incoming); err != nil {
			log.Printf("Error parsing message: %v\n", err)
			continue
		}

		switch incoming.OrderType {
		case "Buy Package":
			if incoming.OrderService == "Order Init" {
				incoming.RespStatus = "On-Going"

				responseBytes, err := json.Marshal(incoming)
				if err != nil {
					log.Printf("Failed to marshal message: %v\n\n", err)
					continue
				}

				_, _, err = h.Producer.SendMessage(&sarama.ProducerMessage{
					Topic: "user_topic_validate",
					Key:   sarama.ByteEncoder(msg.Key),
					Value: sarama.ByteEncoder(responseBytes),
				})
				if err != nil {
					log.Info().Msg(fmt.Sprintf("Error writing message to topic_validateUser: %v\n", err))
					continue
				}
				log.Info().Msg(fmt.Sprintf("Message sent to topic_validate_user: %s\n\n", string(responseBytes)))

			} else if incoming.OrderService == "Verify User" {

				responseBytes, err := json.Marshal(incoming)
				if err != nil {
					log.Info().Msg(fmt.Sprintf("Failed to marshal message: %v\n", err))
					continue
				}

				if incoming.RespCode == 400 || incoming.RespCode == 401 {
					// update to db set the
					log.Info().Msg(string(responseBytes))
					incoming.RespMessage = "FAILED, user is not valid or verified"
					continue
				}

				_, _, err = h.Producer.SendMessage(&sarama.ProducerMessage{
					Topic: "topic_activate_package",
					Key:   sarama.ByteEncoder(msg.Key),
					Value: sarama.ByteEncoder(responseBytes),
				})
				if err != nil {
					log.Info().Str("Error writing message to topic_activatePackage:", err.Error())
					continue
				}
				log.Info().Str("Message sent to topic_activatePackage: ", string(responseBytes))

			} else if incoming.OrderService == "activatePackage" {
				// Log the completion message
				log.Info().Msg("===============================================================================================")
				log.Info().Msg(fmt.Sprintf("Transaction ID %s for order type '%s' is COMPLETED\n", incoming.TransactionId, incoming.OrderType))
				log.Info().Msg("===============================================================================================")
			}
		default:
			log.Info().Msg(fmt.Sprintf("Received unsupported message format: %v\n\n", incoming))
		}
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
