// team.go
package usecase

import (
	"context"
	"fmt"

	"github.com/PaulLocust/Avito-review/internal/entity"
	"github.com/PaulLocust/Avito-review/internal/repository"
	"github.com/PaulLocust/Avito-review/pkg/logger"
)


type TeamUseCase interface {
	CreateTeam(ctx context.Context, team entity.Team) error
	GetTeam(ctx context.Context, teamName string) (*entity.Team, error)
}

type teamUseCase struct {
	teamRepo repository.TeamRepository
	userRepo repository.UserRepository
	logger   logger.Interface
}

func NewTeamUseCase(teamRepo repository.TeamRepository, userRepo repository.UserRepository, l logger.Interface) TeamUseCase {
	return &teamUseCase{
		teamRepo: teamRepo,
		userRepo: userRepo,
		logger:   l,
	}
}

func (uc *teamUseCase) CreateTeam(ctx context.Context, team entity.Team) error {
	uc.logger.Info("Creating team: %s with %d members", team.Name, len(team.Members))

	// Проверяем существование команды
	exists, err := uc.teamRepo.TeamExists(ctx, team.Name)
	if err != nil {
		uc.logger.Error("Failed to check team existence: %v", err)
		return fmt.Errorf("teamUseCase - CreateTeam - TeamExists: %w", err)
	}
	if exists {
		uc.logger.Warn("Team already exists: %s", team.Name)
		return entity.NewAppError(entity.ErrorTeamExists, "team already exists")
	}

	// Создаем команду и пользователей
	err = uc.teamRepo.CreateTeam(ctx, &team)
	if err != nil {
		uc.logger.Error("Failed to create team: %v", err)
		return fmt.Errorf("teamUseCase - CreateTeam - CreateTeam: %w", err)
	}

	uc.logger.Info("Team created successfully: %s", team.Name)
	return nil
}

func (uc *teamUseCase) GetTeam(ctx context.Context, teamName string) (*entity.Team, error) {
	uc.logger.Debug("Getting team: %s", teamName)

	team, err := uc.teamRepo.GetTeam(ctx, teamName)
	if err != nil {
		uc.logger.Warn("Team not found: %s", teamName)
		return nil, entity.NewAppError(entity.ErrorNotFound, "team not found")
	}

	uc.logger.Debug("Team found: %s with %d members", teamName, len(team.Members))
	return team, nil
}