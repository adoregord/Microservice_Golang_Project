package kafka

import "github.com/IBM/sarama"

// NewKafkaProducer creates and returns a new Kafka producer
func NewKafkaProducer(brokers []string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}
	return producer, nil
}

// SendMessage sends a message to the specified Kafka topic
func SendMessage(producer sarama.SyncProducer, topic string, key, message []byte) (int32, int64, error) {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(message),
	}

	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		return 0, 0, err
	}

	return partition, offset, nil
}
