package service

import (
	"bank_test/internal/db/models"
	"bank_test/internal/transport/http/schemas"
)

// AccountService is the interface for the account service. It defines the business logic for the account service.
type AccountService interface {
	CreateAccount(account *schemas.CreateAccountRequest) *models.Account // CreateAccount creates a new account
	GetAccountByID(id string) (*models.Account, error)                   // GetAccountByID retrieves an account by its ID
	GetAllAccounts() []models.Account                                    // GetAllAccounts retrieves all accounts
}

// TransactionService is the interface for the transaction service. It defines the business logic for the transaction service.
type TransactionService interface {
	CreateTransaction(accountId string, transaction *schemas.CreateTransactionRequest) (*models.Transaction, error) // CreateTransaction creates a new transaction
	GetTransactionsByAccountID(accountId string) ([]models.Transaction, error)                                      // GetTransactionsByAccountID retrieves all transactions for an account
	Transfer(from string, to string, amount float64) error                                                          // Transfer transfers money from one account to another
}
