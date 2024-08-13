package main

import (
	"log"
	"order_microservice/internal/provider/database"
	"order_microservice/internal/provider/initialize"
)

func main() {
	database, err := database.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	initialize.StartConsumer(database)
}
