package domain

type OrderRequest struct {
	ID        int    `json:"-"`
	OrderType string `json:"orderType" binding:"required" validate:"noblank,min=1"`
	UserID    string `json:"-"`
	ItemID    int    `json:"itemID" binding:"required" validate:"noblank,min=1"`
	Amount    int    `json:"amount" validate:"gte=1"`
}
