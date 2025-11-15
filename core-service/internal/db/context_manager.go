package db

import (
	"context"
	"core-service/internal/logger"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ContextKey string

const (
	EngineKey ContextKey = "db.engine"
)

type ContextManager struct {
	pool *pgxpool.Pool
}

func NewContextManager(pool *pgxpool.Pool) *ContextManager {
	return &ContextManager{
		pool: pool,
	}
}

type Engine interface {
	pgxscan.Querier
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

type Transactioner interface {
	Do(ctx context.Context, f func(ctx context.Context) error) error
}

type EngineFactory interface {
	Get(ctx context.Context) Engine
}

func (cm *ContextManager) putEngineInContext(ctx context.Context, engine Engine) context.Context {
	return context.WithValue(ctx, EngineKey, engine)
}

func (cm *ContextManager) begin(ctx context.Context) (context.Context, error) {
	_, ok := ctx.Value(EngineKey).(pgx.Tx)
	if ok {
		return ctx, nil
	}

	tx, err := cm.pool.Begin(ctx)
	if err != nil {
		return ctx, err
	}

	return cm.putEngineInContext(ctx, tx), nil
}

func (cm *ContextManager) commit(ctx context.Context) error {
	tx, ok := ctx.Value(EngineKey).(pgx.Tx)
	if !ok {
		return nil
	}

	return tx.Commit(ctx)
}

func (cm *ContextManager) rollback(ctx context.Context) error {
	tx, ok := ctx.Value(EngineKey).(pgx.Tx)
	if !ok {
		return nil
	}

	return tx.Rollback(ctx)
}

func (cm *ContextManager) Do(ctx context.Context, f func(ctx context.Context) error) (err error) {
	txCtx, err := cm.begin(ctx)
	if err != nil {
		return err
	}

	detCtx := context.WithoutCancel(txCtx)
	defer func() {
		if p := recover(); p != nil {
			logger.Errorf(detCtx, "panic occurred: %v", p)
			cm.rollback(detCtx)
			panic(p)
		}
		if err != nil {
			logger.Errorf(detCtx, "error in tx occurred: %v", err)
			innerErr := cm.rollback(txCtx)
			if innerErr != nil {
				logger.Errorf(detCtx, "failed to rollback transaction: %v", err)
			}
		} else {
			err = cm.commit(txCtx)
			if err != nil {
				logger.Errorf(detCtx, "failed to commit transaction: %v", err)
			}
		}
	}()

	err = f(txCtx)

	return err
}

func (cm *ContextManager) Get(ctx context.Context) Engine {
	if engine, ok := ctx.Value(EngineKey).(Engine); ok {
		return engine
	}
	return cm.pool
}
