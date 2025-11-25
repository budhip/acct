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

func TestCategoryRepositoryTestSuite(t *testing.T) {
	t.Helper()
	suite.Run(t, new(categoryTestSuite))
}

type categoryTestSuite struct {
	suite.Suite
	t    *testing.T
	db   *sql.DB
	mock sqlmock.Sqlmock
	repo CategoryRepository
}

func (suite *categoryTestSuite) SetupTest() {
	var err error

	suite.db, suite.mock, err = sqlmock.New()
	require.NoError(suite.T(), err)

	suite.t = suite.T()
	suite.repo = NewMySQLRepository(suite.db).GetCategoryRepository()
}

func (suite *categoryTestSuite) TearDownTest() {
	defer suite.db.Close()
}

func (suite *categoryTestSuite) TestRepository_CheckCategoryByCode() {
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
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryCategoryIsExistByCode)).WillReturnRows(rows)
				},
			},
			wantErr: false,
		},
		{
			name: "test data not found",
			args: args{
				ctx: context.TODO(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryCategoryIsExistByCode)).WillReturnError(sql.ErrNoRows)
				},
			},
			wantErr: true,
		},
		{
			name: "test error result",
			args: args{
				ctx: context.TODO(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryCategoryIsExistByCode)).WillReturnError(assert.AnError)
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.args.setupMocks()

			err := suite.repo.CheckCategoryByCode(tt.args.ctx, tt.args.code)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *categoryTestSuite) TestRepository_GetByCode() {
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
						[]string{"id", "code", "name", "description", "coa_type_code", "status", "created_at", "updated_at"}).
						AddRow(1, "666", "ENT", "this is description", "coa_type_code", "status", time.Now(), time.Now())
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryCategoryGetByCode)).WillReturnRows(rows)
				},
			},
			wantErr: false,
		},
		{
			name: "test data not found",
			args: args{
				ctx: context.TODO(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryCategoryGetByCode)).WillReturnError(sql.ErrNoRows)
				},
			},
			wantErr: false,
		},
		{
			name: "test error result",
			args: args{
				ctx: context.TODO(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryCategoryGetByCode)).WillReturnError(assert.AnError)
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.args.setupMocks()

			_, err := suite.repo.GetByCode(tt.args.ctx, tt.args.code)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *categoryTestSuite) TestRepository_Create() {
	type args struct {
		ctx context.Context
		req models.CreateCategoryIn
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
				req: models.CreateCategoryIn{
					Code:        "test",
					Name:        "test",
					Description: "test",
				},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryCategoryCreate)).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "error no row",
			args: args{
				ctx: context.TODO(),
				req: models.CreateCategoryIn{},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryCategoryCreate)).WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
		{
			name: "error result",
			args: args{
				ctx: context.TODO(),
				req: models.CreateCategoryIn{},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryCategoryCreate)).WillReturnResult(sqlmock.NewErrorResult(assert.AnError))
			},
			wantErr: true,
		},
		{
			name: "error db",
			args: args{
				ctx: context.TODO(),
				req: models.CreateCategoryIn{},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryCategoryCreate)).WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.doMock(tt.args)

			err := suite.repo.Create(tt.args.ctx, &tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *categoryTestSuite) TestRepository_List() {
	testCases := []struct {
		name    string
		wantErr bool
		doMock  func()
	}{
		{
			name: "success",
			doMock: func() {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryCategoryList)).
					WillReturnRows(
						sqlmock.
							NewRows([]string{"id", "code", "name", "description", "coa_type_code", "status", "created_at", "updated_at"}).
							AddRow(1, "code", "name", "desc", "coa_type_code", "status", time.Now(), time.Now()),
					)
			},
			wantErr: false,
		},
		{
			name: "error row",
			doMock: func() {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryCategoryList)).
					WillReturnRows(
						sqlmock.
							NewRows([]string{"id", "code", "name", "description", "coa_type_code", "status", "created_at", "updated_at"}).
							AddRow(1, "code", "name", "desc", "coa_type_code", "status", time.Now(), time.Now()).RowError(0, assert.AnError),
					)
			},
			wantErr: true,
		},
		{
			name: "error scan row",
			doMock: func() {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryCategoryList)).
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
					ExpectQuery(regexp.QuoteMeta(queryCategoryList)).
					WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		suite.t.Run(tc.name, func(t *testing.T) {
			tc.doMock()

			_, err := suite.repo.List(context.Background())
			assert.Equal(t, tc.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *categoryTestSuite) TestRepository_GetByCoaCode() {
	testCases := []struct {
		name    string
		wantErr bool
		code    string
		doMock  func()
	}{
		{
			name: "success",
			code: "123",
			doMock: func() {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryCategoryListByCoa)).
					WillReturnRows(
						sqlmock.
							NewRows([]string{"code", "name"}).
							AddRow("code", "name"),
					)
			},
			wantErr: false,
		},
		{
			name: "error row",
			code: "123",
			doMock: func() {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryCategoryListByCoa)).
					WillReturnRows(
						sqlmock.
							NewRows([]string{"code", "name"}).
							AddRow("code", "name").RowError(0, assert.AnError),
					)
			},
			wantErr: true,
		},
		{
			name: "error scan row",
			code: "123",
			doMock: func() {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryCategoryListByCoa)).
					WillReturnRows(
						sqlmock.NewRows([]string{"InvalidColumn"}).AddRow(nil),
					)
			},
			wantErr: true,
		},
		{
			name: "error db",
			code: "123",
			doMock: func() {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryCategoryListByCoa)).
					WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		suite.t.Run(tc.name, func(t *testing.T) {
			tc.doMock()

			_, err := suite.repo.GetByCoaCode(context.Background(), tc.code)
			assert.Equal(t, tc.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *categoryTestSuite) TestRepository_Update() {
	type args struct {
		ctx        context.Context
		req        models.UpdateCategoryIn
		setupMocks func()
	}
	testCases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test success",
			args: args{ctx: context.TODO(),
				req: models.UpdateCategoryIn{
					Name:          "Kas Teller 111",
					Code: 		   "111",	
					Description:   "Kas Teller ceritanya 111",
					CoaTypeCode:   "LIA",
				},
				setupMocks: func() {
					suite.mock.ExpectExec(regexp.QuoteMeta(queryCeCategoryUpdate)).WillReturnResult(sqlmock.NewResult(0, 1))
				},
			},
			wantErr: false,
		},
		{name: "test error db",
			args: args{
				ctx: context.TODO(),
				req: models.UpdateCategoryIn{
					Name:          "Kas Teller 111",
					Code: 		   "111",	
					Description:   "Kas Teller ceritanya 111",
					CoaTypeCode:   "LIA",
				},
				setupMocks: func() {
					suite.mock.ExpectExec(regexp.QuoteMeta(queryCeCategoryUpdate)).WillReturnError(assert.AnError)
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.args.setupMocks()
			err := suite.repo.Update(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
