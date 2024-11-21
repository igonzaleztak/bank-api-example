package errors

import "net/http"

// APIError represents an error that is returned to the client.
type APIError struct {
	Code       string `json:"code"`    // error code that can be used to identify the error
	Message    string `json:"message"` // detailed description of the error
	HTTPStatus int    `json:"-"`       // http status code. It is not included in the response body
}

// NewAPIError creates a new APIError.
func NewAPIError(code string, message string, httpStatus int) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// Error returns the error message.
func (e *APIError) Error() string {
	return e.Message
}

var (

	// ErrInvalidBody is returned when the request body is invalid.
	ErrInvalidBody = NewAPIError("INVALID_BODY", "invalid request body", http.StatusBadRequest)

	// ErrAccountIdIsMissing is returned when the account id is missing.
	ErrAccountIdIsMissing = NewAPIError("ACCOUNT_ID_MISSING", "account id is missing", http.StatusBadRequest)

	// ErrAccountNotFound is returned when an account is not found.
	ErrAccountNotFound = NewAPIError("ACCOUNT_NOT_FOUND", "account not found", http.StatusBadRequest)

	// ErrInvalidAccountID is returned when an account id is invalid.
	ErrInvalidAccountID = NewAPIError("INVALID_ACCOUNT_ID", "invalid account id. Must be UUID format", http.StatusBadRequest)

	// ErrInsufficientBalance is returned when an account has insufficient balance.
	ErrInsufficientBalance = NewAPIError("INSUFFICIENT_BALANCE", "insufficient balance", http.StatusBadRequest)

	// ErrInvalidAmount is returned when an amount is invalid.
	INVALID_AMOUNT = NewAPIError("INVALID_AMOUNT", "invalid amount", http.StatusBadRequest)

	// ErrUkwnown is returned when an unknown error occurs.
	ErrUnknown = NewAPIError("UNKNOWN", "unknown error", http.StatusInternalServerError)
)
