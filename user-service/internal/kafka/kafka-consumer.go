package kafka

import (
	"context"
	"errors"
	"log"
	"user_microservice/internal/usecase"

	"github.com/IBM/sarama"
)

type KafkaConsumer struct {
	consumer sarama.ConsumerGroup
	handler  sarama.ConsumerGroupHandler
	topics   []string
}

type MessageHandler struct {
	Producer sarama.SyncProducer
	usecase  usecase.UserUsecaseInterface
}

func NewMessageHandler(producer sarama.SyncProducer, Usecase usecase.UserUsecaseInterface) *MessageHandler {
	return &MessageHandler{
		Producer: producer,
		usecase:  Usecase,
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
		if err := h.usecase.SendMessage(context.Background(), msg); err != nil {
			log.Print(err)
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
