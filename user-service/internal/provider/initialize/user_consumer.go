package initialize

import (
	"context"
	"log"
	"user_microservice/internal/domain"
	"user_microservice/internal/kafka"
	"user_microservice/internal/usecase"

	"sync"
)

// StartConsume for consuming the message that were sent to topic 'user_topic_validate'
func StartConsume() {
	kafkaConfig := domain.KafkaConfig{
		Brokers: []string{"127.0.0.1:29092"}, // My Kafka broker address
		GroupID: "user-consumer-group",
		Topic:   "user_topic_validate", // message that want to be listened
	}

	// setup kafka producer for kafka for user and package
	producer, err := kafka.NewKafkaProducer(kafkaConfig.Brokers)
	if err != nil {
		log.Printf("Failed to produce new producer for kafka: %s\n", err)
		return
	}
	defer producer.Close()

	uc := usecase.NewUserUsecase(producer)
	handler := kafka.NewMessageHandler(producer, uc)

	consumer, err := kafka.NewKafkaConsumer(kafkaConfig.Brokers, kafkaConfig.GroupID, []string{kafkaConfig.Topic}, handler)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Print("Listening for messages...")

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := consumer.Consume(ctx); err != nil {
			log.Printf("Error from consumer: %v", err)
		}
	}()
	wg.Wait()
}
