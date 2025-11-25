package mysql

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestEntityRepositoryTestSuite(t *testing.T) {
	t.Helper()
	suite.Run(t, new(entityTestSuite))
}

type entityTestSuite struct {
	suite.Suite
	t    *testing.T
	db   *sql.DB
	mock sqlmock.Sqlmock
	repo EntityRepository
}

func (suite *entityTestSuite) SetupTest() {
	var err error

	suite.db, suite.mock, err = sqlmock.New()
	require.NoError(suite.T(), err)

	suite.t = suite.T()
	suite.repo = NewMySQLRepository(suite.db).GetEntityRepository()
}

func (suite *entityTestSuite) TearDownTest() {
	defer suite.db.Close()
}

func (suite *entityTestSuite) TestRepository_CheckEntityByCode() {
	type args struct {
		ctx        context.Context
		code       string
		status     string
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
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryEntityIsExistByCode)).WillReturnRows(rows)
				},
			},
			wantErr: false,
		},
		{
			name: "test success with status",
			args: args{
				ctx:    context.Background(),
				code:   "211",
				status: "active",
				setupMocks: func() {
					rows := sqlmock.NewRows(
						[]string{"code"}).AddRow("211")
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryEntityIsExistByCode + queryEntityAndStatus)).WillReturnRows(rows)
				},
			},
			wantErr: false,
		},
		{
			name: "test data not found",
			args: args{
				ctx: context.TODO(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryEntityIsExistByCode)).WillReturnError(sql.ErrNoRows)
				},
			},
			wantErr: true,
		},
		{
			name: "test error result",
			args: args{
				ctx: context.TODO(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryEntityIsExistByCode)).WillReturnError(assert.AnError)
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.args.setupMocks()

			err := suite.repo.CheckEntityByCode(tt.args.ctx, tt.args.code, tt.args.status)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func (suite *entityTestSuite) TestRepository_Create() {
	type args struct {
		ctx context.Context
		req models.CreateEntityIn
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
				req: models.CreateEntityIn{
					Code:        "test",
					Name:        "test",
					Description: "test",
					Status:      models.StatusActive,
				},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryEntityCreate)).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "error no result",
			args: args{
				ctx: context.TODO(),
				req: models.CreateEntityIn{},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryEntityCreate)).WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
		{
			name: "error result",
			args: args{
				ctx: context.TODO(),
				req: models.CreateEntityIn{},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryEntityCreate)).WillReturnResult(sqlmock.NewErrorResult(assert.AnError))
			},
			wantErr: true,
		},
		{
			name: "error db",
			args: args{
				ctx: context.TODO(),
				req: models.CreateEntityIn{},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryEntityCreate)).WillReturnError(assert.AnError)
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

func (suite *entityTestSuite) TestRepository_GetByCode() {
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
						[]string{"id", "code", "name", "description", "status", "created_at", "updated_at"}).
						AddRow(1, "666", "ENT", "this is description", "active", time.Now(), time.Now())
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryEntityGetByCode)).WillReturnRows(rows)
				},
			},
			wantErr: false,
		},
		{
			name: "test data not found",
			args: args{
				ctx: context.TODO(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryEntityGetByCode)).WillReturnError(sql.ErrNoRows)
				},
			},
			wantErr: false,
		},
		{
			name: "test error result",
			args: args{
				ctx: context.TODO(),
				setupMocks: func() {
					suite.mock.ExpectQuery(regexp.QuoteMeta(queryEntityGetByCode)).WillReturnError(assert.AnError)
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

func (suite *entityTestSuite) TestRepository_List() {
	testCases := []struct {
		name    string
		wantErr bool
		doMock  func()
	}{
		{
			name: "success get list",
			doMock: func() {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryEntityList)).
					WillReturnRows(
						sqlmock.
							NewRows([]string{"id", "description", "code", "name", "status", "created_at", "updated_at"}).
							AddRow(1, "this is description", "666", "ENT", "active", time.Now(), time.Now()),
					)
			},
			wantErr: false,
		},
		{
			name: "error row",
			doMock: func() {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryEntityList)).
					WillReturnRows(
						sqlmock.
							NewRows([]string{"id", "description", "code", "name", "status", "created_at", "updated_at"}).
							AddRow(1, "this is description", "666", "ENT", "active", time.Now(), time.Now()).RowError(0, assert.AnError),
					)
			},
			wantErr: true,
		},
		{
			name: "failed scan row",
			doMock: func() {
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(queryEntityList)).
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
					ExpectQuery(regexp.QuoteMeta(queryEntityList)).
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

func (suite *entityTestSuite) TestRepository_Update() {
	type args struct {
		ctx        context.Context
		req        models.UpdateEntity
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
				req: models.UpdateEntity{
					Name:        "Lender Yang Baik",
					Description: "Testing",
					Code:        "001",
					Status:      "active",
				},
				setupMocks: func() {
					suite.mock.ExpectExec(regexp.QuoteMeta(queryEntityUpdate)).WillReturnResult(sqlmock.NewResult(0, 1))
				},
			},
			wantErr: false,
		},
		{name: "test error db",
			args: args{
				ctx: context.TODO(),
				req: models.UpdateEntity{
					Name:        "Lender Yang Baik",
					Description: "Testing",
					Code:        "001",
					Status:      "active",
				},
				setupMocks: func() {
					suite.mock.ExpectExec(regexp.QuoteMeta(queryEntityUpdate)).WillReturnError(assert.AnError)
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

func (suite *entityTestSuite) TestRepository_GetByParam() {
	type args struct {
		ctx        context.Context
		in         models.GetEntity
		setupMocks func(req models.GetEntity)
	}

	testCases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test success",
			args: args{
				ctx: context.Background(),
				in: models.GetEntity{
					Code:   "001",
					Name:   "AMF",
					Status: models.StatusActive,
				},
				setupMocks: func(req models.GetEntity) {
					rows := sqlmock.NewRows(
						[]string{"id", "code", "name", "description", "status", "created_at", "updated_at"}).
						AddRow(1, "666", "ENT", "this is description", "active", time.Now(), time.Now())

					query, _, _ := builQueryGetEntityByParam(req)
					suite.mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(rows)
				},
			},
			wantErr: false,
		},
		{
			name: "test data not found",
			args: args{
				ctx: context.Background(),
				in: models.GetEntity{
					Code:   "001",
					Name:   "AWF",
					Status: models.StatusActive,
				},
				setupMocks: func(req models.GetEntity) {
					query, _, _ := builQueryGetEntityByParam(req)
					suite.mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(sql.ErrNoRows)
				},
			},
			wantErr: true,
		},
		{
			name: "test error db",
			args: args{
				ctx: context.Background(),
				in: models.GetEntity{
					Code:   "001",
					Name:   "AWF",
					Status: models.StatusActive,
				},
				setupMocks: func(req models.GetEntity) {
					query, _, _ := builQueryGetEntityByParam(req)
					suite.mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(errors.New("err"))
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.args.setupMocks(models.GetEntity{})

			_, err := suite.repo.GetByParams(tt.args.ctx, tt.args.in)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
