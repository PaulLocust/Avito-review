// pr.go
package usecase

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/PaulLocust/Avito-review/internal/entity"
	"github.com/PaulLocust/Avito-review/internal/repository"
	"github.com/PaulLocust/Avito-review/pkg/logger"
)

type PRUseCase interface {
	CreatePR(ctx context.Context, prID, name, authorID string) (*entity.PullRequest, error)
	MergePR(ctx context.Context, prID string) (*entity.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldUserID string) (*entity.PullRequest, string, error)
}

type prUseCase struct {
	prRepo   repository.PRRepository
	userRepo repository.UserRepository
	teamRepo repository.TeamRepository
	logger   logger.Interface
}

func NewPRUseCase(
	prRepo repository.PRRepository,
	userRepo repository.UserRepository,
	teamRepo repository.TeamRepository,
	l logger.Interface,
) PRUseCase {
	return &prUseCase{
		prRepo:   prRepo,
		userRepo: userRepo,
		teamRepo: teamRepo,
		logger:   l,
	}
}

func (uc *prUseCase) CreatePR(ctx context.Context, prID, name, authorID string) (*entity.PullRequest, error) {
	uc.logger.Info("Creating PR: %s by author %s", prID, authorID)

	// Проверяем существование PR
	existingPR, _ := uc.prRepo.GetPR(ctx, prID)
	if existingPR != nil {
		uc.logger.Warn("PR already exists: %s", prID)
		return nil, entity.NewAppError(entity.ErrorPRExists, "PR already exists")
	}

	// Получаем автора
	author, err := uc.userRepo.GetUser(ctx, authorID)
	if err != nil {
		uc.logger.Warn("Author not found: %s", authorID)
		return nil, entity.NewAppError(entity.ErrorNotFound, "author not found")
	}

	// Получаем активных пользователей команды (исключая автора)
	teamMembers, err := uc.userRepo.GetActiveUsersByTeam(ctx, author.TeamName, authorID)
	if err != nil {
		uc.logger.Error("Failed to get team members: %v", err)
		return nil, fmt.Errorf("prUseCase - CreatePR - GetActiveUsersByTeam: %w", err)
	}

	uc.logger.Debug("Found %d active team members for PR assignment", len(teamMembers))

	// Выбираем до 2 случайных ревьюверов
	reviewers := uc.selectRandomReviewers(teamMembers, 2)
	uc.logger.Info("Selected %d reviewers for PR %s: %v", len(reviewers), prID, reviewers)

	// Создаем PR
	pr := &entity.PullRequest{
		ID:                prID,
		Name:              name,
		AuthorID:          authorID,
		Status:            entity.StatusOpen,
		AssignedReviewers: reviewers,
		CreatedAt:         time.Now(),
	}

	err = uc.prRepo.CreatePR(ctx, pr)
	if err != nil {
		uc.logger.Error("Failed to create PR: %v", err)
		return nil, fmt.Errorf("prUseCase - CreatePR - CreatePR: %w", err)
	}

	uc.logger.Info("PR created successfully: %s", prID)
	return pr, nil
}

func (uc *prUseCase) MergePR(ctx context.Context, prID string) (*entity.PullRequest, error) {
	uc.logger.Info("Merging PR: %s", prID)

	// Получаем PR
	pr, err := uc.prRepo.GetPR(ctx, prID)
	if err != nil {
		uc.logger.Warn("PR not found: %s", prID)
		return nil, entity.NewAppError(entity.ErrorNotFound, "PR not found")
	}

	// Если уже мерджен - возвращаем как есть (идемпотентность)
	if pr.Status == entity.StatusMerged {
		uc.logger.Debug("PR already merged: %s", prID)
		return pr, nil
	}

	// Обновляем статус
	now := time.Now()
	pr.Status = entity.StatusMerged
	pr.MergedAt = &now

	err = uc.prRepo.UpdatePR(ctx, pr)
	if err != nil {
		uc.logger.Error("Failed to merge PR: %v", err)
		return nil, fmt.Errorf("prUseCase - MergePR - UpdatePR: %w", err)
	}

	uc.logger.Info("PR merged successfully: %s", prID)
	return pr, nil
}

