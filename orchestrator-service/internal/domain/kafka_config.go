package domain

import "github.com/IBM/sarama"

type KafkaConfig struct {
	GroupID string `json:"omitempty"`
	Topic   string
	Brokers []string
}

type KafkaConsumer struct {
	Consumer sarama.ConsumerGroup
	Handler  sarama.ConsumerGroupHandler
	Topics   []string
}
