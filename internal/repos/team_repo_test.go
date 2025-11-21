package repos_test

import (
	"context"
	"errors"
	"testing"

	"github.com/forzeyy/avito-autumn/internal/models"
	"github.com/forzeyy/avito-autumn/internal/repos"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestTeamRepo_CreateTeam(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	db := &MockDB{mock: mock}
	repo := repos.NewTeamRepo(db)

	ctx := context.Background()
	teamName := "testteam"

	t.Run("успешное создание команды", func(t *testing.T) {
		mock.ExpectExec(`INSERT INTO teams \(name\) VALUES \(\$1\) ON CONFLICT DO NOTHING`).
			WithArgs(teamName).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err := repo.CreateTeam(ctx, teamName)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка при выполнении запроса", func(t *testing.T) {
		mock.ExpectExec(`INSERT INTO teams \(name\) VALUES \(\$1\) ON CONFLICT DO NOTHING`).
			WithArgs(teamName).
			WillReturnError(errors.New("ошибка базы данных"))

		err := repo.CreateTeam(ctx, teamName)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "не удалось создать команду")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTeamRepo_GetTeam(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	db := &MockDB{mock: mock}
	repo := repos.NewTeamRepo(db)

	ctx := context.Background()
	teamName := "testteam"

	t.Run("успешное получение команды с участниками", func(t *testing.T) {
		expectedMembers := []models.TeamMember{
			{
				UserID:   uuid.New(),
				Username: "user1",
				IsActive: true,
			},
			{
				UserID:   uuid.New(),
				Username: "user2",
				IsActive: false,
			},
		}

		rows := pgxmock.NewRows([]string{"id", "username", "is_active"}).
			AddRow(expectedMembers[0].UserID, expectedMembers[0].Username, expectedMembers[0].IsActive).
			AddRow(expectedMembers[1].UserID, expectedMembers[1].Username, expectedMembers[1].IsActive)

		mock.ExpectQuery(`SELECT id, username, is_active FROM users WHERE team_name = \$1`).
			WithArgs(teamName).
			WillReturnRows(rows)

		team, err := repo.GetTeam(ctx, teamName)

		assert.NoError(t, err)
		assert.Equal(t, teamName, team.Name)
		assert.Equal(t, expectedMembers, team.Members)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка при выполнении запроса", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, username, is_active FROM users WHERE team_name = \$1`).
			WithArgs(teamName).
			WillReturnError(errors.New("ошибка базы данных"))

		team, err := repo.GetTeam(ctx, teamName)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "не удалось получить список участников команды")
		assert.Nil(t, team)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка при сканировании строки", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"id", "username", "is_active"}).
			AddRow("invalid-uuid", "user1", true)

		mock.ExpectQuery(`SELECT id, username, is_active FROM users WHERE team_name = \$1`).
			WithArgs(teamName).
			WillReturnRows(rows)

		team, err := repo.GetTeam(ctx, teamName)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка скана строки участника")
		assert.Nil(t, team)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTeamRepo_IsTeamExists(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	db := &MockDB{mock: mock}
	repo := repos.NewTeamRepo(db)

	ctx := context.Background()
	teamName := "testteam"

	t.Run("команда существует", func(t *testing.T) {
		expected := true

		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM teams WHERE name = \$1\)`).
			WithArgs(teamName).
			WillReturnRows(pgxmock.NewRows([]string{"exists"}).
				AddRow(expected))

		exists, err := repo.IsTeamExists(ctx, teamName)

		assert.NoError(t, err)
		assert.Equal(t, &expected, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("команда не существует", func(t *testing.T) {
		expected := false

		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM teams WHERE name = \$1\)`).
			WithArgs(teamName).
			WillReturnRows(pgxmock.NewRows([]string{"exists"}).
				AddRow(expected))

		exists, err := repo.IsTeamExists(ctx, teamName)

		assert.NoError(t, err)
		assert.Equal(t, &expected, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка при выполнении запроса", func(t *testing.T) {
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM teams WHERE name = \$1\)`).
			WithArgs(teamName).
			WillReturnError(errors.New("ошибка базы данных"))

		exists, err := repo.IsTeamExists(ctx, teamName)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка при получении команды")
		assert.Nil(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
