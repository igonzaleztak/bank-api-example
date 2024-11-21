package enum

// TransactionType is the type for the transaction type enum

type TransactionType string

const (
	// Deposit is the enum value for the deposit transaction type
	Deposit TransactionType = "deposit"

	// Withdrawal is the enum value for the withdrawal transaction type
	Withdrawal TransactionType = "withdrawal"
)

func (t TransactionType) String() string {
	return string(t)
}

func TransactionTypeFromString(s string) TransactionType {
	switch s {
	case "deposit":
		return Deposit
	case "withdrawal":
		return Withdrawal
	default:
		return ""
	}
}
