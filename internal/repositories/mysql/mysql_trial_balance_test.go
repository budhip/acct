package mysql

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestTrialBalanceRepositoryTestSuite(t *testing.T) {
	t.Helper()
	suite.Run(t, new(trialBalanceTestSuite))
}

type trialBalanceTestSuite struct {
	suite.Suite
	t    *testing.T
	db   *sql.DB
	mock sqlmock.Sqlmock
	repo TrialBalanceRepository
}

func (suite *trialBalanceTestSuite) SetupTest() {
	var err error

	suite.db, suite.mock, err = sqlmock.New()
	require.NoError(suite.T(), err)

	suite.t = suite.T()
	suite.repo = NewMySQLRepository(suite.db).GetTrialBalanceRepository()
}

func (suite *trialBalanceTestSuite) TearDownTest() {
	defer suite.db.Close()
}

func (suite *trialBalanceTestSuite) TestRepository_Close() {
	type args struct {
		ctx context.Context
		req models.CloseTrialBalanceRequest
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
				req: models.CloseTrialBalanceRequest{
					Period:   "2023-10",
					ClosedBy: "test",
				},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryTrialBalanceClose)).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "error",
			args: args{
				ctx: context.TODO(),
				req: models.CloseTrialBalanceRequest{
					Period:   "2023-10",
					ClosedBy: "test",
				},
			},
			doMock: func(args args) {
				suite.mock.
					ExpectExec(regexp.QuoteMeta(queryTrialBalanceClose)).
					WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range testCases {
		tt := tt
		suite.t.Run(tt.name, func(t *testing.T) {
			tt.doMock(tt.args)

			err := suite.repo.Close(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
