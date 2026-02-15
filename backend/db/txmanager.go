package db

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TxManager interface {
	RunInTx(ctx context.Context, fn func(q Querier) error) error
}

type PgxTxManager struct {
	pool    *pgxpool.Pool
	queries *Queries
}

func NewTxManager(pool *pgxpool.Pool, queries *Queries) *PgxTxManager {
	return &PgxTxManager{pool: pool, queries: queries}
}

func (m *PgxTxManager) RunInTx(ctx context.Context, fn func(q Querier) error) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			slog.Error("failed to rollback transaction", "error", err)
		}
	}()

	qtx := m.queries.WithTx(tx)
	if err := fn(qtx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
