package main

import (
	"context"
	"delivery_microservice/internal/provider/initialize"
	"log"
	"os"
	"os/signal"
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
