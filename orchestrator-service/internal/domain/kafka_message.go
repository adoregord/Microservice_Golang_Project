package domain

// IncomingMessage represents the structure of the incoming messages
type IncomingMessage struct {
	OrderType     string `json:"orderType"`
	OrderService  string `json:"orderService,omitempty"`
	TransactionId string `json:"transactionId"`
	UserId        string `json:"userId"`
	PackageId     string `json:"packageId"`
	RespStatus    string `json:"respStatus,omitempty"`
	RespMessage   string `json:"respMessage,omitempty"`
	RespCode      int    `json:"respCode,omitempty"`
}
