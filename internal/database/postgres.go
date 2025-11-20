package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
}

func ConnectDatabase(dsn string) (*DB, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("не удалось спарсить конфиг базы данных: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать пул базы данных: %v", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("не удалось подключиться к базе данных: %v", err)
	}

	return &DB{pool: pool}, nil
}

func (db *DB) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return db.pool.Exec(ctx, query, args...)
}

func (db *DB) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return db.pool.QueryRow(ctx, query, args...)
}

func (db *DB) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	rows, err := db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (db *DB) Close() {
	db.pool.Close()
}
