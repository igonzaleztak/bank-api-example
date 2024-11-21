package service

import (
	errors "bank_test/internal/api_errors"
	"bank_test/internal/db"
	"bank_test/internal/db/memory"
	"bank_test/internal/db/models"
	"bank_test/internal/helpers"
	"bank_test/internal/transport/http/schemas"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

// transactionSuite defines the test suite for the transaction service.
type transactionSuite struct {
	db db.DatabaseAdapter
	as AccountService
	ts TransactionService
	suite.Suite
}

func (s *transactionSuite) SetupTest() {
	logger := zap.NewExample().Sugar()

	s.db = memory.NewInMemoryDatabase(logger)
	s.as = NewAccountService(logger, s.db)
	s.ts = NewTransactionService(logger, s.db)
}

// TestCreateTransaction tests the creation of transactions.
func (s *transactionSuite) TestCreateTransaction() {
	account := []schemas.CreateAccountRequest{
		{
			Owner:          "Alice",
			InitialBalance: helpers.PointerValue(float64(100)),
		},
		{
			Owner:          "Bob",
			InitialBalance: helpers.PointerValue(float64(0)),
		},
		{
			Owner:          "Charlie",
			InitialBalance: helpers.PointerValue(float64(50)),
		},
	}

	storedAccounts := make([]*models.Account, 3)

	for i, acc := range account {
		storedAccounts[i] = s.as.CreateAccount(&acc)
	}

	s.Run("ok: withdrawal and deposit", func() {
		// Define input data
		inputs := []struct {
			accountId   string
			transaction *schemas.CreateTransactionRequest
			expectedBal float64
		}{
			{
				accountId: storedAccounts[0].ID,
				transaction: &schemas.CreateTransactionRequest{
					Type:   "withdrawal",
					Amount: helpers.PointerValue(float64(10)),
				},
				expectedBal: float64(90), // Initial balance 100 - 10
			},
			{
				accountId: storedAccounts[0].ID,
				transaction: &schemas.CreateTransactionRequest{
					Type:   "deposit",
					Amount: helpers.PointerValue(float64(20)),
				},
				expectedBal: float64(110), // Previous balance 90 + 20
			},
		}

		// Execute transactions and verify results
		for _, input := range inputs {
			tx, err := s.ts.CreateTransaction(input.accountId, input.transaction)
			s.NoError(err)
			s.NotNil(tx)

			// Verify transaction details
			s.Equal(input.transaction.Type, tx.Type.String())
			s.Equal(*input.transaction.Amount, tx.Amount)
			s.Equal(input.accountId, tx.AccountID)

			// Verify updated account balance
			account, err := s.as.GetAccountByID(input.accountId)
			s.NoError(err)
			s.Equal(input.expectedBal, account.Balance)
		}
	})

	s.Run("not ok: account not found", func() {
		tx, err := s.ts.CreateTransaction(uuid.NewString(), &schemas.CreateTransactionRequest{
			Type:   "deposit",
			Amount: helpers.PointerValue(float64(10)),
		})
		s.Error(err)
		apiError, ok := err.(*errors.APIError)
		s.Require().True(ok)
		s.Equal(apiError, errors.ErrAccountNotFound)
		s.Nil(tx)
	})

	s.Run("not ok: insufficient balance", func() {
		tx, err := s.ts.CreateTransaction(storedAccounts[1].ID, &schemas.CreateTransactionRequest{
			Type:   "withdrawal",
			Amount: helpers.PointerValue(float64(10)),
		})
		s.Error(err)
		apiError, ok := err.(*errors.APIError)
		s.Require().True(ok)

		s.Equal(apiError, errors.ErrInsufficientBalance)
		s.Nil(tx)
	})
}

// TestConcurrentTransactions tests the concurrent execution of transactions.
func (s *transactionSuite) TestConcurrentTransactions() {
	// Create an account with an initial balance
	initialBalance := float64(100)
	accountReq := &schemas.CreateAccountRequest{
		Owner:          "TestUser",
		InitialBalance: &initialBalance,
	}
	account := s.as.CreateAccount(accountReq)
	s.Require().NotNil(account)

	// Define transactions to process concurrently
	transactions := []schemas.CreateTransactionRequest{
		{Type: "deposit", Amount: helpers.PointerValue(float64(50))},
		{Type: "withdrawal", Amount: helpers.PointerValue(float64(20))},
		{Type: "deposit", Amount: helpers.PointerValue(float64(30))},
		{Type: "withdrawal", Amount: helpers.PointerValue(float64(60))},
	}

	var wg sync.WaitGroup
	results := make(chan error, len(transactions))

	// Execute transactions concurrently
	for _, tx := range transactions {
		wg.Add(1)
		go func(transaction schemas.CreateTransactionRequest) {
			defer wg.Done()
			// Create transaction and update balance in sequence
			_, err := s.ts.CreateTransaction(account.ID, &transaction)
			results <- err
		}(tx)
	}

	// Wait for all transactions to complete
	wg.Wait()
	close(results)

	// Collect errors
	var failedTransactions int
	for err := range results {
		if err != nil {
			failedTransactions++
			s.T().Logf("Error: %v", err)
		}
	}

	// Verify the final balance
	finalAccount, err := s.as.GetAccountByID(account.ID)
	s.Require().NoError(err)

	expectedBalance := float64(100) // Adjust based on successful transactions
	s.Equal(expectedBalance, finalAccount.Balance)
	s.T().Logf("Failed transactions: %d", failedTransactions)
}

// TestGetTransactionsByAccountID tests the retrieval of transactions by account ID.
func (s *transactionSuite) TestGetTransactionsByAccountID() {
	account := []schemas.CreateAccountRequest{
		{
			Owner:          "Alice",
			InitialBalance: helpers.PointerValue(float64(100)),
		},
		{
			Owner:          "Bob",
			InitialBalance: helpers.PointerValue(float64(0)),
		},
		{
			Owner:          "Charlie",
			InitialBalance: helpers.PointerValue(float64(50)),
		},
	}

	storedAccounts := make([]*models.Account, 3)

	for i, acc := range account {
		storedAccounts[i] = s.as.CreateAccount(&acc)
	}

	// Define input data
	inputs := []struct {
		accountId    string
		transactions []schemas.CreateTransactionRequest
	}{
		{
			accountId: storedAccounts[0].ID,
			transactions: []schemas.CreateTransactionRequest{
				{Type: "deposit", Amount: helpers.PointerValue(float64(10))},
				{Type: "withdrawal", Amount: helpers.PointerValue(float64(20))},
			},
		},
		{
			accountId: storedAccounts[1].ID,
			transactions: []schemas.CreateTransactionRequest{
				{Type: "deposit", Amount: helpers.PointerValue(float64(10))},
				{Type: "withdrawal", Amount: helpers.PointerValue(float64(5))},
			},
		},
	}

	// Execute transactions
	for _, input := range inputs {
		for _, tx := range input.transactions {
			_, err := s.ts.CreateTransaction(input.accountId, &tx)
			s.Require().NoError(err)
		}
	}

	// Get transactions for each account
	for _, input := range inputs {
		txs, err := s.ts.GetTransactionsByAccountID(input.accountId)
		s.Require().NoError(err)
		s.Len(txs, len(input.transactions))

		// Verify transaction details
		for i, tx := range txs {
			s.Equal(input.transactions[i].Type, tx.Type.String())
			s.Equal(*input.transactions[i].Amount, tx.Amount)
			s.Equal(input.accountId, tx.AccountID)
		}
	}
}

// TestTransfer tests the transfer of funds between accounts.
func (s *transactionSuite) TestTransfer() {
	s.Run("ok: successful transfer", func() {
		accounts := []schemas.CreateAccountRequest{
			{Owner: "Alice", InitialBalance: helpers.PointerValue(float64(100))},
			{Owner: "Bob", InitialBalance: helpers.PointerValue(float64(50))},
		}

		storedAccounts := make(map[string]*models.Account)

		for _, acc := range accounts {
			account := s.as.CreateAccount(&acc)
			s.Require().NotNil(account)
			storedAccounts[account.Owner] = account
		}

		from := storedAccounts["Alice"]
		to := storedAccounts["Bob"]
		amount := float64(30)

		err := s.ts.Transfer(from.ID, to.ID, amount)
		s.NoError(err)

		// Validate balances
		fromAccount, err := s.as.GetAccountByID(from.ID)
		s.Require().NoError(err)
		s.Equal(float64(70), fromAccount.Balance) // 100 - 30 = 70

		toAccount, err := s.as.GetAccountByID(to.ID)
		s.Require().NoError(err)
		s.Equal(float64(80), toAccount.Balance) // 50 + 30 = 80
	})

	s.Run("not ok: insufficient balance", func() {
		accounts := []schemas.CreateAccountRequest{
			{Owner: "Alice", InitialBalance: helpers.PointerValue(float64(100))},
			{Owner: "Bob", InitialBalance: helpers.PointerValue(float64(50))},
		}

		storedAccounts := make(map[string]*models.Account)

		for _, acc := range accounts {
			account := s.as.CreateAccount(&acc)
			s.Require().NotNil(account)
			storedAccounts[account.Owner] = account
		}

		from := storedAccounts["Alice"]
		to := storedAccounts["Bob"]
		amount := float64(200)

		err := s.ts.Transfer(from.ID, to.ID, amount)
		s.Error(err)
		apiError, ok := err.(*errors.APIError)
		s.Require().True(ok)
		s.Equal(apiError, errors.ErrInsufficientBalance)

		// Validate balances
		fromAccount, err := s.as.GetAccountByID(from.ID)
		s.Require().NoError(err)
		s.Equal(float64(100), fromAccount.Balance) // No change

		toAccount, err := s.as.GetAccountByID(to.ID)
		s.Require().NoError(err)
		s.Equal(float64(50), toAccount.Balance) // No change
	})

	s.Run("not ok: account not found", func() {
		err := s.ts.Transfer(uuid.NewString(), uuid.NewString(), float64(10))
		s.Error(err)
		apiError, ok := err.(*errors.APIError)
		s.Require().True(ok)
		s.Equal(apiError, errors.ErrAccountNotFound)
	})
}

// TestConcurrentTransfers tests the concurrent execution of transfers.
func (s *transactionSuite) TestConcurrentTransfers() {
	// Setup: create accounts
	alice := s.as.CreateAccount(&schemas.CreateAccountRequest{
		Owner:          "Alice",
		InitialBalance: helpers.PointerValue(float64(1000)),
	})
	bob := s.as.CreateAccount(&schemas.CreateAccountRequest{
		Owner:          "Bob",
		InitialBalance: helpers.PointerValue(float64(100)),
	})
	s.Require().NotNil(alice)
	s.Require().NotNil(bob)

	// Define the transfer parameters
	transferAmount := float64(10)
	concurrentTransfers := 50

	var wg sync.WaitGroup
	errs := make(chan error, concurrentTransfers)

	// Simulate concurrent transfers
	for i := 0; i < concurrentTransfers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := s.ts.Transfer(alice.ID, bob.ID, transferAmount)
			errs <- err
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errs)

	// Check for errors
	for err := range errs {
		s.NoError(err)
	}

	// Verify final balances
	aliceAccount, err := s.as.GetAccountByID(alice.ID)
	s.Require().NoError(err)
	bobAccount, err := s.as.GetAccountByID(bob.ID)
	s.Require().NoError(err)

	expectedAliceBalance := 1000 - float64(concurrentTransfers)*transferAmount
	expectedBobBalance := 100 + float64(concurrentTransfers)*transferAmount

	s.Equal(expectedAliceBalance, aliceAccount.Balance)
	s.Equal(expectedBobBalance, bobAccount.Balance)
}

func TestTransactionSuite(t *testing.T) {
	suite.Run(t, new(transactionSuite))
}
