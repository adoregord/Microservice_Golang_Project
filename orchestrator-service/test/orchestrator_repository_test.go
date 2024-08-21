package test

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"microservice_orchestrator/internal/domain"
	"microservice_orchestrator/internal/repository"
	"testing"
)

func TestOrchestratorRepo_GetNextTopic_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewOrchestratorRepo(db)

	incomingMessage := &domain.Message{
		OrderType:    "order_type_1",
		OrderService: "order_service_1",
	}

	query := `SELECT kafka_topic FROM orches_routes WHERE step_type = \$1 AND step_name = \$2`

	mock.ExpectQuery(query).WithArgs(incomingMessage.OrderType, incomingMessage.OrderService).
		WillReturnRows(sqlmock.NewRows([]string{"kafka_topic"}).AddRow("next_topic"))

	ctx := context.TODO()
	topic := repo.GetNextTopic(incomingMessage, false, ctx)

	assert.Equal(t, "next_topic", topic)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrchestratorRepo_GetNextTopic_Retry_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewOrchestratorRepo(db)

	incomingMessage := &domain.Message{
		OrderType:    "order_type_1",
		OrderService: "order_service_1",
	}

	// Use a regexp to match the query, ignoring whitespace differences
	queryRegexp := `(?s)SELECT\s+kafka_topic\s+FROM\s+orches_routes\s+WHERE\s+id\s*=\s*\(\s*SELECT\s+max\(id\)\s+FROM\s+orches_routes\s+WHERE\s+step_type\s*=\s*\$1\s+AND\s+id\s*<\s*\(\s*SELECT\s+id\s+FROM\s+orches_routes\s+WHERE\s+step_type\s*=\s*\$1\s+AND\s+step_name\s*=\s*\$2\s*\)\s*\)`

	mock.ExpectQuery(queryRegexp).WithArgs(incomingMessage.OrderType, incomingMessage.OrderService).
		WillReturnRows(sqlmock.NewRows([]string{"kafka_topic"}).AddRow("retry_topic"))

	ctx := context.TODO()
	topic := repo.GetNextTopic(incomingMessage, true, ctx)

	assert.Equal(t, "retry_topic", topic)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrchestratorRepo_GetNextTopic_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewOrchestratorRepo(db)

	incomingMessage := &domain.Message{
		OrderType:    "order_type_1",
		OrderService: "order_service_1",
	}

	query := `SELECT kafka_topic FROM orches_routes WHERE step_type = \$1 AND step_name = \$2`

	mock.ExpectQuery(query).WithArgs(incomingMessage.OrderType, incomingMessage.OrderService).
		WillReturnError(sql.ErrNoRows)

	ctx := context.TODO()
	topic := repo.GetNextTopic(incomingMessage, false, ctx)

	assert.Equal(t, "order_topic", topic)
	assert.Equal(t, 404, incomingMessage.RespCode)
	assert.Equal(t, "Not Found", incomingMessage.RespStatus)
	assert.Equal(t, "FAILED data of that step type in database 🚨💀", incomingMessage.RespMessage)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrchestratorRepo_GetNextTopic_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewOrchestratorRepo(db)

	incomingMessage := &domain.Message{
		OrderType:    "order_type_1",
		OrderService: "order_service_1",
	}

	query := `SELECT kafka_topic FROM orches_routes WHERE step_type = \$1 AND step_name = \$2`

	mock.ExpectQuery(query).WithArgs(incomingMessage.OrderType, incomingMessage.OrderService).
		WillReturnError(errors.New("some error"))

	ctx := context.TODO()
	topic := repo.GetNextTopic(incomingMessage, false, ctx)

	assert.Equal(t, "some error", topic)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrchestratorRepo_GetRollbackTopic_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewOrchestratorRepo(db)

	incomingMessage := &domain.Message{
		OrderType:    "order_type_1",
		OrderService: "order_service_1",
	}

	query := `SELECT rollback FROM orches_routes WHERE step_type = \$1 AND step_name = \$2`

	mock.ExpectQuery(query).WithArgs(incomingMessage.OrderType, incomingMessage.OrderService).
		WillReturnRows(sqlmock.NewRows([]string{"rollback"}).AddRow("rollback_topic"))

	ctx := context.TODO()
	topic := repo.GetRollbackTopic(incomingMessage, ctx)

	assert.Equal(t, "rollback_topic", topic)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrchestratorRepo_GetRollbackTopic_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewOrchestratorRepo(db)

	incomingMessage := &domain.Message{
		OrderType:    "order_type_1",
		OrderService: "order_service_1",
	}

	query := `SELECT rollback FROM orches_routes WHERE step_type = \$1 AND step_name = \$2`

	mock.ExpectQuery(query).WithArgs(incomingMessage.OrderType, incomingMessage.OrderService).
		WillReturnError(sql.ErrNoRows)

	ctx := context.TODO()
	topic := repo.GetRollbackTopic(incomingMessage, ctx)

	assert.Equal(t, "", topic)
	assert.Equal(t, 404, incomingMessage.RespCode)
	assert.Equal(t, "Not Found", incomingMessage.RespStatus)
	assert.Equal(t, "FAILED data of that step type in database 🚨💀", incomingMessage.RespMessage)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrchestratorRepo_GetRollbackTopic_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewOrchestratorRepo(db)

	incomingMessage := &domain.Message{
		OrderType:    "order_type_1",
		OrderService: "order_service_1",
	}

	query := `SELECT rollback FROM orches_routes WHERE step_type = \$1 AND step_name = \$2`

	mock.ExpectQuery(query).WithArgs(incomingMessage.OrderType, incomingMessage.OrderService).
		WillReturnError(errors.New("some error"))

	ctx := context.TODO()
	topic := repo.GetRollbackTopic(incomingMessage, ctx)

	assert.Equal(t, "", topic)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrchestratorRepo_OrchesLog_Success(t *testing.T) {
	// Setup
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewOrchestratorRepo(db)

	// Test data
	message := domain.Message{
		OrderID:      1,
		OrderType:    "order_type_1",
		OrderService: "order_service_1",
		UserId:       "user_1",
		ItemID:       1,
		Amount:       100,
		RespCode:     200,
		RespStatus:   "Success",
		RespMessage:  "Order processed successfully",
	}

	// Expected query
	expectedQuery := `
	INSERT INTO 
	    orches_logs \(order_id, order_type, step_name, user_id, item_id,
	                amount, resp_code, resp_status, resp_message,
	                payload\)
	VALUES \(\$1, \$2, \$3, \$4, \$5, \$6, \$7, \$8, \$9, \$10\)
	`

	// Setup expectation
	mock.ExpectExec(expectedQuery).
		WithArgs(
			message.OrderID,
			message.OrderType,
			message.OrderService,
			message.UserId,
			message.ItemID,
			message.Amount,
			message.RespCode,
			message.RespStatus,
			message.RespMessage,
			sqlmock.AnyArg(), // Untuk payload, karena ini adalah struct
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute
	ctx := context.Background()
	err = repo.OrchesLog(message, ctx)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrchestratorRepo_OrchesLog_Error(t *testing.T) {
	// Setup
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewOrchestratorRepo(db)

	// Test data
	message := domain.Message{
		OrderID:      1,
		OrderType:    "order_type_1",
		OrderService: "order_service_1",
		UserId:       "user_1",
		ItemID:       1,
		Amount:       100,
		RespCode:     200,
		RespStatus:   "Success",
		RespMessage:  "Order processed successfully",
	}

	// Expected query
	expectedQuery := `
	INSERT INTO 
	    orches_logs \(order_id, order_type, step_name, user_id, item_id,
	                amount, resp_code, resp_status, resp_message,
	                payload\)
	VALUES \(\$1, \$2, \$3, \$4, \$5, \$6, \$7, \$8, \$9, \$10\)
	`

	// Setup expectation
	mock.ExpectExec(expectedQuery).
		WithArgs(
			message.OrderID,
			message.OrderType,
			message.OrderService,
			message.UserId,
			message.ItemID,
			message.Amount,
			message.RespCode,
			message.RespStatus,
			message.RespMessage,
			sqlmock.AnyArg(), // Untuk payload, karena ini adalah struct
		).
		WillReturnError(errors.New("insert error"))

	// Execute
	ctx := context.Background()
	err = repo.OrchesLog(message, ctx)

	// Assert
	assert.Error(t, err)
	assert.EqualError(t, err, "insert error")
	assert.NoError(t, mock.ExpectationsWereMet())
}
