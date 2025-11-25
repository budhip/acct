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
	"github.com/stretchr/testify/assert"
)

func (suite *accountingTestSuite) TestRepository_GetOpeningBalanceByDate() {
	date, _ := atime.NowZeroTime()
	type args struct {
		ctx           context.Context
		accountNumber string
		date          time.Time
	}
	testCases := []struct {
		name       string
		args       args
		setupMocks func(a args)
		wantErr    bool
	}{
		{
			name: "success case - get account opening balance by date",
			args: args{
				ctx:           context.TODO(),
				accountNumber: "",
				date:          date,
			},
			setupMocks: func(a args) {
				rows := sqlmock.
					NewRows([]string{"closing_balance"}).
					AddRow(0)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetOpeningBalanceDate)).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "error case - QueryContext",
			args: args{
				ctx:           context.TODO(),
				accountNumber: "",
				date:          date,
			},
			setupMocks: func(a args) {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetOpeningBalanceDate)).
					WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - row scan",
			args: args{
				ctx:           context.TODO(),
				accountNumber: "",
				date:          date,
			},
			setupMocks: func(a args) {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetOpeningBalanceDate)).
					WillReturnRows(sqlmock.NewRows([]string{"InvalidColumn"}).AddRow(nil))
			},
			wantErr: true,
		},
		{
			name: "error case - row error",
			args: args{
				ctx:           context.TODO(),
				accountNumber: "",
				date:          date,
			},
			setupMocks: func(a args) {
				rows := sqlmock.
					NewRows([]string{"closing_balance"}).
					AddRow(0).RowError(0, assert.AnError)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetOpeningBalanceDate)).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(tt.args)

			_, err := suite.repo.GetOpeningBalanceByDate(tt.args.ctx, tt.args.accountNumber, tt.args.date)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountingTestSuite) TestRepository_GetLastOpeningBalance() {
	date, _ := atime.NowZeroTime()
	type args struct {
		ctx           context.Context
		accountNumber string
		date          time.Time
	}
	testCases := []struct {
		name       string
		args       args
		setupMocks func(a args)
		wantErr    bool
	}{
		{
			name: "success case - get account opening balance by date",
			args: args{
				ctx:           context.TODO(),
				accountNumber: "",
				date:          date,
			},
			setupMocks: func(a args) {
				rows := sqlmock.
					NewRows([]string{"closing_balance"}).
					AddRow(0)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetLastOpeningBalance)).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "error case - QueryContext",
			args: args{
				ctx:           context.TODO(),
				accountNumber: "",
				date:          date,
			},
			setupMocks: func(a args) {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetLastOpeningBalance)).
					WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - row scan",
			args: args{
				ctx:           context.TODO(),
				accountNumber: "",
				date:          date,
			},
			setupMocks: func(a args) {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetLastOpeningBalance)).
					WillReturnRows(sqlmock.NewRows([]string{"InvalidColumn"}).AddRow(nil))
			},
			wantErr: true,
		},
		{
			name: "error case - row error",
			args: args{
				ctx:           context.TODO(),
				accountNumber: "",
				date:          date,
			},
			setupMocks: func(a args) {
				rows := sqlmock.
					NewRows([]string{"closing_balance"}).
					AddRow(0).RowError(0, assert.AnError)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetLastOpeningBalance)).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(tt.args)

			_, err := suite.repo.GetLastOpeningBalance(tt.args.ctx, tt.args.accountNumber, tt.args.date)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountingTestSuite) TestRepository_CalculateOpeningClosingBalance() {
	date, _ := atime.NowZeroTime()
	type args struct {
		ctx           context.Context
		accountNumber string
		date          time.Time
	}
	testCases := []struct {
		name       string
		args       args
		setupMocks func(a args)
		wantErr    bool
	}{
		{
			name: "success case - calculate debit credit opening closing",
			args: args{
				ctx:           context.TODO(),
				accountNumber: "",
				date:          date,
			},
			setupMocks: func(a args) {
				query, _, _ := buildCalculateOpeningClosingBalance(a.accountNumber, a.date)
				rows := sqlmock.
					NewRows([]string{
						`account_number`,
						`entity_code`,
						`category_code`,
						`sub_category_code`,
						`opening`,
						`closing`,
					}).
					AddRow("", "", 0, 0, 0, 0)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "error case - QueryContext",
			args: args{
				ctx:           context.TODO(),
				accountNumber: "",
				date:          date,
			},
			setupMocks: func(a args) {
				query, _, _ := buildCalculateOpeningClosingBalance(a.accountNumber, a.date)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - row scan",
			args: args{
				ctx:           context.TODO(),
				accountNumber: "",
				date:          date,
			},
			setupMocks: func(a args) {
				query, _, _ := buildCalculateOpeningClosingBalance(a.accountNumber, a.date)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(sqlmock.NewRows([]string{"InvalidColumn"}).AddRow(nil))
			},
			wantErr: true,
		},
		{
			name: "error case - row error",
			args: args{
				ctx:           context.TODO(),
				accountNumber: "",
				date:          date,
			},
			setupMocks: func(a args) {
				query, _, _ := buildCalculateOpeningClosingBalance(a.accountNumber, a.date)
				rows := sqlmock.
					NewRows([]string{
						`account_number`,
						`entity_code`,
						`category_code`,
						`sub_category_code`,
						`debit`,
						`credit`,
						`opening`,
						`closing`,
					}).
					AddRow("", "", "", "", 0, 0, 0, 0).
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

			_, err := suite.repo.CalculateOpeningClosingBalance(tt.args.ctx, tt.args.accountNumber, tt.args.date)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountingTestSuite) TestRepository_CalculateOpeningClosingBalanceFromAccountBalance() {
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
			name: "success case",
			args: args{
				ctx: context.TODO(),
				in:  models.CalculateTrialBalance{},
			},
			setupMocks: func(a args) {
				rows := sqlmock.
					NewRows([]string{"opening_balance"}).
					AddRow(0)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryCalculateOpeningClosingBalanceFromAccountBalance)).
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
					ExpectQuery(regexp.QuoteMeta(queryCalculateOpeningClosingBalanceFromAccountBalance)).
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
					ExpectQuery(regexp.QuoteMeta(queryCalculateOpeningClosingBalanceFromAccountBalance)).
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
					ExpectQuery(regexp.QuoteMeta(queryCalculateOpeningClosingBalanceFromAccountBalance)).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(tt.args)

			_, err := suite.repo.CalculateOpeningClosingBalanceFromAccountBalance(tt.args.ctx, tt.args.in)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountingTestSuite) TestRepository_GetOneAccountBalanceDaily() {
	date, _ := atime.NowZeroTime()
	type args struct {
		ctx           context.Context
		accountNumber string
		date          time.Time
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
				ctx:           context.TODO(),
				accountNumber: "",
				date:          date,
			},
			setupMocks: func(a args) {
				rows := sqlmock.
					NewRows([]string{
						"aadb.account_number",
						"aadb.entity_code",
						"aadb.category_code",
						"aadb.sub_category_code",
						"aadb.closing_balance",
						"aadb.closing_balance",
					}).
					AddRow("21100100000001", "001", "211", "21101", 0, 0)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetOneAccountBalanceDaily)).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "error case - QueryContext",
			args: args{
				ctx:           context.TODO(),
				accountNumber: "",
				date:          date,
			},
			setupMocks: func(a args) {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetOneAccountBalanceDaily)).
					WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - row scan",
			args: args{
				ctx:           context.TODO(),
				accountNumber: "",
				date:          date,
			},
			setupMocks: func(a args) {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetOneAccountBalanceDaily)).
					WillReturnRows(sqlmock.NewRows([]string{"InvalidColumn"}).AddRow(nil))
			},
			wantErr: true,
		},
		{
			name: "error case - row error",
			args: args{
				ctx:           context.TODO(),
				accountNumber: "",
				date:          date,
			},
			setupMocks: func(a args) {
				rows := sqlmock.
					NewRows([]string{"closing_balance"}).
					AddRow(0).RowError(0, assert.AnError)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetOneAccountBalanceDaily)).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(tt.args)

			_, err := suite.repo.GetOneAccountBalanceDaily(tt.args.ctx, tt.args.accountNumber, tt.args.date)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountingTestSuite) TestRepository_InsertAccountBalanceDaily() {
	type args struct {
		ctx context.Context
		req []models.AccountBalanceDaily
	}
	valueStrings := "(?, ?, ?, ?, ?, ?, ?, ?, ?)"
	query := fmt.Sprintf(queryInsertAccountBalanceDailyOld, valueStrings)

	testCases := []struct {
		name    string
		args    args
		wantErr bool
		doMock  func(args args)
	}{
		{
			name: "success case - insert account balance daily",
			args: args{
				ctx: context.TODO(),
				req: []models.AccountBalanceDaily{
					{},
				},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(query)).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "error case - database error",
			args: args{
				ctx: context.TODO(),
				req: []models.AccountBalanceDaily{
					{},
				},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(query)).WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
		// {
		// 	name: "error case - result error",
		// 	args: args{
		// 		ctx: context.TODO(),
		// 		req: []models.AccountBalanceDaily{},
		// 	},
		// 	doMock: func(args args) {
		// 		suite.mock.
		// 			ExpectExec(regexp.QuoteMeta(queryInsertAccountBalanceDaily)).WillReturnResult(sqlmock.NewErrorResult(assert.AnError))
		// 	},
		// 	wantErr: true,
		// },
		// {
		// 	name: "error case - no affected row",
		// 	args: args{
		// 		ctx: context.TODO(),
		// 		req: []models.AccountBalanceDaily{},
		// 	},
		// 	doMock: func(args args) {
		// 		suite.mock.
		// 			ExpectExec(regexp.QuoteMeta(queryInsertAccountBalanceDaily)).WillReturnResult(sqlmock.NewResult(0, 0))
		// 	},
		// 	wantErr: true,
		// },
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.doMock(tt.args)

			err := suite.repo.InsertAccountBalanceDaily(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
