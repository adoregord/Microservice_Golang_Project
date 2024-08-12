package kafka

import (
	"context"
	"errors"
	"log"
	"microservice_orchestrator/internal/usecase"

	"github.com/IBM/sarama"
)

type KafkaConsumer struct {
	consumer sarama.ConsumerGroup
	handler  sarama.ConsumerGroupHandler
	topics   []string
}

type MessageHandler struct {
	Producer sarama.SyncProducer
	usecase usecase.OrchestratorUsecaseInterface
}

func NewMessageHandler(producer sarama.SyncProducer, usecase usecase.OrchestratorUsecaseInterface) *MessageHandler {
	return &MessageHandler{
		Producer: producer,
		usecase: usecase,
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
		err := h.usecase.ViewOrchesSteps(context.Background(), msg)
		if err != nil {
			return err
		}
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

// switch incoming.OrderType {
// case "Buy Package":
// 	if incoming.OrderService == "Order Init" {
// 		incoming.RespStatus = "On-Going"

// 		responseBytes, err := json.Marshal(incoming)
// 		if err != nil {
// 			log.Printf("Failed to marshal message: %v\n\n", err)
// 			continue
// 		}

// 		_, _, err = h.Producer.SendMessage(&sarama.ProducerMessage{
// 			Topic: "user_topic_validate",
// 			Key:   sarama.ByteEncoder(msg.Key),
// 			Value: sarama.ByteEncoder(responseBytes),
// 		})
// 		if err != nil {
// 			log.Printf("Error writing message to topic_validateUser: %v\n", err)
// 			continue
// 		}
// 		log.Printf("Message sent to topic_validate_user: %s\n\n", string(responseBytes))

// 	} else if incoming.OrderService == "Verify User" {

// 		responseBytes, err := json.Marshal(incoming)
// 		if err != nil {
// 			log.Printf("Failed to marshal message: %v\n", err)
// 			continue
// 		}

// 		if incoming.RespCode == 400 || incoming.RespCode == 401{
// 			// update to db set the
// 			log.Println(string(responseBytes))
// 			incoming.RespMessage = "FAILED, user is not valid or verified"
// 			continue
// 		}

// 		_, _, err = h.Producer.SendMessage(&sarama.ProducerMessage{
// 			Topic: "topic_activate_package",
// 			Key:   sarama.ByteEncoder(msg.Key),
// 			Value: sarama.ByteEncoder(responseBytes),
// 		})
// 		if err != nil {
// 			log.Printf("Error writing message to topic_activatePackage: %v\n", err)
// 			continue
// 		}
// 		log.Printf("Message sent to topic_activatePackage: %s\n\n", string(responseBytes))

// 	} else if incoming.OrderService == "activatePackage" {
// 		// Log the completion message
// 		log.Printf("===============================================================================================")
// 		log.Printf("Transaction ID %s for order type '%s' is COMPLETED\n", incoming.TransactionId, incoming.OrderType)
// 		log.Printf("===============================================================================================")
// 	}
// default:
// 	log.Printf("Received unsupported message format: %v\n\n", incoming)
// }
