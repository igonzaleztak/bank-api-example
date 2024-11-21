package tests

import (
	errors "bank_test/internal/api_errors"
	"bank_test/internal/db/models"
	"bank_test/internal/helpers"
	"bank_test/internal/transport/http/schemas"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type integrationSuite struct {
	baseURL string
	c       *testcontainers.Container
	suite.Suite
}

func (s *integrationSuite) SetupSuite() {
	viper.AutomaticEnv()

	port := viper.GetString("PORT")
	health_port := viper.GetString("HEALTH_PORT")
	exposedPorts := []string{fmt.Sprintf("%s/tcp", port), fmt.Sprintf("%s/tcp", health_port)}

	env := map[string]string{
		"PORT":        port,
		"HEALTH_PORT": health_port,
		"LOG_LEVEL":   "info",
	}

	natHealthPort, err := nat.NewPort("tcp", health_port)
	s.Require().NoError(err)

	// run api image in docker
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    "../",
			Dockerfile: "Dockerfile",
		},
		Name:         "integration_test_api",
		ExposedPorts: exposedPorts,
		Env:          env,
		WaitingFor:   wait.ForHTTP("/health").WithPort(natHealthPort),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	s.Require().NoError(err)

	s.c = &container

	mappedPort, err := container.MappedPort(ctx, nat.Port(port))
	s.Require().NoError(err)
	s.baseURL = fmt.Sprintf("http://localhost:%s", mappedPort.Port())
}

func (s *integrationSuite) TearDownSuite() {
	err := testcontainers.TerminateContainer(*s.c)
	s.Require().NoError(err)
}

func (s *integrationSuite) TestCreateAccount() {
	endpoint := "/accounts"

	s.Run("ok: create account", func() {
		input := []struct {
			in                 *schemas.CreateAccountRequest
			err                error
			expectedStatusCode int
		}{
			{
				in: &schemas.CreateAccountRequest{
					Owner:          "John Doe",
					InitialBalance: helpers.PointerValue(100.0),
				},
				expectedStatusCode: 201,
			},
		}

		for _, tt := range input {
			jsonData, err := json.Marshal(tt.in)
			s.Require().NoError(err)

			resp, err := http.Post(s.baseURL+endpoint, "application/json", bytes.NewBuffer(jsonData))
			s.Require().NoError(err)
			s.Require().Equal(tt.expectedStatusCode, resp.StatusCode)
		}
	})

	s.Run("fail: create account with invalid body", func() {
		input := []struct {
			in                 *schemas.CreateAccountRequest
			out                *errors.APIError
			err                error
			expectedStatusCode int
		}{
			{
				in: &schemas.CreateAccountRequest{
					Owner: "John Doe",
				},
				out:                errors.ErrInvalidBody,
				expectedStatusCode: errors.ErrInvalidBody.HTTPStatus,
			},
			{
				in: &schemas.CreateAccountRequest{
					InitialBalance: helpers.PointerValue(100.0),
				},
				out:                errors.ErrInvalidBody,
				expectedStatusCode: errors.ErrInvalidBody.HTTPStatus,
			},
			{
				in:                 &schemas.CreateAccountRequest{},
				out:                errors.ErrInvalidBody,
				expectedStatusCode: errors.ErrInvalidBody.HTTPStatus,
			},
		}

		for _, tt := range input {
			jsonData, err := json.Marshal(tt.in)
			s.Require().NoError(err)

			resp, err := http.Post(s.baseURL+endpoint, "application/json", bytes.NewBuffer(jsonData))
			s.Require().NoError(err)

			defer resp.Body.Close()

			responseJSON, _ := io.ReadAll(resp.Body)

			var body errors.APIError
			err = json.Unmarshal(responseJSON, &body)
			s.Require().NoError(err)

			helpers.PrettyPrintStruct(body)

			s.Equal(tt.expectedStatusCode, resp.StatusCode)
			s.Equal(tt.out.Code, body.Code)
		}

	})
}

