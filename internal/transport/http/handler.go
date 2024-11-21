package http

import (
	errors "bank_test/internal/api_errors"
	"bank_test/internal/db"
	"bank_test/internal/helpers"
	"bank_test/internal/service"
	"bank_test/internal/transport/http/binding"
	"bank_test/internal/transport/http/schemas"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type handler struct {
	logger *zap.SugaredLogger
	db     db.DatabaseAdapter

	// services
	as service.AccountService
	ts service.TransactionService
}

// newHandler creates a new handler.
func newHandler(logger *zap.SugaredLogger, db db.DatabaseAdapter) *handler {
	// initiate services
	as := service.NewAccountService(logger, db)
	ts := service.NewTransactionService(logger, db)

	return &handler{logger: logger, db: db, as: as, ts: ts}
}

// createAccount is an endpoint that creates a new account.
func (h *handler) createAccount(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("create account endpoint called")

	h.logger.Debugf("decoding request body")
	var body schemas.CreateAccountRequest
	if err := binding.DecodeJSONBody(r, &body); err != nil {
		e := errors.ErrInvalidBody
		e.Message = fmt.Sprintf("failed to decode request body: %v", err)
		h.wrapError(w, r, e)
		return
	}
	h.logger.Debugf("request body decoded successfully: %s", helpers.PrettyPrintStructResponse(body))

	h.logger.Debugf("creating account for owner %s", body.Owner)
	acc := h.as.CreateAccount(&body)
	h.logger.Info("account created successfully")
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, acc)
}

// getAccount is an endpoint that retrieves an account by its id.
func (h *handler) getAccount(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("get account endpoint called")

	// decode the account id from the request
	h.logger.Debugf("decoding account id from the request")
	accID := chi.URLParam(r, "id")
	if accID == "" {
		h.wrapError(w, r, errors.ErrAccountIdIsMissing)
		return
	}

	// check if the account id is valid uuid
	if err := uuid.Validate(accID); err != nil {
		h.wrapError(w, r, errors.ErrInvalidAccountID)
		return
	}

	h.logger.Debugf("account id decoded successfully: %s", accID)

	h.logger.Debugf("getting account with id %s", accID)
	acc, err := h.as.GetAccountByID(accID)
	if err != nil {
		h.wrapError(w, r, err)
		return
	}
	h.logger.Info("account retrieved successfully")
	render.Status(r, http.StatusOK)
	render.JSON(w, r, acc)
}

// getAllAccounts is an endpoint that retrieves all accounts.
func (h *handler) getAllAccounts(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("get all accounts endpoint called")

	h.logger.Info("getting all accounts")
	accs := h.as.GetAllAccounts()
	h.logger.Info("all accounts retrieved successfully")
	render.Status(r, http.StatusOK)
	render.JSON(w, r, accs)
}

// createTransaction creates a new transaction by either depositing money or withdrawing it.
func (h *handler) createTransaction(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("create transaction endpoint called")

	// decode the account id from the request
	h.logger.Debugf("decoding account id from the request")
	accID := chi.URLParam(r, "id")
	if accID == "" {
		h.wrapError(w, r, errors.ErrAccountIdIsMissing)
		return
	}

	// check if the account id is valid uuid
	if err := uuid.Validate(accID); err != nil {
		h.wrapError(w, r, errors.ErrInvalidAccountID)
		return
	}
	h.logger.Debugf("account id decoded successfully: %s", accID)

	h.logger.Debugf("decoding request body")
	var body schemas.CreateTransactionRequest
	if err := binding.DecodeJSONBody(r, &body); err != nil {
		bodyErr := errors.ErrInvalidBody
		bodyErr.Message = fmt.Sprintf("failed to decode request body: %v", err)
		h.wrapError(w, r, bodyErr)
		return
	}
	h.logger.Debugf("request body decoded successfully: %s", helpers.PrettyPrintStructResponse(body))

	h.logger.Debugf("creating transaction for account %s", accID)
	acc, err := h.ts.CreateTransaction(accID, &body)
	if err != nil {
		h.wrapError(w, r, err)
		return
	}
	h.logger.Info("transaction created successfully")
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, acc)
}

// getTransactionsByAccountID retrieves all transactions for the account.
func (h *handler) getTransactionsByAccountID(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("get transactions by account id endpoint called")

	// decode the account id from the request
	h.logger.Debugf("decoding account id from the request")
	accID := chi.URLParam(r, "id")
	if accID == "" {
		h.wrapError(w, r, errors.ErrAccountIdIsMissing)
		return
	}

	// check if the account id is valid uuid
	if err := uuid.Validate(accID); err != nil {
		h.wrapError(w, r, errors.ErrInvalidAccountID)
		return
	}

	h.logger.Debugf("account id decoded successfully: %s", accID)

	h.logger.Debugf("getting all transactions for account with id %s", accID)
	txs, err := h.ts.GetTransactionsByAccountID(accID)
	if err != nil {
		h.wrapError(w, r, err)
		return
	}
	h.logger.Info("all transactions retrieved successfully")
	render.Status(r, http.StatusOK)
	render.JSON(w, r, txs)
}

// transfer is an endpoint that transfers money from one account to another.
func (h *handler) transfer(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("transfer endpoint called")

	// decode the request body
	h.logger.Debugf("decoding request body")
	var body schemas.TransferRequest
	if err := binding.DecodeJSONBody(r, &body); err != nil {
		bodyErr := errors.ErrInvalidBody
		bodyErr.Message = fmt.Sprintf("failed to decode request body: %v", err)
		h.wrapError(w, r, bodyErr)
		return
	}
	h.logger.Debugf("request body decoded successfully: %s", helpers.PrettyPrintStructResponse(body))

	h.logger.Debugf("transferring money from account %s to account %s", body.FromAccountId, body.ToAccountId)
	if err := h.ts.Transfer(body.FromAccountId, body.ToAccountId, *body.Amount); err != nil {
		h.wrapError(w, r, err)
		return
	}
	h.logger.Info("money transferred successfully")
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, schemas.OkResponse{Message: "money transferred successfully"})
}

// wrapError logs the error and writes it to the response.
func (h *handler) wrapError(w http.ResponseWriter, r *http.Request, err error) {
	apiError, ok := err.(*errors.APIError)
	if !ok {
		unknownError := errors.ErrUnknown
		unknownError.Message = err.Error()
		h.logger.Error(unknownError)
		render.JSON(w, r, unknownError)
		return
	}

	h.logger.Error(apiError.Message)
	render.Status(r, apiError.HTTPStatus)
	render.JSON(w, r, apiError)
}
