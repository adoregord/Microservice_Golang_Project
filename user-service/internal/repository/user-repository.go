package repository

import (
	"database/sql"
	"user_microservice/internal/domain"
)

type UserRepoInterface interface {
	ValidateUser(user *domain.User) bool
}

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) UserRepoInterface {
	return UserRepo{
		db: db,
	}
}

func (repo UserRepo) ValidateUser(user *domain.User) bool {
	query := `
	select username
	from account a
	where username = $1`

	err := repo.db.QueryRow(query, user.Username).Scan(&user.Username)

	return err == nil
}
