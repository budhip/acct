package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"testing"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestAccountingRepositoryTestSuite(t *testing.T) {
	t.Helper()
	suite.Run(t, new(accountingTestSuite))
}

type accountingTestSuite struct {
	suite.Suite
	t    *testing.T
	db   *sql.DB
	mock sqlmock.Sqlmock
	repo AccountingRepository
}

func (suite *accountingTestSuite) SetupTest() {
	var err error
	suite.db, suite.mock, err = sqlmock.New()
	require.NoError(suite.T(), err)
	suite.t = suite.T()
	suite.repo = NewMySQLRepository(suite.db).GetAccountingRepository()
}

func (suite *accountingTestSuite) TearDownTest() {
	defer suite.db.Close()
}

func (suite *accountingTestSuite) TestRepository_GetTrialBalance() {
	type args struct {
		ctx  context.Context
		opts models.TrialBalanceFilterOptions
	}
	defaultColumns := []string{
		"coa_id",
		"coa_type_name",
		"category_code",
		"category_name",
		"sub_category_code",
		"sub_category_name",
		"debit",
		"credit",
		"opening",
		"closing",
	}
	dec, _ := decimal.NewFromString("100000")
	defaultDecimal := decimal.NewNullDecimal(dec)
	date := atime.Now()

	testCases := []struct {
		name       string
		args       args
		setupMocks func(a args)
		wantErr    bool
		expected   []models.GetTrialBalanceOut
	}{
		{
			name: "success get trial balance",
			args: args{
				ctx: context.TODO(),
				opts: models.TrialBalanceFilterOptions{
					EntityCode: "000",
					StartDate:  date,
					EndDate:    date,
				},
			},
			setupMocks: func(a args) {
				listQuery, _, _ := buildGetTrialBalanceQuery(a.opts)
				rows := sqlmock.
					NewRows(defaultColumns).
					AddRow(
						"000",
						"asset",
						"A.111",
						"Kas Teller",
						"A.111.01",
						"Kas Teller Point",
						defaultDecimal,
						defaultDecimal,
						defaultDecimal,
						defaultDecimal,
					).
					AddRow(
						"000",
						"asset",
						"A.111",
						"Kas Teller",
						"A.111.01",
						"Kas Teller Point",
						decimal.NewFromFloat(0),
						decimal.NewFromFloat(0),
						decimal.NewFromFloat(0),
						decimal.NewFromFloat(0),
					)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(listQuery)).
					WillReturnRows(rows)
			},
			expected: []models.GetTrialBalanceOut{
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
				listQuery, _, _ := buildGetTrialBalanceQuery(a.opts)
				rows := sqlmock.
					NewRows(defaultColumns).
					AddRow(
						"000",
						"asset",
						"A.111",
						"Kas Teller",
						"A.111.01",
						"Kas Teller Point",
						defaultDecimal,
						defaultDecimal,
						defaultDecimal,
						defaultDecimal,
					).RowError(0, assert.AnError)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(listQuery)).
					WillReturnRows(rows)
			},
			expected: []models.GetTrialBalanceOut{},
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
				listQuery, _, _ := buildGetTrialBalanceQuery(a.opts)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(listQuery)).
					WillReturnRows(sqlmock.NewRows([]string{"InvalidColumn"}).AddRow(nil))
			},
			expected: []models.GetTrialBalanceOut{},
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
				listQuery, _, _ := buildGetTrialBalanceQuery(a.opts)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(listQuery)).
					WillReturnError(assert.AnError)
			},
			expected: []models.GetTrialBalanceOut{},
			wantErr:  true,
		},
	}

	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(tt.args)

			_, _, err := suite.repo.GetTrialBalance(tt.args.ctx, tt.args.opts)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountingTestSuite) TestRepository_InsertJournalDetail() {
	type args struct {
		ctx context.Context
		req []models.CreateJournalDetail
	}
	valueStrings := "(?, ?, ?, CURRENT_TIMESTAMP(6), CURRENT_TIMESTAMP(6))"
	queryInsertJournalDetail = fmt.Sprintf(queryInsertJournalDetail, valueStrings)

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
				req: []models.CreateJournalDetail{
					{},
					{},
				},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryInsertJournalDetail)).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "error no row",
			args: args{
				ctx: context.TODO(),
				req: []models.CreateJournalDetail{},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryInsertJournalDetail)).WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
		{
			name: "error result",
			args: args{
				ctx: context.TODO(),
				req: []models.CreateJournalDetail{},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryInsertJournalDetail)).WillReturnResult(sqlmock.NewErrorResult(assert.AnError))
			},
			wantErr: true,
		},
		{
			name: "error db",
			args: args{
				ctx: context.TODO(),
				req: []models.CreateJournalDetail{},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryInsertJournalDetail)).WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.doMock(tt.args)

			err := suite.repo.InsertJournalDetail(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountingTestSuite) TestRepository_GetSubLedger() {
	date := atime.Now()

	type args struct {
		ctx  context.Context
		opts models.SubLedgerFilterOptions
	}
	testCases := []struct {
		name       string
		args       args
		setupMocks func(a args)
		wantErr    bool
		expected   []models.GetSubLedgerOut
	}{
		{
			name: "success get sub ledger",
			args: args{
				ctx: context.TODO(),
				opts: models.SubLedgerFilterOptions{
					AccountNumber: "22201000000008",
					Email:         "ono@amartha.com",
					StartDate:     date,
					EndDate:       date,
					Limit:         1,
				},
			},
			setupMocks: func(a args) {
				query, _, _ := buildSubLedgerQuery(a.opts)
				rows := sqlmock.
					NewRows([]string{
						`t.transaction_id`,
						`COALESCE(ajd.reference_number, '')`,
						`ajd.transaction_date`,
						`ajd.order_type`,
						`ajd.transaction_type`,
						`s.description as narrative`,
						`COALESCE(ajd.metadata, '{}') metadata`,
						`case when ajd.is_debit = TRUE then s.amount else 0 end as debit`,
						`case when ajd.is_debit = FALSE then s.amount else 0 end as credit`,
						`ajd.created_at`,
						`ajd.updated_at`,
					}).
					AddRow("", "", time.Time{}, "", "", "", nil, 0, 0, time.Time{}, time.Time{})
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(rows)
			},
			expected: []models.GetSubLedgerOut{{
				TransactionID:   "",
				TransactionDate: time.Time{},
				TransactionType: "",
				Narrative:       "",
				Metadata:        &models.Metadata{},
				Debit:           decimal.New(0, 0),
				Credit:          decimal.New(0, 0),
				CreatedAt:       time.Time{},
				UpdatedAt:       time.Time{},
			},
			},
			wantErr: false,
		},
		{
			name: "error case - row error",
			args: args{
				ctx: context.TODO(),
				opts: models.SubLedgerFilterOptions{
					AccountNumber: "22201000000008",
					Email:         "ono@amartha.com",
					StartDate:     date,
					EndDate:       date,
					Limit:         1,
				},
			},
			setupMocks: func(a args) {
				query, _, _ := buildSubLedgerQuery(a.opts)
				rows := sqlmock.
					NewRows([]string{
						`t.transaction_id`,
						`COALESCE(ajd.reference_number, '')`,
						`ajd.transaction_date`,
						`ajd.order_type`,
						`ajd.transaction_type`,
						`s.description as narrative`,
						`COALESCE(ajd.metadata, '{}') metadata`,
						`case when ajd.is_debit = TRUE then s.amount else 0 end as debit`,
						`case when ajd.is_debit = FALSE then s.amount else 0 end as credit`,
						`ajd.created_at`,
						`ajd.updated_at`,
					}).
					AddRow("", "", time.Time{}, "", "", "", nil, 0, 0, time.Time{}, time.Time{}).RowError(0, assert.AnError)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
		{
			name: "error case - scan row",
			args: args{
				ctx: context.TODO(),
				opts: models.SubLedgerFilterOptions{
					AccountNumber: "22201000000008",
					Email:         "tono@amartha.com",
					StartDate:     date,
					EndDate:       date,
					Limit:         1,
				},
			},
			setupMocks: func(a args) {
				query, _, _ := buildSubLedgerQuery(a.opts)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(sqlmock.NewRows([]string{"InvalidColumn"}).AddRow(nil))
			},
			wantErr: true,
		},
		{
			name: "error case - QueryContext",
			args: args{
				ctx: context.TODO(),
				opts: models.SubLedgerFilterOptions{
					AccountNumber: "22201000000008",
					Email:         "ono@amartha.com",
					StartDate:     date,
					EndDate:       date,
					Limit:         1,
				},
			},
			setupMocks: func(a args) {
				query, _, _ := buildSubLedgerQuery(a.opts)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnError(assert.AnError)
			},
			expected: []models.GetSubLedgerOut{},
			wantErr:  true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(tt.args)

			_, err := suite.repo.GetSubLedger(tt.args.ctx, tt.args.opts)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountingTestSuite) TestRepository_GetSubLedgerStream() {
	date := atime.Now()

	type args struct {
		ctx  context.Context
		opts models.SubLedgerFilterOptions
	}
	testCases := []struct {
		name       string
		args       args
		setupMocks func(a args)
		wantErr    bool
	}{
		{
			name: "success get sub ledger",
			args: args{
				ctx: context.TODO(),
				opts: models.SubLedgerFilterOptions{
					AccountNumber: "22201000000008",
					Email:         "ono@amartha.com",
					StartDate:     date,
					EndDate:       date,
				},
			},
			setupMocks: func(a args) {
				query, _, _ := buildSubLedgerQuery(a.opts)
				rows := sqlmock.
					NewRows([]string{
						`t.transaction_id`,
						`COALESCE(ajd.reference_number, '')`,
						`ajd.transaction_date`,
						`ajd.order_type`,
						`ajd.transaction_type`,
						`s.description as narrative`,
						`case when ajd.is_debit = TRUE then s.amount else 0 end as debit`,
						`case when ajd.is_debit = FALSE then s.amount else 0 end as credit`,
						`ajd.created_at`,
						`ajd.updated_at`,
					}).
					AddRow("", "", time.Time{}, "", "", "", 0, 0, time.Time{}, time.Time{})

				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(rows).WillReturnRows(&sqlmock.Rows{})
			},
			wantErr: false,
		},
		{
			name: "error case - row error",
			args: args{
				ctx: context.TODO(),
				opts: models.SubLedgerFilterOptions{
					AccountNumber: "22201000000008",
					Email:         "ono@amartha.com",
					StartDate:     date,
					EndDate:       date,
				},
			},
			setupMocks: func(a args) {
				query, _, _ := buildSubLedgerQuery(a.opts)
				rows := sqlmock.
					NewRows([]string{
						`t.transaction_id`,
						`COALESCE(ajd.reference_number, '')`,
						`ajd.transaction_date`,
						`ajd.order_type`,
						`ajd.transaction_type`,
						`s.description as narrative`,
						`case when ajd.is_debit = TRUE then s.amount else 0 end as debit`,
						`case when ajd.is_debit = FALSE then s.amount else 0 end as credit`,
						`ajd.created_at`,
						`ajd.updated_at`,
					}).
					AddRow("", "", time.Time{}, "", "", "", 0, 0, time.Time{}, time.Time{}).RowError(0, assert.AnError)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
		{
			name: "error case - scan row",
			args: args{
				ctx: context.TODO(),
				opts: models.SubLedgerFilterOptions{
					AccountNumber: "22201000000008",
					Email:         "tono@amartha.com",
					StartDate:     date,
					EndDate:       date,
				},
			},
			setupMocks: func(a args) {
				query, _, _ := buildSubLedgerQuery(a.opts)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(sqlmock.NewRows([]string{"InvalidColumn"}).AddRow(nil))
			},
			wantErr: true,
		},
		{
			name: "error case - QueryContext",
			args: args{
				ctx: context.TODO(),
				opts: models.SubLedgerFilterOptions{
					AccountNumber: "22201000000008",
					Email:         "ono@amartha.com",
					StartDate:     date,
					EndDate:       date,
				},
			},
			setupMocks: func(a args) {
				query, _, _ := buildSubLedgerQuery(a.opts)
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

			chanData := suite.repo.GetSubLedgerStream(tt.args.ctx, tt.args.opts)
			for result := range chanData {
				assert.Equal(t, tt.wantErr, result.Err != nil)
			}

			if err := suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountingTestSuite) TestRepository_GetSubLedgerCount() {
	date := atime.Now()
	type args struct {
		ctx  context.Context
		opts models.SubLedgerFilterOptions
	}
	testCases := []struct {
		name       string
		args       args
		setupMocks func(a args)
		wantErr    bool
	}{
		{
			name: "success case -  get sub ledger count",
			args: args{
				ctx: context.TODO(),
				opts: models.SubLedgerFilterOptions{
					AccountNumber: "22201000000008",
					Email:         "ono@amartha.com",
					StartDate:     date,
					EndDate:       date,
					Limit:         1,
				},
			},
			setupMocks: func(a args) {
				query, _, _ := buildCountSubLedgerQuery(a.opts)
				rows := sqlmock.
					NewRows([]string{"count"}).
					AddRow(1)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "error case - QueryRowContext",
			args: args{
				ctx: context.TODO(),
				opts: models.SubLedgerFilterOptions{
					AccountNumber: "22201000000008",
					Email:         "ono@amartha.com",
					StartDate:     date,
					EndDate:       date,
					Limit:         1,
				},
			},
			setupMocks: func(a args) {
				query, _, _ := buildCountSubLedgerQuery(a.opts)
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

			_, err := suite.repo.GetSubLedgerCount(tt.args.ctx, tt.args.opts)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountingTestSuite) TestRepository_GetJournalDetailByTransactionId() {
	defaultDecimal := decimal.NewNullDecimal(decimal.NewFromFloat(100000))
	type args struct {
		ctx context.Context
		req string
	}
	testCases := []struct {
		name       string
		args       args
		setupMocks func(a args)
		wantErr    bool
		expected   []models.GetJournalDetailOut
	}{
		{
			name: "success get sub ledger",
			args: args{
				ctx: context.TODO(),
				req: "12c5692e-cfbd-4cee-a4ce-86eac1447d48",
			},
			setupMocks: func(a args) {
				query, _, _ := getJournalDetailQuery(a.req)
				rows := sqlmock.
					NewRows([]string{
						`t.transaction_id`,
						`ajd.journal_id`,
						`aa.account_number`,
						`COALESCE(aa.name, '') account_name`,
						`COALESCE(aa.alt_id, '') alt_id`,
						`aa.entity_code`,
						`aa.entity_name`,
						`aa.sub_category_code`,
						`aa.sub_category_name`,
						`ajd.transaction_type`,
						`s.amount`,
						`ajd.transaction_date`,
						`COALESCE(s.description, '') AS narrative`,
						`ajd.is_debit`,
					}).
					AddRow("", "", "", "", "", "", "", "", "", "", defaultDecimal, time.Time{}, "", true)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "error case - row error",
			args: args{
				ctx: context.TODO(),
				req: "12c5692e-cfbd-4cee-a4ce-86eac1447d48",
			},
			setupMocks: func(a args) {
				query, _, _ := getJournalDetailQuery(a.req)
				rows := sqlmock.
					NewRows([]string{
						`t.transaction_id`,
						`ajd.journal_id`,
						`aa.account_number`,
						`COALESCE(aa.name, '') account_name`,
						`COALESCE(aa.alt_id, '') alt_id`,
						`aa.entity_code`,
						`aa.entity_name`,
						`aa.sub_category_code`,
						`aa.sub_category_name`,
						`ajd.transaction_type`,
						`s.amount`,
						`ajd.transaction_date`,
						`COALESCE(s.description, '') AS narrative`,
						`ajd.is_debit`,
					}).
					AddRow("", "", "", "", "", "", "", "", "", "", defaultDecimal, time.Time{}, "", true).RowError(0, assert.AnError)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
		{
			name: "error case - scan row",
			args: args{
				ctx: context.TODO(),
				req: "12c5692e-cfbd-4cee-a4ce-86eac1447d48",
			},
			setupMocks: func(a args) {
				query, _, _ := getJournalDetailQuery(a.req)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(sqlmock.NewRows([]string{"InvalidColumn"}).AddRow(nil))
			},
			wantErr: true,
		},
		{
			name: "error case - QueryContext",
			args: args{
				ctx: context.TODO(),
				req: "12c5692e-cfbd-4cee-a4ce-86eac1447d48",
			},
			setupMocks: func(a args) {
				query, _, _ := getJournalDetailQuery(a.req)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnError(assert.AnError)
			},
			expected: []models.GetJournalDetailOut{},
			wantErr:  true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(tt.args)

			_, err := suite.repo.GetJournalDetailByTransactionId(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountingTestSuite) TestRepository_GetSubLedgerAccounts() {
	date := atime.Now()
	filter := models.SubLedgerAccountsFilterOptions{
		EntityCode:      "001",
		Search:          "121001000000009",
		SearchBy:        "account_number",
		StartDate:       date,
		EndDate:         date,
		AfterCreatedAt:  &date,
		BeforeCreatedAt: &date,
		Limit:           1,
	}
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
	type args struct {
		ctx  context.Context
		opts models.SubLedgerAccountsFilterOptions
	}
	testCases := []struct {
		name       string
		args       args
		setupMocks func(a args)
		wantErr    bool
		expected   []models.GetSubLedgerAccountsOut
	}{
		{
			name: "success case - get sub ledger accounts",
			args: args{
				ctx:  context.TODO(),
				opts: filter,
			},
			setupMocks: func(a args) {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetCategorySubCategoryCOAType)).
					WillReturnRows(rowsCategorySubCategoryCOAType)
				query, _, _ := buildSubLedgerAccountsQuery(a.opts)
				rows := sqlmock.
					NewRows([]string{
						`aa.account_number`,
						`aa.name as account_name`,
						`coalesce(aa.alt_id, "") alt_id`,
						`aa.sub_category_code`,
						`0 as number_of_data`,
						`aa.created_at`,
					}).
					AddRow("121001000000009", "Cash in Transit - Disburse Modal", "", "12101", 0, time.Time{})
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(rows)
			},
			expected: []models.GetSubLedgerAccountsOut{
				{
					AccountNumber:   "121001000000009",
					AccountName:     "Cash in Transit - Disburse Modal",
					AltId:           "",
					SubCategoryCode: "12101",
					SubCategoryName: "Cash in Transit",
					TotalRowData:    0,
					CreatedAt:       time.Time{},
				},
			},
			wantErr: false,
		},
		{
			name: "error case - row error",
			args: args{
				ctx:  context.TODO(),
				opts: filter,
			},
			setupMocks: func(a args) {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetCategorySubCategoryCOAType)).
					WillReturnRows(rowsCategorySubCategoryCOAType)
				query, _, _ := buildSubLedgerAccountsQuery(a.opts)
				rows := sqlmock.
					NewRows([]string{
						`aa.account_number`,
						`aa.name as account_name`,
						`coalesce(aa.alt_id, "") alt_id`,
						`aa.sub_category_code`,
						`0 as number_of_data`,
						`aa.created_at`,
					}).
					AddRow("121001000000009", "Cash in Transit - Disburse Modal", "", "12101", 0, time.Time{}).RowError(0, assert.AnError)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
		{
			name: "error case - scan row",
			args: args{
				ctx:  context.TODO(),
				opts: filter,
			},
			setupMocks: func(a args) {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetCategorySubCategoryCOAType)).
					WillReturnRows(rowsCategorySubCategoryCOAType)
				query, _, _ := buildSubLedgerAccountsQuery(a.opts)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(sqlmock.NewRows([]string{"InvalidColumn"}).AddRow(nil))
			},
			wantErr: true,
		},
		{
			name: "error case - QueryContext",
			args: args{
				ctx:  context.TODO(),
				opts: filter,
			},
			setupMocks: func(a args) {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetCategorySubCategoryCOAType)).
					WillReturnRows(rowsCategorySubCategoryCOAType)
				query, _, _ := buildSubLedgerAccountsQuery(a.opts)
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

			_, err := suite.repo.GetSubLedgerAccounts(tt.args.ctx, tt.args.opts)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountingTestSuite) TestRepository_GetSubLedgerAccountsCount() {
	filter := models.SubLedgerAccountsFilterOptions{
		EntityCode: "001",
		Limit:      1,
	}
	type args struct {
		ctx  context.Context
		opts models.SubLedgerAccountsFilterOptions
	}
	testCases := []struct {
		name       string
		args       args
		setupMocks func(a args)
		wantErr    bool
	}{
		{
			name: "success case -  get sub ledger accounts count",
			args: args{
				ctx:  context.TODO(),
				opts: filter,
			},
			setupMocks: func(a args) {
				query, _, _ := buildCountSubLedgerAccountsQuery(a.opts)
				rows := sqlmock.
					NewRows([]string{"count"}).
					AddRow(1)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "error case - QueryRowContext",
			args: args{
				ctx:  context.TODO(),
				opts: filter,
			},
			setupMocks: func(a args) {
				query, _, _ := buildCountSubLedgerAccountsQuery(a.opts)
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

			_, err := suite.repo.GetSubLedgerAccountsCount(tt.args.ctx, tt.args.opts)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountingTestSuite) TestRepository_GetSubLedgerAccountTotalTransaction() {
	date := atime.Now()
	filter := models.SubLedgerAccountsFilterOptions{
		EntityCode: "001",
		Search:     "121001000000009",
		SearchBy:   "account_number",
		StartDate:  date,
		EndDate:    date,
		Limit:      1,
	}
	type args struct {
		ctx  context.Context
		opts models.SubLedgerAccountsFilterOptions
	}
	testCases := []struct {
		name       string
		args       args
		setupMocks func(a args)
		wantErr    bool
		expected   int
	}{
		{
			name: "success case - get sub ledger accounts",
			args: args{
				ctx:  context.TODO(),
				opts: filter,
			},
			setupMocks: func(a args) {
				query, _, _ := buildSubLedgerAccountTotalTransactionQuery(a.opts)
				rows := sqlmock.
					NewRows([]string{`number_of_data`}).AddRow(10)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(rows)
			},
			expected: 10,
			wantErr:  false,
		},
		{
			name: "error case - row error",
			args: args{
				ctx:  context.TODO(),
				opts: filter,
			},
			setupMocks: func(a args) {
				query, _, _ := buildSubLedgerAccountTotalTransactionQuery(a.opts)
				rows := sqlmock.
					NewRows([]string{`number_of_data`}).AddRow(10).RowError(0, assert.AnError)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
		{
			name: "error case - scan row",
			args: args{
				ctx:  context.TODO(),
				opts: filter,
			},
			setupMocks: func(a args) {
				query, _, _ := buildSubLedgerAccountTotalTransactionQuery(a.opts)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(sqlmock.NewRows([]string{"InvalidColumn"}).AddRow(nil))
			},
			wantErr: true,
		},
		{
			name: "error case - QueryContext",
			args: args{
				ctx:  context.TODO(),
				opts: filter,
			},
			setupMocks: func(a args) {
				query, _, _ := buildSubLedgerAccountTotalTransactionQuery(a.opts)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - QueryContext - no rows",
			args: args{
				ctx:  context.TODO(),
				opts: filter,
			},
			setupMocks: func(a args) {
				query, _, _ := buildSubLedgerAccountTotalTransactionQuery(a.opts)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnError(models.ErrNoRows)
			},
			wantErr: false,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(tt.args)

			total, err := suite.repo.GetSubLedgerAccountTotalTransaction(tt.args.ctx, tt.args.opts)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.expected, total)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountingTestSuite) TestRepository_GetAccountBalancePeriodStart() {
	type args struct {
		ctx  context.Context
		opts models.SubLedgerFilterOptions
	}

	date := atime.Now()
	testCases := []struct {
		name       string
		args       args
		setupMocks func(a args)
		wantErr    bool
	}{
		{
			name: "success case - get account balance period start",
			args: args{
				ctx: context.TODO(),
				opts: models.SubLedgerFilterOptions{
					AccountNumber: "22201000000008",
					Email:         "tono@amartha.com",
					StartDate:     date,
					EndDate:       date,
					Limit:         1,
				},
			},
			setupMocks: func(a args) {
				query, _, _ := getAccountBalancePeriodStart(a.opts.AccountNumber, a.opts.StartDate)
				rows := sqlmock.
					NewRows([]string{"balance_period_start"}).
					AddRow(1)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "error case - row error",
			args: args{
				ctx: context.TODO(),
				opts: models.SubLedgerFilterOptions{
					AccountNumber: "22201000000008",
					Email:         "tono@amartha.com",
					StartDate:     date,
					EndDate:       date,
					Limit:         1,
				},
			},
			setupMocks: func(a args) {
				query, _, _ := getAccountBalancePeriodStart(a.opts.AccountNumber, a.opts.StartDate)
				rows := sqlmock.
					NewRows([]string{"balance_period_start"}).
					AddRow(1).RowError(0, assert.AnError)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
		{
			name: "error case - scan row",
			args: args{
				ctx: context.TODO(),
				opts: models.SubLedgerFilterOptions{
					AccountNumber: "22201000000008",
					Email:         "tono@amartha.com",
					StartDate:     date,
					EndDate:       date,
					Limit:         1,
				},
			},
			setupMocks: func(a args) {
				query, _, _ := getAccountBalancePeriodStart(a.opts.AccountNumber, a.opts.StartDate)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(sqlmock.NewRows([]string{"InvalidColumn"}).AddRow(nil))
			},
			wantErr: true,
		},
		{
			name: "error case - QueryContext",
			args: args{
				ctx: context.TODO(),
				opts: models.SubLedgerFilterOptions{
					AccountNumber: "22201000000008",
					Email:         "ono@amartha.com",
					StartDate:     date,
					EndDate:       date,
					Limit:         1,
				},
			},
			setupMocks: func(a args) {
				query, _, _ := getAccountBalancePeriodStart(a.opts.AccountNumber, a.opts.StartDate)
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

			_, err := suite.repo.GetAccountBalancePeriodStart(tt.args.ctx, tt.args.opts.AccountNumber, tt.args.opts.StartDate)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountingTestSuite) TestRepository_GetTrialBalanceDetails() {
	type args struct {
		ctx  context.Context
		opts models.TrialBalanceDetailsFilterOptions
	}
	defaultColumns := []string{
		"account_number",
		"name",
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
			name: "success get trial balance details",
			args: args{
				ctx: context.TODO(),
				opts: models.TrialBalanceDetailsFilterOptions{
					SubCategoryCode: "21108",
					StartDate:       date,
					EndDate:         date,
				},
			},
			setupMocks: func(a args) {
				listQuery, _, _ := buildGetListTrialBalanceDetailQuery(a.opts)
				rows := sqlmock.
					NewRows(defaultColumns).
					AddRow(
						"000",
						"Shinji Takeru",
						defaultDecimal,
						defaultDecimal,
						defaultDecimal,
						defaultDecimal,
					).
					AddRow(
						"001",
						"Shinji Takeru 1",
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
			name: "error row",
			args: args{
				ctx: context.TODO(),
				opts: models.TrialBalanceDetailsFilterOptions{
					SubCategoryCode: "21108",
					StartDate:       date,
					EndDate:         date,
				},
			},
			setupMocks: func(a args) {
				listQuery, _, _ := buildGetListTrialBalanceDetailQuery(a.opts)
				rows := sqlmock.
					NewRows(defaultColumns).
					AddRow(
						"123456",
						"Shinji Takeru",
						defaultDecimal,
						defaultDecimal,
						defaultDecimal,
						defaultDecimal,
					).RowError(0, assert.AnError)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(listQuery)).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
		{
			name: "error scan row",
			args: args{
				ctx: context.TODO(),
				opts: models.TrialBalanceDetailsFilterOptions{
					SubCategoryCode: "21108",
					StartDate:       date,
					EndDate:         date,
				},
			},
			setupMocks: func(a args) {
				listQuery, _, _ := buildGetListTrialBalanceDetailQuery(a.opts)
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
				opts: models.TrialBalanceDetailsFilterOptions{
					SubCategoryCode: "21108",
					StartDate:       date,
					EndDate:         date,
				},
			},
			setupMocks: func(a args) {
				listQuery, _, _ := buildGetListTrialBalanceDetailQuery(a.opts)
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

			_, err := suite.repo.GetTrialBalanceDetails(tt.args.ctx, tt.args.opts)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountingTestSuite) TestRepository_InsertTransaction() {
	valueStrings := "(?, ?, ?)"
	queryInsertTransaction = fmt.Sprintf(queryInsertTransaction, valueStrings)

	testCases := []struct {
		name    string
		wantErr bool
		doMock  func()
	}{
		{
			name: "success",
			doMock: func() {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryInsertTransaction)).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "error no row",
			doMock: func() {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryInsertTransaction)).WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
		{
			name: "error result",
			doMock: func() {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryInsertTransaction)).WillReturnResult(sqlmock.NewErrorResult(assert.AnError))
			},
			wantErr: true,
		},
		{
			name: "error db",
			doMock: func() {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryInsertTransaction)).WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.doMock()

			err := suite.repo.InsertTransaction(context.Background(), []models.CreateTransaction{})
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountingTestSuite) TestRepository_ToggleForeignKeyChecks() {
	defaultQuery := "SET FOREIGN_KEY_CHECKS = 0;"

	testCases := []struct {
		isEnable bool
		name     string
		wantErr  bool
		doMock   func()
	}{
		{
			name:     "success enable",
			isEnable: true,
			doMock: func() {
				suite.mock.
					ExpectExec(regexp.QuoteMeta("SET FOREIGN_KEY_CHECKS = 1;")).WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: false,
		},
		{
			name: "error db",
			doMock: func() {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(defaultQuery)).WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.doMock()

			err := suite.repo.ToggleForeignKeyChecks(context.Background(), tt.isEnable)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountingTestSuite) TestRepository_InsertSplit() {
	valueStrings := "(?, ?, ?, ?, ?, ?)"
	queryInsertSplit = fmt.Sprintf(queryInsertSplit, valueStrings)

	testCases := []struct {
		name    string
		wantErr bool
		doMock  func()
	}{
		{
			name: "success",
			doMock: func() {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryInsertSplit)).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "error no row",
			doMock: func() {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryInsertSplit)).WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
		{
			name: "error result",
			doMock: func() {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryInsertSplit)).WillReturnResult(sqlmock.NewErrorResult(assert.AnError))
			},
			wantErr: true,
		},
		{
			name: "error db",
			doMock: func() {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryInsertSplit)).WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.doMock()

			err := suite.repo.InsertSplit(context.Background(), []models.CreateSplit{})
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountingTestSuite) TestRepository_InsertSplitAccount() {
	valueStrings := "(?, ?)"
	queryInsertSplitAccount = fmt.Sprintf(queryInsertSplitAccount, valueStrings)

	testCases := []struct {
		name    string
		wantErr bool
		doMock  func()
	}{
		{
			name: "success",
			doMock: func() {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryInsertSplitAccount)).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "error no row",
			doMock: func() {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryInsertSplitAccount)).WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
		{
			name: "error result",
			doMock: func() {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryInsertSplitAccount)).WillReturnResult(sqlmock.NewErrorResult(assert.AnError))
			},
			wantErr: true,
		},
		{
			name: "error db",
			doMock: func() {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryInsertSplitAccount)).WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.doMock()

			err := suite.repo.InsertSplitAccount(context.Background(), []models.CreateSplitAccount{})
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountingTestSuite) TestRepository_GetOneSplitAccount() {
	type args struct {
		ctx           context.Context
		accountNumber string
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
			},
			setupMocks: func(a args) {
				rows := sqlmock.
					NewRows([]string{"is_exist"}).
					AddRow(true)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetOneSplitAccount)).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "error case - QueryContext",
			args: args{
				ctx:           context.TODO(),
				accountNumber: "",
			},
			setupMocks: func(a args) {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetOneSplitAccount)).
					WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(tt.args)

			_, err := suite.repo.GetOneSplitAccount(tt.args.ctx, tt.args.accountNumber)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountingTestSuite) TestRepository_CheckTransactionIdIsExist() {
	type args struct {
		ctx           context.Context
		transactionId string
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
				transactionId: "",
			},
			setupMocks: func(a args) {
				rows := sqlmock.
					NewRows([]string{"is_exist"}).
					AddRow(true)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryCheckTransactionIdIsExist)).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "error case - QueryContext",
			args: args{
				ctx:           context.TODO(),
				transactionId: "",
			},
			setupMocks: func(a args) {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryCheckTransactionIdIsExist)).
					WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(tt.args)

			_, err := suite.repo.CheckTransactionIdIsExist(tt.args.ctx, tt.args.transactionId)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
