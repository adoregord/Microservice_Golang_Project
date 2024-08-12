package usecase

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"user_microservice/internal/domain"
	"user_microservice/internal/repository"

	"github.com/rs/zerolog/log"
)

type UserUsecase struct {
	UserRepo repository.UserRepoInterface
}

func NewUserUsecase(repo repository.UserRepoInterface) *UserUsecase {
	return &UserUsecase{
		UserRepo: repo,
	}
}

func (uc UserUsecase) ValidateUser(username string) bool {
	type responseStk struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	payload := domain.User{
		Username: username,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return false
	}

	// Make the POST request to hit outbound request
	log.Info().Str("Name", payload.Username).Msg("Hitting outbound api")
	response, err := http.Post("https://7a70c146-33cd-4f4b-9b33-38cc4824afa0.mock.pstmn.io/chekuser", "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return false
	}
	defer response.Body.Close()

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return false
	}

	// Unmarshal the response JSON into responseStk struct
	var responseStruct responseStk
	err = json.Unmarshal(body, &responseStruct)
	if err != nil {
		return false
	}

	//ok := uc.UserRepo.ValidateUser(&payload)

	// Check if the status is success
	if responseStruct.Status == "ok" {
		return true
	}

	return false
}
