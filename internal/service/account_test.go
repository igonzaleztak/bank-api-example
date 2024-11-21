package service

import (
	errors "bank_test/internal/api_errors"
	"bank_test/internal/db"
	"bank_test/internal/db/memory"
	"bank_test/internal/db/models"
	"bank_test/internal/helpers"
	"bank_test/internal/transport/http/schemas"
	"testing"

	"github.com/google/uuid"
	"github.com/ledongthuc/goterators"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type accountSuite struct {
	db db.DatabaseAdapter
	as AccountService
	suite.Suite
}

func (s *accountSuite) SetupTest() {
	logger := zap.NewExample().Sugar()

	s.db = memory.NewInMemoryDatabase(logger)
	s.as = NewAccountService(logger, s.db)
}

func (s *accountSuite) TestCreateAccount() {
	s.Run("ok", func() {
		inputData := []struct {
			in  *schemas.CreateAccountRequest
			out *models.Account
		}{
			{
				in: &schemas.CreateAccountRequest{
					Owner:          "Alice",
					InitialBalance: helpers.PointerValue(float64(20)),
				},
				out: &models.Account{
					Owner:   "Alice",
					Balance: 20,
				},
			},
			{
				in: &schemas.CreateAccountRequest{
					Owner:          "Bob",
					InitialBalance: helpers.PointerValue(float64(100)),
				},
				out: &models.Account{
					Owner:   "Bob",
					Balance: 100,
				},
			},
		}

		for _, data := range inputData {
			acc := s.as.CreateAccount(data.in)
			s.Equal(data.out.Owner, acc.Owner)
			s.Equal(data.out.Balance, acc.Balance)

			// check if the account was saved in the database
			accFromDB, err := s.db.GetAccountByID(acc.ID)
			s.NoError(err)
			s.Equal(data.out.Owner, accFromDB.Owner)
			s.Equal(data.out.Balance, accFromDB.Balance)
		}
	})
}

func (s *accountSuite) TestGetAccountByID() {
	// create an account
	account := schemas.CreateAccountRequest{
		Owner:          "Alice",
		InitialBalance: helpers.PointerValue(float64(20)),
	}
	createdAccount := s.as.CreateAccount(&account)

	s.Run("ok", func() {
		inputData := []struct {
			in  string
			out *models.Account
			err error
		}{
			{
				in:  createdAccount.ID,
				out: createdAccount,
				err: nil,
			},
			{
				in:  uuid.NewString(),
				out: nil,
				err: errors.ErrAccountNotFound,
			},
		}

		for _, data := range inputData {
			acc, err := s.as.GetAccountByID(data.in)
			s.Equal(data.err, err)

			if data.out != nil {
				s.Equal(data.out.Owner, acc.Owner)
				s.Equal(data.out.Balance, acc.Balance)
			}
		}
	})
}

func (s *accountSuite) TestGetAllAccounts() {
	// create accounts
	accounts := []schemas.CreateAccountRequest{
		{
			Owner:          "Alice",
			InitialBalance: helpers.PointerValue(float64(20)),
		},
		{
			Owner:          "Bob",
			InitialBalance: helpers.PointerValue(float64(100)),
		},
	}

	for _, acc := range accounts {
		s.as.CreateAccount(&acc)
	}

	s.Run("ok", func() {
		accs := s.as.GetAllAccounts()
		s.Equal(len(accounts), len(accs))

		responseOwners := goterators.Map(accs, func(acc models.Account) string {
			return acc.Owner
		})
		expectedOwners := goterators.Map(accounts, func(acc schemas.CreateAccountRequest) string {
			return acc.Owner
		})

		includesAllOwners := goterators.Include(responseOwners, expectedOwners)
		s.True(includesAllOwners)
	})
}

func TestAccountSuite(t *testing.T) {
	suite.Run(t, new(accountSuite))
}
