package dbutil

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func Test_db_Exec(t *testing.T) {
	type args struct {
		ctx  context.Context
		sql  string
		args []interface{}
	}
	tests := []struct {
		name    string
		args    args
		doMock  func(a *args, w sql.Result, h *dbUtilHelper)
		want    sql.Result
		wantErr bool
	}{
		{
			name: "success exec",
			args: args{
				sql:  "insert into table(a, b, c) values (1, 2, 3)",
				args: nil,
			},
			doMock: func(a *args, w sql.Result, h *dbUtilHelper) {
				h.mockPrimary.ExpectExec(a.sql).WillReturnResult(w)
			},
			want: sqlmock.NewResult(0, 0),
		},
		{
			name: "failed exec",
			args: args{
				sql:  "insert into table(a, b, c) values (1, 2, 3)",
				args: nil,
			},
			doMock: func(a *args, w sql.Result, h *dbUtilHelper) {
				h.mockPrimary.ExpectExec(a.sql).WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHelper(t)

			if tt.doMock != nil {
				tt.doMock(&tt.args, tt.want, &h)
			}

			d := DbConn{
				primary:  h.primary,
				replicas: h.replicas,
			}
			_, err := d.Exec(tt.args.sql, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Exec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err = h.mockPrimary.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

			for i, slave := range h.mockReplicas {
				if err = slave.ExpectationsWereMet(); err != nil {
					t.Errorf("on slave index: %d, there were unfulfilled expectations: %s", i, err)
				}
			}
		})
	}
}

func Test_db_Prepare(t *testing.T) {
	type args struct {
		ctx  context.Context
		sql  string
		args []interface{}
	}
	tests := []struct {
		name    string
		args    args
		doMock  func(a *args, h *dbUtilHelper)
		wantErr bool
	}{
		{
			name: "success prepare",
			args: args{
				sql:  "insert into table(a, b, c) values (?, ?, ?)",
				args: nil,
			},
			doMock: func(a *args, h *dbUtilHelper) {
				h.mockPrimary.ExpectPrepare(a.sql)
			},
		},
		{
			name: "failed prepare",
			args: args{
				sql:  "insert into table(a, b, c) values (?, ?, ?)",
				args: nil,
			},
			doMock: func(a *args, h *dbUtilHelper) {
				h.mockPrimary.ExpectPrepare(a.sql).WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHelper(t)

			if tt.doMock != nil {
				tt.doMock(&tt.args, &h)
			}

			d := DbConn{
				primary:  h.primary,
				replicas: h.replicas,
			}
			_, err := d.Prepare(tt.args.sql)
			if (err != nil) != tt.wantErr {
				t.Errorf("Prepare() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err = h.mockPrimary.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

			for i, slave := range h.mockReplicas {
				if err = slave.ExpectationsWereMet(); err != nil {
					t.Errorf("on slave index: %d, there were unfulfilled expectations: %s", i, err)
				}
			}
		})
	}
}

func Test_db_Query(t *testing.T) {
	type args struct {
		ctx  context.Context
		sql  string
		args []interface{}
	}
	tests := []struct {
		name    string
		args    args
		doMock  func(a *args, h *dbUtilHelper)
		wantErr bool
	}{
		{
			name: "success query",
			args: args{
				sql:  `insert into table(a, b, c) values (?, ?, ?) returning ("a", "b", "c")`,
				args: nil,
			},
			doMock: func(a *args, h *dbUtilHelper) {
				h.mockPrimary.
					ExpectQuery(a.sql).
					WillReturnRows(sqlmock.NewRows([]string{"a", "b", "c"}))
			},
		},
		{
			name: "failed query",
			args: args{
				sql:  `insert into table(a, b, c) values (?, ?, ?) returning ("a", "b", "c")`,
				args: nil,
			},
			doMock: func(a *args, h *dbUtilHelper) {
				h.mockPrimary.ExpectQuery(a.sql).WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHelper(t)

			if tt.doMock != nil {
				tt.doMock(&tt.args, &h)
			}

			d := DbConn{
				primary:  h.primary,
				replicas: h.replicas,
			}
			_, err := d.Query(tt.args.sql, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Query() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err = h.mockPrimary.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

			for i, slave := range h.mockReplicas {
				if err := slave.ExpectationsWereMet(); err != nil {
					t.Errorf("on slave index: %d, there were unfulfilled expectations: %s", i, err)
				}
			}
		})
	}
}

func Test_db_QueryRow(t *testing.T) {
	type args struct {
		ctx  context.Context
		sql  string
		args []interface{}
	}
	tests := []struct {
		name    string
		args    args
		doMock  func(a *args, h *dbUtilHelper)
		wantErr bool
	}{
		{
			name: "success query - on primary",
			args: args{
				sql:  `update tables set col=1 where id=1 returning col`,
				args: nil,
			},
			doMock: func(a *args, h *dbUtilHelper) {
				h.mockPrimary.
					ExpectQuery(a.sql).
					WillReturnRows(sqlmock.NewRows([]string{"1"}))
			},
		},
		{
			name: "success query - on replica",
			args: args{
				sql:  `select * from table limit 1`,
				args: nil,
			},
			doMock: func(a *args, h *dbUtilHelper) {
				h.mockReplicas[1].
					ExpectQuery(a.sql).
					WillReturnRows(sqlmock.NewRows([]string{"a", "b", "c"}))
			},
		},
		{
			name: "failed query",
			args: args{
				sql:  `select * from table limit 1`,
				args: nil,
			},
			doMock: func(a *args, h *dbUtilHelper) {
				h.mockReplicas[1].ExpectQuery(a.sql).
					WillReturnError(assert.AnError)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHelper(t)

			if tt.doMock != nil {
				tt.doMock(&tt.args, &h)
			}

			d := DbConn{
				primary:  h.primary,
				replicas: h.replicas,
			}
			_ = d.QueryRow(tt.args.sql, tt.args.args...)

			if err := h.mockPrimary.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

			for i, slave := range h.mockReplicas {
				if err := slave.ExpectationsWereMet(); err != nil {
					t.Errorf("on slave index: %d, there were unfulfilled expectations: %s", i, err)
				}
			}
		})
	}
}

func Test_isUseDBMaster(t *testing.T) {
	type args struct {
		ctx   context.Context
		query string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "select statement",
			args: args{
				query: "select * from tables limit 1000",
			},
			want: false,
		},
		{
			name: "update statement",
			args: args{
				query: "update tables set id=1 where id=2",
			},
			want: true,
		},
		{
			name: "select statement - but override using context as primary",
			args: args{
				ctx:   NewContextUsePrimaryDB(context.Background()),
				query: "select * from tables limit 10",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, isUseDBPrimary(tt.args.ctx, tt.args.query), "isUseDBMaster(%v, %v)", tt.args.ctx, tt.args.query)
		})
	}
}
