package initialize

import (
	"context"
	"database/sql"
	"log"
	"order_microservice/internal/domain"
	"order_microservice/internal/handler"
	"order_microservice/internal/kafka"
	"order_microservice/internal/provider/routes"
	"order_microservice/internal/repository"
	"order_microservice/internal/usecase"
	"sync"
)

// StartConsumer starts consuming messages from the Kafka topic
func StartConsumer(db *sql.DB) {
	kafkaConfig := domain.KafkaConfig{
		Brokers: []string{"127.0.0.1:29092"},
		GroupID: "Order_Kafka_Consumer",
		Topic:   "order_topic", // topic that want to be listened
	}

	// setup kafka producer for kafka for user and package
	producer, err := kafka.NewKafkaProducer(kafkaConfig.Brokers)
	if err != nil {
		log.Printf("Failed to produce new producer for kafka: %s\n", err)
		return
	}
	defer producer.Close()

	repo := repository.NewOrderRepo(db)
	uc := usecase.NewOrderUsecase(repo, producer)
	h := handler.NewOrderHandler(uc)
	routes.SetupRoutes(h).Run("127.0.0.1:8081")

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
