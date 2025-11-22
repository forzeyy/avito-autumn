package repos_test

import (
	"context"
	"errors"
	"testing"

	"github.com/forzeyy/avito-autumn/internal/database"
	"github.com/forzeyy/avito-autumn/internal/models"
	"github.com/forzeyy/avito-autumn/internal/repos"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

type MockDB struct {
	mock pgxmock.PgxPoolIface
}

func (m *MockDB) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return m.mock.QueryRow(ctx, query, args...)
}

func (m *MockDB) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	return m.mock.Query(ctx, query, args...)
}

func (m *MockDB) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return m.mock.Exec(ctx, query, args...)
}

func (m *MockDB) BeginTx(ctx context.Context, txOptions *pgx.TxOptions) (pgx.Tx, error) {
	return m.mock.BeginTx(ctx, *txOptions)
}

func (m *MockDB) WithinTx(ctx context.Context, txFunc database.TxFunc, txOptions *pgx.TxOptions) error {
	tx, err := m.mock.BeginTx(ctx, *txOptions)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	err = txFunc(tx)
	if err != nil {
		rbErr := tx.Rollback(ctx)
		if rbErr != nil {
			return err
		}
		return err
	}

	return tx.Commit(ctx)
}

func TestUserRepo_GetUser(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	db := &MockDB{mock: mock}
	repo := repos.NewUserRepo(db)
	ctx := context.Background()
	userID := "userid"

	t.Run("успешное получение пользователя", func(t *testing.T) {
		expectedUser := &models.User{
			ID:       userID,
			Username: "testtt",
			TeamName: "cool_team",
			IsActive: true,
		}

		mock.ExpectQuery(`SELECT id, username, team_name, is_active FROM users WHERE id = \$1`).
			WithArgs(userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "username", "team_name", "is_active"}).
				AddRow(expectedUser.ID, expectedUser.Username, expectedUser.TeamName, expectedUser.IsActive))

		user, err := repo.GetUser(ctx, userID)

		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("пользователь не найден", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, username, team_name, is_active FROM users WHERE id = \$1`).
			WithArgs(userID).
			WillReturnError(pgx.ErrNoRows)

		user, err := repo.GetUser(ctx, userID)

		assert.Error(t, err)
		assert.Equal(t, "пользователь не найден", err.Error())
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка при выполнении запроса", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, username, team_name, is_active FROM users WHERE id = \$1`).
			WithArgs(userID).
			WillReturnError(errors.New("ошибка базы данных"))

		user, err := repo.GetUser(ctx, userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "не удалось получить пользователя")
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepo_UpsertUser(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	db := &MockDB{mock: mock}
	repo := repos.NewUserRepo(db)

	ctx := context.Background()
	user := &models.User{
		ID:       "cool_id",
		Username: "cool_username",
		TeamName: "cool_team",
		IsActive: true,
	}

	t.Run("успешное создание/обновление пользователя", func(t *testing.T) {
		mock.ExpectExec(`INSERT INTO users \(id, username, team_name, is_active\) VALUES \(\$1, \$2, \$3, \$4\) ON CONFLICT \(id\) DO UPDATE SET username = EXCLUDED\.username, team_name = EXCLUDED\.team_name, is_active = EXCLUDED\.is_active`).
			WithArgs(user.ID, user.Username, user.TeamName, user.IsActive).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err := repo.UpsertUser(ctx, user)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка при выполнении запроса", func(t *testing.T) {
		mock.ExpectExec(`INSERT INTO users \(id, username, team_name, is_active\) VALUES \(\$1, \$2, \$3, \$4\) ON CONFLICT \(id\) DO UPDATE SET username = EXCLUDED\.username, team_name = EXCLUDED\.team_name, is_active = EXCLUDED\.is_active`).
			WithArgs(user.ID, user.Username, user.TeamName, user.IsActive).
			WillReturnError(errors.New("ошибка базы данных"))

		err := repo.UpsertUser(ctx, user)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка при создании/обновлении пользователя")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepo_SetUserActive(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	db := &MockDB{mock: mock}
	repo := repos.NewUserRepo(db)

	ctx := context.Background()
	userID := "userid"
	isActive := false

	t.Run("успешное изменение статуса пользователя", func(t *testing.T) {
		expectedUser := &models.User{
			ID:       userID,
			Username: "good_username",
			TeamName: "good_teamname",
			IsActive: isActive,
		}

		mock.ExpectQuery(`UPDATE users SET is_active = \$1 WHERE id = \$2 RETURNING id, username, team_name, is_active`).
			WithArgs(isActive, userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "username", "team_name", "is_active"}).
				AddRow(expectedUser.ID, expectedUser.Username, expectedUser.TeamName, expectedUser.IsActive))

		user, err := repo.SetUserActive(ctx, userID, isActive)

		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("пользователь не найден", func(t *testing.T) {
		mock.ExpectQuery(`UPDATE users SET is_active = \$1 WHERE id = \$2 RETURNING id, username, team_name, is_active`).
			WithArgs(isActive, userID).
			WillReturnError(pgx.ErrNoRows)

		user, err := repo.SetUserActive(ctx, userID, isActive)

		assert.Error(t, err)
		assert.Equal(t, "пользователь не найден", err.Error())
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка при выполнении запроса", func(t *testing.T) {
		mock.ExpectQuery(`UPDATE users SET is_active = \$1 WHERE id = \$2 RETURNING id, username, team_name, is_active`).
			WithArgs(isActive, userID).
			WillReturnError(errors.New("ошибка базы данных"))

		user, err := repo.SetUserActive(ctx, userID, isActive)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "не получилось изменить статус пользователя")
		assert.Nil(t, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepo_GetActiveUsersByTeam(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	db := &MockDB{mock: mock}
	repo := repos.NewUserRepo(db)

	ctx := context.Background()
	teamName := "team123"

	t.Run("успешное получение активных пользователей команды", func(t *testing.T) {
		expectedUsers := []models.User{
			{
				ID:       "userid1",
				Username: "user1",
				TeamName: teamName,
				IsActive: true,
			},
			{
				ID:       "userid2",
				Username: "user2",
				TeamName: teamName,
				IsActive: false,
			},
		}

		rows := pgxmock.NewRows([]string{"id", "username", "team_name", "is_active"}).
			AddRow(expectedUsers[0].ID, expectedUsers[0].Username, expectedUsers[0].TeamName, expectedUsers[0].IsActive).
			AddRow(expectedUsers[1].ID, expectedUsers[1].Username, expectedUsers[1].TeamName, expectedUsers[1].IsActive)

		mock.ExpectQuery(`SELECT id, username, team_name, is_active FROM users WHERE team_name = \$1`).
			WithArgs(teamName).
			WillReturnRows(rows)

		users, err := repo.GetActiveUsersByTeam(ctx, teamName)

		assert.NoError(t, err)
		assert.Equal(t, expectedUsers, users)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка при выполнении запроса", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, username, team_name, is_active FROM users WHERE team_name = \$1`).
			WithArgs(teamName).
			WillReturnError(errors.New("ошибка базы данных"))

		users, err := repo.GetActiveUsersByTeam(ctx, teamName)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка при получении активных юзеров команды")
		assert.Nil(t, users)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка при сканировании строки", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"id", "username", "team_name", "is_active"}).
			AddRow("invalid-uuid", "user1", teamName, true)

		mock.ExpectQuery(`SELECT id, username, team_name, is_active FROM users WHERE team_name = \$1`).
			WithArgs(teamName).
			WillReturnRows(rows)

		users, err := repo.GetActiveUsersByTeam(ctx, teamName)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка при скане строки")
		assert.Nil(t, users)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
