package repository

import (
	"context"
	"database/sql"
	"log"
	"microservice_orchestrator/internal/domain"
)

type OrchestratorRepoInterface interface {
	ViewOrchesSteps
	OrchesLog
}

type ViewOrchesSteps interface {
	ViewOrchesSteps(incoming_message *domain.Message, kontek context.Context) string
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

func (repo OrchestratorRepo) ViewOrchesSteps(incoming_message *domain.Message, kontek context.Context) string {
	query := `
	select
		kafka_topic
	from
		orches_routes
	where
		step_type = $1
		and
		step_name = $2
	`
	var topic string
	err := repo.db.QueryRowContext(kontek, query, incoming_message.OrderType, incoming_message.OrderService).Scan(&topic)
	if err != nil {
		if err == sql.ErrNoRows {
			incoming_message.RespCode = 404
			incoming_message.RespStatus = "Not Found"
			incoming_message.RespMessage = "FAILED data of that step type in database 🚨💀"
			log.Printf("No topic found for step_type: %s and step_name: %s", incoming_message.OrderType, incoming_message.OrderService)
			return ""
		}
		return ""
	}

	log.Printf("topic dari db: %s", topic)
	return topic
}

func (repo OrchestratorRepo) OrchesLog(message domain.Message, kontek context.Context) error {

	query := `
	INSERT INTO 
	    orches_logs (order_type, step_name, user_id, item_id,
	                amount, resp_code, resp_status, resp_message,
	                payload)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := repo.db.ExecContext(kontek, query, message.OrderType, message.OrderService, message.UserId, message.ItemID, message.Amount, message.RespCode, message.RespStatus, message.RespMessage, message)
	if err != nil {
		return err
	}

	log.Printf("message: %v", message)

	return nil
}
