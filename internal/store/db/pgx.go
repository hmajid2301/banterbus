package db

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	*Queries
	pool *pgxpool.Pool
}

func NewDB(pool *pgxpool.Pool) (DB, error) {
	queries := New(pool)
	store := DB{
		pool:    pool,
		Queries: queries,
	}

	return store, nil
}
