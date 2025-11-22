// usecase.go
package usecase

import (
	"github.com/PaulLocust/Avito-review/internal/repository"
	"github.com/PaulLocust/Avito-review/pkg/logger"
)

type UseCases struct {
	Team TeamUseCase
	User UserUseCase  
	PR   PRUseCase
}

func NewUseCases(
	teamRepo repository.TeamRepository,
	userRepo repository.UserRepository,
	prRepo repository.PRRepository,
	l logger.Interface,
) *UseCases {
	return &UseCases{
		Team: NewTeamUseCase(teamRepo, userRepo, l),
		User: NewUserUseCase(userRepo, prRepo, l),
		PR:   NewPRUseCase(prRepo, userRepo, teamRepo, l),
	}
}