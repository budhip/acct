package dbutil

import (
	"context"
	"database/sql"
	"sync/atomic"
	"time"
)

type DB interface {
	Stmt
	Begin() (*sql.Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Close() error
	PingContext(ctx context.Context) error
	SetConnMaxLifetime(d time.Duration)
	SetMaxIdleConns(n int)
	SetMaxOpenConns(n int)
	Stats() sql.DBStats
}

// Stmt is a sql prepared statement
type Stmt interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type DbConn struct {
	primary  *sql.DB
	replicas []*sql.DB

	counter uint64
}

// New creates read write query splitting DbConn connection pool.
//
// To override behaviour DbConn connection and directly use primary connection,
// create context using `dbutil.NewContextUsePrimaryDB(parentCtx)`
// and pass this value inside service/repository function or directly pass ctx to Stmt func.
func New(primary *sql.DB, replicas ...*sql.DB) *DbConn {
	return &DbConn{
		primary:  primary,
		replicas: replicas,
	}
}

func (d DbConn) Primary() *sql.DB {
	return d.primary
}

func (d DbConn) Replicas() []*sql.DB {
	return d.replicas
}

func (d DbConn) selectReplica() *sql.DB {
	totalSlaves := uint64(len(d.replicas))

	if totalSlaves == 0 {
		return d.primary
	}

	// round-robin
	index := atomic.AddUint64(&d.counter, 1) % totalSlaves
	return d.replicas[index]
}
