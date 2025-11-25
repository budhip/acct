package dbutil

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/mock/gomock"
)

type dbUtilHelper struct {
	mockCtrl *gomock.Controller

	primary     *sql.DB
	mockPrimary sqlmock.Sqlmock

	replicas     []*sql.DB
	mockReplicas []sqlmock.Sqlmock
}

func newHelper(t *testing.T) (h dbUtilHelper) {
	t.Helper()

	primary, mockPrimary, err := sqlmock.New(
		sqlmock.MonitorPingsOption(true),
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	h.primary = primary
	h.mockPrimary = mockPrimary

	for i := 0; i < 2; i++ {
		slave, mockSlave, errNewSlave := sqlmock.New(
			sqlmock.MonitorPingsOption(true),
			sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if errNewSlave != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", errNewSlave)
		}

		h.replicas = append(h.replicas, slave)
		h.mockReplicas = append(h.mockReplicas, mockSlave)
	}

	return
}

func TestNew(t *testing.T) {
	helper := newHelper(t)

	type args struct {
		primary  *sql.DB
		replicas []*sql.DB
	}
	tests := []struct {
		name string
		args args
		want DB
	}{
		{
			name: "success init connection",
			args: args{
				primary:  helper.primary,
				replicas: helper.replicas,
			},
			want: New(helper.primary, helper.replicas...),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.primary, tt.args.replicas...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_db_Primary(t *testing.T) {
	helper := newHelper(t)

	type fields struct {
		primary  *sql.DB
		replicas []*sql.DB
		counter  uint64
	}
	tests := []struct {
		name   string
		fields fields
		want   *sql.DB
	}{
		{
			name: "success get primary",
			fields: fields{
				primary:  helper.primary,
				replicas: helper.replicas,
			},
			want: helper.primary,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := DbConn{
				primary:  tt.fields.primary,
				replicas: tt.fields.replicas,
				counter:  tt.fields.counter,
			}
			if got := d.Primary(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Master() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_db_selectReplica(t *testing.T) {
	helper := newHelper(t)

	type fields struct {
		primary  *sql.DB
		replicas []*sql.DB
		counter  uint64
	}
	tests := []struct {
		name   string
		fields fields
		want   *sql.DB
	}{
		{
			name: "success get replica",
			fields: fields{
				primary:  helper.primary,
				replicas: helper.replicas,
			},
			want: helper.replicas[1], // round-robin after counter value 0
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := DbConn{
				primary:  tt.fields.primary,
				replicas: tt.fields.replicas,
				counter:  tt.fields.counter,
			}
			if got := d.selectReplica(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Slave() = %v, want %v", got, tt.want)
			}
		})
	}
}
