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

func TestSubCategoryRepositoryTestSuite(t *testing.T) {
	t.Helper()
	suite.Run(t, new(subCategoryTestSuite))
}

type subCategoryTestSuite struct {
	suite.Suite
	t    *testing.T
	db   *sql.DB
	mock sqlmock.Sqlmock
	repo SubCategoryRepository
}

func (suite *subCategoryTestSuite) SetupTest() {
	var err error

	suite.db, suite.mock, err = sqlmock.New()
	require.NoError(suite.T(), err)

	suite.t = suite.T()
	suite.repo = NewMySQLRepository(suite.db).GetSubCategoryRepository()
}

func (suite *subCategoryTestSuite) TearDownTest() {
	defer suite.db.Close()
}

func (suite *subCategoryTestSuite) TestRepository_CheckSubCategoryByCodeAndCategoryCode() {
	type args struct {
		ctx                   context.Context
		code, subCategoryCode string
		setupMocks            func()
	}

	testCases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test success",
			args: args{
				ctx:             context.Background(),
				code:            "211",
				subCategoryCode: "100",
				setupMocks: func() {
					rows := sqlmock.NewRows(
						[]string{"id", "code", "name", "description", "category_code", "account_type", "default_product_type_code", "default_currency", "status", "created_at", "updated_at"}).
						AddRow(1, "code", "name", "description", "category_code", "account_type", "default_product_type_code", "default_currency", "status", nil, nil)
					suite.mock.ExpectQuery(regexp.QuoteMeta(querySubCategoryIsExistByCode)).WillReturnRows(rows)
				},
			},
			wantErr: false,
		},
		{
			name: "test data not found",
			args: args{
				ctx: context.TODO(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(querySubCategoryIsExistByCode)).WillReturnError(sql.ErrNoRows)
				},
			},
			wantErr: true,
		},
		{
			name: "test error result",
			args: args{
				ctx: context.TODO(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(querySubCategoryIsExistByCode)).WillReturnError(assert.AnError)
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.args.setupMocks()

			_, err := suite.repo.CheckSubCategoryByCodeAndCategoryCode(tt.args.ctx, tt.args.code, tt.args.subCategoryCode)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *subCategoryTestSuite) TestRepository_GetByCode() {
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
						[]string{"id", "code", "name", "description", "category_code", "account_type", "product_type_code", "currency", "status", "created_at", "updated_at"}).
						AddRow(1, "code", "name", "description", "category_code", "account_type", "product_type_code", "currency", "status", time.Now(), time.Now())
					suite.mock.ExpectQuery(regexp.QuoteMeta(querySubCategoryGetByCode)).WithArgs("211").WillReturnRows(rows)
				},
			},
			wantErr: false,
		},
		{
			name: "test data not found",
			args: args{
				ctx: context.TODO(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(querySubCategoryGetByCode)).WillReturnError(sql.ErrNoRows)
				},
			},
			wantErr: false,
		},
		{
			name: "test error result",
			args: args{
				ctx: context.TODO(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(querySubCategoryGetByCode)).WillReturnError(assert.AnError)
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

func (suite *subCategoryTestSuite) TestRepository_Create() {
	type args struct {
		ctx context.Context
		req models.CreateSubCategory
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
				req: models.CreateSubCategory{
					Code:        "test",
					Name:        "test",
					Description: "test",
				},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(querySubCategoryCreate)).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "error no row",
			args: args{
				ctx: context.TODO(),
				req: models.CreateSubCategory{},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(querySubCategoryCreate)).WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
		{
			name: "error result",
			args: args{
				ctx: context.TODO(),
				req: models.CreateSubCategory{},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(querySubCategoryCreate)).WillReturnResult(sqlmock.NewErrorResult(assert.AnError))
			},
			wantErr: true,
		},
		{
			name: "error db",
			args: args{
				ctx: context.TODO(),
				req: models.CreateSubCategory{},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(querySubCategoryCreate)).WillReturnError(assert.AnError)
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

func (suite *subCategoryTestSuite) TestRepository_GetAll() {
	testCases := []struct {
		name    string
		wantErr bool
		doMock  func()
	}{
		{
			name: "success get all",
			doMock: func() {
				rows := sqlmock.NewRows(
					[]string{"id", "category_code", "code", "name", "description", "account_type", "default_product_type_code", "product_type_name", "default_currency", "status", "created_at", "updated_at"}).
					AddRow(1, "category_code", "code", "name", "description", "account_type", "default_product_type_code", "product_type_name", "default_currency", "status", nil, nil)
				suite.mock.ExpectQuery(regexp.QuoteMeta(queryGetAllSubCategory)).WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "error RowError",
			doMock: func() {
				rows := sqlmock.NewRows(
					[]string{"id", "category_code", "code", "name", "description", "account_type", "default_product_type_code", "product_type_name", "default_currency", "status", "created_at", "updated_at"}).
					AddRow(1, "category_code", "code", "name", "description", "account_type", "default_product_type_code", "product_type_name", "default_currency", "status", nil, nil).RowError(0, assert.AnError)
				suite.mock.ExpectQuery(regexp.QuoteMeta(queryGetAllSubCategory)).WillReturnRows(rows)
			},
			wantErr: true,
		},
		{
			name: "error row",
			doMock: func() {
				suite.mock.ExpectQuery(regexp.QuoteMeta(queryGetAllSubCategory)).WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "failed scan row",
			doMock: func() {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetAllSubCategory)).
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
					ExpectQuery(regexp.QuoteMeta(queryGetAllSubCategory)).
					WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		suite.t.Run(tc.name, func(t *testing.T) {
			tc.doMock()
			_, err := suite.repo.GetAll(context.Background(), models.GetAllSubCategoryParam{
				CategoryCode: "10000",
			})
			assert.Equal(t, tc.wantErr, err != nil)
			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *subCategoryTestSuite) TestRepository_GetByAccountType() {
	testCases := []struct {
		name           string
		accountType    string
		dbExpectations func(mock sqlmock.Sqlmock)
		expectedResult *models.SubCategory
		expectedError  error
	}{
		{
			name:        "Valid Account Type",
			accountType: "valid_account_type",
			dbExpectations: func(mock sqlmock.Sqlmock) {
				// Expect a SELECT query and return a single row
				rows := sqlmock.NewRows(
					[]string{"id", "code", "name", "description", "account_type", "default_product_type_code", "category_code", "default_currency", "status", "created_at", "updated_at"}).
					AddRow(1, "code", "name", "description", "account_type", "default_product_type_code", "category_code", "default_currency", "status", nil, nil)
				mock.ExpectQuery(regexp.QuoteMeta(querySubCategoryGetByAccountType)).WithArgs("valid_account_type", "active").WillReturnRows(rows)
			},
			expectedResult: &models.SubCategory{
				ID:              1,
				Code:            "code",
				Name:            "name",
				Description:     "description",
				ProductTypeCode: "default_product_type_code",
				AccountType:     "account_type",
				CategoryCode:    "category_code",
				Currency:        "default_currency",
				Status:          "status",
			},
			expectedError: nil,
		},
		{
			name:        "No Rows Found",
			accountType: "nonexistent_account_type",
			dbExpectations: func(mock sqlmock.Sqlmock) {
				// Expect a SELECT query with no rows returned
				mock.ExpectQuery("SELECT").
					WithArgs("nonexistent_account_type", "active").
					WillReturnError(sql.ErrNoRows)
			},
			expectedResult: nil,
			expectedError:  models.ErrNoRows,
		},
		{
			name:        "Database Error",
			accountType: "db_error_account_type",
			dbExpectations: func(mock sqlmock.Sqlmock) {
				// Expect a SELECT query and return a database error
				mock.ExpectQuery("SELECT").
					WithArgs("db_error_account_type", "active").
					WillReturnError(assert.AnError)
			},
			expectedResult: nil,
			expectedError:  assert.AnError,
		},
	}
	for _, tc := range testCases {
		suite.t.Run(tc.name, func(t *testing.T) {
			tc.dbExpectations(suite.mock)

			subCat, err := suite.repo.GetByAccountType(context.Background(), tc.accountType)
			assert.Equal(t, tc.expectedError, err)
			assert.Equal(t, tc.expectedResult, subCat)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *subCategoryTestSuite) TestRepository_Update() {
	type args struct {
		ctx context.Context
		req models.UpdateSubCategory
	}
	testCases := []struct {
		name    string
		args    args
		wantErr bool
		doMock  func(args args)
	}{
		{
			name: "success case",
			args: args{ctx: context.TODO(),
				req: models.UpdateSubCategory{
					Name:            "RETAIL",
					Description:     "Retail",
					Code:            "10000",
					Status:          "active",
					ProductTypeCode: &[]string{"1001"}[0],
				},
			},
			doMock: func(args args) {
				suite.mock.ExpectExec(regexp.QuoteMeta(querySubCategoryUpdate)).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{name: "error case - database error",
			args: args{
				ctx: context.TODO(),
				req: models.UpdateSubCategory{
					Name:            "RETAIL",
					Description:     "Retail",
					Code:            "10000",
					Status:          "active",
					ProductTypeCode: &[]string{"1001"}[0],
				},
			},
			doMock: func(args args) {
				suite.mock.ExpectExec(regexp.QuoteMeta(querySubCategoryUpdate)).WillReturnError(assert.AnError)
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

func (suite *subCategoryTestSuite) TestRepository_GetLatestSubCategCode() {
	type args struct {
		ctx        context.Context
		code       string
		setupMocks func()
	}
	testCases := []struct {
		name    string
		code    string
		args    args
		wantErr bool
	}{
		{
			name: "success case",
			code: "131",
			args: args{
				ctx: context.Background(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryGetLatestSubCategoryCode)).
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
			code: "131",
			args: args{
				ctx: context.TODO(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryGetLatestSubCategoryCode)).WillReturnError(assert.AnError)
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.args.setupMocks()

			_, err := suite.repo.GetLatestSubCategCode(tt.args.ctx, tt.args.code)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}

}
