package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"microservice_orchestrator/internal/domain"
)

type OrchestratorRepoInterface interface {
	ViewOrchesSteps
	ViewOrchesFailStep
	OrchesLog
}

type ViewOrchesSteps interface {
	GetNextTopic(incoming_message *domain.Message, isRetry bool, kontek context.Context) string
}
type ViewOrchesFailStep interface {
	GetRollbackTopic(incoming_message *domain.Message, kontek context.Context) string
}
type OrchesLog interface {
	OrchesLog(message domain.Message, kontek context.Context) error
}

type OrchestratorRepo struct {
	db *sql.DB
}

func NewOrchestratorRepo(db *sql.DB) OrchestratorRepoInterface {
	return OrchestratorRepo{
		db: db,
	}
}

func (repo OrchestratorRepo) GetNextTopic(incoming_message *domain.Message, isRetry bool, kontek context.Context) string {
	var query string
	if isRetry {
		// Instead of going to the previous step, stay on the current step
		query = `
            SELECT kafka_topic 
			FROM orches_routes
			WHERE id = (
				SELECT max(id) 
				FROM orches_routes 
				WHERE step_type = $1 
				  AND id < (
					  SELECT id 
					  FROM orches_routes 
					  WHERE step_type = $1 
						AND step_name = $2
				  )
			);
        `
	} else {
		query = `
            SELECT kafka_topic FROM orches_routes
            WHERE step_type = $1 AND step_name = $2
        `
	}
	var topic string
	err := repo.db.QueryRowContext(kontek, query, incoming_message.OrderType, incoming_message.OrderService).Scan(&topic)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			incoming_message.RespCode = 404
			incoming_message.RespStatus = "Not Found"
			incoming_message.RespMessage = "FAILED data of that step type in database 🚨💀"
			log.Printf("No topic found for step_type: %s and step_name: %s", incoming_message.OrderType, incoming_message.OrderService)
			return "order_topic"
		}
		log.Println(err.Error())
		return err.Error()
	}
	return topic
}

func (repo OrchestratorRepo) GetRollbackTopic(incoming_message *domain.Message, ctx context.Context) string {
	query := `
		SELECT rollback FROM orches_routes
		WHERE step_type = $1 AND step_name = $2
	`

	var rollbackTopic string
	err := repo.db.QueryRowContext(ctx, query, incoming_message.OrderType, incoming_message.OrderService).Scan(&rollbackTopic)
	if err != nil || rollbackTopic == "" {
		if errors.Is(err, sql.ErrNoRows) {
			incoming_message.RespCode = 404
			incoming_message.RespStatus = "Not Found"
			incoming_message.RespMessage = "FAILED data of that step type in database 🚨💀"
			log.Printf("No topic found for step_type: %s and step_name: %s", incoming_message.OrderType, incoming_message.OrderService)
			return ""
		}
		return ""
	}
	return rollbackTopic
}

func (repo OrchestratorRepo) OrchesLog(message domain.Message, kontek context.Context) error {

	query := `
	INSERT INTO 
	    orches_logs (order_id, order_type, step_name, user_id, item_id,
	                amount, resp_code, resp_status, resp_message,
	                payload)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	payload, _ := json.Marshal(message)

	_, err := repo.db.ExecContext(kontek, query, message.OrderID, message.OrderType, message.OrderService, message.UserId, message.ItemID, message.Amount, message.RespCode, message.RespStatus, message.RespMessage, payload)
	if err != nil {
		return err
	}

	return nil
}
