package db

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxUUID "github.com/vgarvardt/pgx-google-uuid/v5"
	"golang.org/x/exp/rand"
)

type DB struct {
	*Queries
	pool       *pgxpool.Pool
	maxRetries int
	baseDelay  time.Duration
}

func NewPool(ctx context.Context, dbURI string) (*pgxpool.Pool, error) {
	pgxConfig, err := pgxpool.ParseConfig(dbURI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse db uri: %w", err)
	}

	pgxConfig.ConnConfig.Tracer = otelpgx.NewTracer()

	pgxConfig.AfterConnect = func(_ context.Context, conn *pgx.Conn) error {
		pgxUUID.Register(conn.TypeMap())
		return nil
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}

	return pool, nil
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

func (db *DB) TransactionWithIsolationLevel(
	ctx context.Context,
	isolationLevel pgx.TxIsoLevel,
	fn func(*Queries) error,
) error {
	var err error

	for attempt := 1; attempt <= db.maxRetries; attempt++ {
		conn, err := db.pool.Acquire(ctx)
		if err != nil {
			return err
		}

		tx, err := conn.BeginTx(ctx, pgx.TxOptions{
			IsoLevel: isolationLevel,
		})
		if err != nil {
			conn.Release()
			return err
		}

		retryingTx := NewRetryingDBTX(tx, db.maxRetries, db.baseDelay)
		err = fn(New(retryingTx))

		if err != nil {
			_ = tx.Rollback(ctx)
			conn.Release()

			if !isRetryableErr(err) || ctx.Err() != nil {
				return err
			}

			sleepWithBackoff(ctx, attempt, db.baseDelay)
			continue
		}

		err = tx.Commit(ctx)
		conn.Release()

		if err == nil {
			return nil
		}

		if !isRetryableErr(err) || ctx.Err() != nil {
			return err
		}

		sleepWithBackoff(ctx, attempt, db.baseDelay)
	}

	return err
}

func (db *DB) TransactionWithRepeatableRead(ctx context.Context, fn func(*Queries) error) error {
	return db.TransactionWithIsolationLevel(ctx, pgx.RepeatableRead, fn)
}

func isRetryableErr(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.SQLState() {
		case "40P01",
			"40001",
			"08006",
			"08000",
			"08003":
			return true
		}
	}

	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func IsLockConflict(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.SQLState() == "55P03"
	}
	return false
}

func IsSerializationFailure(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.SQLState() == "40001"
	}
	return false
}

func sleepWithBackoff(ctx context.Context, attempt int, baseDelay time.Duration) {
	delay := baseDelay * time.Duration(1<<(attempt-1))
	jitter := time.Duration(rand.Int63n(int64(delay / 2)))
	select {
	case <-time.After(delay + jitter):
	case <-ctx.Done():
	}
}
