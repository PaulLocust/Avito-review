// team.go
package postgresql

import (
	"context"
	"fmt"

	"github.com/PaulLocust/Avito-review/internal/entity"
	"github.com/PaulLocust/Avito-review/internal/repository"
	"github.com/PaulLocust/Avito-review/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

type teamRepo struct {
	db     *pgxpool.Pool
	logger logger.Interface
}

func NewTeamRepository(db *pgxpool.Pool, l logger.Interface) repository.TeamRepository {
	return &teamRepo{db: db, logger: l}
}

func (r *teamRepo) CreateTeam(ctx context.Context, team *entity.Team) error {
	r.logger.Debug("Creating team: %s with %d members", team.Name, len(team.Members))

	tx, err := r.db.Begin(ctx)
	if err != nil {
		r.logger.Error("Failed to begin transaction for team creation: %v", err)
		return fmt.Errorf("teamRepo - CreateTeam - Begin: %w", err)
	}
	defer tx.Rollback(ctx)

	// Создаем команду
	_, err = tx.Exec(ctx, "INSERT INTO teams (name) VALUES ($1) ON CONFLICT (name) DO NOTHING", team.Name)
	if err != nil {
		r.logger.Error("Failed to insert team: %v", err)
		return fmt.Errorf("teamRepo - CreateTeam - Insert team: %w", err)
	}

	// Создаем/обновляем пользователей
	for _, member := range team.Members {
		_, err = tx.Exec(ctx, `
			INSERT INTO users (id, username, team_name, is_active) 
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO UPDATE SET 
				username = EXCLUDED.username,
				team_name = EXCLUDED.team_name,
				is_active = EXCLUDED.is_active
		`, member.UserID, member.Username, team.Name, member.IsActive)
		if err != nil {
			r.logger.Error("Failed to insert user %s: %v", member.UserID, err)
			return fmt.Errorf("teamRepo - CreateTeam - Insert user %s: %w", member.UserID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		r.logger.Error("Failed to commit transaction for team creation: %v", err)
		return fmt.Errorf("teamRepo - CreateTeam - Commit: %w", err)
	}

	r.logger.Info("Team created successfully: %s with %d members", team.Name, len(team.Members))
	return nil
}

func (r *teamRepo) GetTeam(ctx context.Context, name string) (*entity.Team, error) {
	r.logger.Debug("Getting team: %s", name)

	var team entity.Team
	team.Name = name

	rows, err := r.db.Query(ctx, `
		SELECT id, username, is_active 
		FROM users 
		WHERE team_name = $1
		ORDER BY username
	`, name)
	if err != nil {
		r.logger.Error("Failed to query team members: %v", err)
		return nil, fmt.Errorf("teamRepo - GetTeam - Query: %w", err)
	}
	defer rows.Close()

	memberCount := 0
	for rows.Next() {
		var member entity.TeamMember
		err := rows.Scan(&member.UserID, &member.Username, &member.IsActive)
		if err != nil {
			r.logger.Error("Failed to scan team member: %v", err)
			return nil, fmt.Errorf("teamRepo - GetTeam - Scan: %w", err)
		}
		team.Members = append(team.Members, member)
		memberCount++
	}

	if memberCount == 0 {
		r.logger.Warn("Team not found: %s", name)
		return nil, fmt.Errorf("team not found")
	}

	r.logger.Debug("Found team %s with %d members", name, memberCount)
	return &team, nil
}

func (r *teamRepo) TeamExists(ctx context.Context, name string) (bool, error) {
	r.logger.Debug("Checking if team exists: %s", name)

	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM teams WHERE name = $1)", name).Scan(&exists)
	if err != nil {
		r.logger.Error("Failed to check team existence: %v", err)
		return false, fmt.Errorf("teamRepo - TeamExists: %w", err)
	}

	r.logger.Debug("Team %s exists: %v", name, exists)
	return exists, nil
}