func (uc *prUseCase) ReassignReviewer(ctx context.Context, prID, oldUserID string) (*entity.PullRequest, string, error) {
	uc.logger.Info("Reassigning reviewer %s from PR %s", oldUserID, prID)

	// Получаем PR
	pr, err := uc.prRepo.GetPR(ctx, prID)
	if err != nil {
		uc.logger.Warn("PR not found: %s", prID)
		return nil, "", entity.NewAppError(entity.ErrorNotFound, "PR not found")
	}

	// Проверяем что PR не мерджен
	if pr.Status == entity.StatusMerged {
		uc.logger.Warn("Cannot reassign on merged PR: %s", prID)
		return nil, "", entity.NewAppError(entity.ErrorPRMerged, "cannot reassign on merged PR")
	}

	// Проверяем что старый ревьювер назначен на PR
	if !uc.containsReviewer(pr.AssignedReviewers, oldUserID) {
		uc.logger.Warn("Reviewer %s not assigned to PR %s", oldUserID, prID)
		return nil, "", entity.NewAppError(entity.ErrorNotAssigned, "reviewer is not assigned to this PR")
	}

	// Получаем информацию о старом ревьювере
	oldReviewer, err := uc.userRepo.GetUser(ctx, oldUserID)
	if err != nil {
		uc.logger.Warn("Old reviewer not found: %s", oldUserID)
		return nil, "", entity.NewAppError(entity.ErrorNotFound, "old reviewer not found")
	}

	// Получаем активных пользователей команды (исключая старого ревьювера и автора PR)
	teamMembers, err := uc.userRepo.GetActiveUsersByTeam(ctx, oldReviewer.TeamName, oldUserID)
	if err != nil {
		uc.logger.Error("Failed to get team members: %v", err)
		return nil, "", fmt.Errorf("prUseCase - ReassignReviewer - GetActiveUsersByTeam: %w", err)
	}

	// Исключаем автора PR из кандидатов
	teamMembers = uc.filterOutUser(teamMembers, pr.AuthorID)

	// Исключаем уже назначенных ревьюверов
	teamMembers = uc.filterOutReviewers(teamMembers, pr.AssignedReviewers)

	uc.logger.Debug("Found %d candidate reviewers for replacement", len(teamMembers))

	if len(teamMembers) == 0 {
		uc.logger.Warn("No replacement candidates found for reviewer %s in PR %s", oldUserID, prID)
		return nil, "", entity.NewAppError(entity.ErrorNoCandidate, "no active replacement candidate in team")
	}

	// Выбираем случайного нового ревьювера
	newReviewer := teamMembers[rand.Intn(len(teamMembers))]
	uc.logger.Info("Replacing reviewer %s with %s in PR %s", oldUserID, newReviewer.ID, prID)

	// Заменяем ревьювера
	err = uc.prRepo.ReplaceReviewer(ctx, prID, oldUserID, newReviewer.ID)
	if err != nil {
		uc.logger.Error("Failed to replace reviewer: %v", err)
		return nil, "", fmt.Errorf("prUseCase - ReassignReviewer - ReplaceReviewer: %w", err)
	}

	// Обновляем список ревьюверов в возвращаемом объекте
	for i, reviewer := range pr.AssignedReviewers {
		if reviewer == oldUserID {
			pr.AssignedReviewers[i] = newReviewer.ID
			break
		}
	}

	uc.logger.Info("Reviewer reassigned successfully in PR %s", prID)
	return pr, newReviewer.ID, nil
}

// Вспомогательные методы (без изменений)
func (uc *prUseCase) selectRandomReviewers(users []entity.User, max int) []string {
	if len(users) == 0 {
		return []string{}
	}

	// Перемешиваем пользователей
	shuffled := make([]entity.User, len(users))
	copy(shuffled, users)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	// Берем до max пользователей
	count := len(shuffled)
	if count > max {
		count = max
	}

	var reviewers []string
	for i := 0; i < count; i++ {
		reviewers = append(reviewers, shuffled[i].ID)
	}

	return reviewers
}

func (uc *prUseCase) containsReviewer(reviewers []string, userID string) bool {
	for _, reviewer := range reviewers {
		if reviewer == userID {
			return true
		}
	}
	return false
}

func (uc *prUseCase) filterOutUser(users []entity.User, userID string) []entity.User {
	var result []entity.User
	for _, user := range users {
		if user.ID != userID {
			result = append(result, user)
		}
	}
	return result
}

func (uc *prUseCase) filterOutReviewers(users []entity.User, reviewers []string) []entity.User {
	var result []entity.User
	for _, user := range users {
		if !uc.containsReviewer(reviewers, user.ID) {
			result = append(result, user)
		}
	}
	return result
}