package schemas

// CreateAccountRequest is the request schema for the CreateAccount endpoint.
// It is used to create a new account.
type CreateAccountRequest struct {
	Owner          string   `json:"owner" validate:"required"`
	InitialBalance *float64 `json:"initial_balance" validate:"required"`
}

// CreateTransactionRequest is the request schema for the CreateTransaction endpoint.
// It is used to create a new transaction for an account.
type CreateTransactionRequest struct {
	Type   string   `json:"type" validate:"required,oneof=deposit withdrawal"`
	Amount *float64 `json:"amount" validate:"required,gt=0"`
}

// TransferRequest is the request schema for the Transfer endpoint.
// It is used to transfer money from one account to another.
type TransferRequest struct {
	FromAccountId string   `json:"from_account_id" validate:"required"`
	ToAccountId   string   `json:"to_account_id" validate:"required"`
	Amount        *float64 `json:"amount" validate:"required,gt=0"`
}
