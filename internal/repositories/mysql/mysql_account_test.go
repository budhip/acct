package mysql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestAccountRepositoryTestSuite(t *testing.T) {
	t.Helper()
	suite.Run(t, new(accountTestSuite))
}

type accountTestSuite struct {
	suite.Suite
	t    *testing.T
	db   *sql.DB
	mock sqlmock.Sqlmock
	repo AccountRepository
}

func (suite *accountTestSuite) SetupTest() {
	var err error

	suite.db, suite.mock, err = sqlmock.New()
	require.NoError(suite.T(), err)

	suite.t = suite.T()
	suite.repo = NewMySQLRepository(suite.db).GetAccountRepository()
}

func (suite *accountTestSuite) TearDownTest() {
	defer suite.db.Close()
}

func (suite *accountTestSuite) TestRepository_Create() {
	type args struct {
		ctx        context.Context
		req        models.CreateAccount
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
				ctx: context.TODO(),
				req: models.CreateAccount{},
				setupMocks: func() {
					suite.mock.ExpectExec(regexp.QuoteMeta(queryAccountCreate)).WillReturnResult(sqlmock.NewResult(0, 1))
				},
			},
			wantErr: false,
		},
		{
			name: "test error result",
			args: args{
				ctx: context.TODO(),
				req: models.CreateAccount{},
				setupMocks: func() {
					suite.mock.ExpectExec(regexp.QuoteMeta(queryAccountCreate)).WillReturnResult(sqlmock.NewErrorResult(assert.AnError))
				},
			},
			wantErr: true,
		},
		{
			name: "test error no row affected",
			args: args{
				ctx: context.TODO(),
				req: models.CreateAccount{},
				setupMocks: func() {
					suite.mock.ExpectExec(regexp.QuoteMeta(queryAccountCreate)).WillReturnResult(sqlmock.NewResult(0, 0))
				},
			},
			wantErr: true,
		},
		{
			name: "test error db",
			args: args{
				ctx: context.TODO(),
				req: models.CreateAccount{},
				setupMocks: func() {
					suite.mock.ExpectExec(regexp.QuoteMeta(queryAccountCreate)).WillReturnError(assert.AnError)
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.args.setupMocks()

			err := suite.repo.Create(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountTestSuite) TestRepository_Update() {
	type args struct {
		ctx        context.Context
		req        models.UpdateAccount
		setupMocks func(args)
	}
	testCases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test success",
			args: args{ctx: context.TODO(),
				req: models.UpdateAccount{
					Name:          "Lender Yang Baik",
					OwnerID:       "12345",
					AltID:         "534534534555353523523423423",
					AccountNumber: "22200100000001",
				},
				setupMocks: func(a args) {
					listQuery, _, _ := buildUpdateAccountQuery(a.req)
					suite.mock.ExpectExec(regexp.QuoteMeta(listQuery)).WillReturnResult(sqlmock.NewResult(0, 1))
				},
			},
			wantErr: false,
		},
		{
			name: "test error db",
			args: args{
				ctx: context.TODO(),
				req: models.UpdateAccount{
					Name:          "Lender Yang Baik",
					OwnerID:       "12345",
					AltID:         "534534534555353523523423423",
					AccountNumber: "22200100000001",
				},
				setupMocks: func(a args) {
					listQuery, _, _ := buildUpdateAccountQuery(a.req)
					suite.mock.ExpectExec(regexp.QuoteMeta(listQuery)).WillReturnError(assert.AnError)
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.args.setupMocks(tt.args)

			err := suite.repo.Update(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountTestSuite) TestRepository_UpdateEntity() {
	type args struct {
		ctx        context.Context
		req        models.UpdateAccountEntity
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
				req: models.UpdateAccountEntity{
					AccountNumber: "22200100000001",
					EntityCode:    "001",
				},
				setupMocks: func() {
					suite.mock.ExpectExec(regexp.QuoteMeta(queryUpdateAccountEntity)).WillReturnResult(sqlmock.NewResult(0, 1))
				},
			},
			wantErr: false,
		},
		{
			name: "test error db",
			args: args{
				ctx: context.TODO(),
				req: models.UpdateAccountEntity{
					AccountNumber: "22200100000001",
					EntityCode:    "001",
				},
				setupMocks: func() {
					suite.mock.ExpectExec(regexp.QuoteMeta(queryUpdateAccountEntity)).WillReturnError(assert.AnError)
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.args.setupMocks()

			err := suite.repo.UpdateEntity(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountTestSuite) TestRepository_UpdateLegacyId() {
	type args struct {
		ctx        context.Context
		req        models.UpdateLegacyId
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
				req: models.UpdateLegacyId{
					LegacyId:      nil,
					AccountNumber: "22200100000001",
				},
				setupMocks: func() {
					suite.mock.ExpectExec(regexp.QuoteMeta(queryAccountUpdateLegacyId)).WillReturnResult(sqlmock.NewResult(0, 1))
				},
			},
			wantErr: false,
		},
		{name: "test error db",
			args: args{
				ctx: context.TODO(),
				req: models.UpdateLegacyId{
					LegacyId:      nil,
					AccountNumber: "22200100000001",
				},
				setupMocks: func() {
					suite.mock.ExpectExec(regexp.QuoteMeta(queryAccountUpdateLegacyId)).WillReturnError(assert.AnError)
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.args.setupMocks()

			err := suite.repo.UpdateLegacyId(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountTestSuite) TestRepository_UpdateAltId() {
	type args struct {
		ctx        context.Context
		req        models.UpdateAltId
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
				req: models.UpdateAltId{
					AltId:         "123321",
					AccountNumber: "22200100000001",
				},
				setupMocks: func() {
					suite.mock.ExpectExec(regexp.QuoteMeta(queryAccountUpdateAltId)).WillReturnResult(sqlmock.NewResult(0, 1))
				},
			},
			wantErr: false,
		},
		{name: "test error db",
			args: args{
				ctx: context.TODO(),
				req: models.UpdateAltId{
					AccountNumber: "22200100000001",
				},
				setupMocks: func() {
					suite.mock.ExpectExec(regexp.QuoteMeta(queryAccountUpdateAltId)).WillReturnError(assert.AnError)
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.args.setupMocks()

			err := suite.repo.UpdateAltId(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountTestSuite) TestRepository_UpdateBySubCategory() {
	type args struct {
		ctx        context.Context
		req        models.UpdateBySubCategory
		setupMocks func()
	}

	query := queryAccountUpdateBySubCategory + ` product_type_code = ?, ` + ` currency = ?, ` + queryAccountUpdateBySubCategoryWhere

	testCases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test success",
			args: args{ctx: context.TODO(),
				req: models.UpdateBySubCategory{
					ProductTypeCode: &[]string{"1009"}[0],
					Currency:        &[]string{"IDR"}[0],
					Code:            "1000",
				},
				setupMocks: func() {
					suite.mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnResult(sqlmock.NewResult(0, 1))
				},
			},
			wantErr: false,
		},
		{
			name: "test error db",
			args: args{
				ctx: context.TODO(),
				req: models.UpdateBySubCategory{
					ProductTypeCode: &[]string{"1009"}[0],
					Currency:        &[]string{"IDR"}[0],
					Code:            "1000",
				},
				setupMocks: func() {
					suite.mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnError(assert.AnError)
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.args.setupMocks()

			err := suite.repo.UpdateBySubCategory(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountTestSuite) TestRepository_GetOneByAccountNumber() {
	var emptyJson = new(*map[string]interface{})
	type args struct {
		ctx           context.Context
		accountNumber string
		setupMocks    func()
	}

	testCases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test success",
			args: args{ctx: context.TODO(),
				accountNumber: "21100100000001",
				setupMocks: func() {
					rows := sqlmock.NewRows([]string{"account_number", "account_name", "owner_id", "category_code", "category_name",
						"coa_type_code", "coa_type_name", "sub_category_code", "sub_category_name", "entity_code", "entity_name", "product_type_code", "product_type_name",
						"currency", "status", "alt_id", "created_at", "updated_at", "legacy_id", "metadata", "account_type"}).
						AddRow("21100100000001", "Lender Yang Baik", "121212", "211", "Marketplace Payable (Lender Balance)",
							"LIA", "Liability", "21101", "Lender Balance - Individual Non RDL", "001", "PT. Amartha Mikro Fintek (AMF)", "1001", "Group Loan",
							"IDR", "active", "535235235235325", time.Now(), time.Now(), emptyJson, emptyJson, "INDIVIDU")
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryAccountByNumber)).WillReturnRows(rows)
				},
			},
			wantErr: false,
		},
		{name: "test data not found",
			args: args{
				ctx: context.TODO(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryAccountByNumber)).WillReturnError(sql.ErrNoRows)
				},
			},
			wantErr: true,
		},
		{
			name: "test error result",
			args: args{
				ctx: context.TODO(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryAccountByNumber)).WillReturnError(assert.AnError)
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.args.setupMocks()

			_, err := suite.repo.GetOneByAccountNumber(tt.args.ctx, tt.args.accountNumber)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountTestSuite) TestRepository_GetOneByLegacyID() {
	var emptyJson = new(*map[string]interface{})
	type args struct {
		ctx           context.Context
		accountNumber string
		setupMocks    func()
	}

	testCases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test success",
			args: args{ctx: context.TODO(),
				accountNumber: "21100100000001",
				setupMocks: func() {
					rows := sqlmock.NewRows([]string{"account_number", "account_name", "owner_id", "category_code", "category_name",
						"coa_type_code", "coa_type_name", "sub_category_code", "sub_category_name", "entity_code", "entity_name", "product_type_code", "product_type_name",
						"currency", "status", "alt_id", "created_at", "updated_at", "legacy_id", "metadata", "account_type"}).
						AddRow("21100100000001", "Lender Yang Baik", "121212", "211", "Marketplace Payable (Lender Balance)",
							"LIA", "Liability", "21101", "Lender Balance - Individual Non RDL", "001", "PT. Amartha Mikro Fintek (AMF)", "1001", "Group Loan",
							"IDR", "active", "535235235235325", time.Now(), time.Now(), emptyJson, emptyJson, "INDIVIDU")
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryAccountByLegacyID)).WillReturnRows(rows)
				},
			},
			wantErr: false,
		},
		{
			name: "test data not found",
			args: args{
				ctx: context.TODO(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryAccountByLegacyID)).WillReturnError(sql.ErrNoRows)
				},
			},
			wantErr: true,
		},
		{
			name: "test error result",
			args: args{
				ctx: context.TODO(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryAccountByLegacyID)).WillReturnError(assert.AnError)
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.args.setupMocks()

			_, err := suite.repo.GetOneByLegacyID(tt.args.ctx, tt.args.accountNumber)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountTestSuite) TestRepository_GetAccountList() {
	date := atime.Now()
	defaultColumns := []string{
		"account_number",
		"account_name",
		"category_code",
		"sub_category_code",
		"entity_code",
		"entity_name",
		"product_type_code",
		"product_type_name",
		"alt_id",
		"owner_id",
		"status",
		"legacy_id",
		"created_at",
		"updated_at",
		"t24_account_number",
	}
	rows := sqlmock.
		NewRows(defaultColumns).
		AddRow(
			"11200100000001",
			"Kas BP Love You",
			"112",
			"11201",
			"001",
			"AMF",
			"1001",
			"Group Loan",
			"5345345345553535235234234232",
			"123456789",
			"active",
			nil,
			date,
			date,
			"123456789",
		)
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
	res := []models.GetAccountOut{
		{
			AccountNumber:   "11200100000001",
			AccountName:     "Kas BP Love You",
			CoaTypeCode:     "AST",
			CoaTypeName:     "Liability",
			CategoryCode:    "112",
			CategoryName:    "KAS BP",
			SubCategoryCode: "11201",
			SubCategoryName: "Kas BP",
			EntityCode:      "001",
			EntityName:      "AMF",
			ProductTypeCode: "1001",
			ProductTypeName: "Group Loan",
			AltID:           "5345345345553535235234234232",
			OwnerID:         "123456789",
			Status:          "active",
			LegacyId:        nil,
			CreatedAt:       date,
			UpdatedAt:       date,
		},
	}
	type args struct {
		ctx  context.Context
		opts models.AccountFilterOptions
	}
	testCases := []struct {
		name       string
		args       args
		setupMocks func(a args)
		wantErr    bool
		expected   []models.GetAccountOut
	}{
		{
			name: "success case - get account list sort created_at",
			args: args{
				ctx: context.TODO(),
				opts: models.AccountFilterOptions{
					Search:          "11201",
					CoaTypeCode:     "AST",
					EntityCode:      "001",
					CategoryCode:    "112",
					SubCategoryCode: "11201",
					ProductTypeCode: "1001",
				},
			},
			setupMocks: func(a args) {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetCategorySubCategoryCOAType)).
					WillReturnRows(rowsCategorySubCategoryCOAType)

				listQuery, _, _ := buildAccountListQuery(a.opts)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(listQuery)).
					WillReturnRows(rows)
			},
			expected: res,
			wantErr:  false,
		},
		{
			name: "error case - row error",
			args: args{
				ctx: context.TODO(),
				opts: models.AccountFilterOptions{
					CoaTypeCode:     "AST",
					EntityCode:      "001",
					CategoryCode:    "112",
					SubCategoryCode: "11201",
					ProductTypeCode: "1001",
				},
			},
			setupMocks: func(a args) {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetCategorySubCategoryCOAType)).
					WillReturnRows(rowsCategorySubCategoryCOAType)
				listQuery, _, _ := buildAccountListQuery(a.opts)
				rows := sqlmock.
					NewRows(defaultColumns).
					AddRow(
						"11200100000001",
						"Kas BP Love You",
						"112",
						"11201",
						"001",
						"AMF",
						"1001",
						"Group Loan",
						"5345345345553535235234234232",
						"123456789",
						"active",
						nil,
						date,
						date,
						"",
					).RowError(0, assert.AnError)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(listQuery)).
					WillReturnRows(rows)
			},
			expected: []models.GetAccountOut{},
			wantErr:  true,
		},
		{
			name: "error case - failed scan row",
			args: args{
				ctx: context.TODO(),
				opts: models.AccountFilterOptions{
					CoaTypeCode:     "AST",
					EntityCode:      "001",
					CategoryCode:    "112",
					SubCategoryCode: "11201",
					ProductTypeCode: "1001",
				},
			},
			setupMocks: func(a args) {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetCategorySubCategoryCOAType)).
					WillReturnRows(rowsCategorySubCategoryCOAType)
				listQuery, _, _ := buildAccountListQuery(a.opts)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(listQuery)).
					WillReturnRows(sqlmock.NewRows([]string{"InvalidColumn"}).AddRow(nil))
			},
			expected: []models.GetAccountOut{},
			wantErr:  true,
		},
		{
			name: "error case - database error",
			args: args{
				ctx: context.TODO(),
				opts: models.AccountFilterOptions{
					CoaTypeCode:     "AST",
					EntityCode:      "001",
					CategoryCode:    "112",
					SubCategoryCode: "11201",
					ProductTypeCode: "1001",
				},
			},
			setupMocks: func(a args) {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetCategorySubCategoryCOAType)).
					WillReturnRows(rowsCategorySubCategoryCOAType)
				listQuery, _, _ := buildAccountListQuery(a.opts)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(listQuery)).
					WillReturnError(assert.AnError)
			},
			expected: []models.GetAccountOut{},
			wantErr:  true,
		},
		{
			name: "error case - database error GetCategorySubCategoryCOAType",
			args: args{
				ctx: context.TODO(),
				opts: models.AccountFilterOptions{
					CoaTypeCode:     "AST",
					EntityCode:      "001",
					CategoryCode:    "112",
					SubCategoryCode: "11201",
					ProductTypeCode: "1001",
				},
			},
			setupMocks: func(a args) {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryGetCategorySubCategoryCOAType)).
					WillReturnError(assert.AnError)
			},
			expected: []models.GetAccountOut{},
			wantErr:  true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(tt.args)

			_, err := suite.repo.GetAccountList(tt.args.ctx, tt.args.opts)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountTestSuite) TestRepository_GetAccountListCount() {
	type args struct {
		ctx  context.Context
		opts models.AccountFilterOptions
	}

	testCases := []struct {
		name       string
		args       args
		setupMocks func(a args)
		wantErr    bool
		expected   int
	}{
		{
			name: "success count get account list",
			args: args{
				ctx: context.TODO(),
				opts: models.AccountFilterOptions{
					Search:          "11201",
					CoaTypeCode:     "AST",
					EntityCode:      "001",
					CategoryCode:    "112",
					SubCategoryCode: "11201",
				},
			},
			setupMocks: func(a args) {
				listQuery, _, _ := buildCountAccountListQuery(a.opts)
				rows := sqlmock.
					NewRows([]string{"count"}).
					AddRow(1)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(listQuery)).
					WillReturnRows(rows)
			},
			expected: 1,
			wantErr:  false,
		},
		{
			name: "error count get account list",
			args: args{
				ctx: context.TODO(),
				opts: models.AccountFilterOptions{
					Search:          "11201",
					CoaTypeCode:     "AST",
					EntityCode:      "001",
					CategoryCode:    "112",
					SubCategoryCode: "11201",
				},
			},
			setupMocks: func(a args) {
				listQuery, _, _ := buildCountAccountListQuery(a.opts)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(listQuery)).
					WillReturnError(assert.AnError)
			},
			expected: 0,
			wantErr:  true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(tt.args)

			_, err := suite.repo.GetAccountListCount(tt.args.ctx, tt.args.opts)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountTestSuite) TestRepository_CheckExistByParam() {
	type args struct {
		ctx        context.Context
		param      models.AccountFilterOptions
		setupMocks func()
	}

	testCases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test success check alt id exist",
			args: args{ctx: context.TODO(),
				param: models.AccountFilterOptions{AltID: "21100100000001"},
				setupMocks: func() {
					rows := sqlmock.NewRows([]string{"id"}).AddRow("123456")
					query, _, _ := buildQueryCheckExistByParam(models.AccountFilterOptions{AltID: "21100100000001"})
					suite.mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(rows)
				},
			},
			wantErr: false,
		},
		{
			name: "test success check entity and subcategory exist",
			args: args{ctx: context.TODO(),
				param: models.AccountFilterOptions{EntityCode: "001", SubCategoryCode: "21101"},
				setupMocks: func() {
					rows := sqlmock.NewRows([]string{"id"}).AddRow("123456")
					query, _, _ := buildQueryCheckExistByParam(models.AccountFilterOptions{EntityCode: "001", SubCategoryCode: "21101"})
					suite.mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(rows)
				},
			},
			wantErr: false,
		},
		{
			name: "test data not found",
			args: args{
				ctx:   context.TODO(),
				param: models.AccountFilterOptions{AltID: "21100100000001"},
				setupMocks: func() {
					query, _, _ := buildQueryCheckExistByParam(models.AccountFilterOptions{AltID: "21100100000001"})
					suite.mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(sql.ErrNoRows)
				},
			},
			wantErr: false,
		},
		{
			name: "test error result",
			args: args{
				ctx:   context.TODO(),
				param: models.AccountFilterOptions{AltID: "21100100000001"},
				setupMocks: func() {
					query, _, _ := buildQueryCheckExistByParam(models.AccountFilterOptions{AltID: "21100100000001"})
					suite.mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(assert.AnError)
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.args.setupMocks()

			_, err := suite.repo.CheckExistByParam(tt.args.ctx, tt.args.param)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountTestSuite) TestRepository_BulkInsertAccount() {
	type args struct {
		ctx context.Context
		req []models.CreateAccount
	}
	valueStrings := "(?, ?)"
	queryBulkInsertAcctAccount := fmt.Sprintf(queryBulkInsertAccount, valueStrings)

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
				req: []models.CreateAccount{{}},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryBulkInsertAcctAccount)).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "error no row",
			args: args{
				ctx: context.TODO(),
				req: []models.CreateAccount{{}},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryBulkInsertAcctAccount)).WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
		{
			name: "error result",
			args: args{
				ctx: context.TODO(),
				req: []models.CreateAccount{{}},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryBulkInsertAcctAccount)).WillReturnResult(sqlmock.NewErrorResult(assert.AnError))
			},
			wantErr: true,
		},
		{
			name: "error db",
			args: args{
				ctx: context.TODO(),
				req: []models.CreateAccount{{}},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryBulkInsertAcctAccount)).WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.doMock(tt.args)

			err := suite.repo.BulkInsertAccount(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountTestSuite) TestRepository_BulkInsertAcctAccount() {
	type args struct {
		ctx context.Context
		req []models.CreateAccount
	}
	valueStrings := "(NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, '{}'), NULLIF(?, '{}'), CURRENT_TIMESTAMP(6), CURRENT_TIMESTAMP(6))"
	queryBulkInsertAcctAccount := fmt.Sprintf(queryBulkInsertAcctAccount, valueStrings)

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
				req: []models.CreateAccount{{}},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryBulkInsertAcctAccount)).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "error no row",
			args: args{
				ctx: context.TODO(),
				req: []models.CreateAccount{{}},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryBulkInsertAcctAccount)).WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
		{
			name: "error result",
			args: args{
				ctx: context.TODO(),
				req: []models.CreateAccount{{}},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryBulkInsertAcctAccount)).WillReturnResult(sqlmock.NewErrorResult(assert.AnError))
			},
			wantErr: true,
		},
		{
			name: "error db",
			args: args{
				ctx: context.TODO(),
				req: []models.CreateAccount{{}},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryBulkInsertAcctAccount)).WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.doMock(tt.args)

			err := suite.repo.BulkInsertAcctAccount(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountTestSuite) TestRepository_CreateLenderAccount() {
	type args struct {
		ctx        context.Context
		req        models.CreateLenderAccount
		setupMocks func()
	}

	testCases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success casae",
			args: args{
				ctx: context.TODO(),
				req: models.CreateLenderAccount{},
				setupMocks: func() {
					suite.mock.ExpectExec(regexp.QuoteMeta(queryCreateLenderAccount)).WillReturnResult(sqlmock.NewResult(0, 1))
				},
			},
			wantErr: false,
		},
		{
			name: "error case - return result error",
			args: args{
				ctx: context.TODO(),
				req: models.CreateLenderAccount{},
				setupMocks: func() {
					suite.mock.ExpectExec(regexp.QuoteMeta(queryCreateLenderAccount)).WillReturnResult(sqlmock.NewErrorResult(assert.AnError))
				},
			},
			wantErr: true,
		},
		{
			name: "error case - no row affected",
			args: args{
				ctx: context.TODO(),
				req: models.CreateLenderAccount{},
				setupMocks: func() {
					suite.mock.ExpectExec(regexp.QuoteMeta(queryCreateLenderAccount)).WillReturnResult(sqlmock.NewResult(0, 0))
				},
			},
			wantErr: true,
		},
		{
			name: "error case - query error",
			args: args{
				ctx: context.TODO(),
				req: models.CreateLenderAccount{},
				setupMocks: func() {
					suite.mock.ExpectExec(regexp.QuoteMeta(queryCreateLenderAccount)).WillReturnError(assert.AnError)
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.args.setupMocks()

			err := suite.repo.CreateLenderAccount(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountTestSuite) TestRepository_GetAllAccountNumbersByParam() {
	type args struct {
		ctx  context.Context
		opts models.GetAllAccountNumbersByParamIn
	}
	defaultColumns := []string{
		`aa.owner_id`,
		`aa.account_number`,
		`coalesce(aa.alt_id, "") alt_id`,
		`coalesce(aa.name,"") account_name`,
		`coalesce(asc2.account_type,'') account_type`,
		`coalesce(aa.entity_code,"") entity_code`,
		`coalesce(aa.product_type_code, "") product_type_code`,
		`coalesce(aa.category_code,"") category_code`,
		`coalesce(aa.sub_category_code,"") sub_category_code`,
		`coalesce(aa.currency,"") currency`,
		`coalesce(aa.status,"") status`,
		`coalesce(aa.legacy_id, "{}") legacy_id`,
		`coalesce(aa.metadata, "{}") metadata`,
		`aa.created_at`,
	}
	var emptyJson = new(*map[string]interface{})
	row := []driver.Value{
		"1170006831580",
		"114001000000012",
		"1170006831580",
		"Bank Mandiri Escrow",
		"BANK_HO_MANDIRI",
		"001",
		"",
		"114",
		"11402",
		"IDR",
		"active",
		emptyJson,
		emptyJson,
		time.Now(),
	}
	testCases := []struct {
		name       string
		args       args
		setupMocks func(a args)
		wantErr    bool
		expected   []models.GetAllAccountNumbersByParamOut
	}{
		{
			name: "success get all account numbers by param",
			args: args{
				ctx: context.TODO(),
				opts: models.GetAllAccountNumbersByParamIn{
					OwnerId:         "12345678910",
					SubCategoryCode: "12345",
				},
			},
			setupMocks: func(a args) {
				query, _, _ := buildQueryGetAllAccountNumbersByParam(a.opts)
				rows := sqlmock.
					NewRows(defaultColumns).
					AddRow(row...)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(rows)
			},
			expected: []models.GetAllAccountNumbersByParamOut{
				{
					AccountNumber:   "11200100000001",
					AccountType:     "LENDER_INSTITUSI_NON_RDL",
					SubCategoryCode: "21103",
					CreatedAt:       atime.Now(),
				},
			},
			wantErr: false,
		},
		{
			name: "error row",
			args: args{
				ctx: context.TODO(),
				opts: models.GetAllAccountNumbersByParamIn{
					OwnerId: "JanganDipaksain",
				},
			},
			setupMocks: func(a args) {
				query, _, _ := buildQueryGetAllAccountNumbersByParam(a.opts)
				rows := sqlmock.
					NewRows(defaultColumns).
					AddRow(row...).RowError(0, assert.AnError)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(rows)
			},
			expected: []models.GetAllAccountNumbersByParamOut{},
			wantErr:  true,
		},
		{
			name: "error scan row",
			args: args{
				ctx: context.TODO(),
				opts: models.GetAllAccountNumbersByParamIn{
					OwnerId: "JanganDipaksain",
				},
			},
			setupMocks: func(a args) {
				query, _, _ := buildQueryGetAllAccountNumbersByParam(a.opts)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(sqlmock.NewRows([]string{"InvalidColumn"}).AddRow(nil))
			},
			expected: []models.GetAllAccountNumbersByParamOut{},
			wantErr:  true,
		},
		{
			name: "error database",
			args: args{
				ctx: context.TODO(),
				opts: models.GetAllAccountNumbersByParamIn{
					OwnerId: "JanganDipaksain",
				},
			},
			setupMocks: func(a args) {
				query, _, _ := buildQueryGetAllAccountNumbersByParam(a.opts)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnError(assert.AnError)
			},
			expected: []models.GetAllAccountNumbersByParamOut{},
			wantErr:  true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(tt.args)

			_, err := suite.repo.GetAllAccountNumbersByParam(tt.args.ctx, tt.args.opts)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountTestSuite) TestRepository_GetInvestedAccountNumberByCIHAccountNumber() {
	type args struct {
		ctx           context.Context
		accountNumber string
		setupMocks    func()
	}

	testCases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success case",
			args: args{ctx: context.TODO(),
				accountNumber: "211001000381110",
				setupMocks: func() {
					rows := sqlmock.NewRows([]string{"cih_account_number", "invested_account_number", "receivables_account_number"}).
						AddRow("211001000381110", "212001000000001", "142001000000001")
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryGetLenderAccountByCIHAccountNumber)).WillReturnRows(rows)
				},
			},
			wantErr: false,
		},
		{
			name: "error case - error result",
			args: args{
				ctx: context.TODO(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryGetLenderAccountByCIHAccountNumber)).WillReturnError(assert.AnError)
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.args.setupMocks()

			_, err := suite.repo.GetLenderAccountByCIHAccountNumber(tt.args.ctx, tt.args.accountNumber)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountTestSuite) TestRepository_CreateLoanAccount() {
	type args struct {
		ctx        context.Context
		req        models.CreateLoanAccount
		setupMocks func()
	}
	testCases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success casae",
			args: args{
				ctx: context.TODO(),
				req: models.CreateLoanAccount{},
				setupMocks: func() {
					suite.mock.ExpectExec(regexp.QuoteMeta(queryCreateLoanAccount)).WillReturnResult(sqlmock.NewResult(0, 1))
				},
			},
			wantErr: false,
		},
		{
			name: "error case - return result error",
			args: args{
				ctx: context.TODO(),
				req: models.CreateLoanAccount{},
				setupMocks: func() {
					suite.mock.ExpectExec(regexp.QuoteMeta(queryCreateLoanAccount)).WillReturnResult(sqlmock.NewErrorResult(assert.AnError))
				},
			},
			wantErr: true,
		},
		{
			name: "error case - no row affected",
			args: args{
				ctx: context.TODO(),
				req: models.CreateLoanAccount{},
				setupMocks: func() {
					suite.mock.ExpectExec(regexp.QuoteMeta(queryCreateLoanAccount)).WillReturnResult(sqlmock.NewResult(0, 0))
				},
			},
			wantErr: true,
		},
		{
			name: "error case - query error",
			args: args{
				ctx: context.TODO(),
				req: models.CreateLoanAccount{},
				setupMocks: func() {
					suite.mock.ExpectExec(regexp.QuoteMeta(queryCreateLoanAccount)).WillReturnError(assert.AnError)
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.args.setupMocks()

			err := suite.repo.CreateLoanAccount(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountTestSuite) TestRepository_CheckLegacyIdIsExist() {
	type args struct {
		ctx           context.Context
		accountNumber string
		setupMocks    func()
	}
	testCases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test success",
			args: args{ctx: context.TODO(),
				accountNumber: "21100100000001",
				setupMocks: func() {
					rows := sqlmock.NewRows([]string{"legacy_id"}).AddRow(`{"t24AccountNumber":"21100100000001"}`)
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryCheckLegacyId)).WillReturnRows(rows)
				},
			},
			wantErr: false,
		},
		{
			name: "test error result",
			args: args{
				ctx: context.TODO(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryCheckLegacyId)).WillReturnError(assert.AnError)
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.args.setupMocks()

			_, err := suite.repo.CheckLegacyIdIsExist(tt.args.ctx, tt.args.accountNumber)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountTestSuite) TestRepository_CheckAccountNumberIsExist() {
	type args struct {
		ctx           context.Context
		accountNumber string
		setupMocks    func()
	}
	testCases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test success",
			args: args{ctx: context.TODO(),
				accountNumber: "21100100000001",
				setupMocks: func() {
					rows := sqlmock.NewRows([]string{"account_number", "entity_code"}).AddRow(1, "001")
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryCheckAccountNumber)).WillReturnRows(rows)
				},
			},
			wantErr: false,
		},
		{
			name: "test error result",
			args: args{
				ctx: context.TODO(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryCheckAccountNumber)).WillReturnError(assert.AnError)
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.args.setupMocks()

			_, err := suite.repo.CheckAccountNumberIsExist(tt.args.ctx, tt.args.accountNumber)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *accountTestSuite) TestRepository_GetLoanAdvanceAccountByLoanAccount() {
	type args struct {
		ctx               context.Context
		loanAccountNumber string
		setupMocks        func()
	}

	testCases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test success",
			args: args{ctx: context.TODO(),
				loanAccountNumber: "21100100000001",
				setupMocks: func() {
					rows := sqlmock.NewRows([]string{"loan_account_number", "loan_advance_payment_account_number"}).
						AddRow("21100100000001", "21100100000002")
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryGetLoanAdvanceAccountByLoanAccount)).WillReturnRows(rows)
				},
			},
			wantErr: false,
		},
		{
			name: "test data not found",
			args: args{
				ctx: context.TODO(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryGetLoanAdvanceAccountByLoanAccount)).WillReturnError(sql.ErrNoRows)
				},
			},
			wantErr: true,
		},
		{
			name: "test error result",
			args: args{
				ctx: context.TODO(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryGetLoanAdvanceAccountByLoanAccount)).WillReturnError(assert.AnError)
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.args.setupMocks()

			_, err := suite.repo.GetLoanAdvanceAccountByLoanAccount(tt.args.ctx, tt.args.loanAccountNumber)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
