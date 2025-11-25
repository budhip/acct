package mysql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestLoanPartnerAccountRepositoryTestSuite(t *testing.T) {
	t.Helper()
	suite.Run(t, new(loanPartnerAccountRepositoryTestSuite))
}

type loanPartnerAccountRepositoryTestSuite struct {
	suite.Suite
	t    *testing.T
	db   *sql.DB
	mock sqlmock.Sqlmock
	repo LoanPartnerAccountRepository
}

func (suite *loanPartnerAccountRepositoryTestSuite) SetupTest() {
	var err error

	suite.db, suite.mock, err = sqlmock.New()
	require.NoError(suite.T(), err)

	suite.t = suite.T()
	suite.repo = NewMySQLRepository(suite.db).GetLoanPartnerAccountRepository()
}

func (suite *loanPartnerAccountRepositoryTestSuite) TearDownTest() {
	defer suite.db.Close()
}

func (suite *loanPartnerAccountRepositoryTestSuite) TestRepository_Create() {
	type args struct {
		ctx context.Context
		req models.LoanPartnerAccount
	}
	testCases := []struct {
		name    string
		args    args
		wantErr bool
		doMock  func(args args)
	}{
		{
			name: "success",
			args: args{
				ctx: context.TODO(),
				req: models.LoanPartnerAccount{
					PartnerId:           "efishery",
					LoanKind:            "EFISHERY_LOAN",
					AccountNumber:       "22100100000001",
					AccountType:         "INTERNAL_ACCOUNTS_REVENUE_AMARTHA",
					EntityCode:          "001",
					LoanSubCategoryCode: "13101",
				},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryLoanPartnerAccountCreate)).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "error no result",
			args: args{
				ctx: context.TODO(),
				req: models.LoanPartnerAccount{},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryLoanPartnerAccountCreate)).WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
		{
			name: "error result",
			args: args{
				ctx: context.TODO(),
				req: models.LoanPartnerAccount{},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryLoanPartnerAccountCreate)).WillReturnResult(sqlmock.NewErrorResult(assert.AnError))
			},
			wantErr: true,
		},
		{
			name: "error db",
			args: args{
				ctx: context.TODO(),
				req: models.LoanPartnerAccount{},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryLoanPartnerAccountCreate)).WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.doMock(tt.args)

			err := suite.repo.Create(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *loanPartnerAccountRepositoryTestSuite) TestRepository_Update() {
	type args struct {
		ctx        context.Context
		req        models.UpdateLoanPartnerAccount
		setupMocks func(req models.UpdateLoanPartnerAccount)
	}

	testCases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success case",
			args: args{ctx: context.TODO(),
				req: models.UpdateLoanPartnerAccount{
					PartnerId:     "efishery",
					LoanKind:      "BOILERX_LOAN",
					AccountNumber: "1001",
				},
				setupMocks: func(req models.UpdateLoanPartnerAccount) {
					query, _, _ := queryAccountLoanPartnerUpdate(req)
					suite.mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnResult(sqlmock.NewResult(0, 1))
				},
			},
			wantErr: false,
		},
		{
			name: "error case - database error",
			args: args{
				ctx: context.TODO(),
				req: models.UpdateLoanPartnerAccount{
					PartnerId:     "efishery",
					LoanKind:      "BOILERX_LOAN",
					AccountNumber: "1001",
				},
				setupMocks: func(req models.UpdateLoanPartnerAccount) {
					query, _, _ := queryAccountLoanPartnerUpdate(req)
					suite.mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnError(assert.AnError)
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.args.setupMocks(tt.args.req)

			err := suite.repo.Update(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *loanPartnerAccountRepositoryTestSuite) TestRepository_BulkInsertLoanPartnerAccount() {
	type args struct {
		ctx context.Context
		req []models.LoanPartnerAccount
	}

	valueStrings := "(NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''))"
	expectedQuery := fmt.Sprintf(queryBulkLoanPartnerAccountCreate, valueStrings)

	testCases := []struct {
		name    string
		args    args
		wantErr bool
		doMock  func()
	}{
		{
			name: "success",
			args: args{
				ctx: context.TODO(),
				req: []models.LoanPartnerAccount{{PartnerId: "123", LoanKind: "loan", AccountNumber: "acc", AccountType: "type", EntityCode: "001", LoanSubCategoryCode: "TEST_123"}},
			},
			doMock: func() {
				suite.mock.ExpectExec(regexp.QuoteMeta(expectedQuery)).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "error no row",
			args: args{
				ctx: context.TODO(),
				req: []models.LoanPartnerAccount{{PartnerId: "123", LoanKind: "loan", AccountNumber: "acc", AccountType: "type", EntityCode: "001", LoanSubCategoryCode: "TEST_123"}},
			},
			doMock: func() {
				suite.mock.ExpectExec(regexp.QuoteMeta(expectedQuery)).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
		{
			name: "error result",
			args: args{
				ctx: context.TODO(),
				req: []models.LoanPartnerAccount{{PartnerId: "123", LoanKind: "loan", AccountNumber: "acc", AccountType: "type", EntityCode: "001", LoanSubCategoryCode: "TEST_123"}},
			},
			doMock: func() {
				suite.mock.ExpectExec(regexp.QuoteMeta(expectedQuery)).
					WillReturnResult(sqlmock.NewErrorResult(assert.AnError))
			},
			wantErr: true,
		},
		{
			name: "error db",
			args: args{
				ctx: context.TODO(),
				req: []models.LoanPartnerAccount{{PartnerId: "123", LoanKind: "loan", AccountNumber: "acc", AccountType: "type", EntityCode: "001", LoanSubCategoryCode: "TEST_123"}},
			},
			doMock: func() {
				suite.mock.ExpectExec(regexp.QuoteMeta(expectedQuery)).
					WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range testCases {
		suite.T().Run(tt.name, func(t *testing.T) {
			tt.doMock()

			err := suite.repo.BulkInsertLoanPartnerAccount(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *loanPartnerAccountRepositoryTestSuite) TestRepository_GetByParams() {
	type args struct {
		ctx  context.Context
		opts models.GetLoanPartnerAccountByParamsIn
	}
	defaultColumns := []string{
		`partner_id`,
		`loan_kind`,
		`account_number`,
		`account_type`,
		`entity_code`,
		`loan_sub_category_code`,
		`created_at`,
		`updated_at`,
	}
	row := []driver.Value{
		"efishery",
		"EFISHERY_LOAN",
		"22100100000001",
		"INTERNAL_ACCOUNTS_REVENUE_AMARTHA",
		"001",
		"13101",
		atime.Now(),
		atime.Now(),
	}
	testCases := []struct {
		name       string
		args       args
		setupMocks func(a args)
		wantErr    bool
		expected   []models.LoanPartnerAccount
	}{
		{
			name: "success get all account numbers by param",
			args: args{
				ctx: context.TODO(),
				opts: models.GetLoanPartnerAccountByParamsIn{
					PartnerId:           "efishery",
					LoanKind:            "EFISHERY_LOAN",
					AccountNumber:       "22100100000001",
					AccountType:         "INTERNAL_ACCOUNTS_REVENUE_AMARTHA",
					EntityCode:          "001",
					LoanSubCategoryCode: "13101",
				},
			},
			setupMocks: func(a args) {
				query, _, _ := buildQueryGetLoanPartnerAccountByParam(a.opts)
				rows := sqlmock.
					NewRows(defaultColumns).
					AddRow(row...)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(rows)
			},
			expected: []models.LoanPartnerAccount{
				{
					PartnerId:           "efishery",
					LoanKind:            "EFISHERY_LOAN",
					AccountNumber:       "22100100000001",
					AccountType:         "INTERNAL_ACCOUNTS_REVENUE_AMARTHA",
					EntityCode:          "001",
					LoanSubCategoryCode: "13101",
					CreatedAt:           atime.Now(),
					UpdatedAt:           atime.Now(),
				},
			},
			wantErr: false,
		},
		{
			name: "error row",
			args: args{
				ctx:  context.TODO(),
				opts: models.GetLoanPartnerAccountByParamsIn{},
			},
			setupMocks: func(a args) {
				query, _, _ := buildQueryGetLoanPartnerAccountByParam(a.opts)
				rows := sqlmock.
					NewRows(defaultColumns).
					AddRow(row...).RowError(0, assert.AnError)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
		{
			name: "error scan row",
			args: args{
				ctx:  context.TODO(),
				opts: models.GetLoanPartnerAccountByParamsIn{},
			},
			setupMocks: func(a args) {
				query, _, _ := buildQueryGetLoanPartnerAccountByParam(a.opts)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(sqlmock.NewRows([]string{"InvalidColumn"}).AddRow(nil))
			},
			wantErr: true,
		},
		{
			name: "error database",
			args: args{
				ctx:  context.TODO(),
				opts: models.GetLoanPartnerAccountByParamsIn{},
			},
			setupMocks: func(a args) {
				query, _, _ := buildQueryGetLoanPartnerAccountByParam(a.opts)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(tt.args)

			_, err := suite.repo.GetByParams(tt.args.ctx, tt.args.opts)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
