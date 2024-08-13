package repository

import (
	"context"
	"database/sql"
	"errors"
	"log"
)

type OrchestratorRepoInterface interface {
	ViewOrchesSteps
}

type ViewOrchesSteps interface{
	ViewOrchesSteps(step_type string, step_name string, kontek context.Context) (string, error)
}

type OrchestratorRepo struct {
	db *sql.DB
}

func NewOrchestratorRepo(db *sql.DB) OrchestratorRepoInterface{
	return OrchestratorRepo{
		db: db,
	}
}

func (repo OrchestratorRepo) ViewOrchesSteps(step_type string, step_name string, kontek context.Context) (string, error) {
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
	err := repo.db.QueryRowContext(kontek, query, step_type, step_name).Scan(&topic)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No topic found for step_type: %s and step_name: %s", step_type, step_name)
			return "", errors.New("no data found in database 💀")
		}
		return "", err
	}

	log.Printf("topic dari db: %s", topic)
	return topic, nil
}
