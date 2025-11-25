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

func TestProductTypeRepositoryTestSuite(t *testing.T) {
	t.Helper()
	suite.Run(t, new(productTypeTestSuite))
}

type productTypeTestSuite struct {
	suite.Suite
	t    *testing.T
	db   *sql.DB
	mock sqlmock.Sqlmock
	repo ProductTypeRepository
}

func (suite *productTypeTestSuite) SetupTest() {
	var err error

	suite.db, suite.mock, err = sqlmock.New()
	require.NoError(suite.T(), err)

	suite.t = suite.T()
	suite.repo = NewMySQLRepository(suite.db).GetProductTypeRepository()
}

func (suite *productTypeTestSuite) TearDownTest() {
	defer suite.db.Close()
}

func (suite *productTypeTestSuite) TestRepository_Create() {
	type args struct {
		ctx context.Context
		req models.CreateProductTypeRequest
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
				req: models.CreateProductTypeRequest{
					Code:       "1001",
					Name:       "Chickin",
					Status:     "active",
					EntityCode: "001",
				},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryProductTypeCreate)).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "error no row",
			args: args{
				ctx: context.TODO(),
				req: models.CreateProductTypeRequest{},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryProductTypeCreate)).WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
		{
			name: "error result",
			args: args{
				ctx: context.TODO(),
				req: models.CreateProductTypeRequest{},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryProductTypeCreate)).WillReturnResult(sqlmock.NewErrorResult(assert.AnError))
			},
			wantErr: true,
		},
		{
			name: "error db",
			args: args{
				ctx: context.TODO(),
				req: models.CreateProductTypeRequest{},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryProductTypeCreate)).WillReturnError(assert.AnError)
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

func (suite *productTypeTestSuite) TestRepository_GetByCode() {
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
						[]string{"id", "code", "name", "status", "entity_code", "created_at", "updated_at"}).
						AddRow(1, "101", "Chikin", "active", "entity_code", time.Now(), time.Now())
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryProductTypeGetByCode)).WillReturnRows(rows)
				},
			},
			wantErr: false,
		},
		{
			name: "test data not found",
			args: args{
				ctx: context.TODO(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryProductTypeGetByCode)).WillReturnError(sql.ErrNoRows)
				},
			},
			wantErr: false,
		},
		{
			name: "test error result",
			args: args{
				ctx: context.TODO(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryProductTypeGetByCode)).WillReturnError(assert.AnError)
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

func (suite *productTypeTestSuite) TestRepository_List() {
	testCases := []struct {
		name    string
		wantErr bool
		doMock  func()
	}{
		{
			name: "success get list",
			doMock: func() {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryProductTypeList)).
					WillReturnRows(
						sqlmock.
							NewRows([]string{"id", "code", "name", "status", "entity_code", "created_at", "updated_at"}).
							AddRow(1, "101", "Chikin", "active", "entity_code", time.Now(), time.Now()),
					)
			},
			wantErr: false,
		},
		{
			name: "error row",
			doMock: func() {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryProductTypeList)).
					WillReturnRows(
						sqlmock.
							NewRows([]string{"id", "code", "name", "status", "entity_code", "created_at", "updated_at"}).
							AddRow(1, "101", "Chikin", "active", "entity_code", time.Now(), time.Now()).RowError(0, assert.AnError),
					)
			},
			wantErr: true,
		},
		{
			name: "failed scan row",
			doMock: func() {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryProductTypeList)).
					WillReturnRows(
						sqlmock.NewRows([]string{"InvalidColumn"}).AddRow(nil),
					)
			},
			wantErr: true,
		},
		{
			name: "failed from db",
			doMock: func() {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryProductTypeList)).
					WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range testCases {
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.doMock()

			_, err := suite.repo.List(context.Background())
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *productTypeTestSuite) TestRepository_CheckProductTypeIsExist() {
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
			name: "success case",
			args: args{
				ctx:  context.Background(),
				code: "1001",
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryCheckProductTypeIsExist)).
						WillReturnRows(
							sqlmock.
								NewRows([]string{"id"}).
								AddRow(1))
				},
			},
			wantErr: false,
		},
		{
			name: "error case - error result",
			args: args{
				ctx: context.TODO(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryCheckProductTypeIsExist)).WillReturnError(assert.AnError)
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.args.setupMocks()

			err := suite.repo.CheckProductTypeIsExist(tt.args.ctx, tt.args.code)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *productTypeTestSuite) TestRepository_Update() {
	type args struct {
		ctx        context.Context
		req        models.UpdateProductType
		setupMocks func(req models.UpdateProductType)
	}

	testCases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success case",
			args: args{ctx: context.TODO(),
				req: models.UpdateProductType{
					Name:   "Chickin",
					Status: "active",
					Code:   "1001",
				},
				setupMocks: func(req models.UpdateProductType) {
					query, _, _ := queryProductTypeUpdate(req)
					suite.mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnResult(sqlmock.NewResult(0, 1))
				},
			},
			wantErr: false,
		},
		{
			name: "error case - database error",
			args: args{
				ctx: context.TODO(),
				req: models.UpdateProductType{
					Name:       "Chickin",
					Status:     "active",
					Code:       "1001",
					EntityCode: "001",
				},
				setupMocks: func(req models.UpdateProductType) {
					query, _, _ := queryProductTypeUpdate(req)
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

func (suite *productTypeTestSuite) TestRepository_GetLatestProductCode() {
	type args struct {
		ctx        context.Context
		setupMocks func()
	}
	testCases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success case",
			args: args{
				ctx: context.Background(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryGetLatestProductCode)).
						WillReturnRows(
							sqlmock.
								NewRows([]string{"code"}).
								AddRow(1))
				},
			},
			wantErr: false,
		},
		{
			name: "error case - error result",
			args: args{
				ctx: context.TODO(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryGetLatestProductCode)).WillReturnError(assert.AnError)
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.args.setupMocks()

			_, err := suite.repo.GetLatestProductCode(tt.args.ctx)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
