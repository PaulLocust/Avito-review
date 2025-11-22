// user.go
package postgresql

import (
	"context"
	"fmt"

	"github.com/PaulLocust/Avito-review/internal/entity"
	"github.com/PaulLocust/Avito-review/internal/repository"
	"github.com/PaulLocust/Avito-review/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepo struct {
	db     *pgxpool.Pool
	logger logger.Interface
}

func NewUserRepository(db *pgxpool.Pool, l logger.Interface) repository.UserRepository {
	return &userRepo{db: db, logger: l}
}

func (r *userRepo) CreateOrUpdateUser(ctx context.Context, user *entity.User) error {
	r.logger.Debug("Creating or updating user: %s", user.ID)

	_, err := r.db.Exec(ctx, `
		INSERT INTO users (id, username, team_name, is_active) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET 
			username = EXCLUDED.username,
			team_name = EXCLUDED.team_name,
			is_active = EXCLUDED.is_active
	`, user.ID, user.Username, user.TeamName, user.IsActive)

	if err != nil {
		r.logger.Error("Failed to create or update user %s: %v", user.ID, err)
		return fmt.Errorf("userRepo - CreateOrUpdateUser: %w", err)
	}

	r.logger.Debug("User created or updated successfully: %s", user.ID)
	return nil
}

func (r *userRepo) GetUser(ctx context.Context, id string) (*entity.User, error) {
	r.logger.Debug("Getting user: %s", id)

	var user entity.User
	err := r.db.QueryRow(ctx, `
		SELECT id, username, team_name, is_active 
		FROM users 
		WHERE id = $1
	`, id).Scan(&user.ID, &user.Username, &user.TeamName, &user.IsActive)

	if err != nil {
		r.logger.Warn("User not found: %s", id)
		return nil, fmt.Errorf("userRepo - GetUser: %w", err)
	}

	r.logger.Debug("User found: %s (team: %s, active: %v)", user.ID, user.TeamName, user.IsActive)
	return &user, nil
}

func (r *userRepo) UpdateUser(ctx context.Context, user *entity.User) error {
	r.logger.Debug("Updating user: %s", user.ID)

	_, err := r.db.Exec(ctx, `
		UPDATE users 
		SET username = $1, team_name = $2, is_active = $3 
		WHERE id = $4
	`, user.Username, user.TeamName, user.IsActive, user.ID)

	if err != nil {
		r.logger.Error("Failed to update user %s: %v", user.ID, err)
		return fmt.Errorf("userRepo - UpdateUser: %w", err)
	}

	r.logger.Debug("User updated successfully: %s", user.ID)
	return nil
}

func (r *userRepo) GetActiveUsersByTeam(ctx context.Context, teamName string, excludeUserID string) ([]entity.User, error) {
	r.logger.Debug("Getting active users by team: %s, exclude: %s", teamName, excludeUserID)

	var users []entity.User

	query := `
		SELECT id, username, team_name, is_active 
		FROM users 
		WHERE team_name = $1 AND is_active = true
	`
	args := []interface{}{teamName}

	if excludeUserID != "" {
		query += " AND id != $2"
		args = append(args, excludeUserID)
	}

	query += " ORDER BY id"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to query active users by team: %v", err)
		return nil, fmt.Errorf("userRepo - GetActiveUsersByTeam - Query: %w", err)
	}
	defer rows.Close()

	userCount := 0
	for rows.Next() {
		var user entity.User
		err := rows.Scan(&user.ID, &user.Username, &user.TeamName, &user.IsActive)
		if err != nil {
			r.logger.Error("Failed to scan user: %v", err)
			return nil, fmt.Errorf("userRepo - GetActiveUsersByTeam - Scan: %w", err)
		}
		users = append(users, user)
		userCount++
	}

	r.logger.Debug("Found %d active users in team %s", userCount, teamName)
	return users, nil
}
