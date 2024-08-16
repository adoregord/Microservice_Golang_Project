package repository

import (
	"context"
	"database/sql"
	"errors"
	"order_microservice/internal/domain"
	"time"
)

type OrderRepoInterface interface {
	StoreOrderDetails
	UpdateOrderDetails
	EditRetryOrder
}
type StoreOrderDetails interface {
	StoreOrderDetails(orderReq domain.OrderRequest, kontek context.Context) (*domain.Message, error)
}
type UpdateOrderDetails interface {
	UpdateOrderDetails(msg domain.Message, kontek context.Context) error
}
type EditRetryOrder interface {
	EditRetryOrder(orderRetryReq *domain.RetryOrder, kontek context.Context) (*domain.Message, error)
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
	returning id, order_type, user_id, item_id, amount, resp_code, resp_message
	`
	err := repo.db.QueryRowContext(kontek, query, orderReq.OrderType, orderReq.UserID, orderReq.ItemID, orderReq.Amount).
		Scan(
			&message.OrderID,
			&message.OrderType,
			&message.UserId,
			&message.ItemID,
			&message.Amount,
			&message.RespCode,
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
		resp_code = $3,
		resp_message = $4
		where id = $5
	`

	updated_time := time.Now().Format("2006-01-02 15:04:05.000")

	_, err := repo.db.ExecContext(kontek, query, msg.Total, updated_time, msg.RespCode, msg.RespMessage, msg.OrderID)
	if err != nil {
		return err
	}

	return nil
}

func (repo OrderRepo) EditRetryOrder(orderRetryReq *domain.RetryOrder, kontek context.Context) (*domain.Message, error) {
	var msg domain.Message
	query := `
	update orders 
	set
	order_type = $1,
	item_id = $2,
	amount = $3
	where
	id = $4 and user_id = $5 and resp_message like '%FAILED%'
	returning id, order_type, user_id, item_id, amount, resp_code, resp_message
	`

	err := repo.db.QueryRowContext(kontek, query, orderRetryReq.OrderType, orderRetryReq.ItemID, orderRetryReq.Amount, orderRetryReq.OrderID, orderRetryReq.UserID).Scan(
		&msg.OrderID,
		&msg.OrderType,
		&msg.UserId,
		&msg.ItemID,
		&msg.Amount,
		&msg.RespCode,
		&msg.RespMessage,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("theres no order for this id or ure not allowed to retry")
		}
		return nil, err
	}

	return &msg, nil
}
