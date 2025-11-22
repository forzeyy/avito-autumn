package repos

import (
	"context"
	"errors"
	"fmt"

	"github.com/forzeyy/avito-autumn/internal/models"
	"github.com/jackc/pgx/v5"
)

type PRRepo interface {
	CreatePR(ctx context.Context, pr *models.PullRequest) error
	GetPRByID(ctx context.Context, prID string) (*models.PullRequest, error)
	GetPRsByReviewer(ctx context.Context, userID string) ([]models.PullRequestShort, error)
	UpdatePRStatus(ctx context.Context, prID string, status models.Status) (*models.PullRequest, error)
	ReplaceReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error
	IsPRMerged(ctx context.Context, prID string) (*bool, error)
	GetTotalPRCount(ctx context.Context) (int, error)
	GetReviewCountByUser(ctx context.Context) ([]models.UserStats, error)
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
	err := prr.db.WithinTx(ctx, txFunc, &pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("ошибка транзакции при создании пулл реквеста: %v", err)
	}

	return nil
}

func (prr *prRepo) GetPRByID(ctx context.Context, prID string) (*models.PullRequest, error) {
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

	reviewersQuery := `
		SELECT reviewer_id
		FROM pr_reviewers
		WHERE pr_id = $1
	`

	reviewersRows, err := prr.db.Query(ctx, reviewersQuery, prID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении ревьюеров: %v", err)
	}
	defer reviewersRows.Close()

	var reviewers []string
	for reviewersRows.Next() {
		var reviewerID string
		err := reviewersRows.Scan(&reviewerID)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании ревьюера: %v", err)
		}
		reviewers = append(reviewers, reviewerID)
	}

	pr.AssignedReviewers = reviewers

	return &pr, nil
}

func (prr *prRepo) GetPRsByReviewer(ctx context.Context, userID string) ([]models.PullRequestShort, error) {
	var prs []models.PullRequestShort

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
		var pr models.PullRequestShort
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

func (prr *prRepo) UpdatePRStatus(ctx context.Context, prID string, status models.Status) (*models.PullRequest, error) {
	query := `
		UPDATE pull_requests
		SET status = $1, merged_at = CASE WHEN $1 = 'MERGED'
										  THEN CURRENT_TIMESTAMP ELSE merged_at END
		WHERE id = $2
	`
	_, err := prr.db.Exec(ctx, query, status, prID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при обновлении статуса пулл реквеста: %v", err)
	}
	return prr.GetPRByID(ctx, prID)
}

func (prr *prRepo) ReplaceReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {
	txFunc := func(tx pgx.Tx) error {
		var count int
		err := tx.QueryRow(ctx,
			"SELECT COUNT(*) FROM pr_reviewers WHERE pr_id = $1 AND reviewer_id = $2",
			prID, oldReviewerID).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check old reviewer existence: %w", err)
		}
		if count == 0 {
			return errors.New("NOT_ASSIGNED")
		}

		err = tx.QueryRow(ctx,
			"SELECT COUNT(*) FROM pr_reviewers WHERE pr_id = $1 AND reviewer_id = $2",
			prID, newReviewerID).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check new reviewer existence: %w", err)
		}

		if count > 0 {
			_, err = tx.Exec(ctx,
				"DELETE FROM pr_reviewers WHERE pr_id = $1 AND reviewer_id = $2",
				prID, oldReviewerID)
			if err != nil {
				return fmt.Errorf("failed to delete old reviewer: %w", err)
			}
		} else {
			result, err := tx.Exec(ctx,
				"UPDATE pr_reviewers SET reviewer_id = $1 WHERE pr_id = $2 AND reviewer_id = $3",
				newReviewerID, prID, oldReviewerID)
			if err != nil {
				return fmt.Errorf("failed to replace reviewer: %w", err)
			}

			rowsAffected := result.RowsAffected()
			if rowsAffected == 0 {
				return errors.New("NOT_ASSIGNED")
			}
		}
		return nil
	}
	err := prr.db.WithinTx(ctx, txFunc, &pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("ошибка транзакции при создании пулл реквеста: %v", err)
	}

	return nil
}

func (prr *prRepo) IsPRMerged(ctx context.Context, prID string) (*bool, error) {
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

func (prr *prRepo) GetTotalPRCount(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM pull_requests`
	err := prr.db.QueryRow(ctx, query).Scan(&count)
	return count, err
}

func (prr *prRepo) GetReviewCountByUser(ctx context.Context) ([]models.UserStats, error) {
	query := `
        SELECT 
            u.user_id,
            u.username,
            COALESCE(review_stats.count, 0) AS review_count,
            u.is_active
        FROM users u
        LEFT JOIN (
            SELECT reviewer_id, COUNT(*) AS count
            FROM pr_reviewers
            GROUP BY reviewer_id
        ) AS review_stats ON u.user_id = review_stats.reviewer_id
        ORDER BY review_count DESC, u.username
    `

	rows, err := prr.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []models.UserStats
	for rows.Next() {
		var s models.UserStats
		err := rows.Scan(&s.UserID, &s.Username, &s.ReviewCount, &s.IsActive)
		if err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}
