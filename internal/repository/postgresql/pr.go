// pr.go
package postgresql

import (
	"context"
	"fmt"
	"time"

	"github.com/PaulLocust/Avito-review/internal/entity"
	"github.com/PaulLocust/Avito-review/internal/repository"
	"github.com/PaulLocust/Avito-review/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

type prRepo struct {
	db     *pgxpool.Pool
	logger logger.Interface
}

func NewPRRepository(db *pgxpool.Pool, l logger.Interface) repository.PRRepository {
	return &prRepo{db: db, logger: l}
}

func (r *prRepo) CreatePR(ctx context.Context, pr *entity.PullRequest) error {
	r.logger.Debug("Creating PR: %s with %d reviewers", pr.ID, len(pr.AssignedReviewers))

	tx, err := r.db.Begin(ctx)
	if err != nil {
		r.logger.Error("Failed to begin transaction for PR creation: %v", err)
		return fmt.Errorf("prRepo - CreatePR - Begin: %w", err)
	}
	defer tx.Rollback(ctx)

	// Создаем PR
	_, err = tx.Exec(ctx, `
		INSERT INTO pull_requests (id, name, author_id, status, created_at) 
		VALUES ($1, $2, $3, $4, $5)
	`, pr.ID, pr.Name, pr.AuthorID, pr.Status, pr.CreatedAt)
	if err != nil {
		r.logger.Error("Failed to insert PR: %v", err)
		return fmt.Errorf("prRepo - CreatePR - Insert PR: %w", err)
	}

	// Добавляем ревьюверов
	for _, reviewerID := range pr.AssignedReviewers {
		_, err = tx.Exec(ctx, `
			INSERT INTO pr_reviewers (pr_id, user_id) 
			VALUES ($1, $2)
		`, pr.ID, reviewerID)
		if err != nil {
			r.logger.Error("Failed to insert reviewer %s: %v", reviewerID, err)
			return fmt.Errorf("prRepo - CreatePR - Insert reviewer %s: %w", reviewerID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		r.logger.Error("Failed to commit transaction for PR creation: %v", err)
		return fmt.Errorf("prRepo - CreatePR - Commit: %w", err)
	}

	r.logger.Info("PR created successfully: %s with %d reviewers", pr.ID, len(pr.AssignedReviewers))
	return nil
}

func (r *prRepo) GetPR(ctx context.Context, id string) (*entity.PullRequest, error) {
	r.logger.Debug("Getting PR: %s", id)

	var pr entity.PullRequest
	var mergedAt *time.Time

	// Получаем основную информацию о PR
	err := r.db.QueryRow(ctx, `
		SELECT id, name, author_id, status, created_at, merged_at
		FROM pull_requests 
		WHERE id = $1
	`, id).Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &mergedAt)

	if err != nil {
		r.logger.Warn("PR not found: %s", id)
		return nil, fmt.Errorf("prRepo - GetPR - Query PR: %w", err)
	}

	// Конвертируем mergedAt
	if mergedAt != nil {
		pr.MergedAt = mergedAt
	}

	// Получаем список ревьюверов
	rows, err := r.db.Query(ctx, `
		SELECT user_id 
		FROM pr_reviewers 
		WHERE pr_id = $1
		ORDER BY user_id
	`, id)
	if err != nil {
		r.logger.Error("Failed to query PR reviewers: %v", err)
		return nil, fmt.Errorf("prRepo - GetPR - Query reviewers: %w", err)
	}
	defer rows.Close()

	reviewerCount := 0
	for rows.Next() {
		var reviewerID string
		err := rows.Scan(&reviewerID)
		if err != nil {
			r.logger.Error("Failed to scan reviewer: %v", err)
			return nil, fmt.Errorf("prRepo - GetPR - Scan reviewer: %w", err)
		}
		pr.AssignedReviewers = append(pr.AssignedReviewers, reviewerID)
		reviewerCount++
	}

	r.logger.Debug("Found PR %s with %d reviewers", id, reviewerCount)
	return &pr, nil
}

func (r *prRepo) UpdatePR(ctx context.Context, pr *entity.PullRequest) error {
	r.logger.Debug("Updating PR: %s, status: %s", pr.ID, pr.Status)

	_, err := r.db.Exec(ctx, `
		UPDATE pull_requests 
		SET name = $1, status = $2, merged_at = $3
		WHERE id = $4
	`, pr.Name, pr.Status, pr.MergedAt, pr.ID)

	if err != nil {
		r.logger.Error("Failed to update PR %s: %v", pr.ID, err)
		return fmt.Errorf("prRepo - UpdatePR: %w", err)
	}

	r.logger.Debug("PR updated successfully: %s", pr.ID)
	return nil
}

func (r *prRepo) GetPRsByReviewer(ctx context.Context, userID string) ([]entity.PullRequest, error) {
	r.logger.Debug("Getting PRs by reviewer: %s", userID)

	var prs []entity.PullRequest

	rows, err := r.db.Query(ctx, `
		SELECT p.id, p.name, p.author_id, p.status, p.created_at, p.merged_at
		FROM pull_requests p
		JOIN pr_reviewers pr ON p.id = pr.pr_id
		WHERE pr.user_id = $1
		ORDER BY p.created_at DESC
	`, userID)
	if err != nil {
		r.logger.Error("Failed to query PRs by reviewer: %v", err)
		return nil, fmt.Errorf("prRepo - GetPRsByReviewer - Query: %w", err)
	}
	defer rows.Close()

	prCount := 0
	for rows.Next() {
		var pr entity.PullRequest
		var mergedAt *time.Time

		err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &mergedAt)
		if err != nil {
			r.logger.Error("Failed to scan PR: %v", err)
			return nil, fmt.Errorf("prRepo - GetPRsByReviewer - Scan: %w", err)
		}

		if mergedAt != nil {
			pr.MergedAt = mergedAt
		}

		// Получаем ревьюверов для этого PR
		reviewerRows, err := r.db.Query(ctx, `
			SELECT user_id FROM pr_reviewers WHERE pr_id = $1
		`, pr.ID)
		if err != nil {
			r.logger.Error("Failed to query reviewers for PR %s: %v", pr.ID, err)
			return nil, fmt.Errorf("prRepo - GetPRsByReviewer - Query reviewers: %w", err)
		}

		reviewerCount := 0
		for reviewerRows.Next() {
			var reviewerID string
			err := reviewerRows.Scan(&reviewerID)
			if err != nil {
				reviewerRows.Close()
				r.logger.Error("Failed to scan reviewer for PR %s: %v", pr.ID, err)
				return nil, fmt.Errorf("prRepo - GetPRsByReviewer - Scan reviewer: %w", err)
			}
			pr.AssignedReviewers = append(pr.AssignedReviewers, reviewerID)
			reviewerCount++
		}
		reviewerRows.Close()

		prs = append(prs, pr)
		prCount++
	}

	r.logger.Debug("Found %d PRs for reviewer %s", prCount, userID)
	return prs, nil
}

func (r *prRepo) AddReviewer(ctx context.Context, prID, userID string) error {
	r.logger.Debug("Adding reviewer %s to PR %s", userID, prID)

	_, err := r.db.Exec(ctx, `
		INSERT INTO pr_reviewers (pr_id, user_id) 
		VALUES ($1, $2)
		ON CONFLICT (pr_id, user_id) DO NOTHING
	`, prID, userID)

	if err != nil {
		r.logger.Error("Failed to add reviewer %s to PR %s: %v", userID, prID, err)
		return fmt.Errorf("prRepo - AddReviewer: %w", err)
	}

	r.logger.Debug("Reviewer %s added to PR %s", userID, prID)
	return nil
}

func (r *prRepo) RemoveReviewer(ctx context.Context, prID, userID string) error {
	r.logger.Debug("Removing reviewer %s from PR %s", userID, prID)

	_, err := r.db.Exec(ctx, `
		DELETE FROM pr_reviewers 
		WHERE pr_id = $1 AND user_id = $2
	`, prID, userID)

	if err != nil {
		r.logger.Error("Failed to remove reviewer %s from PR %s: %v", userID, prID, err)
		return fmt.Errorf("prRepo - RemoveReviewer: %w", err)
	}

	r.logger.Debug("Reviewer %s removed from PR %s", userID, prID)
	return nil
}

func (r *prRepo) ReplaceReviewer(ctx context.Context, prID, oldUserID, newUserID string) error {
	r.logger.Debug("Replacing reviewer %s with %s in PR %s", oldUserID, newUserID, prID)

	tx, err := r.db.Begin(ctx)
	if err != nil {
		r.logger.Error("Failed to begin transaction for reviewer replacement: %v", err)
		return fmt.Errorf("prRepo - ReplaceReviewer - Begin: %w", err)
	}
	defer tx.Rollback(ctx)

	// Удаляем старого ревьювера
	_, err = tx.Exec(ctx, `
		DELETE FROM pr_reviewers 
		WHERE pr_id = $1 AND user_id = $2
	`, prID, oldUserID)
	if err != nil {
		r.logger.Error("Failed to remove old reviewer %s: %v", oldUserID, err)
		return fmt.Errorf("prRepo - ReplaceReviewer - Remove old: %w", err)
	}

	// Добавляем нового ревьювера
	_, err = tx.Exec(ctx, `
		INSERT INTO pr_reviewers (pr_id, user_id) 
		VALUES ($1, $2)
	`, prID, newUserID)
	if err != nil {
		r.logger.Error("Failed to add new reviewer %s: %v", newUserID, err)
		return fmt.Errorf("prRepo - ReplaceReviewer - Add new: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		r.logger.Error("Failed to commit transaction for reviewer replacement: %v", err)
		return fmt.Errorf("prRepo - ReplaceReviewer - Commit: %w", err)
	}

	r.logger.Info("Reviewer replaced successfully: %s -> %s in PR %s", oldUserID, newUserID, prID)
	return nil
}
