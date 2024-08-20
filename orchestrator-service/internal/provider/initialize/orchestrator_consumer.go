package initialize

import (
	"context"
	"database/sql"
	"log"
	"microservice_orchestrator/internal/domain"
	"microservice_orchestrator/internal/kafka"
	"microservice_orchestrator/internal/repository"
	"microservice_orchestrator/internal/usecase"
	"sync"
)

// StartConsumer starts consuming messages from the Kafka topic
func StartConsumer(db *sql.DB) {
	kafkaConfig := domain.KafkaConfig{
		Brokers: []string{"127.0.0.1:29092"},
		GroupID: "orchestrator_kafka_consumer",
		Topic:   "orchestrator_topic", // topic that want to be listened
	}

	// setup kafka producer for kafka for user and package
	producer, err := kafka.NewKafkaProducer(kafkaConfig.Brokers)
	if err != nil {
		log.Printf("Failed to produce new producer for kafka: %s\n", err)
		return
	}
	defer producer.Close()

	repo := repository.NewOrchestratorRepo(db)
	uc := usecase.NewOrchestratorUsecase(producer, repo)

	handler := kafka.NewMessageHandler(producer, uc)

	consumer, err := kafka.NewKafkaConsumer(kafkaConfig.Brokers, kafkaConfig.GroupID, []string{kafkaConfig.Topic}, handler)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer consumer.Close()

	log.Print("Listening for messages...")

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
