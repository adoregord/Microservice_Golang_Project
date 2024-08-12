package main

import (
	"context"
	"os"
	"os/signal"
	provider "user_microservice/internal/provider/db"
	"user_microservice/internal/provider/initialize"
	"user_microservice/internal/repository"
	"user_microservice/internal/usecase"

	"github.com/rs/zerolog/log"

	"syscall"
)

func main() {
	database, err := provider.DBConnection()
	if err != nil {
		log.Fatal().Err(err)
	}
	defer database.Close()

	userRepo := repository.NewUserRepo(database)
	userUsecase := usecase.NewUserUsecase(userRepo)

	initialize.StartConsumeFromOrder(userUsecase)
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	<-sigchan

	// log.Println("Received termination signal. Initiating shutdown...")
	log.Print("Received termination signal. Initiating shutdown...")
	cancel()

	log.Print("Consumer has shut down")
}
