package godbledger

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/darcys22/godbledger/godbledger/core"
	"github.com/darcys22/godbledger/godbledger/db/mysqldb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestGoDBLedgerTestSuite(t *testing.T) {
	t.Helper()
	suite.Run(t, new(goDBLedgerTestSuite))
}

type goDBLedgerTestSuite struct {
	suite.Suite
	t    *testing.T
	db   *sql.DB
	mock sqlmock.Sqlmock
	repo GoDBLedger
}

func (suite *goDBLedgerTestSuite) SetupTest() {
	var err error
	suite.db, suite.mock, err = sqlmock.New()
	require.NoError(suite.T(), err)
	suite.t = suite.T()
	db := mysqldb.Database{DB: suite.db}
	suite.repo = New(&db)
}

func (suite *goDBLedgerTestSuite) TearDownTest() {
	defer suite.db.Close()
}

var queryGetAccount = `SELECT * FROM accounts WHERE account_id = ? LIMIT 1`

func (suite *goDBLedgerTestSuite) TestGoDBLedger_InsertAccount() {
	queryInsertAccount := `INSERT INTO accounts(account_id, name) VALUES(?,?)`

	type args struct {
		ctx       context.Context
		accountID string
	}
	testCases := []struct {
		name       string
		args       args
		setupMocks func(a args)
		wantErr    bool
	}{
		{
			name: "success case - insert account",
			args: args{
				ctx:       context.TODO(),
				accountID: "211001000000048",
			},
			setupMocks: func(a args) {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetAccount)).
					WillReturnError(sql.ErrNoRows)
				suite.mock.ExpectBegin()
				prep := suite.mock.ExpectPrepare(regexp.QuoteMeta(queryInsertAccount))
				prep.ExpectExec().WillReturnResult(sqlmock.NewResult(0, 1))
				suite.mock.ExpectCommit()
			},
			wantErr: false,
		},
	}

	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(tt.args)

			err := suite.repo.InsertAccount(tt.args.ctx, tt.args.accountID)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *goDBLedgerTestSuite) TestGoDBLedger_GetAccount() {
	type args struct {
		ctx       context.Context
		accountId string
	}
	testCases := []struct {
		name       string
		args       args
		setupMocks func(a args)
		wantErr    bool
		expected   *core.Account
	}{
		{
			name: "success case - account is exist",
			args: args{
				ctx:       context.TODO(),
				accountId: "211001000000048",
			},
			setupMocks: func(a args) {
				rows := sqlmock.
					NewRows([]string{"account_id", "name"}).
					AddRow("211001000000048", "211001000000048")
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetAccount)).
					WillReturnRows(rows)
			},
			expected: &core.Account{
				Code: "211001000000048",
				Name: "211001000000048",
			},
			wantErr: false,
		},
		{
			name: "error case - account not found",
			args: args{
				ctx:       context.TODO(),
				accountId: "211001000000048",
			},
			setupMocks: func(a args) {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetAccount)).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(tt.args)

			resp, err := suite.repo.GetAccount(tt.args.ctx, tt.args.accountId)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, resp, tt.expected)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *goDBLedgerTestSuite) TestGoDBLedger_FindTransaction() {
	qeueryGetTransaction := `SELECT t.transaction_id,
	t.postdate,
	t.description,
	u.user_id,
	u.username
FROM   transactions AS t
	JOIN users AS u
		ON t.poster_user_id = u.user_id
WHERE  t.transaction_id = ?
LIMIT  1`

	type args struct {
		ctx   context.Context
		txnID string
	}
	testCases := []struct {
		name       string
		args       args
		setupMocks func(a args)
		wantErr    bool
		expected   *core.Account
	}{
		{
			name: "error case - transaction not found",
			args: args{
				ctx:   context.TODO(),
				txnID: "211001000000048",
			},
			setupMocks: func(a args) {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(qeueryGetTransaction)).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(tt.args)

			_, err := suite.repo.FindTransaction(tt.args.ctx, tt.args.txnID)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *goDBLedgerTestSuite) TestGoDBLedger_GetCurrency() {
	query := `SELECT * FROM currencies WHERE name = ? LIMIT 1`

	type args struct {
		ctx      context.Context
		currency string
	}
	testCases := []struct {
		name       string
		args       args
		setupMocks func(a args)
		wantErr    bool
		expected   Currency
	}{
		{
			name: "success case - get currency IDR",
			args: args{
				ctx:      context.TODO(),
				currency: "IDR",
			},
			setupMocks: func(a args) {
				rows := sqlmock.
					NewRows([]string{"name", "decimals"}).
					AddRow("IDR", 2)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(rows)
			},
			expected: CurrencyIDR,
			wantErr:  false,
		},
		{
			name: "success case - get currency USD",
			args: args{
				ctx:      context.TODO(),
				currency: "IDR",
			},
			setupMocks: func(a args) {
				rows := sqlmock.
					NewRows([]string{"name", "decimals"}).
					AddRow("USD", 2)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(rows)
			},
			expected: CurrencyUSD,
			wantErr:  false,
		},
		{
			name: "error case - currency not found",
			args: args{
				ctx:      context.TODO(),
				currency: "IDR",
			},
			setupMocks: func(a args) {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnError(sql.ErrTxDone)
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(tt.args)

			resp, err := suite.repo.GetCurrency(tt.args.ctx, tt.args.currency)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, resp, tt.expected)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
