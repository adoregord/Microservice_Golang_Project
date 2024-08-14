package repository

import (
	"context"
	"database/sql"
	"order_microservice/internal/domain"
	"time"
)

type OrderRepoInterface interface {
	StoreOrderDetails
	UpdateOrderDetails
}
type StoreOrderDetails interface {
	StoreOrderDetails(orderReq domain.OrderRequest, kontek context.Context) (*domain.Message, error)
}
type UpdateOrderDetails interface {
	UpdateOrderDetails(msg domain.Message, kontek context.Context) error
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
		orders (order_type, user_id, item_id, amount) 
	values 
		($1, $2, $3, $4)
	returning id, order_type, user_id, item_id, amount, resp_status, resp_message
	`
	err := repo.db.QueryRowContext(kontek, query, orderReq.OrderType, orderReq.UserID, orderReq.ItemID, orderReq.Amount).
		Scan(
			&message.OrderID,
			&message.OrderType,
			&message.UserId,
			&message.ItemID,
			&message.Amount,
			&message.RespStatus,
			&message.RespMessage,
		)
	if err != nil {
		return nil, err
	}

	return &message, nil
}

func (repo OrderRepo) UpdateOrderDetails(msg domain.Message, kontek context.Context) error {
	query := `
	update 
		orders
	SET 
		total = $1,
		updated_at = $2,
		resp_status = $3,
		resp_message = $4
		where id = $5
	`

	updated_time := time.Now().Format("2006-01-02 15:04:05.000")

	_, err := repo.db.ExecContext(kontek, query, msg.Total, updated_time, msg.RespStatus, msg.RespMessage, msg.OrderID)
	if err != nil {
		return err
	}

	return nil
}