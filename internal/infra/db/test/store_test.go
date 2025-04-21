package test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/NhutHuyDev/sgbank/internal/infra/db"
	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := db.NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	n := 5
	amount := int64(10)

	errs := make(chan error)
	results := make(chan db.TransferTxResult)

	for i := 0; i < n; i++ {
		go func() {
			result, err := store.TransferTx(context.Background(), db.TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})

			errs <- err
			results <- result
		}()
	}

	existed := make(map[int]bool)
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		// Check transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, transfer.FromAccountID, account1.ID)
		require.Equal(t, transfer.ToAccountID, account2.ID)
		require.Equal(t, transfer.Amount, amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// Check entries
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, fromEntry.AccountID, account1.ID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, toEntry.AccountID, account2.ID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// Check accounts
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, account1.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, account2.ID, toAccount.ID)

		// check accounts's balance
		diff1 := account1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - account2.Balance

		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0)

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	// check the final updated balance of accounts
	updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	require.Equal(t, account1.Balance-amount*int64(n), updatedAccount1.Balance)
	require.Equal(t, account2.Balance+amount*int64(n), updatedAccount2.Balance)
}

// Testing Deadlock by order matters
func TestTransferTxDeadlock(t *testing.T) {
	store := db.NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	n := 20
	amount := int64(10)

	errs := make(chan error)

	for i := 0; i < n; i++ {
		fromAccountID := account1.ID
		toAccountID := account2.ID

		if i%2 == 1 {
			fromAccountID = account2.ID
			toAccountID = account1.ID
		}

		go func() {
			_, err := store.TransferTx(context.Background(), db.TransferTxParams{
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				Amount:        amount,
			})

			errs <- err
		}()
	}

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}

	// check the final updated balance of accounts
	updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	require.Equal(t, account1.Balance, updatedAccount1.Balance)
	require.Equal(t, account2.Balance, updatedAccount2.Balance)
}

func TestTransferTxCheckBalance(t *testing.T) {
	store := db.NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	fmt.Printf("account1's balance: %v \n", account1.Balance)
	fmt.Printf("account2's balance: %v \n", account2.Balance)

	n := 20
	amount := int64(100)
	var wg sync.WaitGroup

	// Use buffered channels to prevent blocking
	errs := make(chan error, n)
	results := make(chan db.TransferTxResult, n)

	for i := 0; i < n; i++ {
		wg.Add(1) // Increment the counter

		go func() {
			defer wg.Done() // Decrement the counter when the goroutine completes
			result, err := store.TransferTx(context.Background(), db.TransferTxParams{
				FromAccountID: account2.ID,
				ToAccountID:   account1.ID,
				Amount:        amount,
			})

			errs <- err
			results <- result
		}()
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Close channels after all goroutines are done
	close(errs)
	close(results)

	// Collect and check the errors and results
	var successCount, failCount int
	for err := range errs {
		if err != nil {
			failCount++
		} else {
			successCount++
		}
	}

	fmt.Printf("total number of unsuccessful transfer transactions: %v \n", failCount)
	fmt.Printf("total number of successful transfer transactions: %v \n", successCount)

	require.Greater(t, failCount, 0)
	require.Less(t, successCount, n)

	// check the final updated balance of accounts
	updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	require.Equal(t, account1.Balance+amount*int64(successCount), updatedAccount1.Balance)
	require.Equal(t, account2.Balance-amount*int64(successCount), updatedAccount2.Balance)
}
