package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Store interface {
	TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error)
	Querier
}

type StoreSQL struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) Store {
	return &StoreSQL{
		db:      db,
		Queries: New(db),
	}
}

func (store *StoreSQL) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := store.WithTx(tx)
	err = fn(q)

	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("error: %v, rollback error: %v", err, rbErr)
		}

		return err
	}

	return tx.Commit()
}
