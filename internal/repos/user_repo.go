package repos

import (
	"context"
	"errors"
	"fmt"

	"github.com/forzeyy/avito-autumn/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserRepo interface {
	GetUser(ctx context.Context, userID uuid.UUID) (*models.User, error)
	UpsertUser(ctx context.Context, user *models.User) error
	SetUserActive(ctx context.Context, userID uuid.UUID, isActive bool) (*models.User, error)
	GetActiveUsersByTeam(ctx context.Context, teamName string) ([]models.User, error)
}

type userRepo struct {
	db DBInterface
}

func NewUserRepo(db DBInterface) UserRepo {
	return &userRepo{
		db: db,
	}
}

func (ur *userRepo) GetUser(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var user models.User

	query := `
		SELECT id, username, team_name, is_active
		FROM users
		WHERE id = $1
	`
	row := ur.db.QueryRow(ctx, query, userID)

	err := row.Scan(&user.ID, &user.Username, &user.TeamName, &user.IsActive)
	if err == pgx.ErrNoRows {
		return nil, errors.New("пользователь не найден")
	}
	if err != nil {
		return nil, fmt.Errorf("не удалось получить пользователя: %v", err)
	}
	return &user, nil
}

func (ur *userRepo) UpsertUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, username, team_name, is_active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id)
		DO UPDATE SET
			username = EXCLUDED.username,
			team_name = EXCLUDED.team_name,
			is_active = EXCLUDED.is_active
	`

	_, err := ur.db.Exec(ctx, query, user.ID, user.Username, user.TeamName, user.IsActive)
	if err != nil {
		return fmt.Errorf("ошибка при создании/обновлении пользователя: %v", err)
	}
	return nil
}

func (ur *userRepo) SetUserActive(ctx context.Context, userID uuid.UUID, isActive bool) (*models.User, error) {
	var user models.User

	query := `
		UPDATE users
		SET is_active = $1
		WHERE id = $2
		RETURNING id, username, team_name, is_active
	`
	row := ur.db.QueryRow(ctx, query, isActive, userID)
	err := row.Scan(&user.ID, &user.Username, &user.TeamName, &user.IsActive)
	if err == pgx.ErrNoRows {
		return nil, errors.New("пользователь не найден")
	}
	if err != nil {
		return nil, fmt.Errorf("не получилось изменить статус пользователя: %v", err)
	}

	return &user, nil
}

func (ur *userRepo) GetActiveUsersByTeam(ctx context.Context, teamName string) ([]models.User, error) {
	var activeUsers []models.User

	query := `
		SELECT id, username, team_name, is_active
		FROM users
		WHERE team_name = $1
	`
	rows, err := ur.db.Query(ctx, query, teamName)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении активных юзеров команды %v: %v", teamName, err)
	}

	defer rows.Close()
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.TeamName, &user.IsActive)
		if err != nil {
			return nil, fmt.Errorf("ошибка при скане строки: %v", err)
		}
		activeUsers = append(activeUsers, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка сканирования строк: %v", err)
	}

	return activeUsers, nil
}
