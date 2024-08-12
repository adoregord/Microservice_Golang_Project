package main

import (
	"fmt"
	"order_microservice/internal/provider/routes"
)

func main() {
	err := routes.SetupRoutes().Run("127.0.0.1:8081")
	if err != nil {
		fmt.Println(err.Error())
	}
}
