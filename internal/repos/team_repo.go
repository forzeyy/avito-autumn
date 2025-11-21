package repos

import (
	"context"
	"fmt"

	"github.com/forzeyy/avito-autumn/internal/models"
)

type TeamRepo interface {
	CreateTeam(ctx context.Context, teamName string) error
	GetTeam(ctx context.Context, teamName string) (*models.Team, error)
	IsTeamExists(ctx context.Context, teamName string) (*bool, error)
}

type teamRepo struct {
	db DBInterface
}

func NewTeamRepo(db DBInterface) TeamRepo {
	return &teamRepo{
		db: db,
	}
}

func (tr *teamRepo) CreateTeam(ctx context.Context, teamName string) error {
	query := `
		INSERT INTO teams (name)
		VALUES ($1)
		ON CONFLICT DO NOTHING
	`
	_, err := tr.db.Exec(ctx, query, teamName)
	if err != nil {
		return fmt.Errorf("не удалось создать команду: %v", err)
	}
	return nil
}

func (tr *teamRepo) GetTeam(ctx context.Context, teamName string) (*models.Team, error) {
	var team models.Team

	query := `
		SELECT id, username, is_active
		FROM users
		WHERE team_name = $1
	`
	rows, err := tr.db.Query(ctx, query, teamName)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить список участников команды %v: %v", teamName, err)
	}

	defer rows.Close()
	for rows.Next() {
		var member models.TeamMember
		err := rows.Scan(&member.UserID, &member.Username, &member.IsActive)
		if err != nil {
			return nil, fmt.Errorf("ошибка скана строки участника: %v", err)
		}
		team.Members = append(team.Members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка скана строк: %v", err)
	}

	team.Name = teamName
	return &team, nil
}

func (tr *teamRepo) IsTeamExists(ctx context.Context, teamName string) (*bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(SELECT 1
					  FROM teams
					  WHERE name = $1)
	`
	row := tr.db.QueryRow(ctx, query, teamName)
	err := row.Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении команды: %v", err)
	}

	return &exists, nil
}
