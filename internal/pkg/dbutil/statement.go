package dbutil

import (
	"context"
	"database/sql"
	"regexp"
)

var (
	contextKeyUsePrimary = "usePrimaryDBConn"
	mutationRegex        = regexp.MustCompile(`(?i)\b(INSERT|UPDATE|DELETE|MERGE|RETURNING)\b`)
)

// NewContextUsePrimaryDB returns a new context with a value to tell the dbutil to use master DbConn connection.
func NewContextUsePrimaryDB(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKeyUsePrimary, true)
}

func isContextUsePrimary(ctx context.Context) bool {
	if ctx == nil {
		return false
	}

	if v, ok := ctx.Value(contextKeyUsePrimary).(bool); ok {
		return v
	}

	return false
}

func isMutationQuery(query string) bool {
	// We can't control what query values will be passed into the functions `Query` and `QueryRow`.
	// For example, someone might do `DbConn.Query("INSERT INTO t(id, id2) VALUES 2,3 RETURNING id,id2;")` and expect it to work.
	// Therefore, we need to make sure we support these statements too,
	return mutationRegex.MatchString(query)
}

func isUseDBPrimary(ctx context.Context, query string) bool {
	return isContextUsePrimary(ctx) || isMutationQuery(query)
}

func (d DbConn) Exec(query string, args ...interface{}) (sql.Result, error) {
	return d.ExecContext(context.Background(), query, args...)
}

func (d DbConn) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return d.Primary().ExecContext(ctx, query, args...)
}

func (d DbConn) Prepare(query string) (*sql.Stmt, error) {
	return d.PrepareContext(context.Background(), query)
}

func (d DbConn) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	// Currently, PrepareContext returns a `*sql.Stmt` bound to a single database instance.
	// So, all executions of this statement will use that same instance,
	// which doesn't leverage load balancing across multiple instances. so:
	//
	// TODO: prepare the statement on all instances to enable load balancing
	// TODO: implement Stmt interface to support load-balanced query executions
	//
	// Note: Instead of implementing the tasks above, it might be preferable to use a middleware proxy
	// to handle HA and load balancing when using multiple read replicas.
	if isUseDBPrimary(ctx, query) {
		return d.Primary().PrepareContext(ctx, query)
	}

	return d.selectReplica().PrepareContext(ctx, query)
}

func (d DbConn) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return d.QueryContext(context.Background(), query, args...)
}

func (d DbConn) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if isUseDBPrimary(ctx, query) {
		return d.Primary().QueryContext(ctx, query, args...)
	}

	return d.selectReplica().QueryContext(ctx, query, args...)
}

func (d DbConn) QueryRow(query string, args ...interface{}) *sql.Row {
	return d.QueryRowContext(context.Background(), query, args...)
}

func (d DbConn) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if isUseDBPrimary(ctx, query) {
		return d.Primary().QueryRowContext(ctx, query, args...)
	}

	return d.selectReplica().QueryRowContext(ctx, query, args...)
}
