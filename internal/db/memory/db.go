package memory

import (
	errors "bank_test/internal/api_errors"
	"bank_test/internal/db/models"
	"bank_test/internal/enum"
	"bank_test/internal/helpers"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// inMemoryDatabase is an in-memory implementation of the database.
//
// We could use another packages such as go-memdb, but for the sake of simplicity I have
// decided to implement a simple in-memory database.
type inMemoryDatabase struct {
	mu     sync.RWMutex
	logger *zap.SugaredLogger

	accounts     map[string]models.Account
	transactions map[string][]models.Transaction
}

// NewInMemoryDatabase creates a new in-memory database.
func NewInMemoryDatabase(logger *zap.SugaredLogger) *inMemoryDatabase {
	return &inMemoryDatabase{
		logger: logger,

		accounts:     make(map[string]models.Account),
		transactions: make(map[string][]models.Transaction),
	}
}

// CreateAccount creates a new account in the database.
func (d *inMemoryDatabase) CreateAccount(account *models.Account) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.logger.Debugf("storing account with id '%s' in memory database: %s", account.ID, helpers.PrettyPrintStructResponse(account))
	d.accounts[account.ID] = *account
	d.transactions[account.ID] = make([]models.Transaction, 0)
	d.logger.Debugf("account with id '%s' stored in memory database", account.ID)
}

// GetAccountByID retrieves an account from the database by its id.
func (d *inMemoryDatabase) GetAccountByID(id string) (*models.Account, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	d.logger.Debugf("getting account with id '%s' from memory database", id)
	acc, ok := d.accounts[id]
	if !ok {
		d.logger.Error(fmt.Sprintf("account with id '%s' not found", id))
		return nil, errors.ErrAccountNotFound
	}
	d.logger.Debugf("account with id '%s' retrieved from memory database: %s", id, helpers.PrettyPrintStructResponse(acc))
	return &acc, nil
}

// GetAllAccounts retrieves all accounts from the database.
func (d *inMemoryDatabase) GetAllAccounts() []models.Account {
	d.mu.RLock()
	defer d.mu.RUnlock()

	d.logger.Debugf("getting all accounts from memory database")
	accounts := make([]models.Account, 0, len(d.accounts))
	for _, acc := range d.accounts {
		accounts = append(accounts, acc)
	}
	d.logger.Debugf("all accounts retrieved from memory database: %s", helpers.PrettyPrintStructResponse(accounts))
	return accounts
}

// CreateTransaction creates a new transaction in the database.
func (d *inMemoryDatabase) CreateTransaction(transaction *models.Transaction) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.logger.Debugf("getting account with id '%s' from memory database", transaction.AccountID)
	account, ok := d.accounts[transaction.AccountID]
	if !ok {
		d.logger.Error(fmt.Sprintf("account with id '%s' not found", transaction.AccountID))
		return errors.ErrAccountNotFound
	}
	d.logger.Debugf("account with id '%s' retrieved from memory database: %s", transaction.AccountID, helpers.PrettyPrintStructResponse(account))

	d.logger.Debugf("updating account balance")
	switch transaction.Type {
	case enum.Deposit:
		account.Balance += transaction.Amount
	case enum.Withdrawal:
		if account.Balance < transaction.Amount {
			d.logger.Error(fmt.Sprintf("insufficient balance for account with id '%s'", transaction.AccountID))
			return errors.ErrInsufficientBalance
		}
		account.Balance -= transaction.Amount
	default:
		d.logger.Error(fmt.Sprintf("invalid transaction type '%s'", transaction.Type))
		return errors.ErrUnknown
	}

	d.accounts[transaction.AccountID] = account
	d.logger.Debugf("account balance updated: %f", account.Balance)

	d.logger.Debugf("storing transaction with id '%s' in memory database: %s", transaction.ID, helpers.PrettyPrintStructResponse(transaction))
	d.transactions[transaction.AccountID] = append(d.transactions[transaction.AccountID], *transaction)
	d.logger.Debugf("transaction with id '%s' stored in memory database", transaction.ID)
	return nil
}

// GetTransactionsByAccountID retrieves all transactions for an account from the database.
func (d *inMemoryDatabase) GetTransactionsByAccountID(id string) ([]models.Transaction, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	d.logger.Debugf("getting all transactions for account with id '%s' from memory database", id)
	txs, ok := d.transactions[id]
	if !ok {
		return nil, errors.ErrAccountNotFound
	}
	d.logger.Debugf("all transactions for account with id '%s' retrieved from memory database: %s", id, helpers.PrettyPrintStructResponse(txs))
	return txs, nil
}
