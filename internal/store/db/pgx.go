package db

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/exp/rand"
)

type DB struct {
	*Queries
	pool       *pgxpool.Pool
	maxRetries int
	baseDelay  time.Duration
}

func NewDB(pool *pgxpool.Pool, maxRetries int, baseDelay time.Duration) *DB {
	retryingDB := NewRetryingDBTX(pool, maxRetries, baseDelay)

	return &DB{
		Queries:    New(retryingDB),
		pool:       pool,
		maxRetries: maxRetries,
		baseDelay:  baseDelay,
	}
}

type RetryingDBTX struct {
	db         DBTX
	maxRetries int
	baseDelay  time.Duration
}

func NewRetryingDBTX(db DBTX, maxRetries int, baseDelay time.Duration) *RetryingDBTX {
	return &RetryingDBTX{
		db:         db,
		maxRetries: maxRetries,
		baseDelay:  baseDelay,
	}
}

func (r *RetryingDBTX) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	var result pgconn.CommandTag
	var err error

	for attempt := 1; attempt <= r.maxRetries; attempt++ {
		result, err = r.db.Exec(ctx, sql, args...)
		if err == nil || !isRetryableErr(err) || ctx.Err() != nil {
			break
		}
		sleepWithBackoff(ctx, attempt, r.baseDelay)
	}

	return result, err
}

func (r *RetryingDBTX) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	var rows pgx.Rows
	var err error

	for attempt := 1; attempt <= r.maxRetries; attempt++ {
		rows, err = r.db.Query(ctx, sql, args...)
		if err == nil || !isRetryableErr(err) || ctx.Err() != nil {
			break
		}
		sleepWithBackoff(ctx, attempt, r.baseDelay)
	}

	return rows, err
}

func (r *RetryingDBTX) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return r.db.QueryRow(ctx, sql, args...)
}

func (db *DB) TransactionWithRetry(ctx context.Context, fn func(*Queries) error) error {
	var err error

	for attempt := 1; attempt <= db.maxRetries; attempt++ {
		err = pgx.BeginFunc(ctx, db.pool, func(tx pgx.Tx) error {
			retryingTx := NewRetryingDBTX(tx, db.maxRetries, db.baseDelay)
			return fn(New(retryingTx))
		})

		if err == nil {
			return nil
		}

		if !isRetryableErr(err) || ctx.Err() != nil {
			break
		}

		sleepWithBackoff(ctx, attempt, db.baseDelay)
	}

	return err
}

func isRetryableErr(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.SQLState() {
		case "40P01",
			"08006",
			"08000",
			"08003":
			return true
		}
	}

	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func sleepWithBackoff(ctx context.Context, attempt int, baseDelay time.Duration) {
	delay := baseDelay * time.Duration(1<<(attempt-1))
	jitter := time.Duration(rand.Int63n(int64(delay / 2)))
	select {
	case <-time.After(delay + jitter):
	case <-ctx.Done():
	}
}
