package db

import (
	"bank_test/internal/db/memory"
	"bank_test/internal/db/models"

	"go.uber.org/zap"
)

// Database is the interface for the database layer.
//
// By using this interface, we can easily swap out the underlying database implementation.
type DatabaseAdapter interface {
	// Account methods
	CreateAccount(account *models.Account)             // CreateAccount creates a new account
	GetAccountByID(id string) (*models.Account, error) // GetAccountByID retrieves an account by its ID
	GetAllAccounts() []models.Account                  // GetAllAccounts retrieves all accounts

	// Transaction methods
	CreateTransaction(transaction *models.Transaction) error            // CreateTransaction creates a new transaction
	GetTransactionsByAccountID(id string) ([]models.Transaction, error) // GetTransactionsByAccountID retrieves all transactions for an account
}

// NewDatabaseAdapter creates a new database adapter. In this case there is only one implementation: an in-memory database.
// However, in the future we could add other implementations such as a SQL database. By modifying this function, we can easily
// switch between different database implementations.
func NewDatabaseAdapter(logger *zap.SugaredLogger) DatabaseAdapter {
	return memory.NewInMemoryDatabase(logger)
}