func (s *integrationSuite) TestGetAccount() {
	endpoint := "/accounts"

	// create account
	input := &schemas.CreateAccountRequest{
		Owner:          "John Doe",
		InitialBalance: helpers.PointerValue(100.0),
	}

	jsonData, err := json.Marshal(input)
	s.Require().NoError(err)
	resp, err := http.Post(s.baseURL+endpoint, "application/json", bytes.NewBuffer(jsonData))
	s.Require().NoError(err)
	defer resp.Body.Close()

	var account models.Account
	err = json.NewDecoder(resp.Body).Decode(&account)
	s.Require().NoError(err)

	s.Run("ok: get account", func() {
		method := fmt.Sprintf("%s/%s", endpoint, account.ID)
		resp, err := http.Get(s.baseURL + method)
		s.Require().NoError(err)
		s.Require().Equal(200, resp.StatusCode)
		defer resp.Body.Close()

		var body models.Account
		err = json.NewDecoder(resp.Body).Decode(&body)
		s.Require().NoError(err)

		s.Equal(account.ID, body.ID)
		s.Equal(account.Owner, body.Owner)
		s.Equal(account.Balance, body.Balance)
	})

	s.Run("fail: get account with invalid account id", func() {
		method := fmt.Sprintf("%s/%s", endpoint, "invalid_id")
		resp, err := http.Get(s.baseURL + method)
		s.Require().NoError(err)
		s.Require().Equal(400, resp.StatusCode)
		defer resp.Body.Close()

		var body errors.APIError
		err = json.NewDecoder(resp.Body).Decode(&body)
		s.Require().NoError(err)

		s.Equal(errors.ErrInvalidAccountID.Code, body.Code)
	})

	s.Run("fail: get account with account not found", func() {
		method := fmt.Sprintf("%s/%s", endpoint, uuid.NewString())
		resp, err := http.Get(s.baseURL + method)
		s.Require().NoError(err)
		s.Require().Equal(400, resp.StatusCode)
		defer resp.Body.Close()

		var body errors.APIError
		err = json.NewDecoder(resp.Body).Decode(&body)
		s.Require().NoError(err)

		s.Equal(errors.ErrAccountNotFound.Code, body.Code)
	})
}

func (s *integrationSuite) TestGetAllAccounts() {
	endpoint := "/accounts"

	// create 3 accounts
	accounts := []schemas.CreateAccountRequest{
		{
			Owner:          "John Doe",
			InitialBalance: helpers.PointerValue(100.0),
		},
		{
			Owner:          "Jane Doe",
			InitialBalance: helpers.PointerValue(200.0),
		},
		{
			Owner:          "Alice",
			InitialBalance: helpers.PointerValue(300.0),
		},
	}
	for _, account := range accounts {
		jsonData, err := json.Marshal(account)
		s.Require().NoError(err)

		resp, err := http.Post(s.baseURL+endpoint, "application/json", bytes.NewBuffer(jsonData))
		s.Require().NoError(err)
		defer resp.Body.Close()
	}

	s.Run("ok: get all accounts", func() {
		resp, err := http.Get(s.baseURL + endpoint)
		s.Require().NoError(err)
		s.Require().Equal(200, resp.StatusCode)
		defer resp.Body.Close()

		var body []models.Account
		err = json.NewDecoder(resp.Body).Decode(&body)
		s.Require().NoError(err)

		s.NotEmpty(body)
		s.NotZero(len(body)) // in this case we cannot exactly know the length since we are performing multiple tests in parallel
	})
}

func (s *integrationSuite) TestCreateTransaction() {
	endpoint := "/accounts"

	// create 1 account
	account := schemas.CreateAccountRequest{
		Owner:          "transaction account",
		InitialBalance: helpers.PointerValue(100.0),
	}
	jsonData, err := json.Marshal(account)
	s.Require().NoError(err)

	resp, err := http.Post(s.baseURL+"/accounts", "application/json", bytes.NewBuffer(jsonData))
	s.Require().NoError(err)
	defer resp.Body.Close()

	var body models.Account
	err = json.NewDecoder(resp.Body).Decode(&body)
	s.Require().NoError(err)

	s.Run("ok: create transaction", func() {
		path := fmt.Sprintf("%s/%s/transactions", endpoint, body.ID)

		input := []struct {
			in                 *schemas.CreateTransactionRequest
			expectedStatusCode int
		}{
			{
				in: &schemas.CreateTransactionRequest{
					Type:   "deposit",
					Amount: helpers.PointerValue(100.0),
				},
				expectedStatusCode: http.StatusCreated,
			},
			{
				in: &schemas.CreateTransactionRequest{
					Type:   "withdrawal",
					Amount: helpers.PointerValue(50.0),
				},
				expectedStatusCode: http.StatusCreated,
			},
		}

		for _, tt := range input {
			jsonData, err := json.Marshal(tt.in)
			s.Require().NoError(err)

			resp, err := http.Post(s.baseURL+path, "application/json", bytes.NewBuffer(jsonData))
			s.Require().NoError(err)
			s.Require().Equal(tt.expectedStatusCode, resp.StatusCode)

			// get account
			resp, err = http.Get(s.baseURL + fmt.Sprintf("%s/%s", endpoint, body.ID))
			s.Require().NoError(err)
			s.Require().Equal(200, resp.StatusCode)
			defer resp.Body.Close()

			var account models.Account
			err = json.NewDecoder(resp.Body).Decode(&account)
			s.Require().NoError(err)

			expectedBalance := body.Balance
			if *tt.in.Amount > 0 {
				expectedBalance += *tt.in.Amount
			} else {
				expectedBalance -= *tt.in.Amount
			}

			s.Equal(expectedBalance, account.Balance)
		}
	})
}

