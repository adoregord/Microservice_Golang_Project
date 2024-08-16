package domain

type OrderRequest struct {
	ID        int    `json:"-"`
	OrderType string `json:"orderType" binding:"required" validate:"noblank,min=1"`
	UserID    string `json:"-"`
	ItemID    int    `json:"itemID" binding:"required" validate:"noblank,gte=1"`
	Amount    int    `json:"amount" validate:"gte=1"`
}

type RetryOrder struct {
	OrderID   int    `json:"orderID"`
	OrderType string `json:"orderType" binding:"required" validate:"noblank,min=1"`
	UserID    string `json:"-"`
	ItemID    int    `json:"itemID" binding:"required" validate:"noblank,gte=1"`
	Amount    int    `json:"amount" validate:"gte=1"`
}
