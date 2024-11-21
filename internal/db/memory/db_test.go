package memory

import (
	errors "bank_test/internal/api_errors"
	"bank_test/internal/db/models"
	"bank_test/internal/enum"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

// Define the test suite
type InMemoryDatabaseTestSuite struct {
	suite.Suite
	db     *inMemoryDatabase
	logger *zap.SugaredLogger
}

func (suite *InMemoryDatabaseTestSuite) SetupTest() {
	logger := zap.NewExample().Sugar()
	suite.logger = logger
	suite.db = NewInMemoryDatabase(logger)
}

// TestCreateAccountAndGetAccountByID tests account creation and retrieval.
func (suite *InMemoryDatabaseTestSuite) TestCreateAccountAndGetAccountByID() {
	account := &models.Account{
		ID:      "1",
		Owner:   "Alice",
		Balance: 100.0,
	}

	// Create account
	suite.db.CreateAccount(account)

	// Retrieve the account by ID
	retrievedAccount, err := suite.db.GetAccountByID(account.ID)
	suite.Require().NoError(err)
	suite.Equal(account.ID, retrievedAccount.ID)
	suite.Equal(account.Owner, retrievedAccount.Owner)
	suite.Equal(account.Balance, retrievedAccount.Balance)
}

// TestCreateTransactionDeposit tests deposit transaction.
func (suite *InMemoryDatabaseTestSuite) TestCreateTransactionDeposit() {
	account := &models.Account{
		ID:      "2",
		Owner:   "Bob",
		Balance: 50.0,
	}

	// Create account
	suite.db.CreateAccount(account)

	// Create deposit transaction
	transaction := &models.Transaction{
		ID:        "tx1",
		AccountID: account.ID,
		Type:      enum.Deposit,
		Amount:    20.0,
	}

	// Execute transaction
	err := suite.db.CreateTransaction(transaction)
	suite.Require().NoError(err)

	// Verify account balance after deposit
	retrievedAccount, err := suite.db.GetAccountByID(account.ID)
	suite.Require().NoError(err)
	suite.Equal(70.0, retrievedAccount.Balance)
}

// TestCreateTransactionWithdrawal tests withdrawal transaction.
func (suite *InMemoryDatabaseTestSuite) TestCreateTransactionWithdrawal() {
	account := &models.Account{
		ID:      "3",
		Owner:   "Charlie",
		Balance: 100.0,
	}

	// Create account
	suite.db.CreateAccount(account)

	// Create withdrawal transaction
	transaction := &models.Transaction{
		ID:        "tx2",
		AccountID: account.ID,
		Type:      enum.Withdrawal,
		Amount:    30.0,
	}

	// Execute transaction
	err := suite.db.CreateTransaction(transaction)
	suite.Require().NoError(err)

	// Verify account balance after withdrawal
	retrievedAccount, err := suite.db.GetAccountByID(account.ID)
	suite.Require().NoError(err)
	suite.Equal(70.0, retrievedAccount.Balance)
}

// TestCreateTransactionInsufficientBalance tests withdrawal with insufficient balance.
func (suite *InMemoryDatabaseTestSuite) TestCreateTransactionInsufficientBalance() {
	account := &models.Account{
		ID:      "4",
		Owner:   "Dave",
		Balance: 10.0,
	}

	// Create account
	suite.db.CreateAccount(account)

	// Create withdrawal transaction with insufficient balance
	transaction := &models.Transaction{
		ID:        "tx3",
		AccountID: account.ID,
		Type:      enum.Withdrawal,
		Amount:    20.0,
	}

	// Try to execute the transaction and expect an error
	err := suite.db.CreateTransaction(transaction)
	suite.Require().Error(err)
	suite.Equal(errors.ErrInsufficientBalance, err)
}

// TestCreateTransactionInvalidAccount tests transaction on a non-existent account.
func (suite *InMemoryDatabaseTestSuite) TestCreateTransactionInvalidAccount() {
	// Create withdrawal transaction for a non-existent account
	transaction := &models.Transaction{
		ID:        "tx4",
		AccountID: "nonexistent",
		Type:      enum.Withdrawal,
		Amount:    50.0,
	}

	// Try to execute the transaction and expect an error
	err := suite.db.CreateTransaction(transaction)
	suite.Require().Error(err)
	suite.Equal(errors.ErrAccountNotFound, err)
}

// TestGetTransactionsByAccountID tests retrieving transactions by account ID.
func (suite *InMemoryDatabaseTestSuite) TestGetTransactionsByAccountID() {
	account := &models.Account{
		ID:      "5",
		Owner:   "Eve",
		Balance: 200.0,
	}

	// Create account
	suite.db.CreateAccount(account)

	// Create deposit and withdrawal transactions
	depositTransaction := &models.Transaction{
		ID:        "tx5",
		AccountID: account.ID,
		Type:      enum.Deposit,
		Amount:    50.0,
	}
	withdrawalTransaction := &models.Transaction{
		ID:        "tx6",
		AccountID: account.ID,
		Type:      enum.Withdrawal,
		Amount:    30.0,
	}

	// Execute transactions
	err := suite.db.CreateTransaction(depositTransaction)
	suite.Require().NoError(err)
	err = suite.db.CreateTransaction(withdrawalTransaction)
	suite.Require().NoError(err)

	// Retrieve transactions for account
	transactions, err := suite.db.GetTransactionsByAccountID(account.ID)
	suite.Require().NoError(err)
	suite.Len(transactions, 2)
	suite.Equal(depositTransaction.ID, transactions[0].ID)
	suite.Equal(withdrawalTransaction.ID, transactions[1].ID)
}

// TestGetAllAccounts tests retrieving all accounts.
func (suite *InMemoryDatabaseTestSuite) TestGetAllAccounts() {
	account1 := &models.Account{
		ID:      "6",
		Owner:   "Frank",
		Balance: 300.0,
	}
	account2 := &models.Account{
		ID:      "7",
		Owner:   "Grace",
		Balance: 400.0,
	}

	// Create accounts
	suite.db.CreateAccount(account1)
	suite.db.CreateAccount(account2)

	// Get all accounts
	accounts := suite.db.GetAllAccounts()
	suite.Len(accounts, 2)
}

func TestInMemoryDatabaseTestSuite(t *testing.T) {
	suite.Run(t, new(InMemoryDatabaseTestSuite))
}
