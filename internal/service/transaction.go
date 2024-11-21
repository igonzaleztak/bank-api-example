package service

import (
	"bank_test/internal/db"
	"bank_test/internal/db/models"
	"bank_test/internal/enum"
	"bank_test/internal/transport/http/schemas"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// transaction handles all the transaction related operations.
type transaction struct {
	logger *zap.SugaredLogger
	db     db.DatabaseAdapter
}

func NewTransactionService(logger *zap.SugaredLogger, db db.DatabaseAdapter) TransactionService {
	return &transaction{logger: logger, db: db}
}

// CreateTransaction creates a new transaction for the account.
func (s *transaction) CreateTransaction(id string, transaction *schemas.CreateTransactionRequest) (*models.Transaction, error) {
	s.logger.Debugf("creating transaction for account with id %s", id)

	// convert transaction type to enum. It is not necessary to validate the transaction type since it has already
	// been validated in the handler when decoding the request body by using the tag 'oneof'.
	// Additionally, it is not necessary to validate the amount since it has been validated in the handler too.
	txType := enum.TransactionTypeFromString(transaction.Type)

	// create a new transaction and update the account
	txId := uuid.New().String()
	s.logger.Debugf("creating transaction with id %s for account %s", txId, id)
	tx := models.Transaction{
		ID:        txId,
		AccountID: id,
		Type:      txType,
		Amount:    *transaction.Amount,
		Timestamp: time.Now(),
	}
	if err := s.db.CreateTransaction(&tx); err != nil {
		return nil, s.wrapError(err)
	}
	s.logger.Debugf("transaction with id %s created successfully", tx.ID)

	return &tx, nil
}

// GetTransactionsByAccountID retrieves all transactions for the account.
func (s *transaction) GetTransactionsByAccountID(accountID string) ([]models.Transaction, error) {
	s.logger.Debugf("getting all transactions for account with id %s", accountID)
	txs, err := s.db.GetTransactionsByAccountID(accountID)
	if err != nil {
		return nil, s.wrapError(err)
	}
	s.logger.Debugf("all transactions for account with id %s retrieved successfully", accountID)
	return txs, nil
}

// Transfer transfer money from one account to another.
func (s *transaction) Transfer(from string, to string, amount float64) error {
	s.logger.Debugf("transferring %.5f from account %s to account %s", amount, from, to)

	// create a new transaction for the withdrawal
	withdrawalFrom := &models.Transaction{
		ID:        uuid.New().String(),
		AccountID: from,
		Type:      enum.Withdrawal,
		Amount:    amount,
		Timestamp: time.Now(),
	}
	if err := s.db.CreateTransaction(withdrawalFrom); err != nil {
		return s.wrapError(err)
	}

	// create a new transaction for the deposit
	depositTo := &models.Transaction{
		ID:        uuid.New().String(),
		AccountID: to,
		Type:      enum.Deposit,
		Amount:    amount,
		Timestamp: time.Now(),
	}
	if err := s.db.CreateTransaction(depositTo); err != nil {
		return s.wrapError(err)
	}
	s.logger.Debugf("transfer completed successfully")
	return nil
}

// wrapError logs the error and returns it.
func (s *transaction) wrapError(err error) error {
	s.logger.Error(err)
	return err
}
