package domain

type OrderRequest struct {
	ID        int    `json:"id,omitempty"`
	OrderType string `json:"orderType" binding:"required" validate:"noblank,min=1"`
	UserID    string `json:"userID,omitempty"`
	PackageID int    `json:"packageID" binding:"required" validate:"noblank,min=1"`
	Amount    int    `json:"amount" validate:"gte=1"`
}
