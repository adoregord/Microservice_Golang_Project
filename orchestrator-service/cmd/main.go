package main

import (
	"log"
	provider "microservice_orchestrator/internal/provider/db"
	"microservice_orchestrator/internal/provider/initialize"
)

func main() {
	database, err := provider.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	initialize.StartConsumer(database)
}
