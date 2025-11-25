package dbutil

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_db_Begin(t *testing.T) {
	tests := []struct {
		name    string
		doMock  func(h *dbUtilHelper)
		wantErr bool
	}{
		{
			name: "success begin",
			doMock: func(h *dbUtilHelper) {
				h.mockPrimary.ExpectBegin()
			},
		},
		{
			name: "failed begin",
			doMock: func(h *dbUtilHelper) {
				h.mockPrimary.ExpectBegin().WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHelper(t)

			if tt.doMock != nil {
				tt.doMock(&h)
			}

			d := DbConn{
				primary:  h.primary,
				replicas: h.replicas,
			}
			_, err := d.Begin()
			if (err != nil) != tt.wantErr {
				t.Errorf("Begin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err = h.mockPrimary.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

			for _, replica := range h.mockReplicas {
				if err = replica.ExpectationsWereMet(); err != nil {
					t.Errorf("there were unfulfilled expectations: %s", err)
				}
			}
		})
	}
}

func Test_db_Close(t *testing.T) {
	tests := []struct {
		name    string
		doMock  func(h *dbUtilHelper)
		wantErr bool
	}{
		{
			name: "success close",
			doMock: func(h *dbUtilHelper) {
				h.mockPrimary.ExpectClose()
				for _, replica := range h.mockReplicas {
					replica.ExpectClose()
				}
			},
		},
		{
			name: "failed close",
			doMock: func(h *dbUtilHelper) {
				h.mockPrimary.ExpectClose()
				for _, replica := range h.mockReplicas {
					replica.ExpectClose().WillReturnError(assert.AnError)
				}
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHelper(t)

			if tt.doMock != nil {
				tt.doMock(&h)
			}

			d := DbConn{
				primary:  h.primary,
				replicas: h.replicas,
			}
			err := d.Close()
			if (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err = h.mockPrimary.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

			for _, replica := range h.mockReplicas {
				if err = replica.ExpectationsWereMet(); err != nil {
					t.Errorf("there were unfulfilled expectations: %s", err)
				}
			}
		})
	}
}

func Test_db_PingContext(t *testing.T) {
	tests := []struct {
		name    string
		doMock  func(h *dbUtilHelper)
		wantErr bool
	}{
		{
			name: "success ping",
			doMock: func(h *dbUtilHelper) {
				h.mockPrimary.ExpectPing()
				for _, replica := range h.mockReplicas {
					replica.ExpectPing()
				}
			},
		},
		{
			name: "failed ping",
			doMock: func(h *dbUtilHelper) {
				h.mockPrimary.ExpectPing()
				for _, replica := range h.mockReplicas {
					replica.ExpectPing().WillReturnError(assert.AnError)
				}
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHelper(t)

			if tt.doMock != nil {
				tt.doMock(&h)
			}

			d := DbConn{
				primary:  h.primary,
				replicas: h.replicas,
			}
			err := d.PingContext(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("PingContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err = h.mockPrimary.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

			for _, replica := range h.mockReplicas {
				if err = replica.ExpectationsWereMet(); err != nil {
					t.Errorf("there were unfulfilled expectations: %s", err)
				}
			}
		})
	}
}
