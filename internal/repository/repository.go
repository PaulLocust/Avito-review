package repository

import (
	"context"

	"github.com/PaulLocust/Avito-review/internal/entity"
)

// TeamRepository - интерфейс для работы с командами
type TeamRepository interface {
	CreateTeam(ctx context.Context, team *entity.Team) error
	GetTeam(ctx context.Context, name string) (*entity.Team, error)
	TeamExists(ctx context.Context, name string) (bool, error)
}

// UserRepository - интерфейс для работы с пользователями
type UserRepository interface {
	CreateOrUpdateUser(ctx context.Context, user *entity.User) error
	GetUser(ctx context.Context, id string) (*entity.User, error)
	UpdateUser(ctx context.Context, user *entity.User) error
	GetActiveUsersByTeam(ctx context.Context, teamName string, excludeUserID string) ([]entity.User, error)
}

// PRRepository - интерфейс для работы с pull requests
type PRRepository interface {
	CreatePR(ctx context.Context, pr *entity.PullRequest) error
	GetPR(ctx context.Context, id string) (*entity.PullRequest, error)
	UpdatePR(ctx context.Context, pr *entity.PullRequest) error
	GetPRsByReviewer(ctx context.Context, userID string) ([]entity.PullRequest, error)
	AddReviewer(ctx context.Context, prID, userID string) error
	RemoveReviewer(ctx context.Context, prID, userID string) error
	ReplaceReviewer(ctx context.Context, prID, oldUserID, newUserID string) error
}