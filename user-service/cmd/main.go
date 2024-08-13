package main

import (
	"context"
	"os"
	"os/signal"
	"user_microservice/internal/provider/initialize"

	"log"

	"syscall"
)

func main() {

	initialize.StartConsumeFromOrder()
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
