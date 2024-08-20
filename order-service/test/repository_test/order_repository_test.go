package test

import (
	"context"
	"database/sql"
	"order_microservice/internal/domain"
	"order_microservice/internal/repository"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestOrderRepo_StoreOrderDetails_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewOrderRepo(db)

	orderReq := domain.OrderRequest{
		OrderType: "TestOrder",
		UserID:    "123",
		ItemID:    456,
		Amount:    789,
	}

	expectedMessage := &domain.Message{
		OrderID:     1,
		OrderType:   "TestOrder",
		UserId:      "123",
		ItemID:      456,
		Amount:      789,
		RespCode:    200,
		RespMessage: "Order Stored Successfully",
	}

	rows := sqlmock.NewRows([]string{"id", "order_type", "user_id", "item_id", "amount", "resp_code", "resp_message"}).
		AddRow(expectedMessage.OrderID, expectedMessage.OrderType, expectedMessage.UserId, expectedMessage.ItemID, expectedMessage.Amount, expectedMessage.RespCode, expectedMessage.RespMessage)

	mock.ExpectQuery(regexp.QuoteMeta(`
		insert into 
			orders (order_type, user_id, item_id, amount) 
		values 
			($1, $2, $3, $4)
		returning id, order_type, user_id, item_id, amount, resp_code, resp_message`)).
		WithArgs(orderReq.OrderType, orderReq.UserID, orderReq.ItemID, orderReq.Amount).
		WillReturnRows(rows)

	result, err := repo.StoreOrderDetails(orderReq, context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedMessage, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepo_StoreOrderDetails_Failure(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewOrderRepo(db)

	orderReq := domain.OrderRequest{
		OrderType: "TestOrder",
		UserID:    "123",
		ItemID:    456,
		Amount:    789,
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		insert into 
			orders (order_type, user_id, item_id, amount) 
		values 
			($1, $2, $3, $4)
		returning id, order_type, user_id, item_id, amount, resp_code, resp_message`)).
		WithArgs(orderReq.OrderType, orderReq.UserID, orderReq.ItemID, orderReq.Amount).
		WillReturnError(sql.ErrConnDone)

	result, err := repo.StoreOrderDetails(orderReq, context.Background())

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, sql.ErrConnDone.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepo_UpdateOrderDetails_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewOrderRepo(db)

	message := domain.Message{
		OrderID:     1,
		Total:       1000,
		RespCode:    200,
		RespMessage: "Order Updated",
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		update 
			orders
		SET 
			total = $1,
			updated_at = $2,
			resp_code = $3,
			resp_message = $4
		where id = $5`)).
		WithArgs(message.Total, sqlmock.AnyArg(), message.RespCode, message.RespMessage, message.OrderID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.UpdateOrderDetails(message, context.Background())

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepo_UpdateOrderDetails_Failure(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewOrderRepo(db)

	message := domain.Message{
		OrderID:     1,
		Total:       1000,
		RespCode:    200,
		RespMessage: "Order Updated",
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		update 
			orders
		SET 
			total = $1,
			updated_at = $2,
			resp_code = $3,
			resp_message = $4
		where id = $5`)).
		WithArgs(message.Total, sqlmock.AnyArg(), message.RespCode, message.RespMessage, message.OrderID).
		WillReturnError(sql.ErrConnDone)

	err = repo.UpdateOrderDetails(message, context.Background())

	assert.Error(t, err)
	assert.EqualError(t, err, sql.ErrConnDone.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepo_EditRetryOrder_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewOrderRepo(db)

	orderRetryReq := &domain.RetryOrder{
		OrderID:   1,
		OrderType: "RetryOrder",
		UserID:    "123",
		ItemID:    456,
		Amount:    789,
	}

	expectedMessage := &domain.Message{
		OrderID:     1,
		OrderType:   "RetryOrder",
		UserId:      "123",
		ItemID:      456,
		Amount:      789,
		RespCode:    200,
		RespMessage: "Retry Successful",
	}

	rows := sqlmock.NewRows([]string{"id", "order_type", "user_id", "item_id", "amount", "resp_code", "resp_message"}).
		AddRow(expectedMessage.OrderID, expectedMessage.OrderType, expectedMessage.UserId, expectedMessage.ItemID, expectedMessage.Amount, expectedMessage.RespCode, expectedMessage.RespMessage)

	mock.ExpectQuery(regexp.QuoteMeta(`
		update orders 
		set
			order_type = $1,
			item_id = $2,
			amount = $3
		where
			id = $4 and user_id = $5 and resp_message like '%FAILED%'
		returning id, order_type, user_id, item_id, amount, resp_code, resp_message`)).
		WithArgs(orderRetryReq.OrderType, orderRetryReq.ItemID, orderRetryReq.Amount, orderRetryReq.OrderID, orderRetryReq.UserID).
		WillReturnRows(rows)

	result, err := repo.EditRetryOrder(orderRetryReq, context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedMessage, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepo_EditRetryOrder_Failure(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewOrderRepo(db)

	orderRetryReq := &domain.RetryOrder{
		OrderID:   1,
		OrderType: "RetryOrder",
		UserID:    "123",
		ItemID:    456,
		Amount:    789,
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		update orders 
		set
			order_type = $1,
			item_id = $2,
			amount = $3
		where
			id = $4 and user_id = $5 and resp_message like '%FAILED%'
		returning id, order_type, user_id, item_id, amount, resp_code, resp_message`)).
		WithArgs(orderRetryReq.OrderType, orderRetryReq.ItemID, orderRetryReq.Amount, orderRetryReq.OrderID, orderRetryReq.UserID).
		WillReturnError(sql.ErrNoRows)

	result, err := repo.EditRetryOrder(orderRetryReq, context.Background())

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "theres no order for this id or ure not allowed to retry")
	assert.NoError(t, mock.ExpectationsWereMet())
}
