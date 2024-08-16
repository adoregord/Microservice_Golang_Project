package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"payment_microservice/internal/provider/initialize"
	"syscall"
)

func main() {

	initialize.StartConsume()
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
