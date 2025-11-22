// user.go
package usecase

import (
	"context"
	"fmt"

	"github.com/PaulLocust/Avito-review/internal/entity"
	"github.com/PaulLocust/Avito-review/internal/repository"
	"github.com/PaulLocust/Avito-review/pkg/logger"
)

// UserUseCase интерфейс для работы с пользователями
type UserUseCase interface {
	SetUserActive(ctx context.Context, userID string, active bool) (*entity.User, error)
	GetUserReviews(ctx context.Context, userID string) ([]entity.PullRequestShort, error)
}

type userUseCase struct {
	userRepo repository.UserRepository
	prRepo   repository.PRRepository
	logger   logger.Interface
}

func NewUserUseCase(userRepo repository.UserRepository, prRepo repository.PRRepository, l logger.Interface) UserUseCase {
	return &userUseCase{
		userRepo: userRepo,
		prRepo:   prRepo,
		logger:   l,
	}
}

func (uc *userUseCase) SetUserActive(ctx context.Context, userID string, active bool) (*entity.User, error) {
	uc.logger.Info("Setting user %s active=%v", userID, active)

	// Получаем пользователя
	user, err := uc.userRepo.GetUser(ctx, userID)
	if err != nil {
		uc.logger.Warn("User not found: %s", userID)
		return nil, entity.NewAppError(entity.ErrorNotFound, "user not found")
	}

	// Обновляем флаг активности
	user.IsActive = active
	err = uc.userRepo.UpdateUser(ctx, user)
	if err != nil {
		uc.logger.Error("Failed to update user: %v", err)
		return nil, fmt.Errorf("userUseCase - SetUserActive - UpdateUser: %w", err)
	}

	uc.logger.Info("User %s active status updated to %v", userID, active)
	return user, nil
}

func (uc *userUseCase) GetUserReviews(ctx context.Context, userID string) ([]entity.PullRequestShort, error) {
	uc.logger.Debug("Getting reviews for user: %s", userID)

	// Проверяем существование пользователя
	_, err := uc.userRepo.GetUser(ctx, userID)
	if err != nil {
		uc.logger.Warn("User not found: %s", userID)
		return nil, entity.NewAppError(entity.ErrorNotFound, "user not found")
	}

	// Получаем PR где пользователь ревьювер
	prs, err := uc.prRepo.GetPRsByReviewer(ctx, userID)
	if err != nil {
		uc.logger.Error("Failed to get user reviews: %v", err)
		return nil, fmt.Errorf("userUseCase - GetUserReviews - GetPRsByReviewer: %w", err)
	}

	// Конвертируем в короткую версию
	var result []entity.PullRequestShort
	for _, pr := range prs {
		result = append(result, entity.PullRequestShort{
			ID:       pr.ID,
			Name:     pr.Name,
			AuthorID: pr.AuthorID,
			Status:   pr.Status,
		})
	}

	uc.logger.Debug("Found %d PRs for user %s", len(result), userID)
	return result, nil
}
