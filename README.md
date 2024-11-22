# Valsea task

## Overview

This repository aims to design an API to solve the following task: Definition of a simple RESTful API for managing bank accounts and
their transactions. The API must allow users to:

1. Create new accounts.
   - Endpoint: `POST /accounts` 
   - Description: Create a new bank account with an initial balance.
   - Request Body: JSON containing owner and initial_balance.
2. Retrieve Account Details.
   - Endpoint: `GET /accounts/{id}` 
   - Description: Retrieve details of a specific account by ID.
3. List All Accounts.
   - Endpoint: `GET /accounts` 
   - Description: Retrieve a list of all bank accounts.
4. Create a Transaction
   - Endpoint: `POST /accounts/{id}/transactions` 
   - Description: Create a deposit or withdrawal transaction for a specific account.
   - Request Body: JSON containing type (deposit or withdrawal) and amount.
5. Retrieve Transactions for an Account.
   - Endpoint: `GET /accounts/{id}/transactions` 
   - Description: Retrieve all transactions associated with a specific account.
6. Transfer Between Accounts
   - Endpoint: `POST /transfer` 
   - Description: Transfer funds from one account to another.
   - Request Body: JSON containing from_account_id, to_account_id, and amount.

## Design

This section describes how the project has been structured and the line of thought that I have followed to accomplish the definition of the API. 

Below, it can be seen the project's structure

```md
.
├── cmd
│   ├── bootstrap
    │   ├──  bootstrap.go    
    │   └──  log.go
    └──  main.go 
├── internal
│   ├──  api_errors
│   ├──  conf
│   ├──  db
│   ├──  enum
│   ├──  helpers
│   ├──  service
│   └──  transport
│      ├──  http
│      └──  interface.go              
├── tests
├── .air.toml
├── .dockerignore
├── .env
├── .gitignore
├── docker-compose.yaml
├── Dockerfile
├── go.mod
├── go.sum
├── README.md
└── Taskfile.yaml
```

The `cmd` folder contains the `main.go` file, which starts the API. If you look at this file, you'll see that its sole purpose is to call the `Run()` method defined in the bootstrap folder. This function is responsible for setting up the API. Specifically, it handles:

- Configuring the application settings.
- Initializing the logger used for API logging.
- Starting the in-memory database.
- Launching the HTTP server.

To configure the application, it was decided that the most optimal approach would be to use environment variables. This ensures that users can easily modify the tool's settings. Additionally, deploying the API in Docker or Kubernetes simplifies configuration management, as environment variables are easy to define in these platforms. The `.env` file contains the environment variables used by the application.

```bash
PORT=3000 # Define the port in which the API will run
HEALTH_PORT=3001 # Define the port in which the health check will run
LOG_LEVEL=debug # Define the log level of the API. It can be debug or info 
```

As you can see in the `.env` file, two ports are specified: one for the API to handle requests and another for the health check. The decision to use a separate port for the health check allows monitoring systems to independently verify the service's health without accessing the main API endpoints. This approach ensures the application remains operational while minimizing the risk of overloading the primary API or exposing sensitive information.

The folder `internal` contains all the logic of the API. Since no packages are going to be externalized, it makes sense to defined all the packages here. 

`api_errors` defines standard errors that can be returned by the API. All errors are represented by the following structured. Moreover, some common errors have already been defined.

```go
// APIError represents an error that is returned to the client.
type APIError struct {
	Code       string `json:"code"`    // error code that can be used to identify the error
	Message    string `json:"message"` // detailed description of the error
	HTTPStatus int    `json:"-"`       // http status code. It is not included in the response body
}
```

```go
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
```

