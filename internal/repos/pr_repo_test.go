package repos_test

import (
	"context"
	"errors"
	"testing"

	"github.com/forzeyy/avito-autumn/internal/models"
	"github.com/forzeyy/avito-autumn/internal/repos"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestPRRepo_CreatePR(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	db := &MockDB{mock: mock}
	repo := repos.NewPRRepo(db)

	ctx := context.Background()
	pr := &models.PullRequest{
		ID:                uuid.New(),
		Name:              "test_pr",
		AuthorID:          uuid.New(),
		AssignedReviewers: []uuid.UUID{uuid.New(), uuid.New()},
		Status:            models.StatusOpen,
	}

	t.Run("успешное создание пулл реквеста с ревьюерами", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO pull_requests \(id, name, author_id\) VALUES \(\$1, \$2, \$3\)`).
			WithArgs(pr.ID, pr.Name, pr.AuthorID).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		for _, reviewerID := range pr.AssignedReviewers {
			mock.ExpectExec(`INSERT INTO pr_reviewers \(pr_id, reviewer_id\) VALUES \(\$1, \$2\)`).
				WithArgs(pr.ID, reviewerID).
				WillReturnResult(pgxmock.NewResult("INSERT", 1))
		}
		mock.ExpectCommit()

		err := repo.CreatePR(ctx, pr)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка при создании пулл реквеста", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO pull_requests \(id, name, author_id\) VALUES \(\$1, \$2, \$3\)`).
			WithArgs(pr.ID, pr.Name, pr.AuthorID).
			WillReturnError(errors.New("ошибка базы данных"))
		mock.ExpectRollback()

		err := repo.CreatePR(ctx, pr)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка транзакции при создании пулл реквеста")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка при добавлении ревьюера", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO pull_requests \(id, name, author_id\) VALUES \(\$1, \$2, \$3\)`).
			WithArgs(pr.ID, pr.Name, pr.AuthorID).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		mock.ExpectExec(`INSERT INTO pr_reviewers \(pr_id, reviewer_id\) VALUES \(\$1, \$2\)`).
			WithArgs(pr.ID, pr.AssignedReviewers[0]).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		mock.ExpectExec(`INSERT INTO pr_reviewers \(pr_id, reviewer_id\) VALUES \(\$1, \$2\)`).
			WithArgs(pr.ID, pr.AssignedReviewers[1]).
			WillReturnError(errors.New("ошибка добавления ревьюера"))
		mock.ExpectRollback()

		err := repo.CreatePR(ctx, pr)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка транзакции при создании пулл реквеста")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPRRepo_GetPRByID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	db := &MockDB{mock: mock}
	repo := repos.NewPRRepo(db)

	ctx := context.Background()
	prID := uuid.New()

	t.Run("успешное получение пулл реквеста", func(t *testing.T) {
		expectedPR := &models.PullRequest{
			ID:       prID,
			Name:     "test_pr",
			AuthorID: uuid.New(),
			Status:   models.StatusOpen,
		}

		mock.ExpectQuery(`SELECT id, name, author_id, status FROM pull_requests WHERE id = \$1`).
			WithArgs(prID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "name", "author_id", "status"}).
				AddRow(expectedPR.ID, expectedPR.Name, expectedPR.AuthorID, expectedPR.Status))

		pr, err := repo.GetPRByID(ctx, prID)

		assert.NoError(t, err)
		assert.Equal(t, expectedPR, pr)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("пулл реквест не найден", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, name, author_id, status FROM pull_requests WHERE id = \$1`).
			WithArgs(prID).
			WillReturnError(pgx.ErrNoRows)

		pr, err := repo.GetPRByID(ctx, prID)

		assert.Error(t, err)
		assert.Equal(t, "пулл реквест не найден", err.Error())
		assert.Nil(t, pr)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка при выполнении запроса", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, name, author_id, status FROM pull_requests WHERE id = \$1`).
			WithArgs(prID).
			WillReturnError(errors.New("ошибка базы данных"))

		pr, err := repo.GetPRByID(ctx, prID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка при получении пулл реквеста")
		assert.Nil(t, pr)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPRRepo_GetPRsByReviewer(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	db := &MockDB{mock: mock}
	repo := repos.NewPRRepo(db)

	ctx := context.Background()
	userID := uuid.New()

	t.Run("успешное получение пулл реквестов по ревьюеру", func(t *testing.T) {
		expectedPRs := []models.PullRequest{
			{
				ID:       uuid.New(),
				Name:     "pr_1",
				AuthorID: uuid.New(),
				Status:   models.StatusOpen,
			},
			{
				ID:       uuid.New(),
				Name:     "pr_2",
				AuthorID: uuid.New(),
				Status:   models.StatusMerged,
			},
		}

		rows := pgxmock.NewRows([]string{"id", "name", "author_id", "status"}).
			AddRow(expectedPRs[0].ID, expectedPRs[0].Name, expectedPRs[0].AuthorID, expectedPRs[0].Status).
			AddRow(expectedPRs[1].ID, expectedPRs[1].Name, expectedPRs[1].AuthorID, expectedPRs[1].Status)

		mock.ExpectQuery(`SELECT p\.id, p\.name, p\.author_id, p\.status FROM pull_requests p JOIN pr_reviewers r ON p\.id = r\.pr_id WHERE r\.reviewer_id = \$1`).
			WithArgs(userID).
			WillReturnRows(rows)

		prs, err := repo.GetPRsByReviewer(ctx, userID)

		assert.NoError(t, err)
		assert.Equal(t, expectedPRs, prs)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка при выполнении запроса", func(t *testing.T) {
		mock.ExpectQuery(`SELECT p\.id, p\.name, p\.author_id, p\.status FROM pull_requests p JOIN pr_reviewers r ON p\.id = r\.pr_id WHERE r\.reviewer_id = \$1`).
			WithArgs(userID).
			WillReturnError(errors.New("ошибка базы данных"))

		prs, err := repo.GetPRsByReviewer(ctx, userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка при получении пулл реквестов")
		assert.Nil(t, prs)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка при сканировании строки", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"id", "name", "author_id", "status"}).
			AddRow("invalid-uuid", "pr_1", uuid.New(), models.StatusOpen)

		mock.ExpectQuery(`SELECT p\.id, p\.name, p\.author_id, p\.status FROM pull_requests p JOIN pr_reviewers r ON p\.id = r\.pr_id WHERE r\.reviewer_id = \$1`).
			WithArgs(userID).
			WillReturnRows(rows)

		prs, err := repo.GetPRsByReviewer(ctx, userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка при скане строки")
		assert.Nil(t, prs)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPRRepo_UpdatePRStatus(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	db := &MockDB{mock: mock}
	repo := repos.NewPRRepo(db)

	ctx := context.Background()
	prID := uuid.New()
	status := models.StatusMerged

	t.Run("успешное обновление статуса пулл реквеста", func(t *testing.T) {
		updatedPR := &models.PullRequest{
			ID:       prID,
			Name:     "upd_pr",
			AuthorID: uuid.New(),
			Status:   status,
		}

		mock.ExpectExec(`UPDATE pull_requests SET status = \$1, merged_at = CASE WHEN \$1 = 'MERGED' THEN CURRENT_TIMESTAMP ELSE merged_at END WHERE pull_request_id = \$2`).
			WithArgs(status, prID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		mock.ExpectQuery(`SELECT id, name, author_id, status FROM pull_requests WHERE id = \$1`).
			WithArgs(prID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "name", "author_id", "status"}).
				AddRow(updatedPR.ID, updatedPR.Name, updatedPR.AuthorID, updatedPR.Status))

		pr, err := repo.UpdatePRStatus(ctx, prID, status)

		assert.NoError(t, err)
		assert.Equal(t, updatedPR, pr)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка при обновлении статуса", func(t *testing.T) {
		mock.ExpectExec(`UPDATE pull_requests SET status = \$1, merged_at = CASE WHEN \$1 = 'MERGED' THEN CURRENT_TIMESTAMP ELSE merged_at END WHERE pull_request_id = \$2`).
			WithArgs(status, prID).
			WillReturnError(errors.New("ошибка обновления"))

		pr, err := repo.UpdatePRStatus(ctx, prID, status)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка при обновлении статуса пулл реквеста")
		assert.Nil(t, pr)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка при получении обновленного пулл реквеста", func(t *testing.T) {
		mock.ExpectExec(`UPDATE pull_requests SET status = \$1, merged_at = CASE WHEN \$1 = 'MERGED' THEN CURRENT_TIMESTAMP ELSE merged_at END WHERE pull_request_id = \$2`).
			WithArgs(status, prID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		mock.ExpectQuery(`SELECT id, name, author_id, status FROM pull_requests WHERE id = \$1`).
			WithArgs(prID).
			WillReturnError(errors.New("ошибка получения"))

		pr, err := repo.UpdatePRStatus(ctx, prID, status)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка при получении пулл реквеста")
		assert.Nil(t, pr)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPRRepo_ReplaceReviewer(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	db := &MockDB{mock: mock}
	repo := repos.NewPRRepo(db)

	ctx := context.Background()
	prID := uuid.New()
	oldReviewerID := uuid.New()
	newReviewerID := uuid.New()

	t.Run("успешная замена ревьюера", func(t *testing.T) {
		mock.ExpectExec(`UPDATE pr_reviewers SET reviewer_id = \$1 WHERE pr_id = \$2 AND reviewer_id = \$3`).
			WithArgs(newReviewerID, prID, oldReviewerID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := repo.ReplaceReviewer(ctx, prID, oldReviewerID, newReviewerID)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка при замене ревьюера", func(t *testing.T) {
		mock.ExpectExec(`UPDATE pr_reviewers SET reviewer_id = \$1 WHERE pr_id = \$2 AND reviewer_id = \$3`).
			WithArgs(newReviewerID, prID, oldReviewerID).
			WillReturnError(errors.New("ошибка базы данных"))

		err := repo.ReplaceReviewer(ctx, prID, oldReviewerID, newReviewerID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "не удалось заменить ревьюера")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPRRepo_IsPRMerged(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	db := &MockDB{mock: mock}
	repo := repos.NewPRRepo(db)

	ctx := context.Background()
	prID := uuid.New()

	t.Run("пулл реквест смерджен", func(t *testing.T) {
		expected := true

		mock.ExpectQuery(`SELECT status FROM pull_requests WHERE id = \$1`).
			WithArgs(prID).
			WillReturnRows(pgxmock.NewRows([]string{"status"}).
				AddRow(string(models.StatusMerged)))

		isMerged, err := repo.IsPRMerged(ctx, prID)

		assert.NoError(t, err)
		assert.Equal(t, &expected, isMerged)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("пулл реквест не смерджен", func(t *testing.T) {
		expected := false

		mock.ExpectQuery(`SELECT status FROM pull_requests WHERE id = \$1`).
			WithArgs(prID).
			WillReturnRows(pgxmock.NewRows([]string{"status"}).
				AddRow(string(models.StatusOpen)))

		isMerged, err := repo.IsPRMerged(ctx, prID)

		assert.NoError(t, err)
		assert.Equal(t, &expected, isMerged)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("пулл реквест не найден", func(t *testing.T) {
		mock.ExpectQuery(`SELECT status FROM pull_requests WHERE id = \$1`).
			WithArgs(prID).
			WillReturnError(pgx.ErrNoRows)

		isMerged, err := repo.IsPRMerged(ctx, prID)

		assert.Error(t, err)
		assert.Equal(t, "пулл реквест не найден", err.Error())
		assert.Nil(t, isMerged)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ошибка при выполнении запроса", func(t *testing.T) {
		mock.ExpectQuery(`SELECT status FROM pull_requests WHERE id = \$1`).
			WithArgs(prID).
			WillReturnError(errors.New("ошибка базы данных"))

		isMerged, err := repo.IsPRMerged(ctx, prID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "не удалось получить пулл реквест")
		assert.Nil(t, isMerged)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
