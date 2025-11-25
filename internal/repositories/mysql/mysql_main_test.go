package mysql

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestMainRepositoryTestSuite(t *testing.T) {
	t.Helper()
	suite.Run(t, new(mainTestSuite))
}

type mainTestSuite struct {
	suite.Suite
	t    *testing.T
	db   *sql.DB
	mock sqlmock.Sqlmock
	repo SQLRepository
}

func (suite *mainTestSuite) SetupTest() {
	var err error

	suite.db, suite.mock, err = sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(suite.T(), err)

	suite.db.Ping()

	suite.t = suite.T()
	suite.repo = NewMySQLRepository(suite.db)
}

func (suite *mainTestSuite) TearDownTest() {
	defer suite.db.Close()
}

func (suite *mainTestSuite) TestRepository_Atomic() {
	type args struct {
		steps func(ctx context.Context, r SQLRepository) error
	}

	testCases := []struct {
		name       string
		args       args
		setupMocks func()
		wantErr    bool
	}{
		{
			name: "success case",
			args: args{
				steps: func(ctx context.Context, r SQLRepository) error {
					return nil
				},
			},
			setupMocks: func() {
				suite.mock.ExpectBegin()
				suite.mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "error case - begin transaction error",
			args: args{
				steps: func(ctx context.Context, r SQLRepository) error {
					return nil
				},
			},
			setupMocks: func() {
				suite.mock.ExpectBegin().WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - commit transaction error",
			args: args{
				steps: func(ctx context.Context, r SQLRepository) error {
					return nil
				},
			},
			setupMocks: func() {
				suite.mock.ExpectBegin()
				suite.mock.ExpectCommit().WillReturnError(assert.AnError)

			},
			wantErr: true,
		},
		{
			name: "error case - transaction rollback",
			args: args{
				steps: func(ctx context.Context, r SQLRepository) error {
					return sql.ErrTxDone
				},
			},
			setupMocks: func() {
				suite.mock.ExpectBegin()
				suite.mock.ExpectRollback()

			},
			wantErr: true,
		},
		{
			name: "error case - rollback transaction error",
			args: args{
				steps: func(ctx context.Context, r SQLRepository) error {
					return sql.ErrTxDone
				},
			},
			setupMocks: func() {
				suite.mock.ExpectBegin()
				suite.mock.ExpectRollback().WillReturnError(sql.ErrTxDone)

			},
			wantErr: true,
		},
		{
			name: "error case - panic transaction",
			args: args{
				steps: func(ctx context.Context, r SQLRepository) error {
					panic(sql.ErrTxDone)
				},
			},
			setupMocks: func() {
				suite.mock.ExpectBegin()
				suite.mock.ExpectRollback()
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := suite.repo.Atomic(context.TODO(), tt.args.steps)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *mainTestSuite) TestRepository_Ping() {
	testCases := []struct {
		name       string
		setupMocks func()
		wantErr    bool
	}{
		{
			name: "success case",
			setupMocks: func() {
				suite.mock.ExpectPing()
			},
			wantErr: false,
		},
		{
			name: "error case - ping error",
			setupMocks: func() {
				suite.mock.ExpectPing().WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			err := suite.repo.Ping(context.TODO())
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *mainTestSuite) TestRepository_GetAllCategorySubCategoryCOAType() {
	type args struct {
		ctx context.Context
	}
	testCases := []struct {
		name       string
		args       args
		setupMocks func(a args)
		wantErr    bool
	}{
		{
			name: "success case - get all category sub category coa type",
			args: args{
				ctx: context.TODO(),
			},
			setupMocks: func(a args) {
				rows := sqlmock.
					NewRows([]string{
						`category_code`,
						`category_name`,
						`sub_category_code`,
						`sub_category_name`,
						`coa_type_code`,
						`coa_type_name`,
					}).
					AddRow("111", "Cash Point", "11101", "Cash Point", "AST", "Asset")
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetCategorySubCategoryCOAType)).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "error case - QueryContext",
			args: args{
				ctx: context.TODO(),
			},
			setupMocks: func(a args) {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetCategorySubCategoryCOAType)).
					WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - row scan",
			args: args{
				ctx: context.TODO(),
			},
			setupMocks: func(a args) {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetCategorySubCategoryCOAType)).
					WillReturnRows(sqlmock.NewRows([]string{"InvalidColumn"}).AddRow(nil))
			},
			wantErr: true,
		},
		{
			name: "error case - row error",
			args: args{
				ctx: context.TODO(),
			},
			setupMocks: func(a args) {
				rows := sqlmock.
					NewRows([]string{
						`category_code`,
						`category_name`,
						`sub_category_code`,
						`sub_category_name`,
						`coa_type_code`,
						`coa_type_name`,
					}).AddRow("111", "Cash Point", "11101", "Cash Point", "AST", "Asset").
					RowError(0, assert.AnError)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetCategorySubCategoryCOAType)).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(tt.args)

			_, _, _, err := suite.repo.GetAllCategorySubCategoryCOAType(tt.args.ctx)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
