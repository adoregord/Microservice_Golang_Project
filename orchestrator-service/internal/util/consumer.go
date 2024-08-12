package util

import (
	"context"
	"log"
	"microservice_orchestrator/internal/domain"
	"microservice_orchestrator/internal/kafka"
	"sync"
)

// StartConsumer starts consuming messages from the Kafka topic
func StartConsumer() {
	kafkaConfig := domain.KafkaConfig{
		Brokers: []string{"127.0.0.1:29092"},
		GroupID: "Orchestrator_Kafka_Consumer",
		Topic:   "order_topic1", // topic that want to be listened
	}

	// setup kafka producer for kafka for user and package
	producer, err := kafka.NewKafkaProducer(kafkaConfig.Brokers)
	if err != nil {
		log.Printf("Failed to produce new producer for kafka: %s\n", err)
		return
	}
	defer producer.Close()

	handler := kafka.NewMessageHandler(producer)

	consumer, err := kafka.NewKafkaConsumer(kafkaConfig.Brokers, kafkaConfig.GroupID, []string{kafkaConfig.Topic}, handler)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer consumer.Close()

	log.Println("Listening for messages...")

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		// read message from topic
		defer wg.Done()
		if err := consumer.Consume(context.Background()); err != nil {
			log.Printf("Error from consumer: %v", err)
		}
	}()
	wg.Wait()
}
