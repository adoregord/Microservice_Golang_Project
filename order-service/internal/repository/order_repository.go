package repository

import (
	"context"
	"database/sql"
	"order_microservice/internal/domain"
)

type OrderRepoInterface interface {
	StoreOrderDetails
}
type StoreOrderDetails interface {
	StoreOrderDetails(orderReq domain.OrderRequest, kontek context.Context) (*domain.Message, error)
}

type OrderRepo struct {
	db *sql.DB
}

func NewOrderRepo(DB *sql.DB) OrderRepoInterface {
	return OrderRepo{
		db: DB,
	}
}

func (repo OrderRepo) StoreOrderDetails(orderReq domain.OrderRequest, kontek context.Context) (*domain.Message, error) {
	var message domain.Message
	query := `
	insert into 
		orders (order_type, user_id, package_id, amount) 
	values 
		($1, $2, $3, $4)
	returning id, order_type, user_id, package_id, amount, resp_status, resp_message
	`
	err := repo.db.QueryRowContext(kontek, query, orderReq.OrderType, orderReq.UserID, orderReq.PackageID, orderReq.Amount).
	Scan(
		&message.OrderID,
		&message.OrderType,
		&message.UserId,
		&message.PackageID,
		&message.Amount,
		&message.RespStatus,
		&message.RespMessage, 
	)
	if err != nil {
		return nil, err
	}

	return &message, nil
}
