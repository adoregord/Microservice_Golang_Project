package domain

type OrderRequest struct {
	OrderType     string `json:"orderType,required"`
	TransactionID string `json:"transactionId,required"`
	UserId        string `json:"userId,required"`
	PackageId     string `json:"packageId,required"`
}
