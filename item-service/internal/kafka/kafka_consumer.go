package kafka

import (
	"context"
	"errors"
	"item_microservice/internal/usecase"
	"log"

	"github.com/IBM/sarama"
)

type KafkaConsumer struct {
	Consumer sarama.ConsumerGroup
	Handler  sarama.ConsumerGroupHandler
	Topics   []string
}

type MessageHandler struct {
	Producer sarama.SyncProducer
	Usecase  usecase.ItemUsecaseInterface
}

func NewMessageHandler(producer sarama.SyncProducer, usecase usecase.ItemUsecaseInterface) *MessageHandler {
	return &MessageHandler{
		Producer: producer,
		Usecase:  usecase,
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
		if msg.Topic == "activation_key_reserve" {
			if err := h.Usecase.SendMessage(context.Background(), msg); err != nil {
				log.Print(err)
			}
		} else if msg.Topic == "activation_key_rollback" {
			if err := h.Usecase.RollbackItem(context.Background(), msg); err != nil {
				log.Print(err)
			}
		} else if msg.Topic == "item_reserve" {
			if err := h.Usecase.SendMessage(context.Background(), msg); err != nil {
				log.Print(err)
			}
		} else if msg.Topic == "item_rollback" {
			if err := h.Usecase.RollbackItem(context.Background(), msg); err != nil {
				log.Print(err)
			}
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
		Consumer: consumer,
		Topics:   topics,
		Handler:  msg,
	}, nil
}

func (kc *KafkaConsumer) Consume(ctx context.Context) error {
	for {
		err := kc.Consumer.Consume(ctx, kc.Topics, kc.Handler)
		if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

func (kc *KafkaConsumer) Close() error {
	return kc.Consumer.Close()
}

//func (kc *KafkaConsumer) SetHandler(handler sarama.ConsumerGroupHandler) {
//	kc.Handler = handler
//}
