package mysql

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func (suite *accountingTestSuite) TestRepository_CalculateFromAccountBalanceDaily() {
	type args struct {
		ctx context.Context
		in  models.CalculateTrialBalance
	}
	testCases := []struct {
		name       string
		args       args
		setupMocks func(a args)
		wantErr    bool
	}{
		{
			name: "success case - calculate debit credit movement",
			args: args{
				ctx: context.TODO(),
				in:  models.CalculateTrialBalance{},
			},
			setupMocks: func(a args) {
				query, _, _ := buildCalculateFromAccountBalanceDaily(a.in)
				rows := sqlmock.
					NewRows([]string{
						`aadb.entity_code`,
						`aadb.category_code`,
						`aadb.sub_category_code`,
						`debit_movement`,
						`credit_movement`,
					}).
					AddRow("001", "111", "11403", 0, 0)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "error case - QueryContext",
			args: args{
				ctx: context.TODO(),
				in:  models.CalculateTrialBalance{},
			},
			setupMocks: func(a args) {
				query, _, _ := buildCalculateFromAccountBalanceDaily(a.in)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - row scan",
			args: args{
				ctx: context.TODO(),
				in:  models.CalculateTrialBalance{},
			},
			setupMocks: func(a args) {
				query, _, _ := buildCalculateFromAccountBalanceDaily(a.in)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(sqlmock.NewRows([]string{"InvalidColumn"}).AddRow(nil))
			},
			wantErr: true,
		},
		{
			name: "error case - row error",
			args: args{
				ctx: context.TODO(),
				in:  models.CalculateTrialBalance{},
			},
			setupMocks: func(a args) {
				query, _, _ := buildCalculateFromAccountBalanceDaily(a.in)
				rows := sqlmock.
					NewRows([]string{
						`aadb.entity_code`,
						`aadb.category_code`,
						`aadb.sub_category_code`,
						`debit_movement`,
						`credit_movement`,
					}).
					AddRow("001", "111", "11403", 0, 0).
					RowError(0, assert.AnError)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(tt.args)

			_, err := suite.repo.CalculateFromAccountBalanceDaily(tt.args.ctx, tt.args.in)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountingTestSuite) TestRepository_CalculateFromTransactions() {
	type args struct {
		ctx context.Context
		in  models.CalculateTrialBalance
	}
	testCases := []struct {
		name       string
		args       args
		setupMocks func(a args)
		wantErr    bool
	}{
		{
			name: "success case - calculate debit credit movement",
			args: args{
				ctx: context.TODO(),
				in:  models.CalculateTrialBalance{},
			},
			setupMocks: func(a args) {
				query, _, _ := buildCalculateFromTransactions(a.in)
				rows := sqlmock.
					NewRows([]string{
						`entity_code`,
						`category_code`,
						`sub_category_code`,
						`debit_movement`,
						`credit_movement`,
					}).
					AddRow("001", "111", "11403", 0, 0)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "error case - QueryContext",
			args: args{
				ctx: context.TODO(),
				in:  models.CalculateTrialBalance{},
			},
			setupMocks: func(a args) {
				query, _, _ := buildCalculateFromTransactions(a.in)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - row scan",
			args: args{
				ctx: context.TODO(),
				in:  models.CalculateTrialBalance{},
			},
			setupMocks: func(a args) {
				query, _, _ := buildCalculateFromTransactions(a.in)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(sqlmock.NewRows([]string{"InvalidColumn"}).AddRow(nil))
			},
			wantErr: true,
		},
		{
			name: "error case - row error",
			args: args{
				ctx: context.TODO(),
				in:  models.CalculateTrialBalance{},
			},
			setupMocks: func(a args) {
				query, _, _ := buildCalculateFromTransactions(a.in)
				rows := sqlmock.
					NewRows([]string{
						`entity_code`,
						`category_code`,
						`sub_category_code`,
						`debit_movement`,
						`credit_movement`,
					}).
					AddRow("001", "111", "11403", 0, 0).
					RowError(0, assert.AnError)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(tt.args)

			_, err := suite.repo.CalculateFromTransactions(tt.args.ctx, tt.args.in)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountingTestSuite) TestRepository_GetOpeningBalanceFromAccountTrialBalance() {
	type args struct {
		ctx context.Context
		in  models.CalculateTrialBalance
	}
	testCases := []struct {
		name       string
		args       args
		setupMocks func(a args)
		wantErr    bool
	}{
		{
			name: "success case - get account opening balance",
			args: args{
				ctx: context.TODO(),
				in:  models.CalculateTrialBalance{},
			},
			setupMocks: func(a args) {
				rows := sqlmock.
					NewRows([]string{"closing_balance"}).
					AddRow(0)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetOpeingBalanceFromAccountTrialBalance)).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "error case - QueryContext",
			args: args{
				ctx: context.TODO(),
				in:  models.CalculateTrialBalance{},
			},
			setupMocks: func(a args) {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetOpeingBalanceFromAccountTrialBalance)).
					WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - row scan",
			args: args{
				ctx: context.TODO(),
				in:  models.CalculateTrialBalance{},
			},
			setupMocks: func(a args) {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetOpeingBalanceFromAccountTrialBalance)).
					WillReturnRows(sqlmock.NewRows([]string{"InvalidColumn"}).AddRow(nil))
			},
			wantErr: true,
		},
		{
			name: "error case - row error",
			args: args{
				ctx: context.TODO(),
				in:  models.CalculateTrialBalance{},
			},
			setupMocks: func(a args) {
				rows := sqlmock.
					NewRows([]string{"closing_balance"}).
					AddRow(0).RowError(0, assert.AnError)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetOpeingBalanceFromAccountTrialBalance)).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(tt.args)

			_, err := suite.repo.GetOpeningBalanceFromAccountTrialBalance(tt.args.ctx, tt.args.in)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountingTestSuite) TestRepository_InsertAccountTrialBalance() {
	type args struct {
		ctx context.Context
		req []models.AccountTrialBalance
	}
	valueStrings := "(?, ?, ?, ?, ?, ?, ?, ?)"
	queryInsertAccountTrialBalance = fmt.Sprintf(queryInsertAccountTrialBalance, valueStrings)

	testCases := []struct {
		name    string
		args    args
		wantErr bool
		doMock  func(args args)
	}{
		{
			name: "success case - insert account trial balance daily",
			args: args{
				ctx: context.TODO(),
				req: []models.AccountTrialBalance{
					{}, {},
				},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryInsertAccountTrialBalance)).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "error case - database error",
			args: args{
				ctx: context.TODO(),
				req: []models.AccountTrialBalance{
					{}, {},
				},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryInsertAccountTrialBalance)).WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - result error",
			args: args{
				ctx: context.TODO(),
				req: []models.AccountTrialBalance{},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryInsertAccountTrialBalance)).WillReturnResult(sqlmock.NewErrorResult(assert.AnError))
			},
			wantErr: true,
		},
		{
			name: "error case - no affected row",
			args: args{
				ctx: context.TODO(),
				req: []models.AccountTrialBalance{},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryInsertAccountTrialBalance)).WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.doMock(tt.args)

			err := suite.repo.InsertAccountTrialBalance(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountingTestSuite) TestRepository_GetTrialBalanceV2() {
	rowsCategorySubCategoryCOAType := sqlmock.
		NewRows([]string{
			`category_code`,
			`category_name`,
			`sub_category_code`,
			`sub_category_name`,
			`coa_type_code`,
			`coa_type_name`,
		}).
		AddRow("111", "Cash Point", "11101", "Cash Point", "AST", "Asset")
	defaultColumns := []string{
		`aatb.entity_code`,
		`aatb.category_code`,
		`aatb.sub_category_code`,
		`debit_movement`,
		`credit_movement`,
		`opening_balance`,
		`closing_balance`,
	}
	dec, _ := decimal.NewFromString("100000")
	defaultDecimal := decimal.NewNullDecimal(dec)
	date := atime.Now()

	type args struct {
		ctx  context.Context
		opts models.TrialBalanceFilterOptions
	}
	testCases := []struct {
		name       string
		args       args
		setupMocks func(a args)
		wantErr    bool
		expected   []models.GetTrialBalanceV2Out
	}{
		{
			name: "success get trial balance",
			args: args{
				ctx: context.TODO(),
				opts: models.TrialBalanceFilterOptions{
					EntityCode:      "000",
					StartDate:       date,
					EndDate:         date,
					SubCategoryCode: "A.111.01",
				},
			},
			setupMocks: func(a args) {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetCategorySubCategoryCOAType)).
					WillReturnRows(rowsCategorySubCategoryCOAType)
				listQuery, _, _ := buildGetTrialBalanceV2Query(a.opts)
				rows := sqlmock.
					NewRows(defaultColumns).
					AddRow(
						"000",
						"A.111",
						"A.111.01",
						defaultDecimal,
						defaultDecimal,
						defaultDecimal,
						defaultDecimal,
					).
					AddRow(
						"000",
						"A.111",
						"A.111.01",
						decimal.NewFromFloat(0),
						decimal.NewFromFloat(0),
						decimal.NewFromFloat(0),
						decimal.NewFromFloat(0),
					)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(listQuery)).
					WillReturnRows(rows)
			},
			expected: []models.GetTrialBalanceV2Out{
				{
					CoaTypeCode:     "000",
					CoaTypeName:     "asset",
					CategoryCode:    "A.111",
					CategoryName:    "Kas Teller",
					SubCategoryCode: "A.111.01",
					SubCategoryName: "Kas Teller Point",
					OpeningBalance:  defaultDecimal.Decimal,
					DebitMovement:   defaultDecimal.Decimal,
					CreditMovement:  defaultDecimal.Decimal,
					ClosingBalance:  defaultDecimal.Decimal,
				},
			},
			wantErr: false,
		},
		{
			name: "error row",
			args: args{
				ctx: context.TODO(),
				opts: models.TrialBalanceFilterOptions{
					EntityCode: "000",
					StartDate:  date,
					EndDate:    date,
				},
			},
			setupMocks: func(a args) {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetCategorySubCategoryCOAType)).
					WillReturnRows(rowsCategorySubCategoryCOAType)
				listQuery, _, _ := buildGetTrialBalanceV2Query(a.opts)
				rows := sqlmock.
					NewRows(defaultColumns).
					AddRow(
						"000",
						"A.111",
						"A.111.01",
						defaultDecimal,
						defaultDecimal,
						defaultDecimal,
						defaultDecimal,
					).RowError(0, assert.AnError)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(listQuery)).
					WillReturnRows(rows)
			},
			expected: []models.GetTrialBalanceV2Out{},
			wantErr:  true,
		},
		{
			name: "error scan row",
			args: args{
				ctx: context.TODO(),
				opts: models.TrialBalanceFilterOptions{
					EntityCode: "000",
					StartDate:  date,
					EndDate:    date,
				},
			},
			setupMocks: func(a args) {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetCategorySubCategoryCOAType)).
					WillReturnRows(rowsCategorySubCategoryCOAType)
				listQuery, _, _ := buildGetTrialBalanceV2Query(a.opts)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(listQuery)).
					WillReturnRows(sqlmock.NewRows([]string{"InvalidColumn"}).AddRow(nil))
			},
			expected: []models.GetTrialBalanceV2Out{},
			wantErr:  true,
		},
		{
			name: "error database",
			args: args{
				ctx: context.TODO(),
				opts: models.TrialBalanceFilterOptions{
					EntityCode: "000",
					StartDate:  date,
					EndDate:    date,
				},
			},
			setupMocks: func(a args) {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetCategorySubCategoryCOAType)).
					WillReturnRows(rowsCategorySubCategoryCOAType)
				listQuery, _, _ := buildGetTrialBalanceV2Query(a.opts)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(listQuery)).
					WillReturnError(assert.AnError)
			},
			expected: []models.GetTrialBalanceV2Out{},
			wantErr:  true,
		},
	}

	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(tt.args)

			_, _, err := suite.repo.GetTrialBalanceV2(tt.args.ctx, tt.args.opts)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountingTestSuite) TestRepository_GetTrialBalanceSubCategory() {
	type args struct {
		ctx  context.Context
		opts models.TrialBalanceFilterOptions
	}
	defaultColumns := []string{
		"sub_category_code",
		"sub_category_code_name",
		"opening_balance",
		"closing_balance",
		"debit_movement",
		"credit_movement",
	}

	defaultDecimal := decimal.NewNullDecimal(decimal.NewFromInt(100_000))
	date := atime.Now()

	testCases := []struct {
		name       string
		args       args
		setupMocks func(a args)
		wantErr    bool
	}{
		{
			name: "success get trial balance account list",
			args: args{
				ctx: context.TODO(),
				opts: models.TrialBalanceFilterOptions{
					SubCategoryCode: "21108",
					StartDate:       date,
					EndDate:         date,
				},
			},
			setupMocks: func(a args) {
				listQuery, _, _ := buildGetTrialBalanceSubCategory(a.opts)
				rows := sqlmock.
					NewRows(defaultColumns).
					AddRow(
						"21108",
						"Lender's Cash - Earn (Fixed Income)",
						defaultDecimal,
						defaultDecimal,
						defaultDecimal,
						defaultDecimal,
					)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(listQuery)).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "error scan row",
			args: args{
				ctx: context.TODO(),
				opts: models.TrialBalanceFilterOptions{
					SubCategoryCode: "21108",
					StartDate:       date,
					EndDate:         date,
				},
			},
			setupMocks: func(a args) {
				listQuery, _, _ := buildGetTrialBalanceSubCategory(a.opts)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(listQuery)).
					WillReturnRows(sqlmock.NewRows([]string{"InvalidColumn"}).AddRow(nil))
			},
			wantErr: true,
		},
		{
			name: "error database",
			args: args{
				ctx: context.TODO(),
				opts: models.TrialBalanceFilterOptions{
					SubCategoryCode: "21108",
					StartDate:       date,
					EndDate:         date,
				},
			},
			setupMocks: func(a args) {
				listQuery, _, _ := buildGetTrialBalanceSubCategory(a.opts)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(listQuery)).
					WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(tt.args)

			_, err := suite.repo.GetTrialBalanceSubCategory(tt.args.ctx, tt.args.opts)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountingTestSuite) TestRepository_GetTransactionsToday() {
	defaultColumns := []string{
		"transaction_id",
	}
	date := atime.Now()

	type args struct {
		ctx context.Context
		req time.Time
	}
	testCases := []struct {
		name       string
		args       args
		setupMocks func(a args)
		wantErr    bool
	}{
		{
			name: "success case",
			args: args{
				ctx: context.TODO(),
				req: date,
			},
			setupMocks: func(a args) {
				listQuery, _, _ := buildGetTransactionsToday(a.req)
				rows := sqlmock.
					NewRows(defaultColumns).
					AddRow(
						"12345",
					)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(listQuery)).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "error scan row",
			args: args{
				ctx: context.TODO(),
				req: date,
			},
			setupMocks: func(a args) {
				listQuery, _, _ := buildGetTransactionsToday(a.req)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(listQuery)).
					WillReturnRows(sqlmock.NewRows([]string{"InvalidColumn"}).AddRow(nil))
			},
			wantErr: true,
		},
		{
			name: "error database",
			args: args{
				ctx: context.TODO(),
				req: date,
			},
			setupMocks: func(a args) {
				listQuery, _, _ := buildGetTransactionsToday(a.req)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(listQuery)).
					WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(tt.args)

			_, err := suite.repo.GetTransactionsToday(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
