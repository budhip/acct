package mysql

import (
	"context"
	"regexp"
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func (suite *accountingTestSuite) TestRepository_GetBalanceSheet() {
	rowsBalanceSheet := sqlmock.NewRows(
		[]string{
			"ac.coa_type_code",
			"aatb.category_code",
			"ac.name",
			"sum(aatb.closing_balance)",
		}).AddRow("AST", "111", "Cash Point", "100000")
	defaultDate := atime.Now()

	type args struct {
		ctx  context.Context
		opts models.BalanceSheetFilterOptions
	}
	testCases := []struct {
		name       string
		args       args
		setupMocks func(a args)
		wantErr    bool
	}{
		{
			name: "success case - get balance sheet",
			args: args{
				ctx: context.TODO(),
				opts: models.BalanceSheetFilterOptions{
					EntityCode: "111",
					BalanceSheetDate:  defaultDate,
				},
			},
			setupMocks: func(a args) {
				query, _, _ := buildGetBalanceSheetQuery(a.opts)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(rowsBalanceSheet)
			},
			wantErr: false,
		},
		{
			name: "error case - error scan row",
			args: args{
				ctx: context.TODO(),
				opts: models.BalanceSheetFilterOptions{
					EntityCode: "111",
					BalanceSheetDate:  defaultDate,
				},
			},
			setupMocks: func(a args) {
				query, _, _ := buildGetBalanceSheetQuery(a.opts)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(sqlmock.NewRows([]string{"InvalidColumn"}).AddRow(nil))
			},
			wantErr: true,
		},
		{
			name: "error case - error database",
			args: args{
				ctx: context.TODO(),
				opts: models.BalanceSheetFilterOptions{
					EntityCode: "111",
					BalanceSheetDate:  defaultDate,
				},
			},
			setupMocks: func(a args) {
				query, _, _ := buildGetBalanceSheetQuery(a.opts)
				suite.mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		suite.t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks(tc.args)

			_, err := suite.repo.GetBalanceSheet(tc.args.ctx, tc.args.opts)
			assert.Equal(t, tc.wantErr, err != nil)

			if err = suite.mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
