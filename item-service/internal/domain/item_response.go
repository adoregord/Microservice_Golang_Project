package domain

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    Data   `json:"data"`
}

type Data struct {
	ID     int `json:"id"`
	Amount int `json:"amount"`
	Price  int `json:"price"`
}
