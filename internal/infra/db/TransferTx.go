package db

import (
	"context"
	"fmt"
)

type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

func (store *StoreSQL) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		var fromAccount Account

		if arg.FromAccountID < arg.ToAccountID {
			fromAccount, _, err = blockAccounts(ctx, q, arg.FromAccountID, arg.ToAccountID)
			if err != nil {
				return err
			}
		} else {
			_, fromAccount, err = blockAccounts(ctx, q, arg.ToAccountID, arg.FromAccountID)
			if err != nil {
				return err
			}
		}

		if fromAccount.Balance < arg.Amount {
			return fmt.Errorf("the balance of the from account is insufficient")
		}

		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams(arg))
		if err != nil {
			return err
		}

		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		if arg.FromAccountID < arg.ToAccountID {
			result.FromAccount, result.ToAccount, err = updateBalanceForAccounts(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, +arg.Amount)
			if err != nil {
				return err
			}
		} else {
			result.ToAccount, result.FromAccount, err = updateBalanceForAccounts(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return result, err
}

func blockAccounts(
	ctx context.Context,
	q *Queries,
	accountID1 int64,
	accountID2 int64,
) (account1 Account, account2 Account, err error) {
	account1, err = q.GetAccountForUpdate(ctx, accountID1)
	if err != nil {
		return
	}

	account2, err = q.GetAccountForUpdate(ctx, accountID2)
	if err != nil {
		return
	}

	return
}

func updateBalanceForAccounts(
	ctx context.Context,
	q *Queries,
	accountID1 int64,
	amount1 int64,
	accountID2 int64,
	amount2 int64,
) (account1 Account, account2 Account, err error) {
	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID1,
		Amount: amount1,
	})
	if err != nil {
		return
	}

	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID2,
		Amount: amount2,
	})
	if err != nil {
		return
	}

	return
}
