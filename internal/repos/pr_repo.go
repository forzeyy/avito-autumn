package repos

import (
	"context"
	"errors"
	"fmt"

	"github.com/forzeyy/avito-autumn/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type PRRepo interface {
	CreatePR(ctx context.Context, pr *models.PullRequest) error
	GetPRByID(ctx context.Context, prID uuid.UUID) (*models.PullRequest, error)
	GetPRsByReviewer(ctx context.Context, userID uuid.UUID) ([]models.PullRequest, error)
	UpdatePRStatus(ctx context.Context, prID uuid.UUID, status models.Status) (*models.PullRequest, error)
	ReplaceReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID uuid.UUID) error
	IsPRMerged(ctx context.Context, prID uuid.UUID) (*bool, error)
}

type prRepo struct {
	db DBInterface
}

func NewPRRepo(db DBInterface) PRRepo {
	return &prRepo{
		db: db,
	}
}

func (prr *prRepo) CreatePR(ctx context.Context, pr *models.PullRequest) error {
	txFunc := func(tx pgx.Tx) error {
		query := `
			INSERT INTO pull_requests (id, name, author_id)
			VALUES ($1, $2, $3)
		`

		_, err := tx.Exec(ctx, query, pr.ID, pr.Name, pr.AuthorID)
		if err != nil {
			return fmt.Errorf("ошибка при создании пулл реквеста: %v", err)
		}

		if len(pr.AssignedReviewers) > 0 {
			query := `
				INSERT INTO pr_reviewers (pr_id, reviewer_id)
				VALUES ($1, $2)
			`
			for _, reviewerID := range pr.AssignedReviewers {
				_, err := tx.Exec(ctx, query, pr.ID, reviewerID)
				if err != nil {
					return fmt.Errorf("ошибка при добавлении ревьюера: %v", err)
				}
			}
		}
		return nil
	}
	err := prr.db.WithinTx(ctx, txFunc, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("ошибка транзакции при создании пулл реквеста: %v", err)
	}

	return nil
}

func (prr *prRepo) GetPRByID(ctx context.Context, prID uuid.UUID) (*models.PullRequest, error) {
	var pr models.PullRequest

	query := `
		SELECT id, name, author_id, status
		FROM pull_requests
		WHERE id = $1
	`

	row := prr.db.QueryRow(ctx, query, prID)
	err := row.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status)
	if err == pgx.ErrNoRows {
		return nil, errors.New("пулл реквест не найден")
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении пулл реквеста: %v", err)
	}
	return &pr, nil
}

func (prr *prRepo) GetPRsByReviewer(ctx context.Context, userID uuid.UUID) ([]models.PullRequest, error) {
	var prs []models.PullRequest

	query := `
		SELECT p.id, p.name, p.author_id, p.status
		FROM pull_requests p
		JOIN pr_reviewers r ON p.id = r.pr_id
		WHERE r.reviewer_id = $1
	`

	rows, err := prr.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении пулл реквестов: %v", err)
	}

	defer rows.Close()
	for rows.Next() {
		var pr models.PullRequest
		err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status)
		if err != nil {
			return nil, fmt.Errorf("ошибка при скане строки: %v", err)
		}
		prs = append(prs, pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при скане строк: %v", err)
	}
	return prs, nil
}

func (prr *prRepo) UpdatePRStatus(ctx context.Context, prID uuid.UUID, status models.Status) (*models.PullRequest, error) {
	query := `
		UPDATE pull_requests
		SET status = $1, merged_at = CASE WHEN $1 = 'MERGED'
										  THEN CURRENT_TIMESTAMP ELSE merged_at END
		WHERE pull_request_id = $2
	`
	_, err := prr.db.Exec(ctx, query, status, prID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при обновлении статуса пулл реквеста: %v", err)
	}
	return prr.GetPRByID(ctx, prID)
}

func (prr *prRepo) ReplaceReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID uuid.UUID) error {
	query := `
		UPDATE pr_reviewers
		SET reviewer_id = $1
		WHERE pr_id = $2 AND reviewer_id = $3
	`
	_, err := prr.db.Exec(ctx, query, newReviewerID, prID, oldReviewerID)
	if err != nil {
		return fmt.Errorf("не удалось заменить ревьюера: %v", err)
	}
	return nil
}

func (prr *prRepo) IsPRMerged(ctx context.Context, prID uuid.UUID) (*bool, error) {
	var status string
	query := `
		SELECT status
		FROM pull_requests
		WHERE id = $1
	`
	row := prr.db.QueryRow(ctx, query, prID)
	err := row.Scan(&status)
	if err == pgx.ErrNoRows {
		return nil, errors.New("пулл реквест не найден")
	}
	if err != nil {
		return nil, fmt.Errorf("не удалось получить пулл реквест: %v", err)
	}
	isMerged := status == string(models.StatusMerged)
	return &isMerged, nil
}