func (s *integrationSuite) TestGetTransactionByAccountID() {
	endpoint := "/accounts"

	// create 1 account
	account := schemas.CreateAccountRequest{
		Owner:          "get by account id test",
		InitialBalance: helpers.PointerValue(100.0),
	}
	jsonData, err := json.Marshal(account)
	s.Require().NoError(err)

	resp, err := http.Post(s.baseURL+"/accounts", "application/json", bytes.NewBuffer(jsonData))
	s.Require().NoError(err)

	var body models.Account
	err = json.NewDecoder(resp.Body).Decode(&body)
	s.Require().NoError(err)
	defer resp.Body.Close()

	// create 1 transaction
	path := fmt.Sprintf("%s/%s/transactions", endpoint, body.ID)
	transaction := schemas.CreateTransactionRequest{
		Type:   "deposit",
		Amount: helpers.PointerValue(100.0),
	}
	jsonData, err = json.Marshal(transaction)
	s.Require().NoError(err)

	_, err = http.Post(s.baseURL+path, "application/json", bytes.NewBuffer(jsonData))
	s.Require().NoError(err)

	s.Run("ok: get transactions by account id", func() {
		resp, err := http.Get(s.baseURL + fmt.Sprintf("%s/%s/transactions", endpoint, body.ID))
		s.Require().NoError(err)
		s.Require().Equal(200, resp.StatusCode)
		defer resp.Body.Close()

		var transactions []models.Transaction
		err = json.NewDecoder(resp.Body).Decode(&transactions)
		s.Require().NoError(err)

		s.NotEmpty(transactions)
		s.NotZero(len(transactions))

		s.Equal(transaction.Type, transactions[0].Type.String())
		s.Equal(*transaction.Amount, transactions[0].Amount)
		s.Equal(body.ID, transactions[0].AccountID)

	})
}

func (s *integrationSuite) TestTransfer() {
	accountsEndpoint := "/accounts"

	// create 2 accounts
	accounts := []schemas.CreateAccountRequest{
		{
			Owner:          "from",
			InitialBalance: helpers.PointerValue(100.0),
		},
		{
			Owner:          "to",
			InitialBalance: helpers.PointerValue(0.0),
		},
	}

	accountStored := make(map[string]*models.Account)
	for _, account := range accounts {
		jsonData, err := json.Marshal(account)
		s.Require().NoError(err)

		resp, err := http.Post(s.baseURL+accountsEndpoint, "application/json", bytes.NewBuffer(jsonData))
		s.Require().NoError(err)
		defer resp.Body.Close()

		var body models.Account
		err = json.NewDecoder(resp.Body).Decode(&body)
		s.Require().NoError(err)

		accountStored[body.Owner] = &body
	}

	s.Run("ok: transfer", func() {
		body := schemas.TransferRequest{
			FromAccountId: accountStored["from"].ID,
			ToAccountId:   accountStored["to"].ID,
			Amount:        helpers.PointerValue(50.0),
		}

		jsonData, err := json.Marshal(body)
		s.Require().NoError(err)

		resp, err := http.Post(s.baseURL+"/transfer", "application/json", bytes.NewBuffer(jsonData))
		s.Require().NoError(err)
		s.Require().Equal(201, resp.StatusCode)

		// get accounts
		for _, account := range accounts {
			resp, err := http.Get(s.baseURL + fmt.Sprintf("%s/%s", accountsEndpoint, accountStored[account.Owner].ID))
			s.Require().NoError(err)
			s.Require().Equal(200, resp.StatusCode)
			defer resp.Body.Close()

			var acc models.Account
			err = json.NewDecoder(resp.Body).Decode(&acc)
			s.Require().NoError(err)

			expectedBalance := 50.0 // 100 - 50 for one user and 0 + 50 for the other
			s.Equal(expectedBalance, acc.Balance)
		}

	})

}

// TestIntegrationSuite runs the integration test suite.
func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(integrationSuite))
}
