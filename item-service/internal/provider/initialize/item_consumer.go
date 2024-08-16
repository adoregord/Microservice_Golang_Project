package initialize

import (
	"context"
	"item_microservice/internal/domain"
	"item_microservice/internal/kafka"
	"item_microservice/internal/usecase"
	"log"
	"sync"
)

// StartConsume for consuming the message that were sent to topic 'activation_key_reserve'
func StartConsume() {
	kafkaConfig := domain.KafkaConfig{
		Brokers: []string{"127.0.0.1:29092"}, // My Kafka broker address
		GroupID: "item-consumer-group",
		Topic:   "activation_key_reserve", // message that want to be listened
	}

	// setup kafka producer for kafka for user and package
	producer, err := kafka.NewKafkaProducer(kafkaConfig.Brokers)
	if err != nil {
		log.Printf("Failed to produce new producer for kafka: %s\n", err)
		return
	}
	defer producer.Close()

	uc := usecase.NewItemUsecase(producer)
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
