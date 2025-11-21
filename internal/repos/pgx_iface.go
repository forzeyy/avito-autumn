package repos

import (
	"context"

	"github.com/forzeyy/avito-autumn/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DBInterface interface {
	QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row
	Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	WithinTx(ctx context.Context, txFunc database.TxFunc, txOptions pgx.TxOptions) error
}
