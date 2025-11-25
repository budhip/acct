package mysql

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestCOATypeRepositoryTestSuite(t *testing.T) {
	t.Helper()
	suite.Run(t, new(coaTypeTestSuite))
}

type coaTypeTestSuite struct {
	suite.Suite
	t    *testing.T
	db   *sql.DB
	mock sqlmock.Sqlmock
	repo COATypeRepository
}

func (suite *coaTypeTestSuite) SetupTest() {
	var err error

	suite.db, suite.mock, err = sqlmock.New()
	require.NoError(suite.T(), err)

	suite.t = suite.T()
	suite.repo = NewMySQLRepository(suite.db).GetCOATypeRepository()
}

func (suite *coaTypeTestSuite) TearDownTest() {
	defer suite.db.Close()
}

func (suite *coaTypeTestSuite) TestRepository_Create() {
	type args struct {
		ctx context.Context
		in  *models.CreateCOATypeIn
	}

	testCases := []struct {
		name    string
		args    args
		wantErr bool
		doMock  func(args args)
	}{
		{
			name: "test success",
			args: args{

				ctx: context.Background(),
				in: &models.CreateCOATypeIn{
					Code:          "test",
					Name:          "test",
					NormalBalance: "test",
					Status:        "test",
				},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryCOATypeCreate)).
					WithArgs(args.in.Code, args.in.Name, args.in.NormalBalance, args.in.Status).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "test error no row",
			args: args{
				ctx: context.Background(),
				in: &models.CreateCOATypeIn{
					Code:          "test",
					Name:          "test",
					NormalBalance: "test",
					Status:        "test",
				},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryCOATypeCreate)).
					WithArgs(args.in.Code, args.in.Name, args.in.NormalBalance, args.in.Status).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
		{
			name: "test error result",
			args: args{
				ctx: context.Background(),
				in: &models.CreateCOATypeIn{
					Code:          "test",
					Name:          "test",
					NormalBalance: "test",
					Status:        "test",
				},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryCOATypeCreate)).
					WithArgs(args.in.Code, args.in.Name, args.in.NormalBalance, args.in.Status).
					WillReturnResult(sqlmock.NewErrorResult(assert.AnError))
			},
			wantErr: true,
		},
		{
			name: "test error db",
			args: args{
				ctx: context.Background(),
				in: &models.CreateCOATypeIn{
					Code:          "test",
					Name:          "test",
					NormalBalance: "test",
					Status:        "test",
				},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryCOATypeCreate)).
					WithArgs(args.in.Code, args.in.Name, args.in.NormalBalance, args.in.Status).
					WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		suite.t.Run(tc.name, func(t *testing.T) {
			tc.doMock(tc.args)

			err := suite.repo.Create(tc.args.ctx, tc.args.in)
			assert.Equal(t, tc.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *coaTypeTestSuite) TestRepository_CheckCOATypeByCode() {
	type args struct {
		ctx        context.Context
		code       string
		setupMocks func()
	}

	testCases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test success",
			args: args{
				ctx:  context.Background(),
				code: "211",
				setupMocks: func() {
					rows := sqlmock.NewRows(
						[]string{"code"}).AddRow("211")
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryCOATypeIsExistByCode)).WillReturnRows(rows)
				},
			},
			wantErr: false,
		},
		{
			name: "test data not found",
			args: args{
				ctx: context.TODO(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryCOATypeIsExistByCode)).WillReturnError(sql.ErrNoRows)
				},
			},
			wantErr: true,
		},
		{
			name: "test error result",
			args: args{
				ctx: context.TODO(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryCOATypeIsExistByCode)).WillReturnError(assert.AnError)
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.args.setupMocks()

			err := suite.repo.CheckCOATypeByCode(tt.args.ctx, tt.args.code)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *coaTypeTestSuite) TestRepository_GetCOATypeByCode() {
	type args struct {
		ctx        context.Context
		code       string
		setupMocks func()
	}

	testCases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test success",
			args: args{
				ctx:  context.Background(),
				code: "211",
				setupMocks: func() {
					rows := sqlmock.NewRows(
						[]string{"code", "coa_type_name", "normal_balance", "status", "created_at", "updated_at"}).
						AddRow("211", "test", "test", "active", time.Now(), time.Now())
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryGetCOATypeByCode)).WillReturnRows(rows)
				},
			},
			wantErr: false,
		},
		{
			name: "test data not found",
			args: args{
				ctx: context.TODO(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryGetCOATypeByCode)).WillReturnError(sql.ErrNoRows)
				},
			},
			wantErr: false,
		},
		{
			name: "test error result",
			args: args{
				ctx: context.TODO(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryGetCOATypeByCode)).WillReturnError(assert.AnError)
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.args.setupMocks()

			_, err := suite.repo.GetCOATypeByCode(tt.args.ctx, tt.args.code)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *coaTypeTestSuite) TestRepository_GetAll() {
	testCases := []struct {
		name    string
		wantErr bool
		doMock  func()
	}{
		{
			name: "success",
			doMock: func() {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetAllCOAType)).
					WillReturnRows(
						sqlmock.
							NewRows([]string{"code", "coa_type_name", "normal_balance", "status", "created_at", "updated_at"}).
							AddRow("code", "coa_type_name", "normal_balance", "status", time.Now(), time.Now()),
					)
			},
			wantErr: false,
		},
		{
			name: "error row",
			doMock: func() {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetAllCOAType)).
					WillReturnRows(
						sqlmock.
							NewRows([]string{"code", "coa_type_name", "normal_balance", "status", "created_at", "updated_at"}).
							AddRow("code", "coa_type_name", "normal_balance", "status", time.Now(), time.Now()).RowError(0, assert.AnError),
					)
			},
			wantErr: true,
		},
		{
			name: "error scan row",
			doMock: func() {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetAllCOAType)).
					WillReturnRows(
						sqlmock.NewRows([]string{"InvalidColumn"}).AddRow(nil),
					)
			},
			wantErr: true,
		},
		{
			name: "error db",
			doMock: func() {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetAllCOAType)).
					WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		suite.t.Run(tc.name, func(t *testing.T) {
			tc.doMock()

			_, err := suite.repo.GetAll(context.Background())
			assert.Equal(t, tc.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *coaTypeTestSuite) TestRepository_Update() {
	type args struct {
		ctx context.Context
		req models.UpdateCOAType
	}

	testCases := []struct {
		name    string
		args    args
		wantErr bool
		doMock  func(args args)
	}{
		{
			name: "test success",
			args: args{ctx: context.TODO(),
				req: models.UpdateCOAType{
					Name:          "Lender Yang Baik",
					NormalBalance: "debit",
					Code:          "001",
					Status:        "active",
				},
			},
			doMock: func(args args) {
				suite.mock.ExpectExec(regexp.QuoteMeta(queryCOATypeUpdate)).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{name: "test error db",
			args: args{
				ctx: context.TODO(),
				req: models.UpdateCOAType{
					Name:          "Lender Yang Baik",
					NormalBalance: "debit",
					Code:          "001",
					Status:        "active",
				},
			},
			doMock: func(args args) {
				suite.mock.ExpectExec(regexp.QuoteMeta(queryCOATypeUpdate)).WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		suite.t.Run(tc.name, func(t *testing.T) {
			tc.doMock(tc.args)

			err := suite.repo.Update(tc.args.ctx, tc.args.req)
			assert.Equal(t, tc.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
