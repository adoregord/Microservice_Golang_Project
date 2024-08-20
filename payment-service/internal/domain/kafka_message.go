package domain

type Message struct {
	OrderID      int    `json:"orderID"`
	OrderType    string `json:"orderType"`
	OrderService string `json:"orderService,omitempty"`
	UserId       string `json:"userId"`
	ItemID       int    `json:"itemID"`
	RespStatus   string `json:"respStatus,omitempty"`
	RespMessage  string `json:"respMessage,omitempty"`
	RespCode     int    `json:"respCode,omitempty"`
	Total        int    `json:"total,omitempty"`
	Amount       int    `json:"amount"`
	Retry        int    `json:"retry"`
}
