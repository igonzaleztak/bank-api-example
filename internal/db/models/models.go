package models

import (
	"bank_test/internal/enum"
	"time"
)

// Account is the model for the account table
type Account struct {
	ID      string  `json:"id"`
	Owner   string  `json:"owner"`
	Balance float64 `json:"balance"`
}

// Transaction is the model for the transaction table
type Transaction struct {
	ID        string               `json:"id"`
	AccountID string               `json:"account_id"`
	Type      enum.TransactionType `json:"type"` // desposit or withdrawal
	Amount    float64              `json:"amount"`
	Timestamp time.Time            `json:"timestamp"` // timestamp in RFC3339 format
}
