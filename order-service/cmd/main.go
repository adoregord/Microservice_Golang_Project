package main

import (
	"log"
	"order_microservice/internal/provider/database"
	"order_microservice/internal/provider/initialize"
)

func main() {
	db, err := database.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	initialize.StartConsumer(db)
}
