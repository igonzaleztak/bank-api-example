package service

import (
	"bank_test/internal/db"
	"bank_test/internal/db/models"
	"bank_test/internal/transport/http/schemas"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// account handles all the account related operations.
type account struct {
	logger *zap.SugaredLogger
	db     db.DatabaseAdapter
}

// NewAccountService creates a new account service that implements all the business logic for accounts.
func NewAccountService(logger *zap.SugaredLogger, db db.DatabaseAdapter) AccountService {
	return &account{logger: logger, db: db}
}

// CreateAccount creates a new account for the owner.
func (a *account) CreateAccount(account *schemas.CreateAccountRequest) *models.Account {
	a.logger.Debugf("creating account for owner %s", account.Owner)

	a.logger.Debugf("generating account id")
	id := uuid.New()
	a.logger.Debugf("account id generated: %s", id.String())

	acc := models.Account{
		ID:      id.String(),
		Owner:   account.Owner,
		Balance: *account.InitialBalance,
	}

	a.logger.Debugf("saving account to database with id %s", acc.ID)
	a.db.CreateAccount(&acc)
	a.logger.Debugf("account with id %s created successfully", acc.ID)
	return &acc
}

// GetAccountByID retrieves an account by its id.
func (a *account) GetAccountByID(id string) (*models.Account, error) {
	a.logger.Debugf("getting account with id %s", id)
	acc, err := a.db.GetAccountByID(id)
	if err != nil {
		return nil, err
	}
	a.logger.Debugf("account with id %s retrieved successfully", id)
	return acc, nil
}

// GetAllAccounts retrieves all accounts stored in the database.
func (a *account) GetAllAccounts() []models.Account {
	a.logger.Debugf("getting all accounts")
	accounts := a.db.GetAllAccounts()
	a.logger.Debugf("all accounts retrieved successfully")
	return accounts
}
