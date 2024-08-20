package domain

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    Data   `json:"data"`
}

type Data struct {
	ID       int `json:"id"`
	Quantity int `json:"quantity"`
	Price    int `json:"price"`
}