In `conf` you can see how the API reads its configuration, specifically this package reads the environmental variables and loads them in global variable called `GlobalConfig`. To load the configuration into the tool, the package (`viper`)[https://github.com/spf13/viper] has been used.

```go
var GlobalConfig *Config

// Config holds the configuration values for the API
type Config struct {
	Port       string        `mapstructure:"PORT" validate:"required"`        // Port in which the API will listen
	HealthPort string        `mapstructure:"HEALTH_PORT" validate:"required"` // Health port in which the API will listen
	LogLevel   enum.LogLevel `mapstructure:"LOG_LEVEL" validate:"required"`   // Log level for the API: debug, info
}
```

The folder `db` contains the package used to interact with the persistence layer. In this case, keeping in mind that the API should be extendible, an interface has been defined to interact with the database.

```go
// Database is the interface for the database layer.
//
// By using this interface, we can easily swap out the underlying database implementation.
type DatabaseAdapter interface {
	// Account methods
	CreateAccount(account *Account)             // CreateAccount creates a new account
	GetAccountByID(id string) (*Account, error) // GetAccountByID retrieves an account by its ID
	GetAllAccounts() []Account                  // GetAllAccounts retrieves all accounts

	// Transaction methods
	CreateTransaction(transaction *Transaction) error            // CreateTransaction creates a new transaction
	GetTransactionsByAccountID(id string) ([]Transaction, error) // GetTransactionsByAccountID retrieves all transactions for an account
}
```

The previous interface is currently used only by an in-memory database. However, if additional databases are added in the future, they can be easily implemented by adhering to this interface.

```go
// NewDatabaseAdapter creates a new database adapter. In this case there is only one implementation: an in-memory database.
// However, in the future we could add other implementations such as a SQL database. By modifying this function, we can easily
// switch between different database implementations.
func NewDatabaseAdapter(logger *zap.SugaredLogger) DatabaseAdapter {
	return memory.NewInMemoryDatabase(logger)
}
```

Additionally, this package contains the data models that will be stored in the database. Specifically, two entities have been defined: `Account` and `Transaction`

```go
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
```

To store persistently information in the API, a simple in-memory database has been created. This database is composed by two maps: one to store accounts associated with ids and another one to store transactions associated by account ID.

```go
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
```

In the next piece of code it can be seen how an account is stored in memory. In this case, it can be seen the usage of locks to ensure that only one goroutine access the accounts map at a time. The type `sync.Map` could have also been used instead of using regular maps with locks.

```go
// CreateAccount creates a new account in the database.
func (d *inMemoryDatabase) CreateAccount(account *models.Account) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.logger.Debugf("storing account with id '%s' in memory database: %s", account.ID, helpers.PrettyPrintStructResponse(account))
	d.accounts[account.ID] = *account
	d.transactions[account.ID] = make([]models.Transaction, 0)
	d.logger.Debugf("account with id '%s' stored in memory database", account.ID)
}
```

The package `enum` contains enum definitions used by the API. Specifically, two definition: one for log level (`debug` or `info`) and another one for the transaction type (`deposit` or `withdrawal`).

The package `service` contains the business logic of the application. Here, two services have been defined to interact with the accounts and to interact with transactions.

```go
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
```

These services are in charge of recieving information from the transport layer, processing it, and interacting with the database.

Finally, the last package in the `internal` folder is `transport`. This package defines the application's transport layer. Like the database package, it provides an interface to represent this layer, enabling future extensions with additional transport options. At present, only HTTP has been implemented.

```go
// Transporter is an interface for the transport layer. It defines the Serve method that
// will be run by any transport layer implementation (HTTP, gRPC, GraphQL, etc.).
type Transporter interface {
	Serve() error       // starts the transport layer
	HealthCheck() error // starts a health check endpoint that verifies the service is up and running
	Close() error       // handles the graceful shutdown of the transport layer
}

// NewTransporter creates a new transport layer based on the provided type.
func NewTransporter(logger *zap.SugaredLogger, db db.DatabaseAdapter) Transporter {
	return http.NewHttpTransport(logger, db)
}
```

The framework (`chi`)[https://github.com/go-chi/chi] has been used to implement the HTTP server. Additionally, to validate request's bodies the framework (`validator`)[https://github.com/go-playground/validator]. This framework allows users to set multiple rules in the struct tags that can be used to validate the requests. One example of its usage can be seen when creating a transaction. The following struct contains the tag `validate`, which specifies the validation rules. For instance, the field `type` is validated to ensure it is present in the body (`mandatory`) and that its value is either `deposit` or `withdrawal` (`oneof`).

```go
// CreateTransactionRequest is the request schema for the CreateTransaction endpoint.
// It is used to create a new transaction for an account.
type CreateTransactionRequest struct {
	Type   string   `json:"type" validate:"required,oneof=deposit withdrawal"`
	Amount *float64 `json:"amount" validate:"required,gt=0"`
}
```

In the folder `tests`, you can find an integration test of the API. This has been done by dockerizing the API using the library (`testcontainers`)[https://golang.testcontainers.org/] and performing multiple queries to each endpoint of the API.

## Installation and usage

The API can be launch using the tasks defined in the `Taskfile.yaml`, so the package (task)[https://taskfile.dev/] must be installed in your computer. The next commands can be used to run tests and launch the API.

- Run API in development mode (hot reload) by using (Air)[https://github.com/air-verse/air]: `task dev`
- Run API: `task run`
- Run tests: `task test`
- Run race tests: `task race`
- Run integration tests: `task integration_test`
- Run API in docker compose: `task docker`