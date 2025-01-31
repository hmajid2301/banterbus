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
	pool *pgxpool.Pool
}

func NewDB(pool *pgxpool.Pool, maxRetries int, baseDelay time.Duration) *DB {
	retryingDB := NewRetryingDBTX(pool, maxRetries, baseDelay)

	return &DB{
		Queries: New(retryingDB),
		pool:    pool,
	}
}

// TODO: Deal with transactions instead of retrying each query

type RetryingDBTX struct {
	pool       *pgxpool.Pool
	maxRetries int
	baseDelay  time.Duration
}

func NewRetryingDBTX(pool *pgxpool.Pool, maxRetries int, baseDelay time.Duration) *RetryingDBTX {
	return &RetryingDBTX{
		pool:       pool,
		maxRetries: maxRetries,
		baseDelay:  baseDelay,
	}
}

func (r *RetryingDBTX) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	var result pgconn.CommandTag
	err := r.retry(ctx, func() error {
		var innerErr error
		result, innerErr = r.pool.Exec(ctx, sql, args...)
		return innerErr
	})
	return result, err
}

func (r *RetryingDBTX) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	var rows pgx.Rows
	err := r.retry(ctx, func() error {
		var innerErr error
		rows, innerErr = r.pool.Query(ctx, sql, args...)
		return innerErr
	})
	return rows, err
}

func (r *RetryingDBTX) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	var rows pgx.Row
	_ = r.retry(ctx, func() error {
		rows = r.pool.QueryRow(ctx, sql, args...)
		return nil
	})
	return rows
}

func (r *RetryingDBTX) retry(ctx context.Context, fn func() error) error {
	for attempt := 1; attempt <= r.maxRetries; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}
		if !isRetryableErr(err) || ctx.Err() != nil {
			return err
		}

		delay := r.baseDelay * time.Duration(1<<(attempt-1))
		jitter := time.Duration(rand.Int63n(int64(delay / 2)))

		select {
		case <-time.After(delay + jitter):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
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
